# ADR 0002: MHTML Snapshot Strategy using Playwright & CDP

## Context and Problem Statement

Following [ADR 0001](https://www.google.com/search?q=./0001-crawler-system-proposal.md), we decided to use headless browsers for scraping. Initially, saving pages as PDFs was considered. However, PDFs often alter the layout (print media query), break interactive elements, and do not faithfully represent the digital state of the DOM.

We need a format that encapsulates the HTML, CSS, Images, and current JavaScript state into a single, portable file while preserving the exact visual fidelity of the browser session.

## Considered Options

- **PDF Export:** Native Playwright feature. Good for documents, bad for preserving web layout fidelity.
- **Raw HTML (`page.content()`):** Saves the DOM text but creates "broken" local files because relative links to CSS/Images are lost unless downloaded separately (complex file management).
- **WARC (Web ARChive):** Industry standard for archiving (used by Wayback Machine) capturing raw HTTP traffic. It ensures high fidelity and replayability but introduces significant complexity (requires proxy/interception logic) for a simple crawler.
- **MHTML (via CDP):** Uses Chrome DevTools Protocol to snapshot the page into a single file containing all resources.

## Decision Outcome

Chosen option: **MHTML (via CDP)**.

We will use the **Chrome DevTools Protocol (CDP)** command `Page.captureSnapshot` accessed through `playwright-go`.

**Reasoning:**

1. **Single File Portability:** Like PDF, MHTML creates one file per page, making storage and transfer easy.
2. **Fidelity:** Unlike PDF, MHTML preserves the browser view (screen media query) exactly as rendered, including dynamic DOM changes.
3. **Efficiency:** It leverages the browser's internal serialization mechanism rather than requiring us to manually fetch and rewrite asset paths.

## Consequences

- **Good:** Complete preservation of the page state (DOM + Assets) in a single artifact.
- **Good:** No need for complex asset scraping logic (downloading separate `.css`, `.jpg` files).
- **Bad:** MHTML is primarily a Chromium/IE format. While most modern browsers can open it, it is less universal than standard HTML or PDF.
- **Bad:** Requires `CDP` session management, which is a lower-level API than standard Playwright methods.
- **Accepted Technical Debt:** We explicitly accept that MHTML does not capture network-level data (HTTP headers, exact status codes, original timestamps) which creates a dependency on Chromium for rendering and makes "forensic" replay difficult compared to WARC.

## Future Work

- [ ] **Evaluate WARC Migration:** Investigate migrating from MHTML to **WARC (Web ARChive)** format if the project requirements shift towards "legal archiving" or "network traffic replayability".
- [ ] **Storage Abstraction Layer:** Implement an interface that allows switching between output formats (MHTML, WARC, PDF) via configuration, decoupling the crawler logic from the storage format.
