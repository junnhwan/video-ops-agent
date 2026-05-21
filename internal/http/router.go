package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type routerConfig struct {
	agentHandler *AgentHandler
}

type RouterOption func(*routerConfig)

func WithAgentHandler(handler *AgentHandler) RouterOption {
	return func(config *routerConfig) {
		config.agentHandler = handler
	}
}

func NewRouter(options ...RouterOption) *gin.Engine {
	var config routerConfig
	for _, option := range options {
		option(&config)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	if config.agentHandler != nil {
		config.agentHandler.RegisterRoutes(router)
	}

	return router
}
