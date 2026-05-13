/*
 * ============================================================
 *  CHAPTER 62: JVM INTERNALS DEEP DIVE
 * ============================================================
 *  THE TRUE FINAL CHAPTER. This is what separates someone who
 *  "knows Java" from someone who UNDERSTANDS Java. The JVM is
 *  the foundation everything else is built on.
 *
 *  TOPICS:
 *    1. JVM Architecture Overview
 *    2. Class Loading in Detail
 *    3. Runtime Data Areas (Memory Layout)
 *    4. String Pool & Interning
 *    5. Integer Cache & Autoboxing Traps
 *    6. Object Header & Mark Word
 *    7. Synchronization Internals (Biased → Thin → Fat Locks)
 *    8. Garbage Collection Internals
 *    9. JIT Compilation Tiers
 *   10. Class Data Sharing (CDS)
 *   11. JVM Startup & Shutdown
 *   12. Everything-You-Must-Know Summary
 * ============================================================
 */

import java.util.*;

public class Chapter62_JVMInternals {

    // ========================================================
    // 1. JVM ARCHITECTURE
    // ========================================================
    /*
     * ┌───────────────────────────────────────────────────────────┐
     * │                       JVM                                 │
     * │                                                           │
     * │  ┌──────────────────────────────────────────────────┐    │
     * │  │           Class Loading Subsystem                 │    │
     * │  │  Loading → Linking (Verify→Prepare→Resolve) →     │    │
     * │  │  Initialization                                   │    │
     * │  └──────────────────────────────────────────────────┘    │
     * │                         ↓                                 │
     * │  ┌──────────────────────────────────────────────────┐    │
     * │  │           Runtime Data Areas                       │    │
     * │  │  ┌────────┐ ┌────────┐ ┌──────────┐              │    │
     * │  │  │ Heap   │ │ Stack  │ │ Metaspace│              │    │
     * │  │  │(shared)│ │(per    │ │ (shared) │              │    │
     * │  │  │        │ │thread) │ │          │              │    │
     * │  │  └────────┘ └────────┘ └──────────┘              │    │
     * │  │  ┌────────────────┐ ┌────────────────┐           │    │
     * │  │  │ PC Register    │ │ Native Method  │           │    │
     * │  │  │ (per thread)   │ │ Stack          │           │    │
     * │  │  └────────────────┘ └────────────────┘           │    │
     * │  └──────────────────────────────────────────────────┘    │
     * │                         ↓                                 │
     * │  ┌──────────────────────────────────────────────────┐    │
     * │  │           Execution Engine                         │    │
     * │  │  Interpreter → JIT Compiler (C1/C2) → GC          │    │
     * │  └──────────────────────────────────────────────────┘    │
     * │                         ↓                                 │
     * │  ┌──────────────────────────────────────────────────┐    │
     * │  │           Native Interface (JNI)                   │    │
     * │  │  ↔ Native Libraries (.so/.dll/.dylib)              │    │
     * │  └──────────────────────────────────────────────────┘    │
     * └───────────────────────────────────────────────────────────┘
     */

