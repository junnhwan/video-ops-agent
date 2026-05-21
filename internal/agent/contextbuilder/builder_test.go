package contextbuilder

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"video-ops-agent/internal/agent/llm"
	"video-ops-agent/internal/store"
)

func TestBuilderSelectsRecentMessagesToolSummariesAndEvidence(t *testing.T) {
	ctx := context.Background()
	repos := newTestRepositories(t)

	session, err := repos.Sessions.Create(ctx, store.CreateSessionInput{
		UserID:            "operator-1",
		Title:             "hot rank",
		Scenario:          "hot_rank_analysis",
		Status:            store.SessionStatusActive,
		ContextPolicyJSON: `{"max_recent_messages":2,"max_tool_result_chars":120,"max_comments_for_llm":3,"max_comment_chars_each":40}`,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	createMessage(t, repos.Messages, session.ID, store.MessageRoleUser, "old user", "")
	createMessage(t, repos.Messages, session.ID, store.MessageRoleAssistant, "old assistant", "")
	createMessage(t, repos.Messages, session.ID, store.MessageRoleUser, "recent user raw content", "recent user summary")
	createMessage(t, repos.Messages, session.ID, store.MessageRoleAssistant, "recent assistant", "")
	_, err = repos.ToolCalls.Create(ctx, store.CreateToolCallInput{
		SessionID:     session.ID,
		ToolName:      "get_video_detail",
		ArgumentsJSON: `{"video_id":7}`,
		ResultSummary: "video 7: hot by author 3",
		Status:        store.ToolCallStatusSuccess,
	})
	if err != nil {
		t.Fatalf("create tool call: %v", err)
	}

	builder := NewBuilder(repos)
	built, err := builder.Build(ctx, BuildRequest{
		SessionID:        session.ID,
		LatestUserInput:  "当前继续分析视频 7",
		RequiredEvidence: []string{"get_video_detail", "get_hot_videos", "get_video_comments"},
	})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if built.Policy.MaxRecentMessages != 2 {
		t.Fatalf("policy max recent messages = %d, want 2", built.Policy.MaxRecentMessages)
	}
	if len(built.Messages) != 4 {
		t.Fatalf("messages len = %d, want system + 2 recent + latest", len(built.Messages))
	}
	system := built.Messages[0]
	if system.Role != llm.RoleSystem {
		t.Fatalf("first role = %q, want system", system.Role)
	}
	for _, want := range []string{"hot_rank_analysis", "get_video_detail", "get_hot_videos", "video 7: hot"} {
		if !strings.Contains(system.Content, want) {
			t.Fatalf("system prompt missing %q: %s", want, system.Content)
		}
	}
	if strings.Contains(joinMessageContent(built.Messages), "old user") || strings.Contains(joinMessageContent(built.Messages), "old assistant") {
		t.Fatalf("old messages should not be included: %+v", built.Messages)
	}
	if built.Messages[1].Content != "recent user summary" {
		t.Fatalf("recent user content = %q, want summary", built.Messages[1].Content)
	}
	if built.Messages[3].Role != llm.RoleUser || built.Messages[3].Content != "当前继续分析视频 7" {
		t.Fatalf("latest user message = %+v", built.Messages[3])
	}
}

func TestBuilderCompactsToolResultsAndRedactsSensitiveFields(t *testing.T) {
	ctx := context.Background()
	repos := newTestRepositories(t)

	session, err := repos.Sessions.Create(ctx, store.CreateSessionInput{
		UserID:            "operator-1",
		Title:             "comment risk",
		Scenario:          "comment_risk_analysis",
		Status:            store.SessionStatusActive,
		ContextPolicyJSON: `{"max_recent_messages":1,"max_tool_result_chars":220,"max_comments_for_llm":2,"max_comment_chars_each":12}`,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	_, err = repos.ToolCalls.Create(ctx, store.CreateToolCallInput{
		SessionID:     session.ID,
		ToolName:      "get_video_comments",
		ArgumentsJSON: `{"video_id":7}`,
		ResultSummary: "3 comments for video 7",
		ResultJSON: `{
			"api_key":"sk-test-secret",
			"stack_trace":"panic: internal file path",
			"comments":[
				{"id":1,"username":"u1","content":"这是第一条非常非常长的评论内容"},
				{"id":2,"username":"u2","content":"这是第二条也很长的评论内容"},
				{"id":3,"username":"u3","content":"third comment should be dropped"}
			]
		}`,
		Status: store.ToolCallStatusSuccess,
	})
	if err != nil {
		t.Fatalf("create tool call: %v", err)
	}

	builder := NewBuilder(repos)
	built, err := builder.Build(ctx, BuildRequest{
		SessionID:        session.ID,
		LatestUserInput:  "评论区风险如何",
		RequiredEvidence: []string{"get_video_detail", "get_video_comments", "analyze_comment_risk"},
	})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	content := joinMessageContent(built.Messages)
	if strings.Contains(content, "sk-test-secret") || strings.Contains(content, "panic: internal") {
		t.Fatalf("context leaked sensitive fields: %s", content)
	}
	if strings.Contains(content, "third comment should be dropped") {
		t.Fatalf("context did not compact comments: %s", content)
	}
	if !strings.Contains(content, "这是第一条非常非常长") || !strings.Contains(content, "truncated") {
		t.Fatalf("context missing compacted/truncation evidence: %s", content)
	}
	if !strings.Contains(content, "3 comments for video 7") {
		t.Fatalf("context missing tool result summary: %s", content)
	}
}

func TestCompactToolResultIsDeterministicForOversizedJSON(t *testing.T) {
	policy := ContextPolicy{
		MaxToolResultChars:  60,
		MaxCommentsForLLM:   2,
		MaxCommentCharsEach: 20,
	}

	result := CompactToolResult(`{"b":"`+strings.Repeat("x", 120)+`","a":"first"}`, policy)

	if len(result) > 80 {
		t.Fatalf("result len = %d, want compact truncation: %q", len(result), result)
	}
	if !strings.Contains(result, "truncated") {
		t.Fatalf("result = %q, want truncation marker", result)
	}
	if strings.Index(result, `"a"`) > strings.Index(result, `"b"`) {
		t.Fatalf("result should be deterministic with sorted JSON keys: %q", result)
	}
}

type testRepositories struct {
	Sessions  *store.SessionRepository
	Messages  *store.MessageRepository
	ToolCalls *store.ToolCallRepository
}

func newTestRepositories(t *testing.T) Repositories {
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
	return Repositories{
		Sessions:  store.NewSessionRepository(db),
		Messages:  store.NewMessageRepository(db),
		ToolCalls: store.NewToolCallRepository(db),
	}
}

func createMessage(t *testing.T, repo *store.MessageRepository, sessionID uint, role string, content string, summary string) {
	t.Helper()
	if _, err := repo.Create(context.Background(), store.CreateMessageInput{
		SessionID:      sessionID,
		Role:           role,
		Content:        content,
		ContentSummary: summary,
	}); err != nil {
		t.Fatalf("create message: %v", err)
	}
}

func joinMessageContent(messages []llm.Message) string {
	var builder strings.Builder
	for _, message := range messages {
		builder.WriteString(message.Content)
		builder.WriteString("\n")
	}
	return builder.String()
}
