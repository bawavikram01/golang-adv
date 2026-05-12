/*
 * ============================================================
 *  CHAPTER 08: OOP — CLASSES & OBJECTS
 * ============================================================
 *
 *  OBJECT-ORIENTED PROGRAMMING (OOP) is a paradigm based on
 *  the concept of "objects" that contain DATA (fields) and
 *  CODE (methods).
 *
 *  FOUR PILLARS OF OOP:
 *  1. Encapsulation   — bundling data + methods, hiding internals
 *  2. Inheritance     — child class inherits from parent
 *  3. Polymorphism    — one interface, many implementations
 *  4. Abstraction     — hiding complexity, showing essentials
 *
 *  CLASS vs OBJECT:
 *  ─────────────────
 *  CLASS  = Blueprint / Template  (like a house plan)
 *  OBJECT = Instance of a class   (like an actual house)
 *
 *  You can create MANY objects from ONE class.
 *
 *  MEMORY:
 *  - Class definition → loaded into Method Area
 *  - Object → created in HEAP memory
 *  - Reference variable → stored in STACK
 *
 * ============================================================
 */

public class Chapter08_ClassesAndObjects {

    // =====================================================
    //  1. DEFINING A CLASS
    // =====================================================

    // A simple class with fields and methods
    static class Dog {
        // FIELDS (attributes / instance variables)
        String name;
        String breed;
        int age;
        double weight;

        // METHOD (behavior)
        void bark() {
            System.out.println(name + " says: Woof! Woof!");
        }

        void displayInfo() {
            System.out.println("Name: " + name + ", Breed: " + breed
                    + ", Age: " + age + ", Weight: " + weight + "kg");
        }
    }

    // =====================================================
    //  2. CONSTRUCTORS
    // =====================================================

    static class Student {
        String name;
        int age;
        double gpa;

        // DEFAULT CONSTRUCTOR
        // If you don't write ANY constructor, Java provides one:
        // Student() { } — initializes fields to defaults

        // NO-ARG CONSTRUCTOR (explicit)
        Student() {
            this.name = "Unknown";
            this.age = 0;
            this.gpa = 0.0;
            System.out.println("No-arg constructor called");
        }

        // PARAMETERIZED CONSTRUCTOR
        Student(String name, int age, double gpa) {
            this.name = name;   // 'this' refers to current object
            this.age = age;     // distinguishes field from parameter
            this.gpa = gpa;
            System.out.println("Parameterized constructor called for: " + name);
        }

        // CONSTRUCTOR WITH PARTIAL PARAMS
        Student(String name) {
            this(name, 18, 0.0); // CONSTRUCTOR CHAINING with 'this()'
            // this() must be the FIRST statement in the constructor
            System.out.println("Single-param constructor (chained)");
        }

        // COPY CONSTRUCTOR (creates object from another object)
        Student(Student other) {
            this.name = other.name;
            this.age = other.age;
            this.gpa = other.gpa;
            System.out.println("Copy constructor called");
        }

        void display() {
            System.out.println("Student[name=" + name + ", age=" + age + ", gpa=" + gpa + "]");
        }
    }

    // =====================================================
    //  3. THE 'this' KEYWORD
    // =====================================================

    static class Rectangle {
        double length;
        double width;

        Rectangle(double length, double width) {
            // 'this' refers to the CURRENT object
            this.length = length; // this.length = field, length = parameter
            this.width = width;
        }

        double area() {
            return this.length * this.width; // 'this' optional here but clear
        }

        // Method chaining — return 'this'
        Rectangle setLength(double length) {
            this.length = length;
            return this; // enables chaining
        }

        Rectangle setWidth(double width) {
            this.width = width;
            return this;
        }

        void display() {
            System.out.println("Rectangle[" + length + " x " + width
                    + ", area=" + area() + "]");
        }
    }

    // =====================================================
    //  4. ACCESS MODIFIERS ON MEMBERS
    // =====================================================

    static class BankAccount {
        // Access Modifiers:
        // public    → accessible from ANYWHERE
        // private   → accessible ONLY within this class
        // protected → accessible within package + subclasses
        // (default) → accessible within same package only (no keyword)

        private String owner;        // only this class can access
        private double balance;      // only this class can access
        public String accountType;   // anyone can access
        int accountNumber;           // package-private (default)

        public BankAccount(String owner, double balance) {
            this.owner = owner;
            this.balance = balance;
            this.accountType = "Savings";
        }

