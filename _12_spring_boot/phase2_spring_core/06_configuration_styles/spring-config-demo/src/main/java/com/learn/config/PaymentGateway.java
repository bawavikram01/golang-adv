package com.learn.config;

/**
 * Simulates a third-party payment gateway class.
 * We CAN'T add @Component to this (pretend it's from a library JAR).
 * So we register it via @Bean in AppConfig.
 */
public class PaymentGateway {

    private final String apiKey;

    public PaymentGateway(String apiKey) {
        this.apiKey = apiKey;
    }

    public String charge(double amount) {
        return String.format("Charged $%.2f via gateway (key=%s...)", amount, apiKey.substring(0, 4));
    }
}
