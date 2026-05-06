//go:build ignore

// =============================================================================
// LESSON 0.4: ARRAYS, SLICES & MAPS — Go's Data Structures
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Arrays: fixed size, value type (rarely used directly)
// - Slices: the workhorse — header, length, capacity, growth strategy
// - Slice internals: how append works, when it allocates
// - Slice gotchas: shared backing arrays, memory leaks
// - Maps: hash tables, iteration order, concurrent access
// - Map internals: bucket structure, load factor, evacuation
// - make vs new vs literals
//
// THE KEY INSIGHT:
// A slice is a (pointer, length, capacity) header over an underlying array.
// Multiple slices can share the same array. This is powerful but dangerous —
// modifying one slice can affect another. Understanding this model prevents
// the #1 source of subtle Go bugs.
//
// RUN: go run 04_arrays_slices_maps.go
// =============================================================================

package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	fmt.Println("=== ARRAYS, SLICES & MAPS ===")
	fmt.Println()

	arrayFundamentals()
	sliceInternals()
	sliceOperations()
	sliceGotchas()
	mapFundamentals()
	mapPatterns()
	makeVsNew()
}

// =============================================================================
// PART 1: Arrays — Fixed Size, Value Type
// =============================================================================
func arrayFundamentals() {
	fmt.Println("--- ARRAYS ---")

	// Arrays have FIXED size. Size is part of the type!
	// [3]int and [4]int are DIFFERENT types.

	// ─── Declaration ───
	var a [5]int // zero-valued: [0 0 0 0 0]
	a[0] = 10
	a[4] = 50
	fmt.Printf("  var [5]int: %v\n", a)

	// ─── Literal ───
	b := [3]string{"Go", "Rust", "Python"}
	fmt.Printf("  literal: %v\n", b)

	// ─── Ellipsis: compiler counts ───
	c := [...]int{10, 20, 30, 40, 50} // compiler infers [5]int
	fmt.Printf("  [...]: %v (len=%d)\n", c, len(c))

	// ─── Indexed initialization ───
	d := [5]int{1: 10, 3: 30} // only set index 1 and 3
	fmt.Printf("  indexed: %v\n", d)

	// ─── Arrays are VALUE TYPES ───
	// Assigning or passing an array COPIES all elements!
	original := [3]int{1, 2, 3}
	copied := original
	copied[0] = 99
	fmt.Printf("  Value type: original=%v, copy=%v (independent!)\n", original, copied)

	// ─── Arrays are comparable ───
	x := [3]int{1, 2, 3}
	y := [3]int{1, 2, 3}
	fmt.Printf("  Comparable: %v == %v → %v\n", x, y, x == y)

	// WHY ARRAYS ARE RARELY USED:
	// - Fixed size at compile time (not flexible)
	// - Passed by value (copying large arrays is expensive)
	// - Use SLICES instead (99% of the time)
	// Arrays are the BACKING STORE for slices.

	fmt.Println()
}

// =============================================================================
// PART 2: Slice Internals — The Slice Header
// =============================================================================
func sliceInternals() {
	fmt.Println("--- SLICE INTERNALS ---")

	// A slice is a DESCRIPTOR (header) with three fields:
	//
	// type slice struct {
	//     ptr *T  // pointer to the first element in the backing array
	//     len int // number of elements currently in the slice
	//     cap int // capacity: elements from ptr to end of backing array
	// }
	//
	// Size of slice header: 24 bytes on 64-bit (3 × 8 bytes)
	// This is what gets copied when you pass a slice to a function.
	// The BACKING ARRAY is NOT copied — multiple slices can share it.

	// ─── Creating slices ───

	// 1. Literal
	s1 := []int{1, 2, 3, 4, 5}
	fmt.Printf("  Literal: %v (len=%d, cap=%d)\n", s1, len(s1), cap(s1))

	// 2. make(type, length, capacity)
	s2 := make([]int, 3, 10) // len=3, cap=10
	fmt.Printf("  make(3,10): %v (len=%d, cap=%d)\n", s2, len(s2), cap(s2))

	// 3. make(type, length) — capacity = length
	s3 := make([]int, 5) // len=5, cap=5
	fmt.Printf("  make(5): %v (len=%d, cap=%d)\n", s3, len(s3), cap(s3))

	// 4. Slicing an array or slice
	arr := [5]int{10, 20, 30, 40, 50}
	s4 := arr[1:4] // elements at index 1, 2, 3
	fmt.Printf("  arr[1:4]: %v (len=%d, cap=%d)\n", s4, len(s4), cap(s4))
	// cap = 4 because the slice starts at index 1, and the array ends at index 4

	// ─── HOW APPEND WORKS ───
	// If len < cap: append stores at s[len], increments len. O(1).
	// If len == cap: allocate NEW, LARGER backing array, copy old data,
	//                append new element. Old array becomes garbage. O(n).
	//
	// GROWTH STRATEGY (Go 1.18+):
	// - If cap < 256: double capacity
	// - If cap >= 256: grow by ~25% + some constant (smoother growth)
	// - Old strategy (pre-1.18): double if < 1024, then 25%

	s5 := make([]int, 0, 2) // cap=2
	fmt.Printf("  Growth: len=%d cap=%d\n", len(s5), cap(s5))
	for i := 0; i < 10; i++ {
		s5 = append(s5, i)
		fmt.Printf("    append %d: len=%d cap=%d\n", i, len(s5), cap(s5))
	}

	// PERFORMANCE: if you know the size, pre-allocate!
	// make([]T, 0, expectedSize) — avoids repeated growth

	fmt.Println()
}

