package com.learn.aop;

import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class AopDemoApp implements CommandLineRunner {

    private final OrderService orderService;
    private final UserService userService;

    public AopDemoApp(OrderService orderService, UserService userService) {
        this.orderService = orderService;
        this.userService = userService;
    }

    public static void main(String[] args) {
        SpringApplication.run(AopDemoApp.class, args);
    }

    @Override
    public void run(String... args) {
        System.out.println("========================================");
        System.out.println("  SPRING AOP DEMO");
        System.out.println("========================================\n");

        demoAllAdviceTypes();
        demoTimingAspect();
        demoExceptionHandling();
        demoCustomAnnotation();
        demoProxyEvidence();
    }

    private void demoAllAdviceTypes() {
        System.out.println("─── 1. ALL ADVICE TYPES (@Before, @After, @AfterReturning) ──");
        System.out.println();
        System.out.println("  Calling: orderService.placeOrder(\"ORD-100\", 79.99)");
        System.out.println();

        String result = orderService.placeOrder("ORD-100", 79.99);

        System.out.println();
        System.out.println("  Final result: " + result);
        System.out.println();
        System.out.println("  Execution order:");
        System.out.println("    LoggingAspect @Before → TimingAspect @Around (start)");
        System.out.println("    → real method → TimingAspect @Around (end)");
        System.out.println("    → @AfterReturning → @After");
        System.out.println();
    }

    private void demoTimingAspect() {
        System.out.println("─── 2. @AROUND TIMING (only @Timed methods) ────");
        System.out.println();

        System.out.println("  Calling: orderService.cancelOrder(\"ORD-100\")  [@Timed]");
        System.out.println();
        orderService.cancelOrder("ORD-100");
        System.out.println();

        System.out.println("  Calling: orderService.getOrderStatus(\"ORD-100\")  [no @Timed]");
        System.out.println();
        orderService.getOrderStatus("ORD-100");
        System.out.println();
        System.out.println("  → Notice: no timing output for getOrderStatus (no @Timed)!");
        System.out.println();
    }

    private void demoExceptionHandling() {
        System.out.println("─── 3. @AfterThrowing (exception interception) ─");
        System.out.println();
        System.out.println("  Calling: userService.deleteUser(\"admin\") [throws exception]");
        System.out.println();

        try {
            userService.deleteUser("admin");
        } catch (RuntimeException e) {
            System.out.println("  Caught in caller: " + e.getMessage());
        }

        System.out.println();
        System.out.println("  → @AfterThrowing intercepted the exception");
        System.out.println("  → Exception still propagates to the caller");
        System.out.println();
    }

    private void demoCustomAnnotation() {
        System.out.println("─── 4. CUSTOM ANNOTATION @Audited ──────────────");
        System.out.println();
        System.out.println("  Calling: userService.createUser(\"Bob\", \"bob@mail.com\") [@Audited]");
        System.out.println();

        userService.createUser("Bob", "bob@mail.com");
        System.out.println();
        System.out.println("  → AuditAspect reads the @Audited(action=\"CREATE_USER\") value");
        System.out.println();
    }

    private void demoProxyEvidence() {
        System.out.println("─── 5. PROOF: BEANS ARE PROXIED ────────────────");
        System.out.println();
        System.out.println("  orderService class: " + orderService.getClass().getName());
        System.out.println("  userService class:  " + userService.getClass().getName());
        System.out.println();
        System.out.println("  → $$SpringCGLIB$$ = Spring wrapped them in proxies");
        System.out.println("  → The proxy intercepts calls and runs aspect advice");
        System.out.println();
    }
}
