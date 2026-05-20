package com.learn.di;

import org.springframework.stereotype.Component;

/**
 * A simple repository (dependency for services below).
 */
@Component
public class UserRepository {

    public String findUser(String name) {
        return "User(" + name + ")";
    }

    public void save(String name) {
        System.out.println("      [DB] Saved: " + name);
    }
}
