/*
 * ============================================================
 *  CHAPTER 19: GENERICS
 * ============================================================
 *
 *  Generics enable types (classes, interfaces, methods) to be
 *  parameterized. They provide compile-time type safety.
 *
 *  WITHOUT generics: List list = new ArrayList(); → can hold ANYTHING
 *  WITH generics:    List<String> list = new ArrayList<>(); → only Strings
 *
 *  NAMING CONVENTIONS:
 *  T = Type, E = Element, K = Key, V = Value, N = Number, ? = Wildcard
 *
 * ============================================================
 */

import java.util.*;

public class Chapter19_Generics {

    // =====================================================
    //  1. GENERIC CLASS
    // =====================================================

    // T is a type parameter — replaced with actual type when used
    static class Box<T> {
        private T content;

        public Box(T content) { this.content = content; }
        public T getContent() { return content; }
        public void setContent(T content) { this.content = content; }

        @Override
        public String toString() { return "Box[" + content + "]"; }
    }

    // Multiple type parameters
    static class Pair<K, V> {
        private K key;
        private V value;

        public Pair(K key, V value) {
            this.key = key;
            this.value = value;
        }

        public K getKey() { return key; }
        public V getValue() { return value; }

        @Override
        public String toString() { return key + " → " + value; }
    }

    // =====================================================
    //  2. GENERIC METHODS
    // =====================================================

    // Generic method — type parameter before return type
    static <T> void printArray(T[] array) {
        System.out.print("[");
        for (int i = 0; i < array.length; i++) {
            System.out.print(array[i]);
            if (i < array.length - 1) System.out.print(", ");
        }
        System.out.println("]");
    }

    // Generic method with return type
    static <T> T getFirst(List<T> list) {
        if (list == null || list.isEmpty()) return null;
        return list.get(0);
    }

    // Generic method with multiple type params
    static <K, V> Map<K, V> mapOf(K key, V value) {
        Map<K, V> map = new HashMap<>();
        map.put(key, value);
        return map;
    }

    // =====================================================
    //  3. BOUNDED TYPE PARAMETERS
    // =====================================================

    // Upper bound: T must be Number or its subclass
    static <T extends Number> double sum(List<T> list) {
        double total = 0;
        for (T num : list) {
            total += num.doubleValue();
        }
        return total;
    }

    // Multiple bounds: T must implement Comparable AND Serializable
    static <T extends Comparable<T>> T findMax(List<T> list) {
        if (list.isEmpty()) throw new IllegalArgumentException("Empty list");
        T max = list.get(0);
        for (T item : list) {
            if (item.compareTo(max) > 0) max = item;
        }
        return max;
    }

    // =====================================================
    //  4. WILDCARDS
    // =====================================================

    // Unbounded wildcard: ? — accepts any type
    static void printList(List<?> list) {
        for (Object item : list) {
            System.out.print(item + " ");
        }
        System.out.println();
    }

    // Upper bounded wildcard: ? extends Number — read-only producer
    // "PECS: Producer Extends, Consumer Super"
    static double sumOfList(List<? extends Number> list) {
        double total = 0;
        for (Number num : list) {
            total += num.doubleValue();
        }
        return total;
        // list.add(42); // ERROR! Can't add to ? extends (except null)
    }

    // Lower bounded wildcard: ? super Integer — write-only consumer
    static void addNumbers(List<? super Integer> list) {
        list.add(1);
        list.add(2);
        list.add(3);
        // Integer val = list.get(0); // ERROR! Get returns Object
    }

    // =====================================================
    //  5. GENERIC INTERFACE
    // =====================================================

    interface Repository<T> {
        void save(T entity);
        T findById(int id);
        List<T> findAll();
    }

    static class User {
        int id;
        String name;
        User(int id, String name) { this.id = id; this.name = name; }
        @Override
        public String toString() { return "User(" + id + "," + name + ")"; }
    }

    static class UserRepository implements Repository<User> {
        private Map<Integer, User> store = new HashMap<>();

        @Override
        public void save(User user) {
            store.put(user.id, user);
        }

        @Override
        public User findById(int id) {
            return store.get(id);
        }

        @Override
        public List<User> findAll() {
            return new ArrayList<>(store.values());
        }
    }

    // =====================================================
    //  6. GENERIC CLASS WITH BOUNDED TYPE
    // =====================================================

    // A sorted pair that requires Comparable elements
    static class SortedPair<T extends Comparable<T>> {
        T first, second;

        SortedPair(T a, T b) {
            if (a.compareTo(b) <= 0) {
                first = a;
                second = b;
            } else {
                first = b;
                second = a;
            }
        }

        @Override
        public String toString() {
            return "(" + first + ", " + second + ")";
        }
    }

    // =====================================================
    //  7. TYPE ERASURE
    // =====================================================

