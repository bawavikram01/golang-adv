package com.learn.validation.controller;

import com.learn.validation.dto.CreateUserRequest;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.Size;
import java.util.Map;

/**
 * DEMO 1: Basic validation with @Valid + @RequestBody
 * Also: @Validated for path/query param validation
 */
@RestController
@RequestMapping("/api/users")
@Validated  // Required for @PathVariable / @RequestParam validation
public class UserController {

    /**
     * POST /api/users
     * @Valid triggers validation on CreateUserRequest fields.
     * If validation fails → MethodArgumentNotValidException → 400
     */
    @PostMapping
    public ResponseEntity<Map<String, Object>> createUser(
            @Valid @RequestBody CreateUserRequest request) {
        // This code ONLY runs if ALL validations pass
        return ResponseEntity.status(201).body(Map.of(
            "message", "User created successfully",
            "user", Map.of(
                "name", request.name(),
                "email", request.email(),
                "age", request.age()
            )
        ));
    }

    /**
     * GET /api/users/{id}
     * @Min on @PathVariable — validates the path variable itself.
     * Requires @Validated on the class.
     */
    @GetMapping("/{id}")
    public Map<String, Object> getUser(@PathVariable @Min(value = 1, message = "ID must be positive") Long id) {
        return Map.of("id", id, "name", "User " + id, "email", "user" + id + "@example.com");
    }

    /**
     * GET /api/users/search?q=xyz
     * @Size on @RequestParam — validates query parameter length.
     */
    @GetMapping("/search")
    public Map<String, Object> search(
            @RequestParam @Size(min = 2, message = "Search query must be at least 2 characters") String q) {
        return Map.of("query", q, "results", "Found users matching '" + q + "'");
    }
}
