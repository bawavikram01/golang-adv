# 2.4 — Concurrency Control: ACID, Isolation, MVCC, and Locking

> Multiple transactions running simultaneously. Each thinks it's alone.  
> How? This chapter is the beating heart of database correctness.

---

## 1. ACID — The Contract

Every transaction must satisfy ACID:

### Atomicity — All or Nothing

```
Transaction: Transfer $500 from Alice to Bob
  1. UPDATE accounts SET balance = balance - 500 WHERE name = 'Alice';
  2. UPDATE accounts SET balance = balance + 500 WHERE name = 'Bob';

If the system crashes after step 1 but before step 2:
  WITHOUT atomicity: Alice lost $500, Bob got nothing. Money vanished.
  WITH atomicity: the entire transaction is rolled back. Both balances unchanged.

HOW: Write-Ahead Log. On crash, undo partial transactions.
```

### Consistency — Valid State to Valid State

```
Constraints: balance >= 0, total money in system is constant.

If Alice has $300 and tries to send $500:
  CHECK (balance >= 0) fails → transaction aborted → still consistent.

Consistency = application invariants + database constraints.
The database enforces structural consistency (types, FKs, CHECKs).
The application enforces semantic consistency (business rules).
```

### Isolation — Transactions Don't Interfere

```
Transaction A: SELECT balance FROM accounts WHERE name = 'Alice';
Transaction B: UPDATE accounts SET balance = balance - 100 WHERE name = 'Alice';

WITHOUT isolation: A might see the balance mid-update (partially written data).
WITH isolation: A sees EITHER the old balance OR the new balance, never a torn value.

The level of isolation is configurable (isolation levels — see below).
```

### Durability — Committed = Permanent

```
After COMMIT returns success to the client:
  The data MUST survive a crash, power failure, disk failure.

HOW: WAL flushed to disk BEFORE commit returns.
     UPS, RAID, replicas for hardware failures.
```

---

## 2. Transaction Isolation Levels

### The Anomalies

```
DIRTY READ:
  TX1 writes a row (not yet committed).
  TX2 reads that uncommitted data.
  TX1 aborts → TX2 read something that never existed.

NON-REPEATABLE READ:
  TX1 reads a row.
  TX2 modifies and commits it.
  TX1 reads the same row again → different value!

PHANTOM READ:
  TX1 reads rows matching WHERE salary > 100000 (finds 5 rows).
  TX2 inserts a new row with salary = 150000 and commits.
  TX1 re-runs the same query → finds 6 rows! A "phantom" appeared.

WRITE SKEW:
  TX1 reads: Alice is on-call, Bob is on-call (2 doctors on-call).
  TX2 reads: Alice is on-call, Bob is on-call.
  TX1: since Bob is on-call, I'll take Alice off-call.
  TX2: since Alice is on-call, I'll take Bob off-call.
  Both commit → ZERO doctors on-call! Violated invariant.
  Each transaction's logic was correct based on what it read, but combined = wrong.

LOST UPDATE:
  TX1 reads counter = 42.
  TX2 reads counter = 42.
  TX1 writes counter = 43.
  TX2 writes counter = 43.
  → One increment lost!
```

### Isolation Levels vs Anomalies

| Level | Dirty Read | Non-Repeatable Read | Phantom Read | Write Skew |
|-------|-----------|-------------------|-------------|-----------|
| READ UNCOMMITTED | Possible | Possible | Possible | Possible |
| READ COMMITTED | **Prevented** | Possible | Possible | Possible |
| REPEATABLE READ | Prevented | **Prevented** | Possible* | Possible |
| SERIALIZABLE | Prevented | Prevented | **Prevented** | **Prevented** |

*PostgreSQL's REPEATABLE READ actually prevents phantom reads too (it uses snapshot isolation). But SQL standard says phantom reads are allowed.

### READ COMMITTED (PostgreSQL default)

