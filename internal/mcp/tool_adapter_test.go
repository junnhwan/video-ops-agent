package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/gateway"
	"video-ops-agent/internal/store"

	"gorm.io/gorm"
)

func TestToolSchemaMapsToMCPToolSchema(t *testing.T) {
	schema := tools.NewFunctionSchema("get_video_detail", "Fetch video.", map[string]any{
		"type":       "object",
		"properties": map[string]any{"video_id": map[string]any{"type": "integer"}},
		"required":   []string{"video_id"},
	})

	tool := ToolFromSchema(schema)

	if tool.Name != "get_video_detail" || tool.Description != "Fetch video." {
		t.Fatalf("tool = %+v", tool)
	}
	if tool.InputSchema["type"] != "object" {
		t.Fatalf("input schema = %+v", tool.InputSchema)
	}
}

func TestToolAdapterCallExecutesThroughGatewayAndRecordsMCPSource(t *testing.T) {
	ctx := context.Background()
	db := newMCPTestDB(t)
	registry, err := tools.NewRegistry(mcpTestTool{
		name: "get_video_detail",
		result: tools.ToolResult{
			ToolName: "get_video_detail",
			Summary:  "video 101: test",
			Data:     map[string]any{"id": float64(101), "title": "test"},
		},
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	invocations := store.NewGatewayInvocationRepository(db)
	gatewayService := gateway.NewService(gateway.Dependencies{
		Registry:    registry,
		Executor:    tools.NewExecutor(registry, time.Second),
		Invocations: invocations,
	})
	adapter := NewToolAdapter(gatewayService)

	result, err := adapter.CallTool(ctx, "get_video_detail", json.RawMessage(`{"video_id":101}`))
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if len(result.Content) != 1 || result.Content[0].Text != "video 101: test" ||
		result.StructuredContent.ToolName != "get_video_detail" {
		t.Fatalf("result = %+v", result)
	}

	recorded, err := invocations.List(ctx, store.GatewayInvocationFilter{Source: gateway.InvocationSourceMCPClient})
	if err != nil {
		t.Fatalf("list invocations: %v", err)
	}
	if len(recorded) != 1 || recorded[0].ToolName != "get_video_detail" {
		t.Fatalf("recorded invocations = %+v", recorded)
	}
}

func newMCPTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := store.OpenSQLite(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(db); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})
	if err := store.AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}
	return db
}

type mcpTestTool struct {
	name   string
	result tools.ToolResult
}

func (t mcpTestTool) Name() string {
	return t.name
}

func (t mcpTestTool) Schema() tools.ToolSchema {
	return tools.NewFunctionSchema(t.name, "test tool", map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
}

func (t mcpTestTool) Timeout() time.Duration {
	return 0
}

func (t mcpTestTool) Execute(_ context.Context, _ json.RawMessage) (tools.ToolResult, error) {
	if t.result.ToolName != "" {
		return t.result, nil
	}
	return tools.ToolResult{ToolName: t.name, Summary: "ok"}, nil
}
