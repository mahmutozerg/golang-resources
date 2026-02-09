# ADR 0004: Multi-Seed Architecture & Host-Based Rate Limiting Strategy

## Status

**Accepted** (Implemented in v1.0.0)

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

- **Option 3: Dynamic Policy Registry (Chosen)**
  - _Description:_ A centralized, thread-safe registry that stores policy objects (`DomainPolicy`) keyed by the Host. Workers pull the specific policy for their current URL from this registry.
  - _Pros:_ Decouples workers from domains. A fixed worker pool (e.g., 20) can handle an infinite number of domains efficiently.
  - _Cons:_ Introduces state management complexity (`sync.RWMutex` handling).

## Decision Outcome

Chosen option: **Option 3: Dynamic Policy Registry**.

We have implemented a `Registry` (Policy Checker) system. This system lazily initializes and caches policies for each unique host encountered during the crawl.

**Key Architectural Decisions:**

1.  **Registry Pattern:** A `map[string]*DomainPolicy` protected by `sync.RWMutex`.
    - **Key:** The URL Host (e.g., `www.archlinux.org`).
    - **Value:** A struct containing the dedicated `*rate.Limiter` and `*robotstxt.Group`.

2.  **Dependency Injection via Closures:**
    - To avoid circular dependencies between the `policy` package and the `fetcher` package, the `policy` package defines a `FetcherFunc` type. The `fetcher` implementation is injected into the `Registry` constructor as a closure.

3.  **Optimized Lazy Initialization:**
    - Network operations (fetching `robots.txt`) are performed **outside** the write lock to prevent blocking other workers reading existing policies.

4.  **Worker Autonomy:**
    - Workers retrieve the specific limiter for their target host and execute `hostLimiter.Wait(ctx)`. Global limiting is removed.

## Technical Design

The core data structures are defined as follows:

```go
// DomainPolicy holds the policy for a specific domain
type DomainPolicy struct {
    Rule    *robotstxt.Group // Parsed rules (Disallow/Allow)
    Limiter *rate.Limiter    // Dedicated rate limiter per host
}

// Checker (Registry) manages the concurrent access to domain data
type Checker struct {
    hostPolicies map[string]*DomainPolicy
    // The rest remains unchanged
}
```

## Consequences

- **Positive:** **True Parallelism:** `site-a.com` logic is completely isolated from `site-b.com`.
- **Positive:** **Non-Blocking I/O:** Moving the `robots.txt` fetch outside the mutex ensures that network latency does not block the registry for other readers.
- **Positive:** **Robustness:** If `robots.txt` cannot be fetched or parsed, the system falls back to a safe default policy (Allow All + Default Rate Limit).
- **Negative:** **Memory Growth:** The map grows with every visited unique domain.
- **Negative:** **Initial Latency:** The first visit to a new domain incurs a "cold start" penalty for fetching rules.

## Future Work

- [ ] **LRU Eviction:** Implement a mechanism to remove policies for stale domains to cap memory usage.
- [ ] **Persisted Cache:** Store parsed `robots.txt` rules on disk to survive restarts.
