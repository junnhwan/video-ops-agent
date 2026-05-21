package guard

import (
	"strings"
	"testing"
)

func TestDetectScenarioFromUserQuestion(t *testing.T) {
	cases := []struct {
		question string
		want     Scenario
	}{
		{question: "分析视频 123 为什么上热榜", want: ScenarioHotRankAnalysis},
		{question: "评论区有没有攻击和刷屏风险", want: ScenarioCommentRiskAnalysis},
		{question: "作者 8 值得扶持吗", want: ScenarioAuthorProfileAnalysis},
		{question: "#Go 后端 这个标签最近表现怎么样", want: ScenarioTagTrendAnalysis},
		{question: "你好", want: ScenarioGeneral},
	}

	for _, tc := range cases {
		t.Run(tc.question, func(t *testing.T) {
			if got := DetectScenario(tc.question); got != tc.want {
				t.Fatalf("DetectScenario(%q) = %q, want %q", tc.question, got, tc.want)
			}
		})
	}
}

func TestRequiredToolsByScenario(t *testing.T) {
	if got := RequiredTools(ScenarioHotRankAnalysis); strings.Join(got, ",") != "get_video_detail,get_hot_videos,get_video_comments" {
		t.Fatalf("hot rank required tools = %+v", got)
	}
	if got := RequiredTools(ScenarioCommentRiskAnalysis); strings.Join(got, ",") != "get_video_detail,analyze_video_comment_risk" {
		t.Fatalf("comment risk required tools = %+v", got)
	}
	if got := RequiredTools(ScenarioAuthorProfileAnalysis); strings.Join(got, ",") != "get_author_profile,list_author_videos" {
		t.Fatalf("author required tools = %+v", got)
	}
	if got := RequiredTools(ScenarioTagTrendAnalysis); strings.Join(got, ",") != "list_tag_videos" {
		t.Fatalf("tag required tools = %+v", got)
	}
	if got := RequiredTools(ScenarioGeneral); len(got) != 0 {
		t.Fatalf("general required tools = %+v, want empty", got)
	}
}

func TestCheckRequiredEvidenceReportsMissingTools(t *testing.T) {
	check := CheckRequired(
		[]string{"get_video_detail", "get_hot_videos", "get_video_comments"},
		[]string{"get_video_detail"},
	)

	if check.Complete {
		t.Fatalf("expected incomplete evidence")
	}
	if strings.Join(check.MissingTools, ",") != "get_hot_videos,get_video_comments" {
		t.Fatalf("missing tools = %+v", check.MissingTools)
	}

	instruction := RetryInstruction(check.MissingTools)
	if !strings.Contains(instruction, "Evidence is incomplete") || !strings.Contains(instruction, "get_hot_videos") {
		t.Fatalf("instruction = %q", instruction)
	}
}
