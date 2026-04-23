# 3.5 — SQLite

> SQLite is the most deployed database in the world.  
> It's on every phone, every browser, every Mac, every Windows 10+ machine.  
> **Trillions** of SQLite databases are in active use right now.  
> It is not a toy — it is an engineering masterpiece of simplicity.

---

## 1. What SQLite IS (and ISN'T)

```
SQLite IS:
  ✓ A file-format (not a server)
  ✓ Embedded inside your application (linked as a C library, ~300 KB)
  ✓ ACID compliant
  ✓ Zero-configuration, zero-administration
  ✓ Single-file database (one .db file = entire database)
  ✓ Cross-platform (the file format is portable: big-endian/little-endian, 32/64-bit)
  ✓ Public domain (no license, no restrictions)
  ✓ The most tested software in existence (100% MC/DC branch coverage)

SQLite is NOT:
  ✗ A client/server database (no network protocol)
  ✗ Designed for high write concurrency (one writer at a time)
  ✗ A replacement for PostgreSQL/MySQL
  ✗ Designed for multi-machine deployment

Think of SQLite as fopen() on steroids — it "opens" a structured file,
not connects to a server.
```

---

## 2. Architecture — Elegantly Simple

```
┌─────────────────────────────────────────┐
│            Application Process           │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │         SQLite Library             │ │
│  │                                    │ │
│  │  ┌──────────┐  ┌──────────────┐   │ │
│  │  │ SQL       │  │ Virtual      │   │ │
│  │  │ Compiler  │→ │ Machine (VDBE)│   │ │
│  │  │ (parser + │  │ (bytecode    │   │ │
│  │  │ planner)  │  │  interpreter)│   │ │
│  │  └──────────┘  └──────┬───────┘   │ │
│  │                        │           │ │
│  │  ┌──────────────────── ▼ ────────┐ │ │
│  │  │         B-Tree Module          │ │ │
│  │  │  (one B-tree per table/index)  │ │ │
│  │  └────────────┬──────────────────┘ │ │
│  │               │                    │ │
│  │  ┌──────────── ▼ ────────────────┐ │ │
│  │  │         Pager Module           │ │ │
│  │  │  Page cache + transaction mgmt │ │ │
│  │  │  Journal or WAL               │ │ │
│  │  └────────────┬──────────────────┘ │ │
│  │               │                    │ │
│  │  ┌──────────── ▼ ────────────────┐ │ │
│  │  │         OS Interface (VFS)     │ │ │
│  │  │  File I/O, locking, mmap      │ │ │
│  │  └───────────────────────────────┘ │ │
│  └────────────────────────────────────┘ │
└─────────────┬───────────────────────────┘
              │
              ▼
    ┌──────────────────┐
    │ database.db      │  ← Single file on disk
    │ (pages of 4096B) │
    └──────────────────┘

Everything runs in-process. No IPC, no network, no server.
A single function call (sqlite3_exec) goes through the entire stack.
```

### File Format

```
The SQLite database file:
  - Header: first 100 bytes (magic string, page size, versions, etc.)
  - Pages: fixed size (default 4096 bytes, configurable at creation)
  - Page 1: contains the header + root of sqlite_master table
  - sqlite_master: catalog table listing all tables, indexes, etc.

  Page types:
    - B-tree interior pages (table or index)
    - B-tree leaf pages (table = actual row data)
    - Overflow pages (for rows > ~25% of page size)
    - Freelist pages (deleted pages, reusable)
    - Lock-byte page (page at offset 1073741824, for locking)

  Table B-trees: rowid is the key, row data is the value
  Index B-trees: indexed columns are the key, rowid is the value
  WITHOUT ROWID tables: PK directly in the B-tree key (clustered, like InnoDB)
```

---

## 3. Concurrency — WAL vs Rollback Journal

### Rollback Journal (Legacy Default)

```
Write path:
  1. Read original page from database file
  2. COPY original page to rollback journal (.db-journal)
  3. Modify page in-place in cache
  4. On COMMIT: delete journal file → changes are permanent
  5. On ROLLBACK: copy journal pages back → restore original state

Locking levels:
  UNLOCKED → SHARED → RESERVED → PENDING → EXCLUSIVE

  SHARED:    reading (multiple readers allowed)
  RESERVED:  one writer preparing (others can still read)
  PENDING:   writer waiting for readers to finish
  EXCLUSIVE: writing to database file (blocks all readers!)

  Problem: writer blocks ALL readers during write+commit
  
  This is why people say "SQLite doesn't support concurrent writes."
  It's actually: ONE writer, and the writer blocks readers briefly.
```

