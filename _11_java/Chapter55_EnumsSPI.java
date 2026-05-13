/*
 * ============================================================
 *  CHAPTER 55: ADVANCED ENUMS & SERVICE PROVIDER INTERFACE
 * ============================================================
 *  THE FINAL CHAPTER. Enums in Java are far more powerful than
 *  in any other language. SPI lets you build extensible systems.
 *  Master these and you've completed the language journey.
 *
 *  TOPICS:
 *    1. Enum as Full Classes (methods, fields, constructors)
 *    2. Enum with Abstract Methods (Strategy Pattern)
 *    3. Enum State Machine
 *    4. Enum Singleton (safest singleton in Java)
 *    5. EnumSet & EnumMap internals
 *    6. Enum Tricks & Anti-Patterns
 *    7. ServiceLoader / SPI — Plugin Architecture
 *    8. Building an SPI System
 *    9. Real-World SPI Examples
 * ============================================================
 */

import java.util.*;
import java.util.function.*;
import java.util.stream.*;

public class Chapter55_EnumsSPI {

    // ========================================================
    // 1. ENUM WITH BEHAVIOR — Strategy Pattern
    // ========================================================
    // Each constant provides its own implementation

    enum MathOp {
        ADD("+") {
            @Override public double apply(double a, double b) { return a + b; }
        },
        SUBTRACT("-") {
            @Override public double apply(double a, double b) { return a - b; }
        },
        MULTIPLY("*") {
            @Override public double apply(double a, double b) { return a * b; }
        },
        DIVIDE("/") {
            @Override public double apply(double a, double b) {
                if (b == 0) throw new ArithmeticException("Division by zero");
                return a / b;
            }
        },
        POWER("^") {
            @Override public double apply(double a, double b) { return Math.pow(a, b); }
        };

        private final String symbol;

        MathOp(String symbol) { this.symbol = symbol; }

        public abstract double apply(double a, double b);

        public String getSymbol() { return symbol; }

        // Find by symbol
        private static final Map<String, MathOp> BY_SYMBOL = new HashMap<>();
        static {
            for (MathOp op : values()) BY_SYMBOL.put(op.symbol, op);
        }
        public static Optional<MathOp> fromSymbol(String symbol) {
            return Optional.ofNullable(BY_SYMBOL.get(symbol));
        }
    }

    // ========================================================
    // 2. ENUM STATE MACHINE
    // ========================================================

    enum OrderState {
        CREATED {
            @Override public Set<OrderState> allowedTransitions() {
                return EnumSet.of(PAID, CANCELLED);
            }
        },
        PAID {
            @Override public Set<OrderState> allowedTransitions() {
                return EnumSet.of(SHIPPED, REFUNDED);
            }
        },
        SHIPPED {
            @Override public Set<OrderState> allowedTransitions() {
                return EnumSet.of(DELIVERED, RETURNED);
            }
        },
        DELIVERED {
            @Override public Set<OrderState> allowedTransitions() {
                return EnumSet.of(RETURNED);
            }
        },
        CANCELLED {
            @Override public Set<OrderState> allowedTransitions() {
                return EnumSet.noneOf(OrderState.class); // terminal
            }
        },
        REFUNDED {
            @Override public Set<OrderState> allowedTransitions() {
                return EnumSet.noneOf(OrderState.class); // terminal
            }
        },
        RETURNED {
            @Override public Set<OrderState> allowedTransitions() {
                return EnumSet.of(REFUNDED);
            }
        };

        public abstract Set<OrderState> allowedTransitions();

        public boolean canTransitionTo(OrderState next) {
            return allowedTransitions().contains(next);
        }

        public OrderState transitionTo(OrderState next) {
            if (!canTransitionTo(next)) {
                throw new IllegalStateException(
                    "Cannot transition from " + this + " to " + next);
            }
            return next;
        }
    }

    // ========================================================
    // 3. ENUM SINGLETON — The Best Singleton in Java
    // ========================================================

    enum DatabaseConnection {
        INSTANCE;

        private final Map<String, String> config = new HashMap<>();

