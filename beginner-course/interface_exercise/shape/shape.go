package shape

import "fmt"

type Shape interface {
	GetArea() float64
}

type Triangle struct {
	SideLength float64
	BaseLength float64
}

type Square struct {
	SideLength float64
}

func PrintArea(s Shape) {
	fmt.Println(s.GetArea())
}

func (s Square) GetArea() float64 {
	return s.SideLength * s.SideLength
}

func (t Triangle) GetArea() float64 {
	return (t.BaseLength * t.SideLength) / 2
}
