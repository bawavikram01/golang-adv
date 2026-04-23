# 2.1 — Storage Engines: How Databases Store Data on Disk

> You typed `INSERT INTO users VALUES (...)`.  
> What ACTUALLY happened on disk? How did the bits get arranged?  
> This is where we go below the SQL layer and see the machine.

---

## 0. The Fundamental Problem

CPUs operate on data in **registers/cache** (nanoseconds).  
Databases store data on **disk** (milliseconds for HDD, microseconds for SSD).

```
Access latency:
  L1 cache:       ~1 ns
  L2 cache:       ~4 ns
  RAM:            ~100 ns
  SSD random:     ~100,000 ns  (100 μs)
  HDD random:     ~10,000,000 ns  (10 ms)
  Network (LAN):  ~500,000 ns

SSD is 1000x slower than RAM.
HDD is 100,000x slower than RAM.
```

The ENTIRE design of a storage engine is an answer to:  
**"How do we minimize the number of slow disk reads?"**

---

## 1. Pages — The Fundamental Unit

Databases don't read individual rows from disk. They read **pages** (also called blocks).

```
Typical page size:
  PostgreSQL:  8 KB  (compile-time, rarely changed)
  MySQL/InnoDB: 16 KB
  SQLite:      4 KB  (configurable)
  SQL Server:  8 KB
  Oracle:      8 KB  (configurable: 2, 4, 8, 16, 32 KB)
```

Why pages? Because:
1. Disk I/O is expensive per OPERATION, not per byte. Reading 8 KB is almost as fast as reading 1 byte.
2. OS page cache operates on pages (typically 4 KB).
3. SSD internal page size is 4-16 KB. Aligning DB pages to SSD pages avoids write amplification.

### Page Layout (Slotted Page)

Most databases use a **slotted page** design:

```
┌────────────────────────────────────────────────────┐
│                   PAGE HEADER                       │
│  - Page ID (block number)                          │
│  - LSN (Log Sequence Number for recovery)          │
│  - Checksum                                        │
│  - Free space pointer                              │
│  - Number of slots                                 │
│  - Special/flags                                   │
├────────────────────────────────────────────────────┤
│    SLOT ARRAY (grows downward →)                   │
│    [offset₁ | offset₂ | offset₃ | ...]            │
│                                                    │
│              ↓ free space ↓                         │
│                                                    │
│    [tuple₃ data] [tuple₂ data] [tuple₁ data]      │
│    ← TUPLE DATA (grows upward)                     │
├────────────────────────────────────────────────────┤
│                 SPECIAL AREA                        │
│  (B-tree: sibling pointers, etc.)                  │
└────────────────────────────────────────────────────┘
```

**Why slotted pages?**
- The slot array provides **indirection**: other pages/indexes point to (page_id, slot_number), NOT a byte offset.
- When tuples are reorganized within a page (compaction, resizing), only the slot array offsets change. External pointers remain valid.
- Tuples can vary in size.

### PostgreSQL Tuple Layout (HeapTupleHeaderData)

```
Each row on disk has a header (~23 bytes in PostgreSQL):

┌──────────────────────────────────┐
│ t_xmin     (4 bytes)  — Transaction ID that inserted this tuple    │
│ t_xmax     (4 bytes)  — Transaction ID that deleted/updated it     │
│ t_cid      (4 bytes)  — Command ID within the transaction          │
│ t_ctid     (6 bytes)  — Current TID (page, offset) — points to    │
│                          newer version if updated                   │
│ t_infomask (2 bytes)  — Status flags (committed? aborted? etc.)    │
│ t_infomask2(2 bytes)  — Number of attributes + more flags          │
│ t_hoff     (1 byte)   — Offset to user data                       │
│ null bitmap (variable) — Which columns are NULL                    │
├──────────────────────────────────┤
│ PADDING (alignment to 8 bytes)                                     │
├──────────────────────────────────┤
│ USER DATA (actual column values)                                   │
└──────────────────────────────────┘
```

