package pull

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/logx"
	"github.com/0x2e/fusion/repo"
)

var (
	interval   = 30 * time.Minute
	pullLogger = logx.Logger.With("module", "puller")
)

type FeedRepo interface {
	List(filter *repo.FeedListFilter) ([]*model.Feed, error)
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

func (p *Puller) Run() {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		p.PullAll(context.Background(), false)

		<-ticker.C
	}
}

func (p *Puller) PullAll(ctx context.Context, force bool) error {
	logger := logx.LoggerFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, interval/2)
	defer cancel()

	feeds, err := p.feedRepo.List(nil)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			err = nil
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

			logger := logger.With("feed_id", f.ID)
			if err := p.do(ctx, f, force); err != nil {
				logger.Errorln(err)
				if _, ok := err.(net.Error); ok {
					for i := 1; i < 4; i++ {
						logger.Infof("%dth retry", i)
						if p.do(ctx, f, true) == nil {
							break
						}
					}
				}
			}
		}(f)
	}
	wg.Wait()
	return nil
}

func (p *Puller) PullOne(ctx context.Context, id uint) error {
	f, err := p.feedRepo.Get(id)
	if err != nil {
		return err
	}

	return p.do(ctx, f, true)
}
