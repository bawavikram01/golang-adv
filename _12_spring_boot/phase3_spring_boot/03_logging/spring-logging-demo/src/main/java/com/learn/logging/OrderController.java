package com.learn.logging;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.annotation.*;

import java.util.LinkedHashMap;
import java.util.Map;

/**
 * REST controller with proper request/response logging.
 */
@RestController
@RequestMapping("/api/orders")
public class OrderController {

    private static final Logger log = LoggerFactory.getLogger(OrderController.class);

    private final OrderService orderService;

    public OrderController(OrderService orderService) {
        this.orderService = orderService;
    }

    // POST http://localhost:8082/api/orders?orderId=ORD-001&customer=Alice&amount=79.99
    @PostMapping
    public Map<String, Object> placeOrder(
            @RequestParam String orderId,
            @RequestParam String customer,
            @RequestParam double amount) {

        log.info("POST /api/orders — orderId={}, customer={}, amount={}", orderId, customer, amount);

        String result = orderService.placeOrder(orderId, customer, amount);
        orderService.processPayment(orderId, amount);

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("status", "success");
        response.put("message", result);
        response.put("note", "Check console/log file for logging output");

        log.info("Response: status=success for order {}", orderId);
        return response;
    }

    // DELETE http://localhost:8082/api/orders/ORD-001
    @DeleteMapping("/{orderId}")
    public Map<String, Object> cancelOrder(@PathVariable String orderId) {
        log.info("DELETE /api/orders/{}", orderId);

        Map<String, Object> response = new LinkedHashMap<>();
        try {
            String result = orderService.cancelOrder(orderId);
            response.put("status", "success");
            response.put("message", result);
        } catch (Exception e) {
            log.warn("Cancel failed — returning error response for {}", orderId);
            response.put("status", "error");
            response.put("message", e.getMessage());
        }

        return response;
    }

    // GET http://localhost:8082/api/orders/log-levels
    @GetMapping("/log-levels")
    public Map<String, Object> showLogLevels() {
        log.trace("TRACE message — only visible at TRACE level");
        log.debug("DEBUG message — visible at DEBUG level and above");
        log.info("INFO message — visible at INFO level and above");
        log.warn("WARN message — visible at WARN level and above");
        log.error("ERROR message — always visible");

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("configured_level", "DEBUG (for com.learn.logging)");
        response.put("trace_visible", false);
        response.put("debug_visible", true);
        response.put("info_visible", true);
        response.put("warn_visible", true);
        response.put("error_visible", true);
        response.put("note", "TRACE is below our configured DEBUG level, so it's hidden");
        response.put("change_level", "Add --logging.level.com.learn.logging=TRACE to see TRACE");
        return response;
    }
}
