//go:build ignore

// =============================================================================
// LESSON 0.1: VARIABLES, TYPES & CONSTANTS — The Foundation of Everything
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Every way to declare variables in Go
// - All basic types plus their sizes, ranges, and zero values
// - String internals (immutable UTF-8 byte slices)
// - Type conversions (explicit only, no implicit casting)
// - Constants, iota, and untyped constant magic
// - Type definitions vs type aliases
// - Comparability rules
//
// RUN: go run 01_variables_types.go
// =============================================================================

package main

import (
	"fmt"
	"math"
	"unsafe"
)

func main() {
	fmt.Println("=== VARIABLES, TYPES & CONSTANTS ===")
	fmt.Println()

	variableDeclarations()
	numericTypes()
	stringInternals()
	zeroValues()
	typeConversions()
	constantsAndIota()
	typeDefinitionsVsAliases()
	comparabilityRules()
}

// =============================================================================
// PART 1: Variable Declarations — Every Way
// =============================================================================
func variableDeclarations() {
	fmt.Println("--- VARIABLE DECLARATIONS ---")

	// ─── Method 1: var with explicit type ───
	var name string = "Go"
	fmt.Printf("  var name string = %q\n", name)

	// ─── Method 2: var with type inference ───
	var age = 30 // compiler infers int
	fmt.Printf("  var age = %d (type: %T)\n", age, age)

	// ─── Method 3: short declaration := (most common) ───
	// ONLY works inside functions. NOT at package level.
	language := "Go" // type inferred from right side
	fmt.Printf("  language := %q (type: %T)\n", language, language)

	// ─── Method 4: var block (group related vars) ───
	var (
		host    = "localhost"
		port    = 8080
		verbose = true
	)
	fmt.Printf("  block: host=%s port=%d verbose=%v\n", host, port, verbose)

	// ─── Method 5: multiple assignment ───
	x, y, z := 1, 2.0, "three"
	fmt.Printf("  multi: x=%d(%T) y=%f(%T) z=%s(%T)\n", x, x, y, y, z, z)

	// ─── GOTCHA: := creates NEW variables, doesn't reassign ───
	outer := "original"
	{
		outer := "shadowed" // NEW variable in inner scope!
		fmt.Printf("  inner: %s\n", outer)
	}
	fmt.Printf("  outer: %s (unchanged!)\n", outer)
	// The compiler won't warn about this. Use `go vet -shadow` to catch it.

	// ─── GOTCHA: := with multiple returns ───
	// At least ONE variable on the left must be NEW
	a := 10
	a, b := 20, 30 // OK: b is new, a is reassigned
	fmt.Printf("  a=%d, b=%d\n", a, b)

	fmt.Println()
}

// =============================================================================
// PART 2: Numeric Types — Sizes, Ranges, Overflow
// =============================================================================
func numericTypes() {
	fmt.Println("--- NUMERIC TYPES ---")

	// ─── Integer types ───
	// Type       Size     Range
	// int8       1 byte   -128 to 127
	// int16      2 bytes  -32,768 to 32,767
	// int32      4 bytes  -2.1B to 2.1B
	// int64      8 bytes  -9.2×10^18 to 9.2×10^18
	// uint8      1 byte   0 to 255 (alias: byte)
	// uint16     2 bytes  0 to 65,535
	// uint32     4 bytes  0 to 4.2B (alias for Unicode: rune is int32)
	// uint64     8 bytes  0 to 18.4×10^18
	// int        platform 32 or 64 bit (use this by default)
	// uint       platform 32 or 64 bit

	fmt.Printf("  int size: %d bytes\n", unsafe.Sizeof(int(0)))
	fmt.Printf("  int8 range: %d to %d\n", math.MinInt8, math.MaxInt8)
	fmt.Printf("  int64 max: %d\n", math.MaxInt64)

	// ─── OVERFLOW: wraps silently! ───
	var i uint8 = 255
	i++ // wraps to 0, no error!
	fmt.Printf("  uint8(255) + 1 = %d (overflow wraps!)\n", i)

	// ─── Float types ───
	// float32: ~7 decimal digits precision
	// float64: ~15 decimal digits precision (DEFAULT for literals)
	pi := 3.14159265358979 // float64 by default
	var f32 float32 = 3.14
	fmt.Printf("  float64: %.15f\n", pi)
	fmt.Printf("  float32: %.15f (precision loss!)\n", float64(f32))

	// ─── GOTCHA: float comparison ───
	a := 0.1 + 0.2
	fmt.Printf("  0.1 + 0.2 = %.20f (NOT 0.3!)\n", a)
	// Never use == for floats. Use epsilon comparison:
	// math.Abs(a - b) < 1e-9

	// ─── Complex types ───
	c := complex(3, 4) // 3+4i
	fmt.Printf("  complex: %v, real=%.0f, imag=%.0f\n", c, real(c), imag(c))

	// ─── byte and rune ───
	// byte = uint8 (raw byte)
	// rune = int32 (Unicode code point)
	var b byte = 'A'
	var r rune = '世'
	fmt.Printf("  byte 'A' = %d, rune '世' = %d (U+%04X)\n", b, r, r)

	fmt.Println()
}

