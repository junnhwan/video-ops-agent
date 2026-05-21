package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"video-ops-agent/internal/platform/videofeed"
)

const defaultPlatformToolTimeout = 2 * time.Second

type PlatformClient interface {
	GetVideoDetail(ctx context.Context, videoID uint) (videofeed.Video, error)
	GetHotVideos(ctx context.Context, limit int) (videofeed.HotVideosResponse, error)
	GetVideoComments(ctx context.Context, videoID uint, limit int) ([]videofeed.Comment, error)
	GetAuthorProfile(ctx context.Context, authorID uint) (videofeed.AuthorProfile, error)
	ListAuthorVideos(ctx context.Context, authorID uint, limit int) ([]videofeed.Video, error)
	ListTagVideos(ctx context.Context, tagName string, limit int) ([]videofeed.FeedVideoItem, error)
}

func NewDefaultRegistry(client PlatformClient) (*Registry, error) {
	return NewRegistry(DefaultTools(client)...)
}

func DefaultTools(client PlatformClient) []Tool {
	return []Tool{
		newPlatformTool(
			"get_video_detail",
			"Get one video detail from video-feed by video_id.",
			objectSchema(map[string]any{"video_id": integerSchema("Video ID.")}, []string{"video_id"}),
			defaultPlatformToolTimeout,
			func(ctx context.Context, raw json.RawMessage) (any, string, error) {
				var args struct {
					VideoID uint `json:"video_id"`
				}
				if err := decodeArguments(raw, &args); err != nil {
					return nil, "", err
				}
				if args.VideoID == 0 {
					return nil, "", fmt.Errorf("video_id must be greater than 0")
				}
				video, err := client.GetVideoDetail(ctx, args.VideoID)
				if err != nil {
					return nil, "", err
				}
				return video, fmt.Sprintf("video %d: %s by author %d, likes=%d popularity=%d", video.ID, video.Title, video.AuthorID, video.LikesCount, video.Popularity), nil
			},
		),
		newPlatformTool(
			"get_hot_videos",
			"List hot videos from video-feed popularity feed.",
			objectSchema(map[string]any{"limit": integerSchema("Maximum videos to return.")}, []string{}),
			defaultPlatformToolTimeout,
			func(ctx context.Context, raw json.RawMessage) (any, string, error) {
				var args struct {
					Limit int `json:"limit"`
				}
				if err := decodeArguments(raw, &args); err != nil {
					return nil, "", err
				}
				limit := normalizeLimit(args.Limit)
				resp, err := client.GetHotVideos(ctx, limit)
				if err != nil {
					return nil, "", err
				}
				return resp, fmt.Sprintf("%d hot videos, has_more=%t", len(resp.VideoList), resp.HasMore), nil
			},
		),
		newPlatformTool(
			"get_video_comments",
			"List comments for one video from video-feed.",
			objectSchema(map[string]any{
				"video_id": integerSchema("Video ID."),
				"limit":    integerSchema("Maximum comments to return."),
			}, []string{"video_id"}),
			defaultPlatformToolTimeout,
			func(ctx context.Context, raw json.RawMessage) (any, string, error) {
				var args struct {
					VideoID uint `json:"video_id"`
					Limit   int  `json:"limit"`
				}
				if err := decodeArguments(raw, &args); err != nil {
					return nil, "", err
				}
				if args.VideoID == 0 {
					return nil, "", fmt.Errorf("video_id must be greater than 0")
				}
				limit := normalizeLimit(args.Limit)
				comments, err := client.GetVideoComments(ctx, args.VideoID, limit)
				if err != nil {
					return nil, "", err
				}
				return comments, fmt.Sprintf("%d comments for video %d", len(comments), args.VideoID), nil
			},
		),
		newPlatformTool(
			"get_author_profile",
			"Get author profile and aggregate counters from video-feed.",
			objectSchema(map[string]any{"author_id": integerSchema("Author account ID.")}, []string{"author_id"}),
			defaultPlatformToolTimeout,
			func(ctx context.Context, raw json.RawMessage) (any, string, error) {
				var args struct {
					AuthorID uint `json:"author_id"`
				}
				if err := decodeArguments(raw, &args); err != nil {
					return nil, "", err
				}
				if args.AuthorID == 0 {
					return nil, "", fmt.Errorf("author_id must be greater than 0")
				}
				profile, err := client.GetAuthorProfile(ctx, args.AuthorID)
				if err != nil {
					return nil, "", err
				}
				return profile, fmt.Sprintf("author %d: %s, videos=%d total_likes=%d", profile.Account.ID, profile.Account.Username, profile.VideoCount, profile.TotalLikes), nil
			},
		),
		newPlatformTool(
			"list_author_videos",
			"List recent videos for one author from video-feed.",
			objectSchema(map[string]any{
				"author_id": integerSchema("Author account ID."),
				"limit":     integerSchema("Maximum videos to return."),
			}, []string{"author_id"}),
			defaultPlatformToolTimeout,
			func(ctx context.Context, raw json.RawMessage) (any, string, error) {
				var args struct {
					AuthorID uint `json:"author_id"`
					Limit    int  `json:"limit"`
				}
				if err := decodeArguments(raw, &args); err != nil {
					return nil, "", err
				}
				if args.AuthorID == 0 {
					return nil, "", fmt.Errorf("author_id must be greater than 0")
				}
				limit := normalizeLimit(args.Limit)
				videos, err := client.ListAuthorVideos(ctx, args.AuthorID, limit)
				if err != nil {
					return nil, "", err
				}
				return videos, fmt.Sprintf("%d author videos for author %d", len(videos), args.AuthorID), nil
			},
		),
		newPlatformTool(
			"list_tag_videos",
			"List videos under one tag from video-feed.",
			objectSchema(map[string]any{
				"tag_name": stringSchema("Tag name without #."),
				"limit":    integerSchema("Maximum videos to return."),
			}, []string{"tag_name"}),
			defaultPlatformToolTimeout,
			func(ctx context.Context, raw json.RawMessage) (any, string, error) {
				var args struct {
					TagName string `json:"tag_name"`
					Limit   int    `json:"limit"`
				}
				if err := decodeArguments(raw, &args); err != nil {
					return nil, "", err
				}
				tagName := strings.TrimSpace(args.TagName)
				if tagName == "" {
					return nil, "", fmt.Errorf("tag_name is required")
				}
				limit := normalizeLimit(args.Limit)
				videos, err := client.ListTagVideos(ctx, tagName, limit)
				if err != nil {
					return nil, "", err
				}
				return videos, fmt.Sprintf("%d tag videos for %q", len(videos), tagName), nil
			},
		),
		NewVideoCommentRiskTool(client, defaultPlatformToolTimeout),
		NewCommentRiskTool(defaultPlatformToolTimeout),
	}
}

