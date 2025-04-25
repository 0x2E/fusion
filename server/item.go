package server

import (
	"context"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/repo"
)

type ItemRepo interface {
	List(filter repo.ItemFilter, page, pageSize int) ([]*model.Item, int, error)
	Get(id uint) (*model.Item, error)
	Delete(id uint) error
	UpdateUnread(ids []uint, unread *bool) error
	UpdateBookmark(id uint, bookmark *bool) error
}

type Item struct {
	repo ItemRepo
}

func NewItem(repo ItemRepo) *Item {
	return &Item{
		repo: repo,
	}
}

func (i Item) List(ctx context.Context, req *ReqItemList) (*RespItemList, error) {
	filter := repo.ItemFilter{
		Keyword:  req.Keyword,
		FeedID:   req.FeedID,
		GroupID:  req.GroupID,
		Unread:   req.Unread,
		Bookmark: req.Bookmark,
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	data, total, err := i.repo.List(filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*ItemForm, 0, len(data))
	for _, v := range data {
		items = append(items, &ItemForm{
			ID:        v.ID,
			GUID:      v.GUID,
			Title:     v.Title,
			Link:      v.Link,
			Unread:    v.Unread,
			Bookmark:  v.Bookmark,
			PubDate:   v.PubDate,
			UpdatedAt: &v.UpdatedAt,
			Feed: ItemFeed{
				ID:   v.Feed.ID,
				Name: v.Feed.Name,
				Link: v.Feed.Link,
			},
		})
	}
	return &RespItemList{
		Total: &total,
		Items: items,
	}, nil
}

func (i Item) Get(ctx context.Context, req *ReqItemGet) (*RespItemGet, error) {
	data, err := i.repo.Get(req.ID)
	if err != nil {
		return nil, err
	}

	return &RespItemGet{
		ID:        data.ID,
		GUID:      data.GUID,
		Title:     data.Title,
		Link:      data.Link,
		Content:   data.Content,
		Unread:    data.Unread,
		Bookmark:  data.Bookmark,
		PubDate:   data.PubDate,
		UpdatedAt: &data.UpdatedAt,
		Feed: ItemFeed{
			ID:   data.Feed.ID,
			Name: data.Feed.Name,
			Link: data.Feed.Link,
		},
	}, nil
}

func (i Item) Delete(ctx context.Context, req *ReqItemDelete) error {
	return i.repo.Delete(req.ID)
}

func (i Item) UpdateUnread(ctx context.Context, req *ReqItemUpdateUnread) error {
	return i.repo.UpdateUnread(req.IDs, req.Unread)
}

func (i Item) UpdateBookmark(ctx context.Context, req *ReqItemUpdateBookmark) error {
	return i.repo.UpdateBookmark(req.ID, req.Bookmark)
}
