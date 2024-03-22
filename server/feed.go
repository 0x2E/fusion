package server

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/repo"
	"github.com/0x2e/fusion/service/pull"
	"github.com/0x2e/fusion/service/sniff"
)

//go:generate mockgen -destination=feed_mock.go -source=feed.go -package=server

type FeedRepo interface {
	All() ([]*model.Feed, error)
	Get(id uint) (*model.Feed, error)
	Create(feed []*model.Feed) error
	Update(id uint, feed *model.Feed) error
	Delete(id uint) error
}

type ItemInFeedRepo interface {
	DeleteByFeed(id uint) error
}

type Feed struct {
	feedRepo FeedRepo
	itemRepo ItemInFeedRepo
}

func NewFeed(feedRepo FeedRepo, itemRepo ItemInFeedRepo) *Feed {
	return &Feed{
		feedRepo: feedRepo,
		itemRepo: itemRepo,
	}
}

func (f Feed) All(ctx context.Context) (*RespFeedAll, error) {
	data, err := f.feedRepo.All()
	if err != nil {
		return nil, err
	}

	feeds := make([]*FeedForm, 0, len(data))
	for _, v := range data {
		feeds = append(feeds, &FeedForm{
			ID:        v.ID,
			Name:      v.Name,
			Link:      v.Link,
			Failure:   v.Failure,
			Suspended: v.Suspended,
			UpdatedAt: v.UpdatedAt,
			Group:     GroupForm{ID: v.GroupID, Name: v.Group.Name},
		})
	}
	return &RespFeedAll{
		Feeds: feeds,
	}, nil
}

func (f Feed) Get(ctx context.Context, req *ReqFeedGet) (*RespFeedGet, error) {
	data, err := f.feedRepo.Get(req.ID)
	if err != nil {
		return nil, err
	}

	return &RespFeedGet{
		ID:        data.ID,
		Name:      data.Name,
		Link:      data.Link,
		Failure:   data.Failure,
		Suspended: data.Suspended,
		Group:     GroupForm{ID: data.GroupID, Name: data.Group.Name},
	}, nil
}

func (f Feed) Create(ctx context.Context, req *ReqFeedCreate) error {
	feeds := make([]*model.Feed, 0, len(req.Feeds))
	for _, r := range req.Feeds {
		feeds = append(feeds, &model.Feed{
			Name:    r.Name,
			Link:    r.Link,
			GroupID: req.GroupID,
		})
	}
	if len(feeds) == 0 {
		return nil
	}

	if err := f.feedRepo.Create(feeds); err != nil {
		if errors.Is(err, repo.ErrDuplicatedKey) {
			err = NewBizError(err, http.StatusBadRequest, "link is not allowed to be the same as other feeds")
		}
		return err
	}

	puller := pull.NewPuller(repo.NewFeed(repo.DB), repo.NewItem(repo.DB))
	if len(feeds) >= 1 {
		go func() {
			routinePool := make(chan struct{}, 10)
			defer close(routinePool)
			wg := sync.WaitGroup{}
			for _, feed := range feeds {
				routinePool <- struct{}{}
				wg.Add(1)
				go func() {
					// NOTE: do not use the incoming ctx, as it will be Done() automatically
					// by api timeout middleware
					puller.PullOne(context.Background(), feed.ID)
					<-routinePool
					wg.Done()
				}()
			}
			wg.Wait()
		}()
		return nil
	}
	return puller.PullOne(ctx, feeds[0].ID)
}

func (f Feed) CheckValidity(ctx context.Context, req *ReqFeedCheckValidity) (*RespFeedCheckValidity, error) {
	validLinks := make([]ValidityItem, 0)
	parsed, err := pull.Fetch(ctx, req.Link)
	if err == nil && parsed != nil {
		validLinks = append(validLinks, ValidityItem{
			Title: &parsed.Title,
			Link:  &req.Link,
		})
	} else {
		target, err := url.Parse(req.Link)
		if err != nil {
			return nil, err
		}
		sniffed, err := sniff.Sniff(ctx, target)
		if err != nil {
			return nil, err
		}
		for _, l := range sniffed {
			validLinks = append(validLinks, ValidityItem{
				Title: &l.Title,
				Link:  &l.Link,
			})
		}
	}
	return &RespFeedCheckValidity{
		FeedLinks: validLinks,
	}, nil
}

func (f Feed) Update(ctx context.Context, req *ReqFeedUpdate) error {
	data := &model.Feed{
		Name:      req.Name,
		Link:      req.Link,
		Suspended: req.Suspended,
	}
	if req.GroupID != nil {
		data.GroupID = *req.GroupID
	}
	err := f.feedRepo.Update(req.ID, data)
	if errors.Is(err, repo.ErrDuplicatedKey) {
		err = NewBizError(err, http.StatusBadRequest, "link is not allowed to be the same as other feeds")
	}
	return err
}

func (f Feed) Delete(ctx context.Context, req *ReqFeedDelete) error {
	// FIX: transaction
	if err := f.itemRepo.DeleteByFeed(req.ID); err != nil && !errors.Is(err, repo.ErrNotFound) {
		return err
	}
	return f.feedRepo.Delete(req.ID)
}

func (f Feed) Refresh(ctx context.Context, req *ReqFeedRefresh) error {
	pull := pull.NewPuller(repo.NewFeed(repo.DB), repo.NewItem(repo.DB))
	if req.ID != nil {
		return pull.PullOne(ctx, *req.ID)
	}
	if req.All != nil && *req.All {
		// NOTE: do not use the incoming ctx, as it will be Done() automatically
		// by api timeout middleware
		go pull.PullAll(context.Background(), true)
	}
	return nil
}
