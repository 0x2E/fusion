package model

// Group represents a feed group.
type Group struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// Feed represents an RSS/Atom feed.
type Feed struct {
	ID        int64  `json:"id"`
	GroupID   int64  `json:"group_id"`
	Name      string `json:"name"`
	Link      string `json:"link"`
	SiteURL   string `json:"site_url,omitempty"`
	LastBuild int64  `json:"last_build"`
	Failure   string `json:"failure,omitempty"`
	Failures  int64  `json:"failures"`
	Suspended bool   `json:"suspended"`
	Proxy     string `json:"proxy,omitempty"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`

	UnreadCount int64 `json:"unread_count"`
}

// Item represents a feed item.
type Item struct {
	ID        int64  `json:"id"`
	FeedID    int64  `json:"feed_id"`
	GUID      string `json:"guid"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Content   string `json:"content"`
	PubDate   int64  `json:"pub_date"`
	Unread    bool   `json:"unread"`
	CreatedAt int64  `json:"created_at"`
}

// Bookmark represents a saved item snapshot.
type Bookmark struct {
	ID        int64  `json:"id"`
	ItemID    *int64 `json:"item_id"` // nullable
	Link      string `json:"link"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	PubDate   int64  `json:"pub_date"`
	FeedName  string `json:"feed_name"`
	CreatedAt int64  `json:"created_at"`
}
