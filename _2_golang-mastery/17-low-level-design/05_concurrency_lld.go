package lowleveldesign

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// CONCURRENCY PATTERNS FOR LOW-LEVEL DESIGN
// =============================================================================
// Go's concurrency primitives are DESIGN TOOLS. They are part of your LLD.
//
// Patterns covered:
// 1. Worker Pool (bounded parallelism)
// 2. Pipeline (staged processing)
// 3. Fan-Out/Fan-In
// 4. Circuit Breaker
// 5. Semaphore
// 6. Rate Limiter (Token Bucket)

// =============================================================================
// 1. WORKER POOL PATTERN
// =============================================================================
// Process N jobs with M workers (M < N). Controls resource usage.
// Use when: API calls, DB queries, file processing with bounded concurrency.

type Job interface {
	ID() string
	Process() error
}

type Result struct {
	JobID string
	Err   error
}

type WorkerPool struct {
	numWorkers int
	jobs       chan Job
	results    chan Result
	wg         sync.WaitGroup
}

func NewWorkerPool(numWorkers, jobBufferSize int) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		jobs:       make(chan Job, jobBufferSize),
		results:    make(chan Result, jobBufferSize),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-wp.jobs:
			if !ok {
				return // channel closed
			}
			err := job.Process()
			wp.results <- Result{JobID: job.ID(), Err: err}
		}
	}
}

func (wp *WorkerPool) Submit(job Job) {
	wp.jobs <- job
}

func (wp *WorkerPool) Close() {
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
}

func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

// --- Example job ---
type DownloadJob struct {
	url string
}

func (j *DownloadJob) ID() string { return j.url }
func (j *DownloadJob) Process() error {
	fmt.Printf("  Downloading: %s\n", j.url)
	time.Sleep(100 * time.Millisecond) // simulate work
	return nil
}

// =============================================================================
// 2. PIPELINE PATTERN
// =============================================================================
// Chain of processing stages connected by channels.
// Each stage: receive from input channel -> process -> send to output channel.

type PipelineStage[In any, Out any] func(ctx context.Context, input <-chan In) <-chan Out

// Generator: produces values from a slice
func Generate[T any](ctx context.Context, values ...T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, v := range values {
			select {
			case <-ctx.Done():
				return
			case out <- v:
			}
		}
	}()
	return out
}

// Transform: applies a function to each value
func Transform[In any, Out any](ctx context.Context, input <-chan In, fn func(In) Out) <-chan Out {
	out := make(chan Out)
	go func() {
		defer close(out)
		for v := range input {
			select {
			case <-ctx.Done():
				return
			case out <- fn(v):
			}
		}
	}()
	return out
}

// Filter: only passes values that match predicate
func Filter[T any](ctx context.Context, input <-chan T, predicate func(T) bool) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for v := range input {
			if predicate(v) {
				select {
				case <-ctx.Done():
					return
				case out <- v:
				}
			}
		}
	}()
	return out
}

// Example: Number processing pipeline
func ExamplePipeline() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stage 1: Generate numbers
	numbers := Generate(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// Stage 2: Square them
	squared := Transform(ctx, numbers, func(n int) int { return n * n })

	// Stage 3: Filter > 20
	filtered := Filter(ctx, squared, func(n int) bool { return n > 20 })

	// Stage 4: Collect results
	for v := range filtered {
		fmt.Println(v) // 25, 36, 49, 64, 81, 100
	}
}

// =============================================================================
// 3. FAN-OUT / FAN-IN
// =============================================================================
// Fan-Out: One channel feeds multiple goroutines
// Fan-In: Multiple channels merge into one

// Fan-Out: Distribute work across N workers
func FanOut[T any](ctx context.Context, input <-chan T, n int, worker func(T) T) []<-chan T {
	outputs := make([]<-chan T, n)
	for i := 0; i < n; i++ {
		outputs[i] = Transform(ctx, input, worker)
	}
	return outputs
}

