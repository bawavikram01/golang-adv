// =============================================================================
// LESSON 11.2: FUZZING, MOCKING & INTEGRATION TESTING
// =============================================================================
//
// FUZZING (Go 1.18+): The compiler generates random inputs to find crashes.
// MOCKING: Interfaces + manual mocks (no framework needed in Go).
// INTEGRATION: Testing real dependencies with testcontainers or build tags.
//
// RUN FUZZ:
//   go test -fuzz=FuzzParseEmail -fuzztime=10s ./11-advanced-testing/
// =============================================================================

package advancedtesting

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// =============================================================================
// CODE UNDER TEST
// =============================================================================

type Email struct {
	Local  string
	Domain string
}

func ParseEmail(s string) (Email, error) {
	if s == "" {
		return Email{}, fmt.Errorf("empty email")
	}
	if !utf8.ValidString(s) {
		return Email{}, fmt.Errorf("invalid UTF-8")
	}

	at := strings.LastIndex(s, "@")
	if at < 1 {
		return Email{}, fmt.Errorf("missing or invalid @")
	}

	local := s[:at]
	domain := s[at+1:]

	if domain == "" {
		return Email{}, fmt.Errorf("empty domain")
	}
	if !strings.Contains(domain, ".") {
		return Email{}, fmt.Errorf("domain must have a dot")
	}
	if len(local) > 64 {
		return Email{}, fmt.Errorf("local part too long")
	}
	if len(domain) > 255 {
		return Email{}, fmt.Errorf("domain too long")
	}

	return Email{Local: local, Domain: domain}, nil
}

// =============================================================================
// FUZZ TEST — Finds edge cases you'd never think of
// =============================================================================
//
// The fuzzer starts with your "seed corpus" and mutates inputs.
// If it finds a crash (panic, hang, or unexpected error), it saves the input
// to testdata/fuzz/FuzzParseEmail/ for regression testing.
//
// RUN: go test -fuzz=FuzzParseEmail -fuzztime=30s

func FuzzParseEmail(f *testing.F) {
	// Seed corpus — provide diverse starting points
	seeds := []string{
		"user@example.com",
		"a@b.c",
		"user+tag@domain.co.uk",
		"",
		"@",
		"@@",
		"user@",
		"@domain.com",
		"user@domain",
		"a@b.c.d.e.f",
		"very.long.local.part@example.com",
		"user@sub.domain.example.com",
		"名前@例え.jp", // unicode
	}

	for _, s := range seeds {
		f.Add(s) // add to seed corpus
	}

	// The fuzz function receives random mutations of the seeds
	f.Fuzz(func(t *testing.T, input string) {
		email, err := ParseEmail(input)
		if err != nil {
			return // errors are fine — we're looking for panics
		}

		// INVARIANT CHECKS — if parsing succeeds, these MUST be true:
		if email.Local == "" {
			t.Error("successful parse returned empty local part")
		}
		if email.Domain == "" {
			t.Error("successful parse returned empty domain")
		}
		if !strings.Contains(email.Domain, ".") {
			t.Errorf("domain %q has no dot", email.Domain)
		}

		// Round-trip check: reconstruct and re-parse
		reconstructed := email.Local + "@" + email.Domain
		email2, err := ParseEmail(reconstructed)
		if err != nil {
			t.Errorf("round-trip failed: ParseEmail(%q) = %v", reconstructed, err)
		}
		if email2.Local != email.Local || email2.Domain != email.Domain {
			t.Errorf("round-trip mismatch: got %+v, want %+v", email2, email)
		}
	})
}

// =============================================================================
// MOCKING — Interface-based, no framework needed
// =============================================================================
//
// Go's approach: define interfaces at the CONSUMER, not the producer.
// Create manual mock implementations for testing.

// Interface defined where it's USED (not where it's implemented)
type UserStore interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	Save(ctx context.Context, user *User) error
}

type NotificationService interface {
	SendWelcomeEmail(ctx context.Context, email string) error
}

// Production service that depends on these interfaces
type UserRegistrar struct {
	store   UserStore
	notify  NotificationService
}

func NewUserRegistrar(store UserStore, notify NotificationService) *UserRegistrar {
	return &UserRegistrar{store: store, notify: notify}
}

func (r *UserRegistrar) Register(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	if err := r.store.Save(ctx, user); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	// Non-critical — don't fail registration if email fails
	if err := r.notify.SendWelcomeEmail(ctx, user.Email); err != nil {
		// Log but don't return error
		_ = err
	}
	return nil
}

