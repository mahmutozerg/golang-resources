package config

import (
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var skipExts = map[string]struct{}{
	".pdf": {}, ".doc": {}, ".docx": {}, ".xls": {}, ".xlsx": {}, ".ppt": {}, ".pptx": {},
	".odt": {}, ".ods": {}, ".odp": {}, ".rtf": {}, ".txt": {}, ".csv": {}, ".tsv": {},
	".epub": {}, ".mobi": {}, ".azw3": {}, ".djvu": {},

	".zip": {}, ".rar": {}, ".7z": {}, ".tar": {}, ".gz": {}, ".tgz": {}, ".bz2": {},
	".xz": {}, ".iso": {}, ".dmg": {}, ".img": {}, ".toast": {}, ".vcd": {},

	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {}, ".bmp": {}, ".tiff": {},
	".ico": {}, ".svg": {}, ".heic": {}, ".psd": {}, ".ai": {}, ".raw": {}, ".cr2": {},

	".mp3": {}, ".wav": {}, ".flac": {}, ".aac": {}, ".ogg": {}, ".wma": {}, ".m4a": {}, ".opus": {},

	".mp4": {}, ".mkv": {}, ".avi": {}, ".mov": {}, ".wmv": {}, ".flv": {}, ".webm": {},
	".mpeg": {}, ".mpg": {}, ".m4v": {}, ".3gp": {}, ".ts": {},

	".exe": {}, ".msi": {}, ".apk": {}, ".app": {}, ".deb": {}, ".rpm": {}, ".jar": {},
	".bin": {}, ".sh": {}, ".bat": {}, ".cmd": {}, ".ps1": {}, ".pkg": {},

	".json": {}, ".xml": {}, ".yaml": {}, ".sql": {}, ".db": {}, ".sqlite": {},
	".ttf": {}, ".otf": {}, ".woff": {}, ".woff2": {},

	".dwg": {}, ".dxf": {}, ".stl": {}, ".obj": {}, ".fbx": {}, ".blend": {},
}

func ShouldSkipLink(u *url.URL) bool {

	ext := strings.ToLower(path.Ext(u.Path))

	if _, exists := skipExts[ext]; exists {
		return true
	}

	if ext == "" {
		return isContentTypeFile(u.String())
	}

	return false
}

func isContentTypeFile(link string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		return true
	}

	req.Header.Set("User-Agent", "*")

	resp, err := client.Do(req)
	if err != nil {
		return true
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		return true
	}

	return false
}
