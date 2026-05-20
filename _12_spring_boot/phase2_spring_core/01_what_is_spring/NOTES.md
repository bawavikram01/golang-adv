# Phase 2.1 — What is Spring & Why

---

## The ONE Sentence Answer

**Spring is a framework that creates objects for you, wires them together, and manages their lifecycle — so you focus only on business logic.**

That's it. Everything else in Spring is built on top of this core idea.

---

## The Problem Spring Solves

### Without Spring: YOU manage everything

In a real application, you have dozens of classes that depend on each other:

```
UserController
    → needs UserService
        → needs UserRepository
            → needs DataSource
                → needs connection pool config
                    → needs properties from a file

OrderController
    → needs OrderService
        → needs OrderRepository (same DataSource)
        → needs UserService (same instance!)
        → needs PaymentGateway
            → needs HttpClient
                → needs SSL config
```

**Without Spring, YOU manually:**
1. Create every object
2. Pass dependencies through constructors
3. Ensure singletons are shared (not duplicated)
4. Handle lifecycle (startup, shutdown)
5. Manage configuration loading
6. Handle cross-cutting concerns (logging, transactions)

```java
// WITHOUT SPRING — The manual nightmare
DataSource ds = new HikariDataSource(loadConfig("db.properties"));
UserRepository userRepo = new UserRepository(ds);
OrderRepository orderRepo = new OrderRepository(ds);
PaymentGateway gateway = new PaymentGateway(new HttpClient(new SSLConfig()));
UserService userService = new UserService(userRepo);
OrderService orderService = new OrderService(orderRepo, userService, gateway);
UserController userController = new UserController(userService);
OrderController orderController = new OrderController(orderService);
// ... and 50 more classes ...
// What if UserService's constructor changes? Fix it everywhere!
```

### With Spring: Spring manages everything

```java
// WITH SPRING — You declare. Spring wires.
@Component
class UserRepository { ... }

@Component
class UserService {
    @Autowired UserRepository repo; // Spring injects this
}

@RestController
class UserController {
    @Autowired UserService service; // Spring injects this
}

// That's it. Spring figures out the order, creates everything, wires everything.
```

---

## The Core Principles

### 1. Inversion of Control (IoC)

**Normal control:** Your code creates its own dependencies.
```java
class OrderService {
    private EmailService email = new EmailService(); // YOU control creation
}
```

**Inverted control:** The framework creates dependencies and gives them to you.
```java
class OrderService {
    private EmailService email; // SPRING controls creation, you just receive it

    public OrderService(EmailService email) {
        this.email = email; // Spring passes it in
    }
}
```

**"Don't call us, we'll call you."** (Hollywood Principle)

You don't go to the store (create objects). The delivery service (Spring) brings them to your door.

### 2. Dependency Injection (DI)

DI is the *mechanism* that implements IoC. Spring "injects" (passes/sets) dependencies into your objects.

Three ways to inject:
| Method | Syntax | Recommended? |
|--------|--------|--------------|
| **Constructor** | `public MyService(Repo repo)` | ✅ YES — immutable, testable |
| **Setter** | `@Autowired public void setRepo(Repo repo)` | ⚠️ Sometimes — optional deps |
| **Field** | `@Autowired private Repo repo` | ❌ Avoid — hard to test |

### 3. The Spring Container (ApplicationContext)

The container is the **brain** of Spring. It:
1. **Reads** your configuration (annotations, XML, Java config)
2. **Creates** all beans (objects managed by Spring)
3. **Wires** dependencies (injects beans into each other)
4. **Manages** lifecycle (init, destroy, scopes)
5. **Stores** singletons in a registry (HashMap internally)

```
Your Code (annotations) ──→ Spring Container ──→ Fully wired application
                              │
                              ├── Creates beans
                              ├── Resolves dependencies
                              ├── Injects values
                              └── Manages lifecycle
```

---

## Spring vs Spring Boot

| | Spring Framework | Spring Boot |
|---|---|---|
| **What** | The core dependency injection + ecosystem | Opinionated auto-configuration layer ON TOP of Spring |
| **Config** | You configure everything manually | Auto-configures sensible defaults |
| **Server** | You deploy to external Tomcat/Jetty | Embedded server (just run the JAR) |
| **Dependencies** | You pick every library + version | Starters bundle compatible versions |
| **Analogy** | A car engine + parts (you assemble) | A fully assembled car (just drive) |

