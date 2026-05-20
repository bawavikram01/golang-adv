package com.learn.ioc;

import org.springframework.stereotype.Component;

/**
 * Implementation 2: SMS Notification.
 * 
 * This exists alongside EmailNotification.
 * Both implement NotificationService.
 * Spring uses @Primary on EmailNotification to resolve the conflict.
 */
@Component
public class SmsNotification implements NotificationService {

    public SmsNotification() {
        System.out.println("  [IoC] SmsNotification bean created by Spring");
    }

    @Override
    public void send(String to, String message) {
        System.out.println("    📱 [SMS] → " + to + ": " + message);
    }

    @Override
    public String getType() {
        return "SMS";
    }
}