// =============================================================================
// PART 3: Slice Operations
// =============================================================================
func sliceOperations() {
	fmt.Println("--- SLICE OPERATIONS ---")

	// ─── Append ───
	s := []int{1, 2, 3}
	s = append(s, 4)       // append one
	s = append(s, 5, 6, 7) // append multiple
	other := []int{8, 9, 10}
	s = append(s, other...) // append another slice
	fmt.Printf("  Append: %v\n", s)

	// ─── Slicing syntax: s[low:high:max] ───
	// s[low:high]     — from low to high-1, cap = cap(s) - low
	// s[low:high:max] — from low to high-1, cap = max - low (LIMITS cap!)
	data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	sub1 := data[2:5]   // {2,3,4}, cap=8 (shares backing array!)
	sub2 := data[2:5:5] // {2,3,4}, cap=3 (limited! append won't overwrite data)
	fmt.Printf("  s[2:5]:   %v cap=%d\n", sub1, cap(sub1))
	fmt.Printf("  s[2:5:5]: %v cap=%d (full slice expression)\n", sub2, cap(sub2))

	// ─── Copy ───
	src := []int{1, 2, 3, 4, 5}
	dst := make([]int, 3)
	n := copy(dst, src) // copies min(len(dst), len(src)) elements
	fmt.Printf("  Copy: %v (copied %d elements)\n", dst, n)

	// copy can also shift elements within the same slice:
	s2 := []int{1, 2, 3, 4, 5}
	copy(s2[1:], s2[2:]) // shift left: [1, 3, 4, 5, 5]
	s2 = s2[:len(s2)-1]  // truncate: [1, 3, 4, 5]
	fmt.Printf("  Delete index 1 (shift): %v\n", s2)

	// ─── Delete element (order-preserving) — Go 1.21+ ───
	s3 := []int{1, 2, 3, 4, 5}
	s3 = slices.Delete(s3, 1, 2) // delete index 1 (inclusive) to 2 (exclusive)
	fmt.Printf("  slices.Delete: %v\n", s3)

	// ─── Delete element (no order, faster) ───
	s4 := []int{1, 2, 3, 4, 5}
	// swap with last, then truncate
	s4[1] = s4[len(s4)-1]
	s4 = s4[:len(s4)-1]
	fmt.Printf("  Swap-delete: %v (order lost)\n", s4)

	// ─── Insert at position ───
	s5 := []int{1, 2, 3, 5}
	s5 = slices.Insert(s5, 3, 4) // insert 4 at index 3
	fmt.Printf("  slices.Insert: %v\n", s5)

	// ─── Contains, Index (Go 1.21+) ───
	fmt.Printf("  Contains(3): %v\n", slices.Contains(s5, 3))
	fmt.Printf("  Index(4): %d\n", slices.Index(s5, 4))

	// ─── Sort (Go 1.21+) ───
	unsorted := []int{5, 3, 1, 4, 2}
	slices.Sort(unsorted)
	fmt.Printf("  Sorted: %v\n", unsorted)

	// ─── Nil slice vs empty slice ───
	var nilSlice []int          // nil (ptr=nil, len=0, cap=0)
	emptySlice := []int{}       // non-nil (ptr!=nil, len=0, cap=0)
	makeSlice := make([]int, 0) // non-nil
	fmt.Printf("  nil slice: %v (nil=%v)\n", nilSlice, nilSlice == nil)
	fmt.Printf("  empty:     %v (nil=%v)\n", emptySlice, emptySlice == nil)
	fmt.Printf("  make(0):   %v (nil=%v)\n", makeSlice, makeSlice == nil)
	// All three: len=0, safe to append, range, len.
	// Only difference: nil check and JSON marshaling (nil→null, empty→[])

	fmt.Println()
}

