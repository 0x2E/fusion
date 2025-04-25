package server

import "time"

type FeedForm struct {
	ID          uint      `json:"id"`
	Name        *string   `json:"name"`
	Link        *string   `json:"link"`
	Failure     *string   `json:"failure"`
	Suspended   *bool     `json:"suspended"`
	ReqProxy    *string   `json:"req_proxy"`
	UpdatedAt   time.Time `json:"updated_at"`
	UnreadCount int       `json:"unread_count"`
	Group       GroupForm `json:"group"`
}

type ReqFeedList struct {
	HaveUnread   *bool `query:"have_unread"`
	HaveBookmark *bool `query:"have_bookmark"`
}

type RespFeedList struct {
	Feeds []*FeedForm `json:"feeds"`
}

type ReqFeedGet struct {
	ID uint `param:"id" validate:"required"`
}

type RespFeedGet FeedForm

type FeedRequestOptions struct {
	Proxy *string `json:"proxy"`
}

type ReqFeedCheckValidity struct {
	Link           string             `json:"link" validate:"required"`
	RequestOptions FeedRequestOptions `json:"request_options"`
}

type ValidityItem struct {
	Title *string `json:"title"`
	Link  *string `json:"link"`
}

type RespFeedCheckValidity struct {
	FeedLinks []ValidityItem `json:"feed_links"`
}

type ReqFeedCreate struct {
	Feeds []struct {
		Name           *string            `json:"name" validate:"required"`
		Link           *string            `json:"link" validate:"required"`
		RequestOptions FeedRequestOptions `json:"request_options"`
	} `json:"feeds" validate:"required"`
	GroupID uint `json:"group_id" validate:"required"`
}

type ReqFeedUpdate struct {
	ID        uint    `param:"id" validate:"required"`
	Name      *string `json:"name"`
	Link      *string `json:"link"`
	Suspended *bool   `json:"suspended"`
	ReqProxy  *string `json:"req_proxy"`
	GroupID   *uint   `json:"group_id"`
}

type ReqFeedDelete struct {
	ID uint `param:"id" validate:"required"`
}

type ReqFeedRefresh struct {
	ID  *uint `json:"id"`
	All *bool `json:"all"`
}
