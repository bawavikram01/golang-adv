# 3.4 — Microsoft SQL Server

> SQL Server is the backbone of the Microsoft enterprise stack.  
> If you work in a .NET/Windows shop, this is your world.  
> Its unique features — columnstore, In-Memory OLTP, Query Store — are world-class.

---

## 1. Architecture

```
SQL Server uses a single-process, multi-thread architecture.

┌────────────────────────────────────────────────┐
│             SQL Server Process (sqlservr)        │
│                                                  │
│  ┌──────────────────────────────────────────┐   │
│  │         Memory (Buffer Pool)              │   │
│  │  ┌──────────┐ ┌────────────┐ ┌────────┐ │   │
│  │  │Data Pages│ │Plan Cache  │ │Log Cache│ │   │
│  │  │(8KB pages)│ │(compiled   │ │         │ │   │
│  │  │          │ │ query plans)│ │         │ │   │
│  │  └──────────┘ └────────────┘ └────────┘ │   │
│  │  ┌──────────────┐ ┌───────────────────┐  │   │
│  │  │Lock Manager   │ │Memory Grants      │  │   │
│  │  │               │ │(sort/hash/query)  │  │   │
│  │  └──────────────┘ └───────────────────┘  │   │
│  └──────────────────────────────────────────┘   │
│                                                  │
│  ┌──────────────────────┐                       │
│  │ SQLOS (SQL Server OS) │  ← Cooperative scheduler │
│  │ - Task scheduling     │     Not preemptive!      │
│  │ - Memory management   │     Threads yield voluntarily │
│  │ - I/O                 │                       │
│  │ - CPU scheduling      │                       │
│  └──────────────────────┘                       │
│                                                  │
│  Worker threads: max_worker_threads (default 0 = auto) │
│  On 64-bit: 512 workers for ≤4 CPUs, more for larger   │
│                                                  │
│  Databases:                                      │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────┐ │
│  │master  │ │msdb    │ │tempdb  │ │User DB   │ │
│  │(system)│ │(jobs)  │ │(temp)  │ │(.mdf/.ldf│ │
│  └────────┘ └────────┘ └────────┘ └──────────┘ │
└────────────────────────────────────────────────┘

Key system databases:
  master:  server-level metadata, logins, linked servers
  msdb:    SQL Agent jobs, backup history, alerts
  tempdb:  temp tables, sort spills, version store (created fresh on restart)
  model:   template for new databases
```

### SQLOS — SQL Server Operating System

```
SQL Server has its own OS layer (SQLOS) that sits above Windows/Linux:

- Non-preemptive scheduling: threads YIELD voluntarily
  (unlike OS threads which are preempted by scheduler)
- Memory management: SQL Server manages its own memory, bypassing OS caching
- I/O: asynchronous I/O subsystem

Why?
  - Better control over thread scheduling for database workloads
  - Avoids context switch overhead of OS scheduling
  - Can detect and handle blocking/waiting more efficiently

  Wait stats: SQLOS tracks what threads are WAITING for
  → This is the primary performance tuning tool in SQL Server

SELECT wait_type, waiting_tasks_count, wait_time_ms,
       signal_wait_time_ms
FROM sys.dm_os_wait_stats
WHERE waiting_tasks_count > 0
ORDER BY wait_time_ms DESC;

-- Top waits tell you the bottleneck:
-- PAGEIOLATCH_*:    disk I/O too slow (need more RAM or faster disk)
-- CXPACKET/CXCONSUMER: parallelism waits (usually benign)
-- LCK_M_*:          blocking/locking issues
-- WRITELOG:         transaction log write bottleneck
-- SOS_SCHEDULER_YIELD: CPU pressure
-- ASYNC_NETWORK_IO: clients consuming results slowly
```

---

## 2. Storage Architecture

```
Database files:
  .mdf:  Primary data file (one per database)
  .ndf:  Secondary data files (optional, for spreading I/O)
  .ldf:  Transaction log file

Pages and Extents:
  Page = 8 KB (fixed, cannot change)
  Extent = 8 contiguous pages = 64 KB
  
  Mixed extents: pages from different objects (small tables)
  Uniform extents: all 8 pages from same object (large tables)

Page types:
  Data pages, Index pages, IAM (Index Allocation Map),
  PFS (Page Free Space), GAM/SGAM (extent allocation),
  Text/Image pages (LOB data)

Row storage:
  Row = header (4 bytes) + fixed-length columns + null bitmap + 
        variable-length offset array + variable-length columns
  Max row size: 8060 bytes (rest goes to row overflow pages)
```

