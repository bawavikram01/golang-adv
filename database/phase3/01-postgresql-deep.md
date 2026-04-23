# 3.1 — PostgreSQL Deep Mastery

> PostgreSQL is the **most advanced open-source relational database**.  
> If you master one database to god level, make it this one.  
> Everything you learned in Phase 2 — this is where it becomes real.

---

## 1. Architecture Overview

```
                     Client Application
                           │
                           ▼ TCP/IP (port 5432)
                    ┌──────────────┐
                    │  Postmaster   │  (main daemon — listens, forks)
                    │  (PID 1)      │
                    └──────┬───────┘
                           │ fork() per connection
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ Backend 1 │    │ Backend 2 │    │ Backend 3 │   ← one OS process per connection
    │ (client A)│    │ (client B)│    │ (client C)│
    └──────────┘    └──────────┘    └──────────┘
          │                │                │
          └────────────────┼────────────────┘
                           │
                    ┌──────▼──────┐
                    │ Shared Memory│
                    │ ┌──────────┐│
                    │ │Buffer Pool││  ← shared_buffers
                    │ │WAL Buffers││  ← wal_buffers
                    │ │Lock Table ││
                    │ │CLOG/XACT  ││  ← commit status cache
                    │ │Proc Array ││  ← snapshot info
                    │ └──────────┘│
                    └─────────────┘
                           │
     ┌─────────────────────┼──────────────────────┐
     ▼                     ▼                      ▼
┌──────────┐        ┌──────────┐          ┌──────────────┐
│Background│        │Background│          │ Background    │
│Writer    │        │WAL Writer│          │ Autovacuum    │
│(bgwriter)│        │          │          │ Launcher +    │
│          │        │Flushes   │          │ Workers       │
│Flushes   │        │WAL       │          │               │
│dirty     │        │buffers   │          │Cleans dead    │
│pages     │        │to disk   │          │tuples         │
└──────────┘        └──────────┘          └──────────────┘
     │                     │
     ▼                     ▼
┌──────────┐        ┌──────────┐
│Checkpointer│      │WAL Archiver│
│           │        │           │
│Periodic   │        │Copies WAL │
│full flush │        │segments to│
│+ checkpoint│       │archive    │
│record     │        └──────────┘
└──────────┘

     ┌──────────┐     ┌──────────────┐
     │Stats     │     │Logical Repli-│
     │Collector │     │cation Worker │
     └──────────┘     └──────────────┘
```

### Process Model (NOT threads)

```
PostgreSQL uses one OS PROCESS per connection (not threads).

Why processes (not threads)?
  1. Crash isolation:   a segfault in one backend doesn't kill others
  2. Simpler code:      no thread-safety concerns in legacy C code
  3. OS scheduling:     OS handles process scheduling, NUMA, etc.

Downsides:
  - High memory overhead per connection (~5-10 MB idle)
  - fork() cost on connection (~1-5 ms)
  - Limited to ~hundreds of connections (use pooler!)
  - Context switch overhead between processes

PostgreSQL 17+: considering background worker threads for internal parallelism
Industry trend: MySQL, Oracle use thread-per-connection → lower overhead
```

---

## 2. System Catalogs — The Metadata Database

PostgreSQL stores all metadata (table definitions, indexes, types, functions) in **system catalog tables** — regular tables you can query.

