package store

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type GatewayInvocationRepository struct {
	db *gorm.DB
}

type CreateGatewayInvocationInput struct {
	Source        string
	SessionID     *uint
	MessageID     *uint
	SkillID       string
	SkillVersion  string
	ToolName      string
	ArgumentsJSON string
	ResultJSON    string
	ResultSummary string
	LatencyMS     int64
	Status        string
	ErrorMessage  string
}

type GatewayInvocationFilter struct {
	Source    string
	ToolName  string
	SessionID *uint
	SkillID   string
	Status    string
	Limit     int
}

func NewGatewayInvocationRepository(db *gorm.DB) *GatewayInvocationRepository {
	return &GatewayInvocationRepository{db: db}
}

func (r *GatewayInvocationRepository) Create(ctx context.Context, input CreateGatewayInvocationInput) (GatewayToolInvocation, error) {
	source := strings.TrimSpace(input.Source)
	if source == "" {
		return GatewayToolInvocation{}, fmt.Errorf("source is required")
	}
	toolName := strings.TrimSpace(input.ToolName)
	if toolName == "" {
		return GatewayToolInvocation{}, fmt.Errorf("tool_name is required")
	}
	if strings.TrimSpace(input.ArgumentsJSON) == "" {
		return GatewayToolInvocation{}, fmt.Errorf("arguments_json is required")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = ToolCallStatusSuccess
	}

	invocation := GatewayToolInvocation{
		Source:        source,
		SessionID:     input.SessionID,
		MessageID:     input.MessageID,
		SkillID:       strings.TrimSpace(input.SkillID),
		SkillVersion:  strings.TrimSpace(input.SkillVersion),
		ToolName:      toolName,
		ArgumentsJSON: input.ArgumentsJSON,
		ResultJSON:    input.ResultJSON,
		ResultSummary: input.ResultSummary,
		LatencyMS:     input.LatencyMS,
		Status:        status,
		ErrorMessage:  input.ErrorMessage,
	}
	if err := r.db.WithContext(ctx).Create(&invocation).Error; err != nil {
		return GatewayToolInvocation{}, fmt.Errorf("create gateway tool invocation: %w", err)
	}
	return invocation, nil
}

func (r *GatewayInvocationRepository) Get(ctx context.Context, id uint) (GatewayToolInvocation, error) {
	if id == 0 {
		return GatewayToolInvocation{}, fmt.Errorf("invocation id is required")
	}
	var invocation GatewayToolInvocation
	if err := r.db.WithContext(ctx).First(&invocation, id).Error; err != nil {
		return GatewayToolInvocation{}, fmt.Errorf("get gateway tool invocation %d: %w", id, err)
	}
	return invocation, nil
}

func (r *GatewayInvocationRepository) List(ctx context.Context, filter GatewayInvocationFilter) ([]GatewayToolInvocation, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	query := r.db.WithContext(ctx).Order("id DESC").Limit(limit)
	if source := strings.TrimSpace(filter.Source); source != "" {
		query = query.Where("source = ?", source)
	}
	if toolName := strings.TrimSpace(filter.ToolName); toolName != "" {
		query = query.Where("tool_name = ?", toolName)
	}
	if filter.SessionID != nil {
		query = query.Where("session_id = ?", *filter.SessionID)
	}
	if skillID := strings.TrimSpace(filter.SkillID); skillID != "" {
		query = query.Where("skill_id = ?", skillID)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("status = ?", status)
	}

	var invocations []GatewayToolInvocation
	if err := query.Find(&invocations).Error; err != nil {
		return nil, fmt.Errorf("list gateway tool invocations: %w", err)
	}
	reverseGatewayInvocations(invocations)
	return invocations, nil
}

func reverseGatewayInvocations(invocations []GatewayToolInvocation) {
	for left, right := 0, len(invocations)-1; left < right; left, right = left+1, right-1 {
		invocations[left], invocations[right] = invocations[right], invocations[left]
	}
}
