# 3.3 — Oracle Database

> Oracle is the enterprise juggernaut — powering banks, telecoms, and governments.  
> You don't need to use Oracle daily, but understanding its architecture  
> makes you literate in the most commercially successful database ever built.

---

## 1. Architecture — Instance + Database

```
Oracle separates the INSTANCE (memory + processes) from the DATABASE (files).

Instance = SGA + Background Processes (lives in memory)
Database = Datafiles + Redo logs + Control files (lives on disk)

One instance can mount one database.
RAC: multiple instances mount the SAME database (shared storage).

┌──────────────────────────────────────────────────┐
│                ORACLE INSTANCE                     │
│                                                    │
│  ┌──────────────── SGA (System Global Area) ────┐ │
│  │                                               │ │
│  │ ┌──────────────┐ ┌────────────────────┐       │ │
│  │ │ Database      │ │ Shared Pool         │       │ │
│  │ │ Buffer Cache  │ │ ┌────────────────┐ │       │ │
│  │ │ (like PG's    │ │ │Library Cache   │ │       │ │
│  │ │ shared_buffers│ │ │(parsed SQL,    │ │       │ │
│  │ │ )             │ │ │execution plans)│ │       │ │
│  │ └──────────────┘ │ ├────────────────┤ │       │ │
│  │                   │ │Data Dictionary │ │       │ │
│  │ ┌──────────────┐ │ │Cache           │ │       │ │
│  │ │ Redo Log      │ │ └────────────────┘ │       │ │
│  │ │ Buffer        │ └────────────────────┘       │ │
│  │ └──────────────┘                               │ │
│  │ ┌──────────────┐ ┌────────────────┐            │ │
│  │ │ Large Pool    │ │ Java Pool       │            │ │
│  │ │ (RMAN, parallel│ │ (Java stored   │            │ │
│  │ │ query buffers)│ │  procedures)   │            │ │
│  │ └──────────────┘ └────────────────┘            │ │
│  └───────────────────────────────────────────────┘ │
│                                                    │
│  ┌── PGA (Program Global Area) ──┐                │
│  │ Per-session private memory     │                │
│  │ - Sort area                    │                │
│  │ - Hash join area               │                │
│  │ - Session variables            │                │
│  │ - Cursor state                 │                │
│  └───────────────────────────────┘                │
│                                                    │
│  Background Processes:                             │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐   │
│  │DBWn  │ │LGWR  │ │CKPT  │ │SMON  │ │PMON  │   │
│  │(write│ │(log  │ │(chk- │ │(sys  │ │(proc │   │
│  │dirty │ │writer│ │point)│ │monit)│ │monit)│   │
│  │pages)│ │)     │ │      │ │      │ │      │   │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘   │
│  ┌──────┐ ┌──────┐ ┌──────┐                      │
│  │ARCn  │ │MMON  │ │RECO  │                      │
│  │(arch-│ │(AWR  │ │(dist │                      │
│  │iver) │ │stats)│ │recov)│                      │
│  └──────┘ └──────┘ └──────┘                      │
└──────────────────────────────────────────────────┘
                    │
                    ▼ Disk
    ┌──────────┬──────────┬──────────┐
    │Datafiles │Online    │Control   │
    │(.dbf)    │Redo Logs │Files     │
    │          │          │          │
    │Tablespace│Circular  │Database  │
    │storage   │redo log  │metadata  │
    │          │groups    │          │
    └──────────┘──────────┘──────────┘
```

### Key Background Processes

```
DBWn (Database Writer):
  - Writes dirty blocks from buffer cache to datafiles
  - Multiple writers: DBW0, DBW1, ...
  - Equivalent to PostgreSQL's bgwriter + checkpointer

LGWR (Log Writer):
  - Writes redo log buffer to online redo log files
  - Triggered on commit, every 3 seconds, or when buffer 1/3 full
  - Equivalent to PostgreSQL's WAL writer

CKPT (Checkpoint):
  - Signals DBWn to write, updates control file and datafile headers
  
SMON (System Monitor):
  - Instance recovery on startup (redo → undo)
  - Coalesces free tablespace extents
  
PMON (Process Monitor):
  - Cleans up after failed user processes
  - Releases locks held by dead sessions
  - Registers database with listener

ARCn (Archiver):
  - Copies filled online redo log files to archive destination
  - Equivalent to PostgreSQL's archive_command

MMON (Manageability Monitor):
  - Collects AWR snapshots
  - Issues alerts (tablespace full, etc.)
```

---

## 2. Oracle MVCC — Undo Segments

