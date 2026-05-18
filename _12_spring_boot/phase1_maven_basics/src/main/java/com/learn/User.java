package com.learn;

/**
 * A simple User class.
 * We'll convert this to/from JSON using Gson (a DEPENDENCY managed by Maven).
 */
public class User {
    private String name;
    private String email;
    private int age;

    public User() {} // Gson needs a no-arg constructor

    public User(String name, String email, int age) {
        this.name = name;
        this.email = email;
        this.age = age;
    }

    // Getters
    public String getName() { return name; }
    public String getEmail() { return email; }
    public int getAge() { return age; }

    @Override
    public String toString() {
        return "User{name='" + name + "', email='" + email + "', age=" + age + "}";
    }
}
