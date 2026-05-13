/*
 * ============================================================
 *  CHAPTER 58: UNSAFE & OFF-HEAP MEMORY
 * ============================================================
 *  sun.misc.Unsafe is Java's backdoor. It lets you do things
 *  the language was designed to prevent. It's dangerous,
 *  unsupported, and ESSENTIAL to understand because half the
 *  Java ecosystem depends on it.
 *
 *  TOPICS:
 *    1. What Is Unsafe?
 *    2. Getting an Unsafe Instance
 *    3. Direct Memory Allocation (Off-Heap)
 *    4. Object Manipulation Without Constructors
 *    5. CAS Operations (Compare-And-Swap)
 *    6. Field Offsets & Direct Field Access
 *    7. Memory Fences / Barriers
 *    8. Array Base & Scale
 *    9. Why Netty, Kafka, Cassandra Use Unsafe
 *   10. The Future: VarHandle, Foreign Memory API
 * ============================================================
 *
 *  WARNING: Unsafe can:
 *    - Crash the JVM with a segfault
 *    - Corrupt memory
 *    - Bypass security
 *    - Break type safety
 *  It exists because sometimes you NEED direct memory control.
 *
 *  NOTE: This code uses reflection to access Unsafe, which
 *  may require --add-opens flags on newer JVMs:
 *    java --add-opens java.base/sun.misc=ALL-UNNAMED Chapter58_Unsafe
 *
 * ============================================================
 */

import java.lang.reflect.Field;
import java.nio.ByteBuffer;
import java.util.concurrent.atomic.AtomicInteger;

public class Chapter58_Unsafe {

    // ========================================================
    // 1. WHAT IS UNSAFE?
    // ========================================================
    /*
     * sun.misc.Unsafe provides:
     *   - Direct memory allocation/deallocation (malloc/free)
     *   - CAS operations on arbitrary memory
     *   - Object creation without constructors
     *   - Direct field access by offset (bypasses access control)
     *   - Memory fences (loadFence, storeFence, fullFence)
     *   - Park/unpark threads (low-level wait/notify)
     *   - Throw exceptions without declaring them
     *
     * WHO USES IT:
     *   - java.util.concurrent (AtomicInteger, ConcurrentHashMap)
     *   - java.nio (DirectByteBuffer)
     *   - Netty (off-heap buffers)
     *   - Kafka (zero-copy networking)
     *   - Cassandra (off-heap memtables)
     *   - Hazelcast (off-heap storage)
     *   - Kryo, Protobuf (fast serialization)
     *   - Mockito (object creation without constructors)
     *   - Spring, Hibernate (field injection)
     */

    // ========================================================
    // 2. GETTING UNSAFE
    // ========================================================

    // Unsafe.getUnsafe() throws SecurityException if caller isn't
    // loaded by bootstrap ClassLoader. We use reflection instead.
    static Object getUnsafe() {
        try {
            Field f = Class.forName("sun.misc.Unsafe").getDeclaredField("theUnsafe");
            f.setAccessible(true);
            return f.get(null);
        } catch (Exception e) {
            System.out.println("  Cannot access Unsafe: " + e.getMessage());
            System.out.println("  Try: java --add-opens java.base/sun.misc=ALL-UNNAMED ...");
            return null;
        }
    }

    // ========================================================
    // SAMPLE CLASSES FOR DEMOS
    // ========================================================

    static class Secret {
        private final int answer = 42;
        private String message = "You shouldn't see this";

        private Secret() {
            // Private constructor — normally can't instantiate
            System.out.println("    Constructor called!");
        }

        @Override
        public String toString() {
            return "Secret{answer=" + answer + ", message='" + message + "'}";
        }
    }

