/*
 * =============================================================
 * STRUCTURAL PATTERN 2: DECORATOR
 * =============================================================
 *
 * INTENT: Attach additional responsibilities to an object
 *         DYNAMICALLY. Provides a flexible alternative to subclassing.
 *
 * ANALOGY: Coffee shop — start with basic coffee, then ADD extras:
 *          milk, sugar, whipped cream. Each is a "decorator".
 *
 * USE WHEN:
 *   - Add behavior to individual objects without affecting others
 *   - Combine behaviors dynamically (mix and match)
 *   - Alternative to a combinatorial explosion of subclasses
 *
 * REAL EXAMPLES: Java I/O streams, Collections.unmodifiableList()
 */

public class DecoratorPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Coffee Shop Example
        // ═══════════════════════════════════════════════════════
        System.out.println("=== COFFEE SHOP DECORATOR ===");

        // Basic coffee
        Coffee coffee = new SimpleCoffee();
        System.out.println(coffee.getDescription() + " → $" + coffee.getCost());

        // Add milk
        coffee = new MilkDecorator(coffee);
        System.out.println(coffee.getDescription() + " → $" + coffee.getCost());

        // Add sugar
        coffee = new SugarDecorator(coffee);
        System.out.println(coffee.getDescription() + " → $" + coffee.getCost());

        // Add whipped cream
        coffee = new WhipCreamDecorator(coffee);
        System.out.println(coffee.getDescription() + " → $" + coffee.getCost());

        // One-liner: stack decorators
        System.out.println("\n--- One-liner stacking ---");
        Coffee fancyCoffee = new WhipCreamDecorator(
                new MilkDecorator(
                        new SugarDecorator(
                                new SimpleCoffee())));
        System.out.println(fancyCoffee.getDescription() + " → $" + fancyCoffee.getCost());

        // ═══════════════════════════════════════════════════════
        // Real-world: Data Stream Processing
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== DATA STREAM DECORATOR ===");

        DataSource file = new FileDataSource();

        // Wrap with encryption
        DataSource encrypted = new EncryptionDecorator(file);
        encrypted.writeData("Secret Message");
        System.out.println("Read: " + encrypted.readData());

        // Wrap with compression on top of encryption
        DataSource compressedEncrypted = new CompressionDecorator(
                new EncryptionDecorator(file));
        compressedEncrypted.writeData("Important Data");
        System.out.println("Read: " + compressedEncrypted.readData());

        // ═══════════════════════════════════════════════════════
        // Real-world: Notification Decorator
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== NOTIFICATION DECORATOR ===");

        Notifier notifier = new BaseNotifier();

        // Just email
        Notifier emailNotifier = new EmailDecorator(notifier);
        emailNotifier.send("Server is down!");

        System.out.println();

        // Email + SMS + Slack — stack decorators!
        Notifier fullNotifier = new SlackDecorator(
                new SmsDecorator(
                        new EmailDecorator(notifier)));
        fullNotifier.send("Critical: Database failure!");
    }
}

// ═══════════════════════════════════════════════════════════════
// COFFEE EXAMPLE
// ═══════════════════════════════════════════════════════════════

// Component interface
interface Coffee {
    String getDescription();
    double getCost();
}

// Concrete component
class SimpleCoffee implements Coffee {
    @Override public String getDescription() { return "Simple Coffee"; }
    @Override public double getCost() { return 2.00; }
}

// Base decorator — holds a reference to the wrapped component
abstract class CoffeeDecorator implements Coffee {
    protected Coffee wrappedCoffee;

    public CoffeeDecorator(Coffee coffee) {
        this.wrappedCoffee = coffee;
    }

    @Override
    public String getDescription() { return wrappedCoffee.getDescription(); }

    @Override
    public double getCost() { return wrappedCoffee.getCost(); }
}

// Concrete decorators — each adds something extra
class MilkDecorator extends CoffeeDecorator {
    public MilkDecorator(Coffee coffee) { super(coffee); }

    @Override
    public String getDescription() { return wrappedCoffee.getDescription() + " + Milk"; }

    @Override
    public double getCost() { return wrappedCoffee.getCost() + 0.50; }
}

