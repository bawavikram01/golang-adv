/*
 * ============================================================
 *  CHAPTER 22: OPTIONAL
 * ============================================================
 *
 *  Optional<T> is a container that may or may not hold a value.
 *  Introduced in Java 8 to avoid NullPointerExceptions.
 *
 *  Instead of returning null, return Optional.empty().
 *  Instead of checking if (x != null), use Optional methods.
 *
 *  RULES:
 *  - NEVER use Optional for fields or method parameters
 *  - Use Optional ONLY as a return type
 *  - NEVER call .get() without checking isPresent() first
 *  - Prefer orElse/orElseGet/map/flatMap over isPresent+get
 *
 * ============================================================
 */

import java.util.*;

public class Chapter22_Optional {

    // =====================================================
    //  DEMO CLASSES
    // =====================================================

    static class User {
        String name;
        String email; // might be null

        User(String name, String email) {
            this.name = name;
            this.email = email;
        }

        // Return Optional instead of nullable value
        Optional<String> getEmail() {
            return Optional.ofNullable(email);
        }

        @Override
        public String toString() { return "User(" + name + ")"; }
    }

    static class UserRepository {
        private Map<Integer, User> users = new HashMap<>();

        UserRepository() {
            users.put(1, new User("Alice", "alice@example.com"));
            users.put(2, new User("Bob", null));
            users.put(3, new User("Charlie", "charlie@example.com"));
        }

        // Return Optional instead of null
        Optional<User> findById(int id) {
            return Optional.ofNullable(users.get(id));
        }
    }

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Creating Optionals ---
        System.out.println("=== CREATING OPTIONALS ===\n");

        Optional<String> present = Optional.of("Hello");     // value must NOT be null
        Optional<String> empty = Optional.empty();           // empty optional
        Optional<String> nullable = Optional.ofNullable(null); // may be null → empty
        Optional<String> nonNull = Optional.ofNullable("Hi");  // not null → present

        System.out.println("present: " + present);
        System.out.println("empty: " + empty);
        System.out.println("nullable: " + nullable);
        System.out.println("nonNull: " + nonNull);

        // Optional.of(null); // NullPointerException! Use ofNullable for potentially null values

        // --- 2. Checking and Getting Values ---
        System.out.println("\n=== CHECKING VALUES ===\n");

        // isPresent / isEmpty (Java 11+)
        System.out.println("present.isPresent(): " + present.isPresent()); // true
        System.out.println("empty.isPresent(): " + empty.isPresent());     // false
        System.out.println("empty.isEmpty(): " + empty.isEmpty());         // true (Java 11)

        // ifPresent — execute action if value exists
        present.ifPresent(val -> System.out.println("Value is: " + val));
        empty.ifPresent(val -> System.out.println("This won't print"));

        // --- 3. Getting Values Safely ---
        System.out.println("\n=== GETTING VALUES ===\n");

        // orElse — return default if empty
        String val1 = present.orElse("default");
        String val2 = empty.orElse("default");
        System.out.println("present.orElse: " + val1);   // "Hello"
        System.out.println("empty.orElse: " + val2);      // "default"

        // orElseGet — lazy default (computed only if needed)
        String val3 = empty.orElseGet(() -> "computed default");
        System.out.println("empty.orElseGet: " + val3);

        // orElseThrow — throw if empty
        try {
            String val4 = empty.orElseThrow(() -> new RuntimeException("No value!"));
        } catch (RuntimeException e) {
            System.out.println("orElseThrow: " + e.getMessage());
        }

        // get() — DANGEROUS! Throws NoSuchElementException if empty
        // String bad = empty.get(); // NEVER do this without checking!

        // --- 4. Transforming with map and flatMap ---
        System.out.println("\n=== MAP AND FLATMAP ===\n");

        Optional<String> name = Optional.of("alice");

        // map — transform the value
        Optional<String> upper = name.map(String::toUpperCase);
        System.out.println("map(toUpper): " + upper); // Optional[ALICE]

        Optional<Integer> length = name.map(String::length);
        System.out.println("map(length): " + length); // Optional[5]

        // map on empty returns empty
        Optional<String> emptyUpper = empty.map(String::toUpperCase);
        System.out.println("empty.map: " + emptyUpper); // Optional.empty

