# Phase 3.3 — Logging (SLF4J + Logback)

---

## Why Logging Matters

`System.out.println()` is NOT logging:
- Can't turn it off in production
- No timestamps, no severity levels
- Can't route different messages to different files
- Can't control verbosity per package

Proper logging gives you: levels, timestamps, structured output, file rotation, per-package control.

---

## Spring Boot's Logging Stack

```
Your Code → SLF4J (API/facade) → Logback (implementation)
```

| Layer | Role | Provided by |
|-------|------|-------------|
| **SLF4J** | Logging API (interface) | `spring-boot-starter` includes it |
| **Logback** | Actual logging engine | Spring Boot's default implementation |

You code against **SLF4J** (the interface). If you ever swap Logback for Log4j2, your code doesn't change.

---

## Basic Usage

```java
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

@Service
public class OrderService {

    // One logger per class (standard pattern)
    private static final Logger log = LoggerFactory.getLogger(OrderService.class);

    public void placeOrder(String orderId, double amount) {
        log.info("Placing order {} for ${}", orderId, amount);     // INFO
        log.debug("Order details: id={}, amount={}", orderId, amount);  // DEBUG
        log.warn("Low stock detected for order {}", orderId);      // WARN
        log.error("Payment failed for order {}", orderId);         // ERROR

        try {
            // ...
        } catch (Exception e) {
            log.error("Failed to process order {}: {}", orderId, e.getMessage(), e);
            // Passing 'e' as last arg prints full stack trace
        }
    }
}
```

### Key rules:
- Use `{}` placeholders (NOT string concatenation) — efficient, no toString() if level disabled
- One `Logger` per class (static final)
- Pass exception as LAST argument for stack trace

---

## Log Levels (Severity Order)

```
TRACE → DEBUG → INFO → WARN → ERROR
 (most verbose)              (least verbose)
```

| Level | When to use |
|-------|-------------|
| `TRACE` | Very detailed diagnostic (method entry/exit, loop iterations) |
| `DEBUG` | Useful during development (variable values, flow decisions) |
| `INFO` | Important business events (startup, order placed, user login) |
| `WARN` | Something unexpected but recoverable (retries, deprecation) |
| `ERROR` | Something broke (exceptions, failed operations) |

Setting level = INFO means: INFO, WARN, ERROR are shown. DEBUG and TRACE are hidden.

---

## Configuring Log Levels

### In application.properties:

```properties
# Root level (applies to everything)
logging.level.root=INFO

# Per-package level
logging.level.com.learn.logging=DEBUG
logging.level.org.springframework=WARN
logging.level.org.hibernate.SQL=DEBUG

# Spring framework internals (usually keep at WARN)
logging.level.org.springframework.web=INFO
```

### Via command line:
```bash
java -jar app.jar --logging.level.com.learn=DEBUG
```

---

## Log Output Format

Default Spring Boot format:
```
2026-05-20T14:48:03.041+05:30  INFO 118133 --- [app-name] [main] c.l.logging.OrderService : Order placed
│                              │    │          │           │       │                         │
│                              │    │          │           │       │                         └─ Message
│                              │    │          │           │       └─ Logger name (abbreviated)
│                              │    │          │           └─ Thread name
│                              │    │          └─ Application name
│                              │    └─ PID
│                              └─ Level
└─ Timestamp (ISO-8601)
```

### Custom format:
```properties
# Custom console pattern
logging.pattern.console=%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n

# Custom file pattern
logging.pattern.file=%d{yyyy-MM-dd HH:mm:ss} [%thread] %-5level %logger{36} - %msg%n
```

---

## Logging to Files

```properties
# Log to a file (in addition to console)
logging.file.name=app.log

# OR log to a directory (creates spring.log)
logging.file.path=/var/log/myapp

# Max file size before rotation
logging.logback.rollingpolicy.max-file-size=10MB

# Keep 7 days of log files
logging.logback.rollingpolicy.max-history=7

# Total size cap
logging.logback.rollingpolicy.total-size-cap=1GB
```

---

## Custom Logback Configuration

For advanced needs, create `src/main/resources/logback-spring.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<configuration>
    <!-- Use Spring Boot defaults as base -->
    <include resource="org/springframework/boot/logging/logback/defaults.xml"/>

    <!-- Console output -->
    <appender name="CONSOLE" class="ch.qos.logback.core.ConsoleAppender">
        <encoder>
            <pattern>%d{HH:mm:ss} %highlight(%-5level) %cyan(%logger{36}) - %msg%n</pattern>
        </encoder>
    </appender>

    <!-- File output with rotation -->
    <appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
        <file>logs/app.log</file>
        <rollingPolicy class="ch.qos.logback.core.rolling.TimeBasedRollingPolicy">
            <fileNamePattern>logs/app.%d{yyyy-MM-dd}.log</fileNamePattern>
            <maxHistory>30</maxHistory>
        </rollingPolicy>
        <encoder>
            <pattern>%d{yyyy-MM-dd HH:mm:ss} [%thread] %-5level %logger{36} - %msg%n</pattern>
        </encoder>
    </appender>

    <!-- Package-specific levels -->
    <logger name="com.learn" level="DEBUG"/>
    <logger name="org.springframework" level="WARN"/>

    <root level="INFO">
        <appender-ref ref="CONSOLE"/>
        <appender-ref ref="FILE"/>
    </root>
</configuration>
```

**Note:** Use `logback-spring.xml` (not `logback.xml`) to get Spring Boot features like profile-specific sections.

---

## Profile-Specific Logging in logback-spring.xml

```xml
<springProfile name="dev">
    <root level="DEBUG">
        <appender-ref ref="CONSOLE"/>
    </root>
</springProfile>

<springProfile name="prod">
    <root level="WARN">
        <appender-ref ref="FILE"/>
    </root>
</springProfile>
```

---

## Lombok @Slf4j (Optional Shortcut)

With Lombok, skip the boilerplate:

```java
@Slf4j  // Generates: private static final Logger log = LoggerFactory.getLogger(...)
@Service
public class OrderService {
    public void placeOrder() {
        log.info("Order placed");  // 'log' is auto-generated
    }
}
```

We won't use Lombok yet — learning the manual way first.

---

## Common Anti-Patterns

```java
// ❌ BAD: String concatenation (toString() called even if level disabled)
log.debug("User " + user.getName() + " logged in at " + timestamp);

// ✅ GOOD: Parameterized (no toString() if debug is disabled)
log.debug("User {} logged in at {}", user.getName(), timestamp);

// ❌ BAD: Checking level before logging (redundant with parameterized)
if (log.isDebugEnabled()) {
    log.debug("Value: {}", value);
}

// ✅ GOOD: Just log it (SLF4J handles the check internally)
log.debug("Value: {}", value);

// ❌ BAD: Catching exception and only logging message
catch (Exception e) {
    log.error("Failed: " + e.getMessage());  // Lost stack trace!
}

// ✅ GOOD: Pass exception as last argument
catch (Exception e) {
    log.error("Failed to process: {}", e.getMessage(), e);  // Full stack trace
}
```

---

## Key Takeaways

1. **SLF4J** = logging API; **Logback** = default implementation (both included automatically)
2. **`LoggerFactory.getLogger(MyClass.class)`** — one logger per class
3. **Use `{}` placeholders** — never concatenate strings in log statements
4. **Levels**: TRACE < DEBUG < INFO < WARN < ERROR
5. **Configure per-package**: `logging.level.com.myapp=DEBUG`
6. **File logging**: `logging.file.name=app.log` (auto-rotates)
7. **Advanced**: `logback-spring.xml` for custom patterns, multiple appenders, profile-specific config
