# 1.2 — SQL Beginner: The Complete Foundation

> SQL is how you **talk** to a relational database.
> It looks simple. It's not. The difference between a beginner and a god is understanding what happens UNDER each statement.

---

## 0. Setup: Get a Database Running

```bash
# Option 1: PostgreSQL via Docker (recommended)
docker run --name pg -e POSTGRES_PASSWORD=learn -p 5432:5432 -d postgres:16

# Connect
psql -h localhost -U postgres
# Or use: pgcli, DBeaver, DataGrip

# Option 2: SQLite (zero setup)
sqlite3 learn.db
```

Everything below uses **PostgreSQL syntax**. Differences with MySQL/SQLite noted where important.

---

## 1. Data Definition Language (DDL)

DDL defines the **structure** (schema) of your database.

### CREATE TABLE

```sql
CREATE TABLE department (
    id          SERIAL PRIMARY KEY,        -- auto-incrementing integer (PG)
    name        VARCHAR(100) NOT NULL,
    budget      DECIMAL(15, 2) DEFAULT 0,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE employee (
    id          SERIAL PRIMARY KEY,
    first_name  VARCHAR(50) NOT NULL,
    last_name   VARCHAR(50) NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    salary      DECIMAL(10, 2) CHECK (salary > 0),
    hire_date   DATE NOT NULL DEFAULT CURRENT_DATE,
    dept_id     INTEGER REFERENCES department(id)  -- foreign key
                    ON DELETE SET NULL
                    ON UPDATE CASCADE,
    is_active   BOOLEAN DEFAULT TRUE
);
```

**What's actually happening under the hood:**
- The database creates a **heap file** (or clustered index) to store rows
- `PRIMARY KEY` creates a **unique B+ tree index** on that column
- `UNIQUE` creates another **B+ tree index**
- `REFERENCES` stores a constraint that's checked on INSERT/UPDATE/DELETE
- `SERIAL` creates a **sequence object** that auto-generates values

### PostgreSQL vs MySQL Data Type Differences

| Concept | PostgreSQL | MySQL |
|---------|-----------|-------|
| Auto-increment | `SERIAL` / `GENERATED ALWAYS AS IDENTITY` | `AUTO_INCREMENT` |
| Boolean | `BOOLEAN` (true/false) | `TINYINT(1)` (0/1), `BOOLEAN` alias |
| Variable string | `VARCHAR(n)` / `TEXT` (no perf diff) | `VARCHAR(n)` / `TEXT` (different) |
| JSON | `JSON` / `JSONB` (binary, indexable) | `JSON` (since 5.7) |
| UUID | `UUID` type | `CHAR(36)` or `BINARY(16)` |
| Array | `INTEGER[]`, `TEXT[]` native | Not supported |

### ALTER TABLE

```sql
-- Add a column
ALTER TABLE employee ADD COLUMN phone VARCHAR(20);

-- Drop a column
ALTER TABLE employee DROP COLUMN phone;

-- Rename a column
ALTER TABLE employee RENAME COLUMN first_name TO fname;

-- Change column type
ALTER TABLE employee ALTER COLUMN salary TYPE NUMERIC(12, 2);

-- Add a constraint
ALTER TABLE employee ADD CONSTRAINT salary_positive CHECK (salary > 0);

-- Drop a constraint
ALTER TABLE employee DROP CONSTRAINT salary_positive;

-- Add an index (not technically DDL, but related)
CREATE INDEX idx_employee_dept ON employee(dept_id);
```

**Critical knowledge:**
- `ALTER TABLE` takes **locks**. On a table with millions of rows:
  - Adding a column with a DEFAULT → PostgreSQL 11+ is instant (before 11, it rewrote the entire table!)
  - Adding a NOT NULL constraint → scans the entire table
  - Creating an index → `CREATE INDEX CONCURRENTLY` avoids locking (PostgreSQL)
- In production, you NEVER run ALTER TABLE casually. You use tools like `gh-ost` (MySQL) or `pg_repack` (PostgreSQL).

### DROP and TRUNCATE

