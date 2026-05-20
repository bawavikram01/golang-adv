package com.learn.events;

import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

/**
 * CONDITIONAL LISTENER — only fires when order amount > $50.
 * Demonstrates SpEL condition in @EventListener.
 */
@Component
public class VipListener {

    @EventListener(condition = "#event.amount > 50.0")
    public void onLargeOrder(OrderPlacedEvent event) {
        System.out.println("    ⭐ VipListener [CONDITIONAL]: Large order detected! $"
                + String.format("%.2f", event.getAmount()) + " → adding VIP points for " + event.getCustomerName());
    }
}
