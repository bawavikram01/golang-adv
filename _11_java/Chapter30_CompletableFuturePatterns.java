/*
 * ============================================================
 *  CHAPTER 30: COMPLETABLEFUTURE & CONCURRENCY PATTERNS
 * ============================================================
 *  CompletableFuture (Java 8+) = modern async programming
 *
 *  Why CompletableFuture over Future?
 *    Future.get() blocks. CompletableFuture lets you:
 *    → Chain transformations (thenApply, thenCompose)
 *    → Combine results (thenCombine, allOf, anyOf)
 *    → Handle errors (exceptionally, handle)
 *    → Run callbacks (thenAccept, thenRun)
 *    → All without blocking!
 * ============================================================
 */

import java.util.concurrent.*;
import java.util.List;
import java.util.stream.*;

public class Chapter30_CompletableFuturePatterns {

    // Simulated async service calls
    static CompletableFuture<String> fetchUser(int id) {
        return CompletableFuture.supplyAsync(() -> {
            sleep(300);
            return "User-" + id;
        });
    }

    static CompletableFuture<String> fetchOrder(String user) {
        return CompletableFuture.supplyAsync(() -> {
            sleep(200);
            return "Order for " + user;
        });
    }

    static CompletableFuture<Double> fetchPrice(String item) {
        return CompletableFuture.supplyAsync(() -> {
            sleep(100);
            return 29.99;
        });
    }

