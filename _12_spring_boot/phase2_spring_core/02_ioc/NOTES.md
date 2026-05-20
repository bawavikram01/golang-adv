# Phase 2.2 — Inversion of Control (IoC)

---

## The Big Idea in One Line

**IoC means: YOU don't control when objects are created or how they're connected. The FRAMEWORK does.**

---

## Traditional Control vs Inverted Control

### Traditional (YOU are in control)

```java
class OrderService {
    // YOU decide WHAT to create, WHEN to create, HOW to configure
    private EmailService email = new EmailService("smtp.gmail.com", 587);
    private OrderRepository repo = new OrderRepository(new DataSource("jdbc:mysql://..."));
    private PaymentGateway payment = new PaymentGateway(new HttpClient());

    public void placeOrder(Order order) {
        repo.save(order);
        payment.charge(order);
        email.send(order.getUserEmail(), "Confirmed!");
    }
}
```

Problems:
- `OrderService` is God — it knows how to create everything
- Can't test without real email/payment/database
- Can't swap implementations (what if you want SMS instead of email?)
- Every class is a factory for its own dependencies

### Inverted (FRAMEWORK is in control)

```java
@Service
class OrderService {
    // YOU declare what you NEED. Spring decides what/when/how to provide.
    private final EmailService email;
    private final OrderRepository repo;
    private final PaymentGateway payment;

    // Dependencies arrive from OUTSIDE. You didn't create them.
    public OrderService(EmailService email, OrderRepository repo, PaymentGateway payment) {
        this.email = email;
        this.repo = repo;
        this.payment = payment;
    }

    public void placeOrder(Order order) {
        repo.save(order);
        payment.charge(order);
        email.send(order.getUserEmail(), "Confirmed!");
    }
}
```

What changed:
- `OrderService` has NO idea how `EmailService` is configured
- It doesn't know if `repo` uses MySQL, Postgres, or an in-memory DB
- It doesn't care. It just USES the dependencies it receives.
- **Control over creation is INVERTED — moved from the class to the framework**

---

## The Hollywood Principle

> *"Don't call us, we'll call you."*

In traditional code: Your class calls out to create its own stuff.
In IoC: The framework calls YOUR class and provides everything it needs.

| | Traditional | IoC (Spring) |
|---|---|---|
| Who creates objects? | The class itself | The framework (container) |
| Who decides which implementation? | Hardcoded in the class | Configuration (annotations, profiles) |
| Who manages lifecycle? | You (`new`/`close`) | The container (creates & destroys) |
| Who connects objects? | You wire in main() | The container via DI |

---

## Three Forms of IoC in Spring

### Form 1: Dependency Injection (DI)
The container **provides dependencies** to your object.

```java
@Component
class UserService {
    private final UserRepository repo; // Container provides this

    public UserService(UserRepository repo) {
        this.repo = repo;
    }
}
```

### Form 2: Event-Driven Invocation
The container **calls your method** when an event happens.

```java
@EventListener
public void onUserRegistered(UserRegisteredEvent event) {
    // Spring calls THIS METHOD when the event fires. You don't poll.
}
```

### Form 3: Lifecycle Callbacks
The container **calls your code** at specific lifecycle stages.

```java
@Component
class CacheService {
    @PostConstruct  // Spring calls this AFTER creating the bean
    public void warmUpCache() {
        // Load data into cache
    }

    @PreDestroy  // Spring calls this BEFORE shutting down
    public void flushCache() {
        // Save cache to disk
    }
}
```

In ALL forms: **You don't initiate. The framework does, at the right time.**

---

## The IoC Container (ApplicationContext)

The container is the engine that implements IoC. Its job:

