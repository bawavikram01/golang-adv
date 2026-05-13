/*
 * ============================================================
 *  CHAPTER 47: ADVANCED GENERICS
 * ============================================================
 *  Chapter 19 covered basic generics. This goes DEEP.
 *
 *  TOPICS:
 *    1. Type Erasure — what really happens at runtime
 *    2. Type Tokens & Super Type Tokens
 *    3. F-Bounded Polymorphism (recursive generics)
 *    4. Wildcard Capture & Helper Methods
 *    5. Intersection Types
 *    6. Generic Singletons & Checked Casts
 *    7. Heterogeneous Type-Safe Containers
 *    8. Builder Pattern with Generics
 *    9. Common Pitfalls & Tricks
 * ============================================================
 */

import java.lang.reflect.*;
import java.util.*;
import java.util.function.*;

public class Chapter47_AdvancedGenerics {

    // ========================================================
    // 1. TYPE ERASURE — what the compiler REALLY does
    // ========================================================
    //
    // At compile time: List<String> and List<Integer> are different
    // At runtime: both become just List (raw type)
    //
    // The compiler:
    //   1. Replaces type parameters with bounds (or Object)
    //   2. Inserts casts where needed
    //   3. Generates bridge methods for polymorphism
    //
    // CONSEQUENCE: you CANNOT do these at runtime:
    //   new T()           → type unknown at runtime
    //   new T[]           → same
    //   instanceof List<String>  → erased
    //   T.class           → not a thing
    //   static T field    → T is per-instance

    static <T> void erasureDemo() {
        // This is what the compiler turns generics into:
        // List<String> list → List list (raw)
        // T val → Object val (or bound if bounded)

        List<String> strings = new ArrayList<>();
        List<Integer> ints = new ArrayList<>();

        // Both are just ArrayList at runtime!
        System.out.println("  String list class: " + strings.getClass());
        System.out.println("  Int list class:    " + ints.getClass());
        System.out.println("  Same class?        " + (strings.getClass() == ints.getClass()));
    }

    // ========================================================
    // 2. TYPE TOKENS — Store type info at runtime
    // ========================================================
    // Since generics are erased, how do we know the type at runtime?
    // Pass Class<T> as a "type token"

    // Type-safe heterogeneous container (Effective Java, Item 33)
    static class TypeSafeMap {
        private Map<Class<?>, Object> map = new HashMap<>();

        <T> void put(Class<T> type, T value) {
            map.put(type, type.cast(value));  // runtime type check
        }

        <T> T get(Class<T> type) {
            return type.cast(map.get(type));
        }
    }

    // Generic factory using type token
    static <T> T createInstance(Class<T> clazz) throws Exception {
        return clazz.getDeclaredConstructor().newInstance();
    }

    // ========================================================
    // 3. SUPER TYPE TOKEN — Capture generic type info
    // ========================================================
    // Class<T> can't represent List<String>.class (doesn't exist!)
    // Neal Gafter's Super Type Token trick:

    static abstract class TypeReference<T> {
        private final Type type;

        protected TypeReference() {
            // Anonymous subclass captures generic info in superclass
            Type superClass = getClass().getGenericSuperclass();
            this.type = ((ParameterizedType) superClass).getActualTypeArguments()[0];
        }

        public Type getType() { return type; }

        @Override
        public String toString() { return type.getTypeName(); }
    }

    // ========================================================
    // 4. F-BOUNDED POLYMORPHISM (Recursive Generics)
    // ========================================================
    // Pattern: class Foo<T extends Foo<T>>
    // Used for: fluent APIs, builders, Comparable, enums

    static abstract class BaseBuilder<T extends BaseBuilder<T>> {
        private String name;
        private int value;

        @SuppressWarnings("unchecked")
        T self() { return (T) this; }

        T name(String name) { this.name = name; return self(); }
        T value(int value) { this.value = value; return self(); }

        @Override
        public String toString() { return name + "=" + value; }
    }

    // Subclass inherits fluent methods that return the RIGHT type
    static class AdvancedBuilder extends BaseBuilder<AdvancedBuilder> {
        private String extra;

