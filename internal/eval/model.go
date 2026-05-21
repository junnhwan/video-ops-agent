package eval

import "time"

const (
	ModeBaseline   = "baseline"
	ModeSkillGuard = "skill_guard"
)

type SummaryFilter struct {
	SkillID string
}

type Summary struct {
	ToolCallSuccessRate                        float64  `json:"tool_call_success_rate"`
	ToolCallErrorCount                         int      `json:"tool_call_error_count"`
	UnauthorizedToolCallCount                  int      `json:"unauthorized_tool_call_count"`
	GuardRetryCount                            *int     `json:"guard_retry_count"`
	EvidenceCompleteFinalAnswerCount           int      `json:"evidence_complete_final_answer_count"`
	EvidenceIncompleteFinalAnswerRejectedCount *int     `json:"evidence_incomplete_final_answer_rejected_count"`
	AverageToolLatencyMS                       float64  `json:"average_tool_latency_ms"`
	AverageRoundCount                          *float64 `json:"average_round_count"`
	AverageToolCallCount                       float64  `json:"average_tool_call_count"`
	SkillSuccessCount                          int      `json:"skill_success_count"`
	SkillFailureCount                          int      `json:"skill_failure_count"`
	UnsupportedMetrics                         []string `json:"unsupported_metrics,omitempty"`
}

type CreateRunInput struct {
	Mode    string `json:"mode"`
	SkillID string `json:"skill_id,omitempty"`
}

type EvalRun struct {
	ID        uint      `json:"id"`
	Mode      string    `json:"mode"`
	SkillID   string    `json:"skill_id,omitempty"`
	Summary   Summary   `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
}
