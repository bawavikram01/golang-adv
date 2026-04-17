//go:build ignore

// =============================================================================
// LESSON 0.8: GOROUTINES & CHANNELS — Concurrency Foundations
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Goroutines: lightweight concurrent functions
// - Channels: typed conduits for communication
// - Unbuffered vs buffered channels
// - select for multiplexing
// - sync.WaitGroup for coordination
// - Common patterns: done channel, fan-out, pipeline
// - Goroutine leaks and how to avoid them
//
// THE KEY INSIGHT:
// "Don't communicate by sharing memory; share memory by communicating."
// Go's concurrency model is based on Communicating Sequential Processes (CSP).
// Goroutines are the units of execution, channels are the communication pipes.
//
// NOTE: This covers FUNDAMENTALS. Advanced concurrency (mutexes, atomics,
// advanced patterns, context deep-dive) is in 02-advanced-concurrency/.
//
// RUN: go run 08_goroutines_channels.go
// =============================================================================

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== GOROUTINES & CHANNELS ===")
	fmt.Println()

	goroutineBasics()
	channelBasics()
	bufferedChannels()
	channelDirections()
	rangeOverChannels()
	selectStatement()
	waitGroupPattern()
	doneChannelPattern()
	pipelinePattern()
	goroutineLeaks()
	commonMistakes()
}

// =============================================================================
// PART 1: Goroutine Basics
// =============================================================================
func goroutineBasics() {
	fmt.Println("--- GOROUTINE BASICS ---")

	// A goroutine is a lightweight thread managed by the Go runtime.
	// Start one with the `go` keyword.
	// Cost: ~2KB initial stack (vs ~1MB for OS thread)
	// You can easily run 100,000+ goroutines.

	// ─── Basic goroutine ───
	go func() {
		// This runs concurrently
		fmt.Println("  Hello from goroutine!")
	}()

	// ─── GOTCHA: main doesn't wait for goroutines! ───
	// If main() returns, ALL goroutines are killed immediately.
	// We need synchronization (channels, WaitGroup, etc.)
	time.Sleep(10 * time.Millisecond) // crude: don't do this in production

	// ─── Goroutine with arguments ───
	// IMPORTANT: Pass loop variables as arguments, don't capture them!
	for i := 0; i < 3; i++ {
		go func(n int) { // n is a COPY of i
			fmt.Printf("  Goroutine %d\n", n)
		}(i)
	}
	time.Sleep(10 * time.Millisecond)

	// NOTE: Since Go 1.22, loop variables are per-iteration,
	// so the closure capture bug is fixed for `for i := range ...`

	// ─── How goroutines work internally ───
	// - Go runtime has its own scheduler (M:N scheduling)
	// - Goroutines (G) are multiplexed onto OS threads (M)
	// - Scheduler is cooperative + preemptive (since Go 1.14)
	// - When a goroutine blocks (I/O, channel, etc.), the runtime
	//   parks it and schedules another goroutine on the same thread

	fmt.Println()
}

// =============================================================================
// PART 2: Channel Basics
// =============================================================================
func channelBasics() {
	fmt.Println("--- CHANNEL BASICS ---")

	// A channel is a typed conduit for passing values between goroutines.
	// Think of it as a pipe: one goroutine sends, another receives.

	// ─── Create with make ───
	ch := make(chan string) // unbuffered channel of strings

	// ─── Send and receive ───
	go func() {
		ch <- "hello" // send (blocks until someone receives)
	}()

	msg := <-ch // receive (blocks until someone sends)
	fmt.Printf("  Received: %s\n", msg)

	// ─── UNBUFFERED channels: synchronization point ───
	// sender blocks until receiver is ready
	// receiver blocks until sender sends
	// They synchronize: both goroutines meet at the channel

	// ─── Multiple values ───
	ch2 := make(chan int)
	go func() {
		for i := 0; i < 5; i++ {
			ch2 <- i * i
		}
		close(ch2) // signal: no more values
	}()

	// Receive until channel closed
	for val := range ch2 {
		fmt.Printf("  Squared: %d\n", val)
	}

	// ─── Channel zero value: nil ───
	var nilCh chan int
	fmt.Printf("  nil channel: %v\n", nilCh)
	// Sending to nil channel blocks forever
	// Receiving from nil channel blocks forever
	// This is useful with select (disable a case by setting channel to nil)

	fmt.Println()
}

