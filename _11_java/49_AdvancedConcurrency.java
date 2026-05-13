/*
 * ============================================================
 *  CHAPTER 49: ADVANCED CONCURRENCY II
 * ============================================================
 *  Beyond basics: the weapons that gods wield.
 *
 *  TOPICS:
 *    1. ForkJoinPool — work-stealing parallelism
 *    2. StampedLock — optimistic locking
 *    3. Phaser — flexible barrier
 *    4. LongAdder/LongAccumulator — high-contention counters
 *    5. CompletionService — process results as they arrive
 *    6. Lock-free algorithms (CAS patterns)
 *    7. ThreadLocal deep dive
 *    8. Exchanger, TransferQueue
 * ============================================================
 */

import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.*;
import java.util.concurrent.locks.*;

public class Chapter49_AdvancedConcurrency {

    // ========================================================
    // 1. FORK/JOIN FRAMEWORK
    // ========================================================
    // Divide-and-conquer parallelism with work-stealing.
    // Each thread has a deque. When a thread runs out of work,
    // it steals from the tail of another thread's deque.

    // RecursiveTask<V> — returns a value
    static class ParallelSum extends RecursiveTask<Long> {
        private final int[] arr;
        private final int start, end;
        private static final int THRESHOLD = 10_000;

        ParallelSum(int[] arr, int start, int end) {
            this.arr = arr; this.start = start; this.end = end;
        }

        @Override
        protected Long compute() {
            if (end - start <= THRESHOLD) {
                // Base case: compute directly
                long sum = 0;
                for (int i = start; i < end; i++) sum += arr[i];
                return sum;
            }

            // Recursive case: split
            int mid = start + (end - start) / 2;
            ParallelSum left = new ParallelSum(arr, start, mid);
            ParallelSum right = new ParallelSum(arr, mid, end);

            left.fork();           // submit left to pool (async)
            long rightResult = right.compute();  // compute right in THIS thread
            long leftResult = left.join();       // wait for left

            return leftResult + rightResult;
        }
    }

    // RecursiveAction — no return value
    static class ParallelSort extends RecursiveAction {
        private final int[] arr;
        private final int start, end;
        private static final int THRESHOLD = 10_000;

        ParallelSort(int[] arr, int start, int end) {
            this.arr = arr; this.start = start; this.end = end;
        }

        @Override
        protected void compute() {
            if (end - start <= THRESHOLD) {
                Arrays.sort(arr, start, end);
                return;
            }
            int mid = start + (end - start) / 2;
            invokeAll(
                new ParallelSort(arr, start, mid),
                new ParallelSort(arr, mid, end)
            );
            // Merge
            merge(arr, start, mid, end);
        }

        private void merge(int[] a, int lo, int mid, int hi) {
            int[] temp = Arrays.copyOfRange(a, lo, mid);
            int i = 0, j = mid, k = lo;
            while (i < temp.length && j < hi) {
                a[k++] = temp[i] <= a[j] ? temp[i++] : a[j++];
            }
            while (i < temp.length) a[k++] = temp[i++];
        }
    }

    // ========================================================
    // 2. STAMPED LOCK — optimistic reads
    // ========================================================
    // Faster than ReentrantReadWriteLock for read-heavy workloads.
    // Supports OPTIMISTIC reads (no locking, just validate).

    static class StampedPoint {
        private double x, y;
        private final StampedLock lock = new StampedLock();

        void move(double deltaX, double deltaY) {
            long stamp = lock.writeLock();
            try {
                x += deltaX;
                y += deltaY;
            } finally {
                lock.unlockWrite(stamp);
            }
        }

        double distanceFromOrigin() {
            // OPTIMISTIC READ — no lock acquired!
            long stamp = lock.tryOptimisticRead();
            double currentX = x, currentY = y;

            // Validate: did a write happen while we were reading?
            if (!lock.validate(stamp)) {
                // Fallback to pessimistic read lock
                stamp = lock.readLock();
                try {
                    currentX = x;
                    currentY = y;
                } finally {
                    lock.unlockRead(stamp);
                }
            }

            return Math.sqrt(currentX * currentX + currentY * currentY);
        }
    }

    // ========================================================
    // 3. PHASER — flexible barrier (reusable, dynamic)
    // ========================================================
    // Like CyclicBarrier but:
    //   → Parties can register/deregister dynamically
    //   → Multiple phases (not just one sync point)
    //   → Can be used as a CountDownLatch too

