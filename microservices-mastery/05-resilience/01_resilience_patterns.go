// =============================================================================
// LESSON 5: RESILIENCE PATTERNS — Surviving Failure in Distributed Systems
// =============================================================================
//
// IN MICROSERVICES, FAILURE IS NOT EXCEPTIONAL — IT'S NORMAL.
//
// Networks fail. Services crash. Databases slow down. Cloud VMs vanish.
// Your system must SURVIVE partial failures gracefully.
//
// This lesson covers every resilience pattern you need:
//   1. Circuit Breaker       ("stop calling a dead service")
//   2. Retry with Backoff    ("try again, but be smart about it")
//   3. Bulkhead              ("isolate failures so they don't spread")
//   4. Timeout               ("don't wait forever")
//   5. Fallback              ("give a degraded but useful response")
//   6. Rate Limiter          ("protect yourself from too much traffic")
//   7. Health Checks         ("know when you're sick before users do")
// =============================================================================

package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// PATTERN 1: Circuit Breaker
// =============================================================================
//
// Inspired by electrical circuits: if too many failures, OPEN the circuit
// and stop sending requests. After a timeout, try one request (HALF-OPEN).
// If it succeeds, CLOSE the circuit. If it fails, OPEN again.
//
// STATES:
//   CLOSED   → requests flow normally. Count failures.
//   OPEN     → all requests fail immediately. No calls to downstream.
//   HALF-OPEN → allow ONE test request. Success → CLOSED, Failure → OPEN.
//
// WHY:
//   ✅ Prevents cascading failures (service A → B → C all fail)
//   ✅ Gives failing services time to recover
//   ✅ Fails fast instead of timing out (better UX)
//   ✅ Reduces load on struggling services
//
// REAL TOOLS: Sony/gobreaker, Netflix Hystrix (Java), resilience4j

type CircuitState int

const (
	StateClosed   CircuitState = iota // Normal operation
	StateOpen                         // Blocking requests
	StateHalfOpen                     // Testing recovery
)

func (s CircuitState) String() string {
	return [...]string{"CLOSED", "OPEN", "HALF-OPEN"}[s]
}

type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failureCount     int
	successCount     int
	failureThreshold int           // failures before opening
	successThreshold int           // successes in half-open before closing
	timeout          time.Duration // how long to stay open before half-open
	lastFailure      time.Time
	onStateChange    func(from, to CircuitState) // callback for monitoring
}

func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		onStateChange: func(from, to CircuitState) {
			fmt.Printf("  ⚡ Circuit: %s → %s\n", from, to)
		},
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Check if we should transition from OPEN to HALF-OPEN
	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.timeout {
		cb.setState(StateHalfOpen)
		cb.successCount = 0
	}

	// If OPEN, fail immediately (fast fail)
	if cb.state == StateOpen {
		cb.mu.Unlock()
		return fmt.Errorf("circuit breaker is OPEN: request blocked")
	}

	cb.mu.Unlock()

	// Execute the actual call
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()

		if cb.state == StateHalfOpen {
			// Any failure in half-open → back to open
			cb.setState(StateOpen)
		} else if cb.failureCount >= cb.failureThreshold {
			cb.setState(StateOpen)
		}
		return err
	}

	// Success
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.setState(StateClosed)
			cb.failureCount = 0
		}
	} else {
		cb.failureCount = 0 // reset on success in closed state
	}

	return nil
}

func (cb *CircuitBreaker) setState(newState CircuitState) {
	if cb.state != newState {
		old := cb.state
		cb.state = newState
		cb.onStateChange(old, newState)
	}
}