```
Oracle MVCC uses UNDO SEGMENTS (rollback segments):

On UPDATE:
  1. Copy old row version to UNDO segment
  2. Modify current block in-place
  3. Write redo for both the undo record and the data change

On SELECT (consistent read):
  If data block SCN > query start SCN:
    → Follow undo chain to reconstruct the version the query should see
    → "CR clone" the block in buffer cache with old values

This is fundamentally different from PostgreSQL:
  PostgreSQL: old version stays in-place, new version added → heap bloat, needs VACUUM
  Oracle: old version moved to undo → current block always has latest → no VACUUM needed!
  
ORA-01555 "Snapshot too old":
  If a long query needs an undo record that's been overwritten:
  → Cannot reconstruct consistent read → error
  Solution: increase undo_retention, size undo tablespace larger
  PostgreSQL equivalent: no direct equivalent (dead tuples stay until vacuumed)
```

### System Change Number (SCN)

```
Oracle uses SCN (System Change Number) instead of transaction IDs.
SCN is a monotonically increasing 48-bit counter (essentially a logical timestamp).

Every committed transaction gets an SCN.
Every data block header stores the SCN of last modification.

Consistent reads:
  Query starts at SCN 1000
  Reads block with SCN 1050 → too new
  → Go to undo, find version at SCN 1000
  
SCN is global across the database (not per-table).
RAC: SCN is synchronized across all instances via GCS (Global Cache Service).
```

---

## 3. Oracle-Specific SQL Features

```sql
-- Hierarchical queries (CONNECT BY — Oracle-specific, before CTEs existed):
SELECT employee_id, manager_id, LEVEL,
       LPAD(' ', 2 * LEVEL) || last_name AS org_chart
FROM employees
START WITH manager_id IS NULL
CONNECT BY PRIOR employee_id = manager_id
ORDER SIBLINGS BY last_name;
-- PostgreSQL equivalent: recursive CTE (WITH RECURSIVE)

-- MERGE (Oracle's UPSERT — also in SQL standard):
MERGE INTO target t
USING source s ON (t.id = s.id)
WHEN MATCHED THEN
    UPDATE SET t.name = s.name, t.updated = SYSDATE
WHEN NOT MATCHED THEN
    INSERT (id, name) VALUES (s.id, s.name);

-- Flashback queries (time travel!):
SELECT * FROM employees AS OF TIMESTAMP (SYSTIMESTAMP - INTERVAL '1' HOUR);
SELECT * FROM employees AS OF SCN 123456789;

-- Flashback table (undo accidental changes):
FLASHBACK TABLE employees TO TIMESTAMP (SYSTIMESTAMP - INTERVAL '30' MINUTE);

-- Analytic functions (Oracle pioneered window functions):
SELECT department_id, salary,
       RATIO_TO_REPORT(salary) OVER (PARTITION BY department_id) AS pct_of_dept,
       LISTAGG(last_name, ', ') WITHIN GROUP (ORDER BY last_name)
           OVER (PARTITION BY department_id) AS all_names
FROM employees;

-- Model clause (spreadsheet-like calculations in SQL):
SELECT * FROM sales
MODEL
    PARTITION BY (product)
    DIMENSION BY (year)
    MEASURES (amount)
RULES (
    amount[2025] = amount[2024] * 1.10  -- project 10% growth
);

-- PL/SQL (Oracle's procedural language):
CREATE OR REPLACE PROCEDURE raise_salary(
    p_emp_id IN NUMBER,
    p_amount IN NUMBER
) AS
    v_current_salary NUMBER;
BEGIN
    SELECT salary INTO v_current_salary
    FROM employees WHERE employee_id = p_emp_id
    FOR UPDATE;  -- lock row
    
    UPDATE employees SET salary = salary + p_amount
    WHERE employee_id = p_emp_id;
    
    COMMIT;
EXCEPTION
    WHEN NO_DATA_FOUND THEN
        RAISE_APPLICATION_ERROR(-20001, 'Employee not found');
END;
/
```

---

## 4. Oracle RAC (Real Application Clusters)

