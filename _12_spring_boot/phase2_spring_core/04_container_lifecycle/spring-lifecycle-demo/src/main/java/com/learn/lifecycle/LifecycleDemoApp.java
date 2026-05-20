package com.learn.lifecycle;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;

@SpringBootApplication
public class LifecycleDemoApp {

    public static void main(String[] args) {
        System.out.println("========================================");
        System.out.println("  SPRING BEAN LIFECYCLE DEMO");
        System.out.println("========================================\n");

        // SpringApplication.run() creates the ApplicationContext
        // which triggers bean instantiation + full lifecycle
        ConfigurableApplicationContext context = SpringApplication.run(LifecycleDemoApp.class, args);

        System.out.println("\n--- Application is RUNNING ---");
        System.out.println("Total beans in container: " + context.getBeanDefinitionCount());
        System.out.println();

        // Fetch our demo bean — it's already fully initialized
        FullLifecycleBean bean = context.getBean(FullLifecycleBean.class);
        bean.doWork();

        System.out.println("\n--- Shutting down (triggers destruction phase) ---\n");
        context.close(); // This triggers @PreDestroy + DisposableBean.destroy()
    }
}
