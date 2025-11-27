package scrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
)

type Scrapper struct {
	PW       *playwright.Playwright
	Browser  playwright.Browser
	Context  playwright.BrowserContext
	Page     playwright.Page
	reqFile  *os.File
	respFile *os.File
}

type NetworkTraffic struct {
	Direction string `json:"direction"`
	Method    string `json:"method,omitempty"`
	Status    int    `json:"status,omitempty"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
}

type ScrapperOptions struct {
	Bwp            []string
	CreateRespFile bool
	CreateReqFile  bool
}

var allowedResources = map[string]bool{
	"document": true,
	"xhr":      true,
	"fetch":    true,
}

func NewScrapper(so ScrapperOptions) (*Scrapper, error) {

	// Install options
	runOption := &playwright.RunOptions{
		SkipInstallBrowsers: len(so.Bwp) != 0,
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

	if len(so.Bwp) != 0 {
		launchOptions.ExecutablePath = playwright.String(so.Bwp[0])
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

	var reqf *os.File = nil
	var respf *os.File = nil

	if so.CreateReqFile {
		reqf, err = createFile("requests.json")
		if err != nil {
			return nil, err
		}
	}
	if so.CreateRespFile {
		respf, err = createFile("response.json")
		if err != nil {
			return nil, err
		}
	}

	return &Scrapper{
		PW:       pw,
		Browser:  browser,
		Context:  context,
		reqFile:  reqf,
		respFile: respf,
	}, nil
}

func (s *Scrapper) Close() error {
	var errs []error
	var err error

	if s.Page != nil {
		if err = s.Page.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.Context != nil {
		if err = s.Context.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.Browser != nil {
		if err = s.Browser.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.PW != nil {
		if err = s.PW.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.reqFile != nil {
		if err = s.reqFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.respFile != nil {
		if err = s.respFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func (s *Scrapper) SetupHooks() {
	s.Context.OnRequest(func(req playwright.Request) {

		if isNetworkType(req.ResourceType()) {

			s.logToFile(NetworkTraffic{
				Direction: "request",
				Method:    req.Method(),
				Type:      req.ResourceType(),
				URL:       req.URL(),
				Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			})

		}

	})

	s.Context.OnResponse(func(res playwright.Response) {

		if isNetworkType(res.Request().ResourceType()) {

			s.logToFile(NetworkTraffic{
				Direction: "response",
				Status:    res.Status(),
				Type:      res.Request().ResourceType(),
				URL:       res.URL(),
				Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			})
		}

	})
}

func (s *Scrapper) logToFile(nt NetworkTraffic) {

	jsonLog, err := json.Marshal(nt)
	if err != nil {
		return
	}
	switch nt.Direction {
	case "response":
		if s.respFile != nil {
			s.respFile.Write(jsonLog)
			s.respFile.Write([]byte("\n"))
			return
		}
	case "request":
		if s.reqFile != nil {
			s.reqFile.Write(jsonLog)
			s.reqFile.Write([]byte("\n"))
			return
		}
	}

	fmt.Println(string(jsonLog))

}
func isNetworkType(t string) bool {
	return allowedResources[t]
}

func createFile(filename string) (*os.File, error) {
	f, err := os.OpenFile(filename,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return f, nil
}
