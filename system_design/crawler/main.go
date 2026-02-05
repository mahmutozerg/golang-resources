package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/playwright-community/playwright-go"
)

func main() {

	f, err := os.Open("seed.txt")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	var baseUrl *url.URL

	if sc.Scan() {
		u, err := url.Parse(sc.Text())
		if err != nil {
			log.Fatalf("failed to parse seed: %v", err)
		}
		baseUrl = u
	}

	if baseUrl == nil {
		log.Fatal("no seed URL found")
	}

	fmt.Println("Base URL:", baseUrl.String())

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	context, err := browser.NewContext()
	if err != nil {
		log.Fatalf("could not create context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	_, err = page.Goto(baseUrl.String(), playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	page.AddStyleTag(playwright.PageAddStyleTagOptions{
		Content: playwright.String(`
			* {
				animation: none !important;
				transition: none !important;
			}
		`),
	})

	entries, err := page.Locator("a").All()
	if err != nil {
		log.Fatalf("could not get entries: %v", err)
	}

	var linksToVisit []string
	for _, entry := range entries {
		href, err := entry.GetAttribute("href")
		if err != nil || href == "" || href == "#" {
			continue
		}
		link, err := baseUrl.Parse(href)
		if err != nil {
			continue
		}

		// İleride option yapacağız bunu
		if link.Host != baseUrl.Host {
			continue
		}

		linksToVisit = append(linksToVisit, link.String())
	}

	fmt.Printf("%d adet link bulundu, indirme başlıyor...\n", len(linksToVisit))

	outDir := path.Join("files", baseUrl.Host, baseUrl.Path)
	os.MkdirAll(outDir, 0755)

	cdp, err := context.NewCDPSession(page)
	if err != nil {
		log.Fatalf("could not create CDP session: %v", err)
	}
	_, err = cdp.Send("Page.enable", nil)
	if err != nil {
		log.Fatalf("could not enable Page domain: %v", err)
	}

	for i, linkStr := range linksToVisit {

		fmt.Println("Archiving:", linkStr)

		_, err = page.Goto(linkStr, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(30000),
		})
		if err != nil {
			log.Println("skip (timeout/error):", err)
			continue
		}

		result, err := cdp.Send("Page.captureSnapshot", map[string]interface{}{
			"format": "mhtml",
		})
		if err != nil {
			log.Println("snapshot failed:", err)
			continue
		}

		mhtml := result.(map[string]interface{})["data"].(string)

		filename := path.Join(outDir, strconv.Itoa(i)+"-"+time.Now().UTC().Format("20060102T150405")+".mhtml")
		err = os.WriteFile(filename, []byte(mhtml), 0644)
		if err != nil {
			log.Println("write failed:", err)
			continue
		}

		fmt.Println("Saved: ", filename)
		break

	}
}
