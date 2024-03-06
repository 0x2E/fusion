package repo

import (
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

func (f Group) All() ([]*model.Group, error) {
	var res []*model.Group
	err := f.db.Find(&res).Error
	return res, err
}

func (f Group) Get(id uint) (*model.Group, error) {
	var res model.Group
	err := f.db.First(&res, id).Error
	return &res, err
}

func (f Group) Create(group *model.Group) error {
	return f.db.Create(group).Error
}

func (f Group) Update(id uint, group *model.Group) error {
	return f.db.Model(&model.Group{}).Where("id = ?", id).Updates(group).Error
}

func (f Group) Delete(id uint) error {
	return f.db.Delete(&model.Group{}, id).Error
}
