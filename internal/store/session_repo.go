package store

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

type CreateSessionInput struct {
	UserID            string
	Title             string
	Scenario          string
	Status            string
	ContextPolicyJSON string
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, input CreateSessionInput) (AgentSession, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return AgentSession{}, fmt.Errorf("user_id is required")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = SessionStatusActive
	}

	session := AgentSession{
		UserID:            strings.TrimSpace(input.UserID),
		Title:             strings.TrimSpace(input.Title),
		Scenario:          strings.TrimSpace(input.Scenario),
		Status:            status,
		ContextPolicyJSON: input.ContextPolicyJSON,
	}
	if err := r.db.WithContext(ctx).Create(&session).Error; err != nil {
		return AgentSession{}, fmt.Errorf("create agent session: %w", err)
	}
	return session, nil
}

func (r *SessionRepository) Get(ctx context.Context, id uint) (AgentSession, error) {
	if id == 0 {
		return AgentSession{}, fmt.Errorf("session id is required")
	}
	var session AgentSession
	if err := r.db.WithContext(ctx).First(&session, id).Error; err != nil {
		return AgentSession{}, fmt.Errorf("get agent session %d: %w", id, err)
	}
	return session, nil
}

func (r *SessionRepository) List(ctx context.Context, userID string, limit int) ([]AgentSession, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Order("updated_at DESC").Limit(limit)
	if strings.TrimSpace(userID) != "" {
		query = query.Where("user_id = ?", strings.TrimSpace(userID))
	}

	var sessions []AgentSession
	if err := query.Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("list agent sessions: %w", err)
	}
	return sessions, nil
}
