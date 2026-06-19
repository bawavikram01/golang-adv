package lowleveldesign

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// LLD PROBLEM: RATE LIMITER (Multiple Algorithms)
// =============================================================================
// Design a rate limiter suitable for an API gateway.
//
// Requirements:
// 1. Multiple algorithms: Token Bucket, Sliding Window, Fixed Window
// 2. Per-client limiting (by API key, IP, user ID)
// 3. Configurable limits (requests/second, requests/minute)
// 4. Thread-safe
// 5. Distributed-ready (interface supports Redis backend)
// 6. HTTP middleware integration
//
// Patterns Used:
// - Strategy Pattern: Multiple rate limiting algorithms
// - Factory Method: Create limiter by type
// - Decorator: Wrap HTTP handlers with rate limiting
// - Proxy: Control access to the underlying service

// =============================================================================
// CORE INTERFACE
// =============================================================================

// RateLimiterAlgo is the strategy interface
type RateLimiterAlgo interface {
	Allow(key string) bool
	AllowN(key string, n int) bool
	Reset(key string)
}

// =============================================================================
// ALGORITHM 1: TOKEN BUCKET
// =============================================================================
// Best for: Allowing bursts while maintaining average rate.
// Used by: AWS, Stripe, most API gateways.
//
// How it works:
// - Bucket holds tokens (max = burst size)
// - Tokens are added at a fixed rate (refill rate)
// - Each request consumes 1 token
// - No tokens = rejected

type TokenBucketLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	rate      float64 // tokens per second
	burstSize int     // max tokens
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

func NewTokenBucketLimiter(rate float64, burstSize int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		buckets:   make(map[string]*bucket),
		rate:      rate,
		burstSize: burstSize,
	}
}

func (l *TokenBucketLimiter) getBucket(key string) *bucket {
	b, exists := l.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     float64(l.burstSize),
			lastRefill: time.Now(),
		}
		l.buckets[key] = b
	}
	return b
}

func (l *TokenBucketLimiter) Allow(key string) bool {
	return l.AllowN(key, 1)
}

func (l *TokenBucketLimiter) AllowN(key string, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b := l.getBucket(key)

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > float64(l.burstSize) {
		b.tokens = float64(l.burstSize)
	}
	b.lastRefill = now

	// Check if enough tokens
	if b.tokens >= float64(n) {
		b.tokens -= float64(n)
		return true
	}
	return false
}

func (l *TokenBucketLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// =============================================================================
// ALGORITHM 2: SLIDING WINDOW LOG
// =============================================================================
// Best for: Precise rate limiting without burst issues.
// Trade-off: Higher memory usage (stores all timestamps).
//
// How it works:
// - Store timestamp of each request
// - Count requests in the last window duration
// - If count >= limit, reject

type SlidingWindowLogLimiter struct {
	mu     sync.Mutex
	logs   map[string][]time.Time
	limit  int
	window time.Duration
}

func NewSlidingWindowLogLimiter(limit int, window time.Duration) *SlidingWindowLogLimiter {
	return &SlidingWindowLogLimiter{
		logs:   make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
}

func (l *SlidingWindowLogLimiter) Allow(key string) bool {
	return l.AllowN(key, 1)
}

func (l *SlidingWindowLogLimiter) AllowN(key string, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-l.window)

	// Remove expired entries
	logs := l.logs[key]
	validStart := 0
	for i, t := range logs {
		if t.After(windowStart) {
			validStart = i
			break
		}
		if i == len(logs)-1 {
			validStart = len(logs)
		}
	}
	logs = logs[validStart:]

	// Check limit
	if len(logs)+n > l.limit {
		l.logs[key] = logs
		return false
	}

	// Add new entries
	for i := 0; i < n; i++ {
		logs = append(logs, now)
	}
	l.logs[key] = logs
	return true
}

func (l *SlidingWindowLogLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.logs, key)
}

// =============================================================================
// ALGORITHM 3: FIXED WINDOW COUNTER
// =============================================================================
// Best for: Simple, memory-efficient limiting.
// Trade-off: Burst at window boundaries (2x limit possible).
//
// How it works:
// - Divide time into fixed windows (e.g., per second)
// - Count requests in current window
// - Reset count at window boundary

type FixedWindowLimiter struct {
	mu      sync.Mutex
	windows map[string]*windowCounter
	limit   int
	window  time.Duration
}

type windowCounter struct {
	count       int
	windowStart time.Time
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		windows: make(map[string]*windowCounter),
		limit:   limit,
		window:  window,
	}
}

func (l *FixedWindowLimiter) Allow(key string) bool {
	return l.AllowN(key, 1)
}

func (l *FixedWindowLimiter) AllowN(key string, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, exists := l.windows[key]
	if !exists || now.Sub(w.windowStart) >= l.window {
		// New window
		l.windows[key] = &windowCounter{
			count:       n,
			windowStart: now.Truncate(l.window),
		}
		return n <= l.limit
	}

	// Same window
	if w.count+n > l.limit {
		return false
	}
	w.count += n
	return true
}

func (l *FixedWindowLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.windows, key)
}

// =============================================================================
// ALGORITHM 4: SLIDING WINDOW COUNTER (Hybrid)
// =============================================================================
// Best for: Good balance of accuracy and memory.
// Combines fixed window + weighted previous window.
//
// How it works:
// - Track counts for current and previous windows
// - Weighted count = prev_count * overlap_ratio + current_count
// - If weighted count >= limit, reject

