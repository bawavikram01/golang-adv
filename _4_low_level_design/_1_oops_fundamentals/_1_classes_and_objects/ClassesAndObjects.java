/*
 * =============================================================
 * MODULE 1: CLASSES AND OBJECTS — The Foundation of Everything
 * =============================================================
 *
 * A CLASS is a blueprint. An OBJECT is a real thing built from that blueprint.
 *
 * Think of it like:
 *   - Class "Car" = the engineering drawing
 *   - Object myCar = a specific car parked in your garage
 *
 * KEY CONCEPTS:
 *   1. Fields (state)      — what the object KNOWS
 *   2. Methods (behavior)  — what the object DOES
 *   3. Constructor         — how the object is BORN
 *   4. `this` keyword      — "me, myself"
 *   5. Access modifiers    — who can see what
 */

public class ClassesAndObjects {

    public static void main(String[] args) {

        // ─── Creating objects ───
        BankAccount acc1 = new BankAccount("Alice", 1000.0);
        BankAccount acc2 = new BankAccount("Bob");  // uses overloaded constructor

        acc1.deposit(500);
        acc1.withdraw(200);
        System.out.println(acc1);  // toString() is called automatically

        acc2.deposit(3000);
        acc2.withdraw(5000);  // insufficient funds
        System.out.println(acc2);

        // ─── Static members belong to the CLASS, not any object ───
        System.out.println("Total accounts created: " + BankAccount.getAccountCount());

        // ─── Object identity vs equality ───
        BankAccount acc3 = acc1;          // same reference (alias)
        BankAccount acc4 = new BankAccount("Alice", 1300.0);

        System.out.println("acc3 == acc1 ? " + (acc3 == acc1));           // true  (same object)
        System.out.println("acc4 == acc1 ? " + (acc4 == acc1));           // false (different object)
        System.out.println("acc4.equals(acc1) ? " + acc4.equals(acc1));   // true  (same content)
    }
}

class BankAccount {

    // ─── Fields (state) ───
    private String owner;
    private double balance;
    private final String accountId;  // `final` → assigned once, never changed

    // ─── Static field — shared across ALL objects ───
    private static int accountCount = 0;

    // ─── Constructor (primary) ───
    public BankAccount(String owner, double initialBalance) {
        this.owner = owner;
        this.balance = initialBalance;
        this.accountId = "ACC-" + (++accountCount);
    }

    // ─── Constructor overloading — default balance ───
    public BankAccount(String owner) {
        this(owner, 0.0);  // delegates to the primary constructor
    }

    // ─── Behavior (methods) ───
    public void deposit(double amount) {
        if (amount <= 0) {
            System.out.println("Deposit amount must be positive.");
            return;
        }
        balance += amount;
        System.out.println(owner + " deposited " + amount + ". Balance: " + balance);
    }

    public void withdraw(double amount) {
        if (amount > balance) {
            System.out.println(owner + " → Insufficient funds! Balance: " + balance);
            return;
        }
        balance -= amount;
        System.out.println(owner + " withdrew " + amount + ". Balance: " + balance);
    }

    // ─── Getter (encapsulated access) ───
    public double getBalance() {
        return balance;
    }

    // ─── Static method — belongs to the class ───
    public static int getAccountCount() {
        return accountCount;
    }

    // ─── toString — human-readable representation ───
    @Override
    public String toString() {
        return "BankAccount{id='" + accountId + "', owner='" + owner + "', balance=" + balance + "}";
    }

    // ─── equals — content-based equality ───
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (obj == null || getClass() != obj.getClass()) return false;
        BankAccount other = (BankAccount) obj;
        return Double.compare(other.balance, balance) == 0
                && owner.equals(other.owner);
    }

    @Override
    public int hashCode() {
        return java.util.Objects.hash(owner, balance);
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ A class is a template; an object is an instance of that template.
 * ✦ `this` refers to the current object.
 * ✦ Constructors initialize objects. You can overload them.
 * ✦ `static` members belong to the class, not to any object.
 * ✦ `final` fields cannot be reassigned after construction.
 * ✦ Always override equals() + hashCode() together.
 * ✦ == checks reference; .equals() checks content.
 *
 * COMPILE & RUN:
 *   javac ClassesAndObjects.java && java ClassesAndObjects
 */
