package videofeed

import "time"

type Video struct {
	ID          uint      `json:"id"`
	AuthorID    uint      `json:"author_id"`
	Username    string    `json:"username"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	PlayURL     string    `json:"play_url"`
	CoverURL    string    `json:"cover_url"`
	LikesCount  int64     `json:"likes_count"`
	Popularity  int64     `json:"popularity"`
	CreatedAt   time.Time `json:"create_time"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Comment struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	VideoID   uint      `json:"video_id"`
	AuthorID  uint      `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Account struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Bio       string `json:"bio,omitempty"`
}

type AuthorProfile struct {
	Account       Account `json:"account"`
	VideoCount    int64   `json:"video_count"`
	TotalLikes    int64   `json:"total_likes"`
	FollowerCount int64   `json:"follower_count"`
	VloggerCount  int64   `json:"vlogger_count"`
}

type FeedAuthor struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type FeedVideoItem struct {
	ID          uint       `json:"id"`
	Author      FeedAuthor `json:"author"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	PlayURL     string     `json:"play_url"`
	CoverURL    string     `json:"cover_url"`
	CreateTime  int64      `json:"create_time"`
	LikesCount  int64      `json:"likes_count"`
	IsLiked     bool       `json:"is_liked"`
}

type HotVideosResponse struct {
	VideoList            []FeedVideoItem `json:"video_list"`
	AsOf                 int64           `json:"as_of"`
	NextOffset           int             `json:"next_offset"`
	HasMore              bool            `json:"has_more"`
	NextLatestPopularity *int64          `json:"next_latest_popularity,omitempty"`
	NextLatestBefore     *time.Time      `json:"next_latest_before,omitempty"`
	NextLatestIDBefore   *uint           `json:"next_latest_id_before,omitempty"`
}
