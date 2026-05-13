/*
 * ============================================================
 *  CHAPTER 33: JVM INTERNALS
 * ============================================================
 *  This chapter teaches HOW Java works under the hood.
 *  Understanding the JVM makes you a truly advanced developer.
 *
 *  TOPICS:
 *    1. JVM Architecture (ClassLoader → Runtime Data Areas → Execution Engine)
 *    2. Memory Model (Heap, Stack, Metaspace, etc.)
 *    3. Garbage Collection
 *    4. Class Loading
 *    5. JIT Compilation
 *    6. JVM Flags and Tuning
 *    7. Runtime inspection with code
 * ============================================================
 *
 *  JVM ARCHITECTURE:
 *  ┌─────────────────────────────────────────────┐
 *  │                  JVM                        │
 *  │  ┌──────────────┐  ┌────────────────────┐  │
 *  │  │ Class Loader  │  │ Execution Engine   │  │
 *  │  │  Bootstrap    │  │  Interpreter       │  │
 *  │  │  Extension    │  │  JIT Compiler      │  │
 *  │  │  Application  │  │  GC                │  │
 *  │  └──────────────┘  └────────────────────┘  │
 *  │  ┌─────────────────────────────────────┐   │
 *  │  │      Runtime Data Areas              │   │
 *  │  │  ┌─────────┐ ┌──────┐ ┌──────────┐ │   │
 *  │  │  │  HEAP   │ │STACK │ │METASPACE │ │   │
 *  │  │  │(objects)│ │(per  │ │(classes, │ │   │
 *  │  │  │         │ │thread│ │ methods) │ │   │
 *  │  │  └─────────┘ └──────┘ └──────────┘ │   │
 *  │  │  ┌────────────────┐ ┌────────────┐ │   │
 *  │  │  │ PC Registers   │ │ Native     │ │   │
 *  │  │  │ (per thread)   │ │ Method     │ │   │
 *  │  │  │                │ │ Stack      │ │   │
 *  │  │  └────────────────┘ └────────────┘ │   │
 *  │  └─────────────────────────────────────┘   │
 *  └─────────────────────────────────────────────┘
 *
 *  MEMORY AREAS:
 *  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 *  HEAP (shared among all threads)
 *    - Young Generation
 *      ├── Eden Space (new objects created here)
 *      ├── Survivor S0
 *      └── Survivor S1
 *    - Old Generation (long-lived objects promoted here)
 *
 *  STACK (one per thread)
 *    - Each method call creates a Stack Frame:
 *      ├── Local Variables
 *      ├── Operand Stack
 *      └── Frame Data (return address, etc.)
 *
 *  METASPACE (replaced PermGen in Java 8)
 *    - Class metadata, method data, constant pool
 *    - Uses native memory (not heap)
 *
 *  GARBAGE COLLECTION:
 *  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 *  Minor GC: cleans Young Gen (fast, frequent)
 *  Major GC: cleans Old Gen (slower, less frequent)
 *  Full GC:  cleans everything (stop-the-world, avoid!)
 *
 *  GC Algorithms:
 *    Serial GC      → single thread, small apps (-XX:+UseSerialGC)
 *    Parallel GC    → multiple threads, throughput (-XX:+UseParallelGC)
 *    G1 GC          → default since Java 9, balanced (-XX:+UseG1GC)
 *    ZGC            → ultra-low latency, Java 11+ (-XX:+UseZGC)
 *    Shenandoah     → low latency, concurrent (-XX:+UseShenandoahGC)
 *
 *  Object Lifecycle:
 *    1. new → allocated in Eden
 *    2. Survives Minor GC → moved to Survivor space
 *    3. Survives enough GCs → promoted to Old Gen
 *    4. No more references → eligible for GC
 *    5. GC reclaims memory
 *
 *  CLASS LOADING:
 *  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 *  Loading → Linking (Verify → Prepare → Resolve) → Initialization
 *
 *  ClassLoader Hierarchy:
 *    Bootstrap CL → loads rt.jar (core Java classes)
 *    Extension CL → loads ext/*.jar
 *    Application CL → loads classpath classes
 *    Custom CL → user-defined loaders
 *
 *  Delegation Model: child asks parent first (parent-first)
 *
 *  JIT COMPILATION:
 *  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 *  1. Bytecode interpreted initially
 *  2. Hot methods detected (invocation count)
 *  3. JIT compiles hot methods to native code
 *  4. C1 compiler (client) = fast compilation
 *  5. C2 compiler (server) = optimized compilation
 *  6. Tiered compilation: C1 → C2 (default since Java 8)
 *
 *  JIT Optimizations:
 *    - Inlining (small methods inserted directly)
 *    - Dead code elimination
 *    - Loop unrolling
 *    - Escape analysis (object may stay on stack)
 *    - Lock coarsening/elision
 * ============================================================
 */

