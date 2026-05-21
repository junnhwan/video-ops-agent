package contextbuilder

import (
	"context"
	"fmt"
	"strings"

	"video-ops-agent/internal/agent/llm"
	"video-ops-agent/internal/store"
)

type Repositories struct {
	Sessions  *store.SessionRepository
	Messages  *store.MessageRepository
	ToolCalls *store.ToolCallRepository
}

type Builder struct {
	repos Repositories
}

func NewBuilder(repos Repositories) *Builder {
	return &Builder{repos: repos}
}

func (b *Builder) Build(ctx context.Context, request BuildRequest) (BuiltContext, error) {
	if request.SessionID == 0 {
		return BuiltContext{}, fmt.Errorf("session_id is required")
	}
	if err := b.validate(); err != nil {
		return BuiltContext{}, err
	}

	session, err := b.repos.Sessions.Get(ctx, request.SessionID)
	if err != nil {
		return BuiltContext{}, err
	}
	policy, err := ParsePolicy(session.ContextPolicyJSON)
	if err != nil {
		return BuiltContext{}, err
	}

	recentMessages, err := b.repos.Messages.ListRecentBySession(ctx, request.SessionID, policy.MaxRecentMessages)
	if err != nil {
		return BuiltContext{}, err
	}
	toolCalls, err := b.repos.ToolCalls.ListRecentBySession(ctx, request.SessionID, 20)
	if err != nil {
		return BuiltContext{}, err
	}

	messages := make([]llm.Message, 0, 1+len(recentMessages)+1)
	messages = append(messages, llm.Message{
		Role:    llm.RoleSystem,
		Content: buildSystemPrompt(session, request.RequiredEvidence, toolCalls, policy),
	})
	for _, message := range recentMessages {
		content := message.Content
		if strings.TrimSpace(message.ContentSummary) != "" {
			content = message.ContentSummary
		}
		messages = append(messages, llm.Message{
			Role:    normalizeMessageRole(message.Role),
			Content: content,
		})
	}
	if strings.TrimSpace(request.LatestUserInput) != "" {
		messages = append(messages, llm.Message{
			Role:    llm.RoleUser,
			Content: request.LatestUserInput,
		})
	}

	return BuiltContext{
		Messages:         messages,
		Policy:           policy,
		RequiredEvidence: request.RequiredEvidence,
	}, nil
}

func (b *Builder) validate() error {
	if b.repos.Sessions == nil {
		return fmt.Errorf("session repository is required")
	}
	if b.repos.Messages == nil {
		return fmt.Errorf("message repository is required")
	}
	if b.repos.ToolCalls == nil {
		return fmt.Errorf("tool call repository is required")
	}
	return nil
}

func buildSystemPrompt(session store.AgentSession, requiredEvidence []string, toolCalls []store.AgentToolCall, policy ContextPolicy) string {
	var builder strings.Builder
	builder.WriteString("You are VideoOps Agent, a short-video content operations diagnosis agent.\n")
	builder.WriteString("Use platform tool evidence only; do not invent metrics or claim unsupported facts.\n")
	builder.WriteString("Keep answers grounded in the provided session memory and tool evidence summaries.\n")
	if strings.TrimSpace(session.Scenario) != "" {
		builder.WriteString("Active scenario: ")
		builder.WriteString(session.Scenario)
		builder.WriteString("\n")
	}
	if len(requiredEvidence) > 0 {
		builder.WriteString("Required evidence:\n")
		for _, evidence := range requiredEvidence {
			if strings.TrimSpace(evidence) == "" {
				continue
			}
			builder.WriteString("- ")
			builder.WriteString(strings.TrimSpace(evidence))
			builder.WriteString("\n")
		}
	}
	if len(toolCalls) > 0 {
		builder.WriteString("Previous tool evidence summaries:\n")
		for _, call := range toolCalls {
			builder.WriteString("- ")
			builder.WriteString(call.ToolName)
			builder.WriteString(" [")
			builder.WriteString(call.Status)
			builder.WriteString("]: ")
			builder.WriteString(toolCallSummary(call, policy))
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func toolCallSummary(call store.AgentToolCall, policy ContextPolicy) string {
	if strings.TrimSpace(call.ResultSummary) != "" {
		return truncateText(call.ResultSummary, policy.MaxToolResultChars)
	}
	if strings.TrimSpace(call.ResultJSON) != "" {
		return CompactToolResult(call.ResultJSON, policy)
	}
	if strings.TrimSpace(call.ErrorMessage) != "" {
		return truncateText(call.ErrorMessage, policy.MaxToolResultChars)
	}
	return "no result summary"
}

func normalizeMessageRole(role string) string {
	switch role {
	case llm.RoleSystem, llm.RoleUser, llm.RoleAssistant, llm.RoleTool:
		return role
	default:
		return llm.RoleUser
	}
}
