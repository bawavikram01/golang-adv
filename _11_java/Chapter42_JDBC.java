/*
 * ============================================================
 *  CHAPTER 42: JDBC & DATABASE
 * ============================================================
 *  JDBC (Java Database Connectivity) = Java API for SQL databases.
 *
 *  ARCHITECTURE:
 *    Java App → JDBC API → JDBC Driver → Database
 *
 *  KEY INTERFACES (java.sql):
 *    DriverManager    — manages JDBC drivers
 *    Connection       — database connection
 *    Statement        — execute SQL
 *    PreparedStatement — parameterized SQL (prevents SQL injection!)
 *    ResultSet        — query results
 *    CallableStatement — stored procedures
 *
 *  NOTE: This uses SQLite for portability (no server needed).
 *  For real projects, use PostgreSQL, MySQL, etc.
 *
 *  TO RUN: Download sqlite-jdbc jar and add to classpath:
 *    javac Chapter42_JDBC.java
 *    java -cp .:sqlite-jdbc-3.x.x.jar Chapter42_JDBC
 *
 *  Or just READ this for the concepts — the code patterns are
 *  identical for MySQL, PostgreSQL, Oracle, etc.
 * ============================================================
 */

import java.sql.*;
import java.util.*;

public class Chapter42_JDBC {

    // Connection URL formats for different databases:
    // SQLite:     jdbc:sqlite:mydb.db
    // MySQL:      jdbc:mysql://localhost:3306/mydb
    // PostgreSQL: jdbc:postgresql://localhost:5432/mydb
    // Oracle:     jdbc:oracle:thin:@localhost:1521:mydb
    // H2:         jdbc:h2:mem:testdb (in-memory)

    // Using SQLite (file-based, no server needed)
    private static final String DB_URL = "jdbc:sqlite:chapter42_demo.db";

    // ========================================================
    // 1. GET CONNECTION
    // ========================================================
    static Connection getConnection() throws SQLException {
        return DriverManager.getConnection(DB_URL);
    }

    // ========================================================
    // 2. CREATE TABLE (DDL)
    // ========================================================
    static void createTable(Connection conn) throws SQLException {
        String sql = """
            CREATE TABLE IF NOT EXISTS employees (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                department TEXT NOT NULL,
                salary REAL NOT NULL,
                hire_date TEXT DEFAULT CURRENT_DATE
            )
            """;

        try (Statement stmt = conn.createStatement()) {
            stmt.execute(sql);
            System.out.println("  Table 'employees' created");
        }
    }

    // ========================================================
    // 3. INSERT (with PreparedStatement — ALWAYS use this!)
    // ========================================================
    static int insertEmployee(Connection conn, String name, String dept, double salary)
            throws SQLException {

        String sql = "INSERT INTO employees (name, department, salary) VALUES (?, ?, ?)";

        // PreparedStatement prevents SQL injection!
        // ❌ NEVER: "INSERT ... VALUES ('" + name + "')"  ← SQL INJECTION!
        // ✅ ALWAYS: "INSERT ... VALUES (?, ?, ?)"

        try (PreparedStatement ps = conn.prepareStatement(sql, Statement.RETURN_GENERATED_KEYS)) {
            ps.setString(1, name);
            ps.setString(2, dept);
            ps.setDouble(3, salary);

            int affected = ps.executeUpdate();

            // Get generated ID
            try (ResultSet keys = ps.getGeneratedKeys()) {
                if (keys.next()) {
                    int id = keys.getInt(1);
                    System.out.println("  Inserted: " + name + " (id=" + id + ")");
                    return id;
                }
            }
            return -1;
        }
    }

    // ========================================================
    // 4. SELECT (Query)
    // ========================================================
    static void queryAll(Connection conn) throws SQLException {
        String sql = "SELECT id, name, department, salary FROM employees ORDER BY id";

        try (Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery(sql)) {

            System.out.printf("  %-4s %-15s %-12s %10s%n", "ID", "Name", "Department", "Salary");
            System.out.println("  " + "-".repeat(45));

            while (rs.next()) {
                System.out.printf("  %-4d %-15s %-12s %10.2f%n",
                    rs.getInt("id"),
                    rs.getString("name"),
                    rs.getString("department"),
                    rs.getDouble("salary"));
            }
        }
    }

