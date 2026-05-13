/*
 * ============================================================
 *  CHAPTER 28: MULTITHREADING BASICS
 * ============================================================
 *  Thread = lightweight unit of execution within a process
 *  Process = running program with its own memory space
 *  Thread shares the process's memory (heap) but has own stack
 *
 *  Why multithreading?
 *    → Utilize multi-core CPUs
 *    → Keep UI responsive
 *    → Handle concurrent I/O (servers, file ops)
 *
 *  Two ways to create threads:
 *    1. Extend Thread class
 *    2. Implement Runnable interface (PREFERRED)
 *
 *  Thread Lifecycle:
 *    NEW → RUNNABLE → RUNNING → (BLOCKED/WAITING/TIMED_WAITING) → TERMINATED
 * ============================================================
 */

public class Chapter28_MultithreadingBasics {

    // === WAY 1: Extend Thread ===
    static class MyThread extends Thread {
        private String name;

        MyThread(String name) {
            this.name = name;
        }

        @Override
        public void run() {
            for (int i = 1; i <= 5; i++) {
                System.out.println("  [Thread-" + name + "] Count: " + i);
                try { Thread.sleep(100); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
            }
        }
    }

    // === WAY 2: Implement Runnable (PREFERRED) ===
    static class MyRunnable implements Runnable {
        private String name;

        MyRunnable(String name) {
            this.name = name;
        }

        @Override
        public void run() {
            for (int i = 1; i <= 5; i++) {
                System.out.println("  [Runnable-" + name + "] Count: " + i);
                try { Thread.sleep(100); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
            }
        }
    }

    // === SYNCHRONIZATION ===
    // Without sync, concurrent access causes race conditions

    static class BankAccount {
        private int balance;

        BankAccount(int balance) {
            this.balance = balance;
        }

