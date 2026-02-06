# Web Crawler Demo Architectural Decisions

## Context and Problem Statement

We are building a simple web crawler that saves pages as PDF files and also captures any JSON (or related file formats).

## Considered Options

- Save pages by sending curl requests
- Save pages by using headless browsers
- Save pages by using headless browsers + parallelism

## Decision Outcome

Chosen option: `Save pages by using headless browsers`, because it is easy to implement as a small demo.

### Future work

- [ ] Implement BFS using a queue
- [x] Implement a multi-threaded crawler
- [ ] Implement rate limiting and robots.txt awareness

### Consequences

- Good: Modern pages that heavily rely on JS can be saved.
- Good: For demo/POC, single-page retrieval is trivial.
- Bad: It will require some refactoring to migrate to BFS & multi-threaded options.
