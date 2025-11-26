package scrapper

import (
	"scrapper/constants"
	"scrapper/helper"

	"github.com/playwright-community/playwright-go"
)

type Scrapper struct {
	PW      *playwright.Playwright
	Browser playwright.Browser
	Context playwright.BrowserContext
	Page    playwright.Page
}

func NewScrapper(bwp ...string) (*Scrapper, error) {
	runOption := &playwright.RunOptions{
		SkipInstallBrowsers: true,
	}
	err := playwright.Install(runOption)

	if err != nil {
		return nil, err
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	option := playwright.BrowserTypeLaunchOptions{
		ExecutablePath: playwright.String(bwp[0]),
		Headless:       playwright.Bool(false),
	}

	browser, err := pw.Chromium.Launch(option)
	helper.AssertErrorToNil(err, constants.FailedLaunchLocalBrowser)

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		Locale:    playwright.String("en-US"),
	})

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &Scrapper{
		PW:      pw,
		Browser: browser,
		Context: context,
	}, nil
}

func (s *Scrapper) Close() {
	s.Browser.Close()
	s.PW.Stop()
}
