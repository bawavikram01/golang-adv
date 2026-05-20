import java.lang.annotation.*;
import java.lang.reflect.*;
import java.util.*;

/**
 * PHASE 2.1 — BUILD YOUR OWN MINI-SPRING
 *
 * This is a simplified Spring container that:
 *   1. Scans for @Component classes
 *   2. Creates instances via reflection
 *   3. Resolves and injects dependencies automatically
 *   4. Stores beans as singletons
 *
 * After this, Spring has ZERO magic left for you.
 */

// ============================================================
// Our custom annotations (mirrors Spring's annotations)
// ============================================================

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.TYPE)
@interface Component {
    String value() default "";
}

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.CONSTRUCTOR)
@interface Autowired {}


// ============================================================
// Application classes (annotated — just like in Spring)
// ============================================================

@Component
class AppConfig {
    private String dbUrl = "jdbc:h2:mem:myapp";
    private String appName = "MyApp";

    public String getDbUrl() { return dbUrl; }
    public String getAppName() { return appName; }
    public String toString() { return "AppConfig{db=" + dbUrl + "}"; }
}

@Component
class UserRepository {
    private final AppConfig config;

    @Autowired
    public UserRepository(AppConfig config) {
        this.config = config;
    }

    public String find(String name) {
        return "User(" + name + ") from " + config.getDbUrl();
    }

    public String toString() { return "UserRepository{config=" + config + "}"; }
}

@Component
class NotificationService {
    public void notify(String user, String message) {
        System.out.println("      📧 Notification to " + user + ": " + message);
    }

    public String toString() { return "NotificationService{}"; }
}

@Component
class UserService {
    private final UserRepository repository;
    private final NotificationService notifications;

    @Autowired
    public UserService(UserRepository repository, NotificationService notifications) {
        this.repository = repository;
        this.notifications = notifications;
    }

    public void register(String name) {
        String user = repository.find(name);
        notifications.notify(name, "Welcome to the app!");
        System.out.println("      ✅ Registered: " + user);
    }

    public String toString() { return "UserService{repo=" + repository + ", notif=" + notifications + "}"; }
}


// ============================================================
// THE MINI-SPRING CONTAINER (ApplicationContext equivalent)
// ============================================================

class MiniSpringContainer {
    private final Map<Class<?>, Object> beans = new HashMap<>();
    private final List<Class<?>> componentClasses = new ArrayList<>();

    /**
     * Step 1: Register component classes (simulates component scanning).
     * In real Spring, it scans packages for @Component classes.
     */
    public void register(Class<?>... classes) {
        for (Class<?> clazz : classes) {
            if (clazz.isAnnotationPresent(Component.class)) {
                componentClasses.add(clazz);
                System.out.println("  [SCAN] Found @Component: " + clazz.getSimpleName());
            } else {
                System.out.println("  [SCAN] Skipped (no @Component): " + clazz.getSimpleName());
            }
        }
    }

    /**
     * Step 2: Create all beans and resolve dependencies.
     * This is the core of Spring's startup process.
     */
    public void refresh() {
        System.out.println("\n  [CONTAINER] Refreshing — creating beans...\n");

        for (Class<?> clazz : componentClasses) {
            if (!beans.containsKey(clazz)) {
                createBean(clazz);
            }
        }

        System.out.println("\n  [CONTAINER] All beans created! Total: " + beans.size());
    }

    /**
     * Creates a bean and resolves its dependencies recursively.
     * This is EXACTLY what Spring does (simplified).
     */
    private Object createBean(Class<?> clazz) {
        // Already created? Return the singleton.
        if (beans.containsKey(clazz)) {
            System.out.println("    [CACHE] " + clazz.getSimpleName() + " already exists — reusing");
            return beans.get(clazz);
        }

        System.out.println("    [CREATE] " + clazz.getSimpleName() + "...");

        try {
            // Find constructor (prefer @Autowired, fallback to default)
            Constructor<?> constructor = findConstructor(clazz);
            Object[] args = resolveDependencies(constructor);

            // Create instance via reflection
            Object bean = constructor.newInstance(args);
            beans.put(clazz, bean);

            System.out.println("    [DONE] " + clazz.getSimpleName() + " ✓");
            return bean;

        } catch (Exception e) {
            throw new RuntimeException("Failed to create bean: " + clazz.getSimpleName(), e);
        }
    }

