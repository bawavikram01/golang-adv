# 6.1 — Build Your Own Database

> You don't truly understand a database until you've built one.  
> Not a toy. A real one — with pages, a buffer pool, a WAL, B+ trees,  
> transactions, and a query engine.  
> This is the crucible. This is where gods are forged.

---

## 1. Architecture of a Database System

```
Every database — from SQLite to Spanner — has these layers:

┌─────────────────────────────────────────────────────────────┐
│                      SQL Interface                          │
│  Parser → AST → Semantic Analysis → Logical Plan            │
├─────────────────────────────────────────────────────────────┤
│                     Query Optimizer                          │
│  Logical Plan → Physical Plan (cost-based or rule-based)    │
├─────────────────────────────────────────────────────────────┤
│                     Execution Engine                         │
│  Volcano/Iterator model or Vectorized/Push-based            │
├─────────────────────────────────────────────────────────────┤
│                   Concurrency Control                        │
│  Locks / MVCC / OCC / Timestamp ordering                    │
├─────────────────────────────────────────────────────────────┤
│                   Buffer Pool Manager                        │
│  Page cache, replacement policy (LRU/Clock), dirty tracking │
├─────────────────────────────────────────────────────────────┤
│                     Storage Engine                           │
│  Disk manager, page layout, B+ tree / LSM-tree              │
├─────────────────────────────────────────────────────────────┤
│                   Recovery Manager (WAL)                     │
│  Write-Ahead Log, ARIES, checkpointing, crash recovery      │
├─────────────────────────────────────────────────────────────┤
│                      Disk / SSD                             │
└─────────────────────────────────────────────────────────────┘

We'll build each layer, bottom-up.
```

---

## 2. Disk Manager & Page Layout

```
Everything starts with PAGES. Databases don't read individual rows.
They read and write fixed-size pages (typically 4KB, 8KB, or 16KB).

Page structure (slotted page — used by PostgreSQL, SQLite, etc.):

┌──────────────────────────────────────────┐  ← Page start
│  Page Header                              │
│  ┌──────────────────────────────────────┐ │
│  │ page_id | num_slots | free_offset    │ │
│  │ prev_page | next_page | lsn          │ │
│  └──────────────────────────────────────┘ │
│                                           │
│  Slot Array (grows downward →)            │
│  ┌─────┬─────┬─────┬─────┬──────────┐   │
│  │ S0  │ S1  │ S2  │ S3  │   ...    │   │
│  │off,l│off,l│off,l│off,l│          │   │
│  └─────┴─────┴─────┴─────┴──────────┘   │
│                                           │
│         ┌── Free Space ──┐                │
│         │                │                │
│  ┌──────┴────────────────┴──────────────┐│
│  │ Tuple Data (grows upward ←)          ││
│  │ [Tuple 3][Tuple 2][Tuple 1][Tuple 0] ││
│  └──────────────────────────────────────┘│
└──────────────────────────────────────────┘  ← Page end

Each slot = (offset, length) pointing to a tuple in the data section.
Slots grow down, tuples grow up. They meet in the middle.
When they meet → page is full.

Tuple layout:
┌──────────┬────────────┬──────────────────────────────┐
│ TupleHdr │ Null bitmap│ Col1 | Col2 | Col3 | ...     │
│ (xmin,   │ (1 bit per │ (fixed-size cols first,      │
│  xmax,   │  column)   │  then variable-length cols   │
│  flags)  │            │  with offset array)           │
└──────────┴────────────┴──────────────────────────────┘
```

### Implementation (Rust)

