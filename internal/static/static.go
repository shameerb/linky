package static

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var embeddedFiles embed.FS

func Files() fs.FS {
	static, err := fs.Sub(embeddedFiles, "dist")
	if err != nil {
		panic(err)
	}
	return static
}
