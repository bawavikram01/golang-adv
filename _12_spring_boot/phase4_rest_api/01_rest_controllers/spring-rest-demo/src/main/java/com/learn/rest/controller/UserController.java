package com.learn.rest.controller;

import com.learn.rest.model.User;
import com.learn.rest.service.UserService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.net.URI;
import java.util.List;

/**
 * DEMO 1: Full CRUD REST Controller
 * 
 * Demonstrates:
 *  - @RestController (auto JSON serialization)
 *  - @RequestMapping (base path)
 *  - @GetMapping, @PostMapping, @PutMapping, @DeleteMapping
 *  - @PathVariable (URL segments)
 *  - @RequestParam (query parameters)
 *  - @RequestBody (JSON body parsing)
 *  - ResponseEntity (status codes + headers)
 */
@RestController
@RequestMapping("/api/users")
public class UserController {

    private final UserService userService;

    // Constructor injection (Spring wires automatically)
    public UserController(UserService userService) {
        this.userService = userService;
    }

    // ─── GET ALL USERS ─────────────────────────────────────────────
    // GET /api/users
    // GET /api/users?name=alice  (filter by name)
    @GetMapping
    public List<User> getAllUsers(@RequestParam(required = false) String name) {
        if (name != null) {
            return userService.findByName(name);
        }
        return userService.findAll();
    }

    // ─── GET USER BY ID ────────────────────────────────────────────
    // GET /api/users/1
    @GetMapping("/{id}")
    public ResponseEntity<User> getUserById(@PathVariable Long id) {
        return userService.findById(id)
                .map(ResponseEntity::ok)                    // 200 OK
                .orElse(ResponseEntity.notFound().build()); // 404 Not Found
    }

    // ─── CREATE USER ───────────────────────────────────────────────
    // POST /api/users  (body: {"name":"Dave","email":"dave@ex.com"})
    @PostMapping
    public ResponseEntity<User> createUser(@RequestBody User user) {
        User saved = userService.save(user);
        URI location = URI.create("/api/users/" + saved.getId());
        return ResponseEntity.created(location).body(saved); // 201 Created
    }

    // ─── UPDATE USER (FULL REPLACE) ───────────────────────────────
    // PUT /api/users/1  (body: {"name":"Alice Updated","email":"new@ex.com"})
    @PutMapping("/{id}")
    public ResponseEntity<User> updateUser(@PathVariable Long id, @RequestBody User user) {
        return userService.update(id, user)
                .map(ResponseEntity::ok)                    // 200 OK
                .orElse(ResponseEntity.notFound().build()); // 404 Not Found
    }

    // ─── DELETE USER ───────────────────────────────────────────────
    // DELETE /api/users/1
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable Long id) {
        if (userService.delete(id)) {
            return ResponseEntity.noContent().build();      // 204 No Content
        }
        return ResponseEntity.notFound().build();           // 404 Not Found
    }
}