        AdvancedBuilder extra(String extra) { this.extra = extra; return this; }

        @Override
        public String toString() { return super.toString() + " extra=" + extra; }
    }

    // Enum uses this pattern: Enum<E extends Enum<E>>
    // Comparable uses this: Comparable<T> where T is self

    // Comparable example
    static class SortableItem<T extends SortableItem<T>> implements Comparable<T> {
        private int priority;

        SortableItem(int priority) { this.priority = priority; }

        @Override
        public int compareTo(T other) {
            return Integer.compare(this.priority, other.priority);
        }
    }

    static class Task extends SortableItem<Task> {
        String name;
        Task(String name, int priority) { super(priority); this.name = name; }

        @Override
        public String toString() { return name + "(p=" + super.priority + ")"; }
    }

    // ========================================================
    // 5. WILDCARD CAPTURE
    // ========================================================
    // Sometimes you need a helper to "capture" the wildcard

    // This WON'T compile:
    // static void swap(List<?> list, int i, int j) {
    //     list.set(i, list.get(j));  // ERROR: can't add Object to List<?>
    // }

    // Fix: capture helper
    static void swap(List<?> list, int i, int j) {
        swapHelper(list, i, j);
    }

    private static <T> void swapHelper(List<T> list, int i, int j) {
        T temp = list.get(i);
        list.set(i, list.get(j));
        list.set(j, temp);
    }

    // ========================================================
    // 6. INTERSECTION TYPES
    // ========================================================
    // T extends A & B — T must implement BOTH

    interface Printable { String toPrintString(); }
    interface Loggable { String toLogString(); }

    static <T extends Printable & Loggable> void printAndLog(T item) {
        System.out.println("  Print: " + item.toPrintString());
        System.out.println("  Log:   " + item.toLogString());
    }

    static class Report implements Printable, Loggable {
        String title;
        Report(String title) { this.title = title; }
        @Override public String toPrintString() { return "[PRINT] " + title; }
        @Override public String toLogString() { return "[LOG] " + title; }
    }

    // Lambda intersection cast (useful for serialization):
    // Runnable r = (Runnable & Serializable) () -> doWork();

    // ========================================================
    // 7. GENERIC SINGLETON FACTORY
    // ========================================================
    // A single instance that works for any generic type (via erasure)

    @SuppressWarnings("unchecked")
    static <T> Comparator<T> reverseOrder() {
        // One object serves ALL types because generics are erased
        return (Comparator<T>) ReverseComparator.INSTANCE;
    }

    private enum ReverseComparator implements Comparator<Comparable<Object>> {
        INSTANCE;

        @Override
        @SuppressWarnings("unchecked")
        public int compare(Comparable<Object> a, Comparable<Object> b) {
            return b.compareTo(a);
        }
    }

    // ========================================================
    // 8. CHECKED CAST PATTERNS
    // ========================================================

    // Safe checked cast with Optional
    static <T> Optional<T> safeCast(Object obj, Class<T> type) {
        return type.isInstance(obj) ? Optional.of(type.cast(obj)) : Optional.empty();
    }

    // Checked collection (runtime type-safe)
    static <E> List<E> checkedList(List<E> list, Class<E> type) {
        return Collections.checkedList(list, type);
    }

    // ========================================================
    // 9. ADVANCED PECS — Producer Extends, Consumer Super
    // ========================================================

