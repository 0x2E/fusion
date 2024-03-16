package pull

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/repo"
)

type FeedRepo interface {
	All() ([]*model.Feed, error)
	Get(id uint) (*model.Feed, error)
	Update(id uint, feed *model.Feed) error
}

type ItemRepo interface {
	Creates(items []*model.Item) error
}

type Puller struct {
	feedRepo FeedRepo
	itemRepo ItemRepo
}

// TODO: cache favicon

func NewPuller(feedRepo FeedRepo, itemRepo ItemRepo) *Puller {
	return &Puller{
		feedRepo: feedRepo,
		itemRepo: itemRepo,
	}
}

var interval = 30 * time.Minute

func (p *Puller) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		p.PullAll(ctx, false)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *Puller) PullAll(ctx context.Context, force bool) error {
	log.Println("start pull-all")
	ctx, cancel := context.WithTimeout(ctx, interval/2)
	defer cancel()
	feeds, err := p.feedRepo.All()
	if err != nil {
		if !errors.Is(err, repo.ErrNotFound) {
			log.Println(err)
		}
		return err
	}
	if len(feeds) == 0 {
		return nil
	}

	routinePool := make(chan struct{}, 10)
	defer close(routinePool)
	wg := sync.WaitGroup{}
	for _, f := range feeds {
		routinePool <- struct{}{}
		wg.Add(1)
		go func(f *model.Feed) {
			defer func() {
				wg.Done()
				<-routinePool
			}()

			if err := p.do(ctx, f, force); err != nil {
				log.Println(err)
			}
		}(f)
	}
	wg.Wait()
	return nil
}

func (p *Puller) PullOne(id uint) error {
	f, err := p.feedRepo.Get(id)
	if err != nil {
		log.Println(err)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return p.do(ctx, f, true)
}