    // Select with WHERE clause
    static void queryByDepartment(Connection conn, String dept) throws SQLException {
        String sql = "SELECT name, salary FROM employees WHERE department = ?";

        try (PreparedStatement ps = conn.prepareStatement(sql)) {
            ps.setString(1, dept);
            try (ResultSet rs = ps.executeQuery()) {
                System.out.println("\n  Employees in " + dept + ":");
                while (rs.next()) {
                    System.out.println("    " + rs.getString("name") + " - $" + rs.getDouble("salary"));
                }
            }
        }
    }

    // ========================================================
    // 5. UPDATE
    // ========================================================
    static void updateSalary(Connection conn, int id, double newSalary) throws SQLException {
        String sql = "UPDATE employees SET salary = ? WHERE id = ?";

        try (PreparedStatement ps = conn.prepareStatement(sql)) {
            ps.setDouble(1, newSalary);
            ps.setInt(2, id);
            int affected = ps.executeUpdate();
            System.out.println("  Updated " + affected + " row(s), id=" + id + " salary→" + newSalary);
        }
    }

    // ========================================================
    // 6. DELETE
    // ========================================================
    static void deleteEmployee(Connection conn, int id) throws SQLException {
        String sql = "DELETE FROM employees WHERE id = ?";

        try (PreparedStatement ps = conn.prepareStatement(sql)) {
            ps.setInt(1, id);
            int affected = ps.executeUpdate();
            System.out.println("  Deleted " + affected + " row(s), id=" + id);
        }
    }

    // ========================================================
    // 7. TRANSACTIONS
    // ========================================================
    static void transferBudget(Connection conn, int fromId, int toId, double amount)
            throws SQLException {

        conn.setAutoCommit(false);  // start transaction

        try {
            // Deduct from source
            try (PreparedStatement ps = conn.prepareStatement(
                    "UPDATE employees SET salary = salary - ? WHERE id = ?")) {
                ps.setDouble(1, amount);
                ps.setInt(2, fromId);
                ps.executeUpdate();
            }

            // Add to destination
            try (PreparedStatement ps = conn.prepareStatement(
                    "UPDATE employees SET salary = salary + ? WHERE id = ?")) {
                ps.setDouble(1, amount);
                ps.setInt(2, toId);
                ps.executeUpdate();
            }

            conn.commit();  // both succeed → commit
            System.out.println("  Transferred $" + amount + " from id=" + fromId + " to id=" + toId);

        } catch (SQLException e) {
            conn.rollback();  // any failure → rollback both
            System.out.println("  Transaction rolled back: " + e.getMessage());
            throw e;
        } finally {
            conn.setAutoCommit(true);  // restore default
        }
    }

    // ========================================================
    // 8. BATCH OPERATIONS
    // ========================================================
    static void batchInsert(Connection conn, List<String[]> employees) throws SQLException {
        String sql = "INSERT INTO employees (name, department, salary) VALUES (?, ?, ?)";

        conn.setAutoCommit(false);
        try (PreparedStatement ps = conn.prepareStatement(sql)) {
            for (String[] emp : employees) {
                ps.setString(1, emp[0]);
                ps.setString(2, emp[1]);
                ps.setDouble(3, Double.parseDouble(emp[2]));
                ps.addBatch();
            }
            int[] results = ps.executeBatch();
            conn.commit();
            System.out.println("  Batch inserted " + results.length + " rows");
        } catch (SQLException e) {
            conn.rollback();
            throw e;
        } finally {
            conn.setAutoCommit(true);
        }
    }

    // ========================================================
    // 9. DAO PATTERN (Data Access Object)
    // ========================================================

