package dependencyinversion

import "testing"

func TestNotificationService_WithEmail(t *testing.T) {
	email := &EmailNotifier{}
	store := NewInMemoryStore()
	svc := NewNotificationService(email, store)

	err := svc.Notify("vikram@x.com", "Hello!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(email.SentMessages) != 1 {
		t.Errorf("expected 1 sent message, got %d", len(email.SentMessages))
	}
	if history := svc.History("vikram@x.com"); len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}
}

func TestNotificationService_WithSMS(t *testing.T) {
	sms := &SMSNotifier{}
	store := NewInMemoryStore()
	// Swap implementation — ZERO changes to NotificationService
	svc := NewNotificationService(sms, store)

	err := svc.Notify("+91-9876543210", "OTP: 1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sms.SentMessages[0] != "SMS to +91-9876543210: OTP: 1234" {
		t.Errorf("unexpected message: %s", sms.SentMessages[0])
	}
}

func TestNotificationService_EmptyRecipient(t *testing.T) {
	svc := NewNotificationService(&EmailNotifier{}, NewInMemoryStore())
	err := svc.Notify("", "Hello")
	if err == nil {
		t.Error("expected error for empty recipient")
	}
}

// FakeNotifier — demonstrates testability via DIP
type FakeNotifier struct {
	Calls []string
}

func (f *FakeNotifier) Send(to, msg string) error {
	f.Calls = append(f.Calls, to+":"+msg)
	return nil
}

func TestNotificationService_WithFake(t *testing.T) {
	fake := &FakeNotifier{}
	store := NewInMemoryStore()
	svc := NewNotificationService(fake, store)

	_ = svc.Notify("test", "msg1")
	_ = svc.Notify("test", "msg2")

	if len(fake.Calls) != 2 {
		t.Errorf("expected 2 calls, got %d", len(fake.Calls))
	}
	if history := svc.History("test"); len(history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(history))
	}
}
