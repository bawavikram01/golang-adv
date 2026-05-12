/*
 * =============================================================
 * UML CLASS DIAGRAMS — The Language of LLD
 * =============================================================
 *
 * In LLD interviews, you will DRAW before you code.
 * This file teaches you UML class diagram notation using
 * ASCII art and Java code equivalents.
 *
 * ELEMENTS:
 *   1. Class box (name, fields, methods)
 *   2. Relationships (association, aggregation, composition, inheritance)
 *   3. Multiplicity (1, 0..1, *, 1..*)
 *   4. Interfaces and abstract classes
 */

import java.util.*;

public class UMLClassDiagrams {

    public static void main(String[] args) {

        System.out.println("=== UML CLASS DIAGRAM NOTATION ===");
        System.out.println();

        // ═══════════════════════════════════════════════════════
        // 1. CLASS BOX
        // ═══════════════════════════════════════════════════════
        System.out.println("1. CLASS BOX:");
        System.out.println("  ┌─────────────────────────┐");
        System.out.println("  │      <<ClassName>>       │");
        System.out.println("  ├─────────────────────────┤");
        System.out.println("  │ - privateField: Type     │  - = private");
        System.out.println("  │ + publicField: Type      │  + = public");
        System.out.println("  │ # protectedField: Type   │  # = protected");
        System.out.println("  │ ~ packageField: Type     │  ~ = package-private");
        System.out.println("  ├─────────────────────────┤");
        System.out.println("  │ + publicMethod(): void   │");
        System.out.println("  │ - privateMethod(): int   │");
        System.out.println("  │ + staticMethod(): String │  underlined = static");
        System.out.println("  └─────────────────────────┘");

        // ═══════════════════════════════════════════════════════
        // 2. RELATIONSHIPS
        // ═══════════════════════════════════════════════════════
        System.out.println();
        System.out.println("2. RELATIONSHIPS:");
        System.out.println();
        System.out.println("  a) INHERITANCE (IS-A): solid line + closed arrow");
        System.out.println("     Dog ──────▷ Animal");
        System.out.println("     Java: class Dog extends Animal");
        System.out.println();
        System.out.println("  b) INTERFACE IMPLEMENTATION: dashed line + closed arrow");
        System.out.println("     ArrayList -----▷ List");
        System.out.println("     Java: class ArrayList implements List");
        System.out.println();
        System.out.println("  c) ASSOCIATION (USES): solid line, open arrow");
        System.out.println("     Teacher ──────> Student");
        System.out.println("     Java: class Teacher { Student student; }");
        System.out.println("     Meaning: Teacher KNOWS about Student");
        System.out.println();
        System.out.println("  d) AGGREGATION (HAS-A, weak): hollow diamond");
        System.out.println("     Department ◇──── Professor");
        System.out.println("     Java: class Department { List<Professor> profs; }");
        System.out.println("     Meaning: Prof can EXIST without Department");
        System.out.println();
        System.out.println("  e) COMPOSITION (HAS-A, strong): filled diamond");
        System.out.println("     House ◆──── Room");
        System.out.println("     Java: class House { List<Room> rooms; }");
        System.out.println("     Meaning: Room CANNOT exist without House");
        System.out.println();
        System.out.println("  f) DEPENDENCY (USES temporarily): dashed arrow");
        System.out.println("     Client -----> Service");
        System.out.println("     Java: void method(Service s) { s.doWork(); }");
        System.out.println("     Meaning: Client uses Service but doesn't hold reference");

        // ═══════════════════════════════════════════════════════
        // 3. MULTIPLICITY
        // ═══════════════════════════════════════════════════════
        System.out.println();
        System.out.println("3. MULTIPLICITY:");
        System.out.println("  ┌────────┬─────────────────────────┐");
        System.out.println("  │ Symbol │ Meaning                 │");
        System.out.println("  ├────────┼─────────────────────────┤");
        System.out.println("  │ 1      │ Exactly one             │");
        System.out.println("  │ 0..1   │ Zero or one (optional)  │");
        System.out.println("  │ *      │ Zero or more            │");
        System.out.println("  │ 1..*   │ One or more             │");
        System.out.println("  │ 3..5   │ Three to five           │");
        System.out.println("  └────────┴─────────────────────────┘");
        System.out.println();
        System.out.println("  Example:");
        System.out.println("    Order 1 ◆──── * LineItem");
        System.out.println("    (One order has many line items)");
        System.out.println("    (Each line item belongs to exactly one order)");

        // ═══════════════════════════════════════════════════════
        // 4. ABSTRACT AND INTERFACE
        // ═══════════════════════════════════════════════════════
        System.out.println();
        System.out.println("4. ABSTRACT CLASS & INTERFACE:");
        System.out.println("  ┌──────────────────────────┐");
        System.out.println("  │  <<abstract>> Shape       │  (or use italic name)");
        System.out.println("  ├──────────────────────────┤");
        System.out.println("  │ # color: String           │");
        System.out.println("  ├──────────────────────────┤");
        System.out.println("  │ + area(): double {abstract}│");
        System.out.println("  │ + display(): void         │");
        System.out.println("  └──────────────────────────┘");
        System.out.println();
        System.out.println("  ┌──────────────────────────┐");
        System.out.println("  │  <<interface>> Drawable    │");
        System.out.println("  ├──────────────────────────┤");
        System.out.println("  │ + draw(): void             │");
        System.out.println("  │ + resize(int): void        │");
        System.out.println("  └──────────────────────────┘");

        // ═══════════════════════════════════════════════════════
        // 5. FULL EXAMPLE: Parking Lot UML
        // ═══════════════════════════════════════════════════════
        System.out.println();
        System.out.println("5. FULL EXAMPLE — Parking Lot:");
        System.out.println();
        System.out.println("  ┌────────────────────┐     ┌──────────────────┐");
        System.out.println("  │ <<singleton>>       │     │ ParkingFloor     │");
        System.out.println("  │ ParkingLot          │     ├──────────────────┤");
        System.out.println("  ├────────────────────┤ 1  *│ - floorId: String│");
        System.out.println("  │ - instance: Parking │◆────│ - spots: List    │");
        System.out.println("  │ - floors: List      │     ├──────────────────┤");
        System.out.println("  │ - tickets: Map      │     │ + findSpot()     │");
        System.out.println("  ├────────────────────┤     └───────┬──────────┘");
        System.out.println("  │ + getInstance()     │             │ 1..*");
        System.out.println("  │ + parkVehicle()     │     ┌───────▼──────────┐");
        System.out.println("  │ + unparkVehicle()   │     │ ParkingSpot      │");
        System.out.println("  └────────────────────┘     ├──────────────────┤");
        System.out.println("                              │ - spotId: String │");
        System.out.println("  ┌──────────────────┐       │ - size: SpotSize │");
        System.out.println("  │ Vehicle           │       │ - vehicle: Vehicl│");
        System.out.println("  ├──────────────────┤       ├──────────────────┤");
        System.out.println("  │ - plate: String   │ 0..1 │ + isAvailable()  │");
        System.out.println("  │ - type: VehicleType◁─────│ + park()         │");
        System.out.println("  └──────────────────┘       │ + unpark()       │");
        System.out.println("                              └──────────────────┘");
        System.out.println();
        System.out.println("  ┌──────────────────────┐");
        System.out.println("  │ <<interface>>         │");
        System.out.println("  │ PricingStrategy       │");
        System.out.println("  ├──────────────────────┤");
        System.out.println("  │ + calculateFee(): dbl │           ┌─────────────┐");
        System.out.println("  └──────────┬───────────┘           │ <<enum>>    │");
        System.out.println("    ▲  ▲                               │ VehicleType │");
        System.out.println("    │  │                               ├─────────────┤");
        System.out.println("    │  └── FlatRatePricing             │ BIKE        │");
        System.out.println("    └───── HourlyPricing               │ CAR         │");
        System.out.println("                                       │ TRUCK       │");
        System.out.println("                                       └─────────────┘");

        // ═══════════════════════════════════════════════════════
        // 6. INTERVIEW APPROACH
        // ═══════════════════════════════════════════════════════
        System.out.println();
        System.out.println("6. LLD INTERVIEW APPROACH (step by step):");
        System.out.println("  ┌─────────────────────────────────────────────┐");
        System.out.println("  │ 1. CLARIFY requirements (ask questions!)   │");
        System.out.println("  │ 2. IDENTIFY entities/nouns → classes       │");
        System.out.println("  │ 3. IDENTIFY actions/verbs → methods        │");
        System.out.println("  │ 4. DEFINE relationships → UML arrows       │");
        System.out.println("  │ 5. APPLY design patterns where appropriate │");
        System.out.println("  │ 6. WRITE code for core logic               │");
        System.out.println("  │ 7. DISCUSS trade-offs and extensibility    │");
        System.out.println("  └─────────────────────────────────────────────┘");
    }
}

/*
 * INTERVIEW TIPS:
 * ─────────────────────────────────────────────────────────────
 * ✦ ALWAYS draw the class diagram FIRST before coding.
 * ✦ Start with 3-4 core classes, then expand.
 * ✦ Use enums for fixed types (VehicleType, SpotSize, OrderStatus).
 * ✦ Mark access modifiers (private fields, public methods).
 * ✦ Show relationships clearly (composition vs aggregation).
 * ✦ Call out design patterns you're using ("I'll use Strategy here").
 *
 * AGGREGATION vs COMPOSITION:
 *   ✦ Aggregation: "has-a" (weak) — child CAN exist alone
 *     University ◇── Professor (professor exists without university)
 *   ✦ Composition: "has-a" (strong) — child CANNOT exist alone
 *     House ◆── Room (room doesn't exist without house)
 *
 * COMPILE & RUN:
 *   javac UMLClassDiagrams.java && java UMLClassDiagrams
 */
