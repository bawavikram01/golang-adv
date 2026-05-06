# 6.2 — Reading Database Source Code

> Reading source code is the highest-leverage activity for understanding databases.  
> Papers tell you the theory. Source code tells you the truth.  
> The gap between them is where real engineering lives.

---

## 1. SQLite — The Perfect Starting Point

```
Why start with SQLite:
  - Single C file (amalgamation): sqlite3.c (~250K lines, but well-structured)
  - No server, no threads, no network — pure database logic
  - Beautifully documented (extensive comments in source)
  - Full SQL database with B-tree, pager, VDBE, query planner
  - Used everywhere: every phone, every browser, every OS

Architecture:
┌─────────────────────────────────────────────────────┐
│  SQL Interface                                       │
│  ┌────────────┐ ┌───────────┐ ┌──────────────────┐  │
│  │ Tokenizer  │→│  Parser   │→│ Code Generator   │  │
│  │ (tokenize.c)│ │(parse.y)  │ │(select.c,        │  │
│  └────────────┘ └───────────┘ │ where.c, etc.)   │  │
│                                └──────────────────┘  │
├─────────────────────────────────────────────────────┤
│  Virtual Machine (VDBE)                              │
│  ┌──────────────────────────────────────────────┐   │
│  │ Bytecode interpreter — like a CPU             │   │
│  │ Opcodes: Column, Seek, Next, Insert, Sort... │   │
│  │ (vdbe.c, vdbeaux.c)                          │   │
│  └──────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────┤
│  B-tree Layer                                        │
│  ┌──────────────────────────────────────────────┐   │
│  │ B-tree for tables (rowid → data)              │   │
│  │ B-tree for indexes (key → rowid)              │   │
│  │ (btree.c — ~10K lines, THE file to read)      │   │
│  └──────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────┤
│  Pager Layer                                         │
│  ┌──────────────────────────────────────────────┐   │
│  │ Page cache, WAL, locking, crash recovery      │   │
│  │ (pager.c, wal.c)                              │   │
│  └──────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────┤
│  OS Interface (VFS)                                  │
│  ┌──────────────────────────────────────────────┐   │
│  │ Abstract file I/O (works on any OS)           │   │
│  │ (os_unix.c, os_win.c)                         │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘

Key files to read (in order):
  1. sqlite.h       — Public API (understand the interface first)
  2. btree.c        — B-tree implementation (core data structure)
  3. pager.c        — Page cache and crash recovery
  4. vdbe.c         — Virtual machine (how queries actually execute)
  5. where.c        — Query planner (how it chooses indexes)
  6. select.c       — SELECT statement compilation
  7. wal.c          — Write-ahead log implementation
```

### How to Read SQLite Source

```bash
# Clone the source:
git clone https://github.com/sqlite/sqlite.git
cd sqlite

# Build with debug symbols:
./configure --enable-debug
make

# Use EXPLAIN to see the VDBE bytecode:
sqlite3 test.db
> CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER);
> INSERT INTO users VALUES (1, 'Alice', 30);
> EXPLAIN SELECT name FROM users WHERE age > 25;

addr  opcode         p1    p2    p3    p4             p5  comment
----  -------------  ----  ----  ----  -------------  --  -------
0     Init           0     12    0                    0   Start
1     OpenRead       0     2     0     3              0   users
2     Rewind         0     10    0                    0
3       Column         0     2     1                    0   users.age
4       Le             2     9     1     (BINARY)       0x54
5       Column         0     1     3                    0   users.name
6       ResultRow      3     1     0                    0   output row
7     Next           0     3     0                    1
8     Halt           0     0     0                    0
...

# Each opcode is a case in the giant switch in vdbe.c
# Open vdbe.c, search for "case OP_Column:" to see how Column works
# This is how SELECT actually executes — instruction by instruction
```

---

## 2. PostgreSQL Source Code

