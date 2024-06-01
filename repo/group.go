package repo

import (
	"errors"

	"github.com/0x2e/fusion/model"

	"gorm.io/gorm"
)

func NewGroup(db *gorm.DB) *Group {
	return &Group{
		db: db,
	}
}

type Group struct {
	db *gorm.DB
}

func (g Group) All() ([]*model.Group, error) {
	var res []*model.Group
	err := g.db.Find(&res).Error
	return res, err
}

func (g Group) Get(id uint) (*model.Group, error) {
	var res model.Group
	err := g.db.First(&res, id).Error
	return &res, err
}

func (g Group) Create(group *model.Group) error {
	return g.db.Create(group).Error
}

func (g Group) Update(id uint, group *model.Group) error {
	return g.db.Model(&model.Group{}).Where("id = ?", id).Updates(group).Error
}

func (g Group) Delete(id uint) error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Feed{}).Where("group_id = ?", id).Update("group_id", 1).Error; err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}

		return tx.Delete(&model.Group{}, id).Error
	})
}
