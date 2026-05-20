# Phase 2.8 — Profiles

---

## What Are Profiles?

Profiles let you activate different beans and configuration based on the **environment**:
- `dev` — local development (H2 database, debug logging, mock services)
- `test` — automated testing (in-memory everything, test doubles)
- `staging` — pre-production (real services, test data)
- `prod` — production (real everything, optimized settings)

Without profiles, you'd need separate builds for each environment. With profiles, **one JAR works everywhere** — just activate the right profile.

---

## Activating Profiles

### 1. application.properties:
```properties
spring.profiles.active=dev
```

### 2. Command-line argument:
```bash
java -jar app.jar --spring.profiles.active=prod
```

### 3. Environment variable:
```bash
export SPRING_PROFILES_ACTIVE=prod
java -jar app.jar
```

### 4. Programmatically:
```java
SpringApplication app = new SpringApplication(MyApp.class);
app.setAdditionalProfiles("dev");
app.run(args);
```

### Multiple profiles (comma-separated):
```bash
java -jar app.jar --spring.profiles.active=prod,metrics,ssl
```

---

## Profile-Specific Property Files

Spring Boot automatically loads profile-specific files:

```
src/main/resources/
├── application.properties             ← always loaded (base/common)
├── application-dev.properties         ← loaded when profile=dev
├── application-prod.properties        ← loaded when profile=prod
└── application-test.properties        ← loaded when profile=test
```

**Merge behavior:** Profile-specific properties **override** base properties. Unspecified properties fall back to the base file.

```properties
# application.properties (base)
app.name=MyApp
app.log-level=INFO
server.port=8080

# application-dev.properties (overrides for dev)
app.log-level=DEBUG
server.port=9090
# app.name is still "MyApp" — inherited from base
```

---

## @Profile on Beans

Conditionally register beans based on the active profile:

```java
public interface NotificationService {
    void send(String message);
}

@Component
@Profile("dev")
public class ConsoleNotifier implements NotificationService {
    public void send(String message) {
        System.out.println("[DEV CONSOLE] " + message);
    }
}

@Component
@Profile("prod")
public class SmsNotifier implements NotificationService {
    public void send(String message) {
        // Real SMS gateway call
    }
}
```

### Profile expressions:
```java
@Profile("dev")              // Active only in dev
@Profile("!prod")            // Active in everything EXCEPT prod
@Profile({"dev", "test"})    // Active in dev OR test
@Profile("prod & ssl")       // Active only when BOTH prod AND ssl are active
```

---

## @Profile on @Configuration Classes

Apply to entire configuration classes:

```java
@Configuration
@Profile("dev")
public class DevConfig {
    @Bean
    public DataSource dataSource() {
        // H2 in-memory for development
        return new EmbeddedDatabaseBuilder()
            .setType(EmbeddedDatabaseType.H2)
            .build();
    }
}

@Configuration
@Profile("prod")
public class ProdConfig {
    @Bean
    public DataSource dataSource() {
        // PostgreSQL connection pool for production
        HikariDataSource ds = new HikariDataSource();
        ds.setJdbcUrl("jdbc:postgresql://prod-server:5432/myapp");
        return ds;
    }
}
```

---

## Default Profile

If NO profile is activated, Spring uses `"default"`:

```java
@Component
@Profile("default")
public class FallbackService implements MyService {
    // Only used when no profile is explicitly set
}
```

You can also set a default profile:
```properties
spring.profiles.default=dev
```

---

## Common Profile Patterns

### Pattern 1: Interface + profile-specific implementations
```
NotificationService (interface)
├── @Profile("dev") → ConsoleNotifier
├── @Profile("test") → MockNotifier
└── @Profile("prod") → SmsNotifier
```

### Pattern 2: Separate config per concern
```
@Configuration @Profile("dev")  → DevDatabaseConfig
@Configuration @Profile("prod") → ProdDatabaseConfig
@Configuration (no profile)     → CommonConfig (always active)
```

### Pattern 3: Feature flags via profiles
```bash
java -jar app.jar --spring.profiles.active=prod,cache,metrics
```
```java
@Configuration
@Profile("cache")
public class CacheConfig { ... }

@Configuration
@Profile("metrics")
public class MetricsConfig { ... }
```

---

## Checking Active Profiles at Runtime

```java
@Component
public class MyService {

    private final Environment environment;

    public MyService(Environment environment) {
        this.environment = environment;
    }

    public void doSomething() {
        String[] profiles = environment.getActiveProfiles();
        if (environment.acceptsProfiles(Profiles.of("prod"))) {
            // production-specific logic
        }
    }
}
```

---

## Key Takeaways

1. **Profiles** = switch beans and config per environment without changing code
2. **Activate**: `spring.profiles.active=dev` (properties, CLI, env var)
3. **Profile files**: `application-{profile}.properties` auto-loaded, overrides base
4. **@Profile("dev")** on `@Component`/`@Configuration` = conditionally register beans
5. **Expressions**: `!prod`, `{"dev","test"}`, `prod & ssl`
6. **One JAR, many environments** — deploy same artifact everywhere
7. **Default profile**: `"default"` active when nothing is explicitly set
