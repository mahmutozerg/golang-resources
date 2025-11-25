package main

import (
	"scrapper/helper"
	"scrapper/scrapper"
)

func main() {

	scr, err := scrapper.NewScrapper()
	helper.AssertErrorToNil("scrapper failed", err)
	defer scr.Close()

}
