/*
 * =============================================================
 * MODULE 6: COMPOSITION vs INHERITANCE — "has-a" vs "is-a"
 * =============================================================
 *
 * GOLDEN RULE: Favor Composition over Inheritance.
 *
 * Inheritance:  "A Dog IS-A Animal"       → tight coupling
 * Composition:  "A Car HAS-A Engine"      → loose coupling
 *
 * WHY COMPOSITION WINS:
 *   1. More flexible — swap behaviors at runtime
 *   2. Avoids fragile base class problem
 *   3. No diamond problem
 *   4. Easier to test (inject mocks)
 *   5. Changes in composed class don't break yours
 *
 * WHEN TO USE INHERITANCE:
 *   - True "is-a" relationship
 *   - You control both parent and child
 *   - You need to override specific behaviors
 */

import java.util.List;
import java.util.ArrayList;

public class CompositionVsInheritance {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // BAD: Inheritance approach — rigid, fragile
        // ═══════════════════════════════════════════════════════
        System.out.println("=== BAD: Inheritance Approach ===");

        RobotDogInheritance roboDog = new RobotDogInheritance();
        roboDog.move();
        roboDog.attack();
        // What if we want a robot that flies instead of walks?
        // We'd need a new class hierarchy... explosion of subclasses!

        // ═══════════════════════════════════════════════════════
        // GOOD: Composition approach — flexible, swappable
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GOOD: Composition Approach ===");

        // Robot that walks and uses a sword
        Robot warrior = new Robot("Warrior",
                new WalkMovement(),
                new SwordAttack());
        warrior.performMove();
        warrior.performAttack();

        // Robot that flies and shoots lasers — just swap components!
        Robot drone = new Robot("Drone",
                new FlyMovement(),
                new LaserAttack());
        drone.performMove();
        drone.performAttack();

        // ─── RUNTIME flexibility: change behavior on the fly ───
        System.out.println("\n=== RUNTIME BEHAVIOR CHANGE ===");
        System.out.println("Warrior upgrades to flying!");
        warrior.setMovementStrategy(new FlyMovement());
        warrior.performMove();

        // ═══════════════════════════════════════════════════════
        // REAL-WORLD EXAMPLE: Building a Computer
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== COMPOSITION: Building a Computer ===");

        CPU cpu = new CPU("AMD Ryzen 9", 12);
        RAM ram = new RAM("DDR5", 32);
        Storage ssd = new Storage("NVMe SSD", 1000);

        Computer myPC = new Computer("Gaming Rig", cpu, ram, ssd);
        myPC.showSpecs();
        myPC.boot();

        // Upgrade RAM? Just compose a new one!
        System.out.println("\n--- Upgrading RAM ---");
        myPC.setRam(new RAM("DDR5", 64));
        myPC.showSpecs();
    }
}

// ═══════════════════════════════════════════════════════════════
// BAD: Inheritance Hierarchy Explosion
// ═══════════════════════════════════════════════════════════════
// Imagine: WalkingRobot, FlyingRobot, SwimmingRobot,
//          WalkingSwordRobot, FlyingLaserRobot, SwimmingSwordRobot...
//          → 3 movements × 3 attacks = 9 classes? Madness!

class RobotBaseInheritance {
    public void move() { System.out.println("  Moving..."); }
    public void attack() { System.out.println("  Attacking..."); }
}

class RobotDogInheritance extends RobotBaseInheritance {
    @Override
    public void move() { System.out.println("  🐕 Walking on 4 legs"); }
    @Override
    public void attack() { System.out.println("  🦷 Biting!"); }
}

// ═══════════════════════════════════════════════════════════════
// GOOD: Composition with Strategy Interfaces
// ═══════════════════════════════════════════════════════════════

// ─── Interfaces for behaviors ───
interface MovementStrategy {
    void move();
}

interface AttackStrategy {
    void attack();
}

// ─── Concrete strategies (plug-and-play) ───
class WalkMovement implements MovementStrategy {
    @Override
    public void move() { System.out.println("  🚶 Walking..."); }
}