    static void phaserDemo() throws InterruptedException {
        Phaser phaser = new Phaser(1);  // "1" = self-registration

        for (int i = 1; i <= 3; i++) {
            int id = i;
            phaser.register();  // dynamic registration
            new Thread(() -> {
                System.out.println("    Worker " + id + " phase 0 (setup)");
                phaser.arriveAndAwaitAdvance();  // wait for all

                System.out.println("    Worker " + id + " phase 1 (process)");
                phaser.arriveAndAwaitAdvance();

                System.out.println("    Worker " + id + " phase 2 (cleanup)");
                phaser.arriveAndDeregister();    // done, leave
            }).start();
        }

        // Main thread advances through phases
        phaser.arriveAndAwaitAdvance();  // phase 0 → 1
        System.out.println("    [Main] Phase 0 complete");

        phaser.arriveAndAwaitAdvance();  // phase 1 → 2
        System.out.println("    [Main] Phase 1 complete");

        phaser.arriveAndDeregister();    // main done
        Thread.sleep(100);
    }

    // ========================================================
    // 4. LONGADDER & LONGACCUMULATOR
    // ========================================================
    // Way faster than AtomicLong under high contention.
    // Uses striped cells — each thread writes to its own cell,
    // cells are summed on read.

    static void longAdderDemo() throws InterruptedException {
        LongAdder adder = new LongAdder();
        AtomicLong atomic = new AtomicLong();

        int threads = 8;
        int iterations = 1_000_000;

        // Benchmark AtomicLong
        long t1 = System.nanoTime();
        Thread[] aThreads = new Thread[threads];
        for (int i = 0; i < threads; i++) {
            aThreads[i] = new Thread(() -> {
                for (int j = 0; j < iterations; j++) atomic.incrementAndGet();
            });
            aThreads[i].start();
        }
        for (Thread t : aThreads) t.join();
        long atomicTime = System.nanoTime() - t1;

        // Benchmark LongAdder
        long t2 = System.nanoTime();
        Thread[] lThreads = new Thread[threads];
        for (int i = 0; i < threads; i++) {
            lThreads[i] = new Thread(() -> {
                for (int j = 0; j < iterations; j++) adder.increment();
            });
            lThreads[i].start();
        }
        for (Thread t : lThreads) t.join();
        long adderTime = System.nanoTime() - t2;

        System.out.println("  AtomicLong: " + atomic.get() + " in " + atomicTime / 1_000_000 + "ms");
        System.out.println("  LongAdder:  " + adder.sum() + " in " + adderTime / 1_000_000 + "ms");
        System.out.println("  LongAdder is ~" + (atomicTime / Math.max(adderTime, 1)) + "x faster");

        // LongAccumulator — custom accumulation function
        LongAccumulator maxAcc = new LongAccumulator(Long::max, Long.MIN_VALUE);
        maxAcc.accumulate(5);
        maxAcc.accumulate(3);
        maxAcc.accumulate(9);
        maxAcc.accumulate(1);
        System.out.println("  LongAccumulator max: " + maxAcc.get());
    }

    // ========================================================
    // 5. COMPLETION SERVICE
    // ========================================================
    // Submit tasks, get results IN ORDER OF COMPLETION (not submission)

    static void completionServiceDemo() throws Exception {
        ExecutorService pool = Executors.newFixedThreadPool(3);
        CompletionService<String> cs = new ExecutorCompletionService<>(pool);

        // Submit tasks with varying durations
        cs.submit(() -> { Thread.sleep(300); return "Task A (300ms)"; });
        cs.submit(() -> { Thread.sleep(100); return "Task B (100ms)"; });
        cs.submit(() -> { Thread.sleep(200); return "Task C (200ms)"; });

        // Get results in completion order (B first, then C, then A)
        for (int i = 0; i < 3; i++) {
            Future<String> f = cs.take();  // blocks until next result
            System.out.println("    Completed: " + f.get());
        }
        pool.shutdown();
    }

    // ========================================================
    // 6. LOCK-FREE STACK (CAS pattern)
    // ========================================================
    // Compare-And-Swap: "change value only if it's still what I expect"

    static class LockFreeStack<T> {
        private final AtomicReference<Node<T>> top = new AtomicReference<>();

        private static class Node<T> {
            final T value;
            final Node<T> next;
            Node(T value, Node<T> next) { this.value = value; this.next = next; }
        }

        void push(T value) {
            Node<T> newNode;
            Node<T> oldTop;
            do {
                oldTop = top.get();
                newNode = new Node<>(value, oldTop);
            } while (!top.compareAndSet(oldTop, newNode));  // CAS retry loop
        }

        T pop() {
            Node<T> oldTop;
            Node<T> newTop;
            do {
                oldTop = top.get();
                if (oldTop == null) return null;
                newTop = oldTop.next;
            } while (!top.compareAndSet(oldTop, newTop));
            return oldTop.value;
        }
    }

    // ========================================================
    // 7. THREADLOCAL — per-thread storage
    // ========================================================
    // Each thread gets its own copy of the variable.
    // Common uses: user context, DB connections, formatters.

    static final ThreadLocal<String> userContext = new ThreadLocal<>();

    // With initial value
    static final ThreadLocal<List<String>> auditLog =
        ThreadLocal.withInitial(ArrayList::new);

    // ⚠️ CRITICAL: Always clean up ThreadLocal in thread pools!
    // Thread pool threads are reused — old values leak to new tasks.

