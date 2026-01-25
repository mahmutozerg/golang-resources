package main

import (
	"fmt"
	"log"
	Ring "toy_dynamodb/ring"
)

func main() {
	// Ring başlatılır ve Replikasyon Sayısı 5 olarak ayarlanır
	ring := Ring.Ring{RCount: 5}
	ring.Init()

	// Node'lar güvenli bir şekilde eklenir
	nodes := []string{"foo", "bar", "bazz", "zapp", "zucc", "bars"}
	for _, n := range nodes {
		if err := ring.AddNode(n); err != nil {
			log.Fatalf("%v", err)
		}
	}

	// Veri yazma (Coordinator 5 replikaya da yazar)
	ring.Put("Mahmut", "Ozer")

	// Veri okuma (Coordinator replikalardan veriyi çeker)
	vals, found := ring.Get("Mahmut")

	if found {
		fmt.Println(vals)
	}
}