```sql
-- Delete the table and all its data permanently
DROP TABLE employee;             -- error if doesn't exist
DROP TABLE IF EXISTS employee;   -- safe
DROP TABLE employee CASCADE;     -- also drops dependent objects (views, FKs)

-- Delete all rows but keep the table structure
TRUNCATE TABLE employee;                  -- much faster than DELETE
TRUNCATE TABLE employee RESTART IDENTITY; -- reset auto-increment
TRUNCATE TABLE employee CASCADE;          -- also truncate referencing tables
```

**DROP vs TRUNCATE vs DELETE:**
| | DROP | TRUNCATE | DELETE |
|---|------|----------|--------|
| Removes structure? | Yes | No | No |
| Removes data? | Yes | Yes | Yes (with WHERE) |
| Logged? | Yes | Minimal (no per-row log) | Full (every row logged) |
| Can rollback? | Yes (in PG) | Yes (in PG) | Yes |
| Fires triggers? | No | No | Yes |
| Speed on 10M rows | Instant | Instant | Very slow |

---

## 2. Data Types — Deep Understanding

### Numeric Types

```sql
-- Exact numeric (use for money, quantities)
SMALLINT          -- 2 bytes, -32768 to 32767
INTEGER (INT)     -- 4 bytes, -2.1B to 2.1B
BIGINT            -- 8 bytes, -9.2 quintillion to 9.2 quintillion
DECIMAL(p, s)     -- exact, arbitrary precision (p=total digits, s=decimal places)
NUMERIC(p, s)     -- same as DECIMAL in PostgreSQL

-- Approximate numeric (use for scientific calculations)
REAL              -- 4 bytes, 6 decimal digits precision
DOUBLE PRECISION  -- 8 bytes, 15 decimal digits precision

-- NEVER use FLOAT/DOUBLE for money!
SELECT 0.1 + 0.2;                          -- 0.3 (DECIMAL)
SELECT 0.1::DOUBLE PRECISION + 0.2::DOUBLE PRECISION; -- 0.30000000000000004
```

**God-level knowledge:**
- `DECIMAL(10,2)` stores EXACTLY — no floating point errors. It's stored as a scaled integer internally.
- `BIGINT` is 8 bytes. `DECIMAL(20,0)` uses MORE storage — it's variable-length.
- For `id` columns: use `BIGINT` not `INTEGER`. You WILL run out of 2.1B on a busy system.
- PostgreSQL's `SERIAL` is just syntax sugar for: create a sequence + set default + grant usage.
- Modern PostgreSQL: prefer `GENERATED ALWAYS AS IDENTITY` over `SERIAL`.

### String Types

```sql
CHAR(n)          -- fixed-length, padded with spaces. Almost never use this.
VARCHAR(n)       -- variable-length, max n characters
TEXT             -- variable-length, unlimited

-- In PostgreSQL: VARCHAR and TEXT are identical in performance.
-- VARCHAR(n) just adds a length check constraint.
-- In MySQL: VARCHAR has a 65535 byte limit, TEXT is stored off-page.
```

### Date/Time Types

```sql
DATE             -- '2026-04-17' (4 bytes)
TIME             -- '14:30:00' (8 bytes)
TIMESTAMP        -- '2026-04-17 14:30:00' (8 bytes, no timezone)
TIMESTAMPTZ      -- '2026-04-17 14:30:00+05:30' (8 bytes, WITH timezone)
INTERVAL         -- '1 year 2 months 3 days 4 hours'

-- ALWAYS use TIMESTAMPTZ for timestamps. ALWAYS.
-- TIMESTAMP without timezone is a bug waiting to happen.

-- PostgreSQL stores TIMESTAMPTZ as microseconds since 2000-01-01 UTC.
-- It converts to/from the session's timezone on input/output.

SELECT NOW();                            -- current timestamp with tz
SELECT CURRENT_DATE;                     -- today's date
SELECT CURRENT_TIMESTAMP;               -- same as NOW()
SELECT '2026-04-17'::DATE + INTERVAL '30 days';  -- date arithmetic
SELECT AGE('2026-04-17', '2000-01-01'); -- interval between dates
SELECT EXTRACT(YEAR FROM NOW());        -- get year component
```

### Boolean

```sql
BOOLEAN   -- true, false, NULL

-- PostgreSQL accepts: true, 't', 'yes', 'y', '1', 'on' -> TRUE
--                     false, 'f', 'no', 'n', '0', 'off' -> FALSE
```