```rust
// A minimal page implementation:

const PAGE_SIZE: usize = 4096;

#[repr(C)]
struct PageHeader {
    page_id: u32,
    num_slots: u16,
    free_space_offset: u16,  // where free space begins (from start)
    data_start_offset: u16,  // where tuple data begins (from end)
    lsn: u64,                // log sequence number (for WAL)
}

struct SlottedPage {
    data: [u8; PAGE_SIZE],
}

impl SlottedPage {
    fn new(page_id: u32) -> Self {
        let mut page = SlottedPage { data: [0u8; PAGE_SIZE] };
        let header = PageHeader {
            page_id,
            num_slots: 0,
            free_space_offset: std::mem::size_of::<PageHeader>() as u16,
            data_start_offset: PAGE_SIZE as u16,
            lsn: 0,
        };
        // Write header to page
        unsafe {
            let header_bytes = std::slice::from_raw_parts(
                &header as *const PageHeader as *const u8,
                std::mem::size_of::<PageHeader>(),
            );
            page.data[..header_bytes.len()].copy_from_slice(header_bytes);
        }
        page
    }

    fn insert_tuple(&mut self, tuple: &[u8]) -> Option<u16> {
        let header = self.header_mut();
        let slot_size = 4u16; // 2 bytes offset + 2 bytes length
        let needed = tuple.len() as u16 + slot_size;
        let free_space = header.data_start_offset - header.free_space_offset;

        if needed > free_space {
            return None; // Page full
        }

        // Write tuple at the end (growing upward)
        header.data_start_offset -= tuple.len() as u16;
        let data_offset = header.data_start_offset;
        self.data[data_offset as usize..data_offset as usize + tuple.len()]
            .copy_from_slice(tuple);

        // Write slot entry (growing downward)
        let slot_offset = header.free_space_offset;
        self.data[slot_offset as usize..slot_offset as usize + 2]
            .copy_from_slice(&data_offset.to_le_bytes());
        self.data[slot_offset as usize + 2..slot_offset as usize + 4]
            .copy_from_slice(&(tuple.len() as u16).to_le_bytes());
        header.free_space_offset += slot_size;

        let slot_id = header.num_slots;
        header.num_slots += 1;
        Some(slot_id)
    }

    fn get_tuple(&self, slot_id: u16) -> Option<&[u8]> {
        let header = self.header();
        if slot_id >= header.num_slots {
            return None;
        }
        let slot_offset = std::mem::size_of::<PageHeader>() + (slot_id as usize * 4);
        let data_offset = u16::from_le_bytes([
            self.data[slot_offset], self.data[slot_offset + 1]
        ]) as usize;
        let data_len = u16::from_le_bytes([
            self.data[slot_offset + 2], self.data[slot_offset + 3]
        ]) as usize;
        Some(&self.data[data_offset..data_offset + data_len])
    }

    // ... header(), header_mut() helpers omitted for brevity
}
```

---

## 3. Buffer Pool Manager

```
The most important component after disk I/O.
The buffer pool caches pages in memory so you don't read from disk every time.

                     ┌────────────────────────────────────┐
                     │         Buffer Pool (RAM)           │
   Page Request      │  ┌────┐ ┌────┐ ┌────┐ ┌────┐      │
   ──────────────→   │  │ P3 │ │ P7 │ │ P1 │ │ P12│      │
                     │  │    │ │    │ │ D  │ │    │      │  D = dirty
                     │  └────┘ └────┘ └────┘ └────┘      │
                     │                                     │
                     │  Page Table: page_id → frame_id     │
                     │  Free List: [frame 4, frame 5, ...] │
                     │  Replacement Policy: LRU / Clock    │
                     └─────────────────┬──────────────────┘
                                       │
                                       │  miss → read from disk
                                       │  evict dirty → write to disk
                                       ▼
                     ┌────────────────────────────────────┐
                     │           Disk (Files)              │
                     │  [P0][P1][P2][P3][P4][P5]...       │
                     └────────────────────────────────────┘

Key operations:
  fetch_page(page_id):
    1. Check page table — if page is in pool, increment pin count, return it
    2. If not in pool, find free frame (or evict via replacement policy)
    3. If evicted page is dirty, write it to disk first
    4. Read requested page from disk into frame
    5. Update page table, set pin count = 1, return page

  unpin_page(page_id, is_dirty):
    Decrement pin count. Mark dirty if modified.
    Page can only be evicted when pin count = 0.

  flush_page(page_id):
    Write page to disk if dirty. Used by WAL before commit.
```

