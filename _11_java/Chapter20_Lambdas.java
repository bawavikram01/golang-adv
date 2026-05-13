/*
 * ============================================================
 *  CHAPTER 20: FUNCTIONAL PROGRAMMING & LAMBDAS
 * ============================================================
 *
 *  Java 8 introduced functional programming features:
 *  1. Lambda Expressions
 *  2. Functional Interfaces
 *  3. Method References
 *  4. Built-in functional interfaces (Predicate, Function, etc.)
 *
 *  LAMBDA SYNTAX:
 *  ──────────────
 *  (parameters) -> expression
 *  (parameters) -> { statements; }
 *
 *  Examples:
 *  () -> 42                          // no params, returns 42
 *  x -> x * 2                       // one param, expression
 *  (x, y) -> x + y                  // two params, expression
 *  (String s) -> s.toUpperCase()    // typed param
 *  (x, y) -> { return x + y; }     // block body with return
 *
 * ============================================================
 */

import java.util.*;
import java.util.function.*;

public class Chapter20_Lambdas {

    // =====================================================
    //  1. FUNCTIONAL INTERFACES
    // =====================================================

    @FunctionalInterface
    interface MathOperation {
        double apply(double a, double b);
    }

    @FunctionalInterface
    interface StringProcessor {
        String process(String input);
    }

    @FunctionalInterface
    interface Validator<T> {
        boolean validate(T value);
    }

    // =====================================================
    //  2. BUILT-IN FUNCTIONAL INTERFACES (java.util.function)
    // =====================================================

    /*
     *  ┌────────────────────┬───────────────┬────────────┬────────────┐
     *  │ Interface          │ Method        │ Input      │ Output     │
     *  ├────────────────────┼───────────────┼────────────┼────────────┤
     *  │ Predicate<T>       │ test(T)       │ T          │ boolean    │
     *  │ Function<T,R>      │ apply(T)      │ T          │ R          │
     *  │ Consumer<T>        │ accept(T)     │ T          │ void       │
     *  │ Supplier<T>        │ get()         │ none       │ T          │
     *  │ UnaryOperator<T>   │ apply(T)      │ T          │ T          │
     *  │ BinaryOperator<T>  │ apply(T,T)    │ T, T       │ T          │
     *  │ BiFunction<T,U,R>  │ apply(T,U)    │ T, U       │ R          │
     *  │ BiPredicate<T,U>   │ test(T,U)     │ T, U       │ boolean    │
     *  │ BiConsumer<T,U>    │ accept(T,U)   │ T, U       │ void       │
     *  └────────────────────┴───────────────┴────────────┴────────────┘
     */

    // =====================================================
    //  3. HELPER METHODS
    // =====================================================

    static double calculate(double a, double b, MathOperation op) {
        return op.apply(a, b);
    }

    static <T> List<T> filter(List<T> list, Predicate<T> predicate) {
        List<T> result = new ArrayList<>();
        for (T item : list) {
            if (predicate.test(item)) result.add(item);
        }
        return result;
    }

    static <T, R> List<R> map(List<T> list, Function<T, R> mapper) {
        List<R> result = new ArrayList<>();
        for (T item : list) {
            result.add(mapper.apply(item));
        }
        return result;
    }

    // For method reference demo
    static boolean isPositive(int n) { return n > 0; }

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Lambda Basics ---
        System.out.println("=== LAMBDA BASICS ===\n");

        // Before lambdas (anonymous inner class):
        MathOperation addOld = new MathOperation() {
            @Override
            public double apply(double a, double b) { return a + b; }
        };

        // With lambdas (concise!):
        MathOperation add = (a, b) -> a + b;
        MathOperation subtract = (a, b) -> a - b;
        MathOperation multiply = (a, b) -> a * b;
        MathOperation divide = (a, b) -> b != 0 ? a / b : 0;
        MathOperation power = (a, b) -> Math.pow(a, b);

