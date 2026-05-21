package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"video-ops-agent/internal/agent/skills"
	"video-ops-agent/internal/gateway"
	"video-ops-agent/internal/store"
)

type Resource struct {
	URI      string `json:"uri"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
}

type ResourceAdapter struct {
	gateway *gateway.Service
	skills  *skills.Service
}

func NewResourceAdapter(gatewayService *gateway.Service, skillService *skills.Service) *ResourceAdapter {
	return &ResourceAdapter{gateway: gatewayService, skills: skillService}
}

func (a *ResourceAdapter) ListResources(context.Context) ([]Resource, error) {
	return []Resource{
		{URI: "videoops://tools", Name: "VideoOps tool catalog", MimeType: "application/json"},
		{URI: "videoops://skills", Name: "VideoOps diagnosis skills", MimeType: "application/json"},
		{URI: "videoops://evidence-rules", Name: "VideoOps evidence rules", MimeType: "application/json"},
	}, nil
}

func (a *ResourceAdapter) ReadResource(ctx context.Context, uri string) ([]byte, error) {
	switch {
	case uri == "videoops://tools":
		if a.gateway == nil {
			return nil, fmt.Errorf("gateway service is required")
		}
		tools, err := a.gateway.ListTools(ctx)
		if err != nil {
			return nil, err
		}
		return marshalResource(map[string]any{"tools": tools})
	case uri == "videoops://skills":
		if a.skills == nil {
			return nil, fmt.Errorf("skill service is required")
		}
		skillList, err := a.skills.List(ctx)
		if err != nil {
			return nil, err
		}
		return marshalResource(map[string]any{"skills": skillList})
	case uri == "videoops://evidence-rules":
		if a.skills == nil {
			return nil, fmt.Errorf("skill service is required")
		}
		skillList, err := a.skills.List(ctx)
		if err != nil {
			return nil, err
		}
		rules := make([]map[string]any, 0, len(skillList))
		for _, skill := range skillList {
			rules = append(rules, map[string]any{
				"skill_id":          skill.ID,
				"scenario":          skill.Scenario,
				"allowed_tools":     skill.AllowedTools,
				"required_evidence": skill.RequiredEvidence,
			})
		}
		return marshalResource(map[string]any{"evidence_rules": rules})
	case strings.HasPrefix(uri, "videoops://sessions/") && strings.HasSuffix(uri, "/trace"):
		if a.gateway == nil {
			return nil, fmt.Errorf("gateway service is required")
		}
		sessionID, err := parseTraceSessionID(uri)
		if err != nil {
			return nil, err
		}
		invocations, err := a.gateway.ListInvocations(ctx, store.GatewayInvocationFilter{SessionID: &sessionID, Limit: 200})
		if err != nil {
			return nil, err
		}
		return marshalResource(map[string]any{"session_id": sessionID, "invocations": invocations})
	default:
		return nil, fmt.Errorf("unknown resource uri %q", uri)
	}
}

func parseTraceSessionID(uri string) (uint, error) {
	trimmed := strings.TrimPrefix(uri, "videoops://sessions/")
	trimmed = strings.TrimSuffix(trimmed, "/trace")
	value, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("invalid trace session id in %q", uri)
	}
	return uint(value), nil
}

func marshalResource(value any) ([]byte, error) {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal resource: %w", err)
	}
	return encoded, nil
}
