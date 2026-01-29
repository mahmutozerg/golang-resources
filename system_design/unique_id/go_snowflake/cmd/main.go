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
