# 4.2 — Distributed SQL / NewSQL

> The question used to be: "Relational OR scale?"  
> NewSQL answered: "Both."  
> These databases provide SQL + ACID + horizontal scaling.  
> They are the future of transactional databases.

---

## 1. The NewSQL Promise

```
Traditional RDBMS (PostgreSQL, MySQL):
  ✓ SQL, ACID, rich queries
  ✗ Single-node write scalability (scale-up only)
  ✗ Manual sharding is painful

NoSQL (Cassandra, DynamoDB, MongoDB):
  ✓ Horizontal scaling, high availability
  ✗ Weak consistency, no cross-shard transactions
  ✗ Limited query capabilities

NewSQL/Distributed SQL (the best of both):
  ✓ SQL interface (PostgreSQL/MySQL wire protocol)
  ✓ ACID transactions across shards
  ✓ Horizontal scaling (add nodes to scale)
  ✓ Strong consistency (serializable/linearizable)
  ✗ Higher latency per operation (network hops between nodes)
  ✗ More complex operations and failure modes
```

---

## 2. Google Spanner — The Pioneer

```
Spanner (2012): first globally-distributed database with external consistency.

Architecture:
  Zone → SpanServer → Tablets (contiguous range of rows)
  
  Directory (group of related rows) → smallest unit of data movement
  Tablets contain directories
  Each tablet is replicated via Paxos across zones

TrueTime:
  GPS receivers + atomic clocks in every datacenter
  API: TT.now() → [earliest, latest]  (uncertainty ε ≈ 1-7ms)
  
  Commit protocol:
    Coordinator picks commit timestamp s
    Wait until TT.now().earliest > s  ("commit-wait")
    → Guarantees s is in the past for ALL observers worldwide
    → External consistency without global locking!

Read-only transactions:
  Pick timestamp = TT.now().latest
  Read from any replica with data at that timestamp
  → NO LOCKS for reads! Lock-free, globally consistent snapshots.

Spanner SQL:
  Full SQL (Spanner was initially key-value, added SQL later)
  GoogleSQL dialect (not quite PostgreSQL/MySQL compatible)
  Cloud Spanner: 99.999% SLA (5 nines!)

Limitations:
  - Only available as Google Cloud service
  - Expensive
  - TrueTime requires specialized hardware (GPS/atomic clocks)
  - Higher latency for writes (cross-region Paxos + commit-wait)
```

---

## 3. CockroachDB — Spanner for Everyone

```
CockroachDB: open-source distributed SQL inspired by Spanner.
PostgreSQL wire protocol compatible.

Architecture:
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │   Node 1      │  │   Node 2      │  │   Node 3      │
  │ ┌──────────┐  │  │ ┌──────────┐  │  │ ┌──────────┐  │
  │ │SQL Layer │  │  │ │SQL Layer │  │  │ │SQL Layer │  │
  │ │(parser,  │  │  │ │(parser,  │  │  │ │(parser,  │  │
  │ │optimizer,│  │  │ │optimizer,│  │  │ │optimizer,│  │
  │ │executor) │  │  │ │executor) │  │  │ │executor) │  │
  │ ├──────────┤  │  │ ├──────────┤  │  │ ├──────────┤  │
  │ │Transact. │  │  │ │Transact. │  │  │ │Transact. │  │
  │ │Layer     │  │  │ │Layer     │  │  │ │Layer     │  │
  │ │(MVCC,    │  │  │ │(MVCC,    │  │  │ │(MVCC,    │  │
  │ │2PC)      │  │  │ │2PC)      │  │  │ │2PC)      │  │
  │ ├──────────┤  │  │ ├──────────┤  │  │ ├──────────┤  │
  │ │Distrib.  │  │  │ │Distrib.  │  │  │ │Distrib.  │  │
  │ │Layer     │  │  │ │Layer     │  │  │ │Layer     │  │
  │ │(ranges,  │  │  │ │(ranges,  │  │  │ │(ranges,  │  │
  │ │Raft,     │  │  │ │Raft,     │  │  │ │Raft,     │  │
  │ │routing)  │  │  │ │routing)  │  │  │ │routing)  │  │
  │ ├──────────┤  │  │ ├──────────┤  │  │ ├──────────┤  │
  │ │Pebble    │  │  │ │Pebble    │  │  │ │Pebble    │  │
  │ │(storage) │  │  │ │(storage) │  │  │ │(storage) │  │
  │ └──────────┘  │  │ └──────────┘  │  │ └──────────┘  │
  └──────────────┘  └──────────────┘  └──────────────┘

Key concepts:
  RANGE: contiguous sorted key range (~512 MB default)
    → Equivalent to Spanner's tablet
    → Each range stored as a Raft group (3 replicas by default)
    → Leaseholder: one replica handles reads/writes (like Raft leader)

  MVCC: every key-value has a timestamp (HLC — Hybrid Logical Clock)
  
  HLC (no TrueTime, no GPS):
    Combines physical clock + logical counter
    Better than Lamport clocks (real-time component)
    Cannot guarantee linearizability like TrueTime
    → Uses "uncertainty intervals" + transaction restarts
       When reading a key with timestamp in the uncertainty window:
       → Restart transaction at a later timestamp (adds latency)

  Transaction protocol:
    Single-range: single Raft consensus (fast)
    Multi-range:  Parallel commits (optimized 2PC)
      → Write intents to all ranges in parallel
      → Transaction record determines commit/abort
      → Very efficient: often single round trip for 2PC

Locality:
  -- Pin data to specific regions:
  ALTER TABLE users CONFIGURE ZONE USING
    constraints = '[+region=us-east]';
  
  -- Geo-partitioned table:
  CREATE TABLE orders (
    region STRING, id UUID, ...
    PRIMARY KEY (region, id)
  ) PARTITION BY LIST (region) (
    PARTITION us VALUES IN ('us'),
    PARTITION eu VALUES IN ('eu')
  );
  ALTER PARTITION us OF TABLE orders CONFIGURE ZONE USING
    constraints = '[+region=us-east]';
```