import java.lang.management.*;
import java.lang.ref.*;
import java.util.*;

public class Chapter33_JVMInternals {

    // === For demonstrating GC ===
    static class HeavyObject {
        byte[] data = new byte[1024 * 100]; // 100KB
        String name;
        HeavyObject(String name) { this.name = name; }
        @Override
        protected void finalize() {
            // finalize is deprecated since Java 9, but shown for learning
            System.out.println("    GC collecting: " + name);
        }
    }

    public static void main(String[] args) {

        // --- 1. Runtime Memory Info ---
        System.out.println("=== RUNTIME MEMORY ===\n");

        Runtime rt = Runtime.getRuntime();
        long mb = 1024 * 1024;
        System.out.println("  Max Memory:   " + rt.maxMemory() / mb + " MB (Xmx)");
        System.out.println("  Total Memory: " + rt.totalMemory() / mb + " MB (current heap)");
        System.out.println("  Free Memory:  " + rt.freeMemory() / mb + " MB");
        System.out.println("  Used Memory:  " + (rt.totalMemory() - rt.freeMemory()) / mb + " MB");
        System.out.println("  Processors:   " + rt.availableProcessors());

        // --- 2. Memory Management Beans ---
        System.out.println("\n=== MEMORY BEANS ===\n");

        MemoryMXBean memBean = ManagementFactory.getMemoryMXBean();
        MemoryUsage heap = memBean.getHeapMemoryUsage();
        MemoryUsage nonHeap = memBean.getNonHeapMemoryUsage();

        System.out.println("  Heap:");
        System.out.println("    Init:      " + heap.getInit() / mb + " MB");
        System.out.println("    Used:      " + heap.getUsed() / mb + " MB");
        System.out.println("    Committed: " + heap.getCommitted() / mb + " MB");
        System.out.println("    Max:       " + heap.getMax() / mb + " MB");

        System.out.println("  Non-Heap (Metaspace, etc.):");
        System.out.println("    Used:      " + nonHeap.getUsed() / mb + " MB");

        // --- 3. Memory Pool Details ---
        System.out.println("\n=== MEMORY POOLS ===\n");
        for (MemoryPoolMXBean pool : ManagementFactory.getMemoryPoolMXBeans()) {
            System.out.printf("  %-30s type=%-8s used=%d KB%n",
                pool.getName(), pool.getType(), pool.getUsage().getUsed() / 1024);
        }

        // --- 4. GC Info ---
        System.out.println("\n=== GARBAGE COLLECTORS ===\n");
        for (GarbageCollectorMXBean gc : ManagementFactory.getGarbageCollectorMXBeans()) {
            System.out.println("  " + gc.getName());
            System.out.println("    Collections: " + gc.getCollectionCount());
            System.out.println("    Time spent:  " + gc.getCollectionTime() + " ms");
        }

        // --- 5. Triggering GC ---
        System.out.println("\n=== GC DEMONSTRATION ===\n");

        System.out.println("  Creating heavy objects...");
        for (int i = 0; i < 10; i++) {
            new HeavyObject("obj-" + i);  // no reference kept → eligible for GC
        }

        System.out.println("  Requesting GC...");
        System.gc();  // HINT only, JVM may ignore

        try { Thread.sleep(500); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }

        System.out.println("  After GC - Used: " +
            (rt.totalMemory() - rt.freeMemory()) / mb + " MB");

        // --- 6. Reference Types ---
        System.out.println("\n=== REFERENCE TYPES ===\n");

        // Strong Reference (normal) — object NOT collected while ref exists
        Object strong = new Object();
        System.out.println("  Strong ref: " + strong);

        // Weak Reference — collected at next GC
        WeakReference<HeavyObject> weak = new WeakReference<>(new HeavyObject("weak-obj"));
        System.out.println("  Weak ref before GC: " + (weak.get() != null));
        System.gc();
        try { Thread.sleep(100); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
        System.out.println("  Weak ref after GC:  " + (weak.get() != null));

        // Soft Reference — collected only when memory is low
        SoftReference<HeavyObject> soft = new SoftReference<>(new HeavyObject("soft-obj"));
        System.out.println("  Soft ref: " + (soft.get() != null) + " (kept until memory pressure)");

        // PhantomReference — can't access object, used for cleanup
        // Used with ReferenceQueue for post-mortem cleanup

        System.out.println("\n  Reference strength: Strong > Soft > Weak > Phantom");

        // --- 7. Stack vs Heap ---
        System.out.println("\n=== STACK vs HEAP ===\n");
        System.out.println("  STACK                          HEAP");
        System.out.println("  ─────                          ────");
        System.out.println("  Primitives & references        Objects (new keyword)");
        System.out.println("  Per thread                     Shared among threads");
        System.out.println("  LIFO (fast)                    GC managed (slower)");
        System.out.println("  Fixed size (-Xss)              Growable (-Xms/-Xmx)");
        System.out.println("  StackOverflowError             OutOfMemoryError");
        System.out.println("  Method frames                  Object instances");

        // --- 8. String Pool ---
        System.out.println("\n=== STRING POOL ===\n");
        String s1 = "hello";       // goes to string pool
        String s2 = "hello";       // reuses from pool
        String s3 = new String("hello");  // new object on heap
        String s4 = s3.intern();   // puts into pool / returns pool ref

        System.out.println("  s1 == s2: " + (s1 == s2) + " (both from pool)");
        System.out.println("  s1 == s3: " + (s1 == s3) + " (pool vs heap)");
        System.out.println("  s1 == s4: " + (s1 == s4) + " (pool vs intern)");

        // --- 9. Class Loading ---
        System.out.println("\n=== CLASS LOADING ===\n");

        ClassLoader cl = Chapter33_JVMInternals.class.getClassLoader();
        System.out.println("  App ClassLoader: " + cl);
        System.out.println("  Parent (Platform): " + cl.getParent());
        System.out.println("  Grandparent (Bootstrap): " + cl.getParent().getParent());

        // String is loaded by bootstrap ClassLoader
        System.out.println("  String loader: " + String.class.getClassLoader() + " (null = bootstrap)");

        // --- 10. Thread Info ---
        System.out.println("\n=== THREAD INFO ===\n");

        ThreadMXBean threadBean = ManagementFactory.getThreadMXBean();
        System.out.println("  Thread count: " + threadBean.getThreadCount());
        System.out.println("  Peak threads: " + threadBean.getPeakThreadCount());
        System.out.println("  Daemon threads: " + threadBean.getDaemonThreadCount());

        System.out.println("  Active threads:");
        for (long id : threadBean.getAllThreadIds()) {
            ThreadInfo info = threadBean.getThreadInfo(id);
            if (info != null) {
                System.out.println("    [" + id + "] " + info.getThreadName() + " - " + info.getThreadState());
            }
        }

        // --- 11. System Properties ---
        System.out.println("\n=== KEY SYSTEM PROPERTIES ===\n");
        String[] props = {"java.version", "java.vendor", "java.home",
            "os.name", "os.arch", "file.encoding", "java.class.path"};
        for (String p : props) {
            String val = System.getProperty(p);
            if (val != null && val.length() > 60) val = val.substring(0, 57) + "...";
            System.out.println("  " + p + " = " + val);
        }

        // --- 12. Important JVM Flags ---
        System.out.println("\n=== KEY JVM FLAGS ===");
        System.out.println("  -Xms256m          Initial heap size");
        System.out.println("  -Xmx1024m         Maximum heap size");
        System.out.println("  -Xss512k          Thread stack size");
        System.out.println("  -XX:+UseG1GC      Use G1 garbage collector");
        System.out.println("  -XX:+UseZGC       Use ZGC (low latency, Java 11+)");
        System.out.println("  -XX:MaxMetaspaceSize=256m   Metaspace limit");
        System.out.println("  -XX:+HeapDumpOnOutOfMemoryError  Dump heap on OOM");
        System.out.println("  -XX:+PrintGCDetails              GC details log");
        System.out.println("  -verbose:gc                      Basic GC logging");
        System.out.println("  -XX:+PrintCompilation            JIT compilation log");

        // --- 13. Memory Leak Patterns ---
        System.out.println("\n=== MEMORY LEAK PATTERNS ===");
        System.out.println("  1. Static collections that grow forever");
        System.out.println("  2. Unclosed resources (streams, connections)");
        System.out.println("  3. Listeners/callbacks not unregistered");
        System.out.println("  4. Inner classes holding outer class reference");
        System.out.println("  5. ThreadLocal not removed after use");
        System.out.println("  6. String.intern() abuse (fills string pool)");
        System.out.println("  7. ClassLoader leaks (common in app servers)");

        System.out.println("\n✓ JVM Internals Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Write a program that causes StackOverflowError. Measure default stack depth.
 * 2. Write a program that causes OutOfMemoryError. Use -Xmx to limit heap.
 * 3. Monitor GC activity: create/discard objects, observe GC beans.
 * 4. Create a custom ClassLoader that loads .class from a custom directory.
 *
 * NEXT: Chapter 34 — Design Patterns: Creational
 */
