package server

import (
	"context"
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

type Group struct {
	repo GroupRepo
}

func NewGroup(repo GroupRepo) *Group {
	return &Group{
		repo: repo,
	}
}

func (g Group) All(ctx context.Context) (*RespGroupAll, error) {
	data, err := g.repo.All()
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

func (g Group) Create(ctx context.Context, req *ReqGroupCreate) (*RespGroupCreate, error) {
	newGroup := &model.Group{
		Name: req.Name,
	}
	err := g.repo.Create(newGroup)
	if err != nil {
		if errors.Is(err, repo.ErrDuplicatedKey) {
			err = NewBizError(err, http.StatusBadRequest, "name is not allowed to be the same as other groups")
		}
		return nil, err
	}
	return &RespGroupCreate{ID: newGroup.ID}, nil
}

func (g Group) Update(ctx context.Context, req *ReqGroupUpdate) error {
	err := g.repo.Update(req.ID, &model.Group{
		Name: req.Name,
	})
	if errors.Is(err, repo.ErrDuplicatedKey) {
		err = NewBizError(err, http.StatusBadRequest, "name is not allowed to be the same as other groups")
	}
	return err
}

func (g Group) Delete(ctx context.Context, req *ReqGroupDelete) error {
	if req.ID == 1 {
		return errors.New("cannot delete the default group")
	}
	return g.repo.Delete(req.ID)
}