type platformTool struct {
	name        string
	description string
	parameters  map[string]any
	timeout     time.Duration
	execute     func(context.Context, json.RawMessage) (any, string, error)
}

func newPlatformTool(name string, description string, parameters map[string]any, timeout time.Duration, execute func(context.Context, json.RawMessage) (any, string, error)) Tool {
	return platformTool{name: name, description: description, parameters: parameters, timeout: timeout, execute: execute}
}

func (t platformTool) Name() string {
	return t.name
}

func (t platformTool) Schema() ToolSchema {
	return NewFunctionSchema(t.name, t.description, t.parameters)
}

func (t platformTool) Timeout() time.Duration {
	return t.timeout
}

func (t platformTool) Execute(ctx context.Context, arguments json.RawMessage) (ToolResult, error) {
	data, summary, err := t.execute(ctx, arguments)
	if err != nil {
		return ToolResult{}, err
	}
	return ToolResult{ToolName: t.name, Data: data, Summary: summary}, nil
}

func decodeArguments(raw json.RawMessage, out any) error {
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}
	return nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

func integerSchema(description string) map[string]any {
	return map[string]any{
		"type":        "integer",
		"description": description,
	}
}

func stringSchema(description string) map[string]any {
	return map[string]any{
		"type":        "string",
		"description": description,
	}
}
