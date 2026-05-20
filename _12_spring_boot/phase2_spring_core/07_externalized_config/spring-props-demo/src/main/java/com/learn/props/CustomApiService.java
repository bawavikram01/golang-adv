package com.learn.props;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

/**
 * Demonstrates @PropertySource — loading from a custom file.
 * The @PropertySource("classpath:custom.properties") is on PropsDemoApp.
 * Properties from custom.properties are now available via @Value.
 */
@Component
public class CustomApiService {

    @Value("${custom.api.key}")
    private String apiKey;

    @Value("${custom.api.base-url}")
    private String baseUrl;

    @Value("${custom.api.rate-limit}")
    private int rateLimit;

    public void printConfig() {
        System.out.println("─── 3. @PROPERTYSOURCE (Custom File) ───────────");
        System.out.println();
        System.out.println("  Loaded from custom.properties:");
        System.out.println("    custom.api.key        = " + mask(apiKey));
        System.out.println("    custom.api.base-url   = " + baseUrl);
        System.out.println("    custom.api.rate-limit = " + rateLimit);
        System.out.println();
        System.out.println("  → @PropertySource(\"classpath:custom.properties\") on config class");
        System.out.println();
    }

    private String mask(String value) {
        if (value == null || value.length() <= 4) return "****";
        return value.substring(0, 4) + "****";
    }
}
