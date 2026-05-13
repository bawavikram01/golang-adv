/*
 * ============================================================
 *  CHAPTER 29: CONCURRENCY — java.util.concurrent
 * ============================================================
 *  The java.util.concurrent package provides high-level
 *  concurrency utilities that are safer and more efficient
 *  than raw Thread + synchronized.
 *
 *  KEY COMPONENTS:
 *    ExecutorService   — thread pool management
 *    Future/Callable   — tasks that return results
 *    Locks (ReentrantLock) — advanced locking
 *    Atomic classes    — lock-free thread safety
 *    Concurrent Collections — thread-safe collections
 *    CountDownLatch, CyclicBarrier, Semaphore — coordination
 *    CompletableFuture — async programming (Java 8+)
 * ============================================================
 */

import java.util.concurrent.*;
import java.util.concurrent.atomic.*;
import java.util.concurrent.locks.*;
import java.util.*;

public class Chapter29_Concurrency {

    // === ReentrantLock ===
    static class LockCounter {
        private int count = 0;
        private final ReentrantLock lock = new ReentrantLock();

        void increment() {
            lock.lock();
            try {
                count++;
            } finally {
                lock.unlock();  // ALWAYS unlock in finally!
            }
        }

        int getCount() { return count; }
    }

    // === ReadWriteLock ===
    // Multiple readers can read concurrently, but writers get exclusive access
    static class ReadWriteCache {
        private final Map<String, String> cache = new HashMap<>();
        private final ReadWriteLock rwLock = new ReentrantReadWriteLock();

        String get(String key) {
            rwLock.readLock().lock();
            try {
                return cache.get(key);
            } finally {
                rwLock.readLock().unlock();
            }
        }

        void put(String key, String value) {
            rwLock.writeLock().lock();
            try {
                cache.put(key, value);
            } finally {
                rwLock.writeLock().unlock();
            }
        }
    }

    public static void main(String[] args) throws Exception {

        // --- 1. ExecutorService ---
        System.out.println("=== EXECUTOR SERVICE ===\n");

        // Fixed thread pool: reuses fixed number of threads
        ExecutorService fixedPool = Executors.newFixedThreadPool(3);
        for (int i = 1; i <= 5; i++) {
            int taskId = i;
            fixedPool.submit(() -> {
                System.out.println("  Task " + taskId + " on " + Thread.currentThread().getName());
            });
        }
        fixedPool.shutdown();          // no new tasks accepted
        fixedPool.awaitTermination(5, TimeUnit.SECONDS);  // wait for completion

        // Other pool types:
        // Executors.newSingleThreadExecutor()   → 1 thread
        // Executors.newCachedThreadPool()        → grows as needed
        // Executors.newScheduledThreadPool(n)    → for delayed/periodic tasks

        // --- 2. Callable + Future ---
        System.out.println("\n=== CALLABLE + FUTURE ===\n");

        ExecutorService pool = Executors.newFixedThreadPool(3);

        // Callable returns a value (unlike Runnable)
        Callable<Integer> task = () -> {
            Thread.sleep(500);
            return 42;
        };

        Future<Integer> future = pool.submit(task);
        System.out.println("  Is done? " + future.isDone());
        Integer result = future.get();  // blocks until result is available
        System.out.println("  Result: " + result);
        System.out.println("  Is done? " + future.isDone());

        // Submit multiple callables
        List<Future<String>> futures = new ArrayList<>();
        for (int i = 1; i <= 5; i++) {
            int id = i;
            futures.add(pool.submit(() -> "Result from task " + id));
        }
        for (Future<String> f : futures) {
            System.out.println("  " + f.get());
        }

        // invokeAll — waits for ALL tasks
        List<Callable<String>> tasks = List.of(
            () -> { Thread.sleep(200); return "A done"; },
            () -> { Thread.sleep(100); return "B done"; },
            () -> { Thread.sleep(300); return "C done"; }
        );
        List<Future<String>> allResults = pool.invokeAll(tasks);
        for (Future<String> f : allResults) {
            System.out.println("  " + f.get());
        }

        // invokeAny — returns first completed result
        String fastest = pool.invokeAny(tasks);
        System.out.println("  Fastest: " + fastest);
        pool.shutdown();

        // --- 3. ScheduledExecutorService ---
        System.out.println("\n=== SCHEDULED EXECUTOR ===\n");

        ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);

