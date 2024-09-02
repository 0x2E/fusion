package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type FeedRequestOptions struct {
	ReqProxy *string `gorm:"req_proxy"`

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
	// Failure is the reason of failure. If it is not null or empty, the fetch processor
	// should skip this feed
	Failure   *string `gorm:"failure;default:''"`
	Suspended *bool   `gorm:"suspended;default:false"`

	FeedRequestOptions

	GroupID uint
	Group   Group
}

func (f Feed) IsFailed() bool {
	return f.Failure != nil && *f.Failure != ""
}

func (f Feed) IsSuspended() bool {
	return f.Suspended != nil && *f.Suspended
}
