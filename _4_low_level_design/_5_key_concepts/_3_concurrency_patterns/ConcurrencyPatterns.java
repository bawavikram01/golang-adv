/*
 * =============================================================
 * KEY CONCEPT: CONCURRENCY PATTERNS IN JAVA
 * =============================================================
 *
 * WHY CONCURRENCY MATTERS FOR LLD:
 *   - Real systems serve multiple users simultaneously
 *   - Thread safety bugs are the hardest to debug
 *   - Interviewers LOVE asking "How would you make this thread-safe?"
 *
 * THIS FILE COVERS:
 *   1. Thread-Safe Singleton (all approaches)
 *   2. Producer-Consumer Pattern
 *   3. Read-Write Lock Pattern
 *   4. Thread Pool Pattern
 *   5. Common thread-safety pitfalls
 */

import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.*;
import java.util.concurrent.locks.*;

public class ConcurrencyPatterns {

    public static void main(String[] args) throws Exception {

        // ═══════════════════════════════════════════════════════
        // 1. THREAD-SAFE SINGLETON
        // ═══════════════════════════════════════════════════════
        System.out.println("=== 1. THREAD-SAFE SINGLETON ===\n");

        // Launch 5 threads to get singleton — all should get same instance
        ExecutorService exec = Executors.newFixedThreadPool(5);
        Set<Integer> hashCodes = ConcurrentHashMap.newKeySet();

        for (int i = 0; i < 5; i++) {
            exec.submit(() -> {
                DatabaseConnection conn = DatabaseConnection.getInstance();
                hashCodes.add(System.identityHashCode(conn));
            });
        }
        exec.shutdown();
        exec.awaitTermination(5, TimeUnit.SECONDS);

        System.out.println("Unique instances created: " + hashCodes.size()
                + " (should be 1)\n");

        // ═══════════════════════════════════════════════════════
        // 2. PRODUCER-CONSUMER
        // ═══════════════════════════════════════════════════════
        System.out.println("=== 2. PRODUCER-CONSUMER ===\n");

        MessageQueue queue = new MessageQueue(3);  // capacity 3

        Thread producer = new Thread(() -> {
            for (int i = 1; i <= 5; i++) {
                queue.produce("Message-" + i);
            }
        }, "Producer");

        Thread consumer = new Thread(() -> {
            for (int i = 1; i <= 5; i++) {
                queue.consume();
            }
        }, "Consumer");

        producer.start();
        consumer.start();
        producer.join();
        consumer.join();

        // ═══════════════════════════════════════════════════════
        // 3. READ-WRITE LOCK
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== 3. READ-WRITE LOCK ===\n");

        ThreadSafeCache cache = new ThreadSafeCache();

        // Multiple readers + 1 writer
        ExecutorService pool = Executors.newFixedThreadPool(4);

        pool.submit(() -> { cache.put("key1", "value1"); });
        pool.submit(() -> { cache.put("key2", "value2"); });
        Thread.sleep(100); // let writes complete
        pool.submit(() -> { System.out.println("  Reader1: key1=" + cache.get("key1")); });
        pool.submit(() -> { System.out.println("  Reader2: key2=" + cache.get("key2")); });

        pool.shutdown();
        pool.awaitTermination(5, TimeUnit.SECONDS);

        // ═══════════════════════════════════════════════════════
        // 4. THREAD POOL
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== 4. THREAD POOL ===\n");

        TaskExecutor taskExecutor = new TaskExecutor(3);  // 3 worker threads

        for (int i = 1; i <= 6; i++) {
            final int taskId = i;
            taskExecutor.submit(() -> {
                System.out.println("  Task-" + taskId + " running on "
                        + Thread.currentThread().getName());
                try { Thread.sleep(200); } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            });
        }

        Thread.sleep(1000);
        taskExecutor.shutdown();

        // ═══════════════════════════════════════════════════════
        // 5. ATOMIC OPERATIONS
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== 5. ATOMIC OPERATIONS ===\n");
        AtomicCounter counter = new AtomicCounter();

        ExecutorService pool2 = Executors.newFixedThreadPool(10);
        for (int i = 0; i < 1000; i++) {
            pool2.submit(counter::increment);
        }
        pool2.shutdown();
        pool2.awaitTermination(5, TimeUnit.SECONDS);

        System.out.println("Counter after 1000 increments from 10 threads: "
                + counter.get() + " (should be 1000)\n");
    }
}

