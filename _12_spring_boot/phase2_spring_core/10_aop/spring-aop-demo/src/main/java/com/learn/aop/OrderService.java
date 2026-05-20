package com.learn.aop;

import org.springframework.stereotype.Service;

/**
 * A normal service class — has NO AOP-related code.
 * The aspects apply transparently via proxying.
 */
@Service
public class OrderService {

    @Timed
    @Audited(action = "PLACE_ORDER")
    public String placeOrder(String orderId, double amount) {
        simulateWork(50);  // Simulate some processing time
        return "Order " + orderId + " placed ($" + String.format("%.2f", amount) + ")";
    }

    @Timed
    public String cancelOrder(String orderId) {
        simulateWork(30);
        return "Order " + orderId + " cancelled";
    }

    public String getOrderStatus(String orderId) {
        // No @Timed — timing aspect won't apply here
        return "Order " + orderId + " status: SHIPPED";
    }

    private void simulateWork(long ms) {
        try { Thread.sleep(ms); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
    }
}
