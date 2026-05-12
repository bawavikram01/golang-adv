/*
 * =============================================================
 * LLD CASE STUDY 7: LRU CACHE
 * =============================================================
 *
 * REQUIREMENTS:
 *   - O(1) get and put operations
 *   - Fixed capacity — evicts least recently used item
 *   - Thread-safe version
 *
 * DATA STRUCTURE:
 *   HashMap + DoublyLinkedList
 *   - HashMap gives O(1) lookup
 *   - DoublyLinkedList gives O(1) insertion/deletion for ordering
 *
 *   HEAD ←→ [Most Recent] ←→ ... ←→ [Least Recent] ←→ TAIL
 *
 * DESIGN PATTERNS USED:
 *   - Proxy (cache acts as proxy to slow data source)
 *   - Strategy (eviction policies could be swappable)
 */

import java.util.*;

public class LRUCache {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Demo 1: Basic LRU Cache
        // ═══════════════════════════════════════════════════════
        System.out.println("=== LRU CACHE DEMO ===\n");

        Cache<Integer, String> cache = new LRUCacheImpl<>(3);

        cache.put(1, "One");
        cache.put(2, "Two");
        cache.put(3, "Three");
        System.out.println("Cache after adding 1,2,3: " + cache);
        // [3=Three, 2=Two, 1=One]

        cache.get(1);  // Access 1 → moves to front
        System.out.println("After get(1):             " + cache);
        // [1=One, 3=Three, 2=Two]

        cache.put(4, "Four");  // Evicts 2 (least recently used)
        System.out.println("After put(4) — evicts 2:  " + cache);
        // [4=Four, 1=One, 3=Three]

        System.out.println("\nget(2) = " + cache.get(2));  // null — evicted
        System.out.println("get(3) = " + cache.get(3));    // Three

        cache.put(5, "Five");  // Evicts 1
        System.out.println("\nAfter put(5) — evicts 1:  " + cache);

        // ═══════════════════════════════════════════════════════
        // Demo 2: Eviction Policy comparison
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== EVICTION POLICIES ===\n");

        EvictionPolicy<Integer> lru = new LRUEvictionPolicy<>();
        EvictionPolicy<Integer> lfu = new LFUEvictionPolicy<>();

        System.out.println("--- LRU (Least Recently Used) ---");
        GenericCache<Integer, String> lruCache = new GenericCache<>(3, lru);
        lruCache.put(1, "A"); lruCache.put(2, "B"); lruCache.put(3, "C");
        lruCache.get(1); lruCache.get(1); lruCache.get(1);  // 1 accessed 3 times
        lruCache.get(2);  // 2 is most recent
        lruCache.put(4, "D");  // LRU evicts 3 (least recently used)
        System.out.println("After heavy get(1), then get(2), then put(4): contains 1? "
                + (lruCache.get(1) != null));

        System.out.println("\n--- LFU (Least Frequently Used) ---");
        GenericCache<Integer, String> lfuCache = new GenericCache<>(3, lfu);
        lfuCache.put(1, "A"); lfuCache.put(2, "B"); lfuCache.put(3, "C");
        lfuCache.get(1); lfuCache.get(1); lfuCache.get(1);  // 1 accessed 3 extra times
        lfuCache.get(2);  // 2 accessed once extra
        lfuCache.put(4, "D");  // LFU evicts 3 (least frequently used)
        System.out.println("After heavy get(1), then get(2), then put(4): contains 1? "
                + (lfuCache.get(1) != null));
    }
}

// ═══════════════════════════════════════════════════════════════
// CACHE INTERFACE
// ═══════════════════════════════════════════════════════════════
interface Cache<K, V> {
    V get(K key);
    void put(K key, V value);
}

// ═══════════════════════════════════════════════════════════════
// DOUBLY LINKED LIST NODE
// ═══════════════════════════════════════════════════════════════
class DLLNode<K, V> {
    K key;
    V value;
    DLLNode<K, V> prev, next;

    DLLNode(K key, V value) {
        this.key = key;
        this.value = value;
    }
}

// ═══════════════════════════════════════════════════════════════
// LRU CACHE — Core Implementation
// ═══════════════════════════════════════════════════════════════
/*
 * WHY HashMap + DoublyLinkedList?
 *
 *   HashMap alone: O(1) lookup but no ordering
 *   LinkedList alone: O(1) move-to-front but O(n) lookup
 *   Together: O(1) for both!
 *
 *   HashMap<Key, Node>  →  find node in O(1)
 *   DoublyLinkedList    →  move/remove node in O(1)
 */
class LRUCacheImpl<K, V> implements Cache<K, V> {
    private final int capacity;
    private final Map<K, DLLNode<K, V>> map;
    private final DLLNode<K, V> head;  // dummy
    private final DLLNode<K, V> tail;  // dummy