```
PostgreSQL: ~1.5M lines of C. The reference implementation of a relational database.

Directory structure (key directories):
  src/
  ├── backend/           # THE important directory
  │   ├── access/        # Storage access methods
  │   │   ├── heap/      # Heap tables (heap_insert, heap_update, etc.)
  │   │   ├── nbtree/    # B-tree index implementation
  │   │   ├── hash/      # Hash index
  │   │   ├── gin/       # GIN (inverted) index
  │   │   ├── gist/      # GiST index
  │   │   └── transam/   # Transaction manager, WAL, MVCC
  │   ├── catalog/       # System catalog (metadata tables)
  │   ├── commands/      # DDL command implementations
  │   ├── executor/      # Query executor (Volcano model)
  │   │   ├── nodeSeqscan.c
  │   │   ├── nodeIndexscan.c
  │   │   ├── nodeHashjoin.c
  │   │   ├── nodeSort.c
  │   │   └── nodeAgg.c
  │   ├── optimizer/     # Query optimizer
  │   │   ├── plan/      # Plan generation
  │   │   ├── path/      # Path generation (access paths)
  │   │   └── util/      # Cost estimation, statistics
  │   ├── parser/        # SQL parser (gram.y — yacc grammar)
  │   ├── postmaster/    # Process manager (one process per connection)
  │   ├── replication/   # Streaming replication, logical replication
  │   ├── storage/       # Buffer pool, lock manager, shared memory
  │   │   ├── buffer/    # Buffer pool manager ★
  │   │   ├── lmgr/      # Lock manager
  │   │   └── smgr/      # Storage manager (file I/O)
  │   └── utils/        # Caches, memory contexts, support
  ├── include/           # Header files (start here for type definitions)
  └── bin/               # Client tools (psql, pg_dump, etc.)

Reading order for understanding core:
  1. src/include/storage/buf_internals.h  — Buffer pool structures
  2. src/backend/storage/buffer/bufmgr.c  — Buffer pool manager
  3. src/include/access/htup_details.h    — Heap tuple layout (xmin/xmax/etc.)
  4. src/backend/access/heap/heapam.c     — Heap access method
  5. src/backend/access/nbtree/nbtinsert.c — B-tree insert
  6. src/backend/access/transam/xlog.c    — WAL implementation
  7. src/backend/executor/execMain.c      — Executor entry point
  8. src/backend/optimizer/plan/planner.c — Query planner entry
```

### Key PostgreSQL Source Code Gems

```c
// === HeapTupleSatisfiesMVCC (src/backend/access/heap/heapam_visibility.c) ===
// THIS is where MVCC visibility is decided.
// For every tuple access, PostgreSQL calls this to check:
//   Is this tuple visible to the current snapshot?

// Simplified logic:
bool HeapTupleSatisfiesMVCC(HeapTuple htup, Snapshot snapshot) {
    // Check xmin (the transaction that created this tuple):
    if (xmin == current_txn)
        // We created it — visible if not deleted by us
        return (xmax == 0 || xmax == current_txn);
    
    if (!TransactionIdDidCommit(xmin))
        return false;  // Creator never committed → invisible
    
    if (XidInMVCCSnapshot(xmin, snapshot))
        return false;  // Creator was in-progress when snapshot taken → invisible
    
    // xmin is committed and was committed before our snapshot → visible
    // Now check xmax (the transaction that deleted this tuple):
    if (xmax == 0)
        return true;   // Not deleted → visible
    
    if (xmax == current_txn)
        return false;  // We deleted it → invisible
    
    if (!TransactionIdDidCommit(xmax))
        return true;   // Deleter didn't commit → still visible
    
    if (XidInMVCCSnapshot(xmax, snapshot))
        return true;   // Deleter was in-progress when snapshot taken → still visible
    
    return false;       // Deleted and committed before our snapshot → invisible
}

// === Buffer Pool (src/backend/storage/buffer/bufmgr.c) ===
// ReadBuffer() → main entry point for fetching a page
// GetVictimBuffer() → clock sweep eviction
// BufferAlloc() → reserve a buffer slot, possibly evicting

// === B-tree Insert (src/backend/access/nbtree/nbtinsert.c) ===
// _bt_doinsert() → entry point for B-tree insertion
// _bt_insertonpg() → insert into a specific page
// _bt_split() → split an overfull page

// === WAL (src/backend/access/transam/xlog.c) ===
// XLogInsert() → append a WAL record
// XLogFlush() → force WAL to disk (called at COMMIT)
// StartupXLOG() → crash recovery entry point (ARIES analysis+redo+undo)
```