```sql
-- Key catalog tables:
pg_class        -- all relations (tables, indexes, views, sequences, materialized views)
pg_attribute    -- all columns of all relations
pg_index        -- all indexes
pg_namespace    -- schemas
pg_type         -- data types
pg_proc         -- functions and procedures
pg_constraint   -- constraints (PK, FK, CHECK, UNIQUE, EXCLUSION)
pg_trigger      -- triggers
pg_depend       -- dependency tracking
pg_stat_*       -- runtime statistics views
pg_settings     -- all configuration parameters
pg_locks        -- current locks

-- Find all tables in current database:
SELECT relname, relkind, reltuples::BIGINT, relpages
FROM pg_class
WHERE relkind = 'r'  -- 'r' = ordinary table
  AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
ORDER BY reltuples DESC;

-- relkind values:
-- 'r' = table, 'i' = index, 'v' = view, 'm' = materialized view,
-- 'S' = sequence, 't' = TOAST table, 'p' = partitioned table

-- Find all columns of a table:
SELECT attname, format_type(atttypid, atttypmod) AS data_type,
       attnotnull AS not_null, atthasdef AS has_default
FROM pg_attribute
WHERE attrelid = 'employee'::regclass
  AND attnum > 0        -- skip system columns
  AND NOT attisdropped   -- skip dropped columns
ORDER BY attnum;

-- Find all indexes on a table:
SELECT i.relname AS index_name,
       pg_get_indexdef(i.oid) AS definition,
       ix.indisunique, ix.indisprimary,
       pg_size_pretty(pg_relation_size(i.oid)) AS size
FROM pg_index ix
JOIN pg_class i ON i.oid = ix.indexrelid
WHERE ix.indrelid = 'employee'::regclass;

-- Find all foreign keys referencing a table:
SELECT conname, conrelid::regclass AS source_table,
       pg_get_constraintdef(oid) AS definition
FROM pg_constraint
WHERE confrelid = 'department'::regclass
  AND contype = 'f';

-- See what's happening RIGHT NOW:
SELECT pid, state, query, wait_event_type, wait_event,
       now() - query_start AS duration
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC;
```

### System Columns (Hidden)

```sql
-- Every table has hidden system columns:
SELECT ctid, xmin, xmax, tableoid, *
FROM employee LIMIT 5;

-- ctid:     (page, offset) — physical tuple location
-- xmin:     transaction ID that inserted this row
-- xmax:     transaction ID that deleted/updated this row (0 if live)
-- tableoid: OID of the table (useful with inheritance)
-- cmin/cmax: command IDs within xmin/xmax transaction
```

---

## 3. VACUUM — The Lifeline of PostgreSQL

### Why VACUUM Exists

```
Because of MVCC, UPDATE and DELETE don't remove old row versions.
They mark them as "dead" (set xmax).
Dead tuples accumulate → table bloats → performance degrades.

VACUUM reclaims space occupied by dead tuples.

Without VACUUM:
  - Tables grow without bound
  - Sequential scans slow down (reading dead tuples)
  - Indexes bloat (entries point to dead tuples)
  - Eventually: TRANSACTION ID WRAPAROUND → data loss!
```

### Types of VACUUM

```sql
-- REGULAR VACUUM (non-blocking):
VACUUM employee;
-- 1. Scans table pages for dead tuples
-- 2. Removes dead tuples, marks space as reusable (within the page)
-- 3. Updates the Free Space Map (FSM)
-- 4. Updates the Visibility Map (VM) for all-visible pages
-- 5. Truncates empty pages at END of file (if possible)
-- 6. Does NOT return space to OS (except trailing empty pages)
-- 7. Does NOT block reads or writes (concurrent safe)

-- VACUUM FULL (blocking, rewrite):
VACUUM FULL employee;
-- 1. Creates a NEW copy of the table (only live tuples)
-- 2. Rebuilds ALL indexes
-- 3. Drops old table file, replaces with new
-- 4. RETURNS SPACE TO OS
-- 5. ACQUIRES ACCESS EXCLUSIVE LOCK — blocks ALL operations!
-- 6. Needs ~2x table size of free disk space

-- VACUUM with options:
VACUUM (VERBOSE) employee;          -- show detailed output
VACUUM (ANALYZE) employee;          -- also update planner statistics
VACUUM (PARALLEL 4) employee;       -- parallel index vacuuming (PG 13+)
VACUUM (INDEX_CLEANUP OFF) employee; -- skip index cleanup (emergency speed)
VACUUM (TRUNCATE OFF) employee;     -- skip file truncation (avoids lock)

-- ANALYZE (statistics only, no dead tuple cleanup):
ANALYZE employee;
ANALYZE employee(salary, dept_id);  -- specific columns only
```

### Autovacuum — The Background Cleaner

