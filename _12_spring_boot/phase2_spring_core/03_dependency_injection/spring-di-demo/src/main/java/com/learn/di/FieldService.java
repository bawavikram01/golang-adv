package com.learn.di;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;

/**
 * ❌ FIELD INJECTION — Avoid in production code.
 *
 * Characteristics:
 *   - Fields set via REFLECTION (bypasses constructor)
 *   - No constructor needed for deps
 *   - Can't make fields final
 *   - HIDDEN dependencies (not visible in API)
 *   - UNTESTABLE without Spring context (fields are null in plain unit tests)
 *   - Easy to accumulate too many deps (no constructor pain signal)
 *
 * ONLY acceptable in: test classes, framework prototypes, slides/tutorials
 */
@Service
public class FieldService {

    // @Autowired directly on fields — Spring uses reflection to set them
    @Autowired
    private UserRepository userRepository;  // Not final! Can't be!

    @Autowired
    @Qualifier("pushSender")  // Picks PushSender specifically
    private MessageSender messageSender;

    // No constructor for deps! Spring uses no-arg constructor + reflection.
    public FieldService() {
        System.out.println("  [FIELD-DI] Created (fields are NULL right now!)");
        // At this point: userRepository == null, messageSender == null
        // Spring will set them via reflection AFTER construction
    }

    public void alertUser(String name, String contact) {
        System.out.println("\n    [FieldService] Alerting: " + name);
        userRepository.save(name);  // Works because Spring set the field via reflection
        messageSender.send(contact, "Alert from field injection!");
    }
}