```
Each STATEMENT sees the latest committed data at statement start.
Two statements in the same transaction can see different data.

TX1:                           TX2:
BEGIN;
SELECT balance FROM accounts   
WHERE name = 'Alice';
→ 1000
                               BEGIN;
                               UPDATE accounts 
                               SET balance = 900
                               WHERE name = 'Alice';
                               COMMIT;

SELECT balance FROM accounts
WHERE name = 'Alice';
→ 900  (sees TX2's commit!)

Different result within the SAME transaction.
Each SELECT gets a FRESH snapshot.
```

### REPEATABLE READ / Snapshot Isolation

```
The transaction gets a SNAPSHOT at the start of the FIRST query.
All reads see the database as of that moment. Committed changes by other
transactions after the snapshot are INVISIBLE.

TX1:                           TX2:
BEGIN ISOLATION LEVEL 
REPEATABLE READ;
SELECT balance FROM accounts
WHERE name = 'Alice';
→ 1000
                               BEGIN;
                               UPDATE accounts
                               SET balance = 900
                               WHERE name = 'Alice';
                               COMMIT;

SELECT balance FROM accounts
WHERE name = 'Alice';
→ 1000  (still sees old value! snapshot was taken at first SELECT)

What if TX1 tries to UPDATE the same row TX2 already modified?
  UPDATE accounts SET balance = balance + 100 WHERE name = 'Alice';
  → ERROR: could not serialize access due to concurrent update
  → TX1 must RETRY the entire transaction.
```

### SERIALIZABLE

```
Transactions behave as if they executed one after another (serially).
No anomaly is possible. The strongest guarantee.

PostgreSQL implements this with Serializable Snapshot Isolation (SSI):
  - Takes a snapshot like REPEATABLE READ
  - Additionally tracks read/write dependencies between transactions
  - If it detects a cycle in the dependency graph → abort one transaction

TX1:                           TX2:
BEGIN ISOLATION LEVEL
SERIALIZABLE;                  BEGIN ISOLATION LEVEL SERIALIZABLE;

SELECT COUNT(*) FROM oncall
WHERE active = true;           SELECT COUNT(*) FROM oncall
→ 2                            WHERE active = true;
                               → 2

UPDATE oncall                  
SET active = false             UPDATE oncall
WHERE name = 'Alice';          SET active = false
                               WHERE name = 'Bob';
COMMIT; → OK                   
                               COMMIT; → ERROR: serialization failure!
                               (detected write skew: both transactions
                                read the same rows and made conflicting writes)
```

---

## 3. Locking

### Lock Types

```
SHARED LOCK (S): "I'm reading this, don't modify it"
  Multiple transactions can hold S locks on the same resource simultaneously.

EXCLUSIVE LOCK (X): "I'm modifying this, nobody else touch it"
  Only one transaction can hold an X lock. Blocks all other locks.

Compatibility matrix:
           Requesting
           S      X
Held S:    ✓      ✗   (shared + shared = OK, shared + exclusive = BLOCK)
Held X:    ✗      ✗   (exclusive blocks everything)
```

### Lock Granularity

```
DATABASE LOCK: lock entire database (only for maintenance)
TABLE LOCK:    lock entire table (DDL operations, LOCK TABLE)
PAGE LOCK:     lock a page (some systems, not common in PostgreSQL)
ROW LOCK:      lock individual row (most common for DML)

Finer granularity = more concurrency, more overhead
Coarser granularity = less concurrency, less overhead

PostgreSQL: row-level locking for DML (INSERT, UPDATE, DELETE)
           table-level locking for DDL (ALTER TABLE, CREATE INDEX)
```

### Intention Locks

