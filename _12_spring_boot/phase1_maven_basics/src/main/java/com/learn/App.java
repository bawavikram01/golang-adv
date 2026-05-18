package com.learn;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import java.util.List;

/**
 * PHASE 1.2 — MAVEN BASICS
 *
 * This class uses Gson (a Google library).
 * We didn't download any JAR file manually.
 * We just listed it in pom.xml → Maven downloaded it for us.
 *
 * In Spring Boot, you'll have 20+ dependencies — imagine managing those by hand!
 */
public class App {
    public static void main(String[] args) {

        System.out.println("=== MAVEN PROJECT RUNNING! ===\n");

        // Gson is available because we declared it in pom.xml
        // Maven downloaded it automatically!
        Gson gson = new GsonBuilder().setPrettyPrinting().create();

        // ---- Java Object → JSON (Serialization) ----
        System.out.println("--- Java Object → JSON ---");
        User user = new User("Alice", "alice@example.com", 28);
        String json = gson.toJson(user);
        System.out.println(json);

        // ---- JSON → Java Object (Deserialization) ----
        System.out.println("\n--- JSON → Java Object ---");
        String inputJson = """
                {
                  "name": "Bob",
                  "email": "bob@example.com",
                  "age": 35
                }
                """;
        User fromJson = gson.fromJson(inputJson, User.class);
        System.out.println(fromJson);

        // ---- List of Objects → JSON ----
        System.out.println("\n--- List → JSON ---");
        List<User> users = List.of(
            new User("Charlie", "charlie@example.com", 22),
            new User("Diana", "diana@example.com", 30)
        );
        System.out.println(gson.toJson(users));

        System.out.println("\n=== KEY POINT ===");
        System.out.println("Gson was NOT downloaded manually.");
        System.out.println("Maven read pom.xml and fetched it from Maven Central.");
        System.out.println("In Spring Boot, Jackson does the same (JSON <-> Java),");
        System.out.println("and it's auto-configured — you don't even need this setup!");
    }
}
