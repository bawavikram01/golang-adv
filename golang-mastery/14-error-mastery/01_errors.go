// =============================================================================
// LESSON 14: ERROR MASTERY — The Art of Go Error Handling
// =============================================================================
//
// Go's error handling is controversial but powerful when done right.
// This lesson covers every technique from basic to god-level.
//
// EVOLUTION:
//   Go 1.0:  errors.New, fmt.Errorf
//   Go 1.13: errors.Is, errors.As, fmt.Errorf %w (wrapping)
//   Go 1.20: errors.Join (multi-error)
//   Go 1.21: slog structured error logging
// =============================================================================

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

// =============================================================================
// LEVEL 1: Sentinel Errors — Package-level error constants
// =============================================================================
// Use for errors that callers need to check with errors.Is().
// Name convention: Err<Noun> (e.g., ErrNotFound, ErrTimeout)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrRateLimit     = errors.New("rate limit exceeded")
)

// =============================================================================
// LEVEL 2: Custom Error Types — When you need structured error data
// =============================================================================
// Use when callers need to extract information (HTTP status, field name, etc.)
// Callers check with errors.As().

// Domain error with operation context
type DomainError struct {
	Op      string // operation: "users.Create", "orders.Process"
	Kind    error  // sentinel error category
	Entity  string // what entity: "user", "order"
	ID      string // which one
	Err     error  // underlying cause (may be nil)
}

func (e *DomainError) Error() string {
	var b strings.Builder
	b.WriteString(e.Op)
	if e.Entity != "" {
		b.WriteString(": " + e.Entity)
		if e.ID != "" {
			b.WriteString("(" + e.ID + ")")
		}
	}
	if e.Err != nil {
		b.WriteString(": " + e.Err.Error())
	}
	return b.String()
}

// Unwrap returns the Kind (sentinel) for errors.Is() checks
func (e *DomainError) Unwrap() error {
	return e.Kind
}

// Builder pattern for fluent error construction
func E(op string, kind error) *DomainError {
	return &DomainError{Op: op, Kind: kind}
}

func (e *DomainError) WithEntity(entity, id string) *DomainError {
	e.Entity = entity
	e.ID = id
	return e
}

func (e *DomainError) WithCause(err error) *DomainError {
	e.Err = err
	return e
}

// =============================================================================
// LEVEL 3: Validation Error — Multiple field errors
// =============================================================================

type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	msgs := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		msgs[i] = f.Error()
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}

// Unwrap makes it work with errors.Is(err, ErrInvalidInput)
func (e *ValidationError) Unwrap() error {
	return ErrInvalidInput
}

type ValidationBuilder struct {
	fields []FieldError
}

func NewValidation() *ValidationBuilder {
	return &ValidationBuilder{}
}

func (v *ValidationBuilder) Check(condition bool, field, message string) *ValidationBuilder {
	if !condition {
		v.fields = append(v.fields, FieldError{Field: field, Message: message})
	}
	return v
}

func (v *ValidationBuilder) Build() error {
	if len(v.fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: v.fields}
}

// =============================================================================
// LEVEL 4: errors.Join — Multiple independent errors (Go 1.20+)
// =============================================================================

func closeResources(resources []interface{ Close() error }) error {
	var errs []error
	for _, r := range resources {
		if err := r.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...) // nil if no errors
}

// errors.Join produces an error that unwraps to ALL constituent errors:
func demonstrateJoin() {
	fmt.Println("\n=== errors.Join (Go 1.20+) ===")

	err1 := fmt.Errorf("database flush failed: %w", ErrNotFound)
	err2 := fmt.Errorf("cache cleanup failed: %w", ErrRateLimit)

	combined := errors.Join(err1, err2)
	fmt.Printf("Combined: %v\n", combined)

	// errors.Is works through Join — checks ALL wrapped errors
	fmt.Printf("Is NotFound: %v\n", errors.Is(combined, ErrNotFound))   // true
	fmt.Printf("Is RateLimit: %v\n", errors.Is(combined, ErrRateLimit)) // true
}

// =============================================================================
// LEVEL 5: Error wrapping chain — Adding context as errors propagate
// =============================================================================
//
// PATTERN: Each layer adds its context with fmt.Errorf and %w
// The error chain preserves the full call stack of context.

