package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/store"
)

type Dependencies struct {
	Registry   *tools.Registry
	Repository *store.SkillRepository
}

type Service struct {
	registry   *tools.Registry
	repository *store.SkillRepository
}

func NewService(deps Dependencies) *Service {
	return &Service{registry: deps.Registry, repository: deps.Repository}
}

func (s *Service) Builtins() []DiagnosisSkill {
	return BuiltinSkills()
}

func (s *Service) List(ctx context.Context) ([]DiagnosisSkill, error) {
	byID := make(map[string]DiagnosisSkill)
	for _, skill := range BuiltinSkills() {
		byID[skill.ID] = skill
	}
	if s.repository != nil {
		records, err := s.repository.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			skill, err := skillFromRecord(record)
			if err != nil {
				return nil, err
			}
			byID[skill.ID] = skill
		}
	}

	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	skills := make([]DiagnosisSkill, 0, len(ids))
	for _, id := range ids {
		skills = append(skills, cloneSkill(byID[id]))
	}
	return skills, nil
}

func (s *Service) Get(ctx context.Context, id string) (DiagnosisSkill, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return DiagnosisSkill{}, fmt.Errorf("skill id is required")
	}
	if s.repository != nil {
		record, err := s.repository.Get(ctx, id)
		if err == nil {
			return skillFromRecord(record)
		}
		if !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return DiagnosisSkill{}, err
		}
	}
	for _, skill := range BuiltinSkills() {
		if skill.ID == id {
			return skill, nil
		}
	}
	return DiagnosisSkill{}, fmt.Errorf("skill %q not found", id)
}

func (s *Service) GetForRuntime(ctx context.Context, id string) (DiagnosisSkill, error) {
	skill, err := s.Get(ctx, id)
	if err != nil {
		return DiagnosisSkill{}, err
	}
	if err := s.EnsureRuntimeUsable(skill); err != nil {
		return DiagnosisSkill{}, err
	}
	return skill, nil
}

func (s *Service) Create(ctx context.Context, skill DiagnosisSkill) error {
	if s.repository == nil {
		return fmt.Errorf("skill repository is required")
	}
	if _, ok := builtinByID(skill.ID); ok {
		return fmt.Errorf("builtin skill %q cannot be created as custom skill", skill.ID)
	}
	if err := s.ValidateDefinition(skill); err != nil {
		return err
	}
	record, err := recordFromSkill(skill)
	if err != nil {
		return err
	}
	return s.repository.Upsert(ctx, record)
}

func (s *Service) Update(ctx context.Context, id string, skill DiagnosisSkill) error {
	if s.repository == nil {
		return fmt.Errorf("skill repository is required")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("skill id is required")
	}
	skill.ID = id
	if err := s.ValidateDefinition(skill); err != nil {
		return err
	}
	record, err := recordFromSkill(skill)
	if err != nil {
		return err
	}
	return s.repository.Upsert(ctx, record)
}

func (s *Service) Enable(ctx context.Context, id string) error {
	return s.setStatus(ctx, id, SkillStatusEnabled)
}

func (s *Service) Disable(ctx context.Context, id string) error {
	return s.setStatus(ctx, id, SkillStatusDisabled)
}

func (s *Service) ValidateDefinition(skill DiagnosisSkill) error {
	if strings.TrimSpace(skill.ID) == "" {
		return fmt.Errorf("skill id is required")
	}
	if strings.TrimSpace(skill.Name) == "" {
		return fmt.Errorf("skill name is required")
	}
	if strings.TrimSpace(skill.Version) == "" {
		return fmt.Errorf("skill version is required")
	}
	status := strings.TrimSpace(skill.Status)
	if status != SkillStatusEnabled && status != SkillStatusDisabled {
		return fmt.Errorf("skill status must be %q or %q", SkillStatusEnabled, SkillStatusDisabled)
	}
	if len(nonEmptyUnique(skill.AllowedTools)) == 0 {
		return fmt.Errorf("allowed_tools is required")
	}
	requiredEvidence := nonEmptyUnique(skill.RequiredEvidence)
	if len(requiredEvidence) == 0 {
		return fmt.Errorf("required_evidence is required")
	}
	if len(nonEmptyUnique(skill.OutputSections)) == 0 {
		return fmt.Errorf("output_sections is required")
	}
	if strings.TrimSpace(skill.PromptTemplate) == "" {
		return fmt.Errorf("prompt_template is required")
	}
	if s.registry != nil {
		if err := s.validateKnownTools(skill.AllowedTools); err != nil {
			return err
		}
		if err := s.validateKnownTools(skill.RequiredEvidence); err != nil {
			return err
		}
	}
	allowed := stringSet(skill.AllowedTools)
	for _, toolName := range requiredEvidence {
		if _, ok := allowed[toolName]; !ok {
			return fmt.Errorf("required_evidence tool %q must be included in allowed_tools", toolName)
		}
	}
	return nil
}

