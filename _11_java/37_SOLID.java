/*
 * ============================================================
 *  CHAPTER 37: SOLID PRINCIPLES
 * ============================================================
 *  5 principles for clean, maintainable OOP code.
 *  Coined by Robert C. Martin ("Uncle Bob").
 *
 *  S — Single Responsibility Principle (SRP)
 *  O — Open/Closed Principle (OCP)
 *  L — Liskov Substitution Principle (LSP)
 *  I — Interface Segregation Principle (ISP)
 *  D — Dependency Inversion Principle (DIP)
 * ============================================================
 */

import java.util.*;

public class Chapter37_SOLID {

    // ========================================================
    // S — SINGLE RESPONSIBILITY PRINCIPLE
    // "A class should have only ONE reason to change"
    // ========================================================

    // ❌ BAD: Employee does too many things
    static class EmployeeBad {
        String name;
        double salary;

        double calculatePay() { return salary; }
        void saveToDatabase() { /* SQL... */ }      // persistence concern
        String generateReport() { return "..."; }   // reporting concern
    }

    // ✅ GOOD: Each class has one job
    static class Employee {
        String name;
        double salary;
        Employee(String name, double salary) { this.name = name; this.salary = salary; }
    }

    static class PayCalculator {
        double calculatePay(Employee emp) { return emp.salary * 1.0; }
        double calculatePayWithBonus(Employee emp, double bonus) { return emp.salary + bonus; }
    }

    static class EmployeeRepository {
        void save(Employee emp) { System.out.println("    Saved: " + emp.name); }
    }

    static class EmployeeReportGenerator {
        String generate(Employee emp) { return "Report for " + emp.name; }
    }

    // ========================================================
    // O — OPEN/CLOSED PRINCIPLE
    // "Open for extension, closed for modification"
    // ========================================================

    // ❌ BAD: Adding new shape requires modifying AreaCalculator
    static class AreaCalculatorBad {
        double calculate(Object shape) {
            if (shape instanceof double[]) return Math.PI * ((double[]) shape)[0] * ((double[]) shape)[0];
            // Adding rectangle? Must modify THIS method every time!
            return 0;
        }
    }

    // ✅ GOOD: New shapes just implement the interface
    interface Shape {
        double area();
    }

    static class CircleShape implements Shape {
        double radius;
        CircleShape(double r) { this.radius = r; }
        @Override public double area() { return Math.PI * radius * radius; }
    }

    static class RectangleShape implements Shape {
        double width, height;
        RectangleShape(double w, double h) { this.width = w; this.height = h; }
        @Override public double area() { return width * height; }
    }

    // Adding triangle? No modification needed — just add new class!
    static class TriangleShape implements Shape {
        double base, height;
        TriangleShape(double b, double h) { this.base = b; this.height = h; }
        @Override public double area() { return 0.5 * base * height; }
    }

    static class AreaCalculator {
        double totalArea(List<Shape> shapes) {
            return shapes.stream().mapToDouble(Shape::area).sum();
        }
    }

    // ========================================================
    // L — LISKOV SUBSTITUTION PRINCIPLE
    // "Subtypes must be substitutable for their base types"
    // ========================================================

    // ❌ BAD: Square breaks Rectangle contract
    static class RectangleBad {
        protected int width, height;
        void setWidth(int w) { this.width = w; }
        void setHeight(int h) { this.height = h; }
        int area() { return width * height; }
    }

    // Square overrides setWidth/setHeight to keep sides equal
    // This BREAKS the contract: setWidth(5) then setHeight(3) should give area=15
    // But Square gives area=9 — SURPRISING to code using Rectangle!

    // ✅ GOOD: Separate types, shared interface
    interface ShapeWithArea {
        int area();
    }

    static class RectangleGood implements ShapeWithArea {
        private int width, height;
        RectangleGood(int w, int h) { this.width = w; this.height = h; }
        @Override public int area() { return width * height; }
    }

    static class SquareGood implements ShapeWithArea {
        private int side;
        SquareGood(int s) { this.side = s; }
        @Override public int area() { return side * side; }
    }

    // ========================================================
    // I — INTERFACE SEGREGATION PRINCIPLE
    // "Don't force clients to depend on methods they don't use"
    // ========================================================

    // ❌ BAD: One fat interface
    interface WorkerBad {
        void work();
        void eat();
        void sleep();
    }
    // A Robot can work but can't eat or sleep!

    // ✅ GOOD: Segregated interfaces
    interface Workable { void work(); }
    interface Eatable { void eat(); }
    interface Sleepable { void sleep(); }

