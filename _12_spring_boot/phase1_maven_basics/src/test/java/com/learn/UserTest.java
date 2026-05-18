package com.learn;

import com.google.gson.Gson;
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

/**
 * JUnit 5 test — Maven runs these automatically during "mvn test".
 * 
 * Notice: JUnit has scope=test in pom.xml, so it's ONLY available here,
 * not in src/main/java. Maven enforces this separation.
 */
class UserTest {

    private final Gson gson = new Gson();

    @Test
    void testSerializeUser() {
        User user = new User("Alice", "alice@example.com", 28);
        String json = gson.toJson(user);

        assertTrue(json.contains("Alice"));
        assertTrue(json.contains("alice@example.com"));
        assertTrue(json.contains("28"));
        System.out.println("✅ Serialization test passed: " + json);
    }

    @Test
    void testDeserializeUser() {
        String json = "{\"name\":\"Bob\",\"email\":\"bob@example.com\",\"age\":35}";
        User user = gson.fromJson(json, User.class);

        assertEquals("Bob", user.getName());
        assertEquals("bob@example.com", user.getEmail());
        assertEquals(35, user.getAge());
        System.out.println("✅ Deserialization test passed: " + user);
    }

    @Test
    void testRoundTrip() {
        User original = new User("Charlie", "charlie@example.com", 22);

        // Object → JSON → Object
        String json = gson.toJson(original);
        User restored = gson.fromJson(json, User.class);

        assertEquals(original.getName(), restored.getName());
        assertEquals(original.getEmail(), restored.getEmail());
        assertEquals(original.getAge(), restored.getAge());
        System.out.println("✅ Round-trip test passed: " + original + " == " + restored);
    }
}
