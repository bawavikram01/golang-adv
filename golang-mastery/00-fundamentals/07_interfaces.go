//go:build ignore

// =============================================================================
// LESSON 0.7: INTERFACES — Go's Most Powerful Feature
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Interface basics: implicit satisfaction (no "implements" keyword)
// - The empty interface (any) and type assertions
// - Interface design principles: small, composable
// - Standard library interfaces everyone must know
// - Type switches for polymorphism
// - The nil interface gotcha
// - Interface embedding and composition
// - When to define interfaces (accept interfaces, return structs)
//
// THE KEY INSIGHT:
// In Go, interfaces are IMPLICIT. A type satisfies an interface just by
// having the right methods — no declaration needed. This means packages
// don't depend on each other: the producer doesn't need to know about
// the consumer's interface. This is Go's decoupling superpower.
//
// RUN: go run 07_interfaces.go
// =============================================================================

package main

import (
	"fmt"
	"math"
	"strings"
)

func main() {
	fmt.Println("=== INTERFACES ===")
	fmt.Println()

	interfaceBasics()
	implicitSatisfaction()
	emptyInterface()
	typeAssertions()
	typeSwitches()
	interfaceComposition()
	stdlibInterfaces()
	nilInterfaceGotcha()
	designPrinciples()
}

// =============================================================================
// PART 1: Interface Basics
// =============================================================================

// An interface defines a SET OF METHODS (a behavior contract).
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Stringer: anything that can present itself as a string
type Stringer interface {
	String() string
}

// ─── Circle satisfies Shape ───
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }
func (c Circle) String() string     { return fmt.Sprintf("Circle(r=%.1f)", c.Radius) }

// ─── Rect satisfies Shape ───
type Rect struct {
	W, H float64
}

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }
func (r Rect) String() string     { return fmt.Sprintf("Rect(%.1f×%.1f)", r.W, r.H) }

func interfaceBasics() {
	fmt.Println("--- INTERFACE BASICS ---")

	// An interface variable holds (type, value) pair
	var s Shape

	s = Circle{Radius: 5}
	fmt.Printf("  %v: area=%.2f, perimeter=%.2f\n", s, s.Area(), s.Perimeter())

	s = Rect{W: 4, H: 3}
	fmt.Printf("  %v: area=%.2f, perimeter=%.2f\n", s, s.Area(), s.Perimeter())

	// ─── Polymorphism: function accepting interface ───
	shapes := []Shape{
		Circle{Radius: 3},
		Rect{W: 5, H: 2},
		Circle{Radius: 1},
	}
	fmt.Printf("  Total area: %.2f\n", totalArea(shapes))

	fmt.Println()
}

func totalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// =============================================================================
// PART 2: Implicit Satisfaction — No "implements"
// =============================================================================
func implicitSatisfaction() {
	fmt.Println("--- IMPLICIT SATISFACTION ---")

	// In Go: NO "implements" keyword. If a type has the right methods,
	// it satisfies the interface automatically.
	//
	// In Java:  class Dog implements Animal { ... }  // explicit
	// In Go:    type Dog struct { ... }               // just add the methods
	//           func (d Dog) Speak() string { ... }   // satisfies Animal implicitly

	// This has HUGE implications:
	// 1. The Shape interface could be defined AFTER Circle and Rect
	// 2. Circle doesn't even need to KNOW about Shape
	// 3. You can define interfaces for types from OTHER packages
	//    (even stdlib types) without modifying them
	//
	// Example: define our own interface that strings.Builder satisfies:
	type Writer interface {
		WriteString(s string) (int, error)
	}

	// strings.Builder already has WriteString — it satisfies Writer!
	var w Writer = &strings.Builder{}
	w.WriteString("hello")
	fmt.Printf("  strings.Builder satisfies our Writer: %q\n",
		w.(*strings.Builder).String())

	// This is how io.Reader works:
	// io.Reader is defined in the io package.
	// os.File, bytes.Buffer, strings.Reader all satisfy it
	// WITHOUT importing or referencing io.Reader.

	fmt.Println("  No 'implements': methods are the contract")
	fmt.Println("  You can define interfaces for types from other packages")
	fmt.Println()
}

// =============================================================================
// PART 3: Empty Interface (any)
// =============================================================================
func emptyInterface() {
	fmt.Println("--- EMPTY INTERFACE ---")

	// interface{} (or `any` since Go 1.18) has zero methods.
	// EVERY type satisfies it — it can hold any value.
	// It's Go's equivalent of Object in Java or void* in C.

	var x any // same as interface{}

	x = 42
	fmt.Printf("  any = int: %v (type: %T)\n", x, x)

	x = "hello"
	fmt.Printf("  any = string: %v (type: %T)\n", x, x)

	x = []int{1, 2, 3}
	fmt.Printf("  any = []int: %v (type: %T)\n", x, x)

	// fmt.Println uses ...any to accept anything:
	// func Println(a ...any) (n int, err error)

	// ─── You lose type safety with any ───
	// Can't do x + 1 (compiler doesn't know the type)
	// Must type-assert to use the value
	// RULE: Use `any` sparingly. Prefer specific interfaces or generics.

	// ─── Practical use: heterogeneous collections ───
	mixed := []any{1, "two", 3.0, true, nil}
	for _, v := range mixed {
		fmt.Printf("    %v (%T)\n", v, v)
	}

	fmt.Println()
}