### LRU-K Replacement Policy

```
Simple LRU evicts the least recently used page.
Problem: a sequential scan pollutes the cache (touches every page once).

LRU-K tracks the K-th most recent access:
  - LRU-1: simple LRU (last access time)
  - LRU-2: evict page whose 2nd-most-recent access is oldest
  - A page accessed only once has K-th access = -∞ → evicted first
  - Prevents sequential scan from evicting frequently-accessed pages

PostgreSQL uses Clock sweep (approximation of LRU):
  - Circular buffer of frames
  - Each frame has a "usage count" (incremented on access, max 5)
  - Clock hand sweeps: if usage_count > 0, decrement and skip
  - If usage_count = 0, evict this frame
  - Simple, O(1), cache-friendly
```

### Implementation Skeleton

```rust
use std::collections::HashMap;
use std::sync::{Arc, Mutex, RwLock};

const BUFFER_POOL_SIZE: usize = 1024; // number of frames

type PageId = u32;
type FrameId = usize;

struct BufferPoolManager {
    pages: Vec<RwLock<Page>>,          // frames holding pages
    page_table: Mutex<HashMap<PageId, FrameId>>,
    free_list: Mutex<Vec<FrameId>>,
    replacer: Mutex<LruKReplacer>,     // eviction policy
    disk_manager: DiskManager,
}

impl BufferPoolManager {
    fn fetch_page(&self, page_id: PageId) -> Option<&RwLock<Page>> {
        let mut page_table = self.page_table.lock().unwrap();

        // 1. Page already in buffer pool?
        if let Some(&frame_id) = page_table.get(&page_id) {
            self.replacer.lock().unwrap().record_access(frame_id);
            self.replacer.lock().unwrap().set_evictable(frame_id, false);
            return Some(&self.pages[frame_id]);
        }

        // 2. Find a frame (free list or eviction)
        let frame_id = {
            let mut free_list = self.free_list.lock().unwrap();
            if let Some(fid) = free_list.pop() {
                fid
            } else {
                // Evict
                let mut replacer = self.replacer.lock().unwrap();
                let victim = replacer.evict()?;
                // If dirty, flush to disk
                let page = self.pages[victim].read().unwrap();
                if page.is_dirty {
                    self.disk_manager.write_page(page.page_id, &page.data);
                }
                // Remove old mapping
                page_table.remove(&page.page_id);
                drop(page);
                victim
            }
        };

        // 3. Read page from disk into frame
        let data = self.disk_manager.read_page(page_id);
        {
            let mut page = self.pages[frame_id].write().unwrap();
            page.page_id = page_id;
            page.data = data;
            page.is_dirty = false;
            page.pin_count = 1;
        }

        // 4. Update page table
        page_table.insert(page_id, frame_id);
        self.replacer.lock().unwrap().record_access(frame_id);
        self.replacer.lock().unwrap().set_evictable(frame_id, false);

        Some(&self.pages[frame_id])
    }

    fn unpin_page(&self, page_id: PageId, is_dirty: bool) {
        let page_table = self.page_table.lock().unwrap();
        if let Some(&frame_id) = page_table.get(&page_id) {
            let mut page = self.pages[frame_id].write().unwrap();
            if is_dirty { page.is_dirty = true; }
            page.pin_count -= 1;
            if page.pin_count == 0 {
                self.replacer.lock().unwrap().set_evictable(frame_id, true);
            }
        }
    }
}
```

---

## 4. B+ Tree Index

