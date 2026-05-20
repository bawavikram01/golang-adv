package com.learn.springdemo;

import org.springframework.stereotype.Service;

/**
 * Business logic layer.
 * @Service is the same as @Component, but semantically means "business logic".
 *
 * This class needs TWO dependencies: UserRepository + NotificationService.
 * Spring resolves BOTH and injects them via the constructor.
 */
@Service
public class UserService {

    private final UserRepository userRepository;
    private final NotificationService notificationService;

    // Spring sees 2 parameters → looks in its container for matching beans → injects them
    public UserService(UserRepository userRepository, NotificationService notificationService) {
        this.userRepository = userRepository;
        this.notificationService = notificationService;
        System.out.println("  ✓ UserService created (deps: UserRepository, NotificationService)");
    }

    public void registerUser(String name) {
        userRepository.save(name);
        notificationService.notify(name, "Welcome to " + name + "'s account!");
        System.out.println("    ✅ User registered: " + userRepository.findByName(name));
    }
}
