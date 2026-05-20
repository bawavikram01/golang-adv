package com.learn.events;

import org.springframework.context.event.EventListener;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

import java.util.UUID;

/**
 * LISTENER 3: Processes payment AND returns a new event (chaining).
 * @Order(3) — runs third.
 *
 * EVENT CHAINING: returning an object from @EventListener publishes it as a new event!
 */
@Component
public class PaymentListener {

    @EventListener
    @Order(3)
    public PaymentProcessedEvent onOrderPlaced(OrderPlacedEvent event) {
        String txnId = "TXN-" + UUID.randomUUID().toString().substring(0, 8);
        System.out.println("    💳 PaymentListener [Order 3]: Charging $" + String.format("%.2f", event.getAmount())
                + " → transaction " + txnId);

        // Returning a new event → Spring will publish it automatically!
        return new PaymentProcessedEvent(event.getOrderId(), txnId);
    }
}
