package com.learn.validation.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;

/**
 * Nested DTO — validated when parent has @Valid.
 */
public record AddressRequest(

    @NotBlank(message = "Street is required")
    String street,

    @NotBlank(message = "City is required")
    String city,

    @NotBlank(message = "State is required")
    @Pattern(regexp = "^[A-Z]{2}$", message = "State must be a 2-letter code (e.g., CA, NY)")
    String state,

    @NotBlank(message = "ZIP code is required")
    @Pattern(regexp = "^\\d{5}(-\\d{4})?$", message = "ZIP must be 5 digits (e.g., 94102 or 94102-1234)")
    String zip

) {}
