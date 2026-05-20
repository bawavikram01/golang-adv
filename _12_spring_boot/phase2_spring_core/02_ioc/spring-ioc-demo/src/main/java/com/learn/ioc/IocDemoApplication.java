package com.learn.ioc;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;

/**
 * PHASE 2.2 — IoC IN ACTION
 *
 * Watch what Spring does:
 *   1. Scans for @Component classes
 *   2. Creates beans in dependency order
 *   3. Calls @PostConstruct after creation
 *   4. Injects dependencies (picks @Primary when multiple exist)
 *   5. On shutdown, calls @PreDestroy
 *
 * YOU control NONE of this. Spring does. That's IoC.
 */
@SpringBootApplication
public class IocDemoApplication {

    public static void main(String[] args) {
        System.out.println("\n╔══════════════════════════════════════════════════╗");
        System.out.println("║   PHASE 2.2 — IoC (Inversion of Control)        ║");
        System.out.println("╚══════════════════════════════════════════════════╝\n");

        System.out.println("=== CONTAINER STARTUP (Spring takes control) ===\n");

        // Spring creates and wires everything
        ConfigurableApplicationContext context = SpringApplication.run(IocDemoApplication.class, args);

        System.out.println("\n=== ALL 3 FORMS OF IoC IN ACTION ===\n");

        // ---- Form 1: Dependency Injection ----
        System.out.println("--- Form 1: Dependency Injection ---");
        System.out.println("  OrderService received its dependencies automatically.\n");

        OrderService orderService = context.getBean(OrderService.class);
        orderService.placeOrder("ORD-001", "alice@example.com");
        System.out.println();
        orderService.placeOrder("ORD-002", "bob@example.com");

        // ---- Form 2: Multiple implementations, Spring chooses ----
        System.out.println("\n--- IoC chose WHICH implementation to inject ---");
        NotificationService injected = context.getBean(OrderService.class)
            .toString().contains("Email") ? null : null; // just for explanation

        // Show both beans exist
        System.out.println("  Beans of type NotificationService in container:");
        context.getBeansOfType(NotificationService.class).forEach((name, bean) ->
            System.out.println("    • " + name + " → " + bean.getType())
        );
        System.out.println("  Spring injected EMAIL because it has @Primary");

        // ---- Form 3: Lifecycle (@PostConstruct already ran at startup) ----
        System.out.println("\n--- Form 3: Lifecycle Callbacks ---");
        DatabaseService db = context.getBean(DatabaseService.class);
        System.out.println("  DatabaseService.isConnected() = " + db.isConnected());
        System.out.println("  (@PostConstruct ran automatically at startup!)");

        // ---- Shutdown: @PreDestroy ----
        System.out.println("\n=== CONTAINER SHUTDOWN (Spring calls @PreDestroy) ===\n");
        context.close();  // Triggers @PreDestroy on all beans

        System.out.println("\n=== IoC SUMMARY ===");
        System.out.println("  1. YOU didn't call 'new' on any service");
        System.out.println("  2. YOU didn't decide creation order");
        System.out.println("  3. YOU didn't choose which NotificationService to use");
        System.out.println("  4. YOU didn't call initialize() or shutdown()");
        System.out.println("  5. Spring controlled ALL of it. That's Inversion of Control.");
    }
}
