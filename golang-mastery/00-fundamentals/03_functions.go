//go:build ignore

// =============================================================================
// LESSON 0.3: FUNCTIONS — First-Class Citizens in Go
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Function signatures: parameters, returns, variadic
// - Multiple return values (Go's hallmark pattern)
// - Named return values (and when to use them)
// - Functions as values (first-class functions)
// - Closures: capturing variables from outer scope
// - Anonymous functions and immediately invoked functions
// - init() functions: package initialization order
// - Recursion and tail calls (Go doesn't optimize them)
//
// THE KEY INSIGHT:
// Functions in Go are first-class values — they can be assigned to variables,
// passed as arguments, and returned from other functions. Combined with
// multiple return values (value, error), this creates Go's distinctive
// style where error handling is explicit and composition is via functions,
// not class hierarchies.
//
// RUN: go run 03_functions.go
// =============================================================================

package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

func main() {
	fmt.Println("=== FUNCTIONS ===")
	fmt.Println()

	basicFunctions()
	multipleReturns()
	namedReturns()
	variadicFunctions()
	functionsAsValues()
	closures()
	anonymousFunctions()
	higherOrderFunctions()
	recursion()
	initFunctions()
}

// =============================================================================
// PART 1: Basic Function Syntax
// =============================================================================

// Parameters with same type can share the type declaration
func add(a, b int) int {
	return a + b
}

// Explicit parameter types when they differ
func greet(name string, times int) string {
	return fmt.Sprintf("Hello %s! (×%d)", name, times)
}

// No return value
func logMessage(msg string) {
	fmt.Printf("  LOG: %s\n", msg)
}

func basicFunctions() {
	fmt.Println("--- BASIC FUNCTIONS ---")

	fmt.Printf("  add(3, 4) = %d\n", add(3, 4))
	fmt.Printf("  greet: %s\n", greet("Vikram", 3))
	logMessage("functions are simple")

	// ─── All arguments are passed BY VALUE ───
	// Go COPIES the argument. Modifying inside the function doesn't affect caller.
	x := 10
	doubleValue(x)
	fmt.Printf("  Pass by value: x = %d (unchanged after doubleValue)\n", x)
	// To modify: pass a pointer (covered in 05_pointers.go)

	fmt.Println()
}

func doubleValue(n int) {
	n *= 2 // modifies local copy only
}

// =============================================================================
// PART 2: Multiple Return Values
// =============================================================================

// The (value, error) pattern — most common in Go
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

// Multiple values for decomposition
func minMax(nums []int) (int, int) {
	if len(nums) == 0 {
		return 0, 0
	}
	min, max := nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return min, max
}

// Return a value and a boolean (comma-ok pattern)
func lookup(key string) (string, bool) {
	m := map[string]string{
		"name": "Vikram",
		"lang": "Go",
	}
	val, ok := m[key]
	return val, ok
}

func multipleReturns() {
	fmt.Println("--- MULTIPLE RETURNS ---")

	// (value, error) pattern
	result, err := divide(10, 3)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  10/3 = %.4f\n", result)
	}

	_, err = divide(10, 0)
	fmt.Printf("  10/0 error: %v\n", err)

	// Multiple values
	min, max := minMax([]int{3, 1, 4, 1, 5, 9, 2, 6})
	fmt.Printf("  min=%d, max=%d\n", min, max)

	// Comma-ok pattern (like map lookups)
	if val, ok := lookup("name"); ok {
		fmt.Printf("  Found: %q\n", val)
	}
	if _, ok := lookup("missing"); !ok {
		fmt.Println("  Not found (comma-ok)")
	}

	fmt.Println()
}

// =============================================================================
// PART 3: Named Return Values
// =============================================================================

// Named returns: variables pre-declared, "naked return" sets them
func rectArea(width, height float64) (area float64, perimeter float64) {
	area = width * height
	perimeter = 2 * (width + height)
	return // naked return — returns named variables
}

