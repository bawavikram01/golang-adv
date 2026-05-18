import java.util.*;
import java.util.function.*;
import java.util.stream.*;

/**
 * PHASE 1.1 — LAMBDA & STREAMS - Modern Spring Style

 * 
 Theory
Lambdas are short anonymous functions. Streams let you process collections in a pipeline (filter → map → collect). Modern Spring uses these extensively — in security configs, repository queries, WebFlux, and more.

Analogy
Lambda = A delivery note without a name. Instead of hiring a full-time employee (named method), you hand a sticky note saying "do this" (lambda).

Stream = A factory conveyor belt. Items flow through stations (filter, transform, collect) and come out as a finished product.
Theory
Lambdas are short anonymous functions. Streams let you process collections in a pipeline (filter → map → collect). Modern Spring uses these extensively — in security configs, repository queries, WebFlux, and more.

Analogy
Lambda = A delivery note without a name. Instead of hiring a full-time employee (named method), you hand a sticky note saying "do this" (lambda).

Stream = A factory conveyor belt. Items flow through stations (filter, transform, collect) and come out as a finished product.

 * Modern Spring uses lambdas everywhere:
 *   - Security: http.csrf(csrf -> csrf.disable())
 *   - Streams: userRepo.findAll().stream().filter(...)
 *   - Functional endpoints: RouterFunction in WebFlux
 *   - Optional: user.orElseThrow(() -> new NotFoundException("..."))
 */

// ============================================================
// Simple data class for examples
// ============================================================
class Employee {
    String name;
    String department;
    double salary;

    Employee(String name, String department, double salary) {
        this.name = name;
        this.department = department;
        this.salary = salary;
    }

    public String toString() {
        return String.format("%-10s | %-12s | $%,.0f", name, department, salary);
    }
}


