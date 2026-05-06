//go:build ignore

// =============================================================================
// LESSON 0.5: POINTERS — Indirection, Sharing & Safety in Go
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - What pointers are and why Go has them
// - Pointer syntax: &, *, declaration
// - Pass-by-value vs pass-by-pointer
// - Pointer vs value receivers on methods
// - When to use pointers (and when NOT to)
// - nil pointers and safe handling
// - Go's pointer limitations (no pointer arithmetic, no dangling pointers)
//
// THE KEY INSIGHT:
// Go is strictly pass-by-value. Pointers exist to let you share data
// across function boundaries WITHOUT copying. But unlike C, Go pointers
// are SAFE: no arithmetic, no dangling references (GC handles lifetime),
// and nil dereference panics immediately (no silent corruption).
//
// RUN: go run 05_pointers.go
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== POINTERS ===")
	fmt.Println()

	pointerBasics()
	passValueVsPointer()
	pointerReceivers()
	pointerToStruct()
	nilPointers()
	whenToUsePointers()
	pointerSafety()
}

// =============================================================================
// PART 1: Pointer Basics
// =============================================================================
func pointerBasics() {
	fmt.Println("--- POINTER BASICS ---")

	// A pointer holds the MEMORY ADDRESS of a value.
	// Two operators:
	//   & (address-of): get the address of a variable
	//   * (dereference):  read/write the value at an address

	x := 42
	p := &x // p is *int, holds the address of x

	fmt.Printf("  x = %d\n", x)
	fmt.Printf("  &x = %p (address in memory)\n", p)
	fmt.Printf("  *p = %d (dereference: read value at address)\n", *p)

	// Modify through the pointer
	*p = 100
	fmt.Printf("  After *p = 100: x = %d (x is changed!)\n", x)

	// ─── Pointer types ───
	var intPtr *int    // pointer to int (zero value: nil)
	var strPtr *string // pointer to string
	fmt.Printf("  Zero pointer: intPtr=%v strPtr=%v\n", intPtr, strPtr)

	// ─── new() creates a pointer to a zero-valued type ───
	np := new(int) // *int, points to 0
	*np = 7
	fmt.Printf("  new(int): %d\n", *np)

	// ─── Pointer to pointer (rarely needed) ───
	pp := &p // **int (pointer to pointer to int)
	fmt.Printf("  **int: **pp = %d\n", **pp)

	fmt.Println()
}

// =============================================================================
// PART 2: Pass-by-Value vs Pass-by-Pointer
// =============================================================================

func doubleByValue(n int) {
	n *= 2 // modifies LOCAL copy
}

func doubleByPointer(n *int) {
	*n *= 2 // modifies the ORIGINAL through the pointer
}

func passValueVsPointer() {
	fmt.Println("--- PASS BY VALUE vs POINTER ---")

	// Go is ALWAYS pass-by-value.
	// When you pass a pointer, the POINTER is copied (not the data).
	// Both the caller and callee point to the same data.

	x := 10
	doubleByValue(x)
	fmt.Printf("  After doubleByValue:   x = %d (unchanged)\n", x)

	doubleByPointer(&x)
	fmt.Printf("  After doubleByPointer: x = %d (doubled!)\n", x)

	// ─── What gets copied for each type? ───
	// int, float, bool, struct  → the actual VALUE (full copy)
	// *T (pointer)              → the address (8 bytes on 64-bit)
	// string                    → (ptr, len) header (16 bytes) — data NOT copied
	// slice                     → (ptr, len, cap) header (24 bytes) — data NOT copied
	// map                       → a pointer internally — data NOT copied
	// interface                 → (type, value) pair (16 bytes)
	// func                      → a pointer
	// channel                   → a pointer
	//
	// TAKEAWAY: slices, maps, channels, and funcs are ALREADY reference-like.
	// You need pointers mainly for STRUCTS and basic types.

	// ─── Slice is already a reference (via header) ───
	s := []int{1, 2, 3}
	modifySlice(s) // modifies backing array
	fmt.Printf("  Slice after modifySlice: %v (modified!)\n", s)

	// But append might not be visible:
	s2 := []int{1, 2, 3}
	appendToSlice(s2)
	fmt.Printf("  Slice after appendToSlice: %v (NOT modified! header not updated)\n", s2)
	// To modify slice length visible to caller: pass *[]int or return the new slice

	fmt.Println()
}

func modifySlice(s []int) {
	s[0] = 99 // modifies backing array (visible to caller)
}

func appendToSlice(s []int) {
	s = append(s, 4) // new header, caller doesn't see it
}

