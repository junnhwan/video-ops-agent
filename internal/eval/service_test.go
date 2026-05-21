package eval

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"video-ops-agent/internal/agent/skills"
	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/gateway"
	"video-ops-agent/internal/store"

	"gorm.io/gorm"
)

func TestServiceSummaryAggregatesTraceMetrics(t *testing.T) {
	ctx := context.Background()
	db := newEvalTestDB(t)
	sessionRepo := store.NewSessionRepository(db)
	invocationRepo := store.NewGatewayInvocationRepository(db)
	skillService := skills.NewService(skills.Dependencies{Registry: newEvalSkillRegistry(t), Repository: store.NewSkillRepository(db)})

	complete, err := sessionRepo.Create(ctx, store.CreateSessionInput{
		UserID:       "operator-1",
		SkillID:      "comment_risk_analysis",
		SkillVersion: "1.0.0",
		Status:       store.SessionStatusActive,
	})
	if err != nil {
		t.Fatalf("create complete session: %v", err)
	}
	incomplete, err := sessionRepo.Create(ctx, store.CreateSessionInput{
		UserID:       "operator-1",
		SkillID:      "comment_risk_analysis",
		SkillVersion: "1.0.0",
		Status:       store.SessionStatusActive,
	})
	if err != nil {
		t.Fatalf("create incomplete session: %v", err)
	}
	seedInvocation(t, ctx, invocationRepo, complete.ID, "comment_risk_analysis", "get_video_detail", store.ToolCallStatusSuccess, 10)
	seedInvocation(t, ctx, invocationRepo, complete.ID, "comment_risk_analysis", "analyze_video_comment_risk", store.ToolCallStatusSuccess, 20)
	seedInvocation(t, ctx, invocationRepo, incomplete.ID, "comment_risk_analysis", "get_hot_videos", store.ToolCallStatusSuccess, 40)
	seedInvocation(t, ctx, invocationRepo, incomplete.ID, "comment_risk_analysis", "analyze_video_comment_risk", store.ToolCallStatusError, 30)

	service := NewService(Dependencies{
		Sessions:    sessionRepo,
		Invocations: invocationRepo,
		Skills:      skillService,
	})
	summary, err := service.Summary(ctx, SummaryFilter{})
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}

	if summary.ToolCallSuccessRate != 0.75 ||
		summary.ToolCallErrorCount != 1 ||
		summary.UnauthorizedToolCallCount != 1 ||
		summary.EvidenceCompleteFinalAnswerCount != 1 ||
		summary.SkillSuccessCount != 1 ||
		summary.SkillFailureCount != 1 {
		t.Fatalf("summary counts = %+v", summary)
	}
	if summary.AverageToolLatencyMS != 25 || summary.AverageToolCallCount != 2 {
		t.Fatalf("summary averages = %+v", summary)
	}
	if len(summary.UnsupportedMetrics) == 0 {
		t.Fatalf("expected unsupported metrics note for non-persisted counters: %+v", summary)
	}

	skillSummary, err := service.Summary(ctx, SummaryFilter{SkillID: "comment_risk_analysis"})
	if err != nil {
		t.Fatalf("skill Summary returned error: %v", err)
	}
	if skillSummary.ToolCallErrorCount != 1 || skillSummary.SkillFailureCount != 1 {
		t.Fatalf("skill summary = %+v", skillSummary)
	}
}

func TestServiceCreatesAndGetsEvalRun(t *testing.T) {
	ctx := context.Background()
	db := newEvalTestDB(t)
	service := NewService(Dependencies{
		Sessions:    store.NewSessionRepository(db),
		Invocations: store.NewGatewayInvocationRepository(db),
		Skills:      skills.NewService(skills.Dependencies{Registry: newEvalSkillRegistry(t)}),
	})

	run, err := service.CreateRun(ctx, CreateRunInput{Mode: ModeSkillGuard})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	if run.ID == 0 || run.Mode != ModeSkillGuard {
		t.Fatalf("run = %+v", run)
	}
	got, err := service.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRun returned error: %v", err)
	}
	if got.ID != run.ID || got.Mode != ModeSkillGuard {
		t.Fatalf("got run = %+v", got)
	}
}

func newEvalTestDB(t *testing.T) *gorm.DB {
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

func newEvalSkillRegistry(t *testing.T) *tools.Registry {
	t.Helper()
	registry, err := tools.NewRegistry(
		evalTestTool{name: "get_video_detail"},
		evalTestTool{name: "get_hot_videos"},
		evalTestTool{name: "get_video_comments"},
		evalTestTool{name: "analyze_video_comment_risk"},
		evalTestTool{name: "analyze_comment_risk"},
		evalTestTool{name: "get_author_profile"},
		evalTestTool{name: "list_author_videos"},
		evalTestTool{name: "list_tag_videos"},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	return registry
}

func seedInvocation(t *testing.T, ctx context.Context, repo *store.GatewayInvocationRepository, sessionID uint, skillID string, toolName string, status string, latencyMS int64) {
	t.Helper()
	if _, err := repo.Create(ctx, store.CreateGatewayInvocationInput{
		Source:        gateway.InvocationSourceAgentRuntime,
		SessionID:     &sessionID,
		SkillID:       skillID,
		SkillVersion:  "1.0.0",
		ToolName:      toolName,
		ArgumentsJSON: `{}`,
		ResultSummary: toolName + " summary",
		LatencyMS:     latencyMS,
		Status:        status,
	}); err != nil {
		t.Fatalf("create invocation %s: %v", toolName, err)
	}
}

type evalTestTool struct {
	name string
}

func (t evalTestTool) Name() string { return t.name }
func (t evalTestTool) Schema() tools.ToolSchema {
	return tools.NewFunctionSchema(t.name, "test tool", map[string]any{"type": "object"})
}
func (t evalTestTool) Timeout() time.Duration { return 0 }
func (t evalTestTool) Execute(context.Context, json.RawMessage) (tools.ToolResult, error) {
	return tools.ToolResult{ToolName: t.name, Summary: "ok"}, nil
}
