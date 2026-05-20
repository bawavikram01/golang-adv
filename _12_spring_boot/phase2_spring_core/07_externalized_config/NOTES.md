# Phase 2.7 — Externalized Configuration

---

## Why Externalize Configuration?

Hardcoding values (database URLs, API keys, feature flags) in Java code is bad:
- Can't change without recompiling
- Different environments need different values (dev/staging/prod)
- Secrets in source code = security risk

Spring Boot's solution: **externalized configuration** via property files, environment variables, and command-line args.

---

## application.properties / application.yml

Spring Boot auto-loads `src/main/resources/application.properties`:

```properties
# application.properties
app.name=My Application
app.version=2.1.0
server.port=8080
spring.datasource.url=jdbc:postgresql://localhost:5432/mydb
```

Or YAML format (`application.yml`):
```yaml
app:
  name: My Application
  version: 2.1.0
server:
  port: 8080
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/mydb
```

Both are equivalent — use whichever you prefer. YAML is more readable for nested properties.

---

## @Value — Inject Individual Properties

```java
@Component
public class MyService {

    @Value("${app.name}")
    private String appName;

    @Value("${server.port}")
    private int port;  // Auto-converts String → int

    @Value("${app.debug:false}")  // Default value after ':'
    private boolean debug;

    @Value("${MISSING_KEY:fallback}")  // If key doesn't exist, use fallback
    private String safe;
}
```

### @Value Features:

| Syntax | Meaning |
|--------|---------|
| `${property.key}` | Inject property value |
| `${key:default}` | Use default if key missing |
| `#{2 + 3}` | SpEL expression (Spring Expression Language) |
| `${key1}/${key2}` | Combine multiple properties |
| `@Value("${items}")` on `List<String>` | Comma-separated → List |

### SpEL Expressions:
```java
@Value("#{${app.timeout} * 1000}")    // Math: multiply by 1000
private long timeoutMs;

@Value("#{systemProperties['user.home']}")  // System property
private String homeDir;

@Value("#{T(java.lang.Math).random()}")     // Static method call
private double random;
```

---

## @ConfigurationProperties — Type-Safe Config (Recommended)

Instead of scattered `@Value` annotations, bind an entire prefix to a POJO:

```java
@ConfigurationProperties(prefix = "app.mail")
public class MailProperties {
    private String host;         // app.mail.host
    private int port;            // app.mail.port
    private String username;     // app.mail.username
    private String password;     // app.mail.password
    private boolean ssl;         // app.mail.ssl

    // Getters and setters required!
}
```

```properties
# application.properties
app.mail.host=smtp.gmail.com
app.mail.port=587
app.mail.username=user@gmail.com
app.mail.password=secret
app.mail.ssl=true
```

### Enabling @ConfigurationProperties:
```java
@SpringBootApplication
@EnableConfigurationProperties(MailProperties.class)  // Option 1
public class MyApp { }

// OR

@ConfigurationProperties(prefix = "app.mail")
@Component  // Option 2: just make it a component
public class MailProperties { ... }
```

### Why prefer @ConfigurationProperties over @Value?
| Feature | @Value | @ConfigurationProperties |
|---------|--------|--------------------------|
| Type safety | Weak (runtime error if wrong type) | Strong (compile-time) |
| Validation | No | Yes (`@Validated` + Jakarta annotations) |
| Nested objects | No | Yes (nested POJOs, lists, maps) |
| IDE support | Minimal | Full autocomplete with metadata |
| Relaxed binding | No | Yes (`app-name` = `appName` = `APP_NAME`) |
| Testability | Harder | Easy (just new the POJO) |

---

## Property Sources Priority (highest wins)

Spring Boot loads properties in this order (higher = overrides lower):

```
1.  Command-line args:          --server.port=9090
2.  SPRING_APPLICATION_JSON:    env variable with JSON
3.  OS Environment variables:   SERVER_PORT=9090
4.  application-{profile}.properties
5.  application.properties
6.  @PropertySource annotations
7.  Default properties (SpringApplication.setDefaultProperties)
```

This means: **environment variables override property files**, and **command-line args override everything**.

### Examples:
```bash
# Override via command line
java -jar app.jar --server.port=9090 --app.name="Production"

# Override via environment variable
export APP_NAME="From Environment"
export SERVER_PORT=9090
java -jar app.jar
```

---

## Relaxed Binding

Spring Boot matches properties flexibly:

| In properties file | Matches field |
|-------------------|---------------|
| `app.database-url` | `databaseUrl` |
| `app.database_url` | `databaseUrl` |
| `app.databaseUrl` | `databaseUrl` |
| `APP_DATABASE_URL` | `databaseUrl` |

All four bind to the same Java field. This is especially useful for environment variables (which are UPPER_SNAKE_CASE).

---

## @PropertySource — Custom Property Files

```java
@Configuration
@PropertySource("classpath:custom.properties")
@PropertySource("classpath:secrets.properties")
public class CustomConfig { }
```

Or multiple:
```java
@PropertySources({
    @PropertySource("classpath:defaults.properties"),
    @PropertySource(value = "classpath:overrides.properties", ignoreResourceNotFound = true)
})
```

**Note:** `@PropertySource` does NOT support YAML — only `.properties` files.

---

## Profiles — Environment-Specific Config

```
src/main/resources/
├── application.properties         (common/default)
├── application-dev.properties     (dev overrides)
├── application-prod.properties    (prod overrides)
└── application-test.properties    (test overrides)
```

Activate a profile:
```properties
# In application.properties
spring.profiles.active=dev
```

Or via command line:
```bash
java -jar app.jar --spring.profiles.active=prod
```

Or via environment:
```bash
export SPRING_PROFILES_ACTIVE=prod
```

---

## Validation with @ConfigurationProperties

```java
@ConfigurationProperties(prefix = "app.mail")
@Validated
public class MailProperties {

    @NotBlank
    private String host;

    @Min(1) @Max(65535)
    private int port;

    @Email
    private String username;

    // App won't start if validation fails!
}
```

Requires `spring-boot-starter-validation` dependency.

---

## Key Takeaways

1. **application.properties** = primary config file (auto-loaded)
2. **@Value("${key}")** = quick injection for simple cases
3. **@ConfigurationProperties** = type-safe, validated, nested config (preferred for groups)
4. **Priority**: CLI args > env vars > profile-specific > application.properties
5. **Relaxed binding**: `my-prop` = `myProp` = `MY_PROP` (all match)
6. **Profiles**: `application-{profile}.properties` for env-specific overrides
7. **Never hardcode** passwords/URLs — externalize everything
