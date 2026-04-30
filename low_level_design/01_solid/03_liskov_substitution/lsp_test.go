package liskov

import (
	"math"
	"testing"
)

func TestRectangle_Area(t *testing.T) {
	r := Rectangle{Width: 5, Height: 3}
	if got := r.Area(); got != 15 {
		t.Errorf("Area() = %v, want 15", got)
	}
	if got := r.Perimeter(); got != 16 {
		t.Errorf("Perimeter() = %v, want 16", got)
	}
}

func TestCircle_Area(t *testing.T) {
	c := Circle{Radius: 7}
	wantArea := math.Pi * 49
	if got := c.Area(); math.Abs(got-wantArea) > 0.01 {
		t.Errorf("Area() = %v, want %v", got, wantArea)
	}
}

func TestTriangle_Area(t *testing.T) {
	tri := Triangle{Base: 10, Height: 5, SideA: 10, SideB: 7, SideC: 7}
	if got := tri.Area(); got != 25 {
		t.Errorf("Area() = %v, want 25", got)
	}
	if got := tri.Perimeter(); got != 24 {
		t.Errorf("Perimeter() = %v, want 24", got)
	}
}

func TestSquare_Area(t *testing.T) {
	s := Square{Side: 4}
	if got := s.Area(); got != 16 {
		t.Errorf("Area() = %v, want 16", got)
	}
	if got := s.Perimeter(); got != 16 {
		t.Errorf("Perimeter() = %v, want 16", got)
	}
}

// TotalArea must work with ANY Shape — this is LSP in action.
func TestTotalArea_AllShapes(t *testing.T) {
	shapes := []Shape{
		Rectangle{Width: 2, Height: 3}, // 6
		Circle{Radius: 1},              // π
		Square{Side: 5},                // 25
	}

	got := TotalArea(shapes)
	want := 6 + math.Pi + 25
	if math.Abs(got-want) > 0.01 {
		t.Errorf("TotalArea() = %v, want %v", got, want)
	}
}

func TestIsLargerThan(t *testing.T) {
	big := Rectangle{Width: 10, Height: 10} // 100
	small := Circle{Radius: 1}              // π ≈ 3.14

	if !IsLargerThan(big, small) {
		t.Error("expected Rectangle(10x10) > Circle(r=1)")
	}
	if IsLargerThan(small, big) {
		t.Error("expected Circle(r=1) < Rectangle(10x10)")
	}
}
