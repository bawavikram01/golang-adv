package com.learn.lifecycle;

import org.springframework.boot.CommandLineRunner;
import org.springframework.context.ApplicationContext;
import org.springframework.stereotype.Component;

/**
 * CommandLineRunner runs AFTER all beans are fully initialized
 * and the ApplicationContext is refreshed.
 *
 * It's perfect for:
 * - Running startup tasks
 * - Printing diagnostics
 * - Testing bean configurations
 */
@Component
public class ContainerExplorer implements CommandLineRunner {

    private final ApplicationContext context;

    public ContainerExplorer(ApplicationContext context) {
        this.context = context;
    }

    @Override
    public void run(String... args) {
        System.out.println("╔══════════════════════════════════════════════╗");
        System.out.println("║   CONTAINER EXPLORATION (CommandLineRunner)  ║");
        System.out.println("╠══════════════════════════════════════════════╣");

        // 1. List all OUR beans (filter out Spring internal beans)
        System.out.println("║");
        System.out.println("║  Our beans in the container:");
        String[] names = context.getBeanDefinitionNames();
        for (String name : names) {
            // Only show our package's beans
            Object bean = context.getBean(name);
            if (bean.getClass().getPackageName().startsWith("com.learn")) {
                System.out.printf("║    • %-30s → %s%n", name, bean.getClass().getSimpleName());
            }
        }

        // 2. Check bean existence
        System.out.println("║");
        System.out.println("║  containsBean(\"fullLifecycleBean\"): " + context.containsBean("fullLifecycleBean"));
        System.out.println("║  containsBean(\"nonExistentBean\"): " + context.containsBean("nonExistentBean"));

        // 3. Environment info
        String[] profiles = context.getEnvironment().getActiveProfiles();
        System.out.println("║");
        System.out.println("║  Active profiles: " + (profiles.length == 0 ? "(none — using default)" : String.join(", ", profiles)));
        System.out.println("║");
        System.out.println("╚══════════════════════════════════════════════╝");
    }
}
