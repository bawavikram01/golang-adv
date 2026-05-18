import java.lang.reflect.*;

/**
 * PHASE 1.3 — PROXY PATTERN
 *
 * A proxy wraps a real object and adds behavior BEFORE/AFTER method calls.
 * This is HOW @Transactional, @Cacheable, and AOP work in Spring.
 *
 * Your code calls: userService.save(user)
 * Actually runs:   proxy.save(user) → BEGIN TX → real.save(user) → COMMIT TX
 */

// ============================================================
// The interface (contract)
// ============================================================
interface PaymentService {
    void processPayment(String orderId, double amount);
    double getBalance();
}


// ============================================================
// The REAL implementation (your business logic)
// ============================================================
class RealPaymentService implements PaymentService {
    private double balance = 10000.0;

    public void processPayment(String orderId, double amount) {
        balance -= amount;
        System.out.println("    💰 Payment processed: $" + amount + " for order " + orderId);
    }

    public double getBalance() {
        return balance;
    }
}


// ============================================================
// VERSION 1: Static Proxy (manual, verbose)
// This is conceptually what Spring creates, but Spring does it dynamically.
// ============================================================
class LoggingPaymentProxy implements PaymentService {
    private final PaymentService real;  // The real service inside

    public LoggingPaymentProxy(PaymentService real) {
        this.real = real;
    }

    public void processPayment(String orderId, double amount) {
        // BEFORE (added by proxy)
        System.out.println("    [LOG] → Calling processPayment(" + orderId + ", $" + amount + ")");
        long start = System.currentTimeMillis();

        // ACTUAL CALL (delegates to real object)
        real.processPayment(orderId, amount);

        // AFTER (added by proxy)
        long elapsed = System.currentTimeMillis() - start;
        System.out.println("    [LOG] ← Completed in " + elapsed + "ms | Balance: $" + real.getBalance());
    }

    public double getBalance() {
        return real.getBalance();
    }
}


// ============================================================
// VERSION 2: JDK Dynamic Proxy (how Spring actually does it!)
// Uses java.lang.reflect.Proxy — the SAME API Spring uses.
// ============================================================
class TransactionalHandler implements InvocationHandler {
    private final Object target;  // The real object

    public TransactionalHandler(Object target) {
        this.target = target;
    }

    @Override
    public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
        // Only wrap methods that "modify" data (Spring uses @Transactional annotation)
        if (method.getName().startsWith("process") || method.getName().startsWith("save")) {
            System.out.println("    [TX] ▶ BEGIN TRANSACTION");
            try {
                Object result = method.invoke(target, args);  // Call real method
                System.out.println("    [TX] ✓ COMMIT TRANSACTION");
                return result;
            } catch (Exception e) {
                System.out.println("    [TX] ✗ ROLLBACK TRANSACTION — Error: " + e.getMessage());
                throw e;
            }
        } else {
            // No transaction needed for read-only methods
            return method.invoke(target, args);
        }
    }
}


// ============================================================
// VERSION 3: Caching Proxy (like @Cacheable)
// ============================================================
class CachingHandler implements InvocationHandler {
    private final Object target;
    private final java.util.Map<String, Object> cache = new java.util.HashMap<>();

    public CachingHandler(Object target) {
        this.target = target;
    }

    @Override
    public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
        // Create a cache key from method name + args
        String key = method.getName() + ":" + java.util.Arrays.toString(args);

        if (method.getName().startsWith("get") && cache.containsKey(key)) {
            System.out.println("    [CACHE] HIT — returning cached value for: " + key);
            return cache.get(key);
        }

        Object result = method.invoke(target, args);

        if (method.getName().startsWith("get")) {
            cache.put(key, result);
            System.out.println("    [CACHE] MISS — stored result for: " + key);
        }

        return result;
    }
}


// ============================================================
// Helper: Create proxies (like Spring does behind the scenes)
// ============================================================
class ProxyFactory {
    @SuppressWarnings("unchecked")
    public static <T> T createTransactionalProxy(T target, Class<T> iface) {
        return (T) Proxy.newProxyInstance(
            iface.getClassLoader(),
            new Class<?>[]{ iface },
            new TransactionalHandler(target)
        );
    }

    @SuppressWarnings("unchecked")
    public static <T> T createCachingProxy(T target, Class<T> iface) {
        return (T) Proxy.newProxyInstance(
            iface.getClassLoader(),
            new Class<?>[]{ iface },
            new CachingHandler(target)
        );
    }
}


public class Step3_Proxy {
    public static void main(String[] args) {

        // ---- Static Proxy (manual) ----
        System.out.println("=== STATIC PROXY (Logging) ===");
        System.out.println("  Every call goes through the proxy first\n");

        PaymentService real = new RealPaymentService();
        PaymentService logged = new LoggingPaymentProxy(real);

        // You call the proxy — it delegates to the real service + adds logging
        logged.processPayment("ORD-001", 150.0);
        System.out.println();
        logged.processPayment("ORD-002", 300.0);


        // ---- Dynamic Proxy (Transactional) — like @Transactional ----
        System.out.println("\n\n=== DYNAMIC PROXY (Transactional) ===");
        System.out.println("  This is EXACTLY how Spring's @Transactional works\n");

        PaymentService realService = new RealPaymentService();
        PaymentService txProxy = ProxyFactory.createTransactionalProxy(realService, PaymentService.class);

        // To your code, it looks like a normal PaymentService call.
        // But the proxy invisibly wraps it in a transaction!
        txProxy.processPayment("ORD-100", 500.0);
        System.out.println();
        txProxy.processPayment("ORD-101", 250.0);

        System.out.println("\n  Balance (no TX needed): $" + txProxy.getBalance());


        // ---- Dynamic Proxy (Caching) — like @Cacheable ----
        System.out.println("\n\n=== DYNAMIC PROXY (Caching) ===");
        System.out.println("  This is EXACTLY how Spring's @Cacheable works\n");

        PaymentService cachedProxy = ProxyFactory.createCachingProxy(new RealPaymentService(), PaymentService.class);

        // First call — cache MISS, calls real method
        System.out.println("  First call:");
        double balance1 = cachedProxy.getBalance();
        System.out.println("  Balance: $" + balance1);

        // Second call — cache HIT, skips real method entirely!
        System.out.println("\n  Second call (same method, same args):");
        double balance2 = cachedProxy.getBalance();
        System.out.println("  Balance: $" + balance2);


        // ---- Show it's the same interface ----
        System.out.println("\n\n=== THE MAGIC ===");
        System.out.println("  logged instanceof PaymentService?  " + (logged instanceof PaymentService));
        System.out.println("  txProxy instanceof PaymentService? " + (txProxy instanceof PaymentService));
        System.out.println("  cachedProxy instanceof PaymentService? " + (cachedProxy instanceof PaymentService));
        System.out.println("\n  All look like PaymentService to your code.");
        System.out.println("  Your code NEVER knows it's talking to a proxy.");
        System.out.println("  This is why Spring annotations 'magically' add behavior.");

        System.out.println("\n=== KEY TAKEAWAY ===");
        System.out.println("  @Transactional → Spring wraps your bean in a TransactionalProxy");
        System.out.println("  @Cacheable     → Spring wraps your bean in a CachingProxy");
        System.out.println("  @Async         → Spring wraps your bean in an AsyncProxy");
        System.out.println("  All use java.lang.reflect.Proxy or CGLIB under the hood.");
    }
}
