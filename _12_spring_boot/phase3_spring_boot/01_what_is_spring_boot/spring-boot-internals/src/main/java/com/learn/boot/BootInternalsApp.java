package com.learn.boot;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * @SpringBootApplication = three annotations in one:
 *   1. @Configuration       → this class can define @Bean methods
 *   2. @EnableAutoConfiguration → turn on auto-config magic
 *   3. @ComponentScan       → scan this package + sub-packages
 *
 * SpringApplication.run() does:
 *   - Creates ApplicationContext
 *   - Runs auto-configuration
 *   - Starts embedded Tomcat
 *   - Registers all beans
 *   - Application is READY
 */
@SpringBootApplication
public class BootInternalsApp {

    public static void main(String[] args) {
        SpringApplication.run(BootInternalsApp.class, args);
    }
}
