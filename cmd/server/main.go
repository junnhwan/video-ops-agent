package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"video-ops-agent/internal/agent/contextbuilder"
	"video-ops-agent/internal/agent/llm"
	agentruntime "video-ops-agent/internal/agent/runtime"
	"video-ops-agent/internal/agent/skills"
	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/config"
	"video-ops-agent/internal/eval"
	"video-ops-agent/internal/gateway"
	httpapi "video-ops-agent/internal/http"
	"video-ops-agent/internal/platform/videofeed"
	"video-ops-agent/internal/store"
)

func main() {
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := store.OpenSQLite(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() {
		if err := store.Close(db); err != nil {
			log.Printf("close database: %v", err)
		}
	}()
	if err := store.AutoMigrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	repos := contextbuilder.Repositories{
		Sessions:  store.NewSessionRepository(db),
		Messages:  store.NewMessageRepository(db),
		ToolCalls: store.NewToolCallRepository(db),
	}
	videoFeedClient, err := videofeed.NewClient(cfg.VideoFeed.BaseURL)
	if err != nil {
		log.Fatalf("create video-feed client: %v", err)
	}
	toolRegistry, err := tools.NewDefaultRegistry(videoFeedClient)
	if err != nil {
		log.Fatalf("create tool registry: %v", err)
	}
	llmClient, err := llm.NewClient(llm.ClientConfig{
		BaseURL: cfg.LLM.BaseURL,
		Model:   cfg.LLM.Model,
		APIKey:  cfg.LLM.APIKey,
	})
	if err != nil {
		log.Fatalf("create llm client: %v", err)
	}
	toolExecutor := tools.NewExecutor(toolRegistry, 2*time.Second)
	skillService := skills.NewService(skills.Dependencies{
		Registry:   toolRegistry,
		Repository: store.NewSkillRepository(db),
	})
	gatewayService := gateway.NewService(gateway.Dependencies{
		Registry:    toolRegistry,
		Executor:    toolExecutor,
		Invocations: store.NewGatewayInvocationRepository(db),
	})
	agentRuntime := agentruntime.NewRuntime(agentruntime.Dependencies{
		LLM:                llmClient,
		ToolRegistry:       toolRegistry,
		ToolExecutor:       toolExecutor,
		ContextBuilder:     contextbuilder.NewBuilder(repos),
		Repositories:       repos,
		SkillService:       skillService,
		InvocationRecorder: gatewayService,
	}, agentruntime.RuntimeConfig{})
	agentHandler := httpapi.NewAgentHandler(repos, agentRuntime)
	gatewayHandler := gateway.NewHandler(gatewayService)
	skillHandler := httpapi.NewSkillHandler(skillService)
	evalService := eval.NewService(eval.Dependencies{
		Sessions:    repos.Sessions,
		Invocations: store.NewGatewayInvocationRepository(db),
		Skills:      skillService,
	})
	evalHandler := eval.NewHandler(evalService)

	server := &http.Server{
		Addr: cfg.Server.Address,
		Handler: httpapi.NewRouter(
			httpapi.WithAgentHandler(agentHandler),
			httpapi.WithGatewayHandler(gatewayHandler),
			httpapi.WithSkillHandler(skillHandler),
			httpapi.WithEvalHandler(evalHandler),
		),
	}

	log.Printf("video-ops-agent listening on %s", cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve: %v", err)
	}
}
