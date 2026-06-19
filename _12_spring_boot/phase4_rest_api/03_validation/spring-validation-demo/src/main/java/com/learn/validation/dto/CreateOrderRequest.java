package com.learn.validation.dto;

import com.learn.validation.validator.NoProfanity;
import jakarta.validation.Valid;
import jakarta.validation.constraints.*;

import java.util.List;

/**
 * Order DTO — demonstrates:
 * - @Positive for numbers
 * - @Valid for nested object validation
 * - @Size on collections
 * - Custom validator (@NoProfanity)
 */
public record CreateOrderRequest(

    @NotBlank(message = "Product name is required")
    @Size(max = 100, message = "Product name cannot exceed 100 characters")
    @NoProfanity  // Custom validator
    String product,

    @Positive(message = "Quantity must be greater than 0")
    int quantity,

    @DecimalMin(value = "0.01", message = "Price must be at least 0.01")
    double price,

    @NotNull(message = "Shipping address is required")
    @Valid  // Cascade validation into nested object
    AddressRequest shippingAddress,

    @Size(max = 5, message = "Maximum 5 tags allowed")
    List<@NotBlank(message = "Tags cannot be blank") String> tags

) {}
