//go:build ignore

// =============================================================================
// LESSON 0.2: CONTROL FLOW — Every Way to Branch and Loop in Go
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - if/else with init statements (Go's secret weapon)
// - switch: expression switch, type switch, tagless switch
// - for: the ONLY loop keyword (but it does everything)
// - range: iterating over everything (string, slice, map, channel)
// - break, continue, goto, labels — and when they're appropriate
// - defer: execution order, closures, stack behavior
//
// THE KEY INSIGHT:
// Go has fewer keywords than most languages but each one is more powerful.
// There's only ONE loop keyword (for) that handles while, do-while, foreach,
// and infinite loops. switch doesn't fall through by default (saving you from
// one of C's most common bugs). And if can have init statements.
//
// RUN: go run 02_control_flow.go
// =============================================================================

package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
)

func main() {
	fmt.Println("=== CONTROL FLOW ===")
	fmt.Println()

	ifElseMastery()
	switchMastery()
	forLoopMastery()
	rangeMastery()
	breakContinueLabels()
	deferMastery()
}

// =============================================================================
// PART 1: if/else — More Powerful Than You Think
// =============================================================================
func ifElseMastery() {
	fmt.Println("--- IF/ELSE ---")

	x := 42

	// ─── Basic if/else ───
	if x > 0 {
		fmt.Println("  positive")
	} else if x < 0 {
		fmt.Println("  negative")
	} else {
		fmt.Println("  zero")
	}

	// ─── if with init statement (Go's killer feature) ───
	// Variables declared in init are scoped to the if/else block.
	// This is THE idiomatic Go pattern for error handling.
	if n := x * 2; n > 50 {
		fmt.Printf("  init statement: n=%d > 50\n", n)
	} else {
		fmt.Printf("  init statement: n=%d <= 50\n", n)
	}
	// n is NOT accessible here — scoped to if/else block

	// ─── THE MOST COMMON GO PATTERN: if err != nil ───
	if _, err := fmt.Fprintf(os.Stdout, "  error check: "); err != nil {
		fmt.Println("write failed:", err)
		return
	}
	fmt.Println("success")

	// WHY init statement matters:
	// Without:
	//   err := doSomething()
	//   if err != nil { ... }
	//   // err still in scope, might accidentally reuse it
	//
	// With:
	//   if err := doSomething(); err != nil { ... }
	//   // err is gone, clean scope

	// ─── Go has NO ternary operator ───
	// No: result = condition ? a : b
	// Use: if/else (explicit, readable)
	// Or for simple cases:
	result := "even"
	if x%2 != 0 {
		result = "odd"
	}
	fmt.Printf("  no ternary: %d is %s\n", x, result)

	// ─── Truthy/falsy: Go doesn't have it ───
	// if 1 { }          // COMPILE ERROR
	// if "" { }         // COMPILE ERROR
	// if nil { }        // COMPILE ERROR
	// Must use explicit boolean expressions:
	// if n != 0 { }     // correct
	// if s != "" { }    // correct
	// if p != nil { }   // correct

	fmt.Println()
}

// =============================================================================
// PART 2: switch — Three Flavors, All Powerful
// =============================================================================
func switchMastery() {
	fmt.Println("--- SWITCH ---")

	// ─── FLAVOR 1: Expression switch ───
	day := "Tuesday"
	switch day {
	case "Monday":
		fmt.Println("  Monday")
	case "Tuesday", "Wednesday", "Thursday": // multiple values per case!
		fmt.Printf("  Midweek: %s\n", day)
	case "Friday":
		fmt.Println("  TGIF!")
	default:
		fmt.Println("  Weekend!")
	}

	// KEY DIFFERENCE FROM C:
	// - NO automatic fallthrough (no break needed!)
	// - In C: every case falls through without break
	// - In Go: each case breaks automatically

	// ─── Explicit fallthrough ───
	n := 3
	switch {
	case n > 0:
		fmt.Print("  positive")
		fallthrough // continues to next case body (skips condition check!)
	case n > -5:
		fmt.Println(" and greater than -5")
	}
	// ⚠️ fallthrough is RARE in production Go code

	// ─── switch with init statement ───
	switch os := runtime.GOOS; os {
	case "darwin":
		fmt.Println("  macOS")
	case "linux":
		fmt.Println("  Linux")
	default:
		fmt.Printf("  Other OS: %s\n", os)
	}

	// ─── FLAVOR 2: Tagless switch (cleaner than if/else chains) ───
	// No expression after `switch` — each case evaluates a boolean
	score := 85
	switch {
	case score >= 90:
		fmt.Println("  Grade: A")
	case score >= 80:
		fmt.Println("  Grade: B")
	case score >= 70:
		fmt.Println("  Grade: C")
	default:
		fmt.Println("  Grade: F")
	}

	// ─── FLAVOR 3: Type switch (key for interfaces) ───
	var value interface{} = "hello"
	switch v := value.(type) {
	case int:
		fmt.Printf("  int: %d\n", v)
	case string:
		fmt.Printf("  string: %q (len=%d)\n", v, len(v))
	case bool:
		fmt.Printf("  bool: %v\n", v)
	case nil:
		fmt.Println("  nil!")
	default:
		fmt.Printf("  unknown type: %T\n", v)
	}

	// Type switch is how you do polymorphic dispatch in Go.
	// Used extensively with error handling, JSON parsing, AST walking.

	fmt.Println()
}

