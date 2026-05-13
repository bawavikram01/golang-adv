/*
 * ============================================================
 *  CHAPTER 10: OOP — INHERITANCE
 * ============================================================
 *
 *  INHERITANCE = Creating a new class from an existing class.
 *
 *  - Parent class (superclass / base class) — the original
 *  - Child class (subclass / derived class)  — extends the parent
 *
 *  The child INHERITS all non-private members of the parent.
 *
 *  WHY?
 *  1. Code Reuse      — don't repeat common code
 *  2. Hierarchies     — model "is-a" relationships
 *  3. Polymorphism    — treat child as parent type
 *
 *  TYPES OF INHERITANCE IN JAVA:
 *  ─────────────────────────────
 *  1. Single:      A → B
 *  2. Multilevel:  A → B → C
 *  3. Hierarchical: A → B, A → C
 *  4. Multiple:    NOT supported with classes (use interfaces)
 *
 *  KEYWORD: extends
 *  class Child extends Parent { }
 *
 * ============================================================
 */

public class Chapter10_Inheritance {

    // =====================================================
    //  1. BASIC INHERITANCE
    // =====================================================

    // Parent (Super) class
    static class Animal {
        String name;
        int age;

        Animal(String name, int age) {
            this.name = name;
            this.age = age;
            System.out.println("  Animal constructor called");
        }

        void eat() {
            System.out.println(name + " is eating.");
        }

        void sleep() {
            System.out.println(name + " is sleeping.");
        }

        @Override
        public String toString() {
            return "Animal{name='" + name + "', age=" + age + "}";
        }
    }

    // Child (Sub) class — INHERITS from Animal
    static class Dog extends Animal {
        String breed;

        Dog(String name, int age, String breed) {
            super(name, age); // MUST call parent constructor first
            this.breed = breed;
            System.out.println("  Dog constructor called");
        }

        // Dog-specific method
        void bark() {
            System.out.println(name + " says: Woof!");
        }

        // METHOD OVERRIDING — redefine parent's method
        @Override
        void eat() {
            System.out.println(name + " the dog is eating kibble.");
        }

        @Override
        public String toString() {
            return "Dog{name='" + name + "', age=" + age + ", breed='" + breed + "'}";
        }
    }

    // Another child class
    static class Cat extends Animal {
        boolean isIndoor;

        Cat(String name, int age, boolean isIndoor) {
            super(name, age);
            this.isIndoor = isIndoor;
        }

        void meow() {
            System.out.println(name + " says: Meow!");
        }

        @Override
        void eat() {
            System.out.println(name + " the cat is eating fish.");
        }
    }

    // =====================================================
    //  2. SUPER KEYWORD
    // =====================================================

    static class Vehicle {
        String brand;
        int speed;

        Vehicle(String brand) {
            this.brand = brand;
            this.speed = 0;
        }

        void accelerate(int amount) {
            speed += amount;
            System.out.println(brand + " accelerating to " + speed + " km/h");
        }

        void displayInfo() {
            System.out.println("Vehicle: " + brand + ", Speed: " + speed);
        }
    }

    static class Car extends Vehicle {
        int numDoors;

        Car(String brand, int numDoors) {
            super(brand); // calls Vehicle(String brand)
            this.numDoors = numDoors;
        }

        @Override
        void accelerate(int amount) {
            // Call parent's version first
            super.accelerate(amount); // Vehicle's accelerate
            if (speed > 200) {
                System.out.println("WARNING: Over speed limit!");
            }
        }

        @Override
        void displayInfo() {
            super.displayInfo(); // call parent's version
            System.out.println("Doors: " + numDoors); // add child-specific info
        }
    }

    // =====================================================
    //  3. CONSTRUCTOR CHAINING IN INHERITANCE
    // =====================================================

    /*
     * Constructor call order:
     * 1. JVM calls Object() constructor first
     * 2. Then parent constructor
     * 3. Then child constructor
     *
     * If you don't write super(), Java inserts super() (no-arg) automatically.
     * If parent has NO no-arg constructor, you MUST explicitly call super(args).
     */

    static class Grandparent {
        Grandparent() {
            System.out.println("    Grandparent constructor");
        }
    }

    static class Parent extends Grandparent {
        Parent() {
            super(); // auto-inserted if omitted
            System.out.println("    Parent constructor");
        }
    }

    static class Child extends Parent {
        Child() {
            super(); // auto-inserted if omitted
            System.out.println("    Child constructor");
        }
    }

    // =====================================================
    //  4. METHOD OVERRIDING RULES
    // =====================================================

    /*
     * Rules for @Override:
     * 1. Same method name and parameter list
     * 2. Return type must be same OR covariant (subtype)
     * 3. Access modifier can be same or WIDER (not narrower)
     *    parent: protected → child: protected or public (NOT private)
     * 4. Cannot override:
     *    - final methods
     *    - static methods (they are hidden, not overridden)
     *    - private methods (not inherited, so nothing to override)
     * 5. Can throw same, fewer, or narrower checked exceptions
     */

    static class Shape {
        protected double area() {
            return 0;
        }

