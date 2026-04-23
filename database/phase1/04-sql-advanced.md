# 1.4 — SQL Advanced: Query Plans, Stored Procedures, and Optimization

> This is where SQL becomes a performance engineering discipline.
> You're not just writing queries — you're reasoning about HOW the database executes them.

---

## 1. EXPLAIN / EXPLAIN ANALYZE — Reading Query Plans

### Why This Matters

The query planner is the **brain** of the database. It takes your SQL, considers hundreds of possible execution strategies, and picks the one it thinks is cheapest.

EXPLAIN shows you what it chose. If you can't read EXPLAIN output, you're flying blind.

### Basic Usage

```sql
-- EXPLAIN: shows the plan WITHOUT executing
EXPLAIN SELECT * FROM employee WHERE dept_id = 1;

-- EXPLAIN ANALYZE: actually EXECUTES the query and shows real timing
EXPLAIN ANALYZE SELECT * FROM employee WHERE dept_id = 1;

-- Full detail
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM employee WHERE dept_id = 1;

-- JSON format (for programmatic analysis, pgMustard, etc.)
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
SELECT e.first_name, d.name
FROM employee e JOIN department d ON e.dept_id = d.id
WHERE e.salary > 100000;
```

### Reading a Query Plan

```
                                    QUERY PLAN
──────────────────────────────────────────────────────────────────────────
Hash Join  (cost=1.09..2.28 rows=4 width=136) (actual time=0.037..0.042 rows=4 loops=1)
  Hash Cond: (e.dept_id = d.id)
  ->  Seq Scan on employee e  (cost=0.00..1.15 rows=5 width=108) (actual time=0.007..0.009 rows=5 loops=1)
        Filter: (salary > 100000)
        Rows Removed by Filter: 5
  ->  Hash  (cost=1.05..1.05 rows=5 width=36) (actual time=0.012..0.012 rows=5 loops=1)
        Buckets: 1024  Batches: 1  Memory Usage: 9kB
        ->  Seq Scan on department d  (cost=0.00..1.05 rows=5 width=36) (actual time=0.003..0.004 rows=5 loops=1)
Planning Time: 0.125 ms
Execution Time: 0.072 ms
```

**How to read this:**
- Read from **inside out** (deepest indent = first executed)
- `cost=start..total` — estimated cost in arbitrary units (sequential page reads)
  - First number: cost to produce first row
  - Second number: cost to produce ALL rows
- `rows=N` — estimated number of rows
- `actual time=start..total` — real wall clock time in milliseconds
- `rows=N` (after actual) — actual rows produced
- `loops=N` — number of times this node was executed

### Scan Types

```
Seq Scan           — Reads every row in the table (full table scan)
                     Used when: no useful index, or table is small, or >10-15% of rows needed

Index Scan         — Traverses B-tree index, then fetches heap row
                     Used when: selective query, few rows needed

Index Only Scan    — Reads ONLY from the index, never touches the heap
                     Used when: all needed columns are in the index (covering index)
                     Requires: visibility map is up-to-date (vacuumed recently)

Bitmap Index Scan  — Builds a bitmap of heap pages from index, then reads pages
  + Bitmap Heap Scan  Used when: moderate selectivity (too many rows for index scan,
                     too few for seq scan). Can combine multiple indexes with AND/OR.

TID Scan           — Directly fetches by physical tuple ID
                     Rare, mostly internal
```

### Join Types

```
Nested Loop       — For each row in outer, scan inner
                    Best when: one side is small, inner side has index
                    Cost: O(N × M) without index, O(N × log M) with index

Hash Join          — Build hash table on smaller input, probe with larger
                    Best when: equijoin, no useful index, sufficient memory
                    Cost: O(N + M) with O(min(N,M)) memory

Merge Join         — Both inputs sorted, walk through simultaneously
                    Best when: both sides already sorted (index), equijoin
                    Cost: O(N log N + M log M) for sorting, O(N + M) for merge
```

### Identifying Problems in Query Plans

