package com.learn.events;

import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class EventsDemoApp implements CommandLineRunner {

    private final OrderService orderService;

    public EventsDemoApp(OrderService orderService) {
        this.orderService = orderService;
    }

    public static void main(String[] args) {
        SpringApplication.run(EventsDemoApp.class, args).close();
    }

    @Override
    public void run(String... args) {
        System.out.println("\n========================================");
        System.out.println("  SPRING EVENTS DEMO");
        System.out.println("========================================\n");

        // ─── Demo 1: Normal order (< $50, VipListener won't fire) ───
        System.out.println("─── ORDER 1: Small order ($29.99) ──────────────");
        System.out.println("  (VipListener has condition: amount > $50)");
        System.out.println();
        orderService.placeOrder("ORD-001", "Alice", 29.99);
        System.out.println();

        // ─── Demo 2: Large order (> $50, VipListener WILL fire) ───
        System.out.println("─── ORDER 2: Large order ($149.99) ─────────────");
        System.out.println("  (VipListener WILL fire for this one)");
        System.out.println();
        orderService.placeOrder("ORD-002", "Bob", 149.99);
        System.out.println();

        // ─── Summary ───
        System.out.println("─── SUMMARY ────────────────────────────────────");
        System.out.println();
        System.out.println("  What happened:");
        System.out.println("    1. OrderService published OrderPlacedEvent");
        System.out.println("    2. EmailListener reacted (Order 1 — runs first)");
        System.out.println("    3. InventoryListener reacted (Order 2 — runs second)");
        System.out.println("    4. PaymentListener reacted (Order 3) → returned PaymentProcessedEvent");
        System.out.println("    5. ReceiptListener reacted to the CHAINED PaymentProcessedEvent");
        System.out.println("    6. VipListener only fired for order > $50 (conditional)");
        System.out.println();
        System.out.println("  Key points:");
        System.out.println("    • OrderService has ZERO knowledge of listeners");
        System.out.println("    • Adding new behavior = add a new listener (Open/Closed Principle)");
        System.out.println("    • Listeners are synchronous by default (publisher waits)");
        System.out.println("    • @Order controls execution sequence");
        System.out.println("    • Returning from @EventListener = chaining (publishes new event)");
        System.out.println();

        System.out.println("─── SHUTTING DOWN (ContextClosedEvent fires) ───\n");
    }
}
