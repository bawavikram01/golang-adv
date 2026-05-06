// =============================================================================
// LESSON 2.1: ADVANCED CHANNEL PATTERNS
// =============================================================================
//
// Channels are Go's primary synchronization primitive. This lesson covers
// patterns you'll use in production systems: fan-out/fan-in, pipelines,
// rate limiting, semaphores, and cancellation.
//
// MENTAL MODEL: A channel is a synchronized queue with optional buffering.
// Unbuffered = rendezvous (both sides block until handshake).
// Buffered = mailbox (sender blocks only when full).
// =============================================================================

package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// =============================================================================
// PATTERN 1: Generator — Function that returns a channel
// =============================================================================
// The function owns the goroutine lifecycle. Caller just reads from channel.
// Always pass context for cancellation support.

func generateNumbers(ctx context.Context, start, count int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // CRITICAL: always close when done producing
		for i := start; i < start+count; i++ {
			select {
			case out <- i:
			case <-ctx.Done():
				return // respect cancellation
			}
		}
	}()
	return out
}

// =============================================================================
// PATTERN 2: Fan-Out / Fan-In
// =============================================================================
// Fan-out: Multiple goroutines read from the same channel (work distribution)
// Fan-in: Multiple channels merged into one channel (result aggregation)

func fanOut(ctx context.Context, input <-chan int, workers int) []<-chan int {
	channels := make([]<-chan int, workers)
	for i := 0; i < workers; i++ {
		channels[i] = processWorker(ctx, input)
	}
	return channels
}

func processWorker(ctx context.Context, input <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range input {
			select {
			case out <- n * n: // simulate processing
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Fan-in: merge multiple channels into one
func fanIn(ctx context.Context, channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	// Start a goroutine for each input channel
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for n := range c {
				select {
				case out <- n:
				case <-ctx.Done():
					return
				}
			}
		}(ch)
	}

	// Close output when all input channels are drained
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// =============================================================================
// PATTERN 3: Pipeline — Chain of processing stages
// =============================================================================
// Each stage: receives from upstream, processes, sends downstream.
// Entire pipeline shuts down cleanly via context cancellation.

func multiply(ctx context.Context, in <-chan int, factor int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * factor:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func filter(ctx context.Context, in <-chan int, predicate func(int) bool) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if predicate(n) {
				select {
				case out <- n:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

func sink(ctx context.Context, in <-chan int) []int {
	var results []int
	for n := range in {
		results = append(results, n)
	}
	return results
}

// =============================================================================
// PATTERN 4: Semaphore — Bounded concurrency with buffered channels
// =============================================================================
// A buffered channel can act as a counting semaphore.

type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(maxConcurrency int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, maxConcurrency)}
}

func (s *Semaphore) Acquire() { s.ch <- struct{}{} }
func (s *Semaphore) Release() { <-s.ch }

func demonstrateSemaphore() {
	fmt.Println("\n=== Semaphore Pattern ===")
	sem := NewSemaphore(3) // max 3 concurrent operations
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()
			fmt.Printf("Worker %d: processing (max 3 concurrent)\n", id)
			time.Sleep(100 * time.Millisecond)
		}(i)
	}
	wg.Wait()
}

// =============================================================================
// PATTERN 5: Rate Limiter — Token bucket via ticker
// =============================================================================

type RateLimiter struct {
	tokens chan struct{}
	quit   chan struct{}
}

func NewRateLimiter(ratePerSecond int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, ratePerSecond), // burst capacity
		quit:   make(chan struct{}),
	}

	// Fill tokens at the specified rate
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(ratePerSecond))
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case rl.tokens <- struct{}{}: // add token if not full
				default: // bucket full, discard
				}
			case <-rl.quit:
				return
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *RateLimiter) Stop() { close(rl.quit) }

// =============================================================================
// PATTERN 6: Or-Done Channel — Read from a channel respecting context
// =============================================================================
// Encapsulates the "select on channel AND done" pattern so callers
// can simply range over the result.

func orDone(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// =============================================================================
// PATTERN 7: Tee Channel — Split one channel into two
// =============================================================================

func tee(ctx context.Context, in <-chan int) (<-chan int, <-chan int) {
	out1 := make(chan int)
	out2 := make(chan int)
	go func() {
		defer close(out1)
		defer close(out2)
		for val := range orDone(ctx, in) {
			// Create local copies for select cases
			o1, o2 := out1, out2
			for i := 0; i < 2; i++ {
				select {
				case o1 <- val:
					o1 = nil // disable after first send
				case o2 <- val:
					o2 = nil
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out1, out2
}

// =============================================================================
// PATTERN 8: Error Group pattern (simplified)
// =============================================================================
// Run multiple goroutines, collect the first error, cancel remaining.

func doWorkWithError(ctx context.Context) error {
	type result struct {
		err error
	}

	tasks := 5
	results := make(chan result, tasks)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < tasks; i++ {
		go func(id int) {
			// Simulate work that might fail
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			if id == 3 {
				results <- result{fmt.Errorf("task %d failed", id)}
				return
			}
			results <- result{nil}
		}(i)
	}

	for i := 0; i < tasks; i++ {
		r := <-results
		if r.err != nil {
			cancel() // cancel remaining work
			return r.err
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Pipeline demo
	fmt.Println("=== Pipeline: generate → multiply → filter ===")
	numbers := generateNumbers(ctx, 1, 20)
	doubled := multiply(ctx, numbers, 2)
	evens := filter(ctx, doubled, func(n int) bool { return n%4 == 0 })
	results := sink(ctx, evens)
	fmt.Printf("Results: %v\n", results)

	// Fan-out/fan-in demo
	fmt.Println("\n=== Fan-Out / Fan-In ===")
	input := generateNumbers(ctx, 1, 10)
	workers := fanOut(ctx, input, 3)
	merged := fanIn(ctx, workers...)
	for v := range merged {
		fmt.Printf("%d ", v)
	}
	fmt.Println()

	// Semaphore demo
	demonstrateSemaphore()

	// Error group demo
	fmt.Println("\n=== Error Group Pattern ===")
	if err := doWorkWithError(ctx); err != nil {
		fmt.Printf("Got expected error: %v\n", err)
	}

	fmt.Println("\n=== Advanced Channel Patterns Complete ===")
}
