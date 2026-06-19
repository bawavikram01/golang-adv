# Phase 4.3 — Bean Validation (Jakarta Validation)

---

## Why Validate?

Never trust client input. Without validation:
- Empty names, invalid emails get into your database
- Negative prices, zero quantities break business logic
- SQL injection, XSS attacks sneak through
- API becomes unreliable — garbage in, garbage out

**Rule: Validate at the boundary (controller layer) before data enters your system.**

---

## Jakarta Bean Validation (JSR 380)

Spring Boot uses **Jakarta Bean Validation** (formerly javax.validation).

```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-validation</artifactId>
</dependency>
```

This brings in **Hibernate Validator** (the reference implementation).

---

## How It Works

```
Client sends JSON → @RequestBody → DTO with annotations → @Valid triggers validation
                                                              ↓
                                          Valid? → proceed to controller method
                                          Invalid? → MethodArgumentNotValidException (400)
```

---

## Built-in Constraint Annotations

### String Constraints:

| Annotation | Purpose | Example |
|-----------|---------|---------|
| `@NotNull` | Cannot be null | `@NotNull String name` |
| `@NotEmpty` | Not null AND not empty string | `@NotEmpty String name` |
| `@NotBlank` | Not null, not empty, not just whitespace | `@NotBlank String name` |
| `@Size(min=, max=)` | String length bounds | `@Size(min=2, max=50)` |
| `@Email` | Valid email format | `@Email String email` |
| `@Pattern(regexp=)` | Matches regex | `@Pattern(regexp="^[A-Z].*")` |

### Number Constraints:

| Annotation | Purpose | Example |
|-----------|---------|---------|
| `@Min(value)` | Minimum value | `@Min(1) int quantity` |
| `@Max(value)` | Maximum value | `@Max(100) int quantity` |
| `@Positive` | Must be > 0 | `@Positive double price` |
| `@PositiveOrZero` | Must be >= 0 | `@PositiveOrZero int stock` |
| `@DecimalMin` | Decimal minimum | `@DecimalMin("0.01")` |
| `@Digits(integer=, fraction=)` | Digit constraints | `@Digits(integer=5, fraction=2)` |

### Date/Time Constraints:

| Annotation | Purpose |
|-----------|---------|
| `@Past` | Must be in the past |
| `@Future` | Must be in the future |
| `@PastOrPresent` | Past or now |
| `@FutureOrPresent` | Future or now |

### Other:

| Annotation | Purpose |
|-----------|---------|
| `@AssertTrue` | Must be true |
| `@AssertFalse` | Must be false |

---

## Using Validation in Controller

```java
@PostMapping
public ResponseEntity<User> createUser(@Valid @RequestBody CreateUserRequest request) {
    // If validation fails, this method is NEVER called
    // Spring throws MethodArgumentNotValidException → 400 Bad Request
}
```

The magic keyword is **`@Valid`** — without it, annotations on the DTO are ignored.

---

## Custom Error Messages

```java
@NotBlank(message = "Name is required")
@Size(min = 2, max = 50, message = "Name must be 2-50 characters")
private String name;

@Email(message = "Please provide a valid email address")
private String email;

@Min(value = 1, message = "Quantity must be at least 1")
private int quantity;
```

---

## Handling Validation Errors (Clean Response)

Default Spring response for validation errors is ugly. Override it:

```java
@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ResponseEntity<Map<String, Object>> handleValidation(
            MethodArgumentNotValidException ex) {
        
        Map<String, String> errors = new LinkedHashMap<>();
        ex.getBindingResult().getFieldErrors().forEach(error ->
            errors.put(error.getField(), error.getDefaultMessage())
        );

        Map<String, Object> body = Map.of(
            "status", 400,
            "error", "Validation Failed",
            "errors", errors
        );
        return ResponseEntity.badRequest().body(body);
    }
}
```

---

## Validating @PathVariable and @RequestParam

```java
@RestController
@Validated  // Required for method-level validation
public class UserController {

    @GetMapping("/users/{id}")
    public User getUser(@PathVariable @Min(1) Long id) { ... }

    @GetMapping("/search")
    public List<User> search(@RequestParam @Size(min=2) String q) { ... }
}
```

Note: `@Validated` on the **class** is required for path/param validation.

---

## Nested Object Validation

```java
public record OrderRequest(
    @NotBlank String product,
    @Min(1) int quantity,
    @Valid AddressDto address  // @Valid cascades into nested object
) {}

public record AddressDto(
    @NotBlank String street,
    @NotBlank String city,
    @Pattern(regexp = "\\d{5}") String zip
) {}
```

---

## Custom Validator (Advanced)

Create your own annotation + validator:

```java
@Target({FIELD, PARAMETER})
@Retention(RUNTIME)
@Constraint(validatedBy = NoSwearWordsValidator.class)
public @interface NoSwearWords {
    String message() default "Contains inappropriate language";
    Class<?>[] groups() default {};
    Class<? extends Payload>[] payload() default {};
}
```

```java
public class NoSwearWordsValidator implements ConstraintValidator<NoSwearWords, String> {
    private static final Set<String> BANNED = Set.of("spam", "fake", "scam");

    @Override
    public boolean isValid(String value, ConstraintValidatorContext ctx) {
        if (value == null) return true;  // Let @NotNull handle nulls
        return BANNED.stream().noneMatch(word -> 
            value.toLowerCase().contains(word));
    }
}
```

---

## Key Takeaways

1. Add `spring-boot-starter-validation` dependency
2. Put constraint annotations on DTO fields (`@NotBlank`, `@Email`, `@Min`, etc.)
3. Add `@Valid` before `@RequestBody` to trigger validation
4. Use `@Validated` on the class for path/param validation
5. Use `@Valid` on nested objects to cascade validation
6. Handle `MethodArgumentNotValidException` for clean error responses
7. Create custom validators for business-specific rules