// ═══════════════════════════════════════════════════════════════
// 1. THREAD-SAFE SINGLETON — Double-Checked Locking
// ═══════════════════════════════════════════════════════════════
/*
 * THREE APPROACHES:
 *
 * (A) Eager — simple but loads even if never used
 *     private static final INSTANCE = new MyClass();
 *
 * (B) Double-Checked Locking — lazy, thread-safe
 *     volatile + synchronized (shown below)
 *
 * (C) Enum — BEST approach, handles serialization
 *     enum Singleton { INSTANCE; }
 *
 * WHY volatile?
 *   Without volatile, Thread B might see a partially constructed object
 *   due to instruction reordering. volatile prevents this.
 */
class DatabaseConnection {
    private static volatile DatabaseConnection instance;

    private DatabaseConnection() {
        System.out.println("  DatabaseConnection created by " + Thread.currentThread().getName());
    }

    public static DatabaseConnection getInstance() {
        if (instance == null) {                  // First check (no lock)
            synchronized (DatabaseConnection.class) {
                if (instance == null) {          // Second check (with lock)
                    instance = new DatabaseConnection();
                }
            }
        }
        return instance;
    }
}

// ═══════════════════════════════════════════════════════════════
// 2. PRODUCER-CONSUMER with wait/notify
// ═══════════════════════════════════════════════════════════════
/*
 * Classic concurrency pattern:
 *   - Producer adds items to a bounded buffer
 *   - Consumer removes items from the buffer
 *   - Producer WAITS when buffer is full
 *   - Consumer WAITS when buffer is empty
 *
 * In real life: Message queues (Kafka, RabbitMQ),
 *               Thread pools (task queue), I/O buffers
 */
class MessageQueue {
    private final Queue<String> buffer;
    private final int capacity;

    public MessageQueue(int capacity) {
        this.buffer = new LinkedList<>();
        this.capacity = capacity;
    }

    public synchronized void produce(String message) {
        while (buffer.size() == capacity) {
            try { wait(); } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                return;
            }
        }
        buffer.add(message);
        System.out.println("  [" + Thread.currentThread().getName()
                + "] Produced: " + message + " | Buffer: " + buffer.size());
        notifyAll();
    }

    public synchronized String consume() {
        while (buffer.isEmpty()) {
            try { wait(); } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                return null;
            }
        }
        String message = buffer.poll();
        System.out.println("  [" + Thread.currentThread().getName()
                + "] Consumed: " + message + " | Buffer: " + buffer.size());
        notifyAll();
        return message;
    }
}

// ═══════════════════════════════════════════════════════════════
// 3. READ-WRITE LOCK
// ═══════════════════════════════════════════════════════════════
/*
 * PROBLEM: synchronized blocks readers too
 *   - Reads don't modify data → multiple reads should be concurrent
 *   - Only writes need exclusive access
 *
 * ReadWriteLock allows:
 *   - Multiple concurrent readers (readLock)
 *   - Exclusive writer access (writeLock)
 *   - Writers block readers AND other writers
 *
 * PERFECT FOR: Caches, config stores, any read-heavy data
 */
class ThreadSafeCache {
    private final Map<String, String> store = new HashMap<>();
    private final ReadWriteLock lock = new ReentrantReadWriteLock();

    public String get(String key) {
        lock.readLock().lock();
        try {
            return store.get(key);
        } finally {
            lock.readLock().unlock();
        }
    }

