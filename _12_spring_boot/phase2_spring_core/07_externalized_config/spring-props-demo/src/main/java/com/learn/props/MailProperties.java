package com.learn.props;

import jakarta.validation.constraints.Max;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.validation.annotation.Validated;

/**
 * Type-safe configuration binding.
 *
 * All properties under "app.mail.*" are automatically mapped to fields:
 *   app.mail.host          → this.host
 *   app.mail.port          → this.port
 *   app.mail.ssl-enabled   → this.sslEnabled  (relaxed binding!)
 *   app.mail.from-name     → this.fromName    (kebab-case → camelCase)
 *
 * @Validated enables Jakarta Bean Validation on startup.
 * If validation fails, the application refuses to start (fail-fast).
 */
@ConfigurationProperties(prefix = "app.mail")
@Validated
public class MailProperties {

    @NotBlank(message = "Mail host must be configured")
    private String host;

    @Min(1) @Max(65535)
    private int port;

    private String username;
    private String fromName;
    private boolean sslEnabled;
    private int poolSize;
    private int timeoutSeconds;

    // ─── Getters and Setters (required for binding) ───────────

    public String getHost() { return host; }
    public void setHost(String host) { this.host = host; }

    public int getPort() { return port; }
    public void setPort(int port) { this.port = port; }

    public String getUsername() { return username; }
    public void setUsername(String username) { this.username = username; }

    public String getFromName() { return fromName; }
    public void setFromName(String fromName) { this.fromName = fromName; }

    public boolean isSslEnabled() { return sslEnabled; }
    public void setSslEnabled(boolean sslEnabled) { this.sslEnabled = sslEnabled; }

    public int getPoolSize() { return poolSize; }
    public void setPoolSize(int poolSize) { this.poolSize = poolSize; }

    public int getTimeoutSeconds() { return timeoutSeconds; }
    public void setTimeoutSeconds(int timeoutSeconds) { this.timeoutSeconds = timeoutSeconds; }
}