    static class HumanWorker implements Workable, Eatable, Sleepable {
        @Override public void work() { System.out.println("    Human working"); }
        @Override public void eat() { System.out.println("    Human eating"); }
        @Override public void sleep() { System.out.println("    Human sleeping"); }
    }

    static class RobotWorker implements Workable {
        @Override public void work() { System.out.println("    Robot working"); }
        // No eat() or sleep() — Robot doesn't need them!
    }

    // ========================================================
    // D — DEPENDENCY INVERSION PRINCIPLE
    // "Depend on abstractions, not concretions"
    // ========================================================

    // ❌ BAD: High-level depends on low-level
    static class MySQLDatabase {
        void save(String data) { System.out.println("    MySQL: saving " + data); }
    }

    static class UserServiceBad {
        private MySQLDatabase db = new MySQLDatabase();  // tightly coupled!
        void createUser(String name) { db.save(name); }
        // Switching to PostgreSQL = rewrite UserService!
    }

    // ✅ GOOD: Both depend on abstraction
    interface Database {
        void save(String data);
        String find(String id);
    }

    static class MySQLDB implements Database {
        @Override public void save(String data) { System.out.println("    MySQL: " + data); }
        @Override public String find(String id) { return "MySQL result"; }
    }

    static class PostgresDB implements Database {
        @Override public void save(String data) { System.out.println("    Postgres: " + data); }
        @Override public String find(String id) { return "Postgres result"; }
    }

    static class UserService {
        private final Database db;  // depends on ABSTRACTION

        UserService(Database db) { this.db = db; }  // injected!

        void createUser(String name) {
            db.save("User: " + name);
        }
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        // --- S: SRP ---
        System.out.println("=== S — SINGLE RESPONSIBILITY ===\n");
        Employee emp = new Employee("Alice", 5000);
        PayCalculator pay = new PayCalculator();
        EmployeeRepository repo = new EmployeeRepository();
        EmployeeReportGenerator report = new EmployeeReportGenerator();

        System.out.println("  Pay: " + pay.calculatePay(emp));
        repo.save(emp);
        System.out.println("  " + report.generate(emp));
        System.out.println("  ✓ Each class has ONE responsibility");

        // --- O: OCP ---
        System.out.println("\n=== O — OPEN/CLOSED ===\n");
        List<Shape> shapes = List.of(
            new CircleShape(5),
            new RectangleShape(4, 6),
            new TriangleShape(3, 8)
        );
        AreaCalculator calc = new AreaCalculator();
        System.out.println("  Total area: " + String.format("%.2f", calc.totalArea(shapes)));
        System.out.println("  ✓ New shapes added without modifying calculator");

        // --- L: LSP ---
        System.out.println("\n=== L — LISKOV SUBSTITUTION ===\n");
        ShapeWithArea rect = new RectangleGood(5, 3);
        ShapeWithArea sq = new SquareGood(4);
        System.out.println("  Rectangle area: " + rect.area());
        System.out.println("  Square area: " + sq.area());
        System.out.println("  ✓ Both substitutable via ShapeWithArea interface");

        // --- I: ISP ---
        System.out.println("\n=== I — INTERFACE SEGREGATION ===\n");
        HumanWorker human = new HumanWorker();
        RobotWorker robot = new RobotWorker();
        human.work();
        human.eat();
        robot.work();
        // robot.eat() → doesn't exist, not forced to implement!
        System.out.println("  ✓ Robot not forced to implement eat/sleep");

        // --- D: DIP ---
        System.out.println("\n=== D — DEPENDENCY INVERSION ===\n");
        // Can switch database without changing UserService!
        UserService service1 = new UserService(new MySQLDB());
        UserService service2 = new UserService(new PostgresDB());

        service1.createUser("Alice");
        service2.createUser("Bob");
        System.out.println("  ✓ Swapped database without changing UserService");

        // --- Summary ---
        System.out.println("\n=== SOLID SUMMARY ===");
        System.out.println("  S  One class = one reason to change");
        System.out.println("  O  Extend behavior without modifying existing code");
        System.out.println("  L  Subtypes must honor parent contracts");
        System.out.println("  I  Many specific interfaces > one fat interface");
        System.out.println("  D  Depend on abstractions, inject dependencies");

        System.out.println("\n✓ SOLID Principles Complete!");
    }
}

/*
 * EXERCISES:
 * 1. SRP: Refactor a class that handles validation, persistence, and logging.
 * 2. OCP: Create a discount system (FlatDiscount, PercentDiscount, BuyOneGetOne).
 * 3. DIP: Create a NotificationService that can use Email, SMS, or Push.
 * 4. Review your previous code and identify SOLID violations. Fix them.
 *
 * NEXT: Chapter 38 — Clean Code & Best Practices
 */
