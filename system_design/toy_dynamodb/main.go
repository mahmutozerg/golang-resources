package main

import (
	"fmt"
	"log"
	Ring "toy_dynamodb/ring"
)

func main() {
	// Ring başlatılır ve Replikasyon Sayısı 5 olarak ayarlanır
	ring := Ring.Ring{ReplicaCount: 5}
	ring.Init()

	// Node'lar güvenli bir şekilde eklenir
	nodes := []string{"foo", "bar", "bazz", "zapp", "zucc", "bars"}
	for _, n := range nodes {
		if err := ring.AddNode(n); err != nil {
			log.Fatalf("%v", err)
		}
	}

	ring.Put("Mahmut", "Ozer", 3)

	// Veri okuma (Coordinator replikalardan veriyi çeker)
	vals, err := ring.Get("Mahmut", 1)

	if err == nil {
		fmt.Println(vals)
	} else {
		log.Fatalf("%v", err)
	}
}
