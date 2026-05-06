# 3.2 — MySQL & MariaDB Deep Dive

> MySQL is the world's most popular open-source database.  
> MariaDB is its community fork, maintained by the original MySQL creator.  
> Understanding InnoDB internals is what separates a MySQL DBA from a MySQL user.

---

## 1. Architecture — MySQL vs PostgreSQL

```
MySQL has a PLUGGABLE STORAGE ENGINE architecture:

    Client
      │
      ▼
┌─────────────────────────────────────┐
│         MySQL Server Layer           │
│  ┌──────────┐  ┌──────────────────┐ │
│  │ Parser   │  │ Optimizer        │ │
│  │          │→ │ (cost-based)     │ │
│  └──────────┘  └──────────────────┘ │
│  ┌──────────────────────────────────┐│
│  │ Executor                         ││
│  └──────────┬───────────────────────┘│
└─────────────┼────────────────────────┘
              │ Handler API (abstract storage interface)
   ┌──────────┼──────────┐
   ▼          ▼          ▼
┌────────┐ ┌────────┐ ┌────────┐
│ InnoDB │ │MyISAM  │ │Memory  │   ← Pluggable storage engines
│(default)│ │(legacy)│ │(temp)  │
└────────┘ └────────┘ └────────┘

Key difference from PostgreSQL:
  PostgreSQL: monolithic — one storage engine tightly integrated
  MySQL: layered — server layer + storage engine API

  Implications:
  - MySQL can't do covering index scans as efficiently (double lookup)
  - MySQL's optimizer has less visibility into storage engine statistics
  - But you can swap engines per table (rarely done in practice)
```

### Thread Model

```
MySQL uses ONE THREAD PER CONNECTION (not process like PostgreSQL).

Advantages over PostgreSQL's process model:
  - Lower memory per connection (~256KB thread stack vs ~5-10MB process)
  - Faster connection creation (thread vs fork)
  - Can handle more connections natively

Disadvantages:
  - A thread crash can bring down the server
  - Thread synchronization complexity (mutexes, latches)
  - Still benefits from connection pooling (ProxySQL, MySQL Router)
```

---

## 2. InnoDB Architecture

```
┌────────────────────────────────────────────┐
│              InnoDB In-Memory               │
│                                            │
│  ┌──────────────────────────────────────┐  │
│  │       Buffer Pool (70-80% of RAM)    │  │
│  │  ┌─────────┐ ┌──────────┐ ┌───────┐ │  │
│  │  │Data Pages│ │Index Pages│ │Undo   │ │  │
│  │  │         │ │          │ │Pages  │ │  │
│  │  └─────────┘ └──────────┘ └───────┘ │  │
│  │  ┌──────────────┐ ┌─────────────┐   │  │
│  │  │Adaptive Hash  │ │Change Buffer│   │  │
│  │  │Index (AHI)    │ │(for sec idx)│   │  │
│  │  └──────────────┘ └─────────────┘   │  │
│  └──────────────────────────────────────┘  │
│                                            │
│  ┌─────────────┐  ┌────────────────────┐   │
│  │ Log Buffer   │  │ Data Dictionary    │   │
│  │ (redo log)   │  │ Cache              │   │
│  └─────────────┘  └────────────────────┘   │
└────────────────────────────────────────────┘
                    │
                    ▼ Disk
    ┌──────────┬──────────┬──────────┬──────────┐
    │.ibd files│ Redo Log │Undo Tblsp│Doublewrite│
    │(tablespce)│(ib_logN) │          │Buffer    │
    │          │          │          │(.dblwr)  │
    └──────────┘──────────┘──────────┘──────────┘
```

### Buffer Pool Deep Dive

```sql
-- InnoDB's buffer pool: THE single most important memory structure

-- Size: 70-80% of available RAM
innodb_buffer_pool_size = 48G  -- on a 64GB machine

-- Multiple buffer pool instances (reduce contention):
innodb_buffer_pool_instances = 8  -- default: 8 if pool > 1GB

-- Page replacement: LRU with midpoint insertion
-- New pages enter at 5/8 from the head (not the head!)
-- Prevents a full table scan from flushing the entire cache
innodb_old_blocks_pct = 37     -- old sublist = 37% of pool (default)
innodb_old_blocks_time = 1000  -- ms before old→young promotion

-- Monitor buffer pool:
SHOW ENGINE INNODB STATUS\G
-- Look for:
-- Buffer pool hit rate: should be > 99.9%
-- Pages read/created/written

SELECT * FROM information_schema.INNODB_BUFFER_POOL_STATS;
```

### Change Buffer (Insert Buffer)

