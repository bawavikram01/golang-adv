# Phase 1.1 — Java Core Refresher

These are the 5 Java concepts that Spring **cannot work without**. Skip any of these and Spring will feel like magic you can't control. Understand these and Spring becomes transparent.

---

## 1. Interfaces — The #1 Concept in Spring

### What
An **interface** is a contract. It says *what* something can do, without saying *how*.

### Why Spring Cares
Spring's entire philosophy is: **"depend on abstractions, not concrete classes."** Every time Spring injects a dependency, it prefers to inject an interface. This is what makes your code swappable, testable, and loosely coupled.

### Analogy
Think of a **power socket** (interface) and **appliances** (implementations). The socket defines the shape (contract). Any appliance that matches the shape can plug in — a TV, a phone charger, a lamp. The socket doesn't care *which* appliance. It just provides power to anything that fits.

Spring is the socket. Your classes are the appliances.

### Key Code Pattern
```java
// INTERFACE — the contract
interface NotificationService {
    void send(String message);
}

// IMPLEMENTATION 1
class EmailNotification implements NotificationService {
    public void send(String message) { /* email logic */ }
}

// IMPLEMENTATION 2
class SmsNotification implements NotificationService {
    public void send(String message) { /* sms logic */ }
}

// CONSUMER — depends on the INTERFACE, not the concrete class
class OrderService {
    private NotificationService notificationService; // <-- INTERFACE type

    // Implementation is INJECTED from outside (this is what Spring does!)
    public OrderService(NotificationService notificationService) {
        this.notificationService = notificationService;
    }
}
```

### Tight Coupling vs Loose Coupling
| | Tight Coupling (Bad) | Loose Coupling (Good) |
|---|---|---|
| **Dependency** | `private EmailService emailService = new EmailService()` | `private NotificationService ns` (interface) |
| **Created by** | The class itself (`new`) | Passed from outside (constructor) |
| **To swap** | Edit source code | Just pass a different implementation |
| **Testable?** | Hard (real email sends) | Easy (pass a mock) |

### Spring Connection
What you do manually: `new OrderService(new EmailNotification())`
In Spring: Spring automatically finds `EmailNotification`, creates it, and injects it into `OrderService`. Zero wiring code.

**See:** `Step1_Interfaces.java`

---

## 2. Generics — Type Safety in Spring

### What
Generics let you write classes and methods that work with **any type**, while keeping compile-time type safety.

### Why Spring Cares
Spring uses generics everywhere:
- `List<User>` — type-safe collections
- `ResponseEntity<Product>` — typed HTTP responses
- `JpaRepository<User, Long>` — typed data repositories
- `Optional<Order>` — safe nullable containers

### Analogy
A **vending machine**. The machine mechanism is the same, but one is loaded with snacks, another with drinks. The machine is generic — the **type parameter** decides what comes out.

### Key Code Pattern
```java
// Generic class — T is a placeholder
class Box<T> {
    private T item;
    public void put(T item) { this.item = item; }
    public T get() { return item; }
}

Box<String> stringBox = new Box<>();  // T becomes String
Box<Integer> intBox = new Box<>();    // T becomes Integer

// Generic interface — EXACTLY like Spring Data!
interface Repository<T, ID> {
    void save(T entity);
    T findById(ID id);
    List<T> findAll();
}

// In Spring you'll write:
// public interface UserRepository extends JpaRepository<User, Long> {}
// Spring auto-generates the implementation!
```

### Key Point
Without generics, you'd need `UserRepository`, `ProductRepository`, `OrderRepository` — each written from scratch. With generics, ONE interface (`JpaRepository<T, ID>`) serves ALL entities.

**See:** `Step2_Generics.java`

---

## 3. Annotations — Spring's Language

### What
Annotations are metadata you attach to classes, methods, or fields using `@`. They don't execute code themselves — but **frameworks read them** and act accordingly.

### Why Spring Cares
Spring is **annotation-driven**. Almost everything is an annotation:
- `@Component` — "Spring, manage this class"
- `@Autowired` — "Spring, inject a dependency here"
- `@RestController` — "This class handles HTTP requests"
- `@GetMapping("/users")` — "Handle GET /users"
- `@Transactional` — "Wrap this in a database transaction"

### Analogy
Sticky notes on a document. The document works on its own, but the sticky notes tell the **reviewer** (framework) what to do with it. A note saying "URGENT" tells the mail system to prioritize it. The letter's content doesn't change.

### Key Code Pattern
```java
// DEFINING a custom annotation
@Retention(RetentionPolicy.RUNTIME)  // Survives until runtime
@Target(ElementType.TYPE)            // Can be placed on classes
@interface Component {
    String value() default "";
}

// USING the annotation
@Component("userService")
class UserService { ... }

// READING the annotation at runtime (what Spring does!)
if (clazz.isAnnotationPresent(Component.class)) {
    Component comp = clazz.getAnnotation(Component.class);
    String name = comp.value();  // "userService"
    // Spring now creates and manages this class
}
```

### Three Parts of the Annotation System
| Part | What | Example |
|---|---|---|
| **Definition** | `@interface MyAnnotation` | Creating the annotation |
| **Usage** | `@MyAnnotation` on a class/method | Applying the label |
| **Processor** | Code that reads annotations via reflection | Spring's container at startup |

### Key Point
Annotations are just **labels**. They do nothing alone. A **processor** (like the Spring container) reads them and acts. Without the processor, `@Component` is meaningless text.

**See:** `Step3_Annotations.java`

---

## 4. Reflection — How Spring Sees Your Code

### What
Reflection lets Java inspect and manipulate classes, methods, and fields **at runtime** — even private ones.

