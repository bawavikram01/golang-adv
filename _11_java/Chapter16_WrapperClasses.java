/*
 * ============================================================
 *  CHAPTER 16: WRAPPER CLASSES & AUTOBOXING
 * ============================================================
 *
 *  Wrapper classes wrap primitive types into objects.
 *  Needed because collections (List, Map) only work with objects.
 *
 *  PRIMITIVE → WRAPPER:
 *  ┌──────────┬────────────┐
 *  │ byte     │ Byte       │
 *  │ short    │ Short      │
 *  │ int      │ Integer    │
 *  │ long     │ Long       │
 *  │ float    │ Float      │
 *  │ double   │ Double     │
 *  │ char     │ Character  │
 *  │ boolean  │ Boolean    │
 *  └──────────┴────────────┘
 *
 *  AUTOBOXING:  primitive → wrapper (automatic)
 *  UNBOXING:    wrapper → primitive (automatic)
 *
 * ============================================================
 */

import java.util.ArrayList;
import java.util.List;

public class Chapter16_WrapperClasses {

    public static void main(String[] args) {

        // =====================================================
        //  1. MANUAL BOXING AND UNBOXING
        // =====================================================

        System.out.println("=== MANUAL BOXING/UNBOXING ===\n");

        // Boxing: primitive → wrapper (manual, old way)
        Integer intObj = Integer.valueOf(42);
        Double dblObj = Double.valueOf(3.14);
        Boolean boolObj = Boolean.valueOf(true);
        Character charObj = Character.valueOf('A');

        // Unboxing: wrapper → primitive (manual)
        int intVal = intObj.intValue();
        double dblVal = dblObj.doubleValue();
        boolean boolVal = boolObj.booleanValue();
        char charVal = charObj.charValue();

        System.out.println("Integer: " + intObj + " → int: " + intVal);
        System.out.println("Double: " + dblObj + " → double: " + dblVal);

        // =====================================================
        //  2. AUTOBOXING AND AUTO-UNBOXING (Java 5+)
        // =====================================================

        System.out.println("\n=== AUTOBOXING ===\n");

        // Autoboxing: primitive → wrapper (automatic)
        Integer autoBoxed = 42;          // equivalent to Integer.valueOf(42)
        Double autoDouble = 3.14;
        Boolean autoBool = true;

        // Auto-unboxing: wrapper → primitive (automatic)
        int autoUnboxed = autoBoxed;     // equivalent to autoBoxed.intValue()
        double autoUnboxDbl = autoDouble;

        System.out.println("Autoboxed Integer: " + autoBoxed);
        System.out.println("Auto-unboxed int: " + autoUnboxed);

        // Autoboxing in expressions
        Integer a = 10, b = 20;
        int sum = a + b;  // a and b auto-unboxed, then added
        System.out.println("a + b = " + sum);

        // Autoboxing with collections
        List<Integer> numbers = new ArrayList<>();
        numbers.add(1);     // autoboxing: int → Integer
        numbers.add(2);
        numbers.add(3);
        int first = numbers.get(0); // auto-unboxing: Integer → int
        System.out.println("List: " + numbers + ", first: " + first);

        // =====================================================
        //  3. PARSING STRINGS TO PRIMITIVES
        // =====================================================

        System.out.println("\n=== PARSING STRINGS ===\n");

        // String → primitive (parse methods)
        int parsedInt = Integer.parseInt("123");
        double parsedDbl = Double.parseDouble("3.14159");
        long parsedLong = Long.parseLong("999999999");
        boolean parsedBool = Boolean.parseBoolean("true");
        float parsedFloat = Float.parseFloat("2.71828");

        System.out.println("parseInt(\"123\"): " + parsedInt);
        System.out.println("parseDouble(\"3.14159\"): " + parsedDbl);
        System.out.println("parseLong: " + parsedLong);
        System.out.println("parseBoolean(\"true\"): " + parsedBool);

        // String → wrapper (valueOf)
        Integer intFromStr = Integer.valueOf("456");
        Double dblFromStr = Double.valueOf("2.718");
        System.out.println("Integer.valueOf(\"456\"): " + intFromStr);

        // Primitive/wrapper → String
        String fromInt = Integer.toString(42);
        String fromDbl = Double.toString(3.14);
        String fromVal = String.valueOf(99);
        String concat = "" + 42; // concatenation (less efficient)
        System.out.println("toString: " + fromInt + ", valueOf: " + fromVal);

        // Different bases
        System.out.println("\n--- Parsing Different Bases ---");
        System.out.println("parseInt(\"FF\", 16) = " + Integer.parseInt("FF", 16));  // 255
        System.out.println("parseInt(\"1010\", 2) = " + Integer.parseInt("1010", 2));  // 10
        System.out.println("parseInt(\"77\", 8) = " + Integer.parseInt("77", 8));     // 63

        // Converting to different base strings
        System.out.println("toBinaryString(255) = " + Integer.toBinaryString(255));
        System.out.println("toOctalString(255) = " + Integer.toOctalString(255));
        System.out.println("toHexString(255) = " + Integer.toHexString(255));

        // =====================================================
        //  4. WRAPPER CLASS UTILITY METHODS
        // =====================================================

        System.out.println("\n=== UTILITY METHODS ===\n");

        // Constants
        System.out.println("Integer.MAX_VALUE: " + Integer.MAX_VALUE);
        System.out.println("Integer.MIN_VALUE: " + Integer.MIN_VALUE);
        System.out.println("Integer.SIZE: " + Integer.SIZE + " bits");
        System.out.println("Integer.BYTES: " + Integer.BYTES + " bytes");
        System.out.println("Double.MAX_VALUE: " + Double.MAX_VALUE);
        System.out.println("Double.NaN: " + Double.NaN);
        System.out.println("Double.POSITIVE_INFINITY: " + Double.POSITIVE_INFINITY);

        // Comparison
        System.out.println("\n--- Comparison ---");
        System.out.println("Integer.compare(5, 10): " + Integer.compare(5, 10));   // -1
        System.out.println("Integer.compare(10, 10): " + Integer.compare(10, 10)); // 0
        System.out.println("Integer.compare(10, 5): " + Integer.compare(10, 5));   // 1
        System.out.println("Integer.max(5, 10): " + Integer.max(5, 10));
        System.out.println("Integer.min(5, 10): " + Integer.min(5, 10));
        System.out.println("Integer.sum(5, 10): " + Integer.sum(5, 10));

        // Character methods
        System.out.println("\n--- Character Methods ---");
        System.out.println("isDigit('5'): " + Character.isDigit('5'));         // true
        System.out.println("isLetter('A'): " + Character.isLetter('A'));       // true
        System.out.println("isLetterOrDigit('5'): " + Character.isLetterOrDigit('5')); // true
        System.out.println("isUpperCase('A'): " + Character.isUpperCase('A')); // true
        System.out.println("isLowerCase('a'): " + Character.isLowerCase('a')); // true
        System.out.println("isWhitespace(' '): " + Character.isWhitespace(' ')); // true
        System.out.println("toUpperCase('a'): " + Character.toUpperCase('a')); // A
        System.out.println("toLowerCase('A'): " + Character.toLowerCase('A')); // a

        // =====================================================
        //  5. CACHING & COMPARISON PITFALLS
        // =====================================================

        System.out.println("\n=== CACHING PITFALL ===\n");

        // Integer caches values from -128 to 127!
        Integer x = 127;
        Integer y = 127;
        System.out.println("127 == 127: " + (x == y));   // TRUE (cached!)

        Integer p = 128;
        Integer q = 128;
        System.out.println("128 == 128: " + (p == q));   // FALSE (not cached!)

        // ALWAYS use .equals() for wrapper comparison!
        System.out.println("128.equals(128): " + p.equals(q)); // TRUE

        // Same caching for: Byte, Short, Integer, Long (-128 to 127)
        // Boolean: TRUE and FALSE cached
        // Character: 0 to 127 cached

        // =====================================================
        //  6. NULL DANGER WITH UNBOXING
        // =====================================================

        System.out.println("\n=== NULL UNBOXING ===\n");

        Integer nullInt = null;
        try {
            int dangerous = nullInt; // NullPointerException on unboxing!
        } catch (NullPointerException e) {
            System.out.println("NPE when unboxing null Integer!");
        }

        // Always check for null before unboxing
        Integer maybNull = null;
        int safe = (maybNull != null) ? maybNull : 0; // default to 0
        System.out.println("Safe unboxing with null check: " + safe);

        // =====================================================
        //  7. NUMBER CLASS (Parent of all numeric wrappers)
        // =====================================================

        System.out.println("\n=== NUMBER CLASS ===\n");

        // All numeric wrappers extend Number
        Number num = Integer.valueOf(42);
        System.out.println("intValue: " + num.intValue());
        System.out.println("doubleValue: " + num.doubleValue());
        System.out.println("longValue: " + num.longValue());
        System.out.println("floatValue: " + num.floatValue());

        // Polymorphic use
        Number[] nums = {42, 3.14, 100L, 2.71f};
        for (Number n : nums) {
            System.out.printf("  %s → double: %.2f%n",
                    n.getClass().getSimpleName(), n.doubleValue());
        }

        // =====================================================
        //  8. PRACTICAL EXAMPLES
        // =====================================================

        System.out.println("\n=== PRACTICAL EXAMPLES ===\n");

        // Count digits in a string
        String text = "Hello 123 World 456";
        int digitCount = 0;
        for (char c : text.toCharArray()) {
            if (Character.isDigit(c)) digitCount++;
        }
        System.out.println("Digits in '" + text + "': " + digitCount);

        // Validate integer input
        String[] inputs = {"42", "hello", "99", "3.14", "-7"};
        System.out.println("\nValidating inputs:");
        for (String input : inputs) {
            try {
                int val = Integer.parseInt(input);
                System.out.println("  '" + input + "' → valid integer: " + val);
            } catch (NumberFormatException e) {
                System.out.println("  '" + input + "' → NOT a valid integer");
            }
        }
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Convert a char array to uppercase using Character methods.
 *
 *  2. Write a method that takes a String and returns true if
 *     it represents a valid integer (handle all edge cases).
 *
 *  3. What's the output? Why?
 *     Integer a = 100, b = 100;
 *     Integer c = 200, d = 200;
 *     System.out.println(a == b);  // ???
 *     System.out.println(c == d);  // ???
 *
 *  4. Write a method that safely unboxes a list of Integer values,
 *     replacing nulls with a default value.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 17 — Collections Framework
 * ============================================================
 */
