/*
 * =============================================================
 * SOLID PRINCIPLE 4: INTERFACE SEGREGATION PRINCIPLE (ISP)
 * =============================================================
 *
 * "No client should be forced to depend on methods it doesn't use."
 *
 * Translation: Many small, focused interfaces > one fat interface.
 *
 * Violations:
 *   - A class implements an interface but leaves methods empty
 *   - A class throws UnsupportedOperationException for some methods
 *   - You feel the need to "stub out" interface methods
 */

public class InterfaceSegregation {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // BAD: Fat interface forces empty implementations
        // ═══════════════════════════════════════════════════════
        System.out.println("=== BAD: Fat Interface ===");

        BadMultiFunctionDevice badPrinter = new BadOldPrinter();
        badPrinter.print("Hello");
        badPrinter.scan();   // does nothing — forced to implement!
        badPrinter.fax();    // does nothing — forced to implement!
        badPrinter.email();  // does nothing — forced to implement!
        System.out.println("  ^ OldPrinter was forced to implement scan/fax/email (all empty)");

        // ═══════════════════════════════════════════════════════
        // GOOD: Small, focused interfaces
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GOOD: Segregated Interfaces ===");

        // Old printer only implements what it CAN do
        Printer oldPrinter = new BasicPrinter();
        oldPrinter.print("Document.pdf");

        // Modern device implements multiple interfaces
        MultiFunctionPrinter modernDevice = new MultiFunctionPrinter();
        modernDevice.print("Report.pdf");
        modernDevice.scan();
        modernDevice.fax("Contract.pdf");
        modernDevice.sendEmail("boss@company.com", "Report.pdf");

        // ═══════════════════════════════════════════════════════
        // REAL-WORLD: Worker hierarchy
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== REAL-WORLD: Workers ===");

        Workable humanWorker = new HumanWorker("Alice");
        humanWorker.work();
        ((Eatable) humanWorker).eat();
        ((Sleepable) humanWorker).sleep();

        Workable robotWorker = new RobotWorker("RoboX");
        robotWorker.work();
        // robotWorker doesn't implement Eatable or Sleepable — correct!
        System.out.println("  Robot doesn't eat or sleep — and doesn't need to implement them!");
    }
}

// ═══════════════════════════════════════════════════════════════
// BAD: Fat interface — forces everyone to implement everything
// ═══════════════════════════════════════════════════════════════
interface BadMultiFunctionDevice {
    void print(String doc);
    void scan();
    void fax();
    void email();
}

class BadOldPrinter implements BadMultiFunctionDevice {
    @Override public void print(String doc) { System.out.println("  Printing: " + doc); }
    @Override public void scan()  { /* Can't scan! But forced to implement */ }
    @Override public void fax()   { /* Can't fax! But forced to implement */ }
    @Override public void email() { /* Can't email! But forced to implement */ }
    // 3 useless methods = ISP violation
}

// ═══════════════════════════════════════════════════════════════
// GOOD: Segregated interfaces — implement only what you need
// ═══════════════════════════════════════════════════════════════
interface Printer {
    void print(String doc);
}

interface Scanner {
    void scan();
}

interface Faxer {
    void fax(String doc);
}

interface Emailer {
    void sendEmail(String to, String doc);
}

// Basic printer — only implements Printer
class BasicPrinter implements Printer {
    @Override
    public void print(String doc) {
        System.out.println("  🖨️ BasicPrinter printing: " + doc);
    }
}

// Modern device — implements MULTIPLE small interfaces
class MultiFunctionPrinter implements Printer, Scanner, Faxer, Emailer {
    @Override
    public void print(String doc) {
        System.out.println("  🖨️ MultiFunctionPrinter printing: " + doc);
    }

    @Override
    public void scan() {
        System.out.println("  📃 MultiFunctionPrinter scanning...");
    }

    @Override
    public void fax(String doc) {
        System.out.println("  📠 MultiFunctionPrinter faxing: " + doc);
    }

    @Override
    public void sendEmail(String to, String doc) {
        System.out.println("  📧 MultiFunctionPrinter emailing " + doc + " to " + to);
    }
}

// ═══════════════════════════════════════════════════════════════
// REAL-WORLD: Worker example
// ═══════════════════════════════════════════════════════════════
interface Workable {
    void work();
}

interface Eatable {
    void eat();
}

interface Sleepable {
    void sleep();
}

// Human implements all three — makes sense
class HumanWorker implements Workable, Eatable, Sleepable {
    private String name;
    public HumanWorker(String name) { this.name = name; }

    @Override public void work()  { System.out.println("  👷 " + name + " working..."); }
    @Override public void eat()   { System.out.println("  🍕 " + name + " eating..."); }
    @Override public void sleep() { System.out.println("  😴 " + name + " sleeping..."); }
}

// Robot only implements Workable — robots don't eat or sleep
class RobotWorker implements Workable {
    private String id;
    public RobotWorker(String id) { this.id = id; }

    @Override public void work() { System.out.println("  🤖 " + id + " working 24/7..."); }
    // No eat(), no sleep() — because it's not forced to implement them!
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Prefer MANY small interfaces over ONE big interface.
 * ✦ A class should only implement interfaces it fully uses.
 * ✦ If you see empty method implementations → ISP violation.
 * ✦ ISP complements LSP: proper interfaces prevent LSP violations.
 * ✦ Think of interfaces as ROLES or CAPABILITIES:
 *     Printer, Scanner, Faxer (not MultiFunctionDevice)
 *     Workable, Eatable, Sleepable (not Worker)
 *
 * COMPILE & RUN:
 *   javac InterfaceSegregation.java && java InterfaceSegregation
 */