    // ========================================================
    // 2. CLASS LOADING IN DETAIL
    // ========================================================
    /*
     * Three phases of class loading:
     *
     * LOADING:
     *   1. Find the .class file (ClassLoader.findClass)
     *   2. Read the bytes
     *   3. Create a Class object in Metaspace
     *
     * LINKING has three sub-phases:
     *   VERIFICATION:
     *     - Is the bytecode valid? (magic number 0xCAFEBABE)
     *     - Are types correct?
     *     - Are method signatures valid?
     *     - Stack overflow checks (verifier checks frame sizes)
     *     - No illegal access to private members of other classes
     *     - This is why Java is safe even if you hand-craft bytecode
     *
     *   PREPARATION:
     *     - Allocate memory for static fields
     *     - Set to DEFAULT values (0, false, null)
     *     - NOT the programmed values yet!
     *
     *   RESOLUTION:
     *     - Replace symbolic references with direct references
     *     - "java.lang.String" → actual pointer to String class
     *     - Can be lazy (resolved when first used)
     *
     * INITIALIZATION:
     *   - Run static initializer blocks: static { ... }
     *   - Assign static field values
     *   - This is when YOUR code runs for the first time
     *   - JVM guarantees this happens EXACTLY ONCE per class
     *   - Thread-safe (JVM uses internal lock per class)
     *   - This is why static singletons work without synchronization!
     *
     * WHEN DOES INITIALIZATION HAPPEN?
     *   - new instance created
     *   - Static method called
     *   - Static field accessed (except final compile-time constants)
     *   - Reflection (Class.forName)
     *   - Subclass initialized (triggers parent first)
     *   - Main class at JVM start
     *
     * DOES NOT trigger init:
     *   - Accessing a compile-time constant (static final int X = 42)
     *   - Creating an array of the type
     *   - Referencing the Class object without forName
     */

    // ========================================================
    // 3. RUNTIME DATA AREAS
    // ========================================================
    /*
     * HEAP (shared across all threads):
     *   - ALL objects live here (except escape-analyzed ones)
     *   - Divided into generations (for GC):
     *     ┌──────────────────────────────────────────┐
     *     │ Young Generation                          │
     *     │  ┌──────┐ ┌──────────┐ ┌──────────┐     │
     *     │  │ Eden │ │Survivor 0│ │Survivor 1│     │
     *     │  └──────┘ └──────────┘ └──────────┘     │
     *     ├──────────────────────────────────────────┤
     *     │ Old Generation (Tenured)                  │
     *     │  Objects that survived many GC cycles     │
     *     └──────────────────────────────────────────┘
     *
     *   - New objects → Eden
     *   - Survived minor GC → Survivor (S0↔S1, ping-pong)
     *   - Survived many minor GCs → Old gen (tenured)
     *   - -Xms/-Xmx control heap size
     *
     * STACK (per thread):
     *   - One stack per thread
     *   - Contains FRAMES (one per method call)
     *   - Each frame has:
     *     - Local Variables Array (params + locals)
     *     - Operand Stack (computation)
     *     - Constant Pool Reference
     *   - Stack size controlled by -Xss (default 512K-1M)
     *   - StackOverflowError when stack is full
     *
     * METASPACE (replaces PermGen since Java 8):
     *   - Class metadata (bytecode, method data, symbols)
     *   - Constant pool (resolved)
     *   - Method bytecode
     *   - Annotations
     *   - Uses NATIVE memory (not heap)
     *   - Grows dynamically (controlled by -XX:MaxMetaspaceSize)
     *   - ClassLoader GC cleans up classes when loader is collected
     *
     * PC REGISTER (per thread):
     *   - Points to current bytecode instruction
     *   - Undefined during native method execution
     *
     * NATIVE METHOD STACK (per thread):
     *   - Stack for native (C/C++) method calls
     *   - Separate from Java stack
     */

    // ========================================================
    // 4. STRING POOL & INTERNING
    // ========================================================

