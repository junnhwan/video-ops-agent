package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"video-ops-agent/internal/agent/tools"
)

func TestChatSendsOpenAICompatibleRequestAndParsesFinalAnswer(t *testing.T) {
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireLLMRequest(t, r, "secret-key", "/v1/chat/completions")
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		writeLLMJSON(t, w, http.StatusOK, map[string]any{
			"id": "chatcmpl-final",
			"choices": []map[string]any{{
				"index":         0,
				"finish_reason": "stop",
				"message": map[string]any{
					"role":    "assistant",
					"content": "结论：评论区风险较低。",
				},
			}},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL + "/v1", Model: "ops-model", APIKey: "secret-key"})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "分析评论风险"}},
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}

	if captured["model"] != "ops-model" {
		t.Fatalf("model = %#v, want ops-model", captured["model"])
	}
	messages := captured["messages"].([]any)
	if messages[0].(map[string]any)["content"] != "分析评论风险" {
		t.Fatalf("messages = %#v", messages)
	}
	if resp.FinishReason != "stop" || resp.Message.Content != "结论：评论区风险较低。" {
		t.Fatalf("response = %+v", resp)
	}
	if len(resp.Message.ToolCalls) != 0 {
		t.Fatalf("expected no tool calls, got %+v", resp.Message.ToolCalls)
	}
}

func TestChatSendsToolSchemasAndParsesToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireLLMRequest(t, r, "secret-key", "/chat/completions")
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		toolsPayload, ok := request["tools"].([]any)
		if !ok || len(toolsPayload) != 1 {
			t.Fatalf("tools payload = %#v, want one tool schema", request["tools"])
		}
		if request["tool_choice"] != "auto" {
			t.Fatalf("tool_choice = %#v, want auto", request["tool_choice"])
		}
		writeLLMJSON(t, w, http.StatusOK, map[string]any{
			"id": "chatcmpl-tool",
			"choices": []map[string]any{{
				"index":         0,
				"finish_reason": "tool_calls",
				"message": map[string]any{
					"role":    "assistant",
					"content": nil,
					"tool_calls": []map[string]any{{
						"id":   "call_1",
						"type": "function",
						"function": map[string]any{
							"name":      "get_video_detail",
							"arguments": `{"video_id":7}`,
						},
					}},
				},
			}},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, Model: "ops-model", APIKey: "secret-key"})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "分析视频 7"}},
		Tools: []tools.ToolSchema{tools.NewFunctionSchema("get_video_detail", "get video", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"video_id": map[string]any{"type": "integer"},
			},
			"required": []string{"video_id"},
		})},
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}

	if resp.FinishReason != "tool_calls" {
		t.Fatalf("finish reason = %q, want tool_calls", resp.FinishReason)
	}
	if len(resp.Message.ToolCalls) != 1 {
		t.Fatalf("tool calls = %+v, want one", resp.Message.ToolCalls)
	}
	call := resp.Message.ToolCalls[0]
	if call.ID != "call_1" || call.Function.Name != "get_video_detail" || string(call.Function.Arguments) != `{"video_id":7}` {
		t.Fatalf("tool call = %+v", call)
	}
}

func TestChatReturnsAPIErrorWithoutLeakingAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireLLMRequest(t, r, "secret-key", "/chat/completions")
		writeLLMJSON(t, w, http.StatusUnauthorized, map[string]any{
			"error": map[string]any{"message": "bad api key"},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, Model: "ops-model", APIKey: "secret-key"})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hello"}}})
	if err == nil {
		t.Fatalf("expected API error")
	}
	if strings.Contains(err.Error(), "secret-key") {
		t.Fatalf("error leaked API key: %q", err.Error())
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "bad api key") {
		t.Fatalf("error = %q, want status and provider message", err.Error())
	}
}

func TestChatReturnsErrorForMalformedToolArguments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeLLMJSON(t, w, http.StatusOK, map[string]any{
			"choices": []map[string]any{{
				"finish_reason": "tool_calls",
				"message": map[string]any{
					"role": "assistant",
					"tool_calls": []map[string]any{{
						"id":   "call_bad",
						"type": "function",
						"function": map[string]any{
							"name":      "get_video_detail",
							"arguments": `{"video_id":`,
						},
					}},
				},
			}},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, Model: "ops-model"})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hello"}}})
	if err == nil {
		t.Fatalf("expected malformed arguments error")
	}
	if !strings.Contains(err.Error(), "tool arguments") {
		t.Fatalf("error = %q, want tool arguments context", err.Error())
	}
}

func requireLLMRequest(t *testing.T, r *http.Request, apiKey string, path string) {
	t.Helper()
	if r.Method != http.MethodPost {
		t.Fatalf("method = %q, want POST", r.Method)
	}
	if r.URL.Path != path {
		t.Fatalf("path = %q, want %s", r.URL.Path, path)
	}
	if r.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("content-type = %q, want application/json", r.Header.Get("Content-Type"))
	}
	if got := r.Header.Get("Authorization"); got != "Bearer "+apiKey {
		t.Fatalf("authorization = %q, want bearer token", got)
	}
}

func writeLLMJSON(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}
