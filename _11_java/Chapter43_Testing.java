/*
 * ============================================================
 *  CHAPTER 43: UNIT TESTING
 * ============================================================
 *  Testing = proving your code works correctly.
 *  Professional developers write tests for all their code.
 *
 *  FRAMEWORKS:
 *    JUnit 5 (Jupiter)  — the standard
 *    Mockito            — mocking framework
 *    AssertJ            — fluent assertions
 *
 *  TEST TYPES:
 *    Unit Test         — test one method/class in isolation
 *    Integration Test  — test multiple components together
 *    End-to-End Test   — test entire system
 *
 *  STRUCTURE (AAA Pattern):
 *    Arrange — set up test data
 *    Act     — call the method
 *    Assert  — verify the result
 *
 *  NOTE: JUnit requires adding junit-jupiter to your classpath.
 *  This file demonstrates the concepts with a manual test runner.
 *  In real projects, use Maven/Gradle with JUnit 5.
 * ============================================================
 */

import java.util.*;

public class Chapter43_Testing {

    // ========================================================
    // Code Under Test
    // ========================================================

    static class Calculator {
        int add(int a, int b) { return a + b; }
        int subtract(int a, int b) { return a - b; }
        int multiply(int a, int b) { return a * b; }

        int divide(int a, int b) {
            if (b == 0) throw new ArithmeticException("Cannot divide by zero");
            return a / b;
        }
    }

    static class StringUtils {
        static String reverse(String s) {
            if (s == null) throw new IllegalArgumentException("Input cannot be null");
            return new StringBuilder(s).reverse().toString();
        }

        static boolean isPalindrome(String s) {
            if (s == null) return false;
            String cleaned = s.replaceAll("[^a-zA-Z0-9]", "").toLowerCase();
            return cleaned.equals(new StringBuilder(cleaned).reverse().toString());
        }

        static String capitalize(String s) {
            if (s == null || s.isEmpty()) return s;
            return s.substring(0, 1).toUpperCase() + s.substring(1).toLowerCase();
        }
    }

    static class UserService {
        private final Map<String, String> users = new HashMap<>();

        String register(String username, String email) {
            if (username == null || username.trim().isEmpty())
                throw new IllegalArgumentException("Username required");
            if (email == null || !email.contains("@"))
                throw new IllegalArgumentException("Valid email required");
            if (users.containsKey(username))
                throw new IllegalStateException("Username taken");

            users.put(username, email);
            return "User " + username + " registered";
        }

        String findEmail(String username) {
            return users.get(username);
        }

        int getUserCount() { return users.size(); }
    }

    // ========================================================
    // Simple Test Framework (mimics JUnit)
    // ========================================================

    static int passed = 0, failed = 0;

    static void assertEquals(Object expected, Object actual, String testName) {
        if (Objects.equals(expected, actual)) {
            System.out.println("  ✓ " + testName);
            passed++;
        } else {
            System.out.println("  ✗ " + testName + " → Expected: " + expected + ", Got: " + actual);
            failed++;
        }
    }

    static void assertTrue(boolean condition, String testName) {
        assertEquals(true, condition, testName);
    }

    static void assertFalse(boolean condition, String testName) {
        assertEquals(false, condition, testName);
    }

    static void assertThrows(Class<? extends Exception> expected, Runnable action, String testName) {
        try {
            action.run();
            System.out.println("  ✗ " + testName + " → No exception thrown");
            failed++;
        } catch (Exception e) {
            if (expected.isInstance(e)) {
                System.out.println("  ✓ " + testName + " → Caught: " + e.getClass().getSimpleName());
                passed++;
            } else {
                System.out.println("  ✗ " + testName + " → Wrong exception: " + e.getClass().getSimpleName());
                failed++;
            }
        }
    }

    static void assertNull(Object obj, String testName) {
        assertEquals(null, obj, testName);
    }

    static void assertNotNull(Object obj, String testName) {
        if (obj != null) {
            System.out.println("  ✓ " + testName);
            passed++;
        } else {
            System.out.println("  ✗ " + testName + " → Was null");
            failed++;
        }
    }

    // ========================================================
    // Test Classes
    // ========================================================

    static void testCalculator() {
        System.out.println("\n--- Calculator Tests ---\n");
        Calculator calc = new Calculator();

        // Basic operations
        assertEquals(5, calc.add(2, 3), "add(2,3) = 5");
        assertEquals(0, calc.add(-1, 1), "add(-1,1) = 0");
        assertEquals(-5, calc.add(-2, -3), "add(-2,-3) = -5");

        assertEquals(1, calc.subtract(3, 2), "subtract(3,2) = 1");
        assertEquals(6, calc.multiply(2, 3), "multiply(2,3) = 6");
        assertEquals(0, calc.multiply(5, 0), "multiply(5,0) = 0");

        assertEquals(5, calc.divide(10, 2), "divide(10,2) = 5");
        assertEquals(-3, calc.divide(9, -3), "divide(9,-3) = -3");

        // Edge case: division by zero
        assertThrows(ArithmeticException.class,
            () -> calc.divide(10, 0),
            "divide by zero throws ArithmeticException");

        // Boundary values
        assertEquals(0, calc.add(Integer.MAX_VALUE, 1), "integer overflow wraps");
    }

