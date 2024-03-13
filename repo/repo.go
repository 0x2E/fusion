package repo

import (
	"errors"

	"github.com/0x2e/fusion/conf"
	"github.com/0x2e/fusion/model"

	"gorm.io/driver/sqlite"
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
	// v0.4.0 adds unique index (guid, feed_id, deleted_at).
	// clear data before AutoMigrate create the unique index.
	// TODO: remove in v1.0.0
	if err := DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec("UPDATE items as i1 SET guid = (SELECT link FROM items as i2 WHERE i1.id = i2.id) WHERE guid = '' OR guid IS NULL").Error
		if err != nil {
			return err
		}
		return tx.Exec("DELETE FROM items WHERE id NOT IN (SELECT MIN(id) FROM items GROUP BY feed_id, deleted_at, guid)").Error
	}); err != nil {
		panic(err)
	}

	// FIX: gorm not auto drop index and change 'not null'
	if err := DB.Debug().AutoMigrate(&model.Feed{}, &model.Group{}, &model.Item{}); err != nil {
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
