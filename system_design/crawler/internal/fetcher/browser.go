package fetcher

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/config"
	"github.com/playwright-community/playwright-go"
)

type PwInstance struct {
	pw             *playwright.Playwright
	browser        playwright.Browser
	context        playwright.BrowserContext
	pages          map[string]playwright.Page
	pageMu         *sync.RWMutex
	OnlySameOrigin bool
}

type CustomBrowserTypeOptions struct {
	LaunchOptions  playwright.BrowserTypeLaunchOptions
	OnlySameOrigin bool
}

type Status int

const (
	CDPSession  Status = iota // 0
	WARCSession               // 1
)

type CrawlJob struct {
	Url   *url.URL
	Depth int
}
type CustomGotoOptions struct {
	GotoOptions              playwright.PageGotoOptions
	SessionType              Status
	AllowInsecureConnections bool
}

func (pwi *PwInstance) GoTo(url string, opt CustomGotoOptions) error {
	pwi.pageMu.RLock()
	p, ok := pwi.pages[url]
	pwi.pageMu.RUnlock()

	if !ok {
		pwi.pageMu.Lock()
		if p, ok = pwi.pages[url]; !ok {
			newPage, err := pwi.context.NewPage()
			if err != nil {
				pwi.pageMu.Unlock()
				return fmt.Errorf("Failed To Open New Page: %w", err)
			}
			pwi.pages[url] = newPage
			p = newPage
		}
		pwi.pageMu.Unlock()
	}

	switch opt.SessionType {
	case Status(CDPSession):
		p.AddStyleTag(playwright.PageAddStyleTagOptions{
			Content: playwright.String(`
			* {
				animation: none !important;
				transition: none !important;
			}
		`),
		})
	case Status(WARCSession):
		return fmt.Errorf("WARCSession Not Implemented")
	}

	resp, err := p.Goto(url, opt.GotoOptions)
	if err != nil {
		return err
	}

	if resp != nil {
		if !(resp.Status() >= 200 && resp.Status() < 300) {
			return fmt.Errorf("HTTP Err (%d): %s", resp.Status(), resp.StatusText())
		}
	}
	return nil
}

func (pwi *PwInstance) LocateLinks(parent CrawlJob, crawlCh chan CrawlJob, errCh chan error, wg *sync.WaitGroup) {
	pwi.pageMu.RLock()
	p, ok := pwi.pages[parent.Url.String()]
	pwi.pageMu.RUnlock()

	if !ok || p == nil {
		errCh <- fmt.Errorf("Failed to find link in the map or parenUrl returned nil : %v", p)
		return
	}

	entries, err := p.Locator("a").All()
	if err != nil {
		errCh <- err
		return
	}

	var linksToPush []*url.URL

	for _, entry := range entries {
		href, err := entry.GetAttribute("href")
		if err != nil || href == "" || href == "#" {
			continue
		}

		relUrl, err := url.Parse(href)
		if err != nil {
			continue
		}
		relUrl.Fragment = ""
		absUrl := parent.Url.ResolveReference(relUrl)
		if pwi.OnlySameOrigin && absUrl.Host != parent.Url.Host {
			continue
		}

		linksToPush = append(linksToPush, absUrl)
	}

	wg.Add(len(linksToPush))

	go func(urls []*url.URL, depth int) {
		for _, u := range urls {
			crawlCh <- CrawlJob{Url: u, Depth: depth}
		}
	}(linksToPush, parent.Depth+1)
}

func (pwi *PwInstance) ClosePage(url string) {
	pwi.pageMu.RLock()
	page, ok := pwi.pages[url]
	pwi.pageMu.RUnlock()

	if !ok {
		log.Println("ClosePage: Page Not Found (Internal Error)")
		return
	}
	pwi.pageMu.Lock()
	delete(pwi.pages, url)
	pwi.pageMu.Unlock()

	err := page.Close()

	if err != nil {
		log.Println("ClosePage: Page Not Found (Internal Error)")
	}

}
func (pwi *PwInstance) FetchMHTML(url string) ([]byte, error) {

	pwi.pageMu.RLock()
	page, ok := pwi.pages[url]
	pwi.pageMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("Page Not Found (Internal Error)")
	}

	cdpSession, err := pwi.context.NewCDPSession(page)
	if err != nil {
		return nil, fmt.Errorf("Failed to start CDP Session: %w", err)
	}
	defer cdpSession.Detach()

	if _, err = cdpSession.Send("Page.enable", nil); err != nil {
		return nil, fmt.Errorf("Failed to Enable CDP : %w", err)
	}

	result, err := cdpSession.Send("Page.captureSnapshot", map[string]any{
		"format": "mhtml",
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to Take Snapshot: %w", err)
	}

	dataMap, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("CDP Response returned none map result %v", dataMap)
	}

	dataStr, ok := dataMap["data"].(string)
	if !ok {
		return nil, fmt.Errorf("CDP data is not string or its nill")
	}

	return []byte(dataStr), nil
}

func (pwi *PwInstance) FetchRobotsContent(url string) ([]byte, error) {
	pwi.pageMu.Lock()
	page, err := pwi.context.NewPage()
	pwi.pageMu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("Failed to open tab for robots.txt: %w", err)
	}

	defer page.Close()

	resp, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(config.GoToRobotsTimeOutMs),
	})

	if err != nil {
		return nil, err
	}
	if resp.Status() == 404 {
		return []byte(""), nil
	}

	if resp.Status() == 401 || resp.Status() == 403 {
		return []byte("User-agent: *\nDisallow: /"), nil
	}

	content, err := page.Locator("body").InnerText()
	if err != nil {
		return nil, fmt.Errorf("Failed to read content: %w", err)
	}

	//fmt.Printf("%s robots txt data", content)
	return []byte(content), nil
}

// Creates Playwright,browser,page,context,rwmu and dummy page
// to make browser stay up
func New(opt CustomBrowserTypeOptions) (*PwInstance, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	browser, err := pw.Chromium.Launch(opt.LaunchOptions)
	if err != nil {
		return nil, err
	}

	context, err := browser.NewContext()
	if err != nil {
		return nil, err
	}

	pages := make(map[string]playwright.Page)

	context.NewPage()
	return &PwInstance{
		pw:             pw,
		browser:        browser,
		context:        context,
		OnlySameOrigin: opt.OnlySameOrigin,
		pages:          pages,
		pageMu:         new(sync.RWMutex),
	}, nil
}

func (pwi *PwInstance) Close() {
	if pwi.pages != nil {
		pwi.pageMu.Lock()
		for key, page := range pwi.pages {
			page.Close()
			delete(pwi.pages, key)
		}
		pwi.pageMu.Unlock()
	}

	if pwi.context != nil {
		pwi.context.Close()
	}

	if pwi.browser != nil {
		pwi.browser.Close()
	}

	if pwi.pw != nil {
		pwi.pw.Stop()
	}
}
