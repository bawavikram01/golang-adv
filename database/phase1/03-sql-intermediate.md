# 1.3 — SQL Intermediate: Window Functions, CTEs, and Set Operations

> This is where you leave 90% of developers behind.
> Window functions alone will change how you think about data.

---

## 1. Window Functions — The Power Tool

### What Are Window Functions?

A window function performs a calculation across a **set of rows related to the current row** — WITHOUT collapsing them into a single row like GROUP BY does.

```
GROUP BY:  many rows → one row per group (aggregate)
WINDOW:    many rows → same number of rows, each with extra computed column
```

### Anatomy of a Window Function

```sql
function_name(args) OVER (
    [PARTITION BY columns]    -- divide rows into groups (like GROUP BY but keeps rows)
    [ORDER BY columns]        -- order within each partition
    [frame_clause]            -- which rows in the partition to consider
)
```

### ROW_NUMBER, RANK, DENSE_RANK, NTILE

```sql
-- Sample data for all examples:
-- employee table with: id, first_name, dept_id, salary

-- ROW_NUMBER: sequential integer, no ties
SELECT first_name, dept_id, salary,
    ROW_NUMBER() OVER (ORDER BY salary DESC) AS row_num
FROM employee;

-- Result:
-- Alice    1   150000   1
-- Grace    1   140000   2
-- Bob      1   130000   3
-- Ivan    NULL 125000   4
-- Dave     1   110000   5
-- ...

-- ROW_NUMBER with PARTITION BY: restart numbering per group
SELECT first_name, dept_id, salary,
    ROW_NUMBER() OVER (PARTITION BY dept_id ORDER BY salary DESC) AS rank_in_dept
FROM employee;

-- Result:
-- Alice    1   150000   1   ← #1 in Engineering
-- Grace    1   140000   2
-- Bob      1   130000   3
-- Dave     1   110000   4
-- Carol    2    95000   1   ← #1 in Sales
-- Eve      2    85000   2
-- Julia    3    92000   1   ← #1 in Marketing
-- Frank    3    72000   2
-- Hannah   4    68000   1   ← #1 in HR

-- RANK: same values get same rank, with gaps
SELECT first_name, salary,
    RANK() OVER (ORDER BY salary DESC) AS rank
FROM employee;
-- If Alice and Bob both have 130000:
-- Alice  130000  1
-- Bob    130000  1
-- Carol  120000  3   ← gap! (not 2)

-- DENSE_RANK: same values get same rank, NO gaps
SELECT first_name, salary,
    DENSE_RANK() OVER (ORDER BY salary DESC) AS dense_rank
FROM employee;
-- Alice  130000  1
-- Bob    130000  1
-- Carol  120000  2   ← no gap

-- NTILE(n): divide rows into n roughly equal buckets
SELECT first_name, salary,
    NTILE(4) OVER (ORDER BY salary DESC) AS quartile
FROM employee;
-- Top 25% = quartile 1, next 25% = quartile 2, etc.
```

### The Killer Use Case: Top-N Per Group

```sql
-- Get the top 2 highest-paid employees per department
SELECT * FROM (
    SELECT first_name, dept_id, salary,
        ROW_NUMBER() OVER (PARTITION BY dept_id ORDER BY salary DESC) AS rn
    FROM employee
) ranked
WHERE rn <= 2;
```

This is the #1 most common window function use case. Learn it cold.

### LAG and LEAD — Access Adjacent Rows

```sql
-- LAG: access the PREVIOUS row's value
-- LEAD: access the NEXT row's value

SELECT
    first_name,
    salary,
    LAG(salary) OVER (ORDER BY salary DESC) AS higher_salary,
    LEAD(salary) OVER (ORDER BY salary DESC) AS lower_salary,
    salary - LAG(salary) OVER (ORDER BY salary DESC) AS diff_from_prev
FROM employee;

-- Result:
-- Alice  150000  NULL     140000  NULL
-- Grace  140000  150000   130000  -10000
-- Bob    130000  140000   125000  -10000
-- ...

-- LAG/LEAD with offset and default
LAG(salary, 2, 0) OVER (...)  -- look 2 rows back, default to 0 if none

-- Real-world: month-over-month revenue change
SELECT
    month,
    revenue,
    LAG(revenue) OVER (ORDER BY month) AS prev_month,
    ROUND((revenue - LAG(revenue) OVER (ORDER BY month))
          / LAG(revenue) OVER (ORDER BY month) * 100, 2) AS pct_change
FROM monthly_revenue;
```

