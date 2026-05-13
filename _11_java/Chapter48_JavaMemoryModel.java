/*
 * ============================================================
 *  CHAPTER 48: JAVA MEMORY MODEL (JMM)
 * ============================================================
 *  The JMM defines how threads interact through memory.
 *  Understanding this is the difference between a developer
 *  who writes "working" concurrent code and one who writes
 *  CORRECT concurrent code.
 *
 *  KEY CONCEPTS:
 *    1. Visibility — when one thread sees another's writes
 *    2. Ordering — can instructions be reordered?
 *    3. Atomicity — is an operation indivisible?
 *    4. Happens-Before — the formal guarantee
 *
 *  THE PROBLEM:
 *    CPU caches, store buffers, and compiler optimizations
 *    can make writes INVISIBLE to other threads and can
 *    REORDER instructions for performance.
 *    The JMM tells you when you're GUARANTEED to see changes.
 * ============================================================
 */

import java.util.concurrent.*;
import java.util.concurrent.atomic.*;
import java.util.concurrent.locks.*;

public class Chapter48_JavaMemoryModel {

    // ========================================================
    // 1. VISIBILITY PROBLEM
    // ========================================================

    // ❌ BROKEN: reader may NEVER see running = false
    // The compiler may hoist the read out of the loop!
    static boolean running = true;  // not volatile!

    static void visibilityBroken() throws InterruptedException {
        Thread reader = new Thread(() -> {
            int count = 0;
            while (running) {  // may be optimized to: while (true)
                count++;
            }
            // May NEVER reach here — infinite loop!
        });

        reader.start();
        Thread.sleep(100);
        running = false;  // writer writes, but reader may not see it
        reader.join(1000);
        if (reader.isAlive()) {
            System.out.println("  ❌ Reader STUCK — didn't see running=false");
            reader.interrupt(); // force stop for demo
        }
    }

    // ✅ FIXED: volatile guarantees visibility
    static volatile boolean runningFixed = true;

    static void visibilityFixed() throws InterruptedException {
        Thread reader = new Thread(() -> {
            int count = 0;
            while (runningFixed) {
                count++;
            }
            System.out.println("  ✓ Reader stopped after " + count + " iterations");
        });

        reader.start();
        Thread.sleep(100);
        runningFixed = false;
        reader.join(1000);
    }

    // ========================================================
    // 2. HAPPENS-BEFORE RELATIONSHIPS
    // ========================================================
    //
    // "If action A happens-before action B, then A's effects
    //  are visible to B and A appears to execute before B"
    //
    // THE 8 HAPPENS-BEFORE RULES:
    //
    // 1. PROGRAM ORDER:
    //    Within a single thread, each statement happens-before
    //    the next statement.
    //
    // 2. MONITOR LOCK:
    //    An unlock on a monitor happens-before every subsequent
    //    lock on that SAME monitor.
    //    synchronized(lock) { x = 1; }  // unlock HB
    //    synchronized(lock) { print(x); }  // lock → sees x=1
    //
    // 3. VOLATILE:
    //    A write to a volatile field happens-before every
    //    subsequent read of that SAME volatile field.
    //    volatile int x;
    //    Thread A: x = 42;       // volatile write HB
    //    Thread B: int y = x;    // volatile read → sees 42
    //
    //    BONUS: volatile also has "piggybacking" — ALL writes
    //    before the volatile write are visible after the
    //    volatile read (not just the volatile variable!)
    //
    // 4. THREAD START:
    //    thread.start() happens-before any action in the started thread.
    //    x = 42;
    //    thread.start();  // thread sees x=42
    //
    // 5. THREAD JOIN:
    //    All actions in a thread happen-before join() returns.
    //    thread.join();  // after this, we see all of thread's writes
    //
    // 6. TRANSITIVITY:
    //    If A HB B and B HB C, then A HB C.
    //
    // 7. INTERRUPT:
    //    thread.interrupt() happens-before the interrupted
    //    thread detects the interrupt.
    //
    // 8. FINALIZER:
    //    Constructor completion happens-before finalize() starts.

    // ========================================================
    // 3. VOLATILE — DEEP DIVE
    // ========================================================