public class Step5_LambdaStreams {
    public static void main(String[] args) {

        // ============================================================
        // PART 1: LAMBDAS
        // ============================================================
        System.out.println("=== 1. LAMBDAS ===\n");

        // OLD WAY: Anonymous inner class (verbose!)
        Runnable oldWay = new Runnable() {
            @Override
            public void run() {
                System.out.println("  Old way: verbose anonymous class");
            }
        };
        oldWay.run();

        // NEW WAY: Lambda (clean!)
        Runnable newWay = () -> System.out.println("  New way: clean lambda!");
        newWay.run();

        System.out.println();

        // Lambda with parameters
        // (a, b) -> a + b  is equivalent to a method that takes two ints and returns their sum
        BiFunction<Integer, Integer, Integer> add = (a, b) -> a + b;
        System.out.println("  add(5, 3) = " + add.apply(5, 3));

        // Lambda as a predicate (boolean condition) — used in Stream.filter()
        Predicate<String> isLong = s -> s.length() > 5;
        System.out.println("  isLong(\"Hi\") = " + isLong.test("Hi"));
        System.out.println("  isLong(\"Spring Boot\") = " + isLong.test("Spring Boot"));

        // Lambda as a transformer — used in Stream.map()
        Function<String, String> toUpper = s -> s.toUpperCase();
        System.out.println("  toUpper(\"spring\") = " + toUpper.apply("spring"));

        // Consumer — takes input, returns nothing — used in .forEach()
        Consumer<String> printer = s -> System.out.println("  → " + s);
        printer.accept("Hello from consumer!");

        // Supplier — takes nothing, returns something — used in .orElseGet()
        Supplier<String> greeting = () -> "Hello, World!";
        System.out.println("  supplier: " + greeting.get());


        // ============================================================
        // PART 2: STREAMS
        // ============================================================
        System.out.println("\n=== 2. STREAMS ===\n");

        List<Employee> employees = List.of(
            new Employee("Alice",   "Engineering", 120000),
            new Employee("Bob",     "Engineering",  95000),
            new Employee("Charlie", "Marketing",    80000),
            new Employee("Diana",   "Engineering", 140000),
            new Employee("Eve",     "Marketing",    70000),
            new Employee("Frank",   "Sales",        85000),
            new Employee("Grace",   "Engineering", 110000)
        );

        System.out.println("All employees:");
        employees.forEach(e -> System.out.println("  " + e));

        // --- filter: Keep only elements that match a condition ---
        System.out.println("\nEngineers only (filter):");
        employees.stream()
            .filter(e -> e.department.equals("Engineering"))
            .forEach(e -> System.out.println("  " + e));

        // --- map: Transform each element ---
        System.out.println("\nAll names uppercased (map):");
        List<String> names = employees.stream()
            .map(e -> e.name.toUpperCase())
            .collect(Collectors.toList());
        System.out.println("  " + names);

        // --- sorted: Order elements ---
        System.out.println("\nSorted by salary descending:");
        employees.stream()
            .sorted((a, b) -> Double.compare(b.salary, a.salary))
            .forEach(e -> System.out.println("  " + e));

        // --- reduce: Combine all elements into one value ---
        double totalSalary = employees.stream()
            .mapToDouble(e -> e.salary)
            .sum();
        System.out.println("\nTotal salary bill: $" + String.format("%,.0f", totalSalary));

        // --- collect + groupingBy: Group elements ---
        System.out.println("\nGrouped by department:");
        Map<String, List<Employee>> byDept = employees.stream()
            .collect(Collectors.groupingBy(e -> e.department));
        byDept.forEach((dept, emps) -> {
            System.out.println("  " + dept + ":");
            emps.forEach(e -> System.out.println("    " + e));
        });

        // --- Chaining it all together ---
        System.out.println("\nTop 3 highest paid engineers:");
        employees.stream()
            .filter(e -> e.department.equals("Engineering"))  // only engineers
            .sorted((a, b) -> Double.compare(b.salary, a.salary))  // highest first
            .limit(3)                                          // top 3
            .forEach(e -> System.out.println("  " + e));


        // ============================================================
        // PART 3: Optional — No More NullPointerException
        // ============================================================
        System.out.println("\n=== 3. OPTIONAL ===\n");

        // Spring Data returns Optional<T> from findById()
        Optional<Employee> found = employees.stream()
            .filter(e -> e.name.equals("Diana"))
            .findFirst();

        // Safe access — no null check needed!
        found.ifPresent(e -> System.out.println("  Found: " + e));

        // Getting a value with a fallback
        String name = found.map(e -> e.name).orElse("Unknown");
        System.out.println("  Name: " + name);

        // Throwing if not found (common pattern in Spring services)
        Optional<Employee> notFound = employees.stream()
            .filter(e -> e.name.equals("Zack"))
            .findFirst();

        String result;
        try {
            result = notFound.map(e -> e.name)
                .orElseThrow(() -> new RuntimeException("Employee not found!"));
        } catch (RuntimeException ex) {
            result = "Exception: " + ex.getMessage();
        }
        System.out.println("  " + result);


        // ============================================================
        // PART 4: How Spring Uses Lambdas
        // ============================================================
        System.out.println("\n=== 4. HOW SPRING USES LAMBDAS (Preview) ===\n");
        System.out.println("  // Spring Security configuration:");
        System.out.println("  http");
        System.out.println("    .csrf(csrf -> csrf.disable())");
        System.out.println("    .authorizeHttpRequests(auth -> auth");
        System.out.println("        .requestMatchers(\"/api/**\").authenticated()");
        System.out.println("        .anyRequest().permitAll()");
        System.out.println("    )");
        System.out.println("    .httpBasic(Customizer.withDefaults());");
        System.out.println();
        System.out.println("  // Spring Data — findById returns Optional");
        System.out.println("  User user = userRepo.findById(id)");
        System.out.println("      .orElseThrow(() -> new UserNotFoundException(id));");

        System.out.println("\n=== KEY TAKEAWAY ===");
        System.out.println("Lambda     = short anonymous function: (params) -> expression");
        System.out.println("Stream     = pipeline: .filter().map().collect()");
        System.out.println("Optional   = safe container that may or may not hold a value");
        System.out.println("Spring uses all three extensively in modern code.");
    }
}
