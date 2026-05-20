/**
 * PHASE 2.2 — IoC: TRADITIONAL vs INVERTED CONTROL
 *
 * Shows the same application written TWO ways:
 *   1. Traditional: Classes create their own dependencies (tight coupling)
 *   2. IoC: Dependencies arrive from outside (loose coupling)
 *
 * You'll SEE why IoC makes code testable, flexible, and maintainable.
 */

import java.util.*;

// ============================================================
// SCENARIO: An Order Processing System
// ============================================================

// ─────────────────────────────────────────────────────
// VERSION 1: TRADITIONAL (Class controls its own deps)
// ─────────────────────────────────────────────────────

class TraditionalPaymentGateway {
    private String apiKey = "sk_live_hardcoded_key"; // Hardcoded!

    public boolean charge(String orderId, double amount) {
        System.out.println("    [PAY] Charged $" + amount + " for " + orderId + " (key: " + apiKey + ")");
        return true;
    }
}

class TraditionalInventory {
    public void reserve(String product, int qty) {
        System.out.println("    [INV] Reserved " + qty + "x " + product);
    }
}

class TraditionalEmailer {
    public void send(String to, String msg) {
        System.out.println("    [EMAIL] → " + to + ": " + msg);
    }
}

// THIS IS THE PROBLEM:
class TraditionalOrderService {
    // Class creates ALL its own dependencies — IT is in control
    private TraditionalPaymentGateway payment = new TraditionalPaymentGateway();
    private TraditionalInventory inventory = new TraditionalInventory();
    private TraditionalEmailer emailer = new TraditionalEmailer();

    public void processOrder(String orderId, String product, double amount, String email) {
        System.out.println("  Processing order: " + orderId);
        payment.charge(orderId, amount);
        inventory.reserve(product, 1);
        emailer.send(email, "Order " + orderId + " confirmed!");
        System.out.println("    ✅ Done\n");
    }
}


// ─────────────────────────────────────────────────────
// VERSION 2: IoC STYLE (Dependencies injected from outside)
// ─────────────────────────────────────────────────────

// Interfaces — contracts (just like Spring expects)
interface PaymentGateway {
    boolean charge(String orderId, double amount);
}

interface InventoryService {
    void reserve(String product, int qty);
}

interface Notifier {
    void notify(String to, String message);
}

// Implementation A: Real payment
class StripePayment implements PaymentGateway {
    private final String apiKey;
    public StripePayment(String apiKey) { this.apiKey = apiKey; }

    public boolean charge(String orderId, double amount) {
        System.out.println("    [STRIPE] Charged $" + amount + " for " + orderId);
        return true;
    }
}

// Implementation B: Fake payment (for testing!)
class FakePayment implements PaymentGateway {
    public boolean charge(String orderId, double amount) {
        System.out.println("    [FAKE-PAY] Simulated charge $" + amount + " (no real money moved)");
        return true;
    }
}

// Implementation: Real inventory
class WarehouseInventory implements InventoryService {
    public void reserve(String product, int qty) {
        System.out.println("    [WAREHOUSE] Reserved " + qty + "x " + product);
    }
}

// Implementation A: Email
class EmailNotifier implements Notifier {
    public void notify(String to, String message) {
        System.out.println("    [EMAIL] → " + to + ": " + message);
    }
}

// Implementation B: SMS
class SmsNotifier implements Notifier {
    public void notify(String to, String message) {
        System.out.println("    [SMS] → " + to + ": " + message);
    }
}

// IoC-STYLE: Class declares what it NEEDS, doesn't create anything
class IoCOrderService {
    private final PaymentGateway payment;
    private final InventoryService inventory;
    private final Notifier notifier;

    // Constructor receives ALL dependencies from OUTSIDE
    // This class has NO IDEA what concrete classes it's using!
    public IoCOrderService(PaymentGateway payment, InventoryService inventory, Notifier notifier) {
        this.payment = payment;
        this.inventory = inventory;
        this.notifier = notifier;
    }

    public void processOrder(String orderId, String product, double amount, String contact) {
        System.out.println("  Processing order: " + orderId);
        payment.charge(orderId, amount);
        inventory.reserve(product, 1);
        notifier.notify(contact, "Order " + orderId + " confirmed!");
        System.out.println("    ✅ Done\n");
    }
}


// ============================================================
// MAIN — Compare the two approaches
// ============================================================
public class Step1_IoCComparison {
    public static void main(String[] args) {

        System.out.println("╔═══════════════════════════════════════════════╗");
        System.out.println("║   TRADITIONAL CONTROL vs IoC                  ║");
        System.out.println("╚═══════════════════════════════════════════════╝\n");

        // ---- Traditional ----
        System.out.println("=== 1. TRADITIONAL (Class controls everything) ===\n");
        TraditionalOrderService traditional = new TraditionalOrderService();
        traditional.processOrder("ORD-001", "Laptop", 999.99, "alice@mail.com");

        System.out.println("  PROBLEMS:");
        System.out.println("  • Can't test without hitting real payment API");
        System.out.println("  • Can't switch from email to SMS without editing the class");
        System.out.println("  • API key is hardcoded inside (security risk)");
        System.out.println("  • Every class is a God object that creates everything\n");


        // ---- IoC: Production configuration ----
        System.out.println("\n=== 2. IoC — PRODUCTION CONFIGURATION ===\n");
        IoCOrderService productionService = new IoCOrderService(
            new StripePayment("sk_live_real_key"),  // Real payment
            new WarehouseInventory(),                // Real inventory
            new EmailNotifier()                      // Email notification
        );
        productionService.processOrder("ORD-002", "Phone", 699.99, "bob@mail.com");


        // ---- IoC: Test configuration (swapped implementations!) ----
        System.out.println("\n=== 3. IoC — TEST CONFIGURATION (Zero code change!) ===\n");
        IoCOrderService testService = new IoCOrderService(
            new FakePayment(),           // Fake payment — no real charges!
            new WarehouseInventory(),     // Can also be mocked
            new SmsNotifier()            // SMS instead of Email — no code change!
        );
        testService.processOrder("TEST-001", "Widget", 9.99, "+1234567890");


        // ---- IoC: Another configuration ----
        System.out.println("\n=== 4. IoC — ANOTHER CONFIGURATION (Still zero code change!) ===\n");
        IoCOrderService anotherConfig = new IoCOrderService(
            new FakePayment(),
            new WarehouseInventory(),
            // Even an inline lambda implementation!
            (to, msg) -> System.out.println("    [SLACK] → #" + to + ": " + msg)
        );
        anotherConfig.processOrder("DEV-001", "Test Item", 0.01, "dev-channel");


        // ---- The point ----
        System.out.println("\n=== THE POINT ===");
        System.out.println("  IoCOrderService was written ONCE.");
        System.out.println("  It ran with 3 COMPLETELY DIFFERENT configurations:");
        System.out.println("    • Production: Stripe + Warehouse + Email");
        System.out.println("    • Testing:    FakePayment + Warehouse + SMS");
        System.out.println("    • Dev:        FakePayment + Warehouse + Slack");
        System.out.println();
        System.out.println("  THE CLASS NEVER CHANGED. Only what's injected changed.");
        System.out.println("  This is the power of Inversion of Control.");
        System.out.println();
        System.out.println("  In Spring:");
        System.out.println("    • Spring is the one doing the 'new StripePayment(...)' part");
        System.out.println("    • 'Profiles' control which impl is used (dev vs prod)");
        System.out.println("    • You never write the wiring code yourself");
    }
}