        System.out.println("10 + 5 = " + calculate(10, 5, add));
        System.out.println("10 - 5 = " + calculate(10, 5, subtract));
        System.out.println("10 * 5 = " + calculate(10, 5, multiply));
        System.out.println("10 / 5 = " + calculate(10, 5, divide));
        System.out.println("2 ^ 10 = " + calculate(2, 10, power));

        // Lambda with block body
        MathOperation modWithLogging = (a, b) -> {
            System.out.print("  Computing " + a + " % " + b + " = ");
            double result = a % b;
            return result;
        };
        System.out.println(modWithLogging.apply(17, 5));

        // String processors
        StringProcessor toUpper = s -> s.toUpperCase();
        StringProcessor reverse = s -> new StringBuilder(s).reverse().toString();
        StringProcessor trim = String::trim; // method reference!

        System.out.println("\ntoUpper: " + toUpper.process("hello"));
        System.out.println("reverse: " + reverse.process("hello"));
        System.out.println("trim: '" + trim.process("  hello  ") + "'");

        // --- 2. Predicate ---
        System.out.println("\n=== PREDICATE ===\n");

        Predicate<Integer> isEven = n -> n % 2 == 0;
        Predicate<Integer> isPos = n -> n > 0;
        Predicate<String> isEmpty = String::isEmpty;

        System.out.println("4 is even? " + isEven.test(4));
        System.out.println("5 is even? " + isEven.test(5));

        // Compose predicates
        Predicate<Integer> isEvenAndPositive = isEven.and(isPos);
        Predicate<Integer> isEvenOrPositive = isEven.or(isPos);
        Predicate<Integer> isOdd = isEven.negate();

        System.out.println("-4 even AND positive? " + isEvenAndPositive.test(-4)); // false
        System.out.println("-4 even OR positive? " + isEvenOrPositive.test(-4));   // true
        System.out.println("5 is odd? " + isOdd.test(5)); // true

        // Filter with predicate
        List<Integer> numbers = Arrays.asList(-3, -2, -1, 0, 1, 2, 3, 4, 5);
        System.out.println("All: " + numbers);
        System.out.println("Even: " + filter(numbers, isEven));
        System.out.println("Positive: " + filter(numbers, isPos));
        System.out.println("Even & Positive: " + filter(numbers, isEvenAndPositive));

        // --- 3. Function ---
        System.out.println("\n=== FUNCTION ===\n");

        Function<String, Integer> strLength = String::length;
        Function<Integer, Integer> doubleIt = n -> n * 2;
        Function<Integer, String> intToStr = n -> "Number: " + n;

        System.out.println("Length of 'Hello': " + strLength.apply("Hello"));
        System.out.println("Double 5: " + doubleIt.apply(5));

        // Compose functions
        Function<Integer, Integer> doubleThenAdd10 = doubleIt.andThen(n -> n + 10);
        Function<Integer, Integer> add10ThenDouble = doubleIt.compose(n -> n + 10);

        System.out.println("double(5) then +10 = " + doubleThenAdd10.apply(5));  // 20
        System.out.println("+10(5) then double = " + add10ThenDouble.apply(5));  // 30

        // Map with function
        List<String> names = Arrays.asList("Alice", "Bob", "Charlie");
        List<Integer> lengths = map(names, String::length);
        List<String> upperNames = map(names, String::toUpperCase);
        System.out.println("Names: " + names);
        System.out.println("Lengths: " + lengths);
        System.out.println("Upper: " + upperNames);

        // --- 4. Consumer ---
        System.out.println("\n=== CONSUMER ===\n");

        Consumer<String> printer = System.out::println;
        Consumer<String> shoutPrinter = s -> System.out.println(s.toUpperCase() + "!!!");

        printer.accept("Hello");
        shoutPrinter.accept("Hello");

        // Chain consumers
        Consumer<String> printThenShout = printer.andThen(shoutPrinter);
        printThenShout.accept("Test");

        // forEach with consumer
        names.forEach(name -> System.out.println("  Hi, " + name));

