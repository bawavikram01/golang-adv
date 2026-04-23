# 6.3 — Research Papers That Shaped Databases

> Every database you use today exists because someone wrote a paper.  
> Codd invented relational theory. Gray invented transactions.  
> Stonebraker built PostgreSQL. Google built the cloud-native era.  
> Read the papers. Stand on the shoulders of giants.

---

## 1. The Foundations (1970s–1990s)

### "A Relational Model of Data for Large Shared Data Banks" — Codd (1970)

```
THE paper. The Big Bang of relational databases.

Before Codd: databases were navigational (IMS, CODASYL).
  Programmers wrote code to traverse pointers between records.
  If the data structure changed → all programs broke.

Codd's insight:
  Separate the LOGICAL representation (relations/tables)
  from the PHYSICAL storage (how data is stored on disk).
  Let the system figure out HOW to retrieve data.
  The programmer specifies WHAT data they want.

Key contributions:
  1. Relations (tables) as the data model
  2. Relational algebra (formal operations: select, project, join)
  3. Data independence (logical schema ≠ physical storage)
  4. Normalization theory (eliminate redundancy)

This paper led to: SQL, every RDBMS ever built, 50+ years of industry.

Read: https://www.seas.upenn.edu/~zives/03f/cis550/codd.pdf
Impact: ★★★★★ — Nothing in databases exists without this paper.
```

### "Access Path Selection in a Relational DBMS" — Selinger et al. (1979)

```
How should a database execute a query? This paper answers that question.
The System R optimizer — the first cost-based query optimizer.

Key contributions:
  1. Cost-based optimization: estimate cost of different plans, pick cheapest
  2. Statistics: use histograms and cardinality estimates
  3. Dynamic programming for join ordering
  4. Interesting orders: a sort order produced by one operation
     might benefit a later operation (e.g., sort-merge join)

This is STILL how PostgreSQL, MySQL, Oracle, SQL Server optimize queries.
Every optimizer lecture starts with this paper.

Read: https://courses.cs.duke.edu/compsci516/cps216/spring03/papers/selinger-etal-1979.pdf
Impact: ★★★★★ — The foundation of every query optimizer.
```

### "ARIES: A Transaction Recovery Method" — Mohan et al. (1992)

```
THE recovery algorithm. How databases survive crashes.

Before ARIES: various ad-hoc recovery schemes, none complete.
ARIES unified everything into one elegant algorithm.

Three phases:
  1. Analysis: scan log, determine what was active at crash
  2. Redo: replay history (restore database to pre-crash state)
  3. Undo: roll back uncommitted transactions

Key innovations:
  - Write-Ahead Logging (WAL): log before modifying pages
  - STEAL/NO-FORCE policy: maximum buffer pool flexibility
  - Physiological logging: physical (page-level) + logical (operation-level)
  - Compensation Log Records (CLRs): make undo idempotent
  - Fuzzy checkpointing: checkpoint without stopping the world

Used by: PostgreSQL, MySQL/InnoDB, SQL Server, DB2, Oracle.
Every database you've used implements ARIES (or a close variant).

Read: https://cs.stanford.edu/people/chr101/cs345/aries.pdf
Impact: ★★★★★ — Without ARIES, databases couldn't reliably recover.
```

### "The Design of POSTGRES" — Stonebraker & Rowe (1986)

```
The design paper for the system that became PostgreSQL.

Michael Stonebraker at UC Berkeley designed POSTGRES (Post-Ingres) with:
  - Extensible type system (user-defined types, operators, functions)
  - No-overwrite storage (time-travel queries — early MVCC)
  - Rules system (instead of triggers)
  - Abstract data types

Many ideas were ahead of their time:
  - JSON support? PostgreSQL's extensibility makes it natural
  - PostGIS? User-defined types + operators + index support
  - Full-text search? Custom access methods

This paper explains WHY PostgreSQL is so extensible —
it was designed from the ground up to be extensible.

Read: https://dsf.berkeley.edu/papers/ERL-M85-95.pdf
Impact: ★★★★☆ — Directly led to the most advanced open-source database.
```

---

## 2. The Distributed Era (2000s–2010s)

### "Bigtable: A Distributed Storage System" — Chang et al. (2006)

