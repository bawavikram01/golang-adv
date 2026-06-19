package com.learn.jackson.config;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * DEMO 4: Global Jackson configuration.
 * 
 * This customizes how ALL JSON is serialized/deserialized in the app.
 */
@Configuration
public class JacksonConfig {

    @Bean
    public ObjectMapper objectMapper() {
        ObjectMapper mapper = new ObjectMapper();

        // Java 8 Date/Time support (LocalDateTime, etc.)
        mapper.registerModule(new JavaTimeModule());

        // Don't write dates as timestamps [1234567890] — use ISO format instead
        mapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);

        // Don't fail if JSON has fields that don't map to Java (globally)
        mapper.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);

        // Skip null fields globally (can be overridden per-field)
        mapper.setDefaultPropertyInclusion(JsonInclude.Include.NON_NULL);

        return mapper;
    }
}
