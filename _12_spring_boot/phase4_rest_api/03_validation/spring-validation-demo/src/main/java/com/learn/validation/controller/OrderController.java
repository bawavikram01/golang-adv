package com.learn.validation.controller;

import com.learn.validation.dto.CreateOrderRequest;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

/**
 * DEMO 2: Nested validation + Custom validator + Collection validation
 */
@RestController
@RequestMapping("/api/orders")
public class OrderController {

    /**
     * POST /api/orders
     * Demonstrates:
     * - @Valid cascading into nested AddressRequest
     * - @NoProfanity custom validator on product name
     * - @Size on List<String> tags
     * - @NotBlank on each element in the list
     */
    @PostMapping
    public ResponseEntity<Map<String, Object>> createOrder(
            @Valid @RequestBody CreateOrderRequest request) {
        
        double total = request.price() * request.quantity();
        
        return ResponseEntity.status(201).body(Map.of(
            "message", "Order placed successfully",
            "product", request.product(),
            "quantity", request.quantity(),
            "total", total,
            "shippingTo", request.shippingAddress().city() + ", " + request.shippingAddress().state()
        ));
    }
}
