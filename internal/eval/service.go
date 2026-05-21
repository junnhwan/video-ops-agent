package eval

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"video-ops-agent/internal/agent/skills"
	"video-ops-agent/internal/store"
)

type Dependencies struct {
	Sessions    *store.SessionRepository
	Invocations *store.GatewayInvocationRepository
	Skills      *skills.Service
}

type Service struct {
	sessions    *store.SessionRepository
	invocations *store.GatewayInvocationRepository
	skills      *skills.Service

	mu     sync.Mutex
	nextID uint
	runs   map[uint]EvalRun
}

func NewService(deps Dependencies) *Service {
	return &Service{
		sessions:    deps.Sessions,
		invocations: deps.Invocations,
		skills:      deps.Skills,
		nextID:      1,
		runs:        make(map[uint]EvalRun),
	}
}

func (s *Service) Summary(ctx context.Context, filter SummaryFilter) (Summary, error) {
	if s.sessions == nil || s.invocations == nil {
		return Summary{}, fmt.Errorf("eval repositories are required")
	}
	skillID := strings.TrimSpace(filter.SkillID)
	invocationFilter := store.GatewayInvocationFilter{Limit: 200}
	if skillID != "" {
		invocationFilter.SkillID = skillID
	}
	invocations, err := s.invocations.List(ctx, invocationFilter)
	if err != nil {
		return Summary{}, err
	}
	sessions, err := s.sessions.List(ctx, "", 100)
	if err != nil {
		return Summary{}, err
	}

	summary := Summary{
		UnsupportedMetrics: []string{
			"guard_retry_count requires persisted runtime events",
			"evidence_incomplete_final_answer_rejected_count requires persisted guard rejections",
			"average_round_count requires persisted run results",
		},
	}
	if len(invocations) > 0 {
		successCount := 0
		var latencySum int64
		sessionToolCounts := make(map[uint]int)
		for _, invocation := range invocations {
			if invocation.Status == store.ToolCallStatusSuccess {
				successCount++
			} else {
				summary.ToolCallErrorCount++
			}
			latencySum += invocation.LatencyMS
			if invocation.SessionID != nil {
				sessionToolCounts[*invocation.SessionID]++
			}
			if s.isUnauthorized(ctx, invocation) {
				summary.UnauthorizedToolCallCount++
			}
		}
		summary.ToolCallSuccessRate = float64(successCount) / float64(len(invocations))
		summary.AverageToolLatencyMS = float64(latencySum) / float64(len(invocations))
		if len(sessionToolCounts) > 0 {
			summary.AverageToolCallCount = float64(len(invocations)) / float64(len(sessionToolCounts))
		}
	}

	successfulToolsBySession := successfulToolsBySession(invocations)
	for _, session := range sessions {
		if strings.TrimSpace(session.SkillID) == "" {
			continue
		}
		if skillID != "" && session.SkillID != skillID {
			continue
		}
		skill, err := s.skills.Get(ctx, session.SkillID)
		if err != nil {
			summary.SkillFailureCount++
			continue
		}
		if hasRequiredEvidence(successfulToolsBySession[session.ID], skill.RequiredEvidence) {
			summary.EvidenceCompleteFinalAnswerCount++
			summary.SkillSuccessCount++
		} else {
			summary.SkillFailureCount++
		}
	}
	return summary, nil
}

func (s *Service) CreateRun(ctx context.Context, input CreateRunInput) (EvalRun, error) {
	mode := strings.TrimSpace(input.Mode)
	if mode == "" {
		mode = ModeSkillGuard
	}
	if mode != ModeBaseline && mode != ModeSkillGuard {
		return EvalRun{}, fmt.Errorf("mode must be %q or %q", ModeBaseline, ModeSkillGuard)
	}
	summary, err := s.Summary(ctx, SummaryFilter{SkillID: input.SkillID})
	if err != nil {
		return EvalRun{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	run := EvalRun{
		ID:        s.nextID,
		Mode:      mode,
		SkillID:   strings.TrimSpace(input.SkillID),
		Summary:   summary,
		CreatedAt: time.Now(),
	}
	s.runs[run.ID] = run
	s.nextID++
	return run, nil
}

func (s *Service) GetRun(_ context.Context, id uint) (EvalRun, error) {
	if id == 0 {
		return EvalRun{}, fmt.Errorf("eval run id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[id]
	if !ok {
		return EvalRun{}, fmt.Errorf("eval run %d not found", id)
	}
	return run, nil
}

func (s *Service) isUnauthorized(ctx context.Context, invocation store.GatewayToolInvocation) bool {
	if s.skills == nil || strings.TrimSpace(invocation.SkillID) == "" {
		return false
	}
	skill, err := s.skills.Get(ctx, invocation.SkillID)
	if err != nil {
		return true
	}
	allowed := make(map[string]struct{}, len(skill.AllowedTools))
	for _, toolName := range skill.AllowedTools {
		allowed[toolName] = struct{}{}
	}
	_, ok := allowed[invocation.ToolName]
	return !ok
}

func successfulToolsBySession(invocations []store.GatewayToolInvocation) map[uint]map[string]struct{} {
	out := make(map[uint]map[string]struct{})
	for _, invocation := range invocations {
		if invocation.SessionID == nil || invocation.Status != store.ToolCallStatusSuccess {
			continue
		}
		tools, ok := out[*invocation.SessionID]
		if !ok {
			tools = make(map[string]struct{})
			out[*invocation.SessionID] = tools
		}
		tools[invocation.ToolName] = struct{}{}
	}
	return out
}

func hasRequiredEvidence(successfulTools map[string]struct{}, required []string) bool {
	for _, toolName := range required {
		if _, ok := successfulTools[toolName]; !ok {
			return false
		}
	}
	return true
}
