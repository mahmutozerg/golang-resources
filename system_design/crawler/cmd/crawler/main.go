package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/config"
	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/fetcher"
	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/storage"
	"github.com/playwright-community/playwright-go"
)

func Must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func StartLocateLinks(pwi *fetcher.PwInstance, baseUrl *url.URL, chSize int) []string {

	urlCh := make(chan string, chSize)
	errCh := make(chan error, chSize)
	var linksToVisit []string

	visited := make(map[string]bool)
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	func() {
		wg.Add(1)
		go pwi.LocateLinks(baseUrl, urlCh, errCh, wg)
	}()

	go func() {
		wg.Wait()
		close(errCh)
		close(urlCh)
	}()

	// Todo implement channel based solution instead of returning the string
	// We'll implement it after we migrate to multi threaded crawler
inf_loop:
	for {
		select {
		case i, ok := <-urlCh:
			if !ok {
				break inf_loop
			}
			if !visited[i] {
				visited[i] = true
				linksToVisit = append(linksToVisit, i)
			}
		case j, ok := <-errCh:
			if !ok {
				break inf_loop

			}
			fmt.Printf("Something went wrong while fetchin urls: %v\n", j)
		}
	}

	return linksToVisit

}
func main() {

	seedUrls, err := config.LoadSeeds("./seed.txt")
	Must(err, "Failed to Load Seeds %v: ")

	fmt.Printf("%v \n", seedUrls)
	pwi, err := fetcher.New(fetcher.CustomBrowserTypeOptions{
		LaunchOptions:  playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(false)},
		OnlySameOrigin: true,
	})

	if err != nil {
		log.Fatalf("Failed to Create Fetcher Instance %v : ", err)
	}
	defer pwi.Close()

	err = pwi.GoTo(seedUrls[0].String(), fetcher.CustomGotoOptions{
		GotoOptions: playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		},
		AllowInsecureConnections: false,
		SessionType:              fetcher.Status(fetcher.CDPSession),
	})

	linksTovisit := StartLocateLinks(pwi, seedUrls[0], 200)

	if len(linksTovisit) == 0 {
		log.Println("No link have found, Closing.")
		return
	}

	for c, j := range linksTovisit {
		parse, err := url.Parse(j)
		if err != nil {
			log.Printf("Failed to parse links to visit entry: %v ", err)
			continue
		}
		outDir := storage.CreateOutDir("../../files", parse)
		filename := path.Join(outDir, time.Now().UTC().Format("20060102T150405")+".mhtml")

		mhtml, err := pwi.FetchMHTML(j)
		if err != nil {
			log.Printf("%v", err)
			continue
		}

		err = os.WriteFile(filename, mhtml, 0644)

		if err != nil {
			log.Printf("%v ", err)
			continue
		}
		if c == 3 {
			break
		}
	}

}
