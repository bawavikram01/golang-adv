package com.learn.logging;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

/**
 * Demonstrates proper logging patterns.
 * Notice: NO System.out.println() — we use SLF4J Logger instead.
 */
@Service
public class OrderService {

    // Standard pattern: one static logger per class
    private static final Logger log = LoggerFactory.getLogger(OrderService.class);

    public String placeOrder(String orderId, String customer, double amount) {
        // DEBUG: detailed info for development/troubleshooting
        log.debug("Entering placeOrder() with orderId={}, customer={}, amount={}", orderId, customer, amount);

        // INFO: important business event
        log.info("Order {} placed by {} for ${}", orderId, customer, String.format("%.2f", amount));

        // Simulate some processing
        if (amount > 1000) {
            // WARN: unusual but not an error
            log.warn("Large order detected: {} (${}) — flagged for review", orderId, String.format("%.2f", amount));
        }

        log.debug("Exiting placeOrder() — returning confirmation for {}", orderId);
        return "Order " + orderId + " confirmed for " + customer;
    }

    public String cancelOrder(String orderId) {
        log.info("Cancelling order {}", orderId);

        try {
            if ("ORD-999".equals(orderId)) {
                throw new IllegalStateException("Order already shipped — cannot cancel");
            }
            log.info("Order {} successfully cancelled", orderId);
            return "Cancelled: " + orderId;
        } catch (Exception e) {
            // ERROR: pass exception as LAST arg for full stack trace
            log.error("Failed to cancel order {}: {}", orderId, e.getMessage(), e);
            throw e;
        }
    }

    public void processPayment(String orderId, double amount) {
        // TRACE: extremely detailed (rarely enabled in production)
        log.trace("processPayment() called — orderId={}", orderId);

        log.debug("Processing payment of ${} for order {}", String.format("%.2f", amount), orderId);

        // Simulate payment steps
        log.debug("Step 1: Validating card...");
        log.debug("Step 2: Charging amount...");
        log.debug("Step 3: Confirming with gateway...");

        log.info("Payment of ${} processed for order {}", String.format("%.2f", amount), orderId);
    }
}