```
Google's Bigtable: the paper that launched the NoSQL movement.

Model:
  - Sparse, distributed, persistent multidimensional sorted map
  - (row_key, column_family:column, timestamp) → value
  - Rows sorted lexicographically by key
  - Tablets (row ranges) distributed across tablet servers

Key ideas:
  - Column families (group related columns for locality)
  - Tablets: auto-split, auto-balance across servers
  - SSTable format (immutable sorted file — basis of LSM trees in practice)
  - GFS (distributed file system) underneath
  - Single-row transactions only (no cross-row ACID)

Inspired: HBase, Cassandra (partially), LevelDB, RocksDB.
The LSM-tree became ubiquitous because of this paper.

Read: https://static.googleusercontent.com/media/research.google.com/en//archive/bigtable-osdi06.pdf
Impact: ★★★★★ — Started the NoSQL era.
```

### "Dynamo: Amazon's Highly Available Key-Value Store" — DeCandia et al. (2007)

```
Amazon's internal key-value store. Prioritize availability above all.

Design decisions (every one is a trade-off):
  - Consistent hashing: distribute data across nodes in a ring
  - Vector clocks: detect conflicts (allow concurrent writes)
  - Sloppy quorum + hinted handoff: write even if some nodes are down
  - Anti-entropy: Merkle trees for background consistency repair
  - Application-level conflict resolution ("last writer wins" OR app decides)

The opposite philosophy from Spanner:
  Spanner: sacrifice latency for consistency
  Dynamo: sacrifice consistency for availability

Inspired: Cassandra (directly), Riak, Voldemort, DynamoDB.
Every discussion of "eventual consistency" starts with Dynamo.

Read: https://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf
Impact: ★★★★★ — Defined the eventually-consistent, AP side of databases.
```

### "Spanner: Google's Globally-Distributed Database" — Corbett et al. (2012)

```
Google Spanner: the first globally-distributed database with external consistency.
"Wait, you can have distribution AND strong consistency?"

The breakthrough: TrueTime API
  - GPS receivers + atomic clocks in every datacenter
  - TrueTime.now() returns an interval [earliest, latest]
  - Uncertainty bounded to ~7ms (much less than network RTT)
  - Commit protocol: wait out the uncertainty interval
  - Result: GLOBAL ordering of transactions (external consistency)

Key ideas:
  - Semi-relational model (SQL + strong reads + ACID)
  - Paxos-based replication (per-shard)
  - Two-phase commit across Paxos groups (distributed ACID)
  - Read-only transactions at a timestamp (no locks needed)
  - Schema-declared hierarchy for locality (interleaved tables)

This paper proved:
  The CAP theorem has nuance. With specialized hardware (atomic clocks),
  you CAN have strong consistency + global distribution + SQL.

Inspired: CockroachDB (without atomic clocks — uses hybrid logical clocks),
  YugabyteDB, Cloud Spanner (the commercial product).

Read: https://static.googleusercontent.com/media/research.google.com/en//archive/spanner-osdi2012.pdf
Impact: ★★★★★ — Changed the industry's understanding of what's possible.
```

---

## 3. Storage & Transaction Innovations

### "The Log-Structured Merge-Tree (LSM-Tree)" — O'Neil et al. (1996)

```
The alternative to B-trees for write-heavy workloads.

Problem: B-trees require random I/O for writes (update a page anywhere on disk).
Solution: Buffer all writes in memory, flush to disk as sorted runs, merge later.

Structure:
  Memory:  MemTable (sorted, in-memory — skip list or red-black tree)
  Level 0: Small sorted files (recently flushed)
  Level 1: Larger sorted files (merged from L0)
  Level N: Very large sorted files

Write path: Write to MemTable → flush to L0 → compaction merges L0→L1→...
Read path: Check MemTable → L0 → L1 → ... (bloom filters skip empty levels)

Trade-offs vs B-tree:
  Writes: LSM much faster (sequential I/O only, ~10-100x for writes)
  Reads: LSM slower (may check multiple levels, ~2-5x for point reads)
  Space: LSM uses more space temporarily (before compaction)
  Space amplification: multiple copies of same key across levels

Used by: RocksDB, LevelDB, Cassandra, HBase, TiKV, CockroachDB, ScyllaDB.

Read: https://www.cs.umb.edu/~poneil/lsmtree.pdf
Impact: ★★★★★ — The foundation of every modern write-optimized store.
```