### WAL Mode (Write-Ahead Logging)

```sql
PRAGMA journal_mode = WAL;

WAL mode reverses the pattern:
  - Writes go to a separate WAL file (.db-wal)
  - Readers read from the main database file + check WAL for newer versions
  - Writers NEVER block readers
  - Readers NEVER block writers
  - ONLY ONE writer at a time (still)

How it works:
  Database file: original pages (not modified during transactions)
  WAL file: append-only log of modified pages
  WAL index (.db-shm): shared memory file for fast WAL page lookup

  Readers: check WAL index, if page is in WAL → use WAL version, else database file
  Writer: append new pages to WAL
  Checkpoint: copy WAL pages back to database file (background, non-blocking)

  Transactions read a SNAPSHOT: they see all WAL frames up to a certain point.
  → True snapshot isolation for readers!

WAL advantages:
  ✓ Readers don't block writers (huge for read-heavy apps)
  ✓ Writes are sequential (faster: append to WAL vs random writes to DB)
  ✓ Better concurrency

WAL disadvantages:
  ✗ Slightly slower reads (must check WAL + database)
  ✗ WAL file can grow large between checkpoints
  ✗ Doesn't work over network filesystems (uses shared memory)
  ✗ Still only ONE writer at a time
```

### WAL2 and BEGIN CONCURRENT (Experimental)

```
SQLite has experimental extensions for better write concurrency:

BEGIN CONCURRENT (experimental):
  Multiple writers can prepare transactions in parallel.
  At COMMIT time, serialize and check for conflicts.
  If no conflict → commit succeeds.
  If conflict → SQLITE_BUSY → application retries.
  
  This is optimistic concurrency control.
  
  Not in mainline SQLite yet, but available in some forks (libSQL).
```

---

## 4. Key PRAGMAs — SQLite Configuration

```sql
-- PRAGMAs are SQLite's configuration system (like SET in PostgreSQL)

-- === Performance-critical PRAGMAs ===

PRAGMA journal_mode = WAL;          -- ALWAYS use WAL mode
PRAGMA synchronous = NORMAL;        -- WAL mode: NORMAL is safe + fast
                                    -- (FULL = fsync every commit, slower)
                                    -- (OFF = no fsync, data loss risk)

PRAGMA cache_size = -64000;         -- 64 MB page cache (negative = KB)
PRAGMA mmap_size = 268435456;       -- memory-map 256 MB of database file
PRAGMA temp_store = MEMORY;         -- temp tables in memory (not disk)
PRAGMA busy_timeout = 5000;         -- wait 5 seconds on lock (not immediate error)

-- === Correctness PRAGMAs ===

PRAGMA foreign_keys = ON;           -- OFF by default (!!) — ALWAYS enable
PRAGMA strict = ON;                 -- SQLite 3.37+ strict type checking

-- === Useful PRAGMAs ===

PRAGMA table_info(employees);       -- like \d in psql
PRAGMA index_list(employees);       -- list indexes
PRAGMA integrity_check;             -- full database consistency check
PRAGMA quick_check;                 -- fast subset of integrity_check
PRAGMA optimize;                    -- run ANALYZE on tables that need it (3.18+)
PRAGMA wal_checkpoint(TRUNCATE);    -- checkpoint and truncate WAL file
PRAGMA page_count;                  -- number of pages in database
PRAGMA page_size;                   -- page size in bytes

-- The "production" pragma set (run at connection open):
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -64000;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 268435456;
```

---

## 5. SQLite Type System — Flexible (Too Flexible?)

