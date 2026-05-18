import java.util.*;

/**
 * PHASE 1.3 — TEMPLATE METHOD PATTERN
 *
 * Define the SKELETON of an algorithm in a base class.
 * Let subclasses fill in specific steps.
 *
 * Spring uses this everywhere:
 *   JdbcTemplate   → connect, execute, map, close (you provide SQL + mapper)
 *   RestTemplate   → connect, send, receive, deserialize (you provide URL + type)
 *   JpaRepository  → all CRUD is handled (you provide entity type)
 */

// ============================================================
// EXAMPLE 1: Data Exporter (like JdbcTemplate pattern)
// ============================================================

abstract class DataExporter {

    // THE TEMPLATE METHOD — defines the algorithm skeleton
    // Marked "final" so subclasses can't change the sequence!
    public final void export() {
        System.out.println("  ┌─── Export Started ───┐");
        List<Map<String, String>> data = fetchData();          // Step 1: Get data
        List<Map<String, String>> processed = processData(data); // Step 2: Transform
        String formatted = formatData(processed);               // Step 3: Format
        writeOutput(formatted);                                 // Step 4: Output
        System.out.println("  └─── Export Complete ───┘\n");
    }

    // Abstract steps — subclasses MUST implement these
    protected abstract List<Map<String, String>> fetchData();
    protected abstract String formatData(List<Map<String, String>> data);
    protected abstract void writeOutput(String data);

    // Hook — optional override (has default behavior)
    protected List<Map<String, String>> processData(List<Map<String, String>> data) {
        return data;  // Default: no processing
    }
}

// Concrete implementation: CSV export
class CsvExporter extends DataExporter {
    protected List<Map<String, String>> fetchData() {
        System.out.println("  │ Fetching from database...");
        return List.of(
            Map.of("name", "Alice", "email", "alice@mail.com"),
            Map.of("name", "Bob", "email", "bob@mail.com")
        );
    }

    protected String formatData(List<Map<String, String>> data) {
        System.out.println("  │ Formatting as CSV...");
        StringBuilder sb = new StringBuilder("name,email\n");
        for (Map<String, String> row : data) {
            sb.append(row.get("name")).append(",").append(row.get("email")).append("\n");
        }
        return sb.toString();
    }

    protected void writeOutput(String data) {
        System.out.println("  │ Writing CSV output:");
        System.out.println("  │ " + data.replace("\n", "\n  │ "));
    }
}

// Concrete implementation: JSON export
class JsonExporter extends DataExporter {
    protected List<Map<String, String>> fetchData() {
        System.out.println("  │ Fetching from database...");
        return List.of(
            Map.of("name", "Charlie", "role", "admin"),
            Map.of("name", "Diana", "role", "user")
        );
    }

    protected String formatData(List<Map<String, String>> data) {
        System.out.println("  │ Formatting as JSON...");
        StringBuilder sb = new StringBuilder("[\n");
        for (int i = 0; i < data.size(); i++) {
            Map<String, String> row = data.get(i);
            sb.append("  {");
            row.forEach((k, v) -> sb.append("\"").append(k).append("\":\"").append(v).append("\","));
            sb.setLength(sb.length() - 1);  // Remove trailing comma
            sb.append("}");
            if (i < data.size() - 1) sb.append(",");
            sb.append("\n");
        }
        sb.append("]");
        return sb.toString();
    }

    protected void writeOutput(String data) {
        System.out.println("  │ Writing JSON output:");
        for (String line : data.split("\n")) {
            System.out.println("  │   " + line);
        }
    }

    // Override the hook! Filter out non-admins
    @Override
    protected List<Map<String, String>> processData(List<Map<String, String>> data) {
        System.out.println("  │ Filtering admins only...");
        return data.stream()
            .filter(row -> "admin".equals(row.get("role")))
            .toList();
    }
}


// ============================================================
// EXAMPLE 2: Simplified JdbcTemplate (shows the Spring connection)
// ============================================================

