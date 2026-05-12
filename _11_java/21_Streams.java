/*
 * ============================================================
 *  CHAPTER 21: STREAMS API
 * ============================================================
 *
 *  Streams (Java 8+) provide a declarative way to process
 *  collections of data — like SQL queries for Java collections.
 *
 *  KEY CONCEPTS:
 *  - A Stream is NOT a data structure — it's a pipeline of operations
 *  - Streams don't modify the source
 *  - Operations are lazy — computed only when terminal op is called
 *  - Streams can only be consumed ONCE
 *
 *  PIPELINE: Source → Intermediate Ops → Terminal Op
 *
 *  INTERMEDIATE (return Stream, lazy):
 *    filter, map, flatMap, distinct, sorted, peek, limit, skip
 *
 *  TERMINAL (trigger execution, produce result):
 *    forEach, collect, reduce, count, min, max, anyMatch,
 *    allMatch, noneMatch, findFirst, findAny, toArray
 *
 * ============================================================
 */

import java.util.*;
import java.util.stream.*;

public class Chapter21_Streams {

    static class Employee {
        String name;
        String department;
        double salary;

        Employee(String name, String department, double salary) {
            this.name = name;
            this.department = department;
            this.salary = salary;
        }

        @Override
        public String toString() {
            return name + "(" + department + ",$" + salary + ")";
        }
    }

