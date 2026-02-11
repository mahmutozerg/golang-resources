package config

import (
	"net/url"
	"path/filepath"
	"strings"
)

var fileExts = map[string]struct{}{
	".pdf": {}, ".ps": {}, ".eps": {}, ".rtf": {}, ".txt": {}, ".md": {},
	".doc": {}, ".docx": {}, ".odt": {}, ".pages": {},
	".xls": {}, ".xlsx": {}, ".ods": {}, ".csv": {}, ".tsv": {},
	".ppt": {}, ".pptx": {}, ".odp": {}, ".key": {},
	".tex": {}, ".epub": {}, ".mobi": {}, ".azw": {}, ".azw3": {},

	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {}, ".bmp": {}, ".tif": {}, ".tiff": {},
	".svg": {}, ".ico": {}, ".heic": {}, ".heif": {}, ".avif": {}, ".raw": {}, ".cr2": {}, ".nef": {},

	".mp3": {}, ".wav": {}, ".flac": {}, ".aac": {}, ".m4a": {}, ".ogg": {}, ".opus": {}, ".wma": {},

	".mp4": {}, ".m4v": {}, ".mkv": {}, ".webm": {}, ".mov": {}, ".avi": {}, ".wmv": {}, ".flv": {},
	".mpeg": {}, ".mpg": {}, ".3gp": {}, ".ts": {},

	".zip": {}, ".rar": {}, ".7z": {}, ".tar": {}, ".gz": {}, ".tgz": {}, ".bz2": {}, ".xz": {}, ".zst": {},
	".iso": {}, ".dmg": {}, ".img": {},

	".exe": {}, ".msi": {}, ".msp": {}, ".appx": {}, ".appxbundle": {}, ".msix": {}, ".msixbundle": {},
	".deb": {}, ".rpm": {}, ".apk": {}, ".aab": {}, ".ipa": {}, ".pkg": {}, ".sh": {}, ".bat": {}, ".cmd": {},

	".bin": {}, ".dat": {}, ".db": {}, ".sqlite": {}, ".sqlite3": {}, ".parquet": {}, ".orc": {}, ".feather": {},
	".sav": {},

	".ttf": {}, ".otf": {}, ".woff": {}, ".woff2": {}, ".eot": {},

	".pem": {}, ".crt": {}, ".cer": {}, ".der": {}, ".pfx": {}, ".p12": {},

	".jar": {}, ".war": {}, ".ear": {},
	".dylib": {}, ".so": {}, ".dll": {},
	".torrent": {},
}

func IsFileByExtension(u *url.URL) bool {
	p := strings.ToLower(u.Path)

	ext := strings.ToLower(filepath.Ext(p))
	if ext == "" {
		return false
	}
	_, ok := fileExts[ext]
	return ok
}
