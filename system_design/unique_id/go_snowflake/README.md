# Go Snowflake: Distributed Unique ID Generator

A high-performance, thread-safe, and "Zero-Allocation" implementation of the Twitter Snowflake algorithm written in Go.

## Features

- **64-Bit Unique ID:** `int64` based, roughly time-sorted (k-ordered) IDs.
- **High Performance:** Capable of generating **~3.6 Million IDs/second** on a single core.
- **Thread-Safe:** Protected against concurrent requests using `sync.Mutex`.
- **Zero-Allocation:** Designed to be Garbage Collector friendly with no unnecessary memory allocations.
- **Clock Rollback Protection:** Provides basic protection against NTP clock regressions.

## Installation & Usage

```bash
go get github.com/mahmutozerg/golang-resources/system_design/unique_id/go_snowflake@latest


```

### Example Usage

```go
package main

import (
	"fmt"
	"time"

	snowflake "github.com/mahmutozerg/golang-resources/system_design/unique_id/go_snowflake"
)

func main() {
	node, err := snowflake.NewNode(1, 1)
	if err != nil {
		panic(err)
	}

	idMap := make(map[int64]bool)
	count := 1000000

	start := time.Now()

	for i := 0; i < count; i++ {
		id, err := node.NextId()
		if err != nil {
			panic(err)
		}

		if _, exists := idMap[id]; exists {
			fmt.Printf("Collision Tespit Edildi ID: %d\n", id)
			fmt.Printf("Sequence doldu ama zaman ilerlemedi\n")
			return
		}
		idMap[id] = true
	}

	duration := time.Since(start)

	fmt.Printf("%d benzersiz ID üretildi.\n", len(idMap))
	fmt.Printf("Geçen Süre: %v\n", duration)
}

```

## ID Structure (Twitter Standard)

| 1 Bit           | 41 Bits           | 5 Bits        | 5 Bits     | 12 Bits  |
| --------------- | ----------------- | ------------- | ---------- | -------- |
| Sign (Always 0) | Timestamp (Epoch) | Datacenter ID | Machine ID | Sequence |

## Benchmark Results

Running the duplicate check on a standard machine:

```text
--- STARTING DUPLICATE CHECK ---
1000000 unique IDs generated.
Duration: 271.984423ms (~3.6M ID/sec)

```
