package skills

const (
	SkillStatusEnabled  = "enabled"
	SkillStatusDisabled = "disabled"
)

type DiagnosisSkill struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Version          string   `json:"version"`
	Status           string   `json:"status"`
	Scenario         string   `json:"scenario"`
	AllowedTools     []string `json:"allowed_tools"`
	RequiredEvidence []string `json:"required_evidence"`
	PromptTemplate   string   `json:"prompt_template"`
	OutputSections   []string `json:"output_sections"`
	RiskNotes        []string `json:"risk_notes,omitempty"`
}
