package storage

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/config"
)

func CreateOutDir(root string, baseUrl *url.URL) string {
	outDir := filepath.Join(root, baseUrl.Host, baseUrl.Path)
	os.MkdirAll(outDir, config.OutputFolderPerm)

	return outDir
}
