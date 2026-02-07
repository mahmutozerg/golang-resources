package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/config"
	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/fetcher"
	checker "github.com/mahmutozerg/golang-resources/system_design/crawler/internal/robot"
	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/storage"
	"github.com/playwright-community/playwright-go"
	"golang.org/x/time/rate"
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

	fmt.Printf("Seed URLs: %v \n", seedUrls)

	pwi, err := fetcher.New(fetcher.CustomBrowserTypeOptions{
		LaunchOptions:  playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(false)},
		OnlySameOrigin: true,
	})

	if err != nil {
		log.Fatalf("Failed to Create Fetcher Instance %v : ", err)
	}
	defer pwi.Close()
	robotChecker := checker.New("*", func(u string) ([]byte, error) {
		return pwi.FetchRobotsContent(u)
	})

	if allowed, _ := robotChecker.IsAllowed(seedUrls[0]); !allowed {
		log.Printf("Seed Url is disalloweb in robots txt.")
		return
	}

	fmt.Println("Visiting Seed URL...")
	err = pwi.GoTo(seedUrls[0].String(), fetcher.CustomGotoOptions{
		GotoOptions: playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		},
		AllowInsecureConnections: false,
		SessionType:              fetcher.Status(fetcher.CDPSession),
	})
	if err != nil {
		log.Fatalf("Seed URL Error: %v", err)
	}

	linksTovisit := StartLocateLinks(pwi, seedUrls[0], 200)
	if len(linksTovisit) == 0 {
		log.Println("No link have found, Closing.")
		return
	}

	fmt.Printf("Found %d link Total. Starting to Download...\n", len(linksTovisit))
	var visitWg *sync.WaitGroup = new(sync.WaitGroup)
	sem := make(chan struct{}, 5)

	limiter := rate.NewLimiter(rate.Limit(1), 1)

	t := time.Now()
	for c, j := range linksTovisit {

		parse, err := url.Parse(j)
		if err != nil {
			log.Printf("Failed to parse %v to visit entry: %v ", j, err)
			continue
		}
		allowed, ts := robotChecker.IsAllowed(parse)

		fmt.Println(ts)
		if !allowed {
			log.Printf("Robots.txt blockage: %s skipped.", j)
			continue
		}

		if ts > 0 {

			newLimit := rate.Every(ts)

			if limiter.Limit() > newLimit {
				fmt.Printf("Slowing down to robots.txt speed: %v \n", ts)
				limiter.SetLimit(newLimit)
			}
		}

		if err := limiter.Wait(context.Background()); err != nil {
			log.Printf("Limiter hatası: %v", err)
			continue
		}
		jitter := time.Duration(rand.Intn(2000)+500) * time.Millisecond
		time.Sleep(jitter)

		outDir := storage.CreateOutDir("../../files", parse)
		filename := filepath.Join(outDir, time.Now().UTC().Format("20060102T150405")+".mhtml")

		sem <- struct{}{}
		visitWg.Add(1)
		go func(targetUrl string, c int, targetFile string) {
			defer visitWg.Done()
			defer func() { <-sem }()
			mhtml, err := pwi.FetchMHTML(targetUrl)
			if err != nil {
				log.Printf("Error (%s): %v", targetUrl, err)
				return
			}

			err = os.WriteFile(targetFile, mhtml, 0644)

			if err != nil {
				log.Printf("Disk Write Error: %v ", err)
				return
			}

			fmt.Printf("[%d] saved: %s\n", c, targetUrl)

		}(j, c, filename)
		if c >= 4 {
			fmt.Println("Test limit (4) exceeded, stopping...")
			break
		}

	}
	visitWg.Wait()
	fmt.Printf("%v", time.Since(t))
}
