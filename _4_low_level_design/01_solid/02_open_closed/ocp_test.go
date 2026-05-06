package openclosed

import "testing"

func TestDiscountStrategies(t *testing.T) {
	tests := []struct {
		name     string
		strategy DiscountStrategy
		price    float64
		want     float64
	}{
		{"no discount", NoDiscount{}, 1000, 1000},
		{"20% off", PercentageDiscount{Percent: 20}, 1000, 800},
		{"50% off", PercentageDiscount{Percent: 50}, 1000, 500},
		{"flat 200 off", FlatDiscount{Amount: 200}, 1000, 800},
		{"flat exceeds price", FlatDiscount{Amount: 1500}, 1000, 0},
		{"BOGO", BuyOneGetOneFree{}, 1000, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewPriceCalculator(tt.strategy)
			got := calc.FinalPrice(tt.price)
			if got != tt.want {
				t.Errorf("FinalPrice(%v) with %s = %v, want %v",
					tt.price, calc.StrategyName(), got, tt.want)
			}
		})
	}
}

// Demonstrate that adding a new discount requires ZERO changes to PriceCalculator.
// This "seasonal" discount didn't exist before — we just added a new struct.
type SeasonalDiscount struct {
	Percent float64
	Label   string
}

func (d SeasonalDiscount) Calculate(price float64) float64 {
	return price * (1 - d.Percent/100)
}

func (d SeasonalDiscount) Name() string { return d.Label }

// We added SeasonalDiscount without touching PriceCalculator at all
func TestExtensibility_NewDiscountWithoutModification(t *testing.T) {
	calc := NewPriceCalculator(SeasonalDiscount{Percent: 30, Label: "Diwali Sale"})
	got := calc.FinalPrice(1000)
	if got != 700 {
		t.Errorf("got %v, want 700", got)
	}
}
