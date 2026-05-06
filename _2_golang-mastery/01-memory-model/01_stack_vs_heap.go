// =============================================================================
// LESSON 1.1: STACK vs HEAP — Where Go Allocates Memory
// =============================================================================
//
// KEY CONCEPT: Go's compiler uses "escape analysis" to decide whether a variable
// lives on the stack (fast, auto-freed) or heap (slower, GC-managed).
//
// The stack is per-goroutine, starts at ~2KB–8KB, and grows dynamically.
// The heap is shared across all goroutines and managed by the garbage collector.
//
// WHY IT MATTERS: Understanding this lets you write zero-allocation hot paths,
// reduce GC pressure, and reason about performance at the hardware level.
//
// RUN WITH: go run -gcflags="-m -m" 01_stack_vs_heap.go
// The -m flag shows escape analysis decisions. Use -m -m for verbose output.
// =============================================================================

package main

import "fmt"

// -----------------------------------------------------------------------------
// CASE 1: Stack allocation — value does NOT escape
// -----------------------------------------------------------------------------
// The compiler sees that `x` never leaves this function's scope.
// It stays on the stack → zero GC pressure, instant cleanup when function returns.
func stackOnly() int {
	x := 42        // allocated on stack
	y := x * 2     // also on stack
	return y        // value is copied to caller — no escape
}

// -----------------------------------------------------------------------------
// CASE 2: Heap allocation — returning a pointer FORCES escape
// -----------------------------------------------------------------------------
// When you return a pointer, the value must survive after the function returns.
// The compiler MUST move it to the heap.
func heapEscape() *int {
	x := 42       // x escapes to heap because we return its address
	return &x      // compiler: "moved to heap: x"
}

// -----------------------------------------------------------------------------
// CASE 3: Interface boxing causes escape
// -----------------------------------------------------------------------------
// When you assign a concrete type to an interface, Go may need to allocate
// on the heap because the interface value needs a stable pointer.
func interfaceEscape() {
	x := 42
	// fmt.Println takes ...interface{} — x gets "boxed" into an interface
	// This causes x to escape to heap
	fmt.Println(x) // compiler: "x escapes to heap"
}

// -----------------------------------------------------------------------------
// CASE 4: Slice growth causes escape
// -----------------------------------------------------------------------------
// Small slices may live on stack, but if the compiler can't prove the size
// at compile time, or if the slice is too large, it escapes.
func sliceEscape() {
	// Compiler knows the size → might stay on stack (implementation-dependent)
	small := make([]int, 3)
	small[0] = 1

	// Compiler can't prove final size → escapes to heap
	dynamic := make([]int, 0)
	for i := 0; i < 10; i++ {
		dynamic = append(dynamic, i) // may trigger reallocation → heap
	}

	_ = small
	_ = dynamic
}

// -----------------------------------------------------------------------------
// CASE 5: Closure captures cause escape
// -----------------------------------------------------------------------------
// If a closure outlives the function that created the captured variable,
// that variable escapes to the heap.
func closureEscape() func() int {
	count := 0 // escapes: closure returned to caller captures this
	return func() int {
		count++
		return count
	}
}

// No escape: closure used only within the function
func closureNoEscape() int {
	x := 10
	double := func() int { return x * 2 } // x does NOT escape
	return double()
}

// -----------------------------------------------------------------------------
// CASE 6: Large objects escape
// -----------------------------------------------------------------------------
// Objects larger than a threshold (typically ~64KB) are allocated on heap
// regardless of escape analysis.
func largeAlloc() {
	// Very large array — too big for stack, goes to heap
	large := make([]byte, 1<<20) // 1MB
	large[0] = 1
	_ = large
}

// -----------------------------------------------------------------------------
// CASE 7: Preventing escape — pass pointers DOWN, not UP
// -----------------------------------------------------------------------------
// PATTERN: Instead of returning a pointer (escapes), accept a pointer parameter
// (caller owns the memory, callee just fills it in).
type Result struct {
	Value int
	Name  string
}

// BAD: Forces heap allocation
func newResult() *Result {
	return &Result{Value: 42, Name: "answer"} // escapes to heap
}

// GOOD: Caller controls allocation (can be stack-allocated)
func fillResult(r *Result) {
	r.Value = 42
	r.Name = "answer"
	// r was allocated by the caller — if caller is also on stack, no heap alloc
}

func main() {
	// Run each case
	_ = stackOnly()

	ptr := heapEscape()
	_ = *ptr

	interfaceEscape()

	sliceEscape()

	counter := closureEscape()
	_ = counter()

	_ = closureNoEscape()

	largeAlloc()

	// Compare the two patterns:
	r1 := newResult()  // heap allocation
	_ = r1

	var r2 Result       // stack allocation (if main doesn't escape)
	fillResult(&r2)     // no heap allocation for r2
	_ = r2

	fmt.Println("=== Run with: go run -gcflags='-m -m' 01_stack_vs_heap.go ===")
	fmt.Println("=== to see escape analysis decisions ===")
}