// =============================================================================
// PART 3: String Internals
// =============================================================================
func stringInternals() {
	fmt.Println("--- STRING INTERNALS ---")

	// Strings in Go are:
	// 1. IMMUTABLE (cannot change individual bytes)
	// 2. UTF-8 encoded byte slices
	// 3. A string header is just (pointer, length) — 16 bytes
	//
	// ┌──────────┐
	// │ pointer ─┼──→ [72, 101, 108, 108, 111]  "Hello"
	// │ length=5 │
	// └──────────┘

	s := "Hello, 世界" // mixed ASCII + multibyte UTF-8
	fmt.Printf("  string: %s\n", s)
	fmt.Printf("  len (bytes): %d\n", len(s))                    // 13 bytes
	fmt.Printf("  rune count:  %d\n", len([]rune(s)))            // 9 characters
	fmt.Printf("  byte[0]: %c (1 byte)\n", s[0])                 // 'H'
	fmt.Printf("  bytes[7:10]: %x (3 bytes = '世')\n", s[7:10])

	// ─── Iterating: bytes vs runes ───
	fmt.Print("  bytes: ")
	for i := 0; i < len(s); i++ {
		fmt.Printf("%02x ", s[i])
	}
	fmt.Println()

	fmt.Print("  runes: ")
	for _, r := range s { // range iterates by RUNE, not byte
		fmt.Printf("%c ", r)
	}
	fmt.Println()

	// ─── String immutability ───
	// s[0] = 'h'  // COMPILE ERROR: cannot assign to s[0]
	// To modify: convert to []byte, modify, convert back
	buf := []byte(s)
	buf[0] = 'h'
	modified := string(buf)
	fmt.Printf("  modified: %s\n", modified)
	// WARNING: []byte(s) copies the data. Not free for large strings.

	// ─── Raw strings ───
	raw := `no \n escape, multi
line works`
	fmt.Printf("  raw: %s\n", raw)

	fmt.Println()
}

// =============================================================================
// PART 4: Zero Values — Go's "Default Initialization"
// =============================================================================
func zeroValues() {
	fmt.Println("--- ZERO VALUES ---")

	// Go initializes ALL variables to their zero value. No uninitialized memory.
	var (
		i    int        // 0
		f    float64    // 0.0
		b    bool       // false
		s    string     // "" (empty string)
		p    *int       // nil
		sl   []int      // nil (nil slice)
		m    map[string]int // nil (nil map)
		ch   chan int    // nil
		fn   func()     // nil
		ifc  interface{} // nil
	)

	fmt.Printf("  int: %d\n", i)
	fmt.Printf("  float64: %f\n", f)
	fmt.Printf("  bool: %v\n", b)
	fmt.Printf("  string: %q\n", s)
	fmt.Printf("  pointer: %v\n", p)
	fmt.Printf("  slice: %v (nil: %v)\n", sl, sl == nil)
	fmt.Printf("  map: %v (nil: %v)\n", m, m == nil)
	fmt.Printf("  chan: %v\n", ch)
	fmt.Printf("  func: %v\n", fn)
	fmt.Printf("  interface: %v\n", ifc)

	// ─── Zero value usability ───
	// Some zero values are directly usable:
	// - sync.Mutex{} → ready to Lock/Unlock
	// - bytes.Buffer{} → ready to Write
	// - strings.Builder{} → ready to Write
	//
	// Some are NOT usable as zero:
	// - nil map: reading OK, writing PANICS
	// - nil slice: reading OK (len=0), append OK, but can't index
	// - nil pointer: dereferencing PANICS

	fmt.Println()
}

// =============================================================================
// PART 5: Type Conversions — Explicit Only
// =============================================================================
func typeConversions() {
	fmt.Println("--- TYPE CONVERSIONS ---")

	// Go has NO implicit type conversion. Everything is explicit.
	// This prevents subtle bugs from automatic widening/narrowing.

	var i int = 42
	var f float64 = float64(i) // explicit
	var u uint = uint(f)       // explicit
	fmt.Printf("  int(%d) → float64(%f) → uint(%d)\n", i, f, u)

	// ─── int ↔ string ───
	// string(65) does NOT give "65"! It gives "A" (rune 65)
	fmt.Printf("  string(65) = %q (rune, not number!)\n", string(rune(65)))
	// Use strconv or fmt.Sprintf for number→string:
	// strconv.Itoa(65)       → "65"
	// fmt.Sprintf("%d", 65)  → "65"

	// ─── Narrowing: data loss ───
	big := int64(256)
	small := int8(big) // truncated!
	fmt.Printf("  int64(%d) → int8(%d) (data loss!)\n", big, small)

	// ─── float → int: truncation, not rounding ───
	fl := 3.9
	n := int(fl)
	fmt.Printf("  float64(%.1f) → int(%d) (truncates, not rounds!)\n", fl, n)

	// ─── []byte ↔ string ───
	bs := []byte("hello")
	s := string(bs)
	fmt.Printf("  []byte↔string: %v ↔ %q\n", bs, s)
	// WARNING: both conversions COPY the data

	fmt.Println()
}

