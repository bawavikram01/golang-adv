/*
 * ============================================================
 *  CHAPTER 53: PERFORMANCE OPTIMIZATION & JMH
 * ============================================================
 *  "Premature optimization is the root of all evil" — Knuth
 *  But god-level Java means knowing WHAT to optimize, HOW the
 *  JVM optimizes for you, and HOW to measure properly.
 *
 *  TOPICS:
 *    1. Why Naive Benchmarks Are Wrong
 *    2. JMH (Java Microbenchmark Harness)
 *    3. JIT Compilation — What the JVM Does Behind Your Back
 *    4. Escape Analysis & Scalar Replacement
 *    5. Loop Optimizations (Unrolling, Vectorization)
 *    6. String Performance
 *    7. Collection Performance
 *    8. Object Creation & GC Pressure
 *    9. Cache-Friendly Code
 *   10. Common Anti-Patterns
 * ============================================================
 *
 *  JMH SETUP (when you want to run real benchmarks):
 *    mvn archetype:generate \
 *      -DinteractiveMode=false \
 *      -DarchetypeGroupId=org.openjdk.jmh \
 *      -DarchetypeArtifactId=jmh-java-benchmark-archetype \
 *      -DgroupId=org.example \
 *      -DartifactId=benchmarks \
 *      -Dversion=1.0
 *
 *  This chapter TEACHES the concepts. Real JMH requires a
 *  Maven/Gradle project. The code below demonstrates the
 *  concepts WITHOUT requiring JMH dependency.
 * ============================================================
 */

import java.util.*;
import java.util.concurrent.*;
import java.util.stream.*;

public class Chapter53_Performance {

    // ========================================================
    // 1. WHY NAIVE BENCHMARKS ARE WRONG
    // ========================================================
    /*
     * NEVER do this:
     *   long start = System.nanoTime();
     *   doSomething();
     *   long elapsed = System.nanoTime() - start;
     *   System.out.println("Took " + elapsed + "ns");
     *
     * WHY IT'S WRONG:
     *   1. JIT hasn't warmed up — first calls are interpreted
     *   2. JIT may ELIMINATE your code (dead code elimination)
     *   3. GC can pause during measurement
     *   4. OS scheduling noise
     *   5. No statistical significance (single measurement)
     *   6. Class loading happens on first use
     *   7. CPU caches are cold
     *
     * JMH handles ALL of these problems:
     *   - Warmup iterations (let JIT optimize)
     *   - Blackhole consumption (prevent dead code elimination)
     *   - Fork new JVMs (clean state)
     *   - Statistical analysis (mean, std dev, percentiles)
     *   - Proper iteration counting
     */

    // ========================================================
    // 2. JMH BENCHMARK STRUCTURE (Reference)
    // ========================================================
    /*
     * JMH ANNOTATIONS (for reference — can't run without JMH dep):
     *
     * @BenchmarkMode(Mode.AverageTime)    // what to measure
     *   - Mode.Throughput     → ops/second
     *   - Mode.AverageTime    → time/op
     *   - Mode.SampleTime     → sampling distribution
     *   - Mode.SingleShotTime → single call (cold start)
     *
     * @OutputTimeUnit(TimeUnit.NANOSECONDS)  // display unit
     *
     * @Warmup(iterations = 5, time = 1)      // warmup
     * @Measurement(iterations = 10, time = 1) // actual measurement
     * @Fork(2)                                // JVM forks
     *
     * @State(Scope.Thread)    // per-thread state
     * @State(Scope.Benchmark) // shared state
     *
     * public class MyBenchmark {
     *     @Benchmark
     *     public int testAdd(Blackhole bh) {
     *         int result = 1 + 2;
     *         bh.consume(result);  // prevent dead code elimination
     *         return result;       // or return it
     *     }
     * }
     *
     * RUN: java -jar benchmarks.jar
     */

