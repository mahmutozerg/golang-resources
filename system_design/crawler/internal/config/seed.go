package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// LoadSeeds reads URL in the given file
// path: "./seed.txt" is relative to the caller
func LoadSeeds(filePath string) ([]*url.URL, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("seed dosyası açılamadı: %w", err)
	}
	defer f.Close()

	var urls []*url.URL
	scanner := bufio.NewScanner(f)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		u, err := url.Parse(line)
		if err != nil {
			fmt.Printf("Warning: Row %d is invalid URL: %s\n", lineNumber, line)
			continue
		}

		urls = append(urls, u)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("File is Valid but no URL Found")
	}

	return urls, nil
}
