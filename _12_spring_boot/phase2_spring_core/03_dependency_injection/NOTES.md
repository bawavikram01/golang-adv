# Phase 2.3 — Dependency Injection (DI)

---

## What is DI?

**Dependency Injection = the mechanism that delivers dependencies to your class from outside.**

IoC is the *principle* ("don't create your own deps"). DI is the *implementation* ("here's how deps arrive").

---

## The Three Injection Methods

### 1. Constructor Injection ✅ (RECOMMENDED)

```java
@Service
public class OrderService {
    private final UserRepository userRepo;     // final = immutable
    private final PaymentGateway payment;

    // Spring calls this constructor and passes matching beans
    public OrderService(UserRepository userRepo, PaymentGateway payment) {
        this.userRepo = userRepo;
        this.payment = payment;
    }
}
```

**Why it's the best:**
| Advantage | Explanation |
|-----------|-------------|
| Immutability | Fields are `final` — can't be changed after construction |
| Required deps are obvious | If it's in the constructor, it's mandatory |
| Testable | Just `new OrderService(mockRepo, mockPayment)` in tests |
| No reflection tricks | Pure Java — no Spring annotations needed on constructor |
| Fail fast | App won't start if dependency is missing |
| No partial construction | Object is always fully initialized |

**Note:** Since Spring 4.3, if a class has **only one constructor**, `@Autowired` is optional. Spring infers it automatically.

---

### 2. Setter Injection ⚠️ (For Optional Dependencies)

```java
@Service
public class ReportService {
    private NotificationService notifier;  // NOT final — can be null

    // @Autowired on setter — Spring calls this after construction
    @Autowired(required = false)  // required=false means it's OPTIONAL
    public void setNotifier(NotificationService notifier) {
        this.notifier = notifier;
    }

    public void generateReport() {
        // Must null-check because dep is optional!
        if (notifier != null) {
            notifier.send("admin", "Report ready");
        }
    }
}
```

**When to use:**
- Dependency is **optional** (the class works without it)
- You need to **re-inject** a dependency later (rare)
- Legacy code that requires a no-arg constructor

**Downsides:**
- Fields can't be `final` (mutable)
- Object can be in an incomplete state
- Must null-check optional deps

---

### 3. Field Injection ❌ (AVOID)

```java
@Service
public class UserService {
    @Autowired
    private UserRepository userRepo;  // Directly injected by Spring via reflection

    @Autowired
    private EmailService emailService;
}
```

**Why it's bad:**
| Problem | Explanation |
|---------|-------------|
| Can't use `final` | Reflection sets the field after construction |
| Hidden dependencies | Can't see deps from constructor signature |
| Untestable without Spring | Can't `new UserService()` in a unit test — fields are null |
| Violates SRP | Easy to add too many deps (no constructor pain) |
| Circular dep hiding | Masks circular dependencies until runtime |

**Only acceptable:** In test classes (`@SpringBootTest`) for convenience.

---

## Comparison Summary

| | Constructor ✅ | Setter ⚠️ | Field ❌ |
|---|---|---|---|
| **Fields `final`?** | Yes | No | No |
| **Required deps?** | Mandatory | Optional | Mandatory (but hidden) |
| **Testable?** | Easy | Moderate | Hard (needs Spring) |
| **Immutable?** | Yes | No | No |
| **Visible deps?** | Clear in signature | Scattered setters | Hidden |
| **Spring annotation needed?** | No (single ctor) | `@Autowired` | `@Autowired` |
| **Use when?** | Always (default choice) | Optional deps only | Tests only |

---

## What Happens Under the Hood

When Spring creates a bean with constructor injection:

```
1. Spring finds UserService has @Service
2. Spring reads its constructor: UserService(UserRepository, EmailService)
3. Spring looks in its container:
     - "Is there a bean of type UserRepository?" → YES → use it
     - "Is there a bean of type EmailService?" → YES → use it
4. Spring calls: new UserService(userRepoBean, emailServiceBean)
5. Stores the result as a singleton
```

With field injection:
```
1. Spring finds UserService has @Service  
2. Spring calls: new UserService() (no-arg constructor)
3. Spring uses REFLECTION to find fields with @Autowired
4. Spring calls: field.setAccessible(true); field.set(bean, dependency)
5. Stores the result
```

Constructor injection is cleaner, faster, and doesn't need reflection hacks.

---

## Multiple Beans of Same Type — Resolving Ambiguity

When Spring finds 2+ beans that match a dependency's type:

```java
interface MessageSender { void send(String msg); }

@Component class EmailSender implements MessageSender { ... }
@Component class SmsSender implements MessageSender { ... }

@Service
class AlertService {
    // ERROR: Which one to inject? Spring doesn't know!
    public AlertService(MessageSender sender) { ... }
}
```

**Solutions (in order of preference):**

### `@Primary` — "This is the default"
```java
@Component
@Primary  // If in doubt, use this one
class EmailSender implements MessageSender { ... }
```

### `@Qualifier` — "I want this specific one"
```java
@Service
class AlertService {
    public AlertService(@Qualifier("smsSender") MessageSender sender) { ... }
    // Bean names default to camelCase class name: SmsSender → "smsSender"
}
```

### Parameter name matching — "Name matches bean name"
```java
@Service
class AlertService {
    // If parameter is named "smsSender", Spring matches by name
    public AlertService(MessageSender smsSender) { ... }
}
```

### Injecting ALL implementations
```java
@Service
class AlertService {
    private final List<MessageSender> allSenders;  // Gets ALL implementations!

    public AlertService(List<MessageSender> allSenders) {
        this.allSenders = allSenders;  // [EmailSender, SmsSender]
    }
}
```

---

## Circular Dependencies

```java
@Component class A { public A(B b) {} }  // A needs B
@Component class B { public B(A a) {} }  // B needs A → CIRCULAR!
```

**Spring Boot 2.6+ rejects this at startup** (good!).

Fix options:
1. **Redesign** (best) — Extract shared logic into a third class C
2. `@Lazy` on one dependency — Creates a proxy, breaks the cycle
3. Setter injection on one side — Allows partial construction

```java
@Component
class A {
    public A(@Lazy B b) { ... }  // Spring injects a lazy proxy, not the real B
}
```

---

## Key Takeaways

1. **Always use constructor injection** — Immutable, testable, clear dependencies
2. **Setter injection** — Only for truly optional dependencies
3. **Field injection** — Avoid in production code, OK in tests
4. **`@Autowired` is optional** on the sole constructor (Spring 4.3+)
5. **Resolve ambiguity** with `@Primary`, `@Qualifier`, or `List<T>`
6. **Circular deps = design smell** — Refactor, don't hack around
