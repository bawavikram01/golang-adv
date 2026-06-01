# Phase 3.2 — Auto-Configuration Deep Dive

---

## Standard Spring Boot Project Structure

```
my-app/
├── pom.xml                              ← Maven build config + dependencies
├── src/
│   ├── main/
│   │   ├── java/
│   │   │   └── com/example/myapp/
│   │   │       ├── MyAppApplication.java       ← @SpringBootApplication (main class)
│   │   │       ├── controller/                  ← REST controllers
│   │   │       ├── service/                     ← Business logic
│   │   │       ├── repository/                  ← Data access
│   │   │       ├── model/                       ← Entities / DTOs
│   │   │       └── config/                      ← @Configuration classes
│   │   └── resources/
│   │       ├── application.properties           ← Config
│   │       ├── application-dev.properties       ← Profile-specific
│   │       ├── static/                          ← Static files (CSS, JS)
│   │       └── templates/                       ← Thymeleaf templates
│   └── test/
│       └── java/
│           └── com/example/myapp/
│               └── MyAppApplicationTests.java   ← Tests
└── target/                              ← Build output (JAR lives here)
```

**Rule:** The main class (`@SpringBootApplication`) must be in the ROOT package. `@ComponentScan` scans that package + everything below it.

---

## How Auto-Configuration Works Internally

### Step 1: `@EnableAutoConfiguration` triggers the process

```java
@SpringBootApplication  // includes @EnableAutoConfiguration
public class MyApp { }
```

### Step 2: Spring reads the auto-configuration registry

In Spring Boot 3.x, auto-configuration classes are listed in:
```
META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports
```

This file (inside spring-boot-autoconfigure.jar) contains ~150 configuration class names:
```
org.springframework.boot.autoconfigure.jackson.JacksonAutoConfiguration
org.springframework.boot.autoconfigure.web.servlet.DispatcherServletAutoConfiguration
org.springframework.boot.autoconfigure.jdbc.DataSourceAutoConfiguration
org.springframework.boot.autoconfigure.orm.jpa.HibernateJpaAutoConfiguration
...
```

### Step 3: Each auto-config class has @Conditional guards

```java
@AutoConfiguration
@ConditionalOnClass(ObjectMapper.class)  // Only if Jackson is on classpath
@ConditionalOnMissingBean(ObjectMapper.class)  // Only if user didn't define their own
public class JacksonAutoConfiguration {

    @Bean
    @ConditionalOnMissingBean
    public ObjectMapper objectMapper() {
        return new ObjectMapper();  // Creates a default ObjectMapper
    }
}
```

### Step 4: Conditions evaluated → beans registered (or skipped)

---

## @Conditional Annotations

| Annotation | Condition |
|-----------|-----------|
| `@ConditionalOnClass` | Class exists on classpath |
| `@ConditionalOnMissingClass` | Class does NOT exist on classpath |
| `@ConditionalOnBean` | Bean already exists in context |
| `@ConditionalOnMissingBean` | Bean does NOT exist (user didn't define it) |
| `@ConditionalOnProperty` | Property has specific value |
| `@ConditionalOnResource` | Resource file exists on classpath |
| `@ConditionalOnWebApplication` | Running as a web app |
| `@ConditionalOnNotWebApplication` | NOT a web app |
| `@ConditionalOnExpression` | SpEL expression evaluates to true |

### The Golden Rule:
```
@ConditionalOnMissingBean = "Only create this bean if the user hasn't defined their own"
```

This is why **you can always override** auto-configured beans:
```java
@Bean  // Your bean takes precedence!
public ObjectMapper objectMapper() {
    return new ObjectMapper().enable(SerializationFeature.INDENT_OUTPUT);
}
// Spring Boot sees you defined ObjectMapper → skips its auto-configured one
```

---

## Examining Auto-Configuration

### 1. Debug report (CLI):
```bash
java -jar app.jar --debug
```

### 2. Actuator conditions endpoint:
```
GET http://localhost:8080/actuator/conditions
```
Returns JSON of all positive/negative matches.

### 3. In code:
```java
@Bean
public CommandLineRunner printAutoConfig(ConditionEvaluationReport report) {
    return args -> {
        report.getConditionAndOutcomesBySource().forEach((source, outcomes) -> {
            System.out.println(source + " → " + (outcomes.isFullMatch() ? "MATCHED" : "SKIPPED"));
        });
    };
}
```

---

## Creating Your Own Auto-Configuration

You can make your own starters! Pattern:

```java
@AutoConfiguration
@ConditionalOnClass(MyLibrary.class)
@ConditionalOnProperty(prefix = "mylib", name = "enabled", havingValue = "true", matchIfMissing = true)
@EnableConfigurationProperties(MyLibProperties.class)
public class MyLibAutoConfiguration {

    @Bean
    @ConditionalOnMissingBean
    public MyLibrary myLibrary(MyLibProperties props) {
        return new MyLibrary(props.getApiKey(), props.getTimeout());
    }
}
```

Register it in: `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports`

---

## Excluding Auto-Configurations

Don't want something auto-configured? Exclude it:

```java
@SpringBootApplication(exclude = {
    DataSourceAutoConfiguration.class,    // Don't auto-configure database
    SecurityAutoConfiguration.class       // Don't auto-configure security
})
public class MyApp { }
```

Or in properties:
```properties
spring.autoconfigure.exclude=org.springframework.boot.autoconfigure.jdbc.DataSourceAutoConfiguration
```

---

## spring-boot-starter-parent

Your POM inherits from `spring-boot-starter-parent`:

```xml
<parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>3.3.0</version>
</parent>
```

What it provides:
- **Dependency management** — all Spring dependency versions pre-defined (no version conflicts)
- **Plugin configuration** — maven-compiler-plugin, spring-boot-maven-plugin pre-configured
- **Default properties** — Java version, encoding, resource filtering
- **Dependency BOM** — imports `spring-boot-dependencies` (manages ~400 dependency versions)

---

## Key Takeaways

1. **Auto-config** = Spring reads a registry of ~150 config classes, evaluates @Conditional, creates beans
2. **@ConditionalOnClass** = "if this library is on classpath, configure it"
3. **@ConditionalOnMissingBean** = "only if user didn't already define this bean"
4. **Your beans always win** — override any auto-configured bean with your own @Bean
5. **`--debug` flag** = shows exactly what was/wasn't auto-configured
6. **Exclude** = `@SpringBootApplication(exclude = ...)` to disable unwanted auto-config
7. **Project structure**: main class at root package, controllers/services/repos in sub-packages
