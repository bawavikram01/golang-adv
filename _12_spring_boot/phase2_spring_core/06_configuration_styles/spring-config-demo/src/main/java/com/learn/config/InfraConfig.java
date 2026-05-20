package com.learn.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * A separate @Configuration class for infrastructure beans.
 * In real apps, you split configs by concern:
 *   - AppConfig        (business beans)
 *   - InfraConfig      (infrastructure: messaging, caching)
 *   - SecurityConfig   (security beans)
 *   - DatabaseConfig   (datasources, connection pools)
 *
 * Spring Boot auto-detects all @Configuration classes in scanned packages.
 */
@Configuration
public class InfraConfig {

    @Bean
    public NotificationService notificationService() {
        // Could read channel from properties in real app
        return new NotificationService("EMAIL");
    }
}
