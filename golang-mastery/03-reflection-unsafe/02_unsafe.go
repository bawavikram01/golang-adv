// =============================================================================
// LESSON 3.2: unsafe PACKAGE — Breaking Go's Type Safety
// =============================================================================
//
// The unsafe package lets you bypass Go's type system for:
//   - Interop with C code (CGo)
//   - Performance-critical zero-copy operations
//   - Inspecting internal data structures
//
// WARNING: unsafe code can crash your program, corrupt memory, and break
// across Go versions. Use ONLY when absolutely necessary.
//
// Three key functions:
//   unsafe.Sizeof(x)    — size in bytes (compile-time constant)
//   unsafe.Alignof(x)   — alignment requirement
//   unsafe.Offsetof(s.f) — byte offset of field f in struct s
//   unsafe.Pointer       — generic pointer (bridges any pointer type)
//   unsafe.Add           — pointer arithmetic (Go 1.17+)
//   unsafe.Slice         — create slice from pointer (Go 1.17+)
// =============================================================================

package main

import (
	"fmt"
	"unsafe"
)

// =============================================================================
// PART 1: Struct Layout and Padding
// =============================================================================
// Go aligns struct fields to their natural alignment. This creates padding.
// Understanding layout lets you optimize memory usage.

// BAD layout: lots of padding
type BadLayout struct {
	a bool   // 1 byte  + 7 bytes padding (next field needs 8-byte alignment)
	b int64  // 8 bytes
	c bool   // 1 byte  + 3 bytes padding
	d int32  // 4 bytes
	e bool   // 1 byte  + 7 bytes padding (struct alignment)
}
// Total: 32 bytes (only 15 bytes of actual data!)

// GOOD layout: minimize padding by ordering fields largest-to-smallest
type GoodLayout struct {
	b int64  // 8 bytes
	d int32  // 4 bytes
	a bool   // 1 byte
	c bool   // 1 byte
	e bool   // 1 byte + 1 byte padding
}
// Total: 16 bytes (same data, half the memory!)

func demonstrateLayout() {
	fmt.Println("=== Struct Layout & Padding ===")

	fmt.Printf("BadLayout  size: %d bytes\n", unsafe.Sizeof(BadLayout{}))
	fmt.Printf("GoodLayout size: %d bytes\n", unsafe.Sizeof(GoodLayout{}))

	// Show field offsets
	var bad BadLayout
	fmt.Printf("\nBadLayout field offsets:\n")
	fmt.Printf("  a (bool):  offset=%d, size=%d, align=%d\n",
		unsafe.Offsetof(bad.a), unsafe.Sizeof(bad.a), unsafe.Alignof(bad.a))
	fmt.Printf("  b (int64): offset=%d, size=%d, align=%d\n",
		unsafe.Offsetof(bad.b), unsafe.Sizeof(bad.b), unsafe.Alignof(bad.b))
	fmt.Printf("  c (bool):  offset=%d, size=%d, align=%d\n",
		unsafe.Offsetof(bad.c), unsafe.Sizeof(bad.c), unsafe.Alignof(bad.c))
	fmt.Printf("  d (int32): offset=%d, size=%d, align=%d\n",
		unsafe.Offsetof(bad.d), unsafe.Sizeof(bad.d), unsafe.Alignof(bad.d))
	fmt.Printf("  e (bool):  offset=%d, size=%d, align=%d\n",
		unsafe.Offsetof(bad.e), unsafe.Sizeof(bad.e), unsafe.Alignof(bad.e))

	var good GoodLayout
	fmt.Printf("\nGoodLayout field offsets:\n")
	fmt.Printf("  b (int64): offset=%d\n", unsafe.Offsetof(good.b))
	fmt.Printf("  d (int32): offset=%d\n", unsafe.Offsetof(good.d))
	fmt.Printf("  a (bool):  offset=%d\n", unsafe.Offsetof(good.a))
	fmt.Printf("  c (bool):  offset=%d\n", unsafe.Offsetof(good.c))
	fmt.Printf("  e (bool):  offset=%d\n", unsafe.Offsetof(good.e))
}

// =============================================================================
// PART 2: unsafe.Pointer — Type-punning and pointer arithmetic
// =============================================================================
//
// Legal conversions (Go spec guarantees these):
//   1. *T → unsafe.Pointer   (any pointer to unsafe.Pointer)
//   2. unsafe.Pointer → *T   (unsafe.Pointer to any pointer)
//   3. unsafe.Pointer → uintptr (for arithmetic, MUST convert back immediately)
//   4. uintptr → unsafe.Pointer (ONLY in the same expression as step 3)
//
// CRITICAL: Never store a uintptr in a variable! The GC doesn't track uintptr,
// so the object could be moved/collected while you hold a stale address.

func demonstratePointerConversion() {
	fmt.Println("\n=== Pointer Conversion ===")

	// Convert between incompatible pointer types
	var x float64 = 3.14

	// View the raw bits of a float64 as a uint64
	bits := *(*uint64)(unsafe.Pointer(&x))
	fmt.Printf("float64 %f = 0x%016x as uint64\n", x, bits)

	// Convert back
	restored := *(*float64)(unsafe.Pointer(&bits))
	fmt.Printf("Restored: %f\n", restored)
}