### JSON

```sql
-- PostgreSQL has two JSON types:
JSON     -- stores as text, re-parsed every access
JSONB    -- stores as binary, faster access, indexable. ALWAYS USE THIS.

CREATE TABLE events (
    id    SERIAL PRIMARY KEY,
    data  JSONB NOT NULL
);

INSERT INTO events (data) VALUES ('{"type": "click", "x": 100, "y": 200}');

-- Access JSON fields
SELECT data->>'type' FROM events;           -- 'click' (as text)
SELECT data->'x' FROM events;               -- 100 (as JSON)
SELECT data#>>'{nested,key}' FROM events;   -- nested access

-- JSONB operators
SELECT * FROM events WHERE data @> '{"type": "click"}';  -- contains
SELECT * FROM events WHERE data ? 'type';                  -- key exists
SELECT * FROM events WHERE data->>'type' = 'click';        -- equality

-- GIN index on JSONB (indexes all keys and values!)
CREATE INDEX idx_events_data ON events USING GIN (data);
```

### UUID

```sql
-- PostgreSQL
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL
);

-- UUIDs are 128-bit, stored as 16 bytes.
-- MUCH better than auto-increment for distributed systems (no coordination needed).
-- But: worse for B-tree index performance (random, not sequential).
-- Solution: UUIDv7 (time-ordered) — best of both worlds.
```

### Arrays (PostgreSQL-specific)

```sql
CREATE TABLE posts (
    id    SERIAL PRIMARY KEY,
    title TEXT,
    tags  TEXT[]    -- array of text
);

INSERT INTO posts (title, tags) VALUES ('DB Post', ARRAY['database', 'sql', 'postgresql']);
INSERT INTO posts (title, tags) VALUES ('DB Post', '{"database", "sql"}');  -- alternative syntax

SELECT * FROM posts WHERE 'sql' = ANY(tags);       -- contains element
SELECT * FROM posts WHERE tags @> ARRAY['sql'];     -- contains array
SELECT * FROM posts WHERE tags && ARRAY['sql','go']; -- overlaps

-- GIN index for array lookups
CREATE INDEX idx_posts_tags ON posts USING GIN (tags);
```

---

## 3. Data Manipulation Language (DML)

### INSERT

```sql
-- Single row
INSERT INTO department (name, budget) VALUES ('Engineering', 1000000);

-- Multiple rows (much faster than separate inserts!)
INSERT INTO department (name, budget) VALUES
    ('Sales', 500000),
    ('Marketing', 750000),
    ('HR', 300000);

-- Insert from a query
INSERT INTO department_archive (name, budget)
SELECT name, budget FROM department WHERE budget < 400000;

-- RETURNING clause (PostgreSQL) — get the inserted data back
INSERT INTO department (name, budget) VALUES ('Research', 2000000)
RETURNING id, name;  -- returns the generated id!

-- INSERT ... ON CONFLICT (UPSERT) — PostgreSQL
INSERT INTO employee (email, first_name, last_name, salary, dept_id)
VALUES ('alice@co.com', 'Alice', 'Smith', 120000, 1)
ON CONFLICT (email) DO UPDATE SET
    salary = EXCLUDED.salary,          -- EXCLUDED refers to the attempted insert row
    first_name = EXCLUDED.first_name;

-- MySQL equivalent:
-- INSERT ... ON DUPLICATE KEY UPDATE salary = VALUES(salary);
```

**Performance knowledge:**
- Single INSERT: ~0.1-1ms per row
- Multi-row INSERT: much faster (1 round-trip, 1 WAL flush)
- COPY (PostgreSQL) / LOAD DATA (MySQL): fastest bulk load (10-100x faster than INSERT)
- INSERT triggers fire ONCE PER ROW — they kill bulk insert performance

```sql
-- Fastest bulk load in PostgreSQL
COPY employee (first_name, last_name, email, salary, dept_id)
FROM '/path/to/data.csv' WITH (FORMAT csv, HEADER true);
```

### UPDATE