```sql
-- BAD: Seq Scan on a large table with a WHERE clause → missing index
Seq Scan on orders  (cost=... rows=1000000)
  Filter: (customer_id = 42)
  Rows Removed by Filter: 999999    ← 999,999 rows read and discarded!
-- FIX: CREATE INDEX idx_orders_customer ON orders(customer_id);

-- BAD: Nested Loop with Seq Scan on inner → O(N²)
Nested Loop
  ->  Seq Scan on orders (rows=1000000)
  ->  Seq Scan on customer (rows=50000)  ← scanned 50000 rows × 1000000 times!
-- FIX: Add index on the join column

-- BAD: Estimated rows vs actual rows are wildly different → stale statistics
Index Scan on orders  (rows=100) (actual rows=500000)
-- FIX: ANALYZE orders;  (updates statistics)

-- BAD: Sort uses disk instead of memory
Sort  (cost=... rows=1000000)
  Sort Key: created_at
  Sort Method: external merge  Disk: 500MB  ← SLOW!
-- FIX: Increase work_mem, or add an index on created_at
```

### The Statistics System

```sql
-- PostgreSQL stores statistics about your data for the optimizer

-- View table statistics
SELECT * FROM pg_stats WHERE tablename = 'employee';

-- Key stats per column:
--   null_frac       — fraction of NULLs
--   n_distinct      — number of distinct values (-1 = unique)
--   most_common_vals — most frequent values
--   most_common_freqs — frequencies of those values
--   histogram_bounds — value distribution in equal-frequency buckets

-- Update statistics manually
ANALYZE employee;           -- specific table
ANALYZE;                    -- all tables
ALTER TABLE employee ALTER COLUMN dept_id SET STATISTICS 1000;  -- more histogram buckets

-- When statistics lie:
-- Correlated columns: optimizer assumes independence
-- Skewed data: most_common_vals helps but not perfectly
-- Functions in WHERE: optimizer can't estimate selectivity of WHERE ABS(x) > 5
```

---

## 2. Index Strategy

### When to Create Indexes

```sql
-- DO index:
-- 1. Columns in WHERE clauses (especially with = or range conditions)
-- 2. Columns in JOIN conditions
-- 3. Columns in ORDER BY (avoids sort)
-- 4. Columns with high selectivity (many distinct values)
-- 5. Foreign keys (speeds up DELETE on referenced table)

-- DON'T index:
-- 1. Small tables (seq scan is faster — no index overhead)
-- 2. Columns with low selectivity (boolean, status with 3 values)
--    EXCEPTION: partial index on rare values
-- 3. Tables with heavy write load and few reads
-- 4. Columns rarely used in queries
```

### Composite Index Column Ordering

```sql
-- THE rule: Equality columns first, then range columns, then sort columns
-- (ESR rule: Equality, Sort, Range)

-- Query: WHERE dept_id = 1 AND salary > 100000 ORDER BY hire_date
CREATE INDEX idx_emp ON employee(dept_id, hire_date, salary);
-- dept_id (equality) → hire_date (sort) → salary (range)

-- The "leftmost prefix" rule:
-- Index on (A, B, C) is useful for:
--   WHERE A = 1                    ✓ (uses A)
--   WHERE A = 1 AND B = 2         ✓ (uses A, B)
--   WHERE A = 1 AND B = 2 AND C = 3  ✓ (uses all)
--   WHERE B = 2                    ✗ (can't skip A!)
--   WHERE A = 1 AND C = 3         △ (uses A, then scans for C)
```

### Covering Indexes (Index-Only Scans)

```sql
-- A covering index includes ALL columns the query needs
-- The database reads ONLY the index, never touches the table heap

-- Query: SELECT dept_id, salary FROM employee WHERE dept_id = 1
CREATE INDEX idx_covering ON employee(dept_id) INCLUDE (salary);
-- INCLUDE columns are stored in leaf pages but not used for searching

-- PostgreSQL pre-13: use a regular composite index
CREATE INDEX idx_covering ON employee(dept_id, salary);
-- Works as covering, but salary is also part of the search tree (slightly larger)
```

### Partial Indexes