    static void sleep(int ms) {
        try { Thread.sleep(ms); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
    }

    public static void main(String[] args) throws Exception {

        // --- 1. Creating CompletableFuture ---
        System.out.println("=== CREATING ===\n");

        // supplyAsync — returns a value
        CompletableFuture<String> cf1 = CompletableFuture.supplyAsync(() -> "Hello");

        // runAsync — no return value
        CompletableFuture<Void> cf2 = CompletableFuture.runAsync(() -> System.out.println("  Running async"));

        // completedFuture — already completed
        CompletableFuture<String> cf3 = CompletableFuture.completedFuture("Already done");

        System.out.println("  cf1: " + cf1.get());
        System.out.println("  cf3: " + cf3.get());

        // --- 2. Transformations ---
        System.out.println("\n=== TRANSFORMATIONS ===\n");

        // thenApply — transform result (like map)
        CompletableFuture<String> upper = CompletableFuture
            .supplyAsync(() -> "hello world")
            .thenApply(String::toUpperCase);
        System.out.println("  thenApply: " + upper.get());

        // thenApply chain
        CompletableFuture<Integer> length = CompletableFuture
            .supplyAsync(() -> "hello")
            .thenApply(String::toUpperCase)
            .thenApply(String::length);
        System.out.println("  Chained: " + length.get());

        // --- 3. Callbacks ---
        System.out.println("\n=== CALLBACKS ===\n");

        // thenAccept — consume result (no return)
        CompletableFuture.supplyAsync(() -> "Consumed value")
            .thenAccept(v -> System.out.println("  thenAccept: " + v))
            .get();

        // thenRun — just run something after (no access to result)
        CompletableFuture.supplyAsync(() -> "done")
            .thenRun(() -> System.out.println("  thenRun: Task completed"))
            .get();

        // --- 4. Chaining Async Operations ---
        System.out.println("\n=== CHAINING (thenCompose) ===\n");

        // thenCompose — chain dependent async operations (like flatMap)
        // fetchUser → fetchOrder (sequential dependency)
        String orderResult = fetchUser(1)
            .thenCompose(user -> fetchOrder(user))
            .get();
        System.out.println("  Composed: " + orderResult);

        // --- 5. Combining Independent Results ---
        System.out.println("\n=== COMBINING (thenCombine) ===\n");

        // thenCombine — combine two independent futures
        CompletableFuture<String> greeting = CompletableFuture
            .supplyAsync(() -> { sleep(200); return "Hello"; })
            .thenCombine(
                CompletableFuture.supplyAsync(() -> { sleep(100); return "World"; }),
                (a, b) -> a + " " + b
            );
        System.out.println("  Combined: " + greeting.get());

        // --- 6. allOf and anyOf ---
        System.out.println("\n=== allOf / anyOf ===\n");

        // allOf — wait for ALL to complete
        CompletableFuture<String> f1 = CompletableFuture.supplyAsync(() -> { sleep(300); return "A"; });
        CompletableFuture<String> f2 = CompletableFuture.supplyAsync(() -> { sleep(200); return "B"; });
        CompletableFuture<String> f3 = CompletableFuture.supplyAsync(() -> { sleep(100); return "C"; });

        CompletableFuture<Void> all = CompletableFuture.allOf(f1, f2, f3);
        all.get();  // wait for all
        System.out.println("  All done: " + f1.get() + ", " + f2.get() + ", " + f3.get());

        // Collect all results into a list
        List<CompletableFuture<String>> futureList = List.of(f1, f2, f3);
        List<String> results = futureList.stream()
            .map(CompletableFuture::join)
            .collect(Collectors.toList());
        System.out.println("  All results: " + results);

        // anyOf — first one to complete wins
        CompletableFuture<Object> any = CompletableFuture.anyOf(
            CompletableFuture.supplyAsync(() -> { sleep(300); return "Slow"; }),
            CompletableFuture.supplyAsync(() -> { sleep(100); return "Fast"; }),
            CompletableFuture.supplyAsync(() -> { sleep(200); return "Medium"; })
        );
        System.out.println("  First: " + any.get());

        // --- 7. Error Handling ---
        System.out.println("\n=== ERROR HANDLING ===\n");

        // exceptionally — handle exception and provide fallback
        CompletableFuture<String> withError = CompletableFuture
            .supplyAsync(() -> {
                if (true) throw new RuntimeException("Oops!");
                return "never";
            })
            .exceptionally(ex -> "Fallback: " + ex.getMessage());
        System.out.println("  exceptionally: " + withError.get());

        // handle — access both result and exception
        CompletableFuture<String> handled = CompletableFuture
            .supplyAsync(() -> {
                if (true) throw new RuntimeException("Error!");
                return "ok";
            })
            .handle((res, ex) -> {
                if (ex != null) return "Handled: " + ex.getMessage();
                return res;
            });
        System.out.println("  handle: " + handled.get());

        // whenComplete — observe result/error without changing it
        CompletableFuture.supplyAsync(() -> "Success")
            .whenComplete((res, ex) -> {
                if (ex != null) System.out.println("  Error: " + ex);
                else System.out.println("  whenComplete: " + res);
            }).get();

        // --- 8. Async Variants ---
        System.out.println("\n=== ASYNC VARIANTS ===\n");
        // Every callback has an Async version that runs on a different thread:
        // thenApply     → thenApplyAsync
        // thenAccept    → thenAcceptAsync
        // thenCompose   → thenComposeAsync
        // thenCombine   → thenCombineAsync

        CompletableFuture.supplyAsync(() -> "Async")
            .thenApplyAsync(s -> s + " variant")
            .thenAcceptAsync(s -> System.out.println("  " + s + " on " + Thread.currentThread().getName()))
            .get();

        // With custom executor
        ExecutorService customPool = Executors.newFixedThreadPool(2);
        CompletableFuture.supplyAsync(() -> "Custom pool", customPool)
            .thenApplyAsync(s -> s + " task", customPool)
            .thenAccept(s -> System.out.println("  " + s))
            .get();
        customPool.shutdown();

        // --- 9. Real-World Pattern: Parallel API Calls ---
        System.out.println("\n=== REAL-WORLD: PARALLEL API CALLS ===\n");

        long start = System.currentTimeMillis();

        CompletableFuture<String> userFuture = fetchUser(42);
        CompletableFuture<Double> priceFuture = fetchPrice("Widget");

        // Both run in parallel, then combine
        String combined = userFuture
            .thenCombine(priceFuture, (user, price) -> user + " bought item for $" + price)
            .get();

        long elapsed = System.currentTimeMillis() - start;
        System.out.println("  " + combined);
        System.out.println("  Completed in " + elapsed + "ms (parallel, not 300+100=400ms)");

        // --- 10. Pattern: Timeout ---
        System.out.println("\n=== TIMEOUT PATTERN ===\n");

        // Using orTimeout (Java 9+) or manual approach for Java 8:
        // Java 9+: future.orTimeout(1, TimeUnit.SECONDS)
        // Java 9+: future.completeOnTimeout("default", 1, TimeUnit.SECONDS)

        CompletableFuture<String> slow = CompletableFuture.supplyAsync(() -> {
            sleep(500);
            return "Slow result";
        });

        CompletableFuture<String> timeout = CompletableFuture.supplyAsync(() -> {
            sleep(200);
            return "Timeout fallback";
        });

        // Race between task and timeout
        String raceResult = (String) CompletableFuture.anyOf(slow, timeout).get();
        System.out.println("  Race winner: " + raceResult);

        // --- Summary ---
        System.out.println("\n=== METHOD SUMMARY ===");
        System.out.println("  CREATE:    supplyAsync, runAsync, completedFuture");
        System.out.println("  TRANSFORM: thenApply (map), thenCompose (flatMap)");
        System.out.println("  CONSUME:   thenAccept, thenRun");
        System.out.println("  COMBINE:   thenCombine, allOf, anyOf");
        System.out.println("  ERRORS:    exceptionally, handle, whenComplete");
        System.out.println("  ASYNC:     every method has *Async variant");

        System.out.println("\n✓ CompletableFuture & Concurrency Patterns Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Simulate 3 API calls (user, orders, recommendations). Fetch all in
 *    parallel, then combine into a single response object.
 * 2. Implement retry logic: if a CompletableFuture fails, retry up to 3 times.
 * 3. Create a pipeline: read file → parse JSON → transform → write result.
 * 4. Implement a circuit breaker pattern using CompletableFuture.
 *
 * NEXT: Chapter 31 — Annotations
 */