    static void threadLocalDemo() throws InterruptedException {
        Thread t1 = new Thread(() -> {
            userContext.set("Alice");
            auditLog.get().add("Login");
            auditLog.get().add("View dashboard");
            System.out.println("    T1 user: " + userContext.get() + " log: " + auditLog.get());
            userContext.remove();  // ALWAYS clean up!
            auditLog.remove();
        });

        Thread t2 = new Thread(() -> {
            userContext.set("Bob");
            auditLog.get().add("Login");
            System.out.println("    T2 user: " + userContext.get() + " log: " + auditLog.get());
            userContext.remove();
            auditLog.remove();
        });

        t1.start(); t2.start();
        t1.join(); t2.join();
    }

    // ========================================================
    // 8. EXCHANGER — two threads swap data
    // ========================================================

    static void exchangerDemo() throws InterruptedException {
        Exchanger<String> exchanger = new Exchanger<>();

        Thread producer = new Thread(() -> {
            try {
                String data = "Produced Data";
                System.out.println("    Producer has: " + data);
                String received = exchanger.exchange(data);  // swap
                System.out.println("    Producer got: " + received);
            } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
        });

        Thread consumer = new Thread(() -> {
            try {
                String token = "Consumer Token";
                System.out.println("    Consumer has: " + token);
                String received = exchanger.exchange(token);  // swap
                System.out.println("    Consumer got: " + received);
            } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
        });

        producer.start(); consumer.start();
        producer.join(); consumer.join();
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) throws Exception {

        // --- 1. ForkJoin ---
        System.out.println("=== FORK/JOIN ===\n");
        int[] data = new int[1_000_000];
        Random rand = new Random(42);
        for (int i = 0; i < data.length; i++) data[i] = rand.nextInt(100);

        ForkJoinPool pool = new ForkJoinPool();
        long sum = pool.invoke(new ParallelSum(data, 0, data.length));
        long expectedSum = 0;
        for (int v : data) expectedSum += v;
        System.out.println("  Parallel sum: " + sum + " (expected: " + expectedSum + ")");
        System.out.println("  Parallelism:  " + pool.getParallelism());

        // Parallel sort
        int[] sortData = {9, 3, 7, 1, 5, 8, 2, 4, 6};
        pool.invoke(new ParallelSort(sortData, 0, sortData.length));
        System.out.println("  Parallel sort: " + Arrays.toString(sortData));

        // --- 2. StampedLock ---
        System.out.println("\n=== STAMPED LOCK ===\n");
        StampedPoint point = new StampedPoint();
        point.move(3, 4);
        System.out.println("  Distance: " + point.distanceFromOrigin() + " (expected: 5.0)");

        // --- 3. Phaser ---
        System.out.println("\n=== PHASER ===\n");
        phaserDemo();

        // --- 4. LongAdder ---
        System.out.println("\n=== LONGADDER vs ATOMICLONG ===\n");
        longAdderDemo();

        // --- 5. CompletionService ---
        System.out.println("\n=== COMPLETION SERVICE ===\n");
        completionServiceDemo();

        // --- 6. Lock-Free Stack ---
        System.out.println("\n=== LOCK-FREE STACK ===\n");
        LockFreeStack<Integer> lfStack = new LockFreeStack<>();
        lfStack.push(1); lfStack.push(2); lfStack.push(3);
        System.out.println("  Pop: " + lfStack.pop());
        System.out.println("  Pop: " + lfStack.pop());
        System.out.println("  Pop: " + lfStack.pop());

        // --- 7. ThreadLocal ---
        System.out.println("\n=== THREADLOCAL ===\n");
        threadLocalDemo();

        // --- 8. Exchanger ---
        System.out.println("\n=== EXCHANGER ===\n");
        exchangerDemo();

        // --- Summary ---
        System.out.println("\n=== TOOL SELECTION GUIDE ===");
        System.out.println("  ForkJoinPool:      CPU-bound divide-and-conquer");
        System.out.println("  StampedLock:       Read-heavy, optimistic reads");
        System.out.println("  Phaser:            Multi-phase barriers, dynamic parties");
        System.out.println("  LongAdder:         High-contention counters");
        System.out.println("  CompletionService: Process results as they arrive");
        System.out.println("  CAS (lock-free):   Ultra-low latency, simple structures");
        System.out.println("  ThreadLocal:       Per-thread state (clean up in pools!)");
        System.out.println("  Exchanger:         Two-thread data swap");

        System.out.println("\n✓ Advanced Concurrency II Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Implement parallel merge sort using ForkJoin and benchmark vs Arrays.sort.
 * 2. Build a read-heavy cache using StampedLock with optimistic reads.
 * 3. Implement a lock-free queue using AtomicReference (Michael-Scott queue).
 * 4. Create a pipeline: stage1 → stage2 → stage3 using Phaser to sync phases.
 *
 * NEXT: Chapter 50 — Functional Programming Deep Dive
 */
