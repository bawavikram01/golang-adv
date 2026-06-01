# Phase 3.1 — What is Spring Boot & Why

---

## The Problem Spring Boot Solves

Before Spring Boot, starting a Spring project required:

```
1. Create Maven/Gradle project manually
2. Add 15-30 dependencies (and hope versions are compatible)
3. Write web.xml (servlet config)
4. Write applicationContext.xml (Spring config)
5. Configure DataSource bean
6. Configure ViewResolver bean
7. Configure DispatcherServlet
8. Configure transaction manager
9. Download and configure Tomcat separately
10. Deploy WAR file to Tomcat
11. Fix version conflicts between Spring modules
12. Configure logging framework
... (20 more steps before "Hello World")
```

**Spring Boot's answer:** All of this should be automatic. You write business code, we handle the plumbing.

---

## Spring Boot = Spring + Opinions + Automation

| Feature | What it does |
|---------|-------------|
| **Starter POMs** | Pre-packaged compatible dependencies (one import = everything you need) |
| **Auto-Configuration** | Detects what's on classpath and configures beans automatically |
| **Embedded Server** | Tomcat/Jetty/Undertow inside your JAR (no external server needed) |
| **Opinionated Defaults** | Sensible defaults that work for 90% of cases |
| **Production-Ready** | Health checks, metrics, externalized config out of the box |

---

## Starters — Pre-Packaged Dependencies

Instead of adding 10 JARs manually, add ONE starter:

| Starter | What it brings |
|---------|---------------|
| `spring-boot-starter-web` | Tomcat + Spring MVC + Jackson (JSON) |
| `spring-boot-starter-data-jpa` | Hibernate + Spring Data JPA + HikariCP |
| `spring-boot-starter-security` | Spring Security + auth filters |
| `spring-boot-starter-test` | JUnit5 + Mockito + AssertJ + Spring Test |
| `spring-boot-starter-validation` | Hibernate Validator + Jakarta Validation |
| `spring-boot-starter-actuator` | Health endpoints + metrics |
| `spring-boot-starter-mail` | JavaMail + Spring Mail abstraction |

```xml
<!-- Just ONE dependency — brings Tomcat + Spring MVC + Jackson + all transitive deps -->
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-web</artifactId>
</dependency>
```

---

## Auto-Configuration — The Magic

When Spring Boot starts, it:

1. Scans the classpath for libraries (Jackson? Hibernate? Tomcat?)
2. If a library is found, it auto-configures beans for it
3. Uses `@Conditional` annotations to decide what to configure

```
Classpath has: jackson-databind.jar
  → Spring Boot auto-creates: ObjectMapper bean

Classpath has: spring-webmvc.jar + tomcat-embed.jar
  → Spring Boot auto-creates: DispatcherServlet + EmbeddedTomcat

Classpath has: HikariCP.jar + h2.jar
  → Spring Boot auto-creates: DataSource bean (H2, HikariCP pool)
```

### You can always override:
```java
// Spring Boot auto-configures ObjectMapper, but you can provide YOUR own:
@Bean
public ObjectMapper objectMapper() {
    return new ObjectMapper()
        .enable(SerializationFeature.INDENT_OUTPUT);
    // YOUR bean takes precedence over auto-configured one
}
```

---

## @SpringBootApplication — Three in One

```java
@SpringBootApplication  // This ONE annotation = three annotations combined
public class MyApp {
    public static void main(String[] args) {
        SpringApplication.run(MyApp.class, args);
    }
}
```

Equivalent to:
```java
@Configuration           // This class is a config class (can have @Bean methods)
@EnableAutoConfiguration // Turn on auto-configuration magic
@ComponentScan           // Scan this package + sub-packages for @Component beans
public class MyApp { ... }
```

---

## Embedded Server — No External Tomcat

**Traditional (before Spring Boot):**
```
Code → WAR file → Deploy to standalone Tomcat → Run Tomcat
```

**Spring Boot:**
```
Code → Executable JAR (includes Tomcat inside) → java -jar app.jar
```

The "fat JAR" contains:
- Your code
- All dependencies
- Embedded Tomcat (or Jetty/Undertow)
- Spring Boot launcher

```bash
# That's it. Tomcat starts automatically on port 8080.
java -jar my-application-1.0.0.jar
```

---

## Spring Boot vs Spring Framework

| | Spring Framework | Spring Boot |
|---|---|---|
| **Setup** | Manual (lots of XML/config) | Automatic (starters + auto-config) |
| **Dependencies** | You manage versions | Starter BOM manages compatibility |
| **Server** | External Tomcat (WAR deployment) | Embedded (executable JAR) |
| **Configuration** | Extensive boilerplate | Convention over configuration |
| **Production** | Add your own health checks | Actuator built-in |
| **Learning** | Know everything to start | Start immediately, learn as you go |

**Spring Boot IS Spring Framework** — it just removes the boilerplate. Every Spring concept you learned (IoC, DI, AOP, Events) works exactly the same.

---

## The Spring Boot Startup Sequence

```
1. main() → SpringApplication.run(MyApp.class, args)
2. Create SpringApplication instance
3. Detect application type (SERVLET, REACTIVE, NONE)
4. Load ApplicationContextInitializers
5. Load ApplicationListeners
6. Create ApplicationContext
7. Run auto-configuration (read META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports)
8. Evaluate @Conditional annotations (skip configs that don't apply)
9. Register all qualifying beans
10. Start embedded web server (if web app)
11. Run ApplicationRunner / CommandLineRunner beans
12. Application is READY
```

---

## Auto-Configuration Report

See exactly what was auto-configured:

```bash
java -jar app.jar --debug
# OR
java -jar app.jar -Ddebug
```

This prints:
```
============================
CONDITIONS EVALUATION REPORT
============================

Positive matches:       (things that WERE auto-configured)
-----------------
   DataSourceAutoConfiguration matched
   JacksonAutoConfiguration matched

Negative matches:       (things that were NOT configured and why)
-----------------
   MongoAutoConfiguration:
      Did not match: @ConditionalOnClass did not find class 'com.mongodb.client.MongoClient'
```

---

## Key Takeaways

1. **Spring Boot = Spring + automation** (not a different framework)
2. **Starters** = pre-packaged compatible dependencies
3. **Auto-configuration** = detects classpath → creates beans automatically
4. **Embedded server** = JAR includes Tomcat, just `java -jar`
5. **@SpringBootApplication** = @Configuration + @EnableAutoConfiguration + @ComponentScan
6. **You can always override** auto-configured beans with your own @Bean
7. **`--debug` flag** shows what was auto-configured and what wasn't
