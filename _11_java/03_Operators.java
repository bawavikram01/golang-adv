/*
 * ============================================================
 *  CHAPTER 03: OPERATORS
 * ============================================================
 *
 *  Operators are symbols that perform operations on operands.
 *
 *  CATEGORIES:
 *  1. Arithmetic      (+, -, *, /, %)
 *  2. Assignment       (=, +=, -=, *=, /=, %=, etc.)
 *  3. Comparison       (==, !=, >, <, >=, <=)
 *  4. Logical          (&&, ||, !)
 *  5. Bitwise          (&, |, ^, ~, <<, >>, >>>)
 *  6. Ternary          (? :)
 *  7. Instanceof       (instanceof)
 *  8. Unary            (++, --, +, -, ~, !)
 *
 *  OPERATOR PRECEDENCE (highest to lowest):
 *  ─────────────────────────────────────────
 *  1.  () [] .                  → Parentheses, array, member access
 *  2.  ++ -- + - ~ ! (type)    → Unary, cast
 *  3.  * / %                    → Multiplicative
 *  4.  + -                      → Additive
 *  5.  << >> >>>                → Shift
 *  6.  < > <= >= instanceof     → Relational
 *  7.  == !=                    → Equality
 *  8.  &                        → Bitwise AND
 *  9.  ^                        → Bitwise XOR
 *  10. |                        → Bitwise OR
 *  11. &&                       → Logical AND
 *  12. ||                       → Logical OR
 *  13. ? :                      → Ternary
 *  14. = += -= *= /= etc.      → Assignment
 *
 *  TIP: When in doubt, use parentheses () to make intent clear!
 *
 * ============================================================
 */

public class Chapter03_Operators {