```sql
-- Index only a subset of rows
CREATE INDEX idx_active_employees ON employee(email) WHERE is_active = TRUE;
-- Smaller index, only used when query includes WHERE is_active = TRUE

-- Great for:
-- Rare conditions: WHERE status = 'error' (1% of rows)
CREATE INDEX idx_error_orders ON orders(created_at) WHERE status = 'error';

-- Soft deletes:
CREATE INDEX idx_undeleted ON users(email) WHERE deleted_at IS NULL;
```

### Expression Indexes

```sql
-- Index the result of an expression
CREATE INDEX idx_lower_email ON employee(LOWER(email));
-- Now this query uses the index:
SELECT * FROM employee WHERE LOWER(email) = 'alice@co.com';

-- Without the expression index, the database can't use a regular index on email
-- because LOWER(email) is a different value than email.

CREATE INDEX idx_year_hired ON employee(EXTRACT(YEAR FROM hire_date));
SELECT * FROM employee WHERE EXTRACT(YEAR FROM hire_date) = 2023;
```

---

## 3. Stored Procedures and Functions

### Functions (PL/pgSQL)

```sql
-- Simple function
CREATE OR REPLACE FUNCTION get_annual_salary(monthly DECIMAL)
RETURNS DECIMAL
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN monthly * 12;
END;
$$;

SELECT first_name, get_annual_salary(salary / 12) FROM employee;

-- Function with multiple return values
CREATE OR REPLACE FUNCTION get_dept_stats(p_dept_id INTEGER)
RETURNS TABLE(emp_count BIGINT, avg_salary DECIMAL, max_salary DECIMAL)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT COUNT(*), AVG(salary)::DECIMAL, MAX(salary)::DECIMAL
    FROM employee
    WHERE dept_id = p_dept_id;
END;
$$;

SELECT * FROM get_dept_stats(1);

-- Function with logic
CREATE OR REPLACE FUNCTION classify_salary(s DECIMAL)
RETURNS TEXT
LANGUAGE plpgsql
IMMUTABLE  -- pure function (same input → same output), helps optimizer
AS $$
BEGIN
    IF s >= 150000 THEN RETURN 'Executive';
    ELSIF s >= 100000 THEN RETURN 'Senior';
    ELSIF s >= 70000 THEN RETURN 'Mid';
    ELSE RETURN 'Junior';
    END IF;
END;
$$;

SELECT first_name, salary, classify_salary(salary) FROM employee;

-- SQL function (simpler, often faster — can be inlined by optimizer)
CREATE OR REPLACE FUNCTION active_employees()
RETURNS SETOF employee
LANGUAGE sql
STABLE
AS $$
    SELECT * FROM employee WHERE is_active = TRUE;
$$;

SELECT * FROM active_employees() WHERE dept_id = 1;
```

### Stored Procedures (PostgreSQL 11+)

```sql
-- Procedures can manage TRANSACTIONS (functions can't)
CREATE OR REPLACE PROCEDURE transfer_employee(
    p_employee_id INTEGER,
    p_new_dept_id INTEGER
)
LANGUAGE plpgsql
AS $$
BEGIN
    -- Update employee's department
    UPDATE employee SET dept_id = p_new_dept_id WHERE id = p_employee_id;

    -- Log the transfer
    INSERT INTO transfer_log (employee_id, new_dept_id, transferred_at)
    VALUES (p_employee_id, p_new_dept_id, NOW());

    -- Could COMMIT or ROLLBACK here
    COMMIT;
END;
$$;

CALL transfer_employee(2, 3);
```

### Function Volatility Categories

```sql
IMMUTABLE  -- Never changes for same input. Can be pre-evaluated, used in indexes.
           -- Example: LOWER(), mathematical functions

STABLE     -- Returns same result within a single query/transaction.
           -- Can read database but won't modify it.
           -- Example: NOW(), current_user

VOLATILE   -- Can return different results each call. May have side effects.
           -- Default. Example: random(), nextval()
```

---

## 4. Triggers

