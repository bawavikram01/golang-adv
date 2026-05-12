/*
 * =============================================================
 * SOLID PRINCIPLE 1: SINGLE RESPONSIBILITY PRINCIPLE (SRP)
 * =============================================================
 *
 * "A class should have only ONE reason to change."
 *
 * Translation: Each class should do ONE thing and do it well.
 *
 * Symptoms of SRP violation:
 *   - Class has methods that belong to different "actors"
 *   - Class name includes "And" or "Manager" (red flag)
 *   - Changes in one area break unrelated features
 *   - The class is growing into a "God object"
 */

import java.util.ArrayList;
import java.util.List;

public class SingleResponsibility {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // BAD: One class does EVERYTHING — invoice, printing, DB
        // ═══════════════════════════════════════════════════════
        System.out.println("=== BAD: SRP Violation ===");
        InvoiceGodClass badInvoice = new InvoiceGodClass("Alice", 100.0);
        badInvoice.calculateTotal();
        badInvoice.printInvoice();
        badInvoice.saveToDatabase();
        // This class has THREE reasons to change:
        // 1. Tax logic changes → modify calculateTotal
        // 2. Print format changes → modify printInvoice
        // 3. Database schema changes → modify saveToDatabase

        // ═══════════════════════════════════════════════════════
        // GOOD: Each class has ONE responsibility
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GOOD: SRP Applied ===");

        Invoice invoice = new Invoice("Bob", 200.0);
        invoice.addItem("Widget", 3, 50.0);
        invoice.addItem("Gadget", 1, 100.0);

        TaxCalculator taxCalc = new TaxCalculator(18.0);
        double total = taxCalc.calculateTotal(invoice);
        System.out.println("Total with tax: $" + String.format("%.2f", total));

        InvoicePrinter printer = new InvoicePrinter();
        printer.print(invoice, total);

        InvoiceRepository repo = new InvoiceRepository();
        repo.save(invoice);
    }
}

// ═══════════════════════════════════════════════════════════════
// BAD: God class — does everything
// ═══════════════════════════════════════════════════════════════
class InvoiceGodClass {
    private String customer;
    private double amount;
    private double taxRate = 18.0;

    public InvoiceGodClass(String customer, double amount) {
        this.customer = customer;
        this.amount = amount;
    }

    // Responsibility 1: Business logic
    public void calculateTotal() {
        double total = amount + (amount * taxRate / 100);
        System.out.println("  Total: $" + total);
    }

    // Responsibility 2: Presentation
    public void printInvoice() {
        System.out.println("  Printing invoice for " + customer + "...");
    }

    // Responsibility 3: Persistence
    public void saveToDatabase() {
        System.out.println("  Saving to database...");
    }
    // 3 responsibilities = 3 reasons to change = BAD
}

// ═══════════════════════════════════════════════════════════════
// GOOD: Each class has exactly ONE responsibility
// ═══════════════════════════════════════════════════════════════

// Responsibility: Hold invoice data (Data Model)
class Invoice {
    private String customer;
    private List<LineItem> items = new ArrayList<>();

    public Invoice(String customer, double baseAmount) {
        this.customer = customer;
    }

    public void addItem(String name, int quantity, double unitPrice) {
        items.add(new LineItem(name, quantity, unitPrice));
    }

    public String getCustomer() { return customer; }
    public List<LineItem> getItems() { return items; }

    public double getSubtotal() {
        return items.stream().mapToDouble(LineItem::getTotal).sum();
    }
}

class LineItem {
    private String name;
    private int quantity;
    private double unitPrice;

    public LineItem(String name, int quantity, double unitPrice) {
        this.name = name;
        this.quantity = quantity;
        this.unitPrice = unitPrice;
    }

    public String getName() { return name; }
    public int getQuantity() { return quantity; }
    public double getUnitPrice() { return unitPrice; }
    public double getTotal() { return quantity * unitPrice; }
}

// Responsibility: Calculate tax (Business Logic)
class TaxCalculator {
    private double taxRate;

    public TaxCalculator(double taxRate) {
        this.taxRate = taxRate;
    }

    public double calculateTotal(Invoice invoice) {
        double subtotal = invoice.getSubtotal();
        return subtotal + (subtotal * taxRate / 100);
    }
}

// Responsibility: Print invoices (Presentation)
class InvoicePrinter {
    public void print(Invoice invoice, double total) {
        System.out.println("  ┌─── INVOICE ───────────────────────┐");
        System.out.println("  │ Customer: " + invoice.getCustomer());
        for (LineItem item : invoice.getItems()) {
            System.out.printf("  │ %-10s %d × $%.2f = $%.2f%n",
                    item.getName(), item.getQuantity(), item.getUnitPrice(), item.getTotal());
        }
        System.out.printf("  │ TOTAL: $%.2f%n", total);
        System.out.println("  └────────────────────────────────────┘");
    }
}

// Responsibility: Persist invoices (Data Access)
class InvoiceRepository {
    public void save(Invoice invoice) {
        System.out.println("  ✓ Invoice for " + invoice.getCustomer() + " saved to database.");
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Each class = ONE job = ONE reason to change.
 * ✦ If you change tax logic → only TaxCalculator changes.
 * ✦ If you change print format → only InvoicePrinter changes.
 * ✦ If you change DB schema → only InvoiceRepository changes.
 * ✦ Invoice (data model) is the most stable — rarely changes.
 *
 * COMPILE & RUN:
 *   javac SingleResponsibility.java && java SingleResponsibility
 */