### FIRST_VALUE, LAST_VALUE, NTH_VALUE

```sql
-- FIRST_VALUE: first value in the window frame
SELECT first_name, dept_id, salary,
    FIRST_VALUE(first_name) OVER (
        PARTITION BY dept_id ORDER BY salary DESC
    ) AS highest_paid_in_dept
FROM employee;

-- LAST_VALUE: CAREFUL — default frame doesn't include all rows!
SELECT first_name, dept_id, salary,
    LAST_VALUE(first_name) OVER (
        PARTITION BY dept_id ORDER BY salary DESC
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING  -- NEED THIS!
    ) AS lowest_paid_in_dept
FROM employee;

-- NTH_VALUE: nth value in the window
SELECT first_name, dept_id, salary,
    NTH_VALUE(salary, 2) OVER (
        PARTITION BY dept_id ORDER BY salary DESC
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    ) AS second_highest_salary
FROM employee;
```

### Running Totals, Moving Averages, Cumulative Operations

```sql
-- Running total
SELECT
    hire_date,
    salary,
    SUM(salary) OVER (ORDER BY hire_date) AS running_total
FROM employee;

-- Running count
SELECT
    hire_date,
    COUNT(*) OVER (ORDER BY hire_date) AS cumulative_hires
FROM employee;

-- Moving average (last 3 rows)
SELECT
    hire_date,
    salary,
    AVG(salary) OVER (
        ORDER BY hire_date
        ROWS BETWEEN 2 PRECEDING AND CURRENT ROW
    ) AS moving_avg_3
FROM employee;

-- Cumulative percentage
SELECT
    first_name,
    salary,
    SUM(salary) OVER (ORDER BY salary DESC) AS running_total,
    ROUND(
        SUM(salary) OVER (ORDER BY salary DESC)
        / SUM(salary) OVER () * 100, 2
    ) AS cumulative_pct
FROM employee;

-- Percent rank and cumulative distribution
SELECT first_name, salary,
    PERCENT_RANK() OVER (ORDER BY salary) AS pct_rank,    -- 0.0 to 1.0
    CUME_DIST() OVER (ORDER BY salary) AS cume_dist       -- 0+ to 1.0
FROM employee;
```

### Window Frame Clauses — The Deep Cut

The **frame** defines which rows relative to the current row are included in the calculation.

```
ROWS BETWEEN <start> AND <end>
RANGE BETWEEN <start> AND <end>
GROUPS BETWEEN <start> AND <end>  -- PostgreSQL 11+
```

Frame boundaries:
```
UNBOUNDED PRECEDING    -- first row of partition
n PRECEDING            -- n rows before current
CURRENT ROW            -- the current row
n FOLLOWING            -- n rows after current
UNBOUNDED FOLLOWING    -- last row of partition
```

```sql
-- Default frame (when ORDER BY is present):
RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
-- This is why LAST_VALUE gives unexpected results without explicit frame!

-- Default frame (when ORDER BY is absent):
ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
-- i.e., the entire partition

-- Explicit frames
SUM(salary) OVER (ORDER BY hire_date ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)
-- Running total (most common)

AVG(salary) OVER (ORDER BY hire_date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW)
-- 7-day moving average

SUM(salary) OVER (ORDER BY hire_date ROWS BETWEEN 1 PRECEDING AND 1 FOLLOWING)
-- Sum of previous + current + next row
```

**ROWS vs RANGE vs GROUPS:**
- `ROWS`: counts physical rows. `2 PRECEDING` = literally 2 rows before.
- `RANGE`: based on value distance. `RANGE BETWEEN 7 PRECEDING AND CURRENT ROW` with an ORDER BY date means "all rows within the last 7 days" (can be more than 7 rows!).
- `GROUPS`: counts peer groups (rows with same ORDER BY value). `2 PRECEDING` = 2 groups of tied values before.

### Named Windows (WINDOW clause)

```sql
-- Avoid repeating the same OVER clause
SELECT
    first_name,
    salary,
    SUM(salary) OVER w AS running_total,
    AVG(salary) OVER w AS running_avg,
    ROW_NUMBER() OVER w AS rn
FROM employee
WINDOW w AS (ORDER BY salary DESC);

-- Can extend named windows
SELECT
    first_name, dept_id, salary,
    RANK() OVER w_dept AS dept_rank,
    SUM(salary) OVER (w_dept ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS running
FROM employee
WINDOW w_dept AS (PARTITION BY dept_id ORDER BY salary DESC);
```

