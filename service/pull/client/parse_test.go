package client_test

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/ptr"
	"github.com/0x2e/fusion/service/pull/client"
)

func TestParseGoFeedItems(t *testing.T) {
	for _, tt := range []struct {
		description string
		feedURL     string
		gfItems     []*gofeed.Item
		expected    []*model.Item
	}{
		{
			description: "converts gofeed items to model items with complete data",
			feedURL:     "https://example.com/feed",
			gfItems: []*gofeed.Item{
				{
					Title:           "Test Item",
					GUID:            "https://example.com/guid",
					Link:            "https://example.com/link",
					Image:           &gofeed.Image{
						URL: "https://example.com/image.png",
					},
					Content:         "<p>This is the content</p>",
					Description:     "This is the description",
					PublishedParsed: mustParseTime("2025-01-01T12:00:00Z"),
				},
			},
			expected: []*model.Item{
				{
					Title:   ptr.To("Test Item"),
					GUID:    ptr.To("https://example.com/guid"),
					Link:    ptr.To("https://example.com/link"),
					Image:   ptr.To("https://example.com/image.png"),
					Content: ptr.To("<p>This is the content</p>"),
					PubDate: mustParseTime("2025-01-01T12:00:00Z"),
					Unread:  ptr.To(true),
				},
			},
		},
		{
			description: "converts relative path links to full URLs",
			feedURL:     "https://example.com/feed",
			gfItems: []*gofeed.Item{
				{
					Title:           "Test Item with Relative Path",
					Link:            "/link",
					Image:           &gofeed.Image{
						URL: "/image.png",
					},
					GUID:            "guid",
					Content:         "<p>This is the content</p>",
					Description:     "This is the description",
					PublishedParsed: mustParseTime("2025-01-01T12:00:00Z"),
				},
			},
			expected: []*model.Item{
				{
					Title:   ptr.To("Test Item with Relative Path"),
					Link:    ptr.To("https://example.com/link"),
					Image:   ptr.To("https://example.com/image.png"),
					GUID:    ptr.To("guid"),
					Content: ptr.To("<p>This is the content</p>"),
					PubDate: mustParseTime("2025-01-01T12:00:00Z"),
					Unread:  ptr.To(true),
				},
			},
		},
		{
			description: "uses description when content is empty",
			feedURL:     "https://example.com/feed",
			gfItems: []*gofeed.Item{
				{
					Title:           "Test Item",
					GUID:            "https://example.com/guid",
					Link:            "https://example.com/link",
					Content:         "", // Empty content
					Description:     "This is the description",
					PublishedParsed: mustParseTime("2025-01-01T12:00:00Z"),
				},
			},
			expected: []*model.Item{
				{
					Title:   ptr.To("Test Item"),
					GUID:    ptr.To("https://example.com/guid"),
					Link:    ptr.To("https://example.com/link"),
					Image:   ptr.To(""),
					Content: ptr.To("This is the description"), // Should use description
					PubDate: mustParseTime("2025-01-01T12:00:00Z"),
					Unread:  ptr.To(true),
				},
			},
		},
		{
			description: "uses link when GUID is empty",
			feedURL:     "https://example.com/feed",
			gfItems: []*gofeed.Item{
				{
					Title:           "Test Item",
					GUID:            "", // Empty GUID
					Link:            "https://example.com/link",
					Content:         "<p>This is the content</p>",
					Description:     "This is the description",
					PublishedParsed: mustParseTime("2025-01-01T12:00:00Z"),
				},
			},
			expected: []*model.Item{
				{
					Title:   ptr.To("Test Item"),
					GUID:    ptr.To("https://example.com/link"), // Should use link
					Link:    ptr.To("https://example.com/link"),
					Image:   ptr.To(""),
					Content: ptr.To("<p>This is the content</p>"),
					PubDate: mustParseTime("2025-01-01T12:00:00Z"),
					Unread:  ptr.To(true),
				},
			},
		},
		{
			description: "handles both empty content and empty GUID",
			feedURL:     "https://example.com/feed",
			gfItems: []*gofeed.Item{
				{
					Title:           "Test Item",
					GUID:            "", // Empty GUID
					Link:            "https://example.com/link",
					Content:         "", // Empty content
					Description:     "This is the description",
					PublishedParsed: mustParseTime("2025-01-01T12:00:00Z"),
				},
			},
			expected: []*model.Item{
				{
					Title:   ptr.To("Test Item"),
					GUID:    ptr.To("https://example.com/link"), // Should use link
					Link:    ptr.To("https://example.com/link"),
					Image:   ptr.To(""),
					Content: ptr.To("This is the description"), // Should use description
					PubDate: mustParseTime("2025-01-01T12:00:00Z"),
					Unread:  ptr.To(true),
				},
			},
		},
		{
			description: "handles multiple items",
			feedURL:     "https://example.com/feed",
			gfItems: []*gofeed.Item{
				{
					Title:           "Item 1",
					GUID:            "guid1",
					Link:            "link1",
					Content:         "content1",
					Description:     "description1",
					PublishedParsed: mustParseTime("2025-01-01T12:00:00Z"),
				},
				{
					Title:           "Item 2",
					GUID:            "guid2",
					Link:            "link2",
					Content:         "content2",
					Description:     "description2",
					PublishedParsed: mustParseTime("2025-01-01T12:00:00Z"),
				},
			},
			expected: []*model.Item{
				{
					Title:   ptr.To("Item 1"),
					GUID:    ptr.To("guid1"),
					Link:    ptr.To("link1"),
					Image:   ptr.To(""),
					Content: ptr.To("content1"),
					PubDate: mustParseTime("2025-01-01T12:00:00Z"),
					Unread:  ptr.To(true),
				},
				{
					Title:   ptr.To("Item 2"),
					GUID:    ptr.To("guid2"),
					Link:    ptr.To("link2"),
					Image:   ptr.To(""),
					Content: ptr.To("content2"),
					PubDate: mustParseTime("2025-01-01T12:00:00Z"),
					Unread:  ptr.To(true),
				},
			},
		},
		{
			description: "returns empty slice for empty input",
			feedURL:     "https://example.com/feed",
			gfItems:     []*gofeed.Item{},
			expected:    []*model.Item{},
		},
		{
			description: "skips nil items in the array",
			feedURL:     "https://example.com/feed",
			gfItems: []*gofeed.Item{
				{
					Title:           "Valid Item",
					GUID:            "valid-guid",
					Link:            "https://example.com/valid",
					Content:         "valid content",
					PublishedParsed: mustParseTime("2025-01-01T12:00:00Z"),
				},
				nil, // Nil item that should be skipped
				{
					Title:           "Another Valid Item",
					GUID:            "another-guid",
					Link:            "https://example.com/another",
					Content:         "another content",
					PublishedParsed: mustParseTime("2025-01-01T12:00:00Z"),
				},
			},
			expected: []*model.Item{
				{
					Title:   ptr.To("Valid Item"),
					GUID:    ptr.To("valid-guid"),
					Link:    ptr.To("https://example.com/valid"),
					Image:   ptr.To(""),
					Content: ptr.To("valid content"),
					PubDate: mustParseTime("2025-01-01T12:00:00Z"),
					Unread:  ptr.To(true),
				},
				{
					Title:   ptr.To("Another Valid Item"),
					GUID:    ptr.To("another-guid"),
					Link:    ptr.To("https://example.com/another"),
					Image:   ptr.To(""),
					Content: ptr.To("another content"),
					PubDate: mustParseTime("2025-01-01T12:00:00Z"),
					Unread:  ptr.To(true),
				},
			},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			result := client.ParseGoFeedItems(tt.feedURL, tt.gfItems)
			assert.Equal(t, tt.expected, result)
		})
	}
}
