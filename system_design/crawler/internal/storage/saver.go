package storage

import (
	"net/url"
	"os"
	"path/filepath"
)

func CreateOutDir(root string, baseUrl *url.URL) string {
	outDir := filepath.Join(root, baseUrl.Host, baseUrl.Path)
	os.MkdirAll(outDir, 0755)

	return outDir
}
