package fetcher

import (
	"fmt"
	"net/url"
	"sync"

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
				return fmt.Errorf("yeni sekme açılamadı: %w", err)
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
			return fmt.Errorf("HTTP Hatası alındı (%d): %s", resp.Status(), resp.StatusText())
		}
	}
	return nil
}

func (pwi *PwInstance) LocateLinks(parentUrl *url.URL, urlCh chan string, errCh chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	pwi.pageMu.RLock()
	p, ok := pwi.pages[parentUrl.String()]
	pwi.pageMu.RUnlock()

	if !ok || parentUrl == nil {
		errCh <- fmt.Errorf("Failed to find link in the map or parenUrl returned nil : %v", parentUrl)
		return
	}

	entries, err := p.Locator("a").All()
	if err != nil {
		errCh <- err
		return
	}

	for _, entry := range entries {
		href, err := entry.GetAttribute("href")
		if err != nil || href == "" || href == "#" {
			continue
		}

		relUrl, err := url.Parse(href)
		if err != nil {
			continue
		}
		absUrl := parentUrl.ResolveReference(relUrl)

		if pwi.OnlySameOrigin && absUrl.Host != parentUrl.Host {
			continue
		}
		urlCh <- absUrl.String()
	}
}

func (pwi *PwInstance) FetchMHTML(url string) ([]byte, error) {
	err := pwi.GoTo(url, CustomGotoOptions{
		GotoOptions: playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(30000),
		},
		SessionType:              Status(CDPSession),
		AllowInsecureConnections: false,
	})

	if err != nil {
		// Handle cleanup at error
		return nil, fmt.Errorf("Navigation Error: %w", err)
	}

	pwi.pageMu.RLock()
	page, ok := pwi.pages[url]
	pwi.pageMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("Page Not Found (Internal Error)")
	}

	defer func() {
		pwi.pageMu.Lock()
		delete(pwi.pages, url)
		pwi.pageMu.Unlock()
		page.Close()
	}()

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
		return nil, fmt.Errorf("CDP datasının formatı string değil veya boş")
	}

	return []byte(dataStr), nil
}

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