---

## 4. TiDB — MySQL-Compatible Distributed SQL

```
TiDB (PingCAP): MySQL wire-protocol compatible distributed SQL.

Architecture (separate compute and storage):

  ┌─────────────────────────────────────┐
  │          TiDB Servers (SQL)          │  ← Stateless SQL layer
  │  Parse → Optimize → Execute          │     (scale horizontally)
  └─────────────┬───────────────────────┘
                │
  ┌─────────────▼───────────────────────┐
  │         PD (Placement Driver)        │  ← Cluster coordinator
  │  - Timestamp oracle (TSO)            │     (Raft-based, 3 nodes)
  │  - Region routing table              │
  │  - Scheduling (load balance, split)  │
  └─────────────┬───────────────────────┘
                │
  ┌─────────────▼───────────────────────┐
  │       TiKV (Distributed KV)          │  ← Storage layer
  │  ┌────────┐ ┌────────┐ ┌────────┐  │     (Raft per region)
  │  │Region 1│ │Region 2│ │Region 3│  │
  │  │(Raft)  │ │(Raft)  │ │(Raft)  │  │
  │  └────────┘ └────────┘ └────────┘  │
  │  Powered by RocksDB per node        │
  └─────────────────────────────────────┘
  
  Optional: TiFlash (columnar store for analytics — HTAP)
  
  TiFlash: columnar replica of TiKV data
  → Analytical queries use TiFlash (column scans)
  → OLTP queries use TiKV (row lookups)
  → Real-time HTAP: one system for both!

Key features:
  - MySQL 5.7 wire protocol (most MySQL ORMs work unchanged)
  - Percolator-based distributed transactions (from Google)
  - Online DDL (add column, add index without downtime)
  - TSO for global timestamps (no TrueTime, but centralized timestamp oracle)

Trade-off: TSO is a single point of latency.
  Every transaction start and commit contacts PD for a timestamp.
  Mitigation: TSO batching, PD is Raft-replicated for availability.
```

---

## 5. YugabyteDB — PostgreSQL-Compatible

```
YugabyteDB: PostgreSQL wire-protocol compatible distributed SQL.

Architecture:
  YQL (Query Layer): PostgreSQL + CQL (Cassandra) compatible
  DocDB (Storage Layer): document-oriented, Raft-replicated tablets
  
  DocDB: forked from RocksDB, stores data as documents per tablet
  Each tablet = Raft group (3 replicas)
  Tablets auto-split and auto-rebalance

PostgreSQL compatibility:
  Reuses PostgreSQL's actual query layer (forked PG code)
  → Best PostgreSQL compatibility among distributed SQL databases
  → Supports PL/pgSQL, extensions (some), pg_dump/pg_restore

Consistency: cluster-level Raft
  Hybrid time (similar to CockroachDB's HLC)
  Serializable isolation (optional, default = snapshot)
  
Geo-distribution:
  Tablespace-based placement:
    CREATE TABLESPACE us_east WITH (replica_placement = 
      '{"num_replicas": 3, "placement_blocks": 
      [{"cloud":"aws","region":"us-east-1","zone":"us-east-1a","min_num_replicas":1}]}');
    CREATE TABLE users (...) TABLESPACE us_east;
```

