package parse

import (
	"github.com/0x2e/fusion/model"

	"github.com/mmcdole/gofeed"
)

func GoFeedItems(gfItems []*gofeed.Item, feedID uint) []*model.Item {
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
			FeedID:  feedID,
		})
	}

	return items
}
