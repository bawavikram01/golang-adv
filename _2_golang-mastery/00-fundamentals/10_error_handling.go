//go:build ignore

// =============================================================================
// LESSON 0.10: ERROR HANDLING — Go's Explicit Error Philosophy
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - The error interface and why Go chose explicit errors
// - Creating errors: errors.New, fmt.Errorf
// - Error wrapping with %w (Go 1.13+)
// - errors.Is and errors.As for inspection
// - Custom error types
// - Sentinel errors
// - panic, recover, and when to use them
// - Production error handling patterns
//
// THE KEY INSIGHT:
// Go treats errors as VALUES, not exceptions. You handle them at every call
// site. This is intentional: it makes error paths visible and forces you
// to think about failure at every step. It's verbose, but it's clear.
//
// NOTE: This covers fundamentals. Advanced error patterns (error trees,
// domain errors, error middleware) are in 14-error-mastery/.
//
// RUN: go run 10_error_handling.go
// =============================================================================

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

func main() {
	fmt.Println("=== ERROR HANDLING ===")
	fmt.Println()

	errorBasics()
	creatingErrors()
	errorWrapping()
	errorsIsAs()
	customErrorTypes()
	sentinelErrors()
	panicAndRecover()
	bestPractices()
}

// =============================================================================
// PART 1: The error Interface
// =============================================================================
func errorBasics() {
	fmt.Println("--- ERROR BASICS ---")

	// The error type is a built-in interface:
	// type error interface {
	//     Error() string
	// }
	// ANY type with an Error() string method is an error.

	// ─── The "comma-error" pattern ───
	// Go functions return errors as the LAST return value:
	val, err := strconv.Atoi("42")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Parsed: %d\n", val)

	// ─── Always check errors ───
	_, err = strconv.Atoi("not_a_number")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	}

	// ─── Why not exceptions? ───
	// In Java/Python: exceptions jump up the stack silently.
	// You don't know which functions might throw.
	// In Go: errors are VALUES in the return signature.
	// You can see exactly where errors come from.
	// Yes, it's verbose. But it's explicit and clear.

	// ─── The zero value of error is nil (no error) ───
	var e error
	fmt.Printf("  nil error: %v, == nil: %v\n", e, e == nil)

	fmt.Println()
}

// =============================================================================
// PART 2: Creating Errors
// =============================================================================

// Sentinel errors: package-level, named error values
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
)

func creatingErrors() {
	fmt.Println("--- CREATING ERRORS ---")

	// ─── errors.New: simple static error ───
	err1 := errors.New("something went wrong")
	fmt.Printf("  errors.New: %v\n", err1)

	// ─── fmt.Errorf: formatted error with context ───
	userID := 42
	err2 := fmt.Errorf("user %d not found in database", userID)
	fmt.Printf("  fmt.Errorf: %v\n", err2)

	// ─── When to use which ───
	// errors.New    → simple, static messages (often as sentinel errors)
	// fmt.Errorf    → dynamic messages with context
	// custom type   → when callers need to extract structured info

	fmt.Println()
}

// =============================================================================
// PART 3: Error Wrapping (Go 1.13+)
// =============================================================================
func errorWrapping() {
	fmt.Println("--- ERROR WRAPPING ---")

	// ─── Wrap errors with %w to add context while preserving the original ───
	original := errors.New("connection refused")
	wrapped := fmt.Errorf("failed to connect to database: %w", original)
	fmt.Printf("  Wrapped: %v\n", wrapped)

	// ─── Unwrap to get the original ───
	unwrapped := errors.Unwrap(wrapped)
	fmt.Printf("  Unwrapped: %v\n", unwrapped)

	// ─── Multi-level wrapping ───
	err := fmt.Errorf("handler failed: %w",
		fmt.Errorf("service error: %w",
			fmt.Errorf("repo error: %w", original)))
	fmt.Printf("  Chain: %v\n", err)
	// Output: "handler failed: service error: repo error: connection refused"

	// ─── %w vs %v ───
	// %w: wraps the error (can be unwrapped with errors.Is/As)
	// %v: formats the error string (NO wrapping, loses the chain)
	//
	// ALWAYS use %w when you want callers to inspect the cause.
	// Use %v only when you intentionally want to hide the underlying error.

	fmt.Println()
}

