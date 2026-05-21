package guard

import "strings"

type Scenario string

const (
	ScenarioHotRankAnalysis       Scenario = "hot_rank_analysis"
	ScenarioCommentRiskAnalysis   Scenario = "comment_risk_analysis"
	ScenarioAuthorProfileAnalysis Scenario = "author_profile_analysis"
	ScenarioTagTrendAnalysis      Scenario = "tag_trend_analysis"
	ScenarioGeneral               Scenario = "general"
)

func DetectScenario(question string) Scenario {
	normalized := strings.ToLower(strings.TrimSpace(question))
	switch {
	case containsAny(normalized, "热榜", "hot rank", "hot_rank", "上榜"):
		return ScenarioHotRankAnalysis
	case containsAny(normalized, "评论区", "评论风险", "攻击", "刷屏", "争议", "comment risk"):
		return ScenarioCommentRiskAnalysis
	case containsAny(normalized, "作者", "扶持", "author", "creator"):
		return ScenarioAuthorProfileAnalysis
	case containsAny(normalized, "标签", "tag trend", "tag_trend") || strings.Contains(normalized, "#"):
		return ScenarioTagTrendAnalysis
	default:
		return ScenarioGeneral
	}
}

func RequiredTools(scenario Scenario) []string {
	switch scenario {
	case ScenarioHotRankAnalysis:
		return []string{"get_video_detail", "get_hot_videos", "get_video_comments"}
	case ScenarioCommentRiskAnalysis:
		return []string{"get_video_detail", "get_video_comments", "analyze_comment_risk"}
	case ScenarioAuthorProfileAnalysis:
		return []string{"get_author_profile", "list_author_videos"}
	case ScenarioTagTrendAnalysis:
		return []string{"list_tag_videos"}
	default:
		return nil
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
