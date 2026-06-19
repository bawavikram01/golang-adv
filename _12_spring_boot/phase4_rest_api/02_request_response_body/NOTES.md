# Phase 4.2 — Request/Response Body & Jackson

---

## How Spring Converts Java ↔ JSON

```
Client sends JSON → Spring uses Jackson → Java Object (deserialization)
Java Object → Spring uses Jackson → JSON sent to client (serialization)
```

**Jackson** is included automatically with `spring-boot-starter-web`.

The key players:
- `ObjectMapper` — the core Jackson class that does all conversion
- `HttpMessageConverter` — Spring's abstraction over Jackson
- `@RequestBody` — tells Spring to deserialize incoming JSON into a Java object
- `@ResponseBody` (implicit in `@RestController`) — serialize return value to JSON

---

## The DTO Pattern (Data Transfer Object)

### Problem: Exposing entities directly is dangerous

```java
// BAD — exposes password, internal IDs, allows mass-assignment
@PostMapping
public User createUser(@RequestBody User user) { ... }
```

### Solution: Use DTOs to control what goes in and what comes out

```
Client → CreateUserRequest (DTO) → Service → User (Entity) → UserResponse (DTO) → Client
```

| Layer | Object | Purpose |
|-------|--------|---------|
| Input | `CreateUserRequest` | Only fields client can set |
| Internal | `User` (entity) | Full object with all fields |
| Output | `UserResponse` | Only fields client should see |

### Benefits of DTOs:
1. **Security** — don't expose passwords, internal IDs
2. **Decoupling** — API contract independent of DB schema
3. **Validation** — validate only input fields
4. **Versioning** — change entity without breaking API
5. **Clarity** — explicit about what data flows where

---

## Jackson Annotations (Most Used)

### On Fields / Getters:

| Annotation | Purpose | Example |
|-----------|---------|---------|
| `@JsonProperty("name")` | Rename field in JSON | `@JsonProperty("user_name") String name` |
| `@JsonIgnore` | Exclude from JSON entirely | `@JsonIgnore String password` |
| `@JsonFormat` | Format dates/times | `@JsonFormat(pattern = "yyyy-MM-dd")` |
| `@JsonInclude` | Skip null/empty values | `@JsonInclude(Include.NON_NULL)` |

### On Classes:

| Annotation | Purpose |
|-----------|---------|
| `@JsonIgnoreProperties(ignoreUnknown = true)` | Ignore extra JSON fields |
| `@JsonNaming(SnakeCaseStrategy.class)` | Use `snake_case` for all fields |
| `@JsonPropertyOrder({"id", "name"})` | Control field order in output |

### Serialization vs Deserialization:

| Annotation | When |
|-----------|------|
| `@JsonProperty(access = READ_ONLY)` | Only in output (serialization) |
| `@JsonProperty(access = WRITE_ONLY)` | Only in input (deserialization) |

---

## Customizing ObjectMapper Globally

```java
@Configuration
public class JacksonConfig {

    @Bean
    public ObjectMapper objectMapper() {
        ObjectMapper mapper = new ObjectMapper();
        mapper.setSerializationInclusion(JsonInclude.Include.NON_NULL);
        mapper.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
        mapper.registerModule(new JavaTimeModule());
        mapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
        return mapper;
    }
}
```

Or via `application.properties`:
```properties
spring.jackson.serialization.write-dates-as-timestamps=false
spring.jackson.default-property-inclusion=non-null
spring.jackson.deserialization.fail-on-unknown-properties=false
spring.jackson.date-format=yyyy-MM-dd HH:mm:ss
```

---

## Java Records as DTOs (Java 17+)

```java
// Immutable, concise, perfect for DTOs
public record CreateUserRequest(String name, String email) {}
public record UserResponse(Long id, String name, String email, LocalDateTime createdAt) {}
```

Records work with Jackson out of the box — no getters/setters needed!

---

## @RequestBody Deep Dive

```java
@PostMapping
public ResponseEntity<UserResponse> create(@RequestBody CreateUserRequest request) {
    // request.name(), request.email() — auto-deserialized from JSON
    User entity = new User(request.name(), request.email());
    User saved = service.save(entity);
    return ResponseEntity.created(...).body(toResponse(saved));
}
```

What happens:
1. Client sends: `{"name":"Alice","email":"alice@ex.com"}`
2. Spring's `MappingJackson2HttpMessageConverter` reads the body
3. Jackson's `ObjectMapper.readValue()` creates `CreateUserRequest`
4. Your method receives a populated object

---

## Nested Objects & Collections

Jackson handles nested JSON automatically:

```java
public record OrderRequest(
    String product,
    int quantity,
    AddressDto shippingAddress    // nested object
) {}

public record AddressDto(String street, String city, String zip) {}
```

```json
{
  "product": "Laptop",
  "quantity": 1,
  "shippingAddress": {
    "street": "123 Main St",
    "city": "NYC",
    "zip": "10001"
  }
}
```

---

## Key Takeaways

1. **Jackson** auto-converts Java ↔ JSON (included in starter-web)
2. **DTOs** separate your API contract from your internal model
3. **`@JsonIgnore`** hides fields, **`@JsonProperty`** renames them
4. **`@JsonFormat`** controls date formatting
5. **Java Records** are ideal for DTOs — immutable, concise
6. **ObjectMapper** can be customized globally via config or properties
7. **Nested objects** and collections are handled automatically
