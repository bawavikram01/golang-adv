package com.learn.autoconfig;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import org.springframework.boot.autoconfigure.condition.ConditionEvaluationReport;
import org.springframework.boot.CommandLineRunner;
import org.springframework.context.ApplicationContext;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.DispatcherServlet;

import java.util.Arrays;
import java.util.Map;
import java.util.TreeMap;

/**
 * Explores what Spring Boot auto-configured and how @Conditional works.
 */
@Component
public class AutoConfigInspector implements CommandLineRunner {

    private final ApplicationContext context;
    private final ConditionEvaluationReport conditionReport;

    public AutoConfigInspector(ApplicationContext context, ConditionEvaluationReport conditionReport) {
        this.context = context;
        this.conditionReport = conditionReport;
    }

    @Override
    public void run(String... args) {
        System.out.println();
        System.out.println("========================================");
        System.out.println("  AUTO-CONFIGURATION DEEP DIVE");
        System.out.println("========================================\n");

        showAutoConfiguredBeans();
        showConditionReport();
        showOverrideDemo();
        showProjectInfo();
    }

    private void showAutoConfiguredBeans() {
        System.out.println("─── 1. KEY AUTO-CONFIGURED BEANS ───────────────");
        System.out.println();

        // These beans exist because spring-boot-starter-web is on classpath
        Map<String, String> autoConfigured = new TreeMap<>();
        autoConfigured.put("ObjectMapper", getBeanInfo(ObjectMapper.class));
        autoConfigured.put("DispatcherServlet", getBeanInfo(DispatcherServlet.class));

        // Check what Jackson features are enabled by default
        ObjectMapper mapper = context.getBean(ObjectMapper.class);
        boolean indentOutput = mapper.isEnabled(SerializationFeature.INDENT_OUTPUT);

        autoConfigured.forEach((name, info) ->
            System.out.println("  ✓ " + name + " → " + info));

        System.out.println();
        System.out.println("  ObjectMapper settings (auto-configured defaults):");
        System.out.println("    INDENT_OUTPUT = " + indentOutput + " (default: false, we'll override this)");
        System.out.println();
    }

    private void showConditionReport() {
        System.out.println("─── 2. CONDITION EVALUATION REPORT ─────────────");
        System.out.println();
        System.out.println("  POSITIVE MATCHES (auto-configs that WERE applied):");

        int count = 0;
        for (var entry : conditionReport.getConditionAndOutcomesBySource().entrySet()) {
            if (entry.getValue().isFullMatch() && entry.getKey().contains("AutoConfiguration")) {
                String name = entry.getKey();
                // Shorten class name
                String shortName = name.substring(name.lastIndexOf('.') + 1);
                if (count < 10) {
                    System.out.println("    ✓ " + shortName);
                }
                count++;
            }
        }
        System.out.println("    ... and " + (count - 10) + " more");
        System.out.println();

        // Show some negative matches
        System.out.println("  NEGATIVE MATCHES (auto-configs that were SKIPPED):");
        int negCount = 0;
        for (var entry : conditionReport.getConditionAndOutcomesBySource().entrySet()) {
            if (!entry.getValue().isFullMatch() && entry.getKey().contains("AutoConfiguration")) {
                String name = entry.getKey();
                String shortName = name.substring(name.lastIndexOf('.') + 1);
                if (negCount < 5) {
                    // Get the reason
                    var outcomes = entry.getValue();
                    String reason = outcomes.iterator().next().getOutcome().getMessage();
                    System.out.println("    ✗ " + shortName);
                    System.out.println("      Reason: " + truncate(reason, 70));
                }
                negCount++;
            }
        }
        System.out.println("    ... and " + (negCount - 5) + " more skipped");
        System.out.println();
        System.out.println("  → Use --debug flag or /actuator/conditions for the full report");
        System.out.println();
    }

    private void showOverrideDemo() {
        System.out.println("─── 3. OVERRIDING AUTO-CONFIGURATION ───────────");
        System.out.println();
        System.out.println("  Our CustomJacksonConfig defines a CUSTOM ObjectMapper:");

        ObjectMapper mapper = context.getBean(ObjectMapper.class);
        boolean indent = mapper.isEnabled(SerializationFeature.INDENT_OUTPUT);

        System.out.println("    INDENT_OUTPUT = " + indent);
        System.out.println();
        if (indent) {
            System.out.println("  → OUR @Bean won! Spring Boot's auto-configured ObjectMapper was SKIPPED");
            System.out.println("  → @ConditionalOnMissingBean saw we defined it → backed off");
        } else {
            System.out.println("  → Auto-configured ObjectMapper is still active (no override)");
        }
        System.out.println();
    }

    private void showProjectInfo() {
        System.out.println("─── 4. ENDPOINTS TO EXPLORE ────────────────────");
        System.out.println();
        System.out.println("  http://localhost:8081/api/demo          → uses our custom ObjectMapper");
        System.out.println("  http://localhost:8081/actuator/conditions → full condition report (JSON)");
        System.out.println("  http://localhost:8081/actuator/beans     → all beans in container");
        System.out.println("  http://localhost:8081/actuator/env       → all config properties");
        System.out.println();
        System.out.println("  Total beans: " + context.getBeanDefinitionCount());
        System.out.println();
        System.out.println("  Press Ctrl+C to stop.");
        System.out.println("========================================\n");
    }

    private String getBeanInfo(Class<?> type) {
        try {
            Object bean = context.getBean(type);
            return bean.getClass().getSimpleName() + " @ " + Integer.toHexString(bean.hashCode());
        } catch (Exception e) {
            return "NOT FOUND";
        }
    }

    private String truncate(String s, int max) {
        return s.length() <= max ? s : s.substring(0, max) + "...";
    }
}
