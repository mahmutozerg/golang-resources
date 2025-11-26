package helper

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func WrapError(err error, msg ...string) error {
	if err == nil {
		return nil
	}

	if len(msg) > 0 {
		return fmt.Errorf("%s: %w", strings.Join(msg, " "), err)
	}

	return err
}

func ValidateNotEmpty(value, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", name)
	}
	return nil
}

func ValidateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return errors.New("url cannot be empty")
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("invalid url: missing scheme or host")
	}

	return nil
}
