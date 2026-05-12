/*
 * =============================================================
 * MODULE 4: POLYMORPHISM — One Interface, Many Forms
 * =============================================================
 *
 * Polymorphism = "many forms". Same method name, different behavior.
 *
 * TWO TYPES:
 *   1. COMPILE-TIME (Static)  → Method Overloading
 *      - Same method name, different parameters
 *      - Resolved at compile time
 *
 *   2. RUNTIME (Dynamic)      → Method Overriding
 *      - Child redefines parent's method
 *      - Resolved at runtime based on actual object type
 *      - THIS is the superpower of OOP
 */

import java.util.ArrayList;
import java.util.List;

public class Polymorphism {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // PART 1: COMPILE-TIME POLYMORPHISM (Method Overloading)
        // ═══════════════════════════════════════════════════════
        System.out.println("=== COMPILE-TIME POLYMORPHISM ===");

        Calculator calc = new Calculator();
        System.out.println("add(2, 3)       = " + calc.add(2, 3));
        System.out.println("add(2, 3, 4)    = " + calc.add(2, 3, 4));
        System.out.println("add(2.5, 3.5)   = " + calc.add(2.5, 3.5));
        System.out.println("add(\"Hi\", \" there\") = " + calc.add("Hi", " there"));

        // ═══════════════════════════════════════════════════════
        // PART 2: RUNTIME POLYMORPHISM (Method Overriding)
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== RUNTIME POLYMORPHISM ===");

        // Parent reference, child object — THE KEY IDEA
        Notification email = new EmailNotification("user@example.com");
        Notification sms   = new SmsNotification("+1234567890");
        Notification push  = new PushNotification("user_device_token");

        // Same method call → different behavior based on ACTUAL object
        String message = "Your order has been shipped!";
        email.send(message);
        sms.send(message);
        push.send(message);

        // ─── The real power: polymorphic collection ───
        System.out.println("\n=== POLYMORPHIC COLLECTION ===");
        List<Notification> channels = new ArrayList<>();
        channels.add(email);
        channels.add(sms);
        channels.add(push);

        // NotificationService doesn't know or care about concrete types!
        NotificationService service = new NotificationService();
        service.notifyAll(channels, "Flash sale: 50% off!");

        // ─── Adding a NEW notification type requires ZERO changes to existing code ───
        System.out.println("\n=== EXTENSIBILITY: Adding Slack ===");
        channels.add(new SlackNotification("#general"));
        service.notifyAll(channels, "New deployment successful!");
    }
}

// ═══════════════════════════════════════════════════════════════
// COMPILE-TIME POLYMORPHISM: Same method, different signatures
// ═══════════════════════════════════════════════════════════════
class Calculator {
    public int add(int a, int b)           { return a + b; }
    public int add(int a, int b, int c)    { return a + b + c; }
    public double add(double a, double b)  { return a + b; }
    public String add(String a, String b)  { return a + b; }
    // Compiler picks the right method based on argument types
}

// ═══════════════════════════════════════════════════════════════
// RUNTIME POLYMORPHISM: Same interface, different implementations
// ═══════════════════════════════════════════════════════════════

// ─── Abstract base ───
abstract class Notification {
    protected String recipient;

    public Notification(String recipient) {
        this.recipient = recipient;
    }

    // The polymorphic method — each subclass implements differently
    public abstract void send(String message);

    // Template method pattern preview: common structure, varying steps
    public void sendWithLogging(String message) {
        System.out.println("[LOG] Preparing to send...");
        send(message);  // polymorphic call!
        System.out.println("[LOG] Sent successfully.");
    }
}

class EmailNotification extends Notification {
    public EmailNotification(String email) { super(email); }

    @Override
    public void send(String message) {
        System.out.println("  📧 EMAIL to " + recipient + ": " + message);
    }
}

class SmsNotification extends Notification {
    public SmsNotification(String phone) { super(phone); }

    @Override
    public void send(String message) {
        // SMS has character limit
        String truncated = message.length() > 160 ? message.substring(0, 157) + "..." : message;
        System.out.println("  📱 SMS to " + recipient + ": " + truncated);
    }
}

class PushNotification extends Notification {
    public PushNotification(String deviceToken) { super(deviceToken); }

    @Override
    public void send(String message) {
        System.out.println("  🔔 PUSH to device[" + recipient + "]: " + message);
    }
}

// Added LATER — existing code doesn't change!
class SlackNotification extends Notification {
    public SlackNotification(String channel) { super(channel); }

    @Override
    public void send(String message) {
        System.out.println("  💬 SLACK to " + recipient + ": " + message);
    }
}

// ─── This class works with ANY Notification — past, present, or future ───
class NotificationService {
    public void notifyAll(List<Notification> channels, String message) {
        for (Notification n : channels) {
            n.send(message);  // runtime polymorphism in action!
        }
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ OVERLOADING (compile-time): same name, different parameter lists.
 *   - Return type alone doesn't distinguish methods.
 *
 * ✦ OVERRIDING (runtime): child redefines parent's method.
 *   - The ACTUAL object type determines which version runs.
 *   - Parent reference + child object = polymorphism.
 *
 * ✦ The killer benefit: you can add new types WITHOUT modifying
 *   existing code. This is the Open/Closed Principle in action.
 *
 * ✦ "Program to an interface, not an implementation."
 *   - NotificationService depends on `Notification` (abstract),
 *     not on EmailNotification, SmsNotification, etc.
 *
 * COMPILE & RUN:
 *   javac Polymorphism.java && java Polymorphism
 */
