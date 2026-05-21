package main

import (
	"context"
	"log"
	"os"
	"time"

	"video-ops-agent/internal/agent/skills"
	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/config"
	"video-ops-agent/internal/gateway"
	"video-ops-agent/internal/mcp"
	"video-ops-agent/internal/platform/videofeed"
	"video-ops-agent/internal/store"
)

func main() {
	log.SetOutput(os.Stderr)

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

	videoFeedClient, err := videofeed.NewClient(cfg.VideoFeed.BaseURL)
	if err != nil {
		log.Fatalf("create video-feed client: %v", err)
	}
	toolRegistry, err := tools.NewDefaultRegistry(videoFeedClient)
	if err != nil {
		log.Fatalf("create tool registry: %v", err)
	}
	gatewayService := gateway.NewService(gateway.Dependencies{
		Registry:    toolRegistry,
		Executor:    tools.NewExecutor(toolRegistry, 2*time.Second),
		Invocations: store.NewGatewayInvocationRepository(db),
	})
	skillService := skills.NewService(skills.Dependencies{
		Registry:   toolRegistry,
		Repository: store.NewSkillRepository(db),
	})
	server := mcp.NewServer(
		mcp.NewToolAdapter(gatewayService),
		mcp.NewResourceAdapter(gatewayService, skillService),
		mcp.NewPromptAdapter(skillService),
	)
	if err := server.Serve(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatalf("serve mcp: %v", err)
	}
}
