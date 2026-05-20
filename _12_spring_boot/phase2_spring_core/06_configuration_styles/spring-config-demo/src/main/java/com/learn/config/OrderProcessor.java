package com.learn.config;

/**
 * Another class we register via @Bean.
 * It depends on PaymentGateway — which is also a @Bean.
 */
public class OrderProcessor {

    private final PaymentGateway gateway;

    public OrderProcessor(PaymentGateway gateway) {
        this.gateway = gateway;
    }

    public String process(String orderId, double amount) {
        String result = gateway.charge(amount);
        return "Order " + orderId + " processed → " + result;
    }
}