```sql
-- Autovacuum is a background process that automatically runs VACUUM and ANALYZE.
-- It is CRITICAL. Never disable it.

-- Key parameters:
autovacuum = on                         -- never turn off
autovacuum_max_workers = 3              -- parallel autovacuum workers
autovacuum_naptime = '1min'             -- how often to check for work

-- When does autovacuum trigger on a table?
-- VACUUM trigger:
--   dead tuples > autovacuum_vacuum_threshold + autovacuum_vacuum_scale_factor × reltuples
--   Default: 50 + 0.20 × reltuples
--   A table with 1M rows: vacuum when > 200,050 dead tuples (20% dead)
--   PROBLEM: for a 100M row table, that's 20M dead tuples before vacuum!

-- ANALYZE trigger:
--   changed tuples > autovacuum_analyze_threshold + autovacuum_analyze_scale_factor × reltuples
--   Default: 50 + 0.10 × reltuples (10% changed)

-- Per-table overrides (critical for large tables):
ALTER TABLE orders SET (
    autovacuum_vacuum_scale_factor = 0.01,   -- vacuum at 1% dead (not 20%)
    autovacuum_vacuum_threshold = 1000,
    autovacuum_analyze_scale_factor = 0.005,
    autovacuum_vacuum_cost_delay = 2         -- less throttling (ms)
);

-- Monitor autovacuum:
SELECT relname, n_live_tup, n_dead_tup,
       ROUND(n_dead_tup::NUMERIC / NULLIF(n_live_tup + n_dead_tup, 0) * 100, 2) AS dead_pct,
       last_vacuum, last_autovacuum, last_analyze, last_autoanalyze
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;
```

### Transaction ID Wraparound — The Silent Killer

```
PostgreSQL transaction IDs (XIDs) are 32-bit unsigned integers: 0 to ~4.2 billion.
XIDs wrap around: after 4.2 billion transactions, XID 1 is reused.

The visibility rule uses XID comparison:
  "Is this tuple's xmin in the past (visible) or future (invisible)?"

After wraparound: old committed transactions look like they're "in the future"
  → Their rows become INVISIBLE → DATA DISAPPEARS.

Prevention: VACUUM FREEZE
  Vacuum replaces old XIDs with a special "frozen" XID that is ALWAYS in the past.
  
  If autovacuum can't keep up (blocked by long-running transactions,
  table locks, or disabled autovacuum):
  
  WARNING: database "mydb" must be vacuumed within 10000000 transactions
  FATAL:   database is not accepting commands to avoid wraparound
  
  The database goes READ-ONLY to force you to vacuum.
  
  To prevent:
  1. Never disable autovacuum
  2. Don't run transactions that last hours/days
  3. Monitor oldest unfrozen XID age:
  
  SELECT datname, age(datfrozenxid) AS xid_age,
         current_setting('autovacuum_freeze_max_age')::INT AS freeze_threshold
  FROM pg_database
  ORDER BY xid_age DESC;
  
  -- If xid_age > 1 billion → you have a problem
  -- autovacuum_freeze_max_age default: 200 million
```

---

## 4. HOT Updates, Fillfactor, and Tuple Chaining

```sql
-- HOT (Heap-Only Tuple) updates:
-- When no indexed column is changed AND the new tuple fits on the same page:
-- → No index update needed → 5-10x faster UPDATE

-- Enable by leaving free space per page:
ALTER TABLE orders SET (fillfactor = 70);
-- 30% of each page reserved for HOT updates

-- Check HOT ratio:
SELECT relname,
       n_tup_upd AS updates,
       n_tup_hot_upd AS hot_updates,
       ROUND(n_tup_hot_upd::NUMERIC / NULLIF(n_tup_upd, 0) * 100, 2) AS hot_pct
FROM pg_stat_user_tables
WHERE n_tup_upd > 0
ORDER BY n_tup_upd DESC;

-- Goal: hot_pct > 90% for frequently-updated tables
-- If low: check which indexes cover updated columns
```

---

## 5. TOAST — Large Value Storage

