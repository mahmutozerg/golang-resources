# AGENTS.md: Development Guidelines for Go Web Crawler

This document provides essential information for agentic coding agents working on this Go & Playwright web crawler project.

## Build, Lint, and Test Commands

### Building and Running
- **Build the main application**: `go build ./cmd/crawler/main.go`
- **Run the crawler**: `go run ./cmd/crawler/main.go`
- **Install dependencies**: `go mod tidy`

### Testing
- **Run all tests**: `go test ./...` (Note: Currently no test files exist)
- **Run tests in specific package**: `go test ./internal/worker/`
- **Run tests with verbose output**: `go test -v ./...`
- **Run single test function**: `go test -run TestFunctionName ./path/to/package`
- **Run tests without caching**: `go test -count=1 ./...`

### Code Quality
- **Format code**: `go fmt ./...`
- **Run static analysis**: `go vet ./...`
- **Check for race conditions**: `go test -race ./...` (when tests are added)
- **Build without running**: `go build -o crawler ./cmd/crawler/main.go`

### Playwright Setup (Required First Time)
```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps
```

## Project Structure and Architecture

### Directory Layout
```
├── cmd/crawler/main.go          # Application entry point
├── internal/
│   ├── config/                  # Configuration constants and utilities
│   │   ├── file_ext.go         # File extension filtering
│   │   └── seed.go             # Seed URL loading
│   ├── fetcher/                 # Playwright browser wrapper
│   │   └── browser.go          # Browser automation logic
│   ├── policy/                  # robots.txt and rate limiting
│   │   └── registery.go        # Per-host policy management
│   ├── state/                   # Shared state management
│   │   └── crawler_state.go    # Visited links tracking
│   ├── storage/                 # File output utilities
│   │   └── saver.go            # Directory creation for MHTML files
│   └── worker/                  # Core crawling logic
│       └── worker.go           # URL processing pipeline
├── docs/decisions/             # Architectural Decision Records
└── files/                      # Output directory for MHTML snapshots
```

### Core Architecture
- **Producer-Consumer Model**: Main goroutine produces jobs, worker pool consumes
- **Per-Host Rate Limiting**: Each domain gets its own rate limiter (solves head-of-line blocking)
- **Dynamic Policy Registry**: Lazy loading of robots.txt with caching
- **Playwright Integration**: Headless browser for modern web content
- **Graceful Shutdown**: Context-based cancellation for clean exits

## Code Style Guidelines

### General Principles
- Follow standard Go conventions and idioms
- Use clear, descriptive names for functions and variables
- Keep functions focused and relatively small
- Prefer composition over inheritance

### Import Organization
- Group imports into three sections: standard library, third-party, internal
- Use absolute import paths with full module name
- Example from codebase:
```go
import (
    "context"
    "fmt"
    "log"
    
    "github.com/mahmutozerg/golang-resources/system_design/crawler/internal/config"
    "github.com/playwright-community/playwright-go"
)
```

### Naming Conventions
- **Packages**: lowercase, single words when possible (e.g., `worker`, `policy`)
- **Constants**: `SCREAMING_SNAKE_CASE` for exported constants
- **Variables**: `camelCase` (e.g., `jobQueue`, `visitWg`)
- **Functions**: `PascalCase` for exported, `camelCase` for unexported
- **Structs**: `PascalCase` (e.g., `PwInstance`, `CrawlJob`)
- **Interfaces**: Often end with `-er` suffix when appropriate

### Error Handling
- Always handle errors explicitly, never ignore them
- Use `fmt.Errorf` for wrapped errors with context
- Include worker ID and context in error messages: `fmt.Errorf("[Worker-%03d] MHTML Error: %v", id, err)`
- Use the `Must()` helper pattern for fatal errors in initialization:
```go
func Must(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %v", msg, err)
    }
}
```
- Log errors with appropriate context levels

### Type Safety and Constants
- Define configuration constants in `internal/config/` package
- Use typed constants for timeouts, limits, and permissions
- Example pattern from codebase:
```go
const (
    JobQueueSize          = 1000
    MaxDepth              = 3
    ConcurrentWorkerCount = 25
    GoToRegularTimeOutMs  = 60000
)
```

### Concurrency Patterns
- Use `sync.RWMutex` for shared state access
- Implement graceful shutdown with `context.Context`
- Use channels for communication between goroutines
- Apply rate limiting per host, not globally
- Use `sync.WaitGroup` for coordinated goroutine termination
- Pattern for worker pools with semaphore limiting:
```go
sem := make(chan struct{}, config.ConcurrentWorkerCount)
sem <- struct{}{}
defer func() { <-sem }()
```

### Logging
- Use structured logging with worker IDs
- Include relevant context (URL, depth, timing)
- Different log levels for different scenarios (info, warning, error)
- Example: `log.Printf("[Worker-%03d] Processing %s at depth %d", id, url, depth)`

### Testing Guidelines (When Adding Tests)
- Place test files in same package as code being tested
- Use `*_test.go` naming convention
- Test both success and failure scenarios
- Mock external dependencies (Playwright, network calls)
- Use table-driven tests for multiple scenarios
- Test concurrent code with race detection enabled

### Configuration Management
- Centralize all configuration in `internal/config/` package
- Use constants for compile-time configuration
- Separate concerns: timeouts, worker counts, file permissions
- Make configuration values easily discoverable and modifiable
- Load seed URLs from `seed.txt` file (one URL per line, `#` for comments)

### Performance Considerations
- Minimize memory allocations in hot paths
- Use connection pooling for browser instances
- Implement proper cleanup in defer statements
- Consider memory usage of visited URLs map
- Use efficient data structures for URL filtering

## Specific Implementation Patterns

### Worker Function Pattern
Core worker functions should follow this signature pattern:
```go
func Process(job CrawlJob, dependencies...) error {
    // 1. Policy check and rate limiting
    // 2. Page navigation with timeout
    // 3. Content extraction and saving
    // 4. Link discovery and queueing
    return nil
}
```

### Error Context Pattern
Always provide context in errors:
```go
return fmt.Errorf("[Worker-%03d] Policy Error (%s): %v", id, url, err)
```

### Resource Management Pattern
Always use defer for cleanup:
```go
defer pwi.ClosePage(urlStr)
defer visitWg.Done()
```

### Double-Checked Locking Pattern
Used in policy registry for efficient lazy initialization:
```go
c.rwMu.RLock()
existingPolicy, ok := c.hostRules[host]
c.rwMu.RUnlock()

if ok {
    return existingPolicy, nil
}

// Expensive operation outside lock
c.rwMu.Lock()
defer c.rwMu.Unlock()
if existingPolicy, ok = c.hostRules[host]; ok {
    return existingPolicy, nil
}
// Create new policy...
```

## Dependencies
- `github.com/playwright-community/playwright-go`: Browser automation
- `github.com/temoto/robotstxt`: robots.txt parsing
- `golang.org/x/time/rate`: Rate limiting implementation
- Go 1.26.0+ required

## Development Workflow
1. Run `go mod tidy` after adding dependencies
2. Use `go fmt ./...` before committing
3. Run `go vet ./...` to catch common issues
4. Test Playwright setup before running crawler
5. Create `seed.txt` file with URLs for testing
6. Check `files/` directory for MHTML output

## Architectural Decision Records
Significant design decisions are documented in `docs/decisions/`:
- 0001-crawler-system-proposal.md
- 0002-mhtml-snapshot-strategy.md
- 0003-multi-page-parallel-crawler.md
- 0004-multi-seed-host-based-limits.md

Consult these ADRs before making major architectural changes.