package main

import (
	"log"
	"os"
	"os/signal"
	"scrapper/helper"
	"scrapper/scrapper"
	"syscall"
)

func main() {
	// Get available browsers
	browsers := helper.GetAvailableBrowsers()

	// Create scrapper
	scr, err := scrapper.NewScrapper(scrapper.ScrapperOptions{
		Bwp:            browsers,
		CreateRespFile: true,
		CreateReqFile:  true,
		UrlScrapperOptions: scrapper.UrlScrapperOptions{
			FollowRedirects: false,
			MaxDepth:        20, // max 20
		},
	})

	if err != nil {
		log.Fatalf("failed to create scrapper: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	scr.SetupHooks()
	scr.NewPage()
	scr.GoTo("https://www.google.com")

	scr.CollectUrls("https://www.youtube.com/watch?v=7fGB-hjc2Gc&t=801s")
	log.Println("Scrapper running... Press Ctrl+C to stop.")

	<-sigCh

	log.Println("Received signal, cleaning up...")

	if err := scr.Close(); err != nil {
		log.Printf("Close error: %v", err)
	}

	log.Println("Shutdown complete.")
}
