// =============================================================================
// LESSON 1.2: ESCAPE ANALYSIS DEEP DIVE & COMPILER OPTIMIZATIONS
// =============================================================================
//
// This file teaches you HOW to read escape analysis output, common traps,
// and techniques to write allocation-free code in hot paths.
//
// RUN: go run -gcflags="-m -l" 02_escape_analysis.go
//    -m: show escape decisions
//    -l: disable inlining (so we can see true escape behavior)
//
// BENCHMARK: go test -bench=. -benchmem -count=5
// =============================================================================

package main

import (
	"encoding/binary"
	"fmt"
	"sync"
)

// =============================================================================
// TECHNIQUE 1: sync.Pool — Reuse heap objects to reduce GC pressure
// =============================================================================
//
// sync.Pool is a concurrent-safe free list. Objects may be collected by GC
// between uses, so never assume the pool has items.

type Buffer struct {
	data []byte
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{data: make([]byte, 0, 4096)}
	},
}

func processWithPool() {
	// Get from pool (or allocate new)
	buf := bufferPool.Get().(*Buffer)

	// Reset without reallocating the underlying array
	buf.data = buf.data[:0]

	// Use the buffer
	buf.data = append(buf.data, "hello world"...)

	// Return to pool for reuse
	bufferPool.Put(buf)
}

// Without pool — every call allocates
func processWithoutPool() {
	buf := &Buffer{data: make([]byte, 0, 4096)} // heap alloc every time
	buf.data = append(buf.data, "hello world"...)
	_ = buf
}

// =============================================================================
// TECHNIQUE 2: Avoid interface{} in hot paths
// =============================================================================
//
// Any time a value is assigned to an interface, it may cause a heap allocation
// (called "boxing"). In hot paths, use concrete types.

// BAD: interface parameter causes boxing
func sumInterface(values ...interface{}) int64 {
	var total int64
	for _, v := range values {
		total += v.(int64) // type assertion cost + boxing cost
	}
	return total
}

// GOOD: concrete type, zero allocations
func sumInt64(values ...int64) int64 {
	var total int64
	for _, v := range values {
		total += v
	}
	return total
}

// =============================================================================
// TECHNIQUE 3: Pre-sized slices and maps
// =============================================================================

// BAD: Unknown size → multiple reallocations
func buildSliceBad(n int) []int {
	var s []int // nil slice, cap=0
	for i := 0; i < n; i++ {
		s = append(s, i) // may reallocate and copy on each grow
	}
	return s
}

// GOOD: Pre-allocate with known capacity
func buildSliceGood(n int) []int {
	s := make([]int, 0, n) // one allocation, exact size
	for i := 0; i < n; i++ {
		s = append(s, i) // never reallocates
	}
	return s
}

// Maps: pre-size to avoid rehashing
func buildMapGood(n int) map[string]int {
	m := make(map[string]int, n) // pre-sized
	for i := 0; i < n; i++ {
		m[fmt.Sprintf("key%d", i)] = i
	}
	return m
}

// =============================================================================
// TECHNIQUE 4: Stack-allocated arrays instead of slices for fixed sizes
// =============================================================================

// Uses a slice → heap allocation
func encodeUint64Slice(v uint64) []byte {
	buf := make([]byte, 8) // escapes to heap
	binary.BigEndian.PutUint64(buf, v)
	return buf
}

// Uses array → stays on stack (if not returned as pointer)
func encodeUint64Array(v uint64) [8]byte {
	var buf [8]byte // stack allocated — array value type
	binary.BigEndian.PutUint64(buf[:], v)
	return buf // copied to caller, still no heap
}

// =============================================================================
// TECHNIQUE 5: String interning & []byte<->string zero-copy
// =============================================================================
//
// String to []byte conversion normally allocates because strings are immutable
// and []byte is mutable. But there are safe patterns to avoid this.

// The compiler optimizes this specific pattern — no allocation:
func lookupInMap(m map[string]int, key []byte) int {
	// Since Go 1.3, map lookup with string(key) where key is []byte
	// does NOT allocate if the result is only used for the lookup
	return m[string(key)] // zero-alloc map lookup!
}

// But this DOES allocate:
func lookupInMapBad(m map[string]int, key []byte) int {
	s := string(key)  // allocates: s escapes
	return m[s]
}

// =============================================================================
// TECHNIQUE 6: Inlining budget — keep functions small for the optimizer
// =============================================================================
//
// The Go compiler inlines small functions (cost < 80 AST nodes).
// Inlined functions can then have their variables participate in the CALLER's
// escape analysis, often avoiding allocations.
//
// Check inlining: go build -gcflags="-m" shows "can inline" / "inlining call"

// This function is small enough to inline:
//
//go:nosplit is NOT needed — the compiler handles this automatically
func add(a, b int) int {
	return a + b // cost ~1, easily inlined
}

// This is too complex to inline (many branches, loops):
func complexFunction(data []int) int {
	result := 0
	for i := 0; i < len(data); i++ {
		switch {
		case data[i] > 100:
			result += data[i] * 3
		case data[i] > 50:
			result += data[i] * 2
		default:
			result += data[i]
		}
		if result > 10000 {
			for j := 0; j < len(data); j++ {
				result -= data[j] / 2
			}
		}
	}
	return result
}

func main() {
	processWithPool()
	processWithoutPool()

	_ = sumInterface(int64(1), int64(2), int64(3))
	_ = sumInt64(1, 2, 3)

	_ = buildSliceBad(100)
	_ = buildSliceGood(100)

	_ = encodeUint64Slice(42)
	_ = encodeUint64Array(42)

	_ = add(1, 2)
	_ = complexFunction([]int{1, 2, 3})

	fmt.Println("=== Lesson 1.2: Escape Analysis Deep Dive ===")
	fmt.Println("Run: go run -gcflags='-m' 02_escape_analysis.go")
	fmt.Println("Compare BAD vs GOOD patterns with benchmarks")
}
