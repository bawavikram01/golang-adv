// Package factory demonstrates the Factory Method pattern.
//
// INTENT: Define an interface for creating objects, but let the factory
// decide which concrete type to instantiate. The client code works with
// the interface — never with concrete types directly.
//
// WHEN TO USE:
//   - You don't know ahead of time which concrete type you need
//   - You want to centralize creation logic
//   - You want to add new types without changing client code
package factory

import "fmt"

// ──────────────────────────────────────────────
// Product interface
// ──────────────────────────────────────────────

type Notification interface {
	Send(to, message string) string
	Type() string
}

// ──────────────────────────────────────────────
// Concrete products
// ──────────────────────────────────────────────

type EmailNotification struct{}

func (e EmailNotification) Send(to, message string) string {
	return fmt.Sprintf("[EMAIL] To: %s | %s", to, message)
}
func (e EmailNotification) Type() string { return "email" }

type SMSNotification struct{}

func (s SMSNotification) Send(to, message string) string {
	return fmt.Sprintf("[SMS] To: %s | %s", to, message)
}
func (s SMSNotification) Type() string { return "sms" }

type PushNotification struct{}

func (p PushNotification) Send(to, message string) string {
	return fmt.Sprintf("[PUSH] To: %s | %s", to, message)
}
func (p PushNotification) Type() string { return "push" }

// ──────────────────────────────────────────────
// Factory function
// ──────────────────────────────────────────────

func NewNotification(channel string) (Notification, error) {
	switch channel {
	case "email":
		return EmailNotification{}, nil
	case "sms":
		return SMSNotification{}, nil
	case "push":
		return PushNotification{}, nil
	default:
		return nil, fmt.Errorf("unknown channel: %s", channel)
	}
}

// ──────────────────────────────────────────────
// Client code — works only with the Notification interface
// ──────────────────────────────────────────────

func SendAlert(channel, to, message string) (string, error) {
	n, err := NewNotification(channel)
	if err != nil {
		return "", err
	}
	return n.Send(to, message), nil
}
