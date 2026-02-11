package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist fallback/index.html
var assets embed.FS

func FrontendFS() (fs.FS, bool, error) {
	if _, err := fs.Stat(assets, "dist/index.html"); err == nil {
		frontendFS, subErr := fs.Sub(assets, "dist")
		if subErr != nil {
			return nil, false, subErr
		}
		return frontendFS, true, nil
	}

	fallbackFS, err := fs.Sub(assets, "fallback")
	if err != nil {
		return nil, false, err
	}

	return fallbackFS, false, nil
}
