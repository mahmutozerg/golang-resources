package scrapper

import (
	"github.com/playwright-community/playwright-go"
)

type Scrapper struct {
	PW      *playwright.Playwright
	Browser playwright.Browser
}

func NewScrapper() (*Scrapper, error) {
	err := playwright.Install()
	if err != nil {
		return nil, err
	}

	pw, err := playwright.Run()

	if err != nil {
		return nil, err
	}

	browser, err := pw.Chromium.Launch()
	if err != nil {
		return nil, err
	}

	return &Scrapper{
		PW:      pw,
		Browser: browser,
	}, nil
}

func (s *Scrapper) Close() {
	s.Browser.Close()
	s.PW.Stop()
}
