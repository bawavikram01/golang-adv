package com.learn.profiles;

import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

/**
 * DEV profile: just prints to console (no real sending).
 * Cheap, fast, no external dependencies.
 */
@Component
@Profile("dev")
public class ConsoleNotificationService implements NotificationService {

    @Override
    public String send(String message) {
        return "[DEV → CONSOLE] " + message;
    }
}
