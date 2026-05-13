/*
 * ============================================================
 *  CHAPTER 56: REFERENCE TYPES & MEMORY MANAGEMENT
 * ============================================================
 *  Java has FOUR reference types, not just one. Understanding
 *  them is the difference between "Java developer" and "Java god."
 *
 *  GC decides what to collect based on REACHABILITY:
 *
 *  ┌─────────────────────────────────────────────────┐
 *  │  Strongest                        Weakest       │
 *  │  Strong → Soft → Weak → Phantom → Unreachable  │
 *  └─────────────────────────────────────────────────┘
 *
 *  TOPICS:
 *    1. Strong References (default)
 *    2. SoftReference — Memory-Sensitive Caching
 *    3. WeakReference — Canonical Maps
 *    4. PhantomReference — Post-Mortem Cleanup
 *    5. ReferenceQueue — Notification on Collection
 *    6. WeakHashMap — Auto-Expiring Cache
 *    7. Cleaner API — Modern Finalizer Replacement
 *    8. Finalizers — Why They're Evil
 *    9. Reference Reachability Rules
 * ============================================================
 */

import java.lang.ref.*;
import java.util.*;
import java.util.concurrent.*;

public class Chapter56_ReferenceTypes {

    // ========================================================
    // REFERENCE TYPE COMPARISON
    // ========================================================
    /*
     * ┌──────────────┬─────────────────┬──────────────────────┬──────────────────────────┐
     * │ Type         │ Collected When  │ Use Case             │ get() after GC           │
     * ├──────────────┼─────────────────┼──────────────────────┼──────────────────────────┤
     * │ Strong       │ Never (if       │ Normal references    │ Always returns object    │
     * │              │ reachable)      │                      │                          │
     * ├──────────────┼─────────────────┼──────────────────────┼──────────────────────────┤
     * │ Soft         │ Before OOM      │ Memory-sensitive     │ null if memory pressure  │
     * │              │ (memory press.) │ caching              │                          │
     * ├──────────────┼─────────────────┼──────────────────────┼──────────────────────────┤
     * │ Weak         │ Next GC cycle   │ Canonical maps,      │ null after any GC        │
     * │              │                 │ listeners, metadata  │                          │
     * ├──────────────┼─────────────────┼──────────────────────┼──────────────────────────┤
     * │ Phantom      │ After finalize  │ Post-mortem cleanup, │ ALWAYS returns null      │
     * │              │                 │ native resource free │ (use ReferenceQueue)     │
     * └──────────────┴─────────────────┴──────────────────────┴──────────────────────────┘
     */

    // ========================================================
    // 1. STRONG REFERENCE — The Default
    // ========================================================

    static void demoStrongReference() {
        System.out.println("--- 1. Strong Reference ---\n");

        // This is what you always use
        Object obj = new Object();  // strong reference
        // obj is NEVER collected while this variable is in scope

        System.out.println("  obj = " + obj);
        System.out.println("  As long as 'obj' variable exists, the object lives");
        System.out.println("  Set obj = null to make it eligible for GC");

        obj = null;  // NOW the object can be collected
        System.out.println("  obj = null → eligible for GC\n");
    }

    // ========================================================
    // 2. SOFT REFERENCE — Memory-Sensitive Cache
    // ========================================================