    public static void main(String[] args) {

        // =====================================================
        //  1. ARITHMETIC OPERATORS
        // =====================================================

        System.out.println("=== ARITHMETIC OPERATORS ===\n");

        int a = 17, b = 5;

        System.out.println("a = " + a + ", b = " + b);
        System.out.println("a + b = " + (a + b));   // 22  (addition)
        System.out.println("a - b = " + (a - b));   // 12  (subtraction)
        System.out.println("a * b = " + (a * b));   // 85  (multiplication)
        System.out.println("a / b = " + (a / b));   // 3   (integer division — truncates!)
        System.out.println("a % b = " + (a % b));   // 2   (modulus — remainder)

        // Division with doubles gives correct result
        System.out.println("17.0 / 5 = " + (17.0 / 5));   // 3.4

        // Modulus is useful for:
        System.out.println("\n--- Modulus Use Cases ---");
        System.out.println("17 is " + (17 % 2 == 0 ? "even" : "odd"));  // odd
        System.out.println("20 is " + (20 % 2 == 0 ? "even" : "odd"));  // even
        System.out.println("Clock: 15:00 in 12h = " + (15 % 12) + ":00"); // 3

        // String concatenation with +
        System.out.println("\n--- String + Operator ---");
        String first = "Hello";
        String second = "World";
        System.out.println(first + " " + second); // "Hello World"
        System.out.println("Age: " + 25);          // int is converted to String
        System.out.println(1 + 2 + " hello");      // "3 hello" (1+2 first, then concat)
        System.out.println("hello " + 1 + 2);      // "hello 12" (concat from left)

        // =====================================================
        //  2. UNARY OPERATORS (++, --, +, -, !)
        // =====================================================

        System.out.println("\n=== UNARY OPERATORS ===\n");

        int x = 10;

        // Pre-increment: increment FIRST, then use the value
        System.out.println("x = " + x);                    // 10
        System.out.println("++x = " + (++x));              // 11 (increment, then return)
        System.out.println("After ++x, x = " + x);         // 11

        // Post-increment: use the value FIRST, then increment
        x = 10;
        System.out.println("\nx = " + x);                   // 10
        System.out.println("x++ = " + (x++));              // 10 (return, then increment)
        System.out.println("After x++, x = " + x);         // 11

        // Pre-decrement and Post-decrement work the same way
        x = 10;
        System.out.println("\n--x = " + (--x));            // 9
        x = 10;
        System.out.println("x-- = " + (x--));              // 10 (returns old, then decrements)
        System.out.println("After x--, x = " + x);         // 9

        // Tricky example
        int m = 5;
        int n = m++ + ++m; // m++ returns 5 (m becomes 6), ++m returns 7 (m becomes 7)
        System.out.println("\nm = 5; n = m++ + ++m;");
        System.out.println("n = " + n);  // 5 + 7 = 12
        System.out.println("m = " + m);  // 7

        // Unary + and -
        int positive = +5;   // positive 5
        int negative = -5;   // negative 5
        int negated = -positive; // -5
        System.out.println("\n+5 = " + positive + ", -5 = " + negative + ", -(+5) = " + negated);

        // =====================================================
        //  3. ASSIGNMENT OPERATORS
        // =====================================================

        System.out.println("\n=== ASSIGNMENT OPERATORS ===\n");

        int val = 100;
        System.out.println("val = " + val);

        val += 10;  // val = val + 10
        System.out.println("val += 10 → " + val);  // 110

        val -= 20;  // val = val - 20
        System.out.println("val -= 20 → " + val);  // 90

        val *= 2;   // val = val * 2
        System.out.println("val *= 2  → " + val);  // 180

        val /= 3;   // val = val / 3
        System.out.println("val /= 3  → " + val);  // 60

        val %= 7;   // val = val % 7
        System.out.println("val %%= 7  → " + val); // 4

        // Compound assignment does implicit casting!
        byte bb = 10;
        bb += 5;    // OK! Equivalent to bb = (byte)(bb + 5)
        // bb = bb + 5; // ERROR! bb + 5 is int, can't assign to byte without cast
        System.out.println("byte compound assignment: " + bb);

        // =====================================================
        //  4. COMPARISON (RELATIONAL) OPERATORS
        // =====================================================

        System.out.println("\n=== COMPARISON OPERATORS ===\n");

        int p = 10, q = 20;
        System.out.println("p = " + p + ", q = " + q);
        System.out.println("p == q: " + (p == q));  // false
        System.out.println("p != q: " + (p != q));  // true
        System.out.println("p > q:  " + (p > q));   // false
        System.out.println("p < q:  " + (p < q));   // true
        System.out.println("p >= q: " + (p >= q));  // false
        System.out.println("p <= q: " + (p <= q));  // true

        // WARNING: == compares REFERENCES for objects, not values!
        String s1 = new String("hello");
        String s2 = new String("hello");
        System.out.println("\ns1 == s2: " + (s1 == s2));        // false (different objects!)
        System.out.println("s1.equals(s2): " + s1.equals(s2));  // true (same content)

        // =====================================================
        //  5. LOGICAL OPERATORS
        // =====================================================

        System.out.println("\n=== LOGICAL OPERATORS ===\n");

        boolean t = true, f = false;

        // && (AND) — both must be true
        System.out.println("true && true:  " + (t && t));   // true
        System.out.println("true && false: " + (t && f));   // false
        System.out.println("false && false:" + (f && f));    // false

        // || (OR) — at least one must be true
        System.out.println("\ntrue || true:  " + (t || t));  // true
        System.out.println("true || false: " + (t || f));    // true
        System.out.println("false || false:" + (f || f));    // false

        // ! (NOT) — inverts
        System.out.println("\n!true:  " + (!t));              // false
        System.out.println("!false: " + (!f));                // true

        // SHORT-CIRCUIT EVALUATION
        // && stops if first operand is false (doesn't evaluate second)
        // || stops if first operand is true (doesn't evaluate second)
        System.out.println("\n--- Short-Circuit ---");
        int sc = 5;
        boolean result = (sc > 10) && (++sc > 5); // ++sc is NEVER executed!
        System.out.println("sc = " + sc + " (still 5, short-circuit skipped ++sc)");

        // & and | (non-short-circuit) — ALWAYS evaluate both sides
        sc = 5;
        result = (sc > 10) & (++sc > 5); // ++sc IS executed!
        System.out.println("sc = " + sc + " (now 6, non-short-circuit evaluated ++sc)");

        // Practical example
        int age = 25;
        double salary = 50000;
        boolean hasLicense = true;
        boolean canRentCar = (age >= 21) && (salary > 30000) && hasLicense;
        System.out.println("\nCan rent car? " + canRentCar);

        // =====================================================
        //  6. BITWISE OPERATORS
        // =====================================================

        System.out.println("\n=== BITWISE OPERATORS ===\n");

        int bit1 = 0b1010;  // 10 in binary
        int bit2 = 0b1100;  // 12 in binary

        System.out.println("bit1 = 1010 (10)");
        System.out.println("bit2 = 1100 (12)");

        // AND (&): 1 only if both bits are 1
        System.out.println("bit1 & bit2 = " + Integer.toBinaryString(bit1 & bit2)
                + " (" + (bit1 & bit2) + ")");  // 1000 (8)

        // OR (|): 1 if either bit is 1
        System.out.println("bit1 | bit2 = " + Integer.toBinaryString(bit1 | bit2)
                + " (" + (bit1 | bit2) + ")");  // 1110 (14)

        // XOR (^): 1 if bits are different
        System.out.println("bit1 ^ bit2 = " + Integer.toBinaryString(bit1 ^ bit2)
                + " (" + (bit1 ^ bit2) + ")");  // 0110 (6)

        // NOT (~): flips all bits
        System.out.println("~bit1 = " + (~bit1) + " (flips all 32 bits)");  // -11

        // Left shift (<<): multiply by 2^n
        System.out.println("\n--- Bit Shifting ---");
        System.out.println("5 << 1 = " + (5 << 1));  // 10  (5 * 2)
        System.out.println("5 << 2 = " + (5 << 2));  // 20  (5 * 4)
        System.out.println("5 << 3 = " + (5 << 3));  // 40  (5 * 8)

        // Right shift (>>): divide by 2^n (preserves sign)
        System.out.println("40 >> 1 = " + (40 >> 1));  // 20  (40 / 2)
        System.out.println("40 >> 2 = " + (40 >> 2));  // 10  (40 / 4)
        System.out.println("-8 >> 1 = " + (-8 >> 1));  // -4  (sign preserved)

        // Unsigned right shift (>>>): fills with 0 (no sign preservation)
        System.out.println("-8 >>> 1 = " + (-8 >>> 1)); // large positive number

        // Practical: Swap two numbers without temp variable using XOR
        System.out.println("\n--- XOR Swap ---");
        int sw1 = 10, sw2 = 20;
        System.out.println("Before: sw1=" + sw1 + ", sw2=" + sw2);
        sw1 = sw1 ^ sw2;
        sw2 = sw1 ^ sw2;
        sw1 = sw1 ^ sw2;
        System.out.println("After:  sw1=" + sw1 + ", sw2=" + sw2);

        // Practical: Check if number is even/odd using AND
        System.out.println("\n--- Even/Odd with & ---");
        for (int i = 1; i <= 6; i++) {
            System.out.println(i + " is " + ((i & 1) == 0 ? "even" : "odd"));
        }

        // =====================================================
        //  7. TERNARY OPERATOR
        // =====================================================

        System.out.println("\n=== TERNARY OPERATOR ===\n");

        // Syntax: condition ? valueIfTrue : valueIfFalse
        int num = 15;
        String parity = (num % 2 == 0) ? "even" : "odd";
        System.out.println(num + " is " + parity);

        // Nested ternary (use sparingly — hard to read)
        int score = 85;
        String grade = (score >= 90) ? "A"
                     : (score >= 80) ? "B"
                     : (score >= 70) ? "C"
                     : (score >= 60) ? "D"
                     : "F";
        System.out.println("Score " + score + " = Grade " + grade);

        // Find max of two numbers
        int max = (a > b) ? a : b;
        System.out.println("Max of " + a + " and " + b + " = " + max);

        // Find max of three numbers
        int c = 8;
        int maxOfThree = (a > b) ? (a > c ? a : c) : (b > c ? b : c);
        System.out.println("Max of " + a + ", " + b + ", " + c + " = " + maxOfThree);

        // =====================================================
        //  8. INSTANCEOF OPERATOR
        // =====================================================

        System.out.println("\n=== INSTANCEOF OPERATOR ===\n");

        // Checks if an object is an instance of a class
        String str = "Hello";
        System.out.println("str instanceof String: " + (str instanceof String));  // true
        System.out.println("str instanceof Object: " + (str instanceof Object));  // true

        Object obj = "I'm a String stored as Object";
        System.out.println("obj instanceof String: " + (obj instanceof String));  // true
        System.out.println("obj instanceof Integer: " + (obj instanceof Integer)); // false

        // null is NOT an instance of anything
        String nullStr = null;
        System.out.println("null instanceof String: " + (nullStr instanceof String)); // false

        // =====================================================
        //  9. OPERATOR PRECEDENCE IN ACTION
        // =====================================================

        System.out.println("\n=== OPERATOR PRECEDENCE ===\n");

        // Without parentheses — follows precedence
        int expr1 = 2 + 3 * 4;      // 3*4 first → 2 + 12 = 14
        System.out.println("2 + 3 * 4 = " + expr1);

        // With parentheses — override precedence
        int expr2 = (2 + 3) * 4;    // 2+3 first → 5 * 4 = 20
        System.out.println("(2 + 3) * 4 = " + expr2);

        // Complex example
        int expr3 = 10 + 20 * 30 / 5 - 15;  // 20*30=600, 600/5=120, 10+120=130, 130-15=115
        System.out.println("10 + 20 * 30 / 5 - 15 = " + expr3);

        // Boolean with mixed operators
        boolean expr4 = 5 > 3 && 10 < 20 || false;  // (5>3)=T && (10<20)=T → T || F → T
        System.out.println("5 > 3 && 10 < 20 || false = " + expr4);

        // TIP: Always use parentheses to make intent clear!
        boolean clear = ((5 > 3) && (10 < 20)) || false;
        System.out.println("Same with parentheses: " + clear);
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Calculate the area and perimeter of a rectangle (l=15, w=8)
 *     using arithmetic operators.
 *
 *  2. Swap two integers using:
 *     a) A temporary variable
 *     b) Arithmetic (a = a+b, b = a-b, a = a-b)
 *     c) XOR (shown above — do it yourself)
 *
 *  3. Write expressions to check if a number is:
 *     a) Between 1 and 100 (inclusive)
 *     b) Divisible by both 3 and 5
 *     c) A positive even number
 *
 *  4. What is the output?
 *     int a = 5, b = 10;
 *     System.out.println(a++ + ++b + a);
 *     (Work it out on paper first, then verify!)
 *
 *  5. Use bit shifting to multiply 7 by 8 without using *.
 *
 *  6. Use the ternary operator to find the absolute value of -42.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 04 — Control Flow
 * ============================================================
 */
