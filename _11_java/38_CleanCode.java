/*
 * ============================================================
 *  CHAPTER 38: CLEAN CODE & BEST PRACTICES
 * ============================================================
 *  Based on "Clean Code" by Robert C. Martin and industry best
 *  practices. This chapter is about WRITING CODE LIKE A PRO.
 *
 *  KEY PRINCIPLES:
 *    1. Naming
 *    2. Methods
 *    3. Error Handling
 *    4. Code Organization
 *    5. Common Anti-Patterns
 *    6. Java-Specific Best Practices
 * ============================================================
 */

import java.util.*;
import java.util.stream.*;

public class Chapter38_CleanCode {

    // ========================================================
    // 1. NAMING — the most important thing in code
    // ========================================================

    // ❌ BAD names:
    //   int d;         // what is d?
    //   String s;      // meaningless
    //   int temp;      // vague
    //   void doStuff() // unclear purpose
    //   List<int[]> l; // cryptic

    // ✅ GOOD names:
    //   int elapsedDays;
    //   String customerName;
    //   int maxRetryCount;
    //   void calculateMonthlyRevenue()
    //   List<int[]> flaggedCells;

    // ❌ BAD
    static int calc(int a, int b, int t) {
        return t == 1 ? a + b : a * b;
    }

    // ✅ GOOD
    enum OperationType { ADD, MULTIPLY }

    static int calculate(int left, int right, OperationType operation) {
        switch (operation) {
            case ADD: return left + right;
            case MULTIPLY: return left * right;
            default: throw new IllegalArgumentException("Unknown: " + operation);
        }
    }

    // Naming conventions in Java:
    //   Classes:    PascalCase     → UserAccount, HttpRequest
    //   Methods:    camelCase      → calculateTotal(), getUserName()
    //   Variables:  camelCase      → itemCount, maxRetries
    //   Constants:  UPPER_SNAKE    → MAX_SIZE, DEFAULT_TIMEOUT
    //   Packages:   lowercase      → com.myapp.service
    //   Booleans:   is/has/can/should prefix → isActive, hasPermission

    // ========================================================
    // 2. METHODS — small, focused, one thing
    // ========================================================

    // ❌ BAD: Method does too many things
    static void processOrderBad(Map<String, Object> order) {
        // validate
        // calculate total
        // apply discount
        // charge payment
        // send email
        // update inventory
        // All in one method = unmaintainable!
    }

    // ✅ GOOD: Each method does ONE thing
    static class OrderProcessor {
        void processOrder(Order order) {
            validate(order);
            double total = calculateTotal(order);
            total = applyDiscount(total, order.discountCode);
            chargePayment(order.paymentMethod, total);
            sendConfirmation(order.email);
        }

        private void validate(Order order) {
            if (order.items.isEmpty()) throw new IllegalArgumentException("Empty order");
        }
        private double calculateTotal(Order order) {
            return order.items.stream().mapToDouble(Item::getPrice).sum();
        }
        private double applyDiscount(double total, String code) {
            if ("SAVE10".equals(code)) return total * 0.9;
            return total;
        }
        private void chargePayment(String method, double amount) {
            System.out.println("    Charged $" + amount + " via " + method);
        }
        private void sendConfirmation(String email) {
            System.out.println("    Confirmation sent to " + email);
        }
    }

    static class Item {
        String name;
        double price;
        Item(String name, double price) { this.name = name; this.price = price; }
        double getPrice() { return price; }
    }

    static class Order {
        List<Item> items;
        String discountCode;
        String paymentMethod;
        String email;

        Order(List<Item> items, String discount, String payment, String email) {
            this.items = items;
            this.discountCode = discount;
            this.paymentMethod = payment;
            this.email = email;
        }
    }

    // ========================================================
    // 3. AVOID ANTI-PATTERNS
    // ========================================================

    // ❌ Magic Numbers
    static double calculateAreaBad(double radius) {
        return 3.14159 * radius * radius;  // what is 3.14159?
    }

    // ✅ Named Constants
    static double calculateArea(double radius) {
        return Math.PI * radius * radius;
    }

    // ❌ Boolean Parameters (hard to read at call site)
    // processFile(file, true, false, true) — what do these mean?

    // ✅ Use enums or builder pattern
    enum FileMode { READ, WRITE }
    enum Compression { ENABLED, DISABLED }

    // ❌ Null Returns
    static List<String> getUsersBad(boolean active) {
        if (!active) return null;  // caller must remember to null-check!
        return List.of("Alice", "Bob");
    }

    // ✅ Return empty collections
    static List<String> getUsers(boolean active) {
        if (!active) return Collections.emptyList();
        return List.of("Alice", "Bob");
    }

    // ❌ Swallowing Exceptions
    static void readFileBad(String path) {
        try {
            // read file
        } catch (Exception e) {
            // empty catch = bug hiding!
        }
    }

    // ========================================================
    // 4. CODE SMELLS & FIXES
    // ========================================================

    // SMELL: Long parameter lists
    // ❌ BAD
    // void createUser(String name, int age, String email, String phone,
    //                 String address, String city, String country) {}

    // ✅ GOOD: Use a parameter object
    static class UserInfo {
        final String name;
        final int age;
        final String email;

        UserInfo(String name, int age, String email) {
            this.name = name;
            this.age = age;
            this.email = email;
        }
    }

