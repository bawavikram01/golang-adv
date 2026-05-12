/*
 * =============================================================
 * CREATIONAL PATTERN 1: SINGLETON
 * =============================================================
 *
 * INTENT: Ensure a class has ONLY ONE instance, and provide
 *         a global point of access to it.
 *
 * USE WHEN:
 *   - Database connection pool (one pool for the app)
 *   - Logger (one logger instance)
 *   - Configuration manager (one config)
 *   - Cache manager
 *
 * IMPLEMENTATIONS SHOWN:
 *   1. Eager initialization
 *   2. Lazy initialization (thread-safe with double-checked locking)
 *   3. Enum singleton (the BEST way in Java)
 */

public class SingletonPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Method 1: Eager Singleton
        // ═══════════════════════════════════════════════════════
        System.out.println("=== EAGER SINGLETON ===");
        EagerLogger log1 = EagerLogger.getInstance();
        EagerLogger log2 = EagerLogger.getInstance();
        log1.log("Hello from log1");
        log2.log("Hello from log2");
        System.out.println("Same instance? " + (log1 == log2));

        // ═══════════════════════════════════════════════════════
        // Method 2: Lazy Singleton (Thread-Safe)
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== LAZY SINGLETON (Thread-Safe) ===");
        DatabaseConnectionPool pool1 = DatabaseConnectionPool.getInstance();
        DatabaseConnectionPool pool2 = DatabaseConnectionPool.getInstance();
        pool1.getConnection();
        System.out.println("Same instance? " + (pool1 == pool2));

        // ═══════════════════════════════════════════════════════
        // Method 3: Enum Singleton (BEST)
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== ENUM SINGLETON (Best) ===");
        AppConfig.INSTANCE.set("db.host", "localhost");
        AppConfig.INSTANCE.set("db.port", "5432");
        System.out.println("db.host = " + AppConfig.INSTANCE.get("db.host"));
        System.out.println("db.port = " + AppConfig.INSTANCE.get("db.port"));
    }
}

// ═══════════════════════════════════════════════════════════════
// Method 1: EAGER — created at class loading time
// ═══════════════════════════════════════════════════════════════
// Pros: Simple, thread-safe
// Cons: Created even if never used
class EagerLogger {
    private static final EagerLogger INSTANCE = new EagerLogger();

    private EagerLogger() {
        System.out.println("  EagerLogger created (only once!)");
    }

    public static EagerLogger getInstance() {
        return INSTANCE;
    }

    public void log(String message) {
        System.out.println("  [LOG] " + message);
    }
}

// ═══════════════════════════════════════════════════════════════
// Method 2: LAZY with Double-Checked Locking
// ═══════════════════════════════════════════════════════════════
// Pros: Created only when needed, thread-safe
// Cons: Slightly more complex
class DatabaseConnectionPool {
    private static volatile DatabaseConnectionPool instance;  // volatile is KEY
    private int poolSize;

    private DatabaseConnectionPool() {
        this.poolSize = 10;
        System.out.println("  ConnectionPool created with " + poolSize + " connections");
    }

    public static DatabaseConnectionPool getInstance() {
        if (instance == null) {                      // 1st check (no lock)
            synchronized (DatabaseConnectionPool.class) {
                if (instance == null) {               // 2nd check (with lock)
                    instance = new DatabaseConnectionPool();
                }
            }
        }
        return instance;
    }

    public void getConnection() {
        System.out.println("  Got connection from pool (size=" + poolSize + ")");
    }
}

// ═══════════════════════════════════════════════════════════════
// Method 3: ENUM SINGLETON — The Josh Bloch recommended way
// ═══════════════════════════════════════════════════════════════
// Pros: Thread-safe, serialization-safe, reflection-safe, SIMPLE
// Cons: Can't do lazy initialization
enum AppConfig {
    INSTANCE;

    private final java.util.Map<String, String> properties = new java.util.HashMap<>();

    public void set(String key, String value) {
        properties.put(key, value);
    }

    public String get(String key) {
        return properties.getOrDefault(key, "not found");
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Singleton = exactly ONE instance of a class.
 * ✦ Private constructor prevents external instantiation.
 * ✦ Static method provides global access point.
 *
 * ✦ For Java, prefer ENUM singleton (Effective Java, Item 3):
 *   - Handles serialization automatically
 *   - Prevents reflection attacks
 *   - Thread-safe by JVM guarantee
 *
 * ✦ Use Double-Checked Locking only if you need lazy init + thread safety.
 *   - `volatile` keyword is CRITICAL (prevents instruction reordering).
 *
 * ⚠️ Singleton is sometimes called an anti-pattern because:
 *   - Global state makes testing harder
 *   - Hides dependencies
 *   Use it sparingly. Prefer Dependency Injection.
 *
 * COMPILE & RUN:
 *   javac SingletonPattern.java && java SingletonPattern
 */