func findUser(id string) (*User, error) {
	// Simulate DB error
	return nil, fmt.Errorf("connection refused")
}

func getUserProfile(id string) (*User, error) {
	user, err := findUser(id)
	if err != nil {
		// Wrap with context: what WE were trying to do + underlying error
		return nil, fmt.Errorf("getUserProfile(%s): %w", id, err)
	}
	return user, nil
}

func handleUserRequest(id string) error {
	_, err := getUserProfile(id)
	if err != nil {
		// Wrap again with our context
		return fmt.Errorf("handleUserRequest: %w", err)
	}
	return nil
}

// Result: "handleUserRequest: getUserProfile(123): connection refused"
// Each layer added its context without losing the original error.

type User struct {
	ID    string
	Name  string
	Email string
}

// =============================================================================
// LEVEL 6: errors.Is vs errors.As — When to use which
// =============================================================================

func demonstrateIsVsAs() {
	fmt.Println("\n=== errors.Is vs errors.As ===")

	// errors.Is — "Is this error (or anything in its chain) equal to X?"
	// Use for sentinel error checks.
	err := E("users.Get", ErrNotFound).WithEntity("user", "123")
	wrappedErr := fmt.Errorf("handler failed: %w", err)

	fmt.Printf("errors.Is(ErrNotFound): %v\n", errors.Is(wrappedErr, ErrNotFound))
	fmt.Printf("errors.Is(ErrForbidden): %v\n", errors.Is(wrappedErr, ErrForbidden))

	// errors.As — "Can I extract a specific error TYPE from the chain?"
	// Use when you need data from a structured error.
	var domainErr *DomainError
	if errors.As(wrappedErr, &domainErr) {
		fmt.Printf("errors.As found DomainError: op=%s entity=%s id=%s\n",
			domainErr.Op, domainErr.Entity, domainErr.ID)
	}

	// Works with standard library errors too:
	_, osErr := os.Open("/nonexistent/file")
	var pathErr *fs.PathError
	if errors.As(osErr, &pathErr) {
		fmt.Printf("PathError: op=%s path=%s\n", pathErr.Op, pathErr.Path)
	}
}

// =============================================================================
// LEVEL 7: Error handling in HTTP — Mapping domain errors to HTTP status
// =============================================================================

func httpStatusFromError(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.Is(err, ErrInvalidInput):
		return 400
	case errors.Is(err, ErrUnauthorized):
		return 401
	case errors.Is(err, ErrForbidden):
		return 403
	case errors.Is(err, ErrAlreadyExists):
		return 409
	case errors.Is(err, ErrRateLimit):
		return 429
	default:
		return 500
	}
}

// Keep user-facing messages safe — never expose internal errors
func userMessage(err error) string {
	switch {
	case errors.Is(err, ErrNotFound):
		return "The requested resource was not found"
	case errors.Is(err, ErrInvalidInput):
		return "Invalid request"
	case errors.Is(err, ErrUnauthorized):
		return "Authentication required"
	case errors.Is(err, ErrForbidden):
		return "You don't have permission"
	case errors.Is(err, ErrRateLimit):
		return "Too many requests, please try again later"
	default:
		return "An internal error occurred" // NEVER expose internal details
	}
}

// =============================================================================
// LEVEL 8: Stack traces — When you need them (and when you don't)
// =============================================================================
//
// Go errors DON'T include stack traces by default (unlike Java/Python).
// This is intentional — stack traces are expensive and usually redundant
// if you wrap errors with context at each layer.
//
// If you NEED stack traces (debugging), here's a lightweight approach:

type StackError struct {
	Err   error
	Stack string
}

func (e *StackError) Error() string { return e.Err.Error() }
func (e *StackError) Unwrap() error { return e.Err }

func WithStack(err error) error {
	if err == nil {
		return nil
	}
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return &StackError{
		Err:   err,
		Stack: string(buf[:n]),
	}
}

// Extract stack if present
func GetStack(err error) (string, bool) {
	var stackErr *StackError
	if errors.As(err, &stackErr) {
		return stackErr.Stack, true
	}
	return "", false
}

// =============================================================================
// LEVEL 9: Error handling patterns to AVOID
// =============================================================================