```sql
-- Audit trigger: log all changes to employee table
CREATE TABLE employee_audit (
    id          SERIAL PRIMARY KEY,
    employee_id INTEGER,
    action      TEXT,
    old_data    JSONB,
    new_data    JSONB,
    changed_at  TIMESTAMPTZ DEFAULT NOW(),
    changed_by  TEXT DEFAULT current_user
);

CREATE OR REPLACE FUNCTION employee_audit_trigger()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO employee_audit (employee_id, action, new_data)
        VALUES (NEW.id, 'INSERT', to_jsonb(NEW));
        RETURN NEW;

    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO employee_audit (employee_id, action, old_data, new_data)
        VALUES (NEW.id, 'UPDATE', to_jsonb(OLD), to_jsonb(NEW));
        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO employee_audit (employee_id, action, old_data)
        VALUES (OLD.id, 'DELETE', to_jsonb(OLD));
        RETURN OLD;
    END IF;
END;
$$;

CREATE TRIGGER trg_employee_audit
AFTER INSERT OR UPDATE OR DELETE ON employee
FOR EACH ROW EXECUTE FUNCTION employee_audit_trigger();

-- Trigger types:
-- BEFORE  — can modify NEW row before it's written (validation, transformation)
-- AFTER   — runs after the write (auditing, notifications)
-- INSTEAD OF — replaces the operation (used on views)
-- FOR EACH ROW    — fires once per row affected
-- FOR EACH STATEMENT — fires once per statement (even if 0 rows affected)

-- Practical trigger: auto-update updated_at
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON employee
FOR EACH ROW EXECUTE FUNCTION update_timestamp();
```

**Trigger caveats (god-level awareness):**
- Triggers fire ONCE PER ROW — bulk inserts become slow
- Triggers are invisible in application code — debugging nightmare
- Trigger chains can cascade unpredictably
- `TRUNCATE` does NOT fire row-level triggers
- Prefer application-level logic for complex business rules
- Use triggers for: audit logs, updated_at timestamps, simple derived data

---

## 5. Full-Text Search (PostgreSQL)

```sql
-- tsvector: processed document (normalized, stemmed, stop-words removed)
-- tsquery: search query

SELECT to_tsvector('english', 'The quick brown fox jumps over the lazy dog');
-- Result: 'brown':3 'dog':9 'fox':4 'jump':5 'lazi':8 'quick':2

SELECT to_tsquery('english', 'quick & fox');
-- Result: 'quick' & 'fox'

-- Search
SELECT * FROM articles
WHERE to_tsvector('english', title || ' ' || body) @@ to_tsquery('english', 'database & optimization');

-- For performance: store the tsvector as a column and index it
ALTER TABLE articles ADD COLUMN search_vector tsvector;

UPDATE articles SET search_vector =
    setweight(to_tsvector('english', COALESCE(title, '')), 'A') ||  -- title weight A (highest)
    setweight(to_tsvector('english', COALESCE(body, '')), 'B');     -- body weight B

CREATE INDEX idx_articles_search ON articles USING GIN(search_vector);

-- Auto-update with trigger
CREATE TRIGGER articles_search_update
BEFORE INSERT OR UPDATE ON articles
FOR EACH ROW EXECUTE FUNCTION
    tsvector_update_trigger(search_vector, 'pg_catalog.english', title, body);

-- Ranking results
SELECT title,
    ts_rank(search_vector, to_tsquery('english', 'database')) AS rank
FROM articles
WHERE search_vector @@ to_tsquery('english', 'database')
ORDER BY rank DESC;

-- Query syntax:
-- &  = AND
-- |  = OR
-- !  = NOT
-- <-> = followed by (phrase search)
-- <2> = within 2 words

to_tsquery('english', 'database & (optimization | tuning)')
to_tsquery('english', 'quick <-> fox')  -- "quick fox" as phrase

-- Fuzzy matching with trigrams
CREATE EXTENSION pg_trgm;
CREATE INDEX idx_name_trgm ON employee USING GIN (first_name gin_trgm_ops);

SELECT * FROM employee WHERE first_name % 'Alce';  -- similarity match
SELECT first_name, similarity(first_name, 'Alce') AS sim
FROM employee
WHERE first_name % 'Alce'
ORDER BY sim DESC;
```

