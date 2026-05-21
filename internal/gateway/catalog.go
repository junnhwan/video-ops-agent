package gateway

import (
	"strings"

	"video-ops-agent/internal/agent/tools"
)

type ToolCatalogItem struct {
	Name        string           `json:"name"`
	DisplayName string           `json:"display_name"`
	Category    string           `json:"category"`
	Description string           `json:"description"`
	ReadOnly    bool             `json:"read_only"`
	Schema      tools.ToolSchema `json:"schema"`
}

func catalogItemFromSchema(schema tools.ToolSchema) ToolCatalogItem {
	name := schema.Function.Name
	return ToolCatalogItem{
		Name:        name,
		DisplayName: displayNameForTool(name),
		Category:    categoryForTool(name),
		Description: schema.Function.Description,
		ReadOnly:    true,
		Schema:      schema,
	}
}

func displayNameForTool(name string) string {
	switch name {
	case "get_video_detail":
		return "视频详情"
	case "get_hot_videos":
		return "热门视频"
	case "get_video_comments":
		return "视频评论"
	case "analyze_video_comment_risk":
		return "评论风险分析"
	case "analyze_comment_risk":
		return "评论文本风险"
	case "get_author_profile":
		return "作者画像"
	case "list_author_videos":
		return "作者视频"
	case "list_tag_videos":
		return "标签视频"
	default:
		return strings.ReplaceAll(name, "_", " ")
	}
}

func categoryForTool(name string) string {
	switch {
	case strings.Contains(name, "comment") || strings.Contains(name, "risk"):
		return "comment"
	case strings.Contains(name, "author"):
		return "author"
	case strings.Contains(name, "tag"):
		return "tag"
	case strings.Contains(name, "video") || strings.Contains(name, "hot"):
		return "video"
	default:
		return "general"
	}
}
