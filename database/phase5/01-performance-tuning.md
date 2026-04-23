# 5.1 — Performance Tuning & Benchmarking

> Performance tuning is where theory meets production.  
> You need a SYSTEMATIC methodology — not random parameter tweaking.  
> The goal: find the bottleneck, fix it, measure, repeat.

---

## 1. The Performance Tuning Methodology

```
Step 1: DEFINE the problem
  "Slow" is not a problem. Be specific:
  "The /api/orders endpoint takes 2.3 seconds at p99 (target: 200ms)"

Step 2: MEASURE (establish baseline)
  - Capture current metrics: QPS, latency (p50/p95/p99), CPU, I/O, connections
  - Identify the BOTTLENECK: is it CPU, I/O, memory, locks, or network?

Step 3: IDENTIFY the slow queries
  pg_stat_statements → top queries by total_exec_time
  slow query log → individual slow query instances

Step 4: ANALYZE the query plan
  EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) SELECT ...
  Understand: is it a sequential scan? nested loop on wrong index? sort spill?

Step 5: FIX one thing at a time
  Add index, rewrite query, tune parameter, restructure table

Step 6: VERIFY the improvement
  Re-measure. Compare p99 before and after.
  Did fixing this query create a different bottleneck?

Step 7: REPEAT until target met

NEVER: skip straight to random configuration changes.
ALWAYS: measure → identify → fix → verify.
```

---

## 2. Finding Slow Queries

### PostgreSQL

```sql
-- pg_stat_statements (must be in shared_preload_libraries):
-- Top 20 queries by TOTAL time (impact on system):
SELECT query,
       calls,
       ROUND(total_exec_time::NUMERIC, 2) AS total_ms,
       ROUND(mean_exec_time::NUMERIC, 2) AS mean_ms,
       ROUND((100 * total_exec_time / SUM(total_exec_time) OVER())::NUMERIC, 2) AS pct,
       rows
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 20;

-- Top queries by MEAN time (individually slow):
SELECT query, calls, ROUND(mean_exec_time::NUMERIC, 2) AS mean_ms,
       ROUND(stddev_exec_time::NUMERIC, 2) AS stddev_ms
FROM pg_stat_statements
WHERE calls > 100
ORDER BY mean_exec_time DESC
LIMIT 20;

-- I/O-heavy queries (read lots of blocks):
SELECT query, calls,
       shared_blks_hit + shared_blks_read AS total_blks,
       shared_blks_read AS disk_reads,
       ROUND(shared_blks_hit::NUMERIC /
             NULLIF(shared_blks_hit + shared_blks_read, 0) * 100, 2) AS cache_hit_pct
FROM pg_stat_statements
ORDER BY shared_blks_read DESC
LIMIT 20;

-- Slow query log (catches individual instances):
-- postgresql.conf:
log_min_duration_statement = 500  -- log queries > 500ms
log_statement = 'none'            -- don't log all statements
auto_explain.log_min_duration = '1s'  -- auto-log plans for slow queries
auto_explain.log_analyze = true
auto_explain.log_buffers = true
```

### MySQL

```sql
-- Enable slow query log:
SET GLOBAL slow_query_log = ON;
SET GLOBAL long_query_time = 0.5;  -- 500ms threshold
SET GLOBAL log_queries_not_using_indexes = ON;

-- Performance Schema (detailed breakdown):
SELECT DIGEST_TEXT, COUNT_STAR, AVG_TIMER_WAIT/1000000000 AS avg_ms,
       SUM_TIMER_WAIT/1000000000 AS total_ms,
       SUM_ROWS_EXAMINED, SUM_ROWS_SENT
FROM performance_schema.events_statements_summary_by_digest
ORDER BY SUM_TIMER_WAIT DESC
LIMIT 20;

-- pt-query-digest (Percona Toolkit):
-- $ pt-query-digest /var/log/mysql/slow.log
```

---

## 3. Reading EXPLAIN Plans Like a Pro

### PostgreSQL EXPLAIN

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT o.id, c.name, o.total
FROM orders o
JOIN customers c ON o.customer_id = c.id
WHERE o.status = 'pending'
  AND o.created_at > '2024-01-01'
ORDER BY o.created_at DESC
LIMIT 20;

