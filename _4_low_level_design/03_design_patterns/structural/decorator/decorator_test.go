package decorator

import "testing"

func TestEspresso_Base(t *testing.T) {
	e := Espresso{}
	if e.Cost() != 100 {
		t.Errorf("Cost() = %v, want 100", e.Cost())
	}
	if e.Description() != "Espresso" {
		t.Errorf("Description() = %q", e.Description())
	}
}

func TestDecorator_SingleTopping(t *testing.T) {
	b := WithMilk(Espresso{})
	if b.Cost() != 120 {
		t.Errorf("Espresso+Milk Cost = %v, want 120", b.Cost())
	}
	if b.Description() != "Espresso + Milk" {
		t.Errorf("Description = %q", b.Description())
	}
}

func TestDecorator_MultipleToppings(t *testing.T) {
	b := WithWhipCream(WithSugar(WithMilk(Latte{})))

	wantCost := 150.0 + 20 + 10 + 30 // Latte + Milk + Sugar + WhipCream
	if b.Cost() != wantCost {
		t.Errorf("Cost = %v, want %v", b.Cost(), wantCost)
	}

	wantDesc := "Latte + Milk + Sugar + Whip Cream"
	if b.Description() != wantDesc {
		t.Errorf("Description = %q, want %q", b.Description(), wantDesc)
	}
}

func TestDecorator_DoubleMilk(t *testing.T) {
	b := WithMilk(WithMilk(Espresso{}))
	if b.Cost() != 140 {
		t.Errorf("double milk Cost = %v, want 140", b.Cost())
	}
}

func TestOrderSummary(t *testing.T) {
	b := WithSugar(Espresso{})
	got := OrderSummary(b)
	want := "Espresso + Sugar = ₹110"
	if got != want {
		t.Errorf("OrderSummary = %q, want %q", got, want)
	}
}

func TestDecorator_LatteFullyLoaded(t *testing.T) {
	b := WithWhipCream(WithSugar(WithMilk(Latte{})))
	got := OrderSummary(b)
	want := "Latte + Milk + Sugar + Whip Cream = ₹210"
	if got != want {
		t.Errorf("OrderSummary = %q, want %q", got, want)
	}
}
