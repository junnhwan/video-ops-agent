package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"video-ops-agent/internal/agent/skills"
	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/store"
)

func TestSkillReadEndpointsReturnBuiltins(t *testing.T) {
	service := newHTTPSkillTestService(t)
	router := NewRouter(WithSkillHandler(NewSkillHandler(service)))

	listResp := performJSON(router, http.MethodGet, "/skills", "")
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listResp.Code, listResp.Body.String())
	}
	var listBody struct {
		Skills []skills.DiagnosisSkill `json:"skills"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listBody.Skills) != 5 {
		t.Fatalf("skill count = %d, want 5: %+v", len(listBody.Skills), listBody.Skills)
	}

	getResp := performJSON(router, http.MethodGet, "/skills/comment_risk_analysis", "")
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", getResp.Code, getResp.Body.String())
	}
	var getBody struct {
		Skill skills.DiagnosisSkill `json:"skill"`
	}
	if err := json.Unmarshal(getResp.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getBody.Skill.ID != "comment_risk_analysis" || len(getBody.Skill.AllowedTools) == 0 {
		t.Fatalf("skill body = %+v", getBody.Skill)
	}
}

func TestSkillWriteEndpointsCreateUpdateEnableAndDisableCustomSkill(t *testing.T) {
	service := newHTTPSkillTestService(t)
	router := NewRouter(WithSkillHandler(NewSkillHandler(service)))

	createBody := `{
		"id":"custom_comment_review",
		"name":"自定义评论复盘",
		"description":"custom skill",
		"version":"1.0.0",
		"status":"enabled",
		"scenario":"comment_risk_analysis",
		"allowed_tools":["get_video_detail","analyze_video_comment_risk"],
		"required_evidence":["get_video_detail"],
		"prompt_template":"Use evidence.",
		"output_sections":["结论","证据"]
	}`
	createResp := performJSON(router, http.MethodPost, "/skills", createBody)
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status = %d body=%s", createResp.Code, createResp.Body.String())
	}

	updateBody := `{
		"name":"自定义评论复盘 v2",
		"description":"custom skill",
		"version":"1.0.1",
		"status":"enabled",
		"scenario":"comment_risk_analysis",
		"allowed_tools":["get_video_detail","analyze_video_comment_risk"],
		"required_evidence":["get_video_detail"],
		"prompt_template":"Use evidence carefully.",
		"output_sections":["结论","证据","建议"]
	}`
	updateResp := performJSON(router, http.MethodPut, "/skills/custom_comment_review", updateBody)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", updateResp.Code, updateResp.Body.String())
	}
	var updateBodyJSON struct {
		Skill skills.DiagnosisSkill `json:"skill"`
	}
	if err := json.Unmarshal(updateResp.Body.Bytes(), &updateBodyJSON); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updateBodyJSON.Skill.Name != "自定义评论复盘 v2" || len(updateBodyJSON.Skill.OutputSections) != 3 {
		t.Fatalf("updated skill = %+v", updateBodyJSON.Skill)
	}

	disableResp := performJSON(router, http.MethodPost, "/skills/custom_comment_review/disable", "")
	if disableResp.Code != http.StatusOK {
		t.Fatalf("disable status = %d body=%s", disableResp.Code, disableResp.Body.String())
	}
	var disableBody struct {
		Skill skills.DiagnosisSkill `json:"skill"`
	}
	if err := json.Unmarshal(disableResp.Body.Bytes(), &disableBody); err != nil {
		t.Fatalf("decode disable response: %v", err)
	}
	if disableBody.Skill.Status != skills.SkillStatusDisabled {
		t.Fatalf("disabled skill = %+v", disableBody.Skill)
	}

	enableResp := performJSON(router, http.MethodPost, "/skills/custom_comment_review/enable", "")
	if enableResp.Code != http.StatusOK {
		t.Fatalf("enable status = %d body=%s", enableResp.Code, enableResp.Body.String())
	}
	var enableBody struct {
		Skill skills.DiagnosisSkill `json:"skill"`
	}
	if err := json.Unmarshal(enableResp.Body.Bytes(), &enableBody); err != nil {
		t.Fatalf("decode enable response: %v", err)
	}
	if enableBody.Skill.Status != skills.SkillStatusEnabled {
		t.Fatalf("enabled skill = %+v", enableBody.Skill)
	}
}

func newHTTPSkillTestService(t *testing.T) *skills.Service {
	t.Helper()
	db, err := store.OpenSQLite(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(db); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})
	if err := store.AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}
	return skills.NewService(skills.Dependencies{
		Registry:   newHTTPTestToolRegistry(t),
		Repository: store.NewSkillRepository(db),
	})
}

func newHTTPTestToolRegistry(t *testing.T) *tools.Registry {
	t.Helper()
	registry, err := tools.NewRegistry(
		httpSkillTestTool{name: "get_video_detail"},
		httpSkillTestTool{name: "get_hot_videos"},
		httpSkillTestTool{name: "get_video_comments"},
		httpSkillTestTool{name: "analyze_video_comment_risk"},
		httpSkillTestTool{name: "analyze_comment_risk"},
		httpSkillTestTool{name: "get_author_profile"},
		httpSkillTestTool{name: "list_author_videos"},
		httpSkillTestTool{name: "list_tag_videos"},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	return registry
}

type httpSkillTestTool struct {
	name string
}

func (t httpSkillTestTool) Name() string {
	return t.name
}

func (t httpSkillTestTool) Schema() tools.ToolSchema {
	return tools.NewFunctionSchema(t.name, "test tool", map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
}

func (t httpSkillTestTool) Timeout() time.Duration {
	return 0
}

func (t httpSkillTestTool) Execute(_ context.Context, _ json.RawMessage) (tools.ToolResult, error) {
	return tools.ToolResult{ToolName: t.name, Summary: "ok"}, nil
}
