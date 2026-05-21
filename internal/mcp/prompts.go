package mcp

import (
	"context"
	"fmt"
	"strings"

	"video-ops-agent/internal/agent/skills"
)

type Prompt struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PromptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PromptResult struct {
	Description string          `json:"description"`
	Messages    []PromptMessage `json:"messages"`
}

type PromptAdapter struct {
	skills *skills.Service
}

func NewPromptAdapter(skillService *skills.Service) *PromptAdapter {
	return &PromptAdapter{skills: skillService}
}

func (a *PromptAdapter) ListPrompts(ctx context.Context) ([]Prompt, error) {
	if a.skills == nil {
		return nil, fmt.Errorf("skill service is required")
	}
	skillList, err := a.skills.List(ctx)
	if err != nil {
		return nil, err
	}
	prompts := make([]Prompt, 0, len(skillList))
	for _, skill := range skillList {
		prompts = append(prompts, Prompt{Name: skill.ID, Description: skill.Description})
	}
	return prompts, nil
}

func (a *PromptAdapter) GetPrompt(ctx context.Context, name string) (PromptResult, error) {
	if a.skills == nil {
		return PromptResult{}, fmt.Errorf("skill service is required")
	}
	skill, err := a.skills.Get(ctx, name)
	if err != nil {
		return PromptResult{}, err
	}
	var content strings.Builder
	content.WriteString(skills.RenderPrompt(skill))
	content.WriteString("Allowed tools:\n")
	for _, toolName := range skill.AllowedTools {
		content.WriteString("- ")
		content.WriteString(toolName)
		content.WriteString("\n")
	}
	content.WriteString("Required evidence:\n")
	for _, toolName := range skill.RequiredEvidence {
		content.WriteString("- ")
		content.WriteString(toolName)
		content.WriteString("\n")
	}
	return PromptResult{
		Description: skill.Description,
		Messages: []PromptMessage{{
			Role:    "user",
			Content: content.String(),
		}},
	}, nil
}
