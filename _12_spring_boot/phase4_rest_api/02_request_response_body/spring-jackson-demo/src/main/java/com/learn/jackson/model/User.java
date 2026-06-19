package com.learn.jackson.model;

import com.fasterxml.jackson.annotation.*;

import java.time.LocalDateTime;

/**
 * Internal entity — contains ALL data including sensitive fields.
 * This is what you'd store in a database.
 * 
 * Jackson annotations demonstrate field-level control.
 */
@JsonIgnoreProperties(ignoreUnknown = true) // Ignore extra JSON fields during deserialization
@JsonPropertyOrder({"id", "name", "email", "role", "createdAt"}) // Output order
public class User {

    private Long id;
    private String name;
    private String email;

    @JsonIgnore  // NEVER expose in JSON (neither input nor output)
    private String password;

    @JsonProperty("role")  // Rename: Java "userRole" → JSON "role"
    private String userRole;

    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")  // Custom date format
    private LocalDateTime createdAt;

    @JsonInclude(JsonInclude.Include.NON_NULL)  // Skip if null
    private String phone;

    public User() {}

    public User(Long id, String name, String email, String password, String userRole) {
        this.id = id;
        this.name = name;
        this.email = email;
        this.password = password;
        this.userRole = userRole;
        this.createdAt = LocalDateTime.now();
    }

    // Getters and Setters
    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }

    public String getEmail() { return email; }
    public void setEmail(String email) { this.email = email; }

    public String getPassword() { return password; }
    public void setPassword(String password) { this.password = password; }

    public String getUserRole() { return userRole; }
    public void setUserRole(String userRole) { this.userRole = userRole; }

    public LocalDateTime getCreatedAt() { return createdAt; }
    public void setCreatedAt(LocalDateTime createdAt) { this.createdAt = createdAt; }

    public String getPhone() { return phone; }
    public void setPhone(String phone) { this.phone = phone; }
}
