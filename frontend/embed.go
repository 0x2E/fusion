package frontend

import (
	"embed"
	"io/fs"
)

// This embed only used in production. Only APIs are used in development,
// so no error

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
