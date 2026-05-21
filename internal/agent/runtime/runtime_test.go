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
	"video-ops-agent/internal/agent/events"
	"video-ops-agent/internal/agent/llm"
	"video-ops-agent/internal/agent/skills"
	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/gateway"
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
		RequiredEvidence: []string{"get_video_detail"},
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

	result, err := rt.Run(ctx, RunRequest{SessionID: session.ID, UserMessage: "普通诊断"})
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

func TestRuntimeAllowsFinalAnswerAfterUsingMaxToolRounds(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session := createRuntimeSession(t, repos, "hot_rank_analysis")

	registry, err := tools.NewRegistry(
		fakeRuntimeTool{name: "get_video_detail", summary: "video detail ready", data: map[string]any{"id": 7}},
		fakeRuntimeTool{name: "get_hot_videos", summary: "hot context ready", data: map[string]any{"count": 10}},
	)
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
					Function: llm.FunctionCall{Name: "get_hot_videos", Arguments: json.RawMessage(`{"limit":10}`)},
				}}},
			},
			{
				FinishReason: "stop",
				Message:      llm.Message{Role: llm.RoleAssistant, Content: "基于视频详情和热榜证据，视频 7 具备上榜基础。"},
			},
		},
	}
	rt := newRuntimeForTest(repos, registry, fakeLLM, RuntimeConfig{MaxToolRounds: 2, MaxGuardRetries: 2, TotalTimeout: 5 * time.Second})

	result, err := rt.Run(ctx, RunRequest{
		SessionID:        session.ID,
		UserMessage:      "分析视频 7 为什么上热榜",
		RequiredEvidence: []string{"get_video_detail", "get_hot_videos"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.ToolCallCount != 2 || result.RoundCount != 3 {
		t.Fatalf("result = %+v, want two tool rounds plus one final answer round", result)
	}
	if result.FinalAnswer != "基于视频详情和热榜证据，视频 7 具备上榜基础。" {
		t.Fatalf("final answer = %q", result.FinalAnswer)
	}
}

func TestRuntimeEvidenceGuardForcesMissingToolBeforeFinalAnswer(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session := createRuntimeSession(t, repos, "hot_rank_analysis")

	registry, err := tools.NewRegistry(fakeRuntimeTool{
		name:    "get_video_detail",
		summary: "video 7: hot by author 3",
		data:    map[string]any{"id": 7},
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{
		responses: []llm.ChatResponse{
			{
				FinishReason: "stop",
				Message:      llm.Message{Role: llm.RoleAssistant, Content: "没有证据也先回答。"},
			},
			{
				FinishReason: "tool_calls",
				Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
					ID:       "call_1",
					Type:     "function",
					Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
				}}},
			},
			{
				FinishReason: "stop",
				Message:      llm.Message{Role: llm.RoleAssistant, Content: "基于视频详情证据，视频 7 具备热榜基础。"},
			},
		},
	}
	rt := newRuntimeForTest(repos, registry, fakeLLM, RuntimeConfig{MaxToolRounds: 3, MaxGuardRetries: 2, TotalTimeout: 5 * time.Second})

	result, err := rt.Run(ctx, RunRequest{
		SessionID:        session.ID,
		UserMessage:      "分析视频 7 为什么上热榜",
		RequiredEvidence: []string{"get_video_detail"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.FinalAnswer != "基于视频详情证据，视频 7 具备热榜基础。" {
		t.Fatalf("final answer = %q", result.FinalAnswer)
	}
	if len(fakeLLM.requests) != 3 {
		t.Fatalf("llm requests = %d, want 3", len(fakeLLM.requests))
	}
	if !strings.Contains(joinLLMContent(fakeLLM.requests[1].Messages), "Evidence is incomplete") ||
		!strings.Contains(joinLLMContent(fakeLLM.requests[1].Messages), "get_video_detail") {
		t.Fatalf("second request missing guard retry instruction: %+v", fakeLLM.requests[1].Messages)
	}

	messages, err := repos.Messages.ListBySession(ctx, session.ID, 10)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("messages = %+v, want only user and accepted final assistant answer", messages)
	}
	if messages[1].Content == "没有证据也先回答。" {
		t.Fatalf("early unsupported final answer should not be persisted")
	}
}

func TestRuntimeStopsRepeatingAlreadySatisfiedEvidenceTools(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session := createRuntimeSession(t, repos, "hot_rank_analysis")

	registry, err := tools.NewRegistry(fakeRuntimeTool{
		name:    "get_video_detail",
		summary: "video 7: hot by author 3",
		data:    map[string]any{"id": 7},
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
					Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
				}}},
			},
			{
				FinishReason: "tool_calls",
				Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
					ID:       "call_repeat",
					Type:     "function",
					Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
				}}},
			},
			{
				FinishReason: "stop",
				Message:      llm.Message{Role: llm.RoleAssistant, Content: "基于已有视频详情证据，视频 7 具备热榜基础。"},
			},
		},
	}
	rt := newRuntimeForTest(repos, registry, fakeLLM, RuntimeConfig{MaxToolRounds: 2, MaxGuardRetries: 2, TotalTimeout: 5 * time.Second})

	result, err := rt.Run(ctx, RunRequest{
		SessionID:        session.ID,
		UserMessage:      "分析视频 7 为什么上热榜",
		RequiredEvidence: []string{"get_video_detail"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.FinalAnswer != "基于已有视频详情证据，视频 7 具备热榜基础。" {
		t.Fatalf("final answer = %q", result.FinalAnswer)
	}
	if result.ToolCallCount != 1 {
		t.Fatalf("tool call count = %d, want only the first non-duplicate call", result.ToolCallCount)
	}
	if len(fakeLLM.requests) != 3 {
		t.Fatalf("llm requests = %d, want 3", len(fakeLLM.requests))
	}
	if !strings.Contains(joinLLMContent(fakeLLM.requests[2].Messages), "Evidence is complete") {
		t.Fatalf("third request missing complete-evidence instruction: %+v", fakeLLM.requests[2].Messages)
	}

	toolCalls, err := repos.ToolCalls.ListBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("list tool calls: %v", err)
	}
	if len(toolCalls) != 1 {
		t.Fatalf("tool calls = %+v, want duplicate tool request skipped", toolCalls)
	}
}

func TestRuntimeRedirectsDuplicateToolCallsToMissingEvidence(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session := createRuntimeSession(t, repos, "comment_risk_analysis")

	registry, err := tools.NewRegistry(
		fakeRuntimeTool{name: "get_video_detail", summary: "video 7 ready", data: map[string]any{"id": 7}},
		fakeRuntimeTool{name: "analyze_video_comment_risk", summary: "low risk", data: map[string]any{"risk_level": "low"}},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{
		responses: []llm.ChatResponse{
			{
				FinishReason: "tool_calls",
				Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
					ID:       "call_detail",
					Type:     "function",
					Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
				}}},
			},
			{
				FinishReason: "tool_calls",
				Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
					ID:       "call_repeat",
					Type:     "function",
					Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
				}}},
			},
			{
				FinishReason: "tool_calls",
				Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
					ID:       "call_risk",
					Type:     "function",
					Function: llm.FunctionCall{Name: "analyze_video_comment_risk", Arguments: json.RawMessage(`{"video_id":7,"limit":20}`)},
				}}},
			},
			{
				FinishReason: "stop",
				Message:      llm.Message{Role: llm.RoleAssistant, Content: "基于评论和规则扫描，风险较低。"},
			},
		},
	}
	rt := newRuntimeForTest(repos, registry, fakeLLM, RuntimeConfig{MaxToolRounds: 3, MaxGuardRetries: 2, TotalTimeout: 5 * time.Second})

	result, err := rt.Run(ctx, RunRequest{
		SessionID:        session.ID,
		UserMessage:      "分析视频 7 的评论风险",
		RequiredEvidence: []string{"get_video_detail", "analyze_video_comment_risk"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.FinalAnswer != "基于评论和规则扫描，风险较低。" {
		t.Fatalf("final answer = %q", result.FinalAnswer)
	}
	if result.ToolCallCount != 2 {
		t.Fatalf("tool call count = %d, want original detail call plus risk call", result.ToolCallCount)
	}
	if !strings.Contains(joinLLMContent(fakeLLM.requests[2].Messages), "Evidence is incomplete") ||
		!strings.Contains(joinLLMContent(fakeLLM.requests[2].Messages), "analyze_video_comment_risk") {
		t.Fatalf("third request missing missing-evidence redirect: %+v", fakeLLM.requests[2].Messages)
	}

	toolCalls, err := repos.ToolCalls.ListBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("list tool calls: %v", err)
	}
	if len(toolCalls) != 2 || toolCalls[0].ToolName != "get_video_detail" || toolCalls[1].ToolName != "analyze_video_comment_risk" {
		t.Fatalf("tool calls = %+v, want duplicate detail call skipped before risk", toolCalls)
	}
}

func TestRuntimeSelectedSkillLimitsSchemasAndPersistsSkillTrace(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session, err := repos.Sessions.Create(ctx, store.CreateSessionInput{
		UserID:       "operator-1",
		Scenario:     "comment_risk_analysis",
		SkillID:      "comment_risk_analysis",
		SkillVersion: "1.0.0",
		Status:       store.SessionStatusActive,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	registry, err := tools.NewRegistry(
		fakeRuntimeTool{name: "get_video_detail", summary: "video ready", data: map[string]any{"id": 7}},
		fakeRuntimeTool{name: "analyze_video_comment_risk", summary: "low risk", data: map[string]any{"risk_level": "low"}},
		fakeRuntimeTool{name: "get_hot_videos", summary: "hot videos", data: map[string]any{"count": 10}},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{responses: []llm.ChatResponse{
		{
			FinishReason: "tool_calls",
			Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
				ID:       "call_detail",
				Type:     "function",
				Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
			}}},
		},
		{
			FinishReason: "tool_calls",
			Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
				ID:       "call_risk",
				Type:     "function",
				Function: llm.FunctionCall{Name: "analyze_video_comment_risk", Arguments: json.RawMessage(`{"video_id":7,"limit":20}`)},
			}}},
		},
		{
			FinishReason: "stop",
			Message:      llm.Message{Role: llm.RoleAssistant, Content: "风险较低。"},
		},
	}}
	rt := newRuntimeForTestWithSkills(repos, registry, fakeLLM, skills.NewService(skills.Dependencies{Registry: registry}), RuntimeConfig{MaxToolRounds: 3, TotalTimeout: 5 * time.Second})

	result, err := rt.Run(ctx, RunRequest{SessionID: session.ID, UserMessage: "请分析视频 7 的评论风险"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.FinalAnswer != "风险较低。" {
		t.Fatalf("final answer = %q", result.FinalAnswer)
	}

	firstToolNames := toolSchemaNames(fakeLLM.requests[0].Tools)
	if strings.Join(firstToolNames, ",") != "analyze_video_comment_risk,get_video_detail" {
		t.Fatalf("first request tool schemas = %+v, want only comment-risk skill tools", firstToolNames)
	}
	if strings.Contains(joinLLMContent(fakeLLM.requests[0].Messages), "get_hot_videos") {
		t.Fatalf("skill context should not require disallowed hot tools: %+v", fakeLLM.requests[0].Messages)
	}

	toolCalls, err := repos.ToolCalls.ListBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("list tool calls: %v", err)
	}
	if len(toolCalls) != 2 {
		t.Fatalf("tool calls = %+v, want two", toolCalls)
	}
	for _, call := range toolCalls {
		if call.SkillID != "comment_risk_analysis" || call.SkillVersion != "1.0.0" {
			t.Fatalf("tool call missing skill metadata: %+v", call)
		}
	}
}