```sql
-- Update specific rows
UPDATE employee SET salary = salary * 1.10 WHERE dept_id = 1;

-- Update with a join (PostgreSQL syntax)
UPDATE employee e
SET salary = salary * 1.15
FROM department d
WHERE e.dept_id = d.id AND d.name = 'Engineering';

-- RETURNING clause
UPDATE employee SET salary = salary * 1.10
WHERE id = 42
RETURNING id, salary AS new_salary;
```

**What happens internally on UPDATE (PostgreSQL):**
1. Find the row(s) matching WHERE clause
2. DON'T modify the old row — mark it as "dead" (set xmax)
3. INSERT a new version of the row with updated values
4. Update all indexes that reference this row (unless HOT update applies)
5. The old row version stays until VACUUM cleans it up

This is MVCC! Updates are actually INSERT + DELETE internally.

### DELETE

```sql
-- Delete specific rows
DELETE FROM employee WHERE is_active = FALSE;

-- Delete with a subquery
DELETE FROM employee
WHERE dept_id IN (SELECT id FROM department WHERE budget < 100000);

-- Delete with join (PostgreSQL)
DELETE FROM employee e
USING department d
WHERE e.dept_id = d.id AND d.name = 'Defunct';

-- RETURNING clause
DELETE FROM employee WHERE id = 42 RETURNING *;

-- Delete all rows (use TRUNCATE instead for speed)
DELETE FROM employee;
```

### MERGE / UPSERT (SQL:2003 standard)

```sql
-- PostgreSQL 15+ supports MERGE
MERGE INTO employee AS target
USING new_employee_data AS source
ON target.email = source.email
WHEN MATCHED THEN
    UPDATE SET salary = source.salary, first_name = source.first_name
WHEN NOT MATCHED THEN
    INSERT (first_name, last_name, email, salary, dept_id)
    VALUES (source.first_name, source.last_name, source.email, source.salary, source.dept_id);
```

---

## 4. Data Query Language (DQL) — SELECT

### The Logical Order of SQL Execution

This is **THE most important thing** to understand about SQL. The clauses execute in this order (NOT the order you write them):

```
1. FROM        ← Which tables? Cross joins happen here.
2. JOIN        ← Combine tables. ON conditions applied.
3. WHERE       ← Filter rows (before grouping).
4. GROUP BY    ← Group rows into buckets.
5. HAVING      ← Filter groups (after grouping).
6. SELECT      ← Choose columns, compute expressions.
7. DISTINCT    ← Remove duplicates.
8. ORDER BY    ← Sort results.
9. LIMIT/OFFSET ← Paginate.
```

This is why you CANNOT use a column alias in WHERE:
```sql
-- WRONG: WHERE filters before SELECT computes the alias
SELECT salary * 12 AS annual_salary FROM employee WHERE annual_salary > 100000;

-- RIGHT: repeat the expression, or use a subquery/CTE
SELECT salary * 12 AS annual_salary FROM employee WHERE salary * 12 > 100000;
```

### Basic SELECT

```sql
-- All columns
SELECT * FROM employee;

-- Specific columns
SELECT first_name, last_name, salary FROM employee;

-- Expressions and aliases
SELECT
    first_name || ' ' || last_name AS full_name,     -- string concatenation (PG)
    salary * 12 AS annual_salary,
    ROUND(salary / 160, 2) AS hourly_rate
FROM employee;

-- DISTINCT — remove duplicate rows
SELECT DISTINCT dept_id FROM employee;

-- DISTINCT ON (PostgreSQL-specific) — first row per group
SELECT DISTINCT ON (dept_id) dept_id, first_name, salary
FROM employee
ORDER BY dept_id, salary DESC;  -- gets highest-paid per department
```

### WHERE — Filtering Rows

