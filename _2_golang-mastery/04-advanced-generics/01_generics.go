// =============================================================================
// LESSON 4: ADVANCED GENERICS — Beyond the Basics
// =============================================================================
//
// Go 1.18+ generics (type parameters) enable type-safe reusable code.
// This lesson covers advanced patterns: constraints, type sets,
// recursive types, monadic patterns, and real-world data structures.
//
// KEY INSIGHT: Go generics use CONSTRAINTS (interfaces with type sets),
// not templates (C++) or type erasure (Java). The compiler generates
// specialized code per type (monomorphization) OR uses dictionaries.
// =============================================================================

package main

import (
	"cmp"
	"fmt"
	"sync"
)

// =============================================================================
// PART 1: Custom Constraints
// =============================================================================

// Basic constraint: any type that can be added
type Addable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 |
		~string // string concatenation with +
}

// The ~ means "underlying type" — allows named types:
type Meters float64  // Meters has underlying type float64
type UserID int64    // UserID has underlying type int64

func Sum[T Addable](values []T) T {
	var total T
	for _, v := range values {
		total += v
	}
	return total
}

// Constraint with methods — any type that can describe itself
type Stringer interface {
	String() string
}

// Combine type sets with method requirements
type OrderedStringer interface {
	cmp.Ordered
	fmt.Stringer
}

// =============================================================================
// PART 2: Generic Data Structures
// =============================================================================

// --- Generic Result type (like Rust's Result<T, E>) ---
type Result[T any] struct {
	value T
	err   error
	ok    bool
}

func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, ok: true}
}

func Err[T any](err error) Result[T] {
	return Result[T]{err: err, ok: false}
}

func (r Result[T]) Unwrap() (T, error) {
	if r.ok {
		return r.value, nil
	}
	return r.value, r.err
}

func (r Result[T]) Map(fn func(T) T) Result[T] {
	if r.ok {
		return Ok(fn(r.value))
	}
	return r
}

func (r Result[T]) FlatMap(fn func(T) Result[T]) Result[T] {
	if r.ok {
		return fn(r.value)
	}
	return r
}

// --- Generic Optional type ---
type Optional[T any] struct {
	value T
	valid bool
}

func Some[T any](v T) Optional[T]   { return Optional[T]{value: v, valid: true} }
func None[T any]() Optional[T]      { return Optional[T]{} }

func (o Optional[T]) Get() (T, bool) { return o.value, o.valid }

func (o Optional[T]) OrElse(fallback T) T {
	if o.valid {
		return o.value
	}
	return fallback
}

func (o Optional[T]) Map(fn func(T) T) Optional[T] {
	if o.valid {
		return Some(fn(o.value))
	}
	return o
}

// --- Generic concurrent-safe Map ---
type SyncMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{m: make(map[K]V)}
}

func (sm *SyncMap[K, V]) Get(key K) (V, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	v, ok := sm.m[key]
	return v, ok
}

func (sm *SyncMap[K, V]) Set(key K, value V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m[key] = value
}

func (sm *SyncMap[K, V]) GetOrSet(key K, defaultVal V) V {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if v, ok := sm.m[key]; ok {
		return v
	}
	sm.m[key] = defaultVal
	return defaultVal
}

// --- Generic Binary Search Tree ---
type BST[T cmp.Ordered] struct {
	root *bstNode[T]
}

type bstNode[T cmp.Ordered] struct {
	value       T
	left, right *bstNode[T]
}

func (t *BST[T]) Insert(val T) {
	t.root = insertNode(t.root, val)
}

func insertNode[T cmp.Ordered](node *bstNode[T], val T) *bstNode[T] {
	if node == nil {
		return &bstNode[T]{value: val}
	}
	if val < node.value {
		node.left = insertNode(node.left, val)
	} else if val > node.value {
		node.right = insertNode(node.right, val)
	}
	return node
}

func (t *BST[T]) InOrder() []T {
	var result []T
	inOrderTraversal(t.root, &result)
	return result
}

func inOrderTraversal[T cmp.Ordered](node *bstNode[T], result *[]T) {
	if node == nil {
		return
	}
	inOrderTraversal(node.left, result)
	*result = append(*result, node.value)
	inOrderTraversal(node.right, result)
}

func (t *BST[T]) Search(val T) bool {
	return searchNode(t.root, val)
}

func searchNode[T cmp.Ordered](node *bstNode[T], val T) bool {
	if node == nil {
		return false
	}
	if val == node.value {
		return true
	}
	if val < node.value {
		return searchNode(node.left, val)
	}
	return searchNode(node.right, val)
}

// =============================================================================
// PART 3: Generic Functional Utilities
// =============================================================================