```sql
-- Rows > ~2 KB trigger TOAST:
-- 1. Try to compress the value (pglz or lz4)
-- 2. If still too big, slice into chunks in a separate TOAST table

-- Check TOAST table:
SELECT relname, reltoastrelid::regclass AS toast_table
FROM pg_class WHERE relname = 'articles';

-- TOAST strategies per column:
-- PLAIN:    no TOAST (small fixed types: int, float)
-- EXTENDED: compress + external (default for text, jsonb, bytea)
-- EXTERNAL: external without compression
-- MAIN:     compress, avoid external if possible

-- Change strategy:
ALTER TABLE articles ALTER COLUMN body SET STORAGE EXTERNAL;
-- Use EXTERNAL for pre-compressed data (images, compressed files)

-- PostgreSQL 14+: LZ4 compression (faster than default pglz)
ALTER TABLE articles ALTER COLUMN body SET COMPRESSION lz4;
-- Or set default: default_toast_compression = 'lz4'

-- TOAST performance impact:
-- Accessing a TOASTed column = extra I/O to read TOAST table chunks
-- SELECT * on tables with large text/jsonb triggers TOAST reads for EVERY row
-- → Only SELECT the columns you actually need!
```

---

## 6. Replication

### Streaming Replication (Physical)

```
Primary → copies exact WAL bytes → Standby

Primary server:                 Standby server:
┌────────────┐                 ┌────────────┐
│ Accept R/W │                 │ Read-only   │
│ Generate WAL├───WAL stream──→│ Apply WAL   │
│            │                 │ (replay)    │
└────────────┘                 └────────────┘

# Primary: postgresql.conf
wal_level = replica              # or 'logical' for logical replication too
max_wal_senders = 10             # max number of standby connections
wal_keep_size = '1GB'            # keep WAL segments for slow standbys

# Primary: pg_hba.conf
host replication replicator standby_ip/32 scram-sha-256

# Standby: create base backup + configure
pg_basebackup -h primary -U replicator -D /var/lib/postgresql/data -Fp -Xs -P

# Standby: postgresql.conf (auto-configured by pg_basebackup in PG12+)
primary_conninfo = 'host=primary port=5432 user=replicator password=...'
# standby.signal file present → enters standby mode

Synchronous replication:
  # On primary:
  synchronous_standby_names = 'FIRST 1 (standby1, standby2)'
  # Commit waits until at least 1 standby confirms WAL receipt
  # → Zero data loss, but higher commit latency
  
  # synchronous_commit options:
  # on       — wait for local WAL flush (default)
  # remote_write — wait for standby to receive (not flush)
  # remote_apply — wait for standby to replay (strongest)
  # off      — don't wait for WAL flush (fastest, risk of data loss)
```

### Logical Replication

```
Primary → decodes WAL into logical changes → Subscriber

Unlike streaming replication:
  - Can replicate SPECIFIC TABLES (not entire database)
  - Can replicate between DIFFERENT PostgreSQL versions
  - Subscriber can have its OWN tables, indexes, different schema
  - Can replicate to different databases on same server
  - Subscriber is WRITABLE (careful: no conflict detection!)

# Publisher (primary):
CREATE PUBLICATION my_pub FOR TABLE orders, customers;
-- Or: FOR ALL TABLES

# Subscriber:
CREATE SUBSCRIPTION my_sub
  CONNECTION 'host=primary dbname=mydb user=replicator'
  PUBLICATION my_pub;

-- Check replication status:
-- On publisher:
SELECT * FROM pg_stat_replication;
-- On subscriber:
SELECT * FROM pg_stat_subscription;

-- Replication slots (prevent WAL from being recycled before subscriber reads it):
SELECT * FROM pg_replication_slots;
-- WARNING: an inactive slot prevents WAL cleanup → disk fills up!
-- Monitor and drop unused slots.
```

---

## 7. Connection Pooling with PgBouncer

```ini
# pgbouncer.ini
[databases]
mydb = host=localhost port=5432 dbname=mydb

[pgbouncer]
listen_port = 6432
listen_addr = 0.0.0.0
auth_type = scram-sha-256
auth_file = /etc/pgbouncer/userlist.txt

pool_mode = transaction        # session | transaction | statement
default_pool_size = 20         # PG connections per user/database pair
max_client_conn = 1000         # max client connections to PgBouncer
reserve_pool_size = 5          # extra connections for burst
reserve_pool_timeout = 3       # seconds before using reserve pool

# Connection aging:
server_lifetime = 3600         # close PG connections after 1 hour
server_idle_timeout = 600      # close idle PG connections after 10 min

# Monitoring:
SHOW pools;    -- connection stats
SHOW stats;    -- query stats
SHOW clients;  -- connected clients
SHOW servers;  -- backend connections
```