-- Sample output:
Limit  (cost=1234.56..1234.78 rows=20 width=52) (actual time=12.3..12.5 rows=20 loops=1)
  Buffers: shared hit=1500 read=45
  -> Sort  (cost=1234.56..1245.67 rows=4444 width=52) (actual time=12.3..12.4 rows=20 loops=1)
        Sort Key: o.created_at DESC
        Sort Method: top-N heapsort  Memory: 27kB
        Buffers: shared hit=1500 read=45
        -> Hash Join  (cost=100.00..1100.00 rows=4444 width=52) (actual time=2.1..11.8 rows=4500 loops=1)
              Hash Cond: (o.customer_id = c.id)
              Buffers: shared hit=1500 read=45
              -> Bitmap Heap Scan on orders o  (cost=55.00..900.00 rows=4444 width=36) (actual time=1.5..9.2 rows=4500 loops=1)
                    Recheck Cond: (status = 'pending' AND created_at > '2024-01-01')
                    Heap Blocks: exact=420
                    Buffers: shared hit=1400 read=45
                    -> Bitmap Index Scan on idx_orders_status_date  (cost=0.00..54.00 rows=4444 width=0) (actual time=1.2..1.2 rows=4500 loops=1)
                          Index Cond: (status = 'pending' AND created_at > '2024-01-01')
                          Buffers: shared hit=15
              -> Hash  (cost=35.00..35.00 rows=1000 width=20) (actual time=0.5..0.5 rows=1000 loops=1)
                    Buckets: 1024  Batches: 1  Memory Usage: 56kB
                    Buffers: shared hit=100
                    -> Seq Scan on customers c  (cost=0.00..35.00 rows=1000 width=20) (actual time=0.01..0.3 rows=1000 loops=1)
                          Buffers: shared hit=100

-- HOW TO READ THIS:
-- 1. Read BOTTOM-UP (execution starts at innermost nodes)
-- 2. Compare "rows" estimate vs "actual rows" → if wildly off, ANALYZE the table
-- 3. Look at "Buffers: shared read=45" → 45 blocks from DISK (cache miss)
-- 4. "Sort Method: top-N heapsort Memory: 27kB" → sorted in memory (good)
-- 5. "Bitmap Heap Scan" → used index, then fetched matching heap pages

