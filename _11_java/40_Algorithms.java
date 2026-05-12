/*
 * ============================================================
 *  CHAPTER 40: ALGORITHMS IN JAVA
 * ============================================================
 *  Master these algorithms and you can solve any coding problem.
 *
 *  COVERED:
 *    1. Sorting: Bubble, Selection, Insertion, Merge, Quick
 *    2. Searching: Linear, Binary
 *    3. Recursion & Backtracking
 *    4. Dynamic Programming
 *    5. Two Pointers / Sliding Window
 *    6. Graph Algorithms: BFS, DFS, Dijkstra
 *    7. Big-O Complexity Analysis
 * ============================================================
 */

import java.util.*;

public class Chapter40_Algorithms {

    // ========================================================
    // 1. SORTING ALGORITHMS
    // ========================================================

    // Bubble Sort — O(n²) — swap adjacent if wrong order
    static void bubbleSort(int[] arr) {
        int n = arr.length;
        for (int i = 0; i < n - 1; i++) {
            boolean swapped = false;
            for (int j = 0; j < n - 1 - i; j++) {
                if (arr[j] > arr[j + 1]) {
                    int tmp = arr[j]; arr[j] = arr[j + 1]; arr[j + 1] = tmp;
                    swapped = true;
                }
            }
            if (!swapped) break;  // optimization: already sorted
        }
    }

    // Selection Sort — O(n²) — find min, place at front
    static void selectionSort(int[] arr) {
        for (int i = 0; i < arr.length - 1; i++) {
            int minIdx = i;
            for (int j = i + 1; j < arr.length; j++) {
                if (arr[j] < arr[minIdx]) minIdx = j;
            }
            int tmp = arr[i]; arr[i] = arr[minIdx]; arr[minIdx] = tmp;
        }
    }

    // Insertion Sort — O(n²) — insert each element into sorted portion
    static void insertionSort(int[] arr) {
        for (int i = 1; i < arr.length; i++) {
            int key = arr[i];
            int j = i - 1;
            while (j >= 0 && arr[j] > key) {
                arr[j + 1] = arr[j];
                j--;
            }
            arr[j + 1] = key;
        }
    }

    // Merge Sort — O(n log n) — divide, sort halves, merge
    static void mergeSort(int[] arr, int left, int right) {
        if (left >= right) return;
        int mid = left + (right - left) / 2;
        mergeSort(arr, left, mid);
        mergeSort(arr, mid + 1, right);
        merge(arr, left, mid, right);
    }

    private static void merge(int[] arr, int left, int mid, int right) {
        int[] temp = new int[right - left + 1];
        int i = left, j = mid + 1, k = 0;

        while (i <= mid && j <= right) {
            temp[k++] = arr[i] <= arr[j] ? arr[i++] : arr[j++];
        }
        while (i <= mid) temp[k++] = arr[i++];
        while (j <= right) temp[k++] = arr[j++];

        System.arraycopy(temp, 0, arr, left, temp.length);
    }

    // Quick Sort — O(n log n) avg, O(n²) worst — partition around pivot
    static void quickSort(int[] arr, int low, int high) {
        if (low >= high) return;
        int pivot = partition(arr, low, high);
        quickSort(arr, low, pivot - 1);
        quickSort(arr, pivot + 1, high);
    }

    private static int partition(int[] arr, int low, int high) {
        int pivot = arr[high];
        int i = low - 1;
        for (int j = low; j < high; j++) {
            if (arr[j] < pivot) {
                i++;
                int tmp = arr[i]; arr[i] = arr[j]; arr[j] = tmp;
            }
        }
        int tmp = arr[i + 1]; arr[i + 1] = arr[high]; arr[high] = tmp;
        return i + 1;
    }

    // ========================================================
    // 2. SEARCHING
    // ========================================================

    // Binary Search — O(log n) — array MUST be sorted
    static int binarySearch(int[] arr, int target) {
        int left = 0, right = arr.length - 1;
        while (left <= right) {
            int mid = left + (right - left) / 2;
            if (arr[mid] == target) return mid;
            if (arr[mid] < target) left = mid + 1;
            else right = mid - 1;
        }
        return -1;  // not found
    }

    // ========================================================
    // 3. RECURSION & BACKTRACKING
    // ========================================================

    // Fibonacci with memoization
    static long fib(int n, long[] memo) {
        if (n <= 1) return n;
        if (memo[n] != 0) return memo[n];
        memo[n] = fib(n - 1, memo) + fib(n - 2, memo);
        return memo[n];
    }

    // Permutations using backtracking
    static void permutations(String str, int left, int right, List<String> result) {
        if (left == right) {
            result.add(str);
            return;
        }
        char[] chars = str.toCharArray();
        for (int i = left; i <= right; i++) {
            char tmp = chars[left]; chars[left] = chars[i]; chars[i] = tmp;
            permutations(new String(chars), left + 1, right, result);
        }
    }

