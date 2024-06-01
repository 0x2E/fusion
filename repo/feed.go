package repo

import (
	"errors"

	"github.com/0x2e/fusion/model"

	"gorm.io/gorm"
)

func NewFeed(db *gorm.DB) *Feed {
	return &Feed{
		db: db,
	}
}

type Feed struct {
	db *gorm.DB
}

func (f Feed) All() ([]*model.Feed, error) {
	var res []*model.Feed
	err := f.db.Model(&model.Feed{}).Joins("Group").Find(&res).Error
	return res, err
}

func (f Feed) Get(id uint) (*model.Feed, error) {
	var res model.Feed
	err := f.db.Model(&model.Feed{}).Joins("Group").First(&res, id).Error
	return &res, err
}

func (f Feed) Create(data []*model.Feed) error {
	return f.db.Create(data).Error
}

func (f Feed) Update(id uint, feed *model.Feed) error {
	return f.db.Model(&model.Feed{}).Where("id = ?", id).Updates(feed).Error
}

func (f Feed) Delete(id uint) error {
	return f.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Item{}).Where("feed_id = ?", id).Delete(&model.Item{}).Error; err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		return tx.Delete(&model.Feed{}, id).Error
	})
}