func (s *Service) EnsureRuntimeUsable(skill DiagnosisSkill) error {
	if err := s.ValidateDefinition(skill); err != nil {
		return err
	}
	if skill.Status != SkillStatusEnabled {
		return fmt.Errorf("skill %q is disabled", skill.ID)
	}
	return nil
}

func (s *Service) validateKnownTools(toolNames []string) error {
	for _, toolName := range nonEmptyUnique(toolNames) {
		if _, ok := s.registry.Get(toolName); !ok {
			return fmt.Errorf("unknown tool %q", toolName)
		}
	}
	return nil
}

func nonEmptyUnique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range nonEmptyUnique(values) {
		set[value] = struct{}{}
	}
	return set
}

func (s *Service) setStatus(ctx context.Context, id string, status string) error {
	if s.repository == nil {
		return fmt.Errorf("skill repository is required")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("skill id is required")
	}
	if err := s.repository.SetStatus(ctx, id, status); err == nil {
		return nil
	} else if !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return err
	}

	builtin, ok := builtinByID(id)
	if !ok {
		return fmt.Errorf("skill %q not found", id)
	}
	builtin.Status = status
	record, err := recordFromSkill(builtin)
	if err != nil {
		return err
	}
	return s.repository.Upsert(ctx, record)
}

func builtinByID(id string) (DiagnosisSkill, bool) {
	for _, skill := range BuiltinSkills() {
		if skill.ID == strings.TrimSpace(id) {
			return skill, true
		}
	}
	return DiagnosisSkill{}, false
}

func recordFromSkill(skill DiagnosisSkill) (store.DiagnosisSkillRecord, error) {
	allowedToolsJSON, err := marshalStringSlice(skill.AllowedTools)
	if err != nil {
		return store.DiagnosisSkillRecord{}, err
	}
	requiredEvidenceJSON, err := marshalStringSlice(skill.RequiredEvidence)
	if err != nil {
		return store.DiagnosisSkillRecord{}, err
	}
	outputSectionsJSON, err := marshalStringSlice(skill.OutputSections)
	if err != nil {
		return store.DiagnosisSkillRecord{}, err
	}
	riskNotesJSON := ""
	if len(skill.RiskNotes) > 0 {
		riskNotesJSON, err = marshalStringSlice(skill.RiskNotes)
		if err != nil {
			return store.DiagnosisSkillRecord{}, err
		}
	}
	return store.DiagnosisSkillRecord{
		ID:                   strings.TrimSpace(skill.ID),
		Name:                 strings.TrimSpace(skill.Name),
		Description:          strings.TrimSpace(skill.Description),
		Version:              strings.TrimSpace(skill.Version),
		Status:               strings.TrimSpace(skill.Status),
		Scenario:             strings.TrimSpace(skill.Scenario),
		AllowedToolsJSON:     allowedToolsJSON,
		RequiredEvidenceJSON: requiredEvidenceJSON,
		PromptTemplate:       strings.TrimSpace(skill.PromptTemplate),
		OutputSectionsJSON:   outputSectionsJSON,
		RiskNotesJSON:        riskNotesJSON,
	}, nil
}

func skillFromRecord(record store.DiagnosisSkillRecord) (DiagnosisSkill, error) {
	allowedTools, err := unmarshalStringSlice(record.AllowedToolsJSON)
	if err != nil {
		return DiagnosisSkill{}, fmt.Errorf("decode allowed tools for skill %q: %w", record.ID, err)
	}
	requiredEvidence, err := unmarshalStringSlice(record.RequiredEvidenceJSON)
	if err != nil {
		return DiagnosisSkill{}, fmt.Errorf("decode required evidence for skill %q: %w", record.ID, err)
	}
	outputSections, err := unmarshalStringSlice(record.OutputSectionsJSON)
	if err != nil {
		return DiagnosisSkill{}, fmt.Errorf("decode output sections for skill %q: %w", record.ID, err)
	}
	var riskNotes []string
	if strings.TrimSpace(record.RiskNotesJSON) != "" {
		riskNotes, err = unmarshalStringSlice(record.RiskNotesJSON)
		if err != nil {
			return DiagnosisSkill{}, fmt.Errorf("decode risk notes for skill %q: %w", record.ID, err)
		}
	}
	return DiagnosisSkill{
		ID:               record.ID,
		Name:             record.Name,
		Description:      record.Description,
		Version:          record.Version,
		Status:           record.Status,
		Scenario:         record.Scenario,
		AllowedTools:     allowedTools,
		RequiredEvidence: requiredEvidence,
		PromptTemplate:   record.PromptTemplate,
		OutputSections:   outputSections,
		RiskNotes:        riskNotes,
	}, nil
}

func marshalStringSlice(values []string) (string, error) {
	encoded, err := json.Marshal(nonEmptyUnique(values))
	if err != nil {
		return "", fmt.Errorf("marshal string slice: %w", err)
	}
	return string(encoded), nil
}

func unmarshalStringSlice(raw string) ([]string, error) {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}
