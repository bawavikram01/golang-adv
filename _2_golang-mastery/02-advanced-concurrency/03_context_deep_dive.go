// =============================================================================
// LESSON 2.3: CONTEXT DEEP DIVE — The Backbone of Go Concurrency
// =============================================================================
//
// context.Context controls cancellation, deadlines, and request-scoped values
// across goroutine trees. Every production Go program uses it extensively.
//
// RULES:
// 1. Always pass context as the first parameter: func Foo(ctx context.Context, ...)
// 2. Never store context in a struct (except in rare framework code)
// 3. Never pass nil context — use context.TODO() if unsure
// 4. Context values are for request-scoped data (trace IDs, auth tokens),
//    NOT for passing function parameters
// =============================================================================

package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// =============================================================================
// PATTERN 1: Cascading cancellation
// =============================================================================
// When a parent context is cancelled, ALL children are cancelled too.
// This creates a tree of cancellation that propagates automatically.

func demonstrateCascadingCancel() {
	fmt.Println("=== Cascading Cancellation ===")

	// Root context
	root, rootCancel := context.WithCancel(context.Background())

	// Child contexts — automatically cancelled when root is cancelled
	child1, _ := context.WithCancel(root)
	child2, _ := context.WithCancel(root)
	grandchild, _ := context.WithCancel(child1)

	// Launch workers on each context
	go worker(root, "root")
	go worker(child1, "child1")
	go worker(child2, "child2")
	go worker(grandchild, "grandchild")

	time.Sleep(200 * time.Millisecond)

	fmt.Println("Cancelling root — all children should stop")
	rootCancel()

	time.Sleep(100 * time.Millisecond) // let workers notice cancellation
	fmt.Println()
}

func worker(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("  [%s] stopped: %v\n", name, ctx.Err())
			return
		case <-time.After(50 * time.Millisecond):
			// working...
		}
	}
}

// =============================================================================
// PATTERN 2: Deadline propagation
// =============================================================================
// WithDeadline sets an absolute time. WithTimeout sets a relative duration.
// The effective deadline is the MINIMUM of parent and child deadlines.

func demonstrateDeadline() {
	fmt.Println("=== Deadline Propagation ===")

	// Parent has 500ms deadline
	parent, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Child tries to set 1s deadline — won't work!
	// Effective deadline is still 500ms (parent's deadline comes first)
	child, childCancel := context.WithTimeout(parent, 1*time.Second)
	defer childCancel()

	dl, ok := child.Deadline()
	fmt.Printf("Child deadline set: %v, deadline: %v from now\n", ok, time.Until(dl).Round(time.Millisecond))

	// Simulate work
	select {
	case <-child.Done():
		fmt.Printf("Child cancelled: %v\n", child.Err())
	}
	fmt.Println()
}

// =============================================================================
// PATTERN 3: Context values — Request-scoped data
// =============================================================================
// Use custom unexported key types to prevent collisions across packages.

// CORRECT: unexported type prevents other packages from setting this key
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFrom(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func handleRequest(ctx context.Context) {
	reqID, _ := RequestIDFrom(ctx)
	userID, _ := UserIDFrom(ctx)
	fmt.Printf("  Processing request %s for user %d\n", reqID, userID)
}

// =============================================================================
// PATTERN 4: context.AfterFunc (Go 1.21+)
// =============================================================================
// Register a callback that runs when context is cancelled.
// Useful for cleanup without blocking in a goroutine.

func demonstrateAfterFunc() {
	fmt.Println("=== context.AfterFunc ===")

	ctx, cancel := context.WithCancel(context.Background())

	// Register cleanup to run when ctx is cancelled
	stop := context.AfterFunc(ctx, func() {
		fmt.Println("  Cleanup function called after cancellation!")
	})

	// Can prevent the callback from running:
	_ = stop // stop() would prevent the callback

	cancel()
	time.Sleep(50 * time.Millisecond) // let callback run
	fmt.Println()
}

// =============================================================================
// PATTERN 5: context.WithCancelCause (Go 1.20+)
// =============================================================================
// Attach a specific error reason to cancellation.

func demonstrateCancelCause() {
	fmt.Println("=== WithCancelCause ===")

	ctx, cancel := context.WithCancelCause(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		// Cancel with a specific reason
		cancel(fmt.Errorf("database connection lost"))
	}()

	<-ctx.Done()
	fmt.Printf("  Cancelled: %v\n", ctx.Err())
	fmt.Printf("  Cause: %v\n", context.Cause(ctx))
	fmt.Println()
}

// =============================================================================
// PATTERN 6: Graceful shutdown with context
// =============================================================================

type Server struct {
	name string
}

func (s *Server) Start(ctx context.Context) error {
	fmt.Printf("  [%s] Starting...\n", s.name)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("  [%s] Shutting down gracefully...\n", s.name)
			// Perform cleanup
			time.Sleep(50 * time.Millisecond)
			fmt.Printf("  [%s] Shutdown complete\n", s.name)
			return ctx.Err()
		case <-ticker.C:
			fmt.Printf("  [%s] Processing...\n", s.name)
		}
	}
}

func demonstrateGracefulShutdown() {
	fmt.Println("=== Graceful Shutdown ===")

	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	// Start multiple servers
	errs := make(chan error, 2)
	go func() { errs <- (&Server{"HTTP"}).Start(ctx) }()
	go func() { errs <- (&Server{"gRPC"}).Start(ctx) }()

	// Wait for both to finish
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil && !errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("  Unexpected error: %v\n", err)
		}
	}
	fmt.Println()
}

func main() {
	demonstrateCascadingCancel()
	demonstrateDeadline()

	fmt.Println("=== Context Values ===")
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-abc-123")
	ctx = WithUserID(ctx, 42)
	handleRequest(ctx)
	fmt.Println()

	demonstrateAfterFunc()
	demonstrateCancelCause()
	demonstrateGracefulShutdown()
}
