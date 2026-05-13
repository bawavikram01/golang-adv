/*
 * ============================================================
 *  CHAPTER 09: OOP — ENCAPSULATION & PACKAGES
 * ============================================================
 *
 *  ENCAPSULATION = Data Hiding + Bundling
 *  ──────────────────────────────────────
 *  - HIDE internal state (private fields)
 *  - EXPOSE controlled access (public methods)
 *  - VALIDATE data before setting
 *
 *  Benefits:
 *  1. Data Protection  — prevent invalid states
 *  2. Flexibility      — change internals without breaking external code
 *  3. Maintainability  — changes in one place
 *  4. Testability      — easy to test isolated units
 *
 *  PACKAGES = Directories for organizing classes
 *  ─────────────────────────────────────────────
 *  - Convention: reverse domain name (com.company.project)
 *  - Prevent naming conflicts
 *  - Control access (package-private)
 *
 *  ACCESS MODIFIERS SUMMARY:
 *  ┌──────────────┬────────┬─────────┬──────────┬───────────┐
 *  │ Modifier     │ Class  │ Package │ Subclass │ World     │
 *  ├──────────────┼────────┼─────────┼──────────┼───────────┤
 *  │ public       │   ✓    │    ✓    │    ✓     │    ✓      │
 *  │ protected    │   ✓    │    ✓    │    ✓     │    ✗      │
 *  │ (default)    │   ✓    │    ✓    │    ✗     │    ✗      │
 *  │ private      │   ✓    │    ✗    │    ✗     │    ✗      │
 *  └──────────────┴────────┴─────────┴──────────┴───────────┘
 *
 * ============================================================
 */

public class Chapter09_EncapsulationAndPackages {

    // =====================================================
    //  1. ENCAPSULATION — PROPER JAVA BEAN
    // =====================================================

    // A well-encapsulated class (Java Bean pattern)
    static class Employee {
        // ALL fields are PRIVATE — this is encapsulation!
        private String name;
        private int age;
        private double salary;
        private String department;

        // Constructor
        public Employee(String name, int age, double salary, String department) {
            setName(name);          // use setters for validation
            setAge(age);
            setSalary(salary);
            setDepartment(department);
        }

        // GETTERS — provide read access
        public String getName() { return name; }
        public int getAge() { return age; }
        public double getSalary() { return salary; }
        public String getDepartment() { return department; }

        // SETTERS — provide write access WITH VALIDATION
        public void setName(String name) {
            if (name == null || name.trim().isEmpty()) {
                throw new IllegalArgumentException("Name cannot be null or empty");
            }
            this.name = name.trim();
        }

        public void setAge(int age) {
            if (age < 18 || age > 100) {
                throw new IllegalArgumentException("Age must be between 18 and 100");
            }
            this.age = age;
        }

        public void setSalary(double salary) {
            if (salary < 0) {
                throw new IllegalArgumentException("Salary cannot be negative");
            }
            this.salary = salary;
        }

        public void setDepartment(String department) {
            this.department = department;
        }

        // BUSINESS METHODS
        public double getAnnualSalary() {
            return salary * 12;
        }

        public void giveRaise(double percentage) {
            if (percentage < 0 || percentage > 100) {
                throw new IllegalArgumentException("Raise percentage must be 0-100");
            }
            this.salary += this.salary * (percentage / 100);
        }

        @Override
        public String toString() {
            return String.format("Employee{name='%s', age=%d, salary=%.2f, dept='%s'}",
                    name, age, salary, department);
        }
    }

    // =====================================================
    //  2. IMMUTABLE CLASS
    // =====================================================

    // An immutable class — once created, cannot be changed
    // Rules for immutability:
    // 1. Class is final (can't be extended)
    // 2. All fields are private and final
    // 3. No setters
    // 4. Deep copy mutable fields in constructor and getter

    static final class ImmutablePerson {
        private final String name;
        private final int age;
        private final int[] scores; // mutable field — needs special handling

        public ImmutablePerson(String name, int age, int[] scores) {
            this.name = name;
            this.age = age;
            // DEEP COPY — don't store the original reference!
            this.scores = scores.clone();
        }

