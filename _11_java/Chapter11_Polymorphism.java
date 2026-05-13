/*
 * ============================================================
 *  CHAPTER 11: OOP — POLYMORPHISM
 * ============================================================
 *
 *  POLYMORPHISM = "Many Forms"
 *  One interface, multiple implementations.
 *
 *  TWO TYPES:
 *  ──────────
 *  1. COMPILE-TIME (Static) Polymorphism  → Method Overloading
 *     - Same method name, different parameters
 *     - Resolved at COMPILE time
 *
 *  2. RUNTIME (Dynamic) Polymorphism → Method Overriding
 *     - Parent reference, child object
 *     - Resolved at RUNTIME (dynamic dispatch)
 *     - The JVM decides which method to call based on actual object type
 *
 *  UPCASTING & DOWNCASTING:
 *  ────────────────────────
 *  Upcasting:   Child → Parent  (implicit, always safe)
 *  Downcasting: Parent → Child  (explicit, needs instanceof check)
 *
 * ============================================================
 */

public class Chapter11_Polymorphism {

    // =====================================================
    //  1. COMPILE-TIME POLYMORPHISM (Overloading)
    // =====================================================

    static class Calculator {
        // Same name, different parameter TYPES
        static int add(int a, int b) {
            return a + b;
        }

        static double add(double a, double b) {
            return a + b;
        }

        // Same name, different NUMBER of params
        static int add(int a, int b, int c) {
            return a + b + c;
        }

        // Same name, different PARAMETER ORDER
        static String add(String s, int n) {
            return s + n;
        }

        static String add(int n, String s) {
            return n + s;
        }
    }

    // =====================================================
    //  2. RUNTIME POLYMORPHISM (Overriding)
    // =====================================================

    static class Shape {
        String color;

        Shape(String color) {
            this.color = color;
        }

        double area() {
            return 0;
        }

        double perimeter() {
            return 0;
        }

        void draw() {
            System.out.println("Drawing a shape in " + color);
        }

        @Override
        public String toString() {
            return getClass().getSimpleName() + "{color=" + color + "}";
        }
    }

    static class CircleShape extends Shape {
        double radius;

        CircleShape(String color, double radius) {
            super(color);
            this.radius = radius;
        }

        @Override
        double area() {
            return Math.PI * radius * radius;
        }

        @Override
        double perimeter() {
            return 2 * Math.PI * radius;
        }

        @Override
        void draw() {
            System.out.println("Drawing a circle (r=" + radius + ") in " + color);
        }
    }

    static class RectangleShape extends Shape {
        double length, width;

        RectangleShape(String color, double length, double width) {
            super(color);
            this.length = length;
            this.width = width;
        }

        @Override
        double area() {
            return length * width;
        }

        @Override
        double perimeter() {
            return 2 * (length + width);
        }

        @Override
        void draw() {
            System.out.println("Drawing a rectangle (" + length + "x" + width + ") in " + color);
        }
    }

    static class TriangleShape extends Shape {
        double a, b, c; // sides

        TriangleShape(String color, double a, double b, double c) {
            super(color);
            this.a = a;
            this.b = b;
            this.c = c;
        }

        @Override
        double area() {
            double s = (a + b + c) / 2; // Heron's formula
            return Math.sqrt(s * (s - a) * (s - b) * (s - c));
        }

        @Override
        double perimeter() {
            return a + b + c;
        }

        @Override
        void draw() {
            System.out.println("Drawing a triangle (" + a + "," + b + "," + c + ") in " + color);
        }
    }

    // =====================================================
    //  3. TYPE CHECKING & CASTING
    // =====================================================

    static void describeAnimal(Shape shape) {
        // instanceof — checks runtime type
        if (shape instanceof CircleShape) {
            CircleShape c = (CircleShape) shape; // downcast
            System.out.println("It's a circle with radius " + c.radius);
        } else if (shape instanceof RectangleShape) {
            RectangleShape r = (RectangleShape) shape;
            System.out.println("It's a rectangle " + r.length + "x" + r.width);
        } else if (shape instanceof TriangleShape) {
            TriangleShape t = (TriangleShape) shape;
            System.out.println("It's a triangle with sides " + t.a + "," + t.b + "," + t.c);
        } else {
            System.out.println("Unknown shape");
        }
    }

    // =====================================================
    //  4. POLYMORPHIC METHOD (accepts parent type)
    // =====================================================

    static void printShapeInfo(Shape s) {
        // This method doesn't care WHAT shape it is!
        // It calls the overridden methods — JVM figures out which one
        s.draw();
        System.out.printf("  Area: %.2f, Perimeter: %.2f%n", s.area(), s.perimeter());
    }

    static double totalArea(Shape[] shapes) {
        double total = 0;
        for (Shape s : shapes) {
            total += s.area(); // calls the correct overridden area() for each shape
        }
        return total;
    }

    // =====================================================
    //  5. COVARIANT RETURN TYPES
    // =====================================================

    static class Animal {
        Animal create() {
            return new Animal();
        }

        @Override
        public String toString() { return "Animal"; }
    }

    static class DogAnimal extends Animal {
        @Override
        DogAnimal create() {  // return type is DogAnimal (subtype of Animal) — COVARIANT!
            return new DogAnimal();
        }

        @Override
        public String toString() { return "Dog"; }
    }

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Compile-Time Polymorphism ---
        System.out.println("=== COMPILE-TIME POLYMORPHISM ===\n");

