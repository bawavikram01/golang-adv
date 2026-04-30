# The 10 Pillars of Any Database

> When you encounter **any** database — new or old, SQL or NoSQL, distributed or embedded —  
> ask these 10 questions. The answers tell you everything.

---

## The Framework

```
 1. DATA MODEL            — How does it organize data?
 2. STORAGE ENGINE        — How does it store data on disk?
 3. INDEXING              — How does it find data fast?
 4. QUERY PROCESSING      — How does it execute queries?
 5. TRANSACTIONS          — How does it handle concurrent access?
 6. DURABILITY & RECOVERY — How does it survive crashes?
 7. REPLICATION           — How does it copy data across nodes?
 8. PARTITIONING/SHARDING — How does it scale beyond one machine?
 9. MEMORY MANAGEMENT     — How does it use RAM?
10. OPERATIONAL SURFACE   — How do you run it in production?
```

---

## 1. Data Model

**Question:** How does this database organize and represent data?

```
Relational (rows/tables):  PostgreSQL, MySQL, SQLite, Oracle, SQL Server
Document (JSON/BSON):      MongoDB, CouchDB, Couchbase, FerretDB
Key-Value:                 Redis, DynamoDB, etcd, Memcached
Wide-Column:               Cassandra, ScyllaDB, HBase, Bigtable
Graph:                     Neo4j, Neptune, JanusGraph, ArangoDB
Time-Series:               TimescaleDB, InfluxDB, QuestDB, Prometheus
Columnar (OLAP):           ClickHouse, DuckDB, BigQuery, Redshift
Vector:                    pgvector, Pinecone, Milvus, Qdrant
```

**What to learn:** Schema design patterns, data access patterns, trade-offs of each model.

---

## 2. Storage Engine

**Question:** How does this database persist data to disk?

```
B+ Tree:       PostgreSQL, MySQL/InnoDB, SQLite, Oracle, SQL Server
               → Balanced tree, O(log N) reads, good for reads + writes
               → Data sorted by key, great for range scans

LSM-Tree:      RocksDB, Cassandra, HBase, LevelDB, Pebble, TiKV
               → Log-structured, sequential writes only
               → Great for write-heavy workloads (10-100x write throughput)
               → Compaction needed (size-tiered, leveled, FIFO)

In-Memory:     Redis, VoltDB, Memcached, SAP HANA
               → Everything in RAM, optional persistence to disk
               → Microsecond latency

Columnar:      ClickHouse, DuckDB, Parquet, BigQuery
               → Data stored by column, not row
               → 10-100x compression, scans only needed columns

Hybrid:        CockroachDB (LSM via Pebble), TiDB (LSM via TiKV)

Key concepts:
  - Page layout (slotted pages, fixed-size)
  - WAL (Write-Ahead Log) — always understand how writes hit disk
  - Compaction (LSM) or VACUUM (PostgreSQL) — garbage collection
  - Bloom filters (LSM optimization — skip files that don't have your key)
  - Copy-on-write (LMDB, BoltDB)
```

---

## 3. Indexing

**Question:** How does this database find data without scanning everything?

```
B+ Tree index:      The default everywhere (PostgreSQL, MySQL, SQLite)
Hash index:         O(1) exact lookup, no range scans (Redis, Memcached)
Inverted index:     Full-text search (Elasticsearch, Lucene, PostgreSQL GIN)
R-tree / GiST:     Spatial data (PostGIS, MongoDB geospatial)
BRIN:               Block range index — huge tables, sorted data (PostgreSQL)
Bloom filter:       Probabilistic — "definitely not here" (LSM optimization)
HNSW / IVF:         Vector similarity search (pgvector, Pinecone, Milvus)
Skip list:          Ordered in-memory (Redis sorted sets)

Key concepts:
  - Primary vs secondary index
  - Clustered vs non-clustered (InnoDB: clustered by PK; PostgreSQL: heap)
  - Covering index (index-only scans — no table lookup)
  - Composite index column ordering (leftmost prefix rule)
  - Partial / conditional indexes
  - Index maintenance cost (every write updates every index)
  - When NOT to index (low selectivity, write-heavy, small tables)
```

---

## 4. Query Processing

**Question:** How does a query go from SQL text to actual results?