// =============================================================================
// PART 3: for — The ONLY Loop (But It Does Everything)
// =============================================================================
func forLoopMastery() {
	fmt.Println("--- FOR LOOP ---")

	// Go has ONE loop keyword: `for`. It replaces while, do-while, foreach.

	// ─── Classic for (C-style) ───
	fmt.Print("  C-style: ")
	for i := 0; i < 5; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// ─── while-style (condition only) ───
	fmt.Print("  While-style: ")
	n := 1
	for n < 32 {
		fmt.Print(n, " ")
		n *= 2
	}
	fmt.Println()

	// ─── Infinite loop ───
	// for { ... }  is Go's infinite loop (cleaner than `while(true)`)
	fmt.Print("  Infinite: ")
	count := 0
	for {
		if count >= 3 {
			break
		}
		fmt.Print(count, " ")
		count++
	}
	fmt.Println()

	// ─── do-while equivalent ───
	// Go doesn't have do-while, but you can simulate it:
	fmt.Print("  Do-while: ")
	i := 0
	for {
		fmt.Print(i, " ")
		i++
		if i >= 3 {
			break
		}
	}
	fmt.Println()

	// ─── Multiple variables in for ───
	fmt.Print("  Multi-var: ")
	for i, j := 0, 10; i < j; i, j = i+1, j-1 {
		fmt.Printf("(%d,%d) ", i, j)
	}
	fmt.Println()

	fmt.Println()
}

// =============================================================================
// PART 4: range — Iterate Over Everything
// =============================================================================
func rangeMastery() {
	fmt.Println("--- RANGE ---")

	// range works on: string, array, slice, map, channel

	// ─── range over slice: index + value ───
	fruits := []string{"apple", "banana", "cherry"}
	fmt.Print("  Slice: ")
	for i, v := range fruits {
		fmt.Printf("[%d:%s] ", i, v)
	}
	fmt.Println()

	// ─── range index only (ignore value) ───
	fmt.Print("  Index only: ")
	for i := range fruits {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// ─── range value only (ignore index) ───
	fmt.Print("  Value only: ")
	for _, v := range fruits {
		fmt.Print(v, " ")
	}
	fmt.Println()

	// ─── range over string: iterates RUNES (not bytes!) ───
	fmt.Print("  String runes: ")
	for i, r := range "Go🚀" {
		fmt.Printf("[%d:%c] ", i, r) // i is byte offset, r is rune
	}
	fmt.Println()
	// Note: byte offsets may skip numbers for multi-byte runes

	// ─── range over map: key + value (RANDOM ORDER!) ───
	ages := map[string]int{"Alice": 30, "Bob": 25, "Charlie": 35}
	fmt.Print("  Map: ")
	for k, v := range ages {
		fmt.Printf("%s=%d ", k, v)
	}
	fmt.Println("(random order!)")

	// ─── range over channel: blocks until closed ───
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	ch <- 30
	close(ch)
	fmt.Print("  Channel: ")
	for v := range ch {
		fmt.Print(v, " ")
	}
	fmt.Println()

	// ─── range over integer (Go 1.22+) ───
	fmt.Print("  Integer (1.22+): ")
	for i := range 5 {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// ─── GOTCHA: range creates a COPY of the value ───
	type Item struct{ Name string }
	items := []Item{{"a"}, {"b"}, {"c"}}
	for _, item := range items {
		item.Name = "modified" // modifies the COPY, not the original!
	}
	fmt.Printf("  Range copy gotcha: %v (unchanged!)\n", items)

	// FIX: use index to modify in place:
	for i := range items {
		items[i].Name = strings.ToUpper(items[i].Name)
	}
	fmt.Printf("  Fixed with index: %v\n", items)

	fmt.Println()
}

// =============================================================================
// PART 5: break, continue, Labels — Nested Loop Control
// =============================================================================
func breakContinueLabels() {
	fmt.Println("--- BREAK, CONTINUE, LABELS ---")

	// ─── break: exit the innermost loop ───
	fmt.Print("  break: ")
	for i := 0; i < 10; i++ {
		if i == 5 {
			break
		}
		fmt.Print(i, " ")
	}
	fmt.Println()

	// ─── continue: skip to next iteration ───
	fmt.Print("  continue (skip even): ")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			continue
		}
		fmt.Print(i, " ")
	}
	fmt.Println()

	// ─── Labels: break/continue on an OUTER loop ───
	// Without labels, break only exits the innermost loop.
	// Labels let you break out of nested loops.
	fmt.Print("  Labeled break: ")
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i == 1 && j == 1 {
				break outer // breaks the OUTER loop
			}
			fmt.Printf("(%d,%d) ", i, j)
		}
	}
	fmt.Println()

	// ─── Labels with continue ───
	fmt.Print("  Labeled continue: ")
