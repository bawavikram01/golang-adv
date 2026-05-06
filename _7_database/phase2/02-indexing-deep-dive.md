# 2.2 — Indexing Deep Dive: Every Index Type Explained

> An index is a **separate data structure** that speeds up lookups at the cost of write overhead and storage.
> Choosing the RIGHT index type and column order is the #1 performance skill.

---

## 1. The Cost of Not Indexing

```sql
-- Table: orders (50 million rows, 10 GB on disk)
-- No index on customer_id

SELECT * FROM orders WHERE customer_id = 42;

-- Without index: SEQUENTIAL SCAN
--   Read all 50M rows (1.25 million 8 KB pages)
--   At 500 MB/s SSD sequential: ~20 seconds
--   Filter 49,999,980 rows to find ~20 matches

-- With B-tree index on customer_id:
--   3 index page reads + 20 heap page reads
--   At 100 μs per random read: ~2.3 ms
--   ~8,700x faster
```

---

## 2. Primary vs Secondary Indexes

### Primary Index (Clustered)

The table data is physically sorted by the index key. There can be only ONE.

```
MySQL/InnoDB: primary key IS the clustered index.
  The table itself is a B+ tree sorted by PK.
  Leaf pages contain the ACTUAL ROW DATA.

  ┌── root ──┐
  │  [30|60]  │
  └─┬────┬───┘
    ▼    ▼
  ┌──────┐  ┌──────┐
  │10|20 │  │40|50 │   internal: just keys + pointers
  └─┬──┬─┘  └─┬──┬─┘
    ▼  ▼      ▼  ▼
  ┌──────────────────┐
  │ PK=10: full row  │   leaf: ACTUAL row data
  │ PK=20: full row  │
  └──────────────────┘

Consequence: secondary indexes in InnoDB store the PRIMARY KEY (not a heap pointer).
  Looking up by secondary index = traverse secondary B-tree → get PK → traverse primary B-tree.
  This is a "double lookup" — one reason why small PKs (INT) are better than UUIDs in InnoDB.

PostgreSQL: does NOT have clustered indexes by default.
  The heap is unordered. CLUSTER command physically reorders once (but not maintained).
  Primary key index points to heap (page_id, offset) just like any other index.
```

### Secondary Index (Non-Clustered)

An additional B-tree on non-PK columns. Points back to the row location.

```
PostgreSQL secondary index:
  Leaf entries: (key_value, TID) where TID = (heap_page, offset)
  
  SELECT * FROM employee WHERE email = 'alice@co.com'
  1. Traverse email B-tree → find TID (page 42, slot 3)
  2. Read heap page 42 → get row from slot 3
  
MySQL/InnoDB secondary index:
  Leaf entries: (key_value, PRIMARY_KEY_VALUE)
  
  SELECT * FROM employee WHERE email = 'alice@co.com'  
  1. Traverse email B-tree → find PK = 1
  2. Traverse primary key B-tree with PK=1 → find row in leaf
  → TWO B-tree traversals!
```

---

## 3. Composite (Multi-Column) Indexes

```sql
CREATE INDEX idx_dept_salary ON employee(dept_id, salary);
```

This creates a B-tree sorted by `dept_id` FIRST, then by `salary` within each dept_id.

```
Conceptual sorted order:
  (1, 50000)
  (1, 75000)
  (1, 120000)
  (2, 60000)
  (2, 95000)
  (3, 80000)
  ...

This index is useful for:
  WHERE dept_id = 1                          ✓ (equality on first column)
  WHERE dept_id = 1 AND salary > 100000      ✓ (equality + range on prefix)
  WHERE dept_id = 1 ORDER BY salary          ✓ (equality + sort on next column)
  WHERE dept_id BETWEEN 1 AND 3             ✓ (range on first column)
  
This index is NOT useful for:
  WHERE salary > 100000                      ✗ (second column without first)
  ORDER BY salary                            ✗ (sort on second without first)
  WHERE salary > 100000 AND dept_id = 1      ✓ (optimizer rearranges to dept_id = 1 AND salary > 100000)
```

### The ESR Rule (Equality → Sort → Range)

For optimal composite index design:

```
Given query:
  WHERE dept_id = 1 AND hire_date > '2025-01-01' ORDER BY salary

Optimal index column order:
  1. EQUALITY columns: dept_id (= 1, exact match)
  2. SORT columns: salary (ORDER BY)
  3. RANGE columns: hire_date (> '2025-01-01')

CREATE INDEX idx_optimal ON employee(dept_id, salary, hire_date);

WHY this order?
  - dept_id = 1 narrows to a contiguous section of the index
  - Within that section, data is sorted by salary → ORDER BY is "free" (no sort needed)
  - hire_date range filter applied last (still benefits from index but as a filter)

WRONG order:
  CREATE INDEX ON employee(dept_id, hire_date, salary);
  - hire_date range breaks the sort order → salary ORDER BY requires a separate sort
```

---

## 4. Covering Indexes & Index-Only Scans

```sql
-- Query only needs dept_id and salary:
SELECT dept_id, salary FROM employee WHERE dept_id = 1;

-- Regular index on employee(dept_id):
--   1. Traverse B-tree → find TIDs for dept_id = 1
--   2. Fetch EACH heap page to get salary column
--   → Many random reads to the heap!

-- Covering index:
CREATE INDEX idx_covering ON employee(dept_id) INCLUDE (salary);
--   1. Traverse B-tree → leaf has dept_id AND salary
--   2. Return directly from index. NEVER touch the heap.
--   → Index-Only Scan!

-- INCLUDE (PostgreSQL 11+) stores columns in leaf pages but NOT in the B-tree search structure.
-- Benefit: doesn't increase tree height. Columns are just "passengers" in the leaf.

-- Check in EXPLAIN:
EXPLAIN SELECT dept_id, salary FROM employee WHERE dept_id = 1;
-- Index Only Scan using idx_covering on employee  ← success!
```

**Caveat:** Index-only scans require the Visibility Map to know the page is all-visible. Run `VACUUM` to keep it up to date.

---

## 5. Partial Indexes

```sql
-- Index only active orders (5% of table)
CREATE INDEX idx_active_orders ON orders(created_at)
WHERE status = 'active';

-- Index size: 5% of a full index. Maintenance cost: 5% of updates.
-- Query MUST include the WHERE condition to use it:
SELECT * FROM orders WHERE status = 'active' AND created_at > '2026-01-01'; ✓
SELECT * FROM orders WHERE created_at > '2026-01-01';  ✗ (can't use partial index)

-- Great patterns:
-- Soft deletes: WHERE deleted_at IS NULL
-- Unprocessed work: WHERE processed = FALSE
-- Recent data: WHERE created_at > '2026-01-01'
-- Rare but queried values: WHERE priority = 'critical'

-- Unique constraint only on non-null values:
CREATE UNIQUE INDEX idx_unique_email ON users(email) WHERE email IS NOT NULL;
```

---

## 6. Expression / Functional Indexes

```sql
-- Index on computed expression
CREATE INDEX idx_lower_email ON users(LOWER(email));

-- Now this works:
SELECT * FROM users WHERE LOWER(email) = 'alice@example.com';

-- Without expression index: can't use any index on email (LOWER changes the value)

-- More examples:
CREATE INDEX idx_year ON events(EXTRACT(YEAR FROM created_at));
CREATE INDEX idx_json_type ON api_logs((payload->>'type'));
CREATE INDEX idx_name_concat ON employee((first_name || ' ' || last_name));
CREATE INDEX idx_date_trunc ON metrics(DATE_TRUNC('hour', ts));
```

---

## 7. Hash Indexes

```sql
CREATE INDEX idx_hash_email ON users USING HASH (email);
```

```
Structure: hash table on disk
  hash(key) → bucket → list of TIDs

  bucket 0: → [(TID₁), (TID₂)]
  bucket 1: → [(TID₃)]
  bucket 2: → []
  bucket 3: → [(TID₄), (TID₅), (TID₆)]
  ...

ONLY supports equality (=). No ranges, no ordering, no multicolumn.

vs B-tree:
  Hash: O(1) lookup (one hash computation + one page read)
  B-tree: O(log N) lookup (3-4 page reads)

Sounds faster, but:
  - Hash indexes can't do ranges (WHERE x > 5, ORDER BY x)
  - Hash indexes can't do prefix matching
  - B-tree equality is already very fast (3-4 reads, top levels cached → ~1-2 reads)
  - Hash indexes had no WAL support before PostgreSQL 10 (not crash-safe!)
  - Hash indexes are auto-growing (need to rehash, can cause stalls)

Verdict: Rarely used. B-tree is better in almost all cases.
         Use only if you have extremely high-cardinality equality-only lookups
         AND profiling confirms the B-tree is a bottleneck.
```

