package client_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/ptr"
	"github.com/0x2e/fusion/service/pull/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func (m *mockHTTPClient) Get(ctx context.Context, link string, options model.FeedRequestOptions) (*http.Response, error) {
	// Store the last feed URL and options for assertions.
	m.lastFeedURL = link
	m.lastOptions = &options

	if m.err != nil {
		return nil, m.err
	}

	return m.resp, nil
}

func TestFeedClientFetchTitle(t *testing.T) {
	for _, tt := range []struct {
		description        string
		feedURL            string
		options            model.FeedRequestOptions
		httpRespBody       string
		httpStatusCode     int
		httpErr            error
		httpBodyReadErrMsg string
		expectedTitle      string
		expectedErrMsg     string
	}{
		{
			description: "fetch title succeeds when HTTP request and RSS parse succeed",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed Title</title>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedTitle:      "Test Feed Title",
			expectedErrMsg:     "",
		},
		{
			description: "fetch title succeeds with default behavior when options are nil",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed Title</title>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedTitle:      "Test Feed Title",
			expectedErrMsg:     "",
		},
		{
			description: "fetch title succeeds when using configured proxy server",
			feedURL:     "https://example.com/feed.xml",
			options: model.FeedRequestOptions{
				ReqProxy: func() *string { s := "http://proxy.example.com:8080"; return &s }(),
			},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed Title via Proxy</title>
    <item>
      <title>Test Item via Proxy</title>
      <link>https://example.com/proxy-item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedTitle:      "Test Feed Title via Proxy",
			expectedErrMsg:     "",
		},
		{
			description:        "fetch title fails when HTTP request returns connection error",
			feedURL:            "https://example.com/feed.xml",
			options:            model.FeedRequestOptions{},
			httpRespBody:       "",
			httpStatusCode:     0, // No status code since request errors
			httpErr:            errors.New("connection refused"),
			httpBodyReadErrMsg: "",
			expectedTitle:      "",
			expectedErrMsg:     "connection refused",
		},
		{
			description:        "fetch title fails when HTTP response has non-200 status code",
			feedURL:            "https://example.com/feed.xml",
			options:            model.FeedRequestOptions{},
			httpRespBody:       "",
			httpStatusCode:     http.StatusNotFound,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedTitle:      "",
			expectedErrMsg:     "got status code 404",
		},
		{
			description:        "fetch title fails when HTTP response body cannot be read",
			feedURL:            "https://example.com/feed.xml",
			options:            model.FeedRequestOptions{},
			httpRespBody:       "",
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "mock body read error",
			expectedTitle:      "",
			expectedErrMsg:     "mock body read error",
		},
		{
			description: "fetch title fails when RSS content cannot be parsed",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<invalid>
  <malformed>
    <content>This is not a valid RSS feed</content>
  </malformed>
</invalid>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedTitle:      "",
			expectedErrMsg:     "Failed to detect feed type",
		},
		{
			description: "fetch title returns empty string when feed has no title",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedTitle:      "",
			expectedErrMsg:     "",
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
				err: tt.httpErr,
			}

			actualTitle, actualErr := client.NewFeedClientWithRequestFn(httpClient.Get).FetchTitle(context.Background(), tt.feedURL, tt.options)

			if tt.expectedErrMsg != "" {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, actualErr)
			}

			assert.Equal(t, tt.expectedTitle, actualTitle)

			assert.Equal(t, tt.feedURL, httpClient.lastFeedURL, "Incorrect feed URL used")
			assert.Equal(t, tt.options, *httpClient.lastOptions, "Incorrect HTTP request options")
		})
	}
}