    static void testStringUtils() {
        System.out.println("\n--- StringUtils Tests ---\n");

        // reverse
        assertEquals("olleh", StringUtils.reverse("hello"), "reverse 'hello'");
        assertEquals("", StringUtils.reverse(""), "reverse empty string");
        assertEquals("a", StringUtils.reverse("a"), "reverse single char");
        assertThrows(IllegalArgumentException.class,
            () -> StringUtils.reverse(null),
            "reverse null throws exception");

        // isPalindrome
        assertTrue(StringUtils.isPalindrome("racecar"), "racecar is palindrome");
        assertTrue(StringUtils.isPalindrome("A man a plan a canal Panama"),
            "phrase palindrome");
        assertFalse(StringUtils.isPalindrome("hello"), "hello is NOT palindrome");
        assertFalse(StringUtils.isPalindrome(null), "null is NOT palindrome");

        // capitalize
        assertEquals("Hello", StringUtils.capitalize("hello"), "capitalize hello");
        assertEquals("A", StringUtils.capitalize("a"), "capitalize single char");
        assertEquals("", StringUtils.capitalize(""), "capitalize empty");
        assertNull(StringUtils.capitalize(null), "capitalize null returns null");
    }

    static void testUserService() {
        System.out.println("\n--- UserService Tests ---\n");

        // Test registration
        UserService service = new UserService();
        assertEquals("User alice registered",
            service.register("alice", "alice@test.com"),
            "register valid user");
        assertEquals(1, service.getUserCount(), "user count after registration");

        // Test find
        assertEquals("alice@test.com", service.findEmail("alice"), "find existing user");
        assertNull(service.findEmail("bob"), "find non-existing user");

        // Test duplicate
        assertThrows(IllegalStateException.class,
            () -> service.register("alice", "alice2@test.com"),
            "duplicate username throws");

        // Test validation
        assertThrows(IllegalArgumentException.class,
            () -> service.register("", "test@test.com"),
            "empty username throws");
        assertThrows(IllegalArgumentException.class,
            () -> service.register(null, "test@test.com"),
            "null username throws");
        assertThrows(IllegalArgumentException.class,
            () -> service.register("bob", "invalid-email"),
            "invalid email throws");
        assertThrows(IllegalArgumentException.class,
            () -> service.register("bob", null),
            "null email throws");
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {
        System.out.println("=== UNIT TESTING ===\n");

        testCalculator();
        testStringUtils();
        testUserService();

        // Results
        System.out.println("\n" + "=".repeat(40));
        System.out.printf("  Results: %d passed, %d failed, %d total%n",
            passed, failed, passed + failed);
        System.out.println("=".repeat(40));

        // JUnit 5 Reference
        System.out.println("\n=== JUnit 5 ANNOTATIONS ===");
        System.out.println("  @Test           Mark test method");
        System.out.println("  @BeforeEach     Run before EACH test");
        System.out.println("  @AfterEach      Run after EACH test");
        System.out.println("  @BeforeAll      Run once before all tests (static)");
        System.out.println("  @AfterAll       Run once after all tests (static)");
        System.out.println("  @DisplayName    Custom test name");
        System.out.println("  @Disabled       Skip this test");
        System.out.println("  @Nested         Group tests in inner class");
        System.out.println("  @ParameterizedTest  Run with multiple inputs");
        System.out.println("  @RepeatedTest   Run N times");
        System.out.println("  @Tag            Categorize tests");
        System.out.println("  @Timeout        Fail if too slow");

        System.out.println("\n=== JUnit 5 ASSERTIONS ===");
        System.out.println("  assertEquals(expected, actual)");
        System.out.println("  assertNotEquals(a, b)");
        System.out.println("  assertTrue(condition)");
        System.out.println("  assertFalse(condition)");
        System.out.println("  assertNull(obj)");
        System.out.println("  assertNotNull(obj)");
        System.out.println("  assertThrows(Exception.class, () -> ...)");
        System.out.println("  assertAll(() -> ..., () -> ...)  // grouped");
        System.out.println("  assertTimeout(Duration.ofSeconds(1), () -> ...)");

        System.out.println("\n=== TESTING BEST PRACTICES ===");
        System.out.println("  1. Test ONE thing per test method");
        System.out.println("  2. Use descriptive test names (should_returnX_when_Y)");
        System.out.println("  3. Follow AAA: Arrange, Act, Assert");
        System.out.println("  4. Test edge cases: null, empty, boundary, negative");
        System.out.println("  5. Test exceptions (assertThrows)");
        System.out.println("  6. Tests should be independent (no shared state)");
        System.out.println("  7. Fast tests = run often = catch bugs early");
        System.out.println("  8. Aim for 80%+ code coverage");
        System.out.println("  9. Use Mockito to mock dependencies");
        System.out.println("  10. Write tests BEFORE code (TDD) for critical logic");

        System.out.println("\n✓ Unit Testing Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Set up JUnit 5 with Maven/Gradle and convert these tests to real JUnit.
 * 2. Write parameterized tests for isPalindrome with multiple inputs.
 * 3. Use Mockito to mock a database dependency in UserService.
 * 4. Aim for 100% coverage on the Calculator class.
 *
 * NEXT: Chapter 44 — Build Tools & Modules
 */
