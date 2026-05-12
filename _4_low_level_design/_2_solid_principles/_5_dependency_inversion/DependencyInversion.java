/*
 * =============================================================
 * SOLID PRINCIPLE 5: DEPENDENCY INVERSION PRINCIPLE (DIP)
 * =============================================================
 *
 * "High-level modules should NOT depend on low-level modules.
 *  Both should depend on ABSTRACTIONS."
 *
 * "Abstractions should not depend on details.
 *  Details should depend on abstractions."
 *
 * Translation: Don't hardcode dependencies. Depend on interfaces.
 *              Inject the concrete implementation from outside.
 *
 * This is the foundation of Dependency Injection (DI).
 */

import java.util.List;
import java.util.ArrayList;

public class DependencyInversion {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // BAD: High-level OrderService depends on low-level MySQLDatabase
        // ═══════════════════════════════════════════════════════
        System.out.println("=== BAD: Direct Dependency ===");
        BadOrderService badService = new BadOrderService();
        badService.placeOrder("Laptop");
        // Problem: Want to switch to PostgreSQL? MODIFY BadOrderService!
        // Problem: Want to unit test? Can't mock the database!

        // ═══════════════════════════════════════════════════════
        // GOOD: Both depend on abstractions
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GOOD: Dependency Inversion ===");

        // Inject MySQL implementation
        Database mysqlDb = new MySQLDatabase();
        NotificationService emailNotif = new EmailNotificationService();
        OrderService service1 = new OrderService(mysqlDb, emailNotif);
        service1.placeOrder("Laptop");

        // Switch to MongoDB — ZERO changes to OrderService!
        System.out.println("\n--- Switching to MongoDB + SMS ---");
        Database mongoDb = new MongoDatabase();
        NotificationService smsNotif = new SmsNotificationService();
        OrderService service2 = new OrderService(mongoDb, smsNotif);
        service2.placeOrder("Phone");

        // Use InMemory for testing — ZERO changes to OrderService!
        System.out.println("\n--- Using InMemory for Testing ---");
        Database testDb = new InMemoryDatabase();
        NotificationService mockNotif = new ConsoleNotificationService();
        OrderService testService = new OrderService(testDb, mockNotif);
        testService.placeOrder("Test Item");

        // ═══════════════════════════════════════════════════════
        // Constructor Injection vs Setter Injection
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== INJECTION TYPES ===");

        // Constructor injection (preferred — guarantees valid state)
        PaymentProcessor processor = new PaymentProcessor(
                new StripePaymentGateway(),
                new FileAuditLogger()
        );
        processor.processPayment(99.99);

        // Setter injection (optional dependencies)
        processor.setFraudDetector(new BasicFraudDetector());
        processor.processPayment(50000.00);  // will trigger fraud check
    }
}

// ═══════════════════════════════════════════════════════════════
// BAD: Tight coupling — high-level depends on low-level
// ═══════════════════════════════════════════════════════════════
class BadOrderService {
    private BadMySQLDatabase database = new BadMySQLDatabase();  // HARDCODED!

    public void placeOrder(String item) {
        database.save("Order: " + item);
        System.out.println("  Order placed: " + item);
    }
    // To change database → must modify THIS class
    // To test → must have real MySQL running
}

class BadMySQLDatabase {
    public void save(String data) {
        System.out.println("  [MySQL] Saving: " + data);
    }
}

// ═══════════════════════════════════════════════════════════════
// GOOD: Both depend on abstractions
// ═══════════════════════════════════════════════════════════════

// ─── Abstractions (interfaces) ───
interface Database {
    void save(String data);
    String find(String id);
}

interface NotificationService {
    void notify(String message);
}

// ─── Low-level implementations ───
class MySQLDatabase implements Database {
    @Override public void save(String data) { System.out.println("  [MySQL] Saving: " + data); }
    @Override public String find(String id) { return "[MySQL] Found: " + id; }
}

