package com.learn.jackson.dto;

import com.fasterxml.jackson.annotation.JsonFormat;

import java.time.LocalDateTime;
import java.util.List;

/**
 * DEMO: Nested objects and collections in JSON.
 */
public record OrderRequest(
    String product,
    int quantity,
    double price,
    AddressDto shippingAddress,    // Nested object
    List<String> tags              // Collection
) {}