```
Problem: TX1 holds a row lock on table T. TX2 wants a table-level X lock on T.
How does TX2 know there's a conflicting row lock without checking EVERY row?

Solution: INTENTION LOCKS — "I intend to lock at a finer level"

Before locking a ROW, first acquire an INTENTION lock on the TABLE:
  IS (Intention Shared): "I'll acquire S locks on some rows"
  IX (Intention Exclusive): "I'll acquire X locks on some rows"

Compatibility:
           IS    IX    S     X
IS:        ✓     ✓     ✓     ✗
IX:        ✓     ✓     ✗     ✗
S:         ✓     ✗     ✓     ✗
X:         ✗     ✗     ✗     ✗

Example:
  TX1: IX lock on table → X lock on row 42
  TX2: IX lock on table → X lock on row 99
  Both succeed! (IX + IX is compatible, different row locks)

  TX3: wants X lock on TABLE → blocked by TX1's IX (table X vs IX = incompatible)
```

### Two-Phase Locking (2PL)

```
The protocol that guarantees serializability:

PHASE 1 — GROWING: acquire locks, never release
PHASE 2 — SHRINKING: release locks, never acquire

        locks held
          ^
          |      /\
          |     /  \
          |    /    \
          |   /      \
          |  /        \
          +--+--------+--→ time
          growing  shrinking

STRICT 2PL (most common):
  Hold ALL locks until COMMIT or ABORT.
  → No shrinking phase — just release everything at once.
  → Prevents cascading aborts.

RIGOROUS 2PL:
  Same as strict 2PL. All locks released after commit.
  (In practice, strict 2PL and rigorous 2PL are the same.)
```

### Deadlock Detection

```
TX1: holds lock on row A, waiting for lock on row B
TX2: holds lock on row B, waiting for lock on row A

→ DEADLOCK! Neither can proceed.

Detection: Wait-For Graph
  TX1 → TX2 (TX1 waits for TX2)
  TX2 → TX1 (TX2 waits for TX1)
  Cycle detected → one transaction must be aborted.

PostgreSQL: checks for deadlocks every deadlock_timeout (default 1s).
  When detected: aborts the transaction that has done the least work.

Prevention strategies:
  1. Lock ordering: always acquire locks in a consistent order
     (e.g., always lock rows by primary key ascending)
  2. Lock timeout: SET lock_timeout = '5s'; — give up after 5 seconds
  3. Nowait: SELECT ... FOR UPDATE NOWAIT; — fail immediately if can't lock
```

### Explicit Locking in PostgreSQL

```sql
-- Row-level locks
SELECT * FROM accounts WHERE id = 1 FOR UPDATE;          -- X lock, block other writers AND FOR UPDATE
SELECT * FROM accounts WHERE id = 1 FOR NO KEY UPDATE;   -- weaker X lock, doesn't block FK checks
SELECT * FROM accounts WHERE id = 1 FOR SHARE;           -- S lock, block writers
SELECT * FROM accounts WHERE id = 1 FOR KEY SHARE;       -- weakest S lock

-- SKIP LOCKED: skip rows that are already locked (for queue processing!)
SELECT * FROM tasks WHERE status = 'pending'
ORDER BY created_at
LIMIT 1
FOR UPDATE SKIP LOCKED;
-- Multiple workers can dequeue without blocking each other!

-- NOWAIT: error immediately if row is locked
SELECT * FROM accounts WHERE id = 1 FOR UPDATE NOWAIT;

-- Table-level locks
LOCK TABLE accounts IN ACCESS EXCLUSIVE MODE;  -- blocks everything
LOCK TABLE accounts IN SHARE MODE;             -- blocks writes, allows reads

-- Advisory locks (application-level custom locks)
SELECT pg_advisory_lock(42);      -- named lock #42, blocks until acquired
SELECT pg_try_advisory_lock(42);  -- returns false immediately if can't acquire
SELECT pg_advisory_unlock(42);    -- release

-- Advisory lock use cases:
-- Prevent duplicate cron jobs: IF pg_try_advisory_lock(hash('daily_report')) THEN run_report()
-- Application-level mutex on a resource
-- Rate limiting per user: pg_try_advisory_lock(user_id)
```

---

## 4. MVCC — Multi-Version Concurrency Control

### The Core Idea