```
The B+ tree is THE index structure for disk-based databases.
Every PostgreSQL, MySQL, SQL Server, Oracle index is a B+ tree by default.

Properties:
  - Balanced (all leaves at same depth)
  - High fan-out (100s of keys per node → tree is shallow)
  - Internal nodes: keys + child pointers (routing only)
  - Leaf nodes: keys + values (row pointers) + sibling pointers
  - Leaf nodes form a doubly-linked list → efficient range scans

For a fan-out of 200:
  Depth 1: 200 keys
  Depth 2: 200 × 200 = 40,000 keys
  Depth 3: 200 × 200 × 200 = 8,000,000 keys
  Depth 4: 200^4 = 1,600,000,000 keys
  → 1.6 BILLION rows indexed in 4 levels (4 disk reads max)
  → With buffer pool, root + first levels are cached → 1-2 disk reads

Structure:
                        ┌─────────────┐
                        │  [30 | 70]  │  ← Internal node (root)
                        └──┬────┬────┬┘
                     <30   │  30-70  │   ≥70
                  ┌────────┘    │    └────────┐
                  ▼             ▼             ▼
           ┌──────────┐ ┌──────────┐ ┌──────────┐
           │[10|20]   │ │[40|50|60]│ │[80|90]   │  ← Internal nodes
           └─┬──┬──┬──┘ └─┬──┬──┬─┘ └─┬──┬──┬──┘
             ▼  ▼  ▼      ▼  ▼  ▼     ▼  ▼  ▼
  Leaves: [5,10]↔[15,20]↔[35,40]↔[50,55]↔[60,65]↔[75,80]↔[85,90,95]
          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
          Doubly-linked list (for ORDER BY / range scans)

Operations:
  Search: start at root, follow pointers based on key comparison → O(log_B N)
  Insert: find leaf, insert key. If leaf full → split into two, push middle key up
  Delete: find leaf, remove key. If leaf underfull → merge or redistribute
  Range scan: find start leaf, follow sibling pointers → sequential I/O
```

### Insert with Split (Walkthrough)

```
Insert key 25 into a B+ tree with max 3 keys per leaf:

Before:
  Root: [30]
  Left leaf: [10, 20]  →  Right leaf: [30, 40, 50]

Step 1: Search → key 25 < 30 → go to left leaf [10, 20]
Step 2: Insert 25 → leaf becomes [10, 20, 25] → fits! Done.

Now insert 27:
Step 1: Search → 27 < 30 → left leaf [10, 20, 25]
Step 2: Insert 27 → leaf would be [10, 20, 25, 27] → OVERFLOW (max 3)
Step 3: Split leaf:
  Left half:  [10, 20]
  Right half: [25, 27]
  Push middle key (25) up to parent
Step 4: Parent root becomes [25, 30]
  Children: [10,20] → [25,27] → [30,40,50]
```

### Implementation Skeleton

```rust
const ORDER: usize = 4; // max keys per node (fanout = ORDER + 1)

enum BPlusTreeNode {
    Internal {
        keys: Vec<i64>,
        children: Vec<PageId>,  // child page IDs
    },
    Leaf {
        keys: Vec<i64>,
        values: Vec<RowId>,     // pointers to actual rows
        next_leaf: Option<PageId>,
        prev_leaf: Option<PageId>,
    },
}

struct BPlusTree {
    root_page_id: PageId,
    buffer_pool: Arc<BufferPoolManager>,
}

impl BPlusTree {
    fn search(&self, key: i64) -> Option<RowId> {
        let mut page_id = self.root_page_id;
        loop {
            let page = self.buffer_pool.fetch_page(page_id)?;
            let node = deserialize_node(&page.read().unwrap().data);
            match node {
                BPlusTreeNode::Internal { keys, children } => {
                    // Binary search for the child to follow
                    let idx = keys.partition_point(|k| *k <= key);
                    self.buffer_pool.unpin_page(page_id, false);
                    page_id = children[idx];
                }
                BPlusTreeNode::Leaf { keys, values, .. } => {
                    let result = keys.binary_search(&key)
                        .ok()
                        .map(|idx| values[idx]);
                    self.buffer_pool.unpin_page(page_id, false);
                    return result;
                }
            }
        }
    }

    fn insert(&self, key: i64, value: RowId) {
        // 1. Find the correct leaf page
        // 2. Insert key-value into leaf
        // 3. If leaf overflows (keys.len() > ORDER):
        //    a. Split leaf into two
        //    b. Push middle key up to parent
        //    c. If parent overflows, recursively split upward
        //    d. If root splits, create new root
        // (Full implementation: ~200-400 lines)
        todo!()
    }

    fn range_scan(&self, start: i64, end: i64) -> Vec<RowId> {
        // 1. Search for start key → find leaf
        // 2. Scan through leaf entries where key <= end
        // 3. Follow next_leaf pointer to next leaf
        // 4. Repeat until key > end
        todo!()
    }
}
```

