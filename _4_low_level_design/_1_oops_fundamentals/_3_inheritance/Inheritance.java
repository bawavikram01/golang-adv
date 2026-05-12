/*
 * =============================================================
 * MODULE 3: INHERITANCE — "is-a" Relationship
 * =============================================================
 *
 * Inheritance lets a child class REUSE and EXTEND the parent class.
 *
 *   Animal (parent / superclass / base class)
 *     ├── Dog   (child / subclass / derived class)
 *     └── Cat
 *
 * WHY?
 *   1. Code reuse — don't repeat common behavior
 *   2. Hierarchical classification — mirrors real-world "is-a"
 *   3. Polymorphism foundation — treat children as parents
 *
 * RULES:
 *   - Java supports SINGLE inheritance (one parent only)
 *   - Use `extends` keyword
 *   - `super` calls parent's constructor or methods
 *   - `final` class → cannot be extended
 *   - `final` method → cannot be overridden
 */

public class Inheritance {

    public static void main(String[] args) {

        // ─── Creating parent and child objects ───
        Shape circle = new Circle("Red", 5.0);
        Shape rectangle = new Rectangle("Blue", 4.0, 6.0);
        Shape square = new Square("Green", 3.0);  // Square IS-A Rectangle IS-A Shape

        // ─── Each shape computes its own area and perimeter ───
        circle.displayInfo();
        System.out.println();
        rectangle.displayInfo();
        System.out.println();
        square.displayInfo();

        // ─── The power: treat all shapes uniformly ───
        System.out.println("\n=== Polymorphic Array ===");
        Shape[] shapes = { circle, rectangle, square };
        double totalArea = 0;
        for (Shape s : shapes) {
            totalArea += s.area();
        }
        System.out.println("Total area of all shapes: " + totalArea);

        // ─── instanceof check ───
        System.out.println("\n=== Type Checking ===");
        for (Shape s : shapes) {
            System.out.println(s.getClass().getSimpleName() + " is a Shape: " + (s instanceof Shape));
        }
        System.out.println("Square is a Rectangle: " + (square instanceof Rectangle)); // true!

        // ─── Method overriding vs hiding ───
        System.out.println("\n=== Constructor Chain (watch super calls) ===");
        new Square("Yellow", 7);  // triggers Shape() → Rectangle() → Square()
    }
}

// ─── PARENT CLASS ───
abstract class Shape {
    protected String color;

    public Shape(String color) {
        this.color = color;
        System.out.println("  Shape constructor called (color=" + color + ")");
    }

    // Abstract methods — subclasses MUST implement
    public abstract double area();
    public abstract double perimeter();

    // Concrete method — inherited as-is (can be overridden)
    public void displayInfo() {
        System.out.println(getClass().getSimpleName() + " [" + color + "]");
        System.out.println("  Area:      " + String.format("%.2f", area()));
        System.out.println("  Perimeter: " + String.format("%.2f", perimeter()));
    }
}

// ─── CHILD CLASS: Circle ───
class Circle extends Shape {
    private double radius;

    public Circle(String color, double radius) {
        super(color);  // MUST call parent constructor first
        this.radius = radius;
    }

    @Override
    public double area() {
        return Math.PI * radius * radius;
    }

    @Override
    public double perimeter() {
        return 2 * Math.PI * radius;
    }
}

// ─── CHILD CLASS: Rectangle ───
class Rectangle extends Shape {
    protected double width;   // protected so Square can access
    protected double height;

    public Rectangle(String color, double width, double height) {
        super(color);
        this.width = width;
        this.height = height;
    }

    @Override
    public double area() {
        return width * height;
    }

    @Override
    public double perimeter() {
        return 2 * (width + height);
    }
}

// ─── GRANDCHILD CLASS: Square IS-A Rectangle ───
class Square extends Rectangle {

    public Square(String color, double side) {
        super(color, side, side);  // a square is a rectangle with equal sides
        System.out.println("  Square constructor called (side=" + side + ")");
    }

    // No need to override area() or perimeter() — Rectangle's versions work!

    // But we CAN add Square-specific behavior:
    public double diagonal() {
        return width * Math.sqrt(2);
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ `extends` establishes an IS-A relationship.
 * ✦ `super(...)` MUST be the first line in a child constructor.
 * ✦ Constructor chain: parent → child (always top-down).
 * ✦ `abstract` class can have both abstract and concrete methods.
 * ✦ `abstract` methods have no body — child MUST implement them.
 * ✦ `@Override` annotation catches typos at compile time.
 * ✦ `protected` fields are visible to subclasses.
 * ✦ Prefer COMPOSITION over inheritance when "has-a" fits better.
 *
 * ⚠️ COMMON PITFALL: "Square extends Rectangle" violates the
 *    Liskov Substitution Principle if Rectangle has setWidth/setHeight
 *    (changing width on a Square would break the square invariant).
 *    We avoid this here by making the fields set only via constructor.
 *
 * COMPILE & RUN:
 *   javac Inheritance.java && java Inheritance
 */
