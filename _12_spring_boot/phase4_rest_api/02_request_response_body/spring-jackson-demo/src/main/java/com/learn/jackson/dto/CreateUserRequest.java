package com.learn.jackson.dto;

/**
 * INPUT DTO — only the fields a client is ALLOWED to send when creating a user.
 * 
 * Using Java Record (Java 17+):
 * - Immutable
 * - Auto-generates constructor, getters, equals, hashCode, toString
 * - Jackson works with records out of the box
 */
public record CreateUserRequest(
    String name,
    String email,
    String password,  // We accept it here but NEVER return it
    String role
) {}