### "Calvin: Fast Distributed Transactions" — Thomson et al. (2012)

```
A radical alternative to traditional distributed transactions.

Conventional approach (Spanner-style):
  Pessimistic locking or OCC + two-phase commit → high latency, complex

Calvin's approach:
  1. Deterministic ordering: all nodes agree on transaction order BEFORE executing
  2. Sequencer: assigns global order to all transactions
  3. Every node executes the SAME transactions in the SAME order
  4. No coordination during execution (order already agreed upon)
  5. No two-phase commit needed!

Trade-offs:
  ✓ Very low latency for transactions
  ✓ No 2PC overhead
  ✗ Transactions must declare read/write sets upfront
  ✗ Aborts are expensive (replay everything after the aborted txn)
  ✗ Single point of ordering (sequencer) can be bottleneck

Inspired: FaunaDB/Fauna, academic systems.
An important alternative mental model to Spanner's approach.

Read: https://cs.yale.edu/homes/thomson/publications/calvin-sigmod12.pdf
Impact: ★★★★☆ — Showed that deterministic execution eliminates coordination.
```

---

## 4. Cloud-Native Databases (2017–2021)

### "Amazon Aurora: Design Considerations for High Throughput Cloud-Native Relational Databases" (2017)

```
Aurora's key insight: the log IS the database.

Traditional: database writes data pages + WAL to storage
Aurora: database writes ONLY the WAL to storage
  → Storage nodes reconstruct pages from WAL on demand
  → 6 copies across 3 AZs, 4/6 quorum writes, 3/6 quorum reads
  → Network I/O reduced by 7.7x (only ship log records, not full pages)

Architecture:
  Compute layer: MySQL/PostgreSQL-compatible, stateless(ish)
  Storage layer: distributed, replicated, auto-healing

Why this matters:
  - Writes are tiny (log records) vs huge (full pages)
  - Storage auto-scales, auto-heals, auto-replicates
  - Failover is fast (new compute, same storage)
  - Backups are continuous and free (always have all log records)

Read: https://web.stanford.edu/class/cs245/readings/aurora.pdf
Impact: ★★★★★ — Redefined cloud-native database architecture.
```

### "CockroachDB: The Resilient Geo-Distributed SQL Database" (2020)

```
Open-source Spanner (without atomic clocks).

Key differences from Spanner:
  - Uses hybrid logical clocks (HLC) instead of TrueTime
  - Uncertainty interval is larger (~250ms vs ~7ms)
  - Serializable Snapshot Isolation (SSI) vs strict serializability
  - Read refreshes: if uncertainty detected, retry read at later timestamp
  - Raft (not Paxos) for replication

Architecture:
  SQL Layer → Distribution Layer → Replication Layer → Storage (Pebble/RocksDB)

Interesting engineering:
  - Range splits: automatic splitting when a range gets too big
  - Leaseholder: one replica serves reads (avoids Raft roundtrip for reads)
  - Follower reads: read from any replica if staleness is acceptable
  - Closed timestamps: enable consistent reads from followers

Read: https://dl.acm.org/doi/pdf/10.1145/3318464.3386134
Impact: ★★★★☆ — Proved Spanner's ideas work without specialized hardware.
```

### "Socrates: The New SQL Server in the Cloud" — Antonopoulos et al. (2019)

```
Microsoft's cloud-native rearchitecture of SQL Server for Azure.

Key idea: disaggregate EVERYTHING.
  Traditional: compute + storage + log on one machine
  Socrates: separate compute, log, page server, storage

  ┌──────────┐  ┌──────────┐  ┌──────────┐
  │ Compute  │  │ Compute  │  │ Compute  │  ← Stateless SQL engines
  └────┬─────┘  └────┬─────┘  └────┬─────┘
       │             │             │
  ┌────┴─────────────┴─────────────┴────┐
  │          Landing Zone (XLOG)         │  ← Fast log storage
  └──────────────────┬──────────────────┘
                     │
  ┌──────────────────┴──────────────────┐
  │          Page Servers                │  ← Reconstruct pages from log
  └──────────────────┬──────────────────┘
                     │
  ┌──────────────────┴──────────────────┐
  │          Azure Storage (XStore)      │  ← Cheap, durable storage
  └─────────────────────────────────────┘

Similar to Aurora but more granular disaggregation.
Enables independent scaling of compute, caching, and storage.

Read: https://www.microsoft.com/en-us/research/uploads/prod/2019/05/socrates.pdf
Impact: ★★★★☆ — Shows the industry consensus on disaggregated architecture.
```