        DatabaseConnection() {
            // Constructor runs exactly once, thread-safe by JVM guarantee
            config.put("url", "jdbc:h2:mem:test");
            config.put("user", "sa");
        }

        public String getConfig(String key) { return config.get(key); }

        public void query(String sql) {
            System.out.println("    [DB] Executing: " + sql);
        }
    }

    /*
     * WHY ENUM SINGLETON IS THE BEST:
     * 1. Thread-safe (JVM guarantees single initialization)
     * 2. Serialization-safe (only one instance, no readResolve needed)
     * 3. Reflection-safe (can't create via reflection)
     * 4. Lazy initialization (loaded when class is accessed)
     * 5. Simple!
     *
     * Compare with double-checked locking singleton:
     *   - Needs volatile + synchronized
     *   - Vulnerable to reflection attack
     *   - Needs readResolve for serialization
     *   - More code, more bugs
     */

    // ========================================================
    // 4. ENUM WITH FUNCTIONAL INTERFACES
    // ========================================================

    enum StringTransform {
        UPPER(String::toUpperCase),
        LOWER(String::toLowerCase),
        TRIM(String::trim),
        REVERSE(s -> new StringBuilder(s).reverse().toString()),
        CAPITALIZE(s -> s.isEmpty() ? s :
            Character.toUpperCase(s.charAt(0)) + s.substring(1).toLowerCase());

        private final UnaryOperator<String> function;

        StringTransform(UnaryOperator<String> function) {
            this.function = function;
        }

        public String apply(String input) { return function.apply(input); }

        // Compose multiple transforms
        public static UnaryOperator<String> compose(StringTransform... transforms) {
            return input -> {
                String result = input;
                for (StringTransform t : transforms) result = t.apply(result);
                return result;
            };
        }
    }

    // ========================================================
    // 5. ENUM FOR TYPE-SAFE FLAGS (replacing bit masks)
    // ========================================================

    enum Permission {
        READ, WRITE, EXECUTE, DELETE, ADMIN;

        // EnumSet is internally a bitmask — O(1) for add/remove/contains!
        // Way better than manual bit manipulation
    }

    static class User {
        final String name;
        final EnumSet<Permission> permissions;

        User(String name, Permission... perms) {
            this.name = name;
            this.permissions = perms.length > 0 ?
                EnumSet.of(perms[0], perms) : EnumSet.noneOf(Permission.class);
        }

        boolean hasPermission(Permission p) { return permissions.contains(p); }
        boolean hasAll(Permission... perms) {
            return permissions.containsAll(EnumSet.of(perms[0], perms));
        }
    }

    // ========================================================
    // 6. SERVICE PROVIDER INTERFACE (SPI)
    // ========================================================

    /*
     * SPI CONCEPT:
     * ──────────────────────────────────────────────────
     * YOU define an INTERFACE (the "service").
     * OTHERS provide IMPLEMENTATIONS (the "providers").
     * ServiceLoader DISCOVERS implementations at runtime.
     *
     * HOW IT WORKS:
     * 1. Define interface: com.example.MyService
     * 2. Providers implement it: com.provider.MyServiceImpl
     * 3. Provider creates file:
     *    META-INF/services/com.example.MyService
     *    containing: com.provider.MyServiceImpl
     * 4. ServiceLoader.load(MyService.class) finds all providers
     *
     * REAL-WORLD EXAMPLES:
     *   - JDBC drivers     (java.sql.Driver)
     *   - Charset providers (java.nio.charset.spi.CharsetProvider)
     *   - XML parsers      (javax.xml.parsers.DocumentBuilderFactory)
     *   - Logging           (java.util.logging / SLF4J)
     *   - File systems      (java.nio.file.spi.FileSystemProvider)
     *   - Cryptography      (java.security.Provider)
     *
     * Module System (Java 9+):
     *   module com.provider {
     *       provides com.example.MyService
     *           with com.provider.MyServiceImpl;
     *   }
     *
     * ──────────────────────────────────────────────────
     */

    // Simulated SPI — define a service interface
    interface TextFormatter {
        String getName();
        String format(String text);
        int priority(); // lower = higher priority
    }

