// =============================================================================
// LESSON 7: ADVANCED DESIGN PATTERNS IN GO
// =============================================================================
//
// Go favors simplicity, but complex systems need patterns.
// These patterns are idiomatic Go — not blind ports from Java/C++.
// =============================================================================

package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// PATTERN 1: Circuit Breaker — Prevent cascading failures
// =============================================================================
// States: Closed (normal) → Open (failing) → Half-Open (testing recovery)

type CircuitState int

const (
	StateClosed   CircuitState = iota // normal operation
	StateOpen                         // rejecting all calls
	StateHalfOpen                     // testing if service recovered
)

type CircuitBreaker struct {
	mu             sync.Mutex
	state          CircuitState
	failures       int
	successes      int
	maxFailures    int
	resetTimeout   time.Duration
	lastFailure    time.Time
	halfOpenMax    int
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		halfOpenMax:  1,
	}
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	switch cb.state {
	case StateOpen:
		// Check if enough time passed to try half-open
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			cb.mu.Unlock()
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	case StateClosed:
		cb.mu.Unlock()
	case StateHalfOpen:
		cb.mu.Unlock()
	}

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.state == StateHalfOpen || cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}

	// Success
	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= cb.halfOpenMax {
			cb.state = StateClosed
			cb.failures = 0
		}
	} else {
		cb.failures = 0 // reset on success
	}
	return nil
}

// =============================================================================
// PATTERN 2: Worker Pool — Bounded goroutine concurrency
// =============================================================================

type Job func() error

type WorkerPool struct {
	jobs    chan Job
	results chan error
	wg      sync.WaitGroup
}

func NewWorkerPool(workers, queueSize int) *WorkerPool {
	wp := &WorkerPool{
		jobs:    make(chan Job, queueSize),
		results: make(chan error, queueSize),
	}

	for i := 0; i < workers; i++ {
		wp.wg.Add(1)
		go func(id int) {
			defer wp.wg.Done()
			for job := range wp.jobs {
				wp.results <- job()
			}
		}(i)
	}
	return wp
}

func (wp *WorkerPool) Submit(job Job) {
	wp.jobs <- job
}

func (wp *WorkerPool) Shutdown() {
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
}

// =============================================================================
// PATTERN 3: Pub/Sub — Event-driven architecture
// =============================================================================

type EventType string

type Event struct {
	Type    EventType
	Payload interface{}
}

type EventHandler func(Event)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{handlers: make(map[EventType][]EventHandler)}
}

func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	handlers := make([]EventHandler, len(eb.handlers[event.Type]))
	copy(handlers, eb.handlers[event.Type])
	eb.mu.RUnlock()

	for _, handler := range handlers {
		handler(event) // could be async with goroutines
	}
}

// =============================================================================
// PATTERN 4: Pipeline with error handling and backpressure
// =============================================================================

type Stage[T any] func(ctx context.Context, in <-chan T) <-chan T

func Pipeline[T any](ctx context.Context, source <-chan T, stages ...Stage[T]) <-chan T {
	current := source
	for _, stage := range stages {
		current = stage(ctx, current)
	}
	return current
}

func TransformStage[T any](transform func(T) T, bufSize int) Stage[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T, bufSize) // buffered for backpressure control
		go func() {
			defer close(out)
			for item := range in {
				select {
				case out <- transform(item):
				case <-ctx.Done():
					return
				}
			}
		}()
		return out
	}
}

// =============================================================================
// PATTERN 5: Retry with exponential backoff and jitter
// =============================================================================

type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     5 * time.Second,
		Multiplier:  2.0,
	}
}

func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	wait := cfg.InitialWait

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt == cfg.MaxRetries {
			break
		}

		// Wait with context cancellation support
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}

		// Exponential backoff
		wait = time.Duration(float64(wait) * cfg.Multiplier)
		if wait > cfg.MaxWait {
			wait = cfg.MaxWait
		}
	}
	return fmt.Errorf("after %d retries: %w", cfg.MaxRetries, lastErr)
}

