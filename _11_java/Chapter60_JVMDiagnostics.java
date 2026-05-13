/*
 * ============================================================
 *  CHAPTER 60: JVM DIAGNOSTICS & TROUBLESHOOTING
 * ============================================================
 *  A god-level Java developer doesn't just write code — they
 *  diagnose production issues at the JVM level. Thread dumps,
 *  heap dumps, GC logs, flight recordings, profiling.
 *
 *  This chapter makes you the person other people call
 *  when the server is on fire.
 *
 *  TOPICS:
 *    1. jps — List Java Processes
 *    2. jstack — Thread Dumps
 *    3. jmap — Heap Analysis
 *    4. jcmd — Swiss Army Knife
 *    5. JFR (Java Flight Recorder)
 *    6. jstat — GC Statistics
 *    7. jinfo — JVM Configuration
 *    8. jconsole / VisualVM — GUI Tools
 *    9. Reading Thread Dumps
 *   10. Reading GC Logs
 *   11. Common Production Issues
 *   12. Programmatic Diagnostics
 * ============================================================
 */

import java.lang.management.*;
import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.locks.*;

public class Chapter60_JVMDiagnostics {

    // ========================================================
    // 1. JPS — List Java Processes
    // ========================================================
    /*
     * jps                   — list PIDs and main class names
     * jps -l                — full main class path
     * jps -v                — JVM arguments
     * jps -m                — arguments to main()
     *
     * EXAMPLE:
     *   $ jps -l
     *   12345 com.example.MyApp
     *   12346 org.apache.catalina.startup.Bootstrap
     *   12347 jdk.jcmd/sun.tools.jps.Jps
     */

    // ========================================================
    // 2. JSTACK — Thread Dumps
    // ========================================================
    /*
     * jstack <pid>          — print all thread stacks
     * jstack -l <pid>       — include lock info
     * jstack -F <pid>       — force (if process hangs)
     *
     * THREAD STATES:
     *   RUNNABLE       — executing or ready to run
     *   BLOCKED        — waiting for monitor lock (synchronized)
     *   WAITING        — waiting indefinitely (wait(), park())
     *   TIMED_WAITING  — waiting with timeout (sleep(), wait(ms))
     *   NEW            — created but not started
     *   TERMINATED     — finished execution
     *
     * READING A THREAD DUMP:
     *
     *   "main" #1 prio=5 os_prio=0 tid=0x00007f... nid=0x1a3 runnable
     *      java.lang.Thread.State: RUNNABLE
     *       at java.io.FileInputStream.readBytes(Native Method)
     *       at java.io.FileInputStream.read(FileInputStream.java:255)
     *       at com.example.MyClass.process(MyClass.java:42)
     *       at com.example.MyClass.main(MyClass.java:10)
     *
     *   "pool-1-thread-3" #15 prio=5 waiting on condition
     *      java.lang.Thread.State: WAITING (parking)
     *       at sun.misc.Unsafe.park(Native Method)
     *       - parking to wait for <0x000000076ab024f8>
     *       at java.util.concurrent.locks.LockSupport.park(LockSupport.java:175)
     *       at java.util.concurrent.FutureTask.awaitDone(FutureTask.java:429)
     *
     * DEADLOCK DETECTION:
     *   jstack automatically prints "Found one Java-level deadlock:"
     *   with the cycle of threads and locks.
     *
     * KILL -3 (alternative):
     *   kill -3 <pid>        — prints thread dump to stdout/stderr
     *   Works even when jstack can't connect
     */

    // ========================================================
    // 3. JMAP — Heap Analysis
    // ========================================================
    /*
     * jmap -histo <pid>                — histogram of objects (count + size)
     * jmap -histo:live <pid>           — only live objects (triggers GC)
     * jmap -dump:format=b,file=heap.hprof <pid>  — full heap dump
     * jmap -dump:live,format=b,file=heap.hprof <pid>  — live objects only
     *
     * HISTOGRAM OUTPUT:
     *   num  #instances  #bytes  class name
     *   1:   5000000     120000000 [B          ← byte arrays
     *   2:   2000000     80000000  java.lang.String
     *   3:   500000      24000000  java.util.HashMap$Node
     *
     * HEAP DUMP ANALYSIS:
     *   Use these tools to open .hprof files:
     *   - Eclipse MAT (Memory Analyzer Tool) — best
     *   - VisualVM
     *   - IntelliJ profiler
     *   - jhat (deprecated)
     *
     * KEY CONCEPTS IN MAT:
     *   Shallow size: memory used by the object itself
     *   Retained size: memory freed if this object is GC'd
     *                  (includes all objects only reachable through it)
     *   Dominator tree: shows what's keeping objects alive
     *   Leak suspects: automatic analysis
     */

