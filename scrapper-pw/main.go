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
	})

	if err != nil {
		log.Fatalf("failed to create scrapper: %v", err)
	}

	// Prepare for SIGINT/SIGTERM (Ctrl+C)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start work
	scr.SetupHooks()
	scr.NewPage()
	scr.GoTo("https://www.google.com")

	log.Println("Scrapper running... Press Ctrl+C to stop.")

	<-sigCh

	log.Println("Received signal, cleaning up...")

	if err := scr.Close(); err != nil {
		log.Printf("Close error: %v", err)
	}

	log.Println("Shutdown complete.")
}
