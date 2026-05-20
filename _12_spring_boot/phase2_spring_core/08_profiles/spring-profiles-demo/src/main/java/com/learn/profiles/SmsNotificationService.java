package com.learn.profiles;

import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

/**
 * PROD profile: sends real SMS/email (simulated here).
 * In reality, this would call Twilio, SendGrid, etc.
 */
@Component
@Profile("prod")
public class SmsNotificationService implements NotificationService {

    @Override
    public String send(String message) {
        return "[PROD → SMS GATEWAY] " + message + " (sent to +1-555-0100)";
    }
}