    // Entity
    static class Employee {
        int id;
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
            return "Employee{id=" + id + ", name='" + name + "', dept='" + department + "', salary=" + salary + "}";
        }
    }

    // DAO Interface
    interface EmployeeDAO {
        Employee findById(int id) throws SQLException;
        List<Employee> findAll() throws SQLException;
        int save(Employee emp) throws SQLException;
        void update(Employee emp) throws SQLException;
        void delete(int id) throws SQLException;
    }

    // DAO Implementation
    static class EmployeeDAOImpl implements EmployeeDAO {
        private final Connection conn;

        EmployeeDAOImpl(Connection conn) { this.conn = conn; }

        @Override
        public Employee findById(int id) throws SQLException {
            String sql = "SELECT * FROM employees WHERE id = ?";
            try (PreparedStatement ps = conn.prepareStatement(sql)) {
                ps.setInt(1, id);
                try (ResultSet rs = ps.executeQuery()) {
                    if (rs.next()) return mapRow(rs);
                    return null;
                }
            }
        }

        @Override
        public List<Employee> findAll() throws SQLException {
            List<Employee> list = new ArrayList<>();
            try (Statement stmt = conn.createStatement();
                 ResultSet rs = stmt.executeQuery("SELECT * FROM employees")) {
                while (rs.next()) list.add(mapRow(rs));
            }
            return list;
        }

        @Override
        public int save(Employee emp) throws SQLException {
            return insertEmployee(conn, emp.name, emp.department, emp.salary);
        }

        @Override
        public void update(Employee emp) throws SQLException {
            updateSalary(conn, emp.id, emp.salary);
        }

        @Override
        public void delete(int id) throws SQLException {
            deleteEmployee(conn, id);
        }

        private Employee mapRow(ResultSet rs) throws SQLException {
            Employee e = new Employee(rs.getString("name"), rs.getString("department"), rs.getDouble("salary"));
            e.id = rs.getInt("id");
            return e;
        }
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== JDBC CONCEPTS ===\n");
        System.out.println("  NOTE: This chapter shows the code patterns.");
        System.out.println("  To actually run, you need a JDBC driver JAR.\n");

        System.out.println("  --- JDBC Workflow ---");
        System.out.println("  1. Load driver (automatic since Java 6)");
        System.out.println("  2. Get Connection via DriverManager");
        System.out.println("  3. Create Statement or PreparedStatement");
        System.out.println("  4. Execute query/update");
        System.out.println("  5. Process ResultSet (for queries)");
        System.out.println("  6. Close resources (try-with-resources!)");

        System.out.println("\n  --- SQL Injection Prevention ---");
        System.out.println("  ❌ NEVER: \"SELECT * FROM users WHERE name='\" + input + \"'\"");
        System.out.println("  ✅ ALWAYS: \"SELECT * FROM users WHERE name=?\" + ps.setString(1, input)");

        System.out.println("\n  --- Connection Pooling ---");
        System.out.println("  Creating connections is expensive.");
        System.out.println("  Use connection pools in production:");
        System.out.println("    HikariCP (fastest, recommended)");
        System.out.println("    Apache DBCP");
        System.out.println("    C3P0");

        System.out.println("\n  --- Transaction Isolation Levels ---");
        System.out.println("  READ_UNCOMMITTED → dirty reads possible");
        System.out.println("  READ_COMMITTED   → no dirty reads (default in most DBs)");
        System.out.println("  REPEATABLE_READ  → no non-repeatable reads");
        System.out.println("  SERIALIZABLE     → strictest, slowest");

        System.out.println("\n  --- DAO Pattern ---");
        System.out.println("  Separate data access logic from business logic.");
        System.out.println("  Entity ↔ DAO Interface ↔ DAO Implementation");
        System.out.println("  Makes switching databases easy (just new DAO impl).");

        System.out.println("\n  --- ORM Alternatives ---");
        System.out.println("  Hibernate / JPA  → full ORM, entity mapping");
        System.out.println("  MyBatis          → SQL mapping");
        System.out.println("  jOOQ             → typesafe SQL builder");
        System.out.println("  Spring Data JPA  → repository pattern");

        System.out.println("\n✓ JDBC & Database Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Download SQLite JDBC driver and run all the CRUD operations.
 * 2. Create a "products" table with categories. Write queries with JOINs.
 * 3. Implement the DAO pattern for a "Book" entity.
 * 4. Add connection pooling using HikariCP.
 *
 * NEXT: Chapter 43 — Unit Testing
 */
