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
	Suspended bool   `json:"suspended"`
	Proxy     string `json:"proxy,omitempty"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`

	FetchState FeedFetchState `json:"fetch_state"`

	UnreadCount int64 `json:"unread_count"`
	ItemCount   int64 `json:"item_count"`
}

// FeedFetchState stores runtime pull metadata for a feed.
// Time fields are Unix seconds; 0 means unknown/unset.
type FeedFetchState struct {
	// ETag is used for If-None-Match conditional requests.
	ETag string `json:"etag,omitempty"`
	// LastModified stores raw HTTP Last-Modified for If-Modified-Since.
	LastModified string `json:"last_modified,omitempty"`
	// CacheControl stores raw HTTP Cache-Control value.
	CacheControl string `json:"cache_control,omitempty"`
	// ExpiresAt is parsed from the HTTP Expires header.
	ExpiresAt int64 `json:"expires_at"`
	// LastCheckedAt is the last fetch attempt time (success or failure).
	LastCheckedAt int64 `json:"last_checked_at"`
	// NextCheckAt is the earliest next fetch time decided by pull policy.
	NextCheckAt int64 `json:"next_check_at"`
	// LastHTTPStatus is the last HTTP status code observed during fetch.
	LastHTTPStatus int `json:"last_http_status"`
	// RetryAfterUntil is derived from Retry-After and blocks fetch before this time.
	RetryAfterUntil int64 `json:"retry_after_until"`
	// LastSuccessAt is the last successful check time (includes 200 and 304).
	LastSuccessAt int64 `json:"last_success_at"`
	// LastErrorAt is the most recent fetch failure time.
	LastErrorAt int64 `json:"last_error_at"`
	// LastError keeps the latest fetch error message.
	LastError string `json:"last_error,omitempty"`
	// ConsecutiveFailures is reset on success and incremented on each failure.
	ConsecutiveFailures int64 `json:"consecutive_failures"`
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
