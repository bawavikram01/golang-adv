# Phase 2.6 — Configuration Styles

---

## Three Ways to Define Beans

Spring gives you three approaches to register beans in the container:

| Style | How | When to Use |
|-------|-----|-------------|
| **Component Scanning** | `@Component` + `@ComponentScan` | Your own classes (90% of the time) |
| **Java Config** | `@Configuration` + `@Bean` methods | Third-party classes, complex creation logic |
| **XML Config** | `<bean>` in .xml files | Legacy apps only (avoid in new projects) |

---

## 1. Component Scanning (Stereotype Annotations)

Spring scans packages and auto-registers classes with these annotations:

```java
@Component       // Generic bean
@Service         // Business logic layer
@Repository      // Data access layer (+ exception translation)
@Controller      // Web MVC controller
@RestController  // REST API controller (@Controller + @ResponseBody)
@Configuration   // Config class (itself is a bean too)
```

### How it works:
```java
@SpringBootApplication  // includes @ComponentScan for current package + sub-packages
public class MyApp { }
```

`@SpringBootApplication` = `@Configuration` + `@EnableAutoConfiguration` + `@ComponentScan`

Spring Boot auto-scans the package of your main class and everything below it.

### Custom scanning:
```java
@ComponentScan(basePackages = {"com.learn.services", "com.learn.repos"})
```

---

## 2. Java Configuration (@Configuration + @Bean)

When you can't add `@Component` to a class (e.g., third-party libraries), use `@Bean`:

```java
@Configuration
public class AppConfig {

    @Bean
    public RestTemplate restTemplate() {
        return new RestTemplate();  // Third-party class — can't annotate it
    }

    @Bean
    public ObjectMapper objectMapper() {
        ObjectMapper mapper = new ObjectMapper();
        mapper.enable(SerializationFeature.INDENT_OUTPUT);
        mapper.registerModule(new JavaTimeModule());
        return mapper;  // Complex creation logic
    }
}
```

### Key Rules:
1. Method name = bean name (unless you specify `@Bean("customName")`)
2. Return type = bean type
3. Method body = factory logic (can be as complex as needed)
4. Parameters = automatically injected dependencies

```java
@Bean
public UserService userService(UserRepository repo, EmailService email) {
    // 'repo' and 'email' are auto-injected from the container
    return new UserService(repo, email);
}
```

---

## 3. @Configuration vs @Component with @Bean

```java
// FULL configuration — @Bean methods are proxied (inter-bean references work)
@Configuration
public class AppConfig {

    @Bean
    public ServiceA serviceA() {
        return new ServiceA();
    }

    @Bean
    public ServiceB serviceB() {
        // Calling serviceA() returns the SAME singleton (proxied!)
        return new ServiceB(serviceA());
    }
}

// LITE mode — no proxying (each call creates a NEW instance!)
@Component
public class LiteConfig {

    @Bean
    public ServiceA serviceA() {
        return new ServiceA();
    }

    @Bean
    public ServiceB serviceB() {
        // WARNING: serviceA() creates a NEW instance (not the singleton!)
        return new ServiceB(serviceA());
    }
}
```

**Rule:** Always use `@Configuration` for config classes, not `@Component`.

---

## @Bean Method Features

### Custom names:
```java
@Bean("myCustomName")
public MyService service() { ... }

@Bean({"name1", "alias1", "alias2"})  // Multiple names
public MyService service() { ... }
```

### Init and destroy methods:
```java
@Bean(initMethod = "init", destroyMethod = "cleanup")
public DataSource dataSource() {
    return new HikariDataSource();
}
```

### Conditional beans:
```java
@Bean
@ConditionalOnProperty(name = "feature.cache", havingValue = "true")
public CacheManager cacheManager() {
    return new RedisCacheManager();
}
```

---

## When Component Scanning vs @Bean?

| Scenario | Use |
|----------|-----|
| Your own service/repo/controller class | `@Component` / `@Service` / `@Repository` |
| Third-party library class (RestTemplate, ObjectMapper) | `@Bean` in `@Configuration` |
| Complex creation logic (builder, conditional setup) | `@Bean` in `@Configuration` |
| Need multiple beans of same type with different config | `@Bean` methods |
| Simple class you own, no special setup | `@Component` |

---

## @Import — Composing Configurations

```java
@Configuration
@Import({DatabaseConfig.class, SecurityConfig.class, CacheConfig.class})
public class AppConfig {
    // Pulls in beans from all imported config classes
}
```

---

## Profile-Specific Configuration

```java
@Configuration
@Profile("dev")
public class DevConfig {
    @Bean
    public DataSource dataSource() {
        return new H2DataSource();  // In-memory for development
    }
}

@Configuration
@Profile("prod")
public class ProdConfig {
    @Bean
    public DataSource dataSource() {
        return new PostgresDataSource();  // Real DB for production
    }
}
```

Activate via: `spring.profiles.active=dev` in application.properties

---

## Key Takeaways

1. **@Component** = auto-detected by scanning (your classes)
2. **@Bean** = explicit factory method in @Configuration (third-party or complex beans)
3. **@Configuration** = full proxy mode (safe inter-bean references)
4. **@SpringBootApplication** = @Configuration + @ComponentScan + @EnableAutoConfiguration
5. **Method name** = bean name; **return type** = bean type; **parameters** = auto-injected
6. In real projects, you'll use BOTH: `@Component` for your code + `@Bean` for libraries
