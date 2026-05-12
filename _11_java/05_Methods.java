/*
 * ============================================================
 *  CHAPTER 05: METHODS (FUNCTIONS)
 * ============================================================
 *
 *  A METHOD is a block of reusable code that performs a specific task.
 *
 *  WHY METHODS?
 *  1. Reusability     — write once, call many times
 *  2. Modularity      — break complex problems into smaller pieces
 *  3. Readability     — give meaningful names to blocks of code
 *  4. Maintainability — fix in one place, fixed everywhere
 *
 *  METHOD SYNTAX:
 *  ──────────────
 *  accessModifier returnType methodName(parameterList) {
 *      // method body
 *      return value; // if returnType is not void
 *  }
 *
 *  COMPONENTS:
 *  - Access modifier: public, private, protected, (default)
 *  - Return type:     int, String, void, double, etc.
 *  - Method name:     camelCase convention
 *  - Parameters:      input values (optional)
 *  - Method body:     the code that runs
 *  - return:          sends value back to caller
 *
 * ============================================================
 */

public class Chapter05_Methods {

    // =====================================================
    //  1. BASIC METHODS
    // =====================================================

    // Method with no parameters and no return value
    static void greet() {
        System.out.println("Hello! Welcome to Java Methods!");
    }

    // Method with parameters
    static void greetUser(String name) {
        System.out.println("Hello, " + name + "!");
    }

    // Method with return value
    static int add(int a, int b) {
        return a + b; // returns the sum
    }

    // Method with multiple parameters and return
    static double calculateArea(double length, double width) {
        return length * width;
    }

    // =====================================================
    //  2. METHOD OVERLOADING
    // =====================================================

    // Same name, DIFFERENT parameters (type, number, or order)
    // This is COMPILE-TIME polymorphism

    static int multiply(int a, int b) {
        return a * b;
    }

    static double multiply(double a, double b) {
        return a * b;
    }

    static int multiply(int a, int b, int c) {
        return a * b * c;
    }

    // Overloading with different parameter ORDER
    static String format(String text, int number) {
        return text + ": " + number;
    }

    static String format(int number, String text) {
        return number + " - " + text;
    }

    // NOTE: Overloading based on RETURN TYPE ALONE is NOT allowed!
    // static double add(int a, int b) { } // ERROR! Same params as add(int, int)

    // =====================================================
    //  3. PASS BY VALUE
    // =====================================================

    // Java is ALWAYS pass-by-value!
    // Primitives: the VALUE is copied
    // Objects: the REFERENCE (address) is copied, NOT the object

    static void tryToChange(int num) {
        num = 999; // changes local copy only
    }

    static void tryToChangeArray(int[] arr) {
        arr[0] = 999; // modifies the ACTUAL array (reference is copied)
    }

    static void tryToReassignArray(int[] arr) {
        arr = new int[]{999, 888}; // reassigns local reference, original unchanged
    }

    // =====================================================
    //  4. RECURSION
    // =====================================================

    // A method that calls itself
    // Every recursive method needs:
    // 1. BASE CASE     — when to stop
    // 2. RECURSIVE CASE — calls itself with a smaller problem

    // Factorial: n! = n × (n-1) × ... × 1
    static long factorial(int n) {
        if (n <= 1) return 1;        // base case
        return n * factorial(n - 1);  // recursive case
    }
    // factorial(5) → 5 * factorial(4)
    //              → 5 * 4 * factorial(3)
    //              → 5 * 4 * 3 * factorial(2)
    //              → 5 * 4 * 3 * 2 * factorial(1)
    //              → 5 * 4 * 3 * 2 * 1
    //              → 120

    // Fibonacci: 0, 1, 1, 2, 3, 5, 8, 13, ...
    static int fibonacci(int n) {
        if (n <= 0) return 0;        // base case 1
        if (n == 1) return 1;        // base case 2
        return fibonacci(n - 1) + fibonacci(n - 2); // recursive case
    }

