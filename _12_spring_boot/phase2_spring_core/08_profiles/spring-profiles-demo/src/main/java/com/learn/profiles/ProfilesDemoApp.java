package com.learn.profiles;

import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.core.env.Environment;

import java.util.Arrays;

@SpringBootApplication
public class ProfilesDemoApp implements CommandLineRunner {

    private final Environment environment;
    private final NotificationService notificationService;
    private final DataSourceService dataSourceService;

    public ProfilesDemoApp(Environment environment,
                           NotificationService notificationService,
                           DataSourceService dataSourceService) {
        this.environment = environment;
        this.notificationService = notificationService;
        this.dataSourceService = dataSourceService;
    }

    public static void main(String[] args) {
        SpringApplication.run(ProfilesDemoApp.class, args);
    }

    @Override
    public void run(String... args) {
        System.out.println("========================================");
        System.out.println("  SPRING PROFILES DEMO");
        System.out.println("========================================\n");

        printActiveProfiles();
        printProfileProperties();
        printProfileBeans();
        printHowToSwitch();
    }

    private void printActiveProfiles() {
        System.out.println("─── 1. ACTIVE PROFILES ─────────────────────────");
        System.out.println();

        String[] active = environment.getActiveProfiles();
        String[] defaults = environment.getDefaultProfiles();

        if (active.length == 0) {
            System.out.println("  Active profiles:  (none explicitly set)");
            System.out.println("  Default profiles: " + Arrays.toString(defaults));
        } else {
            System.out.println("  Active profiles: " + Arrays.toString(active));
        }
        System.out.println();
    }

    private void printProfileProperties() {
        System.out.println("─── 2. PROFILE-SPECIFIC PROPERTIES ────────────");
        System.out.println();

        String appName = environment.getProperty("app.name");
        String dbType = environment.getProperty("app.database.type");
        String dbUrl = environment.getProperty("app.database.url");
        String cacheEnabled = environment.getProperty("app.cache.enabled");
        String logLevel = environment.getProperty("app.log-level");
        String mockServices = environment.getProperty("app.mock-external-services", "not set");

        System.out.println("  app.name           = " + appName + "  (from base)");
        System.out.println("  app.database.type  = " + dbType + "  (profile-specific)");
        System.out.println("  app.database.url   = " + dbUrl);
        System.out.println("  app.cache.enabled  = " + cacheEnabled);
        System.out.println("  app.log-level      = " + logLevel);
        System.out.println("  app.mock-services  = " + mockServices);
        System.out.println();
        System.out.println("  → Profile properties OVERRIDE base application.properties");
        System.out.println();
    }

    private void printProfileBeans() {
        System.out.println("─── 3. PROFILE-SPECIFIC BEANS (@Profile) ───────");
        System.out.println();

        System.out.println("  NotificationService implementation:");
        System.out.println("    Type: " + notificationService.getClass().getSimpleName());
        System.out.println("    Send: " + notificationService.send("Order confirmed!"));
        System.out.println();

        System.out.println("  DataSourceService implementation:");
        System.out.println("    Type: " + dataSourceService.getClass().getSimpleName());
        System.out.println("    Info: " + dataSourceService.getInfo());
        System.out.println();

        System.out.println("  → Different @Profile beans activate per environment");
        System.out.println();
    }

    private void printHowToSwitch() {
        System.out.println("─── 4. TRY SWITCHING PROFILES ──────────────────");
        System.out.println();
        System.out.println("  Run with dev profile:");
        System.out.println("    java -jar target/spring-profiles-demo-1.0.0.jar --spring.profiles.active=dev");
        System.out.println();
        System.out.println("  Run with prod profile:");
        System.out.println("    java -jar target/spring-profiles-demo-1.0.0.jar --spring.profiles.active=prod");
        System.out.println();
        System.out.println("  Run with NO profile (uses @Profile(\"default\") beans):");
        System.out.println("    java -jar target/spring-profiles-demo-1.0.0.jar");
        System.out.println();
    }
}
