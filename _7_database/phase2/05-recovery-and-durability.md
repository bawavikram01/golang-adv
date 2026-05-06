# 2.5 — Recovery & Durability: WAL, ARIES, and Crash Recovery

> The database just crashed. Power went out. Disk had a bad sector.  
> When it comes back, it MUST be in a consistent state with NO data loss.  
> This chapter explains how.

---

## 1. The Recovery Problem

```
At any moment, the database has:
  - Committed transactions whose data is ONLY in the buffer pool (not flushed to disk yet)
  - Active transactions that modified pages in the buffer pool (not committed)
  
If the system crashes:
  - Committed data only in buffer pool → LOST (violates durability)
  - Active transaction data in buffer pool → GARBAGE (violates atomicity)
  
Recovery must:
  1. REDO: Replay committed changes that were in buffer pool but not on disk
  2. UNDO: Remove uncommitted changes that were written to disk
```

---

## 2. WAL — Write-Ahead Logging (Recap + Deep Dive)

### The WAL Contract

```
RULE 1 (WAL Protocol): 
  Before a dirty page is written to disk,
  ALL log records for that page must be flushed to the log disk first.
  
  → Ensures we can REDO any change.

RULE 2 (Force-at-Commit):
  When a transaction commits,
  ALL its log records must be flushed to the log disk.
  
  → Ensures committed data survives crash.
  
These two rules are SUFFICIENT for full crash recovery.
```

### Log Record Types

```
INSERT record:  (LSN, txn_id, INSERT, page_id, offset, new_data)
UPDATE record:  (LSN, txn_id, UPDATE, page_id, offset, old_data, new_data)
DELETE record:  (LSN, txn_id, DELETE, page_id, offset, old_data)
COMMIT record:  (LSN, txn_id, COMMIT)
ABORT record:   (LSN, txn_id, ABORT)
BEGIN record:   (LSN, txn_id, BEGIN)
CLR record:     (LSN, txn_id, CLR, undo_next_LSN)  ← Compensation Log Record (for UNDO)
CHECKPOINT:     (LSN, active_txns, dirty_pages)
END record:     (LSN, txn_id, END)  ← transaction fully cleaned up
```

### LSN (Log Sequence Number)

```
Every log record gets a monotonically increasing LSN.
Every data page has a pageLSN — the LSN of the last log record applied to it.

This is how we know if a page is up-to-date:
  if page.pageLSN >= log_record.LSN:
      this log record already applied to this page → skip (idempotent)
  else:
      apply this log record to the page (REDO)
      
LSNs make recovery IDEMPOTENT — you can replay the log multiple times safely.
```

---

## 3. ARIES Recovery Algorithm

