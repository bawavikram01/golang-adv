/*
 * =============================================================
 * KEY DESIGN PRINCIPLES BEYOND SOLID
 * =============================================================
 *
 * These principles complement SOLID and are essential for
 * writing clean, maintainable code. Know them cold.
 *
 * 1. DRY   — Don't Repeat Yourself
 * 2. KISS  — Keep It Simple, Stupid
 * 3. YAGNI — You Aren't Gonna Need It
 * 4. LoD   — Law of Demeter (Don't talk to strangers)
 * 5. Coupling & Cohesion — the yin and yang of design
 * 6. Composition over Inheritance
 * 7. Program to an Interface
 * 8. Encapsulate What Varies
 */

import java.util.*;

public class DesignPrinciples {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // 1. DRY — Don't Repeat Yourself
        // ═══════════════════════════════════════════════════════
        System.out.println("=== DRY PRINCIPLE ===");

        // BAD: Duplicated validation logic
        System.out.println("  BAD: Validation duplicated in every method");
        // createUser() { if (email == null || !email.contains("@")) throw... }
        // updateUser() { if (email == null || !email.contains("@")) throw... }
        // inviteUser() { if (email == null || !email.contains("@")) throw... }

        // GOOD: Extract to one place
        System.out.println("  GOOD: Validator.validateEmail(email) — ONE place");
        System.out.println("  If validation rules change → change ONE method\n");

        // ═══════════════════════════════════════════════════════
        // 2. KISS — Keep It Simple, Stupid
        // ═══════════════════════════════════════════════════════
        System.out.println("=== KISS PRINCIPLE ===");

        // BAD: Over-engineered
        System.out.println("  BAD: AbstractFactoryBeanProviderStrategyBuilderImpl");
        System.out.println("  GOOD: UserService with 3 methods\n");

        // BAD: Clever one-liner nobody understands
        int result = Arrays.asList(1,2,3,4,5).stream().reduce(0, (a,b) -> a + (b % 2 == 0 ? b * b : 0));
        System.out.println("  BAD (clever): " + result);

        // GOOD: Clear loop
        int sum = 0;
        for (int n : List.of(1,2,3,4,5)) {
            if (n % 2 == 0) sum += n * n;
        }
        System.out.println("  GOOD (clear): " + sum + "\n");

        // ═══════════════════════════════════════════════════════
        // 3. YAGNI — You Aren't Gonna Need It
        // ═══════════════════════════════════════════════════════
        System.out.println("=== YAGNI PRINCIPLE ===");
        System.out.println("  BAD: Building plugin system for app with 0 plugins");
        System.out.println("  BAD: Adding caching before you have performance problems");
        System.out.println("  BAD: Supporting 5 databases when you only use PostgreSQL");
        System.out.println("  GOOD: Build what you need NOW. Refactor when needed.\n");

        // ═══════════════════════════════════════════════════════
        // 4. LAW OF DEMETER — Don't talk to strangers
        // ═══════════════════════════════════════════════════════
        System.out.println("=== LAW OF DEMETER ===");

        BadCustomer badCustomer = new BadCustomer();
        // BAD: Chain of calls — reaching deep into object graph
        // badCustomer.getWallet().getCreditCard().getBank().getName()
        // This means BadCustomer must know about Wallet, CreditCard, and Bank!
        System.out.println("  BAD: customer.getWallet().getCreditCard().getBank().getName()");

        // GOOD: Ask the object directly — it delegates internally
        GoodCustomer goodCustomer = new GoodCustomer("Alice", 100.0);
        goodCustomer.pay(50.0);
        System.out.println("  GOOD: customer.pay(amount) — customer handles wallet internally\n");

        // ═══════════════════════════════════════════════════════
        // 5. COUPLING & COHESION
        // ═══════════════════════════════════════════════════════
        System.out.println("=== COUPLING & COHESION ===");
        System.out.println("  GOAL: LOW coupling + HIGH cohesion");
        System.out.println();
        System.out.println("  HIGH COHESION (GOOD):");
        System.out.println("    UserRepository → save, find, delete (all about user persistence)");
        System.out.println("    EmailService → send, validate, format (all about email)");
        System.out.println();
        System.out.println("  LOW COHESION (BAD):");
        System.out.println("    Utils → sendEmail, calculateTax, formatDate, parseJSON");
        System.out.println("    (unrelated methods dumped together)");
        System.out.println();
        System.out.println("  LOOSE COUPLING (GOOD):");
        System.out.println("    OrderService depends on PaymentGateway interface");
        System.out.println("    (can swap Stripe/PayPal without changing OrderService)");
        System.out.println();
        System.out.println("  TIGHT COUPLING (BAD):");
        System.out.println("    OrderService directly creates new StripePayment()");
        System.out.println("    (changing payment provider requires modifying OrderService)");

