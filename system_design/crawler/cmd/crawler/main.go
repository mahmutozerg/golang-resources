package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/config"
	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/fetcher"
	policy "github.com/mahmutozerg/golang-resources/system_design/crawler/internal/policy"
	crawler_state "github.com/mahmutozerg/golang-resources/system_design/crawler/internal/state"
	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/worker" // Worker paketi eklendi

	"github.com/playwright-community/playwright-go"
)

func Must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func main() {
	// 1. Context ve Sinyal Yönetimi
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nSigterm or Interrupt received, draining queue...")
		cancel()
	}()

	seedUrls, err := config.LoadSeeds("./seed.txt")
	Must(err, "Failed to Load Seeds")
	fmt.Printf("Seed URLs: %v \n", seedUrls)

	pwi, err := fetcher.New(fetcher.CustomBrowserTypeOptions{
		LaunchOptions:  playwright.BrowserTypeLaunchOptions{Headless: new(false)},
		OnlySameOrigin: true,
	})
	if err != nil {
		log.Fatalf("Failed to Create Fetcher Instance: %v", err)
	}
	defer pwi.Close()

	robotChecker := policy.New("MyCrawlerBot", func(u string) ([]byte, error) {
		return pwi.FetchRobotsContent(u)
	}, ctx)

	jobQueue := make(chan fetcher.CrawlJob, config.JobQueueSize)
	visits := crawler_state.VisitedLinks{Visited: make(map[string]bool), Rwmu: new(sync.RWMutex)}
	errCh := make(chan error, config.JobQueueSize)
	var visitWg sync.WaitGroup

	for _, seed := range seedUrls {
		visitWg.Add(1)
		jobQueue <- fetcher.CrawlJob{Url: seed, Depth: 0}
	}

	go func() {
		visitWg.Wait()
		close(jobQueue)
		close(errCh)
	}()

	sem := make(chan struct{}, config.ConcurrentWorkerCount)
	t := time.Now()
	var globalID int = 0
loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Process stopped by user")
			break loop

		case job, ok := <-jobQueue:
			if !ok {
				break loop
			}

			urlStr := job.Url.String()

			visits.Rwmu.Lock()
			if visits.Visited[urlStr] {
				visits.Rwmu.Unlock()
				visitWg.Done()
				continue
			}
			visits.Visited[urlStr] = true
			visits.Rwmu.Unlock()

			if job.Depth > config.MaxDepth {
				visitWg.Done()
				continue
			}

			sem <- struct{}{}
			globalID++
			workerID := globalID

			go func(id int, target fetcher.CrawlJob) {
				defer func() { <-sem }()

				defer visitWg.Done()
				if config.ShouldSkipLink(target.Url) {
					log.Printf("[Worker-%03d] File link detected (by ext): %s, skipping", id, target.Url)
					return
				}

				err := worker.Process(
					target,
					pwi,
					robotChecker,
					errCh,
					jobQueue,
					&visitWg,
					&visits,
					ctx,
					id,
				)

				if err != nil {
					errCh <- err
					log.Printf("[Worker-%03d] Error: %v", id, err)
				}

			}(workerID, job)

		case err, ok := <-errCh:
			if !ok {
				break loop
			}
			log.Printf("Crawler Error: %v", err)
		}
	}
	fmt.Printf("\nTotal Execution Time: %v\n", time.Since(t))
}