**Key insight:** Every row carries its own MVCC metadata (xmin, xmax). That's ~23 bytes of overhead PER ROW. A table with 1 billion rows has ~23 GB of just headers.

---

## 2. Heap Files

A **heap file** is an unordered collection of pages. New rows go wherever there's free space.

```
Heap file organization:
  Page 0: [header][row][row][row][free space]
  Page 1: [header][row][row][free space]
  Page 2: [header][row][row][row][row][free space]
  ...
  Page N: [header][row][free space]
```

PostgreSQL and MySQL/InnoDB both store table data in heap files, but:
- **PostgreSQL**: The heap is the primary storage. Indexes point to (page, offset) in the heap.
- **MySQL/InnoDB**: The table IS a B+ tree clustered by primary key. There's no separate heap.

### Free Space Map (FSM)

How does the database know which page has room for a new row?

PostgreSQL maintains a **Free Space Map** — a separate file that tracks available space per page. When you INSERT, it consults the FSM to find a page with enough room.

```
FSM structure (simplified):
  Page 0: ~7800 bytes free
  Page 1: ~2100 bytes free
  Page 2: ~0 bytes free (full)
  Page 3: ~8000 bytes free (empty)
  ...

INSERT a 200-byte row → finds Page 1 (enough space)
```

### Visibility Map (VM)

PostgreSQL also has a **Visibility Map** — 2 bits per heap page:
- Bit 1: "all tuples on this page are visible to all transactions" (all-visible)
- Bit 2: "all tuples are also frozen" (no transaction ID wraparound needed)

**Why it matters:**
- Index-only scans check the VM. If a page is all-visible, no need to fetch the heap page to check tuple visibility.
- VACUUM can skip all-visible pages.

---

## 3. B-Tree — The King of Indexes

### Why B-Trees?

The B-tree is the single most important data structure in databases. Nearly every index you create is a B-tree (technically B+ tree).

**The problem:** Given millions of rows on disk, find rows where `salary > 100000` without scanning every page.

**Binary search trees** work in memory but fail on disk:
- Each node is one disk read
- A BST with 1M nodes has height ~20 → 20 disk reads per lookup
- Each read is 100μs (SSD) → 2ms per lookup. Too slow.

**B-trees** solve this by having **thousands** of keys per node:
- Each node = one page (8-16 KB)
- With 500 keys per node, height of tree for 1 billion rows ≈ 3-4
- 3-4 disk reads per lookup → 300-400μs. Fast enough.

### B-Tree vs B+ Tree

```
B-TREE:
  - Keys AND values stored in ALL nodes (internal + leaf)
  - Less common in databases

B+ TREE (what databases actually use):
  - Internal nodes store ONLY keys and child pointers (for navigation)
  - ALL data lives in LEAF nodes
  - Leaf nodes are LINKED (doubly-linked list)
```

Why B+ trees win:
1. Internal nodes hold MORE keys (no data pointers taking space) → shorter tree
2. Leaf nodes form a linked list → efficient range scans
3. All lookups have same depth (consistent performance)

### B+ Tree Structure

```
                        ┌────────────────────────┐
              ┌─────────│    [30 | 60 | 90]      │──────────┐
              │         └────────────────────────┘          │
              ▼                    ▼                         ▼
    ┌──────────────┐   ┌──────────────────┐      ┌──────────────┐
    │ [10|15|20|25]│   │ [35|40|45|50|55] │      │[65|70|75|80] │
    └──┬───────────┘   └──┬──────────────┘      └──┬───────────┘
       ▼                   ▼                         ▼
  ┌─────────┐        ┌──────────┐             ┌──────────┐
  │ leaf pg │ ←──→   │ leaf pg  │  ←──→       │ leaf pg  │
  │ data... │        │ data...  │             │ data...  │
  └─────────┘        └──────────┘             └──────────┘
                     leaf pages linked ←──→

INTERNAL NODE (page):
  ┌───────────────────────────────────────────┐
  │ ptr₀ | key₁ | ptr₁ | key₂ | ptr₂ | ...  │
  └───────────────────────────────────────────┘
  ptr₀ points to subtree with all keys < key₁
  ptr₁ points to subtree with key₁ ≤ keys < key₂

LEAF NODE (page):
  ┌────────────────────────────────────────────────────┐
  │ key₁|TID₁ | key₂|TID₂ | key₃|TID₃ | ... | →next  │
  └────────────────────────────────────────────────────┘
  TID = (page_number, offset) pointing to the heap tuple
```

