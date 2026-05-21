package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRegistryLooksUpToolsAndReturnsStableSchemas(t *testing.T) {
	registry, err := NewRegistry(
		staticTool{name: "z_tool", description: "last"},
		staticTool{name: "a_tool", description: "first"},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	tool, ok := registry.Get("a_tool")
	if !ok {
		t.Fatalf("expected a_tool to be registered")
	}
	if tool.Name() != "a_tool" {
		t.Fatalf("tool name = %q, want a_tool", tool.Name())
	}

	schemas := registry.Schemas()
	if len(schemas) != 2 {
		t.Fatalf("schema count = %d, want 2", len(schemas))
	}
	if schemas[0].Function.Name != "a_tool" || schemas[1].Function.Name != "z_tool" {
		t.Fatalf("schemas are not sorted by name: %+v", schemas)
	}
}

func TestRegistryRejectsDuplicateToolNames(t *testing.T) {
	_, err := NewRegistry(
		staticTool{name: "get_video_detail"},
		staticTool{name: "get_video_detail"},
	)
	if err == nil {
		t.Fatalf("expected duplicate tool name error")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error = %q, want duplicate context", err.Error())
	}
}

type staticTool struct {
	name        string
	description string
}

func (t staticTool) Name() string {
	return t.name
}

func (t staticTool) Schema() ToolSchema {
	return NewFunctionSchema(t.name, t.description, map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
}

func (t staticTool) Timeout() time.Duration {
	return 0
}

func (t staticTool) Execute(_ context.Context, _ json.RawMessage) (ToolResult, error) {
	return ToolResult{ToolName: t.name, Data: map[string]string{"ok": "true"}, Summary: "ok"}, nil
}
