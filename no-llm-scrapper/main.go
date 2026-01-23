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
	browserState, err := pwinit.Init(opt)

	if err != nil {
		log.Fatalf("Something went wrong while init %v", err)
	}
	if _, err := browserState.Page.Goto("https://www.google.com"); err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	fmt.Println(browserState.Page.Content())

}