func TestRuntimeSkillRequiredEvidenceDrivesGuardRetry(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session, err := repos.Sessions.Create(ctx, store.CreateSessionInput{
		UserID:       "operator-1",
		Scenario:     "comment_risk_analysis",
		SkillID:      "comment_risk_analysis",
		SkillVersion: "1.0.0",
		Status:       store.SessionStatusActive,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	registry, err := tools.NewRegistry(
		fakeRuntimeTool{name: "get_video_detail", summary: "video ready", data: map[string]any{"id": 7}},
		fakeRuntimeTool{name: "analyze_video_comment_risk", summary: "low risk", data: map[string]any{"risk_level": "low"}},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{responses: []llm.ChatResponse{
		{
			FinishReason: "stop",
			Message:      llm.Message{Role: llm.RoleAssistant, Content: "没有证据也先回答。"},
		},
		{
			FinishReason: "tool_calls",
			Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
				ID:       "call_detail",
				Type:     "function",
				Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
			}}},
		},
		{
			FinishReason: "tool_calls",
			Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
				ID:       "call_risk",
				Type:     "function",
				Function: llm.FunctionCall{Name: "analyze_video_comment_risk", Arguments: json.RawMessage(`{"video_id":7,"limit":20}`)},
			}}},
		},
		{
			FinishReason: "stop",
			Message:      llm.Message{Role: llm.RoleAssistant, Content: "基于评论风险证据，风险较低。"},
		},
	}}
	rt := newRuntimeForTestWithSkills(repos, registry, fakeLLM, skills.NewService(skills.Dependencies{Registry: registry}), RuntimeConfig{MaxToolRounds: 3, MaxGuardRetries: 2, TotalTimeout: 5 * time.Second})

	result, err := rt.Run(ctx, RunRequest{SessionID: session.ID, UserMessage: "请分析视频 7 的评论风险"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.FinalAnswer != "基于评论风险证据，风险较低。" {
		t.Fatalf("final answer = %q", result.FinalAnswer)
	}
	if !strings.Contains(joinLLMContent(fakeLLM.requests[1].Messages), "analyze_video_comment_risk") {
		t.Fatalf("guard retry did not use skill required evidence: %+v", fakeLLM.requests[1].Messages)
	}
}

func TestRuntimeDisabledSkillReturnsClearError(t *testing.T) {
	ctx := context.Background()
	repos, skillRepo := newRuntimeTestRepositoriesAndSkillRepo(t)
	session, err := repos.Sessions.Create(ctx, store.CreateSessionInput{
		UserID:       "operator-1",
		Scenario:     "comment_risk_analysis",
		SkillID:      "comment_risk_analysis",
		SkillVersion: "1.0.0",
		Status:       store.SessionStatusActive,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	registry, err := tools.NewRegistry(
		fakeRuntimeTool{name: "get_video_detail", summary: "video ready", data: map[string]any{"id": 7}},
		fakeRuntimeTool{name: "analyze_video_comment_risk", summary: "low risk", data: map[string]any{"risk_level": "low"}},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	skillService := skills.NewService(skills.Dependencies{Registry: registry, Repository: skillRepo})
	if err := skillService.Disable(ctx, "comment_risk_analysis"); err != nil {
		t.Fatalf("disable skill: %v", err)
	}
	rt := newRuntimeForTestWithSkills(repos, registry, &fakeLLMClient{}, skillService, RuntimeConfig{TotalTimeout: 5 * time.Second})

	_, err = rt.Run(ctx, RunRequest{SessionID: session.ID, UserMessage: "请分析视频 7 的评论风险"})
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("Run error = %v, want disabled skill", err)
	}
}

func TestRuntimeRecordsGatewayInvocationForAgentToolCall(t *testing.T) {
	ctx := context.Background()
	repos, gatewayService, gatewayInvocations := newRuntimeTestRepositoriesAndGatewayRecorder(t)
	session := createRuntimeSession(t, repos, "hot_rank_analysis")
	registry, err := tools.NewRegistry(fakeRuntimeTool{
		name:    "get_video_detail",
		summary: "video 7 ready",
		data:    map[string]any{"id": 7},
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{responses: []llm.ChatResponse{
		{
			FinishReason: "tool_calls",
			Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
				ID:       "call_detail",
				Type:     "function",
				Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
			}}},
		},
		{
			FinishReason: "stop",
			Message:      llm.Message{Role: llm.RoleAssistant, Content: "基于视频详情证据完成分析。"},
		},
	}}
	rt := NewRuntime(Dependencies{
		LLM:                fakeLLM,
		ToolRegistry:       registry,
		ToolExecutor:       tools.NewExecutor(registry, time.Second),
		ContextBuilder:     contextbuilder.NewBuilder(repos),
		Repositories:       repos,
		InvocationRecorder: gatewayService,
	}, RuntimeConfig{MaxToolRounds: 2, TotalTimeout: 5 * time.Second})

	if _, err := rt.Run(ctx, RunRequest{
		SessionID:        session.ID,
		UserMessage:      "分析视频 7",
		RequiredEvidence: []string{"get_video_detail"},
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	invocations, err := gatewayInvocations.List(ctx, store.GatewayInvocationFilter{
		Source:    gateway.InvocationSourceAgentRuntime,
		SessionID: &session.ID,
	})
	if err != nil {
		t.Fatalf("list gateway invocations: %v", err)
	}
	if len(invocations) != 1 {
		t.Fatalf("gateway invocations = %+v, want one", invocations)
	}
	if invocations[0].ToolName != "get_video_detail" ||
		invocations[0].Source != gateway.InvocationSourceAgentRuntime ||
		invocations[0].Status != store.ToolCallStatusSuccess ||
		invocations[0].ResultSummary != "video 7 ready" ||
		invocations[0].MessageID == nil {
		t.Fatalf("gateway invocation = %+v", invocations[0])
	}
}

func TestRuntimeEmitsExecutionEvents(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session := createRuntimeSession(t, repos, "hot_rank_analysis")
	registry, err := tools.NewRegistry(fakeRuntimeTool{
		name:    "get_video_detail",
		summary: "video 7 ready",
		data:    map[string]any{"id": 7},
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{responses: []llm.ChatResponse{
		{
			FinishReason: "tool_calls",
			Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
				ID:       "call_detail",
				Type:     "function",
				Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
			}}},
		},
		{
			FinishReason: "stop",
			Message:      llm.Message{Role: llm.RoleAssistant, Content: "基于视频详情证据完成分析。"},
		},
	}}
	sink := &recordingEventSink{}
	rt := newRuntimeForTest(repos, registry, fakeLLM, RuntimeConfig{MaxToolRounds: 2, TotalTimeout: 5 * time.Second})

	if _, err := rt.Run(ctx, RunRequest{
		SessionID:        session.ID,
		UserMessage:      "分析视频 7",
		RequiredEvidence: []string{"get_video_detail"},
		EventSink:        sink,
	}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	gotTypes := sink.types()
	wantTypes := []string{
		events.TypeAgentStart,
		events.TypeLLMRoundStart,
		events.TypeToolCall,
		events.TypeToolResult,
		events.TypeLLMRoundStart,
		events.TypeFinalAnswer,
	}
	if strings.Join(gotTypes, ",") != strings.Join(wantTypes, ",") {
		t.Fatalf("event types = %+v, want %+v", gotTypes, wantTypes)
	}
	if sink.events[2].ToolName != "get_video_detail" || sink.events[3].Summary != "video 7 ready" ||
		sink.events[5].FinalAnswer != "基于视频详情证据完成分析。" {
		t.Fatalf("events = %+v", sink.events)
	}
}

func TestRuntimeEmitsErrorEventForToolFailureAndContinues(t *testing.T) {
	ctx := context.Background()
	repos := newRuntimeTestRepositories(t)
	session := createRuntimeSession(t, repos, "general")
	registry, err := tools.NewRegistry(fakeRuntimeTool{
		name: "get_video_detail",
		err:  errors.New("platform unavailable"),
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	fakeLLM := &fakeLLMClient{responses: []llm.ChatResponse{
		{
			FinishReason: "tool_calls",
			Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
				ID:       "call_detail",
				Type:     "function",
				Function: llm.FunctionCall{Name: "get_video_detail", Arguments: json.RawMessage(`{"video_id":7}`)},
			}}},
		},
		{
			FinishReason: "stop",
			Message:      llm.Message{Role: llm.RoleAssistant, Content: "工具失败，无法完成证据化分析。"},
		},
	}}
	sink := &recordingEventSink{}
	rt := newRuntimeForTest(repos, registry, fakeLLM, RuntimeConfig{MaxToolRounds: 2, TotalTimeout: 5 * time.Second})

	result, err := rt.Run(ctx, RunRequest{
		SessionID:   session.ID,
		UserMessage: "分析视频 7",
		EventSink:   sink,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.FinalAnswer != "工具失败，无法完成证据化分析。" {
		t.Fatalf("final answer = %q", result.FinalAnswer)
	}

	var foundError bool
	for _, event := range sink.events {
		if event.Type == events.TypeError && strings.Contains(event.Error, "platform unavailable") {
			foundError = true
		}
	}
	if !foundError {
		t.Fatalf("events missing tool failure error event: %+v", sink.events)
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

func newRuntimeForTestWithSkills(repos contextbuilder.Repositories, registry *tools.Registry, llmClient LLMClient, skillService *skills.Service, config RuntimeConfig) *Runtime {
	return NewRuntime(Dependencies{
		LLM:            llmClient,
		ToolRegistry:   registry,
		ToolExecutor:   tools.NewExecutor(registry, time.Second),
		ContextBuilder: contextbuilder.NewBuilder(repos),
		Repositories:   repos,
		SkillService:   skillService,
	}, config)
}

func newRuntimeTestRepositoriesAndSkillRepo(t *testing.T) (contextbuilder.Repositories, *store.SkillRepository) {
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
	}, store.NewSkillRepository(db)
}

func newRuntimeTestRepositoriesAndGatewayRecorder(t *testing.T) (contextbuilder.Repositories, *gateway.Service, *store.GatewayInvocationRepository) {
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
	invocations := store.NewGatewayInvocationRepository(db)
	return contextbuilder.Repositories{
		Sessions:  store.NewSessionRepository(db),
		Messages:  store.NewMessageRepository(db),
		ToolCalls: store.NewToolCallRepository(db),
	}, gateway.NewService(gateway.Dependencies{Invocations: invocations}), invocations
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

func toolSchemaNames(schemas []tools.ToolSchema) []string {
	names := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		names = append(names, schema.Function.Name)
	}
	return names
}

type recordingEventSink struct {
	events []events.RuntimeEvent
}

func (s *recordingEventSink) Emit(_ context.Context, event events.RuntimeEvent) error {
	s.events = append(s.events, event)
	return nil
}

func (s *recordingEventSink) types() []string {
	types := make([]string, 0, len(s.events))
	for _, event := range s.events {
		types = append(types, event.Type)
	}
	return types
}
