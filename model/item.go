package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Item struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:guid"`

	Title    *string    `gorm:"title"`
	GUID     *string    `gorm:"guid;uniqueIndex:guid"`
	Link     *string    `gorm:"link"`
	Content  *string    `gorm:"content"`
	PubDate  *time.Time `gorm:"pub_date"`
	Unread   *bool      `gorm:"unread;default:true;index"`
	Bookmark *bool      `gorm:"bookmark;default:false;index"`

	FeedID uint `gorm:"feed_id;uniqueIndex:guid"`
	Feed   Feed
}
