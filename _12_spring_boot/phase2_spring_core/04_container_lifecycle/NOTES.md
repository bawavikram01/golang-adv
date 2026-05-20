# Phase 2.4 — Spring Container & Bean Lifecycle

---

## ApplicationContext — The Container

`ApplicationContext` is Spring's IoC container. It's the central object that:
- Reads bean definitions (from annotations, Java config, or XML)
- Creates all beans in the correct dependency order
- Manages the complete lifecycle of every bean
- Provides lookup services (`getBean()`)
- Publishes events to listeners

### Container Hierarchy

```
BeanFactory (basic — lazy, minimal)
    └── ApplicationContext (full-featured — eager, events, i18n, AOP)
            ├── AnnotationConfigApplicationContext   (standalone Java config)
            ├── GenericWebApplicationContext         (web apps)
            └── SpringApplication.run()              (Spring Boot picks the right one)
```

You'll almost always interact with `ApplicationContext`. `BeanFactory` is the underlying interface but lacks event support, AOP auto-proxying, etc.

---

## The Complete Bean Lifecycle

Every Spring bean goes through this exact sequence:

```
┌────────────────────────────────────────────────────────────────┐
│                    BEAN LIFECYCLE (Full)                         │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Class loading & Instantiation                               │
│     └─ Spring calls constructor (via reflection)                │
│                                                                 │
│  2. Dependency Injection                                        │
│     └─ Spring sets fields / calls setters / passes ctor args    │
│                                                                 │
│  3. Aware interfaces (optional)                                 │
│     ├─ BeanNameAware.setBeanName()                              │
│     ├─ BeanFactoryAware.setBeanFactory()                        │
│     └─ ApplicationContextAware.setApplicationContext()          │
│                                                                 │
│  4. BeanPostProcessor.postProcessBeforeInitialization()         │
│     └─ Runs for EVERY bean (cross-cutting, like AOP setup)      │
│                                                                 │
│  5. Initialization                                              │
│     ├─ @PostConstruct method                                    │
│     ├─ InitializingBean.afterPropertiesSet()                    │
│     └─ Custom init-method                                       │
│                                                                 │
│  6. BeanPostProcessor.postProcessAfterInitialization()          │
│     └─ AOP proxies are created HERE                             │
│                                                                 │
│  7. ✅ BEAN IS READY — stored in container, available for use   │
│                                                                 │
│  ─── APPLICATION RUNS ───                                       │
│                                                                 │
│  8. Destruction (on context.close() / app shutdown)             │
│     ├─ @PreDestroy method                                       │
│     ├─ DisposableBean.destroy()                                 │
│     └─ Custom destroy-method                                    │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

---

## Lifecycle Hooks You'll Actually Use

| Hook | Annotation/Interface | When | Use Case |
|------|---------------------|------|----------|
| **Post-construct** | `@PostConstruct` | After DI, before use | Open connections, warm caches, validate config |
| **Pre-destroy** | `@PreDestroy` | Before bean is destroyed | Close connections, flush buffers, cleanup |
| **Aware interfaces** | `ApplicationContextAware` | After DI | Access container itself (rare, avoid if possible) |
| **BeanPostProcessor** | `implements BeanPostProcessor` | Before/after init of EVERY bean | Custom annotations, AOP, metrics |

### `@PostConstruct` and `@PreDestroy`
```java
@Component
class CacheService {
    @PostConstruct
    void warmUp() {
        // Called once after bean is fully constructed + injected
        // Perfect for: loading data, opening connections, validation
    }

    @PreDestroy
    void flush() {
        // Called once before bean is destroyed (app shutdown)
        // Perfect for: saving state, closing connections, releasing resources
    }
}
```

---

## BeanPostProcessor — The Power Tool

A `BeanPostProcessor` intercepts the creation of **every bean** in the container. Spring's internals use them heavily:
- `AutowiredAnnotationBeanPostProcessor` → processes `@Autowired`
- `CommonAnnotationBeanPostProcessor` → processes `@PostConstruct`, `@PreDestroy`
- AOP proxies → created in `postProcessAfterInitialization()`

```java
@Component
class MyBeanPostProcessor implements BeanPostProcessor {

    // Called BEFORE @PostConstruct
    public Object postProcessBeforeInitialization(Object bean, String beanName) {
        // Inspect or modify bean before initialization
        return bean;
    }

    // Called AFTER @PostConstruct — AOP proxies wrap beans HERE
    public Object postProcessAfterInitialization(Object bean, String beanName) {
        // Can return a PROXY that wraps the real bean
        return bean;  // or return new Proxy(bean)
    }
}
```

**Key insight:** When you use `@Transactional` or `@Cacheable`, a BeanPostProcessor detects the annotation and **replaces your bean with a proxy** in `postProcessAfterInitialization()`.

---

## Aware Interfaces

Let beans "know" about their environment:

| Interface | Method Called | Provides |
|-----------|-------------|----------|
| `BeanNameAware` | `setBeanName(String)` | The bean's name in the container |
| `BeanFactoryAware` | `setBeanFactory(BeanFactory)` | The factory that created it |
| `ApplicationContextAware` | `setApplicationContext(ApplicationContext)` | Full container access |
| `EnvironmentAware` | `setEnvironment(Environment)` | Properties and profiles |

```java
@Component
class MyBean implements ApplicationContextAware, BeanNameAware {
    private ApplicationContext context;
    private String beanName;

    @Override
    public void setBeanName(String name) {
        this.beanName = name;  // "myBean"
    }

    @Override
    public void setApplicationContext(ApplicationContext ctx) {
        this.context = ctx;  // Full container access
    }
}
```

**Best practice:** Avoid Aware interfaces when possible. They couple your code to Spring. Prefer constructor injection of what you need.

---

## ApplicationContext Common Methods

```java
ApplicationContext ctx = SpringApplication.run(App.class, args);

// Get a bean by type
UserService service = ctx.getBean(UserService.class);

// Get a bean by name
Object bean = ctx.getBean("userService");

// Get a bean by name + type
UserService service = ctx.getBean("userService", UserService.class);

// Get ALL beans of a type
Map<String, MessageSender> senders = ctx.getBeansOfType(MessageSender.class);

// Check if a bean exists
boolean exists = ctx.containsBean("userService");

// Get all bean names
String[] names = ctx.getBeanDefinitionNames();

// Get bean count
int count = ctx.getBeanDefinitionCount();

// Get environment (properties, profiles)
Environment env = ctx.getEnvironment();
String port = env.getProperty("server.port");
String[] activeProfiles = env.getActiveProfiles();
```

---

## Startup Order Summary

```
1. SpringApplication.run() called
2. ApplicationContext created
3. @Configuration classes processed
4. @ComponentScan runs → finds @Component classes
5. Bean definitions registered
6. BeanPostProcessors instantiated FIRST
7. Regular beans instantiated (dependency order)
8. For each bean:
     a. Constructor → b. Inject → c. Aware → d. BPP before → e. @PostConstruct → f. BPP after
9. ApplicationContext refreshed (ready)
10. CommandLineRunner / ApplicationRunner called
11. Application is RUNNING
```

---

## Key Takeaways

1. **ApplicationContext** = the IoC container, creates/manages/destroys all beans
2. **Lifecycle**: Constructor → Inject → Aware → BPP-before → @PostConstruct → BPP-after → READY → @PreDestroy
3. **@PostConstruct** = your initialization hook (most commonly used)
4. **@PreDestroy** = your cleanup hook
5. **BeanPostProcessor** = intercepts ALL beans (how @Autowired and AOP work internally)
6. **Aware interfaces** = let beans access container internals (use sparingly)
