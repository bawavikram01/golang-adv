package com.learn.config;

import org.springframework.stereotype.Service;

/**
 * Style 1: Component Scanning.
 * Spring finds this via @Service (a specialization of @Component).
 * No need for @Bean — auto-registered!
 */
@Service
public class GreetingService {

    public String greet(String name) {
        return "Hello, " + name + "!";
    }
}
