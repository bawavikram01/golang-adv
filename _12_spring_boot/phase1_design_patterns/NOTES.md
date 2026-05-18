# Phase 1.3 — Design Patterns for Spring

We cover the **5 patterns** that Spring uses internally. When you see Spring doing something and wonder "how?", the answer is always one of these patterns.

---

## Why Design Patterns Matter for Spring

| Pattern | Where Spring Uses It |
|---------|---------------------|
| **Singleton** | Every Spring bean is a singleton by default |
| **Factory** | `BeanFactory` / `ApplicationContext` creates all beans |
| **Proxy** | `@Transactional`, `@Cacheable`, AOP — all use proxy objects |
| **Template Method** | `JdbcTemplate`, `RestTemplate`, `JpaRepository` |
| **Observer** | `ApplicationEvent`, `@EventListener` |

If you understand these 5, you understand Spring's architecture.

---

## 1. Singleton — One Instance, Shared Everywhere

### What
Only **one instance** of a class exists in the entire application. Everyone shares it.

### Why Spring Cares
**Every Spring bean is a singleton by default.** When you mark a class with `@Component`, Spring creates ONE instance and injects that same instance everywhere it's needed.

```
@Component
class UserService { ... }

// Controller A gets userService instance #1
// Controller B gets the SAME userService instance #1
// There is only ONE UserService in the entire app
```

### Analogy
A **CEO** of a company. There's only one CEO. Every department references the same person. You don't create a new CEO for every meeting.

### How Java Implements Singleton
```java
class DatabaseConnection {
    private static DatabaseConnection instance;

    private DatabaseConnection() {}  // Private constructor — no one can "new" this

    public static DatabaseConnection getInstance() {
        if (instance == null) {
            instance = new DatabaseConnection();
        }
        return instance;
    }
}
```

### The Spring Way (You Don't Write Singleton Code)
```java
@Component  // Spring automatically makes this a singleton
class UserService { ... }
```
Spring manages the singleton lifecycle for you. You never write the pattern manually.

### Pitfall
Singleton beans **must be stateless** (no mutable instance fields that change per request). Since the same instance is shared across threads, mutable state causes race conditions.

**See:** `Step1_Singleton.java`

---

## 2. Factory — Creating Objects Without "new"

### What
A **Factory** creates objects for you. You tell it *what* you want (by type or name), and it returns the right object. You don't call `new` yourself.

### Why Spring Cares
The entire Spring container is a factory. `ApplicationContext` (which extends `BeanFactory`) is the world's most sophisticated factory:
- You say: "I need a `UserService`"
- Spring says: "Here's one, already created, dependencies injected, ready to use"

```java
// You NEVER do this in Spring:
UserService us = new UserService(new UserRepository(new DataSource(...)));

// Spring's factory does it for you:
@Autowired
UserService us;  // Factory created it, wired it, handed it to you
```

### Analogy
A **pizza restaurant**. You say "Margherita" — the kitchen (factory) makes it. You don't know or care about the recipe, oven temperature, or ingredients. You get a pizza.

### Patterns
| Variant | What | Spring Equivalent |
|---------|------|-------------------|
| Simple Factory | One method, switch on type | `BeanFactory.getBean("name")` |
| Factory Method | Subclasses decide what to create | `@Bean` methods in `@Configuration` |
| Abstract Factory | Family of related objects | Spring profiles + multiple factories |

**See:** `Step2_Factory.java`

---

## 3. Proxy — Do Something Before/After Without Changing Code

### What
A **Proxy** wraps a real object and intercepts calls to it. It can add behavior **before** and **after** the real method runs — without modifying the original class.

### Why Spring Cares
This is Spring's **most powerful pattern**. Used in:
- `@Transactional` — Proxy opens a DB transaction before your method, commits/rollbacks after
- `@Cacheable` — Proxy checks cache before calling your method
- `@Async` — Proxy runs your method in a separate thread
- AOP (`@Before`, `@After`, `@Around`) — Custom cross-cutting logic

```
Your code calls:  userService.save(user)
                       │
                       ▼
             Proxy intercepts:
                ┌──────────────────────┐
                │ BEGIN TRANSACTION    │   ← added by proxy
                │ userService.save()   │   ← your real code
                │ COMMIT TRANSACTION   │   ← added by proxy
                └──────────────────────┘
```

### Analogy
A **security guard at a building entrance**. Every visitor (method call) passes through the guard (proxy). The guard can:
- Check ID before entering (validation)
- Log the visit (logging)
- Deny entry (authorization)
- The actual offices (your code) don't change at all

