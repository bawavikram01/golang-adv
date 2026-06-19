package com.learn.jackson.dto;

import com.fasterxml.jackson.annotation.JsonFormat;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.time.LocalDateTime;

/**
 * OUTPUT DTO — only the fields the client SHOULD see.
 * 
 * Notice: NO password field here! The client never sees it.
 */
public record UserResponse(
    Long id,

    @JsonProperty("full_name")  // Rename for output: "full_name" in JSON
    String name,

    String email,
    String role,
    String phone,

    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    LocalDateTime createdAt
) {}
