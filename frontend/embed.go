package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var build embed.FS

var Content fs.FS

func init() {
	sub, err := fs.Sub(build, "build")
	if err != nil {
		panic(err)
	}
	Content = sub
}