-- RED FLAGS in EXPLAIN:
-- ✗ Seq Scan on large table (missing index)
-- ✗ Nested Loop with inner Seq Scan (missing index for join)
-- ✗ Sort Method: external merge Disk: ...  (work_mem too small → spilled to disk)
-- ✗ rows=1000 actual rows=500000  (bad statistics → run ANALYZE)
-- ✗ Hash Batches: 4  (hash table didn't fit in work_mem)
-- ✗ Buffers: shared read=50000  (massive cache misses → need more shared_buffers)
```

---

## 4. Index Tuning Strategy

```sql
-- The ESR Rule (Phase 2 refresher, applied here):
-- Equality → Sort → Range
-- Build composite indexes in this column order.

-- Example: WHERE status = 'pending' AND created_at > '2024-01-01' ORDER BY created_at
CREATE INDEX idx_orders_status_date ON orders (status, created_at);
-- status = equality (exact match)
-- created_at = both sort AND range → works in this order

-- Find missing indexes (PostgreSQL):
SELECT relname, seq_scan, seq_tup_read,
       idx_scan, idx_tup_fetch,
       seq_scan - idx_scan AS too_many_seq_scans
FROM pg_stat_user_tables
WHERE seq_scan > idx_scan
  AND pg_relation_size(relid) > 10000000  -- > 10 MB
ORDER BY seq_tup_read DESC
LIMIT 20;

-- Find duplicate indexes:
SELECT pg_size_pretty(SUM(pg_relation_size(idx))::BIGINT) AS size,
       (array_agg(idx))[1] AS idx1, (array_agg(idx))[2] AS idx2,
       (array_agg(indkey))[1] AS cols
FROM (
    SELECT indexrelid::regclass AS idx, indrelid, indkey,
           indkey::TEXT
    FROM pg_index
) sub
GROUP BY indrelid, indkey::TEXT
HAVING COUNT(*) > 1
ORDER BY SUM(pg_relation_size(idx)) DESC;

-- Find unused indexes (wasting space and slowing writes):
SELECT indexrelname, idx_scan,
       pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
WHERE idx_scan < 50  -- used fewer than 50 times
  AND indexrelname NOT LIKE '%pkey%'  -- keep primary keys
  AND indexrelname NOT LIKE '%unique%'  -- keep unique constraints
ORDER BY pg_relation_size(indexrelid) DESC;

-- Covering index (Index-Only Scan):
CREATE INDEX idx_orders_covering ON orders (status, created_at)
INCLUDE (id, total);
-- Now queries selecting only id, total can be served from index alone
-- No heap fetch → 2-5x faster for specific queries
```

---

## 5. Connection Pool Sizing

```
Optimal pool size is MUCH SMALLER than people think.

Formula (from PostgreSQL wiki):
  connections = (core_count * 2) + effective_spindle_count
  
  8-core server with SSD:
  connections = (8 * 2) + 1 = 17
  
  → 20-30 connections is often optimal for OLTP
  → Adding MORE connections makes it SLOWER (context switching, lock contention)

HikariCP (Java):
  maximumPoolSize: 10-20 (not 100!)
  minimumIdle: same as max (avoid connection creation overhead)
  connectionTimeout: 30000  (30s — fail fast)
  idleTimeout: 600000  (10 min)
  maxLifetime: 1800000  (30 min — before PostgreSQL idle_in_transaction_session_timeout)

PgBouncer (PostgreSQL):
  default_pool_size = 20
  max_client_conn = 1000  (handle many app connections with few PG connections)
  pool_mode = transaction  (release connection back to pool after each transaction)

ProxySQL (MySQL):
  mysql-max_connections = 2000  (client side)
  hostgroup backend max_connections = 20  (per MySQL server)

The pattern:
  App instances (100s of connections) → Connection pool → Database (20-30 connections)
```

---

## 6. Benchmarking

```bash
# pgbench (PostgreSQL built-in):
# Initialize:
pgbench -i -s 100 mydb       # create tables with scale factor 100 (10M rows)

# Run read-write benchmark:
pgbench -c 20 -j 4 -T 60 mydb
# -c 20: 20 concurrent connections
# -j 4: 4 threads
# -T 60: run for 60 seconds

# Custom benchmark:
pgbench -c 20 -j 4 -T 60 -f custom.sql mydb

# sysbench (MySQL + PostgreSQL):
sysbench oltp_read_write --db-driver=pgsql \
  --pgsql-host=localhost --pgsql-db=mydb \
  --tables=10 --table-size=1000000 \
  --threads=16 --time=300 run

# YCSB (Yahoo! Cloud Serving Benchmark):
# Standard workloads: A (50/50 r/w), B (95/5 r/w), C (100% read), 
#                     D (read latest), E (short ranges), F (read-modify-write)
bin/ycsb load basic -P workloads/workloada
bin/ycsb run basic -P workloads/workloada

# Key metrics to capture:
# Throughput: transactions/sec (TPS) or queries/sec (QPS)
# Latency: p50, p95, p99, p99.9, max
# Error rate: failed transactions / total
# Resource usage: CPU, memory, disk I/O, network I/O

# TPC benchmarks (industry standard):
# TPC-C: OLTP (orders, payments, deliveries)
# TPC-H: OLAP (analytical queries, decision support)
# TPC-DS: data warehouse (complex analytics)
```

---

## 7. Caching Strategies

```
Cache layers (closest to user → furthest):

  Browser/CDN cache → Application cache → Database result cache → Buffer pool → Disk

Application-Level Caching (Redis/Memcached):

Cache-Aside (Lazy Loading):
  1. App checks cache
  2. Cache miss → query database
  3. Store result in cache with TTL
  4. Return to user
  
  read(key):
    value = cache.get(key)
    if value is None:
      value = db.query(key)
      cache.set(key, value, ttl=300)
    return value

Write-Through:
  1. App writes to cache AND database
  2. Cache is always up-to-date
  
  write(key, value):
    db.update(key, value)
    cache.set(key, value)

Write-Behind (Write-Back):
  1. App writes to cache only
  2. Cache asynchronously writes to database
  → Faster writes, risk of data loss if cache crashes

Cache Invalidation (the hard problem):
  TTL-based: expire after N seconds (simple, eventual staleness)
  Event-based: LISTEN/NOTIFY or CDC → invalidate on change
  Version-based: cache key includes version → on update, increment version
  
  "There are only two hard things in computer science:
   cache invalidation and naming things." — Phil Karlton

Materialized Views (database-level caching):
  CREATE MATERIALIZED VIEW dashboard_stats AS
  SELECT date_trunc('hour', created_at) AS hour,
         count(*) AS orders, sum(total) AS revenue
  FROM orders
  GROUP BY hour;
  
  REFRESH MATERIALIZED VIEW CONCURRENTLY dashboard_stats;
  -- CONCURRENTLY: doesn't lock the view during refresh (needs unique index)
```

---

## Key Takeaways

1. **Methodology over intuition.** Measure → identify bottleneck → fix → verify. Never tune randomly.
2. **pg_stat_statements is your #1 tool.** Sort by `total_exec_time` to find the queries with the highest system impact.
3. **Read EXPLAIN bottom-up.** Watch for: estimated vs actual rows mismatch, disk sort spills, sequential scans on large tables, high buffer reads.
4. **Connection pool size should be 20-30**, not hundreds. More connections = more contention = slower for everyone.
5. **Caching is a system design decision,** not a band-aid. Cache-aside with TTL is the safe default. Event-driven invalidation for real-time needs.
6. **Benchmark with realistic data.** pgbench scale factor should match production size. Test at expected QPS, not just max throughput.

---

Next: [02-high-availability.md](02-high-availability.md) →
