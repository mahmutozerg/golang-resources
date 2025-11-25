package helper

import "fmt"

func AssertErrorToNil(message string, err error) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", message, err.Error()))
	}
}

func AssertNotEmpty(value string, name string) {
	if len(value) == 0 {
		panic(fmt.Sprintf("%s cannot be empty", name))
	}
}
