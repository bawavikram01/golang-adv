/*
 * ============================================================
 *  CHAPTER 15: EXCEPTION HANDLING
 * ============================================================
 *
 *  An EXCEPTION is an event that disrupts normal program flow.
 *
 *  EXCEPTION HIERARCHY:
 *  ────────────────────
 *  java.lang.Object
 *  └── java.lang.Throwable
 *      ├── java.lang.Error           (DON'T catch — JVM problems)
 *      │   ├── OutOfMemoryError
 *      │   ├── StackOverflowError
 *      │   └── VirtualMachineError
 *      └── java.lang.Exception       (DO catch — program problems)
 *          ├── IOException            ← CHECKED (must handle)
 *          ├── SQLException           ← CHECKED
 *          ├── FileNotFoundException  ← CHECKED
 *          └── RuntimeException       ← UNCHECKED (optional to handle)
 *              ├── NullPointerException
 *              ├── ArrayIndexOutOfBoundsException
 *              ├── ArithmeticException
 *              ├── NumberFormatException
 *              ├── ClassCastException
 *              └── IllegalArgumentException
 *
 *  CHECKED vs UNCHECKED:
 *  ─────────────────────
 *  CHECKED:    Must be caught or declared (throws). Compiler enforces.
 *              Examples: IOException, SQLException
 *
 *  UNCHECKED:  RuntimeException subclasses. Compiler doesn't enforce.
 *              Examples: NullPointerException, ArithmeticException
 *
 * ============================================================
 */

import java.io.*;

public class Chapter15_ExceptionHandling {

    // =====================================================
    //  1. TRY-CATCH
    // =====================================================

    static void basicTryCatch() {
        System.out.println("--- Basic try-catch ---");

        try {
            int result = 10 / 0; // ArithmeticException!
            System.out.println("This won't print: " + result);
        } catch (ArithmeticException e) {
            System.out.println("Caught: " + e.getMessage());
            // e.getMessage()      → "/ by zero"
            // e.toString()        → "java.lang.ArithmeticException: / by zero"
            // e.printStackTrace() → full stack trace (for debugging)
        }

        System.out.println("Program continues after exception!\n");
    }

    // =====================================================
    //  2. MULTIPLE CATCH BLOCKS
    // =====================================================

    static void multipleCatch() {
        System.out.println("--- Multiple catch ---");

        try {
            String str = null;
            // Try changing these to see different exceptions:
            int[] arr = {1, 2, 3};

            // Uncomment one at a time to test:
            // System.out.println(arr[5]);           // ArrayIndexOutOfBoundsException
            // System.out.println(str.length());     // NullPointerException
            System.out.println(Integer.parseInt("abc")); // NumberFormatException

        } catch (ArrayIndexOutOfBoundsException e) {
            System.out.println("Array index error: " + e.getMessage());
        } catch (NullPointerException e) {
            System.out.println("Null pointer error: " + e.getMessage());
        } catch (NumberFormatException e) {
            System.out.println("Number format error: " + e.getMessage());
        } catch (Exception e) {
            // Catch-all — must be LAST (most general)
            System.out.println("General error: " + e.getMessage());
        }
        System.out.println();
    }

    // =====================================================
    //  3. MULTI-CATCH (Java 7+)
    // =====================================================

    static void multiCatch() {
        System.out.println("--- Multi-catch (Java 7+) ---");

        try {
            // Can catch multiple exceptions in one block
            String value = "abc";
            int num = Integer.parseInt(value);
        } catch (NumberFormatException | ArithmeticException e) {
            // One handler for multiple exception types
            System.out.println("Caught: " + e.getClass().getSimpleName() + ": " + e.getMessage());
        }
        System.out.println();
    }

    // =====================================================
    //  4. TRY-CATCH-FINALLY
    // =====================================================

    static void tryCatchFinally() {
        System.out.println("--- try-catch-finally ---");

        try {
            System.out.println("Try block — opening resource");
            int result = 10 / 2;
            System.out.println("Result: " + result);
        } catch (ArithmeticException e) {
            System.out.println("Catch block — handling error");
        } finally {
            // ALWAYS executes — whether exception occurred or not
            // Used to clean up resources (close files, connections, etc.)
            System.out.println("Finally block — ALWAYS runs (cleanup)");
        }
        System.out.println();

        // finally with exception
        try {
            System.out.println("Try with exception...");
            int x = 10 / 0;
        } catch (ArithmeticException e) {
            System.out.println("Caught exception");
        } finally {
            System.out.println("Finally runs even with exception");
        }
        System.out.println();

        // finally with return
        System.out.println("tryCatchFinallyReturn() = " + tryCatchFinallyReturn());
        System.out.println();
    }

