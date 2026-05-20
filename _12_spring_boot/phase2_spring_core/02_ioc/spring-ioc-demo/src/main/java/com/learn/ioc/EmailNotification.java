package com.learn.ioc;

import org.springframework.context.annotation.Primary;
import org.springframework.stereotype.Component;

/**
 * Implementation 1: Email Notification.
 * 
 * @Primary tells Spring: "If someone asks for a NotificationService
 * and there are multiple implementations, USE THIS ONE by default."
 */
@Component
@Primary
public class EmailNotification implements NotificationService {

    public EmailNotification() {
        System.out.println("  [IoC] EmailNotification bean created by Spring");
    }

    @Override
    public void send(String to, String message) {
        System.out.println("    📧 [EMAIL] → " + to + ": " + message);
    }

    @Override
    public String getType() {
        return "EMAIL";
    }
}
