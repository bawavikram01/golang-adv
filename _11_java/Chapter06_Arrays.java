/*
 * ============================================================
 *  CHAPTER 06: ARRAYS
 * ============================================================
 *
 *  An ARRAY is a fixed-size, ordered collection of elements
 *  of the SAME data type, stored in contiguous memory.
 *
 *  KEY FACTS:
 *  - Index starts at 0
 *  - Size is FIXED after creation (cannot grow/shrink)
 *  - Arrays are OBJECTS in Java (stored in heap)
 *  - Default values: 0 (int), 0.0 (double), false (boolean), null (objects)
 *
 *  MEMORY LAYOUT:
 *      int[] arr = {10, 20, 30, 40, 50};
 *
 *      Index:    [0]  [1]  [2]  [3]  [4]
 *      Value:     10   20   30   40   50
 *      Address:  100  104  108  112  116  (each int = 4 bytes)
 *
 * ============================================================
 */

import java.util.Arrays;

public class Chapter06_Arrays {

    public static void main(String[] args) {

        // =====================================================
        //  1. DECLARING AND CREATING ARRAYS
        // =====================================================

        System.out.println("=== DECLARING ARRAYS ===\n");

        // Method 1: Declare, then allocate
        int[] numbers;            // declare (preferred style)
        numbers = new int[5];     // allocate 5 elements (all initialized to 0)

        // Method 2: Declare and allocate in one line
        double[] prices = new double[3];

        // Method 3: Declare, allocate, AND initialize
        int[] scores = {90, 85, 78, 92, 88};

        // Method 4: Explicit new with values
        String[] fruits = new String[]{"Apple", "Banana", "Cherry"};

        // C-style declaration (valid but NOT recommended in Java)
        int oldStyle[] = {1, 2, 3}; // works but prefer int[]

        // Print array info
        System.out.println("numbers length: " + numbers.length);
        System.out.println("scores: " + Arrays.toString(scores));
        System.out.println("fruits: " + Arrays.toString(fruits));

        // Default values
        System.out.println("\n--- Default Values ---");
        int[] intArr = new int[3];
        double[] dblArr = new double[3];
        boolean[] boolArr = new boolean[3];
        String[] strArr = new String[3];
        System.out.println("int default: " + Arrays.toString(intArr));       // [0, 0, 0]
        System.out.println("double default: " + Arrays.toString(dblArr));     // [0.0, 0.0, 0.0]
        System.out.println("boolean default: " + Arrays.toString(boolArr));   // [false, false, false]
        System.out.println("String default: " + Arrays.toString(strArr));     // [null, null, null]

        // =====================================================
        //  2. ACCESSING AND MODIFYING ELEMENTS
        // =====================================================

        System.out.println("\n=== ACCESSING ELEMENTS ===\n");

        int[] arr = {10, 20, 30, 40, 50};

        // Access by index (0-based)
        System.out.println("First element: arr[0] = " + arr[0]);   // 10
        System.out.println("Last element: arr[4] = " + arr[4]);    // 50
        System.out.println("Last (dynamic): arr[arr.length-1] = " + arr[arr.length - 1]); // 50

        // Modify elements
        arr[2] = 300;
        System.out.println("After arr[2]=300: " + Arrays.toString(arr));

        // ArrayIndexOutOfBoundsException — RUNTIME ERROR
        // arr[5] = 60; // CRASH! Valid indices are 0-4

        // =====================================================
        //  3. ITERATING OVER ARRAYS
        // =====================================================

        System.out.println("\n=== ITERATING ARRAYS ===\n");

        int[] data = {5, 12, 8, 3, 17, 9};

        // Method 1: Classic for loop (use when you need the index)
        System.out.print("for loop: ");
        for (int i = 0; i < data.length; i++) {
            System.out.print(data[i] + " ");
        }
        System.out.println();

        // Method 2: Enhanced for-each (use when you just need values)
        System.out.print("for-each: ");
        for (int val : data) {
            System.out.print(val + " ");
        }
        System.out.println();

        // Method 3: Reverse iteration
        System.out.print("reverse:  ");
        for (int i = data.length - 1; i >= 0; i--) {
            System.out.print(data[i] + " ");
        }
        System.out.println();

        // =====================================================
        //  4. COMMON ARRAY OPERATIONS
        // =====================================================

        System.out.println("\n=== COMMON OPERATIONS ===\n");

        int[] nums = {15, 3, 8, 21, 7, 14, 2, 19};

        // Find sum
        int sum = 0;
        for (int n : nums) sum += n;
        System.out.println("Sum: " + sum);

        // Find average
        double average = (double) sum / nums.length;
        System.out.println("Average: " + average);

        // Find min and max
        int min = nums[0], max = nums[0];
        for (int n : nums) {
            if (n < min) min = n;
            if (n > max) max = n;
        }
        System.out.println("Min: " + min + ", Max: " + max);

        // Count occurrences
        int[] grades = {90, 85, 90, 78, 90, 85, 92};
        int target = 90;
        int count = 0;
        for (int g : grades) {
            if (g == target) count++;
        }
        System.out.println("Count of " + target + ": " + count);

        // Linear search
        int searchVal = 14;
        int foundIndex = -1;
        for (int i = 0; i < nums.length; i++) {
            if (nums[i] == searchVal) {
                foundIndex = i;
                break;
            }
        }
        System.out.println("Search " + searchVal + ": index " + foundIndex);

        // Reverse an array
        int[] original = {1, 2, 3, 4, 5};
        int[] reversed = new int[original.length];
        for (int i = 0; i < original.length; i++) {
            reversed[original.length - 1 - i] = original[i];
        }
        System.out.println("Original: " + Arrays.toString(original));
        System.out.println("Reversed: " + Arrays.toString(reversed));

        // In-place reverse (swap from both ends)
        int[] inPlace = {1, 2, 3, 4, 5};
        for (int i = 0; i < inPlace.length / 2; i++) {
            int temp = inPlace[i];
            inPlace[i] = inPlace[inPlace.length - 1 - i];
            inPlace[inPlace.length - 1 - i] = temp;
        }
        System.out.println("In-place reversed: " + Arrays.toString(inPlace));

        // =====================================================
        //  5. JAVA.UTIL.ARRAYS CLASS
        // =====================================================

        System.out.println("\n=== ARRAYS UTILITY CLASS ===\n");

        int[] unsorted = {38, 27, 43, 3, 9, 82, 10};

        // Sort
        int[] sorted = Arrays.copyOf(unsorted, unsorted.length);
        Arrays.sort(sorted);
        System.out.println("Unsorted: " + Arrays.toString(unsorted));
        System.out.println("Sorted:   " + Arrays.toString(sorted));

        // Binary search (only on SORTED array!)
        int idx = Arrays.binarySearch(sorted, 27);
        System.out.println("binarySearch(27): index " + idx);

        // Fill
        int[] filled = new int[5];
        Arrays.fill(filled, 42);
        System.out.println("Filled: " + Arrays.toString(filled));

        // Copy
        int[] src = {1, 2, 3, 4, 5};
        int[] copy = Arrays.copyOf(src, src.length);       // full copy
        int[] partial = Arrays.copyOfRange(src, 1, 4);     // elements [1, 4) → {2, 3, 4}
        int[] extended = Arrays.copyOf(src, 8);             // extra filled with 0
        System.out.println("Copy: " + Arrays.toString(copy));
        System.out.println("Partial [1,4): " + Arrays.toString(partial));
        System.out.println("Extended: " + Arrays.toString(extended));

        // Equals
        int[] a = {1, 2, 3};
        int[] b = {1, 2, 3};
        int[] c = {3, 2, 1};
        System.out.println("a.equals(b): " + Arrays.equals(a, b));  // true
        System.out.println("a.equals(c): " + Arrays.equals(a, c));  // false

        // =====================================================
        //  6. 2D ARRAYS (Matrix)
        // =====================================================

        System.out.println("\n=== 2D ARRAYS ===\n");

        // Declare and initialize
        int[][] matrix = {
            {1, 2, 3},
            {4, 5, 6},
            {7, 8, 9}
        };

        // Access elements: matrix[row][col]
        System.out.println("matrix[0][0] = " + matrix[0][0]); // 1
        System.out.println("matrix[1][2] = " + matrix[1][2]); // 6
        System.out.println("matrix[2][1] = " + matrix[2][1]); // 8

        // Print 2D array
        System.out.println("\nMatrix:");
        for (int i = 0; i < matrix.length; i++) {
            for (int j = 0; j < matrix[i].length; j++) {
                System.out.printf("%3d", matrix[i][j]);
            }
            System.out.println();
        }

        // Or using Arrays.deepToString()
        System.out.println("deepToString: " + Arrays.deepToString(matrix));

        // 2D array with new
        int[][] grid = new int[3][4]; // 3 rows, 4 columns
        grid[0][0] = 1;
        grid[2][3] = 99;

        // Matrix operations
        System.out.println("\n--- Matrix Addition ---");
        int[][] m1 = {{1, 2}, {3, 4}};
        int[][] m2 = {{5, 6}, {7, 8}};
        int[][] result = new int[2][2];
        for (int i = 0; i < 2; i++) {
            for (int j = 0; j < 2; j++) {
                result[i][j] = m1[i][j] + m2[i][j];
            }
        }
        System.out.println("M1 + M2 = " + Arrays.deepToString(result));

        // Matrix transpose
        System.out.println("\n--- Transpose ---");
        int[][] mat = {{1, 2, 3}, {4, 5, 6}};
        int[][] transpose = new int[3][2]; // rows↔cols
        for (int i = 0; i < mat.length; i++) {
            for (int j = 0; j < mat[0].length; j++) {
                transpose[j][i] = mat[i][j];
            }
        }
        System.out.println("Original:  " + Arrays.deepToString(mat));
        System.out.println("Transpose: " + Arrays.deepToString(transpose));

        // =====================================================
        //  7. JAGGED ARRAYS (Rows of different lengths)
        // =====================================================

        System.out.println("\n=== JAGGED ARRAYS ===\n");

        // Each row can have a different number of columns
        int[][] jagged = new int[3][];  // 3 rows, columns not defined yet
        jagged[0] = new int[]{1, 2};
        jagged[1] = new int[]{3, 4, 5, 6};
        jagged[2] = new int[]{7};

        for (int i = 0; i < jagged.length; i++) {
            System.out.println("Row " + i + " (length=" + jagged[i].length + "): "
                    + Arrays.toString(jagged[i]));
        }

        // =====================================================
        //  8. MULTIDIMENSIONAL ARRAYS (3D+)
        // =====================================================

        System.out.println("\n=== 3D ARRAYS ===\n");

        // Think of it as: array of 2D arrays
        int[][][] cube = {
            {{1, 2}, {3, 4}},
            {{5, 6}, {7, 8}}
        };
        System.out.println("cube[0][1][0] = " + cube[0][1][0]); // 3
        System.out.println("cube[1][0][1] = " + cube[1][0][1]); // 6
        System.out.println("3D: " + Arrays.deepToString(cube));

        // =====================================================
        //  9. ARRAY AS METHOD PARAMETER AND RETURN
        // =====================================================

        System.out.println("\n=== ARRAYS WITH METHODS ===\n");

        int[] input = {5, 3, 8, 1, 9, 2};

        System.out.println("Before sort: " + Arrays.toString(input));
        bubbleSort(input);
        System.out.println("After sort:  " + Arrays.toString(input));

        int[] merged = mergeArrays(new int[]{1, 3, 5}, new int[]{2, 4, 6});
        System.out.println("Merged: " + Arrays.toString(merged));

        // =====================================================
        //  10. COMMON PITFALLS
        // =====================================================

        System.out.println("\n=== COMMON PITFALLS ===\n");

        // Pitfall 1: Array reference vs content comparison
        int[] x = {1, 2, 3};
        int[] y = {1, 2, 3};
        System.out.println("x == y: " + (x == y));               // false! Compares references
        System.out.println("Arrays.equals(x,y): " + Arrays.equals(x, y)); // true! Compares content

        // Pitfall 2: Printing array directly
        System.out.println("Direct print: " + x);                     // [I@hashcode (useless!)
        System.out.println("Arrays.toString: " + Arrays.toString(x)); // [1, 2, 3] (useful!)

        // Pitfall 3: Array assignment copies REFERENCE, not content
        int[] ref1 = {10, 20, 30};
        int[] ref2 = ref1;        // ref2 points to SAME array!
        ref2[0] = 999;
        System.out.println("ref1[0] = " + ref1[0]); // 999! Both point to same array!

        // Fix: use Arrays.copyOf for independent copy
        int[] independent = Arrays.copyOf(ref1, ref1.length);
        independent[0] = 0;
        System.out.println("ref1[0] after independent change: " + ref1[0]); // still 999
    }

    // Methods that work with arrays

    static void bubbleSort(int[] arr) {
        int n = arr.length;
        for (int i = 0; i < n - 1; i++) {
            boolean swapped = false;
            for (int j = 0; j < n - 1 - i; j++) {
                if (arr[j] > arr[j + 1]) {
                    int temp = arr[j];
                    arr[j] = arr[j + 1];
                    arr[j + 1] = temp;
                    swapped = true;
                }
            }
            if (!swapped) break; // optimization: already sorted
        }
    }

    static int[] mergeArrays(int[] a, int[] b) {
        int[] result = new int[a.length + b.length];
        System.arraycopy(a, 0, result, 0, a.length);
        System.arraycopy(b, 0, result, a.length, b.length);
        return result;
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Find the second largest element in an array.
 *
 *  2. Remove duplicates from a sorted array.
 *
 *  3. Rotate an array by k positions to the right.
 *     {1,2,3,4,5} rotated by 2 → {4,5,1,2,3}
 *
 *  4. Check if two arrays are equal (without Arrays.equals).
 *
 *  5. Multiply two 2D matrices.
 *
 *  6. Find the saddle point in a matrix (element which is minimum
 *     in its row and maximum in its column).
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 07 — Strings
 * ============================================================
 */
