package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"video-ops-agent/internal/agent/contextbuilder"
	"video-ops-agent/internal/agent/llm"
	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/store"
)

func TestRuntimeExecutesToolCallsAndPersistsTraceThenFinalAnswer(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session := createRuntimeSession(t, repos, "hot_rank_analysis")

	registry, err := tools.NewRegistry(fakeRuntimeTool{
		name:    "get_video_detail",
		summary: "video 7: hot by author 3",
		data:    map[string]any{"id": 7, "title": "hot"},
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{
		responses: []llm.ChatResponse{
			{
				FinishReason: "tool_calls",
				Message: llm.Message{
					Role: llm.RoleAssistant,
					ToolCalls: []llm.ToolCall{{
						ID:   "call_1",
						Type: "function",
						Function: llm.FunctionCall{
							Name:      "get_video_detail",
							Arguments: json.RawMessage(`{"video_id":7}`),
						},
					}},
				},
			},
			{
				FinishReason: "stop",
				Message: llm.Message{
					Role:    llm.RoleAssistant,
					Content: "视频 7 的热度来自互动表现和作者基础。",
				},
			},
		},
	}
	rt := newRuntimeForTest(repos, registry, fakeLLM, RuntimeConfig{MaxToolRounds: 6, TotalTimeout: 5 * time.Second})

	result, err := rt.Run(ctx, RunRequest{
		SessionID:        session.ID,
		UserMessage:      "分析视频 7 为什么上热榜",
		RequiredEvidence: []string{"get_video_detail", "get_hot_videos"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.FinalAnswer != "视频 7 的热度来自互动表现和作者基础。" {
		t.Fatalf("final answer = %q", result.FinalAnswer)
	}
	if result.ToolCallCount != 1 || result.RoundCount != 2 {
		t.Fatalf("result = %+v, want one tool call and two llm rounds", result)
	}
	if len(fakeLLM.requests) != 2 {
		t.Fatalf("llm requests = %d, want 2", len(fakeLLM.requests))
	}
	if !strings.Contains(joinLLMContent(fakeLLM.requests[1].Messages), "video 7: hot by author 3") {
		t.Fatalf("second LLM request missing persisted tool summary: %+v", fakeLLM.requests[1].Messages)
	}

	messages, err := repos.Messages.ListBySession(ctx, session.ID, 10)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 2 || messages[0].Role != store.MessageRoleUser || messages[1].Role != store.MessageRoleAssistant {
		t.Fatalf("messages = %+v, want persisted user and assistant", messages)
	}

	toolCalls, err := repos.ToolCalls.ListBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("list tool calls: %v", err)
	}
	if len(toolCalls) != 1 {
		t.Fatalf("tool calls = %+v, want one", toolCalls)
	}
	if toolCalls[0].ToolName != "get_video_detail" ||
		toolCalls[0].ArgumentsJSON != `{"video_id":7}` ||
		toolCalls[0].ResultSummary != "video 7: hot by author 3" ||
		toolCalls[0].Status != store.ToolCallStatusSuccess ||
		toolCalls[0].LatencyMS < 0 {
		t.Fatalf("tool call trace = %+v", toolCalls[0])
	}
}

func TestRuntimePersistsToolErrorAndContinues(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session := createRuntimeSession(t, repos, "comment_risk_analysis")

	registry, err := tools.NewRegistry(fakeRuntimeTool{
		name: "get_video_comments",
		err:  errors.New("platform unavailable"),
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{
		responses: []llm.ChatResponse{
			{
				FinishReason: "tool_calls",
				Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
					ID:       "call_1",
					Type:     "function",
					Function: llm.FunctionCall{Name: "get_video_comments", Arguments: json.RawMessage(`{"video_id":7,"limit":50}`)},
				}}},
			},
			{
				FinishReason: "stop",
				Message:      llm.Message{Role: llm.RoleAssistant, Content: "评论接口暂时不可用，无法完成风险结论。"},
			},
		},
	}
	rt := newRuntimeForTest(repos, registry, fakeLLM, RuntimeConfig{MaxToolRounds: 2, TotalTimeout: 5 * time.Second})

	result, err := rt.Run(ctx, RunRequest{SessionID: session.ID, UserMessage: "分析评论风险"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.FinalAnswer == "" {
		t.Fatalf("expected final answer after tool error")
	}

	toolCalls, err := repos.ToolCalls.ListBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("list tool calls: %v", err)
	}
	if len(toolCalls) != 1 {
		t.Fatalf("tool calls = %+v, want one", toolCalls)
	}
	if toolCalls[0].Status != store.ToolCallStatusError || !strings.Contains(toolCalls[0].ErrorMessage, "platform unavailable") {
		t.Fatalf("tool call trace = %+v, want persisted error", toolCalls[0])
	}
}

func TestRuntimeStopsAfterMaxToolRounds(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session := createRuntimeSession(t, repos, "hot_rank_analysis")

	registry, err := tools.NewRegistry(fakeRuntimeTool{name: "get_video_detail", summary: "summary", data: map[string]any{"id": 7}})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{
		responses: []llm.ChatResponse{
			{
				FinishReason: "tool_calls",
				Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
					ID:       "call_1",
					Type:     "function",
					Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
				}}},
			},
			{
				FinishReason: "tool_calls",
				Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
					ID:       "call_2",
					Type:     "function",
					Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":8}`)},
				}}},
			},
		},
	}
	rt := newRuntimeForTest(repos, registry, fakeLLM, RuntimeConfig{MaxToolRounds: 1, TotalTimeout: 5 * time.Second})

	_, err = rt.Run(ctx, RunRequest{SessionID: session.ID, UserMessage: "一直调用工具"})
	if err == nil {
		t.Fatalf("expected max tool rounds error")
	}
	if !strings.Contains(err.Error(), "max tool rounds") {
		t.Fatalf("error = %q, want max tool rounds context", err.Error())
	}

	toolCalls, err := repos.ToolCalls.ListBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("list tool calls: %v", err)
	}
	if len(toolCalls) != 1 {
		t.Fatalf("tool calls = %+v, want exactly one executed call", toolCalls)
	}
}

