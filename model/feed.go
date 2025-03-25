package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type FeedRequestOptions struct {
	ReqProxy *string `gorm:"req_proxy"`

	// LastModified is the value that the server provided in the Last-Modified
	// header on our last successful request to this feed.
	LastModified *string

	// TODO: headers, cookie, etc.
}

type Feed struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_link"`

	Name *string `gorm:"name;not null"`
	Link *string `gorm:"link;not null;uniqueIndex:idx_link"`
	// LastBuild is the last time the content of the feed changed
	LastBuild *time.Time `gorm:"last_build"`
	// Failure is the error message for the last fetch.
	Failure *string `gorm:"failure;default:''"`
	// ConsecutiveFailures is the number of consecutive times we've failed to
	// retrieve this feed.
	ConsecutiveFailures uint `gorm:"consecutive_failures;default:0"`

	// LastModified is the value that the server provided in the Last-Modified
	// header on our last successful request to this feed.
	LastModified *string `gorm:"last_modified_header"`

	Suspended *bool `gorm:"suspended;default:false"`

	FeedRequestOptions

	GroupID uint
	Group   Group
}

func (f Feed) IsSuspended() bool {
	return f.Suspended != nil && *f.Suspended
}
