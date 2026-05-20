package com.learn.ioc;

/**
 * INTERFACE: Defines the contract for notifications.
 * 
 * IoC principle: OrderService will depend on THIS interface,
 * not a concrete class. Spring decides which implementation to inject.
 */
public interface NotificationService {
    void send(String to, String message);
    String getType();
}
