package com.learn.autoconfig;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * OVERRIDING AUTO-CONFIGURATION.
 *
 * Spring Boot auto-configures an ObjectMapper with default settings.
 * But because we define OUR OWN @Bean of type ObjectMapper here,
 * Spring Boot's auto-configured one is SKIPPED.
 *
 * Why? Because JacksonAutoConfiguration uses:
 *   @Bean
 *   @ConditionalOnMissingBean   ← "only if no ObjectMapper bean exists"
 *   public ObjectMapper objectMapper() { ... }
 *
 * Our bean exists → condition fails → auto-config backs off.
 * THIS is the override mechanism.
 */
@Configuration
public class CustomJacksonConfig {

    @Bean
    public ObjectMapper objectMapper() {
        ObjectMapper mapper = new ObjectMapper();

        // Our custom settings (auto-configured one doesn't have these)
        mapper.enable(SerializationFeature.INDENT_OUTPUT);      // Pretty-print JSON
        mapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);  // ISO dates

        return mapper;
    }
}
