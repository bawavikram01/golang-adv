/*
 * =============================================================
 * BEHAVIORAL PATTERN 2: OBSERVER
 * =============================================================
 *
 * INTENT: Define a one-to-many dependency so that when one object
 *         changes state, all dependents are notified automatically.
 *
 * ANALOGY: YouTube subscriptions — when a channel uploads,
 *          ALL subscribers get notified.
 *
 * USE WHEN:
 *   - One object's state change should trigger updates in others
 *   - Event systems, pub-sub, listeners
 *   - GUI event handling, notifications
 *
 * REAL EXAMPLES: Java EventListener, RxJava, MVC pattern,
 *                PropertyChangeListener, Spring ApplicationEvent
 */

import java.util.*;

public class ObserverPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // YouTube-style notification system
        // ═══════════════════════════════════════════════════════
        System.out.println("=== YOUTUBE OBSERVER ===");

        YouTubeChannel techChannel = new YouTubeChannel("TechGuru");

        Subscriber alice = new Subscriber("Alice");
        Subscriber bob = new Subscriber("Bob");
        Subscriber charlie = new Subscriber("Charlie");

        // Subscribe
        techChannel.subscribe(alice);
        techChannel.subscribe(bob);
        techChannel.subscribe(charlie);

        // Upload triggers notification to ALL subscribers
        techChannel.uploadVideo("Java Design Patterns Tutorial");

        System.out.println();

        // Bob unsubscribes
        techChannel.unsubscribe(bob);
        techChannel.uploadVideo("Spring Boot Crash Course");

        // ═══════════════════════════════════════════════════════
        // Real-world: Stock Price Observer
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== STOCK PRICE OBSERVER ===");

        StockExchange exchange = new StockExchange();

        StockDisplay mobileApp = new MobileStockApp("Alice's Phone");
        StockDisplay webDashboard = new WebDashboard("Trading Dashboard");
        StockDisplay alertSystem = new PriceAlertSystem(150.0);  // alert if > 150

        exchange.addObserver(mobileApp);
        exchange.addObserver(webDashboard);
        exchange.addObserver(alertSystem);

        exchange.updatePrice("AAPL", 145.50);
        System.out.println();
        exchange.updatePrice("AAPL", 152.75);  // triggers alert!
        System.out.println();
        exchange.updatePrice("GOOGL", 2800.00);

        // ═══════════════════════════════════════════════════════
        // Event System (generic, reusable)
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GENERIC EVENT SYSTEM ===");

        EventBus eventBus = new EventBus();

        // Different listeners for different events
        eventBus.on("user.login", data -> System.out.println("  📊 Analytics: User logged in: " + data));
        eventBus.on("user.login", data -> System.out.println("  📧 Welcome email sent to: " + data));
        eventBus.on("order.placed", data -> System.out.println("  📦 Fulfillment: Processing " + data));
        eventBus.on("order.placed", data -> System.out.println("  📧 Confirmation email for: " + data));

        eventBus.emit("user.login", "alice@example.com");
        System.out.println();
        eventBus.emit("order.placed", "Order #12345");
    }
}

// ═══════════════════════════════════════════════════════════════
// YOUTUBE OBSERVER
// ═══════════════════════════════════════════════════════════════
interface Observer {
    void update(String channelName, String videoTitle);
}

class YouTubeChannel {
    private String name;
    private List<Observer> subscribers = new ArrayList<>();

    public YouTubeChannel(String name) {
        this.name = name;
    }

    public void subscribe(Observer observer) {
        subscribers.add(observer);
        System.out.println("  ✓ " + observer + " subscribed to " + name);
    }

    public void unsubscribe(Observer observer) {
        subscribers.remove(observer);
        System.out.println("  ✗ " + observer + " unsubscribed from " + name);
    }

    public void uploadVideo(String title) {
        System.out.println("  🎬 " + name + " uploaded: \"" + title + "\"");
        notifySubscribers(title);
    }

    private void notifySubscribers(String title) {
        for (Observer sub : subscribers) {
            sub.update(name, title);
        }
    }
}

class Subscriber implements Observer {
    private String name;

    public Subscriber(String name) { this.name = name; }

    @Override
    public void update(String channelName, String videoTitle) {
        System.out.println("  🔔 " + name + " notified: " + channelName + " → \"" + videoTitle + "\"");
    }

    @Override
    public String toString() { return name; }
}

// ═══════════════════════════════════════════════════════════════
// STOCK PRICE OBSERVER
// ═══════════════════════════════════════════════════════════════
interface StockDisplay {
    void onPriceUpdate(String symbol, double price);
}

class StockExchange {
    private List<StockDisplay> observers = new ArrayList<>();
    private Map<String, Double> prices = new HashMap<>();

    public void addObserver(StockDisplay observer) {
        observers.add(observer);
    }

    public void removeObserver(StockDisplay observer) {
        observers.remove(observer);
    }

    public void updatePrice(String symbol, double price) {
        prices.put(symbol, price);
        System.out.println("  📈 " + symbol + " → $" + price);
        for (StockDisplay obs : observers) {
            obs.onPriceUpdate(symbol, price);
        }
    }
}

class MobileStockApp implements StockDisplay {
    private String device;
    public MobileStockApp(String device) { this.device = device; }

    @Override
    public void onPriceUpdate(String symbol, double price) {
        System.out.println("  📱 [" + device + "] " + symbol + ": $" + price);
    }
}

class WebDashboard implements StockDisplay {
    private String name;
    public WebDashboard(String name) { this.name = name; }

    @Override
    public void onPriceUpdate(String symbol, double price) {
        System.out.println("  🖥️ [" + name + "] Updated " + symbol + " chart: $" + price);
    }
}

class PriceAlertSystem implements StockDisplay {
    private double threshold;
    public PriceAlertSystem(double threshold) { this.threshold = threshold; }

    @Override
    public void onPriceUpdate(String symbol, double price) {
        if (price > threshold) {
            System.out.println("  🚨 ALERT! " + symbol + " exceeded $" + threshold + "! Current: $" + price);
        }
    }
}

// ═══════════════════════════════════════════════════════════════
// GENERIC EVENT BUS (pub-sub)
// ═══════════════════════════════════════════════════════════════
interface EventListener {
    void handle(String data);
}

class EventBus {
    private Map<String, List<EventListener>> listeners = new HashMap<>();

    public void on(String event, EventListener listener) {
        listeners.computeIfAbsent(event, k -> new ArrayList<>()).add(listener);
    }

    public void off(String event, EventListener listener) {
        List<EventListener> list = listeners.get(event);
        if (list != null) list.remove(listener);
    }

    public void emit(String event, String data) {
        System.out.println("  ⚡ Event: " + event);
        List<EventListener> list = listeners.get(event);
        if (list != null) {
            for (EventListener l : list) {
                l.handle(data);
            }
        }
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Observer = Subject (publisher) + Observers (subscribers).
 * ✦ Subject maintains a list of observers and notifies them.
 * ✦ Observers register/unregister dynamically.
 * ✦ Loose coupling: subject doesn't know observer details.
 *
 * ✦ Variants:
 *   - Push model: subject sends data with notification
 *   - Pull model: subject notifies; observer queries for data
 *   - Event bus: decoupled pub-sub with event names
 *
 * ✦ This is the foundation of event-driven architecture.
 *
 * COMPILE & RUN:
 *   javac ObserverPattern.java && java ObserverPattern
 */
