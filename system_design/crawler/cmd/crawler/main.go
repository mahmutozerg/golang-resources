package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	_ "net/http/pprof"
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

	// Graceful Shutdown Sinyalleri
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
		LaunchOptions:  playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(false)},
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
	visits := VisitedLinks{Visited: make(map[string]bool)}
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

			visits.rwMu.Lock()
			if visits.Visited[urlStr] {
				visits.rwMu.Unlock()
				visitWg.Done()
				continue
			}
			visits.Visited[urlStr] = true
			visits.rwMu.Unlock()

			if job.Depth > config.MaxDepth {
				visitWg.Done()
				continue
			}

			globalID++
			workerID := globalID
			sem <- struct{}{}

			go func(id int, target fetcher.CrawlJob) {
				defer visitWg.Done()
				defer func() { <-sem }()
				urlStr := target.Url.String()
				if config.ShouldSkipLink(target.Url) {
					log.Printf("[Worker-%03d] File link detected (by ext): %s, skipping", id, target.Url)
					return
				}

				pol, err := robotChecker.GetPolicy(target.Url)
				if err != nil {
					log.Printf("[Worker-%03d] Policy Error (%s): %v", id, target.Url, err)
					return
				}

				if pol.Rule != nil && !pol.Rule.Test(target.Url.Path) {
					log.Printf("[Worker-%03d]  Robots.txt Disallowed: %s", id, target.Url)
					return
				}

				reservation := pol.Limiter.Reserve()
				if !reservation.OK() {
					log.Printf("[Worker-%03d] Limiter Error: Reservation failed", id)
					return
				}

				delay := reservation.Delay()
				if delay > 3*time.Second {
					reservation.Cancel()

					visits.rwMu.Lock()
					delete(visits.Visited, urlStr)
					visits.rwMu.Unlock()

					log.Printf("[Worker-%03d] Too fast for %s. Re-queuing in %v (Non-blocking)", id, target.Url.Host, delay)
					visitWg.Add(1)

					time.AfterFunc(delay, func() {
						select {
						case <-ctx.Done():
							visitWg.Done()
						case jobQueue <- target:
						}
					})

					return
				}

				if delay > 0 {
					if delay > 3*time.Second {
						log.Printf("[Worker-%03d]  Rate Limit: Waiting %v for %s...", id, delay, target.Url.Host)
					}
					select {
					case <-ctx.Done():
						reservation.Cancel()
						log.Printf("[Worker-%03d]  Shutdown received, cancelling worker wait...", id)
						return
					case <-time.After(delay):
					}
				}

				jitter := time.Duration(rand.Intn(config.JitterMax)+config.JitterMin) * time.Millisecond
				select {
				case <-ctx.Done():
					return
				case <-time.After(jitter):
				}

				outDir := storage.CreateOutDir("../../files", target.Url)
				filename := filepath.Join(outDir, time.Now().UTC().Format("20060102T150405")+".mhtml")

				defer pwi.ClosePage(urlStr)

				err = pwi.GoTo(urlStr, fetcher.CustomGotoOptions{
					GotoOptions: playwright.PageGotoOptions{
						WaitUntil: playwright.WaitUntilStateNetworkidle,
						Timeout:   playwright.Float(config.GoToRegularTimeOutMs),
					},
					SessionType:              fetcher.CDPSession,
					AllowInsecureConnections: false,
				})

				if err != nil {
					log.Printf("[Worker-%03d] GoTo Error (%s): %v", id, urlStr, err)
					return
				}

				mhtml, err := pwi.FetchMHTML(urlStr)
				if err != nil {
					log.Printf("[Worker-%03d]  MHTML Error: %v", id, err)
					return
				}

				err = os.WriteFile(filename, mhtml, config.HostOutputFolderPerm)
				if err != nil {
					log.Printf("[Worker-%03d] Disk Error: %v", id, err)
					return
				}

				fmt.Printf("[Worker-%03d]  Saved: %s (Depth: %d)\n", id, urlStr, target.Depth)

				pwi.LocateLinks(target, jobQueue, errCh, &visitWg)

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