    // ========================================================
    // 4. JCMD — Swiss Army Knife
    // ========================================================
    /*
     * jcmd <pid> help                     — list available commands
     * jcmd <pid> VM.system_properties     — system properties
     * jcmd <pid> VM.flags                 — JVM flags
     * jcmd <pid> VM.uptime               — how long running
     * jcmd <pid> VM.version               — JVM version
     * jcmd <pid> VM.info                  — comprehensive info
     *
     * jcmd <pid> Thread.print             — thread dump (like jstack)
     * jcmd <pid> GC.heap_info             — heap info
     * jcmd <pid> GC.class_histogram       — object histogram (like jmap -histo)
     * jcmd <pid> GC.heap_dump file.hprof  — heap dump (like jmap -dump)
     * jcmd <pid> GC.run                   — trigger GC
     *
     * jcmd <pid> JFR.start duration=60s filename=rec.jfr — start recording
     * jcmd <pid> JFR.dump filename=rec.jfr               — dump recording
     * jcmd <pid> JFR.stop                                 — stop recording
     *
     * jcmd <pid> Compiler.directives_print — JIT directives
     * jcmd <pid> VM.native_memory         — native memory tracking
     *   (requires -XX:NativeMemoryTracking=summary)
     *
     * jcmd is the PREFERRED tool — it replaces jstack, jmap, jinfo
     */

    // ========================================================
    // 5. JFR — Java Flight Recorder
    // ========================================================
    /*
     * JFR is a PRODUCTION-SAFE profiler built into the JVM.
     * Very low overhead (~1-2%). Always-on in production.
     *
     * WHAT IT RECORDS:
     *   - CPU profiling (method sampling)
     *   - Memory allocation/GC events
     *   - Thread events (park, sleep, contention)
     *   - I/O events (file, socket)
     *   - JIT compilation events
     *   - Class loading events
     *   - Exception events
     *   - Custom events (your own!)
     *
     * START RECORDING:
     *   # At JVM start:
     *   java -XX:StartFlightRecording=duration=60s,filename=rec.jfr MyApp
     *
     *   # On running JVM:
     *   jcmd <pid> JFR.start duration=60s filename=rec.jfr
     *
     *   # Programmatic:
     *   try (Recording recording = new Recording()) {
     *       recording.start();
     *       // ... your code ...
     *       recording.stop();
     *       recording.dump(Path.of("recording.jfr"));
     *   }
     *
     * ANALYZE:
     *   - JDK Mission Control (JMC) — official GUI tool
     *   - IntelliJ profiler — reads .jfr files
     *   - jfr command-line:
     *     jfr summary recording.jfr
     *     jfr print --events CPULoad recording.jfr
     *
     * CUSTOM EVENTS:
     *   @Label("My Event")
     *   class MyEvent extends jdk.jfr.Event {
     *       @Label("Message")
     *       String message;
     *   }
     *
     *   MyEvent event = new MyEvent();
     *   event.message = "Something happened";
     *   event.commit();  // recorded by JFR
     */

    // ========================================================
    // 6. JSTAT — GC Statistics
    // ========================================================
    /*
     * jstat -gcutil <pid> 1000     — GC stats every 1 second
     * jstat -gc <pid> 1000         — detailed GC stats
     * jstat -gccause <pid>         — last GC cause
     *
     * OUTPUT (-gcutil):
     *   S0     S1     E      O      M      CCS    YGC    YGCT   FGC   FGCT
     *   0.00   98.45  65.12  45.23  97.00  94.12  15     0.123  2     0.456
     *
     *   S0/S1  = Survivor space 0/1 usage %
     *   E      = Eden space usage %
     *   O      = Old generation usage %
     *   M      = Metaspace usage %
     *   CCS    = Compressed class space %
     *   YGC    = Young GC count
     *   YGCT   = Young GC total time (seconds)
     *   FGC    = Full GC count
     *   FGCT   = Full GC total time (seconds)
     *
     * RED FLAGS:
     *   - FGC keeps increasing → memory leak
     *   - FGCT is high → long pauses
     *   - O near 100% → out of heap
     *   - YGC very frequent → high allocation rate
     */

