package com.learn.rest.controller;

import org.springframework.web.bind.annotation.*;

import java.util.Map;

/**
 * DEMO 2: Various Parameter Binding Examples
 * 
 * Shows different ways to extract data from HTTP requests.
 */
@RestController
@RequestMapping("/api/demo")
public class DemoController {

    // ─── PATH VARIABLES ────────────────────────────────────────────
    // GET /api/demo/greet/Vikram
    @GetMapping("/greet/{name}")
    public Map<String, String> greet(@PathVariable String name) {
        return Map.of("message", "Hello, " + name + "!");
    }

    // Multiple path variables
    // GET /api/demo/math/10/plus/5
    @GetMapping("/math/{a}/plus/{b}")
    public Map<String, Object> add(@PathVariable int a, @PathVariable int b) {
        return Map.of("a", a, "b", b, "sum", a + b);
    }

    // ─── QUERY PARAMETERS ──────────────────────────────────────────
    // GET /api/demo/search?q=spring&page=1&size=5
    @GetMapping("/search")
    public Map<String, Object> search(
            @RequestParam String q,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "10") int size) {
        return Map.of(
            "query", q,
            "page", page,
            "size", size,
            "info", "Would search for '" + q + "' on page " + page
        );
    }

    // ─── REQUEST HEADERS ───────────────────────────────────────────
    // GET /api/demo/headers
    @GetMapping("/headers")
    public Map<String, String> readHeaders(
            @RequestHeader("User-Agent") String userAgent,
            @RequestHeader(value = "X-Custom-Header", required = false) String custom) {
        return Map.of(
            "userAgent", userAgent,
            "customHeader", custom != null ? custom : "not provided"
        );
    }

    // ─── REQUEST BODY (POST with JSON) ─────────────────────────────
    // POST /api/demo/echo  (body: any JSON)
    @PostMapping("/echo")
    public Map<String, Object> echo(@RequestBody Map<String, Object> body) {
        return Map.of(
            "received", body,
            "message", "I echoed back whatever you sent me!"
        );
    }

    // ─── MULTIPLE METHODS ON SAME PATH ─────────────────────────────
    @GetMapping("/resource")
    public Map<String, String> getResource() {
        return Map.of("action", "READ", "info", "This was a GET request");
    }

    @PostMapping("/resource")
    public Map<String, String> postResource(@RequestBody Map<String, String> body) {
        return Map.of("action", "CREATE", "received", body.toString());
    }

    @PutMapping("/resource")
    public Map<String, String> putResource(@RequestBody Map<String, String> body) {
        return Map.of("action", "UPDATE", "received", body.toString());
    }

    @DeleteMapping("/resource")
    public Map<String, String> deleteResource() {
        return Map.of("action", "DELETE", "info", "Resource deleted");
    }
}
