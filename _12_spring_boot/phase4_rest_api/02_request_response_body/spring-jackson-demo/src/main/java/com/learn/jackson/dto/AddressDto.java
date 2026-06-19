package com.learn.jackson.dto;

/**
 * Nested DTO — used inside OrderRequest.
 */
public record AddressDto(
    String street,
    String city,
    String state,
    String zip
) {}