        // Run after 1 second delay
        scheduler.schedule(() -> System.out.println("  Delayed task executed"), 1, TimeUnit.SECONDS);

        // Run periodically: initial delay 0, period 500ms
        ScheduledFuture<?> periodic = scheduler.scheduleAtFixedRate(
            () -> System.out.print("."),
            0, 200, TimeUnit.MILLISECONDS
        );
        Thread.sleep(1500);
        periodic.cancel(false);
        System.out.println("\n  Periodic task cancelled");
        scheduler.shutdown();

        // --- 4. Atomic Classes ---
        System.out.println("\n=== ATOMIC CLASSES ===\n");

        AtomicInteger atomicCount = new AtomicInteger(0);
        AtomicLong atomicLong = new AtomicLong(0);
        AtomicBoolean atomicBool = new AtomicBoolean(false);

        // Thread-safe without locks
        ExecutorService atomicPool = Executors.newFixedThreadPool(5);
        for (int i = 0; i < 10; i++) {
            atomicPool.submit(() -> {
                for (int j = 0; j < 1000; j++) {
                    atomicCount.incrementAndGet();
                }
            });
        }
        atomicPool.shutdown();
        atomicPool.awaitTermination(5, TimeUnit.SECONDS);
        System.out.println("  Atomic count: " + atomicCount.get() + " (expected: 10000)");

        // Atomic operations
        AtomicInteger a = new AtomicInteger(10);
        System.out.println("  getAndAdd(5): " + a.getAndAdd(5) + " → now " + a.get());
        System.out.println("  compareAndSet(15, 20): " + a.compareAndSet(15, 20) + " → now " + a.get());
        System.out.println("  updateAndGet(x -> x * 2): " + a.updateAndGet(x -> x * 2));

        // --- 5. ReentrantLock ---
        System.out.println("\n=== REENTRANT LOCK ===\n");

        LockCounter lockCounter = new LockCounter();
        ExecutorService lockPool = Executors.newFixedThreadPool(5);
        for (int i = 0; i < 10; i++) {
            lockPool.submit(() -> {
                for (int j = 0; j < 1000; j++) lockCounter.increment();
            });
        }
        lockPool.shutdown();
        lockPool.awaitTermination(5, TimeUnit.SECONDS);
        System.out.println("  Lock counter: " + lockCounter.getCount() + " (expected: 10000)");

        // tryLock — non-blocking attempt
        ReentrantLock lock = new ReentrantLock();
        if (lock.tryLock()) {     // returns immediately
            try {
                System.out.println("  Got the lock!");
            } finally {
                lock.unlock();
            }
        }

        // tryLock with timeout
        if (lock.tryLock(1, TimeUnit.SECONDS)) {
            try {
                System.out.println("  Got lock within timeout");
            } finally {
                lock.unlock();
            }
        }

        // --- 6. CountDownLatch ---
        System.out.println("\n=== COUNTDOWN LATCH ===\n");
        // "Wait until N events happen"

        int workerCount = 3;
        CountDownLatch latch = new CountDownLatch(workerCount);
        ExecutorService latchPool = Executors.newFixedThreadPool(workerCount);

        for (int i = 1; i <= workerCount; i++) {
            int id = i;
            latchPool.submit(() -> {
                try {
                    Thread.sleep(id * 200);
                    System.out.println("  Worker " + id + " done");
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                } finally {
                    latch.countDown();  // decrement count
                }
            });
        }

        latch.await();  // blocks until count reaches 0
        System.out.println("  All workers finished!");
        latchPool.shutdown();

