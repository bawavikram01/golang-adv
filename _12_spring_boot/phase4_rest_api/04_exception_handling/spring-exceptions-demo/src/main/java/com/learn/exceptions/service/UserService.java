package com.learn.exceptions.service;

import com.learn.exceptions.exception.*;
import org.springframework.stereotype.Service;

import java.util.*;

/**
 * Service layer — throws custom business exceptions.
 * The service doesn't know about HTTP (no ResponseEntity, no status codes).
 * It only knows business rules.
 */
@Service
public class UserService {

    private final Map<Long, Map<String, Object>> users = new LinkedHashMap<>();
    private long nextId = 1;

    public UserService() {
        createUser("Alice Johnson", "alice@example.com", 1000.0);
        createUser("Bob Smith", "bob@example.com", 500.0);
    }

    public List<Map<String, Object>> findAll() {
        return new ArrayList<>(users.values());
    }

    public Map<String, Object> findById(Long id) {
        Map<String, Object> user = users.get(id);
        if (user == null) {
            // Business exception — will be caught by GlobalExceptionHandler
            throw new ResourceNotFoundException("User", id);
        }
        return user;
    }

    public Map<String, Object> createUser(String name, String email, double balance) {
        // Check for duplicate email
        boolean emailExists = users.values().stream()
                .anyMatch(u -> email.equals(u.get("email")));
        if (emailExists) {
            throw new DuplicateResourceException("User", "email", email);
        }

        Map<String, Object> user = new LinkedHashMap<>();
        user.put("id", nextId);
        user.put("name", name);
        user.put("email", email);
        user.put("balance", balance);
        users.put(nextId, user);
        nextId++;
        return user;
    }

    public void transfer(Long fromId, Long toId, double amount) {
        Map<String, Object> fromUser = findById(fromId);  // throws ResourceNotFoundException
        Map<String, Object> toUser = findById(toId);      // throws ResourceNotFoundException

        double fromBalance = (double) fromUser.get("balance");

        // Business rule: cannot transfer more than you have
        if (amount > fromBalance) {
            throw new BusinessRuleException(
                "Insufficient balance. Available: " + fromBalance + ", Requested: " + amount
            );
        }

        // Business rule: cannot transfer to yourself
        if (fromId.equals(toId)) {
            throw new BusinessRuleException("Cannot transfer to the same account");
        }

        fromUser.put("balance", fromBalance - amount);
        toUser.put("balance", (double) toUser.get("balance") + amount);
    }

    public void deleteById(Long id) {
        if (users.remove(id) == null) {
            throw new ResourceNotFoundException("User", id);
        }
    }

    public void simulateCrash() {
        // Simulates an unexpected error (e.g., DB connection failure)
        throw new RuntimeException("Database connection timeout after 30 seconds");
    }
}