    // N-Queens (classic backtracking)
    static int solveNQueens(int n) {
        int[] count = {0};
        solveNQueensHelper(new int[n], 0, n, count);
        return count[0];
    }

    private static void solveNQueensHelper(int[] queens, int row, int n, int[] count) {
        if (row == n) { count[0]++; return; }
        for (int col = 0; col < n; col++) {
            if (isQueenSafe(queens, row, col)) {
                queens[row] = col;
                solveNQueensHelper(queens, row + 1, n, count);
            }
        }
    }

    private static boolean isQueenSafe(int[] queens, int row, int col) {
        for (int i = 0; i < row; i++) {
            if (queens[i] == col || Math.abs(queens[i] - col) == Math.abs(i - row))
                return false;
        }
        return true;
    }

    // ========================================================
    // 4. DYNAMIC PROGRAMMING
    // ========================================================

    // 0/1 Knapsack
    static int knapsack(int[] weights, int[] values, int capacity) {
        int n = weights.length;
        int[][] dp = new int[n + 1][capacity + 1];

        for (int i = 1; i <= n; i++) {
            for (int w = 0; w <= capacity; w++) {
                dp[i][w] = dp[i - 1][w];  // don't take item i
                if (weights[i - 1] <= w) {
                    dp[i][w] = Math.max(dp[i][w],
                        dp[i - 1][w - weights[i - 1]] + values[i - 1]);
                }
            }
        }
        return dp[n][capacity];
    }

    // Longest Common Subsequence
    static int lcs(String a, String b) {
        int m = a.length(), n = b.length();
        int[][] dp = new int[m + 1][n + 1];

        for (int i = 1; i <= m; i++) {
            for (int j = 1; j <= n; j++) {
                if (a.charAt(i - 1) == b.charAt(j - 1)) {
                    dp[i][j] = dp[i - 1][j - 1] + 1;
                } else {
                    dp[i][j] = Math.max(dp[i - 1][j], dp[i][j - 1]);
                }
            }
        }
        return dp[m][n];
    }

    // Coin Change (minimum coins to make amount)
    static int coinChange(int[] coins, int amount) {
        int[] dp = new int[amount + 1];
        Arrays.fill(dp, amount + 1);
        dp[0] = 0;

        for (int i = 1; i <= amount; i++) {
            for (int coin : coins) {
                if (coin <= i) {
                    dp[i] = Math.min(dp[i], dp[i - coin] + 1);
                }
            }
        }
        return dp[amount] > amount ? -1 : dp[amount];
    }

    // ========================================================
    // 5. TWO POINTERS / SLIDING WINDOW
    // ========================================================

    // Two Sum (sorted array)
    static int[] twoSumSorted(int[] arr, int target) {
        int left = 0, right = arr.length - 1;
        while (left < right) {
            int sum = arr[left] + arr[right];
            if (sum == target) return new int[]{left, right};
            if (sum < target) left++;
            else right--;
        }
        return new int[]{-1, -1};
    }

    // Max sum of subarray of size k (sliding window)
    static int maxSumSubarray(int[] arr, int k) {
        int windowSum = 0, maxSum = Integer.MIN_VALUE;

        for (int i = 0; i < arr.length; i++) {
            windowSum += arr[i];
            if (i >= k - 1) {
                maxSum = Math.max(maxSum, windowSum);
                windowSum -= arr[i - k + 1];
            }
        }
        return maxSum;
    }