    static int tryCatchFinallyReturn() {
        try {
            return 1;
        } finally {
            // finally runs EVEN with return!
            System.out.println("  Finally runs before return!");
            // If finally also has return, it overrides try's return (BAD practice!)
        }
    }

    // =====================================================
    //  5. THROW & THROWS
    // =====================================================

    // throw — explicitly throw an exception
    static void validateAge(int age) {
        if (age < 0) {
            throw new IllegalArgumentException("Age cannot be negative: " + age);
        }
        if (age < 18) {
            throw new IllegalArgumentException("Must be 18 or older. Got: " + age);
        }
        System.out.println("Age " + age + " is valid.");
    }

    // throws — declares that method MAY throw an exception
    // Caller MUST handle it (for checked exceptions)
    static void readFile(String filename) throws IOException {
        // This is a checked exception — must be declared
        if (!new java.io.File(filename).exists()) {
            throw new FileNotFoundException("File not found: " + filename);
        }
        System.out.println("Reading file: " + filename);
    }

    // =====================================================
    //  6. CUSTOM EXCEPTIONS
    // =====================================================

    // Custom checked exception (extends Exception)
    static class InsufficientFundsException extends Exception {
        private double amount;

        InsufficientFundsException(double amount) {
            super("Insufficient funds. Short by: $" + amount);
            this.amount = amount;
        }

        double getAmount() { return amount; }
    }

    // Custom unchecked exception (extends RuntimeException)
    static class InvalidEmailException extends RuntimeException {
        InvalidEmailException(String email) {
            super("Invalid email format: " + email);
        }
    }

    // Using custom exceptions
    static class BankAccount {
        private double balance;

        BankAccount(double balance) { this.balance = balance; }

        void withdraw(double amount) throws InsufficientFundsException {
            if (amount > balance) {
                throw new InsufficientFundsException(amount - balance);
            }
            balance -= amount;
            System.out.println("Withdrawn: $" + amount + ". Balance: $" + balance);
        }
    }

    static void validateEmail(String email) {
        if (email == null || !email.contains("@")) {
            throw new InvalidEmailException(email); // unchecked — no 'throws' needed
        }
        System.out.println("Valid email: " + email);
    }

    // =====================================================
    //  7. TRY-WITH-RESOURCES (Java 7+)
    // =====================================================

    // AutoCloseable resources are automatically closed
    // No need for finally block!

    static class MyResource implements AutoCloseable {
        String name;

        MyResource(String name) {
            this.name = name;
            System.out.println("  Opening resource: " + name);
        }

        void use() {
            System.out.println("  Using resource: " + name);
        }

        @Override
        public void close() {
            // Called automatically at end of try block
            System.out.println("  Closing resource: " + name);
        }
    }

    static void tryWithResources() {
        System.out.println("--- try-with-resources ---");

        // Resources are closed automatically in reverse order
        try (MyResource r1 = new MyResource("DB Connection");
             MyResource r2 = new MyResource("File Handle")) {
            r1.use();
            r2.use();
            // No finally needed! Resources closed automatically
        } catch (Exception e) {
            System.out.println("Error: " + e.getMessage());
        }
        System.out.println("Resources were auto-closed!\n");
    }

    // =====================================================
    //  8. EXCEPTION CHAINING
    // =====================================================

    static void lowLevelMethod() throws Exception {
        throw new Exception("Low-level database error");
    }

    static void highLevelMethod() throws Exception {
        try {
            lowLevelMethod();
        } catch (Exception e) {
            // Wrap low-level exception in a higher-level one
            throw new Exception("Service layer error", e); // 'e' is the cause
        }
    }

    // =====================================================
    //  9. COMMON EXCEPTIONS AND HOW TO AVOID THEM
    // =====================================================