        // Public methods to access private fields (getters/setters)
        public String getOwner() {
            return owner;
        }

        public double getBalance() {
            return balance;
        }

        public void deposit(double amount) {
            if (amount > 0) {
                balance += amount;
                System.out.println("Deposited: $" + amount + " → Balance: $" + balance);
            } else {
                System.out.println("Invalid deposit amount");
            }
        }

        public void withdraw(double amount) {
            if (amount > 0 && amount <= balance) {
                balance -= amount;
                System.out.println("Withdrawn: $" + amount + " → Balance: $" + balance);
            } else {
                System.out.println("Insufficient funds or invalid amount");
            }
        }
    }

    // =====================================================
    //  5. STATIC MEMBERS
    // =====================================================

    static class Counter {
        // INSTANCE variable: each object has its own copy
        int instanceCount;

        // STATIC variable: shared by ALL objects of the class
        static int totalCount = 0;

        Counter() {
            instanceCount = 1;
            totalCount++;  // incremented for every new object
        }

        // STATIC method: belongs to the CLASS, not an object
        // Cannot access instance variables or 'this'
        static int getTotalCount() {
            // return instanceCount; // ERROR! Can't access instance from static
            return totalCount;
        }

        // Instance method: belongs to the OBJECT
        // Can access both static and instance variables
        void display() {
            System.out.println("Instance: " + instanceCount + ", Total: " + totalCount);
        }
    }

    // =====================================================
    //  6. STATIC BLOCK & INSTANCE INITIALIZER
    // =====================================================

    static class InitDemo {
        static int staticVar;
        int instanceVar;

        // STATIC BLOCK — runs ONCE when class is first loaded
        static {
            staticVar = 100;
            System.out.println("  Static block executed. staticVar = " + staticVar);
        }

        // INSTANCE INITIALIZER BLOCK — runs BEFORE each constructor
        {
            instanceVar = 50;
            System.out.println("  Instance initializer block executed. instanceVar = " + instanceVar);
        }

        InitDemo() {
            System.out.println("  Constructor called.");
        }

        // ORDER OF EXECUTION:
        // 1. Static block (once, when class loads)
        // 2. Instance initializer block (each time object is created)
        // 3. Constructor (each time object is created)
    }

    // =====================================================
    //  7. FINAL KEYWORD
    // =====================================================

    static class FinalDemo {
        // final variable: value cannot change after assignment
        final int MAX_SPEED = 200;

        // final blank variable: must be initialized in constructor
        final String color;

        // final static variable: constant
        static final double PI = 3.14159;

        FinalDemo(String color) {
            this.color = color; // blank final must be initialized here
        }

        // final method: cannot be overridden by subclasses
        final void display() {
            System.out.println("Color: " + color + ", Max Speed: " + MAX_SPEED);
        }
    }
    // final class: cannot be extended/inherited
    // final class Immutable { }  — no subclass possible

    // =====================================================
    //  8. OBJECT CLASS METHODS (toString, equals, hashCode)
    // =====================================================

    static class Person {
        String name;
        int age;

        Person(String name, int age) {
            this.name = name;
            this.age = age;
        }

        // Override toString() for readable output
        @Override
        public String toString() {
            return "Person{name='" + name + "', age=" + age + "}";
        }

        // Override equals() for content comparison
        @Override
        public boolean equals(Object obj) {
            if (this == obj) return true;                 // same reference
            if (obj == null || getClass() != obj.getClass()) return false;
            Person other = (Person) obj;
            return age == other.age && name.equals(other.name);
        }

        // Override hashCode() — MUST override if you override equals
        @Override
        public int hashCode() {
            return name.hashCode() * 31 + age;
        }
    }

    // =====================================================
    //  MAIN — Testing Everything
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Creating Objects ---
        System.out.println("=== CREATING OBJECTS ===\n");

        Dog dog1 = new Dog();
        dog1.name = "Buddy";
        dog1.breed = "Golden Retriever";
        dog1.age = 3;
        dog1.weight = 30.5;

        Dog dog2 = new Dog();
        dog2.name = "Max";
        dog2.breed = "German Shepherd";
        dog2.age = 5;
        dog2.weight = 35.0;

        dog1.displayInfo();
        dog1.bark();
        dog2.displayInfo();
        dog2.bark();

