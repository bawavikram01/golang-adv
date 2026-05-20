# Phase 2.10 — Spring AOP (Aspect-Oriented Programming)

---

## What is AOP?

AOP lets you add **cross-cutting concerns** to your code without modifying the actual classes:
- Logging (who called what, when, how long)
- Security (is this user authorized?)
- Transactions (begin/commit/rollback)
- Caching (return cached value instead of executing)
- Performance monitoring (time method execution)

Without AOP, you'd repeat the same logging/security/timing code in every method. AOP extracts that into a separate **Aspect** that applies automatically.

---

## AOP Terminology

| Term | Meaning | Example |
|------|---------|---------|
| **Aspect** | A class containing cross-cutting logic | `@Aspect` class `LoggingAspect` |
| **Advice** | The action taken (the code that runs) | The method inside the aspect |
| **Join Point** | A point in execution where advice can apply | Any method call |
| **Pointcut** | An expression that selects join points | `"execution(* com.learn.aop.service.*.*(..))"` |
| **Target** | The object being advised | Your `OrderService` bean |
| **Proxy** | The wrapper Spring creates around the target | CGLIB/JDK dynamic proxy |
| **Weaving** | The process of applying aspects | At runtime (Spring uses proxies) |

---

## How Spring AOP Works (Proxy-Based)

```
You call:        orderService.placeOrder("ORD-1")
                        │
                        ▼
              ┌──────────────────┐
              │   CGLIB PROXY    │  ← Spring creates this automatically
              │                  │
              │  1. @Before      │  ← aspect advice runs
              │  2. target.placeOrder()  ← real method
              │  3. @After       │  ← aspect advice runs
              │                  │
              └──────────────────┘
```

**Key insight:** When you `@Autowired` a bean that has aspects applied, you get the **proxy**, not the real object. The proxy intercepts calls and runs aspect advice.

---

## Advice Types

| Annotation | When it runs | Use Case |
|-----------|-------------|----------|
| `@Before` | Before the method executes | Logging args, security checks |
| `@After` | After method (regardless of outcome) | Cleanup, finally-style |
| `@AfterReturning` | After method returns successfully | Logging result, caching |
| `@AfterThrowing` | After method throws exception | Error logging, alerting |
| `@Around` | Wraps the method entirely | Timing, transactions, caching, retry |

### @Around is the most powerful:
```java
@Around("pointcutExpression")
public Object measure(ProceedingJoinPoint joinPoint) throws Throwable {
    long start = System.currentTimeMillis();

    Object result = joinPoint.proceed();  // ← calls the real method

    long duration = System.currentTimeMillis() - start;
    System.out.println("Took " + duration + "ms");
    return result;
}
```

---

## Pointcut Expressions

Pointcuts select WHICH methods the aspect applies to:

### execution() — most common:
```java
// Any method in service package
@Before("execution(* com.learn.aop.service.*.*(..))")

// Breakdown:
// execution(
//   *                          ← any return type
//   com.learn.aop.service.     ← package
//   *                          ← any class
//   .*                         ← any method
//   (..)                       ← any parameters
// )

// Only void methods
@Before("execution(void com.learn.aop.service.*.*(..))")

// Methods starting with "get"
@Before("execution(* com.learn.aop.service.*.get*(..))")

// Specific method with specific param types
@Before("execution(* com.learn.aop.service.OrderService.placeOrder(String, double))")
```

### @annotation() — match by annotation:
```java
// Any method annotated with @Timed
@Around("@annotation(com.learn.aop.Timed)")
```

### within() — match by class:
```java
// All methods in OrderService
@Before("within(com.learn.aop.service.OrderService)")

// All methods in any class in service package
@Before("within(com.learn.aop.service.*)")
```

### Combining pointcuts:
```java
@Before("execution(* com.learn.aop.service.*.*(..)) && !execution(* *.get*(..))")
// All service methods EXCEPT getters
```

---

## Named Pointcuts (Reusable)

```java
@Aspect
@Component
public class LoggingAspect {

    // Define reusable pointcut
    @Pointcut("execution(* com.learn.aop.service.*.*(..))")
    public void serviceMethods() {}  // Method name = pointcut name

    @Pointcut("@annotation(com.learn.aop.Timed)")
    public void timedMethods() {}

    // Use the named pointcut
    @Before("serviceMethods()")
    public void logBefore(JoinPoint jp) { ... }

    @After("serviceMethods()")
    public void logAfter(JoinPoint jp) { ... }

    @Around("timedMethods()")
    public Object time(ProceedingJoinPoint pjp) { ... }
}
```

---

## JoinPoint — Accessing Method Details

```java
@Before("serviceMethods()")
public void logBefore(JoinPoint joinPoint) {
    String className = joinPoint.getTarget().getClass().getSimpleName();
    String method = joinPoint.getSignature().getName();
    Object[] args = joinPoint.getArgs();

    System.out.println(className + "." + method + "(" + Arrays.toString(args) + ")");
}
```

---

## Custom Annotations + AOP

Create your own annotations and apply aspects to them:

```java
// Define annotation
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
public @interface Timed { }

// Use it
@Service
public class OrderService {
    @Timed
    public void placeOrder(String id) { ... }
}

// Aspect that targets it
@Aspect @Component
public class TimingAspect {
    @Around("@annotation(com.learn.aop.Timed)")
    public Object time(ProceedingJoinPoint pjp) throws Throwable {
        long start = System.nanoTime();
        Object result = pjp.proceed();
        long ms = (System.nanoTime() - start) / 1_000_000;
        System.out.println(pjp.getSignature().getName() + " took " + ms + "ms");
        return result;
    }
}
```

---

## AOP Limitations in Spring

1. **Only works on Spring beans** — plain `new MyClass()` objects won't be proxied
2. **Only public methods** — private/protected methods can't be intercepted
3. **Self-invocation doesn't work** — calling `this.method()` bypasses the proxy
4. **Proxy-based** — not true bytecode weaving (use AspectJ for that)

```java
@Service
public class MyService {
    public void methodA() {
        this.methodB();  // ⚠️ This bypasses the proxy! Aspects on methodB won't run!
    }

    @Timed
    public void methodB() { ... }
}
```

---

## Spring's Built-in AOP Usage

These Spring features ALL use AOP internally:
- `@Transactional` — wraps method in begin/commit/rollback
- `@Cacheable` — checks cache before executing method
- `@Async` — runs method in a separate thread
- `@Secured` / `@PreAuthorize` — security checks before execution
- `@Retryable` — retries on failure

---

## Key Takeaways

1. **AOP** = extract cross-cutting concerns into reusable Aspects
2. **Spring AOP** = proxy-based (CGLIB wraps your beans)
3. **@Around** = most powerful (wraps entire method, controls execution)
4. **Pointcuts** = expressions that select which methods to advise
5. **Custom annotations** = cleanest approach (`@Timed`, `@Logged`, `@Audited`)
6. **Limitations**: only beans, only public, no self-invocation
7. `@Transactional`, `@Cacheable`, `@Async` — all just AOP under the hood
