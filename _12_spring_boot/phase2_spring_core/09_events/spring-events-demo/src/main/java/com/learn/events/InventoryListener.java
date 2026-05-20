package com.learn.events;

import org.springframework.context.event.EventListener;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

/**
 * LISTENER 2: Updates inventory.
 * @Order(2) — runs second.
 */
@Component
public class InventoryListener {

    @EventListener
    @Order(2)
    public void onOrderPlaced(OrderPlacedEvent event) {
        System.out.println("    📋 InventoryListener [Order 2]: Reserving stock for order " + event.getOrderId());
    }
}