    // ========================================================
    // 7. GC LOG ANALYSIS
    // ========================================================
    /*
     * ENABLE GC LOGGING:
     *   # Java 11+:
     *   java -Xlog:gc*:file=gc.log:time,level,tags MyApp
     *
     *   # Detailed:
     *   java -Xlog:gc*=debug:file=gc.log:time,uptime,level,tags:filecount=5,filesize=10m
     *
     * SAMPLE LOG:
     *   [0.234s][info][gc] GC(0) Pause Young (Normal) (G1 Evacuation Pause)
     *       24M->8M(256M) 3.456ms
     *
     *   [5.678s][info][gc] GC(12) Pause Full (System.gc())
     *       128M->45M(256M) 234.567ms
     *
     * READING THE LOG:
     *   24M->8M(256M) = before_GC → after_GC (heap_size) pause_time
     *
     * ANALYSIS TOOLS:
     *   - GCViewer (open source)
     *   - GCEasy.io (online)
     *   - HP JMeter
     *
     * WHAT TO LOOK FOR:
     *   1. Pause times: are they acceptable for your SLA?
     *   2. Frequency: how often is GC running?
     *   3. Reclaimed: is GC actually freeing memory?
     *   4. Full GC: these are BAD — long pauses
     *   5. Allocation rate: MB/s of new objects
     *   6. Promotion rate: how much moves old → young?
     */

    // ========================================================
    // 8. COMMON PRODUCTION ISSUES
    // ========================================================
    /*
     * ISSUE 1: OutOfMemoryError: Java heap space
     *   CAUSE: Heap is full, GC can't free enough
     *   DEBUG:
     *     1. Take heap dump: jcmd <pid> GC.heap_dump heap.hprof
     *     2. Open in MAT → find biggest retained objects
     *     3. Look for collections that grow forever
     *     4. Check for missing cache eviction
     *   FIX: Fix the leak, or increase -Xmx
     *
     * ISSUE 2: OutOfMemoryError: Metaspace
     *   CAUSE: Too many classes loaded (classloader leak)
     *   DEBUG:
     *     1. jcmd <pid> GC.class_histogram | head -20
     *     2. Look for duplicate classes from different classloaders
     *     3. Common in hot-deploy scenarios
     *   FIX: Fix classloader leak, or increase -XX:MaxMetaspaceSize
     *
     * ISSUE 3: Thread starvation / deadlock
     *   CAUSE: Threads waiting forever for locks
     *   DEBUG:
     *     1. jstack <pid> → look for BLOCKED/WAITING threads
     *     2. jstack detects deadlocks automatically
     *     3. Check for nested synchronized blocks
     *   FIX: Use timeout-based locks, fix lock ordering
     *
     * ISSUE 4: High CPU usage
     *   CAUSE: Busy loops, GC thrashing, or inefficient code
     *   DEBUG:
     *     1. top -H -p <pid> — find CPU-heavy threads
     *     2. Convert thread ID to hex
     *     3. jstack <pid> | grep <hex_tid> — find what it's doing
     *     4. Or use JFR for CPU profiling
     *   FIX: Optimize hot code path
     *
     * ISSUE 5: Long GC pauses
     *   CAUSE: Large heap, bad GC choice, high allocation rate
     *   DEBUG:
     *     1. Enable GC logging
     *     2. Check pause times and frequency
     *     3. Look at object survivor rates
     *   FIX: Tune GC, switch to ZGC/Shenandoah, reduce allocation
     *
     * ISSUE 6: StackOverflowError
     *   CAUSE: Infinite recursion or very deep call stacks
     *   DEBUG: Stack trace shows the recursive method
     *   FIX: Add base case, convert to iteration, or increase -Xss
     */

    // ========================================================
    // MAIN — Programmatic Diagnostics
    // ========================================================

