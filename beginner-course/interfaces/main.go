package main

import "fmt"

/*
Goda'ki interface c#tan baya farklı,
bot GetGreeting()string fonksiyon prototipini karşılayan her fonksiyon (receiver olması lazım)
interface'yi implemente etmiş kabul ediliyor.
*/
type bot interface {
	GetGreeting() string
}

type englishBot struct{}
type spanishBot struct{}

func main() {

	eb := englishBot{}
	sb := spanishBot{}

	printGreeting(eb)
	printGreeting(sb)
}

func printGreeting(b bot) {
	fmt.Println(b.GetGreeting())
}

func (englishBot) GetGreeting() string {
	return "helllo user"
}

func (spanishBot) GetGreeting() string {
	return "hola"
}
