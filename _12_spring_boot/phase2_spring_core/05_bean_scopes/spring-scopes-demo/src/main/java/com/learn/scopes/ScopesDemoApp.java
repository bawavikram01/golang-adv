package com.learn.scopes;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;

@SpringBootApplication
public class ScopesDemoApp {

    public static void main(String[] args) {
        ConfigurableApplicationContext context = SpringApplication.run(ScopesDemoApp.class, args);

        System.out.println("========================================");
        System.out.println("  SPRING BEAN SCOPES DEMO");
        System.out.println("========================================\n");

        // Demo 1: Singleton Scope
        demoSingleton(context);

        // Demo 2: Prototype Scope
        demoPrototype(context);

        // Demo 3: The Prototype Trap
        demoPrototypeTrap(context);

        // Demo 4: Fixing the Prototype Trap with ObjectFactory
        demoPrototypeFix(context);

        context.close();
    }

    private static void demoSingleton(ConfigurableApplicationContext context) {
        System.out.println("─── SINGLETON SCOPE (default) ───────────────────");
        System.out.println();

        SingletonService s1 = context.getBean(SingletonService.class);
        SingletonService s2 = context.getBean(SingletonService.class);
        SingletonService s3 = context.getBean(SingletonService.class);

        System.out.println("  getBean() call 1: " + s1);
        System.out.println("  getBean() call 2: " + s2);
        System.out.println("  getBean() call 3: " + s3);
        System.out.println();
        System.out.println("  s1 == s2 == s3 ? " + (s1 == s2 && s2 == s3));
        System.out.println("  → SAME instance every time (shared)");
        System.out.println();
    }

    private static void demoPrototype(ConfigurableApplicationContext context) {
        System.out.println("─── PROTOTYPE SCOPE ─────────────────────────────");
        System.out.println();

        PrototypeService p1 = context.getBean(PrototypeService.class);
        PrototypeService p2 = context.getBean(PrototypeService.class);
        PrototypeService p3 = context.getBean(PrototypeService.class);

        System.out.println("  getBean() call 1: " + p1);
        System.out.println("  getBean() call 2: " + p2);
        System.out.println("  getBean() call 3: " + p3);
        System.out.println();
        System.out.println("  p1 == p2 ? " + (p1 == p2));
        System.out.println("  p2 == p3 ? " + (p2 == p3));
        System.out.println("  → DIFFERENT instance each time (not shared)");
        System.out.println();

        // Demonstrate statefulness
        p1.addItem("Laptop");
        p1.addItem("Mouse");
        p2.addItem("Keyboard");

        System.out.println("  p1 items (added Laptop, Mouse): " + p1.getItems());
        System.out.println("  p2 items (added Keyboard):      " + p2.getItems());
        System.out.println("  p3 items (added nothing):       " + p3.getItems());
        System.out.println("  → Each prototype has its OWN state");
        System.out.println();
    }

    private static void demoPrototypeTrap(ConfigurableApplicationContext context) {
        System.out.println("─── THE PROTOTYPE TRAP ⚠️ ───────────────────────");
        System.out.println();

        OrderServiceBroken service = context.getBean(OrderServiceBroken.class);

        System.out.println("  Calling placeOrder() 3 times on the SAME singleton service:");
        service.placeOrder("Order-1");
        service.placeOrder("Order-2");
        service.placeOrder("Order-3");
        System.out.println();
        System.out.println("  → PROBLEM: All orders use the SAME cart instance!");
        System.out.println("  → The prototype was injected once into the singleton");
        System.out.println();
    }

    private static void demoPrototypeFix(ConfigurableApplicationContext context) {
        System.out.println("─── FIX: ObjectFactory<T> ───────────────────────");
        System.out.println();

        OrderServiceFixed service = context.getBean(OrderServiceFixed.class);

        System.out.println("  Calling placeOrder() 3 times — each gets a FRESH cart:");
        service.placeOrder("Order-A");
        service.placeOrder("Order-B");
        service.placeOrder("Order-C");
        System.out.println();
        System.out.println("  → FIXED: Each order gets its own fresh prototype cart!");
        System.out.println();
    }
}