---

## 5. Write-Ahead Log (WAL)

```
THE fundamental principle of crash recovery:
"Before modifying any page on disk, write the change to the log first."

Why: If the system crashes mid-write, the log tells us exactly what 
happened and what needs to be undone or redone.

WAL protocol (STEAL/NO-FORCE):
  STEAL: dirty pages CAN be flushed before commit (saves memory)
  NO-FORCE: dirty pages DON'T have to be flushed at commit (saves I/O)
  → Allows maximum buffer pool flexibility
  → Need UNDO (for uncommitted changes on disk) + REDO (for committed
    changes not yet on disk)

Log record types:
┌──────────────────────────────────────────────────────────────┐
│ LSN  | TxnId | Type     | PageId | Offset | Before | After  │
├──────────────────────────────────────────────────────────────┤
│  1   |  T1   | BEGIN    |   -    |   -    |   -    |   -    │
│  2   |  T1   | UPDATE   |   5    |  120   | "old"  | "new"  │
│  3   |  T2   | BEGIN    |   -    |   -    |   -    |   -    │
│  4   |  T2   | INSERT   |   8    |  200   |   -    | "data" │
│  5   |  T1   | COMMIT   |   -    |   -    |   -    |   -    │
│  6   |  T2   | UPDATE   |   3    |   50   | "abc"  | "xyz"  │
│  --- CRASH ---                                                │
│                                                               │
│ Recovery:                                                     │
│   T1: COMMITTED → REDO all T1 changes (ensure on disk)       │
│   T2: NOT committed → UNDO all T2 changes (roll back)        │
│   After recovery: database is consistent.                     │
└──────────────────────────────────────────────────────────────┘
```

### ARIES Recovery Algorithm

```
ARIES (Algorithm for Recovery and Isolation Exploiting Semantics)
THE standard recovery algorithm. Used by PostgreSQL, MySQL, SQL Server, DB2.

Three phases:

Phase 1 — ANALYSIS:
  Scan log forward from last checkpoint.
  Build two tables:
    - Active Transaction Table (ATT): txns that were active at crash
    - Dirty Page Table (DPT): pages that were dirty at crash
  Determines: which txns to redo, which to undo, where to start redo

Phase 2 — REDO:
  Scan log forward from earliest LSN in DPT (the "redo point").
  For each log record:
    If page in DPT AND page's LSN on disk < log record's LSN:
      Reapply the change (redo)
  This restores the database to its exact pre-crash state
  (including uncommitted changes — that's ok, we'll undo those next)

Phase 3 — UNDO:
  Scan log backward.
  For each uncommitted transaction (from ATT):
    Undo all its changes in reverse order.
    Write Compensation Log Records (CLRs) so undo is also crash-safe.
  After undo: all uncommitted changes are rolled back.

Why ARIES is brilliant:
  - REDO is "history repeating" — replays exact order of events
  - UNDO writes CLRs → recovery is idempotent (crash during recovery is safe)
  - Checkpointing reduces recovery time (don't scan from beginning of log)
  - Fine-grained logging (page-level, not table-level)
```

