/*
 * ============================================================
 *  CHAPTER 46: MODERN JAVA — Features from Java 12 to 21+
 * ============================================================
 *  Java evolves every 6 months. This chapter covers the MAJOR
 *  language features added after Java 11 (your current version).
 *
 *  Even if you can't run these yet, you MUST know them —
 *  they're the present and future of the language.
 *
 *  UPGRADE TIP: Install Java 21 (LTS) to run everything here.
 *    sudo apt install openjdk-21-jdk
 *
 *  COVERED:
 *    Java 12: Switch Expressions (preview)
 *    Java 13: Text Blocks (preview)
 *    Java 14: Records (preview), Pattern Matching instanceof
 *    Java 15: Sealed Classes (preview), Text Blocks (final)
 *    Java 16: Records (final), Pattern Matching instanceof (final)
 *    Java 17: Sealed Classes (final) — LTS RELEASE
 *    Java 21: Virtual Threads, Pattern Matching for switch,
 *             Record Patterns, Sequenced Collections — LTS RELEASE
 * ============================================================
 */

import java.util.*;
// import java.util.concurrent.*;  // uncomment for virtual threads on Java 21

public class Chapter46_ModernJava {

    // ========================================================
    // 1. SWITCH EXPRESSIONS (Java 14+)
    // ========================================================
    // Old switch = statement (no return value, fall-through traps)
    // New switch = expression (returns value, no fall-through)

    /*
     * OLD WAY (Java 11):
     *   String result;
     *   switch (day) {
     *       case "MON": case "TUE": case "WED": case "THU": case "FRI":
     *           result = "Weekday";
     *           break;            // MUST remember break!
     *       case "SAT": case "SUN":
     *           result = "Weekend";
     *           break;
     *       default:
     *           result = "Unknown";
     *   }
     *
     * NEW WAY (Java 14+):
     *   String result = switch (day) {
     *       case "MON", "TUE", "WED", "THU", "FRI" -> "Weekday";
     *       case "SAT", "SUN" -> "Weekend";
     *       default -> "Unknown";
     *   };  // note the semicolon — it's an expression!
     *
     * → Arrow syntax (->): no fall-through, returns value
     * → Multiple labels: case "MON", "TUE" (comma-separated)
     * → yield keyword for multi-line blocks:
     *   case "MON" -> {
     *       // multiple statements
     *       yield "Weekday";  // yield = return from switch expression
     *   }
     *
     * → Exhaustiveness: compiler ensures ALL cases are covered
     *   (especially powerful with sealed classes/enums)
     */

    // ========================================================
    // 2. TEXT BLOCKS (Java 15+)
    // ========================================================
    // Multi-line string literals without escape nightmares

    /*
     * OLD WAY (Java 11):
     *   String json = "{\n" +
     *       "  \"name\": \"Alice\",\n" +
     *       "  \"age\": 30\n" +
     *       "}";
     *
     * NEW WAY (Java 15+):
     *   String json = """
     *       {
     *           "name": "Alice",
     *           "age": 30
     *       }
     *       """;
     *
     * Rules:
     * → Opening """ must be followed by newline
     * → Closing """ position determines indentation stripping
     * → Incidental whitespace (common indent) is removed
     * → \n, \t, %s still work inside
     * → Use .formatted() for string interpolation:
     *     """
     *     Hello %s, you are %d years old
     *     """.formatted(name, age);
     */

    // ========================================================
    // 3. RECORDS (Java 16+)
    // ========================================================
    // Immutable data carriers — eliminate boilerplate

    /*
     * OLD WAY (Java 11): 50+ lines for a simple data class
     *   public class Point {
     *       private final int x;
     *       private final int y;
     *       public Point(int x, int y) { this.x = x; this.y = y; }
     *       public int x() { return x; }
     *       public int y() { return y; }
     *       @Override public boolean equals(Object o) { ... }
     *       @Override public int hashCode() { ... }
     *       @Override public String toString() { ... }
     *   }
     *
     * NEW WAY (Java 16+): 1 line!
     *   record Point(int x, int y) {}
     *
     * A record automatically generates:
     *   → private final fields
     *   → canonical constructor
     *   → accessor methods: x(), y() (NOT getX()!)
     *   → equals(), hashCode(), toString()
     *
     * Records can have:
     *   → Custom constructors (compact constructor for validation)
     *   → Methods
     *   → Static fields/methods
     *   → Implement interfaces
     *
     * Records CANNOT:
     *   → Extend classes (implicitly extend java.lang.Record)
     *   → Be extended (implicitly final)
     *   → Have mutable instance fields
     *
     * EXAMPLE:
     *   record Person(String name, int age) {
     *       // Compact constructor — validation
     *       Person {
     *           if (age < 0) throw new IllegalArgumentException("Age < 0");
     *           name = name.trim();  // can modify before assignment
     *       }
     *
     *       // Additional method
     *       String greeting() { return "Hi, I'm " + name; }
     *   }
     *
     *   var p = new Person("Alice", 30);
     *   p.name()  // "Alice" — note: name() not getName()
     *   p.age()   // 30
     *
     * COMMON USE CASES:
     *   → DTOs (Data Transfer Objects)
     *   → Map keys (automatic equals/hashCode)
     *   → Return multiple values from methods
     *   → Pattern matching (Java 21)
     */