        public String getName() { return name; }
        public int getAge() { return age; }

        // Return a COPY — don't expose internal array!
        public int[] getScores() { return scores.clone(); }

        @Override
        public String toString() {
            StringBuilder sb = new StringBuilder();
            sb.append("ImmutablePerson{name='").append(name)
              .append("', age=").append(age).append(", scores=[");
            for (int i = 0; i < scores.length; i++) {
                if (i > 0) sb.append(", ");
                sb.append(scores[i]);
            }
            sb.append("]}");
            return sb.toString();
        }
    }

    // =====================================================
    //  3. BUILDER PATTERN (Advanced Encapsulation)
    // =====================================================

    static class Pizza {
        // Required
        private final String size;
        // Optional
        private final boolean cheese;
        private final boolean mushrooms;
        private final boolean peppers;
        private final boolean onions;

        // Private constructor — can only be created through Builder
        private Pizza(Builder builder) {
            this.size = builder.size;
            this.cheese = builder.cheese;
            this.mushrooms = builder.mushrooms;
            this.peppers = builder.peppers;
            this.onions = builder.onions;
        }

        @Override
        public String toString() {
            return "Pizza{size='" + size + "'" +
                    (cheese ? ", cheese" : "") +
                    (mushrooms ? ", mushrooms" : "") +
                    (peppers ? ", peppers" : "") +
                    (onions ? ", onions" : "") + "}";
        }

        // Static inner Builder class
        static class Builder {
            private final String size;     // required
            private boolean cheese;
            private boolean mushrooms;
            private boolean peppers;
            private boolean onions;

            public Builder(String size) {
                this.size = size;
            }

            public Builder cheese(boolean val) { cheese = val; return this; }
            public Builder mushrooms(boolean val) { mushrooms = val; return this; }
            public Builder peppers(boolean val) { peppers = val; return this; }
            public Builder onions(boolean val) { onions = val; return this; }

            public Pizza build() {
                return new Pizza(this);
            }
        }
    }

    // =====================================================
    //  4. PACKAGES EXPLAINED
    // =====================================================

    /*
     * PACKAGES — Organizing Classes
     * ──────────────────────────────
     *
     * In a real project, you'd organize files into packages:
     *
     *   src/
     *   └── com/
     *       └── myapp/
     *           ├── model/
     *           │   ├── User.java          → package com.myapp.model;
     *           │   └── Product.java       → package com.myapp.model;
     *           ├── service/
     *           │   └── UserService.java   → package com.myapp.service;
     *           └── util/
     *               └── StringUtils.java   → package com.myapp.util;
     *
     * PACKAGE DECLARATION (first line of file):
     *   package com.myapp.model;
     *
     * IMPORT STATEMENTS:
     *   import com.myapp.model.User;         // import specific class
     *   import com.myapp.model.*;             // import all classes in package
     *   import static java.lang.Math.PI;     // static import
     *   import static java.lang.Math.*;      // static import all
     *
     * BUILT-IN PACKAGES:
     *   java.lang    — auto-imported (String, Math, System, Object, etc.)
     *   java.util    — Collections, Scanner, Date, etc.
     *   java.io      — File I/O
     *   java.nio     — New I/O
     *   java.net     — Networking
     *   java.sql     — Database
     *   java.time    — Date/Time API
     *   java.math    — BigDecimal, BigInteger
     *
     * KEY RULES:
     *   - One public class per file
     *   - File name must match public class name
     *   - Package statement must be first (before imports)
     */

    // =====================================================
    //  5. STATIC IMPORTS
    // =====================================================

    // Instead of writing Math.PI, Math.sqrt, etc. every time:
    // import static java.lang.Math.*;
    // Then use: PI, sqrt(x), pow(x,y), etc.

    // =====================================================
    //  MAIN — Testing
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Encapsulation ---
        System.out.println("=== ENCAPSULATION ===\n");

        Employee emp = new Employee("Alice", 30, 75000, "Engineering");
        System.out.println(emp);
        System.out.println("Annual salary: $" + emp.getAnnualSalary());

