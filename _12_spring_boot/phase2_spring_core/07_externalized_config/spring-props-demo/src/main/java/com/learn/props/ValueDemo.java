package com.learn.props;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * Demonstrates @Value injection — the simplest way to read properties.
 * Good for: a few scattered properties.
 * Limitation: no type safety, no validation, harder to test.
 */
@Component
public class ValueDemo {

    // Basic property injection
    @Value("${app.name}")
    private String appName;

    @Value("${app.version}")
    private String version;

    // Auto type conversion (String → int)
    @Value("${app.max-retries}")
    private int maxRetries;

    // Boolean conversion
    @Value("${app.debug}")
    private boolean debug;

    // Default value if property is missing
    @Value("${app.missing.key:DEFAULT_VALUE}")
    private String withDefault;

    // Property placeholder inside a property (resolved recursively)
    @Value("${app.welcome-message}")
    private String welcomeMessage;

    // Comma-separated → List<String>
    @Value("${app.supported-languages}")
    private List<String> languages;

    // SpEL: Spring Expression Language
    @Value("#{${app.max-retries} * 2}")
    private int doubledRetries;

    @Value("#{systemProperties['user.home']}")
    private String userHome;

    public void printValues() {
        System.out.println("─── 1. @VALUE INJECTION ────────────────────────");
        System.out.println();
        System.out.println("  Basic:");
        System.out.println("    app.name         = " + appName);
        System.out.println("    app.version      = " + version);
        System.out.println("    app.max-retries  = " + maxRetries + " (auto: String → int)");
        System.out.println("    app.debug        = " + debug + " (auto: String → boolean)");
        System.out.println();
        System.out.println("  Default values:");
        System.out.println("    ${app.missing.key:DEFAULT_VALUE} = " + withDefault);
        System.out.println();
        System.out.println("  Property placeholder in property:");
        System.out.println("    welcome-message = " + welcomeMessage);
        System.out.println();
        System.out.println("  Comma-separated → List:");
        System.out.println("    supported-languages = " + languages);
        System.out.println();
        System.out.println("  SpEL expressions:");
        System.out.println("    #{max-retries * 2}           = " + doubledRetries);
        System.out.println("    #{systemProperties['user.home']} = " + userHome);
        System.out.println();
    }
}