    // ========================================================
    // 4. SEALED CLASSES (Java 17+)
    // ========================================================
    // Control exactly which classes can extend yours

    /*
     * OLD WAY: Anyone can extend your class. You can't control it.
     *
     * NEW WAY (Java 17+):
     *   sealed interface Shape permits Circle, Rectangle, Triangle {}
     *
     *   record Circle(double radius) implements Shape {}
     *   record Rectangle(double w, double h) implements Shape {}
     *   final class Triangle implements Shape {
     *       double base, height;
     *       // ...
     *   }
     *
     * Subclasses MUST be:
     *   → final       — no further subclassing
     *   → sealed      — further restricted
     *   → non-sealed  — open for extension
     *
     * WHY?
     *   → Compiler knows ALL possible subtypes
     *   → Enables exhaustive pattern matching in switch
     *   → Models algebraic data types (sum types)
     *   → Better domain modeling
     */

    // ========================================================
    // 5. PATTERN MATCHING for instanceof (Java 16+)
    // ========================================================
    // Eliminate redundant casting

    /*
     * OLD WAY (Java 11):
     *   if (obj instanceof String) {
     *       String s = (String) obj;  // redundant cast!
     *       System.out.println(s.length());
     *   }
     *
     * NEW WAY (Java 16+):
     *   if (obj instanceof String s) {
     *       System.out.println(s.length());  // s already cast!
     *   }
     *
     * → Pattern variable 's' is scoped to where it's guaranteed non-null
     * → Works with && (guarded):
     *   if (obj instanceof String s && s.length() > 5) { ... }
     *
     * → Does NOT work with || (because s might not be bound):
     *   // COMPILE ERROR: if (obj instanceof String s || s.isEmpty())
     */

    // ========================================================
    // 6. PATTERN MATCHING for switch (Java 21+)
    // ========================================================
    // switch on types, not just values — GAME CHANGER

    /*
     *   String describe(Object obj) {
     *       return switch (obj) {
     *           case Integer i          -> "Integer: " + i;
     *           case String s           -> "String: " + s;
     *           case int[] arr          -> "Array of length " + arr.length;
     *           case null               -> "null!";
     *           case Point(int x, int y) -> "Point at " + x + "," + y; // record pattern!
     *           default                 -> "Unknown: " + obj.getClass();
     *       };
     *   }
     *
     * GUARDED PATTERNS (when clause):
     *   case Integer i when i > 0  -> "positive: " + i;
     *   case Integer i when i < 0  -> "negative: " + i;
     *   case Integer i             -> "zero";
     *
     * WITH SEALED CLASSES (no default needed!):
     *   sealed interface Shape permits Circle, Rectangle {}
     *   record Circle(double r) implements Shape {}
     *   record Rectangle(double w, double h) implements Shape {}
     *
     *   double area(Shape s) {
     *       return switch (s) {
     *           case Circle(var r)       -> Math.PI * r * r;
     *           case Rectangle(var w, var h) -> w * h;
     *           // NO default needed — compiler knows it's exhaustive!
     *       };
     *   }
     *
     * This is ALGEBRAIC DATA TYPES + PATTERN MATCHING — the holy grail
     * of type-safe programming that Haskell/Rust/Scala have had for years.
     */

    // ========================================================
    // 7. RECORD PATTERNS (Java 21+)
    // ========================================================
    // Destructure records in patterns

    /*
     *   record Point(int x, int y) {}
     *   record Line(Point start, Point end) {}
     *
     *   // Nested destructuring!
     *   if (obj instanceof Line(Point(var x1, var y1), Point(var x2, var y2))) {
     *       double length = Math.sqrt(Math.pow(x2-x1, 2) + Math.pow(y2-y1, 2));
     *   }
     *
     *   // In switch:
     *   switch (shape) {
     *       case Circle(var r) -> Math.PI * r * r;
     *       case Rectangle(var w, var h) -> w * h;
     *   }
     */

