package com.learn.events;

/**
 * A second event — published as a CHAIN from a listener.
 * Demonstrates that listeners can return new events.
 */
public class PaymentProcessedEvent {

    private final String orderId;
    private final String transactionId;

    public PaymentProcessedEvent(String orderId, String transactionId) {
        this.orderId = orderId;
        this.transactionId = transactionId;
    }

    public String getOrderId() { return orderId; }
    public String getTransactionId() { return transactionId; }
}