    static void commonExceptions() {
        System.out.println("--- Common Exceptions ---\n");

        // NullPointerException — accessing null reference
        String str = null;
        // str.length(); // NPE!
        // FIX: Check for null
        if (str != null) {
            System.out.println(str.length());
        }

        // ArrayIndexOutOfBoundsException — invalid index
        int[] arr = {1, 2, 3};
        // arr[5]; // AIOOBE!
        // FIX: Check bounds
        int idx = 5;
        if (idx >= 0 && idx < arr.length) {
            System.out.println(arr[idx]);
        } else {
            System.out.println("Index " + idx + " out of bounds (0-" + (arr.length - 1) + ")");
        }

        // NumberFormatException — invalid number parsing
        String num = "12.3abc";
        // Integer.parseInt(num); // NFE!
        // FIX: Validate before parsing
        try {
            int parsed = Integer.parseInt(num);
        } catch (NumberFormatException e) {
            System.out.println("Cannot parse '" + num + "' as integer");
        }

        // ClassCastException — invalid type cast
        Object obj = "Hello";
        // Integer i = (Integer) obj; // CCE!
        // FIX: Use instanceof
        if (obj instanceof Integer) {
            Integer i = (Integer) obj;
        } else {
            System.out.println("Object is " + obj.getClass().getSimpleName() + ", not Integer");
        }

        // StackOverflowError — infinite recursion
        // void infinite() { infinite(); } // SOE!
        // FIX: Always have a base case in recursion
        System.out.println("Always have base cases in recursion!");
    }

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1-3. try-catch variants ---
        System.out.println("=== TRY-CATCH ===\n");
        basicTryCatch();
        multipleCatch();
        multiCatch();

        // --- 4. finally ---
        System.out.println("=== TRY-CATCH-FINALLY ===\n");
        tryCatchFinally();

        // --- 5. throw & throws ---
        System.out.println("=== THROW & THROWS ===\n");

        try {
            validateAge(25);
            validateAge(15);
        } catch (IllegalArgumentException e) {
            System.out.println("Caught: " + e.getMessage());
        }

        try {
            readFile("nonexistent.txt");
        } catch (IOException e) {
            System.out.println("Caught IOException: " + e.getMessage());
        }
        System.out.println();

        // --- 6. Custom Exceptions ---
        System.out.println("=== CUSTOM EXCEPTIONS ===\n");

        BankAccount account = new BankAccount(500);
        try {
            account.withdraw(200); // OK
            account.withdraw(400); // InsufficientFundsException!
        } catch (InsufficientFundsException e) {
            System.out.println("Caught: " + e.getMessage());
            System.out.println("Short by: $" + e.getAmount());
        }

        System.out.println();

        try {
            validateEmail("user@example.com"); // OK
            validateEmail("invalid-email");     // InvalidEmailException!
        } catch (InvalidEmailException e) {
            System.out.println("Caught: " + e.getMessage());
        }
        System.out.println();

        // --- 7. try-with-resources ---
        System.out.println("=== TRY-WITH-RESOURCES ===\n");
        tryWithResources();

        // --- 8. Exception Chaining ---
        System.out.println("=== EXCEPTION CHAINING ===\n");
        try {
            highLevelMethod();
        } catch (Exception e) {
            System.out.println("High-level: " + e.getMessage());
            System.out.println("Caused by:  " + e.getCause().getMessage());
        }
        System.out.println();

        // --- 9. Common Exceptions ---
        System.out.println("=== COMMON EXCEPTIONS ===\n");
        commonExceptions();

        // --- Best Practices ---
        System.out.println("\n=== BEST PRACTICES ===\n");
        System.out.println("1. Catch specific exceptions, not Exception");
        System.out.println("2. Don't catch and ignore (empty catch block)");
        System.out.println("3. Use try-with-resources for AutoCloseable");
        System.out.println("4. Throw early, catch late");
        System.out.println("5. Use custom exceptions for domain-specific errors");
        System.out.println("6. Don't use exceptions for flow control");
        System.out.println("7. Log the exception, don't swallow it");
        System.out.println("8. Prefer unchecked exceptions for programming errors");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Write a method that divides two numbers and handles:
 *     - Division by zero
 *     - Non-numeric input (if from String parsing)
 *
 *  2. Create a custom `InvalidAgeException` (checked).
 *     Use it in a `Person` class constructor.
 *
 *  3. Write a method that reads integers from a String array,
 *     catches invalid formats, and returns the sum of valid ones.
 *
 *  4. Create a `ResourceManager` class that implements
 *     AutoCloseable and test it with try-with-resources.
 *
 *  5. Create an exception chain: DatabaseException wraps
 *     SQLException wraps a connection error message.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 16 — Wrapper Classes & Autoboxing
 * ============================================================
 */