---

## 6. Dynamic SQL and Prepared Statements

### Prepared Statements

```sql
-- Server-side prepared statements
PREPARE emp_by_dept(INTEGER) AS
SELECT * FROM employee WHERE dept_id = $1;

EXECUTE emp_by_dept(1);
EXECUTE emp_by_dept(2);

DEALLOCATE emp_by_dept;

-- Benefits:
-- 1. Parse + plan ONCE, execute many times (faster for repeated queries)
-- 2. Prevents SQL injection (parameters are not part of the SQL text)

-- In application code (most common usage):
-- Python: cursor.execute("SELECT * FROM employee WHERE id = %s", (42,))
-- Node:   pool.query("SELECT * FROM employee WHERE id = $1", [42])
-- Java:   PreparedStatement ps = conn.prepareStatement("SELECT * FROM employee WHERE id = ?");
```

### Dynamic SQL in PL/pgSQL

```sql
-- EXECUTE for dynamic queries
CREATE OR REPLACE FUNCTION search_employees(
    p_column TEXT,
    p_value TEXT
)
RETURNS SETOF employee
LANGUAGE plpgsql
AS $$
BEGIN
    -- DANGER: p_column is NOT parameterizable (it's an identifier, not a value)
    -- Must validate it!
    IF p_column NOT IN ('first_name', 'last_name', 'email') THEN
        RAISE EXCEPTION 'Invalid column: %', p_column;
    END IF;

    RETURN QUERY EXECUTE
        format('SELECT * FROM employee WHERE %I = $1', p_column)
        USING p_value;
    -- %I = safe identifier quoting
    -- $1 + USING = safe value parameterization
END;
$$;

-- format() specifiers:
-- %I = identifier (quoted as needed): 'my column' → "my column"
-- %L = literal (quoted and escaped): O'Brien → 'O''Brien'
-- %s = simple string (NO escaping — NEVER use for user input!)
```

---

## 7. Table Partitioning

```sql
-- Declarative partitioning (PostgreSQL 10+)

-- Range partitioning (most common — for time-series data)
CREATE TABLE orders (
    id          BIGSERIAL,
    customer_id INTEGER NOT NULL,
    total       DECIMAL(10,2) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, created_at)  -- partition key must be in PK
) PARTITION BY RANGE (created_at);

-- Create partitions
CREATE TABLE orders_2025 PARTITION OF orders
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
CREATE TABLE orders_2026 PARTITION OF orders
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
CREATE TABLE orders_default PARTITION OF orders DEFAULT;

-- Queries automatically route to correct partition (partition pruning):
SELECT * FROM orders WHERE created_at >= '2026-01-01' AND created_at < '2026-04-01';
-- Only scans orders_2026!

-- List partitioning (for categorical data)
CREATE TABLE events (
    id      BIGSERIAL,
    type    TEXT NOT NULL,
    data    JSONB,
    PRIMARY KEY (id, type)
) PARTITION BY LIST (type);

CREATE TABLE events_click PARTITION OF events FOR VALUES IN ('click');
CREATE TABLE events_view PARTITION OF events FOR VALUES IN ('view', 'impression');
CREATE TABLE events_other PARTITION OF events DEFAULT;

-- Hash partitioning (for even distribution)
CREATE TABLE sessions (
    id      UUID NOT NULL,
    user_id INTEGER NOT NULL,
    data    JSONB,
    PRIMARY KEY (id, user_id)
) PARTITION BY HASH (user_id);

CREATE TABLE sessions_0 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE sessions_1 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE sessions_2 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE sessions_3 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 3);

-- Benefits of partitioning:
-- 1. Partition pruning: queries only scan relevant partitions
-- 2. Bulk deletes: DROP TABLE orders_2020 (instant, vs deleting millions of rows)
-- 3. Parallel scans across partitions
-- 4. Different storage/tablespace per partition
-- 5. Independent VACUUM per partition

-- When to partition:
-- Tables > 100 million rows with time-based access patterns
-- Need to efficiently purge old data
-- DON'T partition small tables — overhead isn't worth it
```

