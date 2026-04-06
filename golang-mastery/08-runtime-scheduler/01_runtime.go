// =============================================================================
// LESSON 8: Go RUNTIME & SCHEDULER INTERNALS
// =============================================================================
//
// Understanding the runtime is what separates good Go devs from great ones.
//
// THE GMP MODEL:
//   G = Goroutine (lightweight thread, ~2-8KB stack)
//   M = Machine  (OS thread)
//   P = Processor (logical CPU, holds local run queue)
//
// SCHEDULING:
//   - P count = GOMAXPROCS (default = NumCPU)
//   - Each P has a local run queue (LRQ, 256 goroutines max)
//   - There's also a global run queue (GRQ) when LRQ is full
//   - Work stealing: idle P steals G from other P's LRQ
//   - Preemption: since Go 1.14, goroutines are preempted asynchronously
//     via signals (SIGURG) — no cooperative preemption needed
//
// GOROUTINE STATES:
//   Runnable → Running → Waiting (I/O, channel, mutex) → Runnable
//   Also: Dead (finished), Syscall (blocking OS call)
// =============================================================================

package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// =============================================================================
// PART 1: Runtime Information
// =============================================================================

func printRuntimeInfo() {
	fmt.Println("=== Runtime Information ===")
	fmt.Printf("Go Version:     %s\n", runtime.Version())
	fmt.Printf("GOOS:           %s\n", runtime.GOOS)
	fmt.Printf("GOARCH:         %s\n", runtime.GOARCH)
	fmt.Printf("NumCPU:         %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS:     %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumGoroutine:   %d\n", runtime.NumGoroutine())

	// GC Info
	var gcStats debug.GCStats
	debug.ReadGCStats(&gcStats)
	fmt.Printf("GC Pauses:      %d\n", gcStats.NumGC)
	if len(gcStats.Pause) > 0 {
		fmt.Printf("Last GC Pause:  %v\n", gcStats.Pause[0])
	}
}

// =============================================================================
// PART 2: GOMAXPROCS tuning
// =============================================================================
//
// GOMAXPROCS sets the max number of OS threads running Go code simultaneously.
// Default: runtime.NumCPU()
//
// TUNING:
// - CPU-bound work: GOMAXPROCS = NumCPU (default is optimal)
// - I/O-bound work: Can benefit from GOMAXPROCS > NumCPU
// - Containers: Go auto-detects CPU quota since Go 1.19 (GOEXPERIMENT=containerd)
//   but older versions may see all host CPUs → use uber-go/automaxprocs

func demonstrateGOMAXPROCS() {
	fmt.Println("\n=== GOMAXPROCS Tuning ===")

	prev := runtime.GOMAXPROCS(1) // single thread
	fmt.Printf("Set GOMAXPROCS=1 (was %d)\n", prev)

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sum := 0
			for j := 0; j < 10_000_000; j++ {
				sum += j
			}
			_ = sum
		}(i)
	}
	wg.Wait()
	fmt.Printf("  GOMAXPROCS=1: %v\n", time.Since(start))

	runtime.GOMAXPROCS(runtime.NumCPU()) // restore
	start = time.Now()
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sum := 0
			for j := 0; j < 10_000_000; j++ {
				sum += j
			}
			_ = sum
		}(i)
	}
	wg.Wait()
	fmt.Printf("  GOMAXPROCS=%d: %v\n", runtime.NumCPU(), time.Since(start))
}

// =============================================================================
// PART 3: Goroutine Lifecycle & Stack Growth
// =============================================================================
//
// Goroutine stacks start small (~2-8KB) and grow dynamically up to 1GB.
// When a function needs more stack, Go allocates a new, larger stack and
// copies the old stack contents (including updating all pointers).
// This is called "stack copying" or "contiguous stacks" (since Go 1.4).
//
// IMPLICATION: Pointers to stack variables remain valid because Go
// updates them during stack copying. But uintptr values are NOT updated!

func demonstrateStackGrowth() {
	fmt.Println("\n=== Stack Growth ===")

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Launch goroutines that use significant stack space
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deepRecursion(50) // each goroutine grows its stack
		}()
	}
	wg.Wait()

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	fmt.Printf("Stack in use before: %d KB\n", memBefore.StackInuse/1024)
	fmt.Printf("Stack in use after:  %d KB\n", memAfter.StackInuse/1024)
	fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
}

func deepRecursion(depth int) int {
	if depth == 0 {
		return 1
	}
	// Local variable forces stack frame allocation
	var waste [64]byte
	waste[0] = byte(depth)
	_ = waste
	return deepRecursion(depth - 1)
}

// =============================================================================
// PART 4: runtime.Gosched, runtime.Goexit, runtime.LockOSThread
// =============================================================================