// =============================================================================
// PART 3: Buffered Channels
// =============================================================================
func bufferedChannels() {
	fmt.Println("--- BUFFERED CHANNELS ---")

	// Buffered channel: has internal capacity
	// Sender blocks ONLY when buffer is full
	// Receiver blocks ONLY when buffer is empty
	ch := make(chan int, 3) // buffer of 3

	// Can send 3 values without blocking
	ch <- 1
	ch <- 2
	ch <- 3
	// ch <- 4  // would block! buffer is full

	fmt.Printf("  len=%d, cap=%d\n", len(ch), cap(ch))
	// len = number of elements in buffer
	// cap = buffer capacity

	fmt.Printf("  %d, %d, %d\n", <-ch, <-ch, <-ch)

	// ─── When to use buffered channels ───
	// Unbuffered (make(chan T)):
	//   - Synchronization: sender and receiver meet
	//   - Guaranteed delivery before sender continues
	//
	// Buffered (make(chan T, n)):
	//   - Decouple sender/receiver timing
	//   - Known number of goroutines (e.g., make(chan result, n) for n workers)
	//   - Rate limiting / backpressure
	//
	// Default to unbuffered. Use buffered only when you know why.

	// ─── Semaphore pattern with buffered channel ───
	sem := make(chan struct{}, 3) // allow max 3 concurrent goroutines
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{} // acquire (blocks if 3 already running)
			defer func() { <-sem }()

			fmt.Printf("  Worker %d running\n", id)
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		}(i)
	}
	wg.Wait()

	fmt.Println()
}

// =============================================================================
// PART 4: Channel Directions
// =============================================================================

// send-only channel: can only send, not receive
func produce(ch chan<- int) {
	for i := 0; i < 5; i++ {
		ch <- i
	}
	close(ch)
}

// receive-only channel: can only receive, not send
func consume(ch <-chan int) []int {
	var result []int
	for val := range ch {
		result = append(result, val)
	}
	return result
}

func channelDirections() {
	fmt.Println("--- CHANNEL DIRECTIONS ---")

	// Bidirectional channels can be assigned to directional vars:
	ch := make(chan int) // bidirectional

	go produce(ch)        // auto-converts to chan<- int
	result := consume(ch) // auto-converts to <-chan int
	fmt.Printf("  Produced & consumed: %v\n", result)

	// Direction annotations:
	// chan T    : bidirectional (send and receive)
	// chan<- T  : send-only
	// <-chan T  : receive-only
	//
	// Use directional channels in function signatures to enforce
	// that a function only sends OR receives. Compiler catches mistakes.

	fmt.Println()
}

// =============================================================================
// PART 5: Range Over Channels
// =============================================================================
func rangeOverChannels() {
	fmt.Println("--- RANGE OVER CHANNELS ---")

	ch := make(chan string)

	go func() {
		words := []string{"Go", "is", "awesome"}
		for _, w := range words {
			ch <- w
		}
		close(ch) // MUST close or range blocks forever!
	}()

	// range receives values until channel is closed
	var parts []string
	for word := range ch {
		parts = append(parts, word)
	}
	fmt.Printf("  %v\n", parts)

	// ─── Closed channel behavior ───
	ch2 := make(chan int, 2)
	ch2 <- 42
	close(ch2)

	v1, ok1 := <-ch2 // 42, true (value available)
	v2, ok2 := <-ch2 // 0, false (closed, no more values)
	fmt.Printf("  After close: (%d, %v), (%d, %v)\n", v1, ok1, v2, ok2)

	// RULES:
	// - Sending to closed channel: PANIC
	// - Receiving from closed channel: returns zero value immediately
	// - close() should be called by the SENDER, never the receiver
	// - Only close when the receiver needs to know no more data

	fmt.Println()
}

