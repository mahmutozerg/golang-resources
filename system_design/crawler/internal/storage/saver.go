package storage

import (
	"net/url"
	"os"
	"path"
)

func CreateOutDir(root string, baseUrl *url.URL) string {
	outDir := path.Join(root, baseUrl.Host, baseUrl.Path)
	os.MkdirAll(outDir, 0755)

	return outDir
}