**Spring Boot doesn't replace Spring. It makes Spring easier to use.**

Under every Spring Boot app, it's still the Spring Framework doing the actual work (IoC, DI, AOP, etc.).

---

## The Spring Ecosystem

```
┌────────────────────────────────────────────────────────────┐
│                    SPRING BOOT                              │
│    (auto-config, starters, embedded server)                │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              SPRING FRAMEWORK (Core)                  │  │
│  │                                                      │  │
│  │  ┌─────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐  │  │
│  │  │   IoC   │ │   AOP    │ │  MVC   │ │   Data   │  │  │
│  │  │Container│ │ (Proxy)  │ │ (Web)  │ │  Access  │  │  │
│  │  └─────────┘ └──────────┘ └────────┘ └──────────┘  │  │
│  │                                                      │  │
│  │  ┌─────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐  │  │
│  │  │Security │ │  Events  │ │Testing │ │ Messaging│  │  │
│  │  └─────────┘ └──────────┘ └────────┘ └──────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                            │
│  ┌────────────┐ ┌────────────┐ ┌────────────────────┐     │
│  │Spring Cloud│ │Spring Data │ │ Spring Security    │     │
│  │(Microsvcs) │ │ (JPA/Mongo)│ │ (Auth/Authz)      │     │
│  └────────────┘ └────────────┘ └────────────────────┘     │
└────────────────────────────────────────────────────────────┘
```

---

## What is a "Bean"?

A **bean** = any object that Spring creates and manages.

- NOT a special class. ANY Java class can be a bean.
- Spring stores beans in its container (a Map internally).
- By default, beans are **singletons** (one instance, reused everywhere).

How to tell Spring "this is a bean":
| Method | Example |
|--------|---------|
| `@Component` | `@Component class UserService {}` |
| `@Service` | `@Service class OrderService {}` (same as @Component, semantic) |
| `@Repository` | `@Repository class UserRepo {}` (same, adds DB exception translation) |
| `@Controller` | `@Controller class UserController {}` (same, for web) |
| `@Bean` method | `@Bean public DataSource ds() { return new HikariDataSource(); }` |

---

## The Mental Model

Think of Spring as a **restaurant kitchen**:

| Restaurant | Spring |
|-----------|--------|
| Menu items | Bean definitions (your annotated classes) |
| Head chef | The container (ApplicationContext) |
| Ingredients | Dependencies (other beans) |
| Prep stations | Bean scopes (singleton, prototype) |
| The recipe | Configuration (annotations, properties) |
| Plating & serving | Dependency injection |

You (the waiter/customer) just say: "I need a UserService." The kitchen handles prep, cooking, plating, and delivers a ready-to-use object.

---

## History: Why Spring Was Created

**2002** — Java EE (J2EE) was the standard for enterprise apps. It was:
- Extremely verbose (XML everywhere)
- Required heavy application servers
- Tightly coupled to container APIs
- Complex even for simple tasks

**Rod Johnson** wrote "Expert One-on-One J2EE Design and Development" showing that most J2EE complexity was unnecessary. He created Spring as a **lightweight alternative**.

Spring's promise: **"Simple things should be simple. Complex things should be possible."**

---

## Files in This Module

```
phase2_spring_core/01_what_is_spring/
├── NOTES.md                           ← This file
└── src/
    ├── Step1_ProblemWithoutSpring.java ← The mess without a framework
    └── Step2_ManualDIContainer.java   ← Build your own mini-Spring
```

---

## Key Takeaways

1. **Spring = IoC Container** — It creates objects and wires them together
2. **Bean = any object managed by Spring** — Annotate with `@Component` and Spring manages it
3. **DI = mechanism** — Spring injects dependencies via constructor/setter/field
4. **Container (ApplicationContext) = the brain** — Reads config, creates beans, resolves dependencies
5. **Spring Boot = Spring + auto-config** — Makes Spring easier, but it's still Spring underneath
6. **No magic** — It's just the Factory + Singleton + Proxy patterns you already learned
