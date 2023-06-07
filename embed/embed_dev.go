//go:build !production
// +build !production

package embed

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Content serves content in the /embed directory.
var Content fs.FS

func getEmbedDir() string {
	_, path, _, _ := runtime.Caller(0)
	return filepath.Dir(path)
}

func init() {
	// wd, err := os.Getwd()
	// if err != nil {
	// 	log.Panicf("cannot get current directory: %v", err)
	// }

	// embedDir := filepath.Join(wd, "embed")
	embedDir := getEmbedDir()
	stat, err := os.Stat(embedDir)
	if err != nil {
		log.Panicf("cannot stat embed directory: %v", err)
	}
	if !stat.IsDir() {
		log.Panicf("embed directory is not a directory: %s", embedDir)
	}

	log.Printf("[dev] serving embedded content from %s", embedDir)

	Content = os.DirFS(embedDir)
}