    public LRUCacheImpl(int capacity) {
        this.capacity = capacity;
        this.map = new HashMap<>();
        this.head = new DLLNode<>(null, null);
        this.tail = new DLLNode<>(null, null);
        head.next = tail;
        tail.prev = head;
    }

    @Override
    public V get(K key) {
        DLLNode<K, V> node = map.get(key);
        if (node == null) return null;
        moveToFront(node);
        return node.value;
    }

    @Override
    public void put(K key, V value) {
        DLLNode<K, V> existing = map.get(key);
        if (existing != null) {
            existing.value = value;
            moveToFront(existing);
            return;
        }

        if (map.size() == capacity) {
            evict();
        }

        DLLNode<K, V> newNode = new DLLNode<>(key, value);
        addToFront(newNode);
        map.put(key, newNode);
    }

    private void evict() {
        DLLNode<K, V> lru = tail.prev;  // node just before tail dummy
        removeNode(lru);
        map.remove(lru.key);
        System.out.println("  [EVICT] key=" + lru.key);
    }

    private void moveToFront(DLLNode<K, V> node) {
        removeNode(node);
        addToFront(node);
    }

    private void addToFront(DLLNode<K, V> node) {
        node.next = head.next;
        node.prev = head;
        head.next.prev = node;
        head.next = node;
    }

    private void removeNode(DLLNode<K, V> node) {
        node.prev.next = node.next;
        node.next.prev = node.prev;
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder("[");
        DLLNode<K, V> curr = head.next;
        while (curr != tail) {
            if (curr != head.next) sb.append(", ");
            sb.append(curr.key).append("=").append(curr.value);
            curr = curr.next;
        }
        return sb.append("]").toString();
    }
}

// ═══════════════════════════════════════════════════════════════
// EVICTION POLICY — Strategy Pattern
// ═══════════════════════════════════════════════════════════════
/*
 * Making eviction policy pluggable with Strategy pattern
 * so the same cache structure can support LRU, LFU, FIFO, etc.
 */
interface EvictionPolicy<K> {
    void onAccess(K key);
    void onInsert(K key);
    K evict();
}

class LRUEvictionPolicy<K> implements EvictionPolicy<K> {
    private final LinkedList<K> order = new LinkedList<>();

    @Override
    public void onAccess(K key) {
        order.remove(key);
        order.addFirst(key);
    }

    @Override
    public void onInsert(K key) {
        order.addFirst(key);
    }

    @Override
    public K evict() {
        return order.removeLast();
    }
}

class LFUEvictionPolicy<K> implements EvictionPolicy<K> {
    private final Map<K, Integer> frequency = new HashMap<>();

    @Override
    public void onAccess(K key) {
        frequency.merge(key, 1, Integer::sum);
    }

    @Override
    public void onInsert(K key) {
        frequency.put(key, 1);
    }

    @Override
    public K evict() {
        K leastFrequent = Collections.min(frequency.entrySet(),
                Map.Entry.comparingByValue()).getKey();
        frequency.remove(leastFrequent);
        return leastFrequent;
    }
}

// ═══════════════════════════════════════════════════════════════
// GENERIC CACHE with pluggable eviction
// ═══════════════════════════════════════════════════════════════
class GenericCache<K, V> implements Cache<K, V> {
    private final int capacity;
    private final Map<K, V> store = new HashMap<>();
    private final EvictionPolicy<K> evictionPolicy;

    public GenericCache(int capacity, EvictionPolicy<K> evictionPolicy) {
        this.capacity = capacity;
        this.evictionPolicy = evictionPolicy;
    }

    @Override
    public V get(K key) {
        V value = store.get(key);
        if (value != null) {
            evictionPolicy.onAccess(key);
        }
        return value;
    }

    @Override
    public void put(K key, V value) {
        if (store.containsKey(key)) {
            store.put(key, value);
            evictionPolicy.onAccess(key);
            return;
        }

        if (store.size() == capacity) {
            K evictedKey = evictionPolicy.evict();
            store.remove(evictedKey);
            System.out.println("  [EVICT] key=" + evictedKey);
        }

        store.put(key, value);
        evictionPolicy.onInsert(key);
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * 1. HashMap + DoublyLinkedList = O(1) get & put with ordering
 * 2. Dummy head/tail nodes eliminate null checks in DLL
 * 3. Strategy pattern makes eviction policy pluggable
 * 4. Same structure, different behavior: LRU vs LFU vs FIFO
 *
 * INTERVIEW TIPS:
 *   - Always mention time complexity: O(1) for both get/put
 *   - Draw the HashMap→Node→DLL diagram
 *   - Mention thread-safety with ConcurrentHashMap + locks
 *   - Discuss cache warming, hit ratio, TTL for bonus points
 *
 * COMPILE & RUN:
 *   javac LRUCache.java && java LRUCache
 */