    /*
     * TYPE ERASURE — how generics work at runtime:
     *
     * Generics exist ONLY at compile time!
     * At runtime, all generic info is ERASED.
     *
     * Box<String> → becomes Box (raw) at runtime
     * T → becomes Object (or bound if <T extends Number>)
     *
     * Consequences:
     * 1. Cannot use: new T()      — type unknown at runtime
     * 2. Cannot use: new T[]      — can't create generic array
     * 3. Cannot use: instanceof T — type erased
     * 4. Cannot overload by generic type:
     *    void process(List<String>) and void process(List<Integer>)
     *    both become void process(List) after erasure → conflict!
     */

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Generic Class ---
        System.out.println("=== GENERIC CLASS ===\n");

        Box<String> stringBox = new Box<>("Hello");
        Box<Integer> intBox = new Box<>(42);
        Box<List<String>> listBox = new Box<>(Arrays.asList("A", "B"));

        System.out.println("String box: " + stringBox);
        System.out.println("Integer box: " + intBox);
        System.out.println("List box: " + listBox);

        // Type safety — compiler prevents wrong types
        // stringBox.setContent(42); // ERROR! Expected String

        // Pair
        Pair<String, Integer> nameAge = new Pair<>("Alice", 25);
        Pair<Integer, Boolean> numBool = new Pair<>(1, true);
        System.out.println("Pair: " + nameAge);
        System.out.println("Pair: " + numBool);

        // --- 2. Generic Methods ---
        System.out.println("\n=== GENERIC METHODS ===\n");

        Integer[] intArr = {1, 2, 3, 4, 5};
        String[] strArr = {"Hello", "World"};
        Double[] dblArr = {1.1, 2.2, 3.3};

        printArray(intArr);
        printArray(strArr);
        printArray(dblArr);

        List<String> names = Arrays.asList("Alice", "Bob", "Charlie");
        System.out.println("First: " + getFirst(names));

        Map<String, Integer> singleMap = mapOf("key", 42);
        System.out.println("MapOf: " + singleMap);

        // --- 3. Bounded Types ---
        System.out.println("\n=== BOUNDED TYPES ===\n");

        List<Integer> ints = Arrays.asList(1, 2, 3, 4, 5);
        List<Double> dbls = Arrays.asList(1.1, 2.2, 3.3);
        System.out.println("Sum ints: " + sum(ints));
        System.out.println("Sum doubles: " + sum(dbls));

        // sum(Arrays.asList("hello")); // ERROR! String is not a Number

        System.out.println("Max ints: " + findMax(ints));
        System.out.println("Max strings: " + findMax(names));

        // --- 4. Wildcards ---
        System.out.println("\n=== WILDCARDS ===\n");

        // Unbounded: accepts any List
        printList(ints);
        printList(names);
        printList(dbls);

        // Upper bounded: ? extends Number
        System.out.println("Sum of ints: " + sumOfList(ints));
        System.out.println("Sum of dbls: " + sumOfList(dbls));
        // sumOfList(names); // ERROR! String doesn't extend Number

        // Lower bounded: ? super Integer
        List<Number> numList = new ArrayList<>();
        addNumbers(numList); // adds Integer values
        System.out.println("After addNumbers: " + numList);

        // PECS: Producer Extends, Consumer Super
        System.out.println("\nPECS Rule:");
        System.out.println("  Use <? extends T> for reading (producing values)");
        System.out.println("  Use <? super T> for writing (consuming values)");
        System.out.println("  Use <T> for both reading and writing");

        // --- 5. Generic Interface ---
        System.out.println("\n=== GENERIC INTERFACE ===\n");

        UserRepository repo = new UserRepository();
        repo.save(new User(1, "Alice"));
        repo.save(new User(2, "Bob"));
        System.out.println("Find by id 1: " + repo.findById(1));
        System.out.println("Find all: " + repo.findAll());

        // --- 6. Sorted Pair ---
        System.out.println("\n=== SORTED PAIR ===\n");

        SortedPair<Integer> sp1 = new SortedPair<>(5, 3);
        SortedPair<String> sp2 = new SortedPair<>("Banana", "Apple");
        System.out.println("Sorted pair: " + sp1);
        System.out.println("Sorted pair: " + sp2);

        // --- 7. Type Erasure Demo ---
        System.out.println("\n=== TYPE ERASURE ===\n");

        Box<String> sBox = new Box<>("Test");
        Box<Integer> iBox = new Box<>(42);

        // At runtime, both are just Box
        System.out.println("sBox class: " + sBox.getClass().getName());
        System.out.println("iBox class: " + iBox.getClass().getName());
        System.out.println("Same class? " + (sBox.getClass() == iBox.getClass())); // true!

        // Cannot do: if (sBox instanceof Box<String>) — compile error!
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Create a generic Stack<T> class with push, pop, peek, isEmpty.
 *
 *  2. Create a generic method that swaps two elements in an array.
 *
 *  3. Create a generic Triplet<A, B, C> class.
 *
 *  4. Write a generic method that filters a list based on a predicate.
 *
 *  5. Explain why this doesn't compile:
 *     List<Object> list = new ArrayList<String>();
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 20 — Functional Programming & Lambdas
 * ============================================================
 */