    // ========================================================
    // 6. DIJKSTRA'S SHORTEST PATH
    // ========================================================
    static Map<String, Integer> dijkstra(
            Map<String, List<int[]>> graph,  // node → [(neighbor_index, weight)]
            String[] nodes, String start) {

        Map<String, Integer> dist = new HashMap<>();
        for (String n : nodes) dist.put(n, Integer.MAX_VALUE);
        dist.put(start, 0);

        PriorityQueue<String> pq = new PriorityQueue<>(Comparator.comparingInt(dist::get));
        Set<String> visited = new HashSet<>();
        pq.add(start);

        while (!pq.isEmpty()) {
            String u = pq.poll();
            if (!visited.add(u)) continue;

            for (int[] edge : graph.getOrDefault(u, List.of())) {
                String v = nodes[edge[0]];
                int weight = edge[1];
                int newDist = dist.get(u) + weight;
                if (newDist < dist.get(v)) {
                    dist.put(v, newDist);
                    pq.add(v);
                }
            }
        }
        return dist;
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        // --- 1. Sorting ---
        System.out.println("=== SORTING ===\n");

        int[] a1 = {64, 25, 12, 22, 11};
        bubbleSort(a1);
        System.out.println("  Bubble:    " + Arrays.toString(a1));

        int[] a2 = {64, 25, 12, 22, 11};
        selectionSort(a2);
        System.out.println("  Selection: " + Arrays.toString(a2));

        int[] a3 = {64, 25, 12, 22, 11};
        insertionSort(a3);
        System.out.println("  Insertion: " + Arrays.toString(a3));

        int[] a4 = {64, 25, 12, 22, 11};
        mergeSort(a4, 0, a4.length - 1);
        System.out.println("  Merge:     " + Arrays.toString(a4));

        int[] a5 = {64, 25, 12, 22, 11};
        quickSort(a5, 0, a5.length - 1);
        System.out.println("  Quick:     " + Arrays.toString(a5));

        // --- 2. Searching ---
        System.out.println("\n=== BINARY SEARCH ===\n");
        int[] sorted = {2, 5, 8, 12, 16, 23, 38, 56, 72, 91};
        System.out.println("  Find 23: index " + binarySearch(sorted, 23));
        System.out.println("  Find 50: index " + binarySearch(sorted, 50));

        // --- 3. Recursion ---
        System.out.println("\n=== RECURSION ===\n");
        System.out.println("  Fibonacci(40): " + fib(40, new long[41]));

        List<String> perms = new ArrayList<>();
        permutations("ABC", 0, 2, perms);
        System.out.println("  Permutations of ABC: " + perms);

        System.out.println("  N-Queens(8) solutions: " + solveNQueens(8));

        // --- 4. Dynamic Programming ---
        System.out.println("\n=== DYNAMIC PROGRAMMING ===\n");

        int maxVal = knapsack(new int[]{2, 3, 4, 5}, new int[]{3, 4, 5, 6}, 8);
        System.out.println("  Knapsack(weights=[2,3,4,5], values=[3,4,5,6], cap=8): " + maxVal);

        System.out.println("  LCS(\"ABCBDAB\", \"BDCAB\"): " + lcs("ABCBDAB", "BDCAB"));

        System.out.println("  Coin change([1,5,10,25], 36): " + coinChange(new int[]{1, 5, 10, 25}, 36));

        // --- 5. Two Pointers ---
        System.out.println("\n=== TWO POINTERS ===\n");
        System.out.println("  Two sum sorted [1,3,5,7,9] target=12: "
            + Arrays.toString(twoSumSorted(new int[]{1, 3, 5, 7, 9}, 12)));
        System.out.println("  Max sum k=3 [2,1,5,1,3,2]: "
            + maxSumSubarray(new int[]{2, 1, 5, 1, 3, 2}, 3));

        // --- 6. Dijkstra ---
        System.out.println("\n=== DIJKSTRA'S SHORTEST PATH ===\n");
        String[] nodes = {"A", "B", "C", "D", "E"};
        Map<String, List<int[]>> graph = new HashMap<>();
        graph.put("A", List.of(new int[]{1, 4}, new int[]{2, 2}));    // A→B(4), A→C(2)
        graph.put("B", List.of(new int[]{3, 3}, new int[]{2, 1}));    // B→D(3), B→C(1)
        graph.put("C", List.of(new int[]{1, 1}, new int[]{3, 4}, new int[]{4, 5})); // C→B(1), C→D(4), C→E(5)
        graph.put("D", List.of(new int[]{4, 1}));                     // D→E(1)

        Map<String, Integer> distances = dijkstra(graph, nodes, "A");
        System.out.println("  Shortest from A: " + distances);

        // --- Big-O Summary ---
        System.out.println("\n=== BIG-O COMPLEXITY ===");
        System.out.println("  O(1)       Constant    Hash lookup, array access");
        System.out.println("  O(log n)   Logarithmic Binary search, BST ops");
        System.out.println("  O(n)       Linear      Linear search, iteration");
        System.out.println("  O(n log n) Linearithmic Merge/Quick sort");
        System.out.println("  O(n²)      Quadratic   Bubble/Selection sort, nested loops");
        System.out.println("  O(2^n)     Exponential Recursive fibonacci (no memo)");
        System.out.println("  O(n!)      Factorial   Permutations, brute force TSP");

        System.out.println("\n=== SORTING COMPARISON ===");
        System.out.println("  Bubble:    O(n²)      Stable    In-place   Simple but slow");
        System.out.println("  Selection: O(n²)      Unstable  In-place   Few swaps");
        System.out.println("  Insertion: O(n²)      Stable    In-place   Good for small/nearly sorted");
        System.out.println("  Merge:     O(n log n) Stable    O(n) space Best for linked lists");
        System.out.println("  Quick:     O(n log n) Unstable  In-place   Fastest in practice");

        System.out.println("\n✓ Algorithms Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Implement binary search recursively.
 * 2. Solve: Longest Increasing Subsequence (DP).
 * 3. Implement topological sort for a DAG.
 * 4. Solve: find all subsets of a set (backtracking).
 * 5. Implement A* search algorithm.
 *
 * NEXT: Chapter 41 — Networking & Sockets
 */
