package state

import "testing"

func TestOrder_HappyPath(t *testing.T) {
	o := NewOrder("ORD-1")

	if o.State() != "Pending" {
		t.Fatalf("initial state = %q", o.State())
	}

	if err := o.Next(); err != nil {
		t.Fatalf("Pending -> Confirmed: %v", err)
	}
	if o.State() != "Confirmed" {
		t.Errorf("state = %q, want Confirmed", o.State())
	}

	if err := o.Next(); err != nil {
		t.Fatalf("Confirmed -> Shipped: %v", err)
	}
	if o.State() != "Shipped" {
		t.Errorf("state = %q, want Shipped", o.State())
	}

	if err := o.Next(); err != nil {
		t.Fatalf("Shipped -> Delivered: %v", err)
	}
	if o.State() != "Delivered" {
		t.Errorf("state = %q, want Delivered", o.State())
	}
}

func TestOrder_CannotAdvancePastDelivered(t *testing.T) {
	o := NewOrder("ORD-2")
	_ = o.Next() // -> Confirmed
	_ = o.Next() // -> Shipped
	_ = o.Next() // -> Delivered

	if err := o.Next(); err == nil {
		t.Error("expected error advancing past Delivered")
	}
}

func TestOrder_CancelFromPending(t *testing.T) {
	o := NewOrder("ORD-3")
	if err := o.Cancel(); err != nil {
		t.Fatalf("Cancel from Pending: %v", err)
	}
	if o.State() != "Cancelled" {
		t.Errorf("state = %q, want Cancelled", o.State())
	}
}

func TestOrder_CannotCancelAfterShipped(t *testing.T) {
	o := NewOrder("ORD-4")
	_ = o.Next() // -> Confirmed
	_ = o.Next() // -> Shipped

	if err := o.Cancel(); err == nil {
		t.Error("expected error cancelling shipped order")
	}
}

func TestOrder_HistoryTracking(t *testing.T) {
	o := NewOrder("ORD-5")
	_ = o.Next() // -> Confirmed
	_ = o.Next() // -> Shipped

	expected := []string{"Pending", "Confirmed", "Shipped"}
	if len(o.History) != len(expected) {
		t.Fatalf("history length = %d, want %d", len(o.History), len(expected))
	}
	for i, want := range expected {
		if o.History[i] != want {
			t.Errorf("History[%d] = %q, want %q", i, o.History[i], want)
		}
	}
}

func TestOrder_CancelledCannotProceed(t *testing.T) {
	o := NewOrder("ORD-6")
	_ = o.Cancel()

	if err := o.Next(); err == nil {
		t.Error("expected error advancing cancelled order")
	}
	if err := o.Cancel(); err == nil {
		t.Error("expected error double-cancelling")
	}
}
