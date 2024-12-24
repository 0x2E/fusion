package repo

import (
	"errors"
	"log"

	"github.com/0x2e/fusion/conf"
	"github.com/0x2e/fusion/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	conn, err := gorm.Open(
		sqlite.Open(conf.Conf.DB),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		panic(err)
	}
	DB = conn

	migrage()
	registerCallback()
}

func migrage() {
	// The verison after v0.8.7 will add a unique index to Feed.Link.
	// We must delete any duplicate feeds before AutoMigrate applies the
	// new unique constraint.
	err := DB.Transaction(func(tx *gorm.DB) error {
		// skip when it's the first launch
		if !tx.Migrator().HasTable(&model.Feed{}) || !tx.Migrator().HasTable(&model.Item{}) {
			return nil
		}

		// query duplicate feeds
		dupFeeds := make([]model.Feed, 0)
		err := tx.Model(&model.Feed{}).Where(
			"link IN (?)",
			tx.Model(&model.Feed{}).Select("link").Group("link").
				Having("count(link) > 1"),
		).Order("link, id").Find(&dupFeeds).Error
		if err != nil {
			return err
		}

		// filter out feeds that will be deleted.
		// we've queried with order, so the first one is the one we should keep.
		distinct := map[string]uint{}
		deleteIDs := make([]uint, 0, len(dupFeeds))
		for _, f := range dupFeeds {
			if _, ok := distinct[*f.Link]; !ok {
				distinct[*f.Link] = f.ID
				continue
			}
			deleteIDs = append(deleteIDs, f.ID)
			log.Println("delete duplicate feed: ", f.ID, *f.Name, *f.Link)
		}

		if len(deleteIDs) > 0 {
			// **hard** delete duplicate feeds and their items
			err = tx.Where("id IN ?", deleteIDs).Unscoped().Delete(&model.Feed{}).Error
			if err != nil {
				return err
			}
			return tx.Where("feed_id IN ?", deleteIDs).Unscoped().Delete(&model.Item{}).Error
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	// FIX: gorm not auto drop index and change 'not null'
	if err := DB.AutoMigrate(&model.Feed{}, &model.Group{}, &model.Item{}); err != nil {
		panic(err)
	}

	defaultGroup := "Default"
	if err := DB.Model(&model.Group{}).Where("id = ?", 1).
		FirstOrCreate(&model.Group{ID: 1, Name: &defaultGroup}).Error; err != nil {
		panic(err)
	}
}

func registerCallback() {
	if err := DB.Callback().Query().After("*").Register("convert_error", func(db *gorm.DB) {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			db.Error = ErrNotFound
		}
	}); err != nil {
		panic(err)
	}

	if err := DB.Callback().Create().After("*").Register("convert_error", func(db *gorm.DB) {
		if errors.Is(db.Error, gorm.ErrDuplicatedKey) {
			db.Error = ErrDuplicatedKey
		}
	}); err != nil {
		panic(err)
	}

	if err := DB.Callback().Update().After("*").Register("convert_error", func(db *gorm.DB) {
		if db.Error == nil && db.RowsAffected == 0 {
			db.Error = ErrNotFound
		}
		if errors.Is(db.Error, gorm.ErrDuplicatedKey) {
			db.Error = ErrDuplicatedKey
		}
	}); err != nil {
		panic(err)
	}

	if err := DB.Callback().Delete().After("*").Register("convert_error", func(db *gorm.DB) {
		if db.Error == nil && db.RowsAffected == 0 {
			db.Error = ErrNotFound
		}
		if errors.Is(db.Error, gorm.ErrDuplicatedKey) {
			db.Error = ErrDuplicatedKey
		}
	}); err != nil {
		panic(err)
	}
}
