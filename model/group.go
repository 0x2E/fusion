package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type Group struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_name"`

	Name *string `gorm:"name;not null;uniqueIndex:idx_name"`
}