```
When a secondary index page is NOT in the buffer pool:
  Instead of reading the page from disk just to add an index entry,
  InnoDB caches the change in the CHANGE BUFFER.

  Later, when the page is read for another reason, the buffered
  changes are merged ("merge on read").

  Works for: INSERT, DELETE-marking, purge of secondary indexes
  Does NOT work for: unique indexes (must check uniqueness immediately)
  
  innodb_change_buffer_max_size = 25  -- max % of buffer pool (default 25)
  innodb_change_buffering = all       -- inserts, deletes, purges, changes

  Reduces random I/O for write-heavy workloads with many secondary indexes.
```

---

## 3. InnoDB Locking — The Complete Picture

```
InnoDB uses NEXT-KEY LOCKING (record lock + gap lock) for REPEATABLE READ.
This prevents phantom reads WITHOUT using MVCC snapshots (unlike PostgreSQL).

Lock Types:
  1. Record Lock:  locks a single index record
  2. Gap Lock:     locks the gap BETWEEN two index records
  3. Next-Key Lock: record lock + gap lock on the preceding gap
  4. Insert Intention Lock: special gap lock for INSERT

                Gap   Record   Gap   Record   Gap   Record   Gap
              |......|===R1===|......|===R2===|......|===R3===|......|
              ←gap→  ←rec→    ←gap→  ←rec→    ←gap→  ←rec→    ←gap→
              
  Next-Key Lock on R2 = gap before R2 + record lock on R2
```

### Locking Behavior by Isolation Level

```sql
-- READ COMMITTED: only record locks, no gap locks
-- → Allows phantom reads (new rows appear in range)
-- → Fewer deadlocks, better concurrency

-- REPEATABLE READ (default): next-key locks
-- → Prevents phantoms for locking reads
-- → More deadlocks possible

-- Example: demonstrating next-key locking
CREATE TABLE t (id INT PRIMARY KEY, val INT, INDEX idx_val(val));
INSERT INTO t VALUES (1, 10), (2, 20), (3, 30);

-- Session 1 (REPEATABLE READ):
BEGIN;
SELECT * FROM t WHERE val = 20 FOR UPDATE;
-- Locks:
--   Next-key lock on (10, 20]  (gap from 10 to 20, inclusive of 20)
--   Gap lock on (20, 30)       (gap after 20 to next record)
-- This prevents ANY insert with val between 10 and 30!

-- Session 2:
INSERT INTO t VALUES (4, 15);  -- BLOCKS! (in the gap 10..20)
INSERT INTO t VALUES (5, 25);  -- BLOCKS! (in the gap 20..30)
INSERT INTO t VALUES (6, 5);   -- OK! (outside locked range)
INSERT INTO t VALUES (7, 35);  -- OK! (outside locked range)
```

### Deadlock Detection

```sql
-- InnoDB checks for deadlocks using a wait-for graph.
-- When detected: the transaction with fewest undo log records is rolled back.

-- View last deadlock:
SHOW ENGINE INNODB STATUS\G
-- Search for "LATEST DETECTED DEADLOCK"

-- Common deadlock pattern:
-- Session 1: UPDATE t SET x=1 WHERE id=1; UPDATE t SET x=1 WHERE id=2;
-- Session 2: UPDATE t SET x=1 WHERE id=2; UPDATE t SET x=1 WHERE id=1;
-- → Always lock rows in consistent order to avoid deadlocks.

-- InnoDB deadlock detection can be expensive at high concurrency:
innodb_deadlock_detect = OFF  -- disable if using short lock waits
innodb_lock_wait_timeout = 3  -- seconds before aborting (default 50)
-- With detection OFF + short timeout: deadlocks resolve via timeout
```

---

## 4. InnoDB Clustered Index

```
InnoDB stores data IN the primary key B+ tree (clustered index).

Leaf nodes of primary key = actual row data
This is fundamentally different from PostgreSQL (heap + separate index).

                    [20, 40, 60]
                   /      |       \
            [5,10,15]  [25,30,35]  [45,50,55]
              │          │           │
              ▼          ▼           ▼
          Row data    Row data    Row data
          for 5,10,   for 25,30,  for 45,50,
          15          35          55

Consequences:
  1. PK lookups are ONE B-tree traversal (very fast)
  2. Range scans on PK are sequential I/O (rows physically ordered)
  3. Secondary index leaf nodes store the PK value (not row pointer)
  4. Secondary index lookup = 2 B-tree traversals (index → PK → row)
     This is called a "double lookup" or "bookmark lookup"
  
  5. Random PKs (UUIDs) cause massive page splits and fragmentation
     → Use auto-increment or ULID/UUIDv7 (time-ordered) for InnoDB!

  6. Row order = PK order → changing PK rebuilds entire table
  7. Tables without explicit PK: InnoDB adds a hidden 6-byte row ID

  -- Check secondary index double lookup cost:
  EXPLAIN SELECT * FROM orders WHERE customer_email = 'alice@example.com';
  -- type: ref (secondary index scan)
  -- Internally: idx_email → find PK → PK tree → row data
```