// =============================================================================
// PART 6: select Statement
// =============================================================================
func selectStatement() {
	fmt.Println("--- SELECT STATEMENT ---")

	// select is like switch but for channel operations.
	// It waits on multiple channels simultaneously.

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(5 * time.Millisecond)
		ch1 <- "from ch1"
	}()
	go func() {
		time.Sleep(2 * time.Millisecond)
		ch2 <- "from ch2"
	}()

	// Receives from whichever is ready first
	select {
	case msg := <-ch1:
		fmt.Printf("  %s\n", msg)
	case msg := <-ch2:
		fmt.Printf("  %s\n", msg) // ch2 is faster
	}
	// Drain the other
	<-ch1

	// ─── select with timeout ───
	ch3 := make(chan int)
	select {
	case val := <-ch3:
		fmt.Printf("  Got: %d\n", val)
	case <-time.After(10 * time.Millisecond):
		fmt.Println("  Timeout!")
	}

	// ─── select with default (non-blocking) ───
	ch4 := make(chan int, 1)
	select {
	case val := <-ch4:
		fmt.Printf("  Got: %d\n", val)
	default:
		fmt.Println("  No value ready (non-blocking)")
	}

	// ─── select behavior ───
	// - If multiple cases ready: picks one RANDOMLY
	// - If no case ready + default: runs default (non-blocking)
	// - If no case ready + no default: BLOCKS until one is ready
	// - nil channel case: never selected (useful to disable cases)

	fmt.Println()
}

// =============================================================================
// PART 7: sync.WaitGroup — Waiting for Goroutines
// =============================================================================
func waitGroupPattern() {
	fmt.Println("--- WAITGROUP ---")

	// WaitGroup waits for a collection of goroutines to finish.
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1) // increment counter BEFORE starting goroutine
		go func(id int) {
			defer wg.Done() // decrement counter when done
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			fmt.Printf("  Worker %d done\n", id)
		}(i)
	}

	wg.Wait() // block until counter reaches 0
	fmt.Println("  All workers done")

	// ─── GOTCHAS ───
	// 1. wg.Add() MUST be called before go func(), not inside
	//    (race: main might reach Wait before goroutine calls Add)
	// 2. Don't copy WaitGroup after first use (pass pointer)
	// 3. Counter must not go negative (panic)

	// ─── WaitGroup with error collection ───
	var mu sync.Mutex
	var errors []error

	var wg2 sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			if id == 1 {
				mu.Lock()
				errors = append(errors, fmt.Errorf("worker %d failed", id))
				mu.Unlock()
			}
		}(i)
	}
	wg2.Wait()
	fmt.Printf("  Errors: %v\n", errors)

	fmt.Println()
}

// =============================================================================
// PART 8: Done Channel Pattern
// =============================================================================
func doneChannelPattern() {
	fmt.Println("--- DONE CHANNEL PATTERN ---")

	// Use a channel to signal goroutines to stop.
	// This is the precursor to context.Context.

	done := make(chan struct{}) // empty struct: zero memory

	go func() {
		for {
			select {
			case <-done:
				fmt.Println("  Worker: received stop signal")
				return
			default:
				// do work
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	time.Sleep(20 * time.Millisecond)
	close(done) // broadcast: all receivers get the signal
	time.Sleep(10 * time.Millisecond)

	// Why close(done) instead of done <- struct{}{}?
	// close() broadcasts to ALL receivers.
	// Sending only unblocks ONE receiver.

	// In production: use context.Context instead of done channels.
	// ctx, cancel := context.WithCancel(context.Background())
	// cancel() // same as close(done)
	// <-ctx.Done() // same as <-done

	fmt.Println()
}

// =============================================================================
// PART 9: Pipeline Pattern
// =============================================================================
func pipelinePattern() {
	fmt.Println("--- PIPELINE PATTERN ---")

	// A pipeline is a series of stages connected by channels.
	// Each stage: receives from upstream → processes → sends downstream

	// Stage 1: Generate numbers
	gen := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			for _, n := range nums {
				out <- n
			}
			close(out)
		}()
		return out
	}

	// Stage 2: Square each number
	square := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				out <- n * n
			}
			close(out)
		}()
		return out
	}

	// Stage 3: Filter even numbers
	filterEven := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				if n%2 == 0 {
					out <- n
				}
			}
			close(out)
		}()
		return out
	}

	// Connect the pipeline
	numbers := gen(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	squared := square(numbers)
	evens := filterEven(squared)

	fmt.Print("  Even squares: ")
	for n := range evens {
		fmt.Printf("%d ", n)
	}
	fmt.Println()

	// Pipeline: gen → square → filterEven → consumer
	// Each stage runs in its own goroutine.
	// Data flows through channels.
	// Stages are independent and composable.

	fmt.Println()
}

