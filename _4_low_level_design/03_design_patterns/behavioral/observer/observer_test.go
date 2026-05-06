package observer

import "testing"

func TestEventBus_PublishNotifiesSubscribers(t *testing.T) {
	bus := NewEventBus()
	email := &EmailAlert{Name: "ops"}
	slack := &SlackAlert{Channel: "alerts"}

	bus.Subscribe("deploy", email)
	bus.Subscribe("deploy", slack)

	bus.Publish(Event{Type: "deploy", Payload: "v2.0 released"})

	if len(email.Messages) != 1 {
		t.Fatalf("email got %d messages", len(email.Messages))
	}
	if email.Messages[0] != "EMAIL[deploy]: v2.0 released" {
		t.Errorf("email msg = %q", email.Messages[0])
	}
	if slack.Messages[0] != "SLACK[#alerts]: v2.0 released" {
		t.Errorf("slack msg = %q", slack.Messages[0])
	}
}

func TestEventBus_OnlyNotifiesRelevantSubscribers(t *testing.T) {
	bus := NewEventBus()
	email := &EmailAlert{Name: "ops"}
	logger := &LogAlert{}

	bus.Subscribe("error", email)
	bus.Subscribe("info", logger)

	bus.Publish(Event{Type: "error", Payload: "disk full"})

	if len(email.Messages) != 1 {
		t.Errorf("email should get 1 message, got %d", len(email.Messages))
	}
	if len(logger.Logs) != 0 {
		t.Errorf("logger should get 0 messages, got %d", len(logger.Logs))
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	email := &EmailAlert{Name: "ops"}

	bus.Subscribe("deploy", email)
	bus.Unsubscribe("deploy", email)

	bus.Publish(Event{Type: "deploy", Payload: "should not arrive"})

	if len(email.Messages) != 0 {
		t.Errorf("unsubscribed observer got %d messages", len(email.Messages))
	}
}

func TestEventBus_MultipleEvents(t *testing.T) {
	bus := NewEventBus()
	logger := &LogAlert{}

	bus.Subscribe("error", logger)
	bus.Subscribe("warn", logger)

	bus.Publish(Event{Type: "error", Payload: "oops"})
	bus.Publish(Event{Type: "warn", Payload: "careful"})
	bus.Publish(Event{Type: "info", Payload: "ignored"})

	if len(logger.Logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logger.Logs))
	}
}

func TestEventBus_SubscriberCount(t *testing.T) {
	bus := NewEventBus()
	bus.Subscribe("x", &EmailAlert{Name: "a"})
	bus.Subscribe("x", &EmailAlert{Name: "b"})

	if bus.SubscriberCount("x") != 2 {
		t.Errorf("count = %d, want 2", bus.SubscriberCount("x"))
	}
	if bus.SubscriberCount("y") != 0 {
		t.Errorf("count for unknown = %d", bus.SubscriberCount("y"))
	}
}