    // ========================================================
    // 8. VIRTUAL THREADS (Java 21+) — Project Loom
    // ========================================================
    // Lightweight threads — millions possible, not just thousands

    /*
     * OLD WAY: Platform threads (OS threads)
     *   → Heavy (~1MB stack each)
     *   → Limited (~thousands)
     *   → Thread pools needed
     *
     * NEW WAY: Virtual threads
     *   → Ultra-lightweight (~few KB)
     *   → Millions possible
     *   → Mounted/unmounted on carrier (platform) threads
     *   → Perfect for I/O-bound tasks
     *
     * CREATING VIRTUAL THREADS:
     *
     *   // Way 1: Direct
     *   Thread vt = Thread.ofVirtual().start(() -> {
     *       System.out.println("I'm virtual! " + Thread.currentThread());
     *   });
     *
     *   // Way 2: ExecutorService (recommended)
     *   try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
     *       for (int i = 0; i < 100_000; i++) {
     *           executor.submit(() -> {
     *               Thread.sleep(Duration.ofSeconds(1));
     *               return "done";
     *           });
     *       }
     *   }  // waits for all tasks
     *
     *   // Way 3: Factory
     *   ThreadFactory factory = Thread.ofVirtual().name("vt-", 0).factory();
     *   Thread t = factory.newThread(() -> doWork());
     *
     * KEY DIFFERENCE:
     *   Platform thread blocks OS thread during I/O → wastes resources
     *   Virtual thread blocks but unmounts → carrier thread does other work
     *
     * RULES:
     *   → Don't pool virtual threads (create-per-task instead)
     *   → Avoid synchronized blocks in virtual threads (use ReentrantLock)
     *   → Best for I/O-bound, not CPU-bound work
     *   → Works with existing APIs (JDBC, HTTP, etc.) transparently
     */

    // ========================================================
    // 9. SEQUENCED COLLECTIONS (Java 21+)
    // ========================================================

    /*
     * Before Java 21: no common interface for "has first/last element"
     *   List.get(0) / List.get(size()-1)      — different from
     *   Deque.getFirst() / Deque.getLast()     — different from
     *   SortedSet.first() / SortedSet.last()   — no consistency!
     *
     * Java 21: SequencedCollection interface
     *   interface SequencedCollection<E> extends Collection<E> {
     *       SequencedCollection<E> reversed();
     *       void addFirst(E e);
     *       void addLast(E e);
     *       E getFirst();
     *       E getLast();
     *       E removeFirst();
     *       E removeLast();
     *   }
     *
     *   List, Deque, SortedSet, LinkedHashSet all implement this now.
     *
     *   Similarly: SequencedMap (LinkedHashMap, SortedMap)
     *     map.firstEntry(), map.lastEntry(), map.reversed()
     */

    // ========================================================
    // 10. OTHER NOTABLE FEATURES
    // ========================================================

    /*
     * LOCAL VARIABLE TYPE INFERENCE — var (Java 10+)
     *   YOU ALREADY HAVE THIS on Java 11!
     *   var list = new ArrayList<String>();  // inferred as ArrayList<String>
     *   var stream = list.stream();          // inferred type
     *   // Only for local variables, NOT fields/parameters/return types
     *
     * HELPFUL NULLPOINTEREXCEPTION (Java 14+)
     *   Before: "NullPointerException" — WHERE?
     *   After:  "Cannot invoke String.length() because the return
     *            value of Person.getName() is null" — MUCH better!
     *
     * UNNAMED VARIABLES — _ (Java 22+)
     *   try { ... } catch (Exception _) { log("failed"); }
     *   map.forEach((_, v) -> process(v));
     *   case Point(var x, _) -> useX(x);  // don't care about y
     *
     * STRING TEMPLATES (Java 21 preview, Java 23+)
     *   String msg = STR."Hello \{name}, you are \{age} years old";
     *   // Type-safe, injection-safe string interpolation
     *   // (Still evolving as of Java 23)
     *
     * STRUCTURED CONCURRENCY (Java 21 preview)
     *   try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {
     *       Subtask<String> user  = scope.fork(() -> fetchUser());
     *       Subtask<Integer> order = scope.fork(() -> fetchOrder());
     *       scope.join().throwIfFailed();
     *       return new Response(user.get(), order.get());
     *   }
     *   // All subtasks are children of the scope
     *   // If one fails, all are cancelled → no leaked threads
     *
     * SCOPED VALUES (Java 21 preview)
     *   static final ScopedValue<User> CURRENT_USER = ScopedValue.newInstance();
     *   ScopedValue.where(CURRENT_USER, user).run(() -> handleRequest());
     *   // Like ThreadLocal but for virtual threads, safer, immutable
     */