// --- Mock implementations ---

type MockUserStore struct {
	SaveFn    func(ctx context.Context, user *User) error
	GetByIDFn func(ctx context.Context, id int64) (*User, error)

	// Track calls for assertions
	SaveCalls []User
}

func (m *MockUserStore) Save(ctx context.Context, user *User) error {
	m.SaveCalls = append(m.SaveCalls, *user)
	if m.SaveFn != nil {
		return m.SaveFn(ctx, user)
	}
	return nil
}

func (m *MockUserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("not found")
}

type MockNotifier struct {
	SendFn    func(ctx context.Context, email string) error
	SendCalls []string
}

func (m *MockNotifier) SendWelcomeEmail(ctx context.Context, email string) error {
	m.SendCalls = append(m.SendCalls, email)
	if m.SendFn != nil {
		return m.SendFn(ctx, email)
	}
	return nil
}

// --- Tests with mocks ---

func TestUserRegistrar_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		store := &MockUserStore{}
		notifier := &MockNotifier{}
		registrar := NewUserRegistrar(store, notifier)

		user := &User{Name: "Vikram", Email: "v@test.com", Age: 25}
		err := registrar.Register(context.Background(), user)
		assertNoError(t, err)

		// Verify interactions
		assertEqual(t, len(store.SaveCalls), 1)
		assertEqual(t, store.SaveCalls[0].Name, "Vikram")
		assertEqual(t, len(notifier.SendCalls), 1)
		assertEqual(t, notifier.SendCalls[0], "v@test.com")
	})

	t.Run("validation failure doesn't call store", func(t *testing.T) {
		store := &MockUserStore{}
		notifier := &MockNotifier{}
		registrar := NewUserRegistrar(store, notifier)

		user := &User{Name: "", Email: "invalid", Age: -1}
		err := registrar.Register(context.Background(), user)

		if err == nil {
			t.Fatal("expected validation error")
		}
		assertEqual(t, len(store.SaveCalls), 0) // store should NOT be called
	})

	t.Run("store failure returns error", func(t *testing.T) {
		store := &MockUserStore{
			SaveFn: func(ctx context.Context, user *User) error {
				return fmt.Errorf("connection refused")
			},
		}
		notifier := &MockNotifier{}
		registrar := NewUserRegistrar(store, notifier)

		user := &User{Name: "Vikram", Email: "v@test.com", Age: 25}
		err := registrar.Register(context.Background(), user)

		if err == nil {
			t.Fatal("expected store error")
		}
		if !strings.Contains(err.Error(), "connection refused") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("notification failure doesn't fail registration", func(t *testing.T) {
		store := &MockUserStore{}
		notifier := &MockNotifier{
			SendFn: func(ctx context.Context, email string) error {
				return fmt.Errorf("SMTP timeout")
			},
		}
		registrar := NewUserRegistrar(store, notifier)

		user := &User{Name: "Vikram", Email: "v@test.com", Age: 25}
		err := registrar.Register(context.Background(), user)

		assertNoError(t, err) // registration should succeed despite email failure
	})
}

// =============================================================================
// BENCHMARK WITH DIFFERENT STRATEGIES
// =============================================================================

func BenchmarkParseEmail(b *testing.B) {
	inputs := []string{
		"simple@example.com",
		"user+tag@sub.domain.example.com",
		"a@b.co",
	}

	for _, input := range inputs {
		b.Run(input, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ParseEmail(input)
			}
		})
	}
}

// =============================================================================
// TESTING TIME-DEPENDENT CODE
// =============================================================================
// Don't use time.Now() directly — inject a clock interface.

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type MockClock struct {
	CurrentTime time.Time
}

func (m MockClock) Now() time.Time { return m.CurrentTime }

type TokenService struct {
	clock Clock
}

func (ts *TokenService) IsExpired(expiresAt time.Time) bool {
	return ts.clock.Now().After(expiresAt)
}

func TestTokenExpiry(t *testing.T) {
	fixedTime := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	mock := MockClock{CurrentTime: fixedTime}
	svc := &TokenService{clock: mock}

	t.Run("not expired", func(t *testing.T) {
		future := fixedTime.Add(1 * time.Hour)
		assertEqual(t, svc.IsExpired(future), false)
	})

	t.Run("expired", func(t *testing.T) {
		past := fixedTime.Add(-1 * time.Hour)
		assertEqual(t, svc.IsExpired(past), true)
	})
}