// =============================================================================
// PATTERN 2: Retry with Exponential Backoff + Jitter
// =============================================================================
//
// When a call fails, retry it. But DON'T retry immediately:
//   Attempt 1: wait 100ms
//   Attempt 2: wait 200ms
//   Attempt 3: wait 400ms
//   Attempt 4: wait 800ms
//
// ADD JITTER: randomize the wait time so all clients don't retry at the
// exact same time (thundering herd problem).
//
// RULES:
//   ✅ Only retry TRANSIENT errors (timeout, 503), NOT 400/404
//   ✅ Set a max retry count (don't retry forever)
//   ✅ Use idempotent operations only (safe to repeat)
//   ❌ Don't retry non-idempotent operations (e.g., create payment)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
	WithJitter bool
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
		WithJitter: true,
	}
}

func RetryWithBackoff(config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		lastErr = operation()
		if lastErr == nil {
			if attempt > 0 {
				fmt.Printf("  ✓ Succeeded on attempt %d\n", attempt+1)
			}
			return nil
		}

		if attempt == config.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt))
		if delay > float64(config.MaxDelay) {
			delay = float64(config.MaxDelay)
		}

		// Add jitter (±25%)
		if config.WithJitter {
			jitter := delay * 0.25 * (rand.Float64()*2 - 1) // -25% to +25%
			delay += jitter
		}

		actualDelay := time.Duration(delay)
		fmt.Printf("  ↻ Attempt %d failed: %v (retrying in %v)\n", attempt+1, lastErr, actualDelay)
		time.Sleep(actualDelay)
	}

	return fmt.Errorf("all %d retries exhausted: %w", config.MaxRetries+1, lastErr)
}

// =============================================================================
// PATTERN 3: Bulkhead
// =============================================================================
//
// Inspired by ship bulkheads: compartments that contain flooding.
// If one section floods, the ship doesn't sink.
//
// In microservices: limit concurrent calls to a dependency.
// If service A calls 3 downstream services, give each its own thread pool.
// If service B is slow, only its pool fills up — services C and D unaffected.
//
// IMPLEMENTATION: Semaphore (bounded channel in Go)

type Bulkhead struct {
	name    string
	sem     chan struct{}
	timeout time.Duration
}

func NewBulkhead(name string, maxConcurrent int, timeout time.Duration) *Bulkhead {
	return &Bulkhead{
		name:    name,
		sem:     make(chan struct{}, maxConcurrent),
		timeout: timeout,
	}
}

func (b *Bulkhead) Execute(fn func() error) error {
	// Try to acquire a permit
	select {
	case b.sem <- struct{}{}:
		// Got a permit
		defer func() { <-b.sem }()
		return fn()

	case <-time.After(b.timeout):
		return fmt.Errorf("bulkhead '%s' full: %d concurrent calls (rejected)", b.name, cap(b.sem))
	}
}

// =============================================================================
// PATTERN 4: Rate Limiter (Token Bucket)
// =============================================================================
//
// Control how many requests per second a service accepts.
// Protects against DDoS, misbehaving clients, and cascading load.
//
// TOKEN BUCKET ALGORITHM:
//   - Bucket holds N tokens (capacity).
//   - Each request consumes 1 token.
//   - Tokens are added at a fixed rate (e.g., 10/second).
//   - If bucket is empty → reject request (429 Too Many Requests).
//   - Allows short bursts up to bucket capacity.

type RateLimiter struct {
	tokens     int64
	capacity   int64
	refillRate time.Duration // time between adding one token
	mu         sync.Mutex
	lastRefill time.Time
}

