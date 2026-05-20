package com.learn.springdemo;

import org.springframework.stereotype.Repository;

/**
 * Data access layer.
 * @Repository is the same as @Component, but semantically means "data access".
 * 
 * NOTICE: We declare a dependency (AppConfig) in the constructor.
 * Spring automatically provides it — we don't call "new AppConfig()".
 */
@Repository
public class UserRepository {

    private final AppConfig config;

    // CONSTRUCTOR INJECTION — Spring sees this constructor,
    // knows UserRepository needs an AppConfig, and passes it in.
    public UserRepository(AppConfig config) {
        this.config = config;
        System.out.println("  ✓ UserRepository created (db: " + config.getDbUrl() + ")");
    }

    public String findByName(String name) {
        return "User(" + name + ") from " + config.getDbUrl();
    }

    public void save(String name) {
        System.out.println("    [DB] Saved user: " + name);
    }
}
