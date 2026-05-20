package com.learn.di;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;

/**
 * ⚠️ SETTER INJECTION — For optional dependencies.
 *
 * Characteristics:
 *   - Dependencies are NOT final (can be null, can change)
 *   - Can be marked required=false (truly optional)
 *   - Object constructed first, then setters called
 *   - Must null-check optional deps before using
 *   - @Autowired IS required on the setter method
 */
@Service
public class SetterService {

    // Required dep — injected via constructor (hybrid approach)
    private final UserRepository userRepository;

    // Optional dep — injected via setter (may or may not be present)
    private MessageSender messageSender;

    // Constructor for required deps
    public SetterService(UserRepository userRepository) {
        this.userRepository = userRepository;
        System.out.println("  [SETTER-DI] Created (setter deps not yet injected!)");
    }

    // Setter for optional dep — @Autowired required here
    // @Qualifier selects a SPECIFIC bean by name (not @Primary)
    @Autowired
    @Qualifier("smsSender")  // Picks SmsSender specifically, ignoring @Primary
    public void setMessageSender(MessageSender messageSender) {
        this.messageSender = messageSender;
        System.out.println("  [SETTER-DI] Setter called — injected: " + messageSender.getChannel());
    }

    public void notifyUser(String name, String contact) {
        System.out.println("\n    [SetterService] Notifying: " + name);
        String user = userRepository.findUser(name);
        System.out.println("      Found: " + user);

        // Must null-check because setter injection deps can be null!
        if (messageSender != null) {
            messageSender.send(contact, "Hello from setter injection!");
        } else {
            System.out.println("      (No sender configured — skipping notification)");
        }
    }
}