### Implementation Skeleton

```rust
#[derive(Clone)]
struct LogRecord {
    lsn: u64,
    txn_id: u32,
    record_type: LogRecordType,
    page_id: Option<PageId>,
    offset: Option<u16>,
    before_image: Option<Vec<u8>>,  // for UNDO
    after_image: Option<Vec<u8>>,   // for REDO
    prev_lsn: Option<u64>,         // previous LSN for this txn
}

enum LogRecordType {
    Begin,
    Commit,
    Abort,
    Update,
    Insert,
    Delete,
    Checkpoint,
    CompensationLogRecord { undo_next_lsn: u64 },
}

struct WALManager {
    log_file: File,
    current_lsn: AtomicU64,
    buffer: Mutex<Vec<u8>>,  // log buffer (batch writes for performance)
}

impl WALManager {
    fn append(&self, record: &LogRecord) -> u64 {
        let lsn = self.current_lsn.fetch_add(1, Ordering::SeqCst);
        let mut buffer = self.buffer.lock().unwrap();
        let serialized = serialize_log_record(record);
        buffer.extend_from_slice(&serialized);
        // Flush to disk on commit or when buffer is full
        lsn
    }

    fn flush(&self, up_to_lsn: u64) {
        // Force all log records up to this LSN to disk
        // Called before commit returns to client
        let mut buffer = self.buffer.lock().unwrap();
        self.log_file.write_all(&buffer).unwrap();
        self.log_file.sync_all().unwrap(); // fsync!
        buffer.clear();
    }

    fn recover(&self, buffer_pool: &BufferPoolManager) {
        // Phase 1: Analysis
        let (active_txns, dirty_pages) = self.analysis_phase();
        
        // Phase 2: Redo
        self.redo_phase(&dirty_pages, buffer_pool);
        
        // Phase 3: Undo
        self.undo_phase(&active_txns, buffer_pool);
    }
}
```

---

## 6. MVCC (Multi-Version Concurrency Control)

```
Readers never block writers. Writers never block readers.
Each transaction sees a SNAPSHOT of the database at its start time.

Implementation approaches:

PostgreSQL style (append-only, in-place versions):
  Every tuple has: xmin (creating txn), xmax (deleting txn)
  UPDATE = mark old tuple dead (set xmax) + INSERT new tuple
  Visibility check: tuple visible if xmin committed AND xmax not committed
  Garbage: old versions accumulate → VACUUM cleans them

MySQL/InnoDB style (undo log):
  Main table always has latest version
  UPDATE = modify in-place, store old version in undo log
  Readers reconstruct old versions by applying undo records backward
  Less bloat than PostgreSQL, but undo log can grow

Visibility check (simplified Snapshot Isolation):
  Transaction T starts → records snapshot: all committed txn IDs at start time
  
  For each tuple:
    if tuple.xmin NOT in snapshot (not yet committed when T started) → invisible
    if tuple.xmin aborted → invisible
    if tuple.xmax committed AND in snapshot → invisible (deleted before T started)
    otherwise → visible

  This gives each transaction a consistent, frozen view of the database.
```

```rust
// Simplified MVCC visibility check:
struct TransactionManager {
    active_txns: Mutex<HashSet<u32>>,
    next_txn_id: AtomicU32,
}

struct Snapshot {
    txn_id: u32,
    active_at_start: HashSet<u32>,  // txns that were active when snapshot taken
    min_active: u32,                 // smallest active txn ID
}

impl Snapshot {
    fn is_visible(&self, tuple: &Tuple) -> bool {
        let xmin_visible = tuple.xmin < self.min_active  // committed before any active txn
            || (tuple.xmin < self.txn_id && !self.active_at_start.contains(&tuple.xmin));
        
        if !xmin_visible { return false; }

        match tuple.xmax {
            None => true,  // not deleted
            Some(xmax) => {
                // Deleted, but by whom?
                if xmax == self.txn_id { return false; }  // we deleted it
                if self.active_at_start.contains(&xmax) { return true; }  // deleter hadn't committed
                if xmax > self.txn_id { return true; }  // deleted after our snapshot
                false  // deleted and committed before our snapshot
            }
        }
    }
}
```

