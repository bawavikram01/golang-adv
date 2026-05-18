import java.util.*;
import java.util.function.Supplier;

/**
 * PHASE 1.3 — FACTORY PATTERN
 *
 * Creates objects WITHOUT you calling "new" directly.
 * Spring's ApplicationContext is essentially a giant factory.
 */

// ============================================================
// The products — different notification types
// ============================================================
interface Notification {
    void send(String to, String message);
}

class EmailNotification implements Notification {
    public void send(String to, String message) {
        System.out.println("  📧 Email to " + to + ": " + message);
    }
}

class SmsNotification implements Notification {
    public void send(String to, String message) {
        System.out.println("  📱 SMS to " + to + ": " + message);
    }
}

class PushNotification implements Notification {
    public void send(String to, String message) {
        System.out.println("  🔔 Push to " + to + ": " + message);
    }
}

class SlackNotification implements Notification {
    public void send(String to, String message) {
        System.out.println("  💬 Slack to #" + to + ": " + message);
    }
}


// ============================================================
// VERSION 1: Simple Factory (switch/if-else)
// ============================================================
class SimpleNotificationFactory {
    public static Notification create(String type) {
        return switch (type.toLowerCase()) {
            case "email" -> new EmailNotification();
            case "sms"   -> new SmsNotification();
            case "push"  -> new PushNotification();
            case "slack" -> new SlackNotification();
            default -> throw new IllegalArgumentException("Unknown type: " + type);
        };
    }
}


// ============================================================
// VERSION 2: Registry-Based Factory (how Spring actually works!)
// ============================================================
class NotificationRegistry {
    // Map of name → how to create it (Supplier = factory function)
    private final Map<String, Supplier<Notification>> registry = new HashMap<>();

    // Register a creator (like Spring's @Bean or @Component scanning)
    public void register(String name, Supplier<Notification> creator) {
        registry.put(name, creator);
    }

    // Get a notification by name (like ApplicationContext.getBean("name"))
    public Notification get(String name) {
        Supplier<Notification> creator = registry.get(name);
        if (creator == null) {
            throw new IllegalArgumentException("No bean registered with name: " + name);
        }
        return creator.get();
    }

    public Set<String> getRegisteredNames() {
        return registry.keySet();
    }
}


// ============================================================
// VERSION 3: Spring-Style BeanFactory (simplified)
// ============================================================
class MiniSpringFactory {
    private final Map<String, Object> singletons = new HashMap<>();
    private final Map<String, Supplier<?>> beanDefinitions = new HashMap<>();

    // Like @Bean in @Configuration class
    public <T> void registerBean(String name, Supplier<T> creator) {
        beanDefinitions.put(name, creator);
    }

    // Like applicationContext.getBean(name)
    @SuppressWarnings("unchecked")
    public <T> T getBean(String name) {
        // Check if singleton already exists
        if (singletons.containsKey(name)) {
            System.out.println("    [Factory] Returning cached bean: " + name);
            return (T) singletons.get(name);
        }

        // Create new instance
        Supplier<?> creator = beanDefinitions.get(name);
        if (creator == null) {
            throw new RuntimeException("No bean definition for: " + name);
        }

        Object bean = creator.get();
        singletons.put(name, bean);  // Cache as singleton
        System.out.println("    [Factory] Created new bean: " + name);
        return (T) bean;
    }

    // Like @Autowired by type
    @SuppressWarnings("unchecked")
    public <T> T getBean(Class<T> type) {
        for (Object bean : singletons.values()) {
            if (type.isInstance(bean)) {
                return (T) bean;
            }
        }
        // Try creating from definitions
        for (Map.Entry<String, Supplier<?>> entry : beanDefinitions.entrySet()) {
            Object bean = entry.getValue().get();
            if (type.isInstance(bean)) {
                singletons.put(entry.getKey(), bean);
                System.out.println("    [Factory] Created bean by type: " + type.getSimpleName());
                return (T) bean;
            }
        }
        throw new RuntimeException("No bean of type: " + type.getSimpleName());
    }
}


public class Step2_Factory {
    public static void main(String[] args) {

        // ---- Simple Factory ----
        System.out.println("=== SIMPLE FACTORY ===");
        System.out.println("  (You say WHAT you want, factory creates it)\n");

        Notification n1 = SimpleNotificationFactory.create("email");
        Notification n2 = SimpleNotificationFactory.create("sms");
        Notification n3 = SimpleNotificationFactory.create("slack");

        n1.send("alice@mail.com", "Welcome!");
        n2.send("+1234567890", "Your OTP is 4521");
        n3.send("general", "Deploy succeeded!");


        // ---- Registry Factory ----
        System.out.println("\n=== REGISTRY FACTORY ===");
        System.out.println("  (Register creators, get by name — like Spring component scanning)\n");

        NotificationRegistry registry = new NotificationRegistry();
        registry.register("email", EmailNotification::new);       // Method reference as Supplier
        registry.register("sms", SmsNotification::new);
        registry.register("push", PushNotification::new);
        registry.register("slack", SlackNotification::new);

        System.out.println("  Registered: " + registry.getRegisteredNames());

        registry.get("email").send("bob@mail.com", "Invoice attached");
        registry.get("push").send("user-42", "New message from Alice");


        // ---- Spring-Style BeanFactory ----
        System.out.println("\n=== SPRING-STYLE BEAN FACTORY ===");
        System.out.println("  (This is what ApplicationContext does under the hood)\n");

        MiniSpringFactory ctx = new MiniSpringFactory();

        // Register bean definitions (like @Bean methods in a @Configuration class)
        ctx.registerBean("emailService", EmailNotification::new);
        ctx.registerBean("smsService", SmsNotification::new);
        ctx.registerBean("pushService", PushNotification::new);

        // First call → creates the bean
        Notification email = ctx.getBean("emailService");
        email.send("charlie@mail.com", "Hello from factory!");

        // Second call → returns cached singleton (same instance)
        Notification emailAgain = ctx.getBean("emailService");
        emailAgain.send("diana@mail.com", "Same instance!");

        System.out.println("\n  Same instance? " + (email == emailAgain));  // true!

        // Get by type (like @Autowired without @Qualifier)
        System.out.println();
        Notification sms = ctx.getBean("smsService");
        sms.send("+9876543210", "Got by name");

        System.out.println("\n=== KEY TAKEAWAY ===");
        System.out.println("  Factory pattern = someone else creates objects for you.");
        System.out.println("  Spring's ApplicationContext is a Factory + Singleton Registry.");
        System.out.println("  You declare beans (@Component, @Bean) → Spring creates them.");
        System.out.println("  You request beans (@Autowired) → Spring hands them to you.");
    }
}
