package server

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/repo"
	"github.com/0x2e/fusion/service/pull"
	"github.com/0x2e/fusion/service/sniff"

	"gorm.io/gorm"
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

func (f Feed) All() (*RespFeedAll, error) {
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

func (f Feed) Get(req *ReqFeedGet) (*RespFeedGet, error) {
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

func (f Feed) Create(req *ReqFeedCreate) error {
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
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = errors.New("link is not allowed to be the same as other feeds")
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
					puller.PullOne(feed.ID)
					<-routinePool
					wg.Done()
				}()
			}
			wg.Wait()
		}()
		return nil
	}
	return puller.PullOne(feeds[0].ID)
}

func (f Feed) CheckValidity(req *ReqFeedCheckValidity) (*RespFeedCheckValidity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

func (f Feed) Update(req *ReqFeedUpdate) error {
	data := &model.Feed{
		Name:      req.Name,
		Link:      req.Link,
		Suspended: req.Suspended,
	}
	if req.GroupID != nil {
		data.GroupID = *req.GroupID
	}
	err := f.feedRepo.Update(req.ID, data)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		err = errors.New("link is not allowed to be the same as other feeds")
	}
	return err
}

func (f Feed) Delete(req *ReqFeedDelete) error {
	// FIX: transaction
	if err := f.itemRepo.DeleteByFeed(req.ID); err != nil {
		return err
	}
	return f.feedRepo.Delete(req.ID)
}

func (f Feed) Refresh(req *ReqFeedRefresh) error {
	pull := pull.NewPuller(repo.NewFeed(repo.DB), repo.NewItem(repo.DB))
	if req.ID != nil {
		return pull.PullOne(*req.ID)
	}
	if req.All != nil && *req.All {
		go pull.PullAll(context.Background())
	}
	return nil
}
