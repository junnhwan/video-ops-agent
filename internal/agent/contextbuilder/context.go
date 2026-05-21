package contextbuilder

import (
	"encoding/json"
	"fmt"

	"video-ops-agent/internal/agent/llm"
)

const (
	defaultMaxRecentMessages  = 6
	defaultMaxToolResultChars = 4000
	defaultMaxCommentsForLLM  = 50
	defaultMaxCommentChars    = 300
)

type ContextPolicy struct {
	MaxRecentMessages   int `json:"max_recent_messages"`
	MaxToolResultChars  int `json:"max_tool_result_chars"`
	MaxCommentsForLLM   int `json:"max_comments_for_llm"`
	MaxCommentCharsEach int `json:"max_comment_chars_each"`
}

type BuildRequest struct {
	SessionID        uint
	LatestUserInput  string
	RequiredEvidence []string
	SkillPrompt      string
}

type BuiltContext struct {
	Messages         []llm.Message
	Policy           ContextPolicy
	RequiredEvidence []string
}

func DefaultPolicy() ContextPolicy {
	return ContextPolicy{
		MaxRecentMessages:   defaultMaxRecentMessages,
		MaxToolResultChars:  defaultMaxToolResultChars,
		MaxCommentsForLLM:   defaultMaxCommentsForLLM,
		MaxCommentCharsEach: defaultMaxCommentChars,
	}
}

func ParsePolicy(raw string) (ContextPolicy, error) {
	policy := DefaultPolicy()
	if raw == "" {
		return policy, nil
	}
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return ContextPolicy{}, fmt.Errorf("parse context policy json: %w", err)
	}
	return normalizePolicy(policy), nil
}

func normalizePolicy(policy ContextPolicy) ContextPolicy {
	defaults := DefaultPolicy()
	if policy.MaxRecentMessages <= 0 {
		policy.MaxRecentMessages = defaults.MaxRecentMessages
	}
	if policy.MaxToolResultChars <= 0 {
		policy.MaxToolResultChars = defaults.MaxToolResultChars
	}
	if policy.MaxCommentsForLLM <= 0 {
		policy.MaxCommentsForLLM = defaults.MaxCommentsForLLM
	}
	if policy.MaxCommentCharsEach <= 0 {
		policy.MaxCommentCharsEach = defaults.MaxCommentCharsEach
	}
	return policy
}
