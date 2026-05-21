package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/store"
)

type Dependencies struct {
	Registry    *tools.Registry
	Executor    *tools.Executor
	Invocations *store.GatewayInvocationRepository
}

type Service struct {
	registry    *tools.Registry
	executor    *tools.Executor
	invocations *store.GatewayInvocationRepository
}

type CallToolInput struct {
	ToolName     string
	Arguments    json.RawMessage
	Source       string
	SessionID    *uint
	MessageID    *uint
	SkillID      string
	SkillVersion string
}

type CallToolOutput struct {
	Invocation store.GatewayToolInvocation `json:"invocation"`
	Result     tools.ToolResult            `json:"result"`
}

func NewService(deps Dependencies) *Service {
	return &Service{
		registry:    deps.Registry,
		executor:    deps.Executor,
		invocations: deps.Invocations,
	}
}

func (s *Service) ListTools(_ context.Context) ([]ToolCatalogItem, error) {
	if s.registry == nil {
		return nil, fmt.Errorf("tool registry is required")
	}
	schemas := s.registry.Schemas()
	items := make([]ToolCatalogItem, 0, len(schemas))
	for _, schema := range schemas {
		items = append(items, catalogItemFromSchema(schema))
	}
	return items, nil
}

func (s *Service) GetTool(ctx context.Context, name string) (ToolCatalogItem, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ToolCatalogItem{}, fmt.Errorf("tool name is required")
	}
	items, err := s.ListTools(ctx)
	if err != nil {
		return ToolCatalogItem{}, err
	}
	for _, item := range items {
		if item.Name == name {
			return item, nil
		}
	}
	return ToolCatalogItem{}, fmt.Errorf("unknown tool %q", name)
}

func (s *Service) CallTool(ctx context.Context, input CallToolInput) (CallToolOutput, error) {
	if s.executor == nil {
		return CallToolOutput{}, fmt.Errorf("tool executor is required")
	}
	if s.invocations == nil {
		return CallToolOutput{}, fmt.Errorf("gateway invocation repository is required")
	}
	toolName := strings.TrimSpace(input.ToolName)
	if toolName == "" {
		return CallToolOutput{}, fmt.Errorf("tool name is required")
	}
	source := strings.TrimSpace(input.Source)
	if source == "" {
		source = InvocationSourceManualConsole
	}
	arguments := input.Arguments
	if len(strings.TrimSpace(string(arguments))) == 0 {
		arguments = json.RawMessage(`{}`)
	}

	started := time.Now()
	result, err := s.executor.Execute(ctx, toolName, arguments)
	latencyMS := time.Since(started).Milliseconds()

	record := store.CreateGatewayInvocationInput{
		Source:        source,
		SessionID:     input.SessionID,
		MessageID:     input.MessageID,
		SkillID:       input.SkillID,
		SkillVersion:  input.SkillVersion,
		ToolName:      toolName,
		ArgumentsJSON: string(arguments),
		LatencyMS:     latencyMS,
	}
	if err != nil {
		record.Status = statusForToolError(err)
		record.ErrorMessage = err.Error()
	} else {
		record.Status = store.ToolCallStatusSuccess
		record.ResultSummary = result.Summary
		if result.Data != nil {
			encoded, marshalErr := json.Marshal(result.Data)
			if marshalErr != nil {
				record.Status = store.ToolCallStatusError
				record.ErrorMessage = fmt.Sprintf("marshal tool result: %v", marshalErr)
			} else {
				record.ResultJSON = string(encoded)
			}
		}
	}

	invocation, createErr := s.invocations.Create(ctx, record)
	if createErr != nil {
		return CallToolOutput{}, createErr
	}
	if err != nil {
		return CallToolOutput{Invocation: invocation}, err
	}
	return CallToolOutput{Invocation: invocation, Result: result}, nil
}

func (s *Service) ListInvocations(ctx context.Context, filter store.GatewayInvocationFilter) ([]store.GatewayToolInvocation, error) {
	if s.invocations == nil {
		return nil, fmt.Errorf("gateway invocation repository is required")
	}
	return s.invocations.List(ctx, filter)
}

func (s *Service) GetInvocation(ctx context.Context, id uint) (store.GatewayToolInvocation, error) {
	if s.invocations == nil {
		return store.GatewayToolInvocation{}, fmt.Errorf("gateway invocation repository is required")
	}
	return s.invocations.Get(ctx, id)
}

func statusForToolError(err error) string {
	lowered := strings.ToLower(err.Error())
	if strings.Contains(lowered, "timeout") || strings.Contains(lowered, "deadline") {
		return store.ToolCallStatusTimeout
	}
	return store.ToolCallStatusError
}
