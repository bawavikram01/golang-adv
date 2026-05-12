/*
 * =============================================================
 * STRUCTURAL PATTERN 3: FACADE
 * =============================================================
 *
 * INTENT: Provide a SIMPLE interface to a complex subsystem.
 *
 * ANALOGY: Hotel concierge — you say "I want dinner",
 *          they handle restaurant reservation, taxi, payment.
 *          You don't deal with each subsystem individually.
 *
 * USE WHEN:
 *   - Complex system with many interacting classes
 *   - Client doesn't need to know internal details
 *   - You want to reduce coupling between client and subsystem
 *
 * REAL EXAMPLES: SLF4J (logging facade), JDBC, Spring Boot auto-config
 */

public class FacadePattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // WITHOUT Facade: Client must know every subsystem
        // ═══════════════════════════════════════════════════════
        System.out.println("=== WITHOUT FACADE (Complex) ===");
        // You'd have to do this every time:
        CPU cpu = new CPU();
        Memory memory = new Memory();
        HardDrive hd = new HardDrive();
        cpu.freeze();
        memory.load(0, hd.read(0, 1024));
        cpu.jump(0);
        cpu.execute();
        // Imagine doing this everywhere in your codebase!

        // ═══════════════════════════════════════════════════════
        // WITH Facade: One simple call
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== WITH FACADE (Simple) ===");
        ComputerFacade computer = new ComputerFacade();
        computer.start();
        computer.shutdown();

        // ═══════════════════════════════════════════════════════
        // Real-world: E-Commerce Order Facade
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== REAL-WORLD: Order Facade ===");

        OrderFacade orderSystem = new OrderFacade();
        boolean success = orderSystem.placeOrder("user123", "PROD-456", 2, "4111111111111111");

        System.out.println("\nOrder " + (success ? "SUCCEEDED ✓" : "FAILED ✗"));

        // Client doesn't need to know about inventory, payment,
        // shipping, or notification subsystems!
    }
}

// ═══════════════════════════════════════════════════════════════
// COMPUTER SUBSYSTEM (complex internals)
// ═══════════════════════════════════════════════════════════════
class CPU {
    public void freeze()  { System.out.println("  CPU: Freezing processor"); }
    public void jump(long position) { System.out.println("  CPU: Jumping to " + position); }
    public void execute() { System.out.println("  CPU: Executing instructions"); }
    public void halt()    { System.out.println("  CPU: Halting"); }
}

class Memory {
    public void load(long position, String data) {
        System.out.println("  Memory: Loading data at position " + position);
    }
    public void free() { System.out.println("  Memory: Freeing memory"); }
}

class HardDrive {
    public String read(long lba, int size) {
        System.out.println("  HardDrive: Reading " + size + " bytes from sector " + lba);
        return "boot_data";
    }
}

// ═══════════════════════════════════════════════════════════════
// FACADE: Simplifies the complex subsystem
// ═══════════════════════════════════════════════════════════════
class ComputerFacade {
    private CPU cpu;
    private Memory memory;
    private HardDrive hardDrive;

    public ComputerFacade() {
        this.cpu = new CPU();
        this.memory = new Memory();
        this.hardDrive = new HardDrive();
    }

    // ONE simple method replaces a complex sequence
    public void start() {
        System.out.println("  Starting computer...");
        cpu.freeze();
        memory.load(0, hardDrive.read(0, 1024));
        cpu.jump(0);
        cpu.execute();
        System.out.println("  ✓ Computer started!");
    }

    public void shutdown() {
        System.out.println("  Shutting down...");
        cpu.halt();
        memory.free();
        System.out.println("  ✓ Computer shut down!");
    }
}

// ═══════════════════════════════════════════════════════════════
// REAL-WORLD: E-Commerce Subsystems
// ═══════════════════════════════════════════════════════════════
class InventoryService {
    public boolean checkStock(String productId, int quantity) {
        System.out.println("  [Inventory] Checking stock for " + productId + " (qty=" + quantity + ")");
        return true;  // simulated: in stock
    }

    public void reserveItems(String productId, int quantity) {
        System.out.println("  [Inventory] Reserved " + quantity + "x " + productId);
    }
}

class PaymentService {
    public boolean processPayment(String userId, double amount, String cardNumber) {
        // Only use last 4 digits for display (security!)
        String maskedCard = "****" + cardNumber.substring(cardNumber.length() - 4);
        System.out.println("  [Payment] Processing $" + amount + " for " + userId + " (card: " + maskedCard + ")");
        return true;  // simulated success
    }
}

class ShippingService {
    public String createShipment(String userId, String productId, int quantity) {
        String trackingId = "SHIP-" + System.currentTimeMillis() % 10000;
        System.out.println("  [Shipping] Shipment " + trackingId + " created for " + userId);
        return trackingId;
    }
}

class NotificationServiceFacade {
    public void sendOrderConfirmation(String userId, String trackingId) {
        System.out.println("  [Notification] Order confirmation sent to " + userId + " (tracking: " + trackingId + ")");
    }
}

// ═══════════════════════════════════════════════════════════════
// ORDER FACADE: One clean method for the entire order flow
// ═══════════════════════════════════════════════════════════════
class OrderFacade {
    private InventoryService inventory = new InventoryService();
    private PaymentService payment = new PaymentService();
    private ShippingService shipping = new ShippingService();
    private NotificationServiceFacade notification = new NotificationServiceFacade();

    public boolean placeOrder(String userId, String productId, int quantity, String cardNumber) {
        System.out.println("  Processing order for " + userId + "...");

        // Step 1: Check inventory
        if (!inventory.checkStock(productId, quantity)) {
            System.out.println("  ✗ Out of stock!");
            return false;
        }

        // Step 2: Process payment
        double amount = quantity * 29.99;  // simplified pricing
        if (!payment.processPayment(userId, amount, cardNumber)) {
            System.out.println("  ✗ Payment failed!");
            return false;
        }

        // Step 3: Reserve items
        inventory.reserveItems(productId, quantity);

        // Step 4: Create shipment
        String trackingId = shipping.createShipment(userId, productId, quantity);

        // Step 5: Send notification
        notification.sendOrderConfirmation(userId, trackingId);

        return true;
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Facade provides a SIMPLE interface to a COMPLEX subsystem.
 * ✦ It doesn't hide the subsystem — advanced users can still access it.
 * ✦ Reduces coupling: client → facade → subsystems.
 * ✦ Facade is NOT a wrapper for one class (that's Adapter).
 *   Facade orchestrates MULTIPLE classes.
 *
 * ✦ Difference from Adapter:
 *   - Adapter: makes incompatible interface compatible (1:1)
 *   - Facade: simplifies complex subsystem (1:many)
 *
 * COMPILE & RUN:
 *   javac FacadePattern.java && java FacadePattern
 */
