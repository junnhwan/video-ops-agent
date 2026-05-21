package tools

import (
	"context"
	"encoding/json"
	"time"
)

type Tool interface {
	Name() string
	Schema() ToolSchema
	Timeout() time.Duration
	Execute(ctx context.Context, arguments json.RawMessage) (ToolResult, error)
}

type ToolSchema struct {
	Type     string             `json:"type"`
	Function ToolFunctionSchema `json:"function"`
}

type ToolFunctionSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ToolResult struct {
	ToolName string `json:"tool_name"`
	Data     any    `json:"data"`
	Summary  string `json:"summary"`
}

func NewFunctionSchema(name string, description string, parameters map[string]any) ToolSchema {
	return ToolSchema{
		Type: "function",
		Function: ToolFunctionSchema{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
}
