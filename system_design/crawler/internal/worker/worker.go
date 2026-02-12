package worker

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
	policy "github.com/mahmutozerg/golang-resources/system_design/crawler/internal/policy"
	crawler_state "github.com/mahmutozerg/golang-resources/system_design/crawler/internal/state"
	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/storage"
	"github.com/playwright-community/playwright-go"
	"golang.org/x/time/rate"
)

func Process(job fetcher.CrawlJob,
	pwi *fetcher.PwInstance,
	checker *policy.Checker,
	errCh chan error,
	jobCh chan fetcher.CrawlJob,
	visitWg *sync.WaitGroup,
	visits *crawler_state.VisitedLinks,
	ctx context.Context,
	id int) error {

	urlStr := job.Url.String()
	defer pwi.ClosePage(urlStr)

	reservation, err := processPolicy(job.Url, checker, id)
	if reservation == nil && err != nil {
		return err
	} else if reservation == nil && err == nil {
		return fmt.Errorf("[Worker-%03d] Reservation Returned Nil but no Error", id)
	}

	delay := reservation.Delay()

	if delay > 3*time.Second {
		reservation.Cancel()

		visits.Rwmu.Lock()
		delete(visits.Visited, urlStr)
		visits.Rwmu.Unlock()

		log.Printf("[Worker-%03d] Too fast for %s. Re-queuing in %v (Non-blocking)", id, job.Url.Host, delay)

		visitWg.Add(1)

		time.AfterFunc(delay, func() {
			select {
			case <-ctx.Done():
				visitWg.Done()
			case jobCh <- job:
			}
		})

		return nil
	}

	if delay > 0 {
		select {
		case <-ctx.Done():
			reservation.Cancel()
			log.Printf("[Worker-%03d] Shutdown received, cancelling wait...", id)
			return nil
		case <-time.After(delay):
		}
	}

	jitter := time.Duration(rand.Intn(config.JitterMax)+config.JitterMin) * time.Millisecond
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(jitter):
	}

	err = pwi.GoTo(urlStr, fetcher.CustomGotoOptions{
		GotoOptions: playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(config.GoToRegularTimeOutMs),
		},
		SessionType:              fetcher.CDPSession,
		AllowInsecureConnections: false,
	})
	if err != nil {
		return err
	}

	err = processMHTML(pwi, job.Url, id)
	if err != nil {
		return err
	}
	pwi.LocateLinks(job, jobCh, errCh, visitWg)
	return nil
}

func processPolicy(url *url.URL, checker *policy.Checker, id int) (*rate.Reservation, error) {

	pol, err := checker.GetPolicy(url)
	if err != nil {
		return nil, fmt.Errorf("[Worker-%03d] Policy Error (%s): %v", id, url, err)
	}

	if pol.Rule != nil && !pol.Rule.Test(url.Path) {
		return nil, fmt.Errorf("[Worker-%03d]  Robots.txt Disallowed: %s", id, url)

	}

	reservation := pol.Limiter.Reserve()
	if !reservation.OK() {
		return nil, fmt.Errorf("[Worker-%03d] Limiter Error: Reservation failed", id)
	}
	return reservation, nil
}
func processMHTML(pwi *fetcher.PwInstance, url *url.URL, id int) error {
	urlStr := url.String()
	outDir := storage.CreateOutDir("../../files", url)
	filename := filepath.Join(outDir, time.Now().UTC().Format("20060102T150405")+".mhtml")

	mhtml, err := pwi.FetchMHTML(urlStr)
	if err != nil {
		return fmt.Errorf("[Worker-%03d]  MHTML Error: %v", id, err)
	}

	err = os.WriteFile(filename, mhtml, config.HostOutputFolderPerm)
	if err != nil {
		return fmt.Errorf("[Worker-%03d] Disk Error: %v", id, err)

	}

	return nil
}
