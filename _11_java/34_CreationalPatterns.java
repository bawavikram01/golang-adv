/*
 * ============================================================
 *  CHAPTER 34: DESIGN PATTERNS — CREATIONAL
 * ============================================================
 *  Design Patterns = proven solutions to common problems.
 *  Gang of Four (GoF) categorized 23 patterns in 3 groups:
 *    CREATIONAL  — how objects are created (this chapter)
 *    STRUCTURAL  — how objects are composed (Chapter 35)
 *    BEHAVIORAL  — how objects communicate (Chapter 36)
 *
 *  CREATIONAL PATTERNS:
 *    1. Singleton       — one instance only
 *    2. Factory Method  — delegate creation to subclasses
 *    3. Abstract Factory — family of related objects
 *    4. Builder         — step-by-step construction
 *    5. Prototype       — clone existing objects
 * ============================================================
 */

import java.util.*;

public class Chapter34_CreationalPatterns {

    // ========================================================
    // 1. SINGLETON — ensure only one instance exists
    // ========================================================

    // Thread-safe, lazy initialization (Bill Pugh Singleton)
    static class Database {
        private String connectionInfo = "Connected to DB";

        private Database() {
            System.out.println("    Database instance created");
        }

        // Inner class not loaded until getInstance() called
        private static class Holder {
            private static final Database INSTANCE = new Database();
        }

        public static Database getInstance() {
            return Holder.INSTANCE;
        }

        public String query(String sql) {
            return "Result of: " + sql;
        }
    }

    // Enum Singleton (Joshua Bloch's recommended approach)
    enum Logger {
        INSTANCE;

        public void log(String message) {
            System.out.println("    [LOG] " + message);
        }
    }

    // ========================================================
    // 2. FACTORY METHOD — let subclasses decide what to create
    // ========================================================

    // Product interface
    interface Notification {
        void send(String message);
    }

    // Concrete products
    static class EmailNotification implements Notification {
        @Override
        public void send(String msg) { System.out.println("    📧 Email: " + msg); }
    }

    static class SMSNotification implements Notification {
        @Override
        public void send(String msg) { System.out.println("    📱 SMS: " + msg); }
    }

    static class PushNotification implements Notification {
        @Override
        public void send(String msg) { System.out.println("    🔔 Push: " + msg); }
    }

    // Factory
    static class NotificationFactory {
        public static Notification create(String type) {
            switch (type.toLowerCase()) {
                case "email": return new EmailNotification();
                case "sms":   return new SMSNotification();
                case "push":  return new PushNotification();
                default: throw new IllegalArgumentException("Unknown type: " + type);
            }
        }
    }

    // ========================================================
    // 3. ABSTRACT FACTORY — family of related objects
    // ========================================================

    // Abstract products
    interface Button { void render(); }
    interface TextField { void render(); }

    // Concrete products — Light theme
    static class LightButton implements Button {
        @Override
        public void render() { System.out.println("    [Light Button]"); }
    }
    static class LightTextField implements TextField {
        @Override
        public void render() { System.out.println("    [Light TextField]"); }
    }

    // Concrete products — Dark theme
    static class DarkButton implements Button {
        @Override
        public void render() { System.out.println("    [Dark Button]"); }
    }
    static class DarkTextField implements TextField {
        @Override
        public void render() { System.out.println("    [Dark TextField]"); }
    }

    // Abstract Factory
    interface UIFactory {
        Button createButton();
        TextField createTextField();
    }

    static class LightThemeFactory implements UIFactory {
        @Override
        public Button createButton() { return new LightButton(); }
        @Override
        public TextField createTextField() { return new LightTextField(); }
    }

    static class DarkThemeFactory implements UIFactory {
        @Override
        public Button createButton() { return new DarkButton(); }
        @Override
        public TextField createTextField() { return new DarkTextField(); }
    }

    // ========================================================
    // 4. BUILDER — step-by-step construction
    // ========================================================

    static class HttpRequest {
        private final String method;
        private final String url;
        private final Map<String, String> headers;
        private final String body;
        private final int timeout;

        private HttpRequest(Builder builder) {
            this.method = builder.method;
            this.url = builder.url;
            this.headers = Collections.unmodifiableMap(builder.headers);
            this.body = builder.body;
            this.timeout = builder.timeout;
        }

        @Override
        public String toString() {
            return method + " " + url + " headers=" + headers +
                " body=" + (body != null ? body : "none") + " timeout=" + timeout;
        }

        // Static inner Builder class
        static class Builder {
            private final String method;  // required
            private final String url;     // required
            private Map<String, String> headers = new HashMap<>();
            private String body;
            private int timeout = 30;

            Builder(String method, String url) {
                this.method = method;
                this.url = url;
            }

            Builder header(String key, String value) {
                headers.put(key, value);
                return this;  // fluent API
            }

            Builder body(String body) {
                this.body = body;
                return this;
            }

            Builder timeout(int seconds) {
                this.timeout = seconds;
                return this;
            }

            HttpRequest build() {
                return new HttpRequest(this);
            }
        }
    }

