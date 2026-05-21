package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestExecutorRunsRegisteredTool(t *testing.T) {
	registry, err := NewRegistry(staticTool{name: "ok_tool"})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	executor := NewExecutor(registry, 2*time.Second)
	result, err := executor.Execute(context.Background(), "ok_tool", json.RawMessage(`{"id":1}`))
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.ToolName != "ok_tool" || result.Summary != "ok" {
		t.Fatalf("result = %+v", result)
	}
}

func TestExecutorReturnsErrorForUnknownTool(t *testing.T) {
	registry, err := NewRegistry(staticTool{name: "known_tool"})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	executor := NewExecutor(registry, time.Second)
	_, err = executor.Execute(context.Background(), "missing_tool", json.RawMessage(`{}`))
	if err == nil {
		t.Fatalf("expected unknown tool error")
	}
	if !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("error = %q, want unknown tool context", err.Error())
	}
}

func TestExecutorEnforcesPerToolTimeout(t *testing.T) {
	registry, err := NewRegistry(slowTool{name: "slow_tool", timeout: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}

	executor := NewExecutor(registry, time.Second)
	started := time.Now()
	_, err = executor.Execute(context.Background(), "slow_tool", json.RawMessage(`{}`))
	elapsed := time.Since(started)

	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("error = %q, want timeout context", err.Error())
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("timeout took %s, want executor to return promptly", elapsed)
	}
}

type slowTool struct {
	name    string
	timeout time.Duration
}

func (t slowTool) Name() string {
	return t.name
}

func (t slowTool) Schema() ToolSchema {
	return NewFunctionSchema(t.name, "slow", map[string]any{"type": "object"})
}

func (t slowTool) Timeout() time.Duration {
	return t.timeout
}

func (t slowTool) Execute(_ context.Context, _ json.RawMessage) (ToolResult, error) {
	time.Sleep(time.Second)
	return ToolResult{ToolName: t.name, Summary: "late"}, nil
}
