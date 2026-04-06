// =============================================================================
// LESSON 1.3: THE Go GARBAGE COLLECTOR — How It Works Under the Hood
// =============================================================================
//
// Go uses a CONCURRENT, TRI-COLOR, MARK-AND-SWEEP collector.
// Understanding this is crucial for writing low-latency systems.
//
// KEY CONCEPTS:
//   - Tri-color marking: white (unreachable), grey (to scan), black (reachable)
//   - Write barrier: ensures GC correctness during concurrent mutation
//   - STW (Stop-The-World): Go minimizes this to <1ms typically
//   - GOGC: controls GC trigger threshold (default 100 = trigger at 2x heap)
//   - GOMEMLIMIT: soft memory limit (Go 1.19+) — better than GOGC alone
//
// RUN: GODEBUG=gctrace=1 go run 03_gc_internals.go
// =============================================================================

package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
	"unsafe"
)

// =============================================================================
// CONCEPT 1: GC Phases
// =============================================================================
//
// Phase 1: Mark Setup     (STW) — Enable write barrier (~10-30μs)
// Phase 2: Marking        (Concurrent) — Trace from roots, mark reachable objects
// Phase 3: Mark Termination (STW) — Disable write barrier, cleanup (~60-90μs)
// Phase 4: Sweeping       (Concurrent) — Reclaim unmarked (white) objects
//
// The write barrier tells the GC when pointers change during concurrent marking,
// so it doesn't miss reachable objects.

// =============================================================================
// CONCEPT 2: Reading GC Stats
// =============================================================================

func printGCStats() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	fmt.Println("=== GC Statistics ===")
	fmt.Printf("HeapAlloc    = %d MB    (current heap usage)\n", stats.HeapAlloc/1024/1024)
	fmt.Printf("HeapSys      = %d MB    (heap memory obtained from OS)\n", stats.HeapSys/1024/1024)
	fmt.Printf("HeapObjects  = %d        (number of allocated objects)\n", stats.HeapObjects)
	fmt.Printf("NumGC        = %d        (completed GC cycles)\n", stats.NumGC)
	fmt.Printf("PauseTotalNs = %d μs     (total STW pause time)\n", stats.PauseTotalNs/1000)
	fmt.Printf("GCSys        = %d KB     (GC metadata memory)\n", stats.GCSys/1024)

	// Last 5 GC pause times
	fmt.Println("\nLast GC pause times:")
	for i := 0; i < 5 && i < int(stats.NumGC); i++ {
		idx := (int(stats.NumGC) - 1 - i + 256) % 256
		fmt.Printf("  GC #%d: %d μs\n", stats.NumGC-uint32(i), stats.PauseNs[idx]/1000)
	}
}

// =============================================================================
// CONCEPT 3: GC Tuning with GOGC and GOMEMLIMIT
// =============================================================================

func demonstrateGOGC() {
	fmt.Println("\n=== GOGC Tuning ===")

	// GOGC=100 (default): GC triggers when heap doubles
	// GOGC=200: GC triggers when heap triples (less frequent GC, more memory)
	// GOGC=50: GC triggers at 1.5x heap (more frequent GC, less memory)
	// GOGC=off: Disable GC entirely (for short-lived programs)

	// Read current GOGC
	old := debug.SetGCPercent(100)
	fmt.Printf("Previous GOGC: %d\n", old)

	// Set aggressive GC for memory-constrained environments
	debug.SetGCPercent(50)
	fmt.Println("Set GOGC=50 (GC triggers at 1.5x live heap)")

	// GOMEMLIMIT (Go 1.19+): Soft memory limit
	// Better than GOGC alone — GC becomes more aggressive as you approach the limit
	debug.SetMemoryLimit(256 * 1024 * 1024) // 256 MB soft limit
	fmt.Println("Set GOMEMLIMIT=256MB")

	// Restore
	debug.SetGCPercent(int(old))
}

// =============================================================================
// CONCEPT 4: Pointer vs Non-pointer types and GC scanning cost
// =============================================================================
//
// The GC must scan every pointer in live objects to find reachable objects.
// Structs with fewer pointers → less GC work.

