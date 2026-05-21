package gateway

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/store"

	"gorm.io/gorm"
)

func TestServiceCatalogReturnsSortedToolsWithCategories(t *testing.T) {
	registry, err := tools.NewRegistry(
		gatewayTestTool{name: "get_video_detail", description: "Fetch video."},
		gatewayTestTool{name: "analyze_video_comment_risk", description: "Analyze comments."},
		gatewayTestTool{name: "get_author_profile", description: "Fetch author."},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	service := NewService(Dependencies{Registry: registry})

	catalog, err := service.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}

	if len(catalog) != 3 {
		t.Fatalf("catalog count = %d, want 3", len(catalog))
	}
	names := []string{catalog[0].Name, catalog[1].Name, catalog[2].Name}
	wantNames := []string{"analyze_video_comment_risk", "get_author_profile", "get_video_detail"}
	for i := range wantNames {
		if names[i] != wantNames[i] {
			t.Fatalf("catalog names = %+v, want %+v", names, wantNames)
		}
	}
	if catalog[0].Category != "comment" || catalog[1].Category != "author" || catalog[2].Category != "video" {
		t.Fatalf("catalog categories = %q, %q, %q", catalog[0].Category, catalog[1].Category, catalog[2].Category)
	}
	if !catalog[0].ReadOnly || catalog[0].DisplayName == "" || catalog[0].Schema.Function.Name != "analyze_video_comment_risk" {
		t.Fatalf("catalog item = %+v", catalog[0])
	}
}

func TestServiceCallToolRecordsSuccessInvocation(t *testing.T) {
	ctx := context.Background()
	db := newGatewayTestDB(t)
	session, err := store.NewSessionRepository(db).Create(ctx, store.CreateSessionInput{UserID: "operator-1"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	registry, err := tools.NewRegistry(gatewayTestTool{
		name:        "get_video_detail",
		description: "Fetch video.",
		result: tools.ToolResult{
			ToolName: "get_video_detail",
			Summary:  "video 101: test",
			Data:     map[string]any{"id": float64(101), "title": "test"},
		},
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	service := NewService(Dependencies{
		Registry:    registry,
		Executor:    tools.NewExecutor(registry, time.Second),
		Invocations: store.NewGatewayInvocationRepository(db),
	})

	output, err := service.CallTool(ctx, CallToolInput{
		ToolName:     "get_video_detail",
		Arguments:    json.RawMessage(`{"video_id":101}`),
		Source:       InvocationSourceManualConsole,
		SessionID:    &session.ID,
		SkillID:      "comment_risk_analysis",
		SkillVersion: "1.0.0",
	})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}

	if output.Result.Summary != "video 101: test" || output.Invocation.ID == 0 {
		t.Fatalf("output = %+v", output)
	}
	if output.Invocation.Source != InvocationSourceManualConsole ||
		output.Invocation.ToolName != "get_video_detail" ||
		output.Invocation.Status != store.ToolCallStatusSuccess ||
		output.Invocation.SkillID != "comment_risk_analysis" {
		t.Fatalf("invocation = %+v", output.Invocation)
	}
	if output.Invocation.ResultJSON == "" || output.Invocation.ResultSummary != "video 101: test" {
		t.Fatalf("invocation result fields = %+v", output.Invocation)
	}

	invocations, err := service.ListInvocations(ctx, store.GatewayInvocationFilter{
		Source:    InvocationSourceManualConsole,
		SessionID: &session.ID,
	})
	if err != nil {
		t.Fatalf("ListInvocations returned error: %v", err)
	}
	if len(invocations) != 1 || invocations[0].ToolName != "get_video_detail" {
		t.Fatalf("invocations = %+v", invocations)
	}
}

func newGatewayTestDB(t *testing.T) *gorm.DB {
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

type gatewayTestTool struct {
	name        string
	description string
	result      tools.ToolResult
}

func (t gatewayTestTool) Name() string {
	return t.name
}

func (t gatewayTestTool) Schema() tools.ToolSchema {
	return tools.NewFunctionSchema(t.name, t.description, map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
}

func (t gatewayTestTool) Timeout() time.Duration {
	return 0
}

func (t gatewayTestTool) Execute(_ context.Context, _ json.RawMessage) (tools.ToolResult, error) {
	if t.result.ToolName != "" || t.result.Summary != "" || t.result.Data != nil {
		return t.result, nil
	}
	return tools.ToolResult{ToolName: t.name, Summary: "ok", Data: map[string]any{"ok": true}}, nil
}
