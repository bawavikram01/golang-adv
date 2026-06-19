package com.learn.jackson.dto;

/**
 * INPUT DTO for partial updates (PUT/PATCH).
 */
public record UpdateUserRequest(
    String name,
    String email,
    String phone
) {}