### B+ Tree Operations

#### Lookup: `WHERE id = 42`
```
1. Start at root page (always cached in buffer pool)
2. Binary search within page: 42 is between 30 and 60 → follow ptr₁
3. Read internal page, binary search: 42 is between 40 and 45 → follow pointer
4. Read leaf page, binary search: find key=42, get TID=(page 15, slot 3)
5. Read heap page 15, slot 3 → return the row

Total: ~3-4 page reads (root usually cached → 2-3 actual disk reads)
```

#### Range Scan: `WHERE id BETWEEN 30 AND 70`
```
1. Descend to leaf containing key=30 (same as point lookup)
2. Scan RIGHT through leaf pages following →next pointers
3. Stop when key > 70

This is why leaf linking matters — range scans are sequential reads.
```

#### Insert
```
1. Find correct leaf page
2. If leaf has room: insert key in sorted position. Done.
3. If leaf is FULL: SPLIT
   a. Create new leaf page
   b. Move upper half of keys to new page
   c. Insert median key into parent (as separator)
   d. If parent is also full: split parent (propagates up)
   e. If root splits: create new root (tree grows taller)
```

#### Delete
```
1. Find key in leaf page, remove it
2. If leaf is less than half full: MERGE or REDISTRIBUTE
   a. Try to merge with sibling (combine two pages into one)
   b. If merge impossible, redistribute keys from sibling
   c. Update parent separator key
   d. If parent becomes underfull: propagate up
3. In practice: databases often DON'T merge (lazy deletion)
   — just mark space as free, let future inserts reuse it
   — this is why indexes BLOAT over time (need REINDEX periodically)
```

### B+ Tree Math

```
Given:
  Page size: 8 KB (8192 bytes)
  Page header: 24 bytes
  Key: 8 bytes (BIGINT)
  Pointer: 6 bytes (page number + offset)
  Usable space: ~8168 bytes

Internal node capacity:
  Each (key + pointer) pair = 14 bytes
  Plus one extra pointer = 6 bytes
  Keys per internal node ≈ (8168 - 6) / 14 ≈ 583

Leaf node capacity (with TID pointers):
  Each (key + TID) = 8 + 6 = 14 bytes
  Keys per leaf ≈ 8168 / 14 ≈ 583

For 1 BILLION rows:
  Leaf pages:     1,000,000,000 / 583 ≈ 1,715,266 pages
  Level 2 pages:  1,715,266 / 583 ≈ 2,943 pages
  Level 1 pages:  2,943 / 583 ≈ 6 pages
  Root:           1 page

  HEIGHT = 4 levels → max 4 page reads per lookup
  
  Total index size: ~1,715,266 × 8 KB ≈ 13 GB
```

### Why B+ Trees Are Cache-Friendly

```
Access pattern for a lookup:
  Root page → always in buffer pool (hot page)
  Level 1 pages → probably in buffer pool (few pages)
  Level 2 pages → might be in buffer pool
  Leaf page → might need disk read
  Heap page → likely needs disk read

In practice, for many workloads:
  - Top 2-3 levels are always cached
  - Most lookups need 0-1 actual disk reads
  - Sequential leaf scans are very efficient (adjacent pages)
```

---

## 4. LSM Trees (Log-Structured Merge Trees)

### The Write-Optimized Alternative

B-trees are great for reads but every write requires:
1. Find the page (random read)
2. Modify the page (random write)
3. Write the page back (another random write if WAL + data page)

Random writes are expensive. LSM trees solve this.

### How LSM Trees Work

