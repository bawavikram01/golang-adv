/*
 * ============================================================
 *  CHAPTER 02: VARIABLES & DATA TYPES
 * ============================================================
 *
 *  WHAT IS A VARIABLE?
 *  --------------------
 *  A variable is a named container that stores data in memory.
 *  Think of it as a labeled box that holds a value.
 *
 *  Syntax: dataType variableName = value;
 *  Example: int age = 25;
 *
 *  NAMING RULES:
 *  1. Can contain letters, digits, _ and $
 *  2. Must start with a letter, _ or $ (NOT a digit)
 *  3. Cannot use Java reserved keywords (int, class, for, etc.)
 *  4. Case-sensitive (age != Age != AGE)
 *
 *  NAMING CONVENTIONS:
 *  - Variables & Methods: camelCase     → myVariableName
 *  - Classes:            PascalCase    → MyClassName
 *  - Constants:          UPPER_SNAKE   → MAX_VALUE
 *  - Packages:           lowercase     → com.example.myapp
 *
 *  TWO CATEGORIES OF DATA TYPES:
 *  ==============================
 *
 *  1. PRIMITIVE TYPES (8 types) — stored directly in memory
 *     ┌──────────────┬───────┬─────────────────────────────┬───────────────────┐
 *     │ Type         │ Size  │ Range                       │ Default Value     │
 *     ├──────────────┼───────┼─────────────────────────────┼───────────────────┤
 *     │ byte         │ 1B    │ -128 to 127                 │ 0                 │
 *     │ short        │ 2B    │ -32,768 to 32,767           │ 0                 │
 *     │ int          │ 4B    │ -2^31 to 2^31-1             │ 0                 │
 *     │ long         │ 8B    │ -2^63 to 2^63-1             │ 0L                │
 *     │ float        │ 4B    │ ~7 decimal digits            │ 0.0f              │
 *     │ double       │ 8B    │ ~15 decimal digits           │ 0.0d              │
 *     │ char         │ 2B    │ 0 to 65,535 (Unicode)       │ '\u0000'          │
 *     │ boolean      │ 1bit  │ true or false               │ false             │
 *     └──────────────┴───────┴─────────────────────────────┴───────────────────┘
 *
 *  2. REFERENCE TYPES — store address (reference) to object in heap
 *     - String, Arrays, Classes, Interfaces, etc.
 *     - Default value: null
 *
 *  MEMORY MODEL:
 *  ==============
 *  - Primitives → stored in STACK (fast, fixed size)
 *  - Reference  → reference in STACK, object in HEAP
 *
 *     STACK                  HEAP
 *    ┌──────────┐          ┌──────────────────┐
 *    │ age = 25 │          │  String object:  │
 *    │ x = 3.14 │          │  "Hello"         │
 *    │ name ────┼────────> │                  │
 *    └──────────┘          └──────────────────┘
 *
 * ============================================================
 */

public class Chapter02_VariablesAndDataTypes {