        // Cannot be overridden
        final void printType() {
            System.out.println("I am a shape");
        }

        // Static methods are HIDDEN, not overridden
        static void staticMethod() {
            System.out.println("Shape static method");
        }
    }

    static class Circle extends Shape {
        double radius;

        Circle(double radius) {
            this.radius = radius;
        }

        @Override
        public double area() { // protected → public is OK (wider access)
            return Math.PI * radius * radius;
        }

        // This HIDES Shape.staticMethod(), not overrides it
        static void staticMethod() {
            System.out.println("Circle static method");
        }
    }

    // =====================================================
    //  5. MULTILEVEL INHERITANCE
    // =====================================================

    static class LivingBeing {
        void breathe() { System.out.println("Breathing..."); }
    }

    static class Human extends LivingBeing {
        void speak() { System.out.println("Speaking..."); }
    }

    static class Student extends Human {
        void study() { System.out.println("Studying..."); }
    }
    // Student inherits: breathe() from LivingBeing, speak() from Human

    // =====================================================
    //  6. THE OBJECT CLASS (Root of all classes)
    // =====================================================

    /*
     * Every class in Java implicitly extends java.lang.Object
     *
     * Object class provides:
     * - toString()    → string representation
     * - equals()      → content equality
     * - hashCode()    → hash code for collections
     * - getClass()    → runtime class information
     * - clone()       → create a copy (requires Cloneable)
     * - finalize()    → called before garbage collection (deprecated)
     * - wait/notify() → thread synchronization
     */

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Basic Inheritance ---
        System.out.println("=== BASIC INHERITANCE ===\n");

        System.out.println("Creating a Dog:");
        Dog dog = new Dog("Buddy", 3, "Labrador");
        System.out.println(dog);
        dog.eat();    // overridden version
        dog.sleep();  // inherited from Animal
        dog.bark();   // Dog-specific

        System.out.println();

        Cat cat = new Cat("Whiskers", 5, true);
        cat.eat();    // overridden version
        cat.sleep();  // inherited
        cat.meow();   // Cat-specific

        // --- 2. super keyword ---
        System.out.println("\n=== SUPER KEYWORD ===\n");

        Car car = new Car("BMW", 4);
        car.accelerate(100);
        car.accelerate(120); // triggers warning
        car.displayInfo();

        // --- 3. Constructor Chaining ---
        System.out.println("\n=== CONSTRUCTOR CHAINING ===\n");
        System.out.println("Creating Child object:");
        Child child = new Child();
        // Notice order: Grandparent → Parent → Child

        // --- 4. Overriding ---
        System.out.println("\n=== METHOD OVERRIDING ===\n");

        Circle circle = new Circle(5);
        System.out.println("Circle area: " + circle.area());
        circle.printType(); // inherited, not overridden (final)

        // Static method hiding
        Shape.staticMethod();   // "Shape static method"
        Circle.staticMethod();  // "Circle static method"

        // --- 5. Multilevel Inheritance ---
        System.out.println("\n=== MULTILEVEL INHERITANCE ===\n");
        Student student = new Student();
        student.breathe(); // from LivingBeing
        student.speak();   // from Human
        student.study();   // from Student

        // --- 6. Object class methods ---
        System.out.println("\n=== OBJECT CLASS ===\n");
        System.out.println("dog.getClass(): " + dog.getClass());
        System.out.println("dog.getClass().getSimpleName(): " + dog.getClass().getSimpleName());
        System.out.println("dog instanceof Dog: " + (dog instanceof Dog));
        System.out.println("dog instanceof Animal: " + (dog instanceof Animal));
        System.out.println("dog instanceof Object: " + (dog instanceof Object));

        // --- IS-A relationship ---
        System.out.println("\n=== IS-A RELATIONSHIP ===\n");
        System.out.println("Dog IS-A Animal: " + (dog instanceof Animal));   // true
        System.out.println("Cat IS-A Animal: " + (cat instanceof Animal));   // true
        System.out.println("Dog IS-A Object: " + (dog instanceof Object));   // true
        // "Animal IS-A Dog" → false (parent is NOT a child type)

        // A child class reference can be assigned to parent type
        Animal animalRef = dog; // upcasting (safe, implicit)
        animalRef.eat();  // calls Dog's overridden eat() — dynamic dispatch!
        // animalRef.bark(); // ERROR: Animal doesn't know about bark()
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Create a hierarchy: Shape → Rectangle → Square.
 *     Override area() and perimeter() in each.
 *
 *  2. Create: Vehicle → Car, Vehicle → Motorcycle.
 *     Add relevant fields and methods. Test inheritance.
 *
 *  3. Create a 3-level hierarchy:
 *     Employee → Manager → Director.
 *     Each adds relevant fields and overrides a describe() method.
 *
 *  4. What's the output?
 *     class A { A() { System.out.println("A"); } }
 *     class B extends A { B() { System.out.println("B"); } }
 *     class C extends B { C() { System.out.println("C"); } }
 *     new C(); → ???
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 11 — Polymorphism
 * ============================================================
 */
