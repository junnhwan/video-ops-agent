package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"video-ops-agent/internal/platform/videofeed"
)

const (
	RiskLevelLow    = "low"
	RiskLevelMedium = "medium"
	RiskLevelHigh   = "high"
)

type CommentRiskReport struct {
	VideoID     uint                 `json:"video_id"`
	RiskLevel   string               `json:"risk_level"`
	Total       int                  `json:"total"`
	Findings    []CommentRiskFinding `json:"findings"`
	RiskSummary string               `json:"risk_summary"`
}

type VideoCommentRiskResult struct {
	VideoID  uint                `json:"video_id"`
	Comments []videofeed.Comment `json:"comments"`
	Report   CommentRiskReport   `json:"report"`
}

type CommentRiskFinding struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	CommentID  uint   `json:"comment_id,omitempty"`
	Username   string `json:"username,omitempty"`
	Content    string `json:"content,omitempty"`
	Count      int    `json:"count,omitempty"`
	Matched    string `json:"matched,omitempty"`
	Suggestion string `json:"suggestion"`
}

type commentRiskComment struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Content  string `json:"content"`
}

type commentRiskTool struct {
	timeout time.Duration
}

type videoCommentRiskTool struct {
	client  PlatformClient
	timeout time.Duration
}

func NewCommentRiskTool(timeout time.Duration) Tool {
	return commentRiskTool{timeout: timeout}
}

func NewVideoCommentRiskTool(client PlatformClient, timeout time.Duration) Tool {
	return videoCommentRiskTool{client: client, timeout: timeout}
}

func (t commentRiskTool) Name() string {
	return "analyze_comment_risk"
}

func (t commentRiskTool) Schema() ToolSchema {
	return NewFunctionSchema(
		t.Name(),
		"Analyze comment risk with deterministic rules: repeated content, sensitive words, excessive mentions, and negative keywords.",
		objectSchema(map[string]any{
			"video_id": integerSchema("Video ID."),
			"comments": map[string]any{
				"type":        "array",
				"description": "Comments to scan.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":       integerSchema("Comment ID."),
						"username": stringSchema("Comment author username."),
						"content":  stringSchema("Comment content."),
					},
					"required":             []string{"content"},
					"additionalProperties": true,
				},
			},
		}, []string{"video_id", "comments"}),
	)
}

func (t commentRiskTool) Timeout() time.Duration {
	return t.timeout
}

func (t commentRiskTool) Execute(_ context.Context, arguments json.RawMessage) (ToolResult, error) {
	var args struct {
		VideoID  uint                 `json:"video_id"`
		Comments []commentRiskComment `json:"comments"`
	}
	if err := decodeArguments(arguments, &args); err != nil {
		return ToolResult{}, err
	}
	if args.VideoID == 0 {
		return ToolResult{}, fmt.Errorf("video_id must be greater than 0")
	}
	if len(args.Comments) == 0 {
		return ToolResult{}, fmt.Errorf("comments must not be empty")
	}

	report := analyzeCommentRisk(args.VideoID, args.Comments)
	return ToolResult{
		ToolName: t.Name(),
		Data:     report,
		Summary:  fmt.Sprintf("%s comment risk for video %d with %d findings across %d comments", report.RiskLevel, report.VideoID, len(report.Findings), report.Total),
	}, nil
}

func analyzeCommentRisk(videoID uint, comments []commentRiskComment) CommentRiskReport {
	findings := make([]CommentRiskFinding, 0)
	findings = append(findings, repeatedContentFindings(comments)...)

	var sensitiveFinding *CommentRiskFinding
	var mentionFinding *CommentRiskFinding
	var negativeFinding *CommentRiskFinding

	for _, comment := range comments {
		trimmed := strings.TrimSpace(comment.Content)
		if trimmed == "" {
			continue
		}
		for _, word := range sensitiveWords {
			if strings.Contains(trimmed, word) {
				if sensitiveFinding == nil {
					sensitiveFinding = &CommentRiskFinding{
						Type:       "sensitive_word",
						Severity:   RiskLevelHigh,
						CommentID:  comment.ID,
						Username:   comment.Username,
						Content:    trimmed,
						Matched:    word,
						Suggestion: "人工复核敏感词命中内容，必要时隐藏评论或提醒作者。",
					}
				}
				sensitiveFinding.Count++
				break
			}
		}
		mentions := mentionCount(trimmed)
		if mentions >= 4 {
			if mentionFinding == nil {
				mentionFinding = &CommentRiskFinding{
					Type:       "excessive_mentions",
					Severity:   RiskLevelMedium,
					CommentID:  comment.ID,
					Username:   comment.Username,
					Content:    trimmed,
					Suggestion: "检查是否存在拉踩、引战或刷屏式召唤用户。",
				}
			}
			mentionFinding.Count++
		}
		for _, word := range negativeWords {
			if strings.Contains(trimmed, word) {
				if negativeFinding == nil {
					negativeFinding = &CommentRiskFinding{
						Type:       "negative_keyword",
						Severity:   RiskLevelMedium,
						CommentID:  comment.ID,
						Username:   comment.Username,
						Content:    trimmed,
						Matched:    word,
						Suggestion: "观察负面反馈是否集中，必要时补充说明或运营澄清。",
					}
				}
				negativeFinding.Count++
				break
			}
		}
	}
	if sensitiveFinding != nil {
		findings = append(findings, *sensitiveFinding)
	}
	if mentionFinding != nil {
		findings = append(findings, *mentionFinding)
	}
	if negativeFinding != nil {
		findings = append(findings, *negativeFinding)
	}

	riskLevel := RiskLevelLow
	for _, finding := range findings {
		if finding.Severity == RiskLevelHigh {
			riskLevel = RiskLevelHigh
			break
		}
		if finding.Severity == RiskLevelMedium {
			riskLevel = RiskLevelMedium
		}
	}

	return CommentRiskReport{
		VideoID:     videoID,
		RiskLevel:   riskLevel,
		Total:       len(comments),
		Findings:    findings,
		RiskSummary: summarizeRisk(riskLevel, len(findings), len(comments)),
	}
}

