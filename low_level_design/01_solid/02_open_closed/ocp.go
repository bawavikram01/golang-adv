// Package openclosed demonstrates the Open/Closed Principle.
//
// PRINCIPLE: Software entities should be open for extension but closed for
// modification. You add new behavior by adding new types — not by editing
// existing code.
//
// Go idiom: Define an interface. Existing code programs against the interface.
// New behavior = new struct that satisfies the interface. Zero existing code
// changes.
//
// EXAMPLE: A discount calculator that supports multiple discount strategies.
// Adding a new discount type requires ZERO changes to existing code.
package openclosed

import "fmt"

// ──────────────────────────────────────────────
// The abstraction — open for extension
// ──────────────────────────────────────────────

// DiscountStrategy calculates a discount for a given price.
// To add a new discount type, just create a new struct implementing this.
type DiscountStrategy interface {
	Calculate(price float64) float64
	Name() string
}

// ──────────────────────────────────────────────
// Concrete strategies — each is a separate extension
// ──────────────────────────────────────────────

// NoDiscount — full price.
type NoDiscount struct{}

func (d NoDiscount) Calculate(price float64) float64 { return price }
func (d NoDiscount) Name() string                    { return "No Discount" }

// PercentageDiscount — e.g., 20% off.
type PercentageDiscount struct {
	Percent float64
}

func (d PercentageDiscount) Calculate(price float64) float64 {
	return price * (1 - d.Percent/100)
}

func (d PercentageDiscount) Name() string {
	return fmt.Sprintf("%.0f%% Off", d.Percent)
}

// FlatDiscount — e.g., ₹500 off.
type FlatDiscount struct {
	Amount float64
}

func (d FlatDiscount) Calculate(price float64) float64 {
	result := price - d.Amount
	if result < 0 {
		return 0
	}
	return result
}

func (d FlatDiscount) Name() string {
	return fmt.Sprintf("₹%.0f Flat Off", d.Amount)
}

// BuyOneGetOneFree — 50% off, effectively.
type BuyOneGetOneFree struct{}

func (d BuyOneGetOneFree) Calculate(price float64) float64 { return price / 2 }
func (d BuyOneGetOneFree) Name() string                    { return "Buy 1 Get 1 Free" }

// ──────────────────────────────────────────────
// The closed part — this code NEVER changes when you add discounts
// ──────────────────────────────────────────────

// PriceCalculator uses any discount strategy. Adding a new discount
// type requires ZERO modifications here.
type PriceCalculator struct {
	strategy DiscountStrategy
}

func NewPriceCalculator(strategy DiscountStrategy) *PriceCalculator {
	return &PriceCalculator{strategy: strategy}
}

func (c *PriceCalculator) FinalPrice(price float64) float64 {
	return c.strategy.Calculate(price)
}

func (c *PriceCalculator) StrategyName() string {
	return c.strategy.Name()
}
