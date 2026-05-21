package eval

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"video-ops-agent/internal/agent/skills"
	"video-ops-agent/internal/store"

	"github.com/gin-gonic/gin"
)

func TestHandlerSummaryAndRunEndpoints(t *testing.T) {
	db := newEvalTestDB(t)
	service := NewService(Dependencies{
		Sessions:    store.NewSessionRepository(db),
		Invocations: store.NewGatewayInvocationRepository(db),
		Skills:      skills.NewService(skills.Dependencies{Registry: newEvalSkillRegistry(t)}),
	})
	router := newEvalTestRouter(NewHandler(service))

	summaryResp := performEvalJSON(router, http.MethodGet, "/eval/summary", "")
	if summaryResp.Code != http.StatusOK {
		t.Fatalf("summary status = %d body=%s", summaryResp.Code, summaryResp.Body.String())
	}
	var summaryBody struct {
		Summary Summary `json:"summary"`
	}
	if err := json.Unmarshal(summaryResp.Body.Bytes(), &summaryBody); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
	if len(summaryBody.Summary.UnsupportedMetrics) == 0 {
		t.Fatalf("summary = %+v", summaryBody.Summary)
	}

	skillResp := performEvalJSON(router, http.MethodGet, "/eval/skills/comment_risk_analysis/summary", "")
	if skillResp.Code != http.StatusOK {
		t.Fatalf("skill summary status = %d body=%s", skillResp.Code, skillResp.Body.String())
	}

	createResp := performEvalJSON(router, http.MethodPost, "/eval/runs", `{"mode":"skill_guard","skill_id":"comment_risk_analysis"}`)
	if createResp.Code != http.StatusOK {
		t.Fatalf("create run status = %d body=%s", createResp.Code, createResp.Body.String())
	}
	var createBody struct {
		Run EvalRun `json:"run"`
	}
	if err := json.Unmarshal(createResp.Body.Bytes(), &createBody); err != nil {
		t.Fatalf("decode create run: %v", err)
	}
	if createBody.Run.ID == 0 || createBody.Run.Mode != ModeSkillGuard {
		t.Fatalf("created run = %+v", createBody.Run)
	}

	getResp := performEvalJSON(router, http.MethodGet, "/eval/runs/1", "")
	if getResp.Code != http.StatusOK {
		t.Fatalf("get run status = %d body=%s", getResp.Code, getResp.Body.String())
	}
}

func newEvalTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	handler.RegisterRoutes(router)
	return router
}

func performEvalJSON(handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}