```
Pipeline:
  SQL text → Lexer/Parser → AST → Semantic Analysis
  → Logical Plan → Optimizer → Physical Plan → Executor → Results

Optimizer types:
  Cost-based:  PostgreSQL, MySQL 8+, CockroachDB, Oracle
               Estimates cost of each plan, picks cheapest
  Rule-based:  Some NoSQL systems, older databases
               Apply fixed transformation rules

Execution models:
  Volcano (iterator):  Pull one tuple at a time (PostgreSQL, MySQL)
  Vectorized:          Process batches of ~2048 values (DuckDB, ClickHouse)
  Push-based:          Operators push data to parent (some modern engines)
  Compiled/JIT:        Compile query to native code (PostgreSQL JIT, HyPer)

Key concepts:
  - EXPLAIN / EXPLAIN ANALYZE (read this for EVERY slow query)
  - Join algorithms: nested loop, hash join, sort-merge join
  - Predicate pushdown, projection pushdown
  - Cardinality estimation (histograms, n-distinct, MCV)
  - Statistics (ANALYZE in PostgreSQL, ANALYZE TABLE in MySQL)
  - Parallel query execution
```

---

## 5. Transactions & Concurrency

**Question:** What happens when multiple connections read/write simultaneously?

```
ACID:
  Atomicity:   All or nothing (commit/rollback)
  Consistency: Constraints always hold
  Isolation:   Transactions don't interfere with each other
  Durability:  Committed data survives crashes

Isolation levels:
  READ UNCOMMITTED → dirty reads possible (almost nobody uses this)
  READ COMMITTED   → default in PostgreSQL, Oracle
  REPEATABLE READ  → default in MySQL/InnoDB
  SERIALIZABLE     → as-if-sequential execution

Concurrency control mechanisms:
  2PL (Two-Phase Locking):  Acquire locks → release locks. Simple but blocking.
  MVCC:                     Multiple versions of each row. Readers never block writers.
    PostgreSQL: xmin/xmax on tuples, VACUUM cleans old versions
    MySQL:      Undo logs + read views, purge thread cleans
  OCC:                      No locks during execution, validate at commit.

Anomalies to understand:
  - Dirty read, non-repeatable read, phantom read
  - Write skew (the subtle one — only prevented by SERIALIZABLE)
  - Lost update

Some databases have LIMITED transactions:
  - Redis: single-threaded, atomic commands, MULTI/EXEC
  - Cassandra: lightweight transactions (Paxos) for single partition
  - DynamoDB: single-item atomic, or TransactWriteItems (25 items max)
  - MongoDB: multi-document transactions (since 4.0, replica set required)
```

---

## 6. Durability & Recovery

**Question:** If the power goes out mid-write, what happens?

```
WAL (Write-Ahead Log):
  Rule: BEFORE modifying any data page, write the change to the log.
  On crash: replay the log to recover.
  Used by: PostgreSQL, MySQL, SQLite, CockroachDB, virtually everyone.

Recovery algorithm:
  ARIES (Analysis → Redo → Undo):
    PostgreSQL, MySQL, SQL Server, DB2, Oracle
    1. Analysis: figure out what was active at crash
    2. Redo: replay all changes (restore pre-crash state)
    3. Undo: roll back uncommitted transactions

Checkpointing:
  Periodically flush dirty pages to disk + record position in WAL.
  Reduces recovery time (don't replay from beginning of log).

Other durability mechanisms:
  - Double-write buffer (InnoDB — prevents torn pages)
  - Group commit (batch multiple transactions' flushes together)
  - fsync (the system call that actually guarantees data is on disk)
  - RDB/AOF (Redis — periodic snapshots + append-only file)

PITR (Point-in-Time Recovery):
  Restore a backup + replay WAL up to a specific timestamp.
  Supported by: PostgreSQL (WAL archiving), MySQL (binlog).
```

---

## 7. Replication

**Question:** How does this database copy data to other nodes?

```
Replication topologies:
  Single-leader:    One writer, N readers (PostgreSQL, MySQL, MongoDB)
  Multi-leader:     Multiple writers (Galera, CockroachDB, Cassandra)
  Leaderless:       Any node accepts writes (Cassandra, DynamoDB, Riak)

Sync vs async:
  Synchronous:   Leader waits for replica confirmation → no data loss, higher latency
  Asynchronous:  Leader doesn't wait → possible data loss on failover, lower latency
  Semi-sync:     Wait for at least 1 replica (MySQL semi-sync)

Replication methods:
  Physical/streaming: Ship WAL bytes (PostgreSQL streaming replication)
  Logical:            Ship row-level changes (PostgreSQL logical replication, MySQL binlog)
  State machine:      Ship commands via consensus (Raft in CockroachDB, etcd)

Key concepts:
  - Replication lag (how far behind are replicas?)
  - Failover (automatic vs manual, split-brain prevention)
  - Read-your-writes consistency (read from leader after write)
  - Consensus protocols: Raft (CockroachDB, etcd), Paxos (Spanner)
```

---

## 8. Partitioning / Sharding

