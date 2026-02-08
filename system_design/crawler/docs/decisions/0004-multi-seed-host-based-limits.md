# ADR 0004: Multi-Seed Architecture & Host-Based Rate Limiting Strategy

## Status

**Proposed** (Targeting Release v1.0.0)

## Context and Problem Statement

Currently, the crawler operates with a **Global Rate Limiter** and a single `robots.txt` checker instance. While this architecture is robust for single-domain crawling, it presents a significant blocking issue when multiple seed URLs from different domains are provided (Multi-Seed).

**The Bottleneck:**
If the seed list contains disparate domains (e.g., `google.com` and `yahoo.com`) and one domain requires a high `Crawl-delay` (e.g., 5 seconds), the global limiter halts the entire crawling process. Fast domains are forced to wait for the delays imposed by slow domains, resulting in severe throughput degradation ("Head-of-Line Blocking").

We need to architect a system where concurrency is managed **per host**, allowing the crawler to process multiple domains in parallel while strictly adhering to each domain's individual `robots.txt` policies.

## Considered Options

- **Option 1: Global Limiter (Status Quo)**
  - _Pros:_ Simple implementation.
  - _Cons:_ Unsuitable for multi-seed operations; slow domains block fast ones.

- **Option 2: Fixed Worker-per-Domain**
  - _Description:_ Spawn a dedicated goroutine loop for every new domain found.
  - _Pros:_ Strict isolation.
  - _Cons:_ Not scalable. Crawling 10,000 unique domains would spawn 10,000 goroutines, causing high memory usage and scheduler thrashing.

- **Option 3: Dynamic Policy Registry (Proposed)**
  - _Description:_ A centralized, thread-safe registry that stores policy objects (`RateLimiter` + `RobotRules`) keyed by the Host. Workers pull the specific policy for their current URL from this registry.
  - _Pros:_ Decouples workers from domains. A fixed worker pool (e.g., 20) can handle an infinite number of domains efficiently.
  - _Cons:_ Introduces state management complexity (`sync.RWMutex` handling).

## Decision Outcome

Chosen option: **Option 3: Dynamic Policy Registry**.

We propose to implement a `DomainRegistry` system for v1.0. This system will lazily initialize and cache policies for each unique host encountered during the crawl.

**Proposed Architectural Changes:**

1.  **Registry Pattern:** Introduce a `map[string]*DomainData` protected by `sync.RWMutex`.
    - **Key:** The URL Host (e.g., `www.archlinux.org`).
    - **Value:** A struct containing the dedicated `*rate.Limiter` and `*robot.Group`.

2.  **Lazy Initialization (Just-In-Time):**
    - Policies will not be pre-loaded.
    - When a worker picks up a URL, it will query the registry.
    - If the host is new, the worker will (under a lock) fetch `robots.txt`, calculate the specific delay, instantiate a new `rate.Limiter`, and store it.

3.  **Worker Autonomy:**
    - Workers will no longer wait on a global `limiter.Wait()`.
    - Instead, they will retrieve the specific limiter for their target host and execute `hostLimiter.Wait(ctx)`.

## Technical Design (Draft)

We intend to define a structure similar to:

```go
// DomainData holds the policy for a specific domain
type DomainData struct {
    Limiter     *rate.Limiter
    RobotRules  *robot.Group
    LastVisited time.Time
}

// Registry manages the concurrent access to domain data
type Registry struct {
    data map[string]*DomainData
    mu   sync.RWMutex
}
```

### The "GetOrInit" Flow

To handle high concurrency safely, we will implement the **Double-Checked Locking** pattern:

1. **Read Lock:** Check if host exists. If yes, return.
2. **Write Lock:** If not, lock the map.
3. **Re-Check:** Verify existence again (to handle race conditions).
4. **Fetch & Create:** Download `robots.txt`, parse rules, create `rate.Limiter`.
5. **Store:** Save to map and Unlock.

## Consequences

- **Positive:** **True Parallelism:** `site-a.com` being slow will no longer block `site-b.com`.
- **Positive:** **Politeness:** We will respect `Crawl-delay` individually for every site.
- **Positive:** **Resource Efficiency:** Memory usage grows with _active_ domains, not total URLs.
- **Negative:** **Memory Growth:** Long-running crawls visiting millions of unique domains may bloat the map (requires future LRU eviction strategy).
- **Negative:** **Initial Latency:** The first request to any new domain will incur the overhead of fetching `robots.txt`.

## Future Work (Post v1.0)

- [ ] **Persisted Registry:** Save known `robots.txt` rules to disk/DB to avoid re-fetching on restart.
- [ ] **LRU Eviction:** Implement a mechanism to remove policies for stale domains to save RAM.
