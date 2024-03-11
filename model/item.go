package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Item struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt

	Title    *string    `gorm:"title;not null;index"`
	GUID     *string    `gorm:"guid,index"`
	Link     *string    `gorm:"link,index"`
	Content  *string    `gorm:"content"`
	PubDate  *time.Time `gorm:"pub_date"`
	Unread   *bool      `gorm:"unread;default:true;index"`
	Bookmark *bool      `gorm:"bookmark;default:false;index"`

	FeedID uint `gorm:"feed_id;index"`
	Feed   Feed
}