---

## 5. Replication in MySQL

### Binary Log (binlog) — The Foundation

```
MySQL replication is based on the BINARY LOG (binlog), not WAL.

Binlog formats:
  STATEMENT: logs SQL statements (compact, but non-deterministic functions break)
  ROW: logs actual row changes (safe, but larger)
  MIXED: uses STATEMENT by default, switches to ROW when needed

binlog_format = ROW  -- recommended for safety

The binlog serves dual purpose:
  1. Replication (send to replicas)
  2. Point-in-time recovery (replay binlog after backup restore)
```

### Traditional Replication (binlog position)

```
Source → writes binlog → IO thread on Replica reads binlog → 
    writes relay log → SQL thread replays relay log

Source                      Replica
┌──────────┐               ┌──────────┐
│ binlog   │──IO thread───→│relay log │
│          │               │    │     │
│          │               │SQL thread│
│          │               │    ↓     │
│          │               │ apply    │
└──────────┘               └──────────┘

Problems:
  - Must track binlog file + position manually
  - Failover is complex (which position on new source?)
  
# Source:
server-id = 1
log-bin = mysql-bin

# Replica:
server-id = 2
CHANGE REPLICATION SOURCE TO
    SOURCE_HOST = 'source-host',
    SOURCE_USER = 'repl',
    SOURCE_PASSWORD = 'secret',
    SOURCE_LOG_FILE = 'mysql-bin.000003',
    SOURCE_LOG_POS = 154;
START REPLICA;
```

### GTID Replication (Modern)

```
GTID (Global Transaction ID) = source_uuid:transaction_number

Every transaction gets a unique GTID across the cluster.
Replicas know EXACTLY which transactions they've applied.
Failover: just point replica to new source — GTIDs handle positioning.

gtid_mode = ON
enforce_gtid_consistency = ON

# Replica auto-positions:
CHANGE REPLICATION SOURCE TO
    SOURCE_HOST = 'source-host',
    SOURCE_AUTO_POSITION = 1;

-- Check GTID status:
SELECT @@gtid_executed;
-- e.g., "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-100"
```

### Group Replication (InnoDB Cluster)

```
MySQL Group Replication: synchronous multi-source replication using Paxos.

  ┌──────┐  ┌──────┐  ┌──────┐
  │Node 1│──│Node 2│──│Node 3│   Paxos consensus group
  │(R/W) │  │(R/W) │  │(R/W) │   All nodes can accept writes
  └──────┘  └──────┘  └──────┘   Conflicts detected automatically

Modes:
  Single-Primary: one R/W, rest read-only (simpler, recommended)
  Multi-Primary: all R/W (complex, conflict resolution needed)

InnoDB Cluster = Group Replication + MySQL Router + MySQL Shell
  MySQL Router: connection routing + automatic failover
  MySQL Shell: cluster management CLI
```

---

## 6. MySQL Query Optimization

```sql
-- EXPLAIN output (MySQL is different from PostgreSQL):
EXPLAIN SELECT * FROM orders WHERE customer_id = 42;
-- id | select_type | table  | type | possible_keys | key    | rows | Extra
-- 1  | SIMPLE      | orders | ref  | idx_customer  | idx_cust| 50  | NULL

-- EXPLAIN ANALYZE (MySQL 8.0.18+):
EXPLAIN ANALYZE SELECT * FROM orders WHERE status = 'pending' ORDER BY created_at;
-- Returns actual execution times (like PG's EXPLAIN ANALYZE)

-- Key 'type' values (best to worst):
-- system:   table has 1 row
-- const:    at most 1 row (PK/unique lookup)
-- eq_ref:   one row per join key (PK/unique join)
-- ref:      multiple rows per key (non-unique index)
-- range:    index range scan
-- index:    full index scan (reads every entry in the index)
-- ALL:      full table scan

-- Optimizer hints (MySQL 8.0+):
SELECT /*+ NO_INDEX(orders idx_status) */ * FROM orders;
SELECT /*+ JOIN_ORDER(customers, orders) */ ...;
SELECT /*+ SET_VAR(optimizer_switch='mrr=on') */ ...;

-- Index hints (older syntax):
SELECT * FROM orders USE INDEX (idx_status) WHERE status = 'pending';
SELECT * FROM orders FORCE INDEX (idx_date) WHERE created_at > '2024-01-01';
SELECT * FROM orders IGNORE INDEX (idx_status) WHERE status = 'pending';

-- Generated columns + virtual indexes:
ALTER TABLE users ADD COLUMN email_domain VARCHAR(255) 
    GENERATED ALWAYS AS (SUBSTRING_INDEX(email, '@', -1)) VIRTUAL;
CREATE INDEX idx_domain ON users(email_domain);
-- No extra storage — computed at read time, but indexable!
```

