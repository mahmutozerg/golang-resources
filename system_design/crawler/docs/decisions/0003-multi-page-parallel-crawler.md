# ADR 0003: Queue-Based BFS Crawling Architecture with Worker Pool

## Context and Problem Statement

Following the implementation of [ADR 0002](https://github.com/mahmutozerg/golang-resources/blob/main/system_design/crawler/docs/decisions/0002-mhtml-snapshot-strategy.md), our crawler needed to scale beyond sequential execution. The initial single-page approach created a bottleneck where navigation, fetching, and link extraction blocked each other.

We needed an architecture that supports:

1.  **Concurrency:** Processing multiple URLs simultaneously without race conditions.
2.  **Depth Control:** Implementing Breadth-First Search (BFS) to respect crawl limits.
3.  **Resource Safety:** Preventing memory leaks (zombie tabs) and ensuring graceful shutdown.

## Considered Options

- **Sequential Execution:** Safe but inefficient.
- **Fire-and-Forget Async Navigation:** Hard to control depth and rate limits effectively.
- **Full Browser Context per Worker:** Too heavy on memory/CPU.
- **Worker Pool with Page Registry (Chosen):** A hybrid approach using a buffered channel for BFS scheduling and a map-based registry for managing active Playwright pages.

## Decision Outcome

Chosen option: **Worker Pool with Page Registry & BFS Queue**.

We refactored `PwInstance` to manage active pages via a registry, but drove the execution through a concurrent Worker Pool pattern consuming from a BFS Queue.

**Key Architectural Changes:**

1.  **BFS Navigation:** Replaced linear execution with a `jobQueue` channel handling `CrawlJob` structs (URL + Depth).
2.  **Page Registry:** `PwInstance` maintains a `map[string]playwright.Page` protected by `sync.RWMutex`. This allows `GoTo` and `FetchMHTML` to operate on specific tabs safely.
3.  **Concurrency Control:** A semaphore channel (`sem`) limits the number of active browser tabs (e.g., 5) to prevent resource exhaustion.
4.  **Active Cleanup Strategy:** Instead of a background cleaner routine, we utilized `defer` logic within workers to guarantee page closure, combined with Playwright's native navigation timeouts.

## Technical Implementation Details

### 1. State Management

- **Registry:** `map[string]playwright.Page` (Mutex-protected) stores active tabs.
- **Visited Tracking:** A dedicated `VisitedLinks` struct with `sync.RWMutex` prevents cycles and duplicate processing (Check-Lock-Check pattern).
- **Keep-Alive:** A "Dummy Page" is initialized in `New()` to prevent the browser process from terminating during idle rate-limit windows.

### 2. The Worker Lifecycle

The main loop consumes jobs and spawns goroutines. Each worker follows this strict lifecycle:

1.  **Acquire Semaphore:** Block if max tabs reached.
2.  **Defer Cleanup:** Register `defer pwi.ClosePage(url)` immediately to ensure cleanup on any error path.
3.  **Navigate:** Call `pwi.GoTo` (creates tab, registers to map, navigates).
4.  **Snapshot:** Call `pwi.FetchMHTML` (reads CDP session).
5.  **Extract & Schedule:** Call `pwi.LocateLinks` to find child URLs and push them to `jobQueue` (incrementing Depth).

### 3. Graceful Shutdown

Implemented via `context.Context` and `os/signal`:

- On `SIGTERM`/`Interrupt`, the context is canceled.
- The main loop breaks, stopping new job scheduling.
- Existing workers complete (or timeout), and `defer` handlers close browser resources cleanly.

## Consequences

- **Good:** **High Throughput:** Parallel processing limited only by bandwidth and semaphore count.
- **Good:** **Resilience:** "Fail-Closed" logic and timeouts prevent the crawler from hanging on unresponsive sites.
- **Good:** **Memory Efficiency:** Single Browser Context shared across lightweight Pages; tabs are closed immediately after processing.
- **Bad:** **Complexity:** State management (Visited Map, Job Queue, Page Registry) is more complex than a simple recursive crawler.
- **Accepted:** We accept the complexity of manual resource management (`defer ClosePage`) in exchange for granular control over the crawling process.

## Status

**Implemented.**

- [x] **Worker Pool Limits:** Semaphore pattern implemented.
- [x] **Orphan Cleanup:** Solved via `defer` + Timeout (Active Cleanup) instead of background routines.
