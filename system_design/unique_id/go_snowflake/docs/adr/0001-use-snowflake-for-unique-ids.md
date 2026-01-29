# Use Snowflake Algorithm for Distributed Unique ID Generation

## Context and Problem Statement

We need a system to generate unique identifiers for distributed objects (e.g., database rows, event logs).
Standard approaches like UUIDs are too long (128-bit string) and unorderable (random), causing database index fragmentation.
Database auto-increment IDs do not scale horizontally (single point of failure).

We need IDs that are:

1. **Unique** across all distributed nodes.

2. **Time-sortable** (roughly ordered by creation time).

3. **Numerical** (64-bit integers for efficiency).

4. **High performance** (millions of IDs per second).

## Decision Outcome

Chosen option: **"Twitter Snowflake Algorithm"**, because

- It generates **int64** IDs that fit into standard database integer fields.

- It is **k-ordered** (mostly ordered by time), which is friendly to B-Tree indexes.

- It allows distributed generation without coordination (no central server needed).

### ID Structure (64 Bits)

| 1 Bit         | 41 Bits           | 10 Bits | 12 Bits  |
| ------------- | ----------------- | ------- | -------- |
| Sign (Unused) | Timestamp (Epoch) | Node ID | Sequence |

- **Sign:** Always 0.

- **Timestamp:** Milliseconds since custom epoch.

- **Node ID:** Configurable per instance (0-1023).

- **Sequence:** Incremented for IDs generated within the same millisecond (0-4095).

### Consequences

- **Good:** IDs are sortable by time.

- **Good:** Generation is extremely fast (local bitwise operations).

- **Bad:** Depends on system clock. Clock rollbacks can cause ID collisions or system pauses.