// =============================================================================
// PART 6: Constants and iota
// =============================================================================

// ─── Typed constants ───
const MaxRetries int = 3

// ─── Untyped constants: higher precision ───
const Pi = 3.14159265358979323846264338327950288

// ─── iota: auto-incrementing integer ───
const (
	Sunday    = iota // 0
	Monday           // 1
	Tuesday          // 2
	Wednesday        // 3
	Thursday         // 4
	Friday           // 5
	Saturday         // 6
)

// ─── iota for bit flags ───
const (
	ReadPerm   = 1 << iota // 1 (001)
	WritePerm              // 2 (010)
	ExecPerm               // 4 (100)
)

// ─── iota with skip ───
const (
	_  = iota // skip 0
	KB = 1 << (10 * iota) // 1 << 10 = 1024
	MB                     // 1 << 20
	GB                     // 1 << 30
	TB                     // 1 << 40
)

func constantsAndIota() {
	fmt.Println("--- CONSTANTS & IOTA ---")

	fmt.Printf("  MaxRetries: %d\n", MaxRetries)
	fmt.Printf("  Pi (high precision): %.30f\n", Pi)
	fmt.Printf("  Days: Sun=%d Mon=%d Sat=%d\n", Sunday, Monday, Saturday)
	fmt.Printf("  Perms: R=%d W=%d X=%d RWX=%d\n",
		ReadPerm, WritePerm, ExecPerm, ReadPerm|WritePerm|ExecPerm)
	fmt.Printf("  Sizes: KB=%d MB=%d GB=%d TB=%d\n", KB, MB, GB, TB)

	// ─── Untyped constant magic ───
	// Untyped constants have arbitrary precision and adapt to context:
	const x = 1 << 100 // way beyond int64, but OK as untyped
	// fmt.Println(x) // ERROR: can't print, too big for any type
	fmt.Printf("  1<<100 / 1<<99 = %d (untyped arithmetic works!)\n", x>>99)

	// Untyped constants automatically convert:
	const c = 42      // untyped int constant
	var f float64 = c // OK: untyped 42 becomes float64
	var b byte = c    // OK: untyped 42 becomes byte
	fmt.Printf("  Untyped 42: float64=%f byte=%d\n", f, b)

	fmt.Println()
}

// =============================================================================
// PART 7: Type Definitions vs Type Aliases
// =============================================================================

// Type DEFINITION: creates a new distinct type
type Celsius float64
type Fahrenheit float64

// Type ALIAS: just another name for the same type (Go 1.9+)
type Chars = []rune // Chars IS []rune, same type

func typeDefinitionsVsAliases() {
	fmt.Println("--- TYPE DEFINITIONS VS ALIASES ---")

	// ─── Type definition: new type, needs conversion ───
	var temp Celsius = 100
	// var f Fahrenheit = temp  // COMPILE ERROR: different types
	var f Fahrenheit = Fahrenheit(temp*9/5 + 32) // explicit conversion
	fmt.Printf("  %v°C = %v°F\n", temp, f)

	// Can add methods to defined types:
	fmt.Printf("  Celsius.String: %s\n", temp.String())

	// ─── Type alias: same type, no conversion needed ───
	var c Chars = []rune("hello") // Chars and []rune are interchangeable
	fmt.Printf("  Chars: %v\n", c)

	// ─── When to use which ───
	// Type definition: semantic meaning, method attachment, type safety
	//   type UserID int64
	//   type Money int64  // can't accidentally mix UserID and Money
	//
	// Type alias: gradual refactoring, compatibility
	//   type OldName = NewName  // migration helper

	fmt.Println()
}

func (c Celsius) String() string {
	return fmt.Sprintf("%.1f°C", c)
}

