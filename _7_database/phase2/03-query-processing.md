# 2.3 — Query Processing & Optimization

> When you write `SELECT ... FROM ... WHERE ...`, the database doesn't just "run" it.  
> It goes through a **pipeline**: parse → analyze → rewrite → plan → execute.  
> The optimizer chooses from millions of possible strategies. Understanding it is power.

---

## 1. The Query Processing Pipeline

```
    SQL Text
       │
       ▼
  ┌─────────┐    Tokenize, check syntax
  │  PARSER  │    Output: parse tree (AST)
  └────┬─────┘
       ▼
  ┌──────────┐   Resolve names, check types, verify permissions
  │ ANALYZER │   Output: query tree (with resolved references)
  └────┬─────┘
       ▼
  ┌──────────┐   Apply views, rules, security policies
  │ REWRITER │   Output: rewritten query tree
  └────┬─────┘
       ▼
  ┌──────────┐   Generate candidate plans, estimate costs, pick cheapest
  │ PLANNER/ │   Output: physical execution plan (plan tree)
  │OPTIMIZER │
  └────┬─────┘
       ▼
  ┌──────────┐   Execute the plan, return rows
  │ EXECUTOR │   Output: result set
  └──────────┘
```

### Step 1: Parsing

```sql
SELECT e.name, d.name FROM employee e JOIN department d ON e.dept_id = d.id WHERE e.salary > 100000;
```

Parser produces an Abstract Syntax Tree (AST):
```
SelectStmt
├── targetList: [e.name, d.name]
├── fromClause:
│   └── JoinExpr
│       ├── larg: RangeVar(employee AS e)
│       ├── rarg: RangeVar(department AS d)
│       └── quals: OpExpr(e.dept_id = d.id)
└── whereClause: OpExpr(e.salary > 100000)
```

No semantics yet — doesn't know if `employee` table exists or if `salary` is a real column.

### Step 2: Analysis (Semantic Analysis)

- Resolve table names → system catalog lookups
- Resolve column names → check pg_attribute
- Type checking → salary is DECIMAL, 100000 is INTEGER, add implicit cast
- Permission checks → does current user have SELECT on these tables?
- Output: annotated query tree with OIDs (internal object IDs)

### Step 3: Rewriting

- **View expansion:** If you query a view, replace it with the underlying query.
- **Rule application:** PostgreSQL's rule system (e.g., INSTEAD OF rules on views).
- **Row-Level Security:** Inject RLS predicates into WHERE clause.

### Step 4: Planning / Optimization (the hard part)

### Step 5: Execution

---

## 2. Logical Optimization (Before Cost Estimation)

These are **algebraic transformations** that are ALWAYS beneficial.

### Predicate Pushdown

```
BEFORE:
  π_name(σ_salary>100000(employee ⋈ department))
  "Join everything, THEN filter"
  
AFTER:
  π_name(σ_salary>100000(employee) ⋈ department)
  "Filter first, THEN join" — fewer rows to join!
```

```sql
-- The optimizer rewrites:
SELECT e.name, d.name FROM employee e JOIN department d ON e.dept_id = d.id WHERE e.salary > 100000;

-- Internally becomes (conceptually):
SELECT e.name, d.name FROM (SELECT * FROM employee WHERE salary > 100000) e JOIN department d ON e.dept_id = d.id;

-- Pushes the filter before the join → much fewer rows to process
```

### Projection Pushdown

```
BEFORE: Read all 20 columns, join, then project 2 columns
AFTER:  Read only needed columns from each table early

This matters especially for column stores — skip entire column files.
For row stores, less impactful (must read full row from heap anyway).
```

### Constant Folding

```sql
-- Before: WHERE created_at > NOW() - INTERVAL '30 days'
-- After:  WHERE created_at > '2026-03-18 14:30:00+00'  (computed once at plan time)

-- Before: WHERE x = 5 + 3
-- After:  WHERE x = 8
```

### Predicate Simplification

