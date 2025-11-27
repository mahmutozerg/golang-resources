package main

import (
	"log"
	"scrapper/helper"
	"scrapper/scrapper"
)

func main() {
	// Get available browsers
	browsers := helper.GetAvailableBrowsers()

	// Create scrapper
	scr, err := scrapper.NewScrapper(scrapper.ScrapperOptions{
		Bwp:            browsers,
		CreateRespFile: true,
		CreateReqFile:  true,
	})

	if err != nil {
		log.Fatalf("failed to create scrapper: %v", err)
	}
	defer scr.Close()

	scr.SetupHooks()
	scr.NewPage()

	scr.GoTo("https://www.google.com")

}