    static void demoStringPool() {
        System.out.println("--- 4. String Pool & Interning ---\n");

        // String literals are INTERNED — stored in the String Pool
        String a = "hello";        // → string pool
        String b = "hello";        // → SAME object from pool
        String c = new String("hello");  // → new object on HEAP (not pool)
        String d = c.intern();     // → retrieves from pool (same as a)

        System.out.println("  a == b?    " + (a == b) + "   (both from pool)");
        System.out.println("  a == c?    " + (a == c) + "  (c is on heap)");
        System.out.println("  a == d?    " + (a == d) + "   (d.intern() → pool)");
        System.out.println("  a.equals(c)? " + a.equals(c) + "   (content is same)");

        /*
         * STRING POOL HISTORY:
         *   Java 6:  String pool in PermGen (fixed, small) — dangerous!
         *   Java 7+: String pool moved to HEAP — can grow, GC'd
         *   Java 9+: String backed by byte[] (not char[]) — "Compact Strings"
         *            ASCII strings use 1 byte/char (LATIN1 encoding)
         *            Non-ASCII use 2 bytes/char (UTF16 encoding)
         *            → ~50% memory savings for ASCII-heavy apps
         *
         * -XX:StringTableSize=60013 (default ~60K buckets)
         * Use -XX:+PrintStringTableStatistics to see usage
         */

        // Compile-time constant folding
        String s1 = "hello" + " " + "world";  // → single "hello world" at COMPILE time
        String s2 = "hello world";
        System.out.println("  s1 == s2?  " + (s1 == s2) + "   (compile-time concatenation)");

        String part = "world";
        String s3 = "hello " + part;  // → runtime concatenation (NOT pooled)
        System.out.println("  s2 == s3?  " + (s2 == s3) + "  (runtime concat → new object)");

        final String constPart = "world";
        String s4 = "hello " + constPart;  // → compile-time (constPart is final!)
        System.out.println("  s2 == s4?  " + (s2 == s4) + "   (final → compile-time constant)");
        System.out.println();
    }

    // ========================================================
    // 5. INTEGER CACHE & AUTOBOXING TRAPS
    // ========================================================

    static void demoIntegerCache() {
        System.out.println("--- 5. Integer Cache ---\n");

        /*
         * Integer.valueOf(int) caches values from -128 to 127.
         * Autoboxing uses valueOf(), so:
         *   Integer a = 127; → Integer.valueOf(127) → cached
         *   Integer b = 127; → same cached object!
         *
         * This is why == works for small integers but FAILS for large ones.
         * THIS IS ONE OF JAVA'S MOST COMMON TRAPS.
         */

        Integer a = 127;       // cached
        Integer b = 127;       // SAME object
        Integer c = 128;       // NOT cached
        Integer d = 128;       // DIFFERENT object
        Integer e = -128;      // cached (lower bound)
        Integer f = -128;      // same object

        System.out.println("  127 == 127?   " + (a == b) + "  (cached)");
        System.out.println("  128 == 128?   " + (c == d) + " (NOT cached!)");
        System.out.println("  -128 == -128? " + (e == f) + "  (cached)");
        System.out.println("  ALWAYS use .equals() for Integer comparison!");

        /*
         * OTHER CACHED VALUES:
         *   Byte:      all values (-128 to 127) — entire range!
         *   Short:     -128 to 127
         *   Long:      -128 to 127
         *   Character: 0 to 127
         *   Boolean:   TRUE and FALSE (only 2 values)
         *   Float:     NOT cached
         *   Double:    NOT cached
         *
         * Tunable: -XX:AutoBoxCacheMax=<size> (Integer only)
         *   java -XX:AutoBoxCacheMax=1000 → caches 0-1000
         */

        // Autoboxing performance trap
        System.out.println("\n  Autoboxing performance trap:");
        System.out.println("    Long sum = 0L;");
        System.out.println("    for (long i = 0; i < 1_000_000; i++) sum += i;");
        System.out.println("    → Creates ~1,000,000 Long objects! (autoboxing in +=)");
        System.out.println("    → Fix: use primitive 'long sum = 0L;'");
        System.out.println();
    }

