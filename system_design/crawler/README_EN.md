# Go & Playwright Web Crawler

This project is a scalable, polite web crawler built with Go and Playwright. Unlike simple HTML parsers, it renders pages using a real browser engine (Chromium) and saves the complete visual and DOM state as **MHTML** files.

## Features

- **MHTML Snapshots:** Captures the full page state, including CSS, JavaScript execution results, and images, into a single portable file.
- **Host-Based Rate Limiting:** Implements a smart concurrency model. Instead of a global speed limit, it manages rate limits per domain. A slow response from one site does not block the crawling of other faster sites (solves Head-of-Line Blocking).
- **Transparent Delays:** If a site enforces a long `Crawl-delay` (e.g., 30 seconds), the worker logs the waiting time instead of silently freezing, providing clear feedback on the process status.
- **Robots.txt Compliance:** Respects `Disallow` rules and automatically adjusts crawl speed based on `Crawl-delay` directives found in `robots.txt`.
- **Playwright Engine:** Uses headless Chromium to handle modern web features like SPAs and dynamic content rendering.
- **Graceful Shutdown:** Handles system interrupts (CTRL+C) safely, ensuring resources are cleaned up properly before exiting.

## Prerequisites

Ensure you have Go(1.25.6) installed on your machine. You will also need to install the Playwright drivers and dependencies.

1. Clone the repository and navigate to the directory.
2. Download Go modules:

```bash
go mod tidy
```

3. Install Playwright browsers and system dependencies:

```bash
go run [github.com/playwright-community/playwright-go/cmd/playwright@latest](https://github.com/playwright-community/playwright-go/cmd/playwright@latest) install --with-deps

```

## Usage

To start crawling, you need to provide a list of starting URLs.

1. Create a file named `seed.txt` in the root directory and add your target URLs (one per line):

```text
https://go.dev/
https://news.ycombinator.com/
https://www.wikipedia.org/
```

2. Run the crawler:

```bash
go run main.go
```

## Configuration

Configuration settings are defined as constants in the `internal/config` package. You can modify these values in the code to tune the crawler:

- **JobQueueSize:** The capacity of the job channel (Default: 1000).
- **MaxDepth:** How deep the crawler should follow links (Default: 3).
- **ConcurrentWorkerCount:** The number of parallel workers (Default: 5).
- **GoToRegularTimeOutMs:** Maximum time to wait for a page to load.
- **JitterMin/Max:** Adds random delay between requests to behave more human-like.

## Architecture

The system follows a **Producer-Consumer** model enhanced by a "Policy Registry".

1. **Seed Loading:** URLs are read from `seed.txt` and pushed into the job queue.
2. **Worker Pool:** A fixed number of workers consume jobs from the queue.
3. **Policy Registry:** Before fetching a page, the worker checks the Registry.

- The Registry lazily fetches and parses `robots.txt` for the domain if it hasn't been seen before.
- It creates a dedicated `RateLimiter` for that specific host using **Double-Checked Locking** to ensure thread safety without performance bottlenecks.

4. **Fetch & Save:** If allowed, Playwright navigates to the page, captures an MHTML snapshot, and saves it to the disk.
5. **Discovery:** The worker extracts links from the page, filters them (e.g., same-origin policy), and adds new targets to the queue.

## Output

The crawled content is saved in the `../../files` directory relative to the executable. The crawler creates a directory structure that mirrors the URL path.

Example output path:
`../../files/go.dev/doc/tutorial/index.html/20260209T120000.mhtml`