        // --- 2. Constructors ---
        System.out.println("\n=== CONSTRUCTORS ===\n");

        Student s1 = new Student(); // no-arg
        s1.display();

        Student s2 = new Student("Alice", 20, 3.8); // parameterized
        s2.display();

        Student s3 = new Student("Bob"); // single param (chained)
        s3.display();

        Student s4 = new Student(s2); // copy constructor
        s4.display();

        // --- 3. this keyword & Method Chaining ---
        System.out.println("\n=== METHOD CHAINING ===\n");

        Rectangle rect = new Rectangle(5, 3);
        rect.display();

        // Chaining: each setter returns 'this', so we can chain calls
        rect.setLength(10).setWidth(7).display();

        // --- 4. Access Modifiers ---
        System.out.println("\n=== ACCESS MODIFIERS ===\n");

        BankAccount account = new BankAccount("Vikram", 1000);
        // account.balance = 9999; // ERROR: balance is private!
        System.out.println("Owner: " + account.getOwner());      // through getter
        System.out.println("Balance: $" + account.getBalance());  // through getter
        account.deposit(500);
        account.withdraw(200);
        account.withdraw(5000); // insufficient funds

        // --- 5. Static Members ---
        System.out.println("\n=== STATIC MEMBERS ===\n");

        System.out.println("Total before creating: " + Counter.getTotalCount()); // 0

        Counter c1 = new Counter();
        Counter c2 = new Counter();
        Counter c3 = new Counter();

        c1.display();
        System.out.println("Total created: " + Counter.getTotalCount()); // 3

        // Static can be accessed with class name (preferred)
        // or with object reference (not recommended)
        System.out.println("Counter.getTotalCount() = " + Counter.getTotalCount());

        // --- 6. Initialization Order ---
        System.out.println("\n=== INITIALIZATION ORDER ===\n");
        System.out.println("Creating first InitDemo:");
        InitDemo id1 = new InitDemo();
        System.out.println("\nCreating second InitDemo:");
        InitDemo id2 = new InitDemo();
        System.out.println("Notice: static block runs only ONCE");

        // --- 7. Final Keyword ---
        System.out.println("\n=== FINAL KEYWORD ===\n");

        FinalDemo fd = new FinalDemo("Red");
        fd.display();
        // fd.MAX_SPEED = 300; // ERROR: cannot assign to final
        // FinalDemo.PI = 3.0; // ERROR: cannot assign to final static
        System.out.println("PI = " + FinalDemo.PI);

        // --- 8. toString, equals, hashCode ---
        System.out.println("\n=== OBJECT METHODS ===\n");

        Person p1 = new Person("Alice", 25);
        Person p2 = new Person("Alice", 25);
        Person p3 = new Person("Bob", 30);

        // toString
        System.out.println("p1: " + p1);  // calls toString()
        System.out.println("p3: " + p3);

        // equals
        System.out.println("p1.equals(p2): " + p1.equals(p2)); // true (same content)
        System.out.println("p1.equals(p3): " + p1.equals(p3)); // false
        System.out.println("p1 == p2: " + (p1 == p2));          // false (different objects!)

        // hashCode
        System.out.println("p1.hashCode(): " + p1.hashCode());
        System.out.println("p2.hashCode(): " + p2.hashCode()); // same as p1 (same content)

        // --- 9. Null Safety ---
        System.out.println("\n=== NULL SAFETY ===\n");
        Dog nullDog = null;
        // nullDog.bark(); // NullPointerException!
        System.out.println("nullDog is null: " + (nullDog == null));

        // Always check for null before using
        if (nullDog != null) {
            nullDog.bark();
        } else {
            System.out.println("Cannot bark — dog is null!");
        }
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Create a `Car` class with fields: make, model, year, speed.
 *     Add methods: accelerate(int amount), brake(int amount), display().
 *     Test with multiple car objects.
 *
 *  2. Create a `Point` class with x, y coordinates.
 *     Add: distance(Point other), toString(), equals().
 *
 *  3. Create a `Counter` class that tracks how many objects
 *     have been created using a static variable.
 *
 *  4. Create a `Book` class with a copy constructor.
 *     Modify the copy and verify the original is unchanged.
 *
 *  5. Create a `Matrix` class that wraps a 2D array.
 *     Add methods: add(Matrix), multiply(Matrix), transpose(), display().
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 09 — Encapsulation & Packages
 * ============================================================
 */
