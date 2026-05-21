package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"video-ops-agent/internal/platform/videofeed"
)

func TestPlatformToolsExecuteThroughRegistryWithFakeClient(t *testing.T) {
	fake := &fakePlatformClient{
		video: videofeed.Video{ID: 7, AuthorID: 3, Username: "creator", Title: "hot", LikesCount: 99},
		hotVideos: videofeed.HotVideosResponse{
			VideoList: []videofeed.FeedVideoItem{{ID: 7, Title: "hot"}},
			HasMore:   false,
		},
		comments: []videofeed.Comment{{ID: 1, VideoID: 7, Content: "good"}},
		profile: videofeed.AuthorProfile{
			Account:    videofeed.Account{ID: 3, Username: "creator"},
			VideoCount: 5,
		},
		authorVideos: []videofeed.Video{{ID: 7, AuthorID: 3, Title: "hot"}},
		tagVideos:    []videofeed.FeedVideoItem{{ID: 8, Title: "go tag"}},
	}
	registry, err := NewDefaultRegistry(fake)
	if err != nil {
		t.Fatalf("NewDefaultRegistry returned error: %v", err)
	}
	executor := NewExecutor(registry, time.Second)

	cases := []struct {
		name    string
		args    string
		summary string
	}{
		{name: "get_video_detail", args: `{"video_id":7}`, summary: "video 7"},
		{name: "get_hot_videos", args: `{"limit":2}`, summary: "1 hot videos"},
		{name: "get_video_comments", args: `{"video_id":7,"limit":20}`, summary: "1 comments"},
		{name: "get_author_profile", args: `{"author_id":3}`, summary: "author 3"},
		{name: "list_author_videos", args: `{"author_id":3,"limit":20}`, summary: "1 author videos"},
		{name: "list_tag_videos", args: `{"tag_name":"Go","limit":20}`, summary: "1 tag videos"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := executor.Execute(context.Background(), tc.name, json.RawMessage(tc.args))
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if result.ToolName != tc.name {
				t.Fatalf("tool name = %q, want %q", result.ToolName, tc.name)
			}
			if !strings.Contains(result.Summary, tc.summary) {
				t.Fatalf("summary = %q, want substring %q", result.Summary, tc.summary)
			}
			if result.Data == nil {
				t.Fatalf("expected structured data")
			}
		})
	}

	if fake.videoID != 7 || fake.hotLimit != 2 || fake.commentsLimit != 20 ||
		fake.authorID != 3 || fake.authorVideosLimit != 20 || fake.tagName != "Go" || fake.tagLimit != 20 {
		t.Fatalf("fake client saw unexpected calls: %+v", fake)
	}
}

func TestPlatformToolRejectsInvalidArguments(t *testing.T) {
	registry, err := NewDefaultRegistry(&fakePlatformClient{})
	if err != nil {
		t.Fatalf("NewDefaultRegistry returned error: %v", err)
	}
	executor := NewExecutor(registry, time.Second)

	_, err = executor.Execute(context.Background(), "get_video_detail", json.RawMessage(`{"video_id":0}`))
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "video_id") {
		t.Fatalf("error = %q, want video_id context", err.Error())
	}

	_, err = executor.Execute(context.Background(), "list_tag_videos", json.RawMessage(`{"tag_name":"  ","limit":10}`))
	if err == nil {
		t.Fatalf("expected tag_name validation error")
	}
	if !strings.Contains(err.Error(), "tag_name") {
		t.Fatalf("error = %q, want tag_name context", err.Error())
	}
}

type fakePlatformClient struct {
	video        videofeed.Video
	hotVideos    videofeed.HotVideosResponse
	comments     []videofeed.Comment
	profile      videofeed.AuthorProfile
	authorVideos []videofeed.Video
	tagVideos    []videofeed.FeedVideoItem

	videoID           uint
	hotLimit          int
	commentsVideoID   uint
	commentsLimit     int
	authorID          uint
	authorVideosID    uint
	authorVideosLimit int
	tagName           string
	tagLimit          int
}

func (c *fakePlatformClient) GetVideoDetail(_ context.Context, videoID uint) (videofeed.Video, error) {
	c.videoID = videoID
	return c.video, nil
}

func (c *fakePlatformClient) GetHotVideos(_ context.Context, limit int) (videofeed.HotVideosResponse, error) {
	c.hotLimit = limit
	return c.hotVideos, nil
}

func (c *fakePlatformClient) GetVideoComments(_ context.Context, videoID uint, limit int) ([]videofeed.Comment, error) {
	c.commentsVideoID = videoID
	c.commentsLimit = limit
	return c.comments, nil
}

func (c *fakePlatformClient) GetAuthorProfile(_ context.Context, authorID uint) (videofeed.AuthorProfile, error) {
	c.authorID = authorID
	return c.profile, nil
}

func (c *fakePlatformClient) ListAuthorVideos(_ context.Context, authorID uint, limit int) ([]videofeed.Video, error) {
	c.authorVideosID = authorID
	c.authorVideosLimit = limit
	return c.authorVideos, nil
}

func (c *fakePlatformClient) ListTagVideos(_ context.Context, tagName string, limit int) ([]videofeed.FeedVideoItem, error) {
	c.tagName = tagName
	c.tagLimit = limit
	return c.tagVideos, nil
}