// =============================================================================
// PART 3: Pointer Receivers vs Value Receivers
// =============================================================================

type Counter struct {
	n int
}

// Value receiver: gets a COPY of Counter
func (c Counter) Value() int {
	return c.n
}

// Pointer receiver: gets a POINTER to Counter
func (c *Counter) Increment() {
	c.n++ // modifies the original
}

// Value receiver with attempted modification
func (c Counter) IncrementBroken() {
	c.n++ // modifies a COPY, no effect on original
}

func pointerReceivers() {
	fmt.Println("--- POINTER vs VALUE RECEIVERS ---")

	c := Counter{n: 0}
	c.Increment() // Go auto-takes address: (&c).Increment()
	c.Increment()
	fmt.Printf("  After 2× Increment(): n=%d\n", c.Value())

	c.IncrementBroken()
	fmt.Printf("  After IncrementBroken(): n=%d (unchanged!)\n", c.Value())

	// ─── Auto-dereference: Go handles &/* automatically for method calls ───
	p := &Counter{n: 10}
	p.Value()     // Go auto-dereferences: (*p).Value()
	p.Increment() // already a pointer, no conversion needed

	c2 := Counter{n: 5}
	c2.Increment() // Go auto-takes address: (&c2).Increment()

	// ─── RULES FOR CHOOSING RECEIVER TYPE ───
	//
	// USE POINTER RECEIVER (*T) WHEN:
	// 1. Method modifies the receiver (mutates state)
	// 2. Struct is large (avoid copying)
	// 3. The type has methods with pointer receivers (consistency)
	// 4. Any method needs pointer receiver → ALL methods should use it
	//
	// USE VALUE RECEIVER (T) WHEN:
	// 1. Method doesn't modify the receiver (read-only)
	// 2. Struct is small (few fields, basic types)
	// 3. You want the receiver to be immutable within the method
	// 4. The type is map, func, or channel (already references)
	//
	// IMPORTANT: If ANY method uses pointer receiver, ALL should (consistency).
	// A type can be stored in interface differently based on receiver type:
	// - Value receiver methods: callable on both T and *T
	// - Pointer receiver methods: callable only on *T addressable values

	fmt.Println("  Rule: if ANY method mutates → use pointer receiver for ALL")

	fmt.Println()
}

// =============================================================================
// PART 4: Pointer to Struct (Most Common Use)
// =============================================================================

type User struct {
	Name  string
	Email string
	Age   int
}

// Constructor pattern: return a pointer
func NewUser(name, email string, age int) *User {
	return &User{
		Name:  name,
		Email: email,
		Age:   age,
	}
	// &User{} escapes to heap — compiler handles this automatically.
	// In C, this would be a dangling pointer bug. In Go, it's safe!
}

func pointerToStruct() {
	fmt.Println("--- POINTER TO STRUCT ---")

	// ─── Struct literal returns value, & returns pointer ───
	u1 := User{Name: "Alice", Email: "a@test.com", Age: 30} // value
	u2 := &User{Name: "Bob", Email: "b@test.com", Age: 25}  // pointer

	fmt.Printf("  Value: %+v\n", u1)
	fmt.Printf("  Pointer: %+v\n", *u2)

	// ─── Accessing fields: NO explicit dereference needed! ───
	// Go automatically dereferences pointers for field access.
	u2.Name = "Bobby" // same as (*u2).Name = "Bobby"
	fmt.Printf("  Auto-deref: %s\n", u2.Name)

	// ─── Constructor pattern ───
	u3 := NewUser("Vikram", "v@test.com", 25)
	fmt.Printf("  Constructor: %+v\n", *u3)

	// ─── Returning &localVar is SAFE in Go ───
	// The compiler's escape analysis detects this and allocates on the heap.
	// No dangling pointers!
	p := createPointer()
	fmt.Printf("  Returned pointer: %d (safe, GC manages lifetime)\n", *p)

	// ─── Struct with pointer fields ───
	type Node struct {
		Value int
		Next  *Node // self-referential (linked list, trees)
	}
	n1 := &Node{Value: 1}
	n2 := &Node{Value: 2, Next: n1}
	fmt.Printf("  Linked: %d → %d\n", n2.Value, n2.Next.Value)

	fmt.Println()
}

func createPointer() *int {
	x := 42
	return &x // safe! Go allocates x on heap, GC manages it
}

