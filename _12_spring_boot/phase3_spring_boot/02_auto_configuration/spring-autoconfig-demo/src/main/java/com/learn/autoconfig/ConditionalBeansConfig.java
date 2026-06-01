package com.learn.autoconfig;

import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * CUSTOM @Conditional usage — demonstrates how auto-configuration
 * decisions work by writing your own conditional beans.
 */
@Configuration
public class ConditionalBeansConfig {

    /**
     * This bean ONLY exists when app.feature.greeting=true in properties.
     * If the property is missing, matchIfMissing=true means it defaults to active.
     */
    @Bean
    @ConditionalOnProperty(prefix = "app.feature", name = "greeting", havingValue = "true", matchIfMissing = true)
    public GreetingService greetingService() {
        return new GreetingService("Enabled via @ConditionalOnProperty");
    }

    /**
     * This bean ONLY exists when app.feature.premium=true.
     * Since matchIfMissing defaults to false, if the property isn't set, this bean is SKIPPED.
     */
    @Bean
    @ConditionalOnProperty(prefix = "app.feature", name = "premium", havingValue = "true")
    public PremiumService premiumService() {
        return new PremiumService();
    }

    // Simple service classes (inner for demo brevity)
    public static class GreetingService {
        private final String note;

        public GreetingService(String note) { this.note = note; }

        public String greet(String name) { return "Hello, " + name + "!"; }
        public String getNote() { return note; }
    }

    public static class PremiumService {
        public String getFeature() { return "Premium analytics dashboard"; }
    }
}
