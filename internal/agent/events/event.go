package events

const (
	TypeAgentStart   = "agent_start"
	TypeSkillLoaded  = "skill_loaded"
	TypeLLMRoundStart = "llm_round_start"
	TypeToolCall     = "tool_call"
	TypeToolResult   = "tool_result"
	TypeGuardRetry   = "guard_retry"
	TypeFinalAnswer  = "final_answer"
	TypeError        = "error"
)

type RuntimeEvent struct {
	Type          string         `json:"type"`
	SessionID     uint           `json:"session_id"`
	SkillID       string         `json:"skill_id,omitempty"`
	ToolName      string         `json:"tool_name,omitempty"`
	Arguments     map[string]any `json:"arguments,omitempty"`
	Summary       string         `json:"summary,omitempty"`
	Status        string         `json:"status,omitempty"`
	Error         string         `json:"error,omitempty"`
	FinalAnswer   string         `json:"final_answer,omitempty"`
	RoundCount    int            `json:"round_count,omitempty"`
	ToolCallCount int            `json:"tool_call_count,omitempty"`
}
