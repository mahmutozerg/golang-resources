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

const (
	// RequeueThreshold is the delay threshold above which a job is requeued
	// instead of blocking the worker.
	RequeueThreshold = 3 * time.Second
)

// Processor handles the crawling pipeline for a single worker.
type Processor struct {
	Fetcher *fetcher.PwInstance
	Checker *policy.Checker
	ErrCh   chan error
	JobCh   chan fetcher.CrawlJob
	VisitWg *sync.WaitGroup
	Visits  *crawler_state.VisitedLinks
	ID      int
}

// Process handles a single crawl job: rate limiting, fetching, saving MHTML, and link extraction.
// It respects robots.txt rules and implements graceful shutdown via context cancellation.
func (p *Processor) Process(ctx context.Context, job fetcher.CrawlJob) error {
	urlStr := job.Url.String()
	defer func() { p.Fetcher.ClosePage(urlStr) }()

	reservation, err := p.processPolicy(job.Url)
	if reservation == nil && err != nil {
		return err
	}
	if reservation == nil {
		return fmt.Errorf("[Worker-%03d] Reservation returned nil but no error", p.ID)
	}
	defer reservation.Cancel()

	delay := reservation.Delay()

	// Long wait: requeue non-blocking
	if delay > RequeueThreshold {
		reservation.Cancel()

		p.Visits.Rwmu.Lock()
		delete(p.Visits.Visited, urlStr)
		p.Visits.Rwmu.Unlock()

		log.Printf("[Worker-%03d] Too fast for %s. Re-queuing in %v (non-blocking)", p.ID, job.Url.Host, delay)

		p.VisitWg.Add(1)
		time.AfterFunc(delay, func() {
			select {
			case <-ctx.Done():
				p.VisitWg.Done()
			case p.JobCh <- job:
			}
		})

		return nil
	}

	// Short delay: block with context awareness
	if delay > 0 {
		select {
		case <-ctx.Done():
			log.Printf("[Worker-%03d] Shutdown received, cancelling wait...", p.ID)
			return nil
		case <-time.After(delay):
		}
	}

	// Apply jitter to avoid thundering herd
	if err := p.applyJitter(ctx); err != nil {
		return err
	}

	// Navigate to page
	if err := p.Fetcher.GoTo(urlStr, fetcher.CustomGotoOptions{
		GotoOptions: playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(config.GoToRegularTimeOutMs),
		},
		SessionType:              fetcher.CDPSession,
		AllowInsecureConnections: false,
	}); err != nil {
		return err
	}

	// Save MHTML snapshot
	if err := p.processMHTML(job.Url); err != nil {
		return err
	}

	// Extract and queue links
	p.Fetcher.LocateLinks(job, p.JobCh, p.ErrCh, p.VisitWg)

	return nil
}

// processPolicy checks robots.txt and reserves a rate limit slot.
func (p *Processor) processPolicy(u *url.URL) (*rate.Reservation, error) {
	pol, err := p.Checker.GetPolicy(u)
	if err != nil {
		return nil, fmt.Errorf("[Worker-%03d] Policy Error (%s): %v", p.ID, u, err)
	}

	if pol.Rule != nil && !pol.Rule.Test(u.Path) {
		return nil, fmt.Errorf("[Worker-%03d] Robots.txt disallowed: %s", p.ID, u)
	}

	reservation := pol.Limiter.Reserve()
	if !reservation.OK() {
		return nil, fmt.Errorf("[Worker-%03d] Limiter Error: Reservation failed", p.ID)
	}

	return reservation, nil
}

// processMHTML captures and saves an MHTML snapshot of the current page.
func (p *Processor) processMHTML(u *url.URL) error {
	urlStr := u.String()
	outDir := storage.CreateOutDir("../../files", u)
	filename := filepath.Join(outDir, time.Now().UTC().Format("20060102T150405")+".mhtml")

	mhtml, err := p.Fetcher.FetchMHTML(urlStr)
	if err != nil {
		return fmt.Errorf("[Worker-%03d] MHTML Error: %v", p.ID, err)
	}

	if err := os.WriteFile(filename, mhtml, config.HostOutputFolderPerm); err != nil {
		return fmt.Errorf("[Worker-%03d] Disk Error: %v", p.ID, err)
	}

	log.Printf("[Worker-%03d] Saved: %s", p.ID, filename)
	return nil
}

// applyJitter sleeps for a random duration to avoid synchronized requests.
func (p *Processor) applyJitter(ctx context.Context) error {
	// Validate jitter range to prevent overflow
	if config.JitterMax <= config.JitterMin {
		return fmt.Errorf("[Worker-%03d] Config Error: JitterMax must be greater than JitterMin", p.ID)
	}

	rangeMs := config.JitterMax - config.JitterMin + 1
	jitter := time.Duration(rand.Intn(rangeMs)+config.JitterMin) * time.Millisecond

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(jitter):
		return nil
	}
}

// Process is the legacy entry point. Deprecated: Use Processor.Process instead.
// Kept for backward compatibility with existing caller code.
func Process(
	ctx context.Context,
	job fetcher.CrawlJob,
	pwi *fetcher.PwInstance,
	checker *policy.Checker,
	errCh chan error,
	jobCh chan fetcher.CrawlJob,
	visitWg *sync.WaitGroup,
	visits *crawler_state.VisitedLinks,
	workerID int,
) error {
	p := &Processor{
		Fetcher: pwi,
		Checker: checker,
		ErrCh:   errCh,
		JobCh:   jobCh,
		VisitWg: visitWg,
		Visits:  visits,
		ID:      workerID,
	}
	return p.Process(ctx, job)
}