**Question:** How does this database split data across multiple machines?

```
Partitioning (single node — split table into pieces):
  Range:  partition by date range, ID range
  List:   partition by category, region
  Hash:   partition by hash(key) mod N

Sharding (multi-node — distribute across machines):
  Hash-based:       hash(shard_key) → shard (DynamoDB, Cassandra, Redis Cluster)
  Range-based:      key ranges assigned to shards (CockroachDB, HBase, Spanner)
  Directory-based:  lookup table maps key → shard (application-level)

Key concepts:
  - Shard key selection (high cardinality, even distribution, query locality)
  - Hot shards (one shard gets disproportionate traffic)
  - Cross-shard queries (joins across shards = hard + slow)
  - Cross-shard transactions (2PC or saga pattern)
  - Resharding / rebalancing (adding nodes without downtime)
  - Some databases auto-shard: CockroachDB, Cassandra, DynamoDB
  - Some need manual sharding: MySQL (via Vitess), PostgreSQL (via Citus)
```

---

## 9. Memory Management

**Question:** How does this database use RAM?

```
Buffer pool / page cache:
  Most databases cache frequently-accessed pages in memory.
  PostgreSQL: shared_buffers (typically 25% of RAM) + OS page cache
  MySQL:      innodb_buffer_pool_size (typically 70-80% of RAM)

Eviction policies:
  LRU:          Least recently used (simple, vulnerable to sequential scan)
  Clock sweep:  Approximation of LRU (PostgreSQL)
  LRU-K:        Track K-th most recent access (avoids scan pollution)

Other memory areas:
  - Work memory:      Sorts, hash joins, hash aggregates (work_mem in PG)
  - Connection memory: Each connection uses memory (why pooling matters)
  - WAL buffers:       Batch WAL writes before flushing
  - Query plan cache:  Cache compiled query plans (MySQL, SQL Server)

In-memory databases:
  Redis:     Everything in RAM, optional disk persistence
  VoltDB:    All data in memory, WAL for durability
  SAP HANA:  In-memory columnar + row store
```

---

## 10. Operational Surface

**Question:** How do you actually run this database in production?

```
Monitoring:
  - Key metrics: QPS, latency (p50/p95/p99), connections, cache hit ratio,
    replication lag, disk I/O, lock waits, dead tuples, bloat
  - Tools: Prometheus + Grafana, Datadog, PMM (Percona)

Backup & restore:
  - Logical: pg_dump, mysqldump (slow but portable)
  - Physical: pg_basebackup, XtraBackup (fast, for PITR)
  - PITR: restore to any point in time via WAL/binlog replay

Security:
  - Authentication (password, certificate, LDAP, IAM)
  - Authorization (RBAC, row-level security)
  - Encryption (TLS in transit, TDE or LUKS at rest)
  - SQL injection prevention (parameterized queries, always)

Schema migrations:
  - Tools: golang-migrate, Atlas, Flyway, Alembic
  - Zero-downtime patterns: expand-and-contract, CONCURRENTLY
  - Online DDL: gh-ost, pt-online-schema-change, pg_repack

High availability:
  - Failover mechanism (Patroni, Orchestrator, built-in)
  - Connection routing (HAProxy, PgBouncer, ProxySQL)
  - RTO/RPO targets

Configuration tuning:
  - Every database has 5-10 critical knobs
  - PostgreSQL: shared_buffers, work_mem, effective_cache_size, max_connections
  - MySQL: innodb_buffer_pool_size, innodb_log_file_size, max_connections
```

---

## Quick Reference: Apply to Any Database

When learning a new database, fill in this template:

```
Database: ________________

1. Data Model:       [ relational | document | KV | wide-column | graph | ... ]
2. Storage Engine:   [ B+ tree | LSM-tree | in-memory | columnar | ... ]
3. Indexing:         [ B+ tree | hash | inverted | HNSW | ... ]
4. Query Processing: [ cost-based SQL | CQL | command-based | ... ]
5. Transactions:     [ full ACID | single-row | eventual | none ]
6. Durability:       [ WAL + ARIES | AOF | commit log | ... ]
7. Replication:      [ single-leader | multi-leader | leaderless | Raft | ... ]
8. Sharding:         [ auto-range | auto-hash | manual | none ]
9. Memory:           [ buffer pool | all-in-memory | OS page cache | ... ]
10. Operations:      [ monitoring tools | backup method | HA mechanism | ... ]
```

---

> **This is the complete map.**  
> Every file in `phase1/` through `phase6/` teaches details within these 10 pillars.  
> When confused, come back here. Ask: "Which pillar am I learning about?"  
> The answer will always be one of these 10.
