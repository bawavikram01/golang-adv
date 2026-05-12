/*
 * ============================================================
 *  CHAPTER 27: REGULAR EXPRESSIONS
 * ============================================================
 *  Regex = Pattern matching in strings.
 *  Classes: Pattern, Matcher (java.util.regex)
 *
 *  SYNTAX QUICK REFERENCE:
 *  .      → any character           \d → digit [0-9]
 *  *      → 0 or more               \D → non-digit
 *  +      → 1 or more               \w → word char [a-zA-Z0-9_]
 *  ?      → 0 or 1                  \W → non-word char
 *  {n}    → exactly n               \s → whitespace
 *  {n,m}  → n to m times            \S → non-whitespace
 *  ^      → start of string         \b → word boundary
 *  $      → end of string           [abc] → char class
 *  |      → OR                      [^abc] → negated class
 *  ()     → group                   [a-z] → range
 * ============================================================
 */

import java.util.regex.*;
import java.util.List;
import java.util.ArrayList;

public class Chapter27_Regex {

    static void testMatch(String regex, String input) {
        boolean matches = input.matches(regex);
        System.out.printf("  %-20s matches %-15s → %s%n", "\"" + input + "\"", "/" + regex + "/", matches);
    }

    static void findAll(String regex, String input) {
        Pattern pattern = Pattern.compile(regex);
        Matcher matcher = pattern.matcher(input);
        List<String> found = new ArrayList<>();
        while (matcher.find()) {
            found.add(matcher.group());
        }
        System.out.printf("  /%s/ in \"%s\" → %s%n", regex, input, found);
    }

    public static void main(String[] args) {

        // --- 1. Basic Matching ---
        System.out.println("=== BASIC MATCHING ===\n");
        testMatch("hello", "hello");           // true (exact match)
        testMatch("hello", "Hello");           // false (case-sensitive)
        testMatch("(?i)hello", "Hello");       // true (case-insensitive flag)
        testMatch("h.llo", "hello");           // true (. = any char)
        testMatch("h.*o", "hello");            // true (.* = any chars)

        // --- 2. Character Classes ---
        System.out.println("\n=== CHARACTER CLASSES ===\n");
        testMatch("[abc]", "a");               // true
        testMatch("[a-z]+", "hello");          // true
        testMatch("[A-Za-z]+", "Hello");       // true
        testMatch("[0-9]+", "12345");          // true
        testMatch("\\d+", "12345");            // true (\d = digit)
        testMatch("\\w+", "hello_123");        // true (\w = word char)

        // --- 3. Quantifiers ---
        System.out.println("\n=== QUANTIFIERS ===\n");
        testMatch("a?b", "b");                 // true (a is optional)
        testMatch("a?b", "ab");                // true
        testMatch("a+b", "aaab");              // true (1+ a's)
        testMatch("a*b", "b");                 // true (0+ a's)
        testMatch("a{3}b", "aaab");            // true (exactly 3 a's)
        testMatch("a{2,4}b", "aaab");          // true (2-4 a's)

        // --- 4. Finding Matches ---
        System.out.println("\n=== FINDING MATCHES ===\n");
        findAll("\\d+", "I have 3 cats and 12 dogs");
        findAll("[A-Z][a-z]+", "Hello World Java");
        findAll("\\b\\w{4}\\b", "The quick brown fox jumps over");
        findAll("\\S+@\\S+", "email me at user@test.com or admin@site.org");

        // --- 5. Groups ---
        System.out.println("\n=== GROUPS ===\n");
        Pattern datePattern = Pattern.compile("(\\d{4})-(\\d{2})-(\\d{2})");
        Matcher dateMatcher = datePattern.matcher("Today is 2024-12-25 and tomorrow is 2024-12-26");

        while (dateMatcher.find()) {
            System.out.println("Full: " + dateMatcher.group());
            System.out.println("  Year: " + dateMatcher.group(1));
            System.out.println("  Month: " + dateMatcher.group(2));
            System.out.println("  Day: " + dateMatcher.group(3));
        }

        // Named groups
        Pattern namedPattern = Pattern.compile("(?<name>\\w+)=(?<value>\\w+)");
        Matcher namedMatcher = namedPattern.matcher("name=Alice age=25 city=NYC");
        System.out.println("\nNamed groups:");
        while (namedMatcher.find()) {
            System.out.println("  " + namedMatcher.group("name") + " → " + namedMatcher.group("value"));
        }

        // --- 6. String Methods with Regex ---
        System.out.println("\n=== STRING METHODS ===\n");

        // split
        String csv = "apple,,banana, ,cherry";
        String[] parts = csv.split(",\\s*");
        System.out.println("Split: " + java.util.Arrays.toString(parts));

        // replaceAll
        String cleaned = "Hello   World   Java".replaceAll("\\s+", " ");
        System.out.println("Cleaned: " + cleaned);

        String censored = "My phone is 123-456-7890".replaceAll("\\d", "*");
        System.out.println("Censored: " + censored);

        // matches
        System.out.println("\n--- Validation ---");
        System.out.println("Email valid: " + "user@test.com".matches("^[\\w.-]+@[\\w.-]+\\.\\w{2,}$"));
        System.out.println("Phone valid: " + "123-456-7890".matches("\\d{3}-\\d{3}-\\d{4}"));
        System.out.println("IP valid: " + "192.168.1.1".matches("\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}"));

        // --- 7. Common Patterns ---
        System.out.println("\n=== COMMON PATTERNS ===\n");
        System.out.println("Email:    ^[\\\\w.-]+@[\\\\w.-]+\\\\.\\\\w{2,}$");
        System.out.println("Phone:    \\\\d{3}-\\\\d{3}-\\\\d{4}");
        System.out.println("URL:      https?://[\\\\w.-]+(/\\\\S*)?");
        System.out.println("Integer:  -?\\\\d+");
        System.out.println("Decimal:  -?\\\\d+\\\\.\\\\d+");
        System.out.println("Date:     \\\\d{4}-\\\\d{2}-\\\\d{2}");
    }
}

/*
 * EXERCISES:
 * 1. Validate passwords: 8+ chars, uppercase, lowercase, digit, special.
 * 2. Extract all URLs from a text.
 * 3. Find and replace dates from MM/DD/YYYY to YYYY-MM-DD format.
 * 4. Write a regex to validate IP addresses (proper range 0-255).
 *
 * NEXT: Chapter 28 — Multithreading Basics
 */
