package com.learn.exceptions.controller;

import com.learn.exceptions.dto.CreateUserRequest;
import com.learn.exceptions.dto.TransferRequest;
import com.learn.exceptions.service.UserService;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

/**
 * REST Controller — delegates to service, never catches exceptions itself.
 * All exceptions bubble up to GlobalExceptionHandler.
 */
@RestController
@RequestMapping("/api/users")
public class UserController {

    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    // GET /api/users
    @GetMapping
    public List<Map<String, Object>> getAllUsers() {
        return userService.findAll();
    }

    // GET /api/users/1  (throws ResourceNotFoundException if not found)
    @GetMapping("/{id}")
    public Map<String, Object> getUser(@PathVariable Long id) {
        return userService.findById(id);
    }

    // POST /api/users  (throws DuplicateResourceException if email exists)
    @PostMapping
    public ResponseEntity<Map<String, Object>> createUser(@Valid @RequestBody CreateUserRequest request) {
        Map<String, Object> user = userService.createUser(
            request.name(), request.email(), request.balance()
        );
        return ResponseEntity.status(201).body(user);
    }

    // POST /api/users/transfer  (throws BusinessRuleException if insufficient funds)
    @PostMapping("/transfer")
    public Map<String, String> transfer(@RequestBody TransferRequest request) {
        userService.transfer(request.fromUserId(), request.toUserId(), request.amount());
        return Map.of("message", "Transfer successful");
    }

    // DELETE /api/users/1  (throws ResourceNotFoundException if not found)
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable Long id) {
        userService.deleteById(id);
        return ResponseEntity.noContent().build();
    }

    // GET /api/users/crash  (simulates unexpected error → 500)
    @GetMapping("/crash")
    public String crash() {
        userService.simulateCrash();
        return "This will never be reached";
    }
}
