// Package observer demonstrates the Observer pattern.
//
// INTENT: Define a one-to-many dependency so that when one object (subject)
// changes state, all dependents (observers) are notified automatically.
//
// WHEN TO USE:
//   - Event systems, pub/sub
//   - UI updates when data model changes
//   - Notifications when state changes
//
// Go idiom: Subject holds a slice of observer interfaces.
// Observers register/unregister themselves.
package observer

import "fmt"

// ──────────────────────────────────────────────
// Observer interface
// ──────────────────────────────────────────────

type Event struct {
	Type    string
	Payload string
}

type Observer interface {
	OnEvent(event Event)
	ID() string
}

// ──────────────────────────────────────────────
// Subject (Event Bus)
// ──────────────────────────────────────────────

type EventBus struct {
	observers map[string][]Observer // eventType -> observers
}

func NewEventBus() *EventBus {
	return &EventBus{observers: make(map[string][]Observer)}
}

func (eb *EventBus) Subscribe(eventType string, o Observer) {
	eb.observers[eventType] = append(eb.observers[eventType], o)
}

func (eb *EventBus) Unsubscribe(eventType string, o Observer) {
	subs := eb.observers[eventType]
	for i, sub := range subs {
		if sub.ID() == o.ID() {
			eb.observers[eventType] = append(subs[:i], subs[i+1:]...)
			return
		}
	}
}

func (eb *EventBus) Publish(event Event) {
	for _, o := range eb.observers[event.Type] {
		o.OnEvent(event)
	}
}

func (eb *EventBus) SubscriberCount(eventType string) int {
	return len(eb.observers[eventType])
}

// ──────────────────────────────────────────────
// Concrete observers
// ──────────────────────────────────────────────

type EmailAlert struct {
	Name     string
	Messages []string
}

func (e *EmailAlert) OnEvent(event Event) {
	e.Messages = append(e.Messages, fmt.Sprintf("EMAIL[%s]: %s", event.Type, event.Payload))
}

func (e *EmailAlert) ID() string { return "email-" + e.Name }

type SlackAlert struct {
	Channel  string
	Messages []string
}

func (s *SlackAlert) OnEvent(event Event) {
	s.Messages = append(s.Messages, fmt.Sprintf("SLACK[#%s]: %s", s.Channel, event.Payload))
}

func (s *SlackAlert) ID() string { return "slack-" + s.Channel }

type LogAlert struct {
	Logs []string
}

func (l *LogAlert) OnEvent(event Event) {
	l.Logs = append(l.Logs, fmt.Sprintf("LOG: [%s] %s", event.Type, event.Payload))
}

func (l *LogAlert) ID() string { return "logger" }
