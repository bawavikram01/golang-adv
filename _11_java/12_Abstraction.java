/*
 * ============================================================
 *  CHAPTER 12: OOP — ABSTRACTION
 * ============================================================
 *
 *  ABSTRACTION = Hiding implementation details,
 *                showing only essential features.
 *
 *  Two ways to achieve abstraction in Java:
 *  1. ABSTRACT CLASSES  (0-100% abstraction)
 *  2. INTERFACES         (100% abstraction — traditionally)
 *
 *  ┌───────────────────────┬──────────────────┬────────────────────────┐
 *  │ Feature               │ Abstract Class   │ Interface              │
 *  ├───────────────────────┼──────────────────┼────────────────────────┤
 *  │ Keyword               │ abstract class   │ interface              │
 *  │ abstract methods      │ Yes              │ Yes (default)          │
 *  │ concrete methods      │ Yes              │ Yes (default/static)   │
 *  │ constructors          │ Yes              │ No                     │
 *  │ fields                │ Any type         │ public static final    │
 *  │ access modifiers      │ Any              │ public (default)       │
 *  │ inheritance           │ extends (single) │ implements (multiple!) │
 *  │ can instantiate?      │ No               │ No                     │
 *  └───────────────────────┴──────────────────┴────────────────────────┘
 *
 *  WHEN TO USE WHAT:
 *  - Abstract class: shared state + partial implementation + is-a relationship
 *  - Interface: define a capability/contract + multiple inheritance + has-a behavior
 *
 * ============================================================
 */

public class Chapter12_Abstraction {

    // =====================================================
    //  1. ABSTRACT CLASSES
    // =====================================================

    // An abstract class CANNOT be instantiated
    // It can have both abstract and concrete methods
    static abstract class Shape {
        String color;

        // Constructor — yes, abstract classes CAN have constructors!
        Shape(String color) {
            this.color = color;
        }

        // ABSTRACT method — no body, MUST be overridden by subclasses
        abstract double area();
        abstract double perimeter();

        // CONCRETE method — has a body, inherited by subclasses
        void displayInfo() {
            System.out.printf("%s: color=%s, area=%.2f, perimeter=%.2f%n",
                    getClass().getSimpleName(), color, area(), perimeter());
        }
    }

    // Concrete class — MUST implement ALL abstract methods
    static class Circle extends Shape {
        double radius;

        Circle(String color, double radius) {
            super(color);
            this.radius = radius;
        }

        @Override
        double area() { return Math.PI * radius * radius; }

        @Override
        double perimeter() { return 2 * Math.PI * radius; }
    }

    static class Rectangle extends Shape {
        double length, width;

        Rectangle(String color, double length, double width) {
            super(color);
            this.length = length;
            this.width = width;
        }

        @Override
        double area() { return length * width; }

        @Override
        double perimeter() { return 2 * (length + width); }
    }

    // =====================================================
    //  2. INTERFACES
    // =====================================================

    // An interface defines a CONTRACT — what a class MUST do

    interface Drawable {
        // All methods are implicitly "public abstract"
        void draw();
        void resize(double factor);
    }

    interface Printable {
        void print();
    }

    // A class can implement MULTIPLE interfaces!
    interface Saveable {
        void save(String filename);
    }

    // DEFAULT methods (Java 8+) — provide a default implementation
    interface Loggable {
        // Regular abstract method
        void log(String message);

        // DEFAULT method — concrete implementation in interface!
        default void logInfo(String message) {
            log("INFO: " + message);
        }

        default void logError(String message) {
            log("ERROR: " + message);
        }

        // STATIC method in interface
        static String formatLog(String level, String msg) {
            return "[" + level + "] " + msg;
        }
    }

    // =====================================================
    //  3. IMPLEMENTING INTERFACES
    // =====================================================

    static class DrawableCircle extends Circle implements Drawable, Printable, Loggable {

        DrawableCircle(String color, double radius) {
            super(color, radius);
        }

        @Override
        public void draw() {
            System.out.println("Drawing circle with radius " + radius);
        }

        @Override
        public void resize(double factor) {
            radius *= factor;
            System.out.println("Resized circle to radius " + radius);
        }

        @Override
        public void print() {
            System.out.println("Printing: Circle r=" + radius + " in " + color);
        }

        @Override
        public void log(String message) {
            System.out.println("[LOG] " + message);
        }
    }

    // =====================================================
    //  4. INTERFACE INHERITANCE
    // =====================================================

    // Interfaces can extend OTHER interfaces
    interface Readable {
        String read();
    }

    interface Writable {
        void write(String data);
    }

    // Interface extending multiple interfaces
    interface ReadWritable extends Readable, Writable {
        void flush();
    }

    static class File implements ReadWritable {
        private String content = "";

        @Override
        public String read() {
            return content;
        }

        @Override
        public void write(String data) {
            content += data;
        }

        @Override
        public void flush() {
            System.out.println("Flushing: " + content);
        }
    }

    // =====================================================
    //  5. FUNCTIONAL INTERFACE (Single Abstract Method)
    // =====================================================

    // An interface with EXACTLY ONE abstract method
    // Can have default/static methods too
    @FunctionalInterface
    interface MathOperation {
        double operate(double a, double b);

        // default methods don't count
        default void describe() {
            System.out.println("A math operation");
        }
    }

    // Can be implemented with lambda (more in Chapter 20)
    // MathOperation add = (a, b) -> a + b;

    // =====================================================
    //  6. MARKER INTERFACES
    // =====================================================

