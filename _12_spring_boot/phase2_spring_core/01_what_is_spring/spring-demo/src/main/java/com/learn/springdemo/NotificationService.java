package com.learn.springdemo;

import org.springframework.stereotype.Component;

/**
 * Notification service.
 * Simple bean with no dependencies.
 */
@Component
public class NotificationService {

    public NotificationService() {
        System.out.println("  ✓ NotificationService created");
    }

    public void notify(String user, String message) {
        System.out.println("    📧 [NOTIFY] " + user + ": " + message);
    }
}
