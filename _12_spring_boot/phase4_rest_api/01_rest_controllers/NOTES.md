# Phase 4.1 — REST Controllers & HTTP Methods

---

## What is REST?

REST (Representational State Transfer) is an architectural style for APIs:

| Principle | Meaning |
|-----------|---------|
| **Stateless** | Each request contains all info needed (no server-side session) |
| **Resource-based** | URLs represent resources (nouns, not verbs) |
| **HTTP methods** | GET/POST/PUT/DELETE map to CRUD operations |
| **Representations** | Resources returned as JSON (or XML) |
| **Uniform interface** | Consistent URL patterns across the API |

---

## HTTP Methods → CRUD

| HTTP Method | CRUD Operation | Example URL | Purpose |
|-------------|---------------|-------------|---------|
| `GET` | Read | `GET /api/users` | Get all users |
| `GET` | Read | `GET /api/users/42` | Get user by ID |
| `POST` | Create | `POST /api/users` | Create a new user |
| `PUT` | Update (full) | `PUT /api/users/42` | Replace entire user |
| `PATCH` | Update (partial) | `PATCH /api/users/42` | Update specific fields |
| `DELETE` | Delete | `DELETE /api/users/42` | Delete user |

---

## @RestController

```java
@RestController  // = @Controller + @ResponseBody
@RequestMapping("/api/users")  // Base path for all endpoints in this class
public class UserController {

    @GetMapping           // GET /api/users
    @GetMapping("/{id}")  // GET /api/users/42
    @PostMapping          // POST /api/users
    @PutMapping("/{id}")  // PUT /api/users/42
    @PatchMapping("/{id}")// PATCH /api/users/42
    @DeleteMapping("/{id}") // DELETE /api/users/42
}
```

### @RestController vs @Controller:
- `@Controller` = returns view names (HTML templates)
- `@RestController` = returns data directly (auto-serialized to JSON)
- `@RestController` = `@Controller` + `@ResponseBody` on every method

---

## URL Patterns & Parameter Binding

### @PathVariable — extract from URL path:
```java
// GET /api/users/42
@GetMapping("/{id}")
public User getUser(@PathVariable Long id) { ... }

// GET /api/users/42/orders/7
@GetMapping("/{userId}/orders/{orderId}")
public Order getOrder(@PathVariable Long userId, @PathVariable Long orderId) { ... }
```

### @RequestParam — query string parameters:
```java
// GET /api/users?page=0&size=10&sort=name
@GetMapping
public List<User> getUsers(
    @RequestParam(defaultValue = "0") int page,
    @RequestParam(defaultValue = "10") int size,
    @RequestParam(required = false) String sort
) { ... }
```

### @RequestBody — parse JSON body:
```java
// POST /api/users  (body: {"name":"Alice","email":"alice@mail.com"})
@PostMapping
public User createUser(@RequestBody User user) { ... }
```

### @RequestHeader — read HTTP headers:
```java
@GetMapping
public String check(@RequestHeader("Authorization") String authHeader) { ... }
```

---

## HTTP Status Codes

| Code | Meaning | When to Use |
|------|---------|-------------|
| `200 OK` | Success | GET, PUT, PATCH |
| `201 Created` | Resource created | POST |
| `204 No Content` | Success, no body | DELETE |
| `400 Bad Request` | Invalid input | Validation fails |
| `404 Not Found` | Resource doesn't exist | GET/PUT/DELETE with wrong ID |
| `409 Conflict` | Duplicate/conflict | Duplicate email, etc |
| `500 Internal Server Error` | Server bug | Unhandled exceptions |

---

## ResponseEntity — Full Control

```java
@PostMapping
public ResponseEntity<User> createUser(@RequestBody User user) {
    User saved = userService.save(user);
    URI location = URI.create("/api/users/" + saved.getId());

    return ResponseEntity
        .created(location)      // 201 + Location header
        .body(saved);           // Response body
}

@GetMapping("/{id}")
public ResponseEntity<User> getUser(@PathVariable Long id) {
    return userService.findById(id)
        .map(ResponseEntity::ok)           // 200 + body
        .orElse(ResponseEntity.notFound().build());  // 404
}

@DeleteMapping("/{id}")
public ResponseEntity<Void> delete(@PathVariable Long id) {
    userService.delete(id);
    return ResponseEntity.noContent().build();  // 204
}
```

---

## REST API Best Practices

1. **Use nouns, not verbs**: `/api/users` not `/api/getUsers`
2. **Plural names**: `/api/users` not `/api/user`
3. **Nested resources**: `/api/users/42/orders` (orders of user 42)
4. **Use proper HTTP methods**: Don't use GET for creating resources
5. **Return proper status codes**: 201 for create, 204 for delete
6. **Consistent error responses**: Use a standard error format
7. **Version your API**: `/api/v1/users`

---

## Key Takeaways

1. `@RestController` = handles HTTP requests, returns JSON automatically
2. `@RequestMapping("/api/users")` = base path for the controller
3. `@GetMapping`, `@PostMapping`, `@PutMapping`, `@DeleteMapping` = HTTP method handlers
4. `@PathVariable` = URL path segment, `@RequestParam` = query string, `@RequestBody` = JSON body
5. `ResponseEntity` = full control over status code, headers, and body
6. REST = resources (nouns) + HTTP methods (verbs) + proper status codes
