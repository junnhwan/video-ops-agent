package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"video-ops-agent/internal/agent/skills"
	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/gateway"
	"video-ops-agent/internal/store"
)

func TestResourceAdapterReturnsToolsSkillsEvidenceAndTraceResources(t *testing.T) {
	ctx := context.Background()
	db := newMCPTestDB(t)
	registry, err := tools.NewRegistry(mcpTestTool{name: "get_video_detail"})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	invocations := store.NewGatewayInvocationRepository(db)
	sessionID := uint(7)
	if _, err := invocations.Create(ctx, store.CreateGatewayInvocationInput{
		Source:        gateway.InvocationSourceAgentRuntime,
		SessionID:     &sessionID,
		ToolName:      "get_video_detail",
		ArgumentsJSON: `{"video_id":101}`,
		ResultSummary: "video 101: test",
		Status:        store.ToolCallStatusSuccess,
	}); err != nil {
		t.Fatalf("create invocation: %v", err)
	}
	gatewayService := gateway.NewService(gateway.Dependencies{
		Registry:    registry,
		Executor:    tools.NewExecutor(registry, time.Second),
		Invocations: invocations,
	})
	skillService := skills.NewService(skills.Dependencies{
		Registry:   newMCPFullSkillRegistry(t),
		Repository: store.NewSkillRepository(db),
	})
	adapter := NewResourceAdapter(gatewayService, skillService)

	resources, err := adapter.ListResources(ctx)
	if err != nil {
		t.Fatalf("ListResources returned error: %v", err)
	}
	resourceURIs := make([]string, 0, len(resources))
	for _, resource := range resources {
		resourceURIs = append(resourceURIs, resource.URI)
	}
	for _, want := range []string{"videoops://tools", "videoops://skills", "videoops://evidence-rules"} {
		if !containsString(resourceURIs, want) {
			t.Fatalf("resources = %+v, missing %s", resources, want)
		}
	}

	toolsJSON, err := adapter.ReadResource(ctx, "videoops://tools")
	if err != nil {
		t.Fatalf("read tools: %v", err)
	}
	if !strings.Contains(string(toolsJSON), "get_video_detail") {
		t.Fatalf("tools resource = %s", string(toolsJSON))
	}
	skillsJSON, err := adapter.ReadResource(ctx, "videoops://skills")
	if err != nil {
		t.Fatalf("read skills: %v", err)
	}
	if !strings.Contains(string(skillsJSON), "comment_risk_analysis") {
		t.Fatalf("skills resource = %s", string(skillsJSON))
	}
	evidenceJSON, err := adapter.ReadResource(ctx, "videoops://evidence-rules")
	if err != nil {
		t.Fatalf("read evidence rules: %v", err)
	}
	if !strings.Contains(string(evidenceJSON), "required_evidence") {
		t.Fatalf("evidence resource = %s", string(evidenceJSON))
	}
	traceJSON, err := adapter.ReadResource(ctx, "videoops://sessions/7/trace")
	if err != nil {
		t.Fatalf("read trace: %v", err)
	}
	if !strings.Contains(string(traceJSON), "video 101: test") {
		t.Fatalf("trace resource = %s", string(traceJSON))
	}
}

func TestPromptAdapterExposesDiagnosisSkillsAsPrompts(t *testing.T) {
	ctx := context.Background()
	service := skills.NewService(skills.Dependencies{Registry: newMCPFullSkillRegistry(t)})
	adapter := NewPromptAdapter(service)

	prompts, err := adapter.ListPrompts(ctx)
	if err != nil {
		t.Fatalf("ListPrompts returned error: %v", err)
	}
	var found bool
	for _, prompt := range prompts {
		if prompt.Name == "comment_risk_analysis" {
			found = true
		}
	}
	if !found {
		t.Fatalf("prompts missing comment_risk_analysis: %+v", prompts)
	}

	prompt, err := adapter.GetPrompt(ctx, "comment_risk_analysis")
	if err != nil {
		t.Fatalf("GetPrompt returned error: %v", err)
	}
	encoded, err := json.Marshal(prompt)
	if err != nil {
		t.Fatalf("marshal prompt: %v", err)
	}
	if !strings.Contains(string(encoded), "Active diagnosis skill") ||
		!strings.Contains(string(encoded), "analyze_video_comment_risk") {
		t.Fatalf("prompt = %s", string(encoded))
	}
}

func newMCPFullSkillRegistry(t *testing.T) *tools.Registry {
	t.Helper()
	registry, err := tools.NewRegistry(
		mcpTestTool{name: "get_video_detail"},
		mcpTestTool{name: "get_hot_videos"},
		mcpTestTool{name: "get_video_comments"},
		mcpTestTool{name: "analyze_video_comment_risk"},
		mcpTestTool{name: "analyze_comment_risk"},
		mcpTestTool{name: "get_author_profile"},
		mcpTestTool{name: "list_author_videos"},
		mcpTestTool{name: "list_tag_videos"},
	)
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	return registry
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