        // synchronized = only one thread at a time can execute this method
        synchronized void deposit(int amount) {
            int temp = balance;
            try { Thread.sleep(1); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
            balance = temp + amount;
        }

        synchronized void withdraw(int amount) {
            if (balance >= amount) {
                int temp = balance;
                try { Thread.sleep(1); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
                balance = temp - amount;
            }
        }

        synchronized int getBalance() {
            return balance;
        }
    }

    // === SYNCHRONIZED BLOCK (finer control) ===
    static class Counter {
        private int count = 0;
        private final Object lock = new Object();  // dedicated lock object

        void increment() {
            synchronized (lock) {   // sync only the critical section
                count++;
            }
        }

        int getCount() {
            synchronized (lock) {
                return count;
            }
        }
    }

    // === PRODUCER-CONSUMER with wait/notify ===
    static class SharedBuffer {
        private int data;
        private boolean hasData = false;

        synchronized void produce(int value) throws InterruptedException {
            while (hasData) {
                wait();  // release lock and wait until consumed
            }
            data = value;
            hasData = true;
            System.out.println("  Produced: " + value);
            notify();  // wake up consumer
        }

        synchronized int consume() throws InterruptedException {
            while (!hasData) {
                wait();  // release lock and wait until produced
            }
            hasData = false;
            notify();  // wake up producer
            System.out.println("  Consumed: " + data);
            return data;
        }
    }

    // === VOLATILE keyword ===
    // volatile = variable is read from main memory, not CPU cache
    static volatile boolean running = true;

    // === DAEMON THREADS ===
    // Daemon = background thread that doesn't prevent JVM from exiting
    // e.g., garbage collector
    static class DaemonExample extends Thread {
        @Override
        public void run() {
            while (true) {
                System.out.println("  [Daemon] Running in background...");
                try { Thread.sleep(500); } catch (InterruptedException e) { return; }
            }
        }
    }

    public static void main(String[] args) throws InterruptedException {

        // --- 1. Creating Threads ---
        System.out.println("=== CREATING THREADS ===\n");

        // Way 1: Extending Thread
        MyThread t1 = new MyThread("A");
        t1.start();  // start() creates new thread; run() would execute in same thread!
        t1.join();    // wait for t1 to finish

        // Way 2: Implementing Runnable
        Thread t2 = new Thread(new MyRunnable("B"));
        t2.start();
        t2.join();

        // Way 3: Lambda (since Runnable is functional interface)
        Thread t3 = new Thread(() -> {
            System.out.println("  [Lambda] Running in thread: " + Thread.currentThread().getName());
        });
        t3.start();
        t3.join();

        // --- 2. Thread Properties ---
        System.out.println("\n=== THREAD PROPERTIES ===\n");
        Thread current = Thread.currentThread();
        System.out.println("  Name: " + current.getName());
        System.out.println("  ID: " + current.getId());
        System.out.println("  Priority: " + current.getPriority() + " (1=MIN, 5=NORM, 10=MAX)");
        System.out.println("  State: " + current.getState());
        System.out.println("  Is Alive: " + current.isAlive());
        System.out.println("  Is Daemon: " + current.isDaemon());

        // --- 3. Thread Priority ---
        Thread highPriority = new Thread(() -> {
            System.out.println("  [HIGH] Priority: " + Thread.currentThread().getPriority());
        });
        highPriority.setPriority(Thread.MAX_PRIORITY);
        highPriority.start();
        highPriority.join();

        // --- 4. Synchronization Demo ---
        System.out.println("\n=== SYNCHRONIZATION ===\n");
        BankAccount account = new BankAccount(1000);

        Thread depositor = new Thread(() -> {
            for (int i = 0; i < 100; i++) account.deposit(10);
        });
        Thread withdrawer = new Thread(() -> {
            for (int i = 0; i < 100; i++) account.withdraw(5);
        });

        depositor.start();
        withdrawer.start();
        depositor.join();
        withdrawer.join();

        // With sync: always 1000 + (100*10) - (100*5) = 1500
        System.out.println("  Final balance: " + account.getBalance() + " (expected: 1500)");

        // --- 5. Counter with sync block ---
        System.out.println("\n=== SYNC BLOCK ===\n");
        Counter counter = new Counter();

        Thread[] threads = new Thread[10];
        for (int i = 0; i < 10; i++) {
            threads[i] = new Thread(() -> {
                for (int j = 0; j < 1000; j++) counter.increment();
            });
            threads[i].start();
        }
        for (Thread t : threads) t.join();
        System.out.println("  Counter: " + counter.getCount() + " (expected: 10000)");

        // --- 6. wait/notify (Producer-Consumer) ---
        System.out.println("\n=== PRODUCER-CONSUMER ===\n");
        SharedBuffer buffer = new SharedBuffer();

        Thread producer = new Thread(() -> {
            try {
                for (int i = 1; i <= 5; i++) buffer.produce(i);
            } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
        });

        Thread consumer = new Thread(() -> {
            try {
                for (int i = 1; i <= 5; i++) buffer.consume();
            } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
        });

        producer.start();
        consumer.start();
        producer.join();
        consumer.join();

        // --- 7. Volatile ---
        System.out.println("\n=== VOLATILE ===\n");
        Thread worker = new Thread(() -> {
            int count = 0;
            while (running) {
                count++;
            }
            System.out.println("  Worker stopped after " + count + " iterations");
        });
        worker.start();
        Thread.sleep(10);
        running = false;   // visible to worker because of volatile
        worker.join();

        // --- 8. Daemon Thread ---
        System.out.println("\n=== DAEMON THREAD ===\n");
        DaemonExample daemon = new DaemonExample();
        daemon.setDaemon(true);   // must set BEFORE start()
        daemon.start();
        System.out.println("  Daemon started (will die when main exits)");
        Thread.sleep(1500);  // let daemon run for a bit

        // --- 9. Thread States ---
        System.out.println("\n=== THREAD STATES ===");
        System.out.println("  NEW         - Thread created but not started");
        System.out.println("  RUNNABLE    - start() called, ready or running");
        System.out.println("  BLOCKED     - waiting for a monitor lock (sync)");
        System.out.println("  WAITING     - wait(), join(), park()");
        System.out.println("  TIMED_WAIT  - sleep(), wait(timeout), join(timeout)");
        System.out.println("  TERMINATED  - run() completed or exception thrown");

        // --- 10. Common Pitfalls ---
        System.out.println("\n=== PITFALLS ===");
        System.out.println("  1. Calling run() instead of start() → runs in SAME thread");
        System.out.println("  2. Not using volatile/sync → visibility issues");
        System.out.println("  3. Deadlock: Thread A locks X, waits Y; Thread B locks Y, waits X");
        System.out.println("  4. Catching InterruptedException without re-interrupting");
        System.out.println("  5. Forgetting join() → main thread doesn't wait for completion");

        System.out.println("\n✓ Multithreading Basics Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Create 5 threads that print numbers 1-10. Observe interleaving.
 * 2. Implement a thread-safe Stack with push/pop using synchronized.
 * 3. Create a deadlock scenario, then fix it with lock ordering.
 * 4. Implement a countdown latch manually using wait/notify.
 *
 * NEXT: Chapter 29 — Concurrency (java.util.concurrent)
 */