```sql
-- Boolean simplification:
WHERE TRUE AND salary > 100000    →  WHERE salary > 100000
WHERE FALSE OR dept_id = 1        →  WHERE dept_id = 1
WHERE NOT (salary <= 100000)      →  WHERE salary > 100000

-- Contradiction detection:
WHERE dept_id = 1 AND dept_id = 2  →  returns 0 rows immediately (impossible)

-- Redundancy elimination:
WHERE x > 5 AND x > 3  →  WHERE x > 5
WHERE x = 5 AND x > 3  →  WHERE x = 5
```

### Subquery Flattening / Decorrelation

```sql
-- Correlated subquery (slow — runs once per row):
SELECT * FROM employee e
WHERE salary > (SELECT AVG(salary) FROM employee WHERE dept_id = e.dept_id);

-- Optimizer can rewrite as a join (faster — one pass):
SELECT e.* FROM employee e
JOIN (SELECT dept_id, AVG(salary) AS avg_sal FROM employee GROUP BY dept_id) d
ON e.dept_id = d.dept_id
WHERE e.salary > d.avg_sal;
```

---

## 3. Physical Optimization (Cost-Based)

The optimizer generates multiple **physical plans** and picks the cheapest based on **cost estimation**.

### Cost Model

PostgreSQL's cost is measured in **arbitrary units** calibrated to sequential page reads:

```
seq_page_cost    = 1.0    (reading one page sequentially)
random_page_cost = 4.0    (reading one page randomly — 4x slower than sequential)
cpu_tuple_cost   = 0.01   (processing one row)
cpu_index_tuple_cost = 0.005  (processing one index entry)
cpu_operator_cost = 0.0025    (evaluating one operator/function)

Total cost = (pages read) × page_cost + (tuples processed) × tuple_cost

For SSD: set random_page_cost = 1.1 (random and sequential are almost same speed)
         Default 4.0 is for HDD.
```

### Scan Method Selection

```
For: SELECT * FROM employee WHERE dept_id = 1

Option A: Sequential Scan
  cost = pages_in_table × seq_page_cost + rows_in_table × cpu_tuple_cost
  cost = 1000 × 1.0 + 100000 × 0.01 = 2000
  (Must read all pages, check every row)

Option B: Index Scan on idx_dept
  cost = tree_height × random_page_cost + matching_rows × random_page_cost + matching_rows × cpu_cost
  cost = 3 × 4.0 + 500 × 4.0 + 500 × 0.01 = 2017
  (Traverse index, random-read 500 heap pages)

Option C: Bitmap Index Scan + Bitmap Heap Scan
  cost = index_pages × random_page_cost + matching_heap_pages × seq_page_cost
  cost = 3 × 4.0 + 50 × 1.0 = 62
  (Build bitmap, sequentially read 50 heap pages)

Optimizer picks Option C (cost 62) → fastest!
```

### The Selectivity Estimate

**Selectivity** = fraction of rows that match a predicate. This is EVERYTHING.

```
WHERE dept_id = 1

Optimizer checks pg_stats:
  - n_distinct for dept_id = 5 departments
  - most_common_vals = {1, 2, 3, 4, 5}
  - most_common_freqs = {0.35, 0.25, 0.20, 0.15, 0.05}
  
  dept_id = 1 → selectivity = 0.35 (35% of rows)
  
  For 100,000 total rows → estimated 35,000 matching rows

If no stats available, optimizer uses DEFAULT estimates:
  equality:  1/n_distinct or 0.5% (1/200)
  range:     33% (a guess!)
  LIKE 'abc': 0.5%
  
  → This is why ANALYZE is critical. Bad estimates → bad plans.
```

### Histogram for Range Queries

```
WHERE salary > 100000

pg_stats histogram_bounds for salary:
  [30000, 45000, 55000, 65000, 75000, 85000, 95000, 105000, 120000, 150000, 200000]
  
  10 buckets, each with ~10% of rows.
  salary > 100000 falls in bucket 7-8.
  
  Estimated selectivity: ~30% of rows have salary > 100000
  
  If the histogram had 100 buckets (more precise):
  SET default_statistics_target = 1000;  -- more buckets
  ANALYZE employee;
```

---

## 4. Join Optimization

### Join Algorithms

