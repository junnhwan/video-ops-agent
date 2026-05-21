package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type routerConfig struct {
	agentHandler   *AgentHandler
	gatewayHandler routeRegistrar
	skillHandler   routeRegistrar
}

type RouterOption func(*routerConfig)

type routeRegistrar interface {
	RegisterRoutes(router *gin.Engine)
}

func WithAgentHandler(handler *AgentHandler) RouterOption {
	return func(config *routerConfig) {
		config.agentHandler = handler
	}
}

func WithGatewayHandler(handler routeRegistrar) RouterOption {
	return func(config *routerConfig) {
		config.gatewayHandler = handler
	}
}

func WithSkillHandler(handler routeRegistrar) RouterOption {
	return func(config *routerConfig) {
		config.skillHandler = handler
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
	if config.gatewayHandler != nil {
		config.gatewayHandler.RegisterRoutes(router)
	}
	if config.skillHandler != nil {
		config.skillHandler.RegisterRoutes(router)
	}

	return router
}
