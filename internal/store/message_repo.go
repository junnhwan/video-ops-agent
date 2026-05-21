package store

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

type CreateMessageInput struct {
	SessionID      uint
	Role           string
	Content        string
	ContentSummary string
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, input CreateMessageInput) (AgentMessage, error) {
	if input.SessionID == 0 {
		return AgentMessage{}, fmt.Errorf("session_id is required")
	}
	if strings.TrimSpace(input.Role) == "" {
		return AgentMessage{}, fmt.Errorf("role is required")
	}
	if strings.TrimSpace(input.Content) == "" {
		return AgentMessage{}, fmt.Errorf("content is required")
	}

	message := AgentMessage{
		SessionID:      input.SessionID,
		Role:           strings.TrimSpace(input.Role),
		Content:        input.Content,
		ContentSummary: input.ContentSummary,
	}
	if err := r.db.WithContext(ctx).Create(&message).Error; err != nil {
		return AgentMessage{}, fmt.Errorf("create agent message: %w", err)
	}
	return message, nil
}

func (r *MessageRepository) ListBySession(ctx context.Context, sessionID uint, limit int) ([]AgentMessage, error) {
	if sessionID == 0 {
		return nil, fmt.Errorf("session_id is required")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	var messages []AgentMessage
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("id ASC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("list agent messages for session %d: %w", sessionID, err)
	}
	return messages, nil
}