// BAD for GC: many pointers to scan
type NodePointerHeavy struct {
	Value    *int
	Children []*NodePointerHeavy
	Name     *string
	Metadata map[string]*string
}

// GOOD for GC: fewer pointers
type NodeValueHeavy struct {
	Value    int        // not a pointer — GC skips
	Children []int      // indices into a flat slice — not pointers
	Name     string     // string header has 1 pointer, but better than *string
	Metadata [4]int64   // fixed array, zero pointers
}

// BEST: Struct-of-Arrays pattern for GC-friendly data
type NodeSOA struct {
	Values   []int      // one slice header (1 pointer)
	Names    []string   // one slice header (contains pointers, but contiguous)
	// vs Array-of-Structs: each struct element independently scanned
}

func demonstrateGCScanCost() {
	fmt.Println("\n=== GC Scan Cost Comparison ===")
	fmt.Printf("NodePointerHeavy size: %d bytes\n", unsafe.Sizeof(NodePointerHeavy{}))
	fmt.Printf("NodeValueHeavy size:   %d bytes\n", unsafe.Sizeof(NodeValueHeavy{}))

	// Allocate many objects and measure GC pause
	const N = 100_000

	// Pointer-heavy: GC must scan every pointer
	start := time.Now()
	nodes := make([]*NodePointerHeavy, N)
	for i := range nodes {
		v := i
		s := "name"
		nodes[i] = &NodePointerHeavy{Value: &v, Name: &s}
	}
	runtime.GC()
	fmt.Printf("Pointer-heavy GC time: %v\n", time.Since(start))

	// Value-heavy: GC skips non-pointer fields
	start = time.Now()
	vnodes := make([]NodeValueHeavy, N) // flat slice, no individual allocs
	for i := range vnodes {
		vnodes[i] = NodeValueHeavy{Value: i, Name: "name"}
	}
	runtime.GC()
	fmt.Printf("Value-heavy GC time:   %v\n", time.Since(start))

	_ = nodes
	_ = vnodes
}

// =============================================================================
// CONCEPT 5: Finalizers — Last resort cleanup
// =============================================================================
//
// runtime.SetFinalizer registers a function to run when an object is about to
// be collected. USE SPARINGLY — they add GC overhead and are not guaranteed
// to run in any particular order or time.

type ExpensiveResource struct {
	ID   int
	data []byte
}

func NewExpensiveResource(id int) *ExpensiveResource {
	r := &ExpensiveResource{
		ID:   id,
		data: make([]byte, 1024*1024), // 1MB
	}

	// Register finalizer — runs just before GC collects this object
	runtime.SetFinalizer(r, func(r *ExpensiveResource) {
		fmt.Printf("Finalizer: cleaning up resource %d\n", r.ID)
		// Release external resources here (file handles, C memory, etc.)
	})

	return r
}

// =============================================================================
// CONCEPT 6: runtime.KeepAlive — Prevent premature GC
// =============================================================================
//
// When working with unsafe.Pointer or CGo, the GC might collect an object
// before you're done using it (if there are no more Go references).
// runtime.KeepAlive forces the object to stay alive until that point.

func demonstrateKeepAlive() {
	data := make([]byte, 1024)
	// Imagine passing &data[0] to C code via unsafe.Pointer
	// ptr := unsafe.Pointer(&data[0])
	// C.process(ptr)  // data could be GC'd during this call!

	// Solution: keep data alive until after the C call completes
	runtime.KeepAlive(data)
}

func main() {
	printGCStats()
	demonstrateGOGC()
	demonstrateGCScanCost()

	r := NewExpensiveResource(1)
	_ = r
	r = nil // make eligible for GC
	runtime.GC() // force GC to demonstrate finalizer

	demonstrateKeepAlive()

	printGCStats()

	fmt.Println("\n=== Run with: GODEBUG=gctrace=1 go run 03_gc_internals.go ===")
	fmt.Println("=== gctrace output format: ===")
	fmt.Println("gc # @#s #%: #ms+#ms+#ms ms clock, #ms+#ms/#ms/#ms+#ms ms cpu, #->#-># MB, # MB goal, # P")
}
