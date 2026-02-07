# ADR 0003: Multi-Page Parallel Crawling Architecture via Page Registry

## Context and Problem Statement

Following the implementation of [ADR 0002](https://github.com/mahmutozerg/golang-resources/blob/main/system_design/crawler/docs/decisions/0002-mhtml-snapshot-strategy.md), our crawler successfully captures high-fidelity MHTML snapshots. However, the current implementation utilizes a single `playwright.Page` instance within `PwInstance`. This creates a bottleneck: the crawler operates sequentially (single-lane), waiting for one page to navigate and download before processing the next.

Attempting to run concurrent fetch operations on a single `Page` instance causes race conditions, navigation errors, and context mixing. To increase throughput and utilize system resources effectively, we need an architecture that supports concurrent page processing while maintaining strict isolation between different URL contexts.

## Considered Options

- **Sequential Execution (Current Status):** Continue using a single page. Safe but inefficient and non-scalable for large seed lists.
- **Ephemeral Page per Request:** Open a new page, navigate, snapshot, and close it immediately within a single function scope. While thread-safe, this tightly couples navigation and fetching logic, making it harder to implement asynchronous "fire-and-forget" navigation patterns.
- **Worker Pool with Independent Browser Contexts:** Create a full `BrowserContext` for every worker. This provides maximum isolation but incurs significant memory and CPU overhead.
- **Page Registry (Map-based State Management):** Modify `PwInstance` to act as a manager that maintains a registry of active pages (`map[string]Page`). This decouples "Navigation" from "Fetching" and allows multiple tabs to exist simultaneously within a shared browser context.

## Decision Outcome

Chosen option: **Page Registry (Map-based State Management)**.

We will refactor `PwInstance` to manage a thread-safe collection of active pages using a `map[string]playwright.Page` protected by a `sync.RWMutex`.

**Reasoning:**

1. **Isolation:** Each URL is processed in its own dedicated tab (`Page`), preventing DOM leakage or navigation conflicts between concurrent jobs.
2. **Asynchronous Workflow:** This architecture enables a split between `GoToAsync` (Navigating/Registering) and `FetchMHTML` (Snapshotting/Cleaning). A worker can initiate navigation for multiple URLs without blocking the snapshotting process of others.
3. **Resource Efficiency:** We share a single `BrowserContext` (cookies/cache shared) but multiply the `Page` instances (tabs), which is lighter than multiplying Contexts.
4. **State Control:** The registry pattern allows us to track which URLs are currently "in-flight" (being processed).

## Technical Implementation Details

The `PwInstance` struct will be updated to include:

- `pages`: A `map[string]playwright.Page` to store active pages keyed by their target URL.
- `pageMu`: A `sync.RWMutex` to ensure thread-safety when reading/writing to the map.

**Workflow:**

1. **GoToAsync:** Opens a new tab, locks the map, registers the page (`pages[url] = page`), unlocks, and initiates navigation in a goroutine.
2. **FetchMHTML:** Accepts a URL, locks the map (Read), retrieves the corresponding `Page`, performs the MHTML snapshot via CDP, closes the page, and removes the entry from the map (Cleanup).

## Consequences

- **Good:** Significantly higher throughput (parallel crawling).
- **Good:** Thread-safety is enforced via Mutex, preventing panic conditions during concurrent map access.
- **Good:** Modular design allows separate timeouts/retries for navigation vs. snapshotting.
- **Bad:** Increased complexity in error handling. If `GoToAsync` fails or panics, we must ensure the "orphan" page is removed from the registry to prevent memory leaks.
- **Bad:** Higher memory consumption as multiple tabs are open simultaneously.
- **Accepted Complexity:** We explicitly accept the responsibility of managing the lifecycle (Open -> Register -> Close -> Delete) of every page manually.

## Future Work

- [x] **Worker Pool Limits:** Implement a semaphore or worker pool pattern to limit the maximum number of concurrent tabs (e.g., max 10 tabs) to prevent OOM (Out of Memory) crashes.
- [ ] **Orphan Cleanup:** Implement a background routine or timeout mechanism to clean up pages in the registry that hang indefinitely during navigation.
