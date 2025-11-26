package scrapper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

type Scrapper struct {
	PW      *playwright.Playwright
	Browser playwright.Browser
	Context playwright.BrowserContext
	Page    playwright.Page
}

func NewScrapper(bwp []string) (*Scrapper, error) {

	// Install options
	runOption := &playwright.RunOptions{
		SkipInstallBrowsers: len(bwp) != 0,
	}

	if err := playwright.Install(runOption); err != nil {
		return nil, err
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	launchOptions := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	}

	if len(bwp) != 0 {
		launchOptions.ExecutablePath = playwright.String(bwp[0])
	}

	browser, err := pw.Chromium.Launch(launchOptions)
	if err != nil {
		return nil, err
	}

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		Locale:    playwright.String("en-US"),
	})
	if err != nil {
		return nil, err
	}

	return &Scrapper{
		PW:      pw,
		Browser: browser,
		Context: context,
	}, nil
}

func (s *Scrapper) SetupHooks() {
	s.Context.OnRequest(func(req playwright.Request) {
		switch req.ResourceType() {
		case "document", "xhr", "fetch":
			logData := map[string]interface{}{
				"direction": "request",
				"method":    req.Method(),
				"type":      req.ResourceType(),
				"url":       req.URL(),
				"timestamp": time.Now().Format(time.RFC3339),
			}
			jsonLog, _ := json.Marshal(logData)
			fmt.Println(string(jsonLog))
		}
	})

	s.Context.OnResponse(func(res playwright.Response) {
		switch res.Request().ResourceType() {
		case "document", "xhr", "fetch":
			logData := map[string]interface{}{
				"direction": "response",
				"status":    res.Status(),
				"type":      res.Request().ResourceType(),
				"url":       res.URL(),
				"timestamp": time.Now().Format(time.RFC3339),
			}
			jsonLog, _ := json.Marshal(logData)
			fmt.Println(string(jsonLog))
		}
	})
}

func (s *Scrapper) Close() {
	if s.Page != nil {
		s.Page.Close()
	}
	if s.Context != nil {
		s.Context.Close()
	}
	if s.Browser != nil {
		s.Browser.Close()
	}
	if s.PW != nil {
		s.PW.Stop()
	}
}
