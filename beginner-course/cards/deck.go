package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Create a new type of 'deck'
// Which is a slice of strings

type deck []string

// d = değişkenin adresi shallow copy
// deck = veri tipi
// bildiğin this,self
// genel olarak 1 veya 2 harfle ifade edilir

func (d deck) print() {

	for i, card := range d {
		fmt.Println(i, card)
	}

}
func newDeck() deck {
	cards := deck{}

	cardSuits := []string{"Spades", "Kings", "Heards", "Clubs"}
	cardValues := []string{"Ace", "Two", "Three", "Four"}

	// for suit := range cardSuits {
	// 	for val := range cardValues {
	// 		cards = append(cards, cardValues[val]+" of "+cardSuits[suit])
	// 	}
	// }

	for _, suit := range cardSuits {
		for _, val := range cardValues {
			cards = append(cards, val+" of "+suit)
		}
	}
	return cards
}

func deal(d deck, handSize int) (deck, deck) {

	//hand, remaining cards
	return d[:handSize], d[handSize:]
}

func (d deck) toString() string {
	return strings.Join([]string(d), ",")

}
func (d deck) saveToFile(fileName string) error {

	return os.WriteFile(fileName, []byte(d.toString()), 0666)
}

func newDeckFromFile(fileName string) deck {

	fileStatus, err := os.ReadFile(fileName)

	if err != nil {
		log.Fatal("Failed to read file", err.Error())
		os.Exit(1)
	}

	return deck(strings.Split(string(fileStatus), ","))

}
