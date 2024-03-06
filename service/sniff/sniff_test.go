package sniff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatLinkToAbs(t *testing.T) {
	type item struct {
		base string
		link string
		want string
	}
	table := []item{
		{base: "https://x.xx", link: "https://1.xx", want: "https://1.xx"},
		{base: "https://x.xx", link: "", want: "https://x.xx"},
		{base: "https://x.xx/1/", link: "/x/index.xml", want: "https://x.xx/x/index.xml"},
		{base: "https://x.xx/1/", link: "x/index.xml", want: "https://x.xx/1/x/index.xml"},
		{base: "https://x.xx/1", link: "index.xml", want: "https://x.xx/index.xml"},
	}

	for _, tt := range table {
		res := formatLinkToAbs(tt.base, tt.link)
		assert.Equal(t, tt.want, res)
	}
}
