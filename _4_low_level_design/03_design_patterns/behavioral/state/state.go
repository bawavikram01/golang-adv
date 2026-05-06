// Package state demonstrates the State pattern.
//
// INTENT: Allow an object to alter its behavior when its internal state changes.
// The object will appear to change its class.
//
// WHEN TO USE:
//   - Object behavior depends on its state (e.g., vending machine, order lifecycle)
//   - Large switch/if-else blocks based on state
//   - State transitions have rules/guards
//
// Go idiom: State = interface. Each state = struct with transition logic.
// The context delegates behavior to the current state object.
package state

import "fmt"

// ──────────────────────────────────────────────
// State interface
// ──────────────────────────────────────────────

type OrderState interface {
	Name() string
	Next(o *Order) error
	Cancel(o *Order) error
}

// ──────────────────────────────────────────────
// Context — the Order
// ──────────────────────────────────────────────

type Order struct {
	ID      string
	state   OrderState
	History []string
}

func NewOrder(id string) *Order {
	o := &Order{ID: id}
	o.setState(&PendingState{})
	return o
}

func (o *Order) setState(s OrderState) {
	o.state = s
	o.History = append(o.History, s.Name())
}

func (o *Order) State() string { return o.state.Name() }
func (o *Order) Next() error   { return o.state.Next(o) }
func (o *Order) Cancel() error { return o.state.Cancel(o) }

// ──────────────────────────────────────────────
// Concrete states
// ──────────────────────────────────────────────

// PendingState — order just created
type PendingState struct{}

func (s *PendingState) Name() string { return "Pending" }
func (s *PendingState) Next(o *Order) error {
	o.setState(&ConfirmedState{})
	return nil
}
func (s *PendingState) Cancel(o *Order) error {
	o.setState(&CancelledState{})
	return nil
}

// ConfirmedState — payment received
type ConfirmedState struct{}

func (s *ConfirmedState) Name() string { return "Confirmed" }
func (s *ConfirmedState) Next(o *Order) error {
	o.setState(&ShippedState{})
	return nil
}
func (s *ConfirmedState) Cancel(o *Order) error {
	o.setState(&CancelledState{})
	return nil
}

// ShippedState — on the way
type ShippedState struct{}

func (s *ShippedState) Name() string { return "Shipped" }
func (s *ShippedState) Next(o *Order) error {
	o.setState(&DeliveredState{})
	return nil
}
func (s *ShippedState) Cancel(o *Order) error {
	return fmt.Errorf("cannot cancel: order already shipped")
}

// DeliveredState — terminal state
type DeliveredState struct{}

func (s *DeliveredState) Name() string { return "Delivered" }
func (s *DeliveredState) Next(o *Order) error {
	return fmt.Errorf("order already delivered — no next state")
}
func (s *DeliveredState) Cancel(o *Order) error {
	return fmt.Errorf("cannot cancel: order already delivered")
}

// CancelledState — terminal state
type CancelledState struct{}

func (s *CancelledState) Name() string { return "Cancelled" }
func (s *CancelledState) Next(o *Order) error {
	return fmt.Errorf("order is cancelled — cannot proceed")
}
func (s *CancelledState) Cancel(o *Order) error {
	return fmt.Errorf("order already cancelled")
}