---

## 8. Extensions Ecosystem

```sql
-- PostgreSQL's killer feature: EXTENSIONS

-- List installed extensions:
SELECT * FROM pg_available_extensions WHERE installed_version IS NOT NULL;

-- Must-know extensions:

-- 1. pg_stat_statements — query performance tracking
CREATE EXTENSION pg_stat_statements;
-- Add to shared_preload_libraries in postgresql.conf (requires restart)
SELECT query, calls, mean_exec_time, total_exec_time, rows,
       shared_blks_hit, shared_blks_read
FROM pg_stat_statements
ORDER BY total_exec_time DESC LIMIT 20;

-- 2. PostGIS — geospatial
CREATE EXTENSION postgis;
SELECT ST_Distance(
    ST_MakePoint(-122.4194, 37.7749)::geography,
    ST_MakePoint(-73.9857, 40.7484)::geography
) / 1000 AS km;  -- ~4139 km (SF to NYC)

-- 3. pg_trgm — trigram similarity / fuzzy search
CREATE EXTENSION pg_trgm;
SELECT similarity('PostgreSQL', 'Postrgres');  -- 0.5
CREATE INDEX idx_trgm ON users USING GIN (name gin_trgm_ops);

-- 4. pgcrypto — cryptographic functions
CREATE EXTENSION pgcrypto;
SELECT crypt('my_password', gen_salt('bf'));  -- bcrypt hash
SELECT gen_random_uuid();  -- UUIDv4

-- 5. pgvector — vector similarity search (AI/ML embeddings)
CREATE EXTENSION vector;
CREATE TABLE items (id SERIAL, embedding vector(1536));
CREATE INDEX ON items USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
SELECT * FROM items ORDER BY embedding <=> '[0.1, 0.2, ...]' LIMIT 10;

-- 6. TimescaleDB — time-series optimization
-- Converts tables into hypertables partitioned by time
SELECT create_hypertable('metrics', 'timestamp');

-- 7. Citus — distributed PostgreSQL (sharding)
SELECT create_distributed_table('orders', 'customer_id');

-- 8. pg_partman — automated partition management
CREATE EXTENSION pg_partman;
SELECT partman.create_parent('public.events', 'created_at', 'native', 'daily');

-- 9. pg_repack — online table repack (avoid VACUUM FULL locks)
-- $ pg_repack -d mydb -t bloated_table

-- 10. pgaudit — audit logging
CREATE EXTENSION pgaudit;
SET pgaudit.log = 'write, ddl';
```

---

## 9. Performance Tuning — Production Configuration

```sql
-- === CONNECTIONS ===
max_connections = 100                    -- use PgBouncer, keep this LOW
superuser_reserved_connections = 3

-- === MEMORY ===
shared_buffers = '16GB'                  -- 25% of 64GB RAM
effective_cache_size = '48GB'            -- 75% of RAM
work_mem = '64MB'                        -- per sort/hash operation
maintenance_work_mem = '2GB'             -- VACUUM, CREATE INDEX
huge_pages = 'try'

-- === WAL ===
wal_level = 'replica'
wal_buffers = '64MB'
max_wal_size = '4GB'
min_wal_size = '1GB'
checkpoint_completion_target = 0.9
wal_compression = 'lz4'                 -- PG 15+

-- === QUERY PLANNER ===
random_page_cost = 1.1                   -- SSD (default 4.0 is for HDD)
effective_io_concurrency = 200           -- SSD
default_statistics_target = 200          -- more histogram buckets (default 100)

-- === PARALLELISM ===
max_parallel_workers = 8
max_parallel_workers_per_gather = 4
max_parallel_maintenance_workers = 4
parallel_tuple_cost = 0.01
parallel_setup_cost = 100

-- === AUTOVACUUM ===
autovacuum_max_workers = 4
autovacuum_vacuum_cost_delay = '2ms'     -- less throttling on SSD (default 2ms PG 12+)
autovacuum_vacuum_cost_limit = 1000      -- higher limit

-- === LOGGING ===
log_min_duration_statement = '500ms'     -- log slow queries
log_checkpoints = on
log_lock_waits = on
log_temp_files = 0                       -- log ALL temp file usage
log_autovacuum_min_duration = '1s'
```