// Fan-In: Merge multiple channels into one
func FanIn[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	merged := make(chan T)

	// Start a goroutine for each input channel
	output := func(ch <-chan T) {
		defer wg.Done()
		for v := range ch {
			select {
			case <-ctx.Done():
				return
			case merged <- v:
			}
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go output(ch)
	}

	// Close merged channel when all inputs are done
	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

// =============================================================================
// 4. CIRCUIT BREAKER PATTERN
// =============================================================================
// Prevent cascading failures. If a service fails N times, stop calling it
// temporarily and fail fast. After a timeout, try again.
//
// States: CLOSED (normal) -> OPEN (failing fast) -> HALF-OPEN (testing)

type CircuitState int

const (
	StateClosed   CircuitState = iota // Normal operation
	StateOpen                         // Failing fast
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
	failureThreshold int
	successThreshold int // successes needed in half-open to close
	timeout          time.Duration
	lastFailureTime  time.Time
}

func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Check if we should transition from OPEN to HALF-OPEN
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			fmt.Println("[CircuitBreaker] OPEN -> HALF-OPEN (testing)")
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker is OPEN — failing fast")
		}
	}
	cb.mu.Unlock()

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.state == StateHalfOpen {
			// Any failure in half-open goes back to open
			cb.state = StateOpen
			fmt.Println("[CircuitBreaker] HALF-OPEN -> OPEN (still failing)")
		} else if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
			fmt.Printf("[CircuitBreaker] CLOSED -> OPEN (failed %d times)\n", cb.failureCount)
		}
		return err
	}

	// Success
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
			fmt.Println("[CircuitBreaker] HALF-OPEN -> CLOSED (recovered!)")
		}
	} else {
		cb.failureCount = 0 // reset on success in closed state
	}

	return nil
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// =============================================================================
// 5. SEMAPHORE (Bounded Concurrency)
// =============================================================================
// Limit concurrent access to a resource. Go doesn't have built-in semaphores,
// but a buffered channel works perfectly.

type Semaphore struct {
	sem chan struct{}
}

func NewSemaphore(maxConcurrency int) *Semaphore {
	return &Semaphore{
		sem: make(chan struct{}, maxConcurrency),
	}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) Release() {
	<-s.sem
}

// Usage: Limit concurrent HTTP requests
func ExampleSemaphore() {
	sem := NewSemaphore(3) // max 3 concurrent operations
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := sem.Acquire(ctx); err != nil {
				return
			}
			defer sem.Release()

			// Only 3 of these run concurrently
			fmt.Printf("Worker %d processing\n", id)
			time.Sleep(100 * time.Millisecond)
		}(i)
	}
	wg.Wait()
}

// =============================================================================
// 6. TOKEN BUCKET RATE LIMITER
// =============================================================================
// Controls the rate of operations. Tokens are added at a fixed rate.
// Each operation consumes a token. No token = blocked/rejected.

type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens, // start full
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefill = now
}

// Allow: returns true if request is allowed (non-blocking)
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// Wait: blocks until a token is available or context is cancelled
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if tb.Allow() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// retry after small delay
		}
	}
}

// --- Per-client rate limiter (used in API servers) ---
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*TokenBucket
	rate     float64
	burst    float64
}

func NewRateLimiter(rate, burst float64) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*TokenBucket),
		rate:     rate,
		burst:    burst,
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	limiter, exists := rl.limiters[clientID]
	if !exists {
		limiter = NewTokenBucket(rl.burst, rl.rate)
		rl.limiters[clientID] = limiter
	}
	rl.mu.Unlock()

	return limiter.Allow()
}

func ExampleRateLimiter() {
	// 5 requests per second, burst of 10
	limiter := NewRateLimiter(5, 10)

	// Simulate requests from two clients
	for i := 0; i < 15; i++ {
		allowed := limiter.Allow("client-1")
		fmt.Printf("Request %d from client-1: %v\n", i+1, allowed)
	}
}
