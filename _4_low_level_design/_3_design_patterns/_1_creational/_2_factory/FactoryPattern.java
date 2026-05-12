/*
 * =============================================================
 * CREATIONAL PATTERN 2: FACTORY METHOD
 * =============================================================
 *
 * INTENT: Define an interface for creating objects, but let
 *         subclasses decide which class to instantiate.
 *
 * USE WHEN:
 *   - You don't know the exact type at compile time
 *   - Object creation logic is complex
 *   - You want to centralize and encapsulate creation logic
 *   - New types should be addable without modifying existing code (OCP!)
 *
 * VARIATIONS SHOWN:
 *   1. Simple Factory (not a pattern, but useful)
 *   2. Factory Method (the real pattern)
 *   3. Abstract Factory (factory of factories)
 */

import java.util.Map;
import java.util.HashMap;

public class FactoryPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // 1. SIMPLE FACTORY — Centralized creation
        // ═══════════════════════════════════════════════════════
        System.out.println("=== SIMPLE FACTORY ===");
        Notification email = NotificationFactory.create("email");
        Notification sms = NotificationFactory.create("sms");
        Notification push = NotificationFactory.create("push");

        email.send("Hello via email");
        sms.send("Hello via SMS");
        push.send("Hello via push");

        // ═══════════════════════════════════════════════════════
        // 2. FACTORY METHOD — Subclasses decide what to create
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== FACTORY METHOD ===");

        DocumentCreator wordCreator = new WordDocumentCreator();
        DocumentCreator pdfCreator  = new PdfDocumentCreator();
        DocumentCreator htmlCreator = new HtmlDocumentCreator();

        // Client code doesn't know or care about concrete types
        processDocument(wordCreator);
        processDocument(pdfCreator);
        processDocument(htmlCreator);

        // ═══════════════════════════════════════════════════════
        // 3. ABSTRACT FACTORY — Family of related objects
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== ABSTRACT FACTORY ===");

        // Create a Windows-themed UI
        UIFactory windowsFactory = new WindowsUIFactory();
        renderUI(windowsFactory);

        // Create a Mac-themed UI — just swap the factory!
        System.out.println();
        UIFactory macFactory = new MacUIFactory();
        renderUI(macFactory);
    }

    // Client method — works with any DocumentCreator
    static void processDocument(DocumentCreator creator) {
        Document doc = creator.createDocument();  // factory method call
        doc.open();
        doc.save();
    }

    // Client method — works with any UIFactory
    static void renderUI(UIFactory factory) {
        Button btn = factory.createButton();
        Checkbox cb = factory.createCheckbox();
        TextField tf = factory.createTextField();

        btn.render();
        cb.render();
        tf.render();
    }
}

// ═══════════════════════════════════════════════════════════════
// 1. SIMPLE FACTORY
// ═══════════════════════════════════════════════════════════════
interface Notification {
    void send(String message);
}

class EmailNotification implements Notification {
    @Override public void send(String msg) { System.out.println("  📧 Email: " + msg); }
}

class SmsNotification implements Notification {
    @Override public void send(String msg) { System.out.println("  📱 SMS: " + msg); }
}

class PushNotification implements Notification {
    @Override public void send(String msg) { System.out.println("  🔔 Push: " + msg); }
}

// The factory encapsulates creation logic
class NotificationFactory {
    // Use a registry for true OCP — no if/else needed for new types
    private static final Map<String, java.util.function.Supplier<Notification>> registry = new HashMap<>();

    static {
        registry.put("email", EmailNotification::new);
        registry.put("sms", SmsNotification::new);
        registry.put("push", PushNotification::new);
    }

    public static Notification create(String type) {
        java.util.function.Supplier<Notification> supplier = registry.get(type.toLowerCase());
        if (supplier == null) {
            throw new IllegalArgumentException("Unknown notification type: " + type);
        }
        return supplier.get();
    }

    // New types can be registered without modifying this class
    public static void register(String type, java.util.function.Supplier<Notification> supplier) {
        registry.put(type.toLowerCase(), supplier);
    }
}

