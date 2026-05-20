package com.learn.di;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;

@SpringBootApplication
public class DiDemoApplication {

    public static void main(String[] args) {
        System.out.println("\n╔═══════════════════════════════════════════════════════╗");
        System.out.println("║   PHASE 2.3 — DEPENDENCY INJECTION (3 Methods)       ║");
        System.out.println("╚═══════════════════════════════════════════════════════╝\n");

        System.out.println("=== BEAN CREATION (watch the injection order) ===\n");
        ConfigurableApplicationContext ctx = SpringApplication.run(DiDemoApplication.class, args);

        // ────────────────────────────────────────────────────────────
        System.out.println("\n\n═══════════════════════════════════════════════════════");
        System.out.println("  ✅ METHOD 1: CONSTRUCTOR INJECTION (Recommended)");
        System.out.println("═══════════════════════════════════════════════════════");
        System.out.println("  • Fields are final (immutable)");
        System.out.println("  • All deps required — app won't start without them");
        System.out.println("  • Gets @Primary bean (EmailSender) by default");

        ConstructorService constructorService = ctx.getBean(ConstructorService.class);
        constructorService.registerUser("Alice", "alice@mail.com");


        // ────────────────────────────────────────────────────────────
        System.out.println("\n\n═══════════════════════════════════════════════════════");
        System.out.println("  ⚠️ METHOD 2: SETTER INJECTION (Optional deps)");
        System.out.println("═══════════════════════════════════════════════════════");
        System.out.println("  • @Qualifier('smsSender') overrides @Primary");
        System.out.println("  • Field not final — set after construction");
        System.out.println("  • Useful for optional dependencies");

        SetterService setterService = ctx.getBean(SetterService.class);
        setterService.notifyUser("Bob", "+1234567890");


        // ────────────────────────────────────────────────────────────
        System.out.println("\n\n═══════════════════════════════════════════════════════");
        System.out.println("  ❌ METHOD 3: FIELD INJECTION (Avoid!)");
        System.out.println("═══════════════════════════════════════════════════════");
        System.out.println("  • @Qualifier('pushSender') picks PushSender");
        System.out.println("  • Fields set via reflection AFTER construction");
        System.out.println("  • Can't test without Spring — DON'T use in prod");

        FieldService fieldService = ctx.getBean(FieldService.class);
        fieldService.alertUser("Charlie", "device-token-xyz");


        // ────────────────────────────────────────────────────────────
        System.out.println("\n\n═══════════════════════════════════════════════════════");
        System.out.println("  🌟 BONUS: INJECT ALL IMPLEMENTATIONS (List<T>)");
        System.out.println("═══════════════════════════════════════════════════════");
        System.out.println("  • Injects ALL beans of type MessageSender");
        System.out.println("  • Useful for broadcast, strategy pattern, etc.");

        MultiSenderService multiService = ctx.getBean(MultiSenderService.class);
        multiService.broadcastToAll("admin@company.com", "System alert!");


        // ────────────────────────────────────────────────────────────
        System.out.println("\n\n═══════════════════════════════════════════════════════");
        System.out.println("  📋 SUMMARY: Which did each service get?");
        System.out.println("═══════════════════════════════════════════════════════\n");
        System.out.println("  ConstructorService → EmailSender  (via @Primary, default)");
        System.out.println("  SetterService      → SmsSender    (via @Qualifier(\"smsSender\"))");
        System.out.println("  FieldService       → PushSender   (via @Qualifier(\"pushSender\"))");
        System.out.println("  MultiSenderService → ALL 3        (via List<MessageSender>)");
        System.out.println();
        System.out.println("  Same interface, different injections. Zero code change in impls.");
        System.out.println("  RULE: Always prefer constructor injection. Use @Qualifier to pick.");

        ctx.close();
    }
}