    // ========================================================
    // 6. OBJECT HEADER & MARK WORD
    // ========================================================
    /*
     * Every Java object has a header:
     *
     * 64-bit JVM, compressed oops (default):
     *  ┌──────────────────────────────────────┐
     *  │ Mark Word (8 bytes)                   │
     *  │  - hash code (25 bits)               │
     *  │  - GC age (4 bits, max 15)           │
     *  │  - biased lock flag (1 bit)          │
     *  │  - lock state (2 bits)               │
     *  │  - thread ID (for biased locking)    │
     *  ├──────────────────────────────────────┤
     *  │ Klass Pointer (4 bytes, compressed)   │
     *  │  → Points to class metadata           │
     *  ├──────────────────────────────────────┤
     *  │ [Array Length (4 bytes)] — arrays only │
     *  └──────────────────────────────────────┘
     *
     * MARK WORD STATES (64-bit):
     *  ┌────────────────────────────────────────────────────────┐
     *  │ State              │ Content of Mark Word               │
     *  ├────────────────────┼────────────────────────────────────┤
     *  │ Unlocked           │ hash:31 | age:4 | biased:1 | 01  │
     *  │ Biased lock        │ thread:54 | epoch:2 | age:4 | 101│
     *  │ Thin (lightweight) │ stack ptr:62              | 00    │
     *  │ Fat (heavyweight)  │ monitor ptr:62            | 10    │
     *  │ GC marked          │ forwarding ptr:62         | 11    │
     *  └────────────────────┴────────────────────────────────────┘
     *
     * GC AGE: incremented on each minor GC survival.
     *   Default threshold=15 → promoted to old gen.
     *   -XX:MaxTenuringThreshold=15 (max is 15, 4 bits)
     *
     * IDENTITY HASH CODE:
     *   Computed LAZILY on first call to System.identityHashCode()
     *   Stored in mark word
     *   Once hash is computed, biased locking can't be used!
     *   (hash bits and thread ID bits overlap)
     */

    // ========================================================
    // 7. SYNCHRONIZATION INTERNALS
    // ========================================================
    /*
     * When you write 'synchronized (obj)', the JVM doesn't
     * immediately use an OS mutex. It uses LOCK ESCALATION:
     *
     * LEVEL 1: BIASED LOCKING (disabled by default since Java 15)
     *   - First thread to lock biases the object to itself
     *   - Subsequent locks by SAME thread are nearly free (no atomic op)
     *   - If another thread tries to lock, bias is REVOKED → thin lock
     *   - Disabled: -XX:-UseBiasedLocking (default since Java 15)
     *   - Removed in Java 18+
     *
     * LEVEL 2: THIN LOCK (Lightweight/Stack Lock)
     *   - CAS to store lock record pointer in mark word
     *   - If CAS succeeds: locked with minimal overhead
     *   - If CAS fails: brief spin (adaptive spinning)
     *   - If spin fails: inflate to fat lock
     *
     * LEVEL 3: FAT LOCK (Heavyweight/Monitor)
     *   - Full OS mutex (pthread_mutex on Linux)
     *   - Thread is put to sleep (context switch = expensive)
     *   - This is the worst case
     *
     * ADAPTIVE SPINNING:
     *   - Before inflating, spin a few times (busy-wait)
     *   - If the lock was recently held briefly, spin more
     *   - If the lock is usually held long, skip spinning
     *   - JVM learns from history!
     *
     * LOCK COARSENING:
     *   synchronized (x) { a(); }
     *   synchronized (x) { b(); }
     *   → Merged to: synchronized (x) { a(); b(); }
     *
     * LOCK ELISION (via escape analysis):
     *   void foo() {
     *       Object lock = new Object();  // doesn't escape
     *       synchronized (lock) { ... }  // REMOVED by JIT
     *   }
     */

