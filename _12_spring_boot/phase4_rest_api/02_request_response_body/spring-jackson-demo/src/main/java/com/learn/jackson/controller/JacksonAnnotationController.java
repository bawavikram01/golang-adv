package com.learn.jackson.controller;

import com.learn.jackson.model.User;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;

/**
 * DEMO 3: Jackson annotations on the entity itself.
 * 
 * Shows what happens when you serialize a User entity directly:
 * - @JsonIgnore hides password
 * - @JsonProperty("role") renames userRole
 * - @JsonFormat formats the date
 * - @JsonInclude(NON_NULL) skips null phone
 * - @JsonPropertyOrder controls output order
 * - @JsonIgnoreProperties(ignoreUnknown=true) allows extra fields in input
 */
@RestController
@RequestMapping("/api/jackson-demo")
public class JacksonAnnotationController {

    // Shows @JsonIgnore — password is hidden in output
    @GetMapping("/user-with-annotations")
    public User getUserWithAnnotations() {
        User user = new User(1L, "Demo User", "demo@example.com", "SuperSecret!", "ADMIN");
        // phone is null — @JsonInclude(NON_NULL) will skip it
        return user;
    }

    // Shows @JsonIgnoreProperties(ignoreUnknown=true) — extra fields are ignored
    @PostMapping("/ignore-unknown")
    public User postWithExtraFields(@RequestBody User user) {
        // Even if client sends {"name":"A","email":"a@b.com","unknownField":"xyz"}
        // Jackson won't throw an error due to @JsonIgnoreProperties(ignoreUnknown=true)
        user.setId(99L);
        user.setCreatedAt(LocalDateTime.now());
        return user;
    }
}
