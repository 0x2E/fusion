package pull_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/service/pull"
)

// mockReadCloser is a mock io.ReadCloser that can return either data or an error.
type mockReadCloser struct {
	result string
	errMsg string
	reader *strings.Reader
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.errMsg != "" {
		return 0, errors.New(m.errMsg)
	}
	if m.reader == nil {
		m.reader = strings.NewReader(m.result)
	}
	return m.reader.Read(p)
}

func (m *mockReadCloser) Close() error {
	return nil
}

type mockHTTPClient struct {
	resp        *http.Response
	err         error
	lastFeedURL string
	lastOptions *model.FeedRequestOptions
}

func (m *mockHTTPClient) Get(ctx context.Context, link string, options *model.FeedRequestOptions) (*http.Response, error) {
	// Store the last feed URL and options for assertions.
	m.lastFeedURL = link
	m.lastOptions = options

	if m.err != nil {
		return nil, m.err
	}

	return m.resp, nil
}

func TestFeedClientFetch(t *testing.T) {
	for _, tt := range []struct {
		description        string
		feedURL            string
		options            *model.FeedRequestOptions
		httpRespBody       string
		httpStatusCode     int
		httpErrMsg         string
		httpBodyReadErrMsg string
		expectedFeed       *gofeed.Feed
		expectedErrMsg     string
	}{
		{
			description: "fetch succeeds when HTTP request and RSS parse succeed",
			feedURL:     "https://example.com/feed.xml",
			options:     &model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErrMsg:         "",
			httpBodyReadErrMsg: "",
			expectedFeed: &gofeed.Feed{
				Title:       "Test Feed",
				FeedType:    "rss",
				FeedVersion: "2.0",
				Items: []*gofeed.Item{
					{
						Title: "Test Item",
						Link:  "https://example.com/item",
						Links: []string{"https://example.com/item"},
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description: "fetch succeeds with default behavior when options are nil",
			feedURL:     "https://example.com/feed.xml",
			options:     nil,
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErrMsg:         "",
			httpBodyReadErrMsg: "",
			expectedFeed: &gofeed.Feed{
				Title:       "Test Feed",
				FeedType:    "rss",
				FeedVersion: "2.0",
				Items: []*gofeed.Item{
					{
						Title: "Test Item",
						Link:  "https://example.com/item",
						Links: []string{"https://example.com/item"},
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description: "fetch succeeds when using configured proxy server",
			feedURL:     "https://example.com/feed.xml",
			options: &model.FeedRequestOptions{
				ReqProxy: func() *string { s := "http://proxy.example.com:8080"; return &s }(),
			},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed via Proxy</title>
    <item>
      <title>Test Item via Proxy</title>
      <link>https://example.com/proxy-item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErrMsg:         "",
			httpBodyReadErrMsg: "",
			expectedFeed: &gofeed.Feed{
				Title:       "Test Feed via Proxy",
				FeedType:    "rss",
				FeedVersion: "2.0",
				Items: []*gofeed.Item{
					{
						Title: "Test Item via Proxy",
						Link:  "https://example.com/proxy-item",
						Links: []string{"https://example.com/proxy-item"},
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description:        "fetch fails when HTTP request returns connection error",
			feedURL:            "https://example.com/feed.xml",
			options:            &model.FeedRequestOptions{},
			httpRespBody:       "",
			httpStatusCode:     0, // No status code since request errors
			httpErrMsg:         "connection refused",
			httpBodyReadErrMsg: "",
			expectedFeed:       nil,
			expectedErrMsg:     "connection refused",
		},
		{
			description:        "fetch fails when HTTP response has non-200 status code",
			feedURL:            "https://example.com/feed.xml",
			options:            &model.FeedRequestOptions{},
			httpRespBody:       "",
			httpStatusCode:     http.StatusNotFound,
			httpErrMsg:         "",
			httpBodyReadErrMsg: "",
			expectedFeed:       nil,
			expectedErrMsg:     "got status code 404",
		},
		{
			description:        "fetch fails when HTTP response body cannot be read",
			feedURL:            "https://example.com/feed.xml",
			options:            &model.FeedRequestOptions{},
			httpRespBody:       "",
			httpStatusCode:     http.StatusOK,
			httpErrMsg:         "",
			httpBodyReadErrMsg: "mock body read error",
			expectedFeed:       nil,
			expectedErrMsg:     "mock body read error",
		},
		{
			description: "fetch fails when RSS content cannot be parsed",
			feedURL:     "https://example.com/feed.xml",
			options:     &model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<invalid>
  <malformed>
    <content>This is not a valid RSS feed</content>
  </malformed>
</invalid>`,
			httpStatusCode:     http.StatusOK,
			httpErrMsg:         "",
			httpBodyReadErrMsg: "",

			expectedFeed:   nil,
			expectedErrMsg: "Failed to detect feed type",
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			body := &mockReadCloser{
				result: tt.httpRespBody,
				errMsg: tt.httpBodyReadErrMsg,
			}

			httpClient := &mockHTTPClient{
				resp: &http.Response{
					StatusCode: tt.httpStatusCode,
					Status:     http.StatusText(tt.httpStatusCode),
					Body:       body,
				},
				err: func() error {
					if tt.httpErrMsg != "" {
						return errors.New(tt.httpErrMsg)
					}
					return nil
				}(),
			}

			actualFeed, actualErr := pull.NewFeedClient(httpClient.Get).Fetch(context.Background(), tt.feedURL, tt.options)

			if tt.expectedErrMsg != "" {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, actualErr)
			}

			assert.Equal(t, tt.expectedFeed, actualFeed)

			// Verify that the HTTP client received the correct URL.
			assert.Equal(t, tt.feedURL, httpClient.lastFeedURL, "Incorrect feed URL used")

			// Verify that the HTTP client received the correct options.
			if tt.options == nil {
				assert.Nil(t, httpClient.lastOptions, "Expected nil options")
			} else {
				assert.Equal(t, *tt.options, *httpClient.lastOptions, "Incorrect HTTP request options")
			}
		})
	}
}