        // --- 7. CyclicBarrier ---
        System.out.println("\n=== CYCLIC BARRIER ===\n");
        // "Wait until N threads all arrive at a point"

        int parties = 3;
        CyclicBarrier barrier = new CyclicBarrier(parties, () -> {
            System.out.println("  >>> All threads reached barrier! <<<");
        });

        ExecutorService barrierPool = Executors.newFixedThreadPool(parties);
        for (int i = 1; i <= parties; i++) {
            int id = i;
            barrierPool.submit(() -> {
                try {
                    System.out.println("  Thread " + id + " working...");
                    Thread.sleep(id * 200);
                    System.out.println("  Thread " + id + " waiting at barrier");
                    barrier.await();  // wait for others
                    System.out.println("  Thread " + id + " proceeding");
                } catch (Exception e) { Thread.currentThread().interrupt(); }
            });
        }
        barrierPool.shutdown();
        barrierPool.awaitTermination(5, TimeUnit.SECONDS);

        // --- 8. Semaphore ---
        System.out.println("\n=== SEMAPHORE ===\n");
        // Controls access to a limited number of resources

        Semaphore semaphore = new Semaphore(2);  // max 2 concurrent

        ExecutorService semPool = Executors.newFixedThreadPool(5);
        for (int i = 1; i <= 5; i++) {
            int id = i;
            semPool.submit(() -> {
                try {
                    semaphore.acquire();
                    System.out.println("  Thread " + id + " acquired permit");
                    Thread.sleep(500);
                    System.out.println("  Thread " + id + " releasing permit");
                    semaphore.release();
                } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
            });
        }
        semPool.shutdown();
        semPool.awaitTermination(5, TimeUnit.SECONDS);

        // --- 9. Concurrent Collections ---
        System.out.println("\n=== CONCURRENT COLLECTIONS ===\n");

        // ConcurrentHashMap — thread-safe map without locking entire map
        ConcurrentHashMap<String, Integer> concMap = new ConcurrentHashMap<>();
        concMap.put("a", 1);
        concMap.putIfAbsent("b", 2);
        concMap.compute("a", (k, v) -> v + 10);
        System.out.println("  ConcurrentHashMap: " + concMap);

        // CopyOnWriteArrayList — safe for mostly-read scenarios
        CopyOnWriteArrayList<String> cowList = new CopyOnWriteArrayList<>();
        cowList.add("x");
        cowList.add("y");
        // Safe to iterate while modifying (iterator sees snapshot)
        for (String s : cowList) {
            cowList.add("z_" + s);  // no ConcurrentModificationException!
        }
        System.out.println("  CopyOnWriteArrayList: " + cowList);

        // BlockingQueue — thread-safe queue for producer-consumer
        BlockingQueue<String> bq = new LinkedBlockingQueue<>(5);
        bq.put("item1");     // blocks if full
        bq.offer("item2");   // returns false if full
        String item = bq.take();  // blocks if empty
        System.out.println("  BlockingQueue taken: " + item);

        System.out.println("\n  Collection comparison:");
        System.out.println("  HashMap           → ConcurrentHashMap");
        System.out.println("  ArrayList         → CopyOnWriteArrayList");
        System.out.println("  LinkedList        → ConcurrentLinkedQueue");
        System.out.println("  TreeMap           → ConcurrentSkipListMap");
        System.out.println("  PriorityQueue     → PriorityBlockingQueue");
        System.out.println("  ArrayDeque        → LinkedBlockingDeque");

        System.out.println("\n✓ Concurrency Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Create a thread pool of 4 threads. Submit 20 tasks that each compute
 *    a factorial. Collect results using Future.
 * 2. Implement a rate limiter using Semaphore (max N requests per second).
 * 3. Use ConcurrentHashMap to count word frequency from multiple threads.
 * 4. Use CountDownLatch to simulate a race: 5 runners wait for a start signal.
 *
 * NEXT: Chapter 30 — CompletableFuture & Concurrency Patterns
 */