---

## 5. How to Read Database Papers

```
Strategy for reading academic papers:

1. First pass (10 minutes):
   - Read title, abstract, introduction, conclusion
   - Look at figures and diagrams
   - Read section headings
   - Goal: understand WHAT the paper is about and WHY it matters

2. Second pass (1 hour):
   - Read the full paper, skip proofs and detailed algorithms
   - Understand the system architecture
   - Note the key contributions (usually listed in the introduction)
   - Understand the evaluation (what did they measure?)

3. Third pass (2-4 hours, only for important papers):
   - Understand every detail
   - Reproduce the mental model
   - Can you explain this to someone else?
   - What are the weaknesses? What would you do differently?

For database papers specifically:
  - Focus on the ARCHITECTURE section
  - Study the TRADE-OFFS (what did they sacrifice? why?)
  - Look at the EVALUATION (what workload? what hardware? what comparison?)
  - Check: does this match what the real system does today?
    (Papers describe the system at one point in time; code evolves)

Where to find papers:
  - SIGMOD (ACM Conference on Management of Data)
  - VLDB (Very Large Data Bases)
  - ICDE (International Conference on Data Engineering)
  - OSDI / SOSP (for systems papers with database components)
  - arxiv.org/list/cs.DB/recent
  - db.cs.cmu.edu/papers/ (Andy Pavlo's curated list)
  - the morning paper (blog, now archived but excellent backlog)
```

---

## 6. Papers Reading List (Prioritized)

```
Tier 1 — Must Read (understand these cold):
  □ Codd (1970) — Relational model
  □ Selinger et al. (1979) — Query optimization
  □ Mohan et al. (1992) — ARIES recovery
  □ O'Neil et al. (1996) — LSM-tree
  □ DeCandia et al. (2007) — Dynamo
  □ Chang et al. (2006) — Bigtable
  □ Corbett et al. (2012) — Spanner
  □ Verbitski et al. (2017) — Aurora
  □ Hellerstein et al. (2007) — Architecture of a Database System (survey)

Tier 2 — Should Read:
  □ Stonebraker & Rowe (1986) — POSTGRES design
  □ Thomson et al. (2012) — Calvin
  □ Taft et al. (2020) — CockroachDB
  □ Antonopoulos et al. (2019) — Socrates (Azure SQL)
  □ Dageville et al. (2016) — Snowflake
  □ Pavlo & Aslett (2016) — What's Really New with NewSQL?
  □ Stonebraker (2019) — Looking Back at Postgres
  □ Wu et al. (2017) — MVCC evaluation

Tier 3 — Deep Dives:
  □ Graefe (1993) — Query evaluation techniques (volcano model)
  □ Neumann (2011) — Efficiently compiling efficient query plans (HyPer)
  □ Kemper & Neumann (2011) — HyPer: hybrid OLTP/OLAP
  □ Harizopoulos et al. (2008) — OLTP through the looking glass
  □ Berenson et al. (1995) — A critique of ANSI SQL isolation levels
  □ Napa (2021) — Google's analytics system

Estimated time: 1 paper per week → Tier 1 done in ~2 months.
                1 paper per week → All tiers in ~6 months.
```

---

## Key Takeaways

1. **Codd (1970) started everything.** The relational model separates logical from physical — this is why you write SQL instead of traversing pointers. Every database since is a footnote to Codd.
2. **ARIES (1992) is how crash recovery works.** WAL + Analysis + Redo + Undo. Every database implements this. Understand it and you understand durability.
3. **Dynamo (2007) vs. Spanner (2012)** represent the two poles of distributed databases: eventual consistency + availability vs. strong consistency + specialized hardware. Every distributed DB picks a point on this spectrum.
4. **Aurora (2017)** showed that "the log is the database" — ship only WAL records to storage, not full pages. This architecture is now the industry standard for cloud databases.
5. **Read one paper per week.** Start with Tier 1. In 2 months you'll understand more about databases than most senior engineers.

---

Next: [04-books.md](04-books.md) →