    // Power: base^exponent
    static long power(int base, int exponent) {
        if (exponent == 0) return 1;
        return base * power(base, exponent - 1);
    }

    // Sum of digits recursively
    static int sumOfDigits(int n) {
        if (n == 0) return 0;
        return (n % 10) + sumOfDigits(n / 10);
    }

    // Binary search (recursive)
    static int binarySearch(int[] arr, int target, int low, int high) {
        if (low > high) return -1; // not found

        int mid = low + (high - low) / 2; // avoids integer overflow

        if (arr[mid] == target) return mid;
        else if (arr[mid] < target) return binarySearch(arr, target, mid + 1, high);
        else return binarySearch(arr, target, low, mid - 1);
    }

    // =====================================================
    //  5. VARARGS (Variable Arguments)
    // =====================================================

    // Accepts zero or more arguments of the same type
    // Internally treated as an array
    // MUST be the LAST parameter

    static int sum(int... numbers) {
        int total = 0;
        for (int n : numbers) {
            total += n;
        }
        return total;
    }

    // Varargs with other parameters (varargs MUST be last)
    static void printInfo(String label, int... values) {
        System.out.print(label + ": ");
        for (int v : values) {
            System.out.print(v + " ");
        }
        System.out.println();
    }

    // =====================================================
    //  6. STATIC vs INSTANCE METHODS
    // =====================================================

    // Static method: belongs to the CLASS, called with ClassName.method()
    // No need to create an object
    static int square(int n) {
        return n * n;
    }

    // Instance method: belongs to an OBJECT, needs an instance to call
    // (We'll cover this more in OOP chapter)
    int cube(int n) {
        return n * n * n;
    }

    // =====================================================
    //  7. HELPER / UTILITY METHODS
    // =====================================================

    static boolean isPrime(int n) {
        if (n <= 1) return false;
        if (n <= 3) return true;
        if (n % 2 == 0 || n % 3 == 0) return false;
        for (int i = 5; i * i <= n; i += 6) {
            if (n % i == 0 || n % (i + 2) == 0) return false;
        }
        return true;
    }

    static boolean isPalindrome(String str) {
        int left = 0, right = str.length() - 1;
        while (left < right) {
            if (str.charAt(left) != str.charAt(right)) return false;
            left++;
            right--;
        }
        return true;
    }

    static int gcd(int a, int b) {
        while (b != 0) {
            int temp = b;
            b = a % b;
            a = temp;
        }
        return a;
    }

    static int lcm(int a, int b) {
        return (a / gcd(a, b)) * b; // avoids overflow
    }

    static void swap(int[] arr, int i, int j) {
        int temp = arr[i];
        arr[i] = arr[j];
        arr[j] = temp;
    }

    // =====================================================
    //  MAIN METHOD — Testing everything
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Basic Methods ---
        System.out.println("=== BASIC METHODS ===\n");
        greet();
        greetUser("Vikram");
        System.out.println("add(10, 20) = " + add(10, 20));
        System.out.println("area(5.5, 3.0) = " + calculateArea(5.5, 3.0));

        // --- 2. Method Overloading ---
        System.out.println("\n=== METHOD OVERLOADING ===\n");
        System.out.println("multiply(3, 4) = " + multiply(3, 4));          // int version
        System.out.println("multiply(2.5, 3.0) = " + multiply(2.5, 3.0)); // double version
        System.out.println("multiply(2, 3, 4) = " + multiply(2, 3, 4));   // 3-param version
        System.out.println("format(\"Score\", 100) = " + format("Score", 100));
        System.out.println("format(100, \"Score\") = " + format(100, "Score"));

        // --- 3. Pass by Value ---
        System.out.println("\n=== PASS BY VALUE ===\n");

        int num = 10;
        tryToChange(num);
        System.out.println("After tryToChange(10): num = " + num); // still 10

        int[] arr = {1, 2, 3};
        tryToChangeArray(arr);
        System.out.println("After tryToChangeArray: arr[0] = " + arr[0]); // 999 (modified!)

