package calculator

import (
	"math"
)

type Shape interface {
	Area() float64
}

// Circle struct with radius
type Circle struct {
	Radius float64
}

// Rectangle struct with height and width
type Rectangle struct {
	Height float64
	Width  float64
}

// Area implementation for Circle
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Area implementation for Rectangle
func (r Rectangle) Area() float64 {
	return r.Height * r.Width
}

// CalculateTotalArea takes a slice of Shape and returns the sum of their areas
func CalculateTotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, shape := range shapes {
		total += shape.Area()
	}
	return total
}
