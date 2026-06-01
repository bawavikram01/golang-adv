package com.learn.boot;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.time.LocalDateTime;
import java.util.Map;

/**
 * YOUR FIRST REST CONTROLLER.
 *
 * Thanks to auto-configuration:
 *   - Tomcat is running (we didn't configure it)
 *   - DispatcherServlet is registered (we didn't configure it)
 *   - Jackson converts objects to JSON (we didn't configure it)
 *   - This controller is detected via @ComponentScan
 *
 * ALL of that happened because of ONE dependency: spring-boot-starter-web
 */
@RestController
public class HelloController {

    // GET http://localhost:8080/
    @GetMapping("/")
    public Map<String, Object> home() {
        return Map.of(
            "message", "Welcome to Spring Boot!",
            "timestamp", LocalDateTime.now().toString(),
            "info", "This response is auto-serialized to JSON by Jackson (auto-configured)"
        );
    }

    // GET http://localhost:8080/hello?name=Vikram
    @GetMapping("/hello")
    public Map<String, String> hello(@RequestParam(defaultValue = "World") String name) {
        return Map.of(
            "greeting", "Hello, " + name + "!",
            "note", "@RequestParam auto-binds query parameters"
        );
    }

    // GET http://localhost:8080/users/42
    @GetMapping("/users/{id}")
    public Map<String, Object> getUser(@PathVariable int id) {
        return Map.of(
            "id", id,
            "name", "User-" + id,
            "note", "@PathVariable extracts values from the URL path"
        );
    }
}