// =============================================================================
// PART 4: errors.Is and errors.As
// =============================================================================
func errorsIsAs() {
	fmt.Println("--- errors.Is AND errors.As ---")

	// ─── errors.Is: check if error matches a VALUE ───
	// Like == but works through wrapping chains
	err := fmt.Errorf("query failed: %w", ErrNotFound)

	if errors.Is(err, ErrNotFound) {
		fmt.Println("  errors.Is: found ErrNotFound in chain")
	}

	// Without wrapping, you'd need: err == ErrNotFound (doesn't work if wrapped)

	// ─── errors.As: extract a specific TYPE from the chain ───
	fileErr := openMissingFile()
	var pathErr *os.PathError
	if errors.As(fileErr, &pathErr) {
		fmt.Printf("  errors.As: PathError path=%s, op=%s\n", pathErr.Path, pathErr.Op)
	}

	// ─── errors.Is vs errors.As ───
	// errors.Is  → "Is this error (or any cause) this specific VALUE?"
	//              Used with sentinel errors: errors.Is(err, ErrNotFound)
	// errors.As  → "Is this error (or any cause) this TYPE?"
	//              Used with error types: errors.As(err, &pathErr)

	// ─── errors.Is replaces == ───
	// OLD: if err == ErrNotFound { ... }
	// NEW: if errors.Is(err, ErrNotFound) { ... }
	//
	// ─── errors.As replaces type assertion ───
	// OLD: if pathErr, ok := err.(*os.PathError); ok { ... }
	// NEW: if errors.As(err, &pathErr) { ... }

	fmt.Println()
}

func openMissingFile() error {
	_, err := os.Open("/nonexistent/file/path")
	if err != nil {
		return fmt.Errorf("failed to open config: %w", err)
	}
	return nil
}

// =============================================================================
// PART 5: Custom Error Types
// =============================================================================

// Custom error type: carries structured information
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: field %q: %s", e.Field, e.Message)
}

// Another custom error with wrapping support
type APIError struct {
	StatusCode int
	Message    string
	Err        error // underlying error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("API error %d: %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// Unwrap makes this work with errors.Is/As
func (e *APIError) Unwrap() error {
	return e.Err
}

func customErrorTypes() {
	fmt.Println("--- CUSTOM ERROR TYPES ---")

	// ─── Use custom types when callers need structured info ───
	err := validateAge(-5)
	fmt.Printf("  Error: %v\n", err)

	var valErr *ValidationError
	if errors.As(err, &valErr) {
		fmt.Printf("  Field: %s, Message: %s\n", valErr.Field, valErr.Message)
	}

	// ─── Custom error with wrapping ───
	apiErr := &APIError{
		StatusCode: 404,
		Message:    "user not found",
		Err:        ErrNotFound,
	}
	fmt.Printf("  API Error: %v\n", apiErr)

	// errors.Is works through Unwrap:
	if errors.Is(apiErr, ErrNotFound) {
		fmt.Println("  Underlying cause: ErrNotFound")
	}

	// ─── When to use custom error types ───
	// - Callers need to make decisions based on error details
	// - Error carries structured metadata (status code, field name, etc.)
	// - You need to implement Unwrap() for error chaining
	//
	// DON'T create custom types when a simple string error suffices.

	fmt.Println()
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{
			Field:   "age",
			Message: "must be non-negative",
		}
	}
	return nil
}

// =============================================================================
// PART 6: Sentinel Errors
// =============================================================================
func sentinelErrors() {
	fmt.Println("--- SENTINEL ERRORS ---")

	// Sentinel errors: package-level variables used for comparison.
	// Convention: var ErrXxx = errors.New("...")

	// stdlib examples:
	// io.EOF          → end of stream
	// sql.ErrNoRows   → query returned no rows
	// os.ErrNotExist  → file doesn't exist
	// context.Canceled        → context was cancelled
	// context.DeadlineExceeded → deadline passed

	// ─── Using sentinel errors ───
	result, err := findUser("unknown")
	if errors.Is(err, ErrNotFound) {
		fmt.Println("  User not found — return 404")
	} else if err != nil {
		fmt.Printf("  Unexpected error: %v\n", err)
	} else {
		fmt.Printf("  Found: %s\n", result)
	}

	// ─── RULES for sentinel errors ───
	// 1. Export them (ErrXxx, not errXxx) so callers can use errors.Is
	// 2. Use errors.New, not fmt.Errorf (sentinels are static)
	// 3. Don't add dynamic context to sentinels
	//    (wrap them instead: fmt.Errorf("user %d: %w", id, ErrNotFound))
	// 4. Document them in your package docs
	// 5. Be conservative: only create sentinels when callers need to
	//    differentiate between different error conditions

	fmt.Println()
}

func findUser(name string) (string, error) {
	users := map[string]string{
		"alice": "Alice Smith",
		"bob":   "Bob Jones",
	}
	if user, ok := users[name]; ok {
		return user, nil
	}
	return "", fmt.Errorf("findUser(%q): %w", name, ErrNotFound)
}

