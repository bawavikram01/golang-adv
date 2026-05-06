// =============================================================================
// LESSON 11.1: ADVANCED TESTING — Table-Driven, Subtests, TestMain, Helpers
// =============================================================================
//
// Go's testing package is deceptively powerful. Most devs use 10% of it.
// This file covers production-grade testing patterns.
//
// RUN:
//   go test -v -race -count=1 ./11-advanced-testing/
//   go test -v -run=TestParse ./11-advanced-testing/
//   go test -cover -coverprofile=coverage.out ./11-advanced-testing/
//   go tool cover -html=coverage.out
// =============================================================================

package advancedtesting

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// CODE UNDER TEST
// =============================================================================

type User struct {
	Name  string
	Email string
	Age   int
}

func (u User) Validate() error {
	var errs []error
	if u.Name == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	if !strings.Contains(u.Email, "@") {
		errs = append(errs, fmt.Errorf("invalid email"))
	}
	if u.Age < 0 || u.Age > 150 {
		errs = append(errs, fmt.Errorf("age must be 0-150"))
	}
	return errors.Join(errs...)
}

func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	return time.ParseDuration(s)
}

// =============================================================================
// PATTERN 1: Table-Driven Tests — The Go standard
// =============================================================================
// Every test case is a row in a table. Add cases without modifying test logic.

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
		errMsg  string // substring to check
	}{
		{
			name:    "valid user",
			user:    User{Name: "Vikram", Email: "v@test.com", Age: 25},
			wantErr: false,
		},
		{
			name:    "empty name",
			user:    User{Name: "", Email: "v@test.com", Age: 25},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "invalid email",
			user:    User{Name: "Vikram", Email: "invalid", Age: 25},
			wantErr: true,
			errMsg:  "invalid email",
		},
		{
			name:    "negative age",
			user:    User{Name: "Vikram", Email: "v@test.com", Age: -1},
			wantErr: true,
			errMsg:  "age must be 0-150",
		},
		{
			name:    "multiple errors",
			user:    User{Name: "", Email: "bad", Age: 200},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "zero value user",
			user:    User{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		// t.Run creates a subtest — can be run individually:
		// go test -run=TestUserValidate/empty_name
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want containing %q", err, tt.errMsg)
			}
		})
	}
}

// =============================================================================
// PATTERN 2: t.Parallel() — Run tests concurrently
// =============================================================================
// Tests marked Parallel run concurrently with other parallel tests.
// CRITICAL: Capture loop variable (or use Go 1.22+ which fixes this).

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"1s", time.Second, false},
		{"500ms", 500 * time.Millisecond, false},
		{"2h30m", 2*time.Hour + 30*time.Minute, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel() // run each subtest concurrently

			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// PATTERN 3: Test Helpers — Reusable setup/assertion functions
// =============================================================================
// t.Helper() marks the function as a helper so errors report the CALLER's line.

func assertNoError(t *testing.T, err error) {
	t.Helper() // without this, errors point HERE, not the caller
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertError(t *testing.T, err error, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("got error %v, want %v", err, target)
	}
}

func TestWithHelpers(t *testing.T) {
	d, err := ParseDuration("1s")
	assertNoError(t, err)
	assertEqual(t, d, time.Second)
}

// =============================================================================
// PATTERN 4: TestMain — Global setup/teardown
// =============================================================================
// TestMain runs ONCE before all tests in the package. Use for:
//   - Database setup/teardown
//   - Starting test servers
//   - Setting environment variables
//   - Goroutine leak detection

func TestMain(m *testing.M) {
	// SETUP
	fmt.Println("=== GLOBAL SETUP ===")

	// Run tests
	code := m.Run()

	// TEARDOWN
	fmt.Println("=== GLOBAL TEARDOWN ===")

	os.Exit(code)
}

// =============================================================================
// PATTERN 5: t.Cleanup — Deferred cleanup per test
// =============================================================================
// Better than defer — runs even if test is in a subtest and panics.

func setupTempDB(t *testing.T) string {
	t.Helper()
	dbPath := t.TempDir() + "/test.db" // TempDir auto-cleans up!

	// Simulate DB setup
	t.Cleanup(func() {
		// This runs when the test and ALL its subtests finish
		fmt.Printf("  Cleaning up DB: %s\n", dbPath)
	})

	return dbPath
}

func TestWithCleanup(t *testing.T) {
	db := setupTempDB(t)
	t.Logf("Using DB at: %s", db) // only shown with -v flag
}

// =============================================================================
// PATTERN 6: Testing with context and timeouts
// =============================================================================

func slowOperation(ctx context.Context) (string, error) {
	select {
	case <-time.After(100 * time.Millisecond):
		return "done", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func TestSlowOperation(t *testing.T) {
	// Set a test deadline — if the test takes too long, it fails
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := slowOperation(ctx)
	assertNoError(t, err)
	assertEqual(t, result, "done")
}

func TestSlowOperationTimeout(t *testing.T) {
	// Test that the function respects cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := slowOperation(ctx)
	if err == nil {
		t.Error("expected timeout error")
	}
}

// =============================================================================
// PATTERN 7: Golden Files — Snapshot testing
// =============================================================================
// Compare output against a saved "golden" file.
// Update golden files: go test -update-golden

// func TestRenderTemplate(t *testing.T) {
//     output := renderTemplate(data)
//
//     golden := filepath.Join("testdata", t.Name()+".golden")
//
//     if *updateGolden {  // var updateGolden = flag.Bool("update-golden", false, "update golden files")
//         os.WriteFile(golden, []byte(output), 0644)
//         return
//     }
//
//     expected, err := os.ReadFile(golden)
//     assertNoError(t, err)
//     assertEqual(t, output, string(expected))
// }

// =============================================================================
// PATTERN 8: Testing unexported functions
// =============================================================================
// Test files in the SAME package can access unexported symbols.
// Test files in package "foo_test" (external test package) cannot.
//
// Both patterns are valid:
//   foo/foo.go       → package foo
//   foo/foo_test.go  → package foo       (whitebox: accesses unexported)
//   foo/foo_ext_test.go → package foo_test (blackbox: tests public API only)
//
// Use export_test.go to selectively expose unexported things for external tests:
//   foo/export_test.go → package foo
//     var ExportedForTest = unexportedFunc

// =============================================================================
// PATTERN 9: Race detector
// =============================================================================
// go test -race detects data races at runtime.
// ALWAYS run with -race in CI. It has ~2-10x overhead.

type Counter struct {
	n int
}

func (c *Counter) Inc() { c.n++ }
func (c *Counter) Get() int { return c.n }

func TestRace(t *testing.T) {
	// This test PASSES normally but FAILS with -race (without synchronization)
	// Uncomment to see race detection:
	// c := &Counter{}
	// var wg sync.WaitGroup
	// for i := 0; i < 100; i++ {
	//     wg.Add(1)
	//     go func() {
	//         defer wg.Done()
	//         c.Inc() // DATA RACE!
	//     }()
	// }
	// wg.Wait()
}

// =============================================================================
// PATTERN 10: Custom test flags
// =============================================================================
// Register custom flags for your tests (e.g., -integration, -slow)

// var integration = flag.Bool("integration", false, "run integration tests")
//
// func TestDatabaseQuery(t *testing.T) {
//     if !*integration {
//         t.Skip("skipping integration test; use -integration flag")
//     }
//     // ... actual DB test
// }
//
// Also use build tags:
// //go:build integration
//
// Run: go test -tags=integration ./...
