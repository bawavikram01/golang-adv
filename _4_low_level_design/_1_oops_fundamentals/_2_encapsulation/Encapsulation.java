/*
 * =============================================================
 * MODULE 2: ENCAPSULATION — Hide the Guts, Expose the Interface
 * =============================================================
 *
 * Encapsulation = DATA HIDING + CONTROLLED ACCESS
 *
 * Why?
 *   1. Protects internal state from invalid modifications
 *   2. You can change internals without breaking outside code
 *   3. Forces users to interact through a well-defined API
 *
 * ACCESS MODIFIERS:
 *   ┌────────────┬───────┬─────────┬──────────┬───────────┐
 *   │ Modifier   │ Class │ Package │ Subclass │ Everywhere│
 *   ├────────────┼───────┼─────────┼──────────┼───────────┤
 *   │ private    │  ✓    │         │          │           │
 *   │ (default)  │  ✓    │   ✓     │          │           │
 *   │ protected  │  ✓    │   ✓     │    ✓     │           │
 *   │ public     │  ✓    │   ✓     │    ✓     │     ✓     │
 *   └────────────┴───────┴─────────┴──────────┴───────────┘
 */

public class Encapsulation {

    public static void main(String[] args) {

        // ─── BAD: No encapsulation (public fields) ───
        System.out.println("=== BAD: No Encapsulation ===");
        BadEmployee bad = new BadEmployee();
        bad.name = "Alice";
        bad.salary = -50000;  // Negative salary? No validation!
        bad.age = 999;        // 999 years old? Sure, why not!
        System.out.println(bad.name + " earns " + bad.salary + ", age: " + bad.age);

        // ─── GOOD: Encapsulated (private fields + validated setters) ───
        System.out.println("\n=== GOOD: With Encapsulation ===");
        Employee emp = new Employee("Bob", 50000, 30);
        System.out.println(emp);

        emp.setSalary(-100);    // rejected
        emp.setAge(200);        // rejected
        emp.giveRaise(10);      // 10% raise
        System.out.println(emp);

        // ─── Immutable class — the ULTIMATE encapsulation ───
        System.out.println("\n=== IMMUTABLE CLASS ===");
        ImmutablePoint p = new ImmutablePoint(3, 4);
        System.out.println("Point: " + p);
        System.out.println("Distance from origin: " + p.distanceFromOrigin());

        // p.x = 10;  // COMPILE ERROR — fields are private and final
        // No setters exist — you CANNOT modify this object after creation
        ImmutablePoint moved = p.translate(2, 3);  // returns a NEW object
        System.out.println("Translated: " + moved);
        System.out.println("Original unchanged: " + p);
    }
}

// ─── BAD DESIGN: Everything is public ───
class BadEmployee {
    public String name;
    public double salary;
    public int age;
    // Anyone can set anything to anything. Total chaos.
}

// ─── GOOD DESIGN: Properly encapsulated ───
class Employee {
    private String name;
    private double salary;
    private int age;

    public Employee(String name, double salary, int age) {
        this.name = name;
        setSalary(salary);  // reuse validation
        setAge(age);
    }

    // ─── Getters: read-only access ───
    public String getName()   { return name; }
    public double getSalary() { return salary; }
    public int getAge()       { return age; }

    // ─── Setters: write access WITH validation ───
    public void setSalary(double salary) {
        if (salary < 0) {
            System.out.println("  ✗ Salary cannot be negative. Rejected.");
            return;
        }
        this.salary = salary;
    }

    public void setAge(int age) {
        if (age < 18 || age > 120) {
            System.out.println("  ✗ Age must be 18-120. Rejected.");
            return;
        }
        this.age = age;
    }

    // ─── Business logic method ───
    public void giveRaise(double percent) {
        this.salary += this.salary * percent / 100;
        System.out.println("  ✓ " + name + " got a " + percent + "% raise!");
    }

    @Override
    public String toString() {
        return "Employee{name='" + name + "', salary=" + salary + ", age=" + age + "}";
    }
}

// ─── IMMUTABLE CLASS: Cannot be modified after creation ───
// Recipe: (1) final class, (2) private final fields, (3) no setters,
//         (4) return new objects for "modifications"
final class ImmutablePoint {
    private final double x;
    private final double y;

    public ImmutablePoint(double x, double y) {
        this.x = x;
        this.y = y;
    }

    public double getX() { return x; }
    public double getY() { return y; }

    public double distanceFromOrigin() {
        return Math.sqrt(x * x + y * y);
    }

    // Instead of modifying, return a NEW point
    public ImmutablePoint translate(double dx, double dy) {
        return new ImmutablePoint(this.x + dx, this.y + dy);
    }

    @Override
    public String toString() {
        return "(" + x + ", " + y + ")";
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Make all fields PRIVATE. Always. No exceptions.
 * ✦ Provide getters for read access, setters for write access.
 * ✦ Put VALIDATION in setters — this is the whole point.
 * ✦ For maximum safety, make classes IMMUTABLE:
 *     - `final` class, `private final` fields, no setters
 *     - Return new objects instead of modifying state
 * ✦ Immutable objects are thread-safe by default.
 *
 * COMPILE & RUN:
 *   javac Encapsulation.java && java Encapsulation
 */
