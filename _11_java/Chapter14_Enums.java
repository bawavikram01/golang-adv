/*
 * ============================================================
 *  CHAPTER 14: ENUMS
 * ============================================================
 *
 *  ENUM = a special class that represents a group of CONSTANTS
 *         (unchangeable variables).
 *
 *  WHY ENUMS?
 *  - Type safety (can't pass invalid values)
 *  - Built-in methods (name(), ordinal(), values())
 *  - Can have fields, constructors, and methods
 *  - Can implement interfaces
 *  - Used in switch statements
 *
 *  KEY FACTS:
 *  - Enum implicitly extends java.lang.Enum (can't extend anything else)
 *  - Enum constants are implicitly public, static, final
 *  - Constructor is implicitly private (can't instantiate with new)
 *  - All enums get values(), valueOf(), name(), ordinal() for free
 *
 *  NOTE: Records are Java 16+ (not available in Java 11)
 *
 * ============================================================
 */

public class Chapter14_Enums {

    // =====================================================
    //  1. BASIC ENUM
    // =====================================================

    // Simple enum — just constants
    enum Day {
        MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY
    }

    enum Season {
        SPRING, SUMMER, AUTUMN, WINTER
    }

    // =====================================================
    //  2. ENUM WITH FIELDS AND METHODS
    // =====================================================

    // Enums can have fields, constructors, and methods!
    enum Planet {
        MERCURY(3.303e+23, 2.4397e6),
        VENUS(4.869e+24, 6.0518e6),
        EARTH(5.976e+24, 6.37814e6),
        MARS(6.421e+23, 3.3972e6),
        JUPITER(1.9e+27, 7.1492e7),
        SATURN(5.688e+26, 6.0268e7),
        URANUS(8.686e+25, 2.5559e7),
        NEPTUNE(1.024e+26, 2.4746e7);

        private final double mass;    // in kilograms
        private final double radius;  // in meters

        // Constructor is implicitly PRIVATE
        Planet(double mass, double radius) {
            this.mass = mass;
            this.radius = radius;
        }

        // Getters
        double getMass() { return mass; }
        double getRadius() { return radius; }

        // Methods
        double surfaceGravity() {
            final double G = 6.67300E-11;
            return G * mass / (radius * radius);
        }

        double surfaceWeight(double otherMass) {
            return otherMass * surfaceGravity();
        }
    }

    // =====================================================
    //  3. ENUM WITH ABSTRACT METHODS
    // =====================================================

    enum Operation {
        ADD("+") {
            @Override
            public double apply(double x, double y) { return x + y; }
        },
        SUBTRACT("-") {
            @Override
            public double apply(double x, double y) { return x - y; }
        },
        MULTIPLY("*") {
            @Override
            public double apply(double x, double y) { return x * y; }
        },
        DIVIDE("/") {
            @Override
            public double apply(double x, double y) {
                if (y == 0) throw new ArithmeticException("Division by zero");
                return x / y;
            }
        };

        private final String symbol;

        Operation(String symbol) {
            this.symbol = symbol;
        }

        public String getSymbol() { return symbol; }

        // Abstract method — each constant MUST implement
        public abstract double apply(double x, double y);

        @Override
        public String toString() {
            return symbol;
        }
    }

    // =====================================================
    //  4. ENUM IMPLEMENTING INTERFACE
    // =====================================================

    interface Describable {
        String describe();
    }

    enum Color implements Describable {
        RED("#FF0000", "Warm"),
        GREEN("#00FF00", "Cool"),
        BLUE("#0000FF", "Cool"),
        YELLOW("#FFFF00", "Warm"),
        WHITE("#FFFFFF", "Neutral"),
        BLACK("#000000", "Neutral");

        private final String hexCode;
        private final String temperature;

        Color(String hexCode, String temperature) {
            this.hexCode = hexCode;
            this.temperature = temperature;
        }

        public String getHexCode() { return hexCode; }
        public String getTemperature() { return temperature; }

        @Override
        public String describe() {
            return name() + " (" + hexCode + ") - " + temperature;
        }
    }

    // =====================================================
    //  5. ENUM FOR STATE MACHINE
    // =====================================================

    enum TrafficLight {
        RED(30) {
            @Override
            TrafficLight next() { return GREEN; }
        },
        GREEN(25) {
            @Override
            TrafficLight next() { return YELLOW; }
        },
        YELLOW(5) {
            @Override
            TrafficLight next() { return RED; }
        };

        private final int duration; // seconds

        TrafficLight(int duration) {
            this.duration = duration;
        }

        public int getDuration() { return duration; }
        abstract TrafficLight next();
    }

    // =====================================================
    //  6. SINGLETON USING ENUM (Best practice!)
    // =====================================================

