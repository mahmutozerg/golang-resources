# Record architecture decisions

## Context and Problem Statement
We are building a toy DynamoDB-like KV store with a consistent hashing ring, replication, and quorum reads/writes.
Important behavior is encoded in Go code, but the reasons and guarantees are not documented.

## Decision Drivers
- Keep decisions close to code and reviewable via PRs
- Make correctness guarantees explicit (replication, quorum, membership changes)
- Help new contributors understand design intent quickly

## Considered Options
1. Do nothing
2. Record ADRs using MADR-style markdown in the repo
3. Use an external wiki

## Decision Outcome
Chosen option: "Record ADRs in-repo under docs/decisions/", because it keeps decisions versioned with code and easy to review.

## Consequences
- PRs that change distributed behavior should include an ADR update/new ADR
- ADRs become an entry point for understanding system behavior
