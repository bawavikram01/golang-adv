// Package dependencyinversion demonstrates the Dependency Inversion Principle.
//
// PRINCIPLE:
//  1. High-level modules should not depend on low-level modules.
//     Both should depend on abstractions.
//  2. Abstractions should not depend on details.
//     Details should depend on abstractions.
//
// Go idiom: "Accept interfaces, return structs."
//   - Functions/structs accept interface parameters (the abstraction)
//   - Concrete implementations are injected from the outside
//   - Easy to test with mocks/fakes
//
// EXAMPLE: A NotificationService that can send via Email, SMS, or Push.
// The service depends on the Notifier interface, not on any concrete sender.
package dependencyinversion

import (
	"fmt"
	"strings"
)

// ──────────────────────────────────────────────
// Abstraction — the high-level module depends on THIS
// ──────────────────────────────────────────────

// Notifier is the abstraction. High-level code depends on this interface,
// not on any specific implementation.
type Notifier interface {
	Send(to, message string) error
}

// MessageStore abstracts persistence. High-level code doesn't know if it's
// a database, file, or in-memory store.
type MessageStore interface {
	Save(to, message string) error
	GetAll(to string) []string
}

// ──────────────────────────────────────────────
// Low-level implementations (details)
// ──────────────────────────────────────────────

// EmailNotifier — concrete implementation of Notifier
type EmailNotifier struct {
	SentMessages []string
}

func (e *EmailNotifier) Send(to, message string) error {
	e.SentMessages = append(e.SentMessages, fmt.Sprintf("EMAIL to %s: %s", to, message))
	return nil
}

// SMSNotifier — another concrete implementation
type SMSNotifier struct {
	SentMessages []string
}

func (s *SMSNotifier) Send(to, message string) error {
	s.SentMessages = append(s.SentMessages, fmt.Sprintf("SMS to %s: %s", to, message))
	return nil
}

// PushNotifier — yet another implementation
type PushNotifier struct {
	SentMessages []string
}

func (p *PushNotifier) Send(to, message string) error {
	p.SentMessages = append(p.SentMessages, fmt.Sprintf("PUSH to %s: %s", to, message))
	return nil
}

// InMemoryStore — concrete MessageStore
type InMemoryStore struct {
	data map[string][]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string][]string)}
}

func (s *InMemoryStore) Save(to, message string) error {
	s.data[to] = append(s.data[to], message)
	return nil
}

func (s *InMemoryStore) GetAll(to string) []string {
	return s.data[to]
}

// ──────────────────────────────────────────────
// High-level module — depends ONLY on abstractions
// ──────────────────────────────────────────────

// NotificationService is the high-level policy.
// It depends on Notifier and MessageStore interfaces — NOT on concrete types.
// This means:
//   - You can swap Email for SMS without changing this code
//   - You can swap InMemory for Postgres without changing this code
//   - You can test with fakes/mocks trivially
type NotificationService struct {
	notifier Notifier
	store    MessageStore
}

// NewNotificationService — dependency injection via constructor.
func NewNotificationService(n Notifier, s MessageStore) *NotificationService {
	return &NotificationService{notifier: n, store: s}
}

func (ns *NotificationService) Notify(to, message string) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("recipient is required")
	}
	if err := ns.notifier.Send(to, message); err != nil {
		return fmt.Errorf("send failed: %w", err)
	}
	if err := ns.store.Save(to, message); err != nil {
		return fmt.Errorf("store failed: %w", err)
	}
	return nil
}

func (ns *NotificationService) History(to string) []string {
	return ns.store.GetAll(to)
}