### TempDB — The Shared Scratch Space

```sql
-- TempDB is recreated on every SQL Server restart.
-- All users share ONE tempdb.

-- Used for:
-- 1. Temp tables (#local, ##global)
-- 2. Table variables (@table)
-- 3. Sort spills (when memory grant too small)
-- 4. Hash spills
-- 5. Row versioning (RCSI, snapshot isolation)
-- 6. Online index rebuilds
-- 7. Cursors, internal worktables

-- TempDB contention is a common problem:
-- Multiple data files (one per CPU core, up to 8) reduce PFS/GAM contention
ALTER DATABASE tempdb ADD FILE (NAME = 'tempdev2', FILENAME = '...');

-- Check tempdb usage:
SELECT * FROM sys.dm_db_file_space_usage;
```

---

## 3. Columnstore Indexes — SQL Server's Analytics Weapon

```
SQL Server's columnstore indexes: column-oriented storage for analytics.

Traditional (rowstore):      Columnstore:
┌────┬──────┬────────┐      ┌────┐ ┌──────┐ ┌────────┐
│ ID │ Name │ Salary │      │ ID │ │ Name │ │ Salary │
├────┼──────┼────────┤      ├────┤ ├──────┤ ├────────┤
│ 1  │ Alice│ 70000  │      │ 1  │ │ Alice│ │ 70000  │
│ 2  │ Bob  │ 80000  │      │ 2  │ │ Bob  │ │ 80000  │
│ 3  │ Carol│ 90000  │      │ 3  │ │ Carol│ │ 90000  │
└────┴──────┴────────┘      └────┘ └──────┘ └────────┘
                            Each column stored separately
                            + compressed heavily

-- Clustered columnstore (replaces heap/B-tree):
CREATE CLUSTERED COLUMNSTORE INDEX cci_sales ON sales;
-- Entire table stored in column format
-- Best for: fact tables, analytics, data warehouse

-- Nonclustered columnstore (secondary index):
CREATE NONCLUSTERED COLUMNSTORE INDEX ncci_sales ON sales (amount, region, date);
-- Keep rowstore for OLTP, add columnstore for analytics
-- "Real-time operational analytics" — HTAP on one engine!

Architecture:
  Row Group: ~1 million rows (compressed together)
  Column Segment: one column within a row group (compressed, stored as LOB)
  Delta Store: rowstore buffer for recent inserts (batch-compressed later)
  
  Compression: dictionary encoding, RLE, bit-packing
  Typical compression ratio: 5-10x

-- Columnstore + batch mode execution:
-- Processes ~1000 rows at a time (not row-by-row)
-- SIMD-optimized operators
-- 10-100x faster for aggregation queries

-- Check columnstore state:
SELECT * FROM sys.column_store_row_groups;
-- state: OPEN (delta), CLOSED (pending compression), COMPRESSED, TOMBSTONE
```

---

## 4. In-Memory OLTP (Hekaton)

```sql
-- Memory-optimized tables: rows NEVER touch disk (except for durability).
-- Lock-free, latch-free architecture using MVCC + optimistic concurrency.

-- Create a memory-optimized filegroup:
ALTER DATABASE mydb ADD FILEGROUP mem_fg CONTAINS MEMORY_OPTIMIZED_DATA;
ALTER DATABASE mydb ADD FILE (NAME = 'mem_file', FILENAME = '...')
TO FILEGROUP mem_fg;

-- Memory-optimized table:
CREATE TABLE hot_orders (
    id INT NOT NULL PRIMARY KEY NONCLUSTERED HASH WITH (BUCKET_COUNT = 1000000),
    customer_id INT NOT NULL,
    amount DECIMAL(10,2),
    INDEX ix_cust HASH (customer_id) WITH (BUCKET_COUNT = 100000)
) WITH (MEMORY_OPTIMIZED = ON, DURABILITY = SCHEMA_AND_DATA);

-- DURABILITY options:
-- SCHEMA_AND_DATA: persisted (survives restart) — logs to checkpoint files
-- SCHEMA_ONLY: temp table behavior (lost on restart, no I/O!)

-- Natively compiled stored procedures (C code, not interpreted):
CREATE PROCEDURE insert_order
    @id INT, @cust INT, @amount DECIMAL(10,2)
WITH NATIVE_COMPILATION, SCHEMABINDING
AS BEGIN ATOMIC WITH (TRANSACTION ISOLATION LEVEL = SNAPSHOT, LANGUAGE = N'English')
    INSERT INTO hot_orders VALUES (@id, @cust, @amount);
END;

-- Performance gains: 10-30x for OLTP workloads
-- No locks, no latches → extreme concurrency
-- Limitations: no ALTER TABLE, limited data types, max 8060-byte rows
```