    // ========================================================
    // MAIN — Simulated demos (works on Java 11)
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== MODERN JAVA FEATURES (12-21+) ===\n");

        // --- Simulated Switch Expression ---
        System.out.println("--- Switch Expression (Java 14+) ---");
        String day = "SAT";
        // Old way (works on Java 11):
        String type;
        switch (day) {
            case "MON": case "TUE": case "WED": case "THU": case "FRI":
                type = "Weekday"; break;
            default:
                type = "Weekend"; break;
        }
        System.out.println("  " + day + " → " + type);
        System.out.println("  New way: String type = switch(day) { case \"SAT\", \"SUN\" -> \"Weekend\"; ... };");

        // --- Simulated Record ---
        System.out.println("\n--- Record (Java 16+) ---");
        // We simulate with old-style class
        System.out.println("  Old way: class Point { private final int x, y; + constructor + getters + equals + hashCode + toString }");
        System.out.println("  New way: record Point(int x, int y) {}");
        System.out.println("  That's it. ONE LINE. All methods generated.");

        // --- Simulated Sealed Class ---
        System.out.println("\n--- Sealed Classes (Java 17+) ---");
        System.out.println("  sealed interface Shape permits Circle, Rectangle {}");
        System.out.println("  record Circle(double r) implements Shape {}");
        System.out.println("  record Rectangle(double w, double h) implements Shape {}");
        System.out.println("  → Compiler guarantees exhaustive pattern matching");

        // --- Simulated Pattern Matching ---
        System.out.println("\n--- Pattern Matching instanceof (Java 16+) ---");
        Object obj = "Hello World";
        // Old way:
        if (obj instanceof String) {
            String s = (String) obj;
            System.out.println("  Old: cast manually → length = " + s.length());
        }
        // New way explanation:
        System.out.println("  New: if (obj instanceof String s) { s.length(); }  // no cast!");

        // --- Virtual Threads ---
        System.out.println("\n--- Virtual Threads (Java 21+) ---");
        System.out.println("  Platform threads: ~thousands (1MB stack each)");
        System.out.println("  Virtual threads:  ~millions (few KB each)");
        System.out.println("  Thread.ofVirtual().start(() -> doWork());");
        System.out.println("  Executors.newVirtualThreadPerTaskExecutor();");

        // --- Feature Timeline ---
        System.out.println("\n=== JAVA VERSION TIMELINE ===");
        System.out.println("  Java 8  (2014) LTS — Lambdas, Streams, Optional");
        System.out.println("  Java 9  (2017)     — Modules, JShell, List.of()");
        System.out.println("  Java 10 (2018)     — var keyword");
        System.out.println("  Java 11 (2018) LTS — HttpClient, var in lambdas ← YOU ARE HERE");
        System.out.println("  Java 12 (2019)     — Switch expressions (preview)");
        System.out.println("  Java 13 (2019)     — Text blocks (preview)");
        System.out.println("  Java 14 (2020)     — Records (preview), helpful NPE");
        System.out.println("  Java 15 (2020)     — Sealed classes (preview), text blocks (final)");
        System.out.println("  Java 16 (2021)     — Records (final), pattern instanceof (final)");
        System.out.println("  Java 17 (2021) LTS — Sealed classes (final)");
        System.out.println("  Java 20 (2023)     — Scoped values, structured concurrency (preview)");
        System.out.println("  Java 21 (2023) LTS — Virtual threads, pattern switch, record patterns");

        System.out.println("\n  RECOMMENDATION: Upgrade to Java 21 LTS for god-level features");
        System.out.println("  Command: sudo apt install openjdk-21-jdk");

        System.out.println("\n✓ Modern Java Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Install Java 21 and rewrite Chapter 34 (Design Patterns) using
 *    records and sealed classes for the product hierarchies.
 * 2. Rewrite the Chapter 45 Task entity as a record.
 * 3. Refactor all instanceof checks in previous chapters to use
 *    pattern matching instanceof.
 * 4. Create a virtual thread stress test: spawn 1 million virtual
 *    threads, each sleeping 1 second. Compare to platform threads.
 *
 * NEXT: Chapter 47 — Advanced Generics
 */
