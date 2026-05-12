/*
 * =============================================================
 * MODULE 5: ABSTRACTION — Hide Complexity, Show Essentials
 * =============================================================
 *
 * Abstraction = showing WHAT an object does, hiding HOW it does it.
 *
 * Two mechanisms in Java:
 *   1. ABSTRACT CLASSES — partial abstraction (some concrete, some abstract)
 *   2. INTERFACES       — full abstraction (contract only, no state*)
 *
 * When to use which?
 *   ┌──────────────────────┬───────────────────┬──────────────────┐
 *   │                      │  Abstract Class    │    Interface      │
 *   ├──────────────────────┼───────────────────┼──────────────────┤
 *   │ Fields               │  ✓ (any)          │  constants only  │
 *   │ Constructors         │  ✓                │  ✗               │
 *   │ Concrete methods     │  ✓                │  default methods │
 *   │ Multiple inheritance │  ✗ (single only)  │  ✓               │
 *   │ Relationship         │  IS-A             │  CAN-DO / HAS-A  │
 *   └──────────────────────┴───────────────────┴──────────────────┘
 *
 *   Rule of thumb:
 *   - Abstract class → shared state + behavior among related classes
 *   - Interface → a capability that unrelated classes can have
 */

import java.util.ArrayList;
import java.util.List;

public class Abstraction {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // PART 1: ABSTRACT CLASSES
        // ═══════════════════════════════════════════════════════
        System.out.println("=== ABSTRACT CLASSES ===");

        // Payment payment = new Payment(...);  // COMPILE ERROR — can't instantiate abstract
        Payment creditCard = new CreditCardPayment("4111-1111-1111-1111", 99.99);
        Payment upi = new UpiPayment("user@upi", 49.99);

        creditCard.processPayment();
        System.out.println();
        upi.processPayment();

        // ═══════════════════════════════════════════════════════
        // PART 2: INTERFACES — Multiple Capabilities
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== INTERFACES ===");

        // A class can implement MULTIPLE interfaces
        SmartPhone phone = new SmartPhone("iPhone 15");
        phone.call("Mom");
        phone.takePhoto();
        phone.browseWeb("google.com");
        phone.charge();  // default method from interface!

        // ─── Interface as type — program to the interface ───
        System.out.println("\n=== INTERFACE AS TYPE ===");
        List<Camera> cameraDevices = new ArrayList<>();
        cameraDevices.add(phone);
        cameraDevices.add(new DslrCamera("Canon EOS R5"));

        for (Camera cam : cameraDevices) {
            cam.takePhoto();  // SmartPhone and DslrCamera are totally unrelated!
        }

        // ═══════════════════════════════════════════════════════
        // PART 3: INTERFACE EVOLUTION (default & static methods)
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== DEFAULT & STATIC METHODS ===");
        Sortable.printSortInfo();  // static method on interface

        int[] data = {5, 3, 8, 1, 9};
        Sortable bubbleSort = new BubbleSort();
        Sortable quickSort  = new QuickSort();

        System.out.println("BubbleSort is stable? " + bubbleSort.isStable());
        System.out.println("QuickSort is stable?  " + quickSort.isStable());
    }
}

// ═══════════════════════════════════════════════════════════════
// ABSTRACT CLASS: Shared Template for Payments
// ═══════════════════════════════════════════════════════════════
abstract class Payment {
    protected double amount;

    public Payment(double amount) {
        this.amount = amount;
    }

    // Template method — defines the skeleton, subclasses fill in steps
    public final void processPayment() {
        if (validate()) {
            deductAmount();
            sendConfirmation();
        } else {
            System.out.println("  Payment validation failed!");
        }
    }

    // Abstract methods — each payment type implements differently
    protected abstract boolean validate();
    protected abstract void deductAmount();

    // Concrete method — same for all payment types
    protected void sendConfirmation() {
        System.out.println("  ✓ Payment of $" + amount + " confirmed.");
    }
}