func Map[T any, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

func Reduce[T any, U any](slice []T, initial U, fn func(U, T) U) U {
	acc := initial
	for _, v := range slice {
		acc = fn(acc, v)
	}
	return acc
}

func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := keyFn(v)
		result[key] = append(result[key], v)
	}
	return result
}

func Zip[A any, B any](a []A, b []B) []struct {
	First  A
	Second B
} {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	result := make([]struct {
		First  A
		Second B
	}, minLen)
	for i := 0; i < minLen; i++ {
		result[i].First = a[i]
		result[i].Second = b[i]
	}
	return result
}

// =============================================================================
// PART 4: Type assertion with generics
// =============================================================================

// Generic pool with type safety
type Pool[T any] struct {
	pool sync.Pool
}

func NewPool[T any](newFn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() interface{} { return newFn() },
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(v T) {
	p.pool.Put(v)
}

// =============================================================================
// PART 5: Generic middleware/decorator pattern
// =============================================================================

type Handler[Req any, Resp any] func(Req) (Resp, error)

type Middleware[Req any, Resp any] func(Handler[Req, Resp]) Handler[Req, Resp]

func WithLogging[Req any, Resp any]() Middleware[Req, Resp] {
	return func(next Handler[Req, Resp]) Handler[Req, Resp] {
		return func(req Req) (Resp, error) {
			fmt.Printf("  → Request: %v\n", req)
			resp, err := next(req)
			fmt.Printf("  ← Response: %v, err: %v\n", resp, err)
			return resp, err
		}
	}
}

func WithRetry[Req any, Resp any](maxRetries int) Middleware[Req, Resp] {
	return func(next Handler[Req, Resp]) Handler[Req, Resp] {
		return func(req Req) (resp Resp, err error) {
			for i := 0; i <= maxRetries; i++ {
				resp, err = next(req)
				if err == nil {
					return resp, nil
				}
				fmt.Printf("  Retry %d/%d\n", i+1, maxRetries)
			}
			return resp, err
		}
	}
}

func Chain[Req any, Resp any](handler Handler[Req, Resp], middlewares ...Middleware[Req, Resp]) Handler[Req, Resp] {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func main() {
	// Part 1: Custom constraints
	fmt.Println("=== Custom Constraints ===")
	ints := []int{1, 2, 3, 4, 5}
	fmt.Printf("Sum of ints: %d\n", Sum(ints))

	meters := []Meters{1.5, 2.5, 3.0}
	fmt.Printf("Sum of meters: %.1f\n", Sum(meters))

	strings := []string{"hello", " ", "world"}
	fmt.Printf("Sum of strings: %q\n", Sum(strings))

	// Part 2: Generic data structures
	fmt.Println("\n=== Result Type ===")
	r := Ok(42).Map(func(v int) int { return v * 2 })
	v, err := r.Unwrap()
	fmt.Printf("Result: %d, err: %v\n", v, err)

	fmt.Println("\n=== Optional Type ===")
	opt := Some(10).Map(func(v int) int { return v + 5 })
	fmt.Printf("Optional: %v\n", opt.OrElse(0))
	fmt.Printf("None: %v\n", None[int]().OrElse(-1))

	fmt.Println("\n=== Generic BST ===")
	tree := &BST[int]{}
	for _, v := range []int{5, 3, 7, 1, 4, 6, 8} {
		tree.Insert(v)
	}
	fmt.Printf("In-order: %v\n", tree.InOrder())
	fmt.Printf("Search 4: %v, Search 9: %v\n", tree.Search(4), tree.Search(9))

	// String tree
	strTree := &BST[string]{}
	for _, s := range []string{"banana", "apple", "cherry"} {
		strTree.Insert(s)
	}
	fmt.Printf("String BST: %v\n", strTree.InOrder())

	// Part 3: Functional utilities
	fmt.Println("\n=== Functional Utilities ===")
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	doubled := Map(nums, func(n int) int { return n * 2 })
	fmt.Printf("Map(*2): %v\n", doubled)

	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Printf("Filter(even): %v\n", evens)

	sum := Reduce(nums, 0, func(acc, v int) int { return acc + v })
	fmt.Printf("Reduce(sum): %d\n", sum)

	words := []string{"hello", "world", "hi", "hey", "wow"}
	grouped := GroupBy(words, func(s string) byte { return s[0] })
	fmt.Printf("GroupBy(first char): %v\n", grouped)

	zipped := Zip([]int{1, 2, 3}, []string{"a", "b", "c"})
	fmt.Printf("Zip: %v\n", zipped)

	// Part 5: Middleware chain
	fmt.Println("\n=== Generic Middleware Chain ===")
	handler := func(n int) (string, error) {
		return fmt.Sprintf("processed:%d", n), nil
	}
	chain := Chain[int, string](handler, WithLogging[int, string]())
	result, _ := chain(42)
	fmt.Printf("Final result: %s\n", result)
}
