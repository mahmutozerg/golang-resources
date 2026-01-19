package pwinit

import (
	"github.com/playwright-community/playwright-go"
)

type BrowserState struct {
	Browser playwright.Browser
	Page    playwright.Page
}

type CustomInstallOptions struct {
	Skip     bool
	Ep       string
	Headless bool
}

func installOrSkip(skip bool) error {

	runOption := &playwright.RunOptions{
		SkipInstallBrowsers: skip,
	}
	err := playwright.Install(runOption)

	return err
}

func setOptions(options CustomInstallOptions) playwright.BrowserTypeLaunchOptions {

	return playwright.BrowserTypeLaunchOptions{
		ExecutablePath: playwright.String(options.Ep),
		Headless:       playwright.Bool(options.Headless),
	}
}
func Init(opt CustomInstallOptions) (*BrowserState, error) {

	err := installOrSkip(opt.Skip)
	if err != nil {
		return nil, err
	}
	pw, err := playwright.Run()

	options := setOptions(opt)

	browser, err := pw.Chromium.Launch(options)

	if err != nil {
		return nil, err
	}

	page, err := browser.NewPage()
	if err != nil {
		return nil, err
	}
	return &BrowserState{Browser: browser, Page: page}, nil
}
