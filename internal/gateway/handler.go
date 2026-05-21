package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"video-ops-agent/internal/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/gateway")
	group.GET("/tools", h.ListTools)
	group.GET("/tools/:name", h.GetTool)
	group.POST("/tools/:name/call", h.CallTool)
	group.GET("/invocations", h.ListInvocations)
	group.GET("/invocations/:id", h.GetInvocation)
}

func (h *Handler) ListTools(ctx *gin.Context) {
	tools, err := h.service.ListTools(ctx.Request.Context())
	if err != nil {
		writeGatewayError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"tools": tools})
}

func (h *Handler) GetTool(ctx *gin.Context) {
	tool, err := h.service.GetTool(ctx.Request.Context(), ctx.Param("name"))
	if err != nil {
		writeGatewayError(ctx, statusForGatewayError(err), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"tool": tool})
}

func (h *Handler) CallTool(ctx *gin.Context) {
	toolName := strings.TrimSpace(ctx.Param("name"))
	if toolName == "" {
		writeGatewayError(ctx, http.StatusBadRequest, fmt.Errorf("tool name is required"))
		return
	}

	var req CallToolRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeGatewayError(ctx, http.StatusBadRequest, err)
		return
	}
	if len(req.Arguments) == 0 || strings.TrimSpace(string(req.Arguments)) == "null" {
		req.Arguments = json.RawMessage(`{}`)
	}
	if !json.Valid(req.Arguments) {
		writeGatewayError(ctx, http.StatusBadRequest, fmt.Errorf("arguments must be valid json"))
		return
	}

	output, err := h.service.CallTool(ctx.Request.Context(), CallToolInput{
		ToolName:     toolName,
		Arguments:    req.Arguments,
		Source:       req.Source,
		SessionID:    req.SessionID,
		SkillID:      req.SkillID,
		SkillVersion: req.SkillVersion,
	})
	if err != nil {
		writeGatewayError(ctx, statusForGatewayError(err), err)
		return
	}
	ctx.JSON(http.StatusOK, output)
}

func (h *Handler) ListInvocations(ctx *gin.Context) {
	limit, err := optionalGatewayIntQuery(ctx, "limit", 50)
	if err != nil {
		writeGatewayError(ctx, http.StatusBadRequest, err)
		return
	}
	var sessionID *uint
	if raw := strings.TrimSpace(ctx.Query("session_id")); raw != "" {
		parsed, err := parseGatewayUint(raw, "session_id")
		if err != nil {
			writeGatewayError(ctx, http.StatusBadRequest, err)
			return
		}
		sessionID = &parsed
	}

	invocations, err := h.service.ListInvocations(ctx.Request.Context(), store.GatewayInvocationFilter{
		Source:    ctx.Query("source"),
		ToolName:  ctx.Query("tool_name"),
		SessionID: sessionID,
		SkillID:   ctx.Query("skill_id"),
		Status:    ctx.Query("status"),
		Limit:     limit,
	})
	if err != nil {
		writeGatewayError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"invocations": invocations})
}

func (h *Handler) GetInvocation(ctx *gin.Context) {
	id, err := parseGatewayUint(ctx.Param("id"), "id")
	if err != nil {
		writeGatewayError(ctx, http.StatusBadRequest, err)
		return
	}
	invocation, err := h.service.GetInvocation(ctx.Request.Context(), id)
	if err != nil {
		writeGatewayError(ctx, statusForGatewayError(err), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"invocation": invocation})
}

func optionalGatewayIntQuery(ctx *gin.Context, name string, defaultValue int) (int, error) {
	raw := strings.TrimSpace(ctx.Query(name))
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return value, nil
}

func parseGatewayUint(raw string, name string) (uint, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return uint(value), nil
}

func statusForGatewayError(err error) int {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "record not found"), strings.Contains(message, "unknown tool"):
		return http.StatusNotFound
	case strings.Contains(message, "required"), strings.Contains(message, "invalid"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func writeGatewayError(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, gin.H{"error": err.Error()})
}
