# 2.6 — Memory Management: Buffer Pools, Work Memory, and Architecture

> The database is a memory management system that happens to persist to disk.  
> Understanding how it uses RAM is the key to making it fast.

---

## 1. Database Memory Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                        DATABASE PROCESS MEMORY                       │
│                                                                      │
│  ┌───────────────────────────────────────────────────────────┐       │
│  │              SHARED MEMORY                                │       │
│  │                                                           │       │
│  │  ┌──────────────────────┐  ┌──────────────────────────┐  │       │
│  │  │   BUFFER POOL        │  │  WAL BUFFERS             │  │       │
│  │  │   (shared_buffers)   │  │  (wal_buffers)           │  │       │
│  │  │   Default: 128 MB    │  │  Default: ~4 MB          │  │       │
│  │  │   Rec: 25% of RAM    │  │  Rec: 64 MB for busy     │  │       │
│  │  └──────────────────────┘  └──────────────────────────┘  │       │
│  │                                                           │       │
│  │  ┌──────────────────────┐  ┌──────────────────────────┐  │       │
│  │  │   LOCK TABLE         │  │  PROC ARRAY              │  │       │
│  │  │                      │  │  (MVCC snapshot info)     │  │       │
│  │  └──────────────────────┘  └──────────────────────────┘  │       │
│  │                                                           │       │
│  │  ┌──────────────────────┐  ┌──────────────────────────┐  │       │
│  │  │   CLOG / XACT        │  │  SUBTRANS               │  │       │
│  │  │   (commit status)    │  │  (subtransactions)       │  │       │
│  │  └──────────────────────┘  └──────────────────────────┘  │       │
│  └───────────────────────────────────────────────────────────┘       │
│                                                                      │
│  ┌───────────────────────────────────────────────────────────┐       │
│  │              PER-CONNECTION MEMORY (× max_connections)    │       │
│  │                                                           │       │
│  │  ┌──────────────────────┐  ┌──────────────────────────┐  │       │
│  │  │   work_mem           │  │  temp_buffers            │  │       │
│  │  │   (sorts, hashes)    │  │  (temp tables)           │  │       │
│  │  │   Default: 4 MB      │  │  Default: 8 MB           │  │       │
│  │  └──────────────────────┘  └──────────────────────────┘  │       │
│  │                                                           │       │
│  │  ┌──────────────────────┐  ┌──────────────────────────┐  │       │
│  │  │   maintenance_work_mem│ │  query parsing +          │  │       │
│  │  │   (VACUUM, CREATE IDX)│ │  planning memory         │  │       │
│  │  │   Default: 64 MB     │  │                          │  │       │
│  │  └──────────────────────┘  └──────────────────────────┘  │       │
│  └───────────────────────────────────────────────────────────┘       │
│                                                                      │
│  ┌───────────────────────────────────────────────────────────┐       │
│  │              OS PAGE CACHE (kernel manages this)          │       │
│  │              Remaining RAM after shared_buffers            │       │
│  │              Also caches data files, WAL files, etc.      │       │
│  └───────────────────────────────────────────────────────────┘       │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 2. Shared Buffers (Buffer Pool) Deep Dive

### Architecture

```
PostgreSQL shared_buffers is a hash table of in-memory pages:

  Page Table (hash map):
    hash(tablespace_id, relation_id, block_number) → buffer_id
  
  Buffer Descriptors Array (metadata per slot):
    [buffer_id] → {
      tag:        (tablespace, relation, block) — which disk page is here
      state:      valid, dirty, io_in_progress
      usage_count: 0-5 (for clock sweep)
      ref_count:   number of active pins (processes using this page)
      content_lock: shared (read) or exclusive (write)
    }
  
  Buffer Pool Array (actual page data):
    [buffer_id] → 8 KB page content

  Clock sweep hand → moves through buffer descriptors
```

### The Clock Sweep Algorithm (PostgreSQL)

```
PostgreSQL uses a "clock sweep" variant of LRU:

Each buffer has a usage_count (0 to 5):
  - When a buffer is accessed: usage_count = min(usage_count + 1, 5)
  - Clock hand sweeps through buffers:
    If usage_count > 0: decrement by 1, move to next
    If usage_count = 0 AND ref_count = 0: THIS is the victim → evict

              ┌─── clock hand
              ▼
  [ uc=3 ][ uc=0 ][ uc=2 ][ uc=0 ][ uc=1 ][ uc=0 ]
                                              ^
                                          ref_count=0
                                          → VICTIM

  A buffer accessed 5 times survives 5 passes of the clock.
  A buffer accessed once survives 1 pass.
  
  This naturally separates "hot" pages (frequently accessed index roots)
  from "cold" pages (one-time table scan pages).
```

