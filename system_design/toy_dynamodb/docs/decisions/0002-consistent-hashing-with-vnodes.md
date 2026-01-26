# Consistent hashing ring with virtual nodes

## Context and Problem Statement
We need to map keys to nodes in a way that balances load, supports adding nodes, and allows replication.

Current implementation:
- Uses SHA1 hash; takes first 8 bytes as uint64
- Uses 100 virtual spots per physical node
- Maintains sorted ring positions (sortedNodes) and a map position->nodeName (nodeMap)
- For a key, finds the first ring position >= key hash, then walks clockwise collecting distinct node names

## Decision Drivers
- Even distribution across nodes
- Simple, deterministic mapping
- Replica selection should avoid duplicate physical nodes

## Considered Options
1. No virtual nodes (one token per node)
2. Virtual nodes per physical node (vnodes)
3. Rendezvous hashing

## Decision Outcome
Chosen option: "Consistent hashing with vnodes" (VirtualSpotCount=100), because it improves balance while staying simple.

## Consequences
- Adding a node appends 100 positions and re-sorts the ring
- getNode() must deduplicate physical nodes when walking the ring (seenSet)
- Hash choice (SHA1 -> 64-bit) is a design decision; collisions are possible in theory but acceptable for a toy project
