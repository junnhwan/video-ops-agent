package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSkillRepositoryUpsertGetListAndSetStatus(t *testing.T) {
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

	repo := NewSkillRepository(db)
	record := DiagnosisSkillRecord{
		ID:                   "custom_comment_review",
		Name:                 "自定义评论复盘",
		Description:          "custom skill",
		Version:              "1.0.0",
		Status:               "enabled",
		Scenario:             "comment_risk_analysis",
		AllowedToolsJSON:     `["get_video_detail","analyze_video_comment_risk"]`,
		RequiredEvidenceJSON: `["get_video_detail"]`,
		PromptTemplate:       "Use evidence.",
		OutputSectionsJSON:   `["结论","证据"]`,
	}
	if err := repo.Upsert(ctx, record); err != nil {
		t.Fatalf("upsert skill: %v", err)
	}

	got, err := repo.Get(ctx, record.ID)
	if err != nil {
		t.Fatalf("get skill: %v", err)
	}
	if got.ID != record.ID || got.Name != "自定义评论复盘" || got.Status != "enabled" {
		t.Fatalf("got skill = %+v", got)
	}

	record.Name = "自定义评论复盘 v2"
	record.Version = "1.0.1"
	if err := repo.Upsert(ctx, record); err != nil {
		t.Fatalf("update skill: %v", err)
	}
	updated, err := repo.Get(ctx, record.ID)
	if err != nil {
		t.Fatalf("get updated skill: %v", err)
	}
	if updated.Name != "自定义评论复盘 v2" || updated.Version != "1.0.1" {
		t.Fatalf("updated skill = %+v", updated)
	}

	if err := repo.SetStatus(ctx, record.ID, "disabled"); err != nil {
		t.Fatalf("disable skill: %v", err)
	}
	disabled, err := repo.Get(ctx, record.ID)
	if err != nil {
		t.Fatalf("get disabled skill: %v", err)
	}
	if disabled.Status != "disabled" {
		t.Fatalf("disabled status = %q", disabled.Status)
	}

	records, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}
	if len(records) != 1 || records[0].ID != record.ID {
		t.Fatalf("records = %+v", records)
	}
}

func TestSkillRepositoryValidatesRequiredFields(t *testing.T) {
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

	repo := NewSkillRepository(db)
	if err := repo.Upsert(ctx, DiagnosisSkillRecord{}); err == nil {
		t.Fatalf("expected skill validation error")
	}
	if err := repo.SetStatus(ctx, "missing", ""); err == nil {
		t.Fatalf("expected status validation error")
	}
}
