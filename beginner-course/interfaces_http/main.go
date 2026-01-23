package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type logWriter struct{}

func main() {

	resp, err := http.Get("https://www.google.com")

	if err != nil {
		log.Fatalf("Something went wrong while executing get %v", err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Status code not 200 %v", resp.StatusCode)
	}

	io.Copy(os.Stdout, resp.Body)

	lw := logWriter{}

	io.Copy(lw, resp.Body)

}

func (logWriter) Write(bs []byte) (int, error) {

	_, err := fmt.Println(string(bs))

	return len(bs), err
}
