package client

import (
	"github.com/0x2e/fusion/model"

	"github.com/mmcdole/gofeed"
)

func ParseGoFeedItems(gfItems []*gofeed.Item) []*model.Item {
	items := make([]*model.Item, 0, len(gfItems))
	for _, item := range gfItems {
		if item == nil {
			continue
		}

		unread := true
		content := item.Content
		if content == "" {
			content = item.Description
		}
		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}
		items = append(items, &model.Item{
			Title:   &item.Title,
			GUID:    &guid,
			Link:    &item.Link,
			Content: &content,
			PubDate: item.PublishedParsed,
			Unread:  &unread,
		})
	}

	return items
}