```sql
-- SQLite uses "type affinity" — the column type is a SUGGESTION, not enforced.

CREATE TABLE test (
    id INTEGER PRIMARY KEY,  -- aliases to rowid (special, fast)
    name TEXT,
    age INTEGER,
    salary REAL
);

-- This WORKS in SQLite (no error!):
INSERT INTO test VALUES (1, 123, 'not a number', 'also not a number');
-- name gets integer 123, age gets text 'not a number'

-- Type affinities:
-- TEXT:     prefers text storage
-- NUMERIC:  prefers integer or real
-- INTEGER:  prefers integer
-- REAL:     prefers floating point
-- BLOB:     no preference (stores as-is)

-- Column type name → affinity mapping (by substring matching):
-- Contains "INT"        → INTEGER
-- Contains "CHAR/CLOB/TEXT" → TEXT
-- Contains "BLOB" or no type → BLOB (NONE)
-- Contains "REAL/FLOA/DOUB" → REAL
-- Otherwise              → NUMERIC

-- STRICT tables (SQLite 3.37+):
CREATE TABLE orders (
    id INTEGER PRIMARY KEY,
    customer TEXT NOT NULL,
    amount REAL NOT NULL
) STRICT;
-- Now type violations produce errors! Allowed types: INT, INTEGER, REAL, TEXT, BLOB, ANY

-- INTEGER PRIMARY KEY is special:
-- It becomes an alias for the internal rowid (64-bit signed integer)
-- This is SQLite's fastest access path (direct B-tree lookup by rowid)
-- Auto-increment: AUTOINCREMENT keyword prevents rowid reuse (but adds overhead)
```

---

## 6. WITHOUT ROWID Tables

```sql
-- Normal SQLite tables have a hidden rowid. Table B-tree key = rowid.
-- This means PK lookups on non-rowid PKs require TWO B-tree lookups
-- (index → rowid → table B-tree).

-- WITHOUT ROWID eliminates the hidden rowid:
-- Table B-tree key = PRIMARY KEY directly (like InnoDB clustered index)

CREATE TABLE sessions (
    session_token TEXT PRIMARY KEY,
    user_id INTEGER,
    expires_at TEXT
) WITHOUT ROWID;

-- Benefits:
-- 1. PK lookup is ONE B-tree traversal (not two)
-- 2. Less storage for string PKs (no wasted rowid column)
-- 3. Better for tables with large PK + few/small other columns

-- When to use:
-- ✓ Non-integer primary keys (UUIDs, tokens, composite keys)
-- ✓ Tables accessed mostly by PK
-- ✗ NOT for large rows (no overflow page support)
-- ✗ NOT when rowid scan is needed

-- Composite PK example:
CREATE TABLE graph_edges (
    from_node INTEGER,
    to_node INTEGER,
    weight REAL,
    PRIMARY KEY (from_node, to_node)
) WITHOUT ROWID;
-- Scans by from_node are sequential (prefix of PK)!
```

---

## 7. Virtual Tables and Extensions

```sql
-- Virtual tables: tables backed by custom code (like PostgreSQL FDW)

-- FTS5: Full-Text Search (built-in)
CREATE VIRTUAL TABLE articles_fts USING fts5(title, body, content='articles', content_rowid='id');

-- Populate from existing table:
INSERT INTO articles_fts(articles_fts) VALUES ('rebuild');

-- Search:
SELECT * FROM articles_fts WHERE articles_fts MATCH 'database AND (performance OR optimization)';
SELECT * FROM articles_fts WHERE articles_fts MATCH 'NEAR(database performance, 5)';

-- Ranking:
SELECT *, rank FROM articles_fts WHERE articles_fts MATCH 'sqlite' ORDER BY rank;

-- JSON functions (built-in since 3.38):
SELECT json_extract(data, '$.name') FROM documents;
SELECT * FROM documents, json_each(documents.tags);  -- expand JSON array

-- R-Tree (spatial indexing, built-in):
CREATE VIRTUAL TABLE spatial_idx USING rtree(id, min_x, max_x, min_y, max_y);
INSERT INTO spatial_idx VALUES (1, 10.0, 20.0, 30.0, 40.0);
SELECT * FROM spatial_idx WHERE min_x > 5 AND max_x < 25;  -- bounding box query

-- Generate series (built-in):
SELECT value FROM generate_series(1, 100, 5);

-- CSV virtual table (load CSV as a table):
CREATE VIRTUAL TABLE csv_data USING csv(filename='data.csv', header=yes);
SELECT * FROM csv_data WHERE column1 > 100;
```

---

## 8. SQLite in Production — Patterns and Anti-Patterns