    public static void main(String[] args) {

        // =====================================================
        //  1. CREATING STREAMS
        // =====================================================

        System.out.println("=== CREATING STREAMS ===\n");

        // From collection
        List<String> names = Arrays.asList("Alice", "Bob", "Charlie", "Diana");
        Stream<String> nameStream = names.stream();

        // From array
        int[] arr = {1, 2, 3, 4, 5};
        IntStream intStream = Arrays.stream(arr);

        // Stream.of
        Stream<String> ofStream = Stream.of("X", "Y", "Z");

        // Stream.generate (infinite)
        Stream<Double> randoms = Stream.generate(Math::random).limit(5);

        // Stream.iterate (infinite — needs limit)
        Stream<Integer> counting = Stream.iterate(0, n -> n + 2).limit(5);
        System.out.println("Even numbers: " + counting.collect(Collectors.toList()));

        // IntStream.range
        System.out.print("Range [1,5]: ");
        IntStream.range(1, 6).forEach(n -> System.out.print(n + " "));
        System.out.println();

        // =====================================================
        //  2. INTERMEDIATE OPERATIONS
        // =====================================================

        System.out.println("\n=== INTERMEDIATE OPERATIONS ===\n");

        List<Integer> numbers = Arrays.asList(5, 3, 8, 1, 9, 2, 7, 4, 6, 3, 8);

        // filter — keep elements matching predicate
        List<Integer> evens = numbers.stream()
                .filter(n -> n % 2 == 0)
                .collect(Collectors.toList());
        System.out.println("filter(even): " + evens);

        // map — transform each element
        List<Integer> doubled = numbers.stream()
                .map(n -> n * 2)
                .collect(Collectors.toList());
        System.out.println("map(*2): " + doubled);

        // map with type change
        List<String> nameUpper = names.stream()
                .map(String::toUpperCase)
                .collect(Collectors.toList());
        System.out.println("map(toUpper): " + nameUpper);

        // distinct — remove duplicates
        List<Integer> unique = numbers.stream()
                .distinct()
                .collect(Collectors.toList());
        System.out.println("distinct: " + unique);

        // sorted
        List<Integer> sorted = numbers.stream()
                .sorted()
                .collect(Collectors.toList());
        System.out.println("sorted: " + sorted);

        // sorted(comparator)
        List<String> byLength = names.stream()
                .sorted(Comparator.comparingInt(String::length))
                .collect(Collectors.toList());
        System.out.println("sorted by length: " + byLength);

        // limit & skip
        List<Integer> firstThree = numbers.stream().limit(3).collect(Collectors.toList());
        List<Integer> skipThree = numbers.stream().skip(3).collect(Collectors.toList());
        System.out.println("limit(3): " + firstThree);
        System.out.println("skip(3): " + skipThree);

        // peek — intermediate forEach (for debugging)
        System.out.print("peek: ");
        List<Integer> peeked = numbers.stream()
                .filter(n -> n > 5)
                .peek(n -> System.out.print(n + " "))
                .map(n -> n * 10)
                .collect(Collectors.toList());
        System.out.println("→ " + peeked);

        // flatMap — flatten nested structures
        List<List<Integer>> nested = Arrays.asList(
                Arrays.asList(1, 2, 3),
                Arrays.asList(4, 5),
                Arrays.asList(6, 7, 8, 9)
        );
        List<Integer> flat = nested.stream()
                .flatMap(Collection::stream)
                .collect(Collectors.toList());
        System.out.println("flatMap: " + flat);

        // =====================================================
        //  3. TERMINAL OPERATIONS
        // =====================================================

        System.out.println("\n=== TERMINAL OPERATIONS ===\n");

        // forEach
        System.out.print("forEach: ");
        numbers.stream().distinct().sorted().forEach(n -> System.out.print(n + " "));
        System.out.println();

        // count
        long count = numbers.stream().filter(n -> n > 5).count();
        System.out.println("count(>5): " + count);

        // min / max
        Optional<Integer> min = numbers.stream().min(Integer::compareTo);
        Optional<Integer> max = numbers.stream().max(Integer::compareTo);
        System.out.println("min: " + min.orElse(0));
        System.out.println("max: " + max.orElse(0));

        // reduce — combine all elements into one
        int sum = numbers.stream().reduce(0, Integer::sum);
        System.out.println("reduce(sum): " + sum);

        Optional<Integer> product = numbers.stream().reduce((a, b) -> a * b);
        System.out.println("reduce(product): " + product.orElse(0));

        // Concatenate strings
        String joined = names.stream().reduce("", (a, b) -> a + " " + b).trim();
        System.out.println("reduce(concat): " + joined);

        // anyMatch, allMatch, noneMatch
        boolean hasNeg = numbers.stream().anyMatch(n -> n < 0);
        boolean allPos = numbers.stream().allMatch(n -> n > 0);
        boolean noneZero = numbers.stream().noneMatch(n -> n == 0);
        System.out.println("anyMatch(<0): " + hasNeg);
        System.out.println("allMatch(>0): " + allPos);
        System.out.println("noneMatch(==0): " + noneZero);

        // findFirst, findAny
        Optional<Integer> firstEven = numbers.stream().filter(n -> n % 2 == 0).findFirst();
        System.out.println("findFirst(even): " + firstEven.orElse(-1));

        // toArray
        Integer[] array = numbers.stream().distinct().toArray(Integer[]::new);
        System.out.println("toArray: " + Arrays.toString(array));

        // =====================================================
        //  4. COLLECTORS
        // =====================================================

        System.out.println("\n=== COLLECTORS ===\n");

        List<Employee> employees = Arrays.asList(
            new Employee("Alice", "Engineering", 80000),
            new Employee("Bob", "Engineering", 75000),
            new Employee("Charlie", "Marketing", 65000),
            new Employee("Diana", "Marketing", 70000),
            new Employee("Eve", "Engineering", 90000),
            new Employee("Frank", "HR", 60000)
        );

        // toList, toSet
        List<String> empNames = employees.stream()
                .map(e -> e.name)
                .collect(Collectors.toList());
        System.out.println("Names: " + empNames);

        Set<String> departments = employees.stream()
                .map(e -> e.department)
                .collect(Collectors.toSet());
        System.out.println("Departments: " + departments);

        // joining
        String nameString = employees.stream()
                .map(e -> e.name)
                .collect(Collectors.joining(", "));
        System.out.println("Joined: " + nameString);

        // counting, summing, averaging
        long empCount = employees.stream().collect(Collectors.counting());
        double totalSalary = employees.stream().collect(Collectors.summingDouble(e -> e.salary));
        double avgSalary = employees.stream().collect(Collectors.averagingDouble(e -> e.salary));
        System.out.printf("Count: %d, Total: $%.0f, Average: $%.0f%n", empCount, totalSalary, avgSalary);

        // groupingBy
        Map<String, List<Employee>> byDept = employees.stream()
                .collect(Collectors.groupingBy(e -> e.department));
        System.out.println("\nGrouped by department:");
        byDept.forEach((dept, emps) -> System.out.println("  " + dept + ": " + emps));

        // groupingBy with downstream collector
        Map<String, Double> avgByDept = employees.stream()
                .collect(Collectors.groupingBy(
                        e -> e.department,
                        Collectors.averagingDouble(e -> e.salary)));
        System.out.println("\nAvg salary by dept: " + avgByDept);

        Map<String, Long> countByDept = employees.stream()
                .collect(Collectors.groupingBy(e -> e.department, Collectors.counting()));
        System.out.println("Count by dept: " + countByDept);

        // partitioningBy (splits into true/false groups)
        Map<Boolean, List<Employee>> highEarners = employees.stream()
                .collect(Collectors.partitioningBy(e -> e.salary > 70000));
        System.out.println("\nHigh earners (>70k): " + highEarners.get(true));
        System.out.println("Others: " + highEarners.get(false));

        // toMap
        Map<String, Double> salaryMap = employees.stream()
                .collect(Collectors.toMap(e -> e.name, e -> e.salary));
        System.out.println("\nSalary map: " + salaryMap);

        // summarizingDouble
        DoubleSummaryStatistics stats = employees.stream()
                .collect(Collectors.summarizingDouble(e -> e.salary));
        System.out.println("\nSalary stats:");
        System.out.println("  Count: " + stats.getCount());
        System.out.println("  Sum: $" + stats.getSum());
        System.out.println("  Min: $" + stats.getMin());
        System.out.println("  Max: $" + stats.getMax());
        System.out.println("  Avg: $" + stats.getAverage());

        // =====================================================
        //  5. CHAINING — REAL-WORLD PIPELINE
        // =====================================================

        System.out.println("\n=== CHAINING PIPELINE ===\n");

        // Find top 3 highest paid engineers
        List<String> topEngineers = employees.stream()
                .filter(e -> "Engineering".equals(e.department))
                .sorted(Comparator.comparingDouble((Employee e) -> e.salary).reversed())
                .limit(3)
                .map(e -> e.name + " ($" + e.salary + ")")
                .collect(Collectors.toList());
        System.out.println("Top 3 engineers: " + topEngineers);

        // =====================================================
        //  6. PARALLEL STREAMS
        // =====================================================

        System.out.println("\n=== PARALLEL STREAMS ===\n");

        long seqSum = IntStream.rangeClosed(1, 1_000_000).sum();

        long parSum = IntStream.rangeClosed(1, 1_000_000).parallel().sum();

        System.out.println("Sequential sum: " + seqSum);
        System.out.println("Parallel sum: " + parSum);
        System.out.println("Same result: " + (seqSum == parSum));

        System.out.println("\nWhen to use parallel streams:");
        System.out.println("  ✓ Large data sets (100,000+)");
        System.out.println("  ✓ CPU-intensive operations");
        System.out.println("  ✓ No shared mutable state");
        System.out.println("  ✗ Small data sets (overhead > benefit)");
        System.out.println("  ✗ I/O operations");
        System.out.println("  ✗ Order-dependent operations");

        // =====================================================
        //  7. PRIMITIVE STREAMS
        // =====================================================

        System.out.println("\n=== PRIMITIVE STREAMS ===\n");

        // IntStream, LongStream, DoubleStream — avoid boxing
        int sumPrim = IntStream.of(1, 2, 3, 4, 5).sum();
        OptionalInt maxPrim = IntStream.of(1, 2, 3, 4, 5).max();
        double avg = IntStream.rangeClosed(1, 100).average().orElse(0);

        System.out.println("IntStream sum: " + sumPrim);
        System.out.println("IntStream max: " + maxPrim.orElse(0));
        System.out.println("IntStream avg(1-100): " + avg);

        // Convert between object and primitive streams
        List<Integer> intList = IntStream.rangeClosed(1, 5)
                .boxed()  // IntStream → Stream<Integer>
                .collect(Collectors.toList());
        System.out.println("Boxed: " + intList);

        int sumFromList = intList.stream()
                .mapToInt(Integer::intValue)  // Stream<Integer> → IntStream
                .sum();
        System.out.println("mapToInt sum: " + sumFromList);
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Given a list of strings, find the longest string.
 *
 *  2. Given a list of integers, find the sum of squares of
 *     odd numbers greater than 5.
 *
 *  3. Given sentences, create a word frequency map using streams.
 *
 *  4. Flatten a List<List<String>> and remove duplicates, sorted.
 *
 *  5. Given employees, find the highest-paid person in each dept.
 *
 *  6. Implement a simple CSV parser using streams:
 *     Read lines, split by comma, create objects.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 22 — Optional
 * ============================================================
 */