    // An interface with NO methods — just "marks" a class
    // Examples: Serializable, Cloneable
    interface Exportable {
        // empty — just marks that a class can be exported
    }

    static class Report implements Exportable {
        String title;
        Report(String title) { this.title = title; }
    }

    // =====================================================
    //  7. ABSTRACT CLASS vs INTERFACE — PRACTICAL EXAMPLE
    // =====================================================

    // Abstract class: Template Method Pattern
    static abstract class DataProcessor {
        // Template method — defines the skeleton
        final void process() {
            readData();
            validateData();
            transformData();
            saveData();
        }

        abstract void readData();
        abstract void transformData();

        // Hook methods — can be overridden but have default behavior
        void validateData() {
            System.out.println("  Default validation passed.");
        }

        void saveData() {
            System.out.println("  Data saved to default location.");
        }
    }

    static class CSVProcessor extends DataProcessor {
        @Override
        void readData() {
            System.out.println("  Reading CSV file...");
        }

        @Override
        void transformData() {
            System.out.println("  Transforming CSV data...");
        }

        @Override
        void validateData() {
            System.out.println("  Validating CSV format...");
        }
    }

    static class JSONProcessor extends DataProcessor {
        @Override
        void readData() {
            System.out.println("  Reading JSON file...");
        }

        @Override
        void transformData() {
            System.out.println("  Parsing JSON data...");
        }
    }

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Abstract Classes ---
        System.out.println("=== ABSTRACT CLASSES ===\n");

        // Shape shape = new Shape("Red"); // ERROR: can't instantiate abstract!

        Shape circle = new Circle("Red", 5);
        Shape rect = new Rectangle("Blue", 4, 6);

        circle.displayInfo(); // concrete method calling abstract method
        rect.displayInfo();

        // --- 2. Interfaces ---
        System.out.println("\n=== INTERFACES ===\n");

        DrawableCircle dc = new DrawableCircle("Green", 7);
        dc.draw();
        dc.print();
        dc.resize(2);
        dc.draw();
        dc.displayInfo();

        // --- 3. Default and Static Methods ---
        System.out.println("\n=== DEFAULT & STATIC METHODS ===\n");

        dc.log("Direct log");
        dc.logInfo("Something happened");     // uses default method
        dc.logError("Something went wrong");  // uses default method

        // Static method on interface
        String formatted = Loggable.formatLog("WARN", "Low disk space");
        System.out.println(formatted);

        // --- 4. Interface References ---
        System.out.println("\n=== INTERFACE REFERENCES ===\n");

        // An object can be referenced by any interface it implements
        Drawable drawable = dc;
        Printable printable = dc;
        Loggable loggable = dc;

        drawable.draw();
        printable.print();
        loggable.logInfo("Via Loggable reference");

        // --- 5. Multiple Inheritance Through Interfaces ---
        System.out.println("\n=== MULTIPLE INTERFACES ===\n");

        File file = new File();
        file.write("Hello ");
        file.write("World");
        System.out.println("Read: " + file.read());
        file.flush();

        // --- 6. Functional Interfaces ---
        System.out.println("\n=== FUNCTIONAL INTERFACES ===\n");

        // Traditional implementation
        MathOperation addition = new MathOperation() {
            @Override
            public double operate(double a, double b) {
                return a + b;
            }
        };
        System.out.println("5 + 3 = " + addition.operate(5, 3));

        // Lambda (preview — covered fully in Ch 20)
        MathOperation subtraction = (a, b) -> a - b;
        MathOperation multiplication = (a, b) -> a * b;
        MathOperation division = (a, b) -> b != 0 ? a / b : 0;

        System.out.println("5 - 3 = " + subtraction.operate(5, 3));
        System.out.println("5 * 3 = " + multiplication.operate(5, 3));
        System.out.println("5 / 3 = " + division.operate(5, 3));

        // --- 7. Marker Interface ---
        System.out.println("\n=== MARKER INTERFACE ===\n");

        Report report = new Report("Annual Report");
        System.out.println("Is Exportable: " + (report instanceof Exportable));

        // --- 8. Template Method Pattern ---
        System.out.println("\n=== TEMPLATE METHOD (Abstract Class Pattern) ===\n");

        System.out.println("CSV Processing:");
        DataProcessor csv = new CSVProcessor();
        csv.process();

        System.out.println("\nJSON Processing:");
        DataProcessor json = new JSONProcessor();
        json.process();

        // --- Diamond Problem ---
        System.out.println("\n=== DIAMOND PROBLEM ===\n");
        System.out.println("If two interfaces have the same default method,");
        System.out.println("the implementing class MUST override it to resolve ambiguity.");
        System.out.println("This is how Java avoids the 'diamond problem' of multiple inheritance.");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Create an abstract class `Animal` with abstract method
 *     `makeSound()` and concrete method `sleep()`.
 *     Implement Dog, Cat, and Bird subclasses.
 *
 *  2. Create interfaces `Flyable` and `Swimmable`.
 *     Create `Duck` that implements both.
 *     Create `Penguin` that implements only Swimmable.
 *
 *  3. Create interface `Sortable<T>` with method `sort(T[] arr)`.
 *     Implement it in two classes using different algorithms.
 *
 *  4. Create a `Validator<T>` functional interface with method
 *     `boolean validate(T value)`. Use lambdas to create
 *     validators for: positive numbers, non-empty strings, valid emails.
 *
 *  5. Implement the Template Method pattern for a `GameAI`:
 *     abstract class GameAI { collectResources, buildArmy, attack }
 *     Implement OffensiveAI and DefensiveAI.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 13 — Inner Classes
 * ============================================================
 */
