/*
 * ============================================================
 *  CHAPTER 07: STRINGS
 * ============================================================
 *
 *  Strings are one of the most used types in Java.
 *
 *  THREE STRING CLASSES:
 *  ─────────────────────
 *  1. String         → IMMUTABLE (cannot be changed after creation)
 *  2. StringBuilder  → MUTABLE, NOT thread-safe, FAST
 *  3. StringBuffer   → MUTABLE, thread-safe, SLOWER
 *
 *  STRING POOL (String Intern Pool):
 *  ─────────────────────────────────
 *  Java maintains a special memory area in the heap called
 *  the "String Pool" (or Intern Pool).
 *
 *  When you create a String literal:
 *    String s1 = "Hello";  → checks pool, creates if not found
 *    String s2 = "Hello";  → finds "Hello" in pool, reuses it!
 *    s1 == s2 → TRUE (same object in pool)
 *
 *  When you use new:
 *    String s3 = new String("Hello"); → creates NEW object in heap
 *    s1 == s3 → FALSE (different objects!)
 *    s1.equals(s3) → TRUE (same content)
 *
 *     HEAP MEMORY
 *    ┌─────────────────────────────┐
 *    │  String Pool                │
 *    │  ┌──────────────┐           │
 *    │  │  "Hello"  ←──┼─── s1    │
 *    │  │           ←──┼─── s2    │
 *    │  └──────────────┘           │
 *    │                             │
 *    │  ┌──────────────┐           │
 *    │  │  "Hello"  ←──┼─── s3    │  (separate object)
 *    │  └──────────────┘           │
 *    └─────────────────────────────┘
 *
 * ============================================================
 */

import java.util.Arrays;

public class Chapter07_Strings {

