package com.learn.autoconfig;

import com.learn.autoconfig.ConditionalBeansConfig.GreetingService;
import com.learn.autoconfig.ConditionalBeansConfig.PremiumService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.time.LocalDateTime;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Optional;

/**
 * REST controller demonstrating:
 * 1. Custom ObjectMapper is active (pretty-printed JSON)
 * 2. Conditional beans (present or absent based on properties)
 */
@RestController
@RequestMapping("/api")
public class DemoController {

    private final GreetingService greetingService;

    @Autowired(required = false)  // May not exist if condition fails!
    private PremiumService premiumService;

    public DemoController(GreetingService greetingService) {
        this.greetingService = greetingService;
    }

    @GetMapping("/demo")
    public Map<String, Object> demo() {
        Map<String, Object> response = new LinkedHashMap<>();
        response.put("message", "Auto-Configuration Demo");
        response.put("timestamp", LocalDateTime.now().toString());
        response.put("greeting", greetingService.greet("Spring Boot"));
        response.put("greetingNote", greetingService.getNote());
        response.put("premiumAvailable", premiumService != null);
        response.put("premiumFeature", premiumService != null ? premiumService.getFeature() : "DISABLED (app.feature.premium != true)");
        response.put("jsonNote", "Notice: this JSON is PRETTY-PRINTED because we overrode ObjectMapper!");
        return response;
    }

    @GetMapping("/conditional-info")
    public Map<String, Object> conditionalInfo() {
        Map<String, Object> response = new LinkedHashMap<>();
        response.put("concept", "@ConditionalOnProperty");
        response.put("greetingServiceActive", true);
        response.put("greetingServiceReason", "app.feature.greeting not set → matchIfMissing=true → bean created");
        response.put("premiumServiceActive", premiumService != null);
        response.put("premiumServiceReason", premiumService != null
            ? "app.feature.premium=true → bean created"
            : "app.feature.premium not set → matchIfMissing=false (default) → bean SKIPPED");
        response.put("howToEnable", "Add app.feature.premium=true to application.properties");
        return response;
    }
}