### Why Spring Cares
Spring uses reflection to:
1. **Discover** which classes have `@Component` (component scanning)
2. **Create** objects without you calling `new` (`clazz.newInstance()`)
3. **Read** constructor parameters to figure out what dependencies to inject
4. **Inject** values into private fields (`@Autowired` on a private field)
5. **Invoke** methods dynamically (like `@PostConstruct` lifecycle callbacks)

### Analogy
You're a **building inspector**. You can walk into any room (class), open any drawer (field), read any document (method), even if they're locked (private). That's reflection — Java's X-ray vision.

### Key Code Pattern
```java
// 1. Get class info
Class<?> clazz = PaymentService.class;

// 2. Create object WITHOUT "new"
Object bean = clazz.getDeclaredConstructor().newInstance();

// 3. Access PRIVATE field
Field field = clazz.getDeclaredField("provider");
field.setAccessible(true);       // Bypass "private"!
field.set(bean, "Razorpay");     // Set value directly

// 4. Invoke method dynamically
Method method = clazz.getDeclaredMethod("processPayment", double.class);
method.invoke(bean, 250.0);
```

### Mini Spring Container (What Spring Does at Startup)
```
1. Component Scanning  → Find all classes with @Component
2. Reflection          → Create instances (beans) via .newInstance()
3. Read Annotations    → Check for @Autowired fields/constructors
4. Inject Dependencies → Use reflection to set private fields
5. Lifecycle Callbacks → Invoke @PostConstruct methods
```

### Key Point
There is **no magic** in Spring. It's annotations (labels) + reflection (reading labels and acting on them). That's it.

**See:** `Step4_Reflection.java`

---

## 5. Lambda & Streams — Modern Spring Style

### What
**Lambdas** are short anonymous functions: `(params) -> expression`
**Streams** process collections in a pipeline: `.filter().map().collect()`
**Optional** is a safe container that may or may not hold a value.

### Why Spring Cares
Modern Spring uses these extensively:
```java
// Security config
http.csrf(csrf -> csrf.disable())
    .authorizeHttpRequests(auth -> auth.requestMatchers("/api/**").authenticated());

// Repository results
User user = userRepo.findById(id)
    .orElseThrow(() -> new UserNotFoundException(id));

// Reactive (WebFlux)
Flux<User> users = userService.findAll()
    .filter(u -> u.isActive())
    .map(u -> new UserDto(u.getName()));
```

### Analogy
- **Lambda** = A delivery note without a name. Instead of hiring a full-time employee (named method), you hand a sticky note saying "do this."
- **Stream** = A factory conveyor belt. Items flow through stations (filter, transform, collect) and come out as a finished product.

### Key Functional Interfaces
| Interface | Takes | Returns | Spring Usage |
|---|---|---|---|
| `Predicate<T>` | T | boolean | `.filter(user -> user.isActive())` |
| `Function<T,R>` | T | R | `.map(user -> user.getName())` |
| `Consumer<T>` | T | void | `.forEach(user -> log(user))` |
| `Supplier<T>` | nothing | T | `.orElseGet(() -> new User())` |
| `Runnable` | nothing | void | `@Async` tasks |

### Key Stream Operations
| Operation | What | Example |
|---|---|---|
| `filter` | Keep matching elements | `.filter(e -> e.salary > 100000)` |
| `map` | Transform each element | `.map(e -> e.name.toUpperCase())` |
| `sorted` | Order elements | `.sorted(Comparator.comparing(e -> e.name))` |
| `limit` | Take first N | `.limit(5)` |
| `collect` | Gather results | `.collect(Collectors.toList())` |
| `forEach` | Execute action on each | `.forEach(System.out::println)` |
| `reduce` | Combine into one value | `.reduce(0, Integer::sum)` |
| `groupingBy` | Group by a key | `.collect(Collectors.groupingBy(e -> e.dept))` |

**See:** `Step5_LambdaStreams.java`

---

## Challenge Solution

Combines all 5 concepts into one program:
1. **Interface** → `MessageSender` (contract)
2. **Generics** → `List<Message>`, `List<Class<? extends MessageSender>>`
3. **Annotation** → `@DefaultSender` (custom, marks the default implementation)
4. **Reflection** → Finds `@DefaultSender` at runtime, creates instance
5. **Lambda & Streams** → `.filter().map().peek().collect()` pipeline

```
                      @DefaultSender          ← ANNOTATION (label)
                           │
                      EmailSender             ← implements INTERFACE (contract)
                           │
     REFLECTION finds it ──┘  creates instance via .newInstance()
                           │
                     defaultSender            ← typed as INTERFACE (MessageSender)
                           │
     messages.stream()     │                  ← GENERICS (List<Message>)
        .filter(important) │                  ← LAMBDA
        .map(getText)      │                  ← LAMBDA
        .peek(sender::send)┘                  ← LAMBDA calls the INTERFACE method
        .collect(toList())                    ← GENERICS (List<String>)
```

This is a **miniature Spring container**. Spring does exactly this flow.

**See:** `Step6_Challenge.java`

---

## Summary — Why Each Concept Matters for Spring

| Java Concept | Spring Uses It For |
|---|---|
| **Interfaces** | Dependency Injection — depend on contracts, not implementations |
| **Generics** | `JpaRepository<User, Long>`, `ResponseEntity<T>`, type-safe APIs |
| **Annotations** | `@Component`, `@Autowired`, `@GetMapping` — Spring's entire config system |
| **Reflection** | Creating beans, injecting fields, reading annotations at startup |
| **Lambda & Streams** | Security config, repository queries, Optional, WebFlux |