```
Instead of locking rows during reads:
  → Keep MULTIPLE VERSIONS of each row
  → Each transaction sees the version valid at its snapshot time
  → Readers NEVER block writers. Writers NEVER block readers.

This is how PostgreSQL, MySQL/InnoDB, Oracle, and most modern databases work.
```

### PostgreSQL MVCC Implementation

```
Every row (tuple) has hidden system columns:
  xmin  — the transaction ID (XID) that INSERTED this tuple
  xmax  — the transaction ID that DELETED/UPDATED this tuple (0 if live)
  cmin  — command ID within xmin transaction
  cmax  — command ID within xmax transaction
  ctid  — physical location (page, offset) — points to itself, or newer version

Visibility rule (simplified):
  A tuple is VISIBLE to transaction T if:
    1. xmin is committed AND xmin ≤ T's snapshot
    2. AND (xmax is 0 OR xmax is not yet committed OR xmax > T's snapshot)

  "The row was inserted before my snapshot, and either not deleted
   or deleted after my snapshot."
```

### Example: How It Works

```
Initial state: Alice has balance = 1000
  Page 5, slot 1: (xmin=100, xmax=0, balance=1000)

TX 200 starts (READ COMMITTED)
TX 300 starts (READ COMMITTED)

TX 200: UPDATE accounts SET balance = 900 WHERE name = 'Alice'
  1. Mark old tuple: xmax = 200
     Page 5, slot 1: (xmin=100, xmax=200, balance=1000) ← "dead" to TX 200
  2. Insert new tuple:
     Page 5, slot 2: (xmin=200, xmax=0, balance=900) ← new version

TX 300: SELECT balance FROM accounts WHERE name = 'Alice'
  Scans page 5:
    Slot 1: xmin=100 (committed ✓), xmax=200 (not yet committed!) → VISIBLE → balance=1000
    Slot 2: xmin=200 (not yet committed!) → NOT VISIBLE

TX 200: COMMIT

TX 300: SELECT balance FROM accounts WHERE name = 'Alice'
  (New statement in READ COMMITTED → new snapshot)
    Slot 1: xmin=100 ✓, xmax=200 (now committed) → NOT VISIBLE (deleted)
    Slot 2: xmin=200 (now committed) → VISIBLE → balance=900

VACUUM later:
  Slot 1 can be reclaimed — no active transaction can see it anymore.
```

### MySQL/InnoDB MVCC Implementation

```
Different approach from PostgreSQL:

1. The clustered index (primary key B-tree) stores the LATEST committed version.
2. Old versions are stored in the UNDO LOG (rollback segment).
3. Each row has:
   - DB_TRX_ID: transaction ID of last modification
   - DB_ROLL_PTR: pointer to previous version in undo log

Read flow:
  1. Read row from clustered index
  2. Check DB_TRX_ID against my read view (snapshot)
  3. If too new: follow DB_ROLL_PTR to undo log → find older version
  4. Repeat until finding a version visible to my snapshot

┌─────────────────────┐
│ Clustered Index      │    ┌──────────────────────┐
│ PK=1, balance=900   │───→│ Undo Log              │
│ TRX_ID=200          │    │ PK=1, balance=1000    │
│ ROLL_PTR ──────────────→ │ TRX_ID=100            │
└─────────────────────┘    │ ROLL_PTR → older...   │
                           └──────────────────────┘

PostgreSQL vs InnoDB MVCC:
  PostgreSQL:
    + Old versions in heap (same table) → simpler architecture
    − Table bloats with dead tuples → needs VACUUM
    − UPDATE writes full new tuple + all indexes updated
    
  InnoDB:
    + Old versions in separate undo log → main table stays compact
    + UPDATE in-place (only undo log grows)
    − Following undo chain for old versions is costly if long
    − Undo log must be purged periodically
```

### HOT Updates (Heap-Only Tuples) — PostgreSQL Optimization

