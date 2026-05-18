import java.util.*;

/**
 * PHASE 1.3 — OBSERVER PATTERN
 *
 * When something happens, NOTIFY everyone who cares.
 * Publisher doesn't know who's listening — maximum decoupling.
 *
 * Spring's event system (@EventListener, ApplicationEvent) is this pattern.
 */

// ============================================================
// STEP 1: Define Events (what happened)
// ============================================================

// Base event — like Spring's ApplicationEvent
abstract class AppEvent {
    private final Object source;
    private final long timestamp;

    public AppEvent(Object source) {
        this.source = source;
        this.timestamp = System.currentTimeMillis();
    }

    public Object getSource() { return source; }
    public long getTimestamp() { return timestamp; }
}

// Specific events — these are your domain events
class UserRegisteredEvent extends AppEvent {
    private final String username;
    private final String email;

    public UserRegisteredEvent(Object source, String username, String email) {
        super(source);
        this.username = username;
        this.email = email;
    }

    public String getUsername() { return username; }
    public String getEmail() { return email; }
}

class OrderPlacedEvent extends AppEvent {
    private final String orderId;
    private final double amount;

    public OrderPlacedEvent(Object source, String orderId, double amount) {
        super(source);
        this.orderId = orderId;
        this.amount = amount;
    }

    public String getOrderId() { return orderId; }
    public double getAmount() { return amount; }
}


// ============================================================
// STEP 2: Event Listener interface (the observer/subscriber)
// ============================================================

interface EventListener<E extends AppEvent> {
    void onEvent(E event);
}


// ============================================================
// STEP 3: Event Publisher (the subject — like ApplicationEventPublisher)
// ============================================================

class EventPublisher {
    // Map of event type → list of listeners
    private final Map<Class<?>, List<EventListener<?>>> listeners = new HashMap<>();

    // Register a listener for a specific event type
    // In Spring: @EventListener does this automatically!
    public <E extends AppEvent> void subscribe(Class<E> eventType, EventListener<E> listener) {
        listeners.computeIfAbsent(eventType, k -> new ArrayList<>()).add(listener);
    }

    // Publish an event to all listeners of that type
    // In Spring: applicationEventPublisher.publishEvent(event)
    @SuppressWarnings("unchecked")
    public void publish(AppEvent event) {
        List<EventListener<?>> eventListeners = listeners.getOrDefault(event.getClass(), List.of());
        System.out.println("  [Publisher] Publishing " + event.getClass().getSimpleName()
            + " to " + eventListeners.size() + " listener(s)");
        for (EventListener<?> listener : eventListeners) {
            ((EventListener<AppEvent>) listener).onEvent(event);
        }
    }
}


// ============================================================
// STEP 4: Concrete Listeners (the people who react)
// ============================================================

// Listener 1: Send welcome email when user registers
class WelcomeEmailListener implements EventListener<UserRegisteredEvent> {
    public void onEvent(UserRegisteredEvent event) {
        System.out.println("    📧 [WelcomeEmail] Sending welcome email to: " + event.getEmail());
    }
}

// Listener 2: Create default settings when user registers
class DefaultSettingsListener implements EventListener<UserRegisteredEvent> {
    public void onEvent(UserRegisteredEvent event) {
        System.out.println("    ⚙️  [Settings] Creating default settings for: " + event.getUsername());
    }
}

// Listener 3: Notify admin when user registers
class AdminNotificationListener implements EventListener<UserRegisteredEvent> {
    public void onEvent(UserRegisteredEvent event) {
        System.out.println("    🔔 [AdminNotify] New user registered: " + event.getUsername());
    }
}

// Listener 4: Update inventory when order placed
class InventoryListener implements EventListener<OrderPlacedEvent> {
    public void onEvent(OrderPlacedEvent event) {
        System.out.println("    📦 [Inventory] Reserving stock for order: " + event.getOrderId());
    }
}

// Listener 5: Send confirmation when order placed
class OrderConfirmationListener implements EventListener<OrderPlacedEvent> {
    public void onEvent(OrderPlacedEvent event) {
        System.out.println("    ✉️  [Confirmation] Order " + event.getOrderId()
            + " confirmed — $" + event.getAmount());
    }
}

// Listener 6: Analytics tracking
class AnalyticsListener implements EventListener<OrderPlacedEvent> {
    public void onEvent(OrderPlacedEvent event) {
        System.out.println("    📊 [Analytics] Tracking revenue: +$" + event.getAmount());
    }
}