### Double Buffering Problem

```
PostgreSQL shared_buffers AND the OS page cache both cache the same pages!

  App → PG shared_buffers (first cache) → OS page cache (second cache) → disk

When PG reads a page:
  1. Check shared_buffers → miss
  2. Issue read() syscall → OS checks its page cache → HIT (great!) or miss (disk I/O)
  3. Copy page from kernel space to shared_buffers (user space)

When PG writes (evicts) a dirty page:
  1. Write to OS page cache (via write() syscall) — fast
  2. OS eventually flushes to disk (or PG explicitly fsyncs)

The double caching means the SAME page might exist in both caches.
This wastes memory but is unavoidable with the process-per-connection model.

Solution: Don't set shared_buffers too high!
  Recommended: 25% of total RAM (remaining 75% used by OS page cache)
  
  With 64 GB RAM:
    shared_buffers = 16 GB
    OS page cache uses ~40-48 GB
    Connections + maintenance + OS = remaining
  
  Setting shared_buffers = 50 GB would starve the OS page cache
  and actually REDUCE overall performance.
```

### Effective Cache Size

```sql
-- effective_cache_size tells the optimizer how much total cache is available
-- (shared_buffers + OS page cache estimate)
-- It does NOT allocate memory — it's just a hint for the query planner.

-- With 64 GB RAM, shared_buffers = 16 GB:
effective_cache_size = '48GB'  -- 75% of total RAM

-- If too low: optimizer thinks disk is slow → avoids index scans → worse plans
-- If too high: optimizer overestimates cache hits → wrong cost estimates
```

---

## 3. Work Memory (work_mem)

### What It's For

```
work_mem is allocated PER SORT/HASH OPERATION PER CONNECTION.

Operations that use work_mem:
  - ORDER BY (sort)
  - DISTINCT (sort or hash)
  - GROUP BY (hash aggregate)
  - Hash joins (hash table for smaller input)
  - Merge joins (sort phase)
  - Window functions (sort + buffer)
  - UNION / INTERSECT / EXCEPT
  - Recursive CTEs
  
One query can use MULTIPLE work_mem allocations!

Example query:
  SELECT DISTINCT dept_id, AVG(salary)
  FROM employee
  WHERE hire_date > '2024-01-01'
  GROUP BY dept_id     ← hash aggregate: 1 × work_mem
  ORDER BY avg         ← sort: 1 × work_mem
  
  Total: 2 × work_mem per connection running this query
```

### What Happens When work_mem Is Exceeded

```
If the work_mem budget is too small for a sort or hash:

SORT:
  In-memory quicksort → OVERFLOWS to disk (external merge sort)
  → Creates temp files, does multiple merge passes
  → 10-100x slower than in-memory sort

  Visible in EXPLAIN ANALYZE:
    Sort Method: quicksort  Memory: 1024kB      ← in-memory ✓ fast
    Sort Method: external merge  Disk: 50784kB  ← SPILLED TO DISK ✗ slow

HASH JOIN:
  If hash table doesn't fit in work_mem:
  → Multi-batch hash join (partition both inputs to disk, process batch by batch)
  → Much slower

  Visible in EXPLAIN ANALYZE:
    Hash  Batches: 1  Memory Usage: 3000kB   ← single batch ✓ in-memory
    Hash  Batches: 16  Memory Usage: 3000kB  ← 16 batches ✗ spilled
```

### Tuning work_mem

```sql
-- Default: 4 MB (conservative, safe for many connections)
SHOW work_mem;

-- Calculate safe maximum:
-- Available RAM for work = Total RAM - shared_buffers - OS needs
-- With 64 GB RAM, shared_buffers = 16 GB:
--   Available for work: ~32 GB
--   Max connections: 100
--   Avg sorts per query: 3
--   Safe work_mem: 32 GB / (100 × 3) ≈ 100 MB

-- Set globally:
ALTER SYSTEM SET work_mem = '64MB';

-- Set per session (for a complex analytical query):
SET work_mem = '512MB';
-- Run your big query
RESET work_mem;

-- Set per transaction:
SET LOCAL work_mem = '256MB';
-- Resets automatically at end of transaction

-- DANGER: Setting work_mem = 1GB globally with max_connections = 200
-- Worst case: 200 connections × 5 sorts each × 1 GB = 1 TB of memory allocation
-- → OOM killer WILL kill PostgreSQL
```

