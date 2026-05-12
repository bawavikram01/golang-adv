/*
 * =============================================================
 * CREATIONAL PATTERN 4: PROTOTYPE
 * =============================================================
 *
 * INTENT: Create new objects by CLONING existing ones.
 *
 * USE WHEN:
 *   - Object creation is expensive (DB lookup, network call)
 *   - You need copies with slight modifications
 *   - You want to avoid subclasses of a creator class
 *
 * KEY: Deep copy vs Shallow copy
 *   - Shallow: copies references (changes reflect in both)
 *   - Deep: copies the entire object tree (independent copies)
 */

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class PrototypePattern {

    public static void main(String[] args) throws CloneNotSupportedException {

        // ═══════════════════════════════════════════════════════
        // Basic Prototype: GameCharacter
        // ═══════════════════════════════════════════════════════
        System.out.println("=== BASIC PROTOTYPE ===");

        GameCharacter warrior = new GameCharacter("Warrior", 100, 80, 30);
        warrior.addItem("Sword");
        warrior.addItem("Shield");

        // Clone instead of creating from scratch
        GameCharacter warrior2 = warrior.clone();
        warrior2.setName("Warrior-2");
        warrior2.addItem("Potion");  // only warrior2 gets this

        System.out.println("Original:  " + warrior);
        System.out.println("Clone:     " + warrior2);
        System.out.println("Independent? " + (warrior.getItems().size() != warrior2.getItems().size()));

        // ═══════════════════════════════════════════════════════
        // Prototype Registry: Pre-configured templates
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== PROTOTYPE REGISTRY ===");

        CharacterRegistry registry = new CharacterRegistry();

        // Create 5 mages from a prototype — fast!
        for (int i = 1; i <= 3; i++) {
            GameCharacter mage = registry.get("mage");
            mage.setName("Mage-" + i);
            System.out.println("  Created: " + mage);
        }

        // ═══════════════════════════════════════════════════════
        // Real-world: Document Template System
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== DOCUMENT TEMPLATES ===");

        DocumentTemplate reportTemplate = new DocumentTemplate(
                "Monthly Report",
                "Header: Company LLC\nDate: {date}\n\nBody: {content}",
                "Times New Roman", 12
        );
        reportTemplate.addSection("Summary");
        reportTemplate.addSection("Findings");
        reportTemplate.addSection("Recommendations");

        // Clone and customize
        DocumentTemplate janReport = reportTemplate.clone();
        janReport.setTitle("January Report");
        janReport.addSection("January-Specific Data");

        DocumentTemplate febReport = reportTemplate.clone();
        febReport.setTitle("February Report");

        System.out.println("Template:  " + reportTemplate);
        System.out.println("January:   " + janReport);
        System.out.println("February:  " + febReport);
    }
}

// ═══════════════════════════════════════════════════════════════
// Cloneable Game Character with DEEP COPY
// ═══════════════════════════════════════════════════════════════
class GameCharacter implements Cloneable {
    private String name;
    private int health;
    private int attack;
    private int defense;
    private List<String> items;  // mutable field — needs deep copy!

    public GameCharacter(String name, int health, int attack, int defense) {
        this.name = name;
        this.health = health;
        this.attack = attack;
        this.defense = defense;
        this.items = new ArrayList<>();
    }

    // Copy constructor (alternative to clone)
    public GameCharacter(GameCharacter other) {
        this.name = other.name;
        this.health = other.health;
        this.attack = other.attack;
        this.defense = other.defense;
        this.items = new ArrayList<>(other.items);  // deep copy of list
    }

    @Override
    public GameCharacter clone() {
        return new GameCharacter(this);  // delegate to copy constructor
    }

    public void setName(String name) { this.name = name; }
    public void addItem(String item) { this.items.add(item); }
    public List<String> getItems() { return items; }

    @Override
    public String toString() {
        return name + "{hp=" + health + ", atk=" + attack + ", def=" + defense + ", items=" + items + "}";
    }
}

// ═══════════════════════════════════════════════════════════════
// Prototype Registry — stores pre-configured prototypes
// ═══════════════════════════════════════════════════════════════
class CharacterRegistry {
    private Map<String, GameCharacter> prototypes = new HashMap<>();

    public CharacterRegistry() {
        // Pre-configure prototypes
        GameCharacter warrior = new GameCharacter("Warrior", 100, 80, 50);
        warrior.addItem("Sword");
        warrior.addItem("Shield");
        prototypes.put("warrior", warrior);

        GameCharacter mage = new GameCharacter("Mage", 60, 100, 20);
        mage.addItem("Staff");
        mage.addItem("Spellbook");
        prototypes.put("mage", mage);

        GameCharacter archer = new GameCharacter("Archer", 75, 90, 30);
        archer.addItem("Bow");
        archer.addItem("Quiver");
        prototypes.put("archer", archer);
    }

    public GameCharacter get(String type) {
        GameCharacter prototype = prototypes.get(type);
        if (prototype == null) {
            throw new IllegalArgumentException("Unknown character type: " + type);
        }
        return prototype.clone();  // CLONE, never return the original!
    }
}

// ═══════════════════════════════════════════════════════════════
// Document Template — Deep copy with nested collections
// ═══════════════════════════════════════════════════════════════
class DocumentTemplate implements Cloneable {
    private String title;
    private String templateBody;
    private String font;
    private int fontSize;
    private List<String> sections;

    public DocumentTemplate(String title, String templateBody, String font, int fontSize) {
        this.title = title;
        this.templateBody = templateBody;
        this.font = font;
        this.fontSize = fontSize;
        this.sections = new ArrayList<>();
    }

    @Override
    public DocumentTemplate clone() {
        DocumentTemplate copy = new DocumentTemplate(title, templateBody, font, fontSize);
        copy.sections = new ArrayList<>(this.sections);  // deep copy
        return copy;
    }

    public void setTitle(String title) { this.title = title; }
    public void addSection(String section) { this.sections.add(section); }

    @Override
    public String toString() {
        return "Doc{'" + title + "', sections=" + sections + "}";
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Prototype = create by CLONING, not by `new`.
 * ✦ Use when creation is expensive or you need configured templates.
 * ✦ ALWAYS do DEEP COPY of mutable fields (lists, maps, objects).
 * ✦ Copy constructor > Cloneable interface (Joshua Bloch's advice).
 * ✦ Prototype Registry = Map<String, Prototype> for templates.
 *
 * COMPILE & RUN:
 *   javac PrototypePattern.java && java PrototypePattern
 */