```bash
# Clone and explore:
git clone https://github.com/postgres/postgres.git
cd postgres

# Build with debug:
./configure --enable-debug --enable-cassert CFLAGS="-O0 -g"
make -j$(nproc)
make install

# Use GDB to trace a query:
gdb --args ./src/backend/postgres --single -D /tmp/pgdata mydatabase
(gdb) break ExecutorRun
(gdb) run
backend> SELECT * FROM users WHERE id = 1;
# Now step through the executor code

# Or use perf to see where time is spent:
perf record -g -p $(pgrep postgres) -- sleep 10
perf report
```

---

## 3. RocksDB Source Code (C++)

```
RocksDB: The definitive LSM-tree implementation.
Used inside: MySQL (MyRocks), CockroachDB, TiKV, Kafka, many more.

Key files:
  db/
  ├── db_impl/
  │   ├── db_impl.cc           # Main database implementation
  │   ├── db_impl_write.cc     # Write path
  │   └── db_impl_compaction_flush.cc  # Compaction + flush
  ├── memtable/
  │   └── skiplist.h           # SkipList (default memtable structure)
  ├── table/
  │   ├── block_based/
  │   │   └── block_based_table_reader.cc  # SST file reader
  │   └── format.cc           # SST file format
  └── db/
      ├── write_batch.cc       # Atomic batch writes
      ├── version_set.cc       # Version management (which SST files are live)
      └── compaction/
          └── compaction_job.cc  # Compaction logic

Write path:
  1. Write to WAL (write_batch → write to log file)
  2. Insert into MemTable (in-memory skip list)
  3. When MemTable full → flush to SST file (Level 0)
  4. Compaction: merge Level N SSTs → Level N+1 (background thread)

Read path:
  1. Check MemTable (most recent writes)
  2. Check immutable MemTables (being flushed)
  3. Check Level 0 SSTs (may overlap)
  4. Check Level 1+ SSTs (binary search by key range)
  5. Bloom filters skip SSTs that definitely don't have the key
```

---

## 4. DuckDB Source Code (C++)

```
DuckDB: Modern analytical (OLAP) database. "SQLite for analytics."
Clean C++ codebase, excellent for learning vectorized execution.

Key directories:
  src/
  ├── catalog/          # Schema management
  ├── execution/        # Execution engine
  │   ├── operator/     # Physical operators
  │   │   ├── scan/     # Table scan operators
  │   │   ├── join/     # Join operators (hash join, etc.)
  │   │   └── aggregate/ # Aggregation operators
  │   └── expression_executor.cpp  # Expression evaluation
  ├── optimizer/        # Query optimizer
  │   ├── join_order/   # Join order optimization
  │   └── rule/         # Optimization rules
  ├── parser/           # SQL parser
  ├── planner/          # Logical plan generation
  ├── storage/          # Storage engine
  │   ├── buffer/       # Buffer manager
  │   └── table/        # Row groups, column data
  └── common/
      └── vector_operations/  # Vectorized operations ★

What makes DuckDB special to read:
  - Vectorized execution (processes batches of ~2048 values, not one row at a time)
  - Column-oriented storage (data stored by column, not row)
  - Modern C++ (compared to PostgreSQL's C)
  - Push-based execution (operators push data to parent, vs Volcano's pull)
  
Vectorized execution example:
  Instead of: for each row → evaluate age > 25 → if true, output
  DuckDB:     take 2048 ages → SIMD compare all vs 25 → selection vector
              Much faster: CPU cache friendly, branch-prediction friendly
```

---

## 5. CockroachDB Source Code (Go)