// =============================================================================
// PART 8: Comparability Rules
// =============================================================================
func comparabilityRules() {
	fmt.Println("--- COMPARABILITY ---")

	// Go types are comparable (can use == and !=) or not.
	//
	// COMPARABLE:
	// ✓ bool, int*, uint*, float*, complex*, string
	// ✓ pointer, channel
	// ✓ interface (but may panic at runtime)
	// ✓ struct (if ALL fields are comparable)
	// ✓ array (if element type is comparable)
	//
	// NOT COMPARABLE:
	// ✗ slice
	// ✗ map
	// ✗ function
	// (these can only compare to nil)

	a := [3]int{1, 2, 3}
	b := [3]int{1, 2, 3}
	fmt.Printf("  arrays ==: %v\n", a == b) // true

	type Point struct{ X, Y int }
	p1, p2 := Point{1, 2}, Point{1, 2}
	fmt.Printf("  structs ==: %v\n", p1 == p2) // true

	// Slices: can't compare with ==
	// s1 := []int{1, 2}
	// s2 := []int{1, 2}
	// s1 == s2  // COMPILE ERROR
	// Use slices.Equal(s1, s2) from Go 1.21+

	// ─── Map keys must be comparable ───
	// map[[]int]string  // COMPILE ERROR: slice not comparable
	// map[Point]string  // OK: struct with comparable fields

	// ─── Interface comparison gotcha ───
	// Two interfaces can be compared with ==
	// BUT if the underlying type is NOT comparable → PANIC at runtime!
	// var x, y interface{} = []int{1}, []int{1}
	// x == y  // PANIC: comparing uncomparable type []int

	fmt.Println()
}
















































































































































































































































































































































































































































































































































































































































}	fmt.Println()	fmt.Println("  Ordered types: int, float, string only")	// Structs, arrays, interfaces: comparable but NOT ordered!	// Only: int types, float types, string (lexicographic)	// ─── ORDERED types (can use <, >, <=, >=) ───	fmt.Printf("  Struct as map key: %v\n", configMap)	}		{Host: "prod.com", Port: 443}:  "prod",		{Host: "localhost", Port: 8080}: "dev",	configMap := map[Config]string{	}		Port int		Host string	type Config struct {	// Map keys must be comparable	// x == y  // PANIC at runtime! slice is not comparable	// var y interface{} = []int{1, 2}	// var x interface{} = []int{1, 2}	// ⚠️ Interface with non-comparable value → RUNTIME PANIC on ==	fmt.Printf("  Interface: 42 == 42 → %v, 42 == \"42\" → %v\n", i1 == i2, i1 == i3)	var i3 interface{} = "42"	var i2 interface{} = 42	var i1 interface{} = 42	// Interface comparison: compares (type, value) pair	// Fix: use slices.Equal(s1, s2) from Go 1.21+	// s1 == s2  // COMPILE ERROR	// s2 := []int{1, 2, 3}	// s1 := []int{1, 2, 3}	// Slice: NOT comparable	fmt.Printf("  Array: %v == %v → %v\n", a1, a2, a1 == a2)	a2 := [3]int{1, 2, 3}	a1 := [3]int{1, 2, 3}	// Array comparability	fmt.Printf("  Struct: %v == %v → %v\n", p1, p2, p1 == p2)	p2 := Point{1, 2}	p1 := Point{1, 2}	type Point struct{ X, Y int }	// Struct comparability	// NOT COMPARABLE → CANNOT be map keys!	//	// func                                 ❌ (can only compare to nil)	// map                                  ❌ (use maps.Equal)	// slice                                ❌ (use slices.Equal)	// ───────────────────────────────	// NOT COMPARABLE (cannot use ==):	//	// array                                ✅ ONLY if element type is comparable	// struct                               ✅ ONLY if ALL fields are comparable	// interface                            ✅ (compares dynamic type + value)	// channel                              ✅ (compares identity)	// pointer                              ✅ (compares addresses)	// bool, int, float, complex, string    ✅	// ───────────────────────────────	// COMPARABLE (can use == and !=):	//	// This matters for: ==, !=, map keys, and generic constraints.	// Go types are either COMPARABLE or NOT.	fmt.Println("--- COMPARABILITY RULES ---")func comparabilityRules() {// =============================================================================// PART 9: Comparability Rules// =============================================================================}	fmt.Println()	fmt.Printf("  Conversion: int64(789) → UserID = %d\n", uid2)	uid2 := UserID(int64(789))	// You CAN convert between types with the same underlying type:	// ─── Underlying type matters for conversions ───	//   - Rarely needed in application code	//   - byte and rune are aliases in the stdlib	//   - Large-scale refactoring (gradual migration between packages)	// Type Alias:	//	//   - Most of the time, this is what you want	//   - Add methods to types	//   - Prevent mixing incompatible values (UserID vs ProductID)	//   - Create domain types (UserID, Email, Money)	// Type Definition:	// ─── WHEN TO USE WHICH ───	fmt.Printf("  Alias: Seconds=int64, interchangeable: s=%d i=%d\n", s, i)	var i int64 = s // works! They're the same type	var s Seconds = 30	type Seconds = int64 // Seconds IS int64, interchangeable everywhere	// type rune = int32   (defined in Go itself)	// type byte = uint8   (defined in Go itself)	// ─── Type Alias: just another name for the SAME type ───	fmt.Printf("  UserID: %d (new type, can have methods)\n", uid)	// You can add methods to defined types (but not built-in or imported types)	_ = pid	// uid = pid  // COMPILE ERROR! Different types, even though both are int64	var pid ProductID = 456	var uid UserID = 123	type ProductID int64	type UserID int64	// The new type has the same underlying type but is NOT the same type.	// ─── Type Definition: creates a NEW type ───	fmt.Println("--- TYPE DEFINITIONS vs ALIASES ---")func typeDefinitionsAliases() {// =============================================================================// PART 8: Type Definitions vs Type Aliases// =============================================================================}	fmt.Println()	fmt.Printf("  iota reset: A=%d B=%d, C=%d D=%d\n", A, B, C, D)	)		D        // 1		C = iota // 0 again! new block = reset	const (	)		B        // 1		A = iota // 0	const (	// ─── iota resets per const block ───		userPerm&ReadPerm != 0, userPerm&ExecPerm != 0)	fmt.Printf("  Has Read? %v, Has Exec? %v\n",		ReadPerm, WritePerm, ExecPerm, userPerm)	fmt.Printf("  Bit flags: R=%d W=%d X=%d, R|W=%d\n",	userPerm := ReadPerm | WritePerm // 3 (011)	// Combine with OR:	)		ExecPerm               // 4  (100)		WritePerm              // 2  (010)		ReadPerm   = 1 << iota // 1  (001)	const (	// ─── iota with bit flags ───	fmt.Printf("  Size units: KB=%d MB=%d GB=%d TB=%d\n", KB, MB, GB, TB)	)		TB                    // 1 << 40		GB                    // 1 << 30		MB                    // 1 << 20 = 1048576		KB = 1 << (10 * iota) // 1 << 10 = 1024		_  = iota             // 0 (skip zero value)	const (	// ─── iota with expressions ───	fmt.Printf("  iota enum: Red=%d Green=%d Blue=%d\n", Red, Green, Blue)	)		Blue         // 2		Green        // 1 (iota increments automatically)		Red   = iota // 0	const (	// Simple enum:	//	// iota starts at 0 in each const block and increments by 1 per line.	// ─── iota: Go's auto-incrementing constant generator ───	// but it CAN be used in constant expressions.	// `huge` can never be assigned to an int variable (overflow),	fmt.Printf("  (1<<100) / (1<<99) = %d (computed at compile time!)\n", fraction)	const fraction = huge / (1 << 99) // = 2 (computed at compile time)	const huge = 1 << 100           // much bigger than any int type	// They're computed at compile time with full precision!	// Untyped constants have arbitrary precision (at least 256 bits).	// ─── UNTYPED CONSTANT PRECISION ───	fmt.Printf("  Untyped const 42: int8=%d, int64=%d, float64=%f\n", a, b, c)	var c float64 = x  // works! untyped int → float64	var b int64 = x    // works!	var a int8 = x     // works! x fits in int8	// x can be used as ANY integer type without casting:	const z = "hello"    // untyped string constant	const y = 3.14       // untyped float constant — kind: float	const x = 42         // untyped integer constant — kind: int	// It can be used wherever a compatible type is expected.	// An untyped constant has a KIND but no fixed type.	// ─── Untyped constants (the interesting part) ───	const pi float64 = 3.14159265358979	const maxRetries int = 3	// ─── Typed constants ───	fmt.Println("--- CONSTANTS ---")func constantsMastery() {// =============================================================================// PART 7: Constants — Deeper Than You Think// =============================================================================}	fmt.Println()	// The Go philosophy: clarity over convenience. Every conversion is visible.	// In Go: var y float64 = float64(x) // explicit, reader sees the conversion	// In C:  int x = 3; float y = x; // implicit, might lose precision	// ─── WHY NO IMPLICIT CONVERSIONS? ───	fmt.Printf("  Celsius→Fahrenheit: %.0f°C → %.0f°F\n", c, f2)	var f2 Fahrenheit = Fahrenheit(c*9/5 + 32) // explicit conversion	// var f2 Fahrenheit = c   // COMPILE ERROR: different named types	var c Celsius = 100	type Fahrenheit float64	type Celsius float64	// ─── SAME UNDERLYING TYPE conversions ───	fmt.Printf("  []rune→string: %q\n", string(rs))	rs := []rune{72, 101, 108, 108, 111}	// []rune → string: for Unicode	fmt.Printf("  []byte↔string: %v → %q (copies both ways)\n", bs, str)	str := string(bs)	bs := []byte("hello")	// []byte → string and string → []byte: copies data	// Use strconv.Itoa(65) for "65" or fmt.Sprintf("%d", 65)	fmt.Printf("  string(65) = %q (Unicode 'A', NOT \"65\"!)\n", ch)	ch := string(65) // "A", not "65"!	// int → string: gives the Unicode character, NOT the number!	// ─── string conversions ───	fmt.Printf("  float64→int: %f → %d (truncation, not rounding)\n", pi, truncated)	var truncated int = int(pi)	var pi float64 = 3.99	// float → int truncates toward zero	fmt.Printf("  int64→int8: %d → %d (OVERFLOW! silent truncation)\n", big, small)	var small int8 = int8(big) // overflows! No error!	var big int64 = 100000	// ─── DANGER: narrowing conversions truncate silently ───	fmt.Printf("  int→float64→uint: %d → %f → %d\n", i, f, u)	var u uint = uint(f)       // float64 → uint (truncates decimal)	var f float64 = float64(i) // int → float64 (explicit)	var i int = 42	// ─── Numeric conversions ───	// This catches bugs at compile time that other languages miss at runtime.	// Go has NO implicit type conversions. Every conversion is explicit.	fmt.Println("--- TYPE CONVERSIONS ---")func typeConversions() {// =============================================================================// PART 6: Type Conversions — Explicit Only, No Surprises// =============================================================================}	fmt.Println()	fmt.Println("  Design pattern: make your types useful at zero value")	// This is intentional. Strive for this in your own types!	//   strings.Builder{}  — ready to WriteString	//   bytes.Buffer{}     — ready to Write	//   sync.Mutex{}       — ready to Lock/Unlock	// Many stdlib types are useful at zero value:	// ─── Zero-value usability is a DESIGN PATTERN in Go ───	fmt.Printf("  map after make: %v\n", m)	m["key"] = 1	m = make(map[string]int)	// FIX: m = make(map[string]int)	// m["key"] = 1  // PANIC: assignment to entry in nil map	// _ = m["key"]  // returns 0 (safe)	// nil map: safe to read (returns zero value) but PANICS on write!	fmt.Printf("  append to nil slice: %v\n", sl)	sl = append(sl, 1, 2, 3) // append to nil slice is fine	// nil slice: safe to read len, cap, range, append → works!	// ─── Zero value USABILITY ───	fmt.Printf("  interface: %v (nil=%v)\n", ifc, ifc == nil)	fmt.Printf("  func:      %v (nil=%v)\n", fn, fn == nil)	fmt.Printf("  chan:       %v (nil=%v)\n", ch, ch == nil)	fmt.Printf("  map:       %v (nil=%v)\n", m, m == nil)	fmt.Printf("  []int:     %v (nil=%v, len=%d)\n", sl, sl == nil, len(sl))	fmt.Printf("  *int:      %v\n", p)	fmt.Printf("  bool:      %v\n", b)	fmt.Printf("  string:    %q\n", s)	fmt.Printf("  float64:   %f\n", f)	fmt.Printf("  int:       %d\n", i)	)		ifc  interface{}		fn   func()		ch   chan int		m    map[string]int		sl   []int		p    *int		b    bool		s    string		f    float64		i    int	var (	// array             all elements are their zero values	// struct            all fields are their zero values	// interface         nil	// func              nil	// channel           nil (⚠️ blocks forever on send/receive)	// map               nil (⚠️ NOT usable for writes! Must make())	// slice             nil (but len=0, cap=0 — usable for append!)	// pointer           nil	// string            "" (empty string)	// int, float, etc.  0	// bool              false	// ─────────────────────────────	// TYPE              ZERO VALUE	//	// This eliminates an entire class of bugs.	// You can NEVER have an uninitialized variable (unlike C/C++).	// EVERY type in Go has a well-defined zero value.	fmt.Println("--- ZERO VALUES ---")func zeroValues() {// =============================================================================// PART 5: Zero Values — Go Has NO Uninitialized Variables// =============================================================================}	fmt.Println()	fmt.Printf("  []byte: %v → %q\n", data, string(data))	data := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F} // "Hello" in ASCII	// ─── Byte slices for binary data ───	fmt.Printf("  byte: %d (0x%02X)\n", raw, raw)	var raw byte = 0xFF	// byte = uint8. Used for raw binary data, ASCII text.	// if cheap() && expensive() { ... }	// expensive() is NOT called if cheap() returns false:	// && and || short-circuit (like most languages)	// ─── Short-circuit evaluation ───	fmt.Printf("  bool: %v, !bool: negation, no int-to-bool conversion\n", b)	b = !b // negation	b = true	var b bool // false	// if n != 0 { } // explicit comparison required	// if 1 { }  // COMPILE ERROR in Go	// No implicit conversion from int to bool (unlike C).	// bool: true or false. Zero value is false.	fmt.Println("--- BOOL AND BYTE ---")func boolAndByte() {// =============================================================================// PART 4: Bool and Byte// =============================================================================}	fmt.Println()	// For high-perf: work with []byte directly, convert to string at the end.	// PERFORMANCE: string ↔ []byte copies data each time.	_ = back	fmt.Printf("  []byte length: %d, []rune length: %d\n", len(bytes), len(runes))	back := string(bytes) // byte slice → string (COPIES!)	runes := []rune(s)    // string → rune slice (COPIES! one rune per character)	bytes := []byte(s)    // string → byte slice (COPIES!)	// ─── String ↔ []byte ↔ []rune conversions ───	fmt.Printf("  byte 'A' = %d\n", b)	fmt.Printf("  rune '世' = %d (U+%04X)\n", r, r)	var b byte = 'A'          // fits in byte (ASCII)	var r rune = '世'         // single quotes = rune literal	// byte = uint8, represents a single byte	// rune = int32, represents a Unicode code point	// ─── rune type ───	fmt.Printf("  Concat: %q\n", greeting)	greeting := "Hello" + ", " + "World" // 2 allocations	// For building strings: use strings.Builder (see 15-stdlib-mastery)	// + operator creates a NEW string every time (allocates)	// ─── String concatenation ───	fmt.Printf("  Raw string: %q\n", raw[:40])multiple lines`and it can span	raw := `no escape sequences: \n \t \" are literal	// ─── Raw strings: backticks ───	// Must create new string: s = "h" + s[1:]	// s[0] = 'h'  // COMPILE ERROR: strings are immutable	// ─── String immutability ───	// for i := 0; i < len(s); i++ iterates by BYTE — wrong for multi-byte!	fmt.Println()	}		fmt.Printf("(%d:%c) ", i, ch) // i is byte offset	for i, ch := range s {	fmt.Print("  for range: ")	// for range iterates by RUNE (character), not byte	// ─── Iterating correctly ───	fmt.Printf("  s[7] = %d (first byte of '世', NOT the character)\n", s[7])	fmt.Printf("  s[0] = %d (byte 'H')\n", s[0])	// Indexing returns a BYTE, not a character	fmt.Printf("  len() = %d bytes\n", len(s))	fmt.Printf("  String: %q\n", s)	// len() returns BYTES, not characters	s := "Hello, 世界!" // mixing ASCII and multi-byte characters	//   }	//       len int    // length in BYTES (not characters!)	//       ptr *byte  // pointer to immutable byte array	//   type string struct {	// Internal representation:	//	// By convention it holds UTF-8 text, but it can hold arbitrary bytes.	// A string is a read-only slice of bytes.	fmt.Println("--- STRING TYPE ---")func stringType() {// =============================================================================// PART 3: Strings — Immutable UTF-8 Byte Sequences// =============================================================================}	fmt.Println()	fmt.Printf("  complex: %v, real: %f, imag: %f\n", c, real(c), imag(c))	c := complex(3, 4) // 3+4i	// complex128 (two float64)	// complex64  (two float32)	// ──────────────	// COMPLEX TYPES:	fmt.Printf("  IsNaN: %v, IsInf: %v\n", math.IsNaN(nan), math.IsInf(inf, 0))	fmt.Printf("  NaN == NaN? %v (always false!)\n", nan == nan)	fmt.Printf("  Inf: %f, -Inf: %f, NaN: %f\n", inf, negInf, nan)	nan := math.NaN()	negInf := math.Inf(-1)	inf := math.Inf(1)	// ─── Special float values ───	fmt.Printf("  With epsilon: %v\n", math.Abs(a-0.3) < epsilon)	epsilon := 1e-9	// NEVER compare floats with ==. Use an epsilon:	fmt.Printf("  0.1 + 0.2 == 0.3? %v (it's %.17f)\n", a == 0.3, a)	a := 0.1 + 0.2	// ─── Floating point comparison trap ───	fmt.Printf("  float64 max: %e\n", math.MaxFloat64)	fmt.Printf("  float32 max: %e\n", math.MaxFloat32)	// Literal 3.14 is float64 by default.	// RULE: Always use float64 unless you have a specific reason for float32.	//	// float64  ~15-16 decimal digits precision  (8 bytes, IEEE 754) — DEFAULT	// float32  ~6-7 decimal digits precision   (4 bytes, IEEE 754)	// ─────────────────────	// FLOATING POINT TYPES:	fmt.Printf("  int8 overflow: 127 + 1 = %d (wraps!)\n", i8)	i8++ // wraps to -128!	var i8 int8 = 127	fmt.Printf("  uint8 overflow: 255 + 1 = %d (wraps!)\n", u8)	u8++ // wraps to 0, no error!	var u8 uint8 = 255	// ─── Integer overflow wraps silently! ───	fmt.Printf("  int range: %d to %d\n", math.MinInt, math.MaxInt)	fmt.Printf("  int size:  %d bytes (%d bits)\n", unsafe.Sizeof(int(0)), unsafe.Sizeof(int(0))*8)	// - Use sized types for binary protocols and memory-sensitive code	// - Use `int64` when you need guaranteed 64-bit (IDs, timestamps)	// - Use `int` for general integers (loop counters, etc.)	// RULE OF THUMB:	//	// uintptr big enough to hold a pointer  (unsafe, used with unsafe.Pointer)	// uint    platform-dependent (same as int)	// uint64  0 to 18.4×10¹⁸               (8 bytes)	// uint32  0 to 4294967295              (4 bytes)	// uint16  0 to 65535                    (2 bytes)	// uint8   0 to 255                      (1 byte) — alias: byte	//	// int     platform-dependent: 64-bit on 64-bit OS, 32-bit on 32-bit OS	// int64   -9.2×10¹⁸ to 9.2×10¹⁸       (8 bytes)	// int32   -2147483648 to 2147483647    (4 bytes) — alias: rune	// int16   -32768 to 32767              (2 bytes)	// int8    -128 to 127                  (1 byte)	// ──────────────	// INTEGER TYPES:	fmt.Println("--- NUMERIC TYPES ---")func numericTypes() {// =============================================================================// PART 2: Numeric Types — Sizes, Ranges, and Gotchas// =============================================================================}	fmt.Println()	fmt.Printf("  redeclaration: a=%d b=%d c=%d\n", a, b, c)	b, c := 3, 4 // `b` is reassigned, `c` is new. This compiles!	a, b := 1, 2	// := can redeclare a variable IF at least one variable on the left is NEW	// ─── REDECLARATION with := ───	_ = err // explicitly discard error (better than ignoring)	_, err := fmt.Println("  blank identifier: this line prints")	// Discards a value. Used when you must accept a return value but don't need it.	// ─── BLANK IDENTIFIER _ ───	// Use `go vet -shadow` or `staticcheck` to catch this.	// The compiler WON'T warn about shadowing by default.	// FIX: use = instead of := if you want to modify outer variable	fmt.Printf("  shadow outer: %q (unchanged!)\n", value)	}		fmt.Printf("  shadow inner: %q\n", value)		value := "inner" // NEW variable! Shadows outer `value`	{	value := "outer"	// ─── SHADOWING (subtle bug source) ───	fmt.Printf("  after swap: x=%d y=%d\n", x, y)	x, y = y, x // Go evaluates RHS fully before assigning to LHS	// ─── METHOD 5: Swap without temp variable ───	fmt.Printf("  multi-assign: x=%d y=%d\n", x, y)	x, y := 10, 20	// Multiple return values → short declaration shines		maxConns, timeout, keepAlive)	fmt.Printf("  grouped: maxConns=%d timeout=%d keepAlive=%v\n",	)		keepAlive = true		timeout   = 30		maxConns  = 100	var (	// ─── METHOD 4: Multiple declarations ───	fmt.Printf("  short: host=%q retries=%d pi=%f\n", host, retries, pi)	pi := 3.14159       // inferred as float64 (not float32!)	retries := 3        // inferred as int	host := "localhost" // inferred as string	// CANNOT be used at package level!	// USE WHEN: inside functions, type is obvious from the right side	// ─── METHOD 3: Short declaration := (most common) ───	fmt.Printf("  var init: port=%d ratio=%f\n", port, ratio)	var ratio float64 = 0.75 // without type, this would be float64 anyway	var port int = 8080	// USE WHEN: type can't be inferred, or you want to be explicit	// ─── METHOD 2: var with initialization ───		name, count, active, balance)	fmt.Printf("  var (zero values): name=%q count=%d active=%v balance=%f\n",	var balance float64 // zero value: 0.0	var active bool     // zero value: false	var count int       // zero value: 0	var name string     // zero value: ""	// USE WHEN: declaring at package level OR when zero value is the correct initial value	// ─── METHOD 1: var with type (explicit) ───	fmt.Println("--- VARIABLE DECLARATIONS ---")func declarationMastery() {// =============================================================================// PART 1: Variable Declaration — Every Way and When to Use Each// =============================================================================}	comparabilityRules()	typeDefinitionsAliases()	constantsMastery()	typeConversions()	zeroValues()	boolAndByte()	stringType()	numericTypes()	declarationMastery()	fmt.Println()	fmt.Println("=== VARIABLES, TYPES & THE TYPE SYSTEM ===")func main() {)	"unsafe"	"math"	"fmt"import (package main// =============================================================================// RUN: go run 01_variables_types.go//// by large teams. The type system is your documentation.// Every design choice serves one goal: make large codebases maintainable// no classes, no inheritance, no generics (until 1.18), no exceptions.// Go's type system is intentionally simple but has subtle depth. There are// THE KEY INSIGHT://// - Comparable types, ordered types (affects maps and generics)// - Type aliases vs type definitions// - Constants, iota, and untyped constants (surprisingly deep)// - Type conversions vs assertions (and why Go has no implicit conversions)// - Zero values: why Go has NO uninitialized variables// - The complete Go type system: basic, composite, reference, interface// - Every way to declare variables (and when to use which)// WHAT YOU'LL LEARN://// =============================================================================// LESSON 0.1: VARIABLES, TYPES & THE TYPE SYSTEM — Go's Foundation// =============================================================================//go:build ignore