func TestFeedClientFetchDeclaredLink(t *testing.T) {
	for _, tt := range []struct {
		description        string
		httpRespBody       string
		httpStatusCode     int
		httpErr            error
		httpBodyReadErrMsg string
		expectedLink       string
		expectedErrMsg     string
	}{
		{
			description: "fetch declared link succeeds when HTTP request and RSS parse succeed",
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed Title</title>
    <atom:link href="https://example.com/declared-feed.xml" rel="self" type="application/rss+xml" xmlns:atom="http://www.w3.org/2005/Atom"/>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedLink:       "https://example.com/declared-feed.xml",
			expectedErrMsg:     "",
		},
		{
			description: "fetch declared link from RSS 2.0 feed with standard link element",
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed Title</title>
    <link>http://rss2.example.com/</link>
    <description>A dummy RSS news feed.</description>
    <item>
      <title>Test Item</title>
      <link>http://rss2.example.com/article1</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedLink:       "http://rss2.example.com/",
			expectedErrMsg:     "",
		},
		{
			description: "fetch declared link returns empty string when feed has no link",
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed Title</title>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedLink:       "",
			expectedErrMsg:     "",
		},
		{
			description:        "fetch declared link fails when HTTP request returns connection error",
			httpRespBody:       "",
			httpStatusCode:     0, // No status code since request errors
			httpErr:            errors.New("dummy connection refused error"),
			httpBodyReadErrMsg: "",
			expectedLink:       "",
			expectedErrMsg:     "dummy connection refused error",
		},
		{
			description:        "fetch declared link fails when HTTP response body cannot be read",
			httpRespBody:       "",
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "mock body read error",
			expectedLink:       "",
			expectedErrMsg:     "mock body read error",
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
				err: tt.httpErr,
			}

			// The feedURL and options don't matter in the test because we're mocking
			// out the HTTP functionality, but we just need to make sure the HTTP
			// client receives the right values.
			feedURL := "https://dummy.example.com/rss"
			options := model.FeedRequestOptions{}

			actualLink, actualErr := client.NewFeedClientWithRequestFn(httpClient.Get).FetchDeclaredLink(context.Background(), feedURL, options)

			if tt.expectedErrMsg != "" {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, actualErr)
			}

			assert.Equal(t, tt.expectedLink, actualLink)

			assert.Equal(t, feedURL, httpClient.lastFeedURL, "Incorrect feed URL used")
			assert.Equal(t, options, *httpClient.lastOptions, "Incorrect HTTP request options")
		})
	}
}

