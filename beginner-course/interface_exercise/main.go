package main

import (
	"io"
	"log"
	"os"
)

//import "interface_exercise/shape"

func main() {

	// s := shape.Square{SideLength: 4}
	// t := shape.Triangle{BaseLength: 10, SideLength: 5}

	// shape.PrintArea(s)
	// shape.PrintArea(t)

	if len(os.Args) != 2 {
		log.Fatalf("Must supplement a file name %v", len(os.Args))
	}

	f, err := os.Open(os.Args[1])

	if err != nil {
		log.Fatalf("%v", err)
	}

	io.Copy(os.Stdout, f)

}
