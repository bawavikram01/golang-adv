package com.learn.jackson.controller;

import com.learn.jackson.dto.*;
import com.learn.jackson.model.User;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.net.URI;
import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.atomic.AtomicLong;

/**
 * DEMO 1: DTO Pattern in action
 * 
 * Shows:
 * - Input DTO (CreateUserRequest) — what client sends
 * - Output DTO (UserResponse) — what client receives
 * - Entity (User) — internal, never exposed directly
 * - Password accepted but NEVER returned
 */
@RestController
@RequestMapping("/api/users")
public class UserController {

    private final Map<Long, User> users = new LinkedHashMap<>();
    private final AtomicLong idGen = new AtomicLong(1);

    public UserController() {
        // Pre-populate
        User u1 = new User(idGen.getAndIncrement(), "Alice Johnson", "alice@example.com", "secret123", "ADMIN");
        User u2 = new User(idGen.getAndIncrement(), "Bob Smith", "bob@example.com", "pass456", "USER");
        u2.setPhone("+1-555-0123");
        users.put(u1.getId(), u1);
        users.put(u2.getId(), u2);
    }

    // ─── GET ALL (returns DTOs, not entities) ─────────────────────
    @GetMapping
    public List<UserResponse> getAllUsers() {
        return users.values().stream()
                .map(this::toResponse)
                .toList();
    }

    // ─── GET BY ID ────────────────────────────────────────────────
    @GetMapping("/{id}")
    public ResponseEntity<UserResponse> getUser(@PathVariable Long id) {
        User user = users.get(id);
        if (user == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(toResponse(user));
    }

    // ─── CREATE (accepts DTO, returns DTO) ────────────────────────
    @PostMapping
    public ResponseEntity<UserResponse> createUser(@RequestBody CreateUserRequest request) {
        // DTO → Entity (manual mapping)
        User user = new User(
            idGen.getAndIncrement(),
            request.name(),
            request.email(),
            request.password(),  // Stored internally, never returned
            request.role()
        );
        users.put(user.getId(), user);

        // Entity → Response DTO
        URI location = URI.create("/api/users/" + user.getId());
        return ResponseEntity.created(location).body(toResponse(user));
    }

    // ─── UPDATE (accepts UpdateDTO) ───────────────────────────────
    @PutMapping("/{id}")
    public ResponseEntity<UserResponse> updateUser(
            @PathVariable Long id,
            @RequestBody UpdateUserRequest request) {
        User user = users.get(id);
        if (user == null) {
            return ResponseEntity.notFound().build();
        }
        // Only update fields present in the DTO
        if (request.name() != null) user.setName(request.name());
        if (request.email() != null) user.setEmail(request.email());
        if (request.phone() != null) user.setPhone(request.phone());

        return ResponseEntity.ok(toResponse(user));
    }

    // ─── Entity → Response DTO mapping ────────────────────────────
    private UserResponse toResponse(User user) {
        return new UserResponse(
            user.getId(),
            user.getName(),
            user.getEmail(),
            user.getUserRole(),
            user.getPhone(),       // Will be null for some users (excluded by @JsonInclude)
            user.getCreatedAt()
        );
    }
}
