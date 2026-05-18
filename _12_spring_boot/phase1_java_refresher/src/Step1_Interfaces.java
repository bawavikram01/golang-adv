/**
 * PHASE 1.1 — INTERFACES
 * 
 * The #1 concept in Spring. Without this, nothing in Spring makes sense.
 * 
 * Theory
An interface is a contract. It says what something can do, without saying how.

Why does Spring care? Spring's entire philosophy is: "depend on abstractions, not concrete classes." Every time Spring injects a dependency, it prefers to inject an interface. This is what makes your code swappable, testable, and loosely coupled.

Analogy
Think of a power socket (interface) and appliances (implementations). The socket defines the shape (contract). Any appliance that matches the shape can plug in — a TV, a phone charger, a lamp. The socket doesn't care which appliance. It just provides power to anything that fits.

Spring is the socket. Your classes are the appliances.

 * KEY IDEA: Program to an interface, not an implementation.
 */

// ============================================================
// STEP 1: The Problem — Tight Coupling (BAD)
// ============================================================

// Imagine you're building a notification system.
// First, the WRONG way (what Spring helps you avoid):

class EmailService {
    public void send(String message) {
        System.out.println("  Sending EMAIL: " + message);
    }
}

// This class is TIGHTLY COUPLED to EmailService.
// If tomorrow you want SMS instead of email, you must CHANGE this class.
class OrderServiceBad {
    private EmailService emailService = new EmailService(); // <-- HARDCODED dependency

    public void placeOrder(String item) {
        System.out.println("[BAD] Order placed for: " + item);
        emailService.send("Your order for " + item + " is confirmed!");
    }
}


// ============================================================
// STEP 2: The Solution — Interface (GOOD)
// ============================================================

// Define a CONTRACT: "anything that can send a notification"
interface NotificationService {
    void send(String message);
    // No body! Just the signature. This is the CONTRACT.
}

// Implementation 1: Email
class EmailNotification implements NotificationService {
    @Override
    public void send(String message) {
        System.out.println("  📧 EMAIL: " + message);
    }
}

// Implementation 2: SMS
class SmsNotification implements NotificationService {
    @Override
    public void send(String message) {
        System.out.println("  📱 SMS: " + message);
    }
}

// Implementation 3: Push Notification
class PushNotification implements NotificationService {
    @Override
    public void send(String message) {
        System.out.println("  🔔 PUSH: " + message);
    }
}

// Now OrderService depends on the INTERFACE, not a concrete class.
// It doesn't know or care if it's email, SMS, or push.
class OrderServiceGood {
    private NotificationService notificationService; // <-- INTERFACE type

    // The implementation is INJECTED from outside (this is what Spring does!)
    public OrderServiceGood(NotificationService notificationService) {
        this.notificationService = notificationService;
    }

    public void placeOrder(String item) {
        System.out.println("[GOOD] Order placed for: " + item);
        notificationService.send("Your order for " + item + " is confirmed!");
    }
}


// ============================================================
// STEP 3: See the difference
// ============================================================

public class Step1_Interfaces {
    public static void main(String[] args) {

        System.out.println("=== TIGHT COUPLING (Bad Way) ===");
        OrderServiceBad bad = new OrderServiceBad();
        bad.placeOrder("Laptop");
        // Problem: If you want SMS, you must EDIT OrderServiceBad's source code.

        System.out.println();
        System.out.println("=== LOOSE COUPLING (Good Way — Interface) ===");

        // Same OrderServiceGood, different behaviors — NO code change needed!
        OrderServiceGood withEmail = new OrderServiceGood(new EmailNotification());
        withEmail.placeOrder("Laptop");

        System.out.println();

        OrderServiceGood withSms = new OrderServiceGood(new SmsNotification());
        withSms.placeOrder("Phone");

        System.out.println();

        OrderServiceGood withPush = new OrderServiceGood(new PushNotification());
        withPush.placeOrder("Headphones");

        System.out.println();
        System.out.println("=== KEY TAKEAWAY ===");
        System.out.println("OrderServiceGood never changed, but behavior changed!");
        System.out.println("This is EXACTLY what Spring's Dependency Injection does.");
        System.out.println("Spring picks the right implementation and injects it for you.");
    }
}
