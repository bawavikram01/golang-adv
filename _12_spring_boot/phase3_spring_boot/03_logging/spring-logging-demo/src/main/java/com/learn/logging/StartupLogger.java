package com.learn.logging;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

/**
 * Demonstrates all log levels on startup.
 */
@Component
public class StartupLogger implements CommandLineRunner {

    private static final Logger log = LoggerFactory.getLogger(StartupLogger.class);

    @Override
    public void run(String... args) {
        System.out.println();
        System.out.println("========================================");
        System.out.println("  LOGGING DEMO — All 5 levels");
        System.out.println("========================================");
        System.out.println("  Config: logging.level.com.learn.logging=DEBUG");
        System.out.println("  (TRACE is hidden, DEBUG+ is visible)");
        System.out.println("========================================");
        System.out.println();

        log.trace("This is TRACE — you WON'T see this (below configured level)");
        log.debug("This is DEBUG — development diagnostics");
        log.info("This is INFO — important business events");
        log.warn("This is WARN — something unexpected but recoverable");
        log.error("This is ERROR — something broke!");

        System.out.println();
        System.out.println("  ─── ENDPOINTS ────────────────────────────────");
        System.out.println("  POST http://localhost:8082/api/orders?orderId=ORD-001&customer=Alice&amount=79.99");
        System.out.println("  POST http://localhost:8082/api/orders?orderId=ORD-002&customer=Bob&amount=1500");
        System.out.println("  DELETE http://localhost:8082/api/orders/ORD-001");
        System.out.println("  DELETE http://localhost:8082/api/orders/ORD-999  (will error)");
        System.out.println("  GET  http://localhost:8082/api/orders/log-levels");
        System.out.println();
        System.out.println("  Log file: logs/app.log (auto-rotated at 5MB)");
        System.out.println("========================================");
        System.out.println();
    }
}