    static <T extends Comparable<? super T>> T max(List<? extends T> list) {
        // ? extends T → producer (reading FROM list)
        // Comparable<? super T> → T or its ancestor is Comparable
        // This is the MOST FLEXIBLE signature possible
        Iterator<? extends T> it = list.iterator();
        T result = it.next();
        while (it.hasNext()) {
            T item = it.next();
            if (item.compareTo(result) > 0) result = item;
        }
        return result;
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) throws Exception {

        // --- 1. Type Erasure ---
        System.out.println("=== TYPE ERASURE ===\n");
        erasureDemo();

        // --- 2. Type Tokens ---
        System.out.println("\n=== TYPE-SAFE HETEROGENEOUS CONTAINER ===\n");
        TypeSafeMap config = new TypeSafeMap();
        config.put(String.class, "localhost");
        config.put(Integer.class, 8080);
        config.put(Boolean.class, true);

        String host = config.get(String.class);
        Integer port = config.get(Integer.class);
        Boolean debug = config.get(Boolean.class);
        System.out.println("  Host: " + host + ", Port: " + port + ", Debug: " + debug);

        // Factory with type token
        StringBuilder sb = createInstance(StringBuilder.class);
        sb.append("Created via type token!");
        System.out.println("  Factory: " + sb);

        // --- 3. Super Type Token ---
        System.out.println("\n=== SUPER TYPE TOKEN ===\n");
        TypeReference<List<String>> listType = new TypeReference<List<String>>() {};
        TypeReference<Map<String, Integer>> mapType = new TypeReference<Map<String, Integer>>() {};
        System.out.println("  List type: " + listType);
        System.out.println("  Map type:  " + mapType);

        // --- 4. F-Bounded (Recursive Generics) ---
        System.out.println("\n=== F-BOUNDED POLYMORPHISM ===\n");

        // Fluent builder that returns CORRECT subclass type
        AdvancedBuilder built = new AdvancedBuilder()
            .name("test")       // returns AdvancedBuilder, not BaseBuilder
            .value(42)          // returns AdvancedBuilder, not BaseBuilder
            .extra("special");  // only on AdvancedBuilder
        System.out.println("  Built: " + built);

        // Comparable with F-bounded
        List<Task> tasks = new ArrayList<>(List.of(
            new Task("Low", 3),
            new Task("Critical", 1),
            new Task("Medium", 2)
        ));
        Collections.sort(tasks);
        System.out.println("  Sorted tasks: " + tasks);

        // --- 5. Wildcard Capture ---
        System.out.println("\n=== WILDCARD CAPTURE ===\n");
        List<String> letters = new ArrayList<>(List.of("A", "B", "C"));
        System.out.println("  Before swap: " + letters);
        swap(letters, 0, 2);
        System.out.println("  After swap:  " + letters);

        // --- 6. Intersection Types ---
        System.out.println("\n=== INTERSECTION TYPES ===\n");
        Report report = new Report("Q4 Sales");
        printAndLog(report);

        // --- 7. Safe Cast ---
        System.out.println("\n=== SAFE CAST ===\n");
        Object mystery = "I'm a String";
        safeCast(mystery, String.class).ifPresent(s ->
            System.out.println("  Safely cast: " + s.toUpperCase()));
        safeCast(mystery, Integer.class).ifPresentOrElse(
            i -> System.out.println("  Integer: " + i),
            () -> System.out.println("  Not an Integer"));

        // --- 8. Advanced PECS ---
        System.out.println("\n=== ADVANCED PECS ===\n");
        List<Integer> nums = List.of(3, 1, 4, 1, 5, 9, 2, 6);
        System.out.println("  Max: " + max(nums));

        // --- Pitfalls ---
        System.out.println("\n=== GENERICS PITFALLS ===");
        System.out.println("  1. Cannot create generic arrays: new T[] → ILLEGAL");
        System.out.println("  2. Cannot use primitives: List<int> → use List<Integer>");
        System.out.println("  3. Cannot do instanceof with generics: x instanceof List<String>");
        System.out.println("  4. Cannot create instances of type params: new T()");
        System.out.println("  5. Static fields cannot use class type params");
        System.out.println("  6. Cannot catch or throw generic exceptions");
        System.out.println("  7. Overloading with different generic types erases to same signature");
        System.out.println("     void foo(List<String>) and void foo(List<Integer>) → CLASH");

        System.out.println("\n✓ Advanced Generics Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Implement a type-safe event bus: register(EventType<T>, Consumer<T>).
 * 2. Create a generic Pair<A,B> with map() and flatMap() methods.
 * 3. Build a fluent builder hierarchy with 3 levels using F-bounded polymorphism.
 * 4. Explain why List<Dog> is NOT a subtype of List<Animal> (write demo).
 *
 * NEXT: Chapter 48 — Java Memory Model
 */
