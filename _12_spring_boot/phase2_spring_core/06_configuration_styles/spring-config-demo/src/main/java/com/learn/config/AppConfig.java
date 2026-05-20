package com.learn.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * Style 2: Java Configuration.
 *
 * Use @Configuration + @Bean to register beans that:
 *   - Come from third-party libraries (can't annotate their source)
 *   - Require complex initialization logic
 *   - Need specific constructor arguments or configuration
 *
 * KEY: @Configuration creates a CGLIB proxy of this class, which ensures
 * that calling @Bean methods returns the SAME singleton instance.
 */
@Configuration
public class AppConfig {

    /**
     * Register a PaymentGateway bean.
     * We can't put @Component on it (pretend it's from a library).
     * Method name "paymentGateway" becomes the bean name.
     */
    @Bean
    public PaymentGateway paymentGateway() {
        // Complex creation: passing constructor args, configuring...
        return new PaymentGateway("sk_live_abc123xyz");
    }

    /**
     * @Bean method with a parameter: Spring auto-injects 'gateway' from
     * the container. Same as constructor injection!
     */
    @Bean
    public OrderProcessor orderProcessor(PaymentGateway gateway) {
        return new OrderProcessor(gateway);
    }

    /**
     * @Bean with a custom name (not the method name).
     */
    @Bean("applicationName")
    public String appName() {
        return "Spring Config Demo v1.0";
    }
}
