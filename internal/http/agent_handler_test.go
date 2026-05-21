package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"video-ops-agent/internal/agent/contextbuilder"
	"video-ops-agent/internal/agent/runtime"
	"video-ops-agent/internal/store"
)

func TestAgentSessionEndpointsCreateListAndDetailSessions(t *testing.T) {
	repos := newHTTPTestRepositories(t)
	router := NewRouter(WithAgentHandler(NewAgentHandler(repos, &fakeAgentRuntime{})))

	createBody := `{"user_id":"operator-1","title":"hot rank","scenario":"hot_rank_analysis","context_policy":{"max_recent_messages":2}}`
	createResp := performJSON(router, http.MethodPost, "/agent/sessions", createBody)
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status = %d body=%s", createResp.Code, createResp.Body.String())
	}
	var created struct {
		Session store.AgentSession `json:"session"`
	}
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Session.ID == 0 || created.Session.UserID != "operator-1" || created.Session.Scenario != "hot_rank_analysis" {
		t.Fatalf("created session = %+v", created.Session)
	}
	if !strings.Contains(created.Session.ContextPolicyJSON, "max_recent_messages") {
		t.Fatalf("context policy json = %q", created.Session.ContextPolicyJSON)
	}

	if _, err := repos.Messages.Create(context.Background(), store.CreateMessageInput{
		SessionID: created.Session.ID,
		Role:      store.MessageRoleUser,
		Content:   "hello",
	}); err != nil {
		t.Fatalf("create message: %v", err)
	}

	listResp := performJSON(router, http.MethodGet, "/agent/sessions?user_id=operator-1", "")
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listResp.Code, listResp.Body.String())
	}
	var listed struct {
		Sessions []store.AgentSession `json:"sessions"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listed.Sessions) != 1 || listed.Sessions[0].ID != created.Session.ID {
		t.Fatalf("listed sessions = %+v", listed.Sessions)
	}

	detailResp := performJSON(router, http.MethodGet, "/agent/sessions/"+uintString(created.Session.ID), "")
	if detailResp.Code != http.StatusOK {
		t.Fatalf("detail status = %d body=%s", detailResp.Code, detailResp.Body.String())
	}
	var detail struct {
		Session  store.AgentSession   `json:"session"`
		Messages []store.AgentMessage `json:"messages"`
	}
	if err := json.Unmarshal(detailResp.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode detail response: %v", err)
	}
	if detail.Session.ID != created.Session.ID || len(detail.Messages) != 1 || detail.Messages[0].Content != "hello" {
		t.Fatalf("detail = %+v", detail)
	}
}

func TestAgentMessageEndpointRunsRuntime(t *testing.T) {
	repos := newHTTPTestRepositories(t)
	session := createHTTPSession(t, repos)
	fakeRuntime := &fakeAgentRuntime{
		result: runtime.RunResult{SessionID: session.ID, FinalAnswer: "已完成分析", RoundCount: 2, ToolCallCount: 1},
	}
	router := NewRouter(WithAgentHandler(NewAgentHandler(repos, fakeRuntime)))

	resp := performJSON(router, http.MethodPost, "/agent/sessions/"+uintString(session.ID)+"/messages", `{"content":"分析视频 7","required_evidence":["get_video_detail"]}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("message status = %d body=%s", resp.Code, resp.Body.String())
	}

	if fakeRuntime.request.SessionID != session.ID || fakeRuntime.request.UserMessage != "分析视频 7" ||
		len(fakeRuntime.request.RequiredEvidence) != 1 || fakeRuntime.request.RequiredEvidence[0] != "get_video_detail" {
		t.Fatalf("runtime request = %+v", fakeRuntime.request)
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["final_answer"] != "已完成分析" || body["round_count"] != float64(2) || body["tool_call_count"] != float64(1) {
		t.Fatalf("response body = %#v", body)
	}
}

func TestAgentToolCallsEndpointReturnsTrace(t *testing.T) {
	repos := newHTTPTestRepositories(t)
	session := createHTTPSession(t, repos)
	if _, err := repos.ToolCalls.Create(context.Background(), store.CreateToolCallInput{
		SessionID:     session.ID,
		ToolName:      "get_video_detail",
		ArgumentsJSON: `{"video_id":7}`,
		ResultSummary: "video 7",
		Status:        store.ToolCallStatusSuccess,
	}); err != nil {
		t.Fatalf("create tool call: %v", err)
	}
	router := NewRouter(WithAgentHandler(NewAgentHandler(repos, &fakeAgentRuntime{})))

	resp := performJSON(router, http.MethodGet, "/agent/sessions/"+uintString(session.ID)+"/tool-calls", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("tool calls status = %d body=%s", resp.Code, resp.Body.String())
	}
	var body struct {
		ToolCalls []store.AgentToolCall `json:"tool_calls"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.ToolCalls) != 1 || body.ToolCalls[0].ToolName != "get_video_detail" {
		t.Fatalf("tool calls = %+v", body.ToolCalls)
	}
}

func newHTTPTestRepositories(t *testing.T) contextbuilder.Repositories {
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
	return contextbuilder.Repositories{
		Sessions:  store.NewSessionRepository(db),
		Messages:  store.NewMessageRepository(db),
		ToolCalls: store.NewToolCallRepository(db),
	}
}

func createHTTPSession(t *testing.T, repos contextbuilder.Repositories) store.AgentSession {
	t.Helper()
	session, err := repos.Sessions.Create(context.Background(), store.CreateSessionInput{
		UserID: "operator-1",
		Status: store.SessionStatusActive,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	return session
}

func performJSON(handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
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

type fakeAgentRuntime struct {
	request runtime.RunRequest
	result  runtime.RunResult
	err     error
}

func (r *fakeAgentRuntime) Run(_ context.Context, request runtime.RunRequest) (runtime.RunResult, error) {
	r.request = request
	if r.err != nil {
		return runtime.RunResult{}, r.err
	}
	return r.result, nil
}