// ═══════════════════════════════════════════════════════════════
// 2. FACTORY METHOD PATTERN
// ═══════════════════════════════════════════════════════════════

// Product interface
interface Document {
    void open();
    void save();
}

class WordDocument implements Document {
    @Override public void open() { System.out.println("  📝 Opening Word document..."); }
    @Override public void save() { System.out.println("  📝 Saving as .docx"); }
}

class PdfDocument implements Document {
    @Override public void open() { System.out.println("  📄 Opening PDF document..."); }
    @Override public void save() { System.out.println("  📄 Saving as .pdf"); }
}

class HtmlDocument implements Document {
    @Override public void open() { System.out.println("  🌐 Opening HTML document..."); }
    @Override public void save() { System.out.println("  🌐 Saving as .html"); }
}

// Creator — defines the factory method
abstract class DocumentCreator {
    // THE FACTORY METHOD — subclasses override to create specific types
    public abstract Document createDocument();

    // Can have other logic that USES the factory method
    public Document createAndLog() {
        Document doc = createDocument();
        System.out.println("  Created: " + doc.getClass().getSimpleName());
        return doc;
    }
}

// Concrete creators — each knows how to create its type
class WordDocumentCreator extends DocumentCreator {
    @Override public Document createDocument() { return new WordDocument(); }
}

class PdfDocumentCreator extends DocumentCreator {
    @Override public Document createDocument() { return new PdfDocument(); }
}

class HtmlDocumentCreator extends DocumentCreator {
    @Override public Document createDocument() { return new HtmlDocument(); }
}

// ═══════════════════════════════════════════════════════════════
// 3. ABSTRACT FACTORY — Family of related products
// ═══════════════════════════════════════════════════════════════

// Product families
interface Button {
    void render();
}

interface Checkbox {
    void render();
}

interface TextField {
    void render();
}

// Windows family
class WindowsButton implements Button {
    @Override public void render() { System.out.println("  [Windows] Rendering button"); }
}
class WindowsCheckbox implements Checkbox {
    @Override public void render() { System.out.println("  [Windows] Rendering checkbox"); }
}
class WindowsTextField implements TextField {
    @Override public void render() { System.out.println("  [Windows] Rendering text field"); }
}

// Mac family
class MacButton implements Button {
    @Override public void render() { System.out.println("  [Mac] Rendering button"); }
}
class MacCheckbox implements Checkbox {
    @Override public void render() { System.out.println("  [Mac] Rendering checkbox"); }
}
class MacTextField implements TextField {
    @Override public void render() { System.out.println("  [Mac] Rendering text field"); }
}

// Abstract Factory
interface UIFactory {
    Button createButton();
    Checkbox createCheckbox();
    TextField createTextField();
}

class WindowsUIFactory implements UIFactory {
    @Override public Button createButton()       { return new WindowsButton(); }
    @Override public Checkbox createCheckbox()    { return new WindowsCheckbox(); }
    @Override public TextField createTextField()  { return new WindowsTextField(); }
}

class MacUIFactory implements UIFactory {
    @Override public Button createButton()       { return new MacButton(); }
    @Override public Checkbox createCheckbox()    { return new MacCheckbox(); }
    @Override public TextField createTextField()  { return new MacTextField(); }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ SIMPLE FACTORY: One class with a create() method.
 *   Use a registry (Map) to avoid if/else chains.
 *
 * ✦ FACTORY METHOD: Abstract creator, subclasses override createX().
 *   The client works with the abstract creator.
 *
 * ✦ ABSTRACT FACTORY: Creates FAMILIES of related objects.
 *   Windows factory → Windows button + Windows checkbox.
 *   Mac factory → Mac button + Mac checkbox.
 *   Guarantees consistency within a family.
 *
 * ✦ All factory patterns follow OCP — add new types without modify.
 * ✦ All factory patterns follow DIP — client depends on abstractions.
 *
 * COMPILE & RUN:
 *   javac FactoryPattern.java && java FactoryPattern
 */