    /**
     * Finds the constructor to use.
     * Prefers @Autowired constructor, falls back to no-arg constructor.
     */
    private Constructor<?> findConstructor(Class<?> clazz) {
        for (Constructor<?> c : clazz.getDeclaredConstructors()) {
            if (c.isAnnotationPresent(Autowired.class)) {
                return c;
            }
        }
        try {
            return clazz.getDeclaredConstructor(); // no-arg fallback
        } catch (NoSuchMethodException e) {
            throw new RuntimeException("No suitable constructor for: " + clazz.getSimpleName());
        }
    }

    /**
     * Resolves constructor parameters by creating their beans first.
     * This is DEPENDENCY RESOLUTION — the heart of DI.
     */
    private Object[] resolveDependencies(Constructor<?> constructor) {
        Class<?>[] paramTypes = constructor.getParameterTypes();
        Object[] args = new Object[paramTypes.length];

        for (int i = 0; i < paramTypes.length; i++) {
            System.out.println("      [RESOLVE] needs " + paramTypes[i].getSimpleName()
                    + " — creating/fetching...");
            args[i] = createBean(paramTypes[i]);  // Recursive! Create dependency first.
        }

        return args;
    }

    /**
     * Get a bean by type (like applicationContext.getBean(UserService.class))
     */
    @SuppressWarnings("unchecked")
    public <T> T getBean(Class<T> type) {
        T bean = (T) beans.get(type);
        if (bean == null) {
            throw new RuntimeException("No bean of type: " + type.getSimpleName());
        }
        return bean;
    }

    /**
     * Print all managed beans
     */
    public void printBeans() {
        System.out.println("\n  [REGISTRY] Managed beans:");
        beans.forEach((type, instance) ->
            System.out.println("    • " + type.getSimpleName() + " → " + instance)
        );
    }
}


// ============================================================
// MAIN — Using our mini-Spring container
// ============================================================
public class Step2_ManualDIContainer {
    public static void main(String[] args) {

        System.out.println("╔══════════════════════════════════════════════════╗");
        System.out.println("║   BUILDING OUR OWN MINI-SPRING CONTAINER        ║");
        System.out.println("║   (This is what happens when SpringBoot starts) ║");
        System.out.println("╚══════════════════════════════════════════════════╝\n");

        // ---- Step 1: Create the container ----
        System.out.println("=== STEP 1: Component Scanning ===\n");
        MiniSpringContainer container = new MiniSpringContainer();

        // In real Spring: it scans all packages for @Component classes
        container.register(
            AppConfig.class,
            UserRepository.class,
            NotificationService.class,
            UserService.class,
            String.class  // This will be skipped — no @Component!
        );


        // ---- Step 2: Create and wire all beans ----
        System.out.println("\n=== STEP 2: Bean Creation & Dependency Injection ===");
        container.refresh();


        // ---- Step 3: Show what's in the container ----
        System.out.println("\n=== STEP 3: The Container Registry ===");
        container.printBeans();


        // ---- Step 4: Use beans (like @Autowired in your code) ----
        System.out.println("\n\n=== STEP 4: Using Beans (like @Autowired) ===\n");
        UserService userService = container.getBean(UserService.class);
        userService.register("Alice");
        System.out.println();
        userService.register("Bob");


        // ---- Prove singletons ----
        System.out.println("\n\n=== SINGLETON PROOF ===");
        UserService us1 = container.getBean(UserService.class);
        UserService us2 = container.getBean(UserService.class);
        System.out.println("  Same instance? " + (us1 == us2));  // true!


        // ---- Summary ----
        System.out.println("\n\n=== WHAT JUST HAPPENED (same as Spring startup) ===");
        System.out.println("  1. SCAN    → Found classes with @Component");
        System.out.println("  2. RESOLVE → For each class, checked constructor parameters");
        System.out.println("  3. CREATE  → Created dependencies first (recursive), then the bean");
        System.out.println("  4. STORE   → Saved as singleton in a Map<Class, Object>");
        System.out.println("  5. INJECT  → Passed dependencies via constructor");
        System.out.println("  6. READY   → getBean() returns the wired singleton");
        System.out.println("\n  Spring does EXACTLY this + 100 more features (AOP, events, profiles...)");
        System.out.println("  But the CORE is what you just saw. No magic. Just reflection + a Map.");
    }
}
