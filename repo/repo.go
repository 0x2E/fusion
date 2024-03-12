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

	if err := migrage(); err != nil {
		panic(err)
	}

	if err := registerCallback(); err != nil {
		panic(err)
	}

	if conf.Debug {
		DB = DB.Debug()
	}
}

func migrage() error {
	// FIX: gorm not auto drop index and change 'not null'
	if err := DB.AutoMigrate(&model.Feed{}, &model.Group{}, &model.Item{}); err != nil {
		return err
	}

	defaultGroup := "Default"
	if err := DB.Model(&model.Group{}).Where("id = ?", 1).
		FirstOrCreate(&model.Group{ID: 1, Name: &defaultGroup}).Error; err != nil {
		return err
	}

	if err := DB.Table("items as i1").Where("guid = '' OR guid is null").UpdateColumn(
		"guid",
		DB.Table("items as i2").Select("link").Where("i1.id = i2.id"),
	).Error; err != nil {
		return err
	}

	return nil
}

func registerCallback() error {
	if err := DB.Callback().Query().After("*").Register("convert_error", func(db *gorm.DB) {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			db.Error = ErrNotFound
		}
	}); err != nil {
		return err
	}

	if err := DB.Callback().Create().After("*").Register("convert_error", func(db *gorm.DB) {
		if errors.Is(db.Error, gorm.ErrDuplicatedKey) {
			db.Error = ErrDuplicatedKey
		}
	}); err != nil {
		return err
	}

	if err := DB.Callback().Update().After("*").Register("convert_error", func(db *gorm.DB) {
		if db.Error == nil && db.RowsAffected == 0 {
			db.Error = ErrNotFound
		}
		if errors.Is(db.Error, gorm.ErrDuplicatedKey) {
			db.Error = ErrDuplicatedKey
		}
	}); err != nil {
		return err
	}

	if err := DB.Callback().Delete().After("*").Register("convert_error", func(db *gorm.DB) {
		if db.Error == nil && db.RowsAffected == 0 {
			db.Error = ErrNotFound
		}
		if errors.Is(db.Error, gorm.ErrDuplicatedKey) {
			db.Error = ErrDuplicatedKey
		}
	}); err != nil {
		return err
	}
	return nil
}