    // "Providers" would normally be in separate JARs
    static class UpperFormatter implements TextFormatter {
        public String getName() { return "UPPER"; }
        public String format(String text) { return text.toUpperCase(); }
        public int priority() { return 10; }
    }

    static class MarkdownBoldFormatter implements TextFormatter {
        public String getName() { return "BOLD"; }
        public String format(String text) { return "**" + text + "**"; }
        public int priority() { return 20; }
    }

    static class HTMLEscapeFormatter implements TextFormatter {
        public String getName() { return "HTML_ESCAPE"; }
        public String format(String text) {
            return text.replace("&", "&amp;")
                       .replace("<", "&lt;")
                       .replace(">", "&gt;");
        }
        public int priority() { return 5; }
    }

    // Service registry (simulates ServiceLoader)
    static class FormatterRegistry {
        private final List<TextFormatter> formatters = new ArrayList<>();

        void register(TextFormatter formatter) {
            formatters.add(formatter);
            formatters.sort(Comparator.comparingInt(TextFormatter::priority));
        }

        Optional<TextFormatter> getByName(String name) {
            return formatters.stream()
                .filter(f -> f.getName().equals(name))
                .findFirst();
        }

        TextFormatter getHighestPriority() {
            return formatters.get(0);
        }

        List<TextFormatter> getAll() {
            return Collections.unmodifiableList(formatters);
        }

        // Apply all formatters in priority order
        String applyAll(String text) {
            String result = text;
            for (TextFormatter f : formatters) {
                result = f.format(result);
            }
            return result;
        }
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== CHAPTER 55: ADVANCED ENUMS & SPI ===\n");

        // --- 1. Enum Strategy Pattern ---
        System.out.println("--- 1. Enum Strategy (MathOp) ---\n");

        for (MathOp op : MathOp.values()) {
            System.out.printf("    10 %s 3 = %.1f%n", op.getSymbol(), op.apply(10, 3));
        }
        System.out.println("  Lookup '+': " + MathOp.fromSymbol("+"));
        System.out.println("  Lookup '?': " + MathOp.fromSymbol("?"));

        // --- 2. Enum State Machine ---
        System.out.println("\n--- 2. Enum State Machine ---\n");

        OrderState state = OrderState.CREATED;
        System.out.println("  Start: " + state);
        System.out.println("  Can go to PAID? " + state.canTransitionTo(OrderState.PAID));
        System.out.println("  Can go to SHIPPED? " + state.canTransitionTo(OrderState.SHIPPED));

        state = state.transitionTo(OrderState.PAID);
        System.out.println("  → " + state);
        state = state.transitionTo(OrderState.SHIPPED);
        System.out.println("  → " + state);
        state = state.transitionTo(OrderState.DELIVERED);
        System.out.println("  → " + state);

        try {
            state.transitionTo(OrderState.PAID); // illegal!
        } catch (IllegalStateException e) {
            System.out.println("  ✗ " + e.getMessage());
        }

        // Print state diagram
        System.out.println("\n  State Diagram:");
        for (OrderState s : OrderState.values()) {
            Set<OrderState> transitions = s.allowedTransitions();
            if (transitions.isEmpty()) {
                System.out.println("    " + s + " → (terminal)");
            } else {
                System.out.println("    " + s + " → " + transitions);
            }
        }

        // --- 3. Enum Singleton ---
        System.out.println("\n--- 3. Enum Singleton ---\n");

        DatabaseConnection db = DatabaseConnection.INSTANCE;
        System.out.println("  URL: " + db.getConfig("url"));
        db.query("SELECT * FROM users");
        System.out.println("  Same instance? " + (db == DatabaseConnection.INSTANCE));

        // --- 4. Enum + Functional ---
        System.out.println("\n--- 4. Enum Functional Transforms ---\n");

        String input = "  hello WORLD  ";
        for (StringTransform t : StringTransform.values()) {
            System.out.println("    " + t.name() + ": '" + t.apply(input) + "'");
        }

        // Compose
        var trimAndCapitalize = StringTransform.compose(
            StringTransform.TRIM, StringTransform.CAPITALIZE);
        System.out.println("    TRIM+CAPITALIZE: '" + trimAndCapitalize.apply(input) + "'");