    public static void main(String[] args) {

        // =====================================================
        //  1. STRING CREATION
        // =====================================================

        System.out.println("=== STRING CREATION ===\n");

        // Literal (uses String Pool)
        String s1 = "Hello";
        String s2 = "Hello";

        // Using new (creates separate object in heap)
        String s3 = new String("Hello");

        // From char array
        char[] chars = {'J', 'a', 'v', 'a'};
        String s4 = new String(chars);

        // From byte array
        byte[] bytes = {72, 101, 108, 108, 111}; // ASCII for "Hello"
        String s5 = new String(bytes);

        System.out.println("s1: " + s1);
        System.out.println("s4 from chars: " + s4);
        System.out.println("s5 from bytes: " + s5);

        // == vs .equals()
        System.out.println("\n--- == vs .equals() ---");
        System.out.println("s1 == s2: " + (s1 == s2));           // true  (same pool object)
        System.out.println("s1 == s3: " + (s1 == s3));           // false (different objects)
        System.out.println("s1.equals(s3): " + s1.equals(s3));   // true  (same content)

        // intern() — forces string into pool
        String s6 = s3.intern();
        System.out.println("s1 == s3.intern(): " + (s1 == s6));  // true!

        // =====================================================
        //  2. STRING IMMUTABILITY
        // =====================================================

        System.out.println("\n=== STRING IMMUTABILITY ===\n");

        String original = "Hello";
        String modified = original.concat(" World"); // creates NEW string

        System.out.println("original: " + original);   // "Hello" (unchanged!)
        System.out.println("modified: " + modified);    // "Hello World" (new object)

        // Every String method returns a NEW String — original is NEVER modified
        String upper = original.toUpperCase(); // "HELLO" — new object
        System.out.println("upper: " + upper);
        System.out.println("original still: " + original); // still "Hello"

        // Why immutable?
        // 1. Security: Strings used in network connections, file paths, class loading
        // 2. Thread-safe: Can be shared between threads without synchronization
        // 3. Caching: hashCode can be cached (used in HashMap keys)
        // 4. String Pool: Only works because Strings are immutable

        // =====================================================
        //  3. STRING METHODS — EXHAUSTIVE LIST
        // =====================================================

        System.out.println("\n=== STRING METHODS ===\n");

        String str = "  Hello, World! Welcome to Java!  ";

        // --- LENGTH ---
        System.out.println("--- Length & Empty ---");
        System.out.println("length(): " + str.length());        // 35
        System.out.println("isEmpty(): " + str.isEmpty());       // false
        System.out.println("\"\".isEmpty(): " + "".isEmpty());   // true

        // --- CASE ---
        System.out.println("\n--- Case ---");
        System.out.println("toUpperCase(): " + "hello".toUpperCase());   // HELLO
        System.out.println("toLowerCase(): " + "HELLO".toLowerCase());   // hello

        // --- TRIM & STRIP ---
        System.out.println("\n--- Trim/Strip ---");
        System.out.println("trim(): '" + str.trim() + "'");       // removes leading/trailing spaces
        System.out.println("strip(): '" + str.strip() + "'");     // Java 11: Unicode-aware trim
        System.out.println("stripLeading(): '" + str.stripLeading() + "'");
        System.out.println("stripTrailing(): '" + str.stripTrailing() + "'");

        // --- SEARCH ---
        System.out.println("\n--- Search ---");
        String text = "Hello World Hello Java";
        System.out.println("indexOf('o'): " + text.indexOf('o'));           // 4
        System.out.println("indexOf('o', 5): " + text.indexOf('o', 5));     // 7 (start from index 5)
        System.out.println("indexOf(\"World\"): " + text.indexOf("World")); // 6
        System.out.println("lastIndexOf('o'): " + text.lastIndexOf('o'));   // 7
        System.out.println("contains(\"Java\"): " + text.contains("Java")); // true

        // --- COMPARISON ---
        System.out.println("\n--- Comparison ---");
        System.out.println("equals: " + "hello".equals("hello"));                   // true
        System.out.println("equals: " + "hello".equals("Hello"));                   // false
        System.out.println("equalsIgnoreCase: " + "hello".equalsIgnoreCase("Hello"));// true
        System.out.println("compareTo: " + "apple".compareTo("banana"));            // negative (a < b)
        System.out.println("compareTo: " + "banana".compareTo("apple"));            // positive (b > a)
        System.out.println("compareTo: " + "apple".compareTo("apple"));             // 0 (equal)
        System.out.println("startsWith(\"Hello\"): " + text.startsWith("Hello"));   // true
        System.out.println("endsWith(\"Java\"): " + text.endsWith("Java"));         // true

        // --- SUBSTRING ---
        System.out.println("\n--- Substring ---");
        String sub = "Hello World";
        System.out.println("substring(6): " + sub.substring(6));        // "World"
        System.out.println("substring(0,5): " + sub.substring(0, 5));   // "Hello" (end exclusive)

        // --- REPLACE ---
        System.out.println("\n--- Replace ---");
        System.out.println("replace('l','r'): " + "Hello".replace('l', 'r'));             // "Herro"
        System.out.println("replace(\"World\",\"Java\"): " + "Hello World".replace("World", "Java")); // "Hello Java"
        System.out.println("replaceAll regex: " + "a1b2c3".replaceAll("[0-9]", "*"));      // "a*b*c*"
        System.out.println("replaceFirst: " + "aaa".replaceFirst("a", "b"));               // "baa"

        // --- SPLIT ---
        System.out.println("\n--- Split ---");
        String csv = "apple,banana,cherry,date";
        String[] parts = csv.split(",");
        System.out.println("split(','): " + Arrays.toString(parts));

        String words = "Hello   World   Java";
        String[] wordArr = words.split("\\s+"); // split by one or more spaces
        System.out.println("split(spaces): " + Arrays.toString(wordArr));

        // Split with limit
        String limited = "a:b:c:d:e";
        String[] limitedArr = limited.split(":", 3); // max 3 parts
        System.out.println("split(':',3): " + Arrays.toString(limitedArr)); // [a, b, c:d:e]

        // --- JOIN ---
        System.out.println("\n--- Join ---");
        String joined = String.join(", ", "Apple", "Banana", "Cherry");
        System.out.println("join: " + joined);  // "Apple, Banana, Cherry"

        String joinedArr = String.join(" -> ", parts);
        System.out.println("join array: " + joinedArr);

        // --- CHAR ACCESS ---
        System.out.println("\n--- Character Access ---");
        String hello = "Hello";
        System.out.println("charAt(0): " + hello.charAt(0));     // H
        System.out.println("charAt(4): " + hello.charAt(4));     // o
        System.out.println("toCharArray: " + Arrays.toString(hello.toCharArray())); // [H,e,l,l,o]

        // --- CONVERSION ---
        System.out.println("\n--- Conversion ---");
        System.out.println("valueOf(42): " + String.valueOf(42));          // "42"
        System.out.println("valueOf(3.14): " + String.valueOf(3.14));      // "3.14"
        System.out.println("valueOf(true): " + String.valueOf(true));      // "true"
        System.out.println("valueOf(char[]): " + String.valueOf(new char[]{'A','B','C'})); // "ABC"

        // --- FORMAT ---
        System.out.println("\n--- Format ---");
        String formatted = String.format("Name: %s, Age: %d, GPA: %.2f", "Vikram", 25, 3.95);
        System.out.println(formatted);

        // --- REPEAT (Java 11) ---
        System.out.println("\n--- Repeat (Java 11) ---");
        System.out.println("\"ha\".repeat(3): " + "ha".repeat(3));   // "hahaha"
        System.out.println("\"=-\".repeat(10): " + "=-".repeat(10));

        // --- isBlank (Java 11) ---
        System.out.println("\n--- isBlank (Java 11) ---");
        System.out.println("\"\".isBlank(): " + "".isBlank());           // true
        System.out.println("\"  \".isBlank(): " + "   ".isBlank());     // true (whitespace only)
        System.out.println("\" a \".isBlank(): " + " a ".isBlank());    // false

        // =====================================================
        //  4. STRING CONCATENATION PERFORMANCE
        // =====================================================

        System.out.println("\n=== CONCATENATION PERFORMANCE ===\n");

        // BAD: String concatenation in a loop creates many temporary objects
        // Each + creates a new String object!
        String bad = "";
        long start = System.nanoTime();
        for (int i = 0; i < 10000; i++) {
            bad = bad + i; // creates 10000 String objects!
        }
        long badTime = System.nanoTime() - start;
        System.out.println("String + in loop: " + badTime / 1_000_000 + " ms");

        // GOOD: StringBuilder — modifies same object, no copies
        StringBuilder good = new StringBuilder();
        start = System.nanoTime();
        for (int i = 0; i < 10000; i++) {
            good.append(i); // modifies same StringBuilder
        }
        String result = good.toString();
        long goodTime = System.nanoTime() - start;
        System.out.println("StringBuilder:    " + goodTime / 1_000_000 + " ms");
        System.out.println("StringBuilder is ~" + (badTime / Math.max(1, goodTime)) + "x faster!");

        // =====================================================
        //  5. STRINGBUILDER
        // =====================================================

        System.out.println("\n=== STRINGBUILDER ===\n");

        // StringBuilder: mutable, NOT thread-safe, fast
        StringBuilder sb = new StringBuilder("Hello");

        // Append
        sb.append(" World");
        sb.append("!");
        System.out.println("append: " + sb);              // "Hello World!"

        // Insert
        sb.insert(5, ",");
        System.out.println("insert: " + sb);              // "Hello, World!"

        // Replace
        sb.replace(7, 12, "Java");
        System.out.println("replace: " + sb);             // "Hello, Java!!"

        // Delete
        sb.delete(11, 12);
        System.out.println("delete: " + sb);              // "Hello, Java!"

        // deleteCharAt
        sb.deleteCharAt(5); // remove comma
        System.out.println("deleteCharAt: " + sb);        // "Hello Java!"

        // Reverse
        sb.reverse();
        System.out.println("reverse: " + sb);             // "!avaJ olleH"

        // Back to normal
        sb.reverse();

        // charAt, indexOf, length, substring — same as String
        System.out.println("charAt(0): " + sb.charAt(0));
        System.out.println("indexOf(\"Java\"): " + sb.indexOf("Java"));
        System.out.println("length: " + sb.length());
        System.out.println("substring(6): " + sb.substring(6));

        // Capacity
        StringBuilder cap = new StringBuilder(); // default capacity = 16
        System.out.println("\nCapacity: " + cap.capacity());    // 16
        cap.append("Hello World Extra Text Here");
        System.out.println("After append capacity: " + cap.capacity()); // auto-expanded

        // Convert to String
        String finalStr = sb.toString();
        System.out.println("toString: " + finalStr);

        // =====================================================
        //  6. STRINGBUFFER (Thread-safe version of StringBuilder)
        // =====================================================

        System.out.println("\n=== STRINGBUFFER ===\n");

        // Same methods as StringBuilder, but synchronized (thread-safe)
        // Use when multiple threads access the same string
        StringBuffer sbuf = new StringBuffer("Thread");
        sbuf.append(" Safe");
        System.out.println("StringBuffer: " + sbuf);
        System.out.println("Same methods as StringBuilder,  but thread-safe.");
        System.out.println("Use StringBuilder unless you need thread safety.");

        // =====================================================
        //  7. STRING COMPARISON TABLE
        // =====================================================

        System.out.println("\n=== STRING vs STRINGBUILDER vs STRINGBUFFER ===\n");
        System.out.println("┌──────────────┬───────────┬─────────────┬──────────────┐");
        System.out.println("│ Feature      │  String   │StringBuilder│ StringBuffer │");
        System.out.println("├──────────────┼───────────┼─────────────┼──────────────┤");
        System.out.println("│ Mutable?     │    No     │     Yes     │     Yes      │");
        System.out.println("│ Thread-safe? │    Yes*   │     No      │     Yes      │");
        System.out.println("│ Performance  │  Slowest  │   Fastest   │    Medium    │");
        System.out.println("│ String Pool  │    Yes    │     No      │     No       │");
        System.out.println("└──────────────┴───────────┴─────────────┴──────────────┘");
        System.out.println("  *String is thread-safe because it's immutable");

        // =====================================================
        //  8. PRACTICAL EXAMPLES
        // =====================================================

        System.out.println("\n=== PRACTICAL EXAMPLES ===\n");

        // Reverse a string
        String rev = "Java Programming";
        String reversed = new StringBuilder(rev).reverse().toString();
        System.out.println("Reverse: " + reversed);

        // Check palindrome
        String pal = "racecar";
        boolean isPalindrome = pal.equals(new StringBuilder(pal).reverse().toString());
        System.out.println("\"" + pal + "\" palindrome? " + isPalindrome);

        // Count vowels
        String vowelStr = "Hello World";
        int vowelCount = 0;
        for (char ch : vowelStr.toLowerCase().toCharArray()) {
            if ("aeiou".indexOf(ch) != -1) vowelCount++;
        }
        System.out.println("Vowels in \"" + vowelStr + "\": " + vowelCount);

        // Count words
        String sentence = "  The quick  brown fox  jumps  ";
        int wordCount = sentence.trim().split("\\s+").length;
        System.out.println("Words in sentence: " + wordCount);

        // Title case
        String title = "hello world from java";
        StringBuilder titleCase = new StringBuilder();
        for (String w : title.split(" ")) {
            if (titleCase.length() > 0) titleCase.append(" ");
            titleCase.append(Character.toUpperCase(w.charAt(0)))
                     .append(w.substring(1));
        }
        System.out.println("Title case: " + titleCase);

        // Remove duplicates
        String dupes = "programming";
        StringBuilder unique = new StringBuilder();
        for (char ch : dupes.toCharArray()) {
            if (unique.indexOf(String.valueOf(ch)) == -1) {
                unique.append(ch);
            }
        }
        System.out.println("Remove dupes from \"" + dupes + "\": " + unique);

        // Character frequency
        String freq = "mississippi";
        System.out.println("Char frequency in \"" + freq + "\":");
        for (char ch : freq.toCharArray()) {
            int count = freq.length() - freq.replace(String.valueOf(ch), "").length();
            if (freq.indexOf(ch) == freq.lastIndexOf(ch) || freq.indexOf(ch) == freq.indexOf(ch)) {
                // Print each unique char once
            }
        }
        // Better approach with boolean tracking
        boolean[] seen = new boolean[128]; // ASCII
        for (char ch : freq.toCharArray()) {
            if (!seen[ch]) {
                int count = 0;
                for (char c : freq.toCharArray()) {
                    if (c == ch) count++;
                }
                System.out.println("  '" + ch + "': " + count);
                seen[ch] = true;
            }
        }

        // Anagram check
        String a1 = "listen";
        String a2 = "silent";
        char[] c1 = a1.toCharArray();
        char[] c2 = a2.toCharArray();
        Arrays.sort(c1);
        Arrays.sort(c2);
        boolean isAnagram = Arrays.equals(c1, c2);
        System.out.println("\"" + a1 + "\" & \"" + a2 + "\" anagram? " + isAnagram);
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Write a method to count the number of occurrences of a
 *     character in a string.
 *
 *  2. Write a method to find the first non-repeating character.
 *     "stress" → 't'
 *
 *  3. Write a method to check if two strings are rotations.
 *     "abcde" and "cdeab" → true
 *
 *  4. Implement a basic string compression:
 *     "aabcccccaaa" → "a2b1c5a3"
 *
 *  5. Write a method to reverse words in a sentence:
 *     "Hello World" → "World Hello"
 *
 *  6. Check if a string has all unique characters
 *     (without using additional data structures).
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 08 — OOP: Classes & Objects
 * ============================================================
 */
