package com.learn.events;

import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

/**
 * LISTENER 4: Reacts to the CHAINED PaymentProcessedEvent.
 * This demonstrates event chaining — one event triggers another.
 */
@Component
public class ReceiptListener {

    @EventListener
    public void onPaymentProcessed(PaymentProcessedEvent event) {
        System.out.println("    🧾 ReceiptListener: Generating receipt for order " + event.getOrderId()
                + " (txn: " + event.getTransactionId() + ")");
    }
}