    // Volatile guarantees:
    //   ✓ Visibility (all threads see latest value)
    //   ✓ Ordering (no reordering across volatile access)
    //   ✗ Atomicity of compound operations (i++ is NOT atomic even with volatile)

    static volatile int volatileCounter = 0;

    // ❌ THIS IS STILL BROKEN despite volatile!
    // i++ is: read → increment → write (THREE operations)
    // Two threads can both read the same value
    static void volatileNotAtomic() throws InterruptedException {
        volatileCounter = 0;
        Thread t1 = new Thread(() -> { for (int i = 0; i < 10000; i++) volatileCounter++; });
        Thread t2 = new Thread(() -> { for (int i = 0; i < 10000; i++) volatileCounter++; });
        t1.start(); t2.start();
        t1.join(); t2.join();
        System.out.println("  Volatile counter (broken): " + volatileCounter + " (expected 20000)");
    }

    // ✅ FIX: Use AtomicInteger
    static AtomicInteger atomicCounter = new AtomicInteger(0);

    static void atomicCorrect() throws InterruptedException {
        atomicCounter.set(0);
        Thread t1 = new Thread(() -> { for (int i = 0; i < 10000; i++) atomicCounter.incrementAndGet(); });
        Thread t2 = new Thread(() -> { for (int i = 0; i < 10000; i++) atomicCounter.incrementAndGet(); });
        t1.start(); t2.start();
        t1.join(); t2.join();
        System.out.println("  Atomic counter (correct):  " + atomicCounter.get() + " (expected 20000)");
    }

    // ========================================================
    // 4. VOLATILE PIGGYBACKING
    // ========================================================
    // Writes BEFORE a volatile write are visible after the volatile read

    static int a, b, c;
    static volatile boolean ready;

    static void piggybackingDemo() throws InterruptedException {
        a = 0; b = 0; c = 0; ready = false;

        Thread writer = new Thread(() -> {
            a = 1;
            b = 2;
            c = 3;
            ready = true;  // volatile write — publishes ALL above
        });

        Thread reader = new Thread(() -> {
            while (!ready) { Thread.onSpinWait(); }  // volatile read
            // GUARANTEED to see a=1, b=2, c=3
            System.out.println("  Piggyback: a=" + a + " b=" + b + " c=" + c + " (all visible!)");
        });

        writer.start();
        reader.start();
        writer.join();
        reader.join();
    }

    // ========================================================
    // 5. DOUBLE-CHECKED LOCKING (DCL)
    // ========================================================
    // Classic concurrency puzzle — broken without volatile

    static class Singleton {
        // MUST be volatile! Without it, the instance reference may be
        // published before the constructor finishes (instruction reordering)
        private static volatile Singleton instance;

        private final String data;

        private Singleton() {
            this.data = "initialized";
        }

        static Singleton getInstance() {
            Singleton local = instance;  // read volatile only once
            if (local == null) {
                synchronized (Singleton.class) {
                    local = instance;
                    if (local == null) {
                        instance = local = new Singleton();
                    }
                }
            }
            return local;
        }

        // WHY volatile is needed:
        // Without volatile, thread B might see a NON-NULL instance
        // reference but with an UNINITIALIZED data field!
        // The JVM can reorder:
        //   1. Allocate memory
        //   2. Assign reference to instance  ← thread B sees this
        //   3. Run constructor               ← but NOT this yet!
        // Volatile prevents this reordering.
    }

    // Better alternative: use the holder pattern (Chapter 34)
    static class SingletonHolder {
        private SingletonHolder() {}
        private static class Holder {
            static final SingletonHolder INSTANCE = new SingletonHolder();
        }
        static SingletonHolder getInstance() { return Holder.INSTANCE; }
    }

    // ========================================================
    // 6. FINAL FIELDS — special JMM guarantee
    // ========================================================
    // If a field is final AND properly constructed (no 'this' escape),
    // it is guaranteed to be visible to all threads without synchronization.

    static class ImmutablePoint {
        private final int x;  // guaranteed visible after construction
        private final int y;

        ImmutablePoint(int x, int y) {
            this.x = x;
            this.y = y;
            // DO NOT publish 'this' here! If you leak 'this' in the
            // constructor, the guarantee breaks.
        }

        int getX() { return x; }
        int getY() { return y; }
    }

