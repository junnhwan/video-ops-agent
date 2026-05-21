package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestGatewayInvocationRepositoryCreateListAndGet(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("close db: %v", err)
		}
	}()
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	session, err := NewSessionRepository(db).Create(ctx, CreateSessionInput{
		UserID: "operator-1",
		Status: SessionStatusActive,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	repo := NewGatewayInvocationRepository(db)
	manual, err := repo.Create(ctx, CreateGatewayInvocationInput{
		Source:        InvocationSourceManualConsole,
		SessionID:     &session.ID,
		SkillID:       "comment_risk_analysis",
		SkillVersion:  "1.0.0",
		ToolName:      "analyze_video_comment_risk",
		ArgumentsJSON: `{"video_id":101,"limit":50}`,
		ResultJSON:    `{"risk_level":"low"}`,
		ResultSummary: "low comment risk for video 101",
		LatencyMS:     15,
		Status:        ToolCallStatusSuccess,
	})
	if err != nil {
		t.Fatalf("create manual invocation: %v", err)
	}
	if manual.ID == 0 {
		t.Fatalf("expected invocation id to be assigned")
	}

	if _, err := repo.Create(ctx, CreateGatewayInvocationInput{
		Source:        InvocationSourceAgentRuntime,
		SessionID:     &session.ID,
		ToolName:      "get_video_detail",
		ArgumentsJSON: `{"video_id":101}`,
		Status:        ToolCallStatusSuccess,
	}); err != nil {
		t.Fatalf("create agent invocation: %v", err)
	}

	filtered, err := repo.List(ctx, GatewayInvocationFilter{
		Source:    InvocationSourceManualConsole,
		ToolName:  "analyze_video_comment_risk",
		SessionID: &session.ID,
		SkillID:   "comment_risk_analysis",
		Status:    ToolCallStatusSuccess,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("list invocations: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("filtered invocation count = %d, want 1: %+v", len(filtered), filtered)
	}
	if filtered[0].ResultSummary != "low comment risk for video 101" || filtered[0].LatencyMS != 15 {
		t.Fatalf("filtered invocation = %+v", filtered[0])
	}

	got, err := repo.Get(ctx, manual.ID)
	if err != nil {
		t.Fatalf("get invocation: %v", err)
	}
	if got.ToolName != "analyze_video_comment_risk" || got.Source != InvocationSourceManualConsole {
		t.Fatalf("got invocation = %+v", got)
	}
}

func TestGatewayInvocationRepositoryValidatesRequiredFields(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("close db: %v", err)
		}
	}()
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	repo := NewGatewayInvocationRepository(db)
	if _, err := repo.Create(ctx, CreateGatewayInvocationInput{ToolName: "get_video_detail", ArgumentsJSON: `{}`}); err == nil {
		t.Fatalf("expected missing source validation error")
	}
	if _, err := repo.Create(ctx, CreateGatewayInvocationInput{Source: InvocationSourceManualConsole, ArgumentsJSON: `{}`}); err == nil {
		t.Fatalf("expected missing tool name validation error")
	}
	if _, err := repo.Create(ctx, CreateGatewayInvocationInput{Source: InvocationSourceManualConsole, ToolName: "get_video_detail"}); err == nil {
		t.Fatalf("expected missing arguments validation error")
	}
}