    // ========================================================
    // 8. GARBAGE COLLECTION INTERNALS
    // ========================================================
    /*
     * GC ROOT SCAN — What keeps objects alive:
     *   1. Local variables on thread stacks
     *   2. Active threads themselves
     *   3. Static fields of loaded classes
     *   4. JNI references
     *   5. Internal JVM references (class objects, etc.)
     *   6. Synchronized monitors
     *
     * GC ALGORITHMS:
     *
     * MARK-SWEEP-COMPACT:
     *   1. Mark: traverse from roots, mark all reachable objects
     *   2. Sweep: free unmarked objects
     *   3. Compact: move surviving objects together (defragment)
     *
     * COPYING COLLECTOR (used for young gen):
     *   1. Copy all live objects from Eden+S0 to S1
     *   2. Eden+S0 are now completely free
     *   3. Swap S0 and S1 labels
     *   Very fast for young gen (most objects die young)
     *
     * SAFEPOINTS:
     *   GC can only run at "safepoints" — specific locations where
     *   the thread state is known and consistent.
     *   - Method returns
     *   - Loop back-edges (not counted loops though!)
     *   - JNI call returns
     *
     *   All threads must reach a safepoint before GC can proceed.
     *   "Time-To-Safepoint" (TTSP) is a key latency metric.
     *   A thread stuck in a long counted loop delays GC for everyone!
     *
     *   -XX:+PrintSafepointStatistics — see safepoint delays
     *
     * CARD TABLE (for generational GC):
     *   Problem: if old gen object → young gen object,
     *   we'd need to scan ALL of old gen during minor GC.
     *   Solution: divide heap into 512-byte "cards."
     *   When old gen writes a reference, mark the card "dirty."
     *   Minor GC only scans dirty cards in old gen.
     *
     * REMEMBERED SETS (G1GC):
     *   G1 divides heap into regions, not generations.
     *   Each region has a remembered set: which OTHER regions
     *   have pointers INTO this region.
     *   Only scan those during collection.
     *
     * TLAB (Thread-Local Allocation Buffer):
     *   Each thread gets a private chunk of Eden.
     *   Allocation = pointer bump (no synchronization!).
     *   When TLAB fills up, get a new one (requires sync).
     *   This is why Java object allocation is so fast (~10ns).
     */

    // ========================================================
    // 9. JIT COMPILATION TIERS
    // ========================================================
    /*
     * TIERED COMPILATION (default since Java 8):
     *
     *   Tier 0: Interpreter
     *     → Execute bytecode directly
     *     → No compilation cost
     *     → Slowest execution
     *     → Collects basic profiling data
     *
     *   Tier 1: C1 with full optimization (no profiling)
     *     → Simple methods, trivial
     *     → Doesn't collect profile info
     *
     *   Tier 2: C1 with invocation counter
     *     → Collects limited profiling
     *
     *   Tier 3: C1 with full profiling
     *     → Collects detailed profiling data
     *     → This data feeds C2
     *
     *   Tier 4: C2 (Server compiler)
     *     → Aggressive optimizations based on profile
     *     → Inlining, escape analysis, vectorization
     *     → Loop unrolling, dead code elimination
     *     → Speculative optimizations (based on profile)
     *     → Can DEOPTIMIZE if speculation fails
     *
     * TYPICAL PATH: 0 → 3 → 4 (interpreter → C1 profiling → C2)
     *
     * COMPILATION THRESHOLDS:
     *   -XX:CompileThreshold=10000 (invocations before compile)
     *   With tiered: much lower thresholds (~hundreds)
     *
     * ON-STACK REPLACEMENT (OSR):
     *   If a method has a hot loop, the JVM can compile it
     *   and switch to compiled code IN THE MIDDLE of execution.
     *   The loop doesn't need to finish for compilation to kick in.
     *
     * DEOPTIMIZATION:
     *   JIT makes optimistic assumptions (this type is always X).
     *   If assumption is violated → deoptimize back to interpreter.
     *   Then re-profile and re-compile.
     *   Common triggers: unexpected class loading, type change.
     *
     * SEE COMPILATION:
     *   -XX:+PrintCompilation
     *   Output: timestamp compile_id tier method size
     *   Example: 134   42 %  3  MyClass::hotLoop @ 12 (45 bytes)
     *   % = OSR, 3 = tier 3 (C1 full profiling)
     */