    public static void main(String[] args) {

        // =====================================================
        //  1. PRIMITIVE DATA TYPES
        // =====================================================

        System.out.println("=== PRIMITIVE DATA TYPES ===\n");

        // --- INTEGERS ---

        // byte: 1 byte, range -128 to 127
        byte myByte = 127;
        // byte overflow = 128; // ERROR: incompatible types
        System.out.println("byte: " + myByte + " (1 byte, -128 to 127)");

        // short: 2 bytes, range -32768 to 32767
        short myShort = 32000;
        System.out.println("short: " + myShort + " (2 bytes)");

        // int: 4 bytes, most commonly used for integers
        int myInt = 2_000_000_000; // underscores for readability (Java 7+)
        System.out.println("int: " + myInt + " (4 bytes)");

        // long: 8 bytes, suffix L required
        long myLong = 9_223_372_036_854_775_807L; // must end with L
        System.out.println("long: " + myLong + " (8 bytes)");

        // --- FLOATING POINT ---

        // float: 4 bytes, ~7 decimal digits precision, suffix f required
        float myFloat = 3.14159f; // must end with f
        System.out.println("float: " + myFloat + " (4 bytes, ~7 digits)");

        // double: 8 bytes, ~15 decimal digits, default for decimals
        double myDouble = 3.141592653589793;
        System.out.println("double: " + myDouble + " (8 bytes, ~15 digits)");

        // --- CHARACTER ---

        // char: 2 bytes, single character in single quotes
        char myChar = 'A';
        char unicodeChar = '\u0041'; // Unicode for 'A'
        char numericChar = 65;       // ASCII value for 'A'
        System.out.println("char: " + myChar + ", unicode: " + unicodeChar + ", numeric: " + numericChar);

        // --- BOOLEAN ---

        // boolean: true or false only
        boolean isJavaFun = true;
        boolean isFishMammal = false;
        System.out.println("boolean: isJavaFun=" + isJavaFun + ", isFishMammal=" + isFishMammal);

        // =====================================================
        //  2. NUMBER LITERALS (Different Bases)
        // =====================================================

        System.out.println("\n=== NUMBER LITERALS ===\n");

        int decimal = 100;        // Base 10 (normal)
        int binary = 0b1100100;   // Base 2  (prefix 0b)
        int octal = 0144;         // Base 8  (prefix 0)
        int hex = 0x64;           // Base 16 (prefix 0x)

        System.out.println("Decimal: " + decimal);
        System.out.println("Binary 0b1100100: " + binary);
        System.out.println("Octal 0144: " + octal);
        System.out.println("Hex 0x64: " + hex);
        System.out.println("All represent the same value: " + (decimal == binary && binary == octal && octal == hex));

        // Scientific notation
        double sci1 = 1.5e3;  // 1.5 × 10^3 = 1500.0
        double sci2 = 2.5e-2; // 2.5 × 10^-2 = 0.025
        System.out.println("1.5e3 = " + sci1);
        System.out.println("2.5e-2 = " + sci2);

        // =====================================================
        //  3. TYPE CASTING (Converting between types)
        // =====================================================

        System.out.println("\n=== TYPE CASTING ===\n");

        // --- WIDENING (Implicit) — small to large, automatic, NO data loss
        // byte → short → int → long → float → double
        int smallNum = 100;
        double bigNum = smallNum; // int automatically becomes double
        System.out.println("Widening: int " + smallNum + " → double " + bigNum);

        byte b = 42;
        int i = b;      // byte → int (automatic)
        long l = i;     // int → long (automatic)
        float f = l;    // long → float (automatic)
        double d = f;   // float → double (automatic)
        System.out.println("Widening chain: byte " + b + " → int " + i + " → long " + l + " → float " + f + " → double " + d);

        // --- NARROWING (Explicit) — large to small, manual cast needed, POSSIBLE data loss
        double largeDouble = 9.99;
        int truncated = (int) largeDouble; // MUST use (type) cast
        System.out.println("Narrowing: double " + largeDouble + " → int " + truncated + " (decimal lost!)");

        int bigInt = 130;
        byte narrowByte = (byte) bigInt; // 130 overflows byte range!
        System.out.println("Narrowing overflow: int " + bigInt + " → byte " + narrowByte + " (overflow!)");

        // --- char ↔ int conversions
        char letter = 'A';
        int asciiValue = letter; // char → int (widening)
        System.out.println("char 'A' → int: " + asciiValue);

        int num = 66;
        char fromNum = (char) num; // int → char (narrowing)
        System.out.println("int 66 → char: " + fromNum);

        // =====================================================
        //  4. VARIABLES: Declaration, Initialization, Scope
        // =====================================================

        System.out.println("\n=== VARIABLE TYPES ===\n");

        // Declaration (no value yet)
        int declared;
        declared = 10; // Initialize later
        System.out.println("Declared then initialized: " + declared);

        // Declaration + Initialization (in one line)
        int combined = 20;
        System.out.println("Combined: " + combined);

        // Multiple declarations of same type
        int a = 1, c = 3, e = 5;
        System.out.println("Multiple: " + a + ", " + c + ", " + e);

        // CONSTANTS with 'final' keyword — value cannot change
        final double PI = 3.14159265358979;
        final int MAX_USERS = 1000;
        // PI = 3.14; // ERROR: cannot assign a value to final variable
        System.out.println("Constant PI: " + PI);
        System.out.println("Constant MAX_USERS: " + MAX_USERS);

        // var keyword (Java 10+) — type inferred by compiler
        // Note: Java 11 supports this
        var message = "Hello"; // compiler knows this is String
        var number = 42;       // compiler knows this is int
        var pi = 3.14;         // compiler knows this is double
        System.out.println("var message: " + message + " (type: String)");
        System.out.println("var number: " + number + " (type: int)");
        System.out.println("var pi: " + pi + " (type: double)");

        // =====================================================
        //  5. REFERENCE TYPES
        // =====================================================

        System.out.println("\n=== REFERENCE TYPES ===\n");

        // String — a sequence of characters (reference type, NOT primitive)
        String greeting = "Hello, Java!";
        System.out.println("String: " + greeting);
        System.out.println("String length: " + greeting.length());

        // null — means "no object" / "no reference"
        String nothing = null;
        System.out.println("null variable: " + nothing);
        // nothing.length(); // CRASH! NullPointerException!

        // Difference between == and .equals() for Strings
        String s1 = "Hello";
        String s2 = "Hello";
        String s3 = new String("Hello");

        System.out.println("s1 == s2: " + (s1 == s2));       // true (same pool reference)
        System.out.println("s1 == s3: " + (s1 == s3));       // false (different objects)
        System.out.println("s1.equals(s3): " + s1.equals(s3)); // true (same content)

        // =====================================================
        //  6. TYPE INFORMATION
        // =====================================================

        System.out.println("\n=== TYPE INFORMATION ===\n");

        // Getting min/max values of types
        System.out.println("Byte range: " + Byte.MIN_VALUE + " to " + Byte.MAX_VALUE);
        System.out.println("Short range: " + Short.MIN_VALUE + " to " + Short.MAX_VALUE);
        System.out.println("Int range: " + Integer.MIN_VALUE + " to " + Integer.MAX_VALUE);
        System.out.println("Long range: " + Long.MIN_VALUE + " to " + Long.MAX_VALUE);
        System.out.println("Float range: " + Float.MIN_VALUE + " to " + Float.MAX_VALUE);
        System.out.println("Double range: " + Double.MIN_VALUE + " to " + Double.MAX_VALUE);

        // Size in bytes
        System.out.println("\nByte size: " + Byte.SIZE / 8 + " byte(s)");
        System.out.println("Int size: " + Integer.SIZE / 8 + " bytes");
        System.out.println("Long size: " + Long.SIZE / 8 + " bytes");
        System.out.println("Double size: " + Double.SIZE / 8 + " bytes");

        // =====================================================
        //  7. COMMON PITFALLS
        // =====================================================

        System.out.println("\n=== COMMON PITFALLS ===\n");

        // Integer division — truncates, doesn't round
        int result = 7 / 2; // = 3, NOT 3.5!
        System.out.println("7 / 2 = " + result + " (integer division truncates!)");

        // Fix: cast to double first
        double correctResult = 7.0 / 2; // = 3.5
        System.out.println("7.0 / 2 = " + correctResult + " (correct!)");

        // Floating point precision issues
        System.out.println("0.1 + 0.2 = " + (0.1 + 0.2)); // 0.30000000000000004 !
        System.out.println("This is NOT a Java bug — it's IEEE 754 floating point.");

        // Integer overflow — wraps around silently!
        int maxInt = Integer.MAX_VALUE;
        System.out.println("Max int: " + maxInt);
        System.out.println("Max int + 1: " + (maxInt + 1) + " (overflow! wraps to negative)");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Declare variables for your name, age, height (in meters),
 *     and whether you're a student. Print them all.
 *
 *  2. Convert a temperature from Celsius to Fahrenheit:
 *     F = (C × 9/5) + 32
 *     Be careful with integer division!
 *
 *  3. What happens when you assign 200 to a byte variable?
 *     Try it and explain the output.
 *
 *  4. Declare two char variables with 'A' and 'a'.
 *     Print their integer (ASCII) values. What's the difference?
 *
 *  5. Create a final variable and try to reassign it.
 *     What error do you get?
 *
 *  6. Experiment with var keyword — what types does the compiler infer?
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 03 — Operators
 * ============================================================
 */