func NewRateLimiter(capacity int64, ratePerSecond float64) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: time.Duration(float64(time.Second) / ratePerSecond),
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int64(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.capacity {
			rl.tokens = rl.capacity
		}
		rl.lastRefill = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

// =============================================================================
// PATTERN 5: Timeout with Context
// =============================================================================
//
// NEVER make a call without a timeout.
// In microservices, a slow call is worse than a failed call.
// Slow calls consume connections, goroutines, memory, and cascade.
//
// TIMEOUT BUDGET:
//   If your SLA is 500ms and you call 3 services:
//   - Service A: 150ms timeout
//   - Service B: 150ms timeout
//   - Service C: 150ms timeout
//   - Overhead: 50ms
//
// Always propagate cancellation via context.Context.

// =============================================================================
// PATTERN 6: Fallback
// =============================================================================
//
// When a service fails, return a degraded but useful response.
//
// EXAMPLES:
//   - Product recommendations fail → show popular products (cached)
//   - User profile pic fails → show default avatar
//   - Pricing service fails → show last known price (stale but useful)
//   - Search fails → show cached results

type FallbackWrapper struct {
	primary  func() (string, error)
	fallback func() (string, error)
}

func (f *FallbackWrapper) Execute() (string, error) {
	result, err := f.primary()
	if err != nil {
		fmt.Printf("  ↪ Primary failed: %v — using fallback\n", err)
		return f.fallback()
	}
	return result, nil
}

// =============================================================================
// PATTERN 7: Health Check
// =============================================================================
//
// Every microservice must expose health endpoints:
//   GET /health/live   — "Am I running?" (liveness)
//   GET /health/ready  — "Can I serve traffic?" (readiness)
//
// LIVENESS:  Process is alive. If fails → restart container.
// READINESS: Dependencies connected. If fails → remove from load balancer.
//
// CHECK: database ping, cache ping, disk space, memory, goroutine count

type HealthStatus string

const (
	HealthUp       HealthStatus = "UP"
	HealthDown     HealthStatus = "DOWN"
	HealthDegraded HealthStatus = "DEGRADED"
)

type HealthCheck struct {
	Status     HealthStatus      `json:"status"`
	Components map[string]string `json:"components"`
}

type Service struct {
	dbConnected    bool
	cacheConnected bool
	requestCount   atomic.Int64
}

func (s *Service) LivenessCheck() HealthCheck {
	// Just check if the process can respond
	return HealthCheck{
		Status:     HealthUp,
		Components: map[string]string{"process": "alive"},
	}
}

func (s *Service) ReadinessCheck() HealthCheck {
	components := make(map[string]string)
	status := HealthUp

	if s.dbConnected {
		components["database"] = "connected"
	} else {
		components["database"] = "disconnected"
		status = HealthDown
	}

	if s.cacheConnected {
		components["cache"] = "connected"
	} else {
		components["cache"] = "disconnected"
		if status != HealthDown {
			status = HealthDegraded // cache is nice-to-have
		}
	}

	return HealthCheck{Status: status, Components: components}
}

func main() {
	// =========================================================================
	// DEMO 1: Circuit Breaker
	// =========================================================================
	fmt.Println("=== CIRCUIT BREAKER ===")

	cb := NewCircuitBreaker(3, 2, 500*time.Millisecond)
	callCount := 0

	// Simulate an unreliable service
	unreliableCall := func() error {
		callCount++
		if callCount <= 4 {
			return errors.New("connection refused")
		}
		return nil // recovered
	}

	for i := 1; i <= 8; i++ {
		err := cb.Execute(unreliableCall)
		if err != nil {
			fmt.Printf("  Call %d: FAILED — %v\n", i, err)
		} else {
			fmt.Printf("  Call %d: SUCCESS\n", i)
		}

		// After circuit opens, wait for timeout to trigger half-open
		if i == 5 {
			fmt.Println("  ... waiting for circuit timeout ...")
			time.Sleep(600 * time.Millisecond)
		}
	}

	// =========================================================================
	// DEMO 2: Retry with Exponential Backoff
	// =========================================================================
	fmt.Println("\n=== RETRY WITH BACKOFF ===")

	attempt := 0
	err := RetryWithBackoff(RetryConfig{
		MaxRetries: 3,
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		Multiplier: 2.0,
		WithJitter: true,
	}, func() error {
		attempt++
		if attempt < 3 {
			return errors.New("temporary error")
		}
		return nil // succeeds on 3rd try
	})
	fmt.Printf("  Final result: %v\n", err)

	// =========================================================================
	// DEMO 3: Bulkhead
	// =========================================================================
	fmt.Println("\n=== BULKHEAD ===")

	bulkhead := NewBulkhead("payment-service", 2, 100*time.Millisecond)
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := bulkhead.Execute(func() error {
				time.Sleep(200 * time.Millisecond) // simulate slow call
				return nil
			})
			if err != nil {
				fmt.Printf("  Request %d: REJECTED — %v\n", id, err)
			} else {
				fmt.Printf("  Request %d: completed\n", id)
			}
		}(i)
		time.Sleep(10 * time.Millisecond) // stagger slightly
	}
	wg.Wait()

	// =========================================================================
	// DEMO 4: Rate Limiter
	// =========================================================================
	fmt.Println("\n=== RATE LIMITER (Token Bucket) ===")

	limiter := NewRateLimiter(3, 2.0) // 3 burst capacity, 2 tokens/sec refill

	for i := 1; i <= 6; i++ {
		if limiter.Allow() {
			fmt.Printf("  Request %d: ALLOWED\n", i)
		} else {
			fmt.Printf("  Request %d: RATE LIMITED (429)\n", i)
		}
	}

	// Wait for tokens to refill
	fmt.Println("  ... waiting 1 second for token refill ...")
	time.Sleep(1 * time.Second)

	for i := 7; i <= 9; i++ {
		if limiter.Allow() {
			fmt.Printf("  Request %d: ALLOWED\n", i)
		} else {
			fmt.Printf("  Request %d: RATE LIMITED (429)\n", i)
		}
	}

	// =========================================================================
	// DEMO 5: Fallback
	// =========================================================================
	fmt.Println("\n=== FALLBACK PATTERN ===")

	wrapper := &FallbackWrapper{
		primary: func() (string, error) {
			return "", errors.New("recommendation service down")
		},
		fallback: func() (string, error) {
			return "Popular items: [iPhone, MacBook, AirPods]", nil
		},
	}

	result, _ := wrapper.Execute()
	fmt.Printf("  Response: %s\n", result)

	// =========================================================================
	// DEMO 6: Health Check
	// =========================================================================
	fmt.Println("\n=== HEALTH CHECKS ===")

	svc := &Service{dbConnected: true, cacheConnected: false}
	liveness := svc.LivenessCheck()
	readiness := svc.ReadinessCheck()

	fmt.Printf("  Liveness:  %s %v\n", liveness.Status, liveness.Components)
	fmt.Printf("  Readiness: %s %v\n", readiness.Status, readiness.Components)

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println("\n=== RESILIENCE DECISION GUIDE ===")
	fmt.Println("┌──────────────────────────────┬──────────────────────────────────┐")
	fmt.Println("│ Problem                      │ Pattern                          │")
	fmt.Println("├──────────────────────────────┼──────────────────────────────────┤")
	fmt.Println("│ Service keeps failing         │ Circuit Breaker                 │")
	fmt.Println("│ Transient errors              │ Retry + Exponential Backoff     │")
	fmt.Println("│ Cascading failures            │ Bulkhead (isolation)            │")
	fmt.Println("│ Slow calls blocking threads   │ Timeout + Context               │")
	fmt.Println("│ Graceful degradation          │ Fallback                        │")
	fmt.Println("│ Too much traffic              │ Rate Limiter                    │")
	fmt.Println("│ Know when you're unhealthy    │ Health Checks (live + ready)    │")
	fmt.Println("│ All of the above combined     │ Use all together!               │")
	fmt.Println("└──────────────────────────────┴──────────────────────────────────┘")
	fmt.Println()
	fmt.Println("PRODUCTION STACK: Circuit Breaker wrapping Retry wrapping Timeout")
	fmt.Println("  cb.Execute(retry.Execute(timeout.Execute(actualCall)))")
}