// =============================================================================
// PART 4: Slice Gotchas — Where Bugs Hide
// =============================================================================
func sliceGotchas() {
	fmt.Println("--- SLICE GOTCHAS ---")

	// ─── GOTCHA 1: Shared backing array ───
	original := []int{1, 2, 3, 4, 5}
	sub := original[1:3] // {2, 3} — shares the same backing array!
	sub[0] = 99
	fmt.Printf("  Shared array: original=%v (modified by sub!)\n", original)
	// original is now [1, 99, 3, 4, 5] because sub and original share memory

	// FIX: Use full slice expression or copy
	original2 := []int{1, 2, 3, 4, 5}
	safe := make([]int, 2)
	copy(safe, original2[1:3])
	safe[0] = 99
	fmt.Printf("  Safe copy: original=%v, copy=%v\n", original2, safe)

	// ─── GOTCHA 2: Append may or may not create new backing array ───
	a := make([]int, 3, 5) // len=3, cap=5
	a[0], a[1], a[2] = 1, 2, 3
	b := append(a, 4) // fits in cap! NO new array → b shares with a
	b[0] = 99         // changes a[0] too!
	fmt.Printf("  Append in-cap: a=%v, b=%v (shared!)\n", a, b)

	c := make([]int, 3, 3) // len=3, cap=3
	c[0], c[1], c[2] = 1, 2, 3
	d := append(c, 4) // exceeds cap → NEW array → independent
	d[0] = 99
	fmt.Printf("  Append over-cap: c=%v, d=%v (independent)\n", c, d)

	// RULE: Always use the return value of append: s = append(s, ...)

	// ─── GOTCHA 3: Memory leak via slice of large array ───
	// func getFirstByte(data []byte) []byte {
	//     return data[:1]  // holds reference to ENTIRE backing array!
	// }
	// If data is 1GB, the returned 1-byte slice keeps 1GB alive!
	//
	// FIX: copy the data
	// func getFirstByte(data []byte) []byte {
	//     result := make([]byte, 1)
	//     copy(result, data[:1])
	//     return result
	// }
	fmt.Println("  Gotcha 3: Small slice of big array → memory leak (copy to fix)")

	// ─── GOTCHA 4: Range iterates over a copy of the slice header ───
	// But the backing array is still shared!
	s := []int{1, 2, 3}
	for i, v := range s {
		if i == 0 {
			s = append(s, 4) // modifying s during range
		}
		fmt.Printf("    range: i=%d v=%d (s=%v)\n", i, v, s)
	}
	// Range uses the LENGTH captured at the start — won't see the appended element
	fmt.Printf("  After range+append: %v\n", s)

	fmt.Println()
}

// =============================================================================
// PART 5: Maps — Hash Tables
// =============================================================================
func mapFundamentals() {
	fmt.Println("--- MAPS ---")

	// map[KeyType]ValueType
	// Keys MUST be comparable (no slices, maps, or funcs as keys)

	// ─── Creating maps ───

	// 1. Literal
	ages := map[string]int{
		"Alice":   30,
		"Bob":     25,
		"Charlie": 35,
	}
	fmt.Printf("  Literal: %v\n", ages)

	// 2. make
	scores := make(map[string]int) // empty, ready to use
	scores["math"] = 95
	scores["science"] = 88
	fmt.Printf("  make: %v\n", scores)

	// ─── CRUD operations ───
	// Create/Update
	ages["Dave"] = 28
	ages["Alice"] = 31 // update existing

	// Read (returns zero value if key missing)
	age := ages["Alice"]
	missing := ages["Unknown"] // returns 0 (zero value for int)
	fmt.Printf("  Read: Alice=%d, Unknown=%d\n", age, missing)

	// ─── Comma-ok pattern: distinguish "missing" from "zero value" ───
	val, ok := ages["Unknown"]
	fmt.Printf("  Comma-ok: val=%d, ok=%v\n", val, ok)

	if age, ok := ages["Alice"]; ok {
		fmt.Printf("  Found Alice: %d\n", age)
	}

	// Delete
	delete(ages, "Bob")
	fmt.Printf("  After delete(Bob): %v\n", ages)
	delete(ages, "NonExistent") // no-op, no panic

	// ─── Length ───
	fmt.Printf("  len: %d\n", len(ages))

	// ─── Iteration: RANDOM ORDER ───
	// Map iteration order is deliberately randomized by Go runtime.
	// Do NOT depend on iteration order!
	fmt.Print("  Iteration: ")
	for k, v := range ages {
		fmt.Printf("%s=%d ", k, v)
	}
	fmt.Println("(random order!)")

	// ─── nil map behavior ───
	var nilMap map[string]int
	_ = nilMap["key"] // returns zero value (safe)
	_ = len(nilMap)   // returns 0 (safe)
	// nilMap["key"] = 1    // PANIC: assignment to entry in nil map
	// FIX: always initialize before writing
	fmt.Println("  nil map: read=safe, write=PANIC, delete=safe")

	// ─── Map is a reference type ───
	// Passing a map to a function passes a POINTER to the hash table.
	// Modifications inside the function affect the original.
	m := map[string]int{"x": 1}
	modifyMap(m)
	fmt.Printf("  Reference type: %v (modified by function)\n", m)

	// ─── Maps are NOT safe for concurrent access ───
	// Concurrent reads: OK
	// Concurrent read+write: RACE CONDITION (runtime may crash)
	// FIX: use sync.Mutex or sync.RWMutex or sync.Map
	fmt.Println("  Not goroutine-safe! Use sync.Mutex or sync.Map")

	fmt.Println()
}

