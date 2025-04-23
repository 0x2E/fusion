package feedfinder

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHub(t *testing.T) {
	type testItem struct {
		url  *url.URL
		want []FeedLink
	}

	urlBase, _ := url.Parse("https://github.com?xxx=1")
	urlUser, _ := url.Parse("https://github.com/user")
	urlUserFeed := genGitHubUserFeed("user")
	urlUserRepo, _ := url.Parse("https://github.com/user/repo")
	urlRepoFeed := genGitHubRepoFeed("user/repo")

	table := []testItem{
		{url: urlBase, want: githubGlobalFeed},
		{url: urlUser, want: urlUserFeed},
		{url: urlUserRepo, want: urlRepoFeed},
	}

	for _, tt := range table {
		finder := Finder{
			target: tt.url,
		}
		feed, err := finder.githubMatcher(context.Background())
		assert.Nil(t, err)
		assert.ElementsMatch(t, tt.want, feed)
	}
}

func TestReddit(t *testing.T) {
	type testItem struct {
		url  *url.URL
		want []FeedLink
	}

	urlBase, _ := url.Parse("https://www.reddit.com")
	urlSub, _ := url.Parse("https://www.reddit.com/r/homelab")
	urlSubFeed := genRedditSubFeed("homelab")
	urlComment, _ := url.Parse("https://www.reddit.com/r/homelab/comments/1234/xxx_xxx_xx/")
	urlCommentFeed := genRedditCommentFeed("https://www.reddit.com/r/homelab/comments/1234/xxx_xxx_xx/")
	urlUser, _ := url.Parse("https://www.reddit.com/user/x")
	urlUserFeed := genRedditUserFeed("x")
	urlDomainSubmission, _ := url.Parse("https://www.reddit.com/domain/github.com/")
	urlDomainSubmissionFeed := genRedditDomainSubmissionFeed("github.com")

	table := []testItem{
		{url: urlBase, want: redditGlobalFeed},
		{url: urlSub, want: urlSubFeed},
		{url: urlComment, want: urlCommentFeed},
		{url: urlUser, want: urlUserFeed},
		{url: urlDomainSubmission, want: urlDomainSubmissionFeed},
	}

	for _, tt := range table {
		finder := Finder{
			target: tt.url,
		}
		feed, err := finder.redditMatcher(context.Background())
		assert.Nil(t, err)
		assert.ElementsMatch(t, tt.want, feed)
	}
}