---

## 7. MySQL vs PostgreSQL Comparison

```
Feature                  MySQL (InnoDB)                 PostgreSQL
─────────────────────────────────────────────────────────────────
Storage                  Clustered index (PK = data)    Heap + separate indexes
MVCC                     Undo log (rollback segments)   Tuple versioning in heap
Read consistency         Undo to reconstruct old ver.   Read old tuple directly
Bloat from updates       Undo log grows, then purged    Dead tuples in heap
Cleanup                  Purge thread (automatic)       VACUUM (must be tuned!)
Connection model         Thread per connection           Process per connection
Default isolation        REPEATABLE READ                READ COMMITTED
Phantom prevention       Gap/next-key locking           SSI (serializable only)
Replication basis        Binary log (logical)           WAL (physical or logical)
Extensions               Very limited                   Extremely rich ecosystem
JSON support             JSON type, ->> operator        JSONB (binary, indexable)
Full-text search         Basic (InnoDB FTS)             Advanced (tsvector/tsquery)
Partitioning             Range, List, Hash, Key         Range, List, Hash
Window functions         Yes (MySQL 8.0+)               Yes (since forever)
CTEs                     Yes (MySQL 8.0+)               Yes (with recursive)
Parallel query           Limited (MySQL 8.0.14+)        Mature (parallel scans)
Geospatial               Basic spatial indexes          PostGIS (industry standard)
Type system              Less strict (modes matter)     Strict, extensible
Custom types             No                             Yes (CREATE TYPE)
Check constraints        Enforced (MySQL 8.0.16+)       Always enforced
```

---

## 8. MariaDB Differences

```
MariaDB forked from MySQL 5.5 in 2009 (by MySQL creator Monty Widenius).

Compatible but diverging:
  - Shares InnoDB (called "InnoDB" but may be slightly different version)
  - Has Aria engine (crash-safe MyISAM replacement)
  - ColumnStore engine (columnar for analytics)
  - Spider engine (built-in sharding)
  - Galera Cluster (synchronous multi-master, mature)
  - More optimizer features: subquery optimization, derived table merge
  - Sequences (CREATE SEQUENCE — PostgreSQL-like)
  - System-versioned tables (temporal tables, SQL:2011 standard)
  - Oracle PL/SQL compatibility mode
  - No MySQL Router/Shell — uses MaxScale for proxy/routing

MariaDB system-versioned tables:
  CREATE TABLE employees (
      id INT PRIMARY KEY,
      name VARCHAR(100),
      salary DECIMAL(10,2)
  ) WITH SYSTEM VERSIONING;

  -- Query historical data:
  SELECT * FROM employees FOR SYSTEM_TIME AS OF '2024-01-01';
  SELECT * FROM employees FOR SYSTEM_TIME BETWEEN '2024-01-01' AND '2024-06-01';

Galera Cluster (MariaDB):
  Truly synchronous multi-master replication
  All nodes are R/W
  Certification-based conflict detection
  Limitations: no LOCK TABLES, no XA transactions, InnoDB only
  Trade-off: higher write latency (all nodes certify), but zero replication lag
```

---

## Key Takeaways

1. **InnoDB clustered index** fundamentally changes how data is stored — PK order IS row order. Use auto-increment or time-ordered IDs, never random UUIDs.
2. **Next-key locking** in InnoDB prevents phantoms but causes more deadlocks than PostgreSQL's MVCC approach.
3. **Buffer pool = 70-80% of RAM.** This is the single most important InnoDB tuning parameter.
4. **GTID replication** should be standard. Stop using binlog file+position.
5. **`binlog_format = ROW`** is the only safe option for replication.
6. **MySQL's double lookup** for secondary indexes: secondary index → PK → row. Consider covering indexes.
7. **Change buffer** reduces random I/O for non-unique secondary index updates.
8. **MariaDB's Galera Cluster** offers true synchronous multi-master (MySQL Group Replication is the MySQL equivalent).
9. **MySQL's thread model** handles more connections than PostgreSQL, but you should still use ProxySQL.
10. **PostgreSQL is generally more feature-rich** (extensions, types, JSON, FTS), but MySQL/InnoDB has simpler MVCC cleanup (no VACUUM needed).

---

Next: [03-oracle-database.md](03-oracle-database.md) →