func TestFeedClientFetchItems(t *testing.T) {
	for _, tt := range []struct {
		description        string
		feedURL            string
		options            model.FeedRequestOptions
		httpRespBody       string
		httpStatusCode     int
		httpErr            error
		httpBodyReadErrMsg string
		expectedResult     client.FetchItemsResult
		expectedErrMsg     string
	}{
		{
			description: "fetch succeeds with no LastBuild when feed has no updated time",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
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
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedResult: client.FetchItemsResult{
				LastBuild: nil, // UpdatedParsed is nil in this test case
				Items: []*model.Item{
					{
						Title: ptr.To("Test Item"),
						Link:  ptr.To("https://example.com/item"),
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description: "fetch succeeds and populates LastBuild from RSS lastBuildDate",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <lastBuildDate>2025-01-01T12:00:00Z</lastBuildDate>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedResult: client.FetchItemsResult{
				LastBuild: mustParseTime("2025-01-01T12:00:00Z"),
				Items: []*model.Item{
					{
						Title: ptr.To("Test Item"),
						Link:  ptr.To("https://example.com/item"),
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description: "fetch succeeds and populates LastBuild from Atom updated",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Test Feed</title>
  <updated>2025-02-15T15:30:00Z</updated>
  <entry>
    <title>Test Item</title>
    <link href="https://example.com/item"/>
  </entry>
</feed>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedResult: client.FetchItemsResult{
				LastBuild: mustParseTime("2025-02-15T15:30:00Z"),
				Items: []*model.Item{
					{
						Title: ptr.To("Test Item"),
						Link:  ptr.To("https://example.com/item"),
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description: "fetch succeeds with different timezone in lastBuildDate",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <lastBuildDate>2025-01-01T07:00:00-05:00</lastBuildDate>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedResult: client.FetchItemsResult{
				LastBuild: mustParseTime("2025-01-01T12:00:00Z"), // Same time as UTC
				Items: []*model.Item{
					{
						Title: ptr.To("Test Item"),
						Link:  ptr.To("https://example.com/item"),
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description: "fetch succeeds with non-standard time format",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <lastBuildDate>Wed, 01 Jan 2025 12:00:00 GMT</lastBuildDate>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedResult: client.FetchItemsResult{
				LastBuild: mustParseTime("2025-01-01T12:00:00Z"), // Use UTC format since gofeed normalizes to UTC
				Items: []*model.Item{
					{
						Title: ptr.To("Test Item"),
						Link:  ptr.To("https://example.com/item"),
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description: "fetch succeeds with default behavior when options are nil",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <lastBuildDate>2025-01-01T12:00:00Z</lastBuildDate>
    <item>
      <title>Test Item</title>
      <link>https://example.com/item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedResult: client.FetchItemsResult{
				LastBuild: mustParseTime("2025-01-01T12:00:00Z"),
				Items: []*model.Item{
					{
						Title: ptr.To("Test Item"),
						Link:  ptr.To("https://example.com/item"),
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description: "fetch succeeds when using configured proxy server",
			feedURL:     "https://example.com/feed.xml",
			options: model.FeedRequestOptions{
				ReqProxy: func() *string { s := "http://proxy.example.com:8080"; return &s }(),
			},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed via Proxy</title>
    <lastBuildDate>2025-01-01T12:00:00Z</lastBuildDate>
    <item>
      <title>Test Item via Proxy</title>
      <link>https://example.com/proxy-item</link>
    </item>
  </channel>
</rss>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedResult: client.FetchItemsResult{
				LastBuild: mustParseTime("2025-01-01T12:00:00Z"),
				Items: []*model.Item{
					{
						Title: ptr.To("Test Item via Proxy"),
						Link:  ptr.To("https://example.com/proxy-item"),
					},
				},
			},
			expectedErrMsg: "",
		},
		{
			description:        "fetch fails when HTTP request returns connection error",
			feedURL:            "https://example.com/feed.xml",
			options:            model.FeedRequestOptions{},
			httpRespBody:       "",
			httpStatusCode:     0, // No status code since request errors
			httpErr:            errors.New("connection refused"),
			httpBodyReadErrMsg: "",
			expectedResult:     client.FetchItemsResult{},
			expectedErrMsg:     "connection refused",
		},
		{
			description:        "fetch fails when HTTP response has non-200 status code",
			feedURL:            "https://example.com/feed.xml",
			options:            model.FeedRequestOptions{},
			httpRespBody:       "",
			httpStatusCode:     http.StatusNotFound,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedResult:     client.FetchItemsResult{},
			expectedErrMsg:     "got status code 404",
		},
		{
			description:        "fetch fails when HTTP response body cannot be read",
			feedURL:            "https://example.com/feed.xml",
			options:            model.FeedRequestOptions{},
			httpRespBody:       "",
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "mock body read error",
			expectedResult:     client.FetchItemsResult{},
			expectedErrMsg:     "mock body read error",
		},
		{
			description: "fetch fails when RSS content cannot be parsed",
			feedURL:     "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			httpRespBody: `<?xml version="1.0" encoding="UTF-8"?>
<not-a-real-tag>
  <also-a-fake-tag>
    <content>This is not a valid RSS feed</content>
  </also-a-fake-tag>
</not-a-real-tag>`,
			httpStatusCode:     http.StatusOK,
			httpErr:            nil,
			httpBodyReadErrMsg: "",
			expectedResult:     client.FetchItemsResult{},
			expectedErrMsg:     "Failed to detect feed type",
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
				err: tt.httpErr,
			}

			actualResult, actualErr := client.NewFeedClientWithRequestFn(httpClient.Get).FetchItems(context.Background(), tt.feedURL, tt.options)

			if tt.expectedErrMsg != "" {
				require.Error(t, actualErr)
				require.Contains(t, actualErr.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, actualErr)
			}

			if tt.expectedResult.LastBuild != nil {
				require.NotNil(t, actualResult.LastBuild, "LastBuild should not be nil")
				assert.Equal(t, *tt.expectedResult.LastBuild, *actualResult.LastBuild, "LastBuild time doesn't match")
			} else {
				assert.Nil(t, actualResult.LastBuild, "LastBuild should be nil")
			}
			assert.Equal(t, len(tt.expectedResult.Items), len(actualResult.Items))

			if len(tt.expectedResult.Items) > 0 {
				for i, expectedItem := range tt.expectedResult.Items {
					if i < len(actualResult.Items) {
						actualItem := actualResult.Items[i]
						if expectedItem.Title != nil {
							assert.Equal(t, *expectedItem.Title, *actualItem.Title)
						}
						if expectedItem.Link != nil {
							assert.Equal(t, *expectedItem.Link, *actualItem.Link)
						}
					}
				}
			}

			assert.Equal(t, tt.feedURL, httpClient.lastFeedURL, "Incorrect feed URL used")
			assert.Equal(t, tt.options, *httpClient.lastOptions, "Incorrect HTTP request options")
		})
	}
}

// Helper function to parse ISO8601 string to time.Time.
func mustParseTime(iso8601 string) *time.Time {
	t, err := time.Parse(time.RFC3339, iso8601)
	if err != nil {
		panic(err)
	}
	return &t
}
