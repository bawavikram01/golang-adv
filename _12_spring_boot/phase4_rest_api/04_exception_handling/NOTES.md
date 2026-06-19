# Phase 4.4 — Exception Handling in REST APIs

---

## The Problem

Without proper exception handling, your API returns ugly stack traces or generic 500 errors:

```json
{
  "timestamp": "2026-06-01T12:00:00",
  "status": 500,
  "error": "Internal Server Error",
  "path": "/api/users/999"
}
```

Clients have no idea what went wrong. You need **structured, informative error responses**.

---

## Exception Handling Strategy (3 Layers)

```
Layer 1: Custom Business Exceptions (throw them in service layer)
    ↓
Layer 2: @ExceptionHandler methods (catch & convert to response)
    ↓
Layer 3: @ControllerAdvice / @RestControllerAdvice (global handlers)
```

---

## Layer 1: Custom Exceptions

```java
// Base class for all business exceptions
public abstract class BusinessException extends RuntimeException {
    private final String errorCode;

    public BusinessException(String message, String errorCode) {
        super(message);
        this.errorCode = errorCode;
    }
}

// Specific exceptions
public class ResourceNotFoundException extends BusinessException {
    public ResourceNotFoundException(String resource, Object id) {
        super(resource + " not found with id: " + id, "NOT_FOUND");
    }
}

public class DuplicateResourceException extends BusinessException {
    public DuplicateResourceException(String resource, String field, String value) {
        super(resource + " already exists with " + field + ": " + value, "DUPLICATE");
    }
}
```

---

## Layer 2: @ExceptionHandler (Local — per controller)

```java
@RestController
public class UserController {

    @ExceptionHandler(ResourceNotFoundException.class)
    public ResponseEntity<ErrorResponse> handleNotFound(ResourceNotFoundException ex) {
        // Only handles exceptions from THIS controller
    }
}
```

---

## Layer 3: @RestControllerAdvice (Global — all controllers)

```java
@RestControllerAdvice  // = @ControllerAdvice + @ResponseBody
public class GlobalExceptionHandler {

    @ExceptionHandler(ResourceNotFoundException.class)
    public ResponseEntity<ErrorResponse> handleNotFound(ResourceNotFoundException ex) {
        // Handles this exception from ANY controller
    }
}
```

### @ControllerAdvice vs @RestControllerAdvice:
| Annotation | Returns |
|-----------|---------|
| `@ControllerAdvice` | View names (HTML) |
| `@RestControllerAdvice` | JSON/XML directly |

---

## Standard Error Response Structure

```java
public record ErrorResponse(
    LocalDateTime timestamp,
    int status,
    String error,
    String message,
    String path,
    String errorCode,
    Map<String, String> details  // for validation errors
) {}
```

---

## ProblemDetail (RFC 7807) — Spring 6+ Standard

Spring 6 introduced `ProblemDetail` — a built-in class following RFC 7807:

```java
@ExceptionHandler(ResourceNotFoundException.class)
public ProblemDetail handleNotFound(ResourceNotFoundException ex) {
    ProblemDetail problem = ProblemDetail.forStatusAndDetail(
        HttpStatus.NOT_FOUND, ex.getMessage()
    );
    problem.setTitle("Resource Not Found");
    problem.setType(URI.create("https://api.example.com/errors/not-found"));
    problem.setProperty("errorCode", "NOT_FOUND");
    problem.setProperty("timestamp", Instant.now());
    return problem;
}
```

Response:
```json
{
  "type": "https://api.example.com/errors/not-found",
  "title": "Resource Not Found",
  "status": 404,
  "detail": "User not found with id: 999",
  "instance": "/api/users/999",
  "errorCode": "NOT_FOUND",
  "timestamp": "2026-06-01T12:00:00Z"
}
```

---

## Exception Handler Priority Order

1. `@ExceptionHandler` in the **same controller** (local)
2. `@ExceptionHandler` in `@ControllerAdvice` (global)
3. Spring's default handler (generic error page)

More specific exception classes are chosen over generic ones:
- `ResourceNotFoundException` handler wins over `RuntimeException` handler

---

## @ResponseStatus — Quick & Simple

```java
@ResponseStatus(HttpStatus.NOT_FOUND)  // Auto-returns 404 when thrown
public class ResourceNotFoundException extends RuntimeException {
    public ResourceNotFoundException(String message) {
        super(message);
    }
}
```

Simple but limited — no custom body. Use `@ExceptionHandler` for full control.

---

## Common Exceptions to Handle

| Exception | HTTP Status | When |
|-----------|------------|------|
| `ResourceNotFoundException` | 404 | Entity not found by ID |
| `DuplicateResourceException` | 409 | Unique constraint violated |
| `BusinessRuleException` | 422 | Business logic violation |
| `MethodArgumentNotValidException` | 400 | @Valid fails |
| `HttpMessageNotReadableException` | 400 | Malformed JSON |
| `MethodNotAllowedException` | 405 | Wrong HTTP method |
| `Exception` (fallback) | 500 | Unexpected errors |

---

## Key Takeaways

1. **Throw custom exceptions** in service layer (never return error codes)
2. **`@RestControllerAdvice`** catches exceptions globally
3. **`@ExceptionHandler`** maps exception types to HTTP responses
4. **`ProblemDetail`** (RFC 7807) is the modern standard format
5. **Custom error DTOs** give you full control over response structure
6. Handle specific exceptions first, then add a generic fallback
7. Never expose stack traces to clients in production
