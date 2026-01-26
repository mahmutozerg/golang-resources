# Concurrency and locking strategy

## Context and Problem Statement
The ring and nodes are accessed concurrently:
- Ring membership changes (AddNode)
- Read/write operations (Get/Put) which dispatch goroutines to nodes

Current behavior:
- Ring has RWMutex:
  - getNode holds RLock while reading sortedNodes/nodeMap
  - AddNode uses a read-check then write-lock (double-check to avoid TOCTOU)
- Node has RWMutex:
  - Put uses Lock
  - Get uses RLock
- Put/Get spawn goroutines per replica and aggregate results via buffered channels

## Decision Drivers
- Safety under concurrent access
- Keep locks coarse enough to be correct, not so coarse to block all reads
- Avoid deadlocks and races

## Considered Options
1. Single global mutex for everything
2. RWMutex at ring + RWMutex per node (current)
3. Lock-free ring reads with atomic snapshots (more complex)

## Decision Outcome
Chosen option: "RWMutex at ring and nodes", because it keeps implementation simple and safe.

## Consequences
- Membership changes require write-lock and may block concurrent getNode briefly
- Quorum operations may return while goroutines still run (acceptable for toy; consider cancellation later)
- Ring.Init must be called before AddNode/Put/Get; otherwise behavior is undefined
