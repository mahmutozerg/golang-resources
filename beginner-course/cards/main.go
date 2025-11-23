package main

func main() {

	dck := newDeck()

	hand, remainingDeck := deal(dck, 5)

	hand.print()
	remainingDeck.print()

}

func newCard() string {

	return "Ace of Spades"
}