func demonstratePointerArithmetic() {
	fmt.Println("\n=== Pointer Arithmetic ===")

	type Data struct {
		A int32
		B int64
		C int16
	}

	d := Data{A: 100, B: 200, C: 300}

	// Access field B using pointer arithmetic (DON'T do this — use Offsetof!)
	// This is how the runtime accesses struct fields internally.
	bPtr := (*int64)(unsafe.Add(unsafe.Pointer(&d), unsafe.Offsetof(d.B)))
	fmt.Printf("d.B via pointer arithmetic: %d\n", *bPtr)

	// Modify through the computed pointer
	*bPtr = 999
	fmt.Printf("d.B after modification: %d\n", d.B)

	// Access array elements via pointer arithmetic
	arr := [5]int{10, 20, 30, 40, 50}
	elemSize := unsafe.Sizeof(arr[0])

	for i := 0; i < 5; i++ {
		ptr := (*int)(unsafe.Add(unsafe.Pointer(&arr[0]), uintptr(i)*elemSize))
		fmt.Printf("arr[%d] = %d\n", i, *ptr)
	}
}

// =============================================================================
// PART 3: Zero-copy string ↔ []byte conversion
// =============================================================================
//
// Normal string([]byte) and []byte(string) allocate and copy.
// unsafe lets you avoid the copy, but YOU must ensure safety:
//   - Don't modify the []byte after converting to string
//   - The string must not outlive the []byte

// Internal representation (from runtime):
// type stringHeader struct {
//     Data unsafe.Pointer
//     Len  int
// }
// type sliceHeader struct {
//     Data unsafe.Pointer
//     Len  int
//     Cap  int
// }

func stringToBytes(s string) []byte {
	// Go 1.20+: use unsafe.StringData and unsafe.Slice
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func bytesToString(b []byte) string {
	// Go 1.20+: use unsafe.String
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

func demonstrateZeroCopy() {
	fmt.Println("\n=== Zero-Copy String <-> []byte ===")

	original := "Hello, unsafe world!"
	b := stringToBytes(original)
	fmt.Printf("Bytes: %v\n", b)

	// Convert back
	s := bytesToString(b)
	fmt.Printf("String: %s\n", s)

	// WARNING: b and s share the same memory!
	// Modifying b would modify s too (UNSAFE!)
	fmt.Printf("Same data pointer: %v\n",
		unsafe.StringData(original) == unsafe.StringData(s))
}

// =============================================================================
// PART 4: unsafe.Slice — Create slices from raw pointers
// =============================================================================

func demonstrateUnsafeSlice() {
	fmt.Println("\n=== unsafe.Slice ===")

	// Simulate C-style memory: raw array from a pointer
	data := [10]int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Create a Go slice pointing to elements [3:7] without copying
	ptr := &data[3]
	slice := unsafe.Slice(ptr, 4) // 4 elements starting from ptr
	fmt.Printf("Slice from pointer: %v\n", slice)

	// Modification affects original
	slice[0] = 999
	fmt.Printf("Original data[3] after modification: %d\n", data[3])
}

// =============================================================================
// PART 5: Inspecting interface internals
// =============================================================================
//
// An interface value in Go is actually a pair of pointers:
//   - iface: {tab *itab, data unsafe.Pointer}     (non-empty interface)
//   - eface: {_type *_type, data unsafe.Pointer}   (empty interface/any)

type iface struct {
	tab  uintptr // pointer to type+method info
	data uintptr // pointer to actual data
}

func inspectInterface(v interface{}) {
	fmt.Println("\n=== Interface Internals ===")

	// Peek at the interface's internal representation
	ifacePtr := (*iface)(unsafe.Pointer(&v))
	fmt.Printf("Interface type pointer:  0x%x\n", ifacePtr.tab)
	fmt.Printf("Interface data pointer:  0x%x\n", ifacePtr.data)

	// A nil interface has both pointers as 0
	var nilIface interface{}
	nilPtr := (*iface)(unsafe.Pointer(&nilIface))
	fmt.Printf("Nil interface - type: 0x%x, data: 0x%x\n", nilPtr.tab, nilPtr.data)

	// An interface holding a nil pointer is NOT nil!
	var nilErrPtr *BadLayout = nil
	var i interface{} = nilErrPtr
	iPtr := (*iface)(unsafe.Pointer(&i))
	fmt.Printf("Interface with nil ptr - type: 0x%x (not nil!), data: 0x%x\n",
		iPtr.tab, iPtr.data)
	fmt.Printf("i == nil: %v (GOTCHA! This is false because type info exists)\n", i == nil)
}

func main() {
	demonstrateLayout()
	demonstratePointerConversion()
	demonstratePointerArithmetic()
	demonstrateZeroCopy()
	demonstrateUnsafeSlice()
	inspectInterface(42)

	fmt.Println("\n=== KEY TAKEAWAYS ===")
	fmt.Println("1. Order struct fields by size (largest first) to minimize padding")
	fmt.Println("2. unsafe.Pointer is the only legal bridge between pointer types")
	fmt.Println("3. NEVER store uintptr in a variable — GC doesn't track it")
	fmt.Println("4. Zero-copy conversions are fast but dangerous if you break invariants")
	fmt.Println("5. An interface with a nil pointer is NOT a nil interface")
}
