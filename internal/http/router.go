package httpapi

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type routerConfig struct {
	agentHandler   *AgentHandler
	gatewayHandler routeRegistrar
	skillHandler   routeRegistrar
	evalHandler    routeRegistrar
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

func WithEvalHandler(handler routeRegistrar) RouterOption {
	return func(config *routerConfig) {
		config.evalHandler = handler
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
	router.Use(apiKeyMiddleware())

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
	if config.evalHandler != nil {
		config.evalHandler.RegisterRoutes(router)
	}

	return router
}

func apiKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		required := strings.TrimSpace(os.Getenv("VIDEO_OPS_API_KEY"))
		if required == "" {
			c.Next()
			return
		}
		got := c.GetHeader("X-API-Key")
		if got == "" {
			got = c.Query("api_key")
		}
		if got != required {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		c.Next()
	}
}