---

## 6. Other Notable Distributed SQL

### Vitess (MySQL Sharding Middleware)

```
Vitess: sharding layer on top of MySQL (originated at YouTube).
PlanetScale: hosted Vitess platform.

Application → vtgate (proxy) → vttablet (per MySQL instance) → MySQL

Key features:
  - Transparent sharding for MySQL applications
  - Connection pooling
  - Query rewriting (scatter-gather for cross-shard queries)
  - Online schema changes (no locks)
  - NOT full distributed transactions (limited cross-shard)

Use when: you have a large MySQL deployment and need to shard
  without rewriting your application.
```

### Citus (Distributed PostgreSQL)

```
Citus: extension that turns PostgreSQL into a distributed database.
Now part of Azure Cosmos DB for PostgreSQL.

Coordinator node → worker nodes (each running PostgreSQL)

-- Distribute a table by a column:
SELECT create_distributed_table('orders', 'customer_id');

-- Reference tables (replicated to all workers):
SELECT create_reference_table('countries');

-- Queries automatically parallelized across workers
-- Co-located joins: if tables share the same distribution key → local join

Best for: multi-tenant SaaS (tenant_id as distribution key),
  real-time analytics on PostgreSQL.
```

### Neon (Serverless PostgreSQL)

```
Neon: separates PostgreSQL compute from storage.

Compute (PostgreSQL)  →  Pageserver (storage)  →  S3 (durable)
                          ↑
                     WAL → Safekeeper (3-node Paxos for WAL)

Key innovation:
  - Compute scales to zero (serverless — pay for what you use)
  - Instant branching: copy-on-write database branches (like git)
  - Storage is tiered: hot pages in pageserver, cold in S3
  - Point-in-time recovery to any LSN (WAL stored in S3)

Not a distributed SQL database (single-writer), but represents
the "serverless + separation of storage/compute" trend.
```

---

## 7. Comparison Table

```
Database      Wire Protocol  Consistency      Storage         Consensus  License
─────────────────────────────────────────────────────────────────────────────────
Spanner       GoogleSQL      External (linear) Colossus        Paxos      Proprietary
CockroachDB   PostgreSQL     Serializable      Pebble (LSM)    Raft       BSL/Apache2
TiDB          MySQL          Snapshot/Serial   TiKV (RocksDB)  Raft       Apache 2.0
YugabyteDB    PostgreSQL+CQL Snapshot/Serial   DocDB (RocksDB) Raft       Apache 2.0
Vitess        MySQL          Per-shard only    MySQL            N/A        Apache 2.0
Citus         PostgreSQL     Per-shard or 2PC  PostgreSQL       N/A        AGPL/Prop.
Neon          PostgreSQL     Serializable      Custom (S3)      Paxos(WAL) Apache 2.0

When to use which:
  - PostgreSQL app needing scale → CockroachDB or YugabyteDB
  - MySQL app needing scale → TiDB or Vitess
  - Global low-latency + 5-nines → Spanner (if you're on GCP)
  - Multi-tenant SaaS on PostgreSQL → Citus
  - Serverless PostgreSQL → Neon
```

---

## Key Takeaways

1. **Distributed SQL = SQL + ACID + horizontal scale.** The trade-off is higher per-query latency (network hops for consensus).
2. **Spanner's TrueTime is the gold standard** for global consistency, but requires GPS/atomic clocks. CockroachDB and YugabyteDB approximate it with HLC.
3. **CockroachDB's ranges + Raft** is the most widely adopted open-source distributed SQL architecture. Each range is an independent Raft group.
4. **TiDB separates compute (TiDB) from storage (TiKV)** — and adds TiFlash for real-time HTAP (analytics on the same data as OLTP).
5. **The distribution key is everything.** Choose it to minimize cross-shard queries. Co-location (same shard key) enables local joins.
6. **Vitess/Citus are sharding LAYERS** on existing databases. They don't replace the underlying engine, reducing migration risk.
7. **Neon represents the serverless trend** — separation of compute and storage, scale-to-zero, instant branching.

---

Next: [03-key-value-stores.md](03-key-value-stores.md) →
