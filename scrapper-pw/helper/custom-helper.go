package helper

import (
	"fmt"
	"net/url"
	"scrapper/constants"
	"strings"
)

func AssertErrorToNil(err error, message ...string) {
	if err == nil {
		return
	}

	if len(message) > 0 {
		panic(fmt.Sprintf("%s: %s", strings.Join(message, " "), err.Error()))
	}

	panic(err.Error())
}

func AssertNotEmpty(value string, name string) {
	if len(value) == 0 {
		panic(fmt.Sprintf("%s cannot be empty", name))
	}
}

func IsValidURL(u string) error {
	parsed, err := url.ParseRequestURI(u)
	if err != nil {
		return err
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf(constants.InvalidUrl)
	}

	return nil
}