```
When to USE SQLite:
  ✓ Mobile apps (iOS, Android — it's already there)
  ✓ Desktop apps (Electron, native)
  ✓ Embedded systems, IoT
  ✓ CLI tools, local caches
  ✓ Test databases (replace PG in tests for speed)
  ✓ Single-server web apps with low-medium write traffic
  ✓ File format replacement (config, application data)
  ✓ Edge computing (Cloudflare D1, Turso/libSQL)

When NOT to use SQLite:
  ✗ High write concurrency (>100 writes/sec sustained)
  ✗ Multi-server applications needing shared state
  ✗ Very large databases (>1 TB — possible but impractical)
  ✗ Network-accessed database (use PostgreSQL)
  ✗ Heavy concurrent reporting while writing

Production tips:
  1. ALWAYS use WAL mode
  2. ALWAYS set busy_timeout (default 0 = immediate SQLITE_BUSY error)
  3. ALWAYS enable foreign_keys
  4. Keep transactions SHORT (holding a write lock blocks other writers)
  5. Use prepared statements (avoid SQL injection, better performance)
  6. Use connection pools carefully (SQLite connections are cheap, but shared across threads in WAL mode)
  7. Run PRAGMA optimize periodically (or at connection close)
  8. VACUUM periodically to reclaim space and defragment
```

---

## 9. libSQL, Turso, and the SQLite Renaissance

```
SQLite is experiencing a renaissance for server-side use:

libSQL (Turso):
  - Open-source fork of SQLite
  - Adds: server mode (over HTTP), replication, ALTER TABLE extensions
  - Turso: hosted libSQL at the edge (like Cloudflare D1)
  - "Push the database to the edge, close to the user"

Cloudflare D1:
  - SQLite at every Cloudflare edge location
  - Reads are local (fast, no round-trip)
  - Writes replicate to all edges

LiteFS (Fly.io):
  - Distributed SQLite using FUSE filesystem
  - Primary writes, replicated to read replicas

Litestream:
  - Continuous WAL replication to S3/GCS/Azure
  - Point-in-time recovery for SQLite
  - Backup SQLite databases like PostgreSQL's pg_basebackup + PITR

The pattern:
  Instead of one big PostgreSQL → many small SQLite databases
  Each user/tenant gets their own .db file
  Scales horizontally by adding more .db files on more machines

Why this works:
  - SQLite is incredibly fast for single-tenant queries (no network hop)
  - Isolation: one tenant's load doesn't affect others
  - Backup: just copy a file
  - Schema migration: can be done per-tenant, incrementally
```

---

## 10. SQLite vs PostgreSQL

```
Feature                 SQLite                          PostgreSQL
────────────────────────────────────────────────────────────────────
Model                   Embedded library                Client-server
File                    Single file                     Directory of files
Write concurrency       One writer at a time            Many concurrent writers
Read concurrency        Many (WAL mode)                 Many (MVCC)
Max database size       281 TB (theoretical)            Unlimited (practical)
Setup                   Zero configuration              Requires installation
Network                 Not supported                   TCP/IP
Replication             None built-in (Litestream)      Streaming + Logical
Extensions              Limited                         Very rich
Type system             Flexible (affinity)             Strict
JSON                    Good (json_* functions)         Excellent (JSONB)
Full-text search        FTS5 (good)                     tsvector/tsquery (good)
Stored procedures       No                              PL/pgSQL, PL/Python, etc.
Triggers                Yes (limited)                   Yes (full-featured)
Rows per second (read)  Millions (in-process, no IPC)   Hundreds of thousands
ACID                    Yes                             Yes
Cost                    Free (public domain)            Free (PostgreSQL license)
```

---

## Key Takeaways

1. **SQLite is an embedded library, not a server.** It's linked into your application and opens a file.
2. **WAL mode is mandatory** for any serious use — it allows concurrent reads + writes.
3. **`PRAGMA foreign_keys = ON`** — foreign keys are OFF by default. Always enable.
4. **INTEGER PRIMARY KEY = alias for rowid** — the fastest possible access path in SQLite.
5. **One writer at a time** — this is the fundamental limitation. Use transactions briefly.
6. **WITHOUT ROWID** tables give InnoDB-style clustered indexes for non-integer PKs.
7. **SQLite is the most tested software on Earth** — 100% branch coverage, billions of test cases.
8. **The SQLite file format is an archival format** — recommended by the US Library of Congress.
9. **libSQL/Turso/Litestream** are making SQLite viable for server-side edge computing.
10. **Don't dismiss SQLite for "real" applications** — many apps don't need the complexity of a server database.

---

← Back to [01-postgresql-deep.md](01-postgresql-deep.md)