        System.out.println("add(5, 3) = " + Calculator.add(5, 3));           // int version
        System.out.println("add(2.5, 3.7) = " + Calculator.add(2.5, 3.7));   // double version
        System.out.println("add(1, 2, 3) = " + Calculator.add(1, 2, 3));     // 3-param version
        System.out.println("add(\"Score: \", 100) = " + Calculator.add("Score: ", 100));
        System.out.println("add(100, \" points\") = " + Calculator.add(100, " points"));

        // --- 2. Runtime Polymorphism ---
        System.out.println("\n=== RUNTIME POLYMORPHISM ===\n");

        // Parent reference, child object — THIS IS POLYMORPHISM
        Shape s1 = new CircleShape("Red", 5);
        Shape s2 = new RectangleShape("Blue", 4, 6);
        Shape s3 = new TriangleShape("Green", 3, 4, 5);

        // Each calls its own overridden method!
        s1.draw(); // "Drawing a circle..."
        s2.draw(); // "Drawing a rectangle..."
        s3.draw(); // "Drawing a triangle..."

        System.out.println();

        // Polymorphic method call
        printShapeInfo(s1); // doesn't know it's a circle — just calls draw()/area()
        printShapeInfo(s2);
        printShapeInfo(s3);

        // Array of parent type, holding different child objects
        Shape[] shapes = {s1, s2, s3};
        System.out.printf("\nTotal area of all shapes: %.2f%n", totalArea(shapes));

        // --- 3. Upcasting and Downcasting ---
        System.out.println("\n=== UPCASTING & DOWNCASTING ===\n");

        // Upcasting: Child → Parent (implicit, always safe)
        CircleShape circle = new CircleShape("Yellow", 7);
        Shape upcast = circle; // automatic — CircleShape IS-A Shape
        upcast.draw();  // calls CircleShape's draw()
        // upcast.radius; // ERROR! Shape doesn't know about radius

        System.out.println();

        // Downcasting: Parent → Child (explicit, must check with instanceof)
        Shape unknownShape = new RectangleShape("Purple", 10, 5);

        // DANGEROUS without check:
        // CircleShape wrong = (CircleShape) unknownShape; // ClassCastException!

        // SAFE downcasting:
        if (unknownShape instanceof RectangleShape) {
            RectangleShape rect = (RectangleShape) unknownShape;
            System.out.println("Downcast successful! Length: " + rect.length);
        }

        if (unknownShape instanceof CircleShape) {
            System.out.println("This won't print — it's not a circle");
        } else {
            System.out.println("unknownShape is NOT a CircleShape");
        }

        // --- 4. instanceof checks ---
        System.out.println("\n=== INSTANCEOF ===\n");

        for (Shape s : shapes) {
            describeAnimal(s);
        }

        // instanceof hierarchy:
        System.out.println("\ncircle instanceof CircleShape: " + (circle instanceof CircleShape)); // true
        System.out.println("circle instanceof Shape: " + (circle instanceof Shape));                // true
        System.out.println("circle instanceof Object: " + (circle instanceof Object));              // true

        // --- 5. Polymorphism in action: Strategy-like pattern ---
        System.out.println("\n=== POLYMORPHISM POWER ===\n");

        // Create shapes dynamically and process them uniformly
        Shape[] manyShapes = {
            new CircleShape("Red", 3),
            new RectangleShape("Blue", 5, 3),
            new TriangleShape("Green", 5, 5, 5),
            new CircleShape("Yellow", 10),
            new RectangleShape("Orange", 8, 2)
        };

        System.out.println("Processing " + manyShapes.length + " shapes:");
        double grandTotal = 0;
        for (Shape s : manyShapes) {
            System.out.printf("  %-12s → Area: %8.2f%n",
                    s.getClass().getSimpleName(), s.area());
            grandTotal += s.area();
        }
        System.out.printf("  Grand Total Area: %.2f%n", grandTotal);

        // --- 6. Covariant Return Types ---
        System.out.println("\n=== COVARIANT RETURN ===\n");
        Animal animal = new Animal();
        DogAnimal dog = new DogAnimal();
        System.out.println("animal.create(): " + animal.create()); // "Animal"
        System.out.println("dog.create(): " + dog.create());       // "Dog"

        // The return type of dog.create() is DogAnimal (not just Animal)
        DogAnimal created = dog.create(); // works without cast!
        System.out.println("Covariant return works: " + created);

        // --- Summary ---
        System.out.println("\n=== KEY TAKEAWAYS ===\n");
        System.out.println("1. Overloading = compile-time polymorphism (same name, diff params)");
        System.out.println("2. Overriding = runtime polymorphism (parent ref, child obj)");
        System.out.println("3. JVM uses dynamic dispatch to call the right method at runtime");
        System.out.println("4. Upcasting is implicit and safe");
        System.out.println("5. Downcasting needs instanceof check");
        System.out.println("6. Polymorphism lets you write flexible, extensible code");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Create a `Payment` hierarchy: Payment → CreditCard, PayPal, Cash.
 *     Each has a processPayment(double amount) method.
 *     Write a method that processes an array of Payment objects.
 *
 *  2. Create a `Drawable` method that takes Shape[] and calls draw()
 *     on each. Add a new `Pentagon` shape and verify it works
 *     WITHOUT changing the Drawable method.
 *
 *  3. What's the output?
 *     class A { void show() { System.out.println("A"); } }
 *     class B extends A { void show() { System.out.println("B"); } }
 *     A obj = new B();
 *     obj.show(); // ???
 *
 *  4. Explain why this fails:
 *     Shape s = new CircleShape("Red", 5);
 *     s.getRadius(); // ???
 *     How do you fix it?
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 12 — Abstraction
 * ============================================================
 */