// Named returns are BEST for: documenting what's returned
func parseEndpoint(raw string) (host string, port string, err error) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid endpoint: %q", raw)
		return // naked return with error
	}
	host = parts[0]
	port = parts[1]
	return
}

func namedReturns() {
	fmt.Println("--- NAMED RETURNS ---")

	area, perim := rectArea(5, 3)
	fmt.Printf("  Rectangle 5×3: area=%.0f, perimeter=%.0f\n", area, perim)

	host, port, err := parseEndpoint("localhost:8080")
	fmt.Printf("  Parse endpoint: host=%q, port=%q, err=%v\n", host, port, err)

	_, _, err = parseEndpoint("invalid")
	fmt.Printf("  Parse invalid: err=%v\n", err)

	// ─── WHEN TO USE NAMED RETURNS ───
	// ✅ Document what the function returns (self-documenting signature)
	// ✅ Short functions where naked return is clear
	// ✅ With defer to modify return values (see control flow lesson)
	// ❌ Long functions — naked return is unclear, reader must scroll up
	// ❌ When names don't add clarity (func (a, b int) (int, error))
	//
	// RULE: Named returns for documentation, explicit return when possible.

	fmt.Println()
}

// =============================================================================
// PART 4: Variadic Functions
// =============================================================================

// Variadic: ... before the type means "zero or more" arguments
// The parameter is received as a slice
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Variadic must be the LAST parameter
func logWithPrefix(prefix string, messages ...string) {
	for _, msg := range messages {
		fmt.Printf("  [%s] %s\n", prefix, msg)
	}
}

func variadicFunctions() {
	fmt.Println("--- VARIADIC FUNCTIONS ---")

	// Call with individual arguments
	fmt.Printf("  sum(1,2,3) = %d\n", sum(1, 2, 3))
	fmt.Printf("  sum() = %d\n", sum()) // zero arguments OK

	// Spread a slice with ...
	numbers := []int{10, 20, 30, 40}
	fmt.Printf("  sum(slice...) = %d\n", sum(numbers...))

	logWithPrefix("INFO", "server started", "listening on :8080")

	// ─── fmt.Println is variadic: func Println(a ...interface{}) ───
	// That's why it accepts any number of any type!

	// ─── GOTCHA: spreading different slices ───
	a := []int{1, 2}
	b := []int{3, 4}
	// sum(a..., b...) // COMPILE ERROR! Can only spread one slice
	// FIX: combine first
	combined := append(a, b...)
	fmt.Printf("  combined sum = %d\n", sum(combined...))

	fmt.Println()
}

// =============================================================================
// PART 5: Functions as Values (First-Class Functions)
// =============================================================================
func functionsAsValues() {
	fmt.Println("--- FUNCTIONS AS VALUES ---")

	// Functions are values — assign to variables, pass around
	var op func(int, int) int
	op = add
	fmt.Printf("  func variable: op(3,4) = %d\n", op(3, 4))

	// Function type can be named
	type MathFunc func(float64, float64) float64

	var compute MathFunc = math.Pow
	fmt.Printf("  Named func type: pow(2,10) = %.0f\n", compute(2, 10))

	// Functions in a map (dispatch table / strategy pattern)
	operations := map[string]func(int, int) int{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
	}

	for name, fn := range operations {
		fmt.Printf("  %s(10, 3) = %d\n", name, fn(10, 3))
	}

	// Functions in a slice
	pipeline := []func(string) string{
		strings.TrimSpace,
		strings.ToLower,
		func(s string) string { return strings.ReplaceAll(s, " ", "-") },
	}

	input := "  Hello World  "
	for _, fn := range pipeline {
		input = fn(input)
	}
	fmt.Printf("  Pipeline: %q\n", input) // "hello-world"

	fmt.Println()
}