```
WRITES:
  1. Write to WAL (sequential, fast)
  2. Insert into in-memory sorted structure (MemTable — usually a red-black tree or skip list)
  3. When MemTable is full (~64 MB), FLUSH to disk as a sorted file (SSTable)

READS:
  1. Check MemTable first (most recent data)
  2. Check each SSTable on disk from newest to oldest
  3. Return first match found

COMPACTION:
  SSTables accumulate → too many to search → merge them periodically
```

### Detailed Architecture

```
                    WRITE PATH
                    ──────────
          ┌────────────────────────┐
  WRITE──►│      MemTable         │  (in-memory, sorted — e.g., skip list)
          │  (red-black tree /    │
          │   skip list)          │
          └──────────┬────────────┘
                     │ flush when full
                     ▼
          ┌────────────────────────┐
          │  Immutable MemTable    │  (being flushed to disk)
          └──────────┬────────────┘
                     │
    ─ ─ ─ ─ ─ ─ ─ ─ ▼ ─ ─ ─ ─ ─ ─ ─ ─ ─    DISK
                     
    Level 0 (L0):  [SST-5] [SST-4] [SST-3]    (recently flushed, may overlap)
                        │ compaction
    Level 1 (L1):  [   SST-A   |   SST-B   ]   (non-overlapping, sorted)
                        │ compaction
    Level 2 (L2):  [ SST-X | SST-Y | SST-Z ]   (larger, non-overlapping)
                        │
    Level N:       [        much larger       ]  (10x bigger each level)


                    READ PATH
                    ─────────
  READ──► MemTable → L0 SSTables → L1 → L2 → ... → LN
          (check each level, newest first)
```

### SSTable (Sorted String Table)

```
An SSTable is an immutable, sorted file:

┌─────────────────────────────────────────────────┐
│ DATA BLOCKS                                     │
│  Block 1: [key₁:val₁] [key₂:val₂] ...        │
│  Block 2: [key₅:val₅] [key₆:val₆] ...        │
│  ...                                           │
├─────────────────────────────────────────────────┤
│ INDEX BLOCK                                     │
│  [key₁ → offset of block 1]                   │
│  [key₅ → offset of block 2]                   │
│  ...                                           │
├─────────────────────────────────────────────────┤
│ BLOOM FILTER                                    │
│  (probabilistic: "key might be here" / "definitely not here") │
├─────────────────────────────────────────────────┤
│ FOOTER (metadata, offsets to index + filter)    │
└─────────────────────────────────────────────────┘
```

### Compaction Strategies

```
SIZE-TIERED COMPACTION:
  - When N SSTables of similar size accumulate, merge them into one bigger SSTable
  - Pro: simple, good write throughput
  - Con: temporary space amplification during compaction (need space for input + output)
  - Used by: Cassandra (default), HBase

LEVELED COMPACTION:
  - Each level is 10x the size of the previous
  - L0 SSTables compacted into L1
  - When L1 exceeds size limit, some SSTables promoted to L2 (merged with overlapping L2 files)
  - Pro: bounded space amplification, better read performance
  - Con: more write amplification (data rewritten multiple times)
  - Used by: RocksDB (default), LevelDB

FIFO COMPACTION:
  - Just delete the oldest SSTable when space limit is reached
  - For time-series data where old data can be dropped
```

### Bloom Filters — Making LSM Reads Faster

A Bloom filter is a probabilistic data structure that answers:
"Is this key in this SSTable?" → "MAYBE" or "DEFINITELY NOT"

```
How it works:
  - Bit array of m bits, k hash functions
  - INSERT key: set bits at h₁(key), h₂(key), ..., hₖ(key) to 1
  - LOOKUP key: check bits at h₁(key), ..., hₖ(key)
    - Any bit is 0 → key DEFINITELY NOT present (skip this SSTable!)
    - All bits are 1 → key MIGHT be present (check the SSTable)

False positive rate ≈ (1 - e^(-kn/m))^k
  n = number of keys, m = bits, k = hash functions

With 10 bits per key and 7 hash functions: ~0.8% false positive rate
  → 99.2% of unnecessary SSTable reads are avoided!
```

