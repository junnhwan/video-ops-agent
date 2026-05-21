package llm

import (
	"encoding/json"

	"video-ops-agent/internal/agent/tools"
)

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ChatRequest struct {
	Messages []Message          `json:"messages"`
	Tools    []tools.ToolSchema `json:"tools,omitempty"`
}

type ChatResponse struct {
	ID           string  `json:"id"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}
