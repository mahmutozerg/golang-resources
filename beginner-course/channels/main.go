package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {

	links := []string{
		"https://www.google.com",
		"https://www.facebook.com",
		"https://www.golang.org",
	}

	c := make(chan string)

	for _, link := range links {
		go CheckStatus(link, c)
	}

	for l := range c {

		go func() {
			time.Sleep(time.Second)
			CheckStatus(l, c)

		}()
	}
}

func CheckStatus(url string, c chan string) {

	if _, err := http.Get(url); err != nil {
		c <- url
		return
	}

	fmt.Println(url, "is up")
	c <- url

}
