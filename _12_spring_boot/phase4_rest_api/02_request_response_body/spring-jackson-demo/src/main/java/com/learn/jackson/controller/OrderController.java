package com.learn.jackson.controller;

import com.learn.jackson.dto.OrderRequest;
import com.learn.jackson.dto.OrderResponse;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.concurrent.atomic.AtomicLong;

/**
 * DEMO 2: Nested objects and collections in request/response.
 * 
 * Shows Jackson handling:
 * - Nested objects (AddressDto inside OrderRequest)
 * - Lists/arrays (tags)
 * - Computed fields in response (total = price * quantity)
 */
@RestController
@RequestMapping("/api/orders")
public class OrderController {

    private final AtomicLong orderIdGen = new AtomicLong(1000);

    @PostMapping
    public OrderResponse placeOrder(@RequestBody OrderRequest request) {
        // Jackson automatically deserialized the nested JSON into records
        return new OrderResponse(
            orderIdGen.getAndIncrement(),
            request.product(),
            request.quantity(),
            request.price(),
            request.price() * request.quantity(),  // computed field
            request.shippingAddress(),             // nested object passed through
            request.tags(),                        // list passed through
            LocalDateTime.now()
        );
    }
}
