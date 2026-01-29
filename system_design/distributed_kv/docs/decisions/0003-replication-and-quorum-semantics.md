# Replication and quorum semantics

## Context and Problem Statement
We replicate each key to multiple nodes and allow clients to choose read/write quorum sizes.

Current behavior:
- ReplicaCount (N) is configured on Ring (e.g., 5)
- For a key, ring.getNode(key, N) returns up to N *distinct* physical nodes
- Put(key,val,w):
  - launches writes to all replica nodes concurrently
  - returns success once it receives w successful acknowledgements
  - fails if all replica attempts return and successes < w
- Get(key,q):
  - launches reads to all replica nodes concurrently
  - counts a replica as successful only if key exists (ok==true)
  - returns success once it receives q successful "found" responses, returning a map[nodeName]value
  - fails with "not found at any node" only if all replicas respond not-found
  - otherwise fails with "Failed to hit quorum" if responses complete but found < q

## Decision Drivers
- Availability under partial node failures
- Clear definition of what counts as "success" for reads/writes
- Simple behavior that matches Dynamo-style knobs (R/W)

## Considered Options
1. Require all replicas (R=N, W=N)
2. Quorum reads/writes (R and W configurable)
3. Primary-only writes with async replication

## Decision Outcome
Chosen option: "Configurable quorum reads/writes", because the API already exposes w and q and the code implements threshold success.

## Consequences
- Returned GET values may disagree across replicas; current API returns per-node values rather than resolving conflicts
- No read-repair or hinted handoff is implemented
- GET quorum is based on "key exists" rather than "node responded", which changes failure modes (missing key counts as failure)
