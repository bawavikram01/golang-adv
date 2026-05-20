package com.learn.props;

import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.PropertySource;

@SpringBootApplication
@EnableConfigurationProperties({MailProperties.class, DatabaseProperties.class})
@PropertySource("classpath:custom.properties")
public class PropsDemoApp implements CommandLineRunner {

    private final ValueDemo valueDemo;
    private final MailProperties mailProperties;
    private final DatabaseProperties dbProperties;
    private final CustomApiService apiService;

    public PropsDemoApp(ValueDemo valueDemo, MailProperties mailProperties,
                        DatabaseProperties dbProperties, CustomApiService apiService) {
        this.valueDemo = valueDemo;
        this.mailProperties = mailProperties;
        this.dbProperties = dbProperties;
        this.apiService = apiService;
    }

    public static void main(String[] args) {
        SpringApplication.run(PropsDemoApp.class, args);
    }

    @Override
    public void run(String... args) {
        System.out.println("========================================");
        System.out.println("  EXTERNALIZED CONFIGURATION DEMO");
        System.out.println("========================================\n");

        valueDemo.printValues();
        printConfigProperties();
        apiService.printConfig();
        printOverrideDemo(args);
    }

    private void printConfigProperties() {
        System.out.println("─── 2. @ConfigurationProperties (Type-Safe) ────");
        System.out.println();
        System.out.println("  MailProperties (prefix = \"app.mail\"):");
        System.out.println("    host       = " + mailProperties.getHost());
        System.out.println("    port       = " + mailProperties.getPort());
        System.out.println("    username   = " + mailProperties.getUsername());
        System.out.println("    fromName   = " + mailProperties.getFromName());
        System.out.println("    sslEnabled = " + mailProperties.isSslEnabled());
        System.out.println("    poolSize   = " + mailProperties.getPoolSize());
        System.out.println("    timeout    = " + mailProperties.getTimeoutSeconds() + "s");
        System.out.println();
        System.out.println("  DatabaseProperties (prefix = \"app.database\"):");
        System.out.println("    url            = " + dbProperties.getUrl());
        System.out.println("    username       = " + dbProperties.getUsername());
        System.out.println("    password       = " + mask(dbProperties.getPassword()));
        System.out.println("    maxConnections = " + dbProperties.getMaxConnections());
        System.out.println("    driverClass    = " + dbProperties.getDriverClass());
        System.out.println();
        System.out.println("  → Relaxed binding: app.from-name → fromName (kebab → camel)");
        System.out.println("  → Type conversion: \"587\" → int, \"true\" → boolean automatic");
        System.out.println();
    }

    private void printOverrideDemo(String[] args) {
        System.out.println("─── 4. PROPERTY OVERRIDE PRIORITY ──────────────");
        System.out.println();
        System.out.println("  Priority (highest wins):");
        System.out.println("    1. Command-line args:  --app.name=\"CLI Value\"");
        System.out.println("    2. Environment vars:   APP_NAME=\"Env Value\"");
        System.out.println("    3. application-{profile}.properties");
        System.out.println("    4. application.properties");
        System.out.println("    5. @PropertySource files");
        System.out.println();
        System.out.println("  Try running:");
        System.out.println("    java -jar target/spring-props-demo-1.0.0.jar --app.name=\"Overridden!\"");
        System.out.println("    APP_NAME=\"From Env\" java -jar target/spring-props-demo-1.0.0.jar");
        System.out.println();
    }

    private String mask(String value) {
        if (value == null || value.length() <= 3) return "***";
        return value.substring(0, 3) + "***";
    }
}