### B-Tree vs LSM Tree Comparison

| Property | B+ Tree | LSM Tree |
|----------|---------|----------|
| **Write pattern** | Random writes (in-place update) | Sequential writes (append) |
| **Read pattern** | 1 tree traversal | Check multiple levels |
| **Write speed** | Slower (random I/O) | Faster (sequential I/O) |
| **Read speed** | Faster (single lookup) | Slower (multiple lookups) |
| **Space amplification** | Low (~1.5x with fragmentation) | Variable (1x-2x depending on compaction) |
| **Write amplification** | ~2x (WAL + data page) | 10-30x (compaction rewrites) |
| **Predictable latency** | Yes (consistent tree depth) | Less so (compaction spikes) |
| **Compression** | Harder (pages must be updatable) | Easier (SSTables are immutable) |
| **Concurrency** | Fine-grained page locking | Append-only (naturally concurrent) |
| **Used by** | PostgreSQL, MySQL, SQLite, Oracle | RocksDB, LevelDB, Cassandra, HBase |

### When to Use Which

```
B+ Tree (default choice):
  - OLTP with mixed read/write workload
  - Need predictable latency
  - Range scans are common
  - You need strong transactional guarantees

LSM Tree:
  - Write-heavy workloads (logging, IoT, time-series)
  - Write throughput is the bottleneck
  - Data arrives in bursts
  - You can tolerate slightly higher read latency
  - You want better compression (immutable SSTables compress well)
```

---

## 5. Write-Ahead Log (WAL)

### The Durability Problem

```
Scenario: You do INSERT → database writes to buffer pool (in memory) → CRASH!
The data was in memory only → LOST.

Solution: BEFORE modifying the data page, write the change to a LOG on disk.
This is the Write-Ahead Log (WAL), also called:
  - Transaction log (SQL Server)
  - Redo log (MySQL/InnoDB, Oracle)
  - Journal (MongoDB's WiredTiger)
```

### WAL Protocol (The Rule)

**"No page can be written to disk until the corresponding WAL record has been flushed to disk."**

This guarantees:
- If the database crashes, it can replay the WAL to reconstruct any changes that were in the buffer pool but not yet written to data files.
- Data files can be updated LAZILY (checkpoint at convenient times).

### WAL Record Structure

```
┌───────────────────────────────────────────────┐
│ LSN       — Log Sequence Number (monotonic)   │
│ PREV_LSN  — Previous LSN (for undo chain)     │
│ TX_ID     — Transaction that made this change  │
│ TYPE      — INSERT / UPDATE / DELETE / COMMIT  │
│ PAGE_ID   — Which page was modified            │
│ OFFSET    — Where on the page                 │
│ LENGTH    — How many bytes changed            │
│ OLD_DATA  — Previous value (for UNDO)         │
│ NEW_DATA  — New value (for REDO)              │
└───────────────────────────────────────────────┘
```

### WAL Lifecycle

```
1. Transaction begins
2. Each modification:
   a. Generate WAL record with old + new values
   b. Append to WAL buffer (in memory)
   c. Modify the data page in buffer pool (in memory)
3. Transaction COMMIT:
   a. WAL records flushed to disk (fsync) → DURABLE
   b. Return success to client
   c. Data pages still only in buffer pool (dirty)
4. Later: checkpoint writes dirty pages to disk
5. WAL segments before checkpoint can be recycled

Key insight: COMMIT only waits for WAL write, NOT for data page write.
WAL write is SEQUENTIAL → fast (even on HDD: ~100 MB/s sequential).
Data page write would be RANDOM → slow.
```

### Group Commit

```
Problem: fsync after every single transaction is expensive.
Solution: batch multiple transactions' WAL records into one fsync.

  TX₁ writes WAL → 
  TX₂ writes WAL →  → [ONE fsync for all three] → all committed
  TX₃ writes WAL → 

PostgreSQL: commit_delay + commit_siblings parameters
This can turn 1000 individual fsyncs into 100 group fsyncs → 10x throughput improvement.
```