func newRuntimeForTest(repos contextbuilder.Repositories, registry *tools.Registry, llmClient LLMClient, config RuntimeConfig) *Runtime {
	return NewRuntime(Dependencies{
		LLM:            llmClient,
		ToolRegistry:   registry,
		ToolExecutor:   tools.NewExecutor(registry, time.Second),
		ContextBuilder: contextbuilder.NewBuilder(repos),
		Repositories:   repos,
	}, config)
}

func newRuntimeTestRepositories(t *testing.T) contextbuilder.Repositories {
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

func createRuntimeSession(t *testing.T, repos contextbuilder.Repositories, scenario string) store.AgentSession {
	t.Helper()
	session, err := repos.Sessions.Create(context.Background(), store.CreateSessionInput{
		UserID:            "operator-1",
		Scenario:          scenario,
		Status:            store.SessionStatusActive,
		ContextPolicyJSON: `{"max_recent_messages":6,"max_tool_result_chars":1000,"max_comments_for_llm":10,"max_comment_chars_each":120}`,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	return session
}

type fakeLLMClient struct {
	requests  []llm.ChatRequest
	responses []llm.ChatResponse
	err       error
}

func (c *fakeLLMClient) Chat(_ context.Context, request llm.ChatRequest) (llm.ChatResponse, error) {
	c.requests = append(c.requests, request)
	if c.err != nil {
		return llm.ChatResponse{}, c.err
	}
	if len(c.responses) == 0 {
		return llm.ChatResponse{}, errors.New("no fake llm response configured")
	}
	response := c.responses[0]
	c.responses = c.responses[1:]
	return response, nil
}

type fakeRuntimeTool struct {
	name    string
	summary string
	data    any
	err     error
}

func (t fakeRuntimeTool) Name() string {
	return t.name
}

func (t fakeRuntimeTool) Schema() tools.ToolSchema {
	return tools.NewFunctionSchema(t.name, "fake tool", map[string]any{"type": "object"})
}

func (t fakeRuntimeTool) Timeout() time.Duration {
	return time.Second
}

func (t fakeRuntimeTool) Execute(_ context.Context, _ json.RawMessage) (tools.ToolResult, error) {
	if t.err != nil {
		return tools.ToolResult{}, t.err
	}
	return tools.ToolResult{ToolName: t.name, Data: t.data, Summary: t.summary}, nil
}

func joinLLMContent(messages []llm.Message) string {
	var builder strings.Builder
	for _, message := range messages {
		builder.WriteString(message.Content)
		builder.WriteString("\n")
	}
	return builder.String()
}