    enum DatabaseConnection {
        INSTANCE; // only one instance ever created

        private String connectionString = "jdbc:mysql://localhost:3306/mydb";

        public void connect() {
            System.out.println("Connected to: " + connectionString);
        }

        public void disconnect() {
            System.out.println("Disconnected from: " + connectionString);
        }

        public void query(String sql) {
            System.out.println("Executing: " + sql);
        }
    }

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Basic Enum ---
        System.out.println("=== BASIC ENUM ===\n");

        Day today = Day.WEDNESDAY;
        System.out.println("Today: " + today);
        System.out.println("Ordinal: " + today.ordinal()); // 2 (0-based index)
        System.out.println("Name: " + today.name());       // WEDNESDAY

        // Switch with enum
        switch (today) {
            case MONDAY: case TUESDAY: case WEDNESDAY: case THURSDAY: case FRIDAY:
                System.out.println("It's a weekday.");
                break;
            case SATURDAY: case SUNDAY:
                System.out.println("It's the weekend!");
                break;
        }

        // Iterating all values
        System.out.println("\nAll days:");
        for (Day d : Day.values()) {
            System.out.println("  " + d.ordinal() + ": " + d);
        }

        // valueOf — convert String to enum
        Day monday = Day.valueOf("MONDAY");
        System.out.println("\nvalueOf(\"MONDAY\"): " + monday);
        // Day.valueOf("monday"); // IllegalArgumentException! Case-sensitive!

        // Comparing enums — use == (not .equals())
        System.out.println("today == WEDNESDAY: " + (today == Day.WEDNESDAY)); // true

        // --- 2. Enum with Fields ---
        System.out.println("\n=== ENUM WITH FIELDS ===\n");

        double earthWeight = 75.0; // kg
        double mass = earthWeight / Planet.EARTH.surfaceGravity();

        System.out.printf("Your weight on each planet (Earth weight: %.1f kg):%n", earthWeight);
        for (Planet p : Planet.values()) {
            System.out.printf("  %-8s: %6.2f kg%n", p, p.surfaceWeight(mass));
        }

        // --- 3. Enum with Abstract Methods ---
        System.out.println("\n=== ENUM WITH ABSTRACT METHODS ===\n");

        double x = 10, y = 3;
        for (Operation op : Operation.values()) {
            System.out.printf("  %.0f %s %.0f = %.2f%n", x, op, y, op.apply(x, y));
        }

        // --- 4. Enum Implementing Interface ---
        System.out.println("\n=== ENUM IMPLEMENTING INTERFACE ===\n");

        for (Color c : Color.values()) {
            System.out.println("  " + c.describe());
        }

        System.out.println("\nWarm colors:");
        for (Color c : Color.values()) {
            if ("Warm".equals(c.getTemperature())) {
                System.out.println("  " + c.name());
            }
        }

        // --- 5. State Machine ---
        System.out.println("\n=== STATE MACHINE ===\n");

        TrafficLight light = TrafficLight.RED;
        for (int i = 0; i < 6; i++) {
            System.out.println("Light: " + light + " (duration: " + light.getDuration() + "s)");
            light = light.next();
        }

        // --- 6. Singleton ---
        System.out.println("\n=== SINGLETON ENUM ===\n");

        // Always the same instance
        DatabaseConnection db = DatabaseConnection.INSTANCE;
        db.connect();
        db.query("SELECT * FROM users");
        db.disconnect();

        // Same instance
        DatabaseConnection db2 = DatabaseConnection.INSTANCE;
        System.out.println("Same instance: " + (db == db2)); // true

        // --- 7. EnumSet and EnumMap ---
        System.out.println("\n=== ENUM BEST PRACTICES ===\n");
        System.out.println("1. Use enum instead of int constants (type-safe)");
        System.out.println("2. Use enum instead of String constants");
        System.out.println("3. Use enum for singleton pattern");
        System.out.println("4. Enum == comparison is safe (unlike objects)");
        System.out.println("5. Enums are inherently serializable");
        System.out.println("6. Enums can implement interfaces but not extend classes");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Create an enum `Size` (SMALL, MEDIUM, LARGE, XL) with
 *     a price field. Add a method to get the price.
 *
 *  2. Create an enum `Direction` (NORTH, SOUTH, EAST, WEST)
 *     with an `opposite()` method.
 *
 *  3. Create an enum `Priority` (LOW, MEDIUM, HIGH, CRITICAL)
 *     that implements Comparable. Use it to sort tasks.
 *
 *  4. Create a `Calculator` using Operation enum — accept two
 *     numbers and an operation, print the result.
 *
 *  5. Create a `GameState` enum (MENU, PLAYING, PAUSED, GAME_OVER)
 *     with transition methods and validation.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 15 — Exception Handling
 * ============================================================
 */
