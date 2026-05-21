package skills

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/store"
)

func TestBuiltinSkillsContainExpectedDiagnosisMethods(t *testing.T) {
	builtins := BuiltinSkills()

	wantIDs := []string{
		"hot_rank_attribution",
		"comment_risk_analysis",
		"author_support_evaluation",
		"tag_trend_analysis",
		"content_review_summary",
	}
	if len(builtins) != len(wantIDs) {
		t.Fatalf("builtin count = %d, want %d: %+v", len(builtins), len(wantIDs), builtins)
	}
	byID := make(map[string]DiagnosisSkill, len(builtins))
	for _, skill := range builtins {
		byID[skill.ID] = skill
	}
	for _, id := range wantIDs {
		skill, ok := byID[id]
		if !ok {
			t.Fatalf("missing builtin skill %q in %+v", id, builtins)
		}
		if skill.Status != SkillStatusEnabled || skill.Version == "" || len(skill.AllowedTools) == 0 ||
			len(skill.RequiredEvidence) == 0 || len(skill.OutputSections) == 0 || strings.TrimSpace(skill.PromptTemplate) == "" {
			t.Fatalf("builtin skill %q is incomplete: %+v", id, skill)
		}
	}
}

func TestServiceValidationRejectsUnsafeOrUnusableSkills(t *testing.T) {
	registry := newSkillTestRegistry(t)
	service := NewService(Dependencies{Registry: registry})

	if err := service.ValidateDefinition(DiagnosisSkill{
		ID:               "unknown_tool",
		Name:             "unknown",
		Version:          "1.0.0",
		Status:           SkillStatusEnabled,
		AllowedTools:     []string{"missing_tool"},
		RequiredEvidence: []string{"missing_tool"},
		PromptTemplate:   "Use evidence.",
		OutputSections:   []string{"结论"},
	}); err == nil || !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("ValidateDefinition error = %v, want unknown tool", err)
	}

	if err := service.ValidateDefinition(DiagnosisSkill{
		ID:             "empty_evidence",
		Name:           "empty",
		Version:        "1.0.0",
		Status:         SkillStatusEnabled,
		AllowedTools:   []string{"get_video_detail"},
		PromptTemplate: "Use evidence.",
		OutputSections: []string{"结论"},
	}); err == nil || !strings.Contains(err.Error(), "required_evidence") {
		t.Fatalf("ValidateDefinition error = %v, want required evidence", err)
	}

	if err := service.EnsureRuntimeUsable(DiagnosisSkill{
		ID:               "disabled",
		Name:             "disabled",
		Version:          "1.0.0",
		Status:           SkillStatusDisabled,
		AllowedTools:     []string{"get_video_detail"},
		RequiredEvidence: []string{"get_video_detail"},
		PromptTemplate:   "Use evidence.",
		OutputSections:   []string{"结论"},
	}); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("EnsureRuntimeUsable error = %v, want disabled", err)
	}
}

func TestServiceListsBuiltinsAndCustomSkills(t *testing.T) {
	ctx := context.Background()
	db, err := store.OpenSQLite(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer func() {
		if err := store.Close(db); err != nil {
			t.Fatalf("close db: %v", err)
		}
	}()
	if err := store.AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	service := NewService(Dependencies{
		Registry:   newSkillTestRegistry(t),
		Repository: store.NewSkillRepository(db),
	})
	custom := DiagnosisSkill{
		ID:               "custom_comment_review",
		Name:             "自定义评论复盘",
		Version:          "1.0.0",
		Status:           SkillStatusEnabled,
		Scenario:         "comment_risk_analysis",
		AllowedTools:     []string{"get_video_detail", "analyze_video_comment_risk"},
		RequiredEvidence: []string{"get_video_detail"},
		PromptTemplate:   "Use evidence.",
		OutputSections:   []string{"结论", "证据"},
	}
	if err := service.Create(ctx, custom); err != nil {
		t.Fatalf("create custom skill: %v", err)
	}

	listed, err := service.List(ctx)
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}
	if len(listed) != 6 {
		t.Fatalf("skill count = %d, want 6: %+v", len(listed), listed)
	}
	got, err := service.Get(ctx, custom.ID)
	if err != nil {
		t.Fatalf("get custom skill: %v", err)
	}
	if got.Name != "自定义评论复盘" || got.RequiredEvidence[0] != "get_video_detail" {
		t.Fatalf("custom skill = %+v", got)
	}
	builtin, err := service.Get(ctx, "comment_risk_analysis")
	if err != nil {
		t.Fatalf("get builtin skill: %v", err)
	}
	if builtin.ID != "comment_risk_analysis" || builtin.Status != SkillStatusEnabled {
		t.Fatalf("builtin skill = %+v", builtin)
	}

	if err := service.Disable(ctx, custom.ID); err != nil {
		t.Fatalf("disable custom skill: %v", err)
	}
	if _, err := service.GetForRuntime(ctx, custom.ID); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("GetForRuntime error = %v, want disabled", err)
	}
	if err := service.Enable(ctx, custom.ID); err != nil {
		t.Fatalf("enable custom skill: %v", err)
	}
	enabled, err := service.GetForRuntime(ctx, custom.ID)
	if err != nil {
		t.Fatalf("GetForRuntime returned error: %v", err)
	}
	if enabled.Status != SkillStatusEnabled {
		t.Fatalf("enabled skill = %+v", enabled)
	}
}

func newSkillTestRegistry(t *testing.T) *tools.Registry {
	t.Helper()
	registry, err := tools.NewRegistry(
		skillTestTool{name: "get_video_detail"},
		skillTestTool{name: "get_hot_videos"},
		skillTestTool{name: "get_video_comments"},
		skillTestTool{name: "analyze_video_comment_risk"},
		skillTestTool{name: "analyze_comment_risk"},
		skillTestTool{name: "get_author_profile"},
		skillTestTool{name: "list_author_videos"},
		skillTestTool{name: "list_tag_videos"},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	return registry
}

type skillTestTool struct {
	name string
}

func (t skillTestTool) Name() string {
	return t.name
}

func (t skillTestTool) Schema() tools.ToolSchema {
	return tools.NewFunctionSchema(t.name, "test tool", map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
}

func (t skillTestTool) Timeout() time.Duration {
	return 0
}

func (t skillTestTool) Execute(_ context.Context, _ json.RawMessage) (tools.ToolResult, error) {
	return tools.ToolResult{ToolName: t.name, Summary: "ok"}, nil
}
