/*
 * ============================================================
 *  CHAPTER 52: DYNAMIC PROXIES & CLASSLOADERS
 * ============================================================
 *  This is where Java becomes META. You generate classes at
 *  runtime, intercept method calls, and control how the JVM
 *  loads code. Frameworks like Spring, Hibernate, and Mockito
 *  are built on these foundations.
 *
 *  TOPICS:
 *    1. java.lang.reflect.Proxy — Dynamic Proxies
 *    2. InvocationHandler — Intercepting Calls
 *    3. Proxy Patterns: Logging, Caching, Lazy, Retry
 *    4. ClassLoader Hierarchy
 *    5. Custom ClassLoader
 *    6. Context ClassLoader
 *    7. Class Unloading
 *    8. Service Discovery via ClassLoader
 * ============================================================
 */

import java.lang.reflect.*;
import java.util.*;
import java.util.concurrent.*;

public class Chapter52_ProxiesClassLoaders {

    // ========================================================
    // 1. INTERFACES FOR PROXY DEMOS
    // ========================================================

    interface Greeter {
        String greet(String name);
        String farewell(String name);
    }

    interface Calculator {
        int add(int a, int b);
        int multiply(int a, int b);
    }

    interface DataService {
        String fetchData(String key);
    }

    // Real implementations
    static class SimpleGreeter implements Greeter {
        public String greet(String name) { return "Hello, " + name + "!"; }
        public String farewell(String name) { return "Goodbye, " + name + "!"; }
    }

    static class SimpleCalculator implements Calculator {
        public int add(int a, int b) { return a + b; }
        public int multiply(int a, int b) { return a * b; }
    }

    static class SlowDataService implements DataService {
        public String fetchData(String key) {
            // Simulate slow operation
            try { Thread.sleep(50); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
            return "Data for " + key + " at " + System.currentTimeMillis();
        }
    }

    // ========================================================
    // 2. INVOCATION HANDLERS
    // ========================================================

    // --- Logging Proxy ---
    static class LoggingHandler implements InvocationHandler {
        private final Object target;

        LoggingHandler(Object target) { this.target = target; }

        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            System.out.println("    → " + method.getName() + "(" + Arrays.toString(args) + ")");
            long start = System.nanoTime();
            try {
                Object result = method.invoke(target, args);
                long elapsed = (System.nanoTime() - start) / 1_000;
                System.out.println("    ← " + result + " [" + elapsed + "µs]");
                return result;
            } catch (InvocationTargetException e) {
                System.out.println("    ✗ Exception: " + e.getCause());
                throw e.getCause();
            }
        }
    }

    // --- Caching Proxy ---
    static class CachingHandler implements InvocationHandler {
        private final Object target;
        private final Map<String, Object> cache = new ConcurrentHashMap<>();

        CachingHandler(Object target) { this.target = target; }

        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            String key = method.getName() + ":" + Arrays.toString(args);

            if (cache.containsKey(key)) {
                System.out.println("    [CACHE HIT] " + key);
                return cache.get(key);
            }

