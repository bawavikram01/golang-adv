package com.learn.profiles;

import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

/**
 * DEFAULT profile: used when NO profile is explicitly activated.
 * Acts as a safe fallback.
 */
@Component
@Profile("default")
public class DefaultNotificationService implements NotificationService {

    @Override
    public String send(String message) {
        return "[DEFAULT → LOG ONLY] " + message + " (no profile set)";
    }
}
