package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAnalyzeCommentRiskDetectsDeterministicSignals(t *testing.T) {
	registry, err := NewRegistry(NewCommentRiskTool(2 * time.Second))
	if err != nil {
		t.Fatalf("NewRegistry returned error: %v", err)
	}
	executor := NewExecutor(registry, time.Second)

	args := json.RawMessage(`{
		"video_id": 99,
		"comments": [
			{"id": 1, "username": "u1", "content": "垃圾内容 @a @b @c @d"},
			{"id": 2, "username": "u2", "content": "垃圾内容 @a @b @c @d"},
			{"id": 3, "username": "u3", "content": "太差了，骗子"}
		]
	}`)

	result, err := executor.Execute(context.Background(), "analyze_comment_risk", args)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	report, ok := result.Data.(CommentRiskReport)
	if !ok {
		t.Fatalf("data type = %T, want CommentRiskReport", result.Data)
	}

	if report.VideoID != 99 {
		t.Fatalf("video id = %d, want 99", report.VideoID)
	}
	if report.RiskLevel != RiskLevelHigh {
		t.Fatalf("risk level = %q, want %q", report.RiskLevel, RiskLevelHigh)
	}
	if len(report.Findings) < 4 {
		t.Fatalf("findings = %+v, want repeated, sensitive, mention, negative signals", report.Findings)
	}
	if !strings.Contains(result.Summary, "high") || !strings.Contains(result.Summary, "4") {
		t.Fatalf("summary = %q, want high risk with finding count", result.Summary)
	}
}

func TestAnalyzeCommentRiskRejectsMissingComments(t *testing.T) {
	tool := NewCommentRiskTool(time.Second)

	_, err := tool.Execute(context.Background(), json.RawMessage(`{"video_id":99,"comments":[]}`))
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "comments") {
		t.Fatalf("error = %q, want comments context", err.Error())
	}
}
