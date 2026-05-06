# 4.5 — Wide-Column Stores

> Wide-column stores handle massive scale: petabytes, millions of ops/sec.  
> They sacrifice relational flexibility for horizontal write scalability.  
> If you need to write 100K events/sec and never lose data: this is the model.

---

## 1. Apache Cassandra

### Architecture — Masterless Ring

```
Cassandra has NO leader, NO single point of failure.
Every node is equal ("peer-to-peer").

        ┌───── Node A ─────┐
       /                     \
  Node F                    Node B
      |     Gossip          |
      |     protocol        |
  Node E                    Node C
       \                     /
        └───── Node D ─────┘

Data placement: consistent hashing with virtual nodes (vnodes)
  - Each node owns multiple token ranges on the ring
  - Default: 256 vnodes per node
  - Partition key → token (Murmur3 hash) → placed on owning node
  - Replicated to N-1 additional nodes per replication strategy

Replication strategies:
  SimpleStrategy: N replicas on next N-1 nodes clockwise (single DC)
  NetworkTopologyStrategy: specify replicas per datacenter
    CREATE KEYSPACE myks WITH replication = {
      'class': 'NetworkTopologyStrategy',
      'us-east': 3, 'eu-west': 3
    };

Gossip:
  Every second, each node gossips with 1-3 random peers
  Shares: node status, schema version, load, tokens
  Failure detection: φ-accrual detector (adaptive thresholds)
```

### Data Model

```
Keyspace → Table → Partition → Rows → Columns

CREATE KEYSPACE ecommerce WITH replication = {
    'class': 'NetworkTopologyStrategy', 'us-east': 3
};

CREATE TABLE ecommerce.orders_by_customer (
    customer_id UUID,
    order_date TIMESTAMP,
    order_id UUID,
    total DECIMAL,
    status TEXT,
    PRIMARY KEY (customer_id, order_date, order_id)
) WITH CLUSTERING ORDER BY (order_date DESC, order_id ASC);

-- PRIMARY KEY = (partition_key, clustering_columns)
-- customer_id:  partition key → determines which node(s)
-- order_date:   first clustering column → sort order within partition
-- order_id:     second clustering column → tiebreaker

-- Physical storage:
-- Each partition is stored as a CONTIGUOUS sorted run of rows
-- Rows within a partition are sorted by clustering columns

-- Partition: customer_id = Alice
--   ┌─────────────────────────────────────────────────┐
--   │ 2024-06-15 | order-789 | $99.99  | completed    │
--   │ 2024-06-10 | order-456 | $49.99  | shipped      │
--   │ 2024-05-01 | order-123 | $29.99  | completed    │
--   └─────────────────────────────────────────────────┘
--   → Sequential read for one customer's recent orders = FAST
```

### CQL (Cassandra Query Language)

```sql
-- CQL looks like SQL but with strict constraints:

-- Queries MUST include the full partition key:
SELECT * FROM orders_by_customer WHERE customer_id = ?;                    -- OK
SELECT * FROM orders_by_customer WHERE customer_id = ? AND order_date > ?; -- OK
SELECT * FROM orders_by_customer WHERE order_date > '2024-01-01';          -- REJECTED!
-- (Would require scanning ALL partitions → full cluster scan)

-- Clustering column restrictions must be IN ORDER:
SELECT * FROM t WHERE pk = ? AND ck1 = ? AND ck2 > ?;  -- OK
SELECT * FROM t WHERE pk = ? AND ck2 > ?;               -- REJECTED (skipped ck1)

-- No JOINs, no subqueries, no aggregations across partitions
-- GROUP BY only within a partition
-- This is BY DESIGN — it forces you to model data for your queries

-- Lightweight Transactions (LWT) — compare-and-set:
INSERT INTO users (id, email) VALUES (?, ?)
IF NOT EXISTS;  -- uses Paxos (slow! ~4x regular write latency)

-- Materialized Views:
CREATE MATERIALIZED VIEW orders_by_status AS
    SELECT * FROM orders_by_customer
    WHERE status IS NOT NULL AND customer_id IS NOT NULL
    PRIMARY KEY (status, customer_id, order_date, order_id);
-- Automatically maintained by Cassandra (eventual consistency with base table)
-- CAUTION: known consistency issues — many teams avoid MVs
```

