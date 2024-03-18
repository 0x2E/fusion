package pull

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/httpx"

	"github.com/mmcdole/gofeed"
)

func (p *Puller) do(ctx context.Context, f *model.Feed, force bool) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if f.IsSuspended() {
		log.Printf("skip feed %d: suspended\n", f.ID)
		return nil
	}
	if !force {
		if f.IsFailed() {
			log.Printf("skip feed %d: failure exists\n", f.ID)
			return nil
		}
		if time.Since(f.UpdatedAt) < interval {
			log.Printf("skip feed %d: new enough\n", f.ID)
			return nil
		}
	}

	log.Printf("start pull %d", f.ID)
	failure := ""
	fetched, err := Fetch(ctx, *f.Link)
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
		if err := p.itemRepo.Creates(data); err != nil {
			return err
		}
	}
	log.Printf("fetched: %d items\n", len(fetched.Items))
	return p.feedRepo.Update(f.ID, &model.Feed{
		LastBuild: fetched.PublishedParsed,
		Failure:   &failure,
	})
}

func Fetch(ctx context.Context, link string) (*gofeed.Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpx.NewSafeClient().Do(req)
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
