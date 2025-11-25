package main

import "fmt"

func testNumber(nl []int) {
	for _, n := range nl {
		if n%2 == 0 {
			fmt.Println(n, " is even")
		} else {
			fmt.Println(n, " is odd")
		}
	}
}
func main() {

	nl := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	testNumber(nl)
}
