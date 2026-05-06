// =============================================================================
// LESSON 2.2: sync PRIMITIVES — Beyond Mutex
// =============================================================================
//
// The sync package provides low-level primitives that are faster than channels
// for certain patterns. Know WHEN to use each.
//
// RULE OF THUMB:
//   - Channels: for communication, ownership transfer, signaling
//   - Mutex/Atomic: for protecting shared state, counters, caches
// =============================================================================

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// 1. sync.RWMutex — Multiple readers, single writer
// =============================================================================
// Use when reads vastly outnumber writes (e.g., config, caches).
// Multiple goroutines can hold RLock simultaneously.
// Write lock (Lock) is exclusive — blocks all readers and writers.

type SafeConfig struct {
	mu     sync.RWMutex
	values map[string]string
}

func NewSafeConfig() *SafeConfig {
	return &SafeConfig{values: make(map[string]string)}
}

func (c *SafeConfig) Get(key string) (string, bool) {
	c.mu.RLock()         // shared read lock — many goroutines can hold this
	defer c.mu.RUnlock()
	v, ok := c.values[key]
	return v, ok
}

func (c *SafeConfig) Set(key, value string) {
	c.mu.Lock()          // exclusive write lock — blocks all other access
	defer c.mu.Unlock()
	c.values[key] = value
}

// =============================================================================
// 2. sync.Once — Exactly-once initialization
// =============================================================================
// Thread-safe lazy initialization. The function runs exactly once,
// even if called from multiple goroutines simultaneously.
// All callers block until the first call completes.

type Database struct {
	conn string
}

var (
	dbInstance *Database
	dbOnce     sync.Once
)

func GetDB() *Database {
	dbOnce.Do(func() {
		// Expensive initialization — runs exactly once
		fmt.Println("Connecting to database...")
		time.Sleep(100 * time.Millisecond) // simulate connection
		dbInstance = &Database{conn: "postgres://localhost/mydb"}
	})
	return dbInstance
}

// sync.OnceValue (Go 1.21+) — Returns a value, cleaner API
// var getConfig = sync.OnceValue(func() *Config {
//     return loadConfig()
// })

// =============================================================================
// 3. sync.Map — Concurrent map for specific use cases
// =============================================================================
// sync.Map is optimized for TWO specific patterns:
//   1. Write-once, read-many (like a cache that's populated once)
//   2. Non-overlapping keys across goroutines (each goroutine owns its keys)
//
// For other patterns, a regular map + RWMutex is usually faster!

func demonstrateSyncMap() {
	fmt.Println("\n=== sync.Map ===")
	var m sync.Map

	// Store
	m.Store("key1", "value1")
	m.Store("key2", 42)

	// Load
	if v, ok := m.Load("key1"); ok {
		fmt.Printf("key1 = %v\n", v)
	}

	// LoadOrStore — atomic get-or-set
	// Returns existing value if key exists, stores and returns new value if not
	actual, loaded := m.LoadOrStore("key3", "new_value")
	fmt.Printf("key3 = %v, already existed: %v\n", actual, loaded)

	// LoadAndDelete — atomic get-and-remove
	v, loaded := m.LoadAndDelete("key2")
	fmt.Printf("deleted key2 = %v, existed: %v\n", v, loaded)

	// Range — iterate (snapshot-ish, but not perfectly consistent)
	m.Range(func(key, value any) bool {
		fmt.Printf("  %v: %v\n", key, value)
		return true // continue iteration
	})
}

// =============================================================================
// 4. sync.Cond — Conditional waiting
// =============================================================================
// Wait for a condition to become true. Multiple goroutines can wait on
// the same condition. Signal wakes one waiter, Broadcast wakes all.
//
// WARNING: sync.Cond is error-prone. Prefer channels when possible.
// Use sync.Cond when you need Broadcast (channels can't do this easily).

type BoundedQueue struct {
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	items    []int
	maxSize  int
}

func NewBoundedQueue(maxSize int) *BoundedQueue {
	q := &BoundedQueue{
		items:   make([]int, 0, maxSize),
		maxSize: maxSize,
	}
	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)
	return q
}