        // --- 5. Supplier ---
        System.out.println("\n=== SUPPLIER ===\n");

        Supplier<Double> randomSupplier = Math::random;
        Supplier<String> greetSupplier = () -> "Hello, World!";
        Supplier<List<String>> listSupplier = ArrayList::new;

        System.out.println("Random: " + randomSupplier.get());
        System.out.println("Greeting: " + greetSupplier.get());
        System.out.println("New list: " + listSupplier.get());

        // --- 6. BiFunction ---
        System.out.println("\n=== BIFUNCTION ===\n");

        BiFunction<String, Integer, String> repeat = (s, n) -> s.repeat(n);
        BiFunction<Integer, Integer, Integer> max = Math::max;

        System.out.println("repeat('Ha', 3) = " + repeat.apply("Ha", 3));
        System.out.println("max(10, 20) = " + max.apply(10, 20));

        // --- 7. Method References ---
        System.out.println("\n=== METHOD REFERENCES ===\n");

        // Four types of method references:

        // 1. Static method reference: ClassName::staticMethod
        Function<Double, Double> sqrt = Math::sqrt;
        System.out.println("sqrt(16) = " + sqrt.apply(16.0));

        // 2. Instance method of particular object: object::method
        String greeting = "Hello, World!";
        Supplier<Integer> lengthGetter = greeting::length;
        System.out.println("length = " + lengthGetter.get());

        // 3. Instance method of arbitrary object: ClassName::instanceMethod
        Function<String, String> upper = String::toUpperCase;
        System.out.println("upper = " + upper.apply("hello"));

        // 4. Constructor reference: ClassName::new
        Function<String, StringBuilder> sbCreator = StringBuilder::new;
        StringBuilder sb = sbCreator.apply("Hello");
        System.out.println("StringBuilder: " + sb);

        // Using method references in collections
        List<String> words = Arrays.asList("  hello  ", "  world  ", "  java  ");
        List<String> trimmed = map(words, String::trim);
        System.out.println("Trimmed: " + trimmed);

        // --- 8. Closures (Capturing Variables) ---
        System.out.println("\n=== CLOSURES ===\n");

        // Lambdas can capture effectively final variables from enclosing scope
        int multiplier = 3; // effectively final
        Function<Integer, Integer> tripler = n -> n * multiplier;
        System.out.println("Triple 5: " + tripler.apply(5));

        // multiplier = 4; // ERROR! Would make it non-effectively-final

        // --- 9. Practical Examples ---
        System.out.println("\n=== PRACTICAL EXAMPLES ===\n");

        // Sort with lambda
        List<String> sortNames = new ArrayList<>(Arrays.asList("Charlie", "Alice", "Bob"));
        sortNames.sort((a, b) -> a.compareTo(b));
        System.out.println("Sorted: " + sortNames);

        sortNames.sort(Comparator.comparing(String::length));
        System.out.println("By length: " + sortNames);

        // Custom validation
        Validator<String> emailValidator = s -> s != null && s.contains("@") && s.contains(".");
        Validator<Integer> ageValidator = a -> a >= 0 && a <= 150;

        System.out.println("Valid email: " + emailValidator.validate("user@test.com"));
        System.out.println("Valid email: " + emailValidator.validate("invalid"));
        System.out.println("Valid age 25: " + ageValidator.validate(25));
        System.out.println("Valid age -1: " + ageValidator.validate(-1));
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Create a Function pipeline that: trims a string,
 *     converts to lowercase, replaces spaces with hyphens.
 *     Use andThen to chain.
 *
 *  2. Write a generic filter+map method using Predicate + Function.
 *
 *  3. Create a Supplier that generates incrementing IDs.
 *
 *  4. Use BiFunction to create a simple calculator.
 *
 *  5. Implement a retry mechanism using Supplier:
 *     retry(Supplier<T>, int maxAttempts) → T
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 21 — Streams API
 * ============================================================
 */
