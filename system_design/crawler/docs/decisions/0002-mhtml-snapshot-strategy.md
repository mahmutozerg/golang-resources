# ADR 0002: MHTML Snapshot Strategy using Playwright & CDP

## Context and Problem Statement

Following [ADR 0001](./0001-crawler-system-proposal.md), we decided to use headless browsers for scraping. Initially, saving pages as PDFs was considered. However, PDFs often alter the layout, break interactive elements, and do not faithfully represent the digital state of the DOM.

We need a format that encapsulates the HTML, CSS, Images, and current JavaScript state into a single, portable file while preserving the exact visual fidelity of the browser session.

## Considered Options

- **PDF Export:** Native Playwright feature. Good for documents, bad for preserving web layout fidelity.
- **Raw HTML (`page.content()`):** Saves the DOM text but creates "broken" local files because relative links to CSS/Images are lost unless downloaded separately (complex file management).
- **WARC (Web ARChive):** Industry standard for archiving but introduces significant complexity and overhead for a simple crawler.
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
- **Technical Debt:** We are implicitly depending on Chromium-based browsers. Firefox/WebKit implementations in Playwright do not support CDP in the same way.

---
