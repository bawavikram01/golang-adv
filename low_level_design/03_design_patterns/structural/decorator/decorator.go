// Package decorator demonstrates the Decorator pattern.
//
// INTENT: Attach additional responsibilities to an object dynamically.
// Decorators provide a flexible alternative to subclassing for extending
// functionality.
//
// WHEN TO USE:
//   - Add behavior to objects without modifying their code
//   - Combine behaviors in mix-and-match fashion
//   - Avoid class explosion from every possible combination
//
// Go idiom: Wrap an interface with another struct implementing the same interface.
// The decorator delegates to the wrapped object, adding behavior before/after.
//
// EXAMPLE: Coffee shop — base coffee + toppings (milk, sugar, whip cream).
package decorator

import "fmt"

// ──────────────────────────────────────────────
// Component interface
// ──────────────────────────────────────────────

type Beverage interface {
	Cost() float64
	Description() string
}

// ──────────────────────────────────────────────
// Concrete components (base beverages)
// ──────────────────────────────────────────────

type Espresso struct{}

func (e Espresso) Cost() float64       { return 100 }
func (e Espresso) Description() string { return "Espresso" }

type Latte struct{}

func (l Latte) Cost() float64       { return 150 }
func (l Latte) Description() string { return "Latte" }

// ──────────────────────────────────────────────
// Decorators — each wraps a Beverage and adds to it
// ──────────────────────────────────────────────

type MilkDecorator struct {
	wrapped Beverage
}

func WithMilk(b Beverage) Beverage {
	return &MilkDecorator{wrapped: b}
}

func (m *MilkDecorator) Cost() float64 {
	return m.wrapped.Cost() + 20
}

func (m *MilkDecorator) Description() string {
	return m.wrapped.Description() + " + Milk"
}

type SugarDecorator struct {
	wrapped Beverage
}

func WithSugar(b Beverage) Beverage {
	return &SugarDecorator{wrapped: b}
}

func (s *SugarDecorator) Cost() float64 {
	return s.wrapped.Cost() + 10
}

func (s *SugarDecorator) Description() string {
	return s.wrapped.Description() + " + Sugar"
}

type WhipCreamDecorator struct {
	wrapped Beverage
}

func WithWhipCream(b Beverage) Beverage {
	return &WhipCreamDecorator{wrapped: b}
}

func (w *WhipCreamDecorator) Cost() float64 {
	return w.wrapped.Cost() + 30
}

func (w *WhipCreamDecorator) Description() string {
	return w.wrapped.Description() + " + Whip Cream"
}

// ──────────────────────────────────────────────
// Helper — build an order summary
// ──────────────────────────────────────────────

func OrderSummary(b Beverage) string {
	return fmt.Sprintf("%s = ₹%.0f", b.Description(), b.Cost())
}