#### Nested Loop Join

```
For each row in OUTER table:
    For each row in INNER table:
        If join condition matches: emit combined row

                 OUTER (employee)       INNER (department)
                 ┌──────┐              ┌──────┐
  for each row → │ e₁   │ ──────────→ │ d₁   │ check
                 │      │              │ d₂   │ check
                 │      │              │ d₃   │ check
                 ├──────┤              ├──────┤
                 │ e₂   │ ──────────→ │ d₁   │ check
                 │      │              │ d₂   │ check
                 │      │              │ d₃   │ check
                 └──────┘              └──────┘

Complexity: O(N × M) without index on inner
            O(N × log M) with index on inner

Best when:
  - Outer is small (< 1000 rows)
  - Inner has an index on the join column
  - Used for: non-equijoin (theta join), LATERAL
  
PostgreSQL always considers nested loop.
It's the ONLY algorithm that works for non-equijoins (e.g., WHERE a.x > b.y).
```

#### Hash Join

```
PHASE 1 — BUILD: Create hash table on the smaller input
  For each row in BUILD input:
    hash(join_key) → insert into hash table

PHASE 2 — PROBE: Scan the larger input, probe hash table
  For each row in PROBE input:
    hash(join_key) → look up in hash table
    If found: emit combined row

  BUILD (department, smaller)     PROBE (employee, larger)
  ┌────────────────────┐         ┌──────────────┐
  │ Hash Table:        │         │              │
  │  hash(1) → d₁     │  ←───── │ e₁(dept=1)  │ → match!
  │  hash(2) → d₂     │  ←───── │ e₂(dept=3)  │ → match!
  │  hash(3) → d₃     │  ←───── │ e₃(dept=1)  │ → match!
  └────────────────────┘         └──────────────┘

Complexity: O(N + M) time, O(min(N,M)) memory
Must fit smaller input in memory (work_mem).
If doesn't fit: multi-pass hash join (spill to disk — much slower).

Best when:
  - Equijoin (= only, not > or <)
  - No useful index
  - Both tables are large
  - Sufficient work_mem
```

#### Merge Join (Sort-Merge Join)

```
PHASE 1: Sort both inputs by join key (unless already sorted via index)
PHASE 2: Walk through both sorted lists simultaneously

  Sorted employee      Sorted department
  ┌────────────┐      ┌────────────┐
  │ dept=1, e₁ │ ──── │ dept=1, d₁ │  match!
  │ dept=1, e₃ │ ──── │            │  match!
  │ dept=2, e₅ │ ──── │ dept=2, d₂ │  match!
  │ dept=3, e₂ │ ──── │ dept=3, d₃ │  match!
  └────────────┘      └────────────┘

Complexity: O(N log N + M log M) for sorting + O(N + M) for merging
If already sorted (via index): O(N + M) total — fastest!

Best when:
  - Both inputs already sorted (indexes on join columns)
  - Result must be sorted (ORDER BY on join column)
  - Very large data sets (streaming, no memory pressure)
```

### Join Ordering

For a query with N tables, there are **N! possible join orders** (and more with different join algorithms per pair).

```
3 tables: 3! = 6 orderings
5 tables: 5! = 120 orderings
10 tables: 10! = 3,628,800 orderings
15 tables: too many — need heuristics

PostgreSQL:
  - For ≤ 12 tables: exhaustive dynamic programming (considers all orderings)
  - For > 12 tables: GEQO (Genetic Query Optimization) — heuristic search
  - Configurable: geqo_threshold parameter (default 12)

The order MATTERS enormously:
  Bad order: join 1M × 1M = 500K rows, then join 500K × 10 = 500 rows
  Good order: join 1M × 10 = 8 rows, then join 8 × 1M = 8 rows
  → Same result, wildly different performance
```

### Join Selectivity and Cardinality Estimation

