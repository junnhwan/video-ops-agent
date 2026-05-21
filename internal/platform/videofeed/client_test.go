package videofeed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetVideoDetailPostsIDAndDecodesVideo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/video/getDetail")
		requireJSONBody(t, r, map[string]any{"id": float64(42)})
		writeJSON(t, w, http.StatusOK, map[string]any{
			"id":          42,
			"author_id":   7,
			"username":    "alice",
			"title":       "Go Agent",
			"description": "ops report",
			"play_url":    "https://cdn.example/video.mp4",
			"cover_url":   "https://cdn.example/cover.jpg",
			"likes_count": 99,
			"popularity":  123,
			"create_time": "2026-05-21T08:00:00Z",
			"updated_at":  "2026-05-21T08:30:00Z",
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	video, err := client.GetVideoDetail(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetVideoDetail returned error: %v", err)
	}

	if video.ID != 42 || video.AuthorID != 7 || video.Title != "Go Agent" {
		t.Fatalf("decoded video = %+v", video)
	}
	if video.CreatedAt.IsZero() {
		t.Fatalf("expected create_time to decode into CreatedAt")
	}
}

func TestGetHotVideosPostsLimitAndDecodesPopularityResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/feed/listByPopularity")
		requireJSONBody(t, r, map[string]any{"limit": float64(3)})
		writeJSON(t, w, http.StatusOK, map[string]any{
			"video_list": []map[string]any{{
				"id":           11,
				"author":       map[string]any{"id": 2, "username": "bob"},
				"title":        "Hot Video",
				"play_url":     "https://cdn.example/hot.mp4",
				"cover_url":    "https://cdn.example/hot.jpg",
				"create_time":  int64(1779340800000),
				"likes_count":  55,
				"is_liked":     false,
				"description":  "hot context",
			}},
			"as_of":       1779340800,
			"next_offset": 3,
			"has_more":    true,
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.GetHotVideos(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetHotVideos returned error: %v", err)
	}

	if len(resp.VideoList) != 1 || resp.VideoList[0].ID != 11 || resp.NextOffset != 3 || !resp.HasMore {
		t.Fatalf("decoded hot videos response = %+v", resp)
	}
}

func TestGetVideoCommentsTruncatesClientSideLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/comment/listAll")
		requireJSONBody(t, r, map[string]any{"video_id": float64(5)})
		writeJSON(t, w, http.StatusOK, []map[string]any{
			{"id": 1, "video_id": 5, "author_id": 9, "username": "u1", "content": "first", "created_at": "2026-05-21T08:00:00Z"},
			{"id": 2, "video_id": 5, "author_id": 10, "username": "u2", "content": "second", "created_at": "2026-05-21T08:01:00Z"},
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	comments, err := client.GetVideoComments(context.Background(), 5, 1)
	if err != nil {
		t.Fatalf("GetVideoComments returned error: %v", err)
	}

	if len(comments) != 1 || comments[0].Content != "first" {
		t.Fatalf("comments = %+v, want first item only", comments)
	}
}

func TestGetAuthorProfilePostsAccountIDAndDecodesProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/account/getProfile")
		requireJSONBody(t, r, map[string]any{"account_id": float64(8)})
		writeJSON(t, w, http.StatusOK, map[string]any{
			"account":        map[string]any{"id": 8, "username": "creator", "avatar_url": "https://cdn.example/a.png", "bio": "Go dev"},
			"video_count":    12,
			"total_likes":    345,
			"follower_count": 67,
			"vlogger_count":  4,
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	profile, err := client.GetAuthorProfile(context.Background(), 8)
	if err != nil {
		t.Fatalf("GetAuthorProfile returned error: %v", err)
	}

	if profile.Account.ID != 8 || profile.VideoCount != 12 || profile.TotalLikes != 345 {
		t.Fatalf("decoded author profile = %+v", profile)
	}
}

func TestListAuthorVideosTruncatesClientSideLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/video/listByAuthorID")
		requireJSONBody(t, r, map[string]any{"author_id": float64(8)})
		writeJSON(t, w, http.StatusOK, []map[string]any{
			{"id": 1, "author_id": 8, "username": "creator", "title": "one", "play_url": "p1", "cover_url": "c1", "create_time": "2026-05-21T08:00:00Z"},
			{"id": 2, "author_id": 8, "username": "creator", "title": "two", "play_url": "p2", "cover_url": "c2", "create_time": "2026-05-21T08:01:00Z"},
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	videos, err := client.ListAuthorVideos(context.Background(), 8, 1)
	if err != nil {
		t.Fatalf("ListAuthorVideos returned error: %v", err)
	}

	if len(videos) != 1 || videos[0].Title != "one" {
		t.Fatalf("videos = %+v, want first item only", videos)
	}
}

func TestListTagVideosPostsTagAndLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireRequest(t, r, http.MethodPost, "/feed/listByTag")
		requireJSONBody(t, r, map[string]any{"tag_name": "Go", "limit": float64(2)})
		writeJSON(t, w, http.StatusOK, map[string]any{
			"video_list": []map[string]any{{
				"id":          20,
				"author":      map[string]any{"id": 3, "username": "tagger"},
				"title":       "Tagged",
				"play_url":    "p",
				"cover_url":   "c",
				"create_time": int64(1779340800000),
			}},
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	videos, err := client.ListTagVideos(context.Background(), "Go", 2)
	if err != nil {
		t.Fatalf("ListTagVideos returned error: %v", err)
	}

	if len(videos) != 1 || videos[0].ID != 20 || videos[0].Author.Username != "tagger" {
		t.Fatalf("tag videos = %+v", videos)
	}
}

func TestClientReturnsErrorForNonOKResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusInternalServerError, map[string]any{"error": "database unavailable"})
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.GetVideoDetail(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error for non-OK response")
	}
	if !strings.Contains(err.Error(), "500") || !strings.Contains(err.Error(), "database unavailable") {
		t.Fatalf("error = %q, want status and response error", err.Error())
	}
}

func TestClientReturnsErrorForMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.GetVideoDetail(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("error = %q, want decode context", err.Error())
	}
}

func requireRequest(t *testing.T, r *http.Request, method string, path string) {
	t.Helper()
	if r.Method != method {
		t.Fatalf("method = %q, want %q", r.Method, method)
	}
	if r.URL.Path != path {
		t.Fatalf("path = %q, want %q", r.URL.Path, path)
	}
	if got := r.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}
}

func requireJSONBody(t *testing.T, r *http.Request, want map[string]any) {
	t.Helper()

	var got map[string]any
	if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}

	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("body[%s] = %#v, want %#v; full body = %#v", key, got[key], wantValue, got)
		}
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}