---

## 6. Buffer Pool (Buffer Manager)

The buffer pool is the **most critical** component for performance. It's the in-memory cache of disk pages.

### How It Works

```
┌─────────────────────────────────────────────┐
│                BUFFER POOL                   │
│  (shared memory region, e.g., 8 GB)        │
│                                             │
│  Frame 0: [page from orders, block 42]      │
│  Frame 1: [page from users, block 7]        │
│  Frame 2: [page from idx_user_email, blk 3] │
│  Frame 3: [EMPTY]                           │
│  Frame 4: [page from orders, block 43]      │
│  ...                                        │
│  Frame N: [page from products, block 100]   │
│                                             │
│  PAGE TABLE (hash map):                     │
│    (orders, 42) → Frame 0                   │
│    (users, 7)   → Frame 1                   │
│    ...                                      │
│                                             │
│  Each frame has:                            │
│    - pin_count: # of queries using this page│
│    - dirty_flag: modified since read?       │
│    - ref_bit: recently accessed? (for clock)│
└─────────────────────────────────────────────┘
```

### Page Request Flow

```
Query needs page (orders, block 42):

1. Check page table: is (orders, 42) in buffer pool?
   YES → increment pin_count, return pointer. DONE. (buffer hit)
   NO  → continue (buffer miss)

2. Find a victim frame:
   a. Find unpinned frame (pin_count = 0) using replacement policy
   b. If victim is DIRTY: write to disk first (flush)
   c. Evict victim from page table

3. Read requested page from disk into the frame.
4. Update page table: (orders, 42) → this frame.
5. pin_count = 1, return pointer.

When query is done with the page: UNPIN it (decrement pin_count).
```

### Page Replacement Policies

```
LRU (Least Recently Used):
  - Evict the page that hasn't been accessed for the longest time
  - Simple, but vulnerable to "sequential flooding":
    A full table scan reads every page once → flushes the entire buffer pool
    Useful pages for OLTP queries evicted by a single big scan

LRU-K (K=2 or more):
  - Track the last K access times
  - Evict page with oldest K-th access
  - Better than LRU at distinguishing "used twice" vs "used once"

CLOCK (approximation of LRU):
  - Circular buffer of frames, each with a "reference bit"
  - Access sets ref_bit = 1
  - Clock hand sweeps: if ref_bit = 1, set to 0 and move on
                        if ref_bit = 0, this is the victim
  - PostgreSQL uses Clock Sweep (enhanced version)
  - Much cheaper than maintaining a full LRU list (no list maintenance per access)

2Q (Two Queue):
  - Two queues: A1 (short-term, FIFO) and Am (long-term, LRU)
  - New page enters A1
  - If accessed again while in A1, move to Am
  - Protects long-term working set from scan pollution

ARC (Adaptive Replacement Cache):
  - Two LRU lists + two ghost lists
  - Self-tuning: adapts balance between "recency" and "frequency"
  - Used by ZFS, IBM DB2
  - Patented (one reason PostgreSQL uses Clock Sweep instead)
```

### PostgreSQL's Buffer Pool

```sql
-- PostgreSQL shared_buffers setting
-- Default: 128 MB (WAY too low for production)
-- Recommendation: 25% of total RAM (up to ~8-16 GB)
-- OS page cache handles the rest

SHOW shared_buffers;  -- '128MB'
-- Set in postgresql.conf: shared_buffers = '4GB'

-- Check buffer pool hit ratio:
SELECT
    sum(heap_blks_read) AS heap_read,
    sum(heap_blks_hit) AS heap_hit,
    ROUND(sum(heap_blks_hit) / NULLIF(sum(heap_blks_hit) + sum(heap_blks_read), 0) * 100, 2) AS hit_ratio
FROM pg_statio_user_tables;

-- Good: > 99% hit ratio
-- Bad: < 95% → need more shared_buffers or fewer tables
```

---

## 7. Row Store vs Column Store

### Row-Oriented Storage (NSM — N-ary Storage Model)

