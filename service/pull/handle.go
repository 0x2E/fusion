package pull

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/httpx"

	"github.com/mmcdole/gofeed"
)

func (p *Puller) do(ctx context.Context, f *model.Feed, force bool) error {
	logger := pullLogger.With("feed_id", f.ID, "feed_name", f.Name)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if f.IsSuspended() {
		logger.Infoln("skip: suspended")
		return nil
	}
	if !force {
		if f.IsFailed() {
			logger.Infoln("skip: failure exists")
			return nil
		}
		if time.Since(f.UpdatedAt) < interval {
			logger.Infoln("skip: new enough")
			return nil
		}
	}

	failure := ""
	fetched, err := FetchFeed(ctx, f)
	if err != nil {
		failure = err.Error()
		p.feedRepo.Update(f.ID, &model.Feed{Failure: &failure})
		return err
	}
	if fetched == nil {
		return nil
	}
	isLatestBuild := f.LastBuild != nil && fetched.UpdatedParsed != nil &&
		fetched.UpdatedParsed.Equal(*f.LastBuild)
	if len(fetched.Items) != 0 && !isLatestBuild {
		data := make([]*model.Item, 0, len(fetched.Items))
		for _, i := range fetched.Items {
			unread := true
			content := i.Content
			if content == "" {
				content = i.Description
			}
			guid := i.GUID
			if guid == "" {
				guid = i.Link
			}
			data = append(data, &model.Item{
				Title:   &i.Title,
				GUID:    &guid,
				Link:    &i.Link,
				Content: &content,
				PubDate: i.PublishedParsed,
				Unread:  &unread,
				FeedID:  f.ID,
			})
		}
		if err := p.itemRepo.Insert(data); err != nil {
			return err
		}
	}
	logger.Infof("fetched %d items", len(fetched.Items))
	return p.feedRepo.Update(f.ID, &model.Feed{
		LastBuild: fetched.PublishedParsed,
		Failure:   &failure,
	})
}

func FetchFeed(ctx context.Context, f *model.Feed) (*gofeed.Feed, error) {
	resp, err := httpx.FusionRequest(ctx, *f.Link, &f.FeedRequestOptions)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return gofeed.NewParser().ParseString(string(data))
}
