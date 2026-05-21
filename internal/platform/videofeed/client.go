package videofeed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) (*Client, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("parse video-feed base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("video-feed base url must include scheme and host")
	}

	return &Client{
		baseURL: trimmed,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *Client) GetVideoDetail(ctx context.Context, videoID uint) (Video, error) {
	var video Video
	err := c.post(ctx, "/video/getDetail", map[string]any{"id": videoID}, &video)
	return video, err
}

func (c *Client) GetHotVideos(ctx context.Context, limit int) (HotVideosResponse, error) {
	var response HotVideosResponse
	err := c.post(ctx, "/feed/listByPopularity", map[string]any{"limit": limit}, &response)
	return response, err
}

func (c *Client) GetVideoComments(ctx context.Context, videoID uint, limit int) ([]Comment, error) {
	var comments []Comment
	if err := c.post(ctx, "/comment/listAll", map[string]any{"video_id": videoID}, &comments); err != nil {
		return nil, err
	}
	return limitSlice(comments, limit), nil
}

func (c *Client) GetAuthorProfile(ctx context.Context, authorID uint) (AuthorProfile, error) {
	var profile AuthorProfile
	err := c.post(ctx, "/account/getProfile", map[string]any{"account_id": authorID}, &profile)
	return profile, err
}

func (c *Client) ListAuthorVideos(ctx context.Context, authorID uint, limit int) ([]Video, error) {
	var videos []Video
	if err := c.post(ctx, "/video/listByAuthorID", map[string]any{"author_id": authorID}, &videos); err != nil {
		return nil, err
	}
	return limitSlice(videos, limit), nil
}

func (c *Client) ListTagVideos(ctx context.Context, tagName string, limit int) ([]FeedVideoItem, error) {
	var response struct {
		VideoList []FeedVideoItem `json:"video_list"`
	}
	if err := c.post(ctx, "/feed/listByTag", map[string]any{"tag_name": tagName, "limit": limit}, &response); err != nil {
		return nil, err
	}
	return response.VideoList, nil
}

func (c *Client) post(ctx context.Context, path string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode video-feed %s request: %w", path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build video-feed %s request: %w", path, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call video-feed %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return responseError(path, resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode video-feed %s response: %w", path, err)
	}
	return nil
}

func responseError(path string, resp *http.Response) error {
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if readErr != nil {
		return fmt.Errorf("video-feed %s returned status %d and unreadable error body: %w", path, resp.StatusCode, readErr)
	}

	var parsed struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != "" {
		return fmt.Errorf("video-feed %s returned status %d: %s", path, resp.StatusCode, parsed.Error)
	}
	if len(body) == 0 {
		return fmt.Errorf("video-feed %s returned status %d", path, resp.StatusCode)
	}
	return fmt.Errorf("video-feed %s returned status %d: %s", path, resp.StatusCode, strings.TrimSpace(string(body)))
}

func limitSlice[T any](items []T, limit int) []T {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}
