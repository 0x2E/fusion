package server

import "time"

type ItemFeed struct {
	ID   uint    `json:"id"`
	Name *string `json:"name"`
	Link *string `json:"link"`
}

type ItemForm struct {
	ID        uint       `json:"id"`
	Title     *string    `json:"title"`
	Link      *string    `json:"link"`
	GUID      *string    `json:"guid"`
	Content   *string    `json:"content"`
	Unread    *bool      `json:"unread"`
	Bookmark  *bool      `json:"bookmark"`
	PubDate   *time.Time `json:"pub_date"`
	UpdatedAt *time.Time `json:"updated_at"`
	Feed      ItemFeed   `json:"feed"`
}

type ReqItemList struct {
	Paginate
	Keyword  *string `query:"keyword"`
	FeedID   *uint   `query:"feed_id"`
	Unread   *bool   `query:"unread"`
	Bookmark *bool   `query:"bookmark"`
}

type RespItemList struct {
	Total *int        `json:"total"`
	Items []*ItemForm `json:"items"`
}

type ReqItemGet struct {
	ID uint `param:"id" validate:"required"`
}

type RespItemGet ItemForm

type ReqItemDelete struct {
	ID uint `param:"id" validate:"required"`
}

type ReqItemUpdateUnread struct {
	IDs    []uint `json:"ids" validate:"required"`
	Unread *bool  `json:"unread" validate:"required"`
}

type ReqItemUpdateBookmark struct {
	ID       uint  `param:"id" validate:"required"`
	Bookmark *bool `json:"bookmark" validate:"required"`
}
