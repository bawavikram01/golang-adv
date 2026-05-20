package com.learn.di;

/**
 * Interface for message sending.
 * Multiple implementations exist — Spring must choose which to inject.
 */
public interface MessageSender {
    void send(String to, String message);
    String getChannel();
}
