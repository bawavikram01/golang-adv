package com.learn.boot;

import org.springframework.boot.CommandLineRunner;
import org.springframework.context.ApplicationContext;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.DispatcherServlet;

import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * Runs after startup to demonstrate what Spring Boot auto-configured.
 * All these beans were created WITHOUT us writing @Bean methods!
 */
@Component
public class AutoConfigExplorer implements CommandLineRunner {

    private final ApplicationContext context;

    public AutoConfigExplorer(ApplicationContext context) {
        this.context = context;
    }

    @Override
    public void run(String... args) {
        System.out.println();
        System.out.println("========================================");
        System.out.println("  SPRING BOOT AUTO-CONFIGURATION DEMO");
        System.out.println("========================================");
        System.out.println();
        System.out.println("  These beans were AUTO-CONFIGURED (we didn't create them):");
        System.out.println();

        // Jackson ObjectMapper — auto-configured because jackson-databind is on classpath
        ObjectMapper mapper = context.getBean(ObjectMapper.class);
        System.out.println("  ✓ ObjectMapper: " + mapper.getClass().getSimpleName());
        System.out.println("    → Auto-configured because Jackson is on classpath");

        // DispatcherServlet — auto-configured because spring-webmvc is on classpath
        DispatcherServlet dispatcher = context.getBean(DispatcherServlet.class);
        System.out.println("  ✓ DispatcherServlet: " + dispatcher.getClass().getSimpleName());
        System.out.println("    → Auto-configured because spring-webmvc is on classpath");

        // Tomcat — embedded
        System.out.println("  ✓ Embedded Tomcat: running on port 8080");
        System.out.println("    → Auto-configured because tomcat-embed is on classpath");

        // Total beans
        int beanCount = context.getBeanDefinitionCount();
        System.out.println();
        System.out.println("  Total beans in container: " + beanCount);
        System.out.println("  (Most are auto-configured infrastructure beans)");

        System.out.println();
        System.out.println("  ─── TRY THESE URLs ───────────────────────────");
        System.out.println("  http://localhost:8080/              → JSON response");
        System.out.println("  http://localhost:8080/hello?name=Vikram → query param");
        System.out.println("  http://localhost:8080/users/42      → path variable");
        System.out.println("  http://localhost:8080/actuator      → actuator endpoints");
        System.out.println("  http://localhost:8080/actuator/health → health check");
        System.out.println("  http://localhost:8080/actuator/info  → app info");
        System.out.println("  http://localhost:8080/actuator/beans → ALL beans list");
        System.out.println();
        System.out.println("  Press Ctrl+C to stop the server.");
        System.out.println("========================================");
        System.out.println();
    }
}