```
The hardest part of query optimization.

Given: employee (100K rows) JOIN department (50 rows) ON employee.dept_id = department.id

Estimated output rows = 100K × 50 × selectivity
  If dept_id is a FK to department.id: selectivity ≈ 1/50
  Output ≈ 100K × 50 × (1/50) = 100K rows (every employee matches ONE dept)

BUT if the join condition is ON employee.age = department.floor_number:
  Selectivity is much harder to estimate (no FK relationship)
  Postgres assumes 1/max(n_distinct(age), n_distinct(floor)) as a heuristic

Correlated predicates are the KILLJOY:
  WHERE city = 'San Francisco' AND state = 'California'
  Optimizer assumes independence: P(city=SF) × P(state=CA)
  Reality: P(city=SF AND state=CA) = P(city=SF) — they're correlated!
  → Estimate is way too low → picks bad plan

  PostgreSQL 10+: CREATE STATISTICS for multi-column stats
  CREATE STATISTICS stats_city_state ON city, state FROM addresses;
  ANALYZE addresses;
```

---

## 5. Execution Models

### Volcano Model (Iterator / Pull-Based)

```
Every operator has three methods: Open(), Next(), Close()

SELECT e.name FROM employee e JOIN department d ON e.dept_id = d.id WHERE d.name = 'Eng'

Execution tree (rows flow UPWARD):
  
  Project(e.name)          ← keeps calling Next() on child
       │
  Hash Join                ← calls Next() on both children to get rows
    /       \
  Seq Scan   Filter(d.name='Eng')
  (employee)    │
             Seq Scan
             (department)

Flow:
  1. Top node calls Next() on Hash Join
  2. Hash Join calls Next() on right child to BUILD hash table
     - Filter calls Next() on Seq Scan(department) repeatedly
     - Each row: check d.name='Eng', if yes → add to hash table
  3. Hash Join calls Next() on left child (Seq Scan employee)
     - Gets one row at a time, probes hash table
     - If match found → returns it to Project
  4. Project extracts e.name, returns to client
  5. Repeat until Next() returns NULL (no more rows)

Pro: simple, composable, each operator is independent
Con: high overhead per row (virtual function call per Next())
Con: can't use CPU registers/caches effectively (one row at a time)
```

### Vectorized Execution (Column-at-a-Time)

```
Instead of processing ONE row per Next() call, process a VECTOR of rows (typically 1024).

  Next() returns a "column batch":
  ┌──────────┬───────────┬──────────┐
  │ name[0]  │ dept[0]   │ sal[0]   │
  │ name[1]  │ dept[1]   │ sal[1]   │
  │ name[2]  │ dept[2]   │ sal[2]   │
  │ ...      │ ...       │ ...      │
  │ name[1023]│ dept[1023]│ sal[1023]│
  └──────────┴───────────┴──────────┘

Benefits:
  - Amortize Next() overhead across 1024 rows
  - CPU SIMD instructions can process 4-8 values simultaneously
  - Better cache utilization (data stays in L1/L2 cache)
  - 5-10x faster than row-at-a-time for analytical queries

Used by: DuckDB, ClickHouse, Velox (Meta), DataFusion (Apache Arrow)
PostgreSQL: not vectorized (row-at-a-time), but has JIT compilation
```

### JIT Compilation

```
PostgreSQL 11+ can JIT-compile parts of query execution using LLVM:
  - Expression evaluation (WHERE salary * 1.1 > 100000)
  - Tuple deforming (extracting column values from heap tuples)

Instead of interpreting the expression tree for every row,
generate native machine code that runs directly.

Benefit: 20-30% speedup for CPU-intensive queries (complex expressions, many aggregates)
When enabled: only when estimated cost is high (jit_above_cost parameter)

SET jit = on;
SET jit_above_cost = 100000;  -- only JIT-compile expensive queries
```

---

## 6. Parallel Query Execution