---

## 2. Common Table Expressions (CTEs)

### Non-Recursive CTEs

A CTE is a **named temporary result set** that exists for the duration of a single query. Think of it as an inline view.

```sql
-- Basic CTE
WITH dept_stats AS (
    SELECT
        dept_id,
        COUNT(*) AS emp_count,
        AVG(salary) AS avg_salary
    FROM employee
    GROUP BY dept_id
)
SELECT e.first_name, e.salary, ds.avg_salary,
       e.salary - ds.avg_salary AS diff_from_avg
FROM employee e
JOIN dept_stats ds ON e.dept_id = ds.dept_id;

-- Multiple CTEs
WITH
high_earners AS (
    SELECT * FROM employee WHERE salary > 120000
),
engineering AS (
    SELECT * FROM employee WHERE dept_id = 1
)
SELECT * FROM high_earners
INTERSECT
SELECT * FROM engineering;

-- CTE vs Subquery: When to use which?
-- CTEs: better readability, can be referenced multiple times, self-documenting
-- Subqueries: sometimes better optimized (PostgreSQL 12+ inlines simple CTEs)
```

### Recursive CTEs — Traversing Hierarchies

```sql
-- Recursive CTE structure:
WITH RECURSIVE cte_name AS (
    -- Base case (anchor member)
    SELECT ...

    UNION ALL  -- (or UNION to eliminate duplicates)

    -- Recursive case (references cte_name)
    SELECT ...
    FROM cte_name
    JOIN ...
)
SELECT * FROM cte_name;
```

```sql
-- Example 1: Employee hierarchy (org chart)
WITH RECURSIVE org_chart AS (
    -- Base case: top-level managers (no manager)
    SELECT id, first_name, manager_id, 1 AS level,
           first_name::TEXT AS path
    FROM employee
    WHERE manager_id IS NULL

    UNION ALL

    -- Recursive case: employees who report to someone in the result set
    SELECT e.id, e.first_name, e.manager_id, oc.level + 1,
           oc.path || ' → ' || e.first_name
    FROM employee e
    JOIN org_chart oc ON e.manager_id = oc.id
)
SELECT level, path, first_name
FROM org_chart
ORDER BY path;

-- Result:
-- 1  Alice                          Alice
-- 2  Alice → Bob                    Bob
-- 2  Alice → Carol                  Carol
-- 3  Alice → Carol → Eve            Eve
-- 2  Alice → Dave                   Dave
-- 2  Alice → Frank                  Frank
-- 3  Alice → Frank → Julia          Julia
-- 2  Alice → Grace                  Grace
-- 2  Alice → Hannah                 Hannah


-- Example 2: Generate a date series (PostgreSQL has generate_series, but this is portable)
WITH RECURSIVE dates AS (
    SELECT DATE '2026-01-01' AS d
    UNION ALL
    SELECT d + INTERVAL '1 day'
    FROM dates
    WHERE d < DATE '2026-01-31'
)
SELECT d FROM dates;


-- Example 3: Fibonacci sequence
WITH RECURSIVE fib AS (
    SELECT 1 AS n, 0::BIGINT AS a, 1::BIGINT AS b
    UNION ALL
    SELECT n + 1, b, a + b
    FROM fib
    WHERE n < 20
)
SELECT n, a AS fibonacci FROM fib;


-- Example 4: Bill of Materials (BOM) — find all components of a product
WITH RECURSIVE bom AS (
    SELECT component_id, subcomponent_id, quantity, 1 AS depth
    FROM assembly
    WHERE component_id = 'WIDGET-100'

    UNION ALL

    SELECT a.component_id, a.subcomponent_id, a.quantity * bom.quantity, depth + 1
    FROM assembly a
    JOIN bom ON a.component_id = bom.subcomponent_id
)
SELECT * FROM bom;


-- SAFETY: Always add a depth limit to prevent infinite loops!
WITH RECURSIVE traverse AS (
    SELECT id, parent_id, 1 AS depth FROM node WHERE id = 1
    UNION ALL
    SELECT n.id, n.parent_id, t.depth + 1
    FROM node n
    JOIN traverse t ON n.parent_id = t.id
    WHERE t.depth < 100  -- ← SAFETY LIMIT
)
SELECT * FROM traverse;
```

### Materialized vs Not Materialized CTEs

