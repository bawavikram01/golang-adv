package factory

import "testing"

func TestNewNotification(t *testing.T) {
	tests := []struct {
		channel  string
		wantType string
		wantErr  bool
	}{
		{"email", "email", false},
		{"sms", "sms", false},
		{"push", "push", false},
		{"pigeon", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.channel, func(t *testing.T) {
			n, err := NewNotification(tt.channel)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && n.Type() != tt.wantType {
				t.Errorf("Type() = %q, want %q", n.Type(), tt.wantType)
			}
		})
	}
}

func TestSendAlert(t *testing.T) {
	got, err := SendAlert("email", "vikram@x.com", "Server down!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "[EMAIL] To: vikram@x.com | Server down!"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSendAlert_UnknownChannel(t *testing.T) {
	_, err := SendAlert("telegram", "user", "msg")
	if err == nil {
		t.Error("expected error for unknown channel")
	}
}
