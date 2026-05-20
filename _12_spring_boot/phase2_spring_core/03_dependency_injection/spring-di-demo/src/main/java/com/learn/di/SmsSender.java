package com.learn.di;

import org.springframework.stereotype.Component;

/**
 * Implementation 2: SMS sender.
 * NOT @Primary — won't be chosen by default.
 * Can be selected via @Qualifier("smsSender") or parameter name.
 */
@Component
public class SmsSender implements MessageSender {

    @Override
    public void send(String to, String message) {
        System.out.println("      📱 [SMS] → " + to + ": " + message);
    }

    @Override
    public String getChannel() {
        return "SMS";
    }
}
