package com.learn.events;

import org.springframework.context.ApplicationEventPublisher;
import org.springframework.stereotype.Service;

/**
 * PUBLISHER — publishes events without knowing who listens.
 * 
 * OrderService only cares about placing the order.
 * Everything else (email, inventory, audit) is handled by listeners.
 * This keeps OrderService CLEAN and FOCUSED.
 */
@Service
public class OrderService {

    private final ApplicationEventPublisher publisher;

    public OrderService(ApplicationEventPublisher publisher) {
        this.publisher = publisher;
    }

    public void placeOrder(String orderId, String customer, double amount) {
        // ─── Business logic ───
        System.out.println("    📦 OrderService: Order " + orderId + " placed for " + customer + " ($" + String.format("%.2f", amount) + ")");

        // ─── Publish event — listeners react ───
        publisher.publishEvent(new OrderPlacedEvent(orderId, customer, amount));

        // This line runs AFTER all sync listeners have finished
        System.out.println("    📦 OrderService: publishEvent() returned (all sync listeners done)");
    }
}
