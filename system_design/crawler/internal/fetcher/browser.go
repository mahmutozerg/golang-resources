package fetcher

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/playwright-community/playwright-go"
)

// Client Is Responsible to free pointers
// please use defer  PwInstance.Close()
type PwInstance struct {
	pw             *playwright.Playwright
	browser        playwright.Browser
	context        playwright.BrowserContext
	session        playwright.CDPSession
	page           playwright.Page
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

	switch opt.SessionType {
	case Status(CDPSession):
		pwi.page.AddStyleTag(playwright.PageAddStyleTagOptions{
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
	resp, err := pwi.page.Goto(url, opt.GotoOptions)

	if err != nil {
		return err
	}

	// If we try to navigate in to same page we'll get  resp nil
	// and also err nil this fixes the case where we are trying to navigate to the same page we
	// are allready in it
	if resp != nil {
		if !(resp.Status() >= 200 && resp.Status() < 300) {
			return fmt.Errorf("HTTP Hatası alındı (%d): %s", resp.Status(), resp.StatusText())
		}
	}
	return nil
}

func (pwi *PwInstance) LocateLinks(parentUrl *url.URL, urlCh chan string, errCh chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	entries, err := pwi.page.Locator("a").All()
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
		finalLink := absUrl.String()

		urlCh <- finalLink

	}
}

// Use FetchMHTML after you navigated the page
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
		return nil, fmt.Errorf("Navigasyon hatası: %w", err)
	}

	result, err := pwi.session.Send("Page.captureSnapshot", map[string]any{
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

	page, err := context.NewPage()
	if err != nil {
		return nil, err
	}

	// In future we need a interface that covers both cdp and warc
	// we should call interface.NewSession() in order to use them
	cdpSession, err := context.NewCDPSession(page)
	if err != nil {
		return nil, fmt.Errorf("CDP session başlatılamadı: %w", err)
	}

	if _, err = cdpSession.Send("Page.enable", nil); err != nil {
		return nil, fmt.Errorf("CDP Page domain aktif edilemedi: %w", err)
	}

	return &PwInstance{
		pw:             pw,
		browser:        browser,
		context:        context,
		session:        cdpSession,
		OnlySameOrigin: opt.OnlySameOrigin,
		page:           page,
	}, nil
}
func (pwi *PwInstance) Close() {
	if pwi.page != nil {
		pwi.page.Close()
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
