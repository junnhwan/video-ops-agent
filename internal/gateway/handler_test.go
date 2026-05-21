package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/store"

	"github.com/gin-gonic/gin"
)

func TestHandlerListsToolsCallsToolAndListsInvocations(t *testing.T) {
	ctx := context.Background()
	db := newGatewayTestDB(t)
	session, err := store.NewSessionRepository(db).Create(ctx, store.CreateSessionInput{UserID: "operator-1"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	registry, err := tools.NewRegistry(gatewayTestTool{
		name:        "get_video_detail",
		description: "Fetch video.",
		result: tools.ToolResult{
			ToolName: "get_video_detail",
			Summary:  "video 101: test",
			Data:     map[string]any{"id": float64(101), "title": "test"},
		},
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	service := NewService(Dependencies{
		Registry:    registry,
		Executor:    tools.NewExecutor(registry, time.Second),
		Invocations: store.NewGatewayInvocationRepository(db),
	})
	router := newGatewayHandlerTestRouter(NewHandler(service))

	toolsResp := performGatewayJSON(router, http.MethodGet, "/gateway/tools", "")
	if toolsResp.Code != http.StatusOK {
		t.Fatalf("tools status = %d body=%s", toolsResp.Code, toolsResp.Body.String())
	}
	var toolsBody struct {
		Tools []ToolCatalogItem `json:"tools"`
	}
	if err := json.Unmarshal(toolsResp.Body.Bytes(), &toolsBody); err != nil {
		t.Fatalf("decode tools response: %v", err)
	}
	if len(toolsBody.Tools) != 1 || toolsBody.Tools[0].Name != "get_video_detail" || !toolsBody.Tools[0].ReadOnly {
		t.Fatalf("tools body = %+v", toolsBody)
	}

	callResp := performGatewayJSON(
		router,
		http.MethodPost,
		"/gateway/tools/get_video_detail/call",
		`{"source":"manual_console","session_id":`+uintString(session.ID)+`,"skill_id":"comment_risk_analysis","arguments":{"video_id":101}}`,
	)
	if callResp.Code != http.StatusOK {
		t.Fatalf("call status = %d body=%s", callResp.Code, callResp.Body.String())
	}
	var callBody CallToolOutput
	if err := json.Unmarshal(callResp.Body.Bytes(), &callBody); err != nil {
		t.Fatalf("decode call response: %v", err)
	}
	if callBody.Invocation.ID == 0 || callBody.Invocation.Source != InvocationSourceManualConsole ||
		callBody.Invocation.ToolName != "get_video_detail" || callBody.Result.Summary != "video 101: test" {
		t.Fatalf("call body = %+v", callBody)
	}

	listResp := performGatewayJSON(
		router,
		http.MethodGet,
		"/gateway/invocations?source=manual_console&session_id="+uintString(session.ID),
		"",
	)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listResp.Code, listResp.Body.String())
	}
	var listBody struct {
		Invocations []store.GatewayToolInvocation `json:"invocations"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listBody.Invocations) != 1 || listBody.Invocations[0].SkillID != "comment_risk_analysis" {
		t.Fatalf("list body = %+v", listBody)
	}
}

func newGatewayHandlerTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	handler.RegisterRoutes(router)
	return router
}

func performGatewayJSON(handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
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

func uintString(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}
