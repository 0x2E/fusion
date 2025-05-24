package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Item struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_guid"`

	Title    *string    `gorm:"title"`
	GUID     *string    `gorm:"guid;uniqueIndex:idx_guid"`
	Link     *string    `gorm:"link"`
	Image    *string    `gorm:"image"`
	Content  *string    `gorm:"content"`
	PubDate  *time.Time `gorm:"pub_date"`
	Unread   *bool      `gorm:"unread;default:true;index"`
	Bookmark *bool      `gorm:"bookmark;default:false;index"`

	FeedID uint `gorm:"feed_id;uniqueIndex:idx_guid"`
	Feed   Feed
}
