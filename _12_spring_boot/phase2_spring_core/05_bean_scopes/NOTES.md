# Phase 2.5 — Bean Scopes

---

## What is a Bean Scope?

A **scope** defines how many instances of a bean the container creates and how long they live.

When you write `@Component` or `@Bean`, Spring creates **one instance** by default (singleton). But you can change this behavior.

---

## The 6 Bean Scopes

| Scope | Annotation | Instances | Lifetime | Available In |
|-------|-----------|-----------|----------|--------------|
| **singleton** | (default) | 1 per container | Entire app lifetime | All apps |
| **prototype** | `@Scope("prototype")` | New one every time | Until garbage collected | All apps |
| **request** | `@RequestScope` | 1 per HTTP request | Single request | Web apps only |
| **session** | `@SessionScope` | 1 per HTTP session | User session | Web apps only |
| **application** | `@ApplicationScope` | 1 per ServletContext | App lifetime | Web apps only |
| **websocket** | `@Scope("websocket")` | 1 per WebSocket | WebSocket session | WebSocket apps |

You'll use **singleton** (90%+) and **prototype** (occasionally). The web scopes come later when we build REST APIs.

---

## Singleton Scope (Default)

```java
@Component  // Singleton by default — ONE instance shared everywhere
public class DatabaseConnection {
    // Every class that injects this gets the SAME object
}
```

**Characteristics:**
- Created once when the container starts (eager)
- Shared across ALL injection points
- Lives until the container shuts down
- @PreDestroy IS called on shutdown
- **Thread-safety is YOUR responsibility** (multiple threads share same instance)

**When to use:** Stateless services, repositories, configurations — most beans.

---

## Prototype Scope

```java
@Component
@Scope("prototype")  // NEW instance every time it's requested
public class ShoppingCart {
    private List<Item> items = new ArrayList<>();
    // Each user gets their own cart — not shared
}
```

**Characteristics:**
- New instance created **every time** `getBean()` is called or it's injected
- Spring does NOT manage its full lifecycle — **@PreDestroy is NOT called**
- You must clean up prototype beans yourself
- Good for stateful, short-lived objects

**When to use:** Stateful objects, builders, objects that hold per-request/per-user data.

---

## The Prototype Trap ⚠️

The most common mistake in Spring:

```java
@Component  // Singleton (default)
public class OrderService {

    @Autowired
    private ShoppingCart cart;  // Prototype — BUT injected only ONCE!

    public void placeOrder() {
        cart.addItem(...);  // PROBLEM: same cart instance forever!
    }
}
```

**Why?** A singleton is created once → its dependencies are injected once → the prototype is fetched once → same instance forever.

### Solutions:

**1. Inject `ObjectFactory<T>` or `Provider<T>`:**
```java
@Component
public class OrderService {

    @Autowired
    private ObjectFactory<ShoppingCart> cartFactory;

    public void placeOrder() {
        ShoppingCart cart = cartFactory.getObject();  // New instance each time!
    }
}
```

**2. Inject `ApplicationContext` and call `getBean()`:**
```java
@Component
public class OrderService {

    @Autowired
    private ApplicationContext context;

    public void placeOrder() {
        ShoppingCart cart = context.getBean(ShoppingCart.class);  // New each time
    }
}
```

**3. Use `@Lookup` (method injection):**
```java
@Component
public abstract class OrderService {

    @Lookup
    protected abstract ShoppingCart getCart();  // Spring overrides this method

    public void placeOrder() {
        ShoppingCart cart = getCart();  // New instance each time!
    }
}
```

---

## Scope & Lifecycle Summary

| | Singleton | Prototype |
|--|-----------|-----------|
| Instances | 1 | Many |
| Created when | Container starts | Each time requested |
| @PostConstruct | ✅ Called | ✅ Called |
| @PreDestroy | ✅ Called | ❌ NOT called |
| Thread-safe? | You handle it | Naturally (not shared) |
| Memory | Fixed (1 instance) | Can grow (many instances) |

---

## @Scope with proxyMode

For web scopes or injecting narrow-scoped beans into broader scopes:

```java
@Component
@Scope(value = "prototype", proxyMode = ScopedProxyMode.TARGET_CLASS)
public class ShoppingCart { }
```

This creates a **CGLIB proxy** that delegates to a new prototype instance per access. Solves the "prototype in singleton" problem without `ObjectFactory`.

---

## Key Takeaways

1. **Singleton** = default, one instance, shared everywhere, you manage thread-safety
2. **Prototype** = new instance every request, Spring WON'T destroy it for you
3. **Prototype Trap** = injecting prototype into singleton gives you only one instance — use `ObjectFactory`, `Provider`, or `@Lookup`
4. **Web scopes** (request, session) = one per HTTP request/session — we'll use these in Phase 5+
5. **Rule of thumb:** If a bean has state → consider prototype or web scope. If stateless → singleton.
