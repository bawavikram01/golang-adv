/*
 * =============================================================
 * SOLID PRINCIPLE 3: LISKOV SUBSTITUTION PRINCIPLE (LSP)
 * =============================================================
 *
 * "If S is a subtype of T, then objects of type T may be replaced
 *  with objects of type S without altering correctness."
 *
 * Translation: A child class should be usable EVERYWHERE its
 *              parent is used, without surprises.
 *
 * Violations (red flags):
 *   - Overriding a method to throw UnsupportedOperationException
 *   - Overriding a method to do nothing
 *   - Checking `instanceof` before calling a method
 *   - Child weakens parent's promises (preconditions/postconditions)
 */

import java.util.List;
import java.util.ArrayList;

public class LiskovSubstitution {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // BAD: Square extending Rectangle violates LSP
        // ═══════════════════════════════════════════════════════
        System.out.println("=== BAD: LSP Violation ===");

        BadRectangle rect = new BadRectangle(4, 5);
        System.out.println("Rectangle area: " + rect.area());  // 20 ✓

        BadRectangle square = new BadSquare(4);  // Square IS-A Rectangle?
        square.setWidth(5);  // Uh oh — this ALSO changes height in Square
        System.out.println("Square area after setWidth(5): " + square.area());
        // Expected 20 (5×4), got 25 (5×5) — SURPRISE! LSP violated.

        testRectangleArea(new BadRectangle(4, 5));  // works
        testRectangleArea(new BadSquare(4));          // BREAKS!

        // ═══════════════════════════════════════════════════════
        // GOOD: Proper hierarchy that respects LSP
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GOOD: LSP Respected ===");

        Shape rectangle = new GoodRectangle(4, 5);
        Shape goodSquare = new GoodSquare(4);

        System.out.println("Rectangle area: " + rectangle.area());
        System.out.println("Square area: " + goodSquare.area());
        // Both work correctly as Shape — no surprises!

        // ═══════════════════════════════════════════════════════
        // BAD: Bird hierarchy with penguin
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== BAD: Flying Bird + Penguin ===");
        BadBird sparrow = new Sparrow();
        sparrow.fly();  // works

        BadBird penguin = new Penguin();
        // penguin.fly();  // throws exception! LSP VIOLATED!
        System.out.println("Can't call penguin.fly() — would throw exception!");

        // ═══════════════════════════════════════════════════════
        // GOOD: Separate interfaces for separate capabilities
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GOOD: Interface Segregation Fixes LSP ===");

        GoodBird sparrow2 = new GoodSparrow();
        sparrow2.eat();
        if (sparrow2 instanceof Flyable) {
            ((Flyable) sparrow2).fly();
        }

        GoodBird penguin2 = new GoodPenguin();
        penguin2.eat();
        if (penguin2 instanceof Swimmable) {
            ((Swimmable) penguin2).swim();
        }

        // Every GoodBird can eat. Only Flyable birds can fly. Clean!
    }

    // This test PROVES the LSP violation
    static void testRectangleArea(BadRectangle r) {
        r.setWidth(5);
        r.setHeight(4);
        int expected = 20;
        int actual = r.area();
        System.out.println("  Test: setWidth(5), setHeight(4) → area=" + actual
                + " (expected " + expected + ") → " + (actual == expected ? "PASS" : "FAIL!"));
    }
}

// ═══════════════════════════════════════════════════════════════
// BAD: Classic Rectangle-Square LSP Violation
// ═══════════════════════════════════════════════════════════════
class BadRectangle {
    protected int width, height;

    public BadRectangle(int w, int h) { width = w; height = h; }

    public void setWidth(int w)  { width = w; }
    public void setHeight(int h) { height = h; }
    public int area() { return width * height; }
}

class BadSquare extends BadRectangle {
    public BadSquare(int side) { super(side, side); }

    @Override
    public void setWidth(int w)  { width = w; height = w; }  // FORCES side sync
    @Override
    public void setHeight(int h) { width = h; height = h; }  // FORCES side sync
    // setWidth changes BOTH dimensions — violates parent's contract!
}

// ═══════════════════════════════════════════════════════════════
// GOOD: Immutable shapes — LSP preserved
// ═══════════════════════════════════════════════════════════════
interface Shape {
    double area();
}

class GoodRectangle implements Shape {
    private final double width, height;

    public GoodRectangle(double w, double h) { width = w; height = h; }

    @Override
    public double area() { return width * height; }
}

class GoodSquare implements Shape {
    private final double side;

    public GoodSquare(double side) { this.side = side; }

    @Override
    public double area() { return side * side; }
}
// No inheritance between Rectangle and Square.
// Both independently implement Shape. No surprises.

// ═══════════════════════════════════════════════════════════════
// BAD: Not all birds fly!
// ═══════════════════════════════════════════════════════════════
class BadBird {
    public void fly() { System.out.println("  Flying..."); }
    public void eat() { System.out.println("  Eating..."); }
}

class Sparrow extends BadBird {
    @Override public void fly() { System.out.println("  🐦 Sparrow flying high!"); }
}

class Penguin extends BadBird {
    @Override
    public void fly() {
        throw new UnsupportedOperationException("Penguins can't fly!");
        // This BREAKS any code that calls badBird.fly() polymorphically!
    }
}

// ═══════════════════════════════════════════════════════════════
// GOOD: Separate capabilities via interfaces
// ═══════════════════════════════════════════════════════════════
abstract class GoodBird {
    public abstract void eat();
}

interface Flyable {
    void fly();
}

interface Swimmable {
    void swim();
}

class GoodSparrow extends GoodBird implements Flyable {
    @Override public void eat() { System.out.println("  🐦 Sparrow eating seeds..."); }
    @Override public void fly() { System.out.println("  🐦 Sparrow soaring!"); }
}

class GoodPenguin extends GoodBird implements Swimmable {
    @Override public void eat() { System.out.println("  🐧 Penguin eating fish..."); }
    @Override public void swim() { System.out.println("  🐧 Penguin swimming!"); }
    // No fly() method — because penguins CAN'T fly. Honest design.
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Subclass must honor ALL contracts of parent.
 * ✦ If overriding a method to throw exception → LSP violation.
 * ✦ If overriding a method to do nothing → LSP violation.
 * ✦ Immutable objects avoid many LSP issues (no setters to break).
 * ✦ "Rectangle-Square problem" → don't inherit, both implement Shape.
 * ✦ "Penguin-Bird problem" → separate fly() into Flyable interface.
 * ✦ LSP and Interface Segregation Principle often work together.
 *
 * COMPILE & RUN:
 *   javac LiskovSubstitution.java && java LiskovSubstitution
 */
