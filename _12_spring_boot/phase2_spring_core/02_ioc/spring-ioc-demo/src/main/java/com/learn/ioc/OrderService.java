package com.learn.ioc;

import org.springframework.stereotype.Service;

/**
 * DEMONSTRATES IoC:
 * 
 * This class needs a NotificationService and a DatabaseService.
 * It does NOT create them. It does NOT know which notification impl it uses.
 * 
 * Spring INVERTS the control:
 *   - Spring creates the dependencies
 *   - Spring decides which implementation to use (@Primary → Email)
 *   - Spring injects them here
 *   - This class just USES them
 */
@Service
public class OrderService {

    private final NotificationService notifier;  // Interface type — doesn't know if Email or SMS
    private final DatabaseService database;

    // Spring sees this constructor and injects matching beans.
    // Since there are 2 NotificationService impls, Spring picks @Primary (EmailNotification).
    public OrderService(NotificationService notifier, DatabaseService database) {
        this.notifier = notifier;
        this.database = database;
        System.out.println("  [IoC] OrderService created — injected: " + notifier.getType() + " notifier");
    }

    public void placeOrder(String orderId, String userContact) {
        // Use database (which is already initialized via @PostConstruct!)
        String result = database.query("INSERT INTO orders VALUES ('" + orderId + "')");
        System.out.println("    [DB] " + result);

        // Use notification (doesn't know if it's email or SMS!)
        notifier.send(userContact, "Order " + orderId + " has been placed successfully!");
    }
}
