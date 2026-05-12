/*
 * =============================================================
 * CREATIONAL PATTERN 3: BUILDER
 * =============================================================
 *
 * INTENT: Construct complex objects step-by-step.
 *         Same construction process can create different representations.
 *
 * USE WHEN:
 *   - Constructor has too many parameters (telescoping constructor)
 *   - Some parameters are optional
 *   - Object construction requires multiple steps
 *   - You need immutable objects with many fields
 *
 * REAL EXAMPLES: StringBuilder, HttpRequest.Builder, Lombok @Builder
 */

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public class BuilderPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // BAD: Telescoping constructor — unreadable!
        // ═══════════════════════════════════════════════════════
        System.out.println("=== BAD: Telescoping Constructor ===");
        // What does each parameter mean?!
        BadUser bad = new BadUser("Alice", "alice@mail.com", 25, "123 St", "NYC", "USA", "10001", true, false);
        System.out.println("  Created: " + bad);

        // ═══════════════════════════════════════════════════════
        // GOOD: Builder pattern — readable, flexible, immutable
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GOOD: Builder Pattern ===");

        User alice = new User.Builder("Alice", "alice@mail.com")
                .age(25)
                .address("123 Main St")
                .city("New York")
                .country("USA")
                .zipCode("10001")
                .newsletter(true)
                .build();

        System.out.println(alice);

        // Only required fields
        User bob = new User.Builder("Bob", "bob@mail.com")
                .build();

        System.out.println(bob);

        // ═══════════════════════════════════════════════════════
        // REAL-WORLD: Building an HTTP Request
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== REAL-WORLD: HTTP Request Builder ===");

        HttpRequest getRequest = new HttpRequest.Builder("https://api.example.com/users")
                .method("GET")
                .header("Authorization", "Bearer token123")
                .header("Accept", "application/json")
                .timeout(5000)
                .build();

        System.out.println(getRequest);

        HttpRequest postRequest = new HttpRequest.Builder("https://api.example.com/users")
                .method("POST")
                .header("Content-Type", "application/json")
                .body("{\"name\": \"Alice\"}")
                .timeout(10000)
                .build();

        System.out.println(postRequest);

        // ═══════════════════════════════════════════════════════
        // ADVANCED: Builder with validation
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== BUILDER WITH VALIDATION ===");

        Pizza margherita = new Pizza.Builder("Medium")
                .addTopping("Mozzarella")
                .addTopping("Tomato")
                .addTopping("Basil")
                .crustType("Thin")
                .build();

        System.out.println(margherita);

        Pizza meatLovers = new Pizza.Builder("Large")
                .addTopping("Pepperoni")
                .addTopping("Sausage")
                .addTopping("Bacon")
                .addTopping("Ham")
                .crustType("Stuffed")
                .extraCheese(true)
                .build();

        System.out.println(meatLovers);
    }
}

// ═══════════════════════════════════════════════════════════════
// BAD: Telescoping constructor
// ═══════════════════════════════════════════════════════════════
class BadUser {
    String name, email, address, city, country, zipCode;
    int age;
    boolean newsletter, darkMode;

    public BadUser(String name, String email, int age, String address,
                   String city, String country, String zipCode,
                   boolean newsletter, boolean darkMode) {
        this.name = name; this.email = email; this.age = age;
        this.address = address; this.city = city; this.country = country;
        this.zipCode = zipCode; this.newsletter = newsletter;
        this.darkMode = darkMode;
    }

    @Override
    public String toString() { return "BadUser{" + name + ", " + email + "}"; }
}

// ═══════════════════════════════════════════════════════════════
// GOOD: Builder Pattern — immutable object with fluent API
// ═══════════════════════════════════════════════════════════════
class User {
    // All fields are final — immutable after construction
    private final String name;
    private final String email;
    private final int age;
    private final String address;
    private final String city;
    private final String country;
    private final String zipCode;
    private final boolean newsletter;

    // Private constructor — can only be called by Builder
    private User(Builder builder) {
        this.name = builder.name;
        this.email = builder.email;
        this.age = builder.age;
        this.address = builder.address;
        this.city = builder.city;
        this.country = builder.country;
        this.zipCode = builder.zipCode;
        this.newsletter = builder.newsletter;
    }

    // Getters only (no setters — immutable!)
    public String getName()    { return name; }
    public String getEmail()   { return email; }
    public int getAge()        { return age; }