```sql
-- Comparison operators
SELECT * FROM employee WHERE salary > 100000;
SELECT * FROM employee WHERE dept_id != 1;    -- also: <> for not-equal
SELECT * FROM employee WHERE hire_date >= '2025-01-01';

-- Logical operators (AND, OR, NOT)
SELECT * FROM employee
WHERE salary > 80000 AND dept_id = 1;

SELECT * FROM employee
WHERE dept_id = 1 OR dept_id = 2;

SELECT * FROM employee
WHERE NOT is_active;

-- IN — match any value in a list
SELECT * FROM employee WHERE dept_id IN (1, 2, 3);

-- BETWEEN — inclusive range
SELECT * FROM employee WHERE salary BETWEEN 80000 AND 120000;
-- Equivalent to: salary >= 80000 AND salary <= 120000

-- LIKE — pattern matching
SELECT * FROM employee WHERE last_name LIKE 'S%';      -- starts with S
SELECT * FROM employee WHERE email LIKE '%@gmail.com';  -- ends with
SELECT * FROM employee WHERE first_name LIKE '_a%';     -- second char is 'a'
-- % = any characters, _ = exactly one character

-- ILIKE — case-insensitive LIKE (PostgreSQL)
SELECT * FROM employee WHERE first_name ILIKE 'alice';

-- IS NULL / IS NOT NULL
SELECT * FROM employee WHERE dept_id IS NULL;
-- NEVER use = NULL or != NULL — they ALWAYS return NULL (unknown).

-- ANY / ALL with subqueries
SELECT * FROM employee
WHERE salary > ALL (SELECT salary FROM employee WHERE dept_id = 2);
-- Salary greater than EVERY salary in dept 2
```

### NULL — The Billion Dollar Mistake

```sql
-- NULL is NOT a value. It means "unknown" or "missing".
-- NULL compared with ANYTHING is NULL (unknown), not TRUE or FALSE.

SELECT NULL = NULL;      -- NULL (not TRUE!)
SELECT NULL != NULL;     -- NULL (not FALSE!)
SELECT NULL > 5;         -- NULL
SELECT NULL AND TRUE;    -- NULL
SELECT NULL OR TRUE;     -- TRUE  (because TRUE OR anything = TRUE)
SELECT NULL OR FALSE;    -- NULL
SELECT NOT NULL;         -- NULL

-- COALESCE — replace NULL with a default
SELECT COALESCE(dept_id, 0) FROM employee;  -- if dept_id is NULL, use 0
SELECT COALESCE(phone, email, 'no contact') FROM employee; -- first non-null

-- NULLIF — return NULL if two values are equal
SELECT NULLIF(salary, 0);  -- returns NULL if salary is 0 (avoids division by zero)
SELECT total / NULLIF(count, 0);  -- safe division

-- NULL in aggregates
SELECT AVG(salary) FROM employee;  -- NULLs are IGNORED in aggregates
SELECT COUNT(*) FROM employee;     -- counts all rows INCLUDING null
SELECT COUNT(dept_id) FROM employee; -- counts only non-null dept_id values
```

### JOINs

```sql
-- Sample data for examples
-- department: (1, Engineering), (2, Sales), (3, Marketing), (4, Research)
-- employee: Alice(dept 1), Bob(dept 1), Carol(dept 2), Dave(dept NULL)

-- INNER JOIN — only matching rows from both tables
SELECT e.first_name, d.name AS dept_name
FROM employee e
INNER JOIN department d ON e.dept_id = d.id;
-- Result: Alice-Engineering, Bob-Engineering, Carol-Sales
-- Dave excluded (NULL dept_id), Research excluded (no employees)

-- LEFT (OUTER) JOIN — all rows from left, matching rows from right
SELECT e.first_name, d.name AS dept_name
FROM employee e
LEFT JOIN department d ON e.dept_id = d.id;
-- Result: Alice-Engineering, Bob-Engineering, Carol-Sales, Dave-NULL
-- Dave included (with NULL dept_name), Research still excluded

-- RIGHT (OUTER) JOIN — all rows from right, matching rows from left
SELECT e.first_name, d.name AS dept_name
FROM employee e
RIGHT JOIN department d ON e.dept_id = d.id;
-- Result: Alice-Engineering, Bob-Engineering, Carol-Sales,
--         NULL-Marketing, NULL-Research

-- FULL (OUTER) JOIN — all rows from both tables
SELECT e.first_name, d.name AS dept_name
FROM employee e
FULL JOIN department d ON e.dept_id = d.id;
-- Result: Alice-Engineering, Bob-Engineering, Carol-Sales,
--         Dave-NULL, NULL-Marketing, NULL-Research

-- CROSS JOIN — every combination (cartesian product)
SELECT e.first_name, d.name
FROM employee e
CROSS JOIN department d;
-- If 4 employees × 4 departments = 16 rows

-- Self-join — join a table with itself
-- Find employees who earn more than their manager
SELECT e.first_name AS employee, m.first_name AS manager
FROM employee e
JOIN employee m ON e.manager_id = m.id
WHERE e.salary > m.salary;
```