    // SMELL: Deep nesting
    // ❌ BAD
    static String getStatusBad(int code) {
        if (code > 0) {
            if (code < 100) {
                if (code % 2 == 0) {
                    return "even-small";
                } else {
                    return "odd-small";
                }
            } else {
                return "big";
            }
        } else {
            return "invalid";
        }
    }

    // ✅ GOOD: Guard clauses (return early)
    static String getStatus(int code) {
        if (code <= 0) return "invalid";
        if (code >= 100) return "big";
        return code % 2 == 0 ? "even-small" : "odd-small";
    }

    // ========================================================
    // 5. JAVA-SPECIFIC BEST PRACTICES
    // ========================================================

    static void javaBestPractices() {

        // 1. Use StringBuilder for string concatenation in loops
        // ❌ String s = ""; for (...) s += item;
        // ✅ StringBuilder sb = new StringBuilder(); for (...) sb.append(item);

        // 2. Prefer List.of(), Map.of() for immutable collections
        List<String> immutable = List.of("a", "b", "c");

        // 3. Use Optional instead of null
        Optional<String> name = Optional.ofNullable(null);
        String value = name.orElse("default");

        // 4. Use try-with-resources for ALL closeable resources
        // try (var reader = new BufferedReader(...)) { ... }

        // 5. Prefer interfaces over concrete types
        // ❌ ArrayList<String> list = new ArrayList<>();
        // ✅ List<String> list = new ArrayList<>();

        // 6. Use enhanced for-loop or streams over indexed loops
        List<String> items = List.of("a", "b", "c");
        // ❌ for (int i = 0; i < items.size(); i++) { ... }
        // ✅ for (String item : items) { ... }
        // ✅ items.forEach(item -> ...);

        // 7. Use equals() properly
        // ❌ "hello" == str      (compares references)
        // ✅ "hello".equals(str) (compares values)
        // ✅ Objects.equals(a, b) (null-safe)

        // 8. Override equals AND hashCode together
        // If a.equals(b) then a.hashCode() == b.hashCode()

        // 9. Make classes immutable when possible
        // private final fields, no setters, defensive copies

        // 10. Use enums instead of int/String constants
        // ❌ int STATUS_ACTIVE = 1;
        // ✅ enum Status { ACTIVE, INACTIVE }
    }

    // ========================================================
    // 6. CLEAN CODE EXAMPLE: Before & After
    // ========================================================

    // ❌ BEFORE (messy)
    static int f(int[] a) {
        int r = 0;
        for (int i = 0; i < a.length; i++) {
            if (a[i] > 0) {
                if (a[i] % 2 == 0) {
                    r = r + a[i];
                }
            }
        }
        return r;
    }

    // ✅ AFTER (clean)
    static int sumPositiveEvens(int[] numbers) {
        return Arrays.stream(numbers)
            .filter(n -> n > 0)
            .filter(n -> n % 2 == 0)
            .sum();
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== CLEAN CODE DEMONSTRATION ===\n");

        // Clean naming
        int result = calculate(10, 5, OperationType.ADD);
        System.out.println("  10 + 5 = " + result);

        // Clean methods
        System.out.println("\n--- Order Processing ---");
        OrderProcessor processor = new OrderProcessor();
        Order order = new Order(
            List.of(new Item("Book", 15.0), new Item("Pen", 2.50)),
            "SAVE10", "CreditCard", "user@test.com"
        );
        processor.processOrder(order);

        // Guard clauses
        System.out.println("\n--- Guard Clauses ---");
        System.out.println("  Status(50): " + getStatus(50));
        System.out.println("  Status(200): " + getStatus(200));
        System.out.println("  Status(-1): " + getStatus(-1));

        // Empty vs null
        System.out.println("\n--- Empty vs Null ---");
        System.out.println("  Active users: " + getUsers(true));
        System.out.println("  Inactive users: " + getUsers(false) + " (empty, not null!)");

        // Clean vs messy
        int[] numbers = {-3, 2, 5, 8, -1, 4, 7, 6};
        System.out.println("\n--- Clean vs Messy ---");
        System.out.println("  Messy f(): " + f(numbers));
        System.out.println("  Clean sumPositiveEvens(): " + sumPositiveEvens(numbers));

        // --- Print Rules ---
        System.out.println("\n=== CLEAN CODE RULES ===");
        System.out.println("  1. Names should reveal intent");
        System.out.println("  2. Methods should do ONE thing");
        System.out.println("  3. Methods should be SHORT (< 20 lines ideal)");
        System.out.println("  4. Max 2-3 parameters per method");
        System.out.println("  5. Don't return null — use Optional or empty collections");
        System.out.println("  6. Use guard clauses to reduce nesting");
        System.out.println("  7. No magic numbers — use named constants");
        System.out.println("  8. Don't repeat yourself (DRY)");
        System.out.println("  9. Keep it simple (KISS)");
        System.out.println("  10. You ain't gonna need it (YAGNI)");
        System.out.println("  11. Boy Scout Rule: leave code better than you found it");
        System.out.println("  12. Code is read 10x more than written — optimize for readability");

        System.out.println("\n✓ Clean Code & Best Practices Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Take your Chapter 04 code and refactor it applying clean code principles.
 * 2. Find and fix all magic numbers across your previous chapters.
 * 3. Rewrite a nested if-else chain using guard clauses and polymorphism.
 * 4. Create a code review checklist based on this chapter's principles.
 *
 * NEXT: Chapter 39 — Data Structures in Java
 */
