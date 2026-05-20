# Phase 2.9 — Spring Events

---

## What Are Spring Events?

Spring's event system implements the **Observer/Pub-Sub pattern** within the container:
- **Publisher** emits an event (doesn't know who's listening)
- **Listener(s)** react to the event (don't know who published)
- **Decoupled**: publisher and listeners have no direct dependency on each other

This is powerful for cross-cutting concerns: logging, notifications, audit trails, cache invalidation — without polluting your business logic.

---

## The Three Pieces

```
┌─────────────┐         ┌─────────────────────┐         ┌─────────────┐
│  Publisher   │ ──emit──▶ ApplicationEventPublisher ──▶ │  Listener   │
│ (any bean)  │         │    (the container)         │  │ (@EventListener)│
└─────────────┘         └─────────────────────┘         └─────────────┘
```

1. **Event** — a plain Java object (data carrier)
2. **Publisher** — any bean that calls `publisher.publishEvent(event)`
3. **Listener** — any bean method annotated with `@EventListener`

---

## Creating a Custom Event

```java
// Simple POJO — no need to extend anything (since Spring 4.2)
public class OrderPlacedEvent {
    private final String orderId;
    private final double amount;
    private final LocalDateTime timestamp;

    public OrderPlacedEvent(String orderId, double amount) {
        this.orderId = orderId;
        this.amount = amount;
        this.timestamp = LocalDateTime.now();
    }
    // getters...
}
```

Before Spring 4.2, events had to extend `ApplicationEvent`. Now any object works.

---

## Publishing Events

```java
@Service
public class OrderService {

    private final ApplicationEventPublisher publisher;

    public OrderService(ApplicationEventPublisher publisher) {
        this.publisher = publisher;  // Inject the publisher
    }

    public void placeOrder(String orderId, double amount) {
        // Business logic...
        System.out.println("Order " + orderId + " placed!");

        // Publish event — listeners will react
        publisher.publishEvent(new OrderPlacedEvent(orderId, amount));
    }
}
```

**Key:** `ApplicationEventPublisher` is always available — Spring provides it automatically.

---

## Listening to Events

### Method 1: @EventListener (recommended)

```java
@Component
public class EmailNotifier {

    @EventListener
    public void onOrderPlaced(OrderPlacedEvent event) {
        // Reacts to OrderPlacedEvent
        System.out.println("Sending email for order: " + event.getOrderId());
    }
}

@Component
public class InventoryUpdater {

    @EventListener
    public void handleOrder(OrderPlacedEvent event) {
        // Another listener for the SAME event
        System.out.println("Updating inventory for order: " + event.getOrderId());
    }
}
```

**Multiple listeners** can react to the same event — they're all invoked.

### Method 2: Conditional listening

```java
@EventListener(condition = "#event.amount > 100.0")
public void onLargeOrder(OrderPlacedEvent event) {
    // Only fires for orders > $100
}
```

### Method 3: Listen to multiple event types

```java
@EventListener({OrderPlacedEvent.class, OrderCancelledEvent.class})
public void onAnyOrderEvent(Object event) { ... }
```

---

## Synchronous vs Asynchronous Events

By default, events are **synchronous** — the publisher waits for ALL listeners to finish:

```
publisher.publishEvent(event)
    → listener1.handle(event)  // runs first
    → listener2.handle(event)  // runs second
    → returns to publisher     // only now continues
```

### Making events async:

```java
@Configuration
@EnableAsync
public class AsyncConfig { }

@Component
public class SlowListener {

    @Async
    @EventListener
    public void onOrder(OrderPlacedEvent event) {
        // Runs in a separate thread — publisher doesn't wait!
        Thread.sleep(5000); // Doesn't block the publisher
    }
}
```

---

## Built-in Spring Events

Spring publishes these automatically:

| Event | When |
|-------|------|
| `ContextRefreshedEvent` | ApplicationContext is initialized/refreshed |
| `ContextStartedEvent` | Context is started via `context.start()` |
| `ContextStoppedEvent` | Context is stopped via `context.stop()` |
| `ContextClosedEvent` | Context is closed (app shutdown) |
| `ApplicationReadyEvent` | App is fully started and ready (Spring Boot) |
| `ApplicationStartedEvent` | App has started but runners not called yet |

```java
@EventListener
public void onReady(ApplicationReadyEvent event) {
    System.out.println("Application is ready! Took " +
        event.getTimeTaken().toMillis() + "ms");
}
```

---

## Event Ordering

Control listener execution order:

```java
@Component
public class AuditListener {

    @EventListener
    @Order(1)  // Runs first
    public void audit(OrderPlacedEvent event) { ... }
}

@Component
public class NotifyListener {

    @EventListener
    @Order(2)  // Runs second
    public void notify(OrderPlacedEvent event) { ... }
}
```

Lower `@Order` value = higher priority (runs first).

---

## Event Chaining (Returning Events)

A listener can return a new event, which Spring will publish:

```java
@EventListener
public PaymentProcessedEvent onOrder(OrderPlacedEvent event) {
    // Process payment...
    return new PaymentProcessedEvent(event.getOrderId()); // This gets published!
}
```

---

## When to Use Events

| Good Use Cases | Bad Use Cases |
|---------------|---------------|
| Audit logging | Core business flow (hard to debug) |
| Sending notifications (email/SMS) | When you need a return value |
| Cache invalidation | When order of execution is critical |
| Updating search indexes | Simple method calls between 2 classes |
| Analytics/metrics | |

---

## Key Takeaways

1. **Events** = decoupled communication between beans (pub/sub)
2. **Event** = any POJO (no base class needed since Spring 4.2)
3. **Publisher** = inject `ApplicationEventPublisher`, call `.publishEvent()`
4. **Listener** = `@EventListener` on any bean method
5. **Sync by default** — publisher waits for all listeners
6. **@Async + @EnableAsync** = non-blocking event handling
7. **Multiple listeners** per event — add behavior without modifying publisher
8. Use events to **decouple cross-cutting concerns** from business logic