---

## 7. Query Engine — Volcano Model

```
The Volcano (Iterator) model: every operator is an iterator with:
  open()  → initialize
  next()  → return next tuple (or None)
  close() → clean up

Operators are composed into a tree. Data flows bottom-up.

Query: SELECT name, age FROM users WHERE age > 25 ORDER BY age

Execution tree:
  ┌──────────────┐
  │   Sort       │  ← next() returns tuples in age order
  │  (age ASC)   │
  └──────┬───────┘
         │ next()
  ┌──────┴───────┐
  │   Project    │  ← next() strips columns to (name, age)
  │ (name, age)  │
  └──────┬───────┘
         │ next()
  ┌──────┴───────┐
  │   Filter     │  ← next() skips tuples where age <= 25
  │ (age > 25)   │
  └──────┬───────┘
         │ next()
  ┌──────┴───────┐
  │  SeqScan     │  ← next() reads next tuple from users table
  │  (users)     │
  └──────────────┘

Each operator pulls one tuple at a time from its child.
Simple, composable, but function-call overhead per tuple.
```

```rust
trait Executor {
    fn open(&mut self);
    fn next(&mut self) -> Option<Tuple>;
    fn close(&mut self);
}

struct SeqScan {
    table_id: u32,
    current_page: PageId,
    current_slot: u16,
    buffer_pool: Arc<BufferPoolManager>,
}

impl Executor for SeqScan {
    fn open(&mut self) { self.current_page = 0; self.current_slot = 0; }
    fn next(&mut self) -> Option<Tuple> {
        // Read next tuple from table, advancing page/slot
        // When page exhausted, move to next page
        // When no more pages, return None
        todo!()
    }
    fn close(&mut self) {}
}

struct Filter {
    child: Box<dyn Executor>,
    predicate: Box<dyn Fn(&Tuple) -> bool>,
}

impl Executor for Filter {
    fn open(&mut self) { self.child.open(); }
    fn next(&mut self) -> Option<Tuple> {
        loop {
            let tuple = self.child.next()?;
            if (self.predicate)(&tuple) { return Some(tuple); }
        }
    }
    fn close(&mut self) { self.child.close(); }
}

struct Project {
    child: Box<dyn Executor>,
    column_indices: Vec<usize>,
}

impl Executor for Project {
    fn open(&mut self) { self.child.open(); }
    fn next(&mut self) -> Option<Tuple> {
        let tuple = self.child.next()?;
        let projected = self.column_indices.iter()
            .map(|&i| tuple.columns[i].clone())
            .collect();
        Some(Tuple { columns: projected })
    }
    fn close(&mut self) { self.child.close(); }
}

// Hash Join, Sort, Aggregate, etc. follow the same pattern
```

---

## 8. Simple SQL Parser

```
SQL parsing: text → Abstract Syntax Tree (AST)

"SELECT name, age FROM users WHERE age > 25"
         ↓ (Lexer)
Tokens: [SELECT, IDENT("name"), COMMA, IDENT("age"), FROM, 
         IDENT("users"), WHERE, IDENT("age"), GT, INT(25)]
         ↓ (Parser)
AST:
  SelectStatement {
    columns: [Column("name"), Column("age")],
    from: Table("users"),
    where: BinaryExpr {
      left: Column("age"),
      op: GreaterThan,
      right: Literal(25),
    },
    order_by: None,
    limit: None,
  }

Tools for building parsers:
  - Hand-written recursive descent (PostgreSQL, SQLite do this)
  - Parser combinators: nom (Rust), pyparsing (Python)
  - Parser generators: ANTLR, yacc/bison, lalrpop (Rust)
  - sqlparser-rs (Rust crate — full SQL parser, use this to get started)
```

