import java.lang.annotation.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.*;

/**
 * PHASE 1.1 — CHALLENGE SOLUTION
 * 
 * Combines ALL 5 Java concepts:
 *   1. Interface        → MessageSender
 *   2. Generics         → Used in List<Message>, Stream<T>
 *   3. Annotation       → @DefaultSender (custom)
 *   4. Reflection       → Find which class has @DefaultSender, create it
 *   5. Lambda & Streams → Filter important messages, send them
 */

// ============================================================
// CONCEPT 1: INTERFACE — The contract
// ============================================================
interface MessageSender {
    void send(String message);
}


// ============================================================
// CONCEPT 3: CUSTOM ANNOTATION
// ============================================================
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.TYPE)  // Applied to classes
@interface DefaultSender {
    // This marks which implementation should be the default
}


// ============================================================
// CONCEPT 1 (contd): TWO IMPLEMENTATIONS
// ============================================================

@DefaultSender  // <-- This one is marked as the default!
class EmailSender implements MessageSender {
    @Override
    public void send(String message) {
        System.out.println("  📧 [EMAIL] " + message);
    }
}

class SmsSender implements MessageSender {
    @Override
    public void send(String message) {
        System.out.println("  📱 [SMS] " + message);
    }
}


// ============================================================
// CONCEPT 2: GENERICS — A simple Message class
// ============================================================
class Message {
    private String text;
    private boolean important;

    Message(String text, boolean important) {
        this.text = text;
        this.important = important;
    }

    public String getText() { return text; }
    public boolean isImportant() { return important; }

    @Override
    public String toString() {
        return (important ? "⚡" : "  ") + " " + text;
    }
}


// ============================================================
// MAIN — Ties everything together
// ============================================================
public class Step6_Challenge {
    public static void main(String[] args) throws Exception {

        // ===========================================================
        // CONCEPT 4: REFLECTION — Find the @DefaultSender at runtime
        // ===========================================================
        System.out.println("=== STEP 1: Using REFLECTION to find @DefaultSender ===\n");

        // These are the candidate classes (in real Spring, component scanning finds them)
        List<Class<? extends MessageSender>> candidates = List.of(
            EmailSender.class,
            SmsSender.class
        );

        MessageSender defaultSender = null;

        for (Class<? extends MessageSender> clazz : candidates) {
            System.out.println("  Inspecting: " + clazz.getSimpleName());

            if (clazz.isAnnotationPresent(DefaultSender.class)) {
                System.out.println("    ✅ Found @DefaultSender! Creating instance via reflection...");

                // Create instance using reflection — no "new" keyword!
                defaultSender = clazz.getDeclaredConstructor().newInstance();

                System.out.println("    Created: " + defaultSender.getClass().getSimpleName());
            } else {
                System.out.println("    ❌ No @DefaultSender annotation");
            }
        }

        if (defaultSender == null) {
            throw new RuntimeException("No class annotated with @DefaultSender found!");
        }


        // ===========================================================
        // CONCEPT 2 & 5: GENERICS + STREAMS — Filter & send messages
        // ===========================================================
        System.out.println("\n=== STEP 2: Create messages (using GENERICS — List<Message>) ===\n");

        List<Message> messages = List.of(
            new Message("Server is on fire!", true),
            new Message("Weekly newsletter", false),
            new Message("Payment failed for Order #1234", true),
            new Message("New blog post published", false),
            new Message("Database backup failed!", true),
            new Message("User signed up", false)
        );

        System.out.println("  All messages:");
        messages.forEach(m -> System.out.println("    " + m));


        // ===========================================================
        // CONCEPT 5: LAMBDA & STREAMS — Filter important, send them
        // ===========================================================
        System.out.println("\n=== STEP 3: Using STREAMS + LAMBDA to filter & send ===\n");

        System.out.println("  Filtering important messages and sending via " 
            + defaultSender.getClass().getSimpleName() + ":\n");

        // Capture in a final variable for use inside lambda
        final MessageSender sender = defaultSender;

        List<String> sentMessages = messages.stream()
            .filter(Message::isImportant)            // LAMBDA (method reference) — keep only important
            .map(Message::getText)                   // LAMBDA (method reference) — extract text
            .peek(text -> sender.send(text))         // LAMBDA — send each one (side effect)
            .collect(Collectors.toList());            // GENERICS — collect into List<String>

        System.out.println("\n  Summary: Sent " + sentMessages.size() + " important messages");
        System.out.println("  Skipped " + (messages.size() - sentMessages.size()) + " non-important messages");


        // ===========================================================
        // BONUS: Swap the sender without changing ANY logic above
        // ===========================================================
        System.out.println("\n=== BONUS: INTERFACE power — swap to SMS, zero code change ===\n");

        // Just change the implementation — the rest of the code works identically
        MessageSender smsSender = new SmsSender();

        messages.stream()
            .filter(Message::isImportant)
            .map(Message::getText)
            .forEach(text -> smsSender.send(text));


        System.out.println("\n=== ALL 5 CONCEPTS USED ===");
        System.out.println("  1. Interface        → MessageSender (contract)");
        System.out.println("  2. Generics         → List<Message>, List<Class<? extends MessageSender>>");
        System.out.println("  3. Annotation       → @DefaultSender (custom, marks the default impl)");
        System.out.println("  4. Reflection       → Found @DefaultSender at runtime, created instance");
        System.out.println("  5. Lambda & Streams → .filter().map().peek().collect() pipeline");
        System.out.println("\n  This is EXACTLY how Spring works internally. Congratulations!");
    }
}


    //                   @DefaultSender          ← ANNOTATION (label)
    //                        │
    //                   EmailSender             ← implements INTERFACE (contract)
    //                        │
    //  REFLECTION finds it ──┘  creates instance via .newInstance()
    //                        │
    //                  defaultSender            ← typed as INTERFACE (MessageSender)
    //                        │
    //  messages.stream()     │                  ← GENERICS (List<Message>)
    //     .filter(important) │                  ← LAMBDA
    //     .map(getText)      │                  ← LAMBDA
    //     .peek(sender::send)┘                  ← LAMBDA calls the INTERFACE method
    //     .collect(toList())                    ← GENERICS (List<String>)

// This is a miniature Spring container. Spring does exactly this:

// Scans for classes with annotations (@Component)
// Uses reflection to create instances (beans)
// Injects them where needed (via interface type)
// Your code uses lambdas/streams to process data