    // ========================================================
    // 7. MEMORY BARRIERS explained
    // ========================================================
    //
    // The CPU uses store buffers and caches for performance.
    // Memory barriers force ordering:
    //
    // LoadLoad:   no reads reordered before previous reads
    // StoreStore: no writes reordered before previous writes
    // LoadStore:  no writes reordered before previous reads
    // StoreLoad:  no reads reordered before previous writes (most expensive)
    //
    // Volatile write → StoreStore + StoreLoad barrier
    // Volatile read  → LoadLoad + LoadStore barrier
    // synchronized entry → all barriers (full fence)
    // synchronized exit  → all barriers (full fence)

    // ========================================================
    // 8. COMMON CONCURRENT BUGS
    // ========================================================

    // BUG 1: Race condition on check-then-act
    static class LazyInit {
        private Map<String, Object> cache; // not volatile, not synchronized

        // ❌ Two threads might both see null and create two instances
        Object getOrCreateBad(String key) {
            if (cache == null) {      // check
                cache = new ConcurrentHashMap<>();  // act
            }
            return cache.computeIfAbsent(key, k -> new Object());
        }
    }

    // BUG 2: 64-bit non-atomic writes (on 32-bit JVM)
    // long and double writes may NOT be atomic on 32-bit platforms!
    // Use volatile long/double or AtomicLong to be safe.

    // BUG 3: Publishing mutable state without sync
    // static Map<String, String> config = new HashMap<>();
    // Thread A populates config, Thread B reads it — B may see partial data
    // Fix: publish via volatile, synchronized, or use concurrent collection

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) throws Exception {

        // --- 1. Visibility Fixed ---
        System.out.println("=== VISIBILITY ===\n");
        visibilityFixed();

        // --- 2. Volatile is NOT atomic ---
        System.out.println("\n=== VOLATILE vs ATOMIC ===\n");
        volatileNotAtomic();
        atomicCorrect();

        // --- 3. Volatile Piggyback ---
        System.out.println("\n=== VOLATILE PIGGYBACKING ===\n");
        piggybackingDemo();

        // --- 4. Double-Checked Locking ---
        System.out.println("\n=== DOUBLE-CHECKED LOCKING ===\n");
        Singleton s1 = Singleton.getInstance();
        Singleton s2 = Singleton.getInstance();
        System.out.println("  Same instance? " + (s1 == s2));
        System.out.println("  Data: " + s1.data);

        // --- 5. Final Fields ---
        System.out.println("\n=== FINAL FIELD GUARANTEE ===\n");
        ImmutablePoint p = new ImmutablePoint(10, 20);
        System.out.println("  Point: (" + p.getX() + ", " + p.getY() + ")");
        System.out.println("  Final fields visible to all threads without sync!");

        // --- Summary ---
        System.out.println("\n=== HAPPENS-BEFORE SUMMARY ===");
        System.out.println("  1. Program order:  each line HB next (same thread)");
        System.out.println("  2. Monitor lock:   unlock HB subsequent lock (same monitor)");
        System.out.println("  3. Volatile:       write HB subsequent read (same variable)");
        System.out.println("  4. Thread.start(): HB first action in new thread");
        System.out.println("  5. Thread.join():  all actions in thread HB join() returns");
        System.out.println("  6. Transitivity:   A HB B, B HB C → A HB C");

        System.out.println("\n=== WHEN TO USE WHAT ===");
        System.out.println("  volatile:      simple flags, single-writer/multi-reader");
        System.out.println("  AtomicXxx:     counters, CAS operations");
        System.out.println("  synchronized:  compound actions, invariants");
        System.out.println("  Lock:          tryLock, timed lock, multiple conditions");
        System.out.println("  final:         immutable fields (publish safely)");
        System.out.println("  Concurrent*:   thread-safe collections");

        System.out.println("\n✓ Java Memory Model Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Write a program showing visibility bug (remove volatile, observe stuck thread).
 * 2. Implement a publication-safe lazy holder for an expensive resource.
 * 3. Prove that volatile i++ loses updates with a counter test.
 * 4. Implement a lock-free stack using AtomicReference and CAS.
 *
 * NEXT: Chapter 49 — Advanced Concurrency II
 */