// =============================================================================
// PART 4: Type Assertions
// =============================================================================
func typeAssertions() {
	fmt.Println("--- TYPE ASSERTIONS ---")

	// Type assertion extracts the concrete value from an interface.
	// Syntax: value := interfaceVar.(ConcreteType)

	var s Shape = Circle{Radius: 5}

	// ─── Unsafe assertion (panics if wrong type) ───
	c := s.(Circle)
	fmt.Printf("  Unsafe assertion: Circle{Radius: %.1f}\n", c.Radius)

	// ─── Safe assertion with comma-ok ───
	if rect, ok := s.(Rect); ok {
		fmt.Printf("  It's a Rect: %v\n", rect)
	} else {
		fmt.Println("  It's NOT a Rect")
	}

	// ─── Assert to interface (capability check) ───
	// "Does this Shape also implement Stringer?"
	if str, ok := s.(Stringer); ok {
		fmt.Printf("  Also Stringer: %q\n", str.String())
	}

	// ─── PANIC example ───
	// r := s.(Rect) // PANIC: interface conversion: Shape is Circle, not Rect
	// Always use comma-ok form unless you're 100% sure of the type

	fmt.Println()
}

// =============================================================================
// PART 5: Type Switches — Polymorphic Dispatch
// =============================================================================

func describeShape(s Shape) string {
	switch v := s.(type) {
	case Circle:
		return fmt.Sprintf("Circle with radius %.1f", v.Radius)
	case Rect:
		return fmt.Sprintf("Rectangle %.1f×%.1f", v.W, v.H)
	default:
		return fmt.Sprintf("Unknown shape: %T", v)
	}
}

func describeAny(x any) string {
	switch v := x.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int(%d)", v)
	case string:
		return fmt.Sprintf("string(%q)", v)
	case bool:
		return fmt.Sprintf("bool(%v)", v)
	case error:
		return fmt.Sprintf("error(%v)", v)
	case fmt.Stringer:
		return fmt.Sprintf("Stringer(%s)", v.String())
	default:
		return fmt.Sprintf("%T(%v)", v, v)
	}
}

func typeSwitches() {
	fmt.Println("--- TYPE SWITCHES ---")

	shapes := []Shape{Circle{3}, Rect{4, 5}}
	for _, s := range shapes {
		fmt.Printf("  %s\n", describeShape(s))
	}

	// Type switch on any
	values := []any{42, "hello", true, nil, Circle{1}, fmt.Errorf("oops")}
	for _, v := range values {
		fmt.Printf("  %s\n", describeAny(v))
	}

	// NOTE: case order matters for interfaces
	// `error` and `fmt.Stringer` are both interfaces
	// If a type satisfies both, the first matching case wins

	fmt.Println()
}

// =============================================================================
// PART 6: Interface Embedding & Composition
// =============================================================================

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer2 interface {
	Write(p []byte) (n int, err error)
}

// ReadWriter embeds both — requires both Read AND Write
type ReadWriter interface {
	Reader
	Writer2
}

// Closer as a separate behavior
type Closer interface {
	Close() error
}

// ReadWriteCloser composes three interfaces
type ReadWriteCloser interface {
	Reader
	Writer2
	Closer
}

func interfaceComposition() {
	fmt.Println("--- INTERFACE COMPOSITION ---")

	// This is exactly how the standard library works:
	// io.Reader     → Read()
	// io.Writer     → Write()
	// io.Closer     → Close()
	// io.ReadWriter → Read() + Write()
	// io.ReadCloser → Read() + Close()
	// io.ReadWriteCloser → Read() + Write() + Close()
	//
	// Small interfaces compose into larger ones.
	// This is MUCH more flexible than class hierarchies.

	fmt.Println("  Small interfaces compose: Reader + Writer = ReadWriter")
	fmt.Println("  stdlib: io.Reader (1 method) → io.ReadWriteCloser (3 methods)")
	fmt.Println("  Prefer many small interfaces over few large ones")

	fmt.Println()
}