```sql
-- PostgreSQL 12+ can "inline" CTEs (treat them like subqueries for optimization)
-- To force materialization:
WITH my_cte AS MATERIALIZED (
    SELECT ... expensive query ...
)
SELECT * FROM my_cte WHERE ...;

-- To force inlining:
WITH my_cte AS NOT MATERIALIZED (
    SELECT ...
)
SELECT * FROM my_cte WHERE ...;

-- When materialization helps: CTE referenced multiple times, you want it computed once
-- When inlining helps: optimizer can push predicates into the CTE
```

---

## 3. Set Operations

```sql
-- UNION: combine results, remove duplicates
SELECT first_name FROM employee WHERE dept_id = 1
UNION
SELECT first_name FROM employee WHERE salary > 100000;

-- UNION ALL: combine results, KEEP duplicates (faster — no sort/dedup needed)
SELECT first_name FROM employee WHERE dept_id = 1
UNION ALL
SELECT first_name FROM employee WHERE salary > 100000;

-- INTERSECT: rows in both results
SELECT first_name FROM employee WHERE dept_id = 1
INTERSECT
SELECT first_name FROM employee WHERE salary > 100000;
-- Engineers who earn > 100k

-- EXCEPT: rows in first but not in second (set difference)
SELECT first_name FROM employee WHERE dept_id = 1
EXCEPT
SELECT first_name FROM employee WHERE salary > 130000;
-- Engineers who earn ≤ 130k

-- Rules:
-- Both sides must have the same number and type of columns
-- Column names come from the FIRST query
-- ORDER BY applies to the ENTIRE result (put at the end)

SELECT first_name, salary FROM employee WHERE dept_id = 1
UNION ALL
SELECT first_name, salary FROM employee WHERE dept_id = 2
ORDER BY salary DESC;  -- orders the combined result
```

---

## 4. GROUPING SETS, CUBE, ROLLUP

These are for generating **multiple levels of aggregation** in one query.

```sql
-- GROUPING SETS: explicit list of groupings
SELECT dept_id, EXTRACT(YEAR FROM hire_date) AS hire_year, COUNT(*), AVG(salary)
FROM employee
GROUP BY GROUPING SETS (
    (dept_id, hire_year),   -- group by both
    (dept_id),              -- group by department only
    (hire_year),            -- group by year only
    ()                      -- grand total
);

-- ROLLUP: hierarchical aggregation (drills up from detail to total)
SELECT dept_id, EXTRACT(YEAR FROM hire_date) AS hire_year, COUNT(*), AVG(salary)
FROM employee
GROUP BY ROLLUP (dept_id, hire_year);
-- Produces: (dept_id, hire_year), (dept_id), ()
-- Does NOT produce (hire_year) alone — it's hierarchical

-- CUBE: all possible combinations
SELECT dept_id, EXTRACT(YEAR FROM hire_date) AS hire_year, COUNT(*), AVG(salary)
FROM employee
GROUP BY CUBE (dept_id, hire_year);
-- Produces: (dept_id, hire_year), (dept_id), (hire_year), ()

-- GROUPING() function: distinguish NULL from "super-aggregate row"
SELECT
    CASE WHEN GROUPING(dept_id) = 1 THEN 'ALL DEPTS' ELSE dept_id::TEXT END AS dept,
    CASE WHEN GROUPING(hire_year) = 1 THEN 'ALL YEARS' ELSE hire_year::TEXT END AS year,
    COUNT(*), AVG(salary)
FROM (
    SELECT dept_id, EXTRACT(YEAR FROM hire_date) AS hire_year, salary FROM employee
) t
GROUP BY ROLLUP (dept_id, hire_year);
```

---

## 5. LATERAL Joins

A LATERAL join lets the right side of the join **reference columns from the left side**. It's like a correlated subquery, but in the FROM clause.

```sql
-- Without LATERAL (this doesn't work):
-- SELECT * FROM department d
-- JOIN (SELECT * FROM employee WHERE dept_id = d.id LIMIT 2) e;  -- ERROR: can't reference d

-- With LATERAL:
SELECT d.name, top_emp.*
FROM department d
CROSS JOIN LATERAL (
    SELECT first_name, salary
    FROM employee
    WHERE dept_id = d.id
    ORDER BY salary DESC
    LIMIT 2
) top_emp;
-- Gets top 2 employees per department — very efficient!

-- LATERAL with LEFT JOIN (include departments with no employees)
SELECT d.name, top_emp.*
FROM department d
LEFT JOIN LATERAL (
    SELECT first_name, salary
    FROM employee
    WHERE dept_id = d.id
    ORDER BY salary DESC
    LIMIT 2
) top_emp ON TRUE;

-- LATERAL is also great for "apply a function to each row"
SELECT d.name, gs.n
FROM department d
CROSS JOIN LATERAL generate_series(1, 3) AS gs(n);
-- Generates 3 rows per department
```

