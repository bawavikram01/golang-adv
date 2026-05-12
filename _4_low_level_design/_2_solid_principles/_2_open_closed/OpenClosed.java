/*
 * =============================================================
 * SOLID PRINCIPLE 2: OPEN/CLOSED PRINCIPLE (OCP)
 * =============================================================
 *
 * "Software entities should be OPEN for extension
 *  but CLOSED for modification."
 *
 * Translation: Add new behavior by writing NEW code,
 *              not by changing EXISTING code.
 *
 * WHY?
 *   - Existing code is tested and working. Don't touch it.
 *   - New features should not risk breaking old features.
 *   - if/else chains for types → OCP violation.
 */

import java.util.List;
import java.util.ArrayList;

public class OpenClosed {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // BAD: Adding a new shape requires modifying AreaCalculator
        // ═══════════════════════════════════════════════════════
        System.out.println("=== BAD: OCP Violation ===");
        BadAreaCalculator badCalc = new BadAreaCalculator();
        System.out.println("Circle area:    " + badCalc.calculateArea("circle", 5, 0));
        System.out.println("Rectangle area: " + badCalc.calculateArea("rectangle", 4, 6));
        // Want to add Triangle? Must modify BadAreaCalculator's if-else chain!

        // ═══════════════════════════════════════════════════════
        // GOOD: Adding new shapes requires ZERO changes to existing code
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GOOD: OCP Applied ===");

        List<Shape> shapes = new ArrayList<>();
        shapes.add(new Circle(5));
        shapes.add(new Rectangle(4, 6));
        shapes.add(new Triangle(3, 4, 5));  // NEW! No existing code changed!

        AreaCalculator calculator = new AreaCalculator();
        double totalArea = calculator.totalArea(shapes);

        for (Shape s : shapes) {
            System.out.printf("  %-12s area = %.2f%n", s.getClass().getSimpleName(), s.area());
        }
        System.out.printf("  Total area = %.2f%n", totalArea);

        // ═══════════════════════════════════════════════════════
        // REAL-WORLD: Discount system — open for new discount types
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== REAL-WORLD: Discount System ===");

        double originalPrice = 1000.0;

        List<DiscountStrategy> discounts = new ArrayList<>();
        discounts.add(new PercentageDiscount(10));       // 10% off
        discounts.add(new FlatDiscount(50));              // $50 off
        discounts.add(new BuyOneGetOneFreeDiscount());    // BOGO

        PriceCalculator priceCalc = new PriceCalculator();
        for (DiscountStrategy d : discounts) {
            double finalPrice = priceCalc.applyDiscount(originalPrice, d);
            System.out.printf("  %-25s: $%.2f → $%.2f%n",
                    d.getClass().getSimpleName(), originalPrice, finalPrice);
        }
        // Adding SeasonalDiscount, LoyaltyDiscount, etc. requires ZERO changes
        // to PriceCalculator. Just create a new class implementing DiscountStrategy.
    }
}

// ═══════════════════════════════════════════════════════════════
// BAD: Must modify this class for every new shape
// ═══════════════════════════════════════════════════════════════
class BadAreaCalculator {
    public double calculateArea(String shapeType, double a, double b) {
        if (shapeType.equals("circle")) {
            return Math.PI * a * a;
        } else if (shapeType.equals("rectangle")) {
            return a * b;
        }
        // else if ("triangle") ... → MODIFICATION required!
        // else if ("hexagon") ...  → MODIFICATION required!
        // This grows forever. Every addition risks breaking existing logic.
        return 0;
    }
}

// ═══════════════════════════════════════════════════════════════
// GOOD: Abstraction + polymorphism = OCP
// ═══════════════════════════════════════════════════════════════
interface Shape {
    double area();
}

class Circle implements Shape {
    private double radius;
    public Circle(double radius) { this.radius = radius; }
    @Override public double area() { return Math.PI * radius * radius; }
}

class Rectangle implements Shape {
    private double width, height;
    public Rectangle(double w, double h) { this.width = w; this.height = h; }
    @Override public double area() { return width * height; }
}

// NEW shape — no existing code touched!
class Triangle implements Shape {
    private double a, b, c;
    public Triangle(double a, double b, double c) { this.a = a; this.b = b; this.c = c; }
    @Override public double area() {
        double s = (a + b + c) / 2;
        return Math.sqrt(s * (s - a) * (s - b) * (s - c)); // Heron's formula
    }
}

// This class NEVER needs to change for new shapes
class AreaCalculator {
    public double totalArea(List<Shape> shapes) {
        return shapes.stream().mapToDouble(Shape::area).sum();
    }
}

// ═══════════════════════════════════════════════════════════════
// REAL-WORLD: Discount Strategy
// ═══════════════════════════════════════════════════════════════
interface DiscountStrategy {
    double apply(double price);
}

class PercentageDiscount implements DiscountStrategy {
    private double percent;
    public PercentageDiscount(double percent) { this.percent = percent; }
    @Override public double apply(double price) { return price * (1 - percent / 100); }
}

class FlatDiscount implements DiscountStrategy {
    private double amount;
    public FlatDiscount(double amount) { this.amount = amount; }
    @Override public double apply(double price) { return Math.max(0, price - amount); }
}

class BuyOneGetOneFreeDiscount implements DiscountStrategy {
    @Override public double apply(double price) { return price / 2; }
}

// PriceCalculator is CLOSED for modification, OPEN for new discount types
class PriceCalculator {
    public double applyDiscount(double price, DiscountStrategy strategy) {
        return strategy.apply(price);
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ if/else chains based on TYPE → OCP violation.
 * ✦ Use ABSTRACTION (interface/abstract class) + POLYMORPHISM.
 * ✦ New behavior = new class implementing the interface.
 * ✦ Existing code (calculator) never changes.
 * ✦ This is the Strategy Pattern in action.
 *
 * COMPILE & RUN:
 *   javac OpenClosed.java && java OpenClosed
 */
