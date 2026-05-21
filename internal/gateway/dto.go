package gateway

import "encoding/json"

type CallToolRequest struct {
	Arguments    json.RawMessage `json:"arguments"`
	Source       string          `json:"source"`
	SessionID    *uint           `json:"session_id"`
	SkillID      string          `json:"skill_id"`
	SkillVersion string          `json:"skill_version"`
}