// =============================================================================
// PART 7: Standard Library Interfaces Everyone Must Know
// =============================================================================
func stdlibInterfaces() {
	fmt.Println("--- STDLIB INTERFACES ---")

	// INTERFACES EVERY GO DEV MUST KNOW:
	//
	// io.Reader      — Read(p []byte) (n int, err error)
	//                   The most important interface in Go.
	//                   Files, network, HTTP bodies, compressed streams...
	//
	// io.Writer      — Write(p []byte) (n int, err error)
	//                   Counter part to Reader.
	//
	// io.Closer      — Close() error
	//                   Everything that needs cleanup.
	//
	// error          — Error() string
	//                   Built-in interface. ALL errors implement this.
	//
	// fmt.Stringer   — String() string
	//                   How a type presents itself in fmt.Print/Sprintf.
	//
	// sort.Interface — Len() int; Less(i, j int) bool; Swap(i, j int)
	//                   Make any collection sortable.
	//
	// encoding.BinaryMarshaler   — MarshalBinary() ([]byte, error)
	// encoding.BinaryUnmarshaler — UnmarshalBinary(data []byte) error
	//
	// json.Marshaler   — MarshalJSON() ([]byte, error)
	// json.Unmarshaler — UnmarshalJSON(data []byte) error
	//
	// http.Handler     — ServeHTTP(ResponseWriter, *Request)
	//                    The foundation of all Go HTTP servers.
	//
	// context.Context  — Deadline, Done, Err, Value
	//                    Cancellation and request-scoped values.
	//
	// THE error INTERFACE (built-in, not in any package):
	// type error interface {
	//     Error() string
	// }

	// Demonstrate Stringer
	c := Circle{Radius: 5}
	fmt.Printf("  Stringer: %s\n", c) // calls c.String()

	// Demonstrate error interface
	err := fmt.Errorf("something went wrong")
	fmt.Printf("  error: %s\n", err)

	fmt.Println()
}

// =============================================================================
// PART 8: The nil Interface Gotcha
// =============================================================================
func nilInterfaceGotcha() {
	fmt.Println("--- NIL INTERFACE GOTCHA ---")

	// An interface value is nil ONLY when both type AND value are nil.
	//
	// Interface internally: (type, value)
	// nil interface:         (nil, nil)     → == nil is TRUE
	// non-nil with nil val:  (*int, nil)    → == nil is FALSE!

	// ─── Case 1: truly nil interface ───
	var s Shape // (nil, nil) → nil
	fmt.Printf("  nil interface: %v, == nil: %v\n", s, s == nil)

	// ─── Case 2: interface holding a nil pointer (THE GOTCHA) ───
	var c *Circle = nil // c is a nil pointer
	var s2 Shape = c    // s2 = (*Circle, nil) — NOT nil!
	fmt.Printf("  Interface with nil value: %v, == nil: %v (SURPRISE!)\n", s2, s2 == nil)

	// This bites people in error handling:
	err := errorFunction()
	fmt.Printf("  Error == nil: %v (even though return was nil!)\n", err == nil)
	// The function returns (*MyError)(nil), which becomes (error = (*MyError, nil))
	// That is NOT nil!

	// ─── FIX: never assign typed nil to interface ───
	err2 := errorFunctionFixed()
	fmt.Printf("  Fixed == nil: %v\n", err2 == nil)

	fmt.Println()
}

type MyError struct {
	Msg string
}

func (e *MyError) Error() string { return e.Msg }

func errorFunction() error {
	var err *MyError = nil
	return err // returns (*MyError, nil) — NOT a nil error!
}

func errorFunctionFixed() error {
	var err *MyError = nil
	if err != nil {
		return err
	}
	return nil // return an untyped nil — this IS a nil error
}

// =============================================================================
// PART 9: Interface Design Principles
// =============================================================================
func designPrinciples() {
	fmt.Println("--- INTERFACE DESIGN PRINCIPLES ---")

	// ─── PRINCIPLE 1: Accept interfaces, return structs ───
	// Functions should accept the smallest interface they need.
	// DON'T:  func Process(f *os.File) { ... }
	// DO:     func Process(r io.Reader) { ... }
	//
	// Return concrete types so callers know exactly what they get.
	// DON'T:  func NewServer() HTTPHandler { ... }
	// DO:     func NewServer() *Server { ... }
	fmt.Println("  1. Accept interfaces, return structs")

	// ─── PRINCIPLE 2: Keep interfaces small ───
	// The bigger an interface, the weaker the abstraction.
	// — Rob Pike, Go Proverbs
	//
	// io.Reader: 1 method → used everywhere
	// io.ReadWriteCloser: 3 methods → used less
	// Large interface with 20 methods → basically useless
	fmt.Println("  2. Small interfaces (1-3 methods)")

	// ─── PRINCIPLE 3: Define interfaces at the consumer, not producer ───
	// The CONSUMER knows what behavior it needs.
	// // In package "processor":
	// type DataSource interface {
	//     Fetch(ctx context.Context) ([]byte, error)
	// }
	// func Process(src DataSource) { ... }
	//
	// // In package "httpclient":
	// type Client struct { ... }
	// func (c *Client) Fetch(ctx context.Context) ([]byte, error) { ... }
	// // Client doesn't know about DataSource. It just has the right method.
	fmt.Println("  3. Define interfaces at the consumer (not producer)")

	// ─── PRINCIPLE 4: Don't export interfaces for implementation ───
	// DON'T: type UserService interface { ... } (exported for others to implement)
	// DO:    type UserService struct { ... } (exported struct)
	//        Use interfaces where CONSUMED, not where produced.
	fmt.Println("  4. Don't export interfaces for others to implement")

	// ─── PRINCIPLE 5: Composition over declaration ───
	// Compose small interfaces into larger ones as needed:
	// type ReadWriter interface { Reader; Writer }
	// Don't pre-combine every possible combination.
	fmt.Println("  5. Compose small interfaces as needed")

	fmt.Println()
}
