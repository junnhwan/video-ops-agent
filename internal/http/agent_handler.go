package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"video-ops-agent/internal/agent/contextbuilder"
	agentruntime "video-ops-agent/internal/agent/runtime"
	"video-ops-agent/internal/store"

	"github.com/gin-gonic/gin"
)

type AgentRuntime interface {
	Run(ctx context.Context, request agentruntime.RunRequest) (agentruntime.RunResult, error)
}

type AgentHandler struct {
	repos   contextbuilder.Repositories
	runtime AgentRuntime
}

func NewAgentHandler(repos contextbuilder.Repositories, runtime AgentRuntime) *AgentHandler {
	return &AgentHandler{repos: repos, runtime: runtime}
}

func (h *AgentHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/agent/sessions")
	group.POST("", h.CreateSession)
	group.GET("", h.ListSessions)
	group.GET("/:id", h.GetSession)
	group.POST("/:id/messages", h.PostMessage)
	group.GET("/:id/tool-calls", h.ListToolCalls)
}

func (h *AgentHandler) CreateSession(ctx *gin.Context) {
	var req struct {
		UserID            string          `json:"user_id"`
		Title             string          `json:"title"`
		Scenario          string          `json:"scenario"`
		ContextPolicy     json.RawMessage `json:"context_policy"`
		ContextPolicyJSON string          `json:"context_policy_json"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	contextPolicyJSON := strings.TrimSpace(req.ContextPolicyJSON)
	if len(req.ContextPolicy) > 0 && string(req.ContextPolicy) != "null" {
		if !json.Valid(req.ContextPolicy) {
			writeError(ctx, http.StatusBadRequest, fmt.Errorf("context_policy must be valid json"))
			return
		}
		contextPolicyJSON = string(req.ContextPolicy)
	}

	session, err := h.repos.Sessions.Create(ctx.Request.Context(), store.CreateSessionInput{
		UserID:            req.UserID,
		Title:             req.Title,
		Scenario:          req.Scenario,
		Status:            store.SessionStatusActive,
		ContextPolicyJSON: contextPolicyJSON,
	})
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"session": session})
}

func (h *AgentHandler) ListSessions(ctx *gin.Context) {
	limit, err := optionalIntQuery(ctx, "limit", 20)
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	sessions, err := h.repos.Sessions.List(ctx.Request.Context(), ctx.Query("user_id"), limit)
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *AgentHandler) GetSession(ctx *gin.Context) {
	sessionID, ok := parseIDParam(ctx, "id")
	if !ok {
		return
	}
	session, err := h.repos.Sessions.Get(ctx.Request.Context(), sessionID)
	if err != nil {
		writeError(ctx, statusForStoreError(err), err)
		return
	}
	messages, err := h.repos.Messages.ListBySession(ctx.Request.Context(), sessionID, 200)
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"session": session, "messages": messages})
}

func (h *AgentHandler) PostMessage(ctx *gin.Context) {
	sessionID, ok := parseIDParam(ctx, "id")
	if !ok {
		return
	}
	var req struct {
		Content          string   `json:"content"`
		RequiredEvidence []string `json:"required_evidence"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeError(ctx, http.StatusBadRequest, fmt.Errorf("content is required"))
		return
	}
	result, err := h.runtime.Run(ctx.Request.Context(), agentruntime.RunRequest{
		SessionID:        sessionID,
		UserMessage:      req.Content,
		RequiredEvidence: req.RequiredEvidence,
	})
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"session_id":      result.SessionID,
		"final_answer":    result.FinalAnswer,
		"round_count":     result.RoundCount,
		"tool_call_count": result.ToolCallCount,
	})
}

func (h *AgentHandler) ListToolCalls(ctx *gin.Context) {
	sessionID, ok := parseIDParam(ctx, "id")
	if !ok {
		return
	}
	toolCalls, err := h.repos.ToolCalls.ListBySession(ctx.Request.Context(), sessionID)
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"tool_calls": toolCalls})
}

func parseIDParam(ctx *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(ctx.Param(name), 10, 64)
	if err != nil || value == 0 {
		writeError(ctx, http.StatusBadRequest, fmt.Errorf("%s must be a positive integer", name))
		return 0, false
	}
	return uint(value), true
}

func optionalIntQuery(ctx *gin.Context, name string, defaultValue int) (int, error) {
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

func statusForStoreError(err error) int {
	if strings.Contains(err.Error(), "record not found") {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

func writeError(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, gin.H{"error": err.Error()})
}

func uintString(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}