---

## 5. Query Store — Built-in Performance Time Machine

```sql
-- Query Store captures query plans and runtime stats OVER TIME.
-- Think of it as pg_stat_statements + plan history + plan forcing.

-- Enable (on by default in Azure SQL):
ALTER DATABASE mydb SET QUERY_STORE = ON;
ALTER DATABASE mydb SET QUERY_STORE (
    OPERATION_MODE = READ_WRITE,
    MAX_STORAGE_SIZE_MB = 1024,
    INTERVAL_LENGTH_MINUTES = 60,  -- aggregation interval
    CLEANUP_POLICY = (STALE_QUERY_THRESHOLD_DAYS = 30)
);

-- Find regressed queries (plan changed → got slower):
SELECT TOP 20 
    qsq.query_id, qsp.plan_id,
    qsrs.avg_duration / 1000 AS avg_ms,
    qsrs.count_executions,
    qst.query_sql_text
FROM sys.query_store_query qsq
JOIN sys.query_store_plan qsp ON qsq.query_id = qsp.query_id
JOIN sys.query_store_runtime_stats qsrs ON qsp.plan_id = qsrs.plan_id
JOIN sys.query_store_query_text qst ON qsq.query_text_id = qst.query_text_id
ORDER BY qsrs.avg_duration DESC;

-- FORCE a known-good plan:
EXEC sp_query_store_force_plan @query_id = 42, @plan_id = 7;
-- SQL Server will ALWAYS use plan 7 for query 42, regardless of optimizer

-- UNFORCE:
EXEC sp_query_store_unforce_plan @query_id = 42, @plan_id = 7;

-- Why this is revolutionary:
-- Plan regression is one of the hardest database problems.
-- Query Store lets you detect it AND fix it without changing code.
```

---

## 6. Always On Availability Groups

```
SQL Server's HA/DR solution (replacement for older database mirroring):

Primary Replica → synchronous/async → Secondary Replicas

┌──────────┐     ┌──────────┐     ┌──────────┐
│Primary   │────→│Secondary │────→│Secondary │
│(R/W)     │sync │(read-only│async│(DR site) │
│          │     │reporting)│     │          │
└──────────┘     └──────────┘     └──────────┘
      │
  Listener (virtual network name / IP)
  → Clients connect to listener, automatic failover

Availability Group:
  - Group of databases that fail over together
  - Up to 8 secondary replicas (SQL Server 2019)
  - Sync or async data movement
  - Automatic or manual failover
  - Read-only routing to secondaries
  - Seeding: automatic or manual (backup/restore)

Distributed Availability Groups:
  AG1 (datacenter 1) <→ AG2 (datacenter 2)
  Cross-datacenter DR without Windows clustering

Basic Availability Groups (Standard Edition):
  - Only 1 secondary, only 1 database per AG
  - No read-only secondaries
```

---

## 7. T-SQL Unique Features

