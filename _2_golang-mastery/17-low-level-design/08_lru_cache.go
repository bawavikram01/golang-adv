package lowleveldesign

import (
	"container/list"
	"fmt"
	"sync"
)

// =============================================================================
// LLD PROBLEM: LRU CACHE
// =============================================================================
// Design an LRU (Least Recently Used) Cache with O(1) get and put operations.
//
// Requirements:
// 1. Get(key) — returns value if exists, marks as recently used
// 2. Put(key, value) — inserts/updates, evicts LRU item if at capacity
// 3. O(1) time complexity for both operations
// 4. Thread-safe
// 5. Configurable capacity
//
// Data Structures:
// - HashMap (map): O(1) lookup by key
// - Doubly Linked List: O(1) removal and insertion at ends
//   - Head = Most Recently Used (MRU)
//   - Tail = Least Recently Used (LRU)
//
// This is the classic interview question. MEMORIZE the approach.

// =============================================================================
// IMPLEMENTATION
// =============================================================================

type CacheEntry struct {
	key   string
	value interface{}
}

type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List // front = MRU, back = LRU
	mu       sync.RWMutex

	// Metrics
	hits   int64
	misses int64
}

func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		panic("cache capacity must be positive")
	}
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
	}
}

// Get retrieves a value and marks it as recently used
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.items[key]; found {
		// Move to front (most recently used)
		c.order.MoveToFront(elem)
		c.hits++
		return elem.Value.(*CacheEntry).value, true
	}
	c.misses++
	return nil, false
}

// Put inserts or updates a key-value pair
func (c *LRUCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Key already exists — update and move to front
	if elem, found := c.items[key]; found {
		c.order.MoveToFront(elem)
		elem.Value.(*CacheEntry).value = value
		return
	}

	// Evict LRU item if at capacity
	if c.order.Len() >= c.capacity {
		c.evict()
	}

	// Insert new item at front
	entry := &CacheEntry{key: key, value: value}
	elem := c.order.PushFront(entry)
	c.items[key] = elem
}

// Delete removes a key from the cache
func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.items[key]; found {
		c.removeElement(elem)
		return true
	}
	return false
}

// evict removes the least recently used item (must hold lock)
func (c *LRUCache) evict() {
	tail := c.order.Back()
	if tail == nil {
		return
	}
	c.removeElement(tail)
}

func (c *LRUCache) removeElement(elem *list.Element) {
	c.order.Remove(elem)
	entry := elem.Value.(*CacheEntry)
	delete(c.items, entry.key)
}

// Len returns current number of items
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// Stats returns cache hit/miss statistics
func (c *LRUCache) Stats() (hits, misses int64, hitRate float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	total := c.hits + c.misses
	if total == 0 {
		return 0, 0, 0
	}
	return c.hits, c.misses, float64(c.hits) / float64(total)
}

// Keys returns all keys in MRU->LRU order
func (c *LRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, c.order.Len())
	for elem := c.order.Front(); elem != nil; elem = elem.Next() {
		keys = append(keys, elem.Value.(*CacheEntry).key)
	}
	return keys
}

// =============================================================================
// LRU CACHE WITH EXPIRATION (TTL)
// =============================================================================
// Production caches need expiration. Extend the basic LRU with TTL.

// =============================================================================
// GENERIC LRU CACHE (Go 1.22+ with generics)
// =============================================================================

type TypedCacheEntry[K comparable, V any] struct {
	key   K
	value V
}

type TypedLRUCache[K comparable, V any] struct {
	capacity int
	items    map[K]*list.Element
	order    *list.List
	mu       sync.RWMutex
}

func NewTypedLRUCache[K comparable, V any](capacity int) *TypedLRUCache[K, V] {
	return &TypedLRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element, capacity),
		order:    list.New(),
	}
}

func (c *TypedLRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.items[key]; found {
		c.order.MoveToFront(elem)
		return elem.Value.(*TypedCacheEntry[K, V]).value, true
	}
	var zero V
	return zero, false
}

func (c *TypedLRUCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.items[key]; found {
		c.order.MoveToFront(elem)
		elem.Value.(*TypedCacheEntry[K, V]).value = value
		return
	}

	if c.order.Len() >= c.capacity {
		tail := c.order.Back()
		if tail != nil {
			c.order.Remove(tail)
			delete(c.items, tail.Value.(*TypedCacheEntry[K, V]).key)
		}
	}

	entry := &TypedCacheEntry[K, V]{key: key, value: value}
	elem := c.order.PushFront(entry)
	c.items[key] = elem
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================

func ExampleLRUCache() {
	cache := NewLRUCache(3) // capacity of 3

	// Add items
	cache.Put("a", 1)
	cache.Put("b", 2)
	cache.Put("c", 3)
	fmt.Println("Keys (MRU->LRU):", cache.Keys()) // [c, b, a]

	// Access 'a' — moves it to front
	val, _ := cache.Get("a")
	fmt.Println("Get 'a':", val)                  // 1
	fmt.Println("Keys (MRU->LRU):", cache.Keys()) // [a, c, b]

	// Add 'd' — evicts 'b' (LRU)
	cache.Put("d", 4)
	fmt.Println("Keys (MRU->LRU):", cache.Keys()) // [d, a, c]

	// 'b' was evicted
	_, found := cache.Get("b")
	fmt.Println("Get 'b' found:", found) // false

	// Stats
	hits, misses, rate := cache.Stats()
	fmt.Printf("Hits: %d, Misses: %d, Hit Rate: %.2f\n", hits, misses, rate)

	// --- Generic version ---
	typedCache := NewTypedLRUCache[string, int](2)
	typedCache.Put("x", 100)
	typedCache.Put("y", 200)
	v, ok := typedCache.Get("x")
	fmt.Printf("Typed cache get 'x': %d (found: %v)\n", v, ok)
}
