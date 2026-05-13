/*
 * ============================================================
 *  CHAPTER 50: FUNCTIONAL PROGRAMMING DEEP DIVE
 * ============================================================
 *  Java is not a functional language, but you CAN think
 *  functionally. God-level Java means knowing WHEN and HOW.
 *
 *  TOPICS:
 *    1. Pure Functions & Referential Transparency
 *    2. Function Composition & Chaining
 *    3. Currying & Partial Application
 *    4. Monadic Patterns (Optional, Stream, CompletableFuture)
 *    5. Immutable Data Pipelines
 *    6. Algebraic Data Types (simulated)
 *    7. Memoization
 *    8. Functional Error Handling (Either/Result)
 *    9. Lazy Evaluation
 *   10. When NOT to use functional style
 * ============================================================
 */

import java.util.*;
import java.util.function.*;
import java.util.stream.*;

public class Chapter50_FunctionalDeep {

    // ========================================================
    // 1. PURE FUNCTIONS
    // ========================================================
    // Pure = same input → same output, no side effects
    // ✅ Pure: math, transformation, filtering
    // ❌ Impure: I/O, Random, System.currentTimeMillis(), mutation

    static int pureAdd(int a, int b) { return a + b; }  // ✅ always same
    // void log(String s) { System.out.println(s); }     // ❌ side effect

    // ========================================================
    // 2. FUNCTION COMPOSITION
    // ========================================================

    static <A, B, C> Function<A, C> compose(Function<A, B> f, Function<B, C> g) {
        return a -> g.apply(f.apply(a));
    }

    static <T> Function<T, T> pipeline(List<Function<T, T>> functions) {
        return functions.stream()
            .reduce(Function.identity(), Function::andThen);
    }

    // ========================================================
    // 3. CURRYING & PARTIAL APPLICATION
    // ========================================================
    // Currying: f(a, b, c) → f(a)(b)(c)
    // Each call returns a new function taking the next argument

    // Two-argument curry
    static <A, B, R> Function<A, Function<B, R>> curry(BiFunction<A, B, R> fn) {
        return a -> b -> fn.apply(a, b);
    }

    // Three-argument curry
    @FunctionalInterface
    interface TriFunction<A, B, C, R> {
        R apply(A a, B b, C c);
    }

    static <A, B, C, R> Function<A, Function<B, Function<C, R>>> curry3(TriFunction<A, B, C, R> fn) {
        return a -> b -> c -> fn.apply(a, b, c);
    }

    // Partial application: fix some arguments
    static <A, B, R> Function<B, R> partial(BiFunction<A, B, R> fn, A fixedA) {
        return b -> fn.apply(fixedA, b);
    }

    // ========================================================
    // 4. MEMOIZATION
    // ========================================================
    // Cache results of pure functions

    static <T, R> Function<T, R> memoize(Function<T, R> fn) {
        Map<T, R> cache = new ConcurrentHashMap<>();
        return input -> cache.computeIfAbsent(input, fn);
    }

    // Recursive memoization (using array trick for lambda self-reference)
    @SuppressWarnings("unchecked")
    static Function<Integer, Long> memoizedFib() {
        Function<Integer, Long>[] self = new Function[1];
        self[0] = memoize(n -> {
            if (n <= 1) return (long) n;
            return self[0].apply(n - 1) + self[0].apply(n - 2);
        });
        return self[0];
    }

    // ========================================================
    // 5. EITHER / RESULT — Functional Error Handling
    // ========================================================
    // Instead of exceptions, use a type that represents success OR failure.
    // Like Optional but with error information.

    static class Either<L, R> {
        private final L left;   // error
        private final R right;  // success
        private final boolean isRight;

        private Either(L left, R right, boolean isRight) {
            this.left = left; this.right = right; this.isRight = isRight;
        }

        static <L, R> Either<L, R> right(R value) {
            return new Either<>(null, value, true);
        }

        static <L, R> Either<L, R> left(L error) {
            return new Either<>(error, null, false);
        }

        boolean isRight() { return isRight; }
        boolean isLeft() { return !isRight; }
        R getRight() { return right; }
        L getLeft() { return left; }

        <T> Either<L, T> map(Function<R, T> fn) {
            return isRight ? Either.right(fn.apply(right)) : Either.left(left);
        }

        <T> Either<L, T> flatMap(Function<R, Either<L, T>> fn) {
            return isRight ? fn.apply(right) : Either.left(left);
        }

        R orElse(R defaultValue) {
            return isRight ? right : defaultValue;
        }

        @Override
        public String toString() {
            return isRight ? "Right(" + right + ")" : "Left(" + left + ")";
        }
    }

    // Usage: validation pipeline
    static Either<String, Integer> parseInt(String s) {
        try { return Either.right(Integer.parseInt(s)); }
        catch (NumberFormatException e) { return Either.left("Not a number: " + s); }
    }

