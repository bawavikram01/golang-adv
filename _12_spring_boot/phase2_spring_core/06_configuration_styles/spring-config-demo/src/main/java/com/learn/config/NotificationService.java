package com.learn.config;

/**
 * Simulates a notification service — registered via @Bean in InfraConfig.
 */
public class NotificationService {

    private final String channel;

    public NotificationService(String channel) {
        this.channel = channel;
    }

    public String notify(String message) {
        return "[" + channel + "] " + message;
    }
}
