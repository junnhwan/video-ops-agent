package store

import "time"

const (
	SessionStatusActive = "active"
	SessionStatusClosed = "closed"
	SessionStatusError  = "error"

	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
	MessageRoleTool      = "tool"

	ToolCallStatusSuccess = "success"
	ToolCallStatusError   = "error"
	ToolCallStatusTimeout = "timeout"

	InvocationSourceAgentRuntime  = "agent_runtime"
	InvocationSourceManualConsole = "manual_console"
	InvocationSourceMCPClient     = "mcp_client"

	SkillStatusEnabled  = "enabled"
	SkillStatusDisabled = "disabled"
)

type AgentSession struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UserID            string    `gorm:"size:64;index" json:"user_id"`
	Title             string    `gorm:"size:255" json:"title"`
	Scenario          string    `gorm:"size:64;index" json:"scenario"`
	Status            string    `gorm:"size:32;index;not null" json:"status"`
	ContextPolicyJSON string    `gorm:"type:text" json:"context_policy_json,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (AgentSession) TableName() string {
	return "agent_sessions"
}

type AgentMessage struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SessionID      uint      `gorm:"index;not null" json:"session_id"`
	Role           string    `gorm:"size:32;index;not null" json:"role"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	ContentSummary string    `gorm:"type:text" json:"content_summary,omitempty"`
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
}

func (AgentMessage) TableName() string {
	return "agent_messages"
}

type AgentToolCall struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SessionID     uint      `gorm:"index;not null" json:"session_id"`
	MessageID     *uint     `gorm:"index" json:"message_id,omitempty"`
	ToolName      string    `gorm:"size:128;index;not null" json:"tool_name"`
	ArgumentsJSON string    `gorm:"type:text;not null" json:"arguments_json"`
	ResultJSON    string    `gorm:"type:text" json:"result_json,omitempty"`
	ResultSummary string    `gorm:"type:text" json:"result_summary,omitempty"`
	LatencyMS     int64     `gorm:"not null;default:0" json:"latency_ms"`
	Status        string    `gorm:"size:32;index;not null" json:"status"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

func (AgentToolCall) TableName() string {
	return "agent_tool_calls"
}

type GatewayToolInvocation struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Source        string    `gorm:"size:32;index;not null" json:"source"`
	SessionID     *uint     `gorm:"index" json:"session_id,omitempty"`
	MessageID     *uint     `gorm:"index" json:"message_id,omitempty"`
	SkillID       string    `gorm:"size:64;index" json:"skill_id,omitempty"`
	SkillVersion  string    `gorm:"size:32" json:"skill_version,omitempty"`
	ToolName      string    `gorm:"size:128;index;not null" json:"tool_name"`
	ArgumentsJSON string    `gorm:"type:text;not null" json:"arguments_json"`
	ResultJSON    string    `gorm:"type:text" json:"result_json,omitempty"`
	ResultSummary string    `gorm:"type:text" json:"result_summary,omitempty"`
	LatencyMS     int64     `gorm:"not null;default:0" json:"latency_ms"`
	Status        string    `gorm:"size:32;index;not null" json:"status"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

func (GatewayToolInvocation) TableName() string {
	return "gateway_tool_invocations"
}

type DiagnosisSkillRecord struct {
	ID                   string    `gorm:"primaryKey;size:64" json:"id"`
	Name                 string    `gorm:"size:128;not null" json:"name"`
	Description          string    `gorm:"type:text" json:"description"`
	Version              string    `gorm:"size:32;not null" json:"version"`
	Status               string    `gorm:"size:32;index;not null" json:"status"`
	Scenario             string    `gorm:"size:64;index" json:"scenario"`
	AllowedToolsJSON     string    `gorm:"type:text;not null" json:"allowed_tools_json"`
	RequiredEvidenceJSON string    `gorm:"type:text;not null" json:"required_evidence_json"`
	PromptTemplate       string    `gorm:"type:text;not null" json:"prompt_template"`
	OutputSectionsJSON   string    `gorm:"type:text;not null" json:"output_sections_json"`
	RiskNotesJSON        string    `gorm:"type:text" json:"risk_notes_json,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (DiagnosisSkillRecord) TableName() string {
	return "diagnosis_skill_records"
}