        emp.giveRaise(10); // 10% raise
        System.out.println("After 10% raise: " + emp);

        // Validation in action:
        try {
            emp.setAge(5); // too young!
        } catch (IllegalArgumentException e) {
            System.out.println("Validation error: " + e.getMessage());
        }

        try {
            emp.setSalary(-1000);
        } catch (IllegalArgumentException e) {
            System.out.println("Validation error: " + e.getMessage());
        }

        // Without encapsulation, anyone could do:
        // emp.salary = -99999;  // IMPOSSIBLE! salary is private
        // emp.age = -5;         // IMPOSSIBLE! age is private

        // --- 2. Immutable Class ---
        System.out.println("\n=== IMMUTABLE CLASS ===\n");

        int[] originalScores = {90, 85, 78};
        ImmutablePerson ip = new ImmutablePerson("Bob", 25, originalScores);
        System.out.println(ip);

        // Try to modify from outside
        originalScores[0] = 0; // change original array
        System.out.println("After modifying original array: " + ip);
        System.out.println("Score[0] still: " + ip.getScores()[0]); // still 90!

        // Try to modify through getter
        int[] gottenScores = ip.getScores();
        gottenScores[0] = 0; // modify returned array
        System.out.println("After modifying gotten array: " + ip);
        System.out.println("Score[0] still: " + ip.getScores()[0]); // still 90!

        System.out.println("Object is truly immutable!");

        // --- 3. Builder Pattern ---
        System.out.println("\n=== BUILDER PATTERN ===\n");

        // Without builder: new Pizza("Large", true, true, false, false) — unclear!
        // With builder: readable and flexible!
        Pizza pizza1 = new Pizza.Builder("Large")
                .cheese(true)
                .mushrooms(true)
                .build();
        System.out.println(pizza1);

        Pizza pizza2 = new Pizza.Builder("Medium")
                .cheese(true)
                .peppers(true)
                .onions(true)
                .build();
        System.out.println(pizza2);

        Pizza pizza3 = new Pizza.Builder("Small").build(); // minimal pizza
        System.out.println(pizza3);

        // --- 4. Packages Demo ---
        System.out.println("\n=== PACKAGES ===\n");
        System.out.println("java.lang is auto-imported (String, Math, System, etc.)");
        System.out.println("Math.PI = " + Math.PI);
        System.out.println("Math.sqrt(144) = " + Math.sqrt(144));

        // Common imports you'll use:
        System.out.println("\nCommon imports:");
        System.out.println("  import java.util.*;        — Collections, Scanner");
        System.out.println("  import java.io.*;          — File I/O");
        System.out.println("  import java.util.stream.*; — Streams");
        System.out.println("  import java.time.*;        — Date/Time");

        // --- Encapsulation Summary ---
        System.out.println("\n=== ENCAPSULATION BEST PRACTICES ===\n");
        System.out.println("1. Make all fields private");
        System.out.println("2. Provide getters for read access");
        System.out.println("3. Provide setters with validation for write access");
        System.out.println("4. Make fields final if they shouldn't change");
        System.out.println("5. Return copies of mutable objects from getters");
        System.out.println("6. Use Builder pattern for complex object creation");
        System.out.println("7. Consider making classes immutable when possible");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Create a fully encapsulated `BankAccount` class with:
 *     - private fields: accountNumber, owner, balance
 *     - Validation: balance >= 0, owner not empty
 *     - Methods: deposit(), withdraw(), transfer(BankAccount, amount)
 *
 *  2. Create an immutable `Color` class with r, g, b values.
 *     Add a method `mix(Color other)` that returns a NEW Color.
 *
 *  3. Create a `StudentBuilder` using the Builder pattern with
 *     required fields (name, id) and optional fields (email, phone, gpa).
 *
 *  4. What's wrong with this code?
 *     class Broken {
 *         private List<String> items = new ArrayList<>();
 *         public List<String> getItems() { return items; }
 *     }
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 10 — Inheritance
 * ============================================================
 */