**ARIES** (Algorithm for Recovery and Isolation Exploiting Semantics) is the gold standard recovery algorithm used by most databases (DB2, SQL Server, PostgreSQL's approach is ARIES-like).

### Concepts

```
STEAL policy: Can dirty pages from uncommitted transactions be written to disk?
  → YES (steal). Buffer pool can evict dirty pages anytime.
  → Increases buffer pool flexibility.
  → But means disk may contain uncommitted data → need UNDO on crash.

NO-FORCE policy: Must all dirty pages be flushed to disk at commit?
  → NO (no-force). Commit only requires WAL flush.
  → Faster commits (no random I/O to flush data pages).
  → But means committed data may only be in WAL → need REDO on crash.

ARIES uses STEAL + NO-FORCE → needs both REDO and UNDO.
This is the combination that gives you the best RUNTIME performance
(fast commits, flexible buffer pool) at the cost of complex recovery.
```

### Data Structures for Recovery

```
1. WAL (Log) — on disk
   Append-only file of all changes.

2. Transaction Table (in memory)
   Tracks active transactions:
   ┌─────────┬─────────┬──────────┐
   │ txn_id  │ status  │ lastLSN  │
   ├─────────┼─────────┼──────────┤
   │ TX100   │ running │ LSN 450  │
   │ TX101   │ running │ LSN 480  │
   │ TX102   │ committing│ LSN 500│
   └─────────┴─────────┴──────────┘

3. Dirty Page Table (in memory)
   Tracks modified pages NOT yet written to disk:
   ┌──────────┬────────────┐
   │ page_id  │ recLSN     │  (first LSN that dirtied this page)
   ├──────────┼────────────┤
   │ page 42  │ LSN 300    │
   │ page 15  │ LSN 350    │
   │ page 88  │ LSN 410    │
   └──────────┴────────────┘
   recLSN = earliest log record that might need to be applied to this page
```

### Checkpointing

```
A checkpoint saves a snapshot of the transaction table and dirty page table to the WAL.

Purpose: limit how far back recovery needs to scan the log.

Fuzzy Checkpoint (used in practice):
  1. Write a BEGIN_CHECKPOINT record
  2. Write the transaction table and dirty page table to the log
  3. Write an END_CHECKPOINT record
  4. Update the master record (special fixed location on disk) to point to BEGIN_CHECKPOINT
  
  The database does NOT stop during checkpoint.
  Pages are NOT flushed during checkpoint.
  It's just a snapshot of what's active and dirty.

Recovery starts from the last checkpoint's BEGIN_CHECKPOINT LSN.
```

### The Three Phases of ARIES Recovery

```
  ┌──────────────────────────────────────────────────────────┐
  │                        WAL                                │
  │  ... [CKPT] ... [changes] ... [changes] ... [CRASH]      │
  │       ↑                                         ↑        │
  │   checkpoint                                  end of log  │
  └──────────────────────────────────────────────────────────┘

  Phase 1: ANALYSIS  ──→  (forward scan from checkpoint)
  Phase 2: REDO      ──→  (forward scan from min(recLSN))
  Phase 3: UNDO      ←──  (backward scan from end of log)
```

#### Phase 1: ANALYSIS

```
Purpose: Figure out what needs to be redone and undone.

Scan forward from the last checkpoint to the end of the log.

1. Reconstruct the transaction table:
   - BEGIN records → add transaction to table (status: running)
   - COMMIT records → mark transaction as committed
   - ABORT records → mark transaction as aborted
   
2. Reconstruct the dirty page table:
   - For each log record that modifies a page:
     If page NOT in dirty page table → add it (recLSN = this LSN)
   
3. At the end of analysis:
   - Transaction table tells us: which transactions were active at crash
   - Dirty page table tells us: which pages might be out of date on disk
   - The REDO starting point = min(recLSN) across all dirty pages
```

#### Phase 2: REDO (Repeating History)

```
Purpose: Bring ALL pages to the state they were in at crash time.
         Apply changes for BOTH committed AND uncommitted transactions.
         (Uncommitted ones will be undone in Phase 3.)

Scan forward from min(recLSN) to end of log:

For each log record:
  1. Is the page in the dirty page table?
     NO → skip (page was already flushed to disk)
  
  2. Is the log record's LSN < the page's recLSN?
     YES → skip (this change was already on disk before the page became dirty)
  
  3. Read the page from disk. Is page.pageLSN >= this LSN?
     YES → skip (page already has this change)
  
  4. Apply the change. Update page.pageLSN.

After REDO: the database is in exactly the state it was at crash time.
             Including uncommitted transaction changes!
```

#### Phase 3: UNDO (Rolling Back Losers)

```
Purpose: Remove all changes made by transactions that didn't commit.

"Loser" transactions = active (not committed/aborted) from the transaction table.

For each loser transaction, scan backward through its log records:
  1. For each change: apply the INVERSE operation
     INSERT → DELETE the row
     UPDATE → restore old value
     DELETE → re-INSERT the row
  
  2. Write a CLR (Compensation Log Record) for each undo action
     CLR says: "I undid this change" + undo_next_LSN (pointer to next record to undo)
     CLRs make undo REPEATABLE — if we crash DURING recovery, CLRs tell us what was already undone

  3. Write an ABORT record when transaction is fully undone.
  4. Write an END record.

After UNDO: all uncommitted transaction effects are removed.
            Database is in a consistent state.
```

### Recovery Example

```
Log records:
  LSN 1:  [TX1, BEGIN]
  LSN 2:  [TX1, UPDATE page 5, old=A, new=B]
  LSN 3:  [TX2, BEGIN]
  LSN 4:  [TX2, INSERT page 8, new=X]
  LSN 5:  [CHECKPOINT, active={TX1, TX2}, dirty={page 5: recLSN=2, page 8: recLSN=4}]
  LSN 6:  [TX1, UPDATE page 5, old=B, new=C]
  LSN 7:  [TX1, COMMIT]
  LSN 8:  [TX2, UPDATE page 8, old=X, new=Y]
  ───── CRASH ─────

ANALYSIS (scan from checkpoint LSN 5):
  Transaction table:
    TX1: committed (saw COMMIT at LSN 7)
    TX2: running (no commit) → LOSER
  Dirty page table:
    page 5: recLSN=2
    page 8: recLSN=4

REDO (scan from min(recLSN)=2):
  LSN 2: redo if page 5 pageLSN < 2 → apply UPDATE A→B
  LSN 4: redo if page 8 pageLSN < 4 → apply INSERT X
  LSN 6: redo if page 5 pageLSN < 6 → apply UPDATE B→C
  LSN 8: redo if page 8 pageLSN < 8 → apply UPDATE X→Y

  After REDO:
    page 5 = C (TX1's changes, committed)
    page 8 = Y (TX2's changes, NOT committed)

UNDO (loser = TX2):
  LSN 8: undo UPDATE Y→X on page 8, write CLR
  LSN 4: undo INSERT X on page 8, write CLR
  Write ABORT for TX2, then END.

  After UNDO:
    page 5 = C ✓ (TX1 committed)
    page 8 = empty ✓ (TX2 rolled back)
```

---

## 4. PostgreSQL Recovery Specifics

```
PostgreSQL's recovery is ARIES-like but with important differences:

1. NO UNDO LOG
   PostgreSQL doesn't need transaction undo because of MVCC:
   - Old row versions are kept in the heap (not overwritten)
   - Uncommitted transactions' rows are simply invisible to other transactions
   - VACUUM cleans up later
   
   → Recovery only needs REDO (not UNDO)!
   → Simpler than full ARIES.

2. WAL record types:
   - Full page images: after a checkpoint, the FIRST modification to a page 
     writes the ENTIRE page to WAL (not just the diff).
     This protects against torn pages (partial page write).
   
   - Subsequent modifications: only write the diff (efficient).
   
   full_page_writes = on (default, NEVER turn off)

3. Recovery process:
   a. Read pg_control file → find last checkpoint location
   b. Read checkpoint record → get redo start position
   c. Replay WAL from redo position to end
   d. Done! (No undo phase needed)

4. Timeline and WAL segments:
   PostgreSQL WAL is divided into 16 MB segment files:
     000000010000000000000001
     000000010000000000000002
     ...
   
   After recovery from a backup, a new TIMELINE is created
   (prevents confusion between old and new WAL sequences).
```

### Point-in-Time Recovery (PITR)

```
1. Take a base backup (pg_basebackup)
   → Copies all data files + notes the WAL position

2. Archive WAL segments continuously
   archive_command = 'cp %p /backup/wal/%f'

3. To recover to a specific time:
   a. Restore base backup
   b. Configure recovery:
      restore_command = 'cp /backup/wal/%f %p'
      recovery_target_time = '2026-04-17 14:30:00'
   c. Start PostgreSQL → replays WAL up to that time
   
   This can recover from accidental DROP TABLE, bad DML, etc.
   As long as you have the WAL segments covering that period.
```

---

## 5. MySQL/InnoDB Recovery

```
InnoDB uses a classic ARIES model with REDO + UNDO:

REDO LOG (ib_logfile0, ib_logfile1):
  - Fixed-size circular files
  - Stores physical changes (page-level diffs)
  - Used for REDO during crash recovery
  
UNDO LOG (stored in system tablespace or undo tablespaces):
  - Stores old row versions for MVCC + rollback
  - Used for UNDO during crash recovery (rollback uncommitted transactions)
  - Also used at runtime: old versions for consistent reads

DOUBLE-WRITE BUFFER:
  Problem: a 16 KB InnoDB page written to disk might be only partially written 
           if crash occurs during write (torn page).
           The redo log can't fix a torn page (it stores diffs, not full pages).
  
  Solution: before writing pages to their final location,
            write them to a contiguous "doublewrite buffer" area on disk first.
            If a torn page is detected during recovery,
            the clean copy from the doublewrite buffer is used.
  
  PostgreSQL's approach: full_page_writes (store entire page in WAL after checkpoint)
  InnoDB's approach: doublewrite buffer (separate area on disk)

Recovery process:
  1. REDO: replay redo log from last checkpoint
  2. UNDO: rollback uncommitted transactions using undo log
  3. Purge: clean up old undo log entries that are no longer needed
```

---

## 6. Checkpointing Strategies

```
SHARP CHECKPOINT:
  - Stop all operations
  - Flush ALL dirty pages to disk
  - Write checkpoint record
  - Pro: recovery is trivial (everything on disk is consistent)
  - Con: stalls the entire database during checkpoint
  - Used by: nobody in production

FUZZY CHECKPOINT:
  - Write checkpoint record with current dirty page table + transaction table
  - Continue normal operations
  - Dirty pages flushed gradually in background
  - Pro: no stall
  - Con: more complex recovery (must replay more log)
  - Used by: ARIES, PostgreSQL, InnoDB

PostgreSQL checkpoint process:
  1. Spread dirty page writes over checkpoint_completion_target × checkpoint_timeout
     Default: 0.9 × 5 min = writes spread over 4.5 minutes
  2. Avoids I/O spikes
  3. Configurable: checkpoint_timeout (default 5 min), max_wal_size (triggers checkpoint)

InnoDB checkpoint types:
  - Fuzzy checkpoint (normal operation)
  - Sharp checkpoint (only at shutdown: innodb_fast_shutdown = 0)
```

---

## 7. fsync and Durability Guarantees

```
The OS lies to you. A write() call returns success but data is in OS cache, not on disk.
fsync() FORCES the OS to flush data from cache to physical disk.

Without fsync: committed data can be lost on OS crash (even if database is fine).

PostgreSQL:
  wal_sync_method = fdatasync (Linux default)
  Options: fsync, fdatasync, open_datasync, open_sync
  
  fsync = on  (NEVER set this to off in production!)

The fsync problem:
  fsync is the bottleneck for transaction throughput.
  Each commit = one fsync = one disk flush latency.
  
  HDD: ~10 ms per fsync → max ~100 commits/sec per disk
  SSD: ~0.1 ms per fsync → max ~10,000 commits/sec per disk
  
  Group commit: batch multiple transaction WAL writes into one fsync
  → 1000 transactions × 1 fsync instead of 1000 fsyncs
  → Dramatically improves throughput

Battery-Backed Write Cache (BBWC):
  Server-grade RAID controllers have battery-backed cache.
  Write = cache (microseconds) → battery protects data until flush to disk.
  fsync returns immediately because cache IS durable.
  → 100x+ improvement in fsync-heavy workloads.
  
  Consumer SSDs: may or may not honor fsync properly.
  Enterprise SSDs: have power-loss protection capacitors.
```

---

## 8. Practice Questions

**Q1.** After a crash, the recovery system find these log records:
```
LSN 10: [TX1, UPDATE page 3, A→B]
LSN 20: [TX2, INSERT page 7, new=X]
LSN 30: [TX1, COMMIT]
LSN 40: [TX2, UPDATE page 7, X→Y]
--- CRASH ---
```
Which changes are REDOne and which are UNDOne?

<details><summary>Answer</summary>

REDO (all records, including uncommitted):
- LSN 10: UPDATE page 3 A→B (for TX1)
- LSN 20: INSERT page 7 X (for TX2)
- LSN 30: COMMIT TX1 (just a log record)
- LSN 40: UPDATE page 7 X→Y (for TX2)

UNDO (only TX2, which didn't commit):
- LSN 40: undo UPDATE page 7 Y→X
- LSN 20: undo INSERT page 7
- Write ABORT for TX2

Final state:
- page 3 = B (TX1 committed)
- page 7 = empty (TX2 rolled back)
</details>

**Q2.** Why does PostgreSQL not need an UNDO phase during recovery?

<details><summary>Answer</summary>

PostgreSQL uses MVCC with in-place versioning. When a row is updated, the old version remains in the heap with xmin/xmax metadata. Uncommitted transactions' new rows are simply invisible to other transactions via the visibility rules (xmin not committed → not visible).

After crash recovery (REDO only), uncommitted rows exist on disk but are invisible because:
1. Their xmin transaction ID is not in the commit log (pg_xact/pg_clog)
2. VACUUM will eventually remove them

So the UNDO is effectively "lazy" — handled by VACUUM later, not during recovery.
</details>

**Q3.** What is the torn page problem and how do PostgreSQL and InnoDB solve it differently?

<details><summary>Answer</summary>

A torn page occurs when a crash happens mid-write of a page. For example, a 16 KB page is half-written: first 8 KB is new data, last 8 KB is old data. The page is now corrupt and can't be fixed by applying redo log diffs (which assume a consistent base page).

**PostgreSQL:** `full_page_writes = on`. After each checkpoint, the first time a page is modified, the ENTIRE page image is written to WAL. During recovery, this full page image is used as the base, then subsequent diffs are applied on top. No torn page possible.

**InnoDB:** Doublewrite buffer. Before flushing dirty pages to their final data file locations, InnoDB writes them to a sequential "doublewrite buffer" area on disk. If a torn page is detected at recovery, the clean copy from the doublewrite buffer is used to restore the page, then redo log is applied.
</details>

---

## Key Takeaways

1. **WAL protocol:** Log before data. Commit = WAL flush. Data pages written lazily.
2. **ARIES:** three phases — Analysis (what happened?), Redo (replay all), Undo (rollback losers).
3. **ARIES uses STEAL/NO-FORCE** — best runtime performance, complex recovery.
4. **LSNs provide idempotency** — replaying the log multiple times is safe.
5. **PostgreSQL skips UNDO** because MVCC makes uncommitted rows invisible. Simpler recovery.
6. **Checkpoints** limit how much log must be replayed. Fuzzy checkpoints avoid stalls.
7. **fsync** is the durability guarantee. Group commit is the performance solution.
8. **Torn pages** are solved by full-page writes (PG) or doublewrite buffer (InnoDB).

---

Next: [06-memory-management.md](06-memory-management.md) →
