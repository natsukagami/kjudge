//go:build production
// +build production

package embed

import (
	"embed"
	"io/fs"
)

var (
	//go:embed assets/* templates/*
	content embed.FS

	Content fs.FS = content
)