```
CockroachDB: Distributed SQL database. Go codebase.
Excellent for understanding distributed transactions.

Key packages:
  pkg/
  ├── kv/              # Key-value layer
  │   ├── kvserver/    # Raft-based replicated storage
  │   │   ├── store.go           # Store (manages ranges on a node)
  │   │   ├── replica.go         # Replica (one Raft group member)
  │   │   ├── replica_raft.go    # Raft integration
  │   │   └── replica_command.go # Command application
  │   └── kvclient/   # Client-side of KV layer
  ├── sql/             # SQL layer
  │   ├── parser/      # SQL parser
  │   ├── sem/         # Semantic analysis
  │   ├── opt/         # Cost-based query optimizer (Cascades framework)
  │   ├── execinfra/   # Distributed execution infrastructure
  │   └── colexec/     # Vectorized column execution
  ├── storage/         # Storage engine (wraps Pebble, a RocksDB-like LSM)
  └── server/          # Node server, gRPC endpoints

What to focus on:
  1. pkg/kv/kvserver/replica_raft.go → How Raft consensus works in practice
  2. pkg/sql/opt/ → Cascades optimizer (modern, research-grade optimizer)
  3. pkg/kv/kvserver/txnwait/ → Distributed transaction conflict resolution
  4. pkg/roachpb/api.proto → The KV API (protobuf definitions)
```

---

## 6. TiKV Source Code (Rust)

```
TiKV: Distributed key-value store. Rust.
Used as the storage layer for TiDB (distributed MySQL).

Key directories:
  src/
  ├── server/        # gRPC server, coprocessor
  ├── storage/       # Transaction layer
  │   ├── mvcc/      # MVCC implementation (Percolator protocol)
  │   └── txn/       # Transaction logic
  ├── raftstore/     # Raft consensus + region management
  │   ├── store/     # Store manages multiple Raft groups
  │   └── peer.rs    # One Raft peer (one region replica)
  └── coprocessor/   # Push-down computation (execute WHERE on storage nodes)

What to focus on:
  1. src/storage/mvcc/ → Percolator-based distributed transactions
     (Two-phase commit over multiple Raft groups)
  2. src/raftstore/ → Multi-Raft (one Raft group per data range)
  3. Excellent Rust code — learn both database internals and Rust
```

---

## 7. How to Read Source Code Effectively

```
Strategy: Don't read linearly. Trace execution paths.

1. Start with the API:
   - What functions does the user call?
   - SQLite: sqlite3_exec(), sqlite3_prepare_v2()
   - PostgreSQL: exec_simple_query() in postgres.c

2. Trace a simple query end-to-end:
   - "SELECT * FROM t WHERE id = 1"
   - Follow the code from SQL text → parse → plan → execute → return
   - Use GDB breakpoints or printf debugging

3. Focus on one subsystem at a time:
   Week 1: Buffer pool (how pages are cached)
   Week 2: B-tree (how indexes work)
   Week 3: WAL (how crash recovery works)
   Week 4: Executor (how queries run)
   Week 5: Optimizer (how plans are chosen)

4. Read the comments and commit history:
   git log --oneline --all --follow -- src/backend/storage/buffer/bufmgr.c
   # See how the buffer manager evolved over 25+ years

5. Write notes as you read:
   - Draw diagrams of data structures
   - Document function call chains
   - Note where theory diverges from practice

6. Modify and observe:
   - Add printf/logging to trace execution
   - Change a constant and see what breaks
   - Break something on purpose and debug it

Reading order across projects:
  1. SQLite (simplest, best documented)
  2. DuckDB (modern C++, clean architecture)
  3. PostgreSQL (reference implementation, rich in detail)
  4. RocksDB (if interested in LSM trees)
  5. CockroachDB or TiKV (if interested in distributed systems)
```

---

## Key Takeaways

1. **Start with SQLite.** It's the most readable database source code in existence. One C file, beautifully commented, full SQL database. Use EXPLAIN to see the VDBE bytecode, then read vdbe.c.
2. **PostgreSQL's MVCC is in `heapam_visibility.c`.** This single file implements the visibility rules that make transactions work. Read it and you understand MVCC at the deepest level.
3. **Trace, don't read linearly.** Pick a simple query, set breakpoints, and follow execution from SQL text through parser → planner → executor → storage → disk and back.
4. **DuckDB** is the best codebase for understanding modern OLAP: vectorized execution, columnar storage, push-based pipelines. Clean modern C++.
5. **For distributed systems, read CockroachDB (Go) or TiKV (Rust).** They implement Raft consensus, distributed transactions, and range-partitioned storage — the full distributed database stack.

---

Next: [03-research-papers.md](03-research-papers.md) →
