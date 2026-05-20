package com.learn.config;

import org.springframework.stereotype.Repository;

import java.util.List;

/**
 * Style 1: Component Scanning.
 * @Repository = @Component + extra behavior for data access:
 *   - Exception translation (DB exceptions → Spring's DataAccessException)
 */
@Repository
public class UserRepository {

    private final List<String> users = List.of("Alice", "Bob", "Charlie");

    public int count() {
        return users.size();
    }

    public List<String> findAll() {
        return users;
    }
}