---

## 8. GiST (Generalized Search Tree)

```
A framework for building custom balanced tree indexes.
Not a single data structure — it's an INTERFACE that supports:
  - R-trees (spatial data)
  - Full-text search indexes
  - Range type indexes
  - Custom data types

Supports operators:
  <<, >>, &&, @>, <@, =, <->  (containment, overlap, distance, etc.)
```

```sql
-- Spatial: find all restaurants within 5 km
CREATE INDEX idx_location ON restaurants USING GIST (location);
SELECT * FROM restaurants
WHERE ST_DWithin(location, ST_MakePoint(-122.4, 37.8)::geography, 5000);

-- Range types: find all bookings that overlap a date range
CREATE INDEX idx_booking_dates ON bookings USING GIST (date_range);
SELECT * FROM bookings WHERE date_range && '[2026-06-01, 2026-06-15]'::daterange;

-- Exclusion constraints (prevent overlapping ranges):
ALTER TABLE bookings ADD CONSTRAINT no_overlap
    EXCLUDE USING GIST (room_id WITH =, date_range WITH &&);
-- "No two bookings can have the same room_id AND overlapping date_range"
```

---

## 9. GIN (Generalized Inverted Index)

```
An inverted index: maps VALUES to the set of rows containing them.
Perfect for: arrays, full-text search, JSONB, any "contains" query.

Structure:
  value₁ → [TID₁, TID₃, TID₇]
  value₂ → [TID₂, TID₅]
  value₃ → [TID₁, TID₂, TID₃, TID₄]
  ...

Like the index at the back of a book:
  "B-tree" → pages 15, 42, 67, 103
  "index"  → pages 3, 15, 67
```

```sql
-- Full-text search
CREATE INDEX idx_search ON articles USING GIN (search_vector);
SELECT * FROM articles WHERE search_vector @@ to_tsquery('database & optimization');

-- JSONB containment
CREATE INDEX idx_data ON events USING GIN (data);
SELECT * FROM events WHERE data @> '{"type": "click"}';
SELECT * FROM events WHERE data ? 'error';

-- Array containment
CREATE INDEX idx_tags ON posts USING GIN (tags);
SELECT * FROM posts WHERE tags @> ARRAY['sql', 'postgresql'];

-- Trigram similarity (fuzzy text search)
CREATE EXTENSION pg_trgm;
CREATE INDEX idx_name_trgm ON users USING GIN (name gin_trgm_ops);
SELECT * FROM users WHERE name % 'Jonh';  -- finds "John" via similarity
SELECT * FROM users WHERE name ILIKE '%database%';  -- GIN trgm supports LIKE!
```

**GIN vs GiST for text search:**

| | GIN | GiST |
|---|-----|------|
| Build time | Slower (must build posting lists) | Faster |
| Index size | Larger | Smaller (lossy) |
| Lookup speed | Faster (exact posting lists) | Slower (may need recheck) |
| Update speed | Slower (insert into posting lists) | Faster |
| Best for | Read-heavy, accuracy-critical | Write-heavy, approximate OK |

---

## 10. SP-GiST (Space-Partitioned GiST)

```
Supports non-balanced, space-partitioning tree structures:
  - Quad-trees (2D spatial partitioning)
  - k-d trees (k-dimensional space partitioning)
  - Radix trees / tries (string prefixes)

Best for data with natural clustering / hierarchy.
```

```sql
-- Efficient prefix searching on IP addresses
CREATE INDEX idx_ip ON access_log USING SPGIST (ip inet_ops);
SELECT * FROM access_log WHERE ip << '192.168.1.0/24';  -- subnet containment

-- Text prefix search
CREATE INDEX idx_text ON dictionary USING SPGIST (word text_ops);
SELECT * FROM dictionary WHERE word ^@ 'pre';  -- starts with 'pre' (prefix operator)
```

---

## 11. BRIN (Block Range Index)

```
Instead of indexing every row, BRIN indexes every BLOCK RANGE (group of consecutive pages).

For each range of pages, stores: MIN and MAX value of the indexed column.

Structure:
  Pages 0-127:   min=2024-01-01, max=2024-01-15
  Pages 128-255: min=2024-01-15, max=2024-02-01
  Pages 256-383: min=2024-02-01, max=2024-02-15
  ...

Query: WHERE created_at = '2024-01-20'
  → Check: '2024-01-20' between min and max?
  → Pages 0-127: NO (max is Jan 15) → SKIP
  → Pages 128-255: YES → scan those pages
  → Pages 256-383: NO → SKIP
```

