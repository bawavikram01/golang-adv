package com.learn.scopes;

import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import org.springframework.stereotype.Component;

/**
 * SINGLETON scope (the default).
 * - Only ONE instance is created for the entire application.
 * - Every injection point gets the SAME object.
 * - @PreDestroy IS called when the container shuts down.
 */
@Component
public class SingletonService {

    private static int instanceCount = 0;
    private final int id;

    public SingletonService() {
        this.id = ++instanceCount;
    }

    @PostConstruct
    void init() {
        System.out.println("  [Singleton] @PostConstruct called (instance #" + id + ")");
    }

    @PreDestroy
    void cleanup() {
        System.out.println("  [Singleton] @PreDestroy called (instance #" + id + ")");
    }

    @Override
    public String toString() {
        return "SingletonService@" + Integer.toHexString(hashCode()) + " (instance #" + id + ")";
    }
}