func (q *BoundedQueue) Put(item int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// CRITICAL: Always use a loop, not if!
	// Spurious wakeups can happen, and the condition might not be true
	// even after Signal/Broadcast.
	for len(q.items) == q.maxSize {
		q.notFull.Wait() // releases lock, waits, re-acquires lock
	}

	q.items = append(q.items, item)
	q.notEmpty.Signal() // wake one consumer
}

func (q *BoundedQueue) Get() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.items) == 0 {
		q.notEmpty.Wait()
	}

	item := q.items[0]
	q.items = q.items[1:]
	q.notFull.Signal() // wake one producer
	return item
}

// =============================================================================
// 5. sync/atomic — Lock-free operations
// =============================================================================
// Atomic operations are the fastest synchronization primitive.
// Use for simple counters, flags, and pointer swaps.

type AtomicCounter struct {
	count atomic.Int64 // Go 1.19+ typed atomics
}

func (c *AtomicCounter) Increment() int64 {
	return c.count.Add(1)
}

func (c *AtomicCounter) Get() int64 {
	return c.count.Load()
}

// Atomic pointer swap — for lock-free config reload
type Config struct {
	DatabaseURL string
	MaxConns    int
	Debug       bool
}

type AtomicConfig struct {
	current atomic.Pointer[Config] // Go 1.19+ typed atomic pointer
}

func (ac *AtomicConfig) Load() *Config {
	return ac.current.Load()
}

func (ac *AtomicConfig) Store(cfg *Config) {
	ac.current.Store(cfg)
}

// CompareAndSwap — optimistic concurrency control
type AtomicStack struct {
	head atomic.Pointer[node]
}

type node struct {
	value int
	next  *node
}

func (s *AtomicStack) Push(v int) {
	n := &node{value: v}
	for {
		old := s.head.Load()
		n.next = old
		if s.head.CompareAndSwap(old, n) {
			return // successfully pushed
		}
		// CAS failed — another goroutine modified head, retry
	}
}

func (s *AtomicStack) Pop() (int, bool) {
	for {
		old := s.head.Load()
		if old == nil {
			return 0, false
		}
		if s.head.CompareAndSwap(old, old.next) {
			return old.value, true
		}
	}
}

// =============================================================================
// 6. sync.WaitGroup — Advanced patterns
// =============================================================================

// Dynamic WaitGroup — add tasks based on results of other tasks
func crawl(url string, depth int, wg *sync.WaitGroup) {
	defer wg.Done()
	if depth == 0 {
		return
	}

	// Simulate discovering links
	links := []string{url + "/a", url + "/b"}

	for _, link := range links {
		wg.Add(1) // Add BEFORE launching goroutine (not inside!)
		go crawl(link, depth-1, wg)
	}
}

func main() {
	// 1. RWMutex
	fmt.Println("=== RWMutex Config ===")
	cfg := NewSafeConfig()
	cfg.Set("database", "postgres")
	v, _ := cfg.Get("database")
	fmt.Printf("database = %s\n", v)

	// 2. sync.Once
	fmt.Println("\n=== sync.Once ===")
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db := GetDB()
			_ = db
		}()
	}
	wg.Wait()
	fmt.Println("All goroutines got same DB instance")

	// 3. sync.Map
	demonstrateSyncMap()

	// 4. sync.Cond
	fmt.Println("\n=== sync.Cond Bounded Queue ===")
	q := NewBoundedQueue(3)
	var wg2 sync.WaitGroup

	// Producer
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		for i := 0; i < 10; i++ {
			q.Put(i)
			fmt.Printf("Produced: %d\n", i)
		}
	}()

	// Consumer
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		for i := 0; i < 10; i++ {
			v := q.Get()
			fmt.Printf("Consumed: %d\n", v)
		}
	}()
	wg2.Wait()

	// 5. Atomics
	fmt.Println("\n=== Atomic Operations ===")
	counter := &AtomicCounter{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()
	fmt.Printf("Final count: %d (expected 1000)\n", counter.Get())

	// Lock-free stack
	stack := &AtomicStack{}
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	for {
		v, ok := stack.Pop()
		if !ok {
			break
		}
		fmt.Printf("Popped: %d\n", v)
	}
}