func demonstrateRuntimeFuncs() {
	fmt.Println("\n=== Runtime Functions ===")

	// runtime.Gosched — yield the processor
	// Lets other goroutines run. Rarely needed since Go 1.14 (async preemption).
	go func() {
		for i := 0; i < 3; i++ {
			fmt.Printf("  Goroutine: iteration %d\n", i)
			runtime.Gosched() // yield — let other goroutines run
		}
	}()

	// runtime.LockOSThread — pin goroutine to OS thread
	// Required for:
	//   - OpenGL/GPU calls (must be on main thread)
	//   - Some C libraries (thread-local storage)
	//   - Linux namespaces (setns/unshare)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		// This goroutine now runs on a dedicated OS thread
		fmt.Println("  Locked to OS thread for thread-sensitive operations")
	}()

	time.Sleep(100 * time.Millisecond)

	// runtime.Goexit — exit the goroutine, running deferred functions
	// Unlike os.Exit, this only kills the current goroutine and runs defers.
	done := make(chan bool)
	go func() {
		defer func() {
			fmt.Println("  Deferred function ran after Goexit!")
			done <- true
		}()
		fmt.Println("  About to Goexit...")
		runtime.Goexit()
		fmt.Println("  This will never execute")
	}()
	<-done
}

// =============================================================================
// PART 5: Stack Traces & Goroutine Dumps
// =============================================================================

func demonstrateStackTrace() {
	fmt.Println("\n=== Stack Traces ===")

	// Get current goroutine's stack
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false) // false = current goroutine only
	fmt.Printf("Current goroutine stack:\n%s\n", buf[:n])

	// Get all goroutine stacks (useful for debugging deadlocks)
	// buf2 := make([]byte, 1<<20) // 1MB
	// n2 := runtime.Stack(buf2, true) // true = all goroutines
	// fmt.Printf("All goroutine stacks:\n%s\n", buf2[:n2])

	// In production, send SIGQUIT to dump all goroutine stacks:
	// kill -QUIT <pid>
	// Or set GOTRACEBACK=all environment variable
}

// =============================================================================
// PART 6: Goroutine Leak Detection
// =============================================================================

func demonstrateLeakDetection() {
	fmt.Println("\n=== Goroutine Leak Detection ===")

	before := runtime.NumGoroutine()

	// LEAK: This goroutine blocks forever because nobody reads from ch
	ch := make(chan int)
	go func() {
		ch <- 42 // blocks forever — no reader!
	}()

	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()

	fmt.Printf("Goroutines before: %d, after: %d\n", before, after)
	if after > before {
		fmt.Printf("  WARNING: %d goroutine(s) leaked!\n", after-before)
	}

	// FIX: Always ensure goroutines can exit
	// Use context cancellation, buffered channels, or select with done channel

	// PREVENTION: In tests, use goleak:
	//   import "go.uber.org/goleak"
	//   func TestMain(m *testing.M) { goleak.VerifyTestMain(m) }

	// Drain the leak by reading
	go func() { <-ch }()
}

// =============================================================================
// PART 7: Memory Ballast — Reducing GC frequency
// =============================================================================
//
// TECHNIQUE: Allocate a large unused byte slice to increase perceived "live heap"
// This makes GC trigger less frequently (GOGC is based on live heap ratio).
// Largely superseded by GOMEMLIMIT in Go 1.19+, but still used in production.

func demonstrateBallast() {
	fmt.Println("\n=== Memory Ballast ===")

	var stats runtime.MemStats

	// Without ballast
	runtime.GC()
	runtime.ReadMemStats(&stats)
	numGCBefore := stats.NumGC

	// Simulate workload
	for i := 0; i < 100; i++ {
		data := make([]byte, 1<<16) // 64KB allocations
		_ = data
	}
	runtime.ReadMemStats(&stats)
	fmt.Printf("Without ballast: %d GCs\n", stats.NumGC-numGCBefore)

	// With ballast (keep a reference so it's not collected)
	ballast := make([]byte, 100<<20) // 100MB ballast
	_ = ballast[0]                   // touch to ensure it's not optimized away

	numGCBefore = stats.NumGC
	for i := 0; i < 100; i++ {
		data := make([]byte, 1<<16)
		_ = data
	}
	runtime.ReadMemStats(&stats)
	fmt.Printf("With 100MB ballast: %d GCs\n", stats.NumGC-numGCBefore)

	runtime.KeepAlive(ballast)
}

func main() {
	printRuntimeInfo()
	demonstrateGOMAXPROCS()
	demonstrateStackGrowth()
	demonstrateRuntimeFuncs()
	demonstrateStackTrace()
	demonstrateLeakDetection()
	demonstrateBallast()

	fmt.Println("\n=== KEY INSIGHTS ===")
	fmt.Println("1. GMP model: G(goroutines) are multiplexed onto M(threads) via P(processors)")
	fmt.Println("2. Work stealing ensures load balancing across Ps")
	fmt.Println("3. Goroutine stacks start small and grow dynamically")
	fmt.Println("4. runtime.LockOSThread is needed for thread-affinity (OpenGL, namespaces)")
	fmt.Println("5. Always check for goroutine leaks in production")
	fmt.Println("6. GOMEMLIMIT (Go 1.19+) is better than memory ballast for GC tuning")
}
