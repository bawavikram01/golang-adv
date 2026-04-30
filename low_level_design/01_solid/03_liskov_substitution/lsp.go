// Package liskov demonstrates the Liskov Substitution Principle.
//
// PRINCIPLE: If S is a subtype of T, then objects of type T can be replaced
// with objects of type S without breaking correctness.
//
// In Go terms: Any implementation of an interface must honor the full contract
// of that interface — not just the method signatures, but the BEHAVIOR.
//
// CLASSIC VIOLATION: Square implementing Rectangle. Changing width also
// changes height, breaking the expectation that they are independent.
//
// GOOD EXAMPLE: Shape interface where Area() is always correct regardless
// of the concrete type.
package liskov

import "math"

// ──────────────────────────────────────────────
// The interface contract
// ──────────────────────────────────────────────

// Shape defines the behavioral contract:
// - Area() must return the correct geometric area (≥ 0)
// - Perimeter() must return the correct perimeter (≥ 0)
type Shape interface {
	Area() float64
	Perimeter() float64
}

// ──────────────────────────────────────────────
// Implementations — all substitutable for Shape
// ──────────────────────────────────────────────

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

type Triangle struct {
	Base, Height        float64
	SideA, SideB, SideC float64
}

func (t Triangle) Area() float64      { return 0.5 * t.Base * t.Height }
func (t Triangle) Perimeter() float64 { return t.SideA + t.SideB + t.SideC }

// ──────────────────────────────────────────────
// Code that depends on Shape — works with ANY implementation
// ──────────────────────────────────────────────

// TotalArea sums the area of any slice of Shapes.
// This function NEVER needs to know the concrete type.
// If any implementation violates the contract, this breaks.
func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// IsLargerThan compares two shapes by area — works for any pair.
func IsLargerThan(a, b Shape) bool {
	return a.Area() > b.Area()
}

// ──────────────────────────────────────────────
// VIOLATION EXAMPLE (what NOT to do)
// ──────────────────────────────────────────────

// BadSquare violates LSP because SetWidth also changes height.
// Code expecting a Rectangle's Width and Height to be independent breaks.
type BadSquare struct {
	side float64
}

func (s *BadSquare) SetWidth(w float64)  { s.side = w } // also changes "height"
func (s *BadSquare) SetHeight(h float64) { s.side = h } // also changes "width"
func (s *BadSquare) Area() float64       { return s.side * s.side }
func (s *BadSquare) Width() float64      { return s.side }
func (s *BadSquare) Height() float64     { return s.side }

// CorrectSquare — just use a Rectangle with equal sides, or a separate type
// that doesn't pretend to have independent width/height.
type Square struct {
	Side float64
}

func (s Square) Area() float64      { return s.Side * s.Side }
func (s Square) Perimeter() float64 { return 4 * s.Side }
