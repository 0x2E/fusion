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

func (p *Puller) do(ctx context.Context, f *model.Feed) error {
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
	newItemsCount := 0
	isLatestBuild := f.LastBuild != nil && fetched.UpdatedParsed != nil &&
		fetched.UpdatedParsed.Equal(*f.LastBuild)
	if len(fetched.Items) != 0 && !isLatestBuild {
		newItems, err := filterOutNewItems(f.ID, fetched, p.itemRepo)
		if err != nil {
			return err
		}
		newItemsCount = len(newItems)
		if len(newItems) != 0 {
			if err := p.itemRepo.Creates(newItems); err != nil {
				return err
			}
		}
	}
	log.Printf("fetched: %d, new: %d\n", len(fetched.Items), newItemsCount)
	return p.feedRepo.Update(f.ID, &model.Feed{
		LastBuild: fetched.PublishedParsed,
		Failure:   &failure,
	})
}

func Fetch(ctx context.Context, link string) (*gofeed.Feed, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
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

func filterOutNewItems(feedID uint, fetched *gofeed.Feed, r ItemRepo) ([]*model.Item, error) {
	newItems := make([]*model.Item, 0)
	for _, fetchedItem := range fetched.Items {
		exist, err := r.IdentityExist(feedID, fetchedItem.GUID, fetched.Link, fetched.Title)
		if err != nil {
			log.Println(err)
			continue
		}
		if exist {
			continue
		}
		unread := true
		content := fetchedItem.Content
		if content == "" {
			content = fetchedItem.Description
		}
		newItems = append(newItems, &model.Item{
			Title:   &fetchedItem.Title,
			GUID:    &fetchedItem.GUID,
			Link:    &fetchedItem.Link,
			Content: &content,
			PubDate: fetchedItem.PublishedParsed,
			Unread:  &unread,
			FeedID:  feedID,
		})
	}

	return newItems, nil
}