        // --- 5. EnumSet as Bit Flags ---
        System.out.println("\n--- 5. EnumSet Permissions ---\n");

        User admin = new User("admin", Permission.READ, Permission.WRITE,
            Permission.EXECUTE, Permission.DELETE, Permission.ADMIN);
        User reader = new User("reader", Permission.READ);

        System.out.println("  admin has WRITE? " + admin.hasPermission(Permission.WRITE));
        System.out.println("  reader has WRITE? " + reader.hasPermission(Permission.WRITE));
        System.out.println("  admin has READ+WRITE? " +
            admin.hasAll(Permission.READ, Permission.WRITE));

        // EnumSet internals
        System.out.println("\n  EnumSet internals:");
        System.out.println("    EnumSet uses a long bitmask internally (up to 64 enum constants)");
        System.out.println("    RegularEnumSet (≤64) uses a single long");
        System.out.println("    JumboEnumSet (>64) uses long[]");
        System.out.println("    All operations are O(1) bitwise ops!");

        EnumSet<Permission> all = EnumSet.allOf(Permission.class);
        EnumSet<Permission> none = EnumSet.noneOf(Permission.class);
        EnumSet<Permission> complement = EnumSet.complementOf(
            EnumSet.of(Permission.READ, Permission.WRITE));
        System.out.println("    All: " + all);
        System.out.println("    None: " + none);
        System.out.println("    Complement of READ,WRITE: " + complement);

        // EnumMap — array-backed, O(1) get/put, maintains enum order
        System.out.println("\n  EnumMap:");
        EnumMap<Permission, String> descriptions = new EnumMap<>(Permission.class);
        descriptions.put(Permission.READ, "Can read files");
        descriptions.put(Permission.WRITE, "Can write files");
        descriptions.put(Permission.EXECUTE, "Can execute programs");
        descriptions.put(Permission.DELETE, "Can delete files");
        descriptions.put(Permission.ADMIN, "Full control");
        descriptions.forEach((k, v) -> System.out.println("    " + k + " → " + v));

        // --- 6. Enum Tricks ---
        System.out.println("\n--- 6. Enum Tricks ---\n");

        // Trick 1: Reverse lookup with cached map (already shown in MathOp)
        System.out.println("  Trick 1: Cached reverse lookup (see MathOp.fromSymbol)");

        // Trick 2: Implementing interfaces
        System.out.println("  Trick 2: Enums can implement interfaces");
        System.out.println("    enum Color implements Printable { ... }");

        // Trick 3: Enum with generics (doesn't work in Java! Enums can't be generic)
        System.out.println("  Trick 3: Enums CANNOT be generic (Java limitation)");
        System.out.println("    ❌ enum Box<T> { ... } // won't compile");

        // Trick 4: Enum.valueOf with safety
        try {
            MathOp op = MathOp.valueOf("NOSUCH");
        } catch (IllegalArgumentException e) {
            System.out.println("  Trick 4: valueOf throws: " + e.getMessage());
        }
        // Safe valueOf:
        Optional<MathOp> safeOp = Arrays.stream(MathOp.values())
            .filter(o -> o.name().equals("ADD"))
            .findFirst();
        System.out.println("  Safe valueOf: " + safeOp);

        // --- 7. SPI Demo ---
        System.out.println("\n--- 7. Service Provider Interface (SPI) ---\n");

        FormatterRegistry registry = new FormatterRegistry();
        registry.register(new UpperFormatter());
        registry.register(new MarkdownBoldFormatter());
        registry.register(new HTMLEscapeFormatter());

        System.out.println("  Registered formatters:");
        for (TextFormatter f : registry.getAll()) {
            System.out.println("    [" + f.priority() + "] " + f.getName()
                + " → " + f.format("Hello <World>"));
        }

        System.out.println("\n  Highest priority: " + registry.getHighestPriority().getName());
        System.out.println("  By name 'BOLD': " + registry.getByName("BOLD")
            .map(f -> f.format("text")).orElse("not found"));

        // --- 8. Real ServiceLoader usage ---
        System.out.println("\n--- 8. Real ServiceLoader ---\n");