    public void put(String key, String value) {
        lock.writeLock().lock();
        try {
            store.put(key, value);
            System.out.println("  Writer: set " + key + "=" + value);
        } finally {
            lock.writeLock().unlock();
        }
    }
}

// ═══════════════════════════════════════════════════════════════
// 4. THREAD POOL (simplified)
// ═══════════════════════════════════════════════════════════════
/*
 * WHY thread pools?
 *   - Creating threads is expensive (OS-level)
 *   - Reuse a fixed set of threads
 *   - Control resource usage
 *
 * Java provides: Executors.newFixedThreadPool(),
 *   newCachedThreadPool(), newScheduledThreadPool()
 *
 * In interviews, know how to explain the internal working:
 *   - BlockingQueue holds pending tasks
 *   - Worker threads poll from queue
 *   - If all workers busy, new tasks wait in queue
 */
class TaskExecutor {
    private final BlockingQueue<Runnable> taskQueue;
    private final List<Thread> workers;
    private volatile boolean isRunning = true;

    public TaskExecutor(int numThreads) {
        this.taskQueue = new LinkedBlockingQueue<>();
        this.workers = new ArrayList<>();

        for (int i = 0; i < numThreads; i++) {
            Thread worker = new Thread(() -> {
                while (isRunning || !taskQueue.isEmpty()) {
                    try {
                        Runnable task = taskQueue.poll(100, TimeUnit.MILLISECONDS);
                        if (task != null) task.run();
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                        break;
                    }
                }
            }, "Worker-" + i);
            worker.start();
            workers.add(worker);
        }
    }

    public void submit(Runnable task) {
        if (isRunning) taskQueue.offer(task);
    }

    public void shutdown() {
        isRunning = false;
        for (Thread w : workers) {
            try { w.join(2000); } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }
        System.out.println("  ThreadPool shutdown complete.");
    }
}

// ═══════════════════════════════════════════════════════════════
// 5. ATOMIC COUNTER (Compare-And-Swap)
// ═══════════════════════════════════════════════════════════════
/*
 * AtomicInteger uses CAS (Compare-And-Swap) — a CPU instruction
 * that atomically updates a value if it matches the expected value.
 *
 * NO LOCKS needed! This is lock-free concurrency.
 *
 *   Thread A: read counter=5, CAS(5→6) → SUCCESS
 *   Thread B: read counter=5, CAS(5→6) → FAIL (now it's 6), retry
 *   Thread B: read counter=6, CAS(6→7) → SUCCESS
 */
class AtomicCounter {
    private final AtomicInteger count = new AtomicInteger(0);

    public void increment() { count.incrementAndGet(); }
    public int get() { return count.get(); }
}

/*
 * CONCURRENCY CHEAT SHEET:
 * ─────────────────────────────────────────────────────────────
 * | Problem              | Solution                          |
 * |─────────────────────|──────────────────────────────────---|
 * | Single instance      | volatile + DCL / enum singleton   |
 * | Bounded buffer       | wait/notify or BlockingQueue      |
 * | Read-heavy cache     | ReentrantReadWriteLock             |
 * | Simple counter       | AtomicInteger (lock-free)         |
 * | Task execution       | ThreadPoolExecutor / Executors    |
 * | One-time init        | CountDownLatch                    |
 * | Barrier sync         | CyclicBarrier                     |
 * | Semaphore            | Semaphore (limit concurrent access)|
 * | Deadlock prevention  | Lock ordering / tryLock           |
 * ─────────────────────────────────────────────────────────────
 *
 * COMMON PITFALLS:
 *   1. Race condition: check-then-act without synchronization
 *   2. Deadlock: A locks X then Y, B locks Y then X
 *   3. Starvation: low-priority thread never gets lock
 *   4. Visibility: changes not visible across threads (need volatile)
 *   5. Non-atomic ++ : i++ is NOT atomic (read-modify-write)
 *
 * COMPILE & RUN:
 *   javac ConcurrencyPatterns.java && java ConcurrencyPatterns
 */