**Join internals — what the database actually does:**

| Algorithm | When Used | How It Works |
|-----------|----------|--------------|
| **Nested Loop** | Small tables, indexed join column | For each row in outer table, look up matching rows in inner table using index |
| **Hash Join** | Large tables, equijoin, no useful index | Build hash table on smaller table, probe with larger table |
| **Merge Join** | Both inputs sorted (or can be cheaply sorted) | Walk through both sorted inputs simultaneously |

### GROUP BY and Aggregates

```sql
-- Basic aggregation
SELECT
    dept_id,
    COUNT(*) AS num_employees,
    AVG(salary) AS avg_salary,
    MIN(salary) AS min_salary,
    MAX(salary) AS max_salary,
    SUM(salary) AS total_salary
FROM employee
GROUP BY dept_id;

-- IMPORTANT RULE: every column in SELECT must be either:
--   1. In GROUP BY, or
--   2. Inside an aggregate function
-- Otherwise it's an error (except MySQL with ONLY_FULL_GROUP_BY off, which is dangerous)

-- HAVING — filter AFTER grouping
SELECT dept_id, AVG(salary) AS avg_salary
FROM employee
GROUP BY dept_id
HAVING AVG(salary) > 100000;  -- can't use alias in HAVING (in standard SQL)

-- WHERE vs HAVING:
-- WHERE filters individual ROWS before grouping
-- HAVING filters GROUPS after grouping

-- COUNT variations
SELECT
    COUNT(*),              -- count all rows (including NULLs)
    COUNT(dept_id),        -- count non-NULL dept_ids
    COUNT(DISTINCT dept_id) -- count unique non-NULL dept_ids
FROM employee;

-- STRING_AGG (PostgreSQL) / GROUP_CONCAT (MySQL)
SELECT dept_id, STRING_AGG(first_name, ', ' ORDER BY first_name) AS employees
FROM employee
GROUP BY dept_id;
-- Result: 1, "Alice, Bob"
```

### Subqueries

```sql
-- Scalar subquery (returns exactly one value)
SELECT first_name, salary,
       (SELECT AVG(salary) FROM employee) AS avg_salary
FROM employee;

-- Row subquery with IN
SELECT * FROM employee
WHERE dept_id IN (SELECT id FROM department WHERE budget > 500000);

-- Correlated subquery (runs once PER ROW of outer query — can be slow!)
SELECT e.first_name, e.salary
FROM employee e
WHERE e.salary > (
    SELECT AVG(salary) FROM employee WHERE dept_id = e.dept_id
);
-- "Find employees earning more than their department's average"

-- EXISTS — check if subquery returns any rows
SELECT d.name
FROM department d
WHERE EXISTS (SELECT 1 FROM employee e WHERE e.dept_id = d.id);
-- "Departments that have at least one employee"

-- NOT EXISTS
SELECT d.name
FROM department d
WHERE NOT EXISTS (SELECT 1 FROM employee e WHERE e.dept_id = d.id);
-- "Departments with no employees"

-- EXISTS vs IN:
-- EXISTS stops after finding ONE match (short-circuits)
-- IN materializes the entire subquery result
-- For large subquery results, EXISTS is usually faster
```

### ORDER BY and LIMIT

```sql
-- Sort ascending (default) and descending
SELECT * FROM employee ORDER BY salary DESC;
SELECT * FROM employee ORDER BY last_name ASC, first_name ASC;

-- NULLS FIRST / NULLS LAST (PostgreSQL)
SELECT * FROM employee ORDER BY dept_id NULLS LAST;

-- LIMIT and OFFSET
SELECT * FROM employee ORDER BY salary DESC LIMIT 10;           -- top 10
SELECT * FROM employee ORDER BY salary DESC LIMIT 10 OFFSET 20; -- page 3

-- WARNING: LIMIT/OFFSET pagination is SLOW for deep pages!
-- OFFSET 1000000 means the DB reads and discards 1 million rows.
-- Better approach: keyset pagination (cursor-based)
SELECT * FROM employee
WHERE id > 1000  -- last seen id
ORDER BY id
LIMIT 10;
```

