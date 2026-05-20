package com.learn.di;

import org.springframework.context.annotation.Primary;
import org.springframework.stereotype.Component;

/**
 * Implementation 1: Email sender.
 * @Primary = default choice when multiple MessageSender beans exist.
 */
@Component
@Primary
public class EmailSender implements MessageSender {

    @Override
    public void send(String to, String message) {
        System.out.println("      📧 [EMAIL] → " + to + ": " + message);
    }

    @Override
    public String getChannel() {
        return "EMAIL";
    }
}