        // flatMap — when the mapping function returns Optional
        UserRepository repo = new UserRepository();

        Optional<String> aliceEmail = repo.findById(1)   // Optional<User>
                .flatMap(User::getEmail);                  // User::getEmail returns Optional<String>
        System.out.println("Alice email: " + aliceEmail);

        Optional<String> bobEmail = repo.findById(2)
                .flatMap(User::getEmail);
        System.out.println("Bob email: " + bobEmail); // Optional.empty (Bob has no email)

        Optional<String> unknownEmail = repo.findById(99)
                .flatMap(User::getEmail);
        System.out.println("Unknown email: " + unknownEmail); // Optional.empty

        // map vs flatMap:
        // map: takes T → R, wraps result in Optional
        // flatMap: takes T → Optional<R>, doesn't double-wrap

        // --- 5. Filtering ---
        System.out.println("\n=== FILTER ===\n");

        Optional<Integer> age = Optional.of(25);

        Optional<Integer> adult = age.filter(a -> a >= 18);
        Optional<Integer> teen = age.filter(a -> a < 18);

        System.out.println("filter(>=18): " + adult);  // Optional[25]
        System.out.println("filter(<18): " + teen);     // Optional.empty

        // --- 6. Chaining Operations ---
        System.out.println("\n=== CHAINING ===\n");

        // Complex chain: Find user → get email → extract domain → uppercase
        String domain = repo.findById(1)
                .flatMap(User::getEmail)
                .filter(email -> email.contains("@"))
                .map(email -> email.substring(email.indexOf("@") + 1))
                .map(String::toUpperCase)
                .orElse("NO DOMAIN");
        System.out.println("Domain: " + domain);

        // Same for unknown user — gracefully returns default
        String noDomain = repo.findById(99)
                .flatMap(User::getEmail)
                .filter(email -> email.contains("@"))
                .map(email -> email.substring(email.indexOf("@") + 1))
                .map(String::toUpperCase)
                .orElse("NO DOMAIN");
        System.out.println("Unknown domain: " + noDomain);

        // --- 7. Optional with Streams ---
        System.out.println("\n=== OPTIONAL WITH STREAMS ===\n");

        List<Optional<String>> optionals = Arrays.asList(
                Optional.of("Hello"),
                Optional.empty(),
                Optional.of("World"),
                Optional.empty(),
                Optional.of("Java")
        );

        // Filter present values and extract them
        List<String> values = new ArrayList<>();
        for (Optional<String> opt : optionals) {
            opt.ifPresent(values::add);
        }
        System.out.println("Present values: " + values);

        // --- 8. or() method (Java 9+) ---
        System.out.println("\n=== OR METHOD ===\n");

        // or() provides alternative Optional (not just default value)
        Optional<String> result = empty
                .or(() -> Optional.of("Fallback value"));
        System.out.println("or(): " + result);

        // --- 9. Anti-patterns ---
        System.out.println("\n=== ANTI-PATTERNS (DON'T DO THESE!) ===\n");

        System.out.println("❌ optional.get() without isPresent()");
        System.out.println("❌ optional.isPresent() ? optional.get() : default");
        System.out.println("   ✅ Use optional.orElse(default)");
        System.out.println();
        System.out.println("❌ Optional<T> as method parameter");
        System.out.println("   ✅ Use method overloading or nullable param");
        System.out.println();
        System.out.println("❌ Optional<T> as class field");
        System.out.println("   ✅ Use nullable field + getter returning Optional");
        System.out.println();
        System.out.println("❌ Optional.of(nullable) // NPE risk!");
        System.out.println("   ✅ Optional.ofNullable(nullable)");
        System.out.println();
        System.out.println("❌ if (optional.isPresent()) { doSomething(optional.get()); }");
        System.out.println("   ✅ optional.ifPresent(this::doSomething)");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Write a method findLongestString(List<String>) that returns
 *     Optional<String>. Handle empty list gracefully.
 *
 *  2. Chain Optional operations to: find user → get address →
 *     get city → convert to uppercase → return or "UNKNOWN".
 *
 *  3. Convert a list of Optional<Integer> to a list of present
 *     values doubled.
 *
 *  4. Implement a safe division method that returns Optional<Double>.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 23 — File I/O
 * ============================================================
 */
