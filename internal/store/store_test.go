package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenSQLiteAutoMigrateAndRepositories(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("close db: %v", err)
		}
	}()
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	sessionRepo := NewSessionRepository(db)
	messageRepo := NewMessageRepository(db)
	toolCallRepo := NewToolCallRepository(db)

	session, err := sessionRepo.Create(ctx, CreateSessionInput{
		UserID:            "operator-1",
		Title:             "hot rank check",
		Scenario:          "hot_rank_analysis",
		Status:            SessionStatusActive,
		ContextPolicyJSON: `{"max_recent_messages":6}`,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.ID == 0 {
		t.Fatalf("expected session id to be assigned")
	}

	userMessage, err := messageRepo.Create(ctx, CreateMessageInput{
		SessionID:      session.ID,
		Role:           MessageRoleUser,
		Content:        "分析视频 7 为什么上热榜",
		ContentSummary: "hot rank request",
	})
	if err != nil {
		t.Fatalf("create message: %v", err)
	}
	if userMessage.ID == 0 {
		t.Fatalf("expected message id to be assigned")
	}

	toolCall, err := toolCallRepo.Create(ctx, CreateToolCallInput{
		SessionID:     session.ID,
		MessageID:     &userMessage.ID,
		ToolName:      "get_video_detail",
		ArgumentsJSON: `{"video_id":7}`,
		ResultJSON:    `{"id":7,"title":"hot"}`,
		ResultSummary: "video 7: hot",
		LatencyMS:     12,
		Status:        ToolCallStatusSuccess,
	})
	if err != nil {
		t.Fatalf("create tool call: %v", err)
	}
	if toolCall.ID == 0 {
		t.Fatalf("expected tool call id to be assigned")
	}

	gotSession, err := sessionRepo.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if gotSession.UserID != "operator-1" || gotSession.ContextPolicyJSON == "" {
		t.Fatalf("session = %+v", gotSession)
	}

	messages, err := messageRepo.ListBySession(ctx, session.ID, 10)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 1 || messages[0].ContentSummary != "hot rank request" {
		t.Fatalf("messages = %+v", messages)
	}

	toolCalls, err := toolCallRepo.ListBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("list tool calls: %v", err)
	}
	if len(toolCalls) != 1 || toolCalls[0].ResultSummary != "video 7: hot" || toolCalls[0].LatencyMS != 12 {
		t.Fatalf("tool calls = %+v", toolCalls)
	}
}

func TestRepositoriesValidateRequiredFields(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("close db: %v", err)
		}
	}()
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	if _, err := NewSessionRepository(db).Create(ctx, CreateSessionInput{}); err == nil {
		t.Fatalf("expected session validation error")
	}
	if _, err := NewMessageRepository(db).Create(ctx, CreateMessageInput{SessionID: 1, Role: MessageRoleUser}); err == nil {
		t.Fatalf("expected message validation error")
	}
	if _, err := NewToolCallRepository(db).Create(ctx, CreateToolCallInput{SessionID: 1}); err == nil {
		t.Fatalf("expected tool call validation error")
	}
}

func TestMessageRepositoryListsRecentMessagesInChronologicalOrder(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("close db: %v", err)
		}
	}()
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	session, err := NewSessionRepository(db).Create(ctx, CreateSessionInput{UserID: "u1", Status: SessionStatusActive})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	repo := NewMessageRepository(db)
	for _, content := range []string{"m1", "m2", "m3", "m4"} {
		if _, err := repo.Create(ctx, CreateMessageInput{SessionID: session.ID, Role: MessageRoleUser, Content: content}); err != nil {
			t.Fatalf("create message %q: %v", content, err)
		}
	}

	messages, err := repo.ListRecentBySession(ctx, session.ID, 2)
	if err != nil {
		t.Fatalf("ListRecentBySession returned error: %v", err)
	}

	if len(messages) != 2 || messages[0].Content != "m3" || messages[1].Content != "m4" {
		t.Fatalf("messages = %+v, want m3 then m4", messages)
	}
}

func TestToolCallRepositoryListsRecentCallsInChronologicalOrder(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "agent.db"))
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("close db: %v", err)
		}
	}()
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	session, err := NewSessionRepository(db).Create(ctx, CreateSessionInput{UserID: "u1", Status: SessionStatusActive})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	repo := NewToolCallRepository(db)
	for _, toolName := range []string{"t1", "t2", "t3"} {
		if _, err := repo.Create(ctx, CreateToolCallInput{
			SessionID:     session.ID,
			ToolName:      toolName,
			ArgumentsJSON: `{}`,
			Status:        ToolCallStatusSuccess,
		}); err != nil {
			t.Fatalf("create tool call %q: %v", toolName, err)
		}
	}

	toolCalls, err := repo.ListRecentBySession(ctx, session.ID, 2)
	if err != nil {
		t.Fatalf("ListRecentBySession returned error: %v", err)
	}

	if len(toolCalls) != 2 || toolCalls[0].ToolName != "t2" || toolCalls[1].ToolName != "t3" {
		t.Fatalf("toolCalls = %+v, want t2 then t3", toolCalls)
	}
}