### Consistency Levels

```
Tunable consistency per query:

Consistency Level  Nodes contacted   Guarantee
───────────────── ────────────────── ─────────────────
ONE               1                  Fastest, weakest
TWO               2                  Better than ONE
THREE             3                  Better than TWO
QUORUM            ⌊N/2⌋ + 1         Strong if R+W > N (default)
LOCAL_QUORUM      quorum in local DC Best for multi-DC
EACH_QUORUM       quorum in each DC  Strongest multi-DC
ALL               All replicas       Strongest, slowest, least available

Typical production setup (N=3, multi-DC):
  Writes: LOCAL_QUORUM (2/3 in local DC)
  Reads:  LOCAL_QUORUM (2/3 in local DC)
  → Strong consistency within a DC, eventual across DCs
```

### Write Path

```
Client → Coordinator node → Replica nodes

1. Coordinator receives write
2. Determines replica nodes (consistent hashing)
3. Sends write to all replicas simultaneously
4. Each replica:
   a. Write to commit log (WAL — sequential I/O)
   b. Write to MemTable (in-memory sorted structure)
   c. Acknowledge to coordinator
5. Coordinator responds to client when CL is satisfied

MemTable → flush → SSTable (immutable sorted file on disk)
SSTables → compaction → merged SSTables

Compaction strategies:
  SizeTiered (STCS):  merge SSTables of similar size (write-optimized)
  Leveled (LCS):      non-overlapping levels like RocksDB (read-optimized)
  TimeWindow (TWCS):  for time-series (compact within time windows)
```

### Read Path

```
1. Check MemTable (in-memory)
2. Check row cache (if enabled)
3. Check Bloom filters of SSTables (skip files that definitely don't have the key)
4. Check partition index summary (sampled index in memory)
5. Check partition index (find byte offset in SSTable)
6. Read data from SSTable
7. Merge results from MemTable + multiple SSTables (newest wins)

Read performance depends on:
  - Number of SSTables to check (compaction keeps this low)
  - Bloom filter effectiveness (false positive rate)
  - Whether data is in page cache (OS caching)
```

---

## 2. ScyllaDB — Cassandra Rewritten in C++

```
ScyllaDB: drop-in Cassandra replacement, 10x faster.

Key differences:
  Cassandra:                     ScyllaDB:
  Java (GC pauses)              C++ (no GC, predictable latency)
  Thread-per-core               Shard-per-core (Seastar framework)
  Shared memory                 No shared memory (each core owns its data)
  Linux page cache              Custom I/O scheduler + bypass page cache

Shard-per-core architecture:
  Core 0 → owns partitions 0-99    (its own MemTable, SSTables, etc.)
  Core 1 → owns partitions 100-199
  Core 2 → owns partitions 200-299
  ...
  
  No locking between cores → linear scalability with core count
  Cross-core requests use message passing (actor model)

Seastar framework:
  - Cooperative scheduling (no OS thread switching)
  - Future/promise async model
  - Zero-copy networking
  - Custom memory allocator (per-core, no malloc contention)

When to choose ScyllaDB over Cassandra:
  ✓ P99 latency matters (no GC pauses)
  ✓ Need more throughput per node (fewer nodes = less operational cost)
  ✓ Team is comfortable with C++ operations
  Cassandra advantage: larger community, more tooling, JMX monitoring
```

---

## 3. Apache HBase