---

## 6. Pivoting and Unpivoting

### Pivot (rows → columns)

```sql
-- Manual pivot using CASE + aggregate
-- Turn: (dept_id, year, count) rows into columns per year
SELECT
    dept_id,
    COUNT(*) FILTER (WHERE EXTRACT(YEAR FROM hire_date) = 2020) AS hired_2020,
    COUNT(*) FILTER (WHERE EXTRACT(YEAR FROM hire_date) = 2021) AS hired_2021,
    COUNT(*) FILTER (WHERE EXTRACT(YEAR FROM hire_date) = 2022) AS hired_2022,
    COUNT(*) FILTER (WHERE EXTRACT(YEAR FROM hire_date) = 2023) AS hired_2023,
    COUNT(*) FILTER (WHERE EXTRACT(YEAR FROM hire_date) = 2024) AS hired_2024
FROM employee
GROUP BY dept_id
ORDER BY dept_id;

-- PostgreSQL tablefunc extension
CREATE EXTENSION IF NOT EXISTS tablefunc;

SELECT * FROM crosstab(
    'SELECT dept_id, EXTRACT(YEAR FROM hire_date)::INT, COUNT(*)::INT
     FROM employee GROUP BY 1, 2 ORDER BY 1, 2',
    'SELECT generate_series(2020, 2024)'
) AS ct(dept_id INT, y2020 INT, y2021 INT, y2022 INT, y2023 INT, y2024 INT);
```

### Unpivot (columns → rows)

```sql
-- Given a wide table:
-- quarterly_sales(product, q1_sales, q2_sales, q3_sales, q4_sales)

-- Unpivot using UNION ALL
SELECT product, 'Q1' AS quarter, q1_sales AS sales FROM quarterly_sales
UNION ALL
SELECT product, 'Q2', q2_sales FROM quarterly_sales
UNION ALL
SELECT product, 'Q3', q3_sales FROM quarterly_sales
UNION ALL
SELECT product, 'Q4', q4_sales FROM quarterly_sales;

-- Unpivot using LATERAL (more elegant)
SELECT qs.product, x.quarter, x.sales
FROM quarterly_sales qs
CROSS JOIN LATERAL (
    VALUES ('Q1', q1_sales), ('Q2', q2_sales), ('Q3', q3_sales), ('Q4', q4_sales)
) AS x(quarter, sales);
```

---

## 7. JSON Querying (PostgreSQL)

```sql
CREATE TABLE api_logs (
    id SERIAL PRIMARY KEY,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO api_logs (payload) VALUES
('{"method": "POST", "path": "/api/users", "status": 201, "duration_ms": 45, "user": {"id": 1, "role": "admin"}}'),
('{"method": "GET",  "path": "/api/users", "status": 200, "duration_ms": 12, "user": {"id": 2, "role": "viewer"}}'),
('{"method": "POST", "path": "/api/orders", "status": 500, "duration_ms": 2500, "error": "timeout"}');

-- Access operators:
-- ->   returns JSON element
-- ->>  returns element as TEXT
-- #>   path access (returns JSON)
-- #>>  path access (returns TEXT)

SELECT
    payload->>'method' AS method,                    -- 'POST' (text)
    payload->>'status' AS status,                    -- '201' (text!)
    (payload->>'status')::INT AS status_int,         -- 201 (integer)
    payload->'user'->>'role' AS user_role,            -- 'admin'
    payload#>>'{user,role}' AS user_role_alt          -- 'admin' (path syntax)
FROM api_logs;

-- Filtering
SELECT * FROM api_logs WHERE payload->>'method' = 'POST';
SELECT * FROM api_logs WHERE (payload->>'status')::INT >= 500;
SELECT * FROM api_logs WHERE payload @> '{"method": "POST"}';  -- containment
SELECT * FROM api_logs WHERE payload ? 'error';                 -- key exists
SELECT * FROM api_logs WHERE payload ?| ARRAY['error','warning']; -- any key exists

-- Aggregate JSON
SELECT
    payload->>'method' AS method,
    COUNT(*),
    AVG((payload->>'duration_ms')::NUMERIC) AS avg_duration
FROM api_logs
GROUP BY payload->>'method';

-- Modify JSON
UPDATE api_logs
SET payload = payload || '{"processed": true}'::JSONB  -- add/overwrite key
WHERE id = 1;

UPDATE api_logs
SET payload = payload - 'error'  -- remove key
WHERE id = 3;

UPDATE api_logs
SET payload = jsonb_set(payload, '{user,role}', '"superadmin"')  -- set nested key
WHERE id = 1;

-- Indexing JSONB
CREATE INDEX idx_logs_payload ON api_logs USING GIN (payload);           -- index everything
CREATE INDEX idx_logs_method ON api_logs USING BTREE ((payload->>'method')); -- index specific key
CREATE INDEX idx_logs_status ON api_logs USING BTREE (((payload->>'status')::INT)); -- index as int
```

