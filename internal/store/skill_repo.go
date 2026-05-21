package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SkillRepository struct {
	db *gorm.DB
}

func NewSkillRepository(db *gorm.DB) *SkillRepository {
	return &SkillRepository{db: db}
}

func (r *SkillRepository) Upsert(ctx context.Context, record DiagnosisSkillRecord) error {
	if err := validateSkillRecord(record); err != nil {
		return err
	}
	record.ID = strings.TrimSpace(record.ID)
	record.Name = strings.TrimSpace(record.Name)
	record.Version = strings.TrimSpace(record.Version)
	record.Status = strings.TrimSpace(record.Status)
	record.Scenario = strings.TrimSpace(record.Scenario)
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"description",
			"version",
			"status",
			"scenario",
			"allowed_tools_json",
			"required_evidence_json",
			"prompt_template",
			"output_sections_json",
			"risk_notes_json",
			"updated_at",
		}),
	}).Create(&record).Error; err != nil {
		return fmt.Errorf("upsert diagnosis skill %q: %w", record.ID, err)
	}
	return nil
}

func (r *SkillRepository) Get(ctx context.Context, id string) (DiagnosisSkillRecord, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return DiagnosisSkillRecord{}, fmt.Errorf("skill id is required")
	}
	var record DiagnosisSkillRecord
	if err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		return DiagnosisSkillRecord{}, fmt.Errorf("get diagnosis skill %q: %w", id, err)
	}
	return record, nil
}

func (r *SkillRepository) List(ctx context.Context) ([]DiagnosisSkillRecord, error) {
	var records []DiagnosisSkillRecord
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list diagnosis skills: %w", err)
	}
	return records, nil
}

func (r *SkillRepository) SetStatus(ctx context.Context, id string, status string) error {
	id = strings.TrimSpace(id)
	status = strings.TrimSpace(status)
	if id == "" {
		return fmt.Errorf("skill id is required")
	}
	if status != SkillStatusEnabled && status != SkillStatusDisabled {
		return fmt.Errorf("skill status must be %q or %q", SkillStatusEnabled, SkillStatusDisabled)
	}
	result := r.db.WithContext(ctx).Model(&DiagnosisSkillRecord{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("set diagnosis skill %q status: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("diagnosis skill %q not found", id)
	}
	return nil
}

func validateSkillRecord(record DiagnosisSkillRecord) error {
	if strings.TrimSpace(record.ID) == "" {
		return fmt.Errorf("skill id is required")
	}
	if strings.TrimSpace(record.Name) == "" {
		return fmt.Errorf("skill name is required")
	}
	if strings.TrimSpace(record.Version) == "" {
		return fmt.Errorf("skill version is required")
	}
	status := strings.TrimSpace(record.Status)
	if status != SkillStatusEnabled && status != SkillStatusDisabled {
		return fmt.Errorf("skill status must be %q or %q", SkillStatusEnabled, SkillStatusDisabled)
	}
	if !validJSONArray(record.AllowedToolsJSON) {
		return fmt.Errorf("allowed_tools_json must be a json array")
	}
	if !validJSONArray(record.RequiredEvidenceJSON) {
		return fmt.Errorf("required_evidence_json must be a json array")
	}
	if strings.TrimSpace(record.PromptTemplate) == "" {
		return fmt.Errorf("prompt_template is required")
	}
	if !validJSONArray(record.OutputSectionsJSON) {
		return fmt.Errorf("output_sections_json must be a json array")
	}
	if strings.TrimSpace(record.RiskNotesJSON) != "" && !validJSONArray(record.RiskNotesJSON) {
		return fmt.Errorf("risk_notes_json must be a json array")
	}
	return nil
}

func validJSONArray(raw string) bool {
	if strings.TrimSpace(raw) == "" {
		return false
	}
	var values []string
	return json.Unmarshal([]byte(raw), &values) == nil
}
