package server

import (
	"errors"
	"net/http"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/repo"
)

//go:generate mockgen -destination=group_mock.go -source=group.go -package=server

type GroupRepo interface {
	All() ([]*model.Group, error)
	Create(group *model.Group) error
	Update(id uint, group *model.Group) error
	Delete(id uint) error
}

type FeedinGroupRepo interface {
	UpdateGroupID(from uint, to uint) error
}
type Group struct {
	groupRepo GroupRepo
	feedRepo  FeedinGroupRepo
}

func NewGroup(groupRepo GroupRepo, feedRepo FeedinGroupRepo) *Group {
	return &Group{
		groupRepo: groupRepo,
		feedRepo:  feedRepo,
	}
}

func (g Group) All() (*RespGroupAll, error) {
	data, err := g.groupRepo.All()
	if err != nil {
		return nil, err
	}

	groups := make([]*GroupForm, 0, len(data))
	for _, v := range data {
		groups = append(groups, &GroupForm{
			ID:   v.ID,
			Name: v.Name,
		})
	}
	return &RespGroupAll{
		Groups: groups,
	}, nil
}

func (g Group) Create(req *ReqGroupCreate) error {
	newGroup := &model.Group{
		Name: req.Name,
	}
	err := g.groupRepo.Create(newGroup)
	if errors.Is(err, repo.ErrDuplicatedKey) {
		err = NewBizError(err, http.StatusBadRequest, "name is not allowed to be the same as other groups")
	}
	return err
}

func (g Group) Update(req *ReqGroupUpdate) error {
	err := g.groupRepo.Update(req.ID, &model.Group{
		Name: req.Name,
	})
	if errors.Is(err, repo.ErrDuplicatedKey) {
		err = NewBizError(err, http.StatusBadRequest, "name is not allowed to be the same as other groups")
	}
	return err
}

func (g Group) Delete(req *ReqGroupDelete) error {
	if req.ID == 1 {
		return errors.New("cannot delete the default group")
	}
	// FIX: transaction
	if err := g.feedRepo.UpdateGroupID(req.ID, 1); err != nil && !errors.Is(err, repo.ErrNotFound) {
		return err
	}
	return g.groupRepo.Delete(req.ID)
}
