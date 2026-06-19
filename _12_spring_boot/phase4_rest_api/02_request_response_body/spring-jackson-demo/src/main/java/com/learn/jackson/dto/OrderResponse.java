package com.learn.jackson.dto;

import com.fasterxml.jackson.annotation.JsonFormat;

import java.time.LocalDateTime;
import java.util.List;

/**
 * Output DTO for orders — shows how nested objects are serialized.
 */
public record OrderResponse(
    Long orderId,
    String product,
    int quantity,
    double price,
    double total,
    AddressDto shippingAddress,
    List<String> tags,

    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    LocalDateTime orderedAt
) {}
