# Use Write-Ahead Log (WAL) with Base64 Encoding for Persistence

## Context and Problem Statement

Our Distributed Key-Value Store nodes currently store data strictly in memory (RAM). This means any process restart or crash results in complete data loss. We need a mechanism to ensure data durability so that the system state can be restored upon restart.

Additionally, the in-memory map acts as the primary data store. We need to persist operations to disk in an append-only format. However, values stored in the system may contain special characters (like commas or newlines) which would break a simple Comma-Separated Values (CSV) parser in the log file.

How can we ensure strict data durability while maintaining data integrity against parsing errors?

## Decision Drivers

* **Durability:** A successful write acknowledgement must guarantee that data is saved to persistent storage and will survive a restart.
* **Data Integrity:** The log file format must safely handle arbitrary characters (e.g., `,`, `\n`) in values without corruption during replay.
* **Consistency:** The in-memory state must never diverge from the persistent state. If disk write fails, memory must not change.
* **Ordering:** The order of operations in the log must strictly match the order of application in memory.

## Considered Options

* **In-memory only:** No persistence.
* **Write-Back (Async):** Write to RAM first, flush to disk periodically.
* **Plain Text WAL (Write-Through):** Write to disk first as CSV, then RAM.
* **Base64 Encoded WAL (Write-Through):** Write to disk first using Base64 encoding for values, then RAM.

## Decision Outcome

Chosen option: **"Base64 Encoded WAL (Write-Through)"**, because

* It provides strict durability guarantees (no data loss on confirmed writes).
* It solves the parsing issue for special characters (commas/newlines) by encoding values.
* It ensures the log file is always the "Source of Truth".

### Implementation Details

The system will function as a **Log-Structured Storage Engine**:

1. **Record Format:** `COMMAND,KEY,OPTIONAL_VALUE_BASE64\n`
* *Example SET:* `SET,username,bWFobXV0\n`
* *Example DEL:* `DEL,username\n` (Tombstone record)


2. **Write Path (Put/Del):**
* **Lock** (Mutex)
* **Encode** Value (Base64) - *Only for SET*
* **Append** to File
* **fsync** (Flush to physical disk)
* **Update Memory** (Set or Delete in Map) **only if fsync succeeds**
* **Unlock**


3. **Recovery Path (Init):**
* **Open File** (Read-Only)
* **Read Line-by-Line**
* **Parse:** Split by `,`
* **Replay:**
* If `SET`: Decode Base64 -> Insert into Map
* If `DEL`: Delete from Map


* **Close & Re-open** File in Append mode



### Positive Consequences

* **Robustness:** Parsing is immune to delimiter collision; binary data can be stored safely.
* **Strict Durability:** Using `fsync` before memory update ensures zero data loss for acknowledged writes.
* **Simplicity:** The append-only log structure is easy to implement and debug compared to complex binary formats.

### Negative Consequences

* **Performance:** Write latency increases significantly due to the requirement of blocking `fsync` calls for every operation.
* **Storage Overhead:** Base64 encoding increases the size of the log file on disk by approximately 33%.
* **Disk Usage:** Since it is an append-only log without compaction, the file size will grow indefinitely over time (requires future "Log Compaction" mechanism).