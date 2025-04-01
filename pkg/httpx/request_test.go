package httpx_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/httpx"
	"github.com/0x2e/fusion/pkg/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSendRequestFn is a mock implementation of httpx.SendHTTPRequestFn.
type mockSendRequestFn struct {
	response    *http.Response
	err         error
	capturedReq *http.Request
}

func (m *mockSendRequestFn) Do(req *http.Request) (*http.Response, error) {
	m.capturedReq = req
	return m.response, m.err
}

func TestFusionRequestWithRequestSender(t *testing.T) {
	for _, tt := range []struct {
		description          string
		link                 string
		options              model.FeedRequestOptions
		mockResponse         *http.Response
		mockErr              error
		expectedErrMsg       string
		expectedExtraHeaders map[string]string
		ctx                  context.Context
	}{
		{
			description: "successful request",
			link:        "https://example.com/feed.xml",
			options:     model.FeedRequestOptions{},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
			},
			mockErr:        nil,
			expectedErrMsg: "",
			ctx:            context.Background(),
		},
		{
			description: "adds If-Modified-Since header when caller specifies a LastBuild time",
			link:        "https://example.com/feed.xml",
			options: model.FeedRequestOptions{
				LastModified: ptr.To("Wed, 21 Oct 2015 07:28:00 GMT"),
			},
			mockResponse: &http.Response{
				StatusCode: http.StatusNotModified,
				Status:     http.StatusText(http.StatusNotModified),
			},
			mockErr:        nil,
			expectedErrMsg: "",
			expectedExtraHeaders: map[string]string{
				"If-Modified-Since": "Wed, 21 Oct 2015 07:28:00 GMT",
			},
			ctx: context.Background(),
		},
		{
			description:    "handles error from request sender",
			link:           "https://example.com/feed.xml",
			options:        model.FeedRequestOptions{},
			mockResponse:   nil,
			mockErr:        errors.New("connection refused"),
			expectedErrMsg: "connection refused",
			ctx:            context.Background(),
		},
		{
			description:    "handles canceled context",
			link:           "https://example.com/feed.xml",
			options:        model.FeedRequestOptions{},
			mockResponse:   nil,
			mockErr:        context.Canceled,
			expectedErrMsg: "context canceled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel the context immediately
				return ctx
			}(),
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			mockSender := &mockSendRequestFn{
				response: tt.mockResponse,
				err:      tt.mockErr,
			}

			resp, err := httpx.FusionRequestWithRequestSender(tt.ctx, mockSender.Do, tt.link, tt.options)

			if tt.expectedErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponse, resp)
			}

			assert.Equal(t, "GET", mockSender.capturedReq.Method)
			assert.Equal(t, httpx.UserAgentString, mockSender.capturedReq.Header.Get("User-Agent"))
			for expectedHeader, expectedHeaderValue := range tt.expectedExtraHeaders {
				assert.Equal(t, expectedHeaderValue, mockSender.capturedReq.Header.Get(expectedHeader), "wrong value for %s HTTP header", expectedHeader)
			}
			assert.True(t, mockSender.capturedReq.Close)
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