    // ========================================================
    // 10. CLASS DATA SHARING (CDS)
    // ========================================================
    /*
     * CDS speeds up JVM startup by pre-loading class metadata.
     *
     * DEFAULT CDS (Java 12+):
     *   JDK ships with a shared archive of ~1200 core classes.
     *   No setup needed — it just works.
     *
     * APPLICATION CDS (AppCDS):
     *   1. Create a class list:
     *      java -Xshare:off -XX:DumpLoadedClassList=classes.lst -jar app.jar
     *
     *   2. Create the archive:
     *      java -Xshare:dump -XX:SharedClassListFile=classes.lst \
     *           -XX:SharedArchiveFile=app.jsa -jar app.jar
     *
     *   3. Use the archive:
     *      java -Xshare:on -XX:SharedArchiveFile=app.jsa -jar app.jar
     *
     *   → 30-50% faster startup for typical apps!
     *
     * DYNAMIC CDS (Java 13+):
     *   java -XX:ArchiveClassesAtExit=app.jsa -jar app.jar
     *   → Automatically dumps at shutdown
     *   → Next run: java -XX:SharedArchiveFile=app.jsa -jar app.jar
     *
     * HOW IT WORKS:
     *   - Class metadata is memory-mapped from the archive file
     *   - Shared across multiple JVM instances (read-only)
     *   - Skips class loading, verification, parsing
     *   - Reduces memory footprint too (shared pages)
     */

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== CHAPTER 62: JVM INTERNALS DEEP DIVE ===\n");

        demoStringPool();
        demoIntegerCache();

        // --- Class initialization order ---
        System.out.println("--- 6. Class Initialization Order ---\n");

        System.out.println("  Order when 'new Child()' is called:");
        System.out.println("    1. Parent static fields (declaration order)");
        System.out.println("    2. Parent static block");
        System.out.println("    3. Child static fields");
        System.out.println("    4. Child static block");
        System.out.println("    5. Parent instance fields");
        System.out.println("    6. Parent instance block");
        System.out.println("    7. Parent constructor");
        System.out.println("    8. Child instance fields");
        System.out.println("    9. Child instance block");
        System.out.println("   10. Child constructor");
        System.out.println("  Static init happens ONCE; instance init every new().");

        // --- Demonstrating identity hashCode and mark word ---
        System.out.println("\n--- 7. Identity Hash Code ---\n");

        Object obj = new Object();
        int hash1 = System.identityHashCode(obj);
        int hash2 = System.identityHashCode(obj);
        System.out.println("  identityHashCode: " + hash1);
        System.out.println("  Same on 2nd call: " + (hash1 == hash2) + " (stored in mark word)");

        // hashCode() vs identityHashCode()
        String s = new String("test");
        System.out.println("  String.hashCode():     " + s.hashCode() + " (content-based)");
        System.out.println("  identityHashCode(str): " + System.identityHashCode(s) + " (object id)");

        // --- GC Roots demo ---
        System.out.println("\n--- 8. GC Root Demonstration ---\n");

        Object root1 = new Object(); // stack variable → GC root
        List<Object> root2 = new ArrayList<>(); // stack variable → GC root
        root2.add(new Object()); // reachable via root2 → alive

        System.out.println("  root1 on stack → alive");
        System.out.println("  root2.get(0) reachable through root2 → alive");
        root2 = null; // root2 and its contents now unreachable
        System.out.println("  root2 = null → root2 and its elements eligible for GC");

        // --- TLAB demo ---
        System.out.println("\n--- 9. TLAB (Thread-Local Allocation Buffer) ---\n");

        System.out.println("  Object allocation in Java is ~10ns because:");
        System.out.println("  1. Each thread gets a private chunk of Eden (TLAB)");
        System.out.println("  2. Allocation = pointer bump (currentPtr += size)");
        System.out.println("  3. No synchronization needed (it's thread-local!)");
        System.out.println("  4. When TLAB fills, get new one (rare, needs sync)");
        System.out.println("  5. -XX:TLABSize=<size> to control (rarely needed)");

        // --- Safepoints ---
        System.out.println("\n--- 10. Safepoints ---\n");