class CreditCardPayment extends Payment {
    private String cardNumber;

    public CreditCardPayment(String cardNumber, double amount) {
        super(amount);
        this.cardNumber = cardNumber;
    }

    @Override
    protected boolean validate() {
        System.out.println("  Validating credit card: " + cardNumber.substring(cardNumber.length() - 4));
        return cardNumber.length() >= 13;
    }

    @Override
    protected void deductAmount() {
        System.out.println("  Charging $" + amount + " to card ending " + cardNumber.substring(cardNumber.length() - 4));
    }
}

class UpiPayment extends Payment {
    private String upiId;

    public UpiPayment(String upiId, double amount) {
        super(amount);
        this.upiId = upiId;
    }

    @Override
    protected boolean validate() {
        System.out.println("  Validating UPI ID: " + upiId);
        return upiId.contains("@");
    }

    @Override
    protected void deductAmount() {
        System.out.println("  Debiting $" + amount + " from UPI: " + upiId);
    }
}

// ═══════════════════════════════════════════════════════════════
// INTERFACES: Capabilities / Contracts
// ═══════════════════════════════════════════════════════════════
interface Phone {
    void call(String contact);

    // Default method — provides a default implementation
    default void charge() {
        System.out.println("  🔌 Charging via USB-C...");
    }
}

interface Camera {
    void takePhoto();
}

interface WebBrowser {
    void browseWeb(String url);
}

// SmartPhone implements MULTIPLE interfaces — has ALL capabilities
class SmartPhone implements Phone, Camera, WebBrowser {
    private String model;

    public SmartPhone(String model) { this.model = model; }

    @Override
    public void call(String contact) {
        System.out.println("  📞 " + model + " calling " + contact + "...");
    }

    @Override
    public void takePhoto() {
        System.out.println("  📷 " + model + " taking photo...");
    }

    @Override
    public void browseWeb(String url) {
        System.out.println("  🌐 " + model + " browsing " + url + "...");
    }
    // charge() is inherited from Phone's default method
}

// DslrCamera only implements Camera — totally unrelated to SmartPhone
class DslrCamera implements Camera {
    private String model;

    public DslrCamera(String model) { this.model = model; }

    @Override
    public void takePhoto() {
        System.out.println("  📸 " + model + " (DSLR) taking high-res photo...");
    }
}

// ═══════════════════════════════════════════════════════════════
// INTERFACE EVOLUTION: default + static methods (Java 8+)
// ═══════════════════════════════════════════════════════════════
interface Sortable {
    void sort(int[] array);

    // Default method — implementations can override or use as-is
    default boolean isStable() {
        return false;  // default: not stable
    }

    // Static method — utility, not tied to any instance
    static void printSortInfo() {
        System.out.println("Sortable interface defines a sorting contract.");
    }
}

class BubbleSort implements Sortable {
    @Override
    public void sort(int[] array) {
        // bubble sort implementation
    }

    @Override
    public boolean isStable() {
        return true;  // bubble sort IS stable
    }
}

class QuickSort implements Sortable {
    @Override
    public void sort(int[] array) {
        // quicksort implementation
    }
    // isStable() defaults to false — correct for quicksort
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ ABSTRACT CLASS: use for IS-A when classes share state/behavior.
 *   - Can have constructors, fields, concrete and abstract methods.
 *   - Single inheritance only.
 *
 * ✦ INTERFACE: use for CAN-DO (capabilities) across unrelated types.
 *   - Multiple implementation allowed.
 *   - Java 8+: default methods, static methods.
 *   - Java 9+: private methods in interfaces.
 *
 * ✦ "Program to an interface, not an implementation."
 *   - Use `List<Camera>`, not `List<SmartPhone>`.
 *
 * ✦ Template Method pattern = abstract class + final skeleton method
 *   + abstract hook methods. Very powerful.
 *
 * COMPILE & RUN:
 *   javac Abstraction.java && java Abstraction
 */