    static Either<String, Integer> validatePositive(int n) {
        return n > 0 ? Either.right(n) : Either.left("Must be positive: " + n);
    }

    static Either<String, Integer> validateRange(int n) {
        return n <= 100 ? Either.right(n) : Either.left("Too large: " + n);
    }

    // ========================================================
    // 6. LAZY EVALUATION
    // ========================================================
    // Compute only when needed

    static class Lazy<T> {
        private final Supplier<T> supplier;
        private T value;
        private boolean computed = false;

        Lazy(Supplier<T> supplier) { this.supplier = supplier; }

        static <T> Lazy<T> of(Supplier<T> supplier) { return new Lazy<>(supplier); }

        T get() {
            if (!computed) {
                value = supplier.get();
                computed = true;
            }
            return value;
        }

        <R> Lazy<R> map(Function<T, R> fn) {
            return Lazy.of(() -> fn.apply(get()));
        }

        <R> Lazy<R> flatMap(Function<T, Lazy<R>> fn) {
            return Lazy.of(() -> fn.apply(get()).get());
        }
    }

    // ========================================================
    // 7. PATTERN: Predicate Combinators
    // ========================================================

    static <T> Predicate<T> not(Predicate<T> p) { return p.negate(); }

    static <T> Predicate<T> allOf(List<Predicate<T>> predicates) {
        return predicates.stream().reduce(x -> true, Predicate::and);
    }

    static <T> Predicate<T> anyOf(List<Predicate<T>> predicates) {
        return predicates.stream().reduce(x -> false, Predicate::or);
    }

    // ========================================================
    // 8. IMMUTABLE DATA TRANSFORMATIONS
    // ========================================================

    static class Person {
        final String name;
        final int age;
        final String email;

        Person(String name, int age, String email) {
            this.name = name; this.age = age; this.email = email;
        }

        // "Wither" pattern — return new object with one field changed
        Person withName(String name) { return new Person(name, this.age, this.email); }
        Person withAge(int age) { return new Person(this.name, age, this.email); }
        Person withEmail(String email) { return new Person(this.name, this.age, email); }

        @Override
        public String toString() { return name + "(" + age + ", " + email + ")"; }
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        // --- 1. Function Composition ---
        System.out.println("=== FUNCTION COMPOSITION ===\n");

        Function<String, String> trim = String::trim;
        Function<String, String> lower = String::toLowerCase;
        Function<String, String> exclaim = s -> s + "!";

        // Compose: trim → lower → exclaim
        Function<String, String> process = trim.andThen(lower).andThen(exclaim);
        System.out.println("  Composed: " + process.apply("  HELLO WORLD  "));

        // Pipeline of transformations
        Function<Integer, Integer> mathPipeline = pipeline(List.of(
            x -> x + 10,
            x -> x * 2,
            x -> x - 5
        ));
        System.out.println("  Pipeline(5): " + mathPipeline.apply(5) + " = (5+10)*2-5 = 25");

        // --- 2. Currying ---
        System.out.println("\n=== CURRYING ===\n");

        // Curry a two-arg function
        Function<Integer, Function<Integer, Integer>> curriedAdd = curry(Integer::sum);
        Function<Integer, Integer> add10 = curriedAdd.apply(10);
        System.out.println("  add10(5) = " + add10.apply(5));
        System.out.println("  add10(20) = " + add10.apply(20));

        // Three-arg curry
        var curriedFormat = curry3((String prefix, String name, Integer age) ->
            prefix + " " + name + ", age " + age);
        var mr = curriedFormat.apply("Mr.");
        var mrBob = mr.apply("Bob");
        System.out.println("  " + mrBob.apply(30));

        // Partial application
        Function<Integer, Integer> multiplyBy3 = partial((a, b) -> a * b, 3);
        System.out.println("  multiplyBy3(7) = " + multiplyBy3.apply(7));

        // --- 3. Memoization ---
        System.out.println("\n=== MEMOIZATION ===\n");

        Function<Integer, Long> fib = memoizedFib();
        long start = System.nanoTime();
        long result = fib.apply(40);
        long elapsed = System.nanoTime() - start;
        System.out.println("  fib(40) = " + result + " in " + elapsed / 1_000 + "µs (memoized)");

        // General memoization
        Function<String, Integer> expensiveLen = memoize(s -> {
            System.out.println("    Computing length of '" + s + "'...");
            return s.length();
        });
        expensiveLen.apply("hello");  // computes
        expensiveLen.apply("hello");  // cached
        expensiveLen.apply("world");  // computes

        // --- 4. Either (Functional Error Handling) ---
        System.out.println("\n=== EITHER (Result Type) ===\n");

