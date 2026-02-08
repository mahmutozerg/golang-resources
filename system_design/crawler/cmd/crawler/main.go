package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/config"
	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/fetcher"
	policy "github.com/mahmutozerg/golang-resources/system_design/crawler/internal/policy"
	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/storage"
	"github.com/playwright-community/playwright-go"
	"golang.org/x/time/rate"
)

type VisitedLinks struct {
	Visited map[string]bool
	rwMu    sync.RWMutex
}

func Must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}
func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nSigterm or Interrupt received, please wait for recourses to be freed ")
		cancel()
	}()

	seedUrls, err := config.LoadSeeds("./seed.txt")
	Must(err, "Failed to Load Seeds %v: ")

	fmt.Printf("Seed URLs: %v \n", seedUrls)

	pwi, err := fetcher.New(fetcher.CustomBrowserTypeOptions{
		LaunchOptions:  playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)},
		OnlySameOrigin: true,
	})

	if err != nil {
		log.Fatalf("Failed to Create Fetcher Instance %v : ", err)
	}
	defer pwi.Close()
	robotChecker := policy.New("*", func(u string) ([]byte, error) {
		return pwi.FetchRobotsContent(u)
	})

	if allowed, _ := robotChecker.IsAllowed(seedUrls[0]); !allowed {
		log.Printf("Seed Url is disalloweb in robots txt.")
		return
	}

	jobQueue := make(chan fetcher.CrawlJob, 1000)
	visits := VisitedLinks{
		Visited: make(map[string]bool),
	}
	errCh := make(chan error, 1000)
	jobQueue <- fetcher.CrawlJob{Url: seedUrls[0], Depth: 0}
	var visitWg *sync.WaitGroup = new(sync.WaitGroup)
	visitWg.Add(1)
	go func() {
		visitWg.Wait()
		close(jobQueue)
		close(errCh)
	}()
	maxDepth := 2
	sem := make(chan struct{}, 5)

	limiter := rate.NewLimiter(rate.Limit(1), 1)

	t := time.Now()

loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Proces stopped by user")
			break loop
		case job, ok := <-jobQueue:
			if !ok {
				break loop
			}

			urlStr := job.Url.String()
			visits.rwMu.Lock()
			if visits.Visited[urlStr] {
				visits.rwMu.Unlock()
				visitWg.Done()
				continue
			}
			fmt.Printf("New Job starting for %v\n", urlStr)
			visits.Visited[urlStr] = true
			visits.rwMu.Unlock()

			if job.Depth > maxDepth {
				fmt.Printf("%d maxDepth, current depth %d for %s skipped", maxDepth, job.Depth, job.Url)
				visitWg.Done()
				continue
			}
			allowed, ts := robotChecker.IsAllowed(job.Url)
			if !allowed {
				log.Printf("Robots.txt blockage: %s skipped.", job.Url)
				visitWg.Done()
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
				visitWg.Done()
				continue
			}
			jitter := time.Duration(rand.Intn(2000)+500) * time.Millisecond
			time.Sleep(jitter)

			outDir := storage.CreateOutDir("../../files", job.Url)
			filename := filepath.Join(outDir, time.Now().UTC().Format("20060102T150405")+".mhtml")

			sem <- struct{}{}

			go func(target fetcher.CrawlJob, targetFile string) {
				defer visitWg.Done()
				defer func() { <-sem }()
				urlStr := target.Url.String()
				defer pwi.ClosePage(urlStr)

				err := pwi.GoTo(urlStr, fetcher.CustomGotoOptions{
					GotoOptions: playwright.PageGotoOptions{
						WaitUntil: playwright.WaitUntilStateNetworkidle,
						Timeout:   playwright.Float(30000),
					},
					SessionType:              fetcher.Status(fetcher.CDPSession),
					AllowInsecureConnections: false,
				})

				if err != nil {
					log.Printf("Error (%s): %v", urlStr, err)
					return
				}
				mhtml, err := pwi.FetchMHTML(urlStr)
				if err != nil {
					log.Printf("Error (%s): %v", urlStr, err)
					return
				}

				err = os.WriteFile(targetFile, mhtml, 0644)

				if err != nil {
					log.Printf("Disk Write Error: %v ", err)
					return
				}

				fmt.Printf("Depth [%d] saved: %s\n", target.Depth, urlStr)

				pwi.LocateLinks(job, jobQueue, errCh, visitWg)

			}(job, filename)
		case err, ok := <-errCh:
			if !ok {
				break loop
			}
			log.Printf("Crawler Error: %v", err)

		}
	}
	fmt.Printf("%v", time.Since(t))
}
