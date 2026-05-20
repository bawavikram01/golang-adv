package com.learn.di;

import org.springframework.stereotype.Component;

/**
 * Implementation 3: Push notification sender.
 */
@Component
public class PushSender implements MessageSender {

    @Override
    public void send(String to, String message) {
        System.out.println("      🔔 [PUSH] → " + to + ": " + message);
    }

    @Override
    public String getChannel() {
        return "PUSH";
    }
}