class FlyMovement implements MovementStrategy {
    @Override
    public void move() { System.out.println("  ✈️ Flying..."); }
}

class SwimMovement implements MovementStrategy {
    @Override
    public void move() { System.out.println("  🏊 Swimming..."); }
}

class SwordAttack implements AttackStrategy {
    @Override
    public void attack() { System.out.println("  ⚔️ Sword slash!"); }
}

class LaserAttack implements AttackStrategy {
    @Override
    public void attack() { System.out.println("  🔫 Laser beam!"); }
}

// ─── Robot COMPOSES behaviors instead of inheriting them ───
class Robot {
    private String name;
    private MovementStrategy movement;
    private AttackStrategy attack;

    public Robot(String name, MovementStrategy movement, AttackStrategy attack) {
        this.name = name;
        this.movement = movement;
        this.attack = attack;
    }

    public void performMove() {
        System.out.print("  " + name + ": ");
        movement.move();
    }

    public void performAttack() {
        System.out.print("  " + name + ": ");
        attack.attack();
    }

    // Can swap behaviors at RUNTIME!
    public void setMovementStrategy(MovementStrategy movement) {
        this.movement = movement;
    }

    public void setAttackStrategy(AttackStrategy attack) {
        this.attack = attack;
    }
}

// ═══════════════════════════════════════════════════════════════
// REAL-WORLD: Computer HAS-A CPU, RAM, Storage
// ═══════════════════════════════════════════════════════════════
class CPU {
    private String model;
    private int cores;
    public CPU(String model, int cores) { this.model = model; this.cores = cores; }
    public void process() { System.out.println("  CPU [" + model + "] processing with " + cores + " cores..."); }
    @Override public String toString() { return model + " (" + cores + " cores)"; }
}

class RAM {
    private String type;
    private int sizeGB;
    public RAM(String type, int sizeGB) { this.type = type; this.sizeGB = sizeGB; }
    public void load() { System.out.println("  RAM [" + type + "] loading " + sizeGB + "GB..."); }
    @Override public String toString() { return sizeGB + "GB " + type; }
}

class Storage {
    private String type;
    private int sizeGB;
    public Storage(String type, int sizeGB) { this.type = type; this.sizeGB = sizeGB; }
    public void read() { System.out.println("  Storage [" + type + "] reading " + sizeGB + "GB..."); }
    @Override public String toString() { return sizeGB + "GB " + type; }
}

class Computer {
    private String name;
    private CPU cpu;       // HAS-A
    private RAM ram;       // HAS-A
    private Storage storage; // HAS-A

    public Computer(String name, CPU cpu, RAM ram, Storage storage) {
        this.name = name;
        this.cpu = cpu;
        this.ram = ram;
        this.storage = storage;
    }

    public void boot() {
        System.out.println("Booting " + name + "...");
        storage.read();
        ram.load();
        cpu.process();
        System.out.println("  ✓ " + name + " is ready!");
    }

    public void showSpecs() {
        System.out.println(name + " specs: CPU=" + cpu + ", RAM=" + ram + ", Storage=" + storage);
    }

    public void setRam(RAM ram) { this.ram = ram; }  // Easy upgrade!
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ COMPOSITION = object contains other objects (HAS-A).
 * ✦ INHERITANCE = object is a specialized version of parent (IS-A).
 *
 * ✦ Composition benefits:
 *   - Swap behaviors at runtime (strategy pattern)
 *   - No class explosion (M movements × N attacks = M+N classes, not M×N)
 *   - Loose coupling — easy to test and modify
 *
 * ✦ Use inheritance when:
 *   - There's a genuine IS-A relationship
 *   - You want to leverage polymorphism
 *   - The hierarchy is shallow and stable
 *
 * ✦ This module preview's the STRATEGY PATTERN (behavioral design pattern).
 *
 * COMPILE & RUN:
 *   javac CompositionVsInheritance.java && java CompositionVsInheritance
 */