        arr = new int[]{1, 2, 3};
        tryToReassignArray(arr);
        System.out.println("After tryToReassignArray: arr[0] = " + arr[0]); // 1 (unchanged!)

        System.out.println("\nKey insight:");
        System.out.println("- Primitives: value is copied → original unchanged");
        System.out.println("- Objects: reference is copied → can modify contents");
        System.out.println("- But reassigning the reference doesn't affect original");

        // --- 4. Recursion ---
        System.out.println("\n=== RECURSION ===\n");
        System.out.println("factorial(5) = " + factorial(5));     // 120
        System.out.println("factorial(10) = " + factorial(10));   // 3628800

        System.out.print("Fibonacci(0-9): ");
        for (int i = 0; i < 10; i++) {
            System.out.print(fibonacci(i) + " ");
        }
        System.out.println();

        System.out.println("power(2, 10) = " + power(2, 10));     // 1024
        System.out.println("sumOfDigits(12345) = " + sumOfDigits(12345)); // 15

        int[] sorted = {2, 5, 8, 12, 16, 23, 38, 56, 72, 91};
        int idx = binarySearch(sorted, 23, 0, sorted.length - 1);
        System.out.println("binarySearch for 23: index = " + idx); // 5

        // --- 5. Varargs ---
        System.out.println("\n=== VARARGS ===\n");
        System.out.println("sum() = " + sum());               // 0
        System.out.println("sum(5) = " + sum(5));             // 5
        System.out.println("sum(1,2,3) = " + sum(1, 2, 3));   // 6
        System.out.println("sum(1,2,3,4,5) = " + sum(1, 2, 3, 4, 5)); // 15

        printInfo("Scores", 90, 85, 78, 92);
        printInfo("Empty");

        // --- 6. Static vs Instance ---
        System.out.println("\n=== STATIC vs INSTANCE ===\n");
        System.out.println("Static: square(5) = " + square(5)); // no object needed

        Chapter05_Methods obj = new Chapter05_Methods();
        System.out.println("Instance: cube(3) = " + obj.cube(3)); // needs object

        // --- 7. Utility Methods ---
        System.out.println("\n=== UTILITY METHODS ===\n");
        System.out.println("isPrime(17) = " + isPrime(17));       // true
        System.out.println("isPrime(20) = " + isPrime(20));       // false
        System.out.println("isPalindrome(\"racecar\") = " + isPalindrome("racecar")); // true
        System.out.println("isPalindrome(\"hello\") = " + isPalindrome("hello"));     // false
        System.out.println("gcd(48, 18) = " + gcd(48, 18));       // 6
        System.out.println("lcm(12, 18) = " + lcm(12, 18));       // 36

        // --- 8. Method Call Stack ---
        System.out.println("\n=== METHOD CALL STACK ===\n");
        System.out.println("Methods go on the CALL STACK:");
        System.out.println("  main() calls factorial(3)");
        System.out.println("    factorial(3) calls factorial(2)");
        System.out.println("      factorial(2) calls factorial(1)");
        System.out.println("        factorial(1) returns 1        ← base case");
        System.out.println("      factorial(2) returns 2*1 = 2");
        System.out.println("    factorial(3) returns 3*2 = 6");
        System.out.println("  main() gets 6");
        System.out.println("\nIf recursion never hits base case → StackOverflowError!");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Write a method `int max(int a, int b, int c)` that returns
 *     the maximum of three numbers.
 *
 *  2. Write overloaded methods `double area(double radius)` for
 *     circle and `double area(double length, double width)` for
 *     rectangle.
 *
 *  3. Write a recursive method to reverse a string.
 *     reverse("hello") → "olleh"
 *
 *  4. Write a method `boolean isArmstrong(int n)` that checks
 *     if a number is an Armstrong number.
 *
 *  5. Write a method that takes varargs of strings and returns
 *     the longest one.
 *
 *  6. Write a recursive method to compute the Tower of Hanoi
 *     solution for n disks.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 06 — Arrays
 * ============================================================
 */