loop:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if j == 1 {
				continue loop // continues the OUTER loop
			}
			fmt.Printf("(%d,%d) ", i, j)
		}
	}
	fmt.Println()

	// ─── Labels with switch (break out of switch inside a for) ───
	fmt.Print("  Switch in loop: ")
	nums := []int{1, 2, -1, 3}
search:
	for _, n := range nums {
		switch {
		case n < 0:
			fmt.Print("found negative! ")
			break search // without label, break only exits the switch
		default:
			fmt.Printf("%d ", n)
		}
	}
	fmt.Println()

	// ─── goto ───
	// Go HAS goto, but it's rarely used. It cannot jump over variable declarations.
	// Legitimate use: cleanup code in complex state machines (very rare).
	// In 99.9% of code: use break/continue/return instead.
	i := 0
	fmt.Print("  goto: ")
again:
	if i < 3 {
		fmt.Print(i, " ")
		i++
		goto again
	}
	fmt.Println("(please don't do this)")

	fmt.Println()
}

// =============================================================================
// PART 6: defer — Execute on Function Exit
// =============================================================================
func deferMastery() {
	fmt.Println("--- DEFER ---")

	// defer schedules a function call to run when the ENCLOSING FUNCTION returns.
	// NOT when the block/scope exits — when the FUNCTION exits.
	//
	// USE FOR:
	// - Closing files, connections, locks
	// - Cleanup resources
	// - Unlocking mutexes
	// - Recovering from panics

	// ─── Basic defer ───
	fmt.Println("  Defer order:")
	fmt.Println("    first")
	defer fmt.Println("    deferred 1 (runs last)")
	defer fmt.Println("    deferred 2 (runs second-to-last)")
	defer fmt.Println("    deferred 3 (runs first of defers)")
	fmt.Println("    last regular statement")

	// DEFER IS LIFO (Last In, First Out) — stack order!
	// Output: first → last regular → deferred 3 → deferred 2 → deferred 1

	// ─── Arguments are evaluated IMMEDIATELY ───
	x := 10
	defer fmt.Printf("    deferred x = %d (captured at defer time, not at exit)\n", x)
	x = 99 // deferred call still sees 10, not 99!

	// ─── COMMON PATTERN: file handling ───
	// f, err := os.Open("file.txt")
	// if err != nil { return err }
	// defer f.Close()  // guaranteed to close even if later code panics
	// ... use f ...

	// ─── COMMON PATTERN: mutex unlock ───
	// mu.Lock()
	// defer mu.Unlock()  // guaranteed to unlock
	// ... critical section ...

	// ─── GOTCHA: defer in a loop ───
	// Defers don't run until the function exits!
	// In a loop, ALL defers accumulate and run at the end:
	//
	// for _, name := range files {
	//     f, _ := os.Open(name)
	//     defer f.Close()  // ⚠️ ALL files stay open until function returns!
	// }
	//
	// FIX: wrap in a closure or extract to a function:
	// for _, name := range files {
	//     func() {
	//         f, _ := os.Open(name)
	//         defer f.Close()
	//         // process f — f.Close() runs at end of THIS anonymous function
	//     }()
	// }

	// ─── GOTCHA: defer with named return values ───
	result := namedReturnDefer()
	fmt.Printf("    named return with defer: %d (defer modified it!)\n", result)

	// ─── defer for panic recovery (see error handling lesson) ───
	// defer func() {
	//     if r := recover(); r != nil {
	//         fmt.Println("recovered:", r)
	//     }
	// }()

	fmt.Println()
}

// Named return + defer: defer can READ and MODIFY named return values
func namedReturnDefer() (result int) {
	defer func() {
		result *= 2 // modifies the named return value!
	}()
	return 21 // sets result = 21, then defer runs: result = 42
}

// Helper for random number
func init() {
	_ = rand.Int() // ensure rand is used
}
