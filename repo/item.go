package repo

import (
	"errors"

	"github.com/0x2e/fusion/model"

	"gorm.io/gorm"
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

func (i Item) List(filter ItemFilter, offset, count *int) ([]*model.Item, int, error) {
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

	if offset != nil {
		db = db.Offset(*offset)
	}
	if count != nil {
		db = db.Limit(*count)
	}
	err = db.Order("items.created_at desc").Joins("Feed").Find(&res).Error
	return res, int(total), err
}

func (i Item) Get(id uint) (*model.Item, error) {
	var res model.Item
	err := i.db.Joins("Feed").First(&res, id).Error
	return &res, err
}

func (i Item) Creates(items []*model.Item) error {
	return i.db.Create(items).Error
}

func (i Item) Update(id uint, item *model.Item) error {
	return i.db.Model(&model.Item{}).Where("id = ?", id).Updates(item).Error
}

func (i Item) Delete(id uint) error {
	return i.db.Delete(&model.Item{}, id).Error
}

func (i Item) DeleteByFeed(feedID uint) error {
	return i.db.Where("feed_id = ?", feedID).Delete(&model.Item{}).Error
}

func (i Item) IdentityExist(feedID uint, guid, link, title string) (bool, error) { // TODO: optimize
	err := i.db.Model(&model.Item{}).
		Where("feed_id = ? AND (guid = ? OR link = ? OR title = ?)", feedID, guid, link, title).
		First(&model.Item{}).Error
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
	}
	return true, err
}

func (i Item) UpdateUnread(ids []uint, unread *bool) error {
	return i.db.Model(&model.Item{}).Where("id IN ?", ids).Update("unread", unread).Error
}

func (i Item) UpdateBookmark(id uint, bookmark *bool) error {
	return i.db.Model(&model.Item{}).Where("id = ?", id).Update("bookmark", bookmark).Error
}
