/**
 * PHASE 1.1 — GENERICS -  Type Safety in Spring
 * 
 * Theory
Generics let you write classes and methods that work with any type, while keeping type safety. Spring uses generics everywhere — List<User>, ResponseEntity<T>, JpaRepository<User, Long>.

Analogy
A vending machine. The machine mechanism is the same, but one is loaded with snacks, another with drinks. The machine is generic — the type parameter decides what comes out.

 * Spring uses generics EVERYWHERE:
 *   - List<User>
 *   - ResponseEntity<Product>
 *   - JpaRepository<User, Long>
 *   - Optional<Order>
 * 
 * You MUST understand them.
 */

import java.util.ArrayList;
import java.util.List;

// ============================================================
// STEP 1: The Problem — Without Generics
// ============================================================

class OldBox {
    private Object item; // Can hold ANYTHING... but that's the problem

    public void put(Object item) { this.item = item; }

    public Object get() { return item; }
}


// ============================================================
// STEP 2: The Solution — Generics
// ============================================================

// T is a TYPE PARAMETER. It's a placeholder for any type.
// When you create a Box<String>, T becomes String everywhere.
class Box<T> {
    private T item;

    public void put(T item) { this.item = item; }

    public T get() { return item; }
}


// ============================================================
// STEP 3: Real-world — Generic Repository (like Spring Data!)
// ============================================================

// This is EXACTLY the pattern Spring Data JPA uses.
// JpaRepository<Entity, IdType> — you'll see this in Phase 5.

interface Repository<T, ID> {
    void save(T entity);
    T findById(ID id);
    List<T> findAll();
}

// A simple User class
class User {
    private Long id;
    private String name;

    public User(Long id, String name) { this.id = id; this.name = name; }
    public Long getId() { return id; }
    public String getName() { return name; }
    public String toString() { return "User{id=" + id + ", name='" + name + "'}"; }
}

// A simple Product class
class Product {
    private Long id;
    private String title;

    public Product(Long id, String title) { this.id = id; this.title = title; }
    public Long getId() { return id; }
    public String getTitle() { return title; }
    public String toString() { return "Product{id=" + id + ", title='" + title + "'}"; }
}

// UserRepository — Repository<User, Long> means T=User, ID=Long
class UserRepository implements Repository<User, Long> {
    private List<User> store = new ArrayList<>();

    public void save(User entity) { store.add(entity); }

    public User findById(Long id) {
        return store.stream().filter(u -> u.getId().equals(id)).findFirst().orElse(null);
    }

    public List<User> findAll() { return store; }
}

// ProductRepository — same interface, different types!
class ProductRepository implements Repository<Product, Long> {
    private List<Product> store = new ArrayList<>();

    public void save(Product entity) { store.add(entity); }

    public Product findById(Long id) {
        return store.stream().filter(p -> p.getId().equals(id)).findFirst().orElse(null);
    }

    public List<Product> findAll() { return store; }
}


// ============================================================
// STEP 4: Generic Methods
// ============================================================

class Utils {
    // <T> before return type declares a type parameter for THIS METHOD
    public static <T> void printAll(List<T> items) {
        for (T item : items) {
            System.out.println("  → " + item);
        }
    }
}


public class Step2_Generics {
    public static void main(String[] args) {

        System.out.println("=== WITHOUT GENERICS (Unsafe) ===");
        OldBox box = new OldBox();
        box.put("Hello");
        // You must cast, and it can crash at runtime!
        String val = (String) box.get();
        System.out.println("  Got: " + val);
        // box.put(123); // No compile error, but now (String) box.get() would CRASH!

        System.out.println();
        System.out.println("=== WITH GENERICS (Type-Safe) ===");
        Box<String> stringBox = new Box<>();
        stringBox.put("Hello");
        String safe = stringBox.get(); // No casting needed!
        System.out.println("  Got: " + safe);
        // stringBox.put(123); // COMPILE ERROR! The compiler protects you.

        Box<Integer> intBox = new Box<>();
        intBox.put(42);
        int num = intBox.get();
        System.out.println("  Got: " + num);

        System.out.println();
        System.out.println("=== GENERIC REPOSITORY (Spring Data Pattern) ===");

        UserRepository userRepo = new UserRepository();
        userRepo.save(new User(1L, "Alice"));
        userRepo.save(new User(2L, "Bob"));

        System.out.println("All users:");
        Utils.printAll(userRepo.findAll());

        System.out.println("Find user by ID 1: " + userRepo.findById(1L));

        System.out.println();

        ProductRepository productRepo = new ProductRepository();
        productRepo.save(new Product(1L, "MacBook Pro"));
        productRepo.save(new Product(2L, "Mechanical Keyboard"));

        System.out.println("All products:");
        Utils.printAll(productRepo.findAll());

        System.out.println();
        System.out.println("=== KEY TAKEAWAY ===");
        System.out.println("Repository<T, ID> is ONE interface that works for ANY entity.");
        System.out.println("In Spring Data, you'll write: JpaRepository<User, Long>");
        System.out.println("Spring auto-generates the implementation. You write ZERO code.");
    }
}