class SugarDecorator extends CoffeeDecorator {
    public SugarDecorator(Coffee coffee) { super(coffee); }

    @Override
    public String getDescription() { return wrappedCoffee.getDescription() + " + Sugar"; }

    @Override
    public double getCost() { return wrappedCoffee.getCost() + 0.25; }
}

class WhipCreamDecorator extends CoffeeDecorator {
    public WhipCreamDecorator(Coffee coffee) { super(coffee); }

    @Override
    public String getDescription() { return wrappedCoffee.getDescription() + " + Whip Cream"; }

    @Override
    public double getCost() { return wrappedCoffee.getCost() + 0.75; }
}

// ═══════════════════════════════════════════════════════════════
// DATA STREAM EXAMPLE (like Java I/O)
// ═══════════════════════════════════════════════════════════════
interface DataSource {
    void writeData(String data);
    String readData();
}

class FileDataSource implements DataSource {
    private String data;

    @Override
    public void writeData(String data) {
        this.data = data;
        System.out.println("  [File] Wrote: " + data);
    }

    @Override
    public String readData() { return data; }
}

abstract class DataSourceDecorator implements DataSource {
    protected DataSource wrapped;

    public DataSourceDecorator(DataSource source) { this.wrapped = source; }

    @Override
    public void writeData(String data) { wrapped.writeData(data); }

    @Override
    public String readData() { return wrapped.readData(); }
}

class EncryptionDecorator extends DataSourceDecorator {
    public EncryptionDecorator(DataSource source) { super(source); }

    @Override
    public void writeData(String data) {
        System.out.println("  [Encrypt] Encrypting data...");
        super.writeData("🔒[" + data + "]🔒");
    }

    @Override
    public String readData() {
        System.out.println("  [Encrypt] Decrypting data...");
        return super.readData();
    }
}

class CompressionDecorator extends DataSourceDecorator {
    public CompressionDecorator(DataSource source) { super(source); }

    @Override
    public void writeData(String data) {
        System.out.println("  [Compress] Compressing data...");
        super.writeData(data);
    }

    @Override
    public String readData() {
        System.out.println("  [Compress] Decompressing data...");
        return super.readData();
    }
}

// ═══════════════════════════════════════════════════════════════
// NOTIFICATION DECORATOR
// ═══════════════════════════════════════════════════════════════
interface Notifier {
    void send(String message);
}

class BaseNotifier implements Notifier {
    @Override
    public void send(String message) {
        System.out.println("  📋 Log: " + message);
    }
}

abstract class NotifierDecorator implements Notifier {
    protected Notifier wrapped;
    public NotifierDecorator(Notifier notifier) { this.wrapped = notifier; }
    @Override public void send(String message) { wrapped.send(message); }
}

class EmailDecorator extends NotifierDecorator {
    public EmailDecorator(Notifier n) { super(n); }
    @Override public void send(String message) {
        super.send(message);
        System.out.println("  📧 Email sent: " + message);
    }
}

class SmsDecorator extends NotifierDecorator {
    public SmsDecorator(Notifier n) { super(n); }
    @Override public void send(String message) {
        super.send(message);
        System.out.println("  📱 SMS sent: " + message);
    }
}

class SlackDecorator extends NotifierDecorator {
    public SlackDecorator(Notifier n) { super(n); }
    @Override public void send(String message) {
        super.send(message);
        System.out.println("  💬 Slack sent: " + message);
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Decorator wraps an object and adds behavior.
 * ✦ Decorators are stackable — wrap decorator in decorator.
 * ✦ Both decorator and component implement the SAME interface.
 * ✦ Avoids subclass explosion:
 *     Without: CoffeeMilk, CoffeeSugar, CoffeeMilkSugar... (2^N classes)
 *     With: N decorator classes, mix and match freely.
 *
 * ✦ Java I/O is the classic example:
 *     new BufferedReader(new InputStreamReader(new FileInputStream("file")))
 *
 * ✦ vs Inheritance: Decorator is runtime composition.
 *   Inheritance is compile-time. Decorator is more flexible.
 *
 * COMPILE & RUN:
 *   javac DecoratorPattern.java && java DecoratorPattern
 */
