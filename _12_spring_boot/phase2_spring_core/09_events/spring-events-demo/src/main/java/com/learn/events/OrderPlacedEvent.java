package com.learn.events;

import java.time.LocalDateTime;

/**
 * CUSTOM EVENT — a plain POJO carrying data.
 * No need to extend ApplicationEvent (since Spring 4.2).
 */
public class OrderPlacedEvent {

    private final String orderId;
    private final String customerName;
    private final double amount;
    private final LocalDateTime timestamp;

    public OrderPlacedEvent(String orderId, String customerName, double amount) {
        this.orderId = orderId;
        this.customerName = customerName;
        this.amount = amount;
        this.timestamp = LocalDateTime.now();
    }

    public String getOrderId() { return orderId; }
    public String getCustomerName() { return customerName; }
    public double getAmount() { return amount; }
    public LocalDateTime getTimestamp() { return timestamp; }

    @Override
    public String toString() {
        return String.format("OrderPlacedEvent{id=%s, customer=%s, amount=$%.2f}",
                orderId, customerName, amount);
    }
}