// =============================================================================
// PART 6: Closures — Functions That Remember
// =============================================================================
func closures() {
	fmt.Println("--- CLOSURES ---")

	// A closure is a function that captures variables from its surrounding scope.
	// The captured variables survive as long as the closure exists.

	// ─── Counter factory ───
	makeCounter := func() func() int {
		count := 0 // captured by the returned function
		return func() int {
			count++ // modifies the captured variable
			return count
		}
	}

	counter1 := makeCounter()
	counter2 := makeCounter() // separate counter, separate `count`

	fmt.Printf("  counter1: %d, %d, %d\n", counter1(), counter1(), counter1())
	fmt.Printf("  counter2: %d, %d\n", counter2(), counter2()) // independent!

	// ─── Closure captures the VARIABLE, not the value ───
	x := 10
	modify := func() {
		x = 99 // modifies the SAME x
	}
	modify()
	fmt.Printf("  Closure modified x: %d\n", x) // 99

	// ─── CLASSIC GOTCHA: closure in a loop ───
	funcs := make([]func(), 5)
	for i := 0; i < 5; i++ {
		funcs[i] = func() {
			fmt.Print(i, " ") // captures the variable i, not its current value
		}
	}
	fmt.Print("  Loop gotcha: ")
	for _, f := range funcs {
		f() // Go 1.22+ prints 0 1 2 3 4 (fixed!)
		// Before Go 1.22: would print 5 5 5 5 5 (all capture the same i)
	}
	fmt.Println()

	// Pre-1.22 fix: capture by parameter
	funcs2 := make([]func(), 5)
	for i := 0; i < 5; i++ {
		i := i // explicit re-declaration (shadow i with a new i per iteration)
		funcs2[i] = func() {
			fmt.Print(i, " ")
		}
	}
	fmt.Print("  Loop fixed: ")
	for _, f := range funcs2 {
		f()
	}
	fmt.Println()

	// ─── Closure for middleware / decorator ───
	withLogging := func(name string, fn func(int, int) int) func(int, int) int {
		return func(a, b int) int {
			result := fn(a, b)
			fmt.Printf("  %s(%d, %d) = %d\n", name, a, b, result)
			return result
		}
	}

	loggedAdd := withLogging("add", add)
	loggedAdd(5, 3)

	fmt.Println()
}

// =============================================================================
// PART 7: Anonymous Functions
// =============================================================================
func anonymousFunctions() {
	fmt.Println("--- ANONYMOUS FUNCTIONS ---")

	// ─── Inline anonymous function ───
	result := func(a, b int) int {
		return a * b
	}(6, 7) // immediately invoked!
	fmt.Printf("  IIFE: 6×7 = %d\n", result)

	// ─── Anonymous function for sort ───
	people := []struct {
		Name string
		Age  int
	}{
		{"Charlie", 30},
		{"Alice", 25},
		{"Bob", 35},
	}

	sort.Slice(people, func(i, j int) bool {
		return people[i].Age < people[j].Age
	})
	fmt.Printf("  Sorted by age: %v\n", people)

	// ─── Anonymous function for goroutine ───
	// go func() {
	//     fmt.Println("running in goroutine")
	// }()

	// ─── Anonymous function for defer ───
	// defer func() {
	//     if r := recover(); r != nil {
	//         fmt.Println("recovered:", r)
	//     }
	// }()

	fmt.Println()
}

// =============================================================================
// PART 8: Higher-Order Functions
// =============================================================================
func higherOrderFunctions() {
	fmt.Println("--- HIGHER-ORDER FUNCTIONS ---")

	// Higher-order function: takes a function as parameter or returns one

	// ─── apply: takes a function and applies it ───
	apply := func(nums []int, fn func(int) int) []int {
		result := make([]int, len(nums))
		for i, n := range nums {
			result[i] = fn(n)
		}
		return result
	}

	nums := []int{1, 2, 3, 4, 5}
	doubled := apply(nums, func(n int) int { return n * 2 })
	squared := apply(nums, func(n int) int { return n * n })
	fmt.Printf("  doubled: %v\n", doubled)
	fmt.Printf("  squared: %v\n", squared)

	// ─── filter: returns elements matching a predicate ───
	filter := func(nums []int, pred func(int) bool) []int {
		var result []int
		for _, n := range nums {
			if pred(n) {
				result = append(result, n)
			}
		}
		return result
	}

	evens := filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Printf("  evens: %v\n", evens)

	// ─── reduce: fold elements into a single value ───
	reduce := func(nums []int, initial int, fn func(int, int) int) int {
		acc := initial
		for _, n := range nums {
			acc = fn(acc, n)
		}
		return acc
	}

	total := reduce(nums, 0, func(acc, n int) int { return acc + n })
	product := reduce(nums, 1, func(acc, n int) int { return acc * n })
	fmt.Printf("  sum: %d, product: %d\n", total, product)

	// NOTE: Go 1.18+ generics make these type-safe for any type.
	// See 04-advanced-generics for generic Map/Filter/Reduce.

	fmt.Println()
}