```
RAC: Multiple Oracle instances share ONE database on shared storage.

┌──────────┐  ┌──────────┐  ┌──────────┐
│Instance 1│  │Instance 2│  │Instance 3│   ← separate servers
│SGA + PGA │  │SGA + PGA │  │SGA + PGA │
└─────┬────┘  └─────┬────┘  └─────┬────┘
      │             │             │
      │    Interconnect (private high-speed network)
      │     (Cache Fusion — transfers blocks between instances)
      │             │             │
      └─────────────┼─────────────┘
                    │
              ┌─────▼──────┐
              │  Shared     │   ← ASM (Automatic Storage Management)
              │  Storage    │      or shared filesystem
              │  (SAN/NAS)  │
              └────────────┘

Cache Fusion:
  When Instance 1 needs a block held by Instance 2:
  → Block transferred over interconnect (not from disk!)
  → Coordinated by GCS (Global Cache Service) and GES (Global Enqueue Service)

Benefits:
  - Scale-out: add nodes for more processing power
  - High availability: node failure → other nodes continue
  - Single database: no replication lag, no data distribution complexity

Downsides:
  - Extremely expensive ($$$$)
  - Complex to manage
  - Cache fusion overhead limits scalability (diminishing returns > 4-8 nodes)
  - Requires shared storage (SAN)
  - Not the same as sharding (all nodes access all data)
```

---

## 5. Data Guard (Oracle Replication)

```
Oracle Data Guard = standby database for DR and read scaling

Physical Standby:
  - Exact block-for-block copy (redo apply)
  - Can be opened for read (Active Data Guard — extra license $$)
  - Equivalent to PostgreSQL streaming replication

Logical Standby:
  - SQL apply (converts redo to SQL statements)
  - Can have additional tables/indexes
  - Equivalent to PostgreSQL logical replication

Far Sync:
  - Zero-data-loss to a remote site via a lightweight intermediate
  - Primary → Far Sync (local) → Standby (remote)
  - Far Sync just relays redo logs (no full database)

Switchover: planned role swap (primary ↔ standby), zero data loss
Failover: unplanned (primary dies), potential data loss

Protection Modes:
  Maximum Protection: zero data loss, sync redo shipping (halt if standby unreachable)
  Maximum Availability: zero data loss, sync redo (fail to async if standby dies)
  Maximum Performance: async redo, some data loss possible (default)
```

---

## 6. Oracle Exadata

```
Exadata = Oracle's engineered system (appliance):
  Database servers + Storage servers + InfiniBand interconnect

Key innovation: Smart Scan
  Instead of sending data TO the database server for filtering:
  → Push the WHERE clause DOWN to the storage server
  → Storage returns only matching rows
  → Massively reduces data transfer for full table scans

  SELECT * FROM billion_row_table WHERE status = 'active';
  Without Exadata: read entire table to DB server, filter there
  With Exadata smart scan: storage servers filter, return ~1% of data

Other features:
  - Storage indexes: in-memory min/max per storage region
  - Hybrid Columnar Compression (HCC): 10-50x compression for cold data
  - Flash cache: SSDs in storage servers act as extended buffer cache
  - RDMA over InfiniBand: kernel-bypass networking between DB and storage
```

---

## 7. Oracle vs PostgreSQL

```
Feature                 Oracle                          PostgreSQL
────────────────────────────────────────────────────────────────────
Cost                    $$$$$$ (per-core licensing)      Free (open source)
MVCC                    Undo-based (no vacuum needed)   Heap tuple versioning
Bloat                   Undo tablespace growth          Heap+index bloat
Cleanup                 Automatic undo purge            VACUUM required
HA/Scaling              RAC (shared-everything)         Streaming replication
DR                      Data Guard                      pg_basebackup + PITR
Flashback               Full flashback suite            No native equivalent
Partitioning            Very mature, all types          Good (PG 10+)
PL language             PL/SQL (mature, huge ecosystem) PL/pgSQL (similar syntax)
Optimizer               Very mature cost-based          Very good, improving
Parallel query          Highly parallel                 Good (PG 9.6+)
JSON                    JSON (less mature)              JSONB (superior)
Extensions              No extension model              Rich extensions
Geospatial              Oracle Spatial                  PostGIS (often better)
Full-text search        Oracle Text                     tsvector/tsquery
Autonomous features     Autonomous Database (Cloud)     N/A
```

---

## Key Takeaways

1. **Oracle separates Instance (memory) from Database (disk)** — this enables RAC where multiple instances share one database.
2. **Undo-based MVCC** means Oracle NEVER needs VACUUM. Old versions are in undo segments, cleaned automatically.
3. **ORA-01555 "Snapshot too old"** is Oracle's trade-off: long queries may fail if undo is recycled.
4. **PL/SQL** is a very mature procedural language — entire business logic layers run inside Oracle.
5. **RAC** provides shared-everything scale-out, but with diminishing returns and massive cost.
6. **Data Guard** is enterprise-grade replication/DR, far more automated than PostgreSQL's.
7. **Flashback** lets you query past states and UNDO accidental changes — an incredibly powerful feature.
8. **The real cost**: Oracle licensing is per-core and can cost $50K-$500K+ per server. This is why PostgreSQL is eating Oracle's market share.

---

Next: [04-sql-server.md](04-sql-server.md) →
