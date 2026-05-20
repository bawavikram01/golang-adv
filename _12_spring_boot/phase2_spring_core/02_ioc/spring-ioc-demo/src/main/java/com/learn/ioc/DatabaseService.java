package com.learn.ioc;

import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import org.springframework.stereotype.Component;

/**
 * LIFECYCLE CALLBACKS — IoC Form 3
 * 
 * Spring controls when this bean is created AND destroyed.
 * You just tell it what to do at each stage using annotations.
 * 
 * Lifecycle order:
 *   1. Constructor called (bean instantiated)
 *   2. Dependencies injected
 *   3. @PostConstruct called (initialization)
 *   ... application runs ...
 *   4. @PreDestroy called (cleanup before shutdown)
 */
@Component
public class DatabaseService {

    private boolean connected = false;

    public DatabaseService() {
        System.out.println("  [IoC] DatabaseService — Step 1: Constructor called");
    }

    /**
     * Called by Spring AFTER construction + dependency injection.
     * Use for initialization logic (open connections, warm caches, etc.)
     */
    @PostConstruct
    public void initialize() {
        System.out.println("  [IoC] DatabaseService — Step 2: @PostConstruct → Connecting to DB...");
        connected = true;
        System.out.println("  [IoC] DatabaseService — Connected! ✓");
    }

    public String query(String sql) {
        if (!connected) throw new IllegalStateException("Not connected!");
        return "Result of: " + sql;
    }

    public boolean isConnected() {
        return connected;
    }

    /**
     * Called by Spring BEFORE the bean is destroyed (app shutdown).
     * Use for cleanup (close connections, flush buffers, release resources).
     */
    @PreDestroy
    public void shutdown() {
        System.out.println("  [IoC] DatabaseService — @PreDestroy → Closing DB connection...");
        connected = false;
        System.out.println("  [IoC] DatabaseService — Disconnected. Goodbye! ✓");
    }
}