---

## 8. SQL Standards Evolution

| Standard | Key Features |
|----------|-------------|
| SQL-86 | Basic SELECT, INSERT, UPDATE, DELETE |
| SQL-92 | JOINs (INNER, OUTER, CROSS), CASE, CAST, subqueries, string operations |
| SQL:1999 | Recursive CTEs, triggers, OLAP functions, user-defined types, BOOLEAN |
| SQL:2003 | Window functions, MERGE, XML, SEQUENCE, BIGINT, MULTISET |
| SQL:2008 | TRUNCATE, FETCH FIRST, INSTEAD OF triggers |
| SQL:2011 | Temporal tables (system-time, application-time) |
| SQL:2016 | JSON support, row pattern matching (MATCH_RECOGNIZE) |
| SQL:2023 | Property graph queries, JSON enhancements, multi-dimensional arrays |

---

## 9. Practice Exercises

**E1.** Write a query and read its EXPLAIN ANALYZE output. Identify if it's doing a Seq Scan where an Index Scan would be better.

<details><summary>Solution</summary>

```sql
-- First, without index:
EXPLAIN ANALYZE SELECT * FROM employee WHERE salary > 120000;
-- Should show: Seq Scan (because no index on salary)

-- Add index:
CREATE INDEX idx_emp_salary ON employee(salary);

-- Now:
EXPLAIN ANALYZE SELECT * FROM employee WHERE salary > 120000;
-- May still show Seq Scan on small table (optimizer knows it's cheaper)
-- On a large table, would show: Index Scan using idx_emp_salary
```
</details>

**E2.** Create a function that returns the Nth highest salary per department.

<details><summary>Solution</summary>

```sql
CREATE OR REPLACE FUNCTION nth_salary(p_dept_id INTEGER, p_n INTEGER)
RETURNS DECIMAL
LANGUAGE sql
STABLE
AS $$
    SELECT salary FROM employee
    WHERE dept_id = p_dept_id
    ORDER BY salary DESC
    OFFSET p_n - 1
    LIMIT 1;
$$;

SELECT nth_salary(1, 2);  -- 2nd highest in Engineering
```
</details>

**E3.** Create an audit trigger that logs only CHANGED columns on UPDATE.

<details><summary>Solution</summary>

```sql
CREATE OR REPLACE FUNCTION audit_changes()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    changes JSONB := '{}'::JSONB;
    old_val JSONB := to_jsonb(OLD);
    new_val JSONB := to_jsonb(NEW);
    key TEXT;
BEGIN
    FOR key IN SELECT jsonb_object_keys(old_val) LOOP
        IF old_val->key IS DISTINCT FROM new_val->key THEN
            changes := changes || jsonb_build_object(
                key, jsonb_build_object('old', old_val->key, 'new', new_val->key)
            );
        END IF;
    END LOOP;

    IF changes != '{}'::JSONB THEN
        INSERT INTO employee_audit (employee_id, action, new_data, changed_at)
        VALUES (NEW.id, 'UPDATE', changes, NOW());
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_audit_changes
AFTER UPDATE ON employee
FOR EACH ROW EXECUTE FUNCTION audit_changes();
```
</details>

---

## Key Takeaways

1. **EXPLAIN ANALYZE is your X-ray vision.** Run it on every slow query. Learn to read cost, rows, and scan types.
2. **Index strategy**: equality first, then sort, then range. Composite index column order MATTERS.
3. **Covering indexes** (INCLUDE) enable index-only scans — zero heap access.
4. **Partial indexes** for rare conditions are incredibly efficient.
5. **Prepared statements** are a security AND performance feature.
6. **Partitioning** is for big tables with predictable access patterns (especially time-series).
7. **Triggers** are powerful but invisible — use sparingly, prefer application logic for business rules.
8. **Full-text search** in PostgreSQL is production-grade — you may not need Elasticsearch.

---

Next: [05-normalization-and-schema-design.md](05-normalization-and-schema-design.md) →