        // Validation chain using flatMap
        Either<String, Integer> valid = parseInt("42")
            .flatMap(n -> validatePositive(n))
            .flatMap(n -> validateRange(n));
        System.out.println("  Valid '42': " + valid);

        Either<String, Integer> invalid1 = parseInt("abc");
        System.out.println("  Invalid 'abc': " + invalid1);

        Either<String, Integer> invalid2 = parseInt("-5")
            .flatMap(n -> validatePositive(n));
        System.out.println("  Invalid '-5': " + invalid2);

        // Map transformation
        Either<String, String> mapped = parseInt("42")
            .map(n -> n * 2)
            .map(n -> "Result: " + n);
        System.out.println("  Mapped: " + mapped);

        // --- 5. Lazy Evaluation ---
        System.out.println("\n=== LAZY EVALUATION ===\n");

        Lazy<String> lazyGreeting = Lazy.of(() -> {
            System.out.println("    (computing...)");
            return "Hello!";
        });
        System.out.println("  Created lazy (not computed yet)");
        System.out.println("  First access: " + lazyGreeting.get());
        System.out.println("  Second access: " + lazyGreeting.get() + " (cached)");

        // Lazy map chain
        Lazy<Integer> lazyResult = Lazy.of(() -> 5)
            .map(x -> x * 10)
            .map(x -> x + 7);
        System.out.println("  Lazy chain: " + lazyResult.get());

        // --- 6. Predicate Combinators ---
        System.out.println("\n=== PREDICATE COMBINATORS ===\n");

        Predicate<String> nonEmpty = s -> !s.isEmpty();
        Predicate<String> shortEnough = s -> s.length() <= 20;
        Predicate<String> noSpaces = s -> !s.contains(" ");

        Predicate<String> usernameValid = allOf(List.of(nonEmpty, shortEnough, noSpaces));
        System.out.println("  'alice': " + usernameValid.test("alice"));
        System.out.println("  '': " + usernameValid.test(""));
        System.out.println("  'has space': " + usernameValid.test("has space"));

        // --- 7. Immutable Transformations ---
        System.out.println("\n=== IMMUTABLE WITHER PATTERN ===\n");

        Person original = new Person("Alice", 30, "alice@test.com");
        Person updated = original.withAge(31).withEmail("alice@new.com");
        System.out.println("  Original: " + original);
        System.out.println("  Updated:  " + updated + " (original unchanged!)");

        // --- 8. Advanced Stream Patterns ---
        System.out.println("\n=== ADVANCED STREAM PATTERNS ===\n");

        // Unfold: generate sequence from seed
        List<Integer> powers = Stream.iterate(1, x -> x * 2)
            .limit(10)
            .collect(Collectors.toList());
        System.out.println("  Powers of 2: " + powers);

        // Zip (Java doesn't have built-in zip, here's one)
        List<String> names = List.of("Alice", "Bob", "Charlie");
        List<Integer> scores = List.of(95, 87, 92);
        List<String> zipped = IntStream.range(0, Math.min(names.size(), scores.size()))
            .mapToObj(i -> names.get(i) + ": " + scores.get(i))
            .collect(Collectors.toList());
        System.out.println("  Zipped: " + zipped);

        // Sliding window
        List<Integer> data = List.of(1, 2, 3, 4, 5, 6, 7);
        int windowSize = 3;
        List<List<Integer>> windows = IntStream.rangeClosed(0, data.size() - windowSize)
            .mapToObj(i -> data.subList(i, i + windowSize))
            .collect(Collectors.toList());
        System.out.println("  Windows(3): " + windows);

        // --- When NOT to use FP ---
        System.out.println("\n=== WHEN TO USE / NOT USE FP ===");
        System.out.println("  ✅ Data transformation pipelines");
        System.out.println("  ✅ Validation chains");
        System.out.println("  ✅ Configuration / builder patterns");
        System.out.println("  ✅ Event processing / filtering");
        System.out.println("  ✅ Parallel processing (pure functions)");
        System.out.println("  ❌ Mutable state requirements (use OOP)");
        System.out.println("  ❌ Performance-critical inner loops (object allocation)");
        System.out.println("  ❌ Complex business logic with many branches");
        System.out.println("  ❌ When it hurts readability (don't be clever)");

        System.out.println("\n✓ Functional Programming Deep Dive Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Implement a Monad interface with unit, map, flatMap. Make Either and Lazy
 *    implement it.
 * 2. Build a validation library: ValidationResult with accumulating errors
 *    (not short-circuit like Either).
 * 3. Implement a Lens<S,A> for get/set on nested immutable objects.
 * 4. Create a lazy Stream (like Haskell's infinite lists) with take, filter, map.
 *
 * NEXT: Chapter 51 — Bytecode & MethodHandles
 */