```sql
-- PostgreSQL 9.6+ can parallelize queries

EXPLAIN ANALYZE SELECT COUNT(*) FROM orders WHERE total > 100;

-- Gather (workers=4)
--   -> Partial Aggregate  (actual time=0.5..120)
--        -> Parallel Seq Scan on orders  (actual rows=250000 per worker)
--             Filter: (total > 100)
-- Planning Time: 0.2 ms
-- Execution Time: 135 ms (vs ~500 ms without parallelism)

Parallel-safe operations:
  ✓ Seq Scan, Index Scan, Bitmap Scan
  ✓ Hash Join, Nested Loop, Merge Join
  ✓ Aggregation (partial + gather)
  ✓ Sort (each worker sorts a portion)

NOT parallel-safe:
  ✗ Writing data (INSERT, UPDATE, DELETE)
  ✗ Cursors
  ✗ Functions marked PARALLEL UNSAFE

Configuration:
  max_parallel_workers_per_gather = 4  (workers per query node)
  max_parallel_workers = 8            (total across all queries)
  min_parallel_table_scan_size = 8MB  (skip parallelism for small tables)
  parallel_tuple_cost = 0.1           (cost of sending a tuple between processes)
```

---

## 7. Plan Stability & Regression

### The Problem

```
Monday: query runs in 50ms (optimizer chooses index scan)
Tuesday: new data loaded, statistics updated
Tuesday: same query runs in 5 seconds (optimizer chooses seq scan)

Why? The selectivity estimate changed with new data.
The optimizer is doing its best with the NEW statistics.
But the new plan is worse in practice (cardinality estimation error).
```

### Solutions

```sql
-- 1. Plan hints (MySQL, Oracle — NOT PostgreSQL)
SELECT /*+ INDEX(e idx_dept_salary) */ * FROM employee e WHERE dept_id = 1;

-- 2. PostgreSQL: pg_hint_plan extension
LOAD 'pg_hint_plan';
SELECT /*+ SeqScan(e) */ * FROM employee e WHERE dept_id = 1;

-- 3. Optimizer parameter tweaking (discouraged but sometimes necessary)
SET enable_seqscan = off;    -- force index scans (NEVER in production permanently!)
SET enable_hashjoin = off;   -- force merge/nested loop joins

-- 4. Query Store (SQL Server) / pg_stat_statements (PostgreSQL)
-- Continuously track query plans, alert on regressions

-- 5. Prepared statements with plan caching
-- PostgreSQL creates a generic plan after 5 executions of a prepared statement
-- If generic plan is bad, force custom plans:
SET plan_cache_mode = 'force_custom_plan';
```

---

## 8. Practical Query Optimization Workflow

```
Step 1: Identify the slow query
  → pg_stat_statements, slow query log, application APM

Step 2: Run EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
  → Read the plan, find the expensive nodes

Step 3: Check for common problems:
  □ Seq Scan on large table with selective WHERE → missing index
  □ Estimated rows ≠ actual rows → stale stats (run ANALYZE)
  □ Nested Loop with large inner → try enabling/using hash join
  □ Sort with "Sort Method: external merge Disk" → increase work_mem
  □ Rows Removed by Filter: huge number → predicate not pushed down
  □ Hash Join batched to disk → increase work_mem

Step 4: Fix
  → Add index (check ESR rule for column order)
  → ANALYZE to update stats
  → Increase work_mem for sorts/hashes
  → Rewrite query (avoid correlated subquery, use CTE/JOIN instead)
  → Add covering index (INCLUDE) for index-only scan

Step 5: Verify
  → Re-run EXPLAIN ANALYZE
  → Confirm cost decreased AND actual time decreased
```

---

## Key Takeaways

1. **Parsing → Analysis → Rewrite → Plan → Execute.** The planner is where the magic (and bugs) live.
2. **Cost estimation = selectivity × cardinality × I/O model.** Bad stats = bad plans.
3. **Three join algorithms:** Nested Loop (small + indexed), Hash Join (equijoin, no index), Merge Join (pre-sorted).
4. **Join ordering** is exponential — the optimizer uses dynamic programming for ≤12 tables.
5. **Predicate pushdown** is the most impactful logical optimization — filter early.
6. **ANALYZE regularly.** Automatic autovacuum does this, but verify your stats are fresh.
7. **Vectorized execution** (ClickHouse, DuckDB) is 5-10x faster than row-at-a-time (PostgreSQL) for analytics.
8. **Plan regression** is real — monitor query plans over time.

---

Next: [04-concurrency-control.md](04-concurrency-control.md) →
