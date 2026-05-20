package com.learn.springdemo;

import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ApplicationContext;
import org.springframework.context.annotation.Bean;

/**
 * THE MAIN CLASS — Entry point of a Spring Boot application.
 *
 * @SpringBootApplication = Three annotations in one:
 *   @Configuration     → "This class can define @Bean methods"
 *   @EnableAutoConfiguration → "Spring Boot, auto-configure based on dependencies"
 *   @ComponentScan     → "Scan this package and sub-packages for @Component classes"
 */
@SpringBootApplication
public class SpringDemoApplication {

    public static void main(String[] args) {
        System.out.println("\n╔═══════════════════════════════════════════════════╗");
        System.out.println("║   YOUR FIRST SPRING BOOT APPLICATION             ║");
        System.out.println("║   Spring does everything our mini-container did   ║");
        System.out.println("╚═══════════════════════════════════════════════════╝\n");

        System.out.println("--- Spring Container Starting (same steps as our mini-Spring) ---\n");

        // SpringApplication.run() = Creates the ApplicationContext
        // This single line:
        //   1. Scans for @Component classes
        //   2. Creates all beans
        //   3. Resolves all dependencies
        //   4. Injects everything
        //   5. Starts the application
        ApplicationContext context = SpringApplication.run(SpringDemoApplication.class, args);

        System.out.println("\n--- Spring Container Ready! ---\n");

        // Show all our custom beans
        System.out.println("=== BEANS IN THE CONTAINER ===");
        System.out.println("  AppConfig:           " + context.getBean(AppConfig.class));
        System.out.println("  UserRepository:      " + context.getBean(UserRepository.class));
        System.out.println("  NotificationService: " + context.getBean(NotificationService.class));
        System.out.println("  UserService:         " + context.getBean(UserService.class));

        // Prove singletons
        System.out.println("\n=== SINGLETON PROOF ===");
        UserService us1 = context.getBean(UserService.class);
        UserService us2 = context.getBean(UserService.class);
        System.out.println("  Same UserService instance? " + (us1 == us2));

        // Use the service
        System.out.println("\n=== USING THE APPLICATION ===\n");
        UserService userService = context.getBean(UserService.class);
        userService.registerUser("Alice");
        System.out.println();
        userService.registerUser("Bob");

        System.out.println("\n=== COMPARE WITH OUR MINI-SPRING ===");
        System.out.println("  Mini-Spring:  container.register(...) → container.refresh() → container.getBean(...)");
        System.out.println("  Real Spring:  @Component annotations  → SpringApplication.run() → context.getBean(...)");
        System.out.println("  SAME CONCEPT. Real Spring just has auto-scanning + 1000 more features.");

        System.out.println("\n=== TOTAL SPRING-MANAGED BEANS ===");
        String[] allBeans = context.getBeanDefinitionNames();
        System.out.println("  Spring created " + allBeans.length + " beans total");
        System.out.println("  (Most are internal Spring infrastructure beans)\n");
        System.out.println("  Our custom beans:");
        for (String name : allBeans) {
            if (name.startsWith("app") || name.startsWith("user") || name.startsWith("notification")) {
                System.out.println("    • " + name);
            }
        }
    }
}