### Two Types of Proxies in Spring
| Type | How | When |
|------|-----|------|
| **JDK Dynamic Proxy** | Creates proxy implementing the same **interface** | When class implements an interface |
| **CGLIB Proxy** | Creates a **subclass** of your class | When class doesn't implement an interface |

Spring picks the right one automatically. You don't choose.

**See:** `Step3_Proxy.java`

---

## 4. Template Method — Define the Skeleton, Let Subclasses Fill In

### What
Define the **overall algorithm** in a base class, but let subclasses **override specific steps**. The structure is fixed; the details are pluggable.

### Why Spring Cares
Spring is full of "Template" classes:
- `JdbcTemplate` — Handles connection, statement, exception, cleanup. You just provide the SQL.
- `RestTemplate` — Handles HTTP connection, serialization, error handling. You just provide the URL.
- `JpaRepository` — Handles all CRUD boilerplate. You just define the entity type.
- `TransactionTemplate` — Handles begin/commit/rollback. You just provide the logic.

```
Without template:
  1. Get connection        ← boilerplate
  2. Create statement      ← boilerplate
  3. Execute query         ← YOUR CODE (the only unique part)
  4. Map results           ← YOUR CODE
  5. Handle exceptions     ← boilerplate
  6. Close connection      ← boilerplate

With JdbcTemplate:
  jdbcTemplate.query("SELECT * FROM users", (rs) -> new User(rs.getString("name")));
  // Steps 1, 2, 5, 6 are handled for you!
```

### Analogy
A **recipe framework**. The steps are fixed: prep → cook → plate → serve. But *what* you prep and *how* you cook is up to the chef (subclass). The sequence never changes.

**See:** `Step4_TemplateMethod.java`

---

## 5. Observer — When Something Happens, Notify Everyone Who Cares

### What
An object (subject/publisher) maintains a list of dependents (observers/subscribers). When its state changes, it **automatically notifies** all observers.

### Why Spring Cares
Spring has a built-in event system:
- `ApplicationEvent` — The event object
- `ApplicationEventPublisher` — Publishes events
- `@EventListener` — Methods that react to events

```java
// When a user registers...
publisher.publishEvent(new UserRegisteredEvent(user));

// These all react AUTOMATICALLY:
@EventListener
void sendWelcomeEmail(UserRegisteredEvent e) { ... }

@EventListener
void createDefaultSettings(UserRegisteredEvent e) { ... }

@EventListener
void notifyAdmins(UserRegisteredEvent e) { ... }
```

The registration code doesn't know about emails, settings, or admin notifications. It just fires an event. **Loose coupling at its finest.**

### Analogy
A **YouTube channel**. The creator (publisher) uploads a video. All subscribers (observers) get notified. The creator doesn't send individual messages — the platform handles distribution. New subscribers can join anytime without changing the creator's code.

### Benefits
- **Decoupling** — Publisher doesn't know who's listening
- **Extensibility** — Add new listeners without changing the publisher
- **Async possible** — Spring can fire events asynchronously (`@Async @EventListener`)

**See:** `Step5_Observer.java`

---

## How These Patterns Connect in Spring

```
                        ApplicationContext (FACTORY)
                               │
              creates beans as SINGLETONS
                               │
              ┌────────────────┼────────────────┐
              │                │                │
         UserService     OrderService      EventSystem
         (SINGLETON)     (wrapped in        (OBSERVER)
              │           PROXY for           │
              │          @Transactional)       │
              │                │               │
         uses JdbcTemplate     │          publishes
         (TEMPLATE METHOD)     │          OrderPlacedEvent
                               │               │
                          save(order)     @EventListener
                          BEGIN TX        sendConfirmation()
                          INSERT...       updateInventory()
                          COMMIT TX
```

All 5 patterns working together in a single Spring application.

---

## Files in This Module

```
phase1_design_patterns/src/
├── Step1_Singleton.java       ← Singleton (Spring's default bean scope)
├── Step2_Factory.java         ← Factory (how Spring creates beans)
├── Step3_Proxy.java           ← Proxy (how @Transactional, AOP work)
├── Step4_TemplateMethod.java  ← Template Method (JdbcTemplate, RestTemplate)
└── Step5_Observer.java        ← Observer (Spring's event system)
```

---

## Key Takeaways

1. **Singleton** — Spring beans are singletons by default. One instance, shared everywhere. Keep them stateless.
2. **Factory** — `ApplicationContext` is a giant factory that creates, wires, and manages all beans.
3. **Proxy** — Spring wraps your beans in proxies to add `@Transactional`, `@Cacheable`, AOP behavior transparently.
4. **Template Method** — `JdbcTemplate`, `RestTemplate` handle boilerplate; you provide only the unique logic.
5. **Observer** — `@EventListener` reacts to events. Publisher doesn't know who's listening. Maximum decoupling.