// =============================================================================
// PART 5: nil Pointers
// =============================================================================
func nilPointers() {
	fmt.Println("--- NIL POINTERS ---")

	// Zero value of a pointer is nil
	var p *int
	fmt.Printf("  Nil pointer: p=%v, p==nil: %v\n", p, p == nil)

	// ─── Dereferencing nil = PANIC ───
	// *p = 42  // PANIC: runtime error: invalid memory address or nil pointer dereference
	// This is Go's version of a segfault — immediate, not silent corruption.

	// ─── Safe nil checks ───
	if p != nil {
		fmt.Println("  Has value:", *p)
	} else {
		fmt.Println("  Nil: check before dereferencing")
	}

	// ─── Methods can be called on nil receivers (carefully!) ───
	var u *User
	// u.Name  // PANIC! accessing field on nil pointer
	// But a method with nil check can work:
	fmt.Printf("  Nil-safe method: %q\n", safeString(u))

	u = &User{Name: "Vikram"}
	fmt.Printf("  Non-nil method: %q\n", safeString(u))

	// ─── nil pointer in struct ───
	type Config struct {
		DB   *DBConfig
		Port int
	}
	cfg := Config{Port: 8080} // DB is nil
	if cfg.DB != nil {
		fmt.Println("  DB is configured")
	} else {
		fmt.Println("  DB is nil (not configured)")
	}

	fmt.Println()
}

type DBConfig struct {
	Host string
}

func safeString(u *User) string {
	if u == nil {
		return "<nil user>"
	}
	return u.Name
}

// =============================================================================
// PART 6: When to Use Pointers
// =============================================================================
func whenToUsePointers() {
	fmt.Println("--- WHEN TO USE POINTERS ---")

	// ─── USE pointers when: ───
	// 1. You need to modify the original value
	// 2. The struct is large (>64 bytes approximately)
	// 3. The value represents a unique identity (not copyable)
	//    e.g., sync.Mutex, database connection, file handle
	// 4. Interfacing with APIs that require pointers
	// 5. Representing "optional/absent" (nil = not set)

	// ─── DON'T use pointers when: ───
	// 1. The type is small (int, bool, small struct)
	// 2. You want immutability (value copy = can't be modified externally)
	// 3. The type is a map, slice, channel, func (already references)
	// 4. Performance: pointers can hurt cache locality
	//    (pointer chasing forces CPU to fetch from random memory locations)

	// ─── Optional/absent pattern ───
	type UpdateUser struct {
		Name  *string // nil = don't update, non-nil = update to this value
		Email *string
		Age   *int
	}
	name := "NewName"
	update := UpdateUser{Name: &name} // update name, leave email and age unchanged
	fmt.Printf("  Optional fields: name=%v, email=%v, age=%v\n",
		update.Name, update.Email, update.Age)

	// ─── Helper for creating pointer to literal ───
	// Can't do: &"hello" or &42 (address of literal not allowed)
	// Use a helper:
	s := ptr("hello")
	n := ptr(42)
	fmt.Printf("  Ptr helper: *string=%q, *int=%d\n", *s, *n)

	fmt.Println()
}

// Generic pointer helper (Go 1.18+)
func ptr[T any](v T) *T {
	return &v
}

// =============================================================================
// PART 7: Pointer Safety in Go (vs C)
// =============================================================================
func pointerSafety() {
	fmt.Println("--- POINTER SAFETY ---")

	// Go pointers are MUCH safer than C:
	//
	// 1. NO POINTER ARITHMETIC
	//    p++ or p + 4 is NOT allowed (use unsafe.Pointer for rare cases)
	//    This prevents buffer overflows and out-of-bounds access.
	//
	// 2. NO DANGLING POINTERS
	//    GC tracks all references. Memory is freed only when unreachable.
	//    Returning &localVar is safe (compiler promotes to heap).
	//
	// 3. NO UNINITIALIZED POINTERS
	//    Zero value is nil. Dereferencing nil panics immediately.
	//    No reading garbage memory.
	//
	// 4. NO VOID POINTERS
	//    interface{} (or any) is the closest equivalent, but it's type-safe.
	//    You must type-assert to use the value.
	//
	// 5. NO IMPLICIT CASTING
	//    *int and *int64 are different types. Must convert explicitly.
	//
	// The ONLY way to do "unsafe" pointer stuff:
	//   import "unsafe"
	//   unsafe.Pointer — converts to/from any pointer type
	//   This is intentionally named "unsafe" to make it obvious in code review.

	fmt.Println("  No pointer arithmetic → no buffer overflows")
	fmt.Println("  No dangling pointers → GC manages lifetime")
	fmt.Println("  No uninitialized reads → nil dereference panics immediately")
	fmt.Println("  unsafe.Pointer exists but is clearly marked 'unsafe'")

	fmt.Println()
}