```
Problem: every UPDATE creates a new tuple AND updates ALL indexes pointing to the old tuple.
  Table with 10 indexes: 1 UPDATE = 1 heap write + 10 index writes!

HOT (Heap-Only Tuple) optimization:
  IF the updated columns are NOT in ANY index:
    → Store new tuple version on the SAME page
    → DON'T update any indexes
    → Old tuple's ctid points to new tuple (chain within the page)
    → Index still points to old tuple → follows ctid chain to find current version

  Result: 1 heap write, 0 index writes. 10x faster!

When HOT works:
  ✓ Updated column is NOT indexed
  ✓ New tuple fits on the SAME page as old tuple
  ✓ fillfactor has room (set fillfactor < 100 to leave space)

  ALTER TABLE orders SET (fillfactor = 70);
  -- Leaves 30% free space per page for HOT updates
```

---

## 5. Snapshot Isolation vs Serializable Snapshot Isolation

### Snapshot Isolation (SI)

```
Every transaction sees a consistent snapshot of the database as of its start.
Writes by concurrent transactions are invisible.

Prevents: dirty reads, non-repeatable reads, phantom reads
Does NOT prevent: write skew

This is what PostgreSQL's REPEATABLE READ actually implements.
```

### Serializable Snapshot Isolation (SSI)

```
Like snapshot isolation, PLUS tracking of read-write dependencies.

PostgreSQL tracks:
  TX_A reads [rows that TX_B later writes] = rw-dependency A→B
  TX_B reads [rows that TX_A later writes] = rw-dependency B→A
  
  If a CYCLE forms: A→B→A → one transaction must abort.
  This is called a "dangerous structure."

Implementation:
  1. SIRead locks: record what each transaction READ
     (not blocking — just bookkeeping)
  2. On commit: check for rw-dependency cycles
  3. If cycle found: abort the transaction that's easiest to retry

Performance:
  - Overhead of tracking reads (~10-20% slower than REPEATABLE READ)
  - False positives: some aborts are unnecessary (conservative detection)
  - Application must be ready to RETRY aborted transactions
```

---

## 6. Lock-Free & Optimistic Concurrency Control

### Optimistic Concurrency Control (OCC)

```
Assume conflicts are RARE. Don't lock anything during execution.

Phase 1 — READ: Read data, compute results. Track what was read.
Phase 2 — VALIDATE: At commit time, check if anything you read was modified by another committed transaction.
Phase 3 — WRITE: If validation passes → commit. If fails → abort and retry.

Good when: conflicts are rare (mostly read-only workload)
Bad when: conflicts are frequent (high contention → many retries)
```

### Application-Level Optimistic Locking

```sql
-- Version column approach:
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name TEXT,
    price DECIMAL(10,2),
    version INTEGER DEFAULT 1
);

-- Read:
SELECT id, name, price, version FROM products WHERE id = 42;
-- Returns: (42, 'Widget', 29.99, 5)

-- Update with version check:
UPDATE products
SET price = 34.99, version = version + 1
WHERE id = 42 AND version = 5;

-- If affected rows = 1: success, we updated it.
-- If affected rows = 0: someone else modified it! Retry.

-- This is how most web applications handle concurrency:
-- No database-level locks held between HTTP requests.
```

---

## 7. MySQL/InnoDB Locking Details

### InnoDB Lock Types

```
Record Lock:     Locks a single index record.
Gap Lock:        Locks the GAP between index records (prevents phantom inserts).
Next-Key Lock:   Record lock + gap lock on gap before the record.
                 This is InnoDB's DEFAULT lock for REPEATABLE READ.
                 Prevents phantom reads.

Example: Index has values [10, 20, 30]

  Record lock on 20:  locks value 20 only
  Gap lock before 20: locks the gap (10, 20) — prevents INSERT of 15
  Next-key lock on 20: locks (10, 20] — prevents INSERT of 15 AND update to 20

  Gap locks between:
  (-∞, 10) | 10 | (10, 20) | 20 | (20, 30) | 30 | (30, +∞)
```