    // ========================================================
    // 3. JIT COMPILER OPTIMIZATIONS
    // ========================================================
    /*
     * The JVM has TWO JIT compilers (tiered compilation):
     *   C1 (Client) — fast compilation, basic optimizations
     *   C2 (Server) — slow compilation, aggressive optimizations
     *
     * Tiers:
     *   0: Interpreter
     *   1-3: C1 compiled (with varying profiling)
     *   4: C2 compiled (fully optimized)
     *
     * Key C2 optimizations:
     *   1. INLINING — Replace method call with method body
     *      (most important optimization, enables all others)
     *      Default threshold: 35 bytes for hot methods, 325 for all
     *
     *   2. ESCAPE ANALYSIS — Track if object "escapes" the method
     *      If not: allocate on stack (no GC!) or eliminate entirely
     *
     *   3. SCALAR REPLACEMENT — Break object into fields
     *      Point p = new Point(x, y);
     *      → just use 'x' and 'y' directly, no object
     *
     *   4. DEAD CODE ELIMINATION — Remove unreachable/unused code
     *      This is why naive benchmarks fail!
     *
     *   5. LOOP UNROLLING — Copy loop body N times to reduce branching
     *
     *   6. NULL CHECK ELIMINATION — If proven non-null, skip check
     *
     *   7. BOUNDS CHECK ELIMINATION — If proven in-range, skip check
     *
     *   8. LOCK ELISION/COARSENING — Remove unnecessary synchronization
     *      Lock elision: if lock never contended, remove it
     *      Lock coarsening: merge adjacent locks into one
     *
     * USEFUL JVM FLAGS:
     *   -XX:+PrintCompilation      — show what's being compiled
     *   -XX:+UnlockDiagnosticVMOptions -XX:+PrintInlining  — show inlining
     *   -XX:-DoEscapeAnalysis      — disable escape analysis (for testing)
     *   -XX:+PrintGC               — GC events
     */

    // ========================================================
    // DEMONSTRATIONS (showing concepts without JMH)
    // ========================================================

    // --- Helper: poor man's benchmark (aware of its limitations) ---
    @FunctionalInterface
    interface Benchmarkable {
        void run();
    }