### CASE WHEN

```sql
-- Simple CASE
SELECT first_name,
    CASE dept_id
        WHEN 1 THEN 'Engineering'
        WHEN 2 THEN 'Sales'
        ELSE 'Other'
    END AS dept_name
FROM employee;

-- Searched CASE (more flexible)
SELECT first_name, salary,
    CASE
        WHEN salary >= 150000 THEN 'Senior'
        WHEN salary >= 100000 THEN 'Mid'
        WHEN salary >= 60000  THEN 'Junior'
        ELSE 'Entry'
    END AS level
FROM employee;

-- CASE in WHERE (uncommon but valid)
SELECT * FROM employee
WHERE CASE WHEN is_active THEN salary > 80000 ELSE salary > 50000 END;

-- CASE in ORDER BY
SELECT * FROM employee
ORDER BY
    CASE WHEN is_active THEN 0 ELSE 1 END,  -- active first
    salary DESC;
```

---

## 5. Constraints — Data Integrity

```sql
-- All constraints in one table
CREATE TABLE product (
    id          SERIAL PRIMARY KEY,                    -- primary key (unique + not null)
    name        VARCHAR(200) NOT NULL,                 -- not null
    sku         VARCHAR(50) UNIQUE NOT NULL,            -- unique
    price       DECIMAL(10,2) CHECK (price >= 0),       -- check
    category_id INTEGER REFERENCES category(id),        -- foreign key
    quantity    INTEGER DEFAULT 0,                      -- default
    created_at  TIMESTAMPTZ DEFAULT NOW()

    -- Table-level constraints (needed for composite constraints)
    -- CONSTRAINT pk_product PRIMARY KEY (id),
    -- CONSTRAINT uq_sku UNIQUE (sku),
    -- CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES category(id),
    -- CONSTRAINT ck_price CHECK (price >= 0)
);

-- Composite unique constraint
CREATE TABLE enrollment (
    student_id INTEGER REFERENCES student(id),
    course_id  INTEGER REFERENCES course(id),
    grade      CHAR(2),
    PRIMARY KEY (student_id, course_id)  -- composite primary key
);

-- Foreign key actions
ALTER TABLE order_item
ADD CONSTRAINT fk_order
FOREIGN KEY (order_id) REFERENCES orders(id)
    ON DELETE CASCADE       -- delete order → delete its items
    ON UPDATE CASCADE;      -- update order id → update items' order_id

-- Options: CASCADE, SET NULL, SET DEFAULT, RESTRICT, NO ACTION
-- RESTRICT: prevent immediately
-- NO ACTION: prevent at end of statement (allows deferred checks)
-- CASCADE: propagate the change
-- SET NULL: set the FK column to NULL
-- SET DEFAULT: set the FK column to its DEFAULT value
```

**God-level constraint knowledge:**
- Constraints are enforced by the database, not your application. **This is the right place for data integrity.**
- `UNIQUE` columns allow multiple NULLs (NULL ≠ NULL). Use a partial unique index to prevent that.
- `CHECK` constraints can't reference other tables (use triggers for that).
- Foreign keys have overhead: every INSERT/UPDATE/DELETE checks the referenced table.
  - On a high-throughput system, FKs on hot tables can be a bottleneck.
  - Some teams drop FKs in production and enforce in application code (controversial but done at scale).

---

## 6. Practice Exercises

