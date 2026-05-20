package com.learn.profiles;

/**
 * Interface for notification — different implementations per profile.
 */
public interface NotificationService {
    String send(String message);
}