```sql
-- BRIN on naturally ordered data (timestamps, auto-increment IDs)
CREATE INDEX idx_brin_created ON events USING BRIN (created_at);

-- Tiny index! 
-- A B-tree on 1 billion rows: ~8 GB
-- A BRIN on 1 billion rows: ~1 MB (!!)

-- BRIN only works well when data is physically ordered by the indexed column.
-- append-only tables (logs, events) are perfect.
-- Randomly inserted data → BRIN is useless (every range contains every value).
```

**When to use BRIN:**
- Table is append-only or insert order correlates with the column
- Table is very large (100M+ rows)
- You need an index but B-tree is too large
- Queries filter on ranges of the column (dates, IDs)
- Acceptable to scan a few extra pages (not point-exact like B-tree)

---

## 12. Bitmap Indexes

```
NOT a separate index type in PostgreSQL — rather a query execution technique.

The executor can:
1. Scan a B-tree index → build a BITMAP of matching heap pages
2. Scan another B-tree index → build another bitmap
3. AND / OR the bitmaps together
4. Fetch only the matching heap pages

Example:
  SELECT * FROM orders WHERE status = 'active' AND region = 'US';
  
  Without bitmap: 
    Use index on status → fetch rows → filter on region (or vice versa)
  
  With bitmap:
    Bitmap Index Scan on idx_status: pages [1,0,1,1,0,0,1,...] (1=has active)  
    Bitmap Index Scan on idx_region: pages [1,1,0,1,0,1,0,...] (1=has US)
    AND bitmaps:                     pages [1,0,0,1,0,0,0,...] → only 2 pages!
    Bitmap Heap Scan: read only those 2 pages

This is why PostgreSQL can combine MULTIPLE single-column indexes effectively.
Composite indexes are still better for frequent queries, but bitmaps provide flexibility.
```

In Oracle and other databases, bitmap indexes are a real on-disk index type (very compact for low-cardinality columns). PostgreSQL only has "bitmap scans" as an execution strategy.

---

## 13. R-Tree Indexes (Spatial)

```
Implemented via GiST in PostgreSQL.

R-tree indexes multi-dimensional data (points, rectangles, polygons).

Structure: hierachical bounding rectangles.

Root: [entire map area]
  ├── [Northwest quadrant MBR]
  │     ├── [Subregion MBR] → leaf with actual geometries
  │     └── [Subregion MBR] → leaf
  └── [Southeast quadrant MBR]
        ├── [Subregion MBR] → leaf
        └── [Subregion MBR] → leaf

MBR = Minimum Bounding Rectangle

Supports queries:
  - Contains: WHERE box @> point
  - Overlaps: WHERE box1 && box2
  - Distance/nearest neighbor: ORDER BY point <-> query_point LIMIT 5
```

```sql
-- PostGIS spatial index
CREATE INDEX idx_geom ON buildings USING GIST (geom);

-- Find buildings within a polygon
SELECT * FROM buildings
WHERE ST_Within(geom, ST_MakeEnvelope(-122.5, 37.7, -122.3, 37.9, 4326));

-- K-nearest neighbors (uses index!)
SELECT name, ST_Distance(geom, ST_MakePoint(-122.4, 37.8)::geography) AS dist
FROM restaurants
ORDER BY geom <-> ST_MakePoint(-122.4, 37.8)::geometry
LIMIT 5;
```

---

## 14. Index Maintenance & Bloat

### Why Indexes Bloat

```
B-tree operations:
  DELETE: marks leaf entry as dead (doesn't remove it immediately)
  UPDATE: = DELETE old entry + INSERT new entry
  
Over time: dead entries accumulate → pages half-empty → tree too large
  
This is "index bloat."

A table with 1M rows might have an index sized for 3M rows
because 2M old entries haven't been cleaned up.
```

### Detecting Bloat

```sql
-- Check index size vs expected size
SELECT
    indexrelname AS index_name,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
    idx_scan AS times_used,
    idx_tup_read AS tuples_read,
    idx_tup_fetch AS tuples_fetched
FROM pg_stat_user_indexes
ORDER BY pg_relation_size(indexrelid) DESC;

-- pgstattuple extension for precise bloat info
CREATE EXTENSION pgstattuple;
SELECT * FROM pgstatindex('idx_employee_email');
-- avg_leaf_density: if < 70%, index is bloated
-- leaf_fragmentation: if > 30%, consider rebuild
```