---

## 10. Essential Monitoring Queries

```sql
-- Active queries and their state:
SELECT pid, usename, state, wait_event_type, wait_event,
       now() - query_start AS duration, left(query, 80) AS query
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC;

-- Table I/O statistics:
SELECT relname,
       heap_blks_read, heap_blks_hit,
       ROUND(heap_blks_hit::NUMERIC / NULLIF(heap_blks_hit + heap_blks_read, 0) * 100, 2) AS cache_hit_pct,
       idx_blks_read, idx_blks_hit
FROM pg_statio_user_tables
ORDER BY heap_blks_read DESC LIMIT 15;

-- Table bloat estimate:
SELECT relname, n_live_tup, n_dead_tup,
       pg_size_pretty(pg_total_relation_size(relid)) AS total_size,
       ROUND(n_dead_tup::NUMERIC / NULLIF(n_live_tup, 0) * 100, 2) AS bloat_pct
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC LIMIT 15;

-- Index usage (find unused indexes):
SELECT indexrelname, idx_scan, idx_tup_read,
       pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
WHERE idx_scan < 50  -- rarely used
ORDER BY pg_relation_size(indexrelid) DESC;

-- Lock monitoring:
SELECT blocked.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_query,
       blocking_activity.query AS blocking_query
FROM pg_catalog.pg_locks blocked
JOIN pg_catalog.pg_locks blocking ON blocking.locktype = blocked.locktype
    AND blocking.database IS NOT DISTINCT FROM blocked.database
    AND blocking.relation IS NOT DISTINCT FROM blocked.relation
    AND blocking.page IS NOT DISTINCT FROM blocked.page
    AND blocking.tuple IS NOT DISTINCT FROM blocked.tuple
    AND blocking.transactionid IS NOT DISTINCT FROM blocked.transactionid
    AND blocking.pid != blocked.pid
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking.pid
WHERE NOT blocked.granted;

-- Replication lag:
SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn,
       pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replay_lag_bytes
FROM pg_stat_replication;

-- Long-running transactions (MVCC bloat source):
SELECT pid, now() - xact_start AS xact_duration, state, query
FROM pg_stat_activity
WHERE xact_start IS NOT NULL
ORDER BY xact_duration DESC LIMIT 10;
```

---

## 11. Row-Level Security (RLS)

```sql
-- Multi-tenant data isolation without application-level filtering

CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    title TEXT,
    body TEXT
);

-- Enable RLS:
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

-- Policy: each tenant sees only their own data
CREATE POLICY tenant_isolation ON documents
    USING (tenant_id = current_setting('app.tenant_id')::INTEGER);

-- Set tenant context per connection:
SET app.tenant_id = '42';
SELECT * FROM documents;  -- only sees tenant 42's rows!

-- Policy for INSERT:
CREATE POLICY tenant_insert ON documents
    FOR INSERT
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::INTEGER);

-- Table owners BYPASS RLS by default. Force it:
ALTER TABLE documents FORCE ROW LEVEL SECURITY;

-- Multiple policies combine with OR (same command) or AND (different commands)
-- Be careful: no policy = no rows visible (deny by default when RLS is on)
```

---

## 12. Listen/Notify — Built-in Pub/Sub