```
HBase: wide-column store on top of HDFS (Hadoop Distributed FileSystem).
Modeled after Google Bigtable (2006 paper).

Architecture:
  ┌──────────┐
  │ HMaster   │  ← region assignment, DDL, load balancing
  └─────┬────┘
        │
  ┌─────▼─────┐  ┌───────────┐  ┌───────────┐
  │RegionServer│  │RegionServer│  │RegionServer│
  │  ┌──────┐ │  │  ┌──────┐ │  │  ┌──────┐ │
  │  │Region│ │  │  │Region│ │  │  │Region│ │
  │  │Region│ │  │  │Region│ │  │  │Region│ │
  │  └──────┘ │  │  └──────┘ │  │  └──────┘ │
  └───────────┘  └───────────┘  └───────────┘
        │              │              │
  ┌─────▼──────────────▼──────────────▼─────┐
  │            HDFS (Hadoop File System)      │
  │  Distributed, replicated file storage     │
  └──────────────────────────────────────────┘

Data model:
  Table → Region (contiguous row range) → Column Families → Columns
  
  Row key | Column Family: Qualifier | Timestamp | Value
  ─────── | ───────────────────────── | ───────── | ───────
  row1    | info:name                 | t3        | "Alice"
  row1    | info:name                 | t1        | "Al"     ← old version!
  row1    | stats:login_count         | t2        | "42"
  
  - Every cell is VERSIONED (multiple timestamps per cell)
  - Column families are defined at schema time
  - Columns (qualifiers) within a family are dynamic (schema-less)
  - Row key is the ONLY index (design it carefully!)

HBase vs Cassandra:
  HBase: master-based (HMaster), CP (consistent), HDFS-dependent
  Cassandra: masterless, AP (available), self-contained
  HBase: better for Hadoop ecosystem integration
  Cassandra: better for geo-distributed, always-available workloads
```

---

## 4. Google Bigtable

```
Google Bigtable: the original wide-column store (2006 paper).
Available as cloud service (Google Cloud Bigtable).

Key characteristics:
  - Single-row transactions only (no multi-row)
  - Rows sorted lexicographically by row key
  - Column families must be pre-defined
  - Millisecond latency for single-row reads/writes
  - Scales to petabytes, millions of ops/sec

Inspired: HBase (open-source Bigtable clone), Cassandra (some concepts)

Cloud Bigtable:
  - Fully managed (no operations)
  - Integrates with BigQuery, Dataflow, Dataproc
  - Linear scalability: add nodes → proportional throughput increase
  - Use case: IoT telemetry, financial time-series, ad-tech, personalization
```

---

## 5. Wide-Column Data Modeling Rules

```
Rule 1: MODEL FOR YOUR QUERIES, NOT YOUR ENTITIES.
  In relational: normalize data, join at query time.
  In wide-column: denormalize everything into query-specific tables.
  
  If you need:
    "Get orders by customer" → orders_by_customer table (PK: customer_id)
    "Get orders by status"   → orders_by_status table (PK: status)
  → TWO tables with duplicated data. This is normal and expected.

Rule 2: PARTITION KEY IS EVERYTHING.
  It determines data distribution AND what queries are possible.
  Bad partition key → hot partitions → entire cluster slows down.
  
  Good partition keys: customer_id, device_id, user_id
  Bad partition keys: status (few values), country (skewed), timestamp (monotonic)

Rule 3: AVOID UNBOUNDED PARTITION GROWTH.
  A partition with 100 million rows degrades (compaction, reads).
  Use compound partition keys: (customer_id, month)
  → Natural time-based bucketing.

Rule 4: DENORMALIZE. ACCEPT DUPLICATE DATA.
  Storage is cheap. Consistency across copies is managed by your application.
  The write amplification of maintaining multiple views is worth the read performance.
```

---

## Key Takeaways

1. **Cassandra is masterless** — every node is equal. No single point of failure. This is fundamentally different from HBase (has master) and MongoDB (has primary).

2. **Model for your queries.** One table per access pattern. Denormalize aggressively. If you need different views, create different tables with the same data.

3. **Partition key determines EVERYTHING**: which node stores the data, which queries are fast, whether you have hot spots. Get this wrong and nothing else matters.

4. **CQL looks like SQL but isn't.** No JOINs, no subqueries, no cross-partition aggregation. These constraints are features — they guarantee predictable performance.

5. **Tunable consistency** (W+R>N) lets you trade consistency for latency per operation. LOCAL_QUORUM is the sweet spot for multi-datacenter deployments.

6. **ScyllaDB's shard-per-core** eliminates GC pauses and lock contention. It's Cassandra's data model with C++ performance.

7. **HBase lives in the Hadoop ecosystem.** If you're already on HDFS/Hadoop, HBase integrates naturally. Otherwise, Cassandra or ScyllaDB is simpler to operate.

---

Next: [06-graph-databases.md](06-graph-databases.md) →