    public static void main(String[] args) throws Exception {

        System.out.println("=== CHAPTER 60: JVM DIAGNOSTICS ===\n");

        // ====================================================
        // 1. Runtime Information
        // ====================================================
        System.out.println("--- 1. Runtime Info ---\n");

        Runtime runtime = Runtime.getRuntime();
        System.out.println("  Available processors: " + runtime.availableProcessors());
        System.out.printf("  Max memory:     %,d MB%n", runtime.maxMemory() / (1024 * 1024));
        System.out.printf("  Total memory:   %,d MB%n", runtime.totalMemory() / (1024 * 1024));
        System.out.printf("  Free memory:    %,d MB%n", runtime.freeMemory() / (1024 * 1024));
        System.out.printf("  Used memory:    %,d MB%n",
            (runtime.totalMemory() - runtime.freeMemory()) / (1024 * 1024));

        // ====================================================
        // 2. MXBeans — Management Beans
        // ====================================================
        System.out.println("\n--- 2. Management Beans (MXBeans) ---\n");

        // OS info
        OperatingSystemMXBean osMXBean = ManagementFactory.getOperatingSystemMXBean();
        System.out.println("  OS: " + osMXBean.getName() + " " + osMXBean.getVersion());
        System.out.println("  Arch: " + osMXBean.getArch());
        System.out.println("  CPUs: " + osMXBean.getAvailableProcessors());
        System.out.println("  Load avg: " + osMXBean.getSystemLoadAverage());

        // Runtime info
        RuntimeMXBean runtimeMXBean = ManagementFactory.getRuntimeMXBean();
        System.out.println("  JVM: " + runtimeMXBean.getVmName() + " "
            + runtimeMXBean.getVmVersion());
        System.out.println("  Uptime: " + runtimeMXBean.getUptime() + "ms");
        System.out.println("  PID: " + runtimeMXBean.getName().split("@")[0]);
        System.out.println("  Classpath entries: " +
            runtimeMXBean.getClassPath().split(System.getProperty("path.separator")).length);

        // Class loading
        ClassLoadingMXBean classLoadingMXBean = ManagementFactory.getClassLoadingMXBean();
        System.out.println("  Loaded classes: " + classLoadingMXBean.getLoadedClassCount());
        System.out.println("  Total loaded: " + classLoadingMXBean.getTotalLoadedClassCount());
        System.out.println("  Unloaded: " + classLoadingMXBean.getUnloadedClassCount());

        // Compilation
        CompilationMXBean compMXBean = ManagementFactory.getCompilationMXBean();
        if (compMXBean != null) {
            System.out.println("  JIT compiler: " + compMXBean.getName());
            System.out.println("  Compilation time: " + compMXBean.getTotalCompilationTime() + "ms");
        }

        // ====================================================
        // 3. Memory MXBeans
        // ====================================================
        System.out.println("\n--- 3. Memory Details ---\n");

        MemoryMXBean memoryMXBean = ManagementFactory.getMemoryMXBean();
        MemoryUsage heap = memoryMXBean.getHeapMemoryUsage();
        MemoryUsage nonHeap = memoryMXBean.getNonHeapMemoryUsage();

        System.out.printf("  Heap: used=%,dK, committed=%,dK, max=%,dK%n",
            heap.getUsed()/1024, heap.getCommitted()/1024, heap.getMax()/1024);
        System.out.printf("  Non-heap: used=%,dK, committed=%,dK%n",
            nonHeap.getUsed()/1024, nonHeap.getCommitted()/1024);

        // Memory pools
        System.out.println("\n  Memory Pools:");
        for (MemoryPoolMXBean pool : ManagementFactory.getMemoryPoolMXBeans()) {
            MemoryUsage usage = pool.getUsage();
            System.out.printf("    %-30s %s  used=%,dK%n",
                pool.getName(), pool.getType(), usage.getUsed()/1024);
        }

        // ====================================================
        // 4. GC Information
        // ====================================================
        System.out.println("\n--- 4. GC Information ---\n");

        for (GarbageCollectorMXBean gcBean : ManagementFactory.getGarbageCollectorMXBeans()) {
            System.out.println("  Collector: " + gcBean.getName());
            System.out.println("    Collections: " + gcBean.getCollectionCount());
            System.out.println("    Total time: " + gcBean.getCollectionTime() + "ms");
            System.out.println("    Pools: " + Arrays.toString(gcBean.getMemoryPoolNames()));
        }

        // ====================================================
        // 5. Thread Information
        // ====================================================
        System.out.println("\n--- 5. Thread Information ---\n");

        ThreadMXBean threadMXBean = ManagementFactory.getThreadMXBean();
        System.out.println("  Thread count: " + threadMXBean.getThreadCount());
        System.out.println("  Peak threads: " + threadMXBean.getPeakThreadCount());
        System.out.println("  Daemon threads: " + threadMXBean.getDaemonThreadCount());
        System.out.println("  Total started: " + threadMXBean.getTotalStartedThreadCount());

        // Detect deadlocks!
        long[] deadlocked = threadMXBean.findDeadlockedThreads();
        System.out.println("  Deadlocked threads: " +
            (deadlocked == null ? "none" : deadlocked.length));

        // Thread states summary
        Map<Thread.State, Integer> stateCounts = new EnumMap<>(Thread.State.class);
        for (ThreadInfo ti : threadMXBean.getThreadInfo(threadMXBean.getAllThreadIds())) {
            if (ti != null) {
                stateCounts.merge(ti.getThreadState(), 1, Integer::sum);
            }
        }
        System.out.println("  Thread states: " + stateCounts);

        // ====================================================
        // 6. Create a Deadlock for Detection
        // ====================================================
        System.out.println("\n--- 6. Deadlock Detection Demo ---\n");

        Object lockA = new Object();
        Object lockB = new Object();

        Thread t1 = new Thread(() -> {
            synchronized (lockA) {
                try { Thread.sleep(50); } catch (InterruptedException e) {}
                synchronized (lockB) { }
            }
        }, "DeadlockThread-1");

        Thread t2 = new Thread(() -> {
            synchronized (lockB) {
                try { Thread.sleep(50); } catch (InterruptedException e) {}
                synchronized (lockA) { }
            }
        }, "DeadlockThread-2");

        t1.start();
        t2.start();
        Thread.sleep(200);

        // Detect the deadlock
        long[] deadlockedIds = threadMXBean.findDeadlockedThreads();
        if (deadlockedIds != null) {
            System.out.println("  ⚠ DEADLOCK DETECTED!");
            for (ThreadInfo ti : threadMXBean.getThreadInfo(deadlockedIds, true, true)) {
                System.out.println("    Thread: " + ti.getThreadName()
                    + " state=" + ti.getThreadState());
                System.out.println("      Waiting for: " + ti.getLockName());
                System.out.println("      Held by: " + ti.getLockOwnerName());
            }
        } else {
            System.out.println("  No deadlock detected (threads may have resolved)");
        }

        // Cleanup deadlocked threads (they'll be daemon by exit)
        t1.interrupt();
        t2.interrupt();

        // ====================================================
        // 7. Command Cheat Sheet
        // ====================================================
        System.out.println("\n--- 7. Diagnostic Command Cheat Sheet ---\n");

        System.out.println("  QUICK DIAGNOSIS:");
        System.out.println("    jps -lv                         — find your process");
        System.out.println("    jcmd <pid> Thread.print         — thread dump");
        System.out.println("    jcmd <pid> GC.heap_info         — heap summary");
        System.out.println("    jcmd <pid> GC.class_histogram   — object counts");
        System.out.println("    jcmd <pid> VM.flags             — JVM settings");

        System.out.println("\n  DEEP ANALYSIS:");
        System.out.println("    jcmd <pid> GC.heap_dump f.hprof — heap dump");
        System.out.println("    jcmd <pid> JFR.start duration=60s filename=r.jfr");
        System.out.println("    jstat -gcutil <pid> 1000        — GC stats/sec");

        System.out.println("\n  EMERGENCY:");
        System.out.println("    kill -3 <pid>                   — thread dump to stdout");
        System.out.println("    jstack -F <pid>                 — forced thread dump");
        System.out.println("    jmap -dump:live,file=h.hprof <pid> — heap dump");

        System.out.println("\n  STARTUP FLAGS:");
        System.out.println("    -XX:+HeapDumpOnOutOfMemoryError — auto dump on OOM");
        System.out.println("    -XX:HeapDumpPath=/path/         — dump location");
        System.out.println("    -Xlog:gc*:file=gc.log           — GC logging");
        System.out.println("    -XX:+FlightRecorder             — enable JFR");
        System.out.println("    -XX:NativeMemoryTracking=summary — track native mem");

        System.out.println("\n✓ JVM Diagnostics Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Run any Java app, take a thread dump with jstack, and identify
 *    how many threads are in each state.
 * 2. Create a memory leak (e.g., static list that grows forever),
 *    take a heap dump, and find the leak using MAT or VisualVM.
 * 3. Enable GC logging on a Java app and analyze:
 *    - Average pause time
 *    - GC frequency
 *    - Whether Full GCs occur
 * 4. Record a JFR session and find the hottest methods using JMC.
 *
 * NEXT: Chapter 61 — JNI & Native Interop
 */