    @Override
    public String toString() {
        return "User{name='" + name + "', email='" + email + "', age=" + age
                + ", city='" + city + "', newsletter=" + newsletter + "}";
    }

    // ─── STATIC INNER BUILDER CLASS ───
    public static class Builder {
        // Required parameters
        private final String name;
        private final String email;

        // Optional parameters — with defaults
        private int age = 0;
        private String address = "";
        private String city = "";
        private String country = "";
        private String zipCode = "";
        private boolean newsletter = false;

        public Builder(String name, String email) {
            this.name = name;
            this.email = email;
        }

        // Each setter returns the BUILDER (fluent API)
        public Builder age(int age)              { this.age = age; return this; }
        public Builder address(String address)    { this.address = address; return this; }
        public Builder city(String city)          { this.city = city; return this; }
        public Builder country(String country)    { this.country = country; return this; }
        public Builder zipCode(String zipCode)    { this.zipCode = zipCode; return this; }
        public Builder newsletter(boolean flag)   { this.newsletter = flag; return this; }

        public User build() {
            // Validation before construction
            if (name == null || name.isEmpty()) throw new IllegalStateException("Name required");
            if (email == null || !email.contains("@")) throw new IllegalStateException("Valid email required");
            return new User(this);
        }
    }
}

// ═══════════════════════════════════════════════════════════════
// REAL-WORLD: HTTP Request Builder
// ═══════════════════════════════════════════════════════════════
class HttpRequest {
    private final String url;
    private final String method;
    private final java.util.Map<String, String> headers;
    private final String body;
    private final int timeout;

    private HttpRequest(Builder builder) {
        this.url = builder.url;
        this.method = builder.method;
        this.headers = Collections.unmodifiableMap(builder.headers);
        this.body = builder.body;
        this.timeout = builder.timeout;
    }

    @Override
    public String toString() {
        return method + " " + url + "\n    Headers: " + headers
                + (body != null ? "\n    Body: " + body : "")
                + "\n    Timeout: " + timeout + "ms";
    }

    public static class Builder {
        private final String url;
        private String method = "GET";
        private java.util.Map<String, String> headers = new java.util.LinkedHashMap<>();
        private String body;
        private int timeout = 30000;

        public Builder(String url) { this.url = url; }

        public Builder method(String method)  { this.method = method; return this; }
        public Builder header(String k, String v) { this.headers.put(k, v); return this; }
        public Builder body(String body)      { this.body = body; return this; }
        public Builder timeout(int ms)        { this.timeout = ms; return this; }

        public HttpRequest build() { return new HttpRequest(this); }
    }
}

// ═══════════════════════════════════════════════════════════════
// BUILDER WITH COLLECTIONS AND VALIDATION
// ═══════════════════════════════════════════════════════════════
class Pizza {
    private final String size;
    private final String crustType;
    private final List<String> toppings;
    private final boolean extraCheese;

    private Pizza(Builder builder) {
        this.size = builder.size;
        this.crustType = builder.crustType;
        this.toppings = Collections.unmodifiableList(builder.toppings);
        this.extraCheese = builder.extraCheese;
    }

    @Override
    public String toString() {
        return "Pizza{" + size + ", " + crustType + " crust, toppings=" + toppings
                + (extraCheese ? " +EXTRA CHEESE" : "") + "}";
    }

    public static class Builder {
        private final String size;
        private String crustType = "Regular";
        private List<String> toppings = new ArrayList<>();
        private boolean extraCheese = false;

        public Builder(String size) { this.size = size; }

        public Builder crustType(String crust) { this.crustType = crust; return this; }
        public Builder addTopping(String topping) { this.toppings.add(topping); return this; }
        public Builder extraCheese(boolean flag) { this.extraCheese = flag; return this; }

        public Pizza build() {
            if (toppings.isEmpty()) throw new IllegalStateException("At least one topping!");
            return new Pizza(this);
        }
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Builder solves the "telescoping constructor" problem.
 * ✦ Fluent API: each method returns `this` → method chaining.
 * ✦ Creates IMMUTABLE objects (all final fields, no setters).
 * ✦ Separate required params (constructor) from optional (methods).
 * ✦ Validate in build() method before construction.
 * ✦ Use Collections.unmodifiableList() for list fields.
 *
 * ✦ Builder is a static inner class of the product.
 *
 * COMPILE & RUN:
 *   javac BuilderPattern.java && java BuilderPattern
 */