```sql
-- InnoDB locking in action:
-- REPEATABLE READ (default):

-- TX1:
SELECT * FROM orders WHERE total > 100 FOR UPDATE;
-- Acquires NEXT-KEY LOCKS on all matching rows + gaps
-- This prevents TX2 from inserting a row with total > 100 (no phantom!)

-- TX2:
INSERT INTO orders (total) VALUES (150);
-- BLOCKED! Gap lock from TX1 prevents this insert.
```

---

## 8. Common Concurrency Bugs & Solutions

### Bug 1: Lost Update

```sql
-- Two users read, modify, write the same row:
-- TX1: count = SELECT count FROM counters WHERE id = 1;  → 10
-- TX2: count = SELECT count FROM counters WHERE id = 1;  → 10
-- TX1: UPDATE counters SET count = 11 WHERE id = 1;
-- TX2: UPDATE counters SET count = 11 WHERE id = 1;
-- Result: count = 11 (should be 12!)

-- FIX 1: Atomic update (no read-modify-write)
UPDATE counters SET count = count + 1 WHERE id = 1;

-- FIX 2: SELECT ... FOR UPDATE (pessimistic lock)
BEGIN;
SELECT count FROM counters WHERE id = 1 FOR UPDATE;  -- locks the row
UPDATE counters SET count = count + 1 WHERE id = 1;
COMMIT;

-- FIX 3: REPEATABLE READ + retry
-- PostgreSQL will detect the conflict and abort one transaction
```

### Bug 2: Write Skew

```sql
-- Doctors on-call: at least one must be on-call
-- TX1:
BEGIN ISOLATION LEVEL SERIALIZABLE;
SELECT COUNT(*) FROM doctors WHERE on_call = TRUE;  -- sees 2
UPDATE doctors SET on_call = FALSE WHERE name = 'Alice';
COMMIT;

-- TX2:
BEGIN ISOLATION LEVEL SERIALIZABLE;
SELECT COUNT(*) FROM doctors WHERE on_call = TRUE;  -- sees 2
UPDATE doctors SET on_call = FALSE WHERE name = 'Bob';
COMMIT;  -- ERROR: serialization failure

-- FIX: Use SERIALIZABLE isolation level
-- Or: use explicit SELECT ... FOR UPDATE on the rows you're checking
```

### Bug 3: Read-Then-Insert (Duplicate)

```sql
-- Check if username exists, then insert:
-- TX1:
SELECT id FROM users WHERE username = 'alice';  -- not found
INSERT INTO users (username) VALUES ('alice');

-- TX2 (simultaneously):
SELECT id FROM users WHERE username = 'alice';  -- not found  
INSERT INTO users (username) VALUES ('alice');

-- Both insert! Duplicate username!

-- FIX 1: UNIQUE constraint (database enforces it)
-- FIX 2: INSERT ... ON CONFLICT (upsert)
-- FIX 3: SERIALIZABLE isolation level
-- FIX 4: Advisory lock on hash of username
```

---

## Key Takeaways

1. **ACID** is the contract. Atomicity (WAL), Consistency (constraints), Isolation (MVCC/locks), Durability (WAL+fsync).
2. **MVCC** lets readers and writers coexist without blocking. Old versions are kept until no transaction needs them.
3. **PostgreSQL MVCC** = heap-based versioning → needs VACUUM. **InnoDB MVCC** = undo log → needs purge.
4. **READ COMMITTED** (PostgreSQL default) = each statement sees latest commits. Good enough for most apps.
5. **REPEATABLE READ** = snapshot at transaction start. Prevents phantom reads in PostgreSQL (snapshot isolation).
6. **SERIALIZABLE** = true serial execution semantics. Detects write skew. Requires retry logic.
7. **Deadlocks** are detected and resolved by aborting one transaction. Prevent with consistent lock ordering.
8. **FOR UPDATE SKIP LOCKED** = the best pattern for job queues in PostgreSQL.

---

Next: [05-recovery-and-durability.md](05-recovery-and-durability.md) →