```
How most OLTP databases store data:

Page contains ROWS:
┌────────────────────────────────────────────────────────┐
│ [id=1, name="Alice", age=30, salary=150000]            │
│ [id=2, name="Bob",   age=25, salary=130000]            │
│ [id=3, name="Carol", age=28, salary=95000]             │
│ ...                                                    │
└────────────────────────────────────────────────────────┘

GOOD FOR:
  - SELECT * FROM employee WHERE id = 42  (one page read gets everything)
  - INSERT a full row (one page write)
  - OLTP: access patterns are typically "give me all columns of a few rows"

BAD FOR:
  - SELECT AVG(salary) FROM employee  (must read ALL columns of EVERY row  
    to get one column — huge wasted I/O)
```

### Column-Oriented Storage (DSM — Decomposition Storage Model)

```
How OLAP databases store data:

Each COLUMN in its own file/pages:

id column:     [1, 2, 3, 4, 5, 6, 7, ...]
name column:   ["Alice", "Bob", "Carol", ...]
age column:    [30, 25, 28, 32, ...]
salary column: [150000, 130000, 95000, ...]

GOOD FOR:
  - SELECT AVG(salary) FROM employee
    → reads ONLY the salary column file. 4 bytes × N rows vs 100+ bytes × N rows.
    → 25x less I/O!
  - Compression: same data type in sequence → excellent compression
    (run-length encoding, dictionary encoding, delta encoding)
  - Vectorized execution: process 1000 values at once using SIMD

BAD FOR:
  - SELECT * FROM employee WHERE id = 42
    → must read from ALL column files and stitch together. Slow.
  - Single row INSERT: must write to ALL column files.
```

### Column Store Compression Techniques

```
1. RUN-LENGTH ENCODING (RLE):
   Original: [USA, USA, USA, USA, CAN, CAN, UK, UK, UK]
   Encoded:  [(USA, 4), (CAN, 2), (UK, 3)]
   Best when: sorted data with many consecutive repeats

2. DICTIONARY ENCODING:
   Original: ["Engineering", "Sales", "Engineering", "Marketing", "Sales"]
   Dictionary: {0: "Engineering", 1: "Sales", 2: "Marketing"}
   Encoded:    [0, 1, 0, 2, 1]
   Each value: 2 bits instead of ~11 bytes. ~44x compression.

3. DELTA ENCODING:
   Original:  [1000, 1005, 1003, 1010, 1008]
   Encoded:   [1000, +5, -2, +7, -2]
   Deltas are small → fewer bits needed

4. BIT-PACKING:
   If values range 0-15, only need 4 bits per value (not 32 or 64)
   Pack 16 values into one 64-bit word

5. FRAME OF REFERENCE (FOR):
   If values are [10032, 10035, 10028, 10041]:
   Base = 10028, store offsets: [4, 7, 0, 13] (need only 4 bits each)
```

### Databases by Storage Model

```
Row stores (OLTP):         Column stores (OLAP):
  PostgreSQL                 ClickHouse
  MySQL                      Apache Druid
  Oracle                     Amazon Redshift
  SQL Server (default)       Google BigQuery
  SQLite                     DuckDB
                             Apache Parquet (file format)
                             Vertica
                             SQL Server (columnstore indexes)
```

---

## 8. TOAST — The Oversized-Attribute Storage Technique (PostgreSQL)

```
Problem: a page is 8 KB. What if a row has a 1 MB text column?

Answer: TOAST (The Oversized-Attribute Storage Technique)

PostgreSQL automatically:
1. COMPRESSES the value (using LZ compression, ~pglz or lz4)
2. If still too big, SLICES it into chunks stored in a separate TOAST table
3. The main heap row stores a TOAST pointer (18 bytes) instead of the actual data

Thresholds:
  - If row > ~2 KB: try compression
  - If compressed row still > ~2 KB: move to TOAST table

TOAST strategies per column:
  PLAIN    — no TOAST (for small fixed-size types: int, float)
  EXTENDED — compress then external storage (default for text, bytea)
  EXTERNAL — external storage without compression (for pre-compressed data)
  MAIN     — try compression, avoid external if possible

-- Check TOAST settings:
SELECT attname, attstorage FROM pg_attribute WHERE attrelid = 'employee'::regclass;
-- p = PLAIN, x = EXTENDED, e = EXTERNAL, m = MAIN
```

