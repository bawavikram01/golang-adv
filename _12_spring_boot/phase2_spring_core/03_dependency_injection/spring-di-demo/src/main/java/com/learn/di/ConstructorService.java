package com.learn.di;

import org.springframework.stereotype.Service;

/**
 * ✅ CONSTRUCTOR INJECTION — The recommended way.
 *
 * Characteristics:
 *   - Dependencies are FINAL (immutable after construction)
 *   - All deps are REQUIRED (app won't start if missing)
 *   - Clear signature — you see ALL dependencies at a glance
 *   - Easy to test: new ConstructorService(mockRepo, mockSender)
 *   - No @Autowired needed (single constructor, Spring 4.3+)
 */
@Service
public class ConstructorService {

    // FINAL = can't be changed after construction. Guarantees IMMUTABILITY.
    private final UserRepository userRepository;
    private final MessageSender messageSender;  // Gets @Primary (EmailSender)

    // No @Autowired needed! Spring auto-detects the single constructor.
    public ConstructorService(UserRepository userRepository, MessageSender messageSender) {
        this.userRepository = userRepository;
        this.messageSender = messageSender;
        System.out.println("  [CONSTRUCTOR-DI] Created with: "
            + userRepository.getClass().getSimpleName() + ", "
            + messageSender.getChannel());
    }

    public void registerUser(String name, String contact) {
        System.out.println("\n    [ConstructorService] Registering: " + name);
        userRepository.save(name);
        messageSender.send(contact, "Welcome " + name + "!");
    }
}
