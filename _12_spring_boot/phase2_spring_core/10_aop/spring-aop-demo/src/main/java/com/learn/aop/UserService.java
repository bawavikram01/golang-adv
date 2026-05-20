package com.learn.aop;

import org.springframework.stereotype.Service;

/**
 * Another service — demonstrates that the logging aspect applies
 * to ALL classes in the service package (via pointcut expression).
 */
@Service
public class UserService {

    @Audited(action = "CREATE_USER")
    public String createUser(String name, String email) {
        return "User '" + name + "' created with email " + email;
    }

    public String findUser(String name) {
        return "Found user: " + name;
    }

    public void deleteUser(String name) {
        throw new RuntimeException("Cannot delete user '" + name + "': permission denied");
    }
}
