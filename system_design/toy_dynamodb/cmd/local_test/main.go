package main

import (
	"fmt"
	"log"
	"toy_dynamodb/adapter"
	"toy_dynamodb/node"
	"toy_dynamodb/ring"
)

func main() {
	// 1. Ring'i Başlat
	r := ring.Ring{ReplicaCount: 3}
	r.Init()

	node1, _ := node.New("memory-node-1")
	node2, _ := node.New("memory-node-2")
	node3, _ := node.New("memory-node-3")

	r.RegisterClient("node-1", adapter.NewLocalClient(node1))
	r.RegisterClient("node-2", adapter.NewLocalClient(node2))
	r.RegisterClient("node-3", adapter.NewLocalClient(node3))

	fmt.Println("🚀 Sistem 'In-Memory Mock' modunda başlatıldı!")

	err := r.Put("Key", "Value", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ Yazıldı.")

	fmt.Println("--- Okuma Testi ---")
	vals, err := r.Get("Key", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Okunan: %v\n", vals)
}
