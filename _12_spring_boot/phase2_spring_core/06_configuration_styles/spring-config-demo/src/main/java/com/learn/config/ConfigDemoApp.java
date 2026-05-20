package com.learn.config;

import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;

@SpringBootApplication
public class ConfigDemoApp implements CommandLineRunner {

    private final ConfigurableApplicationContext context;

    public ConfigDemoApp(ConfigurableApplicationContext context) {
        this.context = context;
    }

    public static void main(String[] args) {
        SpringApplication.run(ConfigDemoApp.class, args);
    }

    @Override
    public void run(String... args) {
        System.out.println("========================================");
        System.out.println("  CONFIGURATION STYLES DEMO");
        System.out.println("========================================\n");

        demoComponentScanning();
        demoBeanConfig();
        demoConfigurationProxy();
        demoMultipleConfigs();
    }

    private void demoComponentScanning() {
        System.out.println("─── 1. COMPONENT SCANNING ──────────────────────");
        System.out.println();
        System.out.println("  Beans auto-detected via @Component/@Service:");

        // These were found by @ComponentScan automatically
        GreetingService greeting = context.getBean(GreetingService.class);
        System.out.println("    GreetingService.greet(\"World\") = " + greeting.greet("World"));

        UserRepository repo = context.getBean(UserRepository.class);
        System.out.println("    UserRepository.count() = " + repo.count());
        System.out.println();
        System.out.println("  → @Component, @Service, @Repository auto-registered");
        System.out.println();
    }

    private void demoBeanConfig() {
        System.out.println("─── 2. @BEAN IN @CONFIGURATION ─────────────────");
        System.out.println();

        // These were defined in AppConfig.java using @Bean
        PaymentGateway gateway = context.getBean(PaymentGateway.class);
        System.out.println("  PaymentGateway (third-party simulation):");
        System.out.println("    gateway.charge(49.99) = " + gateway.charge(49.99));
        System.out.println();

        // Bean with custom name
        String appName = context.getBean("applicationName", String.class);
        System.out.println("  String bean \"applicationName\" = " + appName);
        System.out.println();

        // Bean with dependencies injected via method params
        OrderProcessor processor = context.getBean(OrderProcessor.class);
        System.out.println("  OrderProcessor (depends on PaymentGateway):");
        System.out.println("    processor.process(\"ORD-123\", 99.99) = " + processor.process("ORD-123", 99.99));
        System.out.println();
        System.out.println("  → @Bean for third-party classes & complex setup");
        System.out.println();
    }

    private void demoConfigurationProxy() {
        System.out.println("─── 3. @CONFIGURATION PROXY BEHAVIOR ───────────");
        System.out.println();

        // Demonstrate that @Configuration class is CGLIB-proxied
        AppConfig config = context.getBean(AppConfig.class);
        System.out.println("  AppConfig class: " + config.getClass().getName());
        System.out.println("  → Notice: $$SpringCGLIB$$ = it's a proxy!");
        System.out.println();

        // The proxy ensures @Bean methods return singletons
        System.out.println("  Singleton guarantee via proxy:");
        PaymentGateway g1 = context.getBean(PaymentGateway.class);
        PaymentGateway g2 = context.getBean(PaymentGateway.class);
        System.out.println("    getBean(PaymentGateway) twice: same? " + (g1 == g2));
        System.out.println("  → @Configuration proxy ensures one instance");
        System.out.println();
    }

    private void demoMultipleConfigs() {
        System.out.println("─── 4. MULTIPLE @CONFIGURATION CLASSES ─────────");
        System.out.println();

        // Beans from separate config classes
        NotificationService notifier = context.getBean(NotificationService.class);
        System.out.println("  From InfraConfig.java:");
        System.out.println("    NotificationService.notify(\"Hello\") = " + notifier.notify("Hello"));
        System.out.println();

        // Show all beans from our package
        System.out.println("  All our beans in the container:");
        String[] names = context.getBeanDefinitionNames();
        for (String name : names) {
            Object bean = context.getBean(name);
            if (bean.getClass().getPackageName().startsWith("com.learn")) {
                System.out.printf("    • %-25s [%s]%n", name,
                        bean.getClass().getSimpleName().contains("CGLIB")
                                ? bean.getClass().getSuperclass().getSimpleName() + " (proxied)"
                                : bean.getClass().getSimpleName());
            }
        }
        System.out.println();
    }
}
