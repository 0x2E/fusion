package repo

import (
	"time"

	"github.com/0x2e/fusion/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewItem(db *gorm.DB) *Item {
	return &Item{
		db: db,
	}
}

type Item struct {
	db *gorm.DB
}

type ItemFilter struct {
	Keyword  *string
	FeedID   *uint
	Unread   *bool
	Bookmark *bool
}

func (i Item) List(filter ItemFilter, page, pageSize int) ([]*model.Item, int, error) {
	var total int64
	var res []*model.Item
	db := i.db.Model(&model.Item{})
	if filter.Keyword != nil {
		expr := "%" + *filter.Keyword + "%"
		db = db.Where("title LIKE ? OR content LIKE ?", expr, expr)
	}
	if filter.FeedID != nil {
		db = db.Where("feed_id = ?", *filter.FeedID)
	}
	if filter.Unread != nil {
		db = db.Where("unread = ?", *filter.Unread)
	}
	if filter.Bookmark != nil {
		db = db.Where("bookmark = ?", *filter.Bookmark)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Joins("Feed").Order("items.pub_date desc, items.created_at desc").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&res).Error
	return res, int(total), err
}

func (i Item) Get(id uint) (*model.Item, error) {
	var res model.Item
	err := i.db.Joins("Feed").First(&res, id).Error
	return &res, err
}

func (i Item) Creates(items []*model.Item) error {
	// limit batchSize to fix 'too many SQL variable' error
	now := time.Now()
	for _, i := range items {
		i.CreatedAt = now
		i.UpdatedAt = now
	}
	return i.db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(items, 5).Error
}

func (i Item) Update(id uint, item *model.Item) error {
	return i.db.Model(&model.Item{}).Where("id = ?", id).Updates(item).Error
}

func (i Item) Delete(id uint) error {
	return i.db.Delete(&model.Item{}, id).Error
}

func (i Item) UpdateUnread(ids []uint, unread *bool) error {
	return i.db.Model(&model.Item{}).Where("id IN ?", ids).Update("unread", unread).Error
}

func (i Item) UpdateBookmark(id uint, bookmark *bool) error {
	return i.db.Model(&model.Item{}).Where("id = ?", id).Update("bookmark", bookmark).Error
}