// =============================================================================
// PART 9: Recursion
// =============================================================================
func recursion() {
	fmt.Println("--- RECURSION ---")

	// Go supports recursion but does NOT optimize tail calls.
	// Deep recursion can stack overflow. Prefer iteration for performance.

	var factorial func(n int) int
	factorial = func(n int) int {
		if n <= 1 {
			return 1
		}
		return n * factorial(n-1)
	}
	fmt.Printf("  factorial(10) = %d\n", factorial(10))

	// ─── Fibonacci (naive recursive) ───
	var fib func(n int) int
	fib = func(n int) int {
		if n <= 1 {
			return n
		}
		return fib(n-1) + fib(n-2)
	}
	fmt.Printf("  fib(10) = %d (naive, O(2^n))\n", fib(10))

	// ─── Fibonacci (iterative — preferred in Go) ───
	fibIter := func(n int) int {
		if n <= 1 {
			return n
		}
		a, b := 0, 1
		for i := 2; i <= n; i++ {
			a, b = b, a+b
		}
		return b
	}
	fmt.Printf("  fib(10) = %d (iterative, O(n))\n", fibIter(10))

	// ─── Recursive tree traversal (where recursion shines) ───
	type TreeNode struct {
		Value int
		Left  *TreeNode
		Right *TreeNode
	}

	root := &TreeNode{
		Value: 1,
		Left:  &TreeNode{Value: 2, Left: &TreeNode{Value: 4}, Right: &TreeNode{Value: 5}},
		Right: &TreeNode{Value: 3, Right: &TreeNode{Value: 6}},
	}

	var inorder func(*TreeNode)
	inorder = func(n *TreeNode) {
		if n == nil {
			return
		}
		inorder(n.Left)
		fmt.Print(n.Value, " ")
		inorder(n.Right)
	}
	fmt.Print("  Inorder traversal: ")
	inorder(root)
	fmt.Println()

	// RULE: Use recursion for tree/graph traversal.
	// Use iteration for everything else (Go has no tail call optimization).

	fmt.Println()
}

// =============================================================================
// PART 10: init() Functions
// =============================================================================

// init() runs automatically before main(), after package-level variables.
// Multiple init() per file are allowed (run in order of appearance).
// Multiple files: init order follows file name sort order.
//
// EXECUTION ORDER:
// 1. Package-level variables (in dependency order)
// 2. init() functions (all of them, in order)
// 3. main() function
//
// USE FOR:
// - Registering database drivers: import _ "github.com/lib/pq"
// - Setting up package-level state
// - Verifying program preconditions
//
// AVOID FOR:
// - Complex logic (hard to test, implicit side effects)
// - Starting goroutines
// - I/O operations

var initOrder []string

func init() {
	initOrder = append(initOrder, "first init")
}

func init() { // yes, multiple init() in one file is valid!
	initOrder = append(initOrder, "second init")
}

func initFunctions() {
	fmt.Println("--- INIT FUNCTIONS ---")
	fmt.Printf("  init order: %v\n", initOrder)
	fmt.Println("  init() runs before main(), used for setup/registration")
	fmt.Println("  Avoid complex logic in init() — hard to test")
	fmt.Println()
}
