/*
 * ============================================================
 *  CHAPTER 18: COLLECTIONS DEEP DIVE
 * ============================================================
 *
 *  This chapter covers Comparable, Comparator, and how
 *  collections work internally.
 *
 * ============================================================
 */

import java.util.*;

public class Chapter18_CollectionsDeepDive {

    // =====================================================
    //  1. COMPARABLE — Natural Ordering
    // =====================================================

    static class Student implements Comparable<Student> {
        String name;
        double gpa;

        Student(String name, double gpa) {
            this.name = name;
            this.gpa = gpa;
        }

        // Natural ordering: by GPA descending
        @Override
        public int compareTo(Student other) {
            return Double.compare(other.gpa, this.gpa); // descending
        }

        // MUST override equals and hashCode for collections
        @Override
        public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof Student)) return false;
            Student s = (Student) o;
            return Double.compare(s.gpa, gpa) == 0 && name.equals(s.name);
        }

        @Override
        public int hashCode() {
            return Objects.hash(name, gpa);
        }

        @Override
        public String toString() {
            return name + "(GPA:" + gpa + ")";
        }
    }

    // =====================================================
    //  2. COMPARATOR — Custom Ordering
    // =====================================================

    // Separate comparison logic from the class
    static class NameComparator implements Comparator<Student> {
        @Override
        public int compare(Student a, Student b) {
            return a.name.compareTo(b.name);
        }
    }

    // =====================================================
    //  3. HOW HASHMAP WORKS INTERNALLY
    // =====================================================

    /*
     * HashMap internals:
     * ──────────────────
     * 1. Uses an array of "buckets" (Node<K,V>[])
     * 2. hashCode() determines which bucket
     * 3. equals() resolves collisions within a bucket
     * 4. Default capacity: 16, load factor: 0.75
     * 5. When 75% full → resize (double capacity)
     * 6. Java 8+: bucket becomes TreeNode after 8 collisions (O(log n))
     *
     *  Buckets:
     *  [0] → null
     *  [1] → Node(key="Alice", val=25) → Node(key="Eve", val=22) → null
     *  [2] → null
     *  [3] → Node(key="Bob", val=30) → null
     *  ...
     *
     * WHY override both equals() AND hashCode()?
     * - hashCode() → finds the bucket
     * - equals()   → finds the exact key within the bucket
     * - If equals() is true, hashCode() MUST be equal
     * - If hashCode() is equal, equals() may be false (collision)
     */

    // Object that works correctly as HashMap key
    static class Point {
        int x, y;

        Point(int x, int y) { this.x = x; this.y = y; }

        @Override
        public boolean equals(Object o) {
            if (this == o) return true;
            if (!(o instanceof Point)) return false;
            Point p = (Point) o;
            return x == p.x && y == p.y;
        }

        @Override
        public int hashCode() {
            return Objects.hash(x, y); // consistent with equals
        }

        @Override
        public String toString() {
            return "(" + x + "," + y + ")";
        }
    }

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Comparable ---
        System.out.println("=== COMPARABLE ===\n");

        List<Student> students = new ArrayList<>(Arrays.asList(
            new Student("Alice", 3.8),
            new Student("Bob", 3.5),
            new Student("Charlie", 3.9),
            new Student("Diana", 3.7)
        ));

        Collections.sort(students); // uses compareTo (natural order: GPA descending)
        System.out.println("By GPA (natural): " + students);

        // --- 2. Comparator ---
        System.out.println("\n=== COMPARATOR ===\n");

        // Named comparator class
        students.sort(new NameComparator());
        System.out.println("By name: " + students);

        // Lambda comparator
        students.sort((a, b) -> Double.compare(a.gpa, b.gpa));
        System.out.println("By GPA ascending (lambda): " + students);

        // Comparator.comparing (Java 8+)
        students.sort(Comparator.comparing(s -> s.name));
        System.out.println("By name (Comparator.comparing): " + students);

        students.sort(Comparator.comparingDouble((Student s) -> s.gpa).reversed());
        System.out.println("By GPA descending: " + students);

        // Chained comparators
        List<Student> moreStudents = new ArrayList<>(Arrays.asList(
            new Student("Alice", 3.8),
            new Student("Bob", 3.8),
            new Student("Charlie", 3.5),
            new Student("Alice", 3.5)
        ));
        moreStudents.sort(Comparator.comparingDouble((Student s) -> s.gpa)
                .reversed()
                .thenComparing(s -> s.name));
        System.out.println("By GPA desc then name: " + moreStudents);

        // --- 3. Using custom object as Map key ---
        System.out.println("\n=== CUSTOM KEY IN MAP ===\n");

        Map<Point, String> grid = new HashMap<>();
        grid.put(new Point(0, 0), "Origin");
        grid.put(new Point(1, 0), "Right");
        grid.put(new Point(0, 1), "Up");

        // This works because we overrode equals() and hashCode()
        System.out.println("Point(0,0): " + grid.get(new Point(0, 0)));
        System.out.println("Point(1,0): " + grid.get(new Point(1, 0)));

        // --- 4. TreeSet with Comparator ---
        System.out.println("\n=== TREESET WITH COMPARATOR ===\n");

        // TreeSet with custom ordering
        TreeSet<Student> byGPA = new TreeSet<>(
                Comparator.comparingDouble((Student s) -> s.gpa).reversed()
                        .thenComparing(s -> s.name)
        );
        byGPA.add(new Student("Alice", 3.8));
        byGPA.add(new Student("Bob", 3.5));
        byGPA.add(new Student("Charlie", 3.9));
        System.out.println("TreeSet by GPA: " + byGPA);

        // --- 5. Unmodifiable / Immutable Collections ---
        System.out.println("\n=== IMMUTABLE COLLECTIONS ===\n");

        // Java 9+ factory methods (truly immutable)
        List<String> immList = List.of("A", "B", "C");
        Set<String> immSet = Set.of("X", "Y", "Z");
        Map<String, Integer> immMap = Map.of("one", 1, "two", 2);

        System.out.println("Immutable List: " + immList);
        System.out.println("Immutable Set: " + immSet);
        System.out.println("Immutable Map: " + immMap);

        try {
            immList.add("D");
        } catch (UnsupportedOperationException e) {
            System.out.println("Can't modify immutable list!");
        }

        // --- 6. Practical: Group students by GPA range ---
        System.out.println("\n=== GROUPING ===\n");

        Map<String, List<Student>> grouped = new TreeMap<>();
        for (Student s : students) {
            String group = s.gpa >= 3.7 ? "Honors" : "Regular";
            grouped.computeIfAbsent(group, k -> new ArrayList<>()).add(s);
        }
        grouped.forEach((group, list) ->
                System.out.println(group + ": " + list));

        // --- 7. Frequency map with merge ---
        System.out.println("\n=== MERGE METHOD ===\n");

        String text = "hello world hello java world hello";
        Map<String, Integer> freq = new HashMap<>();
        for (String word : text.split(" ")) {
            freq.merge(word, 1, Integer::sum); // cleaner than getOrDefault
        }
        System.out.println("Frequency: " + freq);

        // --- 8. Performance comparison ---
        System.out.println("\n=== PERFORMANCE SUMMARY ===\n");
        System.out.println("┌──────────────┬─────────┬──────────┬───────────┐");
        System.out.println("│ Operation    │ArrayList│LinkedList│ HashSet   │");
        System.out.println("├──────────────┼─────────┼──────────┼───────────┤");
        System.out.println("│ get(index)   │  O(1)   │   O(n)   │    N/A    │");
        System.out.println("│ add(end)     │  O(1)*  │   O(1)   │   O(1)    │");
        System.out.println("│ add(middle)  │  O(n)   │   O(1)   │    N/A    │");
        System.out.println("│ remove       │  O(n)   │   O(1)   │   O(1)    │");
        System.out.println("│ contains     │  O(n)   │   O(n)   │   O(1)    │");
        System.out.println("│ Ordering     │  Yes    │   Yes    │    No     │");
        System.out.println("└──────────────┴─────────┴──────────┴───────────┘");
        System.out.println("  * O(1) amortized — occasional resize is O(n)");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Implement Comparable on a Product class (by price).
 *     Sort a list of products.
 *
 *  2. Write multiple Comparators for an Employee class:
 *     by name, by salary, by department then salary.
 *
 *  3. Implement an LRU Cache using LinkedHashMap.
 *
 *  4. Given two lists, find common elements using a Set.
 *
 *  5. Build a frequency map and find the most frequent element.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 19 — Generics
 * ============================================================
 */
