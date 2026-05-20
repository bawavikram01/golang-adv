package com.learn.props;

import jakarta.validation.constraints.NotBlank;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.validation.annotation.Validated;

/**
 * Another @ConfigurationProperties example.
 * Maps all "app.database.*" properties.
 *
 * Demonstrates relaxed binding:
 *   app.database.max-connections → maxConnections
 *   app.database.driver-class    → driverClass
 */
@ConfigurationProperties(prefix = "app.database")
@Validated
public class DatabaseProperties {

    @NotBlank
    private String url;

    @NotBlank
    private String username;

    private String password;
    private int maxConnections;
    private String driverClass;

    // ─── Getters and Setters ──────────────────────────────────

    public String getUrl() { return url; }
    public void setUrl(String url) { this.url = url; }

    public String getUsername() { return username; }
    public void setUsername(String username) { this.username = username; }

    public String getPassword() { return password; }
    public void setPassword(String password) { this.password = password; }

    public int getMaxConnections() { return maxConnections; }
    public void setMaxConnections(int maxConnections) { this.maxConnections = maxConnections; }

    public String getDriverClass() { return driverClass; }
    public void setDriverClass(String driverClass) { this.driverClass = driverClass; }
}