// Functional interface — like Spring's RowMapper<T>
interface RowMapper<T> {
    T mapRow(Map<String, Object> row);
}

class SimpleJdbcTemplate {
    private final String dbUrl;

    public SimpleJdbcTemplate(String dbUrl) {
        this.dbUrl = dbUrl;
    }

    // The template method — handles ALL boilerplate
    // You only provide: SQL query + how to map a row
    public <T> List<T> query(String sql, RowMapper<T> mapper) {
        System.out.println("    1. Getting connection to: " + dbUrl);
        System.out.println("    2. Creating statement...");
        System.out.println("    3. Executing: " + sql);

        // Simulated result set
        List<Map<String, Object>> resultSet = simulateQuery(sql);

        System.out.println("    4. Mapping " + resultSet.size() + " rows...");
        List<T> results = new ArrayList<>();
        for (Map<String, Object> row : resultSet) {
            results.add(mapper.mapRow(row));  // YOU provide this logic
        }

        System.out.println("    5. Closing connection...");
        System.out.println("    6. Done! Returning " + results.size() + " objects.\n");
        return results;
    }

    private List<Map<String, Object>> simulateQuery(String sql) {
        // Simulate a database response
        return List.of(
            Map.of("id", 1, "name", "Alice", "age", 28),
            Map.of("id", 2, "name", "Bob", "age", 35),
            Map.of("id", 3, "name", "Charlie", "age", 22)
        );
    }
}

// A simple User for the JdbcTemplate example
class TemplateUser {
    int id; String name; int age;
    TemplateUser(int id, String name, int age) { this.id = id; this.name = name; this.age = age; }
    public String toString() { return "User{id=" + id + ", name='" + name + "', age=" + age + "}"; }
}


public class Step4_TemplateMethod {
    public static void main(String[] args) {

        System.out.println("=== TEMPLATE METHOD: Data Exporter ===");
        System.out.println("  Same algorithm (fetch→process→format→write), different implementations\n");

        System.out.println("--- CSV Exporter ---");
        new CsvExporter().export();

        System.out.println("--- JSON Exporter (with filtering hook) ---");
        new JsonExporter().export();


        System.out.println("\n=== TEMPLATE METHOD: JdbcTemplate Style ===");
        System.out.println("  Steps 1-3, 5-6 are BOILERPLATE (template handles them)");
        System.out.println("  Step 4 (mapping) is YOUR CODE (you provide a RowMapper)\n");

        SimpleJdbcTemplate jdbc = new SimpleJdbcTemplate("jdbc:mysql://localhost:3306/mydb");

        // YOU only write the mapping logic — everything else is handled!
        List<TemplateUser> users = jdbc.query(
            "SELECT * FROM users",
            row -> new TemplateUser(
                (int) row.get("id"),
                (String) row.get("name"),
                (int) row.get("age")
            )
        );

        System.out.println("  Results:");
        users.forEach(u -> System.out.println("    " + u));


        System.out.println("\n=== HOW THIS MAPS TO SPRING ===");
        System.out.println("  ┌──────────────────────────────────────────────────────┐");
        System.out.println("  │ Our SimpleJdbcTemplate  →  Spring's JdbcTemplate     │");
        System.out.println("  │ Our RowMapper<T>        →  Spring's RowMapper<T>     │");
        System.out.println("  │ jdbc.query(sql, mapper) →  jdbcTemplate.query(...)   │");
        System.out.println("  │                                                      │");
        System.out.println("  │ YOU provide: SQL + how to map a row                  │");
        System.out.println("  │ SPRING handles: connection, statement, exceptions,   │");
        System.out.println("  │                 result iteration, closing resources   │");
        System.out.println("  └──────────────────────────────────────────────────────┘");

        System.out.println("\n=== KEY TAKEAWAY ===");
        System.out.println("  Template Method = fixed skeleton + pluggable steps.");
        System.out.println("  JdbcTemplate, RestTemplate, TransactionTemplate all use this.");
        System.out.println("  You write the UNIQUE part. Spring handles the REPETITIVE part.");
    }
}