    static void demoSoftReference() {
        System.out.println("--- 2. SoftReference ---\n");

        /*
         * SoftReference: GC clears it ONLY when memory is low.
         * JVM guarantees: soft refs are cleared before throwing OOM.
         * The GC tries to keep them as long as possible.
         *
         * PERFECT FOR: caches where you want to use memory when
         * available, but survive memory pressure.
         *
         * JVM flag: -XX:SoftRefLRUPolicyMSPerMB=1000
         *   → Each MB of free heap keeps soft refs alive ~1000ms
         */

        // Create a large object inside a SoftReference
        byte[] largeData = new byte[1024 * 1024]; // 1MB
        SoftReference<byte[]> softRef = new SoftReference<>(largeData);

        System.out.println("  Before clearing strong ref:");
        System.out.println("    softRef.get() != null? " + (softRef.get() != null));

        // Remove strong reference — now only soft-reachable
        largeData = null;

        System.out.println("  After clearing strong ref:");
        System.out.println("    softRef.get() != null? " + (softRef.get() != null));
        System.out.println("    (still alive — GC only clears when memory is low)");

        // Simple soft-reference cache
        Map<String, SoftReference<byte[]>> cache = new HashMap<>();
        cache.put("image1", new SoftReference<>(new byte[1024]));
        cache.put("image2", new SoftReference<>(new byte[1024]));

        // Reading from cache (must check for null!)
        SoftReference<byte[]> ref = cache.get("image1");
        if (ref != null) {
            byte[] data = ref.get();
            if (data != null) {
                System.out.println("  Cache hit: image1 (" + data.length + " bytes)");
            } else {
                System.out.println("  Cache evicted by GC — reload needed");
            }
        }
        System.out.println();
    }

    // ========================================================
    // 3. WEAK REFERENCE — Collected at Next GC
    // ========================================================

    static void demoWeakReference() {
        System.out.println("--- 3. WeakReference ---\n");

        /*
         * WeakReference: GC can collect it at ANY GC cycle.
         * Does NOT prevent garbage collection at all.
         *
         * USE CASES:
         *   - Associating metadata with objects without preventing GC
         *   - Listener/observer lists (avoid memory leaks)
         *   - Canonicalized mappings (WeakHashMap)
         *   - ClassLoader leak prevention
         */

        Object obj = new Object();
        WeakReference<Object> weakRef = new WeakReference<>(obj);

        System.out.println("  Before nulling strong ref:");
        System.out.println("    weakRef.get() = " + weakRef.get());

        obj = null;  // remove strong reference

        System.out.println("  After nulling strong ref (before GC):");
        System.out.println("    weakRef.get() = " + weakRef.get());

        System.gc();  // suggest GC (not guaranteed, but usually works for demo)

        // After GC, weak ref is likely cleared
        System.out.println("  After System.gc():");
        System.out.println("    weakRef.get() = " + weakRef.get() + " (likely null)");
        System.out.println();
    }

    // ========================================================
    // 4. PHANTOM REFERENCE — Post-Mortem Cleanup
    // ========================================================