func antiPatterns() {
	fmt.Println("\n=== Anti-Patterns (DON'T DO THESE) ===")

	// 1. DON'T compare error strings
	// BAD: if err.Error() == "not found"  — fragile, breaks with wrapping
	// GOOD: if errors.Is(err, ErrNotFound)

	// 2. DON'T use panic for expected errors
	// BAD: panic("user not found")
	// GOOD: return nil, ErrNotFound

	// 3. DON'T ignore errors silently
	// BAD: result, _ := riskyOperation()
	// GOOD: result, err := riskyOperation()
	//       if err != nil { return err }

	// 4. DON'T wrap errors without adding context
	// BAD: return fmt.Errorf("error: %w", err)  — adds nothing useful
	// GOOD: return fmt.Errorf("fetching user %d: %w", id, err)

	// 5. DON'T use error types when sentinel errors suffice
	// BAD: type NotFoundError struct{} for simple presence check
	// GOOD: var ErrNotFound = errors.New("not found")

	// 6. DON'T create errors in hot paths — reuse sentinel errors
	// BAD (in hot loop): return fmt.Errorf("invalid value: %d", v)
	// GOOD: return ErrInvalidInput  (then log the details separately)

	fmt.Println("  1. Use errors.Is/As instead of string comparison")
	fmt.Println("  2. Use error returns not panic for expected failures")
	fmt.Println("  3. Never ignore errors (use errcheck linter)")
	fmt.Println("  4. Always add meaningful context when wrapping")
	fmt.Println("  5. Use sentinel errors for simple checks, types for data")
	fmt.Println("  6. Don't allocate errors in hot loops")
}

// =============================================================================
// LEVEL 10: Error logging best practices
// =============================================================================

func demonstrateErrorLogging() {
	fmt.Println("\n=== Error Logging Best Practices ===")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	err := handleUserRequest("123")

	// Log at the TOP of the call stack — not at every level
	// (avoids duplicate log entries for the same error)
	if err != nil {
		// Log with structured context
		logger.Error("request failed",
			"error", err,
			"status", httpStatusFromError(err),
			"user_message", userMessage(err),
		)
	}

	// Check for stack trace
	if stack, ok := GetStack(err); ok {
		logger.Debug("error stack trace", "stack", stack)
	}
}

func main() {
	fmt.Println("=== Error Mastery ===")

	// Level 2: Domain errors
	fmt.Println("\n--- Domain Errors ---")
	err := E("users.Create", ErrAlreadyExists).
		WithEntity("user", "vikram@test.com").
		WithCause(fmt.Errorf("unique constraint violation"))
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Is AlreadyExists: %v\n", errors.Is(err, ErrAlreadyExists))
	fmt.Printf("HTTP Status: %d\n", httpStatusFromError(err))

	// Level 3: Validation
	fmt.Println("\n--- Validation Errors ---")
	valErr := NewValidation().
		Check(false, "name", "is required").
		Check(false, "email", "must contain @").
		Check(true, "age", "must be positive"). // passes — not added
		Build()
	fmt.Printf("Validation: %v\n", valErr)
	fmt.Printf("Is InvalidInput: %v\n", errors.Is(valErr, ErrInvalidInput))

	var ve *ValidationError
	if errors.As(valErr, &ve) {
		fmt.Printf("Failed fields: %d\n", len(ve.Fields))
		for _, f := range ve.Fields {
			fmt.Printf("  - %s: %s\n", f.Field, f.Message)
		}
	}

	// Level 4: errors.Join
	demonstrateJoin()

	// Level 5: Error chain
	fmt.Println("\n--- Error Wrapping Chain ---")
	chainErr := handleUserRequest("123")
	fmt.Printf("Full chain: %v\n", chainErr)

	// Level 6: Is vs As
	demonstrateIsVsAs()

	// Level 7: HTTP mapping
	fmt.Println("\n--- HTTP Status Mapping ---")
	for _, sentinel := range []error{ErrNotFound, ErrInvalidInput, ErrUnauthorized, ErrForbidden, ErrRateLimit} {
		testErr := E("test", sentinel)
		fmt.Printf("  %-20s → %d %s\n", sentinel, httpStatusFromError(testErr), userMessage(testErr))
	}

	// Anti-patterns
	antiPatterns()

	// Logging
	demonstrateErrorLogging()
}
