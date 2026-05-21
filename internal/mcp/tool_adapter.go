package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/gateway"
)

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolCallResult struct {
	Content           []ContentItem    `json:"content"`
	StructuredContent tools.ToolResult `json:"structuredContent"`
}

type ToolAdapter struct {
	gateway *gateway.Service
}

func NewToolAdapter(gatewayService *gateway.Service) *ToolAdapter {
	return &ToolAdapter{gateway: gatewayService}
}

func ToolFromSchema(schema tools.ToolSchema) Tool {
	return Tool{
		Name:        schema.Function.Name,
		Description: schema.Function.Description,
		InputSchema: schema.Function.Parameters,
	}
}

func (a *ToolAdapter) ListTools(ctx context.Context) ([]Tool, error) {
	if a.gateway == nil {
		return nil, fmt.Errorf("gateway service is required")
	}
	items, err := a.gateway.ListTools(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Tool, 0, len(items))
	for _, item := range items {
		out = append(out, ToolFromSchema(item.Schema))
	}
	return out, nil
}

func (a *ToolAdapter) CallTool(ctx context.Context, name string, arguments json.RawMessage) (ToolCallResult, error) {
	if a.gateway == nil {
		return ToolCallResult{}, fmt.Errorf("gateway service is required")
	}
	output, err := a.gateway.CallTool(ctx, gateway.CallToolInput{
		ToolName:  name,
		Arguments: arguments,
		Source:    gateway.InvocationSourceMCPClient,
	})
	if err != nil {
		return ToolCallResult{}, err
	}
	return ToolCallResult{
		Content: []ContentItem{{
			Type: "text",
			Text: output.Result.Summary,
		}},
		StructuredContent: output.Result,
	}, nil
}