        System.out.println("  GC can only happen at safepoints.");
        System.out.println("  Safepoint locations:");
        System.out.println("    ✅ Method return");
        System.out.println("    ✅ Loop back-edges (non-counted loops)");
        System.out.println("    ✅ JNI call boundaries");
        System.out.println("    ❌ Counted loops: for(int i=0; i<n; i++) — NO safepoint!");
        System.out.println("          → A long counted loop can delay GC for everyone!");
        System.out.println("          → Fix: -XX:+UseCountedLoopSafepoints (Java 14+)");
        System.out.println("          → Or use long loop variable: for(long i=0; ...)");

        // --- Everything Summary ---
        System.out.println("\n" + "=".repeat(60));
        System.out.println("  COMPLETE JVM KNOWLEDGE MAP");
        System.out.println("=".repeat(60));
        System.out.println();
        System.out.println("  MEMORY:");
        System.out.println("    Heap (Eden → Survivor → Old) — objects");
        System.out.println("    Metaspace — class metadata (native memory)");
        System.out.println("    Stack — frames per thread (locals + operand)");
        System.out.println("    TLAB — fast allocation per thread");
        System.out.println("    String Pool — interned strings (on heap since Java 7)");
        System.out.println("    Code Cache — JIT compiled code");
        System.out.println("    Direct Memory — ByteBuffer.allocateDirect");
        System.out.println();
        System.out.println("  CLASS LOADING:");
        System.out.println("    Bootstrap → Platform → Application → Custom");
        System.out.println("    Load → Verify → Prepare → Resolve → Initialize");
        System.out.println("    Class identity = ClassLoader + FQN");
        System.out.println();
        System.out.println("  EXECUTION:");
        System.out.println("    Interpreter → C1 (quick) → C2 (optimized)");
        System.out.println("    Inlining, Escape Analysis, Scalar Replacement");
        System.out.println("    OSR for hot loops, Deoptimization on bad speculation");
        System.out.println();
        System.out.println("  GC:");
        System.out.println("    Mark-Sweep-Compact, Copying (young gen)");
        System.out.println("    Card Table, Remembered Sets, Safepoints");
        System.out.println("    G1 (default), ZGC (low-latency), Shenandoah");
        System.out.println();
        System.out.println("  SYNCHRONIZATION:");
        System.out.println("    Biased → Thin (CAS) → Fat (OS mutex)");
        System.out.println("    Adaptive spinning, lock coarsening, lock elision");
        System.out.println();
        System.out.println("  OBJECTS:");
        System.out.println("    Header: Mark Word (8B) + Klass Pointer (4B)");
        System.out.println("    Fields ordered by size for minimal padding");
        System.out.println("    8-byte alignment, compressed oops by default");

        System.out.println("\n" + "=".repeat(60));
        System.out.println("  ALL 62 CHAPTERS COMPLETE.");
        System.out.println("  YOU HAVE MASTERED THE JAVA LANGUAGE AND JVM.");
        System.out.println("=".repeat(60));
        System.out.println();
        System.out.println("  From \"Hello World\" to JVM internals —");
        System.out.println("  there is no Java language topic left uncovered.");
        System.out.println();
        System.out.println("  ✓ The Java Language & JVM: MASTERED.");
    }
}

/*
 * EXERCISES:
 * 1. Use -XX:+PrintCompilation to watch JIT compile your code.
 *    Run a hot loop and identify which tier it compiles at.
 * 2. Demonstrate the Integer cache trap: create a bug where
 *    == works for small numbers but fails for large ones.
 * 3. Use String.intern() to reduce memory usage in a program
 *    that loads millions of duplicate strings from a file.
 * 4. Set up AppCDS for a Java application and measure the
 *    startup time improvement.
 * 5. Read OpenJDK source for HashMap.put() — understand
 *    treeification, hash spreading, and resize mechanics.
 *
 * THE END. FOR REAL THIS TIME.
 */
