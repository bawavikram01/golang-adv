/*
 * ============================================================
 *  CHAPTER 04: CONTROL FLOW
 * ============================================================
 *
 *  Control flow determines the ORDER in which statements execute.
 *
 *  THREE TYPES:
 *  1. SELECTION    — if, if-else, if-else-if, switch
 *  2. ITERATION    — for, while, do-while, for-each
 *  3. BRANCHING    — break, continue, return, labeled breaks
 *
 * ============================================================
 */

public class Chapter04_ControlFlow {

    public static void main(String[] args) {

        // =====================================================
        //  1. IF STATEMENT
        // =====================================================

        System.out.println("=== IF STATEMENT ===\n");

        int temperature = 35;

        // Simple if
        if (temperature > 30) {
            System.out.println("It's hot! (" + temperature + "°C)");
        }

        // if-else
        int age = 17;
        if (age >= 18) {
            System.out.println("You can vote.");
        } else {
            System.out.println("You cannot vote yet. Wait " + (18 - age) + " more year(s).");
        }

        // if-else-if ladder
        int score = 78;
        System.out.println("\nScore: " + score);
        if (score >= 90) {
            System.out.println("Grade: A");
        } else if (score >= 80) {
            System.out.println("Grade: B");
        } else if (score >= 70) {
            System.out.println("Grade: C");
        } else if (score >= 60) {
            System.out.println("Grade: D");
        } else {
            System.out.println("Grade: F");
        }

        // Nested if
        boolean hasTicket = true;
        int passengerAge = 12;
        System.out.println("\n--- Nested If ---");
        if (hasTicket) {
            if (passengerAge < 5) {
                System.out.println("Free entry!");
            } else if (passengerAge < 18) {
                System.out.println("Child ticket: $5");
            } else {
                System.out.println("Adult ticket: $10");
            }
        } else {
            System.out.println("Please buy a ticket first.");
        }

        // Single-line if (no braces) — works but NOT recommended
        if (temperature > 30) System.out.println("\nStill hot (single-line if)");

        // =====================================================
        //  2. SWITCH STATEMENT
        // =====================================================

        System.out.println("\n=== SWITCH STATEMENT ===\n");

        // Classic switch — works with: byte, short, int, char, String, enum
        int dayNum = 3;
        String day;
        switch (dayNum) {
            case 1:
                day = "Monday";
                break;      // break prevents "fall-through"
            case 2:
                day = "Tuesday";
                break;
            case 3:
                day = "Wednesday";
                break;
            case 4:
                day = "Thursday";
                break;
            case 5:
                day = "Friday";
                break;
            case 6:
                day = "Saturday";
                break;
            case 7:
                day = "Sunday";
                break;
            default:        // executed when no case matches
                day = "Invalid day";
                break;
        }
        System.out.println("Day " + dayNum + " = " + day);

        // Switch with String (Java 7+)
        String command = "start";
        switch (command.toLowerCase()) {
            case "start":
                System.out.println("Starting the system...");
                break;
            case "stop":
                System.out.println("Stopping the system...");
                break;
            case "restart":
                System.out.println("Restarting the system...");
                break;
            default:
                System.out.println("Unknown command: " + command);
        }

        // Fall-through behavior (intentional)
        System.out.println("\n--- Fall-Through ---");
        int month = 8;
        String season;
        switch (month) {
            case 12: case 1: case 2:
                season = "Winter";
                break;
            case 3: case 4: case 5:
                season = "Spring";
                break;
            case 6: case 7: case 8:
                season = "Summer";
                break;
            case 9: case 10: case 11:
                season = "Autumn";
                break;
            default:
                season = "Invalid month";
        }
        System.out.println("Month " + month + " = " + season);

        // =====================================================
        //  3. FOR LOOP
        // =====================================================

        System.out.println("\n=== FOR LOOP ===\n");

        // Basic for loop: for (init; condition; update) { body }
        System.out.println("--- Count 1 to 5 ---");
        for (int i = 1; i <= 5; i++) {
            System.out.print(i + " ");
        }
        System.out.println();

        // Count backwards
        System.out.println("--- Count 5 to 1 ---");
        for (int i = 5; i >= 1; i--) {
            System.out.print(i + " ");
        }
        System.out.println();

        // Step by 2
        System.out.println("--- Even numbers 2-20 ---");
        for (int i = 2; i <= 20; i += 2) {
            System.out.print(i + " ");
        }
        System.out.println();

        // Multiple variables in for loop
        System.out.println("--- Multiple variables ---");
        for (int i = 0, j = 10; i < j; i++, j--) {
            System.out.println("i=" + i + ", j=" + j);
        }

        // Infinite loop (careful!)
        // for (;;) { } // runs forever — must have break inside

        // =====================================================
        //  4. WHILE LOOP
        // =====================================================

        System.out.println("\n=== WHILE LOOP ===\n");

        // while: checks condition BEFORE each iteration
        int count = 1;
        System.out.println("--- Count to 5 ---");
        while (count <= 5) {
            System.out.print(count + " ");
            count++;
        }
        System.out.println();

        // Sum of digits
        int number = 12345;
        int sum = 0;
        int temp = number;
        while (temp > 0) {
            sum += temp % 10;    // get last digit
            temp /= 10;         // remove last digit
        }
        System.out.println("Sum of digits of " + number + " = " + sum);

        // Reverse a number
        int original = 12345;
        int reversed = 0;
        temp = original;
        while (temp > 0) {
            reversed = reversed * 10 + temp % 10;
            temp /= 10;
        }
        System.out.println("Reverse of " + original + " = " + reversed);

        // =====================================================
        //  5. DO-WHILE LOOP
        // =====================================================

        System.out.println("\n=== DO-WHILE LOOP ===\n");

        // do-while: executes body AT LEAST ONCE, then checks condition
        int x = 1;
        do {
            System.out.print(x + " ");
            x++;
        } while (x <= 5);
        System.out.println();

        // Key difference: do-while runs at least once even if condition is false
        int y = 100;
        do {
            System.out.println("This prints ONCE even though y > 5. y = " + y);
        } while (y < 5);

        // =====================================================
        //  6. FOR-EACH (Enhanced For) LOOP
        // =====================================================

        System.out.println("\n=== FOR-EACH LOOP ===\n");

        // Used to iterate over arrays and collections
        int[] numbers = {10, 20, 30, 40, 50};

        System.out.println("--- Iterating array ---");
        for (int num : numbers) {
            System.out.print(num + " ");
        }
        System.out.println();

        String[] fruits = {"Apple", "Banana", "Cherry", "Date"};
        for (String fruit : fruits) {
            System.out.println("Fruit: " + fruit);
        }

        // NOTE: for-each cannot modify the array or access the index
        // Use regular for loop when you need the index

        // =====================================================
        //  7. BREAK, CONTINUE, RETURN
        // =====================================================

        System.out.println("\n=== BREAK & CONTINUE ===\n");

        // break: exits the loop immediately
        System.out.println("--- Break ---");
        for (int i = 1; i <= 10; i++) {
            if (i == 6) {
                System.out.println("Breaking at " + i);
                break; // exits the for loop
            }
            System.out.print(i + " ");
        }
        System.out.println();

        // continue: skips current iteration, goes to next
        System.out.println("\n--- Continue ---");
        for (int i = 1; i <= 10; i++) {
            if (i % 3 == 0) {
                continue; // skip multiples of 3
            }
            System.out.print(i + " ");
        }
        System.out.println(" (skipped 3, 6, 9)");

        // =====================================================
        //  8. LABELED BREAKS (for nested loops)
        // =====================================================

        System.out.println("\n=== LABELED BREAKS ===\n");

        // Without label — break only exits inner loop
        System.out.println("--- Without label ---");
        for (int i = 0; i < 3; i++) {
            for (int j = 0; j < 3; j++) {
                if (j == 2) break; // only breaks inner loop
                System.out.print("(" + i + "," + j + ") ");
            }
            System.out.println();
        }

        // With label — break exits the labeled (outer) loop
        System.out.println("--- With label ---");
        outerLoop:
        for (int i = 0; i < 3; i++) {
            for (int j = 0; j < 3; j++) {
                if (i == 1 && j == 1) {
                    System.out.println("\nBreaking outer loop at (" + i + "," + j + ")");
                    break outerLoop; // breaks the OUTER loop
                }
                System.out.print("(" + i + "," + j + ") ");
            }
        }

        // Labeled continue
        System.out.println("\n--- Labeled continue ---");
        outer:
        for (int i = 0; i < 3; i++) {
            for (int j = 0; j < 3; j++) {
                if (j == 1) continue outer; // skip rest of inner AND go to next outer iteration
                System.out.print("(" + i + "," + j + ") ");
            }
        }
        System.out.println();

        // =====================================================
        //  9. NESTED LOOPS — PATTERNS
        // =====================================================

        System.out.println("\n=== NESTED LOOP PATTERNS ===\n");

        int rows = 5;

        // Right triangle
        System.out.println("--- Right Triangle ---");
        for (int i = 1; i <= rows; i++) {
            for (int j = 1; j <= i; j++) {
                System.out.print("* ");
            }
            System.out.println();
        }

        // Inverted triangle
        System.out.println("--- Inverted Triangle ---");
        for (int i = rows; i >= 1; i--) {
            for (int j = 1; j <= i; j++) {
                System.out.print("* ");
            }
            System.out.println();
        }

        // Number pyramid
        System.out.println("--- Number Pyramid ---");
        for (int i = 1; i <= rows; i++) {
            // Print spaces
            for (int s = rows - i; s > 0; s--) {
                System.out.print(" ");
            }
            // Print numbers
            for (int j = 1; j <= i; j++) {
                System.out.print(j + " ");
            }
            System.out.println();
        }

        // Multiplication table
        System.out.println("--- Multiplication Table (1-5) ---");
        for (int i = 1; i <= 5; i++) {
            for (int j = 1; j <= 5; j++) {
                System.out.printf("%4d", i * j);
            }
            System.out.println();
        }

        // =====================================================
        //  10. PRACTICAL EXAMPLES
        // =====================================================

        System.out.println("\n=== PRACTICAL EXAMPLES ===\n");

        // Check if a number is prime
        int checkNum = 29;
        boolean isPrime = true;
        if (checkNum <= 1) {
            isPrime = false;
        } else {
            for (int i = 2; i <= Math.sqrt(checkNum); i++) {
                if (checkNum % i == 0) {
                    isPrime = false;
                    break;
                }
            }
        }
        System.out.println(checkNum + " is " + (isPrime ? "prime" : "not prime"));

        // Fibonacci sequence
        System.out.print("Fibonacci (first 10): ");
        int fib1 = 0, fib2 = 1;
        for (int i = 0; i < 10; i++) {
            System.out.print(fib1 + " ");
            int next = fib1 + fib2;
            fib1 = fib2;
            fib2 = next;
        }
        System.out.println();

        // Factorial
        int n = 10;
        long factorial = 1;
        for (int i = 1; i <= n; i++) {
            factorial *= i;
        }
        System.out.println(n + "! = " + factorial);

        // Find GCD using Euclidean algorithm
        int gcdA = 56, gcdB = 42;
        int tempA = gcdA, tempB = gcdB;
        while (tempB != 0) {
            int remainder = tempA % tempB;
            tempA = tempB;
            tempB = remainder;
        }
        System.out.println("GCD of " + gcdA + " and " + gcdB + " = " + tempA);

        // Print all prime numbers up to N
        int limit = 50;
        System.out.print("Primes up to " + limit + ": ");
        for (int i = 2; i <= limit; i++) {
            boolean prime = true;
            for (int j = 2; j <= Math.sqrt(i); j++) {
                if (i % j == 0) {
                    prime = false;
                    break;
                }
            }
            if (prime) System.out.print(i + " ");
        }
        System.out.println();
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Write a program that prints "FizzBuzz":
 *     - For numbers 1-100
 *     - Print "Fizz" if divisible by 3
 *     - Print "Buzz" if divisible by 5
 *     - Print "FizzBuzz" if divisible by both
 *     - Otherwise print the number
 *
 *  2. Print a diamond pattern of stars with 5 rows.
 *
 *  3. Check if a number is a palindrome (reads same forwards/backwards).
 *
 *  4. Print all Armstrong numbers between 1 and 1000.
 *     (Armstrong: sum of cubes of digits equals the number)
 *     Example: 153 = 1³ + 5³ + 3³
 *
 *  5. Create a simple calculator using switch:
 *     Given two numbers and an operator (+, -, *, /), compute result.
 *
 *  6. Print Pascal's Triangle with 6 rows.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 05 — Methods
 * ============================================================
 */
