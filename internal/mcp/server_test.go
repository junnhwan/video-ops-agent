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

func TestServerHandlesToolsListAndCall(t *testing.T) {
	ctx := context.Background()
	db := newMCPTestDB(t)
	registry, err := tools.NewRegistry(mcpTestTool{
		name: "get_video_detail",
		result: tools.ToolResult{
			ToolName: "get_video_detail",
			Summary:  "video 101: test",
			Data:     map[string]any{"id": float64(101)},
		},
	})
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	gatewayService := gateway.NewService(gateway.Dependencies{
		Registry:    registry,
		Executor:    tools.NewExecutor(registry, time.Second),
		Invocations: store.NewGatewayInvocationRepository(db),
	})
	skillService := skills.NewService(skills.Dependencies{Registry: newMCPFullSkillRegistry(t)})
	server := NewServer(NewToolAdapter(gatewayService), NewResourceAdapter(gatewayService, skillService), NewPromptAdapter(skillService))

	listResp := server.Handle(ctx, JSONRPCRequest{JSONRPC: "2.0", ID: float64(1), Method: "tools/list"})
	if listResp.Error != nil {
		t.Fatalf("tools/list error = %+v", listResp.Error)
	}
	listEncoded, err := json.Marshal(listResp.Result)
	if err != nil {
		t.Fatalf("marshal list result: %v", err)
	}
	if !json.Valid(listEncoded) || !containsJSONText(listEncoded, "get_video_detail") {
		t.Fatalf("tools/list result = %s", string(listEncoded))
	}

	callParams := json.RawMessage(`{"name":"get_video_detail","arguments":{"video_id":101}}`)
	callResp := server.Handle(ctx, JSONRPCRequest{JSONRPC: "2.0", ID: float64(2), Method: "tools/call", Params: callParams})
	if callResp.Error != nil {
		t.Fatalf("tools/call error = %+v", callResp.Error)
	}
	callEncoded, err := json.Marshal(callResp.Result)
	if err != nil {
		t.Fatalf("marshal call result: %v", err)
	}
	if !containsJSONText(callEncoded, "video 101: test") {
		t.Fatalf("tools/call result = %s", string(callEncoded))
	}
}

func containsJSONText(raw []byte, want string) bool {
	return strings.Contains(string(raw), want)
}