---

## 4. Maintenance Work Memory

```sql
-- maintenance_work_mem: for maintenance operations
-- Used by: VACUUM, CREATE INDEX, ALTER TABLE ADD FOREIGN KEY, CLUSTER

-- Default: 64 MB
-- Recommended: 512 MB - 2 GB (only one VACUUM runs per table at a time)

ALTER SYSTEM SET maintenance_work_mem = '1GB';

-- autovacuum_work_mem: separate budget for autovacuum (default: -1 = use maintenance_work_mem)
-- If you have 3 autovacuum workers × 1 GB = 3 GB dedicated to autovacuum

-- Why larger is better for CREATE INDEX:
-- CREATE INDEX must sort all indexed values.
-- work_mem determines if the sort is in-memory or disk.
-- A 10 GB index with maintenance_work_mem = 64 MB → many disk merge passes → slow
-- With maintenance_work_mem = 4 GB → likely single-pass → fast
```

---

## 5. Connection Memory Overhead

```
Each PostgreSQL connection is a SEPARATE OS PROCESS (not a thread!).

Per-connection memory:
  - Stack: ~1-8 MB
  - Catalog cache: ~1-5 MB (cached system catalog lookups)
  - work_mem allocations: variable
  - Query plans: variable
  - Temp buffers: up to temp_buffers (default 8 MB)
  - Libpq buffers: ~1 MB
  
  Rough estimate: 5-10 MB per IDLE connection
                  50-500 MB per ACTIVE connection (with sorts/hashes)

max_connections = 100 (default):
  Idle memory: ~1 GB
  Active memory: up to 50 GB (if all active with large sorts)

max_connections = 5000:
  Idle memory: ~50 GB (just for connections doing nothing!)
  → This is why PostgreSQL can't handle 5000 direct connections.
  → Solution: connection pooling (PgBouncer, pgpool-II)
```

### Connection Pooling

```
PgBouncer sits between the application and PostgreSQL:

  App (1000 connections) → PgBouncer → PostgreSQL (50 connections)

Modes:
  Session pooling:     PG connection assigned per client session (least aggressive)
  Transaction pooling: PG connection assigned per transaction (most common)
  Statement pooling:   PG connection assigned per statement (most aggressive, limitations)

                          ┌──────────────┐
  App Connection 1 ──────→│              │
  App Connection 2 ──────→│  PgBouncer   │──→ PG Connection 1
  App Connection 3 ──────→│  (pool: 20)  │──→ PG Connection 2
  ...                     │              │    ...
  App Connection 500 ────→│              │──→ PG Connection 20
                          └──────────────┘

Transaction pooling:
  1. App sends BEGIN → PgBouncer assigns a PG connection
  2. App sends queries → routed to that PG connection
  3. App sends COMMIT → PG connection released back to pool
  4. Another app's next transaction gets the same PG connection

Caveats with transaction pooling:
  - Can't use session-level features: SET, prepared statements, LISTEN/NOTIFY, temp tables
  - Can't use DEALLOCATE ALL (PgBouncer handles this internally)
  - Application must be stateless between transactions
```

---

## 6. Huge Pages (Large Pages)

```
Normal memory pages: 4 KB
Huge pages: 2 MB (Linux) or 1 GB (Linux, rare)

The CPU has a Translation Lookaside Buffer (TLB) that maps virtual → physical addresses.
TLB has limited entries (~1000-2000).

With 4 KB pages + 16 GB shared_buffers:
  Pages: 16 GB / 4 KB = 4,194,304 pages
  TLB entries: ~1500
  → Massive TLB misses → each miss costs ~10-100 ns
  → 5-10% CPU overhead just for address translation!

With 2 MB huge pages + 16 GB shared_buffers:
  Pages: 16 GB / 2 MB = 8,192 pages
  TLB entries: ~1500
  → Almost everything fits in TLB → near-zero overhead

Enabling huge pages:
  # Linux: allocate huge pages
  sudo sysctl -w vm.nr_hugepages=8500  # slightly more than shared_buffers / 2 MB
  
  # PostgreSQL:
  huge_pages = try   # 'on', 'off', or 'try' (try = use if available)
  
  # Verify:
  cat /proc/meminfo | grep HugePages

Huge pages also can't be swapped to disk → more predictable performance.
```

---

## 7. NUMA-Aware Memory