---

## 8. Practice Exercises

**E1.** Using window functions, show each employee's salary, the department average salary, and the difference. Don't use GROUP BY.

<details><summary>Solution</summary>

```sql
SELECT
    first_name, dept_id, salary,
    ROUND(AVG(salary) OVER (PARTITION BY dept_id), 2) AS dept_avg,
    ROUND(salary - AVG(salary) OVER (PARTITION BY dept_id), 2) AS diff_from_avg
FROM employee
ORDER BY dept_id, salary DESC;
```
</details>

**E2.** For each employee, show the salary of the person who was hired immediately before them and after them (company-wide).

<details><summary>Solution</summary>

```sql
SELECT
    first_name, hire_date, salary,
    LAG(salary) OVER (ORDER BY hire_date) AS prev_hire_salary,
    LEAD(salary) OVER (ORDER BY hire_date) AS next_hire_salary
FROM employee;
```
</details>

**E3.** Using a recursive CTE, build the full org chart showing each employee's level and their chain of command.

<details><summary>Solution</summary>

```sql
WITH RECURSIVE org AS (
    SELECT id, first_name, manager_id, 1 AS level,
           ARRAY[first_name] AS chain
    FROM employee WHERE manager_id IS NULL

    UNION ALL

    SELECT e.id, e.first_name, e.manager_id, o.level + 1,
           o.chain || e.first_name
    FROM employee e
    JOIN org o ON e.manager_id = o.id
    WHERE o.level < 10
)
SELECT level,
       REPEAT('  ', level - 1) || first_name AS org_chart,
       array_to_string(chain, ' → ') AS chain_of_command
FROM org
ORDER BY chain;
```
</details>

**E4.** Calculate a 3-month moving average of hire counts by month.

<details><summary>Solution</summary>

```sql
WITH monthly_hires AS (
    SELECT
        DATE_TRUNC('month', hire_date) AS month,
        COUNT(*) AS hire_count
    FROM employee
    GROUP BY DATE_TRUNC('month', hire_date)
)
SELECT
    month,
    hire_count,
    ROUND(AVG(hire_count) OVER (
        ORDER BY month
        ROWS BETWEEN 2 PRECEDING AND CURRENT ROW
    ), 2) AS moving_avg_3m
FROM monthly_hires
ORDER BY month;
```
</details>

**E5.** Using LATERAL, get each department's top earner and their salary as a percentage of the department budget.

<details><summary>Solution</summary>

```sql
SELECT
    d.name AS dept,
    d.budget,
    top.first_name,
    top.salary,
    ROUND(top.salary / d.budget * 100, 4) AS pct_of_budget
FROM department d
LEFT JOIN LATERAL (
    SELECT first_name, salary
    FROM employee
    WHERE dept_id = d.id
    ORDER BY salary DESC
    LIMIT 1
) top ON TRUE;
```
</details>

---

## Key Takeaways

1. **Window functions** don't reduce rows — they ADD computed columns. Master PARTITION BY + ORDER BY + frame clauses.
2. **ROW_NUMBER + PARTITION BY** for top-N per group is the most valuable SQL pattern.
3. **LAG/LEAD** replace self-joins for accessing previous/next rows.
4. **Frame clause defaults** are tricky: `RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW` when ORDER BY is present.
5. **Recursive CTEs** unlock tree/graph traversal in SQL. Always add depth limits.
6. **LATERAL** is a correlated subquery in FROM — extremely powerful for "for each row, compute something".
7. **GROUPING SETS/CUBE/ROLLUP** replace multiple queries with UNION ALL.

---

Next: [04-sql-advanced.md](04-sql-advanced.md) →