            System.out.println("    [CACHE MISS] " + key);
            Object result = method.invoke(target, args);
            cache.put(key, result);
            return result;
        }
    }

    // --- Retry Proxy ---
    static class RetryHandler implements InvocationHandler {
        private final Object target;
        private final int maxRetries;

        RetryHandler(Object target, int maxRetries) {
            this.target = target;
            this.maxRetries = maxRetries;
        }

        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            Throwable lastException = null;
            for (int attempt = 1; attempt <= maxRetries; attempt++) {
                try {
                    return method.invoke(target, args);
                } catch (InvocationTargetException e) {
                    lastException = e.getCause();
                    System.out.println("    Retry " + attempt + "/" + maxRetries
                        + " failed: " + lastException.getMessage());
                }
            }
            throw lastException;
        }
    }

    // --- Access Control Proxy ---
    static class AccessControlHandler implements InvocationHandler {
        private final Object target;
        private final Set<String> allowedMethods;

        AccessControlHandler(Object target, Set<String> allowedMethods) {
            this.target = target;
            this.allowedMethods = allowedMethods;
        }

        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            if (!allowedMethods.contains(method.getName())) {
                throw new SecurityException("Access denied to: " + method.getName());
            }
            return method.invoke(target, args);
        }
    }

    // ========================================================
    // 3. PROXY FACTORY — Generic Creator
    // ========================================================

    @SuppressWarnings("unchecked")
    static <T> T createProxy(T target, Class<T> iface, InvocationHandler handler) {
        return (T) Proxy.newProxyInstance(
            iface.getClassLoader(),
            new Class[]{iface},
            handler
        );
    }

    // Compose multiple handlers
    static class CompositeHandler implements InvocationHandler {
        private final List<InvocationHandler> handlers;
        private final Object target;

        CompositeHandler(Object target, InvocationHandler... handlers) {
            this.target = target;
            this.handlers = Arrays.asList(handlers);
        }

        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            // Simple: just use logging + delegate
            // A real composite would chain them
            return method.invoke(target, args);
        }
    }

    // ========================================================
    // 4. CUSTOM CLASSLOADER
    // ========================================================

    /*
     * ClassLoader Hierarchy:
     *
     *   Bootstrap ClassLoader (native, loads rt.jar / java.base)
     *          ↓
     *   Platform ClassLoader (Java 9+) / Extension CL (Java 8)
     *          ↓
     *   Application ClassLoader (classpath)
     *          ↓
     *   Custom ClassLoaders (your code)
     *
     * RULES:
     *   1. Parent-delegation: ask parent first, only load if parent can't
     *   2. Visibility: child can see parent's classes, not vice versa
     *   3. Uniqueness: class identity = (ClassLoader, fully qualified name)
     *      Same .class file loaded by two different ClassLoaders → TWO DIFFERENT CLASSES
     */

    // A ClassLoader that transforms class bytes (example: uppercase all strings)
    static class MonitoringClassLoader extends ClassLoader {
        private int loadCount = 0;

        MonitoringClassLoader(ClassLoader parent) {
            super(parent);
        }

        @Override
        public Class<?> loadClass(String name) throws ClassNotFoundException {
            loadCount++;
            // Delegate to parent (standard behavior)
            return super.loadClass(name);
        }

        int getLoadCount() { return loadCount; }
    }

    // ClassLoader that loads class bytes from a Map (useful for testing)
    static class InMemoryClassLoader extends ClassLoader {
        private final Map<String, byte[]> classBytes = new HashMap<>();

        InMemoryClassLoader(ClassLoader parent) { super(parent); }

        void addClass(String name, byte[] bytes) {
            classBytes.put(name, bytes);
        }

        @Override
        protected Class<?> findClass(String name) throws ClassNotFoundException {
            byte[] bytes = classBytes.get(name);
            if (bytes != null) {
                return defineClass(name, bytes, 0, bytes.length);
            }
            throw new ClassNotFoundException(name);
        }
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) throws Exception {

        System.out.println("=== CHAPTER 52: DYNAMIC PROXIES & CLASSLOADERS ===\n");

        // ====================================================
        // 1. Basic Dynamic Proxy
        // ====================================================
        System.out.println("--- 1. Basic Proxy ---\n");

        Greeter realGreeter = new SimpleGreeter();

        // Create a proxy that logs all method calls
        Greeter loggingGreeter = (Greeter) Proxy.newProxyInstance(
            Greeter.class.getClassLoader(),
            new Class[]{Greeter.class},
            new LoggingHandler(realGreeter)
        );

        // These calls go through the proxy
        loggingGreeter.greet("Alice");
        loggingGreeter.farewell("Bob");

        // Verify it's a proxy
        System.out.println("\n  Is proxy? " + Proxy.isProxyClass(loggingGreeter.getClass()));
        System.out.println("  Proxy class: " + loggingGreeter.getClass().getName());

        // ====================================================
        // 2. Caching Proxy
        // ====================================================
        System.out.println("\n--- 2. Caching Proxy ---\n");

        DataService realService = new SlowDataService();
        DataService cachedService = createProxy(
            realService, DataService.class,
            new CachingHandler(realService)
        );

        cachedService.fetchData("key1");  // miss
        cachedService.fetchData("key1");  // hit!
        cachedService.fetchData("key2");  // miss

        // ====================================================
        // 3. Access Control Proxy
        // ====================================================
        System.out.println("\n--- 3. Access Control Proxy ---\n");

        Greeter restrictedGreeter = createProxy(
            realGreeter, Greeter.class,
            new AccessControlHandler(realGreeter, Set.of("greet")) // only greet allowed
        );

        System.out.println("  " + restrictedGreeter.greet("Alice"));  // OK
        try {
            restrictedGreeter.farewell("Alice");  // blocked!
        } catch (SecurityException e) {
            System.out.println("  ✗ " + e.getMessage());
        }

        // ====================================================
        // 4. Multi-Interface Proxy
        // ====================================================
        System.out.println("\n--- 4. Multi-Interface Proxy ---\n");

        // A single proxy implementing multiple interfaces
        Object multiProxy = Proxy.newProxyInstance(
            Chapter52_ProxiesClassLoaders.class.getClassLoader(),
            new Class[]{Greeter.class, Calculator.class},
            (proxy, method, methodArgs) -> {
                System.out.println("    Intercepted: " + method.getDeclaringClass().getSimpleName()
                    + "." + method.getName());
                // Route to the right implementation
                if (method.getDeclaringClass() == Greeter.class) {
                    return method.invoke(new SimpleGreeter(), methodArgs);
                } else if (method.getDeclaringClass() == Calculator.class) {
                    return method.invoke(new SimpleCalculator(), methodArgs);
                }
                // Handle Object methods (toString, etc.)
                if (method.getName().equals("toString")) return "MultiProxy";
                return null;
            }
        );

        Greeter g = (Greeter) multiProxy;
        Calculator c = (Calculator) multiProxy;
        System.out.println("  " + g.greet("World"));
        System.out.println("  add(3,4) = " + c.add(3, 4));

        // ====================================================
        // 5. Proxy Without Target (Virtual Object)
        // ====================================================
        System.out.println("\n--- 5. Virtual Object (No Real Target) ---\n");

        // Sometimes you don't HAVE a real object — the proxy IS the implementation
        // This is how Mockito and MyBatis mappers work
        DataService virtualService = (DataService) Proxy.newProxyInstance(
            DataService.class.getClassLoader(),
            new Class[]{DataService.class},
            (proxy, method, methodArgs) -> {
                if (method.getName().equals("fetchData")) {
                    return "Virtual data for: " + methodArgs[0];
                }
                return null;
            }
        );
        System.out.println("  " + virtualService.fetchData("virtual_key"));

        // ====================================================
        // 6. ClassLoader Hierarchy
        // ====================================================
        System.out.println("\n--- 6. ClassLoader Hierarchy ---\n");

        ClassLoader cl = Chapter52_ProxiesClassLoaders.class.getClassLoader();
        System.out.println("  Class ClassLoader chain:");
        while (cl != null) {
            System.out.println("    " + cl);
            cl = cl.getParent();
        }
        System.out.println("    null (Bootstrap ClassLoader)");

        // String is loaded by bootstrap
        System.out.println("\n  String's ClassLoader: " + String.class.getClassLoader());
        System.out.println("  ArrayList's ClassLoader: " + ArrayList.class.getClassLoader());

        // ====================================================
        // 7. Monitoring ClassLoader
        // ====================================================
        System.out.println("\n--- 7. Custom ClassLoader ---\n");

        MonitoringClassLoader monCL = new MonitoringClassLoader(
            ClassLoader.getSystemClassLoader()
        );

        Class<?> stringClass = monCL.loadClass("java.lang.String");
        Class<?> listClass = monCL.loadClass("java.util.List");
        System.out.println("  Loaded String: " + stringClass.getName());
        System.out.println("  Loaded List: " + listClass.getName());
        System.out.println("  Load count: " + monCL.getLoadCount());

        // ====================================================
        // 8. Class Identity
        // ====================================================
        System.out.println("\n--- 8. Class Identity ---\n");

        /*
         * CRITICAL CONCEPT:
         * Class identity = ClassLoader + Fully Qualified Name
         *
         * Same .class file loaded by two different ClassLoaders
         * produces TWO DIFFERENT Class objects.
         * They CANNOT be cast to each other!
         *
         * This is why you sometimes get:
         *   ClassCastException: com.Foo cannot be cast to com.Foo
         *   (same name but different ClassLoaders!)
         *
         * This is also how:
         *   - App servers isolate web apps
         *   - OSGi bundles manage versions
         *   - Module systems work
         */

        Class<?> c1 = ClassLoader.getSystemClassLoader().loadClass("java.lang.String");
        Class<?> c2 = ClassLoader.getSystemClassLoader().loadClass("java.lang.String");
        System.out.println("  Same ClassLoader, same class: " + (c1 == c2));   // true

        // ====================================================
        // 9. Context ClassLoader
        // ====================================================
        System.out.println("\n--- 9. Context ClassLoader ---\n");

        /*
         * Thread.currentThread().getContextClassLoader()
         *
         * WHY? Parent-delegation breaks when framework code (loaded by
         * parent CL) needs to find user code (loaded by child CL).
         *
         * Example: JDBC. The DriverManager is in java.sql (bootstrap),
         * but JDBC drivers are in the classpath (app CL). Bootstrap CL
         * can't see app CL classes. Solution: use the thread's context
         * ClassLoader to find driver classes.
         *
         * SPI (ServiceLoader) uses context ClassLoader for similar reasons.
         */

        ClassLoader contextCL = Thread.currentThread().getContextClassLoader();
        System.out.println("  Context CL: " + contextCL);
        System.out.println("  Same as system CL? " +
            (contextCL == ClassLoader.getSystemClassLoader()));

        // ====================================================
        // 10. Real-World Patterns
        // ====================================================
        System.out.println("\n--- 10. Real-World Patterns ---\n");

        System.out.println("  Where proxies are used:");
        System.out.println("    • Spring AOP — @Transactional, @Cacheable, @Async");
        System.out.println("    • Hibernate — lazy-loading of entity relationships");
        System.out.println("    • Mockito — mock() creates proxy objects");
        System.out.println("    • MyBatis — mapper interfaces → SQL execution");
        System.out.println("    • RPC — remote method invocation stubs");
        System.out.println("    • Java EE — EJB proxies for lifecycle management");

        System.out.println("\n  Where ClassLoaders are used:");
        System.out.println("    • App servers (Tomcat) — isolate web applications");
        System.out.println("    • OSGi — bundle versioning and isolation");
        System.out.println("    • Hot-reload — load new class versions at runtime");
        System.out.println("    • Plugin systems — load JARs dynamically");
        System.out.println("    • Encryption — decrypt class bytes before loading");

        System.out.println("\n  Proxy limitations:");
        System.out.println("    • Can only proxy INTERFACES (not classes)");
        System.out.println("    • For class proxying, use CGLIB or ByteBuddy");
        System.out.println("    • final classes/methods cannot be proxied");
        System.out.println("    • Method.invoke has overhead (use MethodHandle if hot path)");

        System.out.println("\n✓ Dynamic Proxies & ClassLoaders Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Build a @Timed annotation proxy that measures and logs execution time
 *    of every method call.
 * 2. Create a proxy-based mock framework: mock(Interface.class) returns a
 *    proxy that records calls and can verify them.
 * 3. Write a ClassLoader that loads encrypted .class files (decrypt before
 *    defineClass).
 * 4. Build a simple plugin system: scan a directory for JARs, load them with
 *    a URLClassLoader, find classes implementing a Plugin interface.
 *
 * NEXT: Chapter 53 — Performance & JMH
 */