```
NUMA (Non-Uniform Memory Access):
  In multi-socket servers, each CPU socket has its own "local" memory.
  Accessing local memory: ~100 ns
  Accessing remote memory (another CPU's socket): ~300 ns

  ┌──────────────┐     ┌──────────────┐
  │   CPU 0      │     │   CPU 1      │
  │   ┌──────┐   │     │   ┌──────┐   │
  │   │ Core │   │     │   │ Core │   │
  │   │ Core │   │     │   │ Core │   │
  │   └──────┘   │     │   └──────┘   │
  │   ↕ fast     │     │   ↕ fast     │
  │   ┌──────┐   │←──→│   ┌──────┐   │
  │   │ RAM  │   │ slow│   │ RAM  │   │
  │   │ 32GB │   │     │   │ 32GB │   │
  │   └──────┘   │     │   └──────┘   │
  └──────────────┘     └──────────────┘

Problem: if shared_buffers is allocated on CPU 0's RAM,
         processes on CPU 1 access it at 3x latency.

Solution: interleave shared memory across all NUMA nodes:
  numactl --interleave=all postgres  (spread evenly)
  
  Or in postgresql.conf (PG 14+):
  huge_pages = on
  # PostgreSQL 14+ uses mbind() to interleave shared memory
```

---

## 8. Memory-Mapped I/O (mmap) Databases

```
Alternative to a buffer pool: let the OS handle caching via mmap().

mmap() maps a file directly into the process's virtual address space.
Reading a memory-mapped page: if not in RAM → OS page fault → OS reads from disk.

Databases using mmap: LMDB, MongoDB (WiredTiger has optional mmap), early SQLite

    Traditional (PostgreSQL):          mmap approach (LMDB):
    ┌────────────────────┐            ┌────────────────────┐
    │ App Process         │            │ App Process         │
    │  ┌──────────────┐  │            │  memory-mapped file │
    │  │ Buffer Pool   │  │            │  ┌──────────────┐  │
    │  │ (user space)  │  │            │  │ Virtual Mem   │  │
    │  └──────┬───────┘  │            │  │ = file on disk │  │
    │         │ read()   │            │  └──────┬───────┘  │
    │         ▼          │            │         │ page fault│
    │  ┌──────────────┐  │            │  ┌──────▼───────┐  │
    │  │ OS Page Cache │  │            │  │ OS Page Cache │  │
    │  └──────┬───────┘  │            │  └──────┬───────┘  │
    │         ▼          │            │         ▼          │
    │      [ Disk ]      │            │      [ Disk ]      │
    └────────────────────┘            └────────────────────┘

Pros of mmap:
  + Simpler code (no buffer pool management)
  + OS handles eviction, prefetching, etc.
  + Zero-copy reads (data in page cache, accessed directly)
  + Can exceed available RAM (OS handles paging)

Cons of mmap (why PostgreSQL doesn't use it):
  − Can't control eviction policy (OS uses generic LRU, not database-aware)
  − No control over which pages to keep in memory
  − TLB shootdowns cause performance stalls
  − msync()/fsync() behavior is inconsistent across OSes
  − Can't implement proper WAL protocol (no page-level control)
  − Page faults can stall a query unpredictably
  − Andy Pavlo: "mmap is a bad idea for database management systems"
     (published a paper on this!)
```

---

## 9. PostgreSQL Memory Configuration Cheat Sheet

```
Server: 64 GB RAM, 100 connections expected, SSD storage

# === Shared Memory ===
shared_buffers = '16GB'           # 25% of RAM
effective_cache_size = '48GB'     # 75% of RAM (hint for planner)
huge_pages = 'try'                # Use huge pages if available

# === WAL ===
wal_buffers = '64MB'              # Larger for write-heavy
min_wal_size = '1GB'
max_wal_size = '4GB'

# === Per-Connection Memory ===
work_mem = '64MB'                 # Careful! Per sort/hash per connection
maintenance_work_mem = '2GB'      # For VACUUM, CREATE INDEX
temp_buffers = '32MB'             # Per-connection temp table cache

# === Connections ===
max_connections = 100             # Keep low, use connection pooler
# PgBouncer: pool_size = 20-50 actual PG connections

# === Background Workers ===
max_parallel_workers = 8
max_parallel_workers_per_gather = 4
max_parallel_maintenance_workers = 4

# === VACUUM ===
autovacuum_work_mem = '512MB'     # Independent from maintenance_work_mem
autovacuum_max_workers = 3

# === Cost Parameters (SSD) ===
random_page_cost = 1.1            # SSD: random ≈ sequential
seq_page_cost = 1.0
effective_io_concurrency = 200    # SSD can handle many concurrent reads
```