```sql
-- Lightweight pub/sub without polling

-- Session 1 (listener):
LISTEN new_orders;
-- Blocks until notification arrives or timeout

-- Session 2 (notifier):
NOTIFY new_orders, '{"order_id": 12345, "total": 99.95}';
-- Or from a trigger:
CREATE OR REPLACE FUNCTION notify_new_order() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('new_orders', json_build_object('id', NEW.id, 'total', NEW.total)::TEXT);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_notify_order AFTER INSERT ON orders
FOR EACH ROW EXECUTE FUNCTION notify_new_order();

-- Session 1 receives:
-- Asynchronous notification "new_orders" with payload "{"order_id": 12345, "total": 99.95}"

-- Great for:
-- Real-time cache invalidation
-- Job queue notifications (combine with SKIP LOCKED)
-- Live dashboards

-- Limitations:
-- Payload max 8000 bytes
-- If no listener is connected, notification is lost
-- Not persistent (unlike Kafka)
-- PgBouncer in transaction mode: LISTEN doesn't work (need session mode)
```

---

## 13. Foreign Data Wrappers (FDW)

```sql
-- Query external data sources as if they were PostgreSQL tables

-- postgres_fdw: connect to another PostgreSQL server
CREATE EXTENSION postgres_fdw;

CREATE SERVER remote_server
FOREIGN DATA WRAPPER postgres_fdw
OPTIONS (host 'remote-host', dbname 'remote_db', port '5432');

CREATE USER MAPPING FOR current_user
SERVER remote_server
OPTIONS (user 'remote_user', password 'secret');

CREATE FOREIGN TABLE remote_orders (
    id INTEGER,
    customer_id INTEGER,
    total DECIMAL(10,2)
) SERVER remote_server OPTIONS (table_name 'orders');

-- Now query it like a local table:
SELECT * FROM remote_orders WHERE customer_id = 42;
-- PostgreSQL pushes predicates to the remote server

-- Other FDWs:
-- mysql_fdw:     query MySQL from PostgreSQL
-- oracle_fdw:    query Oracle
-- file_fdw:      query CSV/text files as tables
-- redis_fdw:     query Redis keys
-- multicorn:     write FDWs in Python
```

---

## 14. Advisory Locks — Application-Level Coordination

```sql
-- Session-level advisory locks (held until session ends or explicit unlock):
SELECT pg_advisory_lock(42);          -- blocking acquire
SELECT pg_try_advisory_lock(42);      -- non-blocking (returns true/false)
SELECT pg_advisory_unlock(42);        -- explicit release

-- Transaction-level advisory locks (released at COMMIT/ROLLBACK):
SELECT pg_advisory_xact_lock(42);     -- auto-released at end of transaction

-- Two-key advisory locks (for more specificity):
SELECT pg_advisory_lock(schema_id, record_id);

-- Use cases:

-- 1. Prevent duplicate cron jobs:
SELECT pg_try_advisory_lock(hashtext('daily_report'));
-- Returns false if another instance is already running

-- 2. Application-level mutex on a resource:
BEGIN;
SELECT pg_advisory_xact_lock(hashtext('user:' || user_id::TEXT));
-- Do critical section work
COMMIT;  -- lock auto-released

-- 3. Rate limiting:
SELECT pg_try_advisory_lock(hashtext('api:' || client_ip));
-- If false, rate limit exceeded

-- Check held locks:
SELECT * FROM pg_locks WHERE locktype = 'advisory';
```

---

## Key Takeaways

1. **Process-per-connection** model → connection pooling is mandatory (PgBouncer).
2. **VACUUM is not optional.** Autovacuum must be tuned per table. Monitor dead tuple ratio and XID age.
3. **Transaction ID wraparound** can freeze your database. Monitor `age(datfrozenxid)`.
4. **HOT updates** are free performance — set `fillfactor < 100` on write-heavy tables.
5. **pg_stat_statements** is the #1 tool for finding slow queries. Install it on every server.
6. **Extensions** are PostgreSQL's superpower: PostGIS, pgvector, TimescaleDB, Citus, pg_trgm.
7. **Logical replication** enables table-level selective replication and zero-downtime major version upgrades.
8. **RLS** provides true multi-tenant isolation at the database level.
9. **LISTEN/NOTIFY + FOR UPDATE SKIP LOCKED** = a job queue without Kafka.
10. **random_page_cost = 1.1** for SSD. The default (4.0) is for spinning disks and will cause bad plans.

---

Next: [02-mysql-mariadb.md](02-mysql-mariadb.md) →