        // ═══════════════════════════════════════════════════════
        // 6. TELL, DON'T ASK
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== TELL, DON'T ASK ===");

        // BAD: Ask for data, then decide externally
        Account account = new Account(100);
        // if (account.getBalance() >= 50) { account.setBalance(account.getBalance() - 50); }
        System.out.println("  BAD: if (account.getBalance() >= amount) account.setBalance(...)");

        // GOOD: Tell the object what to do — it decides internally
        account.withdraw(50);
        System.out.println("  GOOD: account.withdraw(50) — account validates internally\n");

        // ═══════════════════════════════════════════════════════
        // Summary Table
        // ═══════════════════════════════════════════════════════
        System.out.println("=== PRINCIPLES SUMMARY ===");
        System.out.println("  ┌──────────────┬─────────────────────────────────────┐");
        System.out.println("  │ Principle    │ One-liner                           │");
        System.out.println("  ├──────────────┼─────────────────────────────────────┤");
        System.out.println("  │ DRY          │ Every piece of knowledge → ONE place│");
        System.out.println("  │ KISS         │ Simplest solution that works        │");
        System.out.println("  │ YAGNI        │ Don't build what you don't need yet │");
        System.out.println("  │ LoD          │ Only talk to immediate friends      │");
        System.out.println("  │ High Cohesion│ Class does ONE related set of things│");
        System.out.println("  │ Low Coupling │ Classes depend on interfaces, not   │");
        System.out.println("  │              │ concrete implementations            │");
        System.out.println("  │ Tell Don't   │ Tell objects what to do, don't ask  │");
        System.out.println("  │   Ask        │ for data and decide externally      │");
        System.out.println("  └──────────────┴─────────────────────────────────────┘");
    }
}

// ═══════════════════════════════════════════════════════════════
// LAW OF DEMETER EXAMPLES
// ═══════════════════════════════════════════════════════════════

// BAD: Forces callers to navigate the object graph
class BadCustomer {
    // Exposing internal structure → callers chain method calls
    // customer.getWallet().getCreditCard().charge(amount)
}

// GOOD: Encapsulate the delegation
class GoodCustomer {
    private String name;
    private double walletBalance;

    public GoodCustomer(String name, double balance) {
        this.name = name;
        this.walletBalance = balance;
    }

    // Caller just says "pay" — doesn't know about wallet internals
    public void pay(double amount) {
        if (walletBalance >= amount) {
            walletBalance -= amount;
            System.out.println("  " + name + " paid $" + amount + ". Remaining: $" + walletBalance);
        } else {
            System.out.println("  " + name + " → Insufficient funds!");
        }
    }
}

// ═══════════════════════════════════════════════════════════════
// TELL, DON'T ASK
// ═══════════════════════════════════════════════════════════════
class Account {
    private double balance;

    public Account(double balance) { this.balance = balance; }

    // GOOD: Object makes its own decision
    public boolean withdraw(double amount) {
        if (amount > balance) {
            System.out.println("  Insufficient funds!");
            return false;
        }
        balance -= amount;
        System.out.println("  Withdrew $" + amount + ". Balance: $" + balance);
        return true;
    }

    public double getBalance() { return balance; }
}

/*
 * PRINCIPLES CHEAT SHEET:
 * ─────────────────────────────────────────────────────────────
 * ✦ DRY: Duplicated code? Extract it.
 * ✦ KISS: Is this the simplest solution? If not, simplify.
 * ✦ YAGNI: Am I building this because I need it NOW?
 * ✦ LoD: Am I chaining more than one dot? (a.b().c().d() = bad)
 * ✦ Cohesion: Does every method in this class serve the same purpose?
 * ✦ Coupling: Would changing class X force me to change class Y?
 * ✦ Tell Don't Ask: Am I getting data just to make a decision?
 *
 * COMPILE & RUN:
 *   javac DesignPrinciples.java && java DesignPrinciples
 */
