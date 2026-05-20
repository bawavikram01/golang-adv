package com.learn.events;

import org.springframework.context.event.EventListener;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

/**
 * LISTENER 1: Sends email notification.
 * @Order(1) — runs first among listeners.
 */
@Component
public class EmailListener {

    @EventListener
    @Order(1)
    public void onOrderPlaced(OrderPlacedEvent event) {
        System.out.println("    ✉️  EmailListener [Order 1]: Sending confirmation email to " + event.getCustomerName()
                + " for order " + event.getOrderId());
    }
}