class MongoDatabase implements Database {
    @Override public void save(String data) { System.out.println("  [MongoDB] Saving: " + data); }
    @Override public String find(String id) { return "[MongoDB] Found: " + id; }
}

class InMemoryDatabase implements Database {
    private List<String> store = new ArrayList<>();
    @Override public void save(String data) { store.add(data); System.out.println("  [InMemory] Saved: " + data); }
    @Override public String find(String id) { return "[InMemory] Found: " + id; }
}

class EmailNotificationService implements NotificationService {
    @Override public void notify(String msg) { System.out.println("  📧 Email: " + msg); }
}

class SmsNotificationService implements NotificationService {
    @Override public void notify(String msg) { System.out.println("  📱 SMS: " + msg); }
}

class ConsoleNotificationService implements NotificationService {
    @Override public void notify(String msg) { System.out.println("  [Console] " + msg); }
}

// ─── High-level module — depends ONLY on abstractions ───
class OrderService {
    private final Database database;           // interface, not implementation!
    private final NotificationService notifier; // interface, not implementation!

    // Constructor Injection — dependencies come from OUTSIDE
    public OrderService(Database database, NotificationService notifier) {
        this.database = database;
        this.notifier = notifier;
    }

    public void placeOrder(String item) {
        database.save("Order: " + item);
        notifier.notify("Order placed: " + item);
        System.out.println("  ✓ Order complete: " + item);
    }
}

// ═══════════════════════════════════════════════════════════════
// Advanced: Multiple injection types
// ═══════════════════════════════════════════════════════════════
interface PaymentGateway {
    boolean charge(double amount);
}

interface AuditLogger {
    void log(String action);
}

interface FraudDetector {
    boolean isFraudulent(double amount);
}

class StripePaymentGateway implements PaymentGateway {
    @Override
    public boolean charge(double amount) {
        System.out.println("  💳 Stripe: Charged $" + amount);
        return true;
    }
}

class FileAuditLogger implements AuditLogger {
    @Override
    public void log(String action) {
        System.out.println("  📝 Audit log: " + action);
    }
}

class BasicFraudDetector implements FraudDetector {
    @Override
    public boolean isFraudulent(double amount) {
        return amount > 10000;  // flag large transactions
    }
}

class PaymentProcessor {
    private final PaymentGateway gateway;       // required — constructor injection
    private final AuditLogger logger;            // required — constructor injection
    private FraudDetector fraudDetector;          // optional — setter injection

    // Constructor injection for REQUIRED dependencies
    public PaymentProcessor(PaymentGateway gateway, AuditLogger logger) {
        this.gateway = gateway;
        this.logger = logger;
    }

    // Setter injection for OPTIONAL dependencies
    public void setFraudDetector(FraudDetector detector) {
        this.fraudDetector = detector;
    }

    public void processPayment(double amount) {
        // Optional fraud check
        if (fraudDetector != null && fraudDetector.isFraudulent(amount)) {
            logger.log("FRAUD DETECTED: $" + amount);
            System.out.println("  ⚠️ Payment blocked — fraud detected!");
            return;
        }

        if (gateway.charge(amount)) {
            logger.log("Payment processed: $" + amount);
        }
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ High-level → uses interfaces, not concrete classes.
 * ✦ Low-level → implements interfaces.
 * ✦ Both point toward the ABSTRACTION layer.
 *
 * ✦ CONSTRUCTOR INJECTION: for required dependencies.
 *   - Object always valid after construction.
 *   - Makes dependencies explicit.
 *
 * ✦ SETTER INJECTION: for optional dependencies.
 *   - Can be changed after construction.
 *
 * ✦ Benefits:
 *   - Swap implementations without changing business logic
 *   - Unit test with mocks/fakes
 *   - Follow Open/Closed Principle automatically
 *
 * COMPILE & RUN:
 *   javac DependencyInversion.java && java DependencyInversion
 */