### Memory Budget Calculation

```
Total RAM: 64 GB
  shared_buffers:     16 GB
  OS page cache:      ~35 GB (remaining after everything else)
  OS + processes:      2 GB
  Connection overhead: 100 × 10 MB = 1 GB (idle baseline)
  work_mem headroom:   100 × 64 MB × 2 sorts = 12.8 GB (worst case)
  maintenance:         2 GB
  autovacuum:         3 × 512 MB = 1.5 GB
  WAL buffers:         64 MB
  ────────────────────────────
  Total worst case:   ~70 GB  ← EXCEEDS 64 GB!
  
  But worst case rarely happens (not all connections doing sorts simultaneously).
  Realistic active load: ~50-55 GB usage. Headroom exists.
  
  If overcommit happens → Linux OOM killer → PostgreSQL DIES.
  Solution: reduce max_connections, use connection pooler, or reduce work_mem.
  
  vm.overcommit_memory = 2  (Linux: don't overcommit, reject if insufficient)
  vm.overcommit_ratio = 90  (allow up to 90% of RAM to be committed)
```

---

## 10. InnoDB Memory Architecture (Comparison)

```
┌──────────────────────────────────────────────┐
│  InnoDB BUFFER POOL (innodb_buffer_pool_size) │
│  Recommended: 70-80% of RAM                  │
│  (Higher than PG because InnoDB replaces     │
│   the OS page cache role more aggressively)  │
│                                              │
│  Contains:                                   │
│  ┌───────────────┐  ┌────────────────────┐  │
│  │ Data pages     │  │ Index pages        │  │
│  │ (clustered idx)│  │ (secondary idx)    │  │
│  └───────────────┘  └────────────────────┘  │
│  ┌───────────────┐  ┌────────────────────┐  │
│  │ Undo pages     │  │ Adaptive Hash Idx  │  │
│  └───────────────┘  └────────────────────┘  │
│  ┌───────────────┐  ┌────────────────────┐  │
│  │ Change buffer  │  │ Lock info          │  │
│  └───────────────┘  └────────────────────┘  │
├──────────────────────────────────────────────┤
│  REDO LOG BUFFER (innodb_log_buffer_size)     │
│  Default: 16 MB                              │
├──────────────────────────────────────────────┤
│  Per-connection:                             │
│  sort_buffer_size (default: 256 KB)          │
│  join_buffer_size (default: 256 KB)          │
│  read_buffer_size, read_rnd_buffer_size      │
│  (These are MUCH smaller than PG's work_mem) │
└──────────────────────────────────────────────┘

Key difference: InnoDB's buffer pool is the PRIMARY cache.
There's less reliance on OS page cache (set innodb_flush_method=O_DIRECT 
to bypass OS cache entirely and avoid double buffering).

PostgreSQL: shared_buffers = 25% + OS page cache = 75%
InnoDB:     buffer_pool = 75% + O_DIRECT (no OS cache for data files)
```

---

## Key Takeaways

1. **shared_buffers = 25% of RAM** in PostgreSQL. The OS page cache handles the rest. Don't go higher.
2. **work_mem is per-operation, not per-connection.** Multiply by max_connections × sorts_per_query for worst case.
3. **Disk spillover kills performance.** Watch for "external merge" sorts and multi-batch hash joins in EXPLAIN.
4. **Connection pooling is mandatory** for PostgreSQL at scale. Use PgBouncer in transaction mode.
5. **Huge pages** eliminate TLB misses — easy 5-10% performance gain for large shared_buffers.
6. **mmap is tempting but wrong** for serious databases — no control over eviction, durability, or page-level operations.
7. **Memory budget carefully.** An OOM kill takes down the whole database.
8. **InnoDB uses 70-80% for buffer pool + O_DIRECT.** PostgreSQL uses 25% + OS page cache. Different philosophies, both work.

---

**Phase 2 Complete!** You now understand:
- How data lives on disk (pages, heap files, B+ trees, LSM trees)
- Every index type and when to use each
- How the query optimizer turns SQL into execution plans
- MVCC, locking, isolation levels, and concurrency anomalies
- Crash recovery with WAL and ARIES
- Memory architecture and tuning

Next: [Phase 3 — Master Specific Database Systems](../phase3/01-postgresql-deep.md) →