func (t videoCommentRiskTool) Name() string {
	return "analyze_video_comment_risk"
}

func (t videoCommentRiskTool) Schema() ToolSchema {
	return NewFunctionSchema(
		t.Name(),
		"Fetch comments for one video from video-feed and analyze comment risk with deterministic rules. Prefer this tool over manually passing comments to analyze_comment_risk.",
		objectSchema(map[string]any{
			"video_id": integerSchema("Video ID."),
			"limit":    integerSchema("Maximum comments to fetch and scan."),
		}, []string{"video_id"}),
	)
}

func (t videoCommentRiskTool) Timeout() time.Duration {
	return t.timeout
}

func (t videoCommentRiskTool) Execute(ctx context.Context, arguments json.RawMessage) (ToolResult, error) {
	var args struct {
		VideoID uint `json:"video_id"`
		Limit   int  `json:"limit"`
	}
	if err := decodeArguments(arguments, &args); err != nil {
		return ToolResult{}, err
	}
	if args.VideoID == 0 {
		return ToolResult{}, fmt.Errorf("video_id must be greater than 0")
	}
	limit := normalizeLimit(args.Limit)
	comments, err := t.client.GetVideoComments(ctx, args.VideoID, limit)
	if err != nil {
		return ToolResult{}, err
	}

	riskComments := make([]commentRiskComment, 0, len(comments))
	for _, comment := range comments {
		riskComments = append(riskComments, commentRiskComment{
			ID:       comment.ID,
			Username: comment.Username,
			Content:  comment.Content,
		})
	}
	report := analyzeCommentRisk(args.VideoID, riskComments)
	result := VideoCommentRiskResult{
		VideoID:  args.VideoID,
		Comments: comments,
		Report:   report,
	}
	return ToolResult{
		ToolName: t.Name(),
		Data:     result,
		Summary:  fmt.Sprintf("%s comment risk for video %d with %d findings across %d comments", report.RiskLevel, report.VideoID, len(report.Findings), report.Total),
	}, nil
}

func repeatedContentFindings(comments []commentRiskComment) []CommentRiskFinding {
	byContent := make(map[string][]commentRiskComment)
	for _, comment := range comments {
		normalized := strings.ToLower(strings.TrimSpace(comment.Content))
		if normalized == "" {
			continue
		}
		byContent[normalized] = append(byContent[normalized], comment)
	}

	keys := make([]string, 0, len(byContent))
	for content := range byContent {
		keys = append(keys, content)
	}
	sort.Strings(keys)

	findings := make([]CommentRiskFinding, 0)
	for _, content := range keys {
		group := byContent[content]
		if len(group) < 2 {
			continue
		}
		first := group[0]
		findings = append(findings, CommentRiskFinding{
			Type:       "repeated_content",
			Severity:   RiskLevelMedium,
			CommentID:  first.ID,
			Username:   first.Username,
			Content:    strings.TrimSpace(first.Content),
			Count:      len(group),
			Suggestion: "检查是否存在复制粘贴式刷屏或带节奏评论。",
		})
	}
	return findings
}

func summarizeRisk(level string, findings int, total int) string {
	switch level {
	case RiskLevelHigh:
		return fmt.Sprintf("评论区存在高风险信号，%d 条规则命中覆盖 %d 条评论。", findings, total)
	case RiskLevelMedium:
		return fmt.Sprintf("评论区存在中等风险信号，%d 条规则命中覆盖 %d 条评论。", findings, total)
	default:
		return fmt.Sprintf("评论区暂未命中明显风险规则，共扫描 %d 条评论。", total)
	}
}

var mentionPattern = regexp.MustCompile(`@\S+`)

func mentionCount(content string) int {
	return len(mentionPattern.FindAllString(content, -1))
}

var sensitiveWords = []string{
	"垃圾",
	"骗子",
	"诈骗",
	"引流",
}

var negativeWords = []string{
	"太差",
	"失望",
	"恶心",
	"避雷",
	"举报",
}