```sql
-- OUTPUT clause (like PostgreSQL's RETURNING, but more powerful):
DELETE FROM orders
OUTPUT DELETED.id, DELETED.total INTO @deleted_orders
WHERE status = 'expired';

-- MERGE (the most powerful MERGE in any database):
MERGE INTO target AS t
USING source AS s ON t.id = s.id
WHEN MATCHED AND s.active = 0 THEN DELETE
WHEN MATCHED THEN UPDATE SET t.name = s.name
WHEN NOT MATCHED BY TARGET THEN INSERT (id, name) VALUES (s.id, s.name)
WHEN NOT MATCHED BY SOURCE THEN DELETE
OUTPUT $action, INSERTED.*, DELETED.*;

-- TRY...CATCH (structured error handling):
BEGIN TRY
    BEGIN TRANSACTION;
    INSERT INTO orders VALUES (...);
    UPDATE inventory SET qty = qty - 1 WHERE product_id = @pid;
    COMMIT;
END TRY
BEGIN CATCH
    ROLLBACK;
    THROW;  -- re-raise the error
END CATCH;

-- STRING_AGG (like PostgreSQL's string_agg):
SELECT department_id, STRING_AGG(name, ', ') AS names
FROM employees GROUP BY department_id;

-- CROSS APPLY / OUTER APPLY (like PostgreSQL's LATERAL JOIN):
SELECT c.name, o.order_date, o.total
FROM customers c
CROSS APPLY (
    SELECT TOP 3 * FROM orders 
    WHERE customer_id = c.id 
    ORDER BY order_date DESC
) o;

-- Temporal tables (system-versioned, similar to MariaDB):
CREATE TABLE employees (
    id INT PRIMARY KEY,
    name NVARCHAR(100),
    salary DECIMAL(10,2),
    valid_from DATETIME2 GENERATED ALWAYS AS ROW START,
    valid_to DATETIME2 GENERATED ALWAYS AS ROW END,
    PERIOD FOR SYSTEM_TIME (valid_from, valid_to)
) WITH (SYSTEM_VERSIONING = ON (HISTORY_TABLE = dbo.employees_history));

SELECT * FROM employees FOR SYSTEM_TIME AS OF '2024-01-01';

-- JSON support:
SELECT id, JSON_VALUE(data, '$.name') AS name
FROM documents
WHERE ISJSON(data) = 1;

-- OPENJSON (parse JSON into rows):
SELECT * FROM OPENJSON(@json_data)
WITH (id INT, name NVARCHAR(100), email NVARCHAR(200));
```

---

## 8. SQL Server vs PostgreSQL

```
Feature                 SQL Server                      PostgreSQL
────────────────────────────────────────────────────────────────────
Cost                    Express free, Standard $$$       Free
OS                      Windows, Linux (2017+)           All platforms
Columnstore             Built-in, mature, batch mode     No native columnstore
In-Memory OLTP          Hekaton (lock-free)              No equivalent
Query Store             Built-in plan management         pg_stat_statements (less)
Temporal tables         Built-in system versioning       No native support
Plan forcing            Query Store + plan guides        pg_hint_plan (extension)
Error handling          TRY...CATCH                      BEGIN...EXCEPTION
Scheduler               SQLOS (cooperative)              OS scheduler
Wait stats              dm_os_wait_stats (excellent)     pg_stat_activity (less)
Tooling                 SSMS (excellent GUI)             pgAdmin (decent)
Replication             Always On AG                     Streaming + Logical
Partitioning            Table + index partitioning       Declarative (PG 10+)
Full-text search        Full-Text Search service         tsvector (built-in)
XML support             Extensive (XML type, XQuery)     Basic
JSON                    JSON functions (improving)        JSONB (superior)
Extensions              CLR integration (.NET)           Rich extension ecosystem
Parallel query          Mature                           Good (PG 9.6+)
Connection model        Thread pool                      Process per connection
MVCC cleanup            Version store in tempdb         VACUUM
```

---

## Key Takeaways

1. **Columnstore indexes** give SQL Server world-class analytics performance — batch mode execution + compression.
2. **In-Memory OLTP (Hekaton)** is a completely lock-free, latch-free engine for extreme OLTP.
3. **Query Store** is the best built-in plan management tool in any database. Force known-good plans.
4. **Wait statistics** (`sys.dm_os_wait_stats`) are the #1 tuning tool — they tell you what SQL Server is waiting for.
5. **Always On Availability Groups** are SQL Server's HA standard. The listener provides transparent failover.
6. **SQLOS cooperative scheduling** means SQL Server manages its own threads — different from PostgreSQL/MySQL.
7. **TempDB** is shared — a single bad query's spill can affect everyone. Size and place it carefully.
8. **CROSS APPLY** is SQL Server's answer to LATERAL JOIN — extremely useful for "top-N per group" queries.
9. **Temporal tables** provide built-in time travel — no trigger-based audit table needed.

---

Next: [05-sqlite.md](05-sqlite.md) →
