package main

import (
	"fmt"
	"log"
	Ring "toy_dynamodb/ring"
)

func main() {
	ring := Ring.Ring{ReplicaCount: 3}
	ring.Init()

	nodes := []string{
		"localhost:50051",
		"localhost:50052",
		"localhost:50053",
	}

	for _, addr := range nodes {
		if err := ring.AddNode(addr); err != nil {
			log.Fatalf("Connection Error %s: %v", addr, err)
		}
		fmt.Printf("Connected to : %s\n", addr)
	}

	fmt.Println("Writing Data with w =2 (Mahmut = Ozer)...")
	err := ring.Put("Mahmut", "Ozer", 2)
	if err != nil {
		log.Fatalf("Write Error:  %v", err)
	}

	fmt.Println("Reading Data with r =2 (Mahmut)...")

	vals, err := ring.Get("Mahmut", 2)
	if err == nil {
		fmt.Printf("Retrieved Values:\n%v\n", vals)
	} else {
		log.Fatalf("Read Error: %v", err)
	}
}