// ============================================================
// STEP 5: Service classes (publishers don't know about listeners)
// ============================================================

class RegistrationService {
    private final EventPublisher publisher;

    public RegistrationService(EventPublisher publisher) {
        this.publisher = publisher;
    }

    public void registerUser(String username, String email) {
        // Business logic
        System.out.println("\n  [RegistrationService] Registering user: " + username);
        // ... save to DB ...

        // Fire event — I DON'T KNOW and DON'T CARE who listens!
        publisher.publish(new UserRegisteredEvent(this, username, email));
    }
}

class OrderingService {
    private final EventPublisher publisher;

    public OrderingService(EventPublisher publisher) {
        this.publisher = publisher;
    }

    public void placeOrder(String orderId, double amount) {
        System.out.println("\n  [OrderingService] Placing order: " + orderId + " ($" + amount + ")");
        // ... save order to DB ...

        // Fire event
        publisher.publish(new OrderPlacedEvent(this, orderId, amount));
    }
}


public class Step5_Observer {
    public static void main(String[] args) {

        System.out.println("=== OBSERVER PATTERN (Spring Event System) ===\n");

        // 1. Create the event bus (like Spring's ApplicationContext)
        EventPublisher publisher = new EventPublisher();

        // 2. Register listeners (Spring does this via @EventListener annotation)
        System.out.println("--- Registering Listeners ---");
        publisher.subscribe(UserRegisteredEvent.class, new WelcomeEmailListener());
        publisher.subscribe(UserRegisteredEvent.class, new DefaultSettingsListener());
        publisher.subscribe(UserRegisteredEvent.class, new AdminNotificationListener());
        publisher.subscribe(OrderPlacedEvent.class, new InventoryListener());
        publisher.subscribe(OrderPlacedEvent.class, new OrderConfirmationListener());
        publisher.subscribe(OrderPlacedEvent.class, new AnalyticsListener());
        System.out.println("  3 listeners for UserRegisteredEvent");
        System.out.println("  3 listeners for OrderPlacedEvent");

        // 3. Create services (they only know about the publisher, not the listeners)
        RegistrationService regService = new RegistrationService(publisher);
        OrderingService orderService = new OrderingService(publisher);

        // 4. Business operations trigger events → listeners react automatically!
        System.out.println("\n--- Business Operations ---");
        regService.registerUser("alice", "alice@example.com");
        regService.registerUser("bob", "bob@example.com");
        orderService.placeOrder("ORD-001", 299.99);
        orderService.placeOrder("ORD-002", 59.99);


        System.out.println("\n\n=== THE DECOUPLING POWER ===");
        System.out.println("  RegistrationService has NO IDEA about:");
        System.out.println("    - WelcomeEmailListener");
        System.out.println("    - DefaultSettingsListener");
        System.out.println("    - AdminNotificationListener");
        System.out.println("  It just publishes an event. Done.");
        System.out.println();
        System.out.println("  Want to add a 4th reaction (e.g., log to audit trail)?");
        System.out.println("  → Add a new listener. ZERO changes to RegistrationService.");
        System.out.println("  This is the Open/Closed Principle in action.");

        System.out.println("\n=== HOW SPRING DOES IT ===");
        System.out.println("  ┌──────────────────────────────────────────────────┐");
        System.out.println("  │ Our EventPublisher → ApplicationEventPublisher   │");
        System.out.println("  │ Our @subscribe()   → @EventListener annotation  │");
        System.out.println("  │ Our AppEvent       → ApplicationEvent            │");
        System.out.println("  │                                                  │");
        System.out.println("  │ In Spring, you just write:                       │");
        System.out.println("  │                                                  │");
        System.out.println("  │   @EventListener                                 │");
        System.out.println("  │   public void handle(UserRegisteredEvent e) {    │");
        System.out.println("  │       // react here                              │");
        System.out.println("  │   }                                              │");
        System.out.println("  │                                                  │");
        System.out.println("  │ Spring auto-discovers and subscribes it!         │");
        System.out.println("  └──────────────────────────────────────────────────┘");

        System.out.println("\n=== KEY TAKEAWAY ===");
        System.out.println("  Observer = publish events, listeners react independently.");
        System.out.println("  Publisher doesn't know listeners. Listeners don't know each other.");
        System.out.println("  Add/remove listeners without touching business logic.");
        System.out.println("  Perfect for: notifications, audit logs, cache invalidation, async tasks.");
    }
}