// =============================================================================
// PART 7: panic and recover
// =============================================================================
func panicAndRecover() {
	fmt.Println("--- PANIC AND RECOVER ---")

	// panic: unrecoverable error. Stops normal execution.
	// recover: catches a panic (only works inside deferred functions).
	//
	// panic unwinds the stack, running deferred functions.
	// If no recover catches it, the program crashes with a stack trace.

	// ─── When to panic ───
	// 1. Truly unrecoverable: corrupted state, programmer error
	// 2. init() failures (can't recover from broken initialization)
	// 3. NEVER for expected errors (file not found, connection failed)
	//
	// RULE: If it can happen in normal operation → return an error.
	//       If it means the program is broken → panic.

	// ─── recover example ───
	fmt.Printf("  Safe divide: %v\n", safeDivide(10, 0))
	fmt.Printf("  Safe divide: %v\n", safeDivide(10, 3))

	// ─── stdlib panics ───
	// These panic because they indicate programmer errors:
	// - Index out of range: a[10] when len(a) == 5
	// - Nil pointer dereference: (*T)(nil).Method()
	// - Send on closed channel
	// - Type assertion failure (without comma-ok)

	// ─── recover converts panic to error ───
	result, err := safeCall(func() {
		panic("something terrible")
	})
	fmt.Printf("  Recovered: result=%v, err=%v\n", result, err)

	fmt.Println()
}

func safeDivide(a, b int) (result string) {
	defer func() {
		if r := recover(); r != nil {
			result = fmt.Sprintf("recovered from panic: %v", r)
		}
	}()
	return fmt.Sprintf("%d / %d = %d", a, b, a/b)
}

func safeCall(fn func()) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	fn()
	return "ok", nil
}

// =============================================================================
// PART 8: Best Practices
// =============================================================================
func bestPractices() {
	fmt.Println("--- BEST PRACTICES ---")

	// ─── 1. Handle errors immediately ───
	// DON'T: result, _ = riskyFunction()   // ignoring error!
	// DO:    result, err := riskyFunction()
	//        if err != nil { return err }
	fmt.Println("  1. Never ignore errors (no _ for error returns)")

	// ─── 2. Add context when wrapping ───
	// DON'T: return err
	// DO:    return fmt.Errorf("fetching user %d: %w", id, err)
	//
	// Each layer adds its context:
	// "handler: get user profile: fetch user 42: connection refused"
	fmt.Println("  2. Wrap with context: fmt.Errorf(\"doing X: %w\", err)")

	// ─── 3. Don't wrap multiple times with same context ───
	// DON'T: return fmt.Errorf("error: %w", fmt.Errorf("error: %w", err))
	// Each function adds ITS context, not repeating what's below.
	fmt.Println("  3. Each layer adds its own unique context")

	// ─── 4. errors.Is/As over == and type assertion ───
	// OLD: if err == io.EOF
	// NEW: if errors.Is(err, io.EOF)
	fmt.Println("  4. Use errors.Is/As (not == or type assertion)")

	// ─── 5. Return early (guard clauses) ───
	// if err != nil {
	//     return err  // return early
	// }
	// // happy path continues (not indented)
	fmt.Println("  5. Return early — keep happy path unindented")

	// ─── 6. Don't panic for normal errors ───
	// panic is for programmer errors, not runtime conditions.
	// If a file might not exist → return error.
	// If a required config is nil → panic (programmer bug).
	fmt.Println("  6. panic = programmer bugs only")

	// ─── 7. Error messages: lowercase, no punctuation ───
	// Good: "opening config file"
	// Bad:  "Error: Failed to open config file."
	// Because errors chain: "handler: service: opening config file"
	// Not: "handler: Error: Failed to open config file."
	fmt.Println("  7. Lowercase, no period, no 'Error:' prefix")

	// ─── 8. Use error groups for concurrent operations ───
	// errgroup.Group (golang.org/x/sync/errgroup)
	// Collects errors from multiple goroutines
	fmt.Println("  8. errgroup for concurrent error collection")

	fmt.Println()

	fmt.Println("=== ERROR HANDLING SUMMARY ===")
	fmt.Println("  error is an interface: Error() string")
	fmt.Println("  Create: errors.New, fmt.Errorf")
	fmt.Println("  Wrap: fmt.Errorf(\"context: %w\", err)")
	fmt.Println("  Check: errors.Is (value), errors.As (type)")
	fmt.Println("  Custom: implement Error() + Unwrap()")
	fmt.Println("  panic/recover → last resort, not control flow")
}
