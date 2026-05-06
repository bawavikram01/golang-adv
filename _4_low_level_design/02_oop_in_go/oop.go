// Package oop demonstrates how Go achieves OOP concepts without classes.
//
// Go is NOT a traditional OOP language — no classes, no inheritance, no constructors.
// Instead it uses:
//   - Structs           → data (like classes without methods)
//   - Methods           → behavior attached to types
//   - Interfaces        → polymorphism (implicit satisfaction)
//   - Embedding         → composition over inheritance
//   - Constructor funcs → NewXxx() pattern
package oop

import "fmt"

// ──────────────────────────────────────────────
// ENCAPSULATION — exported vs unexported
// ──────────────────────────────────────────────

// BankAccount encapsulates balance. External packages cannot access `balance` directly.
type BankAccount struct {
	Owner   string  // exported — accessible outside package
	balance float64 // unexported — only this package can touch it
}

func NewBankAccount(owner string, initial float64) *BankAccount {
	return &BankAccount{Owner: owner, balance: initial}
}

func (a *BankAccount) Deposit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("deposit must be positive")
	}
	a.balance += amount
	return nil
}

func (a *BankAccount) Withdraw(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("withdrawal must be positive")
	}
	if amount > a.balance {
		return fmt.Errorf("insufficient balance")
	}
	a.balance -= amount
	return nil
}

func (a *BankAccount) Balance() float64 { return a.balance }

// ──────────────────────────────────────────────
// POLYMORPHISM — interfaces
// ──────────────────────────────────────────────

// Drawable is satisfied by any type with a Draw() method.
type Drawable interface {
	Draw() string
}

type CircleShape struct{ Radius float64 }
type RectShape struct{ W, H float64 }
type TextBox struct{ Text string }

func (c CircleShape) Draw() string { return fmt.Sprintf("Drawing circle r=%.1f", c.Radius) }
func (r RectShape) Draw() string   { return fmt.Sprintf("Drawing rect %vx%v", r.W, r.H) }
func (t TextBox) Draw() string     { return fmt.Sprintf("Drawing text: %s", t.Text) }

// RenderAll works with ANY Drawable — polymorphism via interface.
func RenderAll(items []Drawable) []string {
	results := make([]string, len(items))
	for i, d := range items {
		results[i] = d.Draw()
	}
	return results
}

// ──────────────────────────────────────────────
// COMPOSITION (embedding) — Go's alternative to inheritance
// ──────────────────────────────────────────────

// Animal is a base type with common behavior.
type Animal struct {
	Name    string
	Species string
}

func (a Animal) Speak() string {
	return fmt.Sprintf("%s (%s) makes a sound", a.Name, a.Species)
}

func (a Animal) String() string {
	return fmt.Sprintf("%s the %s", a.Name, a.Species)
}

// Dog embeds Animal — gets all its methods, can override.
type Dog struct {
	Animal
	Breed string
}

func (d Dog) Speak() string {
	return fmt.Sprintf("%s barks! Woof!", d.Name)
}

func (d Dog) Fetch() string {
	return fmt.Sprintf("%s fetches the ball", d.Name)
}

// Cat embeds Animal with its own override.
type Cat struct {
	Animal
	Indoor bool
}

func (c Cat) Speak() string {
	return fmt.Sprintf("%s purrs... meow", c.Name)
}

// ──────────────────────────────────────────────
// INTERFACE + EMBEDDING together
// ──────────────────────────────────────────────

// Speaker is a shared interface.
type Speaker interface {
	Speak() string
}

// MakeThemSpeak demonstrates polymorphism with embedded types.
func MakeThemSpeak(speakers []Speaker) []string {
	results := make([]string, len(speakers))
	for i, s := range speakers {
		results[i] = s.Speak()
	}
	return results
}