### Fixing Bloat

```sql
-- REINDEX: rebuild an index from scratch
REINDEX INDEX idx_employee_email;           -- locks the table! (ACCESS EXCLUSIVE)
REINDEX INDEX CONCURRENTLY idx_employee_email; -- PostgreSQL 12+, no lock

-- Alternative: create a new index, drop the old one
CREATE INDEX CONCURRENTLY idx_email_new ON employee(email);
DROP INDEX idx_employee_email;
ALTER INDEX idx_email_new RENAME TO idx_employee_email;

-- VACUUM cleans dead tuples from indexes and heap
VACUUM employee;         -- regular vacuum: marks dead tuples as reusable
VACUUM FULL employee;    -- FULL: rewrites entire table + indexes (locks table!)
```

### Index Usage Analysis

```sql
-- Find unused indexes (wasting write performance and disk)
SELECT
    schemaname, tablename, indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS size,
    idx_scan AS scans
FROM pg_stat_user_indexes
WHERE idx_scan = 0        -- never used
AND indexrelname NOT LIKE '%_pkey'  -- don't drop primary keys
ORDER BY pg_relation_size(indexrelid) DESC;

-- Find duplicate indexes
SELECT
    array_agg(indexname) AS duplicate_indexes,
    indrelid::regclass AS table,
    pg_get_indexdef(first_value(indexrelid) OVER w) AS definition
FROM pg_stat_user_indexes
JOIN pg_index USING (indexrelid)
WHERE indisunique = FALSE
GROUP BY indrelid, indkey
HAVING count(*) > 1
WINDOW w AS (PARTITION BY indrelid, indkey ORDER BY indexrelid);

-- Find missing indexes (tables with lots of sequential scans)
SELECT
    schemaname, relname,
    seq_scan,          -- number of sequential scans
    seq_tup_read,      -- rows read by seq scans
    idx_scan,          -- number of index scans
    pg_size_pretty(pg_relation_size(relid)) AS size
FROM pg_stat_user_tables
WHERE seq_scan > 100 AND seq_tup_read > 100000
ORDER BY seq_tup_read DESC;
```

---

## 15. When NOT to Index

```
DON'T create an index when:

1. Table is small (< 10,000 rows)
   → Seq scan fits in a few pages, faster than index indirection

2. Column has very low cardinality AND you query for common values
   → WHERE gender = 'M' on a 50/50 split → index scan reads 50% of table  
     + random I/O overhead → seq scan is faster
   → EXCEPTION: partial index for RARE values: WHERE status = 'error'

3. Table is write-heavy with few reads
   → Every INSERT/UPDATE/DELETE must maintain ALL indexes
   → Each index on a table adds ~30-50% write overhead

4. Column is never used in WHERE, JOIN, or ORDER BY
   → Index provides zero benefit

5. You already have a composite index with this column as a prefix
   → Index on (A, B) covers queries on (A) alone. Don't create a separate index on (A).
   → EXCEPTION: index on (A) alone is smaller → better for A-only queries if B column is large

6. Query always needs a full table scan anyway
   → Aggregation over entire table: SELECT COUNT(*) FROM orders
     No index helps (unless index-only scan on a small index)
```

---

## Summary: Choosing the Right Index

| Task | Index Type |
|------|-----------|
| Equality and range queries on scalar columns | **B-tree** (default) |
| Full-text search | **GIN** with tsvector |
| JSONB containment / key existence | **GIN** |
| Array containment / overlap | **GIN** |
| Fuzzy text matching (LIKE, ILIKE, similarity) | **GIN** with pg_trgm |
| Spatial / geometric queries | **GiST** (R-tree) |
| Range type overlap / exclusion constraints | **GiST** |
| Nearest-neighbor search | **GiST** (with ORDER BY <->) |
| IP address / CIDR lookups | **SP-GiST** or GiST |
| Text prefix search (^@) | **SP-GiST** |
| Append-only time-series data | **BRIN** |
| Large table, naturally ordered column | **BRIN** |
| Equality-only on very high cardinality | **Hash** (rarely worth it) |

---

Next: [03-query-processing.md](03-query-processing.md) →