```
┌─────────────────────────────────────────────────┐
│          ApplicationContext (IoC Container)       │
│                                                  │
│  Input: Bean Definitions                         │
│    • @Component classes (component scanning)     │
│    • @Bean methods (Java configuration)          │
│    • XML config (legacy)                         │
│                                                  │
│  Process:                                        │
│    1. Read all bean definitions                  │
│    2. Determine creation order (dependency graph)│
│    3. Instantiate beans (reflection)             │
│    4. Inject dependencies                        │
│    5. Run lifecycle callbacks (@PostConstruct)   │
│    6. Store as singletons (default)              │
│                                                  │
│  Output: Fully wired, ready-to-use application   │
└─────────────────────────────────────────────────┘
```

### Container Implementations
| Class | When Used |
|-------|-----------|
| `AnnotationConfigApplicationContext` | Standalone (no web server) |
| `GenericWebApplicationContext` | Web applications |
| `SpringApplication.run()` | Spring Boot (auto-detects which to use) |

---

## Bean Definition Sources

| Source | How | Example |
|--------|-----|---------|
| **Component Scanning** | `@Component` on class | `@Service class UserService {}` |
| **Java Config** | `@Bean` method in `@Configuration` class | `@Bean DataSource dataSource() {...}` |
| **XML** (legacy) | `<bean>` tag in XML file | `<bean class="com.UserService"/>` |
| **Programmatic** | Register with context directly | `context.registerBean(MyClass.class)` |

Modern Spring uses **Component Scanning** (90%) + **Java Config** (10%).
XML is legacy — you'll see it in old projects but never write it.

---

## Dependency Resolution — How Spring Figures Out What to Inject

When Spring sees:
```java
public UserService(UserRepository repo, NotificationService notif) { ... }
```

It runs this logic:
1. "UserService needs a `UserRepository` bean" → search container → found → inject
2. "UserService needs a `NotificationService` bean" → search container → found → inject
3. "All dependencies resolved" → create UserService → store as singleton

**What if a dependency doesn't exist?**
```
***************************
APPLICATION FAILED TO START
***************************
Parameter 0 of constructor in UserService required a bean of type 'UserRepository' that could not be found.
```

**What if TWO beans match the same type?** (e.g., two `NotificationService` implementations)
→ Spring throws `NoUniqueBeanDefinitionException`
→ Solution: `@Primary` or `@Qualifier` (covered in Phase 2.8)

---

## Lazy vs Eager Initialization

| | Eager (Default) | Lazy |
|---|---|---|
| When created | At startup (container refresh) | On first use |
| Annotation | None (default) | `@Lazy` |
| Advantage | Fail fast — errors at startup | Faster startup |
| Disadvantage | Slower startup | Errors at runtime |

```java
@Component
@Lazy  // Not created until someone @Autowired it AND calls a method on it
class ExpensiveService {
    public ExpensiveService() {
        // Heavy initialization — only happens when first needed
    }
}
```

---

## Files in This Module

```
phase2_spring_core/02_ioc/
├── NOTES.md                             ← This file
├── src/
│   └── Step1_IoCComparison.java         ← Plain Java: traditional vs IoC style
└── spring-ioc-demo/                     ← Spring Boot project
    ├── pom.xml
    └── src/main/java/com/learn/ioc/
        ├── IocDemoApplication.java      ← Main + lifecycle demos
        ├── DatabaseService.java         ← @PostConstruct / @PreDestroy
        ├── NotificationService.java     ← Interface
        ├── EmailNotification.java       ← Implementation 1
        ├── SmsNotification.java         ← Implementation 2
        └── OrderService.java            ← Consumer (demonstrates IoC)
```

---

## Key Takeaways

1. **IoC = you don't control creation/wiring** — The framework does
2. **You declare needs (constructor params)** — The container fulfills them
3. **ApplicationContext = the IoC container** — Reads config, creates beans, injects deps
4. **Three forms**: Dependency Injection, Event invocation, Lifecycle callbacks
5. **Benefits**: Loose coupling, testability, swappable implementations, centralized config
6. **Default behavior**: Eager creation, singleton scope