    // ========================================================
    // 5. PROTOTYPE — clone existing objects
    // ========================================================

    static class GameCharacter implements Cloneable {
        String name;
        int health;
        int attack;
        List<String> inventory;  // mutable — needs deep copy

        GameCharacter(String name, int health, int attack) {
            this.name = name;
            this.health = health;
            this.attack = attack;
            this.inventory = new ArrayList<>();
        }

        // Deep clone
        @Override
        public GameCharacter clone() {
            try {
                GameCharacter copy = (GameCharacter) super.clone();
                copy.inventory = new ArrayList<>(this.inventory);  // deep copy list
                return copy;
            } catch (CloneNotSupportedException e) {
                throw new AssertionError();
            }
        }

        @Override
        public String toString() {
            return name + " [HP=" + health + " ATK=" + attack + " INV=" + inventory + "]";
        }
    }

    // Prototype Registry
    static class CharacterRegistry {
        private Map<String, GameCharacter> prototypes = new HashMap<>();

        void register(String key, GameCharacter prototype) {
            prototypes.put(key, prototype);
        }

        GameCharacter create(String key) {
            GameCharacter prototype = prototypes.get(key);
            if (prototype == null) throw new IllegalArgumentException("Unknown: " + key);
            return prototype.clone();
        }
    }

    // ========================================================
    // MAIN — demonstrate all patterns
    // ========================================================

    public static void main(String[] args) {

        // --- 1. Singleton ---
        System.out.println("=== SINGLETON ===\n");
        Database db1 = Database.getInstance();
        Database db2 = Database.getInstance();
        System.out.println("  Same instance? " + (db1 == db2));
        System.out.println("  " + db1.query("SELECT * FROM users"));

        Logger.INSTANCE.log("Application started");
        Logger.INSTANCE.log("Enum singleton is thread-safe and serialization-safe");

        // --- 2. Factory Method ---
        System.out.println("\n=== FACTORY METHOD ===\n");
        Notification email = NotificationFactory.create("email");
        Notification sms = NotificationFactory.create("sms");
        Notification push = NotificationFactory.create("push");

        email.send("Welcome!");
        sms.send("Your code is 1234");
        push.send("New message received");

        // --- 3. Abstract Factory ---
        System.out.println("\n=== ABSTRACT FACTORY ===\n");
        String theme = "dark";
        UIFactory factory = theme.equals("dark") ?
            new DarkThemeFactory() : new LightThemeFactory();

        Button button = factory.createButton();
        TextField field = factory.createTextField();
        System.out.println("  Theme: " + theme);
        button.render();
        field.render();

        // --- 4. Builder ---
        System.out.println("\n=== BUILDER ===\n");

        HttpRequest getRequest = new HttpRequest.Builder("GET", "/api/users")
            .header("Accept", "application/json")
            .timeout(10)
            .build();
        System.out.println("  " + getRequest);

        HttpRequest postRequest = new HttpRequest.Builder("POST", "/api/users")
            .header("Content-Type", "application/json")
            .header("Authorization", "Bearer token123")
            .body("{\"name\": \"Alice\"}")
            .timeout(30)
            .build();
        System.out.println("  " + postRequest);

        // --- 5. Prototype ---
        System.out.println("\n=== PROTOTYPE ===\n");

        // Create template
        GameCharacter warrior = new GameCharacter("Warrior", 100, 20);
        warrior.inventory.add("Sword");
        warrior.inventory.add("Shield");

        // Clone and customize
        GameCharacter warrior2 = warrior.clone();
        warrior2.name = "Warrior-2";
        warrior2.inventory.add("Potion");

        System.out.println("  Original: " + warrior);
        System.out.println("  Clone:    " + warrior2);
        System.out.println("  Independent inventories? " +
            (warrior.inventory != warrior2.inventory));

        // Prototype Registry
        System.out.println("\n  --- Prototype Registry ---");
        CharacterRegistry registry = new CharacterRegistry();

        GameCharacter mageTemplate = new GameCharacter("Mage", 60, 35);
        mageTemplate.inventory.add("Staff");
        registry.register("mage", mageTemplate);
        registry.register("warrior", warrior);

        GameCharacter newMage = registry.create("mage");
        newMage.name = "Gandalf";
        System.out.println("  From registry: " + newMage);

        // --- Summary ---
        System.out.println("\n=== WHEN TO USE ===");
        System.out.println("  Singleton:        Global config, logger, connection pool");
        System.out.println("  Factory Method:   Decouple creation, switch implementations");
        System.out.println("  Abstract Factory: UI themes, cross-platform components");
        System.out.println("  Builder:          Complex objects with many optional params");
        System.out.println("  Prototype:        Expensive creation, template-based cloning");

        System.out.println("\n✓ Creational Patterns Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Implement a thread-safe Singleton with double-checked locking.
 * 2. Create a VehicleFactory that produces Car, Bike, Truck objects.
 * 3. Build a Pizza class using Builder pattern with toppings list.
 * 4. Implement Prototype for a Document class with nested objects.
 *
 * NEXT: Chapter 35 — Design Patterns: Structural
 */
