//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	pwinit "scrapper/PwInit"
)

func main() {

	opt := pwinit.CustomInstallOptions{
		Skip:     true,
		Headless: false,
		Ep:       "/usr/bin/brave",
	}
	broserState := pwinit.Init(opt)

	if _, err := broserState.Page.Goto("https://www.google.com"); err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	fmt.Println(broserState.Page.Content())

}