### Setup Data
```sql
CREATE TABLE department (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    budget DECIMAL(15,2)
);

CREATE TABLE employee (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    email VARCHAR(255) UNIQUE,
    salary DECIMAL(10,2),
    hire_date DATE,
    dept_id INTEGER REFERENCES department(id),
    manager_id INTEGER REFERENCES employee(id)
);

INSERT INTO department VALUES
(1, 'Engineering', 2000000),
(2, 'Sales', 800000),
(3, 'Marketing', 600000),
(4, 'HR', 400000),
(5, 'Research', 1500000);

INSERT INTO employee (id, first_name, last_name, email, salary, hire_date, dept_id, manager_id) VALUES
(1, 'Alice',   'Chen',    'alice@co.com',   150000, '2020-03-15', 1, NULL),
(2, 'Bob',     'Smith',   'bob@co.com',     130000, '2021-06-01', 1, 1),
(3, 'Carol',   'Jones',   'carol@co.com',    95000, '2022-01-10', 2, 1),
(4, 'Dave',    'Wilson',  'dave@co.com',     110000, '2021-09-20', 1, 1),
(5, 'Eve',     'Brown',   'eve@co.com',      85000, '2023-04-05', 2, 3),
(6, 'Frank',   'Taylor',  'frank@co.com',    72000, '2023-08-15', 3, 1),
(7, 'Grace',   'Davis',   'grace@co.com',   140000, '2020-11-01', 1, 1),
(8, 'Hannah',  'Miller',  'hannah@co.com',   68000, '2024-01-20', 4, 1),
(9, 'Ivan',    'Garcia',  'ivan@co.com',    125000, '2022-05-15', NULL, NULL),
(10, 'Julia',  'Martinez','julia@co.com',    92000, '2023-02-28', 3, 6);
```

### Exercises

**E1.** Find all employees in the Engineering department earning more than 120000.

<details><summary>Solution</summary>

```sql
SELECT e.first_name, e.last_name, e.salary
FROM employee e
JOIN department d ON e.dept_id = d.id
WHERE d.name = 'Engineering' AND e.salary > 120000;
```
</details>

**E2.** Find departments with no employees.

<details><summary>Solution</summary>

```sql
SELECT d.name
FROM department d
LEFT JOIN employee e ON d.id = e.dept_id
WHERE e.id IS NULL;

-- Or with NOT EXISTS:
SELECT d.name FROM department d
WHERE NOT EXISTS (SELECT 1 FROM employee e WHERE e.dept_id = d.id);
```
</details>

**E3.** Find the average salary per department, only showing departments with avg salary > 100000.

<details><summary>Solution</summary>

```sql
SELECT d.name, AVG(e.salary) AS avg_salary
FROM department d
JOIN employee e ON d.id = e.dept_id
GROUP BY d.name
HAVING AVG(e.salary) > 100000;
```
</details>

**E4.** Find employees who earn more than their manager.

<details><summary>Solution</summary>

```sql
SELECT e.first_name AS employee, e.salary AS emp_salary,
       m.first_name AS manager, m.salary AS mgr_salary
FROM employee e
JOIN employee m ON e.manager_id = m.id
WHERE e.salary > m.salary;
```
</details>

**E5.** Find the highest-paid employee in each department (including department name).

<details><summary>Solution</summary>

```sql
-- Using a subquery
SELECT e.first_name, e.salary, d.name AS dept_name
FROM employee e
JOIN department d ON e.dept_id = d.id
WHERE e.salary = (
    SELECT MAX(salary) FROM employee WHERE dept_id = e.dept_id
);

-- Better using window functions (next chapter!)
```
</details>

**E6.** For each employee, show their salary and what percentage of the department's total salary they represent.

<details><summary>Solution</summary>

```sql
SELECT e.first_name, e.salary, d.name,
       ROUND(e.salary / dept_totals.total * 100, 2) AS pct_of_dept
FROM employee e
JOIN department d ON e.dept_id = d.id
JOIN (
    SELECT dept_id, SUM(salary) AS total
    FROM employee
    WHERE dept_id IS NOT NULL
    GROUP BY dept_id
) dept_totals ON e.dept_id = dept_totals.dept_id
ORDER BY d.name, pct_of_dept DESC;
```
</details>

---

## Key Takeaways

1. **Learn the logical execution order** (FROM → WHERE → GROUP BY → HAVING → SELECT → ORDER BY → LIMIT). This explains 80% of SQL confusion.
2. **NULL is not a value** — it's the absence of a value. Three-valued logic applies.
3. **JOINs are set operations** — INNER = intersection, LEFT = all left + matching right, FULL = everything.
4. **Correlated subqueries** run once per row — watch performance.
5. **Multi-row INSERT** and `COPY` are massively faster than individual inserts.
6. **Always use parameterized queries** in application code — never concatenate strings into SQL (SQL injection).

---

Next: [03-sql-intermediate.md](03-sql-intermediate.md) →