    static void comparativeBenchmark(String name, int warmup, int iterations,
                                     Map<String, Benchmarkable> tests) {
        System.out.println("  Benchmark: " + name);
        System.out.println("  (Warmup: " + warmup + ", Iterations: " + iterations + ")");

        for (Map.Entry<String, Benchmarkable> entry : tests.entrySet()) {
            // Warmup
            for (int i = 0; i < warmup; i++) entry.getValue().run();

            // Measure
            long[] times = new long[iterations];
            for (int i = 0; i < iterations; i++) {
                long start = System.nanoTime();
                entry.getValue().run();
                times[i] = System.nanoTime() - start;
            }

            // Statistics
            long sum = 0;
            long min = Long.MAX_VALUE;
            long max = Long.MIN_VALUE;
            for (long t : times) {
                sum += t;
                min = Math.min(min, t);
                max = Math.max(max, t);
            }
            long avg = sum / iterations;
            System.out.printf("    %-30s avg=%,8dns  min=%,8dns  max=%,8dns%n",
                entry.getKey(), avg, min, max);
        }
        System.out.println();
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== CHAPTER 53: PERFORMANCE OPTIMIZATION ===\n");

        // ====================================================
        // 1. String Concatenation Performance
        // ====================================================
        System.out.println("--- 1. String Performance ---\n");

        final int N = 10_000;

        comparativeBenchmark("String building (" + N + " appends)", 3, 5,
            new LinkedHashMap<>() {{
                put("String += (DON'T DO THIS)", () -> {
                    String s = "";
                    for (int i = 0; i < N; i++) s += "a";
                });
                put("StringBuilder", () -> {
                    StringBuilder sb = new StringBuilder(N);
                    for (int i = 0; i < N; i++) sb.append("a");
                    sb.toString();
                });
                put("StringJoiner", () -> {
                    StringJoiner sj = new StringJoiner("");
                    for (int i = 0; i < N; i++) sj.add("a");
                    sj.toString();
                });
                put("char[] manual", () -> {
                    char[] chars = new char[N];
                    for (int i = 0; i < N; i++) chars[i] = 'a';
                    new String(chars);
                });
            }});

        // ====================================================
        // 2. Collection Performance
        // ====================================================
        System.out.println("--- 2. Collection Performance ---\n");

        final int SIZE = 100_000;
        List<Integer> arrayList = new ArrayList<>(SIZE);
        List<Integer> linkedList = new LinkedList<>();
        for (int i = 0; i < SIZE; i++) {
            arrayList.add(i);
            linkedList.add(i);
        }

        // Iteration
        comparativeBenchmark("Iteration (" + SIZE + " elements)", 3, 5,
            new LinkedHashMap<>() {{
                put("ArrayList (forEach)", () -> {
                    long sum = 0;
                    for (int x : arrayList) sum += x;
                });
                put("LinkedList (forEach)", () -> {
                    long sum = 0;
                    for (int x : linkedList) sum += x;
                });
                put("ArrayList (index)", () -> {
                    long sum = 0;
                    for (int i = 0; i < arrayList.size(); i++) sum += arrayList.get(i);
                });
            }});

        // ====================================================
        // 3. Boxing/Unboxing Cost
        // ====================================================
        System.out.println("--- 3. Boxing/Unboxing ---\n");

        comparativeBenchmark("Sum of " + SIZE + " numbers", 3, 5,
            new LinkedHashMap<>() {{
                put("int[] (primitive)", () -> {
                    int[] arr = new int[SIZE];
                    long sum = 0;
                    for (int i = 0; i < SIZE; i++) { arr[i] = i; sum += arr[i]; }
                });
                put("Integer[] (boxed)", () -> {
                    Integer[] arr = new Integer[SIZE];
                    long sum = 0;
                    for (int i = 0; i < SIZE; i++) { arr[i] = i; sum += arr[i]; }
                });
                put("ArrayList<Integer>", () -> {
                    List<Integer> list = new ArrayList<>(SIZE);
                    long sum = 0;
                    for (int i = 0; i < SIZE; i++) { list.add(i); }
                    for (int x : list) sum += x;
                });
            }});

        // ====================================================
        // 4. Stream vs Loop
        // ====================================================
        System.out.println("--- 4. Stream vs Loop ---\n");

        comparativeBenchmark("Filter+Map+Sum (" + SIZE + ")", 3, 5,
            new LinkedHashMap<>() {{
                put("for-loop", () -> {
                    long sum = 0;
                    for (int x : arrayList) {
                        if (x % 2 == 0) sum += x * 2L;
                    }
                });
                put("stream", () -> {
                    arrayList.stream()
                        .filter(x -> x % 2 == 0)
                        .mapToLong(x -> x * 2L)
                        .sum();
                });
                put("parallelStream", () -> {
                    arrayList.parallelStream()
                        .filter(x -> x % 2 == 0)
                        .mapToLong(x -> x * 2L)
                        .sum();
                });
            }});

        // ====================================================
        // 5. HashMap Performance
        // ====================================================
        System.out.println("--- 5. Map Performance ---\n");

        comparativeBenchmark("Map put+get (" + SIZE + " entries)", 3, 5,
            new LinkedHashMap<>() {{
                put("HashMap", () -> {
                    Map<Integer, Integer> m = new HashMap<>(SIZE);
                    for (int i = 0; i < SIZE; i++) m.put(i, i);
                    for (int i = 0; i < SIZE; i++) m.get(i);
                });
                put("TreeMap", () -> {
                    Map<Integer, Integer> m = new TreeMap<>();
                    for (int i = 0; i < SIZE; i++) m.put(i, i);
                    for (int i = 0; i < SIZE; i++) m.get(i);
                });
                put("LinkedHashMap", () -> {
                    Map<Integer, Integer> m = new LinkedHashMap<>(SIZE);
                    for (int i = 0; i < SIZE; i++) m.put(i, i);
                    for (int i = 0; i < SIZE; i++) m.get(i);
                });
            }});

        // ====================================================
        // 6. Escape Analysis Demo
        // ====================================================
        System.out.println("--- 6. Escape Analysis & Object Allocation ---\n");

        System.out.println("  ESCAPE ANALYSIS determines if an object 'escapes' a method.");
        System.out.println("  If it doesn't escape, the JVM can:");
        System.out.println("    1. Allocate on stack (no GC needed)");
        System.out.println("    2. Scalar replace (decompose into primitives)");
        System.out.println("    3. Eliminate the allocation entirely\n");

        System.out.println("  Example (object does NOT escape):");
        System.out.println("    int sumPoints() {");
        System.out.println("      Point p = new Point(3, 4);  // doesn't escape");
        System.out.println("      return p.x + p.y;           // JVM uses registers");
        System.out.println("    }");
        System.out.println("    → After EA: just returns 3 + 4 = 7, no object created!\n");

        System.out.println("  Example (object DOES escape):");
        System.out.println("    Point getPoint() {");
        System.out.println("      return new Point(3, 4);  // escapes via return");
        System.out.println("    }");
        System.out.println("    → Must allocate on heap\n");

        // ====================================================
        // 7. Cache-Friendly Code
        // ====================================================
        System.out.println("--- 7. Cache-Friendly Code ---\n");

        /*
         * CPU caches are organized in cache lines (typically 64 bytes).
         * Accessing memory sequentially is MUCH faster than random access
         * because of prefetching and cache line reuse.
         *
         * ARRAY OF STRUCTS vs STRUCT OF ARRAYS:
         *
         * Array of Structs (commonly used):
         *   Point[] points = { {x1,y1}, {x2,y2}, ... }
         *   If you only need x values, you still pull y into cache
         *
         * Struct of Arrays (cache-friendly for single-field access):
         *   int[] xs = {x1, x2, ...}
         *   int[] ys = {y1, y2, ...}
         *   Accessing all x values = sequential memory access = fast
         */

        final int MATRIX = 1000;
        int[][] matrix = new int[MATRIX][MATRIX];
        for (int[] row : matrix) Arrays.fill(row, 1);

        comparativeBenchmark("Matrix traversal " + MATRIX + "x" + MATRIX, 3, 5,
            new LinkedHashMap<>() {{
                put("Row-major (cache-friendly)", () -> {
                    long sum = 0;
                    for (int i = 0; i < MATRIX; i++)
                        for (int j = 0; j < MATRIX; j++)
                            sum += matrix[i][j];
                });
                put("Column-major (cache-hostile)", () -> {
                    long sum = 0;
                    for (int j = 0; j < MATRIX; j++)
                        for (int i = 0; i < MATRIX; i++)
                            sum += matrix[i][j];
                });
            }});

        // ====================================================
        // 8. Common Anti-Patterns
        // ====================================================
        System.out.println("--- 8. Performance Anti-Patterns ---\n");

        System.out.println("  ❌ String concatenation in loop (use StringBuilder)");
        System.out.println("  ❌ Using LinkedList (ArrayList is almost always better)");
        System.out.println("  ❌ Excessive autoboxing (use int[], not List<Integer> in hot paths)");
        System.out.println("  ❌ Collections.synchronizedXxx (use ConcurrentHashMap etc.)");
        System.out.println("  ❌ Calling size() in loop condition (may not be cached)");
        System.out.println("  ❌ Catching Exception instead of specific type");
        System.out.println("  ❌ Reflection in hot paths (use MethodHandle or code gen)");
        System.out.println("  ❌ Creating regex Pattern in loop (compile once, reuse)");
        System.out.println("  ❌ Excessive object creation in hot loops");
        System.out.println("  ❌ Using HashMap with bad hashCode() implementations");

        System.out.println("\n  ✅ Pre-size collections: new ArrayList<>(expectedSize)");
        System.out.println("  ✅ Use primitives where possible");
        System.out.println("  ✅ Use StringBuilder for string building");
        System.out.println("  ✅ Use Arrays.sort() (dual-pivot quicksort) over Collections.sort()");
        System.out.println("  ✅ Use EnumMap/EnumSet for enum keys");
        System.out.println("  ✅ Prefer array over ArrayList in tight loops");
        System.out.println("  ✅ Cache expensive computations");
        System.out.println("  ✅ Use lazy initialization for heavy resources");

        // ====================================================
        // 9. GC-Friendly Patterns
        // ====================================================
        System.out.println("\n--- 9. GC Tuning Awareness ---\n");

        System.out.println("  GC Algorithms (match to your workload):");
        System.out.println("  ┌─────────────────┬──────────────────────────────────┐");
        System.out.println("  │ Collector        │ Best For                         │");
        System.out.println("  ├─────────────────┼──────────────────────────────────┤");
        System.out.println("  │ G1GC (default)   │ General purpose, balanced        │");
        System.out.println("  │ ZGC              │ Ultra-low pause (<1ms), large    │");
        System.out.println("  │ Shenandoah       │ Low pause, concurrent compaction │");
        System.out.println("  │ ParallelGC       │ Maximum throughput, batch jobs   │");
        System.out.println("  │ Serial           │ Small heaps, single core         │");
        System.out.println("  │ Epsilon          │ No GC at all (experimental)      │");
        System.out.println("  └─────────────────┴──────────────────────────────────┘");

        System.out.println("\n  GC-friendly coding:");
        System.out.println("  • Reduce allocation rate (fewer objects = fewer GC pauses)");
        System.out.println("  • Short-lived objects are cheap (nursery GC is fast)");
        System.out.println("  • Long-lived objects are OK (promoted, rarely collected)");
        System.out.println("  • MEDIUM-lived objects are expensive (promoted then collected)");
        System.out.println("  • Avoid finalizers and weak references unless necessary");
        System.out.println("  • -Xms = -Xmx reduces heap resizing overhead");

        // ====================================================
        // 10. JVM Tuning Flags
        // ====================================================
        System.out.println("\n--- 10. Key JVM Flags ---\n");

        System.out.println("  Memory:");
        System.out.println("    -Xms512m              Initial heap");
        System.out.println("    -Xmx2g               Max heap");
        System.out.println("    -Xss256k             Thread stack size");
        System.out.println("    -XX:MaxMetaspaceSize  Metaspace limit");

        System.out.println("\n  GC:");
        System.out.println("    -XX:+UseG1GC          Use G1 (default Java 9+)");
        System.out.println("    -XX:+UseZGC           Use ZGC (Java 15+)");
        System.out.println("    -XX:+PrintGCDetails   GC logging");
        System.out.println("    -Xlog:gc              Unified logging (Java 9+)");

        System.out.println("\n  Diagnostics:");
        System.out.println("    -XX:+PrintCompilation          What's being JIT compiled");
        System.out.println("    -XX:+UnlockDiagnosticVMOptions Unlock debug flags");
        System.out.println("    -XX:+PrintInlining             Show inlining decisions");
        System.out.println("    -XX:+HeapDumpOnOutOfMemoryError Auto heap dump");
        System.out.println("    -XX:+FlightRecorder            Enable JFR");

        System.out.println("\n✓ Performance & JMH Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Set up a real JMH project and benchmark ArrayList vs LinkedList
 *    for add-at-beginning, add-at-end, random access, and iteration.
 * 2. Benchmark HashMap vs TreeMap vs ConcurrentHashMap for your workload.
 * 3. Profile an application with jvisualvm or async-profiler.
 *    Find the hottest methods and optimize them.
 * 4. Compare the bytecode (javap -c) of a simple loop vs stream
 *    pipeline and explain why they differ in performance.
 *
 * NEXT: Chapter 54 — Advanced I/O
 */