    static void demoPhantomReference() {
        System.out.println("--- 4. PhantomReference ---\n");

        /*
         * PhantomReference:
         *   - get() ALWAYS returns null (you can never access the object)
         *   - Enqueued in ReferenceQueue AFTER the object is finalized
         *   - Used for cleanup actions AFTER the object is gone
         *   - Safer than finalizers (no resurrection possible)
         *
         * LIFECYCLE:
         *   1. Object becomes phantom-reachable (no strong/soft/weak refs)
         *   2. Object is finalized (if it has a finalizer)
         *   3. PhantomReference is enqueued in ReferenceQueue
         *   4. You poll the queue and do cleanup
         *   5. Object's memory is reclaimed
         *
         * Note: Before Java 9, phantom refs were NOT auto-cleared.
         *       Since Java 9, they ARE auto-cleared after enqueueing.
         */

        ReferenceQueue<Object> queue = new ReferenceQueue<>();
        Object obj = new Object();
        PhantomReference<Object> phantomRef = new PhantomReference<>(obj, queue);

        System.out.println("  phantomRef.get() = " + phantomRef.get() + " (ALWAYS null!)");
        System.out.println("  Queue poll before GC: " + queue.poll());

        obj = null;
        System.gc();

        // Check if phantom ref was enqueued
        try {
            Reference<?> ref = queue.remove(1000); // wait up to 1 second
            if (ref != null) {
                System.out.println("  PhantomRef enqueued after GC! Do cleanup here.");
                ref.clear(); // allow memory reclamation
            } else {
                System.out.println("  PhantomRef not yet enqueued (GC didn't run fully)");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        System.out.println();
    }

    // ========================================================
    // 5. REFERENCE QUEUE — Get Notified When Refs Are Cleared
    // ========================================================

    static void demoReferenceQueue() {
        System.out.println("--- 5. ReferenceQueue ---\n");

        /*
         * ReferenceQueue: GC enqueues cleared references here.
         * You poll/remove from the queue to learn which objects died.
         *
         * Works with: SoftReference, WeakReference, PhantomReference
         * (NOT with strong references — they're never cleared by GC)
         */

        ReferenceQueue<String> queue = new ReferenceQueue<>();

        // Create several weak references
        Map<WeakReference<String>, String> refMap = new HashMap<>();

        // Use new String() to avoid string pool (pool creates strong refs)
        String s1 = new String("one");
        String s2 = new String("two");
        String s3 = new String("three");

        WeakReference<String> ref1 = new WeakReference<>(s1, queue);
        WeakReference<String> ref2 = new WeakReference<>(s2, queue);
        WeakReference<String> ref3 = new WeakReference<>(s3, queue);

        refMap.put(ref1, "data-for-one");
        refMap.put(ref2, "data-for-two");
        refMap.put(ref3, "data-for-three");

        // Kill some strong references
        s1 = null;
        s3 = null;
        // s2 is still alive

        System.gc();

        // Poll queue for cleared references
        Reference<? extends String> cleared;
        int clearedCount = 0;
        while ((cleared = queue.poll()) != null) {
            String associatedData = refMap.remove(cleared);
            System.out.println("  Cleared: " + associatedData);
            clearedCount++;
        }
        System.out.println("  Total cleared: " + clearedCount);
        System.out.println("  Still alive: ref2.get() = " + ref2.get());
        System.out.println();
    }

    // ========================================================
    // 6. WEAKHASHMAP — Auto-Expiring Cache
    // ========================================================

    static void demoWeakHashMap() {
        System.out.println("--- 6. WeakHashMap ---\n");

        /*
         * WeakHashMap: keys are held via WeakReferences.
         * When a key is GC'd, the entry is automatically removed.
         *
         * CRITICAL: Keys must NOT be string literals or interned strings
         * (they're held in the string pool → never GC'd → never removed)
         *
         * USE CASES:
         *   - Associating metadata with objects you don't own
         *   - Listener registry (auto-deregister when listener is GC'd)
         *   - ClassLoader-to-data mappings
         *
         * NOTE: WeakHashMap is NOT thread-safe.
         *       For concurrent use: Collections.synchronizedMap(new WeakHashMap<>())
         *       or build your own with ConcurrentHashMap + WeakReferences.
         */

        WeakHashMap<Object, String> weakMap = new WeakHashMap<>();

        Object key1 = new Object();
        Object key2 = new Object();
        Object key3 = new Object();

        weakMap.put(key1, "value1");
        weakMap.put(key2, "value2");
        weakMap.put(key3, "value3");

        System.out.println("  Before GC: size = " + weakMap.size());

        // Remove strong references to some keys
        key1 = null;
        key3 = null;

        System.gc();

        // WeakHashMap cleans up stale entries on access
        System.out.println("  After GC:  size = " + weakMap.size());
        System.out.println("  key2's value: " + weakMap.get(key2));

        // DANGER: String literals are never GC'd!
        WeakHashMap<String, String> dangerMap = new WeakHashMap<>();
        dangerMap.put("literal", "never removed");  // ← string pool holds strong ref!
        dangerMap.put(new String("dynamic"), "can be removed");

        System.gc();
        System.out.println("\n  String literal key survived GC: " + dangerMap.containsKey("literal"));
        System.out.println("  Dynamic string key survived GC: " + dangerMap.containsKey("dynamic"));
        System.out.println();
    }

    // ========================================================
    // 7. CLEANER API (Java 9+) — Modern Finalizer Replacement
    // ========================================================

    static void demoCleaner() {
        System.out.println("--- 7. Cleaner API ---\n");

        /*
         * BEFORE Java 9: Override finalize() for cleanup.
         * PROBLEM: Finalizers are unpredictable, slow, and dangerous.
         *
         * SINCE Java 9: Use java.lang.ref.Cleaner instead.
         *
         * HOW IT WORKS:
         *   1. Create a Cleaner instance (shared, thread-safe)
         *   2. Register an object + a cleanup Runnable
         *   3. When the object becomes phantom-reachable, the
         *      Runnable is executed on the Cleaner thread
         *
         * RULE: The cleanup action must NOT reference the object
         *       being cleaned (or it would prevent GC → leak!)
         *       Use a STATIC inner class or lambda that captures
         *       only the resource (not 'this').
         */

        // Simulated native resource
        class NativeResource {
            final long handle;
            NativeResource(long handle) { this.handle = handle; }
        }

        // Cleanup action — MUST NOT reference the NativeResourceHolder
        class CleanupAction implements Runnable {
            final long handle;
            CleanupAction(long handle) { this.handle = handle; }

            @Override
            public void run() {
                System.out.println("    [Cleaner] Releasing native handle: " + handle);
                // In real code: freeNativeResource(handle);
            }
        }

        // Usage pattern
        java.lang.ref.Cleaner cleaner = java.lang.ref.Cleaner.create();

        // Register
        NativeResource resource = new NativeResource(12345L);
        java.lang.ref.Cleaner.Cleanable cleanable =
            cleaner.register(resource, new CleanupAction(resource.handle));

        System.out.println("  Resource registered with Cleaner");
        System.out.println("  Manual clean (deterministic): ");
        cleanable.clean();  // Can also clean manually (idempotent)

        // If not cleaned manually, Cleaner does it when object is GC'd
        System.out.println("  (If not cleaned manually, Cleaner runs on GC)");

        /*
         * BEST PRACTICE:
         *   Implement AutoCloseable for deterministic cleanup (try-with-resources)
         *   AND register with Cleaner as safety net.
         *
         *   class MyResource implements AutoCloseable {
         *       private static final Cleaner cleaner = Cleaner.create();
         *       private final Cleaner.Cleanable cleanable;
         *       private final ResourceState state; // static inner class!
         *
         *       MyResource() {
         *           state = new ResourceState(...);
         *           cleanable = cleaner.register(this, state);
         *       }
         *
         *       @Override public void close() { cleanable.clean(); }
         *
         *       private static class ResourceState implements Runnable {
         *           // holds ONLY the native resource, NOT 'this'
         *           public void run() { freeResource(); }
         *       }
         *   }
         */
        System.out.println();
    }

    // ========================================================
    // 8. FINALIZERS — Why They're Evil
    // ========================================================

    static void demoFinalizerEvils() {
        System.out.println("--- 8. Why Finalizers Are Evil ---\n");

        /*
         * @Override protected void finalize() throws Throwable { ... }
         *
         * PROBLEMS:
         *   1. NO GUARANTEE when (or if!) finalize() runs
         *   2. Resurrection: finalize() can make object reachable again!
         *      → Object survives this GC → finalize() never runs again
         *   3. GC overhead: finalizeable objects need TWO GC cycles
         *      → First: mark as finalizable, run finalizer
         *      → Second: actually collect the object
         *   4. Thread safety: finalize() runs on a separate Finalizer thread
         *   5. Exception swallowing: exceptions in finalize() are silently ignored
         *   6. Subclass attacks: malicious subclass can override finalize()
         *      to keep references alive → security vulnerability
         *
         * Since Java 9: finalize() is DEPRECATED
         * Since Java 18: finalize() is DEPRECATED FOR REMOVAL
         *
         * USE INSTEAD:
         *   → try-with-resources (AutoCloseable) for deterministic cleanup
         *   → Cleaner API for safety-net cleanup
         *   → PhantomReference + ReferenceQueue for advanced cases
         */

        System.out.println("  ❌ finalize() — deprecated, unpredictable, dangerous");
        System.out.println("  ✅ AutoCloseable + try-with-resources (deterministic)");
        System.out.println("  ✅ Cleaner API (safety net for missed close())");
        System.out.println("  ✅ PhantomReference + ReferenceQueue (advanced)");
        System.out.println();
    }

    // ========================================================
    // 9. PRACTICAL: Building a Proper Cache
    // ========================================================

    // Thread-safe soft-reference cache with expiry tracking
    static class SoftCache<K, V> {
        private final Map<K, SoftReference<V>> cache = new ConcurrentHashMap<>();
        private final ReferenceQueue<V> queue = new ReferenceQueue<>();
        private final Map<SoftReference<V>, K> reverseMap = new ConcurrentHashMap<>();

        void put(K key, V value) {
            cleanStaleEntries();
            SoftReference<V> ref = new SoftReference<>(value, queue);
            cache.put(key, ref);
            reverseMap.put(ref, key);
        }

        Optional<V> get(K key) {
            cleanStaleEntries();
            SoftReference<V> ref = cache.get(key);
            if (ref == null) return Optional.empty();
            V value = ref.get();
            if (value == null) {
                // Reference was cleared by GC
                cache.remove(key);
                return Optional.empty();
            }
            return Optional.of(value);
        }

        int size() {
            cleanStaleEntries();
            return cache.size();
        }

        @SuppressWarnings("unchecked")
        private void cleanStaleEntries() {
            Reference<? extends V> ref;
            while ((ref = queue.poll()) != null) {
                K key = reverseMap.remove(ref);
                if (key != null) {
                    cache.remove(key);
                }
            }
        }
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== CHAPTER 56: REFERENCE TYPES & MEMORY ===\n");

        demoStrongReference();
        demoSoftReference();
        demoWeakReference();
        demoPhantomReference();
        demoReferenceQueue();
        demoWeakHashMap();
        demoCleaner();
        demoFinalizerEvils();

        // --- 9. SoftCache Demo ---
        System.out.println("--- 9. SoftCache (Practical) ---\n");

        SoftCache<String, byte[]> imageCache = new SoftCache<>();
        imageCache.put("avatar1", new byte[1024]);
        imageCache.put("avatar2", new byte[1024]);

        System.out.println("  Cache size: " + imageCache.size());
        System.out.println("  Get avatar1: " + imageCache.get("avatar1").map(b -> b.length + " bytes").orElse("evicted"));
        System.out.println("  Get missing: " + imageCache.get("nope").orElse(null));

        // --- Reachability Summary ---
        System.out.println("\n--- Reachability Rules ---\n");
        System.out.println("  STRONGLY REACHABLE → can reach via strong refs → never collected");
        System.out.println("  SOFTLY REACHABLE   → only via soft refs → collected before OOM");
        System.out.println("  WEAKLY REACHABLE   → only via weak refs → collected at next GC");
        System.out.println("  PHANTOM REACHABLE  → finalized, phantom ref exists → enqueued");
        System.out.println("  UNREACHABLE        → no references at all → collected");

        System.out.println("\n✓ Reference Types & Memory Management Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Build a WeakReference-based listener/observer system where
 *    listeners are automatically removed when GC'd.
 * 2. Create a LRU cache backed by SoftReferences with a
 *    hard size limit.
 * 3. Use PhantomReference + ReferenceQueue to track when
 *    specific objects are collected (memory leak detector).
 * 4. Implement a resource manager using Cleaner that handles
 *    both manual close() and GC-triggered cleanup.
 *
 * NEXT: Chapter 57 — Java Agents & Instrumentation
 */
