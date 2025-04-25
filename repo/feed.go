package repo

import (
	"errors"

	"github.com/0x2e/fusion/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewFeed(db *gorm.DB) *Feed {
	return &Feed{
		db: db,
	}
}

type Feed struct {
	db *gorm.DB
}

type FeedListFilter struct {
	HaveUnread   *bool
	HaveBookmark *bool
}

func (f Feed) List(filter *FeedListFilter) ([]*model.Feed, error) {
	var res []*model.Feed
	db := f.db.Model(&model.Feed{}).Joins("Group")
	if filter != nil {
		if filter.HaveUnread != nil && *filter.HaveUnread {
			db = db.Joins("inner join items on feeds.id = items.feed_id and items.unread = true").
				Group("feeds.id")
		}
		if filter.HaveBookmark != nil && *filter.HaveBookmark {
			db = db.Joins("inner join items on feeds.id = items.feed_id and items.bookmark = true").
				Group("feeds.id")
		}
	}

	err := db.Find(&res).Error
	if err != nil {
		return nil, err
	}

	// count unread items of each feed.
	// yeah this is stupid, but I don't know how to do it in a single query using GORM.
	ids := make([]uint, 0, len(res))
	for _, feed := range res {
		ids = append(ids, feed.ID)
	}
	var itemUnreadCount []struct {
		FeedID uint  `gorm:"feed_id"`
		Count  int64 `gorm:"count"`
	}
	err = f.db.Model(&model.Item{}).
		Select("feed_id, count(*) as count").
		Where("feed_id in ?", ids).
		Where("unread = true").
		Group("feed_id").
		Find(&itemUnreadCount).Error
	if err != nil {
		return nil, err
	}
	for _, feed := range res {
		for _, count := range itemUnreadCount {
			if feed.ID == count.FeedID {
				feed.UnreadCount = int(count.Count)
				break
			}
		}
	}

	return res, nil
}

func (f Feed) Get(id uint) (*model.Feed, error) {
	var res model.Feed
	err := f.db.Model(&model.Feed{}).Joins("Group").First(&res, id).Error
	return &res, err
}

func (f Feed) Create(data []*model.Feed) error {
	return f.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "link"}, {Name: "deleted_at"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "link", "req_proxy", "group_id"}),
	}).Create(data).Error
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