        System.out.println("  ServiceLoader.load() scans META-INF/services/");
        System.out.println("  Example: loading all available CharsetProviders:");

        // Real ServiceLoader in action (built-in providers)
        java.util.ServiceLoader<java.nio.charset.spi.CharsetProvider> providers =
            java.util.ServiceLoader.load(java.nio.charset.spi.CharsetProvider.class);

        int count = 0;
        for (java.nio.charset.spi.CharsetProvider p : providers) {
            System.out.println("    Found: " + p.getClass().getName());
            count++;
        }
        System.out.println("    Total CharsetProviders: " + count);

        /*
         * TO CREATE YOUR OWN SPI:
         *
         * 1. Define interface:
         *    public interface MyPlugin {
         *        String name();
         *        void execute();
         *    }
         *
         * 2. Create implementation (can be in separate JAR):
         *    public class FooPlugin implements MyPlugin {
         *        public String name() { return "Foo"; }
         *        public void execute() { System.out.println("Foo!"); }
         *    }
         *
         * 3. Create provider config file:
         *    META-INF/services/com.example.MyPlugin
         *    Content: com.example.FooPlugin
         *
         * 4. Load at runtime:
         *    ServiceLoader<MyPlugin> loader = ServiceLoader.load(MyPlugin.class);
         *    for (MyPlugin plugin : loader) {
         *        plugin.execute();
         *    }
         *
         * 5. With Java Modules (Java 9+):
         *    module com.example.foo {
         *        requires com.example.api;
         *        provides com.example.MyPlugin
         *            with com.example.FooPlugin;
         *    }
         */

        // ====================================================
        // FINAL SUMMARY
        // ====================================================
        System.out.println("\n" + "=".repeat(60));
        System.out.println("  CONGRATULATIONS! YOU HAVE COMPLETED ALL 55 CHAPTERS.");
        System.out.println("=".repeat(60));
        System.out.println();
        System.out.println("  You now have god-level knowledge of Java THE LANGUAGE:");
        System.out.println();
        System.out.println("  FOUNDATIONS (Ch 1-12):");
        System.out.println("    Variables, Types, Operators, Control Flow, Methods,");
        System.out.println("    Arrays, Strings, Input/Output");
        System.out.println();
        System.out.println("  OOP (Ch 13-22):");
        System.out.println("    Classes, Inheritance, Polymorphism, Abstraction,");
        System.out.println("    Interfaces, Generics, Inner Classes, Enums, Records");
        System.out.println();
        System.out.println("  CORE LIBRARIES (Ch 23-33):");
        System.out.println("    Collections, Exceptions, I/O, NIO, Threads,");
        System.out.println("    Concurrency, Lambdas, Streams, Optional");
        System.out.println();
        System.out.println("  ADVANCED (Ch 34-45):");
        System.out.println("    Annotations, Reflection, Modules, Networking,");
        System.out.println("    JDBC, Serialization, Regex, Date/Time, Security");
        System.out.println();
        System.out.println("  GOD LEVEL (Ch 46-55):");
        System.out.println("    Modern Java 12-21, Advanced Generics, JMM,");
        System.out.println("    Advanced Concurrency, Functional Programming,");
        System.out.println("    Bytecode & MethodHandles, Proxies & ClassLoaders,");
        System.out.println("    Performance & JIT, Advanced I/O, Enums & SPI");
        System.out.println();
        System.out.println("  NEXT STEPS:");
        System.out.println("    → Build real projects to solidify knowledge");
        System.out.println("    → Contribute to open-source Java projects");
        System.out.println("    → Read JDK source code (java.util.HashMap is great)");
        System.out.println("    → When ready: Spring Boot, Quarkus, or Micronaut");
        System.out.println();
        System.out.println("  ✓ The Java Language: MASTERED.");
    }
}

/*
 * FINAL EXERCISES:
 * 1. Build a complete state machine for a vending machine using enums.
 * 2. Create a plugin system using real ServiceLoader with separate JARs.
 * 3. Implement a command pattern using enums with undo/redo support.
 * 4. Build a DSL (Domain Specific Language) using enum + builder pattern.
 *
 * THE END.
 */
