package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Feed struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt

	Name *string `gorm:"name;not null"`
	Link *string `gorm:"link;not null"`
	// LastBuild is the last time the content of the feed changed
	LastBuild *time.Time `gorm:"last_build"`
	// Failure is the reason of failure. If it is not null or empty, the fetch processor
	// should skip this feed
	Failure   *string `gorm:"failure;default:''"`
	Suspended *bool   `gorm:"suspended;default:false"`

	GroupID uint
	Group   Group
}

func (f Feed) IsFailed() bool {
	return f.Failure != nil && *f.Failure != ""
}

func (f Feed) IsSuspended() bool {
	return f.Suspended != nil && *f.Suspended
}