// =============================================================================
// PART 10: Goroutine Leaks — The #1 Concurrency Bug
// =============================================================================
func goroutineLeaks() {
	fmt.Println("--- GOROUTINE LEAKS ---")

	// A goroutine leak happens when a goroutine never terminates.
	// It stays in memory forever, consuming resources.

	// ─── LEAK: no one reads from channel ───
	// func leaky() {
	//     ch := make(chan int)
	//     go func() {
	//         ch <- 42  // blocks forever: no receiver!
	//     }()
	//     // ch goes out of scope, but goroutine is still blocked
	// }

	// ─── FIX: use buffered channel or ensure receiver exists ───
	noLeak := func() int {
		ch := make(chan int, 1) // buffered: send won't block
		go func() {
			ch <- 42 // doesn't block even without receiver
		}()
		return <-ch
	}
	fmt.Printf("  No leak: %d\n", noLeak())

	// ─── LEAK: infinite loop with no exit ───
	// func leaky2() {
	//     go func() {
	//         for { }  // never stops → goroutine lives forever
	//     }()
	// }

	// ─── FIX: always provide a way to stop ───
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return // clean exit
			default:
				// work
			}
		}
	}()
	close(done) // stop the goroutine

	// RULES TO PREVENT LEAKS:
	// 1. Every goroutine must have a way to terminate
	// 2. The creator is responsible for stopping its goroutines
	// 3. Use context.Context for cancellation
	// 4. Always close channels from the sender side
	// 5. Use go vet and golangci-lint to detect leaks
	// 6. Test with goleak: go.uber.org/goleak

	fmt.Println()
}

// =============================================================================
// PART 11: Common Mistakes
// =============================================================================
func commonMistakes() {
	fmt.Println("--- COMMON MISTAKES ---")

	// ─── MISTAKE 1: Sending to closed channel ───
	// ch := make(chan int)
	// close(ch)
	// ch <- 1  // PANIC: send on closed channel
	fmt.Println("  1. Never send to a closed channel (panic)")

	// ─── MISTAKE 2: Closing a channel twice ───
	// ch := make(chan int)
	// close(ch)
	// close(ch)  // PANIC: close of closed channel
	fmt.Println("  2. Never close a channel twice (panic)")

	// ─── MISTAKE 3: Data race ───
	// var counter int
	// for i := 0; i < 1000; i++ {
	//     go func() { counter++ }()  // RACE: concurrent write
	// }
	// Fix: use sync.Mutex, sync/atomic, or channels
	fmt.Println("  3. Use mutex/atomic/channels for shared state")

	// ─── MISTAKE 4: Using time.Sleep for synchronization ───
	// go doWork()
	// time.Sleep(1 * time.Second)  // "should be enough"... NO!
	// Fix: use WaitGroup, channels, or context
	fmt.Println("  4. Never use time.Sleep for synchronization")

	// ─── MISTAKE 5: Spawning goroutines without limits ───
	// for _, url := range urls {
	//     go fetch(url)  // 1 million URLs = 1 million goroutines
	// }
	// Fix: use semaphore (buffered channel) or worker pool
	fmt.Println("  5. Limit concurrent goroutines (worker pool / semaphore)")

	// Detect races: go run -race myprogram.go
	fmt.Println("  Always test with: go run -race")

	fmt.Println()
}