// =============================================================================
// PATTERN 6: Singleflight — Deduplicate concurrent requests
// =============================================================================
// When 100 goroutines request the same cache key simultaneously,
// only ONE should hit the database. All others wait for that result.

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type SingleFlight struct {
	mu sync.Mutex
	m  map[string]*call
}

func NewSingleFlight() *SingleFlight {
	return &SingleFlight{m: make(map[string]*call)}
}

func (sf *SingleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	sf.mu.Lock()

	if c, ok := sf.m[key]; ok {
		sf.mu.Unlock()
		c.wg.Wait() // wait for the in-flight call
		return c.val, c.err
	}

	c := &call{}
	c.wg.Add(1)
	sf.m[key] = c
	sf.mu.Unlock()

	c.val, c.err = fn() // only one goroutine executes this
	c.wg.Done()

	sf.mu.Lock()
	delete(sf.m, key)
	sf.mu.Unlock()

	return c.val, c.err
}

// =============================================================================
// PATTERN 7: Errgroup-like — Structured concurrency
// =============================================================================

type Group struct {
	wg     sync.WaitGroup
	errMu  sync.Mutex
	err    error
	cancel context.CancelFunc
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

func (g *Group) Go(fn func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := fn(); err != nil {
			g.errMu.Lock()
			if g.err == nil {
				g.err = err
				g.cancel() // cancel all other goroutines
			}
			g.errMu.Unlock()
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}

func main() {
	// Circuit Breaker
	fmt.Println("=== Circuit Breaker ===")
	cb := NewCircuitBreaker(3, 1*time.Second)
	callCount := 0
	for i := 0; i < 7; i++ {
		err := cb.Execute(func() error {
			callCount++
			return fmt.Errorf("service unavailable")
		})
		if errors.Is(err, ErrCircuitOpen) {
			fmt.Printf("  Call %d: REJECTED (circuit open)\n", i+1)
		} else {
			fmt.Printf("  Call %d: %v\n", i+1, err)
		}
	}
	fmt.Printf("  Actual calls made: %d (rest were short-circuited)\n", callCount)

	// Worker Pool
	fmt.Println("\n=== Worker Pool ===")
	pool := NewWorkerPool(3, 10)
	for i := 0; i < 5; i++ {
		id := i
		pool.Submit(func() error {
			fmt.Printf("  Job %d executed\n", id)
			return nil
		})
	}
	pool.Shutdown()

	// Event Bus
	fmt.Println("\n=== Event Bus (Pub/Sub) ===")
	bus := NewEventBus()
	bus.Subscribe("user.created", func(e Event) {
		fmt.Printf("  Handler 1: User created: %v\n", e.Payload)
	})
	bus.Subscribe("user.created", func(e Event) {
		fmt.Printf("  Handler 2: Send welcome email to %v\n", e.Payload)
	})
	bus.Publish(Event{Type: "user.created", Payload: "vikram@example.com"})

	// Singleflight
	fmt.Println("\n=== Singleflight ===")
	sf := NewSingleFlight()
	var wg sync.WaitGroup
	fetchCount := 0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			val, _ := sf.Do("user:1", func() (interface{}, error) {
				fetchCount++
				time.Sleep(50 * time.Millisecond)
				return "user_data", nil
			})
			_ = val
		}(i)
	}
	wg.Wait()
	fmt.Printf("  10 concurrent requests, actual fetches: %d\n", fetchCount)

	// Retry
	fmt.Println("\n=== Retry with Backoff ===")
	ctx := context.Background()
	attempt := 0
	err := Retry(ctx, DefaultRetryConfig(), func() error {
		attempt++
		if attempt < 3 {
			return fmt.Errorf("temporary error")
		}
		return nil
	})
	fmt.Printf("  Succeeded after %d attempts, err: %v\n", attempt, err)
}