func modifyMap(m map[string]int) {
	m["y"] = 2
}

// =============================================================================
// PART 6: Map Patterns
// =============================================================================
func mapPatterns() {
	fmt.Println("--- MAP PATTERNS ---")

	// ─── Set (map[T]bool or map[T]struct{}) ───
	// Go has no set type. Use a map with empty struct for zero-memory values.
	seen := map[string]struct{}{}
	seen["apple"] = struct{}{}
	seen["banana"] = struct{}{}
	if _, ok := seen["apple"]; ok {
		fmt.Println("  Set contains 'apple'")
	}
	// struct{} uses ZERO bytes of storage!

	// ─── Counting / frequency ───
	words := []string{"go", "is", "go", "fun", "go", "is", "great"}
	freq := make(map[string]int)
	for _, w := range words {
		freq[w]++ // zero value of int is 0, so this Just Works™
	}
	fmt.Printf("  Word freq: %v\n", freq)

	// ─── Grouping ───
	type Person struct {
		Name string
		City string
	}
	people := []Person{
		{"Alice", "NYC"}, {"Bob", "SF"}, {"Charlie", "NYC"}, {"Dave", "SF"},
	}
	byCity := make(map[string][]Person)
	for _, p := range people {
		byCity[p.City] = append(byCity[p.City], p) // append to nil slice is fine!
	}
	fmt.Printf("  Group by city: NYC=%d people, SF=%d people\n",
		len(byCity["NYC"]), len(byCity["SF"]))

	// ─── maps.Clone (Go 1.21+): shallow copy ───
	original := map[string]int{"a": 1, "b": 2}
	cloned := maps.Clone(original)
	cloned["c"] = 3
	fmt.Printf("  Clone: original=%v, cloned=%v\n", original, cloned)

	// ─── maps.Equal (Go 1.21+) ───
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"a": 1, "b": 2}
	fmt.Printf("  maps.Equal: %v\n", maps.Equal(m1, m2))

	// ─── Sorted map iteration ───
	m := map[string]int{"banana": 2, "apple": 1, "cherry": 3}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	fmt.Print("  Sorted keys: ")
	for _, k := range keys {
		fmt.Printf("%s=%d ", k, m[k])
	}
	fmt.Println()

	// ─── Map with struct key (composite key) ───
	type Coordinate struct{ X, Y int }
	grid := map[Coordinate]string{
		{0, 0}: "origin",
		{1, 2}: "point A",
	}
	fmt.Printf("  Struct key: %v\n", grid)

	fmt.Println()
}

// =============================================================================
// PART 7: make vs new
// =============================================================================
func makeVsNew() {
	fmt.Println("--- MAKE vs NEW ---")

	// make(): for slices, maps, channels ONLY
	// Returns the TYPE itself (not a pointer)
	// Initializes the internal data structure
	s := make([]int, 5)       // []int, ready to use
	m := make(map[string]int) // map, ready to use
	ch := make(chan int, 10)  // buffered channel, ready to use
	fmt.Printf("  make: slice=%v, map=%v, chan cap=%d\n", s, m, cap(ch))

	// new(): allocates memory for ANY type, returns a POINTER
	// Returns *T (pointer to zero value)
	// Rarely used — prefer &T{} or var t T
	p := new(int) // *int, points to zero-valued int
	*p = 42
	fmt.Printf("  new(int): %d\n", *p)

	// Equivalent alternatives to new (preferred):
	x := 0                        // on stack (if it doesn't escape)
	y := &x                       // pointer to x
	z := &struct{ Name string }{} // pointer to zero-valued struct
	fmt.Printf("  Preferred: y=%d, z=%+v\n", *y, z)

	// ─── SUMMARY ───
	// make: slices, maps, channels (initializes internal structure)
	// new:  anything (allocates, zeros, returns pointer) — rarely needed
	// &T{}: preferred over new(T) for structs
	// var t T: preferred when zero value is useful

	fmt.Println()
}
