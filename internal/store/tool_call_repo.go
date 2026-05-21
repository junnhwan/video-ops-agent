package store

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type ToolCallRepository struct {
	db *gorm.DB
}

type CreateToolCallInput struct {
	SessionID     uint
	MessageID     *uint
	ToolName      string
	ArgumentsJSON string
	ResultJSON    string
	ResultSummary string
	LatencyMS     int64
	Status        string
	ErrorMessage  string
}

func NewToolCallRepository(db *gorm.DB) *ToolCallRepository {
	return &ToolCallRepository{db: db}
}

func (r *ToolCallRepository) Create(ctx context.Context, input CreateToolCallInput) (AgentToolCall, error) {
	if input.SessionID == 0 {
		return AgentToolCall{}, fmt.Errorf("session_id is required")
	}
	if strings.TrimSpace(input.ToolName) == "" {
		return AgentToolCall{}, fmt.Errorf("tool_name is required")
	}
	if strings.TrimSpace(input.ArgumentsJSON) == "" {
		return AgentToolCall{}, fmt.Errorf("arguments_json is required")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = ToolCallStatusSuccess
	}

	toolCall := AgentToolCall{
		SessionID:     input.SessionID,
		MessageID:     input.MessageID,
		ToolName:      strings.TrimSpace(input.ToolName),
		ArgumentsJSON: input.ArgumentsJSON,
		ResultJSON:    input.ResultJSON,
		ResultSummary: input.ResultSummary,
		LatencyMS:     input.LatencyMS,
		Status:        status,
		ErrorMessage:  input.ErrorMessage,
	}
	if err := r.db.WithContext(ctx).Create(&toolCall).Error; err != nil {
		return AgentToolCall{}, fmt.Errorf("create agent tool call: %w", err)
	}
	return toolCall, nil
}

func (r *ToolCallRepository) ListBySession(ctx context.Context, sessionID uint) ([]AgentToolCall, error) {
	if sessionID == 0 {
		return nil, fmt.Errorf("session_id is required")
	}

	var toolCalls []AgentToolCall
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("id ASC").
		Find(&toolCalls).Error; err != nil {
		return nil, fmt.Errorf("list agent tool calls for session %d: %w", sessionID, err)
	}
	return toolCalls, nil
}

func (r *ToolCallRepository) ListRecentBySession(ctx context.Context, sessionID uint, limit int) ([]AgentToolCall, error) {
	if sessionID == 0 {
		return nil, fmt.Errorf("session_id is required")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	var toolCalls []AgentToolCall
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("id DESC").
		Limit(limit).
		Find(&toolCalls).Error; err != nil {
		return nil, fmt.Errorf("list recent agent tool calls for session %d: %w", sessionID, err)
	}
	reverseToolCalls(toolCalls)
	return toolCalls, nil
}

func reverseToolCalls(toolCalls []AgentToolCall) {
	for left, right := 0, len(toolCalls)-1; left < right; left, right = left+1, right-1 {
		toolCalls[left], toolCalls[right] = toolCalls[right], toolCalls[left]
	}
}