---

## 9. Query Optimizer (Intro)

```
The optimizer transforms a logical plan into the most efficient physical plan.

Logical plan → Physical plan:
  SeqScan vs IndexScan (is there a useful index?)
  Nested Loop Join vs Hash Join vs Sort-Merge Join
  Which join order? (for N tables, N! possible orders)

Cost model basics:
  Cost(SeqScan) = num_pages × cost_per_page_read
  Cost(IndexScan) = tree_height × cost_per_page + selectivity × num_pages
  Cost(HashJoin) = build_cost + probe_cost = 3 × (|R| + |S|) pages
  Cost(SortMergeJoin) = sort(R) + sort(S) + merge = O(|R| log|R| + |S| log|S|)

Selectivity estimation (how many rows match a predicate):
  - Histograms: distribution of values in a column
  - Most common values (MCV): exact counts for frequent values
  - Distinct count: for equality predicates → selectivity ≈ 1/n_distinct
  - PostgreSQL: pg_statistic stores all of this (updated by ANALYZE)

Optimization approaches:
  Rule-based: apply heuristic rules (push predicates down, etc.)
  Cost-based: enumerate plans, estimate cost, pick cheapest
  PostgreSQL: cost-based with genetic algorithm for many-table joins (GEQO)
```

---

## 10. Putting It All Together — Project Resources

```
Build projects (increasing difficulty):

1. ★ Let's Build a Simple Database (cstack.github.io/db_tutorial)
   - Build a SQLite clone in C, step by step
   - Covers: REPL, SQL parsing, B-tree, pager, cursor
   - Best starting point. Do this first.

2. ★★ CMU BusTub (github.com/cmu-db/bustub)
   - Course project for CMU 15-445
   - Implement: buffer pool, B+ tree index, query execution, concurrency
   - C++, well-structured, with auto-grader
   - Follow along with Andy Pavlo's lectures (YouTube)

3. ★★★ mini-lsm (github.com/skyzh/mini-lsm)
   - Build an LSM-tree storage engine in Rust
   - Covers: memtable, SST, compaction, bloom filters, MVCC
   - Excellent Rust learning project

4. ★★★ toydb (github.com/erikgrinaker/toydb)
   - Distributed SQL database in Rust
   - Raft consensus, MVCC, SQL engine, B+ tree storage
   - Read the source code — beautifully written

5. ★★★★ chidb (people.cs.uchicago.edu/~adamshaw/chidb)
   - SQLite-like database in C
   - Full B-tree, pager, SQL parser, VM

6. ★★★★★ Your own:
   - Pick a language (Rust, Go, C recommended)
   - Start with key-value store → add B+ tree → add WAL → add SQL
   - Blog about your progress — this IS the resume
```

---

## Key Takeaways

1. **Pages are the atom of database I/O.** Everything — tables, indexes, WAL — is stored in fixed-size pages. Master the slotted page layout.
2. **The buffer pool** is the most performance-critical component. It sits between every query and the disk. LRU-K or Clock replacement prevents sequential scan pollution.
3. **B+ tree** provides O(log N) search with only 3-4 disk reads for billions of rows thanks to high fan-out. Leaf sibling pointers enable efficient range scans.
4. **WAL + ARIES** is the foundation of crash safety. Write the log first, then modify pages. Analysis → Redo → Undo. Every modern database uses this.
5. **Start with cstack's db_tutorial, then CMU BusTub.** Building a database from scratch is the single most educational thing you can do in computer science.

---

Next: [02-source-code-reading.md](02-source-code-reading.md) →