type SlidingWindowCounterLimiter struct {
	mu      sync.Mutex
	windows map[string]*slidingWindow
	limit   int
	window  time.Duration
}

type slidingWindow struct {
	prevCount       int
	currCount       int
	currWindowStart time.Time
}

func NewSlidingWindowCounterLimiter(limit int, window time.Duration) *SlidingWindowCounterLimiter {
	return &SlidingWindowCounterLimiter{
		windows: make(map[string]*slidingWindow),
		limit:   limit,
		window:  window,
	}
}

func (l *SlidingWindowCounterLimiter) Allow(key string) bool {
	return l.AllowN(key, 1)
}

func (l *SlidingWindowCounterLimiter) AllowN(key string, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, exists := l.windows[key]

	if !exists {
		l.windows[key] = &slidingWindow{
			currCount:       n,
			currWindowStart: now.Truncate(l.window),
		}
		return n <= l.limit
	}

	// Check if we've moved to a new window
	elapsed := now.Sub(w.currWindowStart)
	if elapsed >= l.window {
		w.prevCount = w.currCount
		w.currCount = 0
		w.currWindowStart = now.Truncate(l.window)
		elapsed = now.Sub(w.currWindowStart)
	}

	// Calculate weighted count
	overlapRatio := float64(l.window-elapsed) / float64(l.window)
	weightedCount := float64(w.prevCount)*overlapRatio + float64(w.currCount)

	if int(weightedCount)+n > l.limit {
		return false
	}

	w.currCount += n
	return true
}

func (l *SlidingWindowCounterLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.windows, key)
}

// =============================================================================
// RATE LIMITER MIDDLEWARE (Decorator Pattern)
// =============================================================================

type APIHandler func(ctx context.Context, request map[string]string) (map[string]string, error)

// RateLimitMiddlewareFunc wraps an API handler with rate limiting
func RateLimitMiddlewareFunc(limiter RateLimiterAlgo, keyExtractor func(map[string]string) string) func(APIHandler) APIHandler {
	return func(next APIHandler) APIHandler {
		return func(ctx context.Context, request map[string]string) (map[string]string, error) {
			key := keyExtractor(request)
			if !limiter.Allow(key) {
				return map[string]string{
					"error":  "rate limit exceeded",
					"status": "429",
				}, fmt.Errorf("rate limit exceeded for %s", key)
			}
			return next(ctx, request)
		}
	}
}

// =============================================================================
// FACTORY: Create limiter by algorithm name
// =============================================================================

type RateLimiterConfig struct {
	Algorithm string
	Rate      float64
	Limit     int
	Window    time.Duration
	BurstSize int
}

func NewRateLimiterFromConfig(config RateLimiterConfig) RateLimiterAlgo {
	switch config.Algorithm {
	case "token_bucket":
		return NewTokenBucketLimiter(config.Rate, config.BurstSize)
	case "sliding_window_log":
		return NewSlidingWindowLogLimiter(config.Limit, config.Window)
	case "fixed_window":
		return NewFixedWindowLimiter(config.Limit, config.Window)
	case "sliding_window_counter":
		return NewSlidingWindowCounterLimiter(config.Limit, config.Window)
	default:
		// Default to token bucket
		return NewTokenBucketLimiter(config.Rate, config.BurstSize)
	}
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================

func ExampleRateLimiterSystem() {
	fmt.Println("=== Token Bucket (10 req/s, burst 20) ===")
	tb := NewTokenBucketLimiter(10, 20)
	for i := 0; i < 25; i++ {
		allowed := tb.Allow("user-1")
		if !allowed {
			fmt.Printf("  Request %d: REJECTED\n", i+1)
		}
	}

	fmt.Println("\n=== Fixed Window (5 req/sec) ===")
	fw := NewFixedWindowLimiter(5, time.Second)
	for i := 0; i < 8; i++ {
		allowed := fw.Allow("user-2")
		fmt.Printf("  Request %d: %v\n", i+1, allowed)
	}

	fmt.Println("\n=== Sliding Window Log (3 req/sec) ===")
	sw := NewSlidingWindowLogLimiter(3, time.Second)
	for i := 0; i < 5; i++ {
		allowed := sw.Allow("user-3")
		fmt.Printf("  Request %d: %v\n", i+1, allowed)
	}

	fmt.Println("\n=== Middleware Usage ===")
	// Create rate-limited API
	limiter := NewTokenBucketLimiter(2, 5) // 2 req/s, burst 5
	middleware := RateLimitMiddlewareFunc(limiter, func(req map[string]string) string {
		return req["api_key"]
	})

	handler := func(ctx context.Context, req map[string]string) (map[string]string, error) {
		return map[string]string{"status": "200", "data": "success"}, nil
	}

	rateLimitedHandler := middleware(handler)

	for i := 0; i < 7; i++ {
		resp, err := rateLimitedHandler(context.Background(), map[string]string{
			"api_key": "key-123",
		})
		if err != nil {
			fmt.Printf("  API call %d: %s\n", i+1, resp["error"])
		} else {
			fmt.Printf("  API call %d: %s\n", i+1, resp["status"])
		}
	}
}