---

## 9. Copy-on-Write B-Trees

An alternative to WAL-based B-trees, used by LMDB and BoltDB.

```
Traditional B-tree: modify pages in-place + WAL for recovery
Copy-on-Write B-tree: NEVER modify existing pages

UPDATE flow:
1. Copy the leaf page, apply modification to the copy
2. Copy the parent page, update pointer to new leaf copy
3. Copy all ancestors up to the root
4. Atomically swap root pointer to new root copy

Old pages become garbage → collected later.

PROS:
  - No WAL needed (atomic root pointer swap)
  - Readers always see a consistent snapshot (MVCC for free)
  - Never corrupts data (old pages untouched)

CONS:
  - Write amplification: modify 1 leaf → rewrite entire root-to-leaf path
  - Fragmentation: new pages written to different locations (not sequential)
  - More disk writes per update

Used by: LMDB, BoltDB, SQLite (WAL mode is different but related concept)
```

---

## 10. Practice & Exploration

### Exercise 1: Observe Page Layout (PostgreSQL)

```sql
-- Install pageinspect extension
CREATE EXTENSION pageinspect;

-- Create test table
CREATE TABLE test_pages (id serial, data text);
INSERT INTO test_pages SELECT g, repeat('x', 100) FROM generate_series(1, 1000) g;

-- View page header
SELECT * FROM page_header(get_raw_page('test_pages', 0));

-- View tuples on a page
SELECT lp, lp_off, lp_len, t_xmin, t_xmax, t_ctid
FROM heap_page_items(get_raw_page('test_pages', 0));

-- View B-tree index structure
CREATE INDEX idx_test_id ON test_pages(id);
SELECT * FROM bt_metap('idx_test_id');     -- B-tree metadata
SELECT * FROM bt_page_stats('idx_test_id', 1);  -- page stats
SELECT * FROM bt_page_items('idx_test_id', 1);  -- actual index entries
```

### Exercise 2: Calculate B-Tree Height

Given: table with 100 million rows, indexed BIGINT column, 8 KB pages.

<details><summary>Solution</summary>

```
Key size: 8 bytes (BIGINT)
TID: 6 bytes
Internal pointer: 4 bytes (page number, 32-bit)
Page: 8192 bytes, ~24 byte header, usable ~8168 bytes

Leaf: each entry = 8 (key) + 6 (TID) + overhead ≈ 16 bytes
  → ~510 entries per leaf page
  → 100,000,000 / 510 ≈ 196,078 leaf pages

Internal: each entry = 8 (key) + 4 (pointer) + overhead ≈ 14 bytes
  → ~583 entries per internal page

Level 2: 196,078 / 583 ≈ 337 pages
Level 1: 337 / 583 ≈ 1 page
Root: 1 page

HEIGHT = 3 (root → internal → leaf)
3 page reads per point lookup (root usually cached → 2 reads)

Index size: ~196,078 pages × 8 KB ≈ 1.5 GB
```
</details>

---

## Key Takeaways

1. **Pages are the atomic unit** of database I/O. Everything is designed around minimizing page reads.
2. **B+ trees** give O(log_B N) lookups with B ≈ 500 → 3-4 disk reads for billions of rows.
3. **LSM trees** trade read performance for write performance (sequential writes, compaction).
4. **WAL** guarantees durability: log first, data later. Commit = WAL flush, not data flush.
5. **Buffer pool** is the heart of performance. Cache hit ratio > 99% is the goal.
6. **Row stores** for OLTP (few rows, all columns). **Column stores** for OLAP (all rows, few columns).
7. The choice of storage engine shapes EVERYTHING about the database's performance characteristics.

---

Next: [02-indexing-deep-dive.md](02-indexing-deep-dive.md) →