    static class Counter {
        volatile int value = 0;
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) throws Exception {

        System.out.println("=== CHAPTER 58: UNSAFE & OFF-HEAP MEMORY ===\n");

        // ====================================================
        // Since Unsafe is internal API and may not be accessible
        // on all JVM configurations, we demonstrate concepts
        // and show what the code does.
        // ====================================================

        // --- 1. DirectByteBuffer — The Standard Way ---
        System.out.println("--- 1. Off-Heap via DirectByteBuffer (Standard API) ---\n");

        /*
         * ByteBuffer.allocateDirect() allocates memory OUTSIDE the heap.
         * This is the supported way to use off-heap memory.
         *
         * Internally, it uses Unsafe.allocateMemory()!
         *
         * Benefits of off-heap:
         *   - Not subject to GC pauses
         *   - Can be larger than heap
         *   - Can be memory-mapped to files
         *   - Zero-copy I/O possible (DMA)
         */

        // Allocate 1MB off-heap
        ByteBuffer direct = ByteBuffer.allocateDirect(1024 * 1024);
        ByteBuffer heap = ByteBuffer.allocate(1024 * 1024);

        System.out.println("  Direct buffer: isDirect=" + direct.isDirect());
        System.out.println("  Heap buffer:   isDirect=" + heap.isDirect());

        // Write and read
        direct.putInt(0, 42);
        direct.putDouble(4, 3.14159);
        direct.put(12, (byte) 0xFF);

        System.out.println("  Read int: " + direct.getInt(0));
        System.out.println("  Read double: " + direct.getDouble(4));
        System.out.println("  Read byte: " + (direct.get(12) & 0xFF));

        // Performance comparison
        int SIZE = 1_000_000;
        long start, elapsed;

        start = System.nanoTime();
        for (int i = 0; i < SIZE; i++) direct.putInt((i % 256) * 4, i);
        elapsed = System.nanoTime() - start;
        System.out.println("  Direct write " + SIZE + "x: " + elapsed / 1_000 + "µs");

        start = System.nanoTime();
        for (int i = 0; i < SIZE; i++) heap.putInt((i % 256) * 4, i);
        elapsed = System.nanoTime() - start;
        System.out.println("  Heap write " + SIZE + "x:   " + elapsed / 1_000 + "µs");

        // --- 2. What Unsafe Operations Look Like ---
        System.out.println("\n--- 2. Unsafe Operations (Conceptual) ---\n");

        System.out.println("  // Direct memory allocation (like C's malloc)");
        System.out.println("  long address = unsafe.allocateMemory(1024);");
        System.out.println("  unsafe.putInt(address, 42);       // write int at address");
        System.out.println("  int val = unsafe.getInt(address);  // read int from address");
        System.out.println("  unsafe.freeMemory(address);        // like C's free()");
        System.out.println("  // DANGER: forget to free → memory leak (no GC for off-heap!)");

        System.out.println("\n  // Object without constructor");
        System.out.println("  Secret s = (Secret) unsafe.allocateInstance(Secret.class);");
        System.out.println("  // Constructor is NEVER called! Fields have default values.");
        System.out.println("  // This is how Mockito creates mock objects.");

        System.out.println("\n  // CAS (Compare-And-Swap) on raw fields");
        System.out.println("  long offset = unsafe.objectFieldOffset(");
        System.out.println("      Counter.class.getDeclaredField(\"value\"));");
        System.out.println("  unsafe.compareAndSwapInt(counter, offset, expected, newValue);");
        System.out.println("  // This is how AtomicInteger works internally!");

        // --- 3. How AtomicInteger Uses Unsafe ---
        System.out.println("\n--- 3. AtomicInteger Internals (Uses Unsafe) ---\n");

        /*
         * AtomicInteger source code (simplified):
         *
         * public class AtomicInteger {
         *     private static final Unsafe unsafe = Unsafe.getUnsafe();
         *     private static final long valueOffset;
         *
         *     static {
         *         valueOffset = unsafe.objectFieldOffset(
         *             AtomicInteger.class.getDeclaredField("value"));
         *     }
         *
         *     private volatile int value;
         *
         *     public final int getAndIncrement() {
         *         return unsafe.getAndAddInt(this, valueOffset, 1);
         *         // Internally: CAS loop
         *         // do {
         *         //     current = getIntVolatile(obj, offset);
         *         // } while (!compareAndSwapInt(obj, offset, current, current + 1));
         *     }
         * }
         */

        AtomicInteger ai = new AtomicInteger(10);
        System.out.println("  AtomicInteger.getAndIncrement():");
        System.out.println("    Before: " + ai.get());
        int old = ai.getAndIncrement();
        System.out.println("    getAndIncrement returned: " + old);
        System.out.println("    After: " + ai.get());

        boolean casOk = ai.compareAndSet(11, 99);
        System.out.println("    CAS 11→99: " + casOk + ", value=" + ai.get());

        // --- 4. Memory Layout ---
        System.out.println("\n--- 4. Object Memory Layout ---\n");

        /*
         * Every Java object in memory:
         * ┌──────────────────────────────────────────┐
         * │ Mark Word (8 bytes on 64-bit)             │ → hash, GC age, lock state
         * │ Klass Pointer (4/8 bytes)                 │ → pointer to Class metadata
         * │ [Array Length (4 bytes)] — arrays only     │
         * │ Field 1                                    │
         * │ Field 2                                    │
         * │ ... (with alignment padding)               │
         * └──────────────────────────────────────────┘
         *
         * OBJECT HEADER = Mark Word + Klass Pointer
         *   12 bytes with compressed oops (default)
         *   16 bytes without compression
         *
         * ALIGNMENT: Objects are 8-byte aligned
         *   new Object() = 12-byte header → padded to 16 bytes
         *
         * FIELD ORDERING: Not necessarily declaration order!
         *   JVM reorders fields to minimize padding:
         *   - longs/doubles first (8 bytes)
         *   - ints/floats next (4 bytes)
         *   - shorts/chars (2 bytes)
         *   - bytes/booleans (1 byte)
         *   - references (4/8 bytes)
         *
         * USE JOL to see actual layout:
         *   org.openjdk.jol:jol-core
         *   System.out.println(ClassLayout.parseClass(MyClass.class).toPrintable());
         */

        System.out.println("  64-bit JVM with compressed oops (default):");
        System.out.println("    Object header:  12 bytes (mark + klass)");
        System.out.println("    Minimum object: 16 bytes (header + padding)");
        System.out.println("    int field:      +4 bytes");
        System.out.println("    long field:     +8 bytes");
        System.out.println("    reference:      +4 bytes (compressed)");
        System.out.println("    boolean:        +1 byte (but often padded)");
        System.out.println();
        System.out.println("  Example sizes:");
        System.out.println("    new Object()         → 16 bytes");
        System.out.println("    new Integer(42)      → 16 bytes (header + int)");
        System.out.println("    new Long(42L)        → 24 bytes (header + padding + long)");
        System.out.println("    new int[0]           → 16 bytes (header + length)");
        System.out.println("    new int[10]          → 56 bytes (16 + 10*4)");
        System.out.println("    new String(\"hello\") → 40+ bytes (header + ref + hash + coder + byte[])");

        // --- 5. Park/Unpark ---
        System.out.println("\n--- 5. Park/Unpark (Thread Scheduling) ---\n");

        /*
         * LockSupport.park() / unpark() use Unsafe internally.
         *
         * Unlike wait/notify:
         *   - No need for synchronized block
         *   - unpark() before park() = park() returns immediately
         *   - Per-thread permit (binary semaphore)
         *   - Can't cause "missed signal" bugs
         *
         * This is how ReentrantLock, CountDownLatch, etc. work.
         *
         * Unsafe.park(boolean isAbsolute, long time)
         * Unsafe.unpark(Thread thread)
         */

        Thread parkedThread = new Thread(() -> {
            System.out.println("    Thread parking...");
            java.util.concurrent.locks.LockSupport.park();
            System.out.println("    Thread unparked!");
        });
        parkedThread.start();
        Thread.sleep(100);
        java.util.concurrent.locks.LockSupport.unpark(parkedThread);
        parkedThread.join();

        // --- 6. Throwing Checked Exceptions Sneakily ---
        System.out.println("\n--- 6. Sneaky Throws ---\n");

        /*
         * Unsafe.throwException(Throwable) throws a checked exception
         * WITHOUT declaring it. This breaks Java's checked exception system.
         *
         * This is how Lombok's @SneakyThrows works:
         *   @SneakyThrows
         *   void myMethod() {
         *       throw new IOException("boom"); // no 'throws' clause!
         *   }
         *
         * You can also do it with generics erasure:
         */

        System.out.println("  Unsafe.throwException() throws checked without declaring.");
        System.out.println("  Also achievable via generics erasure:");
        System.out.println("    @SuppressWarnings(\"unchecked\")");
        System.out.println("    static <E extends Throwable> void sneaky(Throwable t) throws E {");
        System.out.println("        throw (E) t;  // erased to Throwable at runtime");
        System.out.println("    }");

        // --- 7. The Future ---
        System.out.println("\n--- 7. The Future: Replacing Unsafe ---\n");

        System.out.println("  Java is gradually providing supported alternatives:");
        System.out.println();
        System.out.println("  ┌──────────────────────┬──────────────────────────────────┐");
        System.out.println("  │ Unsafe Feature        │ Supported Replacement            │");
        System.out.println("  ├──────────────────────┼──────────────────────────────────┤");
        System.out.println("  │ CAS operations        │ VarHandle (Java 9+)              │");
        System.out.println("  │ Memory fences          │ VarHandle acquire/release         │");
        System.out.println("  │ Off-heap memory        │ Foreign Memory API (Java 22+)    │");
        System.out.println("  │ Field offsets           │ VarHandle                        │");
        System.out.println("  │ Park/unpark             │ LockSupport (always was)         │");
        System.out.println("  │ allocateInstance        │ No supported replacement yet     │");
        System.out.println("  │ Native function calls   │ Foreign Function API (Java 22+)  │");
        System.out.println("  └──────────────────────┴──────────────────────────────────┘");

        System.out.println("\n  JDK 23+: Unsafe is being deprecated for removal.");
        System.out.println("  Libraries must migrate to VarHandle and Panama (FFM API).");

        System.out.println("\n✓ Unsafe & Off-Heap Memory Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Use DirectByteBuffer to implement a simple off-heap ring buffer
 *    (fixed-size circular buffer with position tracking).
 * 2. Use VarHandle (Chapter 51) to implement a compare-and-swap counter
 *    without AtomicInteger (this is the "supported" way to do what Unsafe does).
 * 3. Read the source of AtomicInteger (OpenJDK) and trace how it uses Unsafe.
 * 4. Use JOL (Java Object Layout) to inspect the memory layout of your
 *    classes: add org.openjdk.jol:jol-core dependency and use
 *    ClassLayout.parseClass(YourClass.class).toPrintable()
 *
 * NEXT: Chapter 59 — Annotation Processing
 */
