/**
 * PHASE 1.3 — SINGLETON PATTERN
 *
 * One instance, shared everywhere.
 * Spring beans are singletons by default — this is the pattern behind it.
 */

// ============================================================
// VERSION 1: Eager Singleton (created immediately)
// ============================================================
class EagerDatabaseConnection {
    // Instance created at class-loading time — guaranteed single instance
    private static final EagerDatabaseConnection INSTANCE = new EagerDatabaseConnection();

    private String url = "jdbc:mysql://localhost:3306/mydb";

    // PRIVATE constructor — nobody can call "new"
    private EagerDatabaseConnection() {
        System.out.println("  [EagerDB] Connection created (happens ONCE)");
    }

    // Global access point
    public static EagerDatabaseConnection getInstance() {
        return INSTANCE;
    }

    public void query(String sql) {
        System.out.println("  [EagerDB] Executing: " + sql + " on " + url);
    }
}


// ============================================================
// VERSION 2: Lazy Singleton (created on first use)
// ============================================================
class LazyDatabaseConnection {
    private static LazyDatabaseConnection instance;  // Starts null

    private String url = "jdbc:postgresql://localhost:5432/mydb";

    private LazyDatabaseConnection() {
        System.out.println("  [LazyDB] Connection created (happens ONCE, on first use)");
    }

    // Created only when first requested
    public static LazyDatabaseConnection getInstance() {
        if (instance == null) {
            instance = new LazyDatabaseConnection();
        }
        return instance;
    }

    public void query(String sql) {
        System.out.println("  [LazyDB] Executing: " + sql + " on " + url);
    }
}


// ============================================================
// VERSION 3: Thread-Safe Singleton (production-grade)
// ============================================================
class ThreadSafeConnection {
    // volatile ensures visibility across threads
    private static volatile ThreadSafeConnection instance;

    private String url = "jdbc:h2:mem:testdb";

    private ThreadSafeConnection() {
        System.out.println("  [ThreadSafe] Connection created (thread-safe, happens ONCE)");
    }

    // Double-checked locking — fast AND safe
    public static ThreadSafeConnection getInstance() {
        if (instance == null) {                   // First check (no lock)
            synchronized (ThreadSafeConnection.class) {
                if (instance == null) {           // Second check (with lock)
                    instance = new ThreadSafeConnection();
                }
            }
        }
        return instance;
    }

    public void query(String sql) {
        System.out.println("  [ThreadSafe] Executing: " + sql + " on " + url);
    }
}


// ============================================================
// HOW SPRING DOES IT (simplified)
// ============================================================
class SpringLikeSingletonRegistry {
    // Spring stores all singletons in a Map (this is the real pattern!)
    private static final java.util.Map<String, Object> singletonMap = new java.util.HashMap<>();

    @SuppressWarnings("unchecked")
    public static <T> T getBean(String name, Class<T> clazz) {
        if (!singletonMap.containsKey(name)) {
            try {
                T instance = clazz.getDeclaredConstructor().newInstance();
                singletonMap.put(name, instance);
                System.out.println("  [Registry] Created new singleton: " + name);
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        } else {
            System.out.println("  [Registry] Returning existing singleton: " + name);
        }
        return (T) singletonMap.get(name);
    }
}

class UserService {
    public void findUser() { System.out.println("  Finding user..."); }
}

class OrderService {
    public void placeOrder() { System.out.println("  Placing order..."); }
}


public class Step1_Singleton {
    public static void main(String[] args) {

        System.out.println("=== EAGER SINGLETON ===");
        EagerDatabaseConnection db1 = EagerDatabaseConnection.getInstance();
        EagerDatabaseConnection db2 = EagerDatabaseConnection.getInstance();
        db1.query("SELECT * FROM users");
        System.out.println("  Same instance? " + (db1 == db2));  // true!

        System.out.println("\n=== LAZY SINGLETON ===");
        System.out.println("  (Not created yet — waiting for first call...)");
        LazyDatabaseConnection lazy1 = LazyDatabaseConnection.getInstance();
        LazyDatabaseConnection lazy2 = LazyDatabaseConnection.getInstance();
        lazy1.query("SELECT * FROM orders");
        System.out.println("  Same instance? " + (lazy1 == lazy2));  // true!

        System.out.println("\n=== THREAD-SAFE SINGLETON ===");
        ThreadSafeConnection ts1 = ThreadSafeConnection.getInstance();
        ThreadSafeConnection ts2 = ThreadSafeConnection.getInstance();
        ts1.query("INSERT INTO products VALUES(...)");
        System.out.println("  Same instance? " + (ts1 == ts2));  // true!

        System.out.println("\n=== SPRING-LIKE SINGLETON REGISTRY ===");
        System.out.println("  (This is how Spring's ApplicationContext actually works)\n");
        UserService us1 = SpringLikeSingletonRegistry.getBean("userService", UserService.class);
        UserService us2 = SpringLikeSingletonRegistry.getBean("userService", UserService.class);
        OrderService os1 = SpringLikeSingletonRegistry.getBean("orderService", OrderService.class);
        System.out.println("\n  us1 == us2? " + (us1 == us2));  // true — singleton!
        us1.findUser();
        os1.placeOrder();

        System.out.println("\n=== KEY TAKEAWAY ===");
        System.out.println("  In Spring, you NEVER write singleton code yourself.");
        System.out.println("  Just annotate with @Component — Spring's registry handles the rest.");
        System.out.println("  RULE: Singleton beans must be STATELESS (no mutable per-request data).");
    }
}
