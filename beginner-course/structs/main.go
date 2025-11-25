package main

import "fmt"

type person struct {
	firstName string
	lastName  string
	contactInfo
}

type contactInfo struct {
	email string
	zip   int
}

func main() {
	newPerson := person{
		firstName: "Mahmut",
		lastName:  "Gozukirmizi",
		contactInfo: contactInfo{
			email: "mahmutozerg@gmail.com",
			zip:   34310,
		},
	}

	newPerson.updateName("Mahmut Ozer")
	newPerson.print()

}

func (p person) print() {
	fmt.Printf("%+v", p)

}

func (p *person) updateName(name string) {
	p.firstName = name
}
