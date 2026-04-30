# Module 3: Data Systems

> Your choice of data store is the most consequential decision in system design.

---

## 3.1 — SQL vs NoSQL: The Full Picture

### Relational Databases (SQL)

**Core properties: ACID**
- **Atomicity** — Transaction is all-or-nothing
- **Consistency** — DB always moves from one valid state to another
- **Isolation** — Concurrent transactions don't interfere
- **Durability** — Committed data survives crashes

**Strengths:**
- Complex queries (JOINs, aggregations, subqueries)
- Strong consistency guarantees
- Well-understood, battle-tested (PostgreSQL since 1996)
- Schema enforcement prevents garbage data
- Rich ecosystem of tools

**Weaknesses:**
- Horizontal scaling is hard (sharding is painful)
- Schema changes on large tables can be slow/risky
- Not ideal for hierarchical or graph-like data

**Major players:**
| Database | Sweet Spot |
|----------|-----------|
| **PostgreSQL** | Best all-around choice. JSON support, extensions, reliability |
| **MySQL** | Web apps, proven at scale (Facebook, Uber) |
| **CockroachDB** | Distributed SQL, survives region failures |
| **TiDB** | MySQL-compatible, horizontally scalable |
| **Spanner** | Google's globally distributed SQL (strongest consistency) |

### NoSQL Databases

**Why NoSQL exists:** Relational databases were designed for a single machine. When data exceeds one machine's capacity, or access patterns don't fit the relational model, NoSQL provides alternatives.

**BASE properties (contrast to ACID):**
- **B**asically **A**vailable
- **S**oft state
- **E**ventual consistency

### NoSQL Categories

#### 1. Key-Value Stores
```
key → value (opaque blob)
```
| Database | Throughput | Notes |
|----------|-----------|-------|
| **Redis** | ~100K ops/sec | In-memory, rich data structures |
| **DynamoDB** | Virtually unlimited | Managed, auto-scaling, single-digit ms |
| **etcd** | ~10K ops/sec | Distributed config, CP system |

**Use when:** Caching, session storage, simple lookups, rate limiting, leaderboards.
**Avoid when:** You need complex queries, relationships, or partial updates.

#### 2. Document Stores
```
key → JSON/BSON document (with nested fields)
```
| Database | Notes |
|----------|-------|
| **MongoDB** | Most popular, flexible schema, aggregation pipeline |
| **Couchbase** | Memcached-compatible caching + document store |
| **Firestore** | Real-time sync, mobile-first |

**Use when:** Varying schemas, rapid prototyping, content management, catalogs.
**Avoid when:** You need multi-document transactions (though MongoDB supports them now), complex JOINs.

#### 3. Wide-Column Stores
```
row_key → { column_family: { column: value, ... }, ... }
```
| Database | Notes |
|----------|-------|
| **Cassandra** | AP system, linear scalability, write-optimized |
| **HBase** | CP system, built on HDFS, Hadoop ecosystem |
| **ScyllaDB** | Cassandra-compatible, C++ (10x faster) |

**Use when:** Time-series data, IoT, high write throughput, need to scale writes to billions/day.
**Avoid when:** Complex queries, need strong consistency across partitions.

#### 4. Graph Databases
```
(Node)-[RELATIONSHIP]->(Node)
```
| Database | Notes |
|----------|-------|
| **Neo4j** | Most popular, Cypher query language |
| **Amazon Neptune** | Managed, supports Gremlin + SPARQL |
| **Dgraph** | Distributed, GraphQL native |

**Use when:** Social networks, recommendation engines, fraud detection, knowledge graphs.
**Avoid when:** Simple CRUD with no relationship queries.

#### 5. Search Engines
```
Documents → Inverted Index → Full-text search
```
| Database | Notes |
|----------|-------|
| **Elasticsearch** | Most popular, distributed, near real-time |
| **OpenSearch** | Fork of Elasticsearch (AWS-backed) |
| **Meilisearch** | Lightweight, fast, typo-tolerant |
| **Typesense** | Easy to use, low latency |

**Use when:** Full-text search, log analysis, analytics, autocomplete.
**Not a primary database** — use as a secondary index synchronized from your source of truth.

### The Decision Framework

```
Need transactions + complex queries?          → SQL (PostgreSQL)
Need simple key lookups at massive scale?      → Key-Value (DynamoDB/Redis)
Flexible schema + document-oriented?           → Document (MongoDB)
Massive write throughput + time-series?         → Wide-Column (Cassandra)
Relationship/graph traversals?                 → Graph (Neo4j)
Full-text search?                              → Search (Elasticsearch)
Need SQL + horizontal scaling?                 → NewSQL (CockroachDB/Spanner)
```

**Pro tip:** Most real systems use **polyglot persistence** — multiple databases for different use cases. E.g., PostgreSQL for transactions + Redis for cache + Elasticsearch for search.

---

## 3.2 — Database Indexing

### B-Tree Index (Default in most SQL DBs)

```
             [50]
           /      \
      [20, 30]     [70, 80]
      / |  \       / |  \
   [10] [25] [35] [60] [75] [90]
```

- Balanced tree, O(log n) lookups
- Great for range queries (WHERE age > 25 AND age < 50)
- Default index type in PostgreSQL, MySQL

### Hash Index

- O(1) exact lookups
- Cannot do range queries
- Used by: memcached, Redis, hash joins in DBs

### LSM Tree + SSTable (Used by NoSQL)

```
Write → Memtable (in-memory sorted tree)
         ↓ (when full, flush)
      SSTable on disk (sorted, immutable)
         ↓ (background compaction)
      Merged SSTables
```

- **Write-optimized** (all writes go to memory first)
- Reads may need to check multiple SSTables
- Used by: LevelDB, RocksDB, Cassandra, HBase

### B-Tree vs LSM Tree

| Feature | B-Tree | LSM Tree |
|---------|--------|----------|
| Read speed | Faster (single seek) | Slower (multiple SSTables) |
| Write speed | Slower (random I/O) | Faster (sequential I/O) |
| Space amplification | Higher (page fragmentation) | Lower (compaction) |
| Write amplification | Lower | Higher (compaction rewrites) |
| Best for | Read-heavy, OLTP | Write-heavy, logs, time-series |

### Indexing Strategies

```sql
-- Single Column Index
CREATE INDEX idx_email ON users(email);

-- Composite Index (order matters!)
CREATE INDEX idx_name_age ON users(last_name, first_name, age);
-- This index works for:
--   WHERE last_name = 'Smith'                          ✓
--   WHERE last_name = 'Smith' AND first_name = 'John'  ✓
--   WHERE first_name = 'John'                           ✗ (leftmost prefix rule)

-- Covering Index (includes all needed columns)
CREATE INDEX idx_cover ON orders(user_id) INCLUDE (total, status);
-- Query can be answered entirely from index, no table lookup

-- Partial Index
CREATE INDEX idx_active ON users(email) WHERE active = true;
-- Smaller index, only covers subset of rows
```

### When NOT to Index

- Columns rarely used in WHERE/JOIN/ORDER BY
- Tables with very few rows
- Columns with low cardinality (e.g., boolean with 50/50 split)
- Write-heavy tables where index maintenance is expensive

---

## 3.3 — Database Sharding (Partitioning)

When one database isn't enough.

### Horizontal Sharding (Most common)

Split **rows** across multiple databases.

```
Users 1-1M       → Shard 1
Users 1M-2M      → Shard 2
Users 2M-3M      → Shard 3
```

### Sharding Strategies

#### 1. Range-Based Sharding
```
user_id 1-1M      → Shard A
user_id 1M-2M     → Shard B
user_id 2M-3M     → Shard C
```
- **Pros:** Range queries are efficient, simple
- **Cons:** Hot spots (if new user_ids cluster on one shard)

#### 2. Hash-Based Sharding
```
shard = hash(user_id) % num_shards
```
- **Pros:** Even distribution
- **Cons:** Range queries require hitting all shards, resharding is painful
- **Better:** Use consistent hashing to minimize data movement

#### 3. Directory-Based Sharding
```
Lookup table: user_id → shard_id
```
- **Pros:** Flexible, can rebalance by updating directory
- **Cons:** Directory is a single point of failure, additional lookup latency

#### 4. Geographic Sharding
```
US users     → US-East shard
EU users     → EU-West shard
Asia users   → AP-South shard
```
- **Pros:** Data locality, compliance (GDPR)
- **Cons:** Cross-region queries are expensive

### Choosing a Shard Key

This is the **single most important decision** in sharding.

**Good shard key properties:**
- High cardinality (many distinct values)
- Even distribution (no hot spots)
- Query isolation (most queries hit one shard)
- Immutable (doesn't change)

**Examples:**
| Use Case | Good Shard Key | Bad Shard Key |
|----------|---------------|---------------|
| Social media | user_id | created_at (recent dates = hot) |
| E-commerce | order_id or user_id | product_id (popular items = hot) |
| Multi-tenant SaaS | tenant_id | user_type (few values) |
| Chat | conversation_id | timestamp |

### Problems with Sharding

1. **JOINs across shards** — Extremely expensive. Denormalize or use application-level joins.
2. **Transactions across shards** — Require distributed transactions (2PC). Avoid if possible.
3. **Resharding** — Adding/removing shards requires data migration. Plan for 2-4x growth.
4. **Hot shards** — One shard gets disproportionate traffic. Must re-shard or split.
5. **Referential integrity** — Foreign keys don't work across shards.

### Vitess: Sharding without the Pain

Used by YouTube, Slack, Square. Sits between app and MySQL:
- Automatic query routing
- Resharding with zero downtime
- Connection pooling
- Schema changes across shards

---

## 3.4 — Replication

### Single-Leader Replication

```
  Writes → [Leader]
              ↓ (replication log)
         [Follower 1] ← Reads
         [Follower 2] ← Reads
         [Follower 3] ← Reads
```

- All writes go to leader
- Leader replicates to followers
- Reads can go to any replica
- **Problem:** Replication lag → stale reads

#### Replication Lag Solutions

| Problem | Solution |
|---------|----------|
| User writes, then reads stale data | **Read-your-writes consistency:** After write, read from leader for N seconds |
| Monotonic reads violated (user sees newer then older data) | **Session consistency:** Same user always reads from same replica |
| Causal ordering violated | **Causal consistency:** Track dependencies between operations |

### Multi-Leader Replication

```
[Leader US] ←→ [Leader EU] ←→ [Leader Asia]
     ↓              ↓              ↓
[Followers]    [Followers]    [Followers]
```

- Multiple leaders accept writes
- Leaders replicate to each other asynchronously
- **Problem:** Write conflicts (same record updated in two regions)

#### Conflict Resolution Strategies

| Strategy | How | Fairness |
|----------|-----|----------|
| **Last Write Wins (LWW)** | Highest timestamp wins | Loses data silently |
| **Merge values** | Combine both writes | Application-specific |
| **Custom logic** | Application resolves | Most flexible |
| **CRDTs** | Conflict-free data structures | Mathematically guaranteed |

### Leaderless Replication

```
Client writes to all N replicas
Client reads from R replicas
Quorum: W + R > N guarantees overlap

Example (N=3, W=2, R=2):
  Write succeeds when 2/3 replicas confirm
  Read returns value from 2/3 replicas (take latest)
  Since 2+2 > 3, at least one read will hit an up-to-date replica
```

- Used by Cassandra, DynamoDB, Riak
- No single point of failure
- Tunable consistency (adjust W and R)

| Configuration | W | R | Property |
|--------------|---|---|----------|
| Strong consistency | N | 1 | All write, one read |
| Strong consistency | ⌈N/2⌉+1 | ⌈N/2⌉+1 | Quorum |
| High availability writes | 1 | N | One write, all read |
| High write throughput | 1 | 1 | Fastest, weakest consistency |

---

## 3.5 — Data Modeling Patterns

### Denormalization

**Normalization:** No duplicate data. JOINs everywhere. Great for consistency, terrible for read performance at scale.

**Denormalization:** Duplicate data to avoid JOINs. Trade write complexity for read speed.

```
Normalized (3 queries):
  users: {id, name}
  posts: {id, user_id, content}
  comments: {id, post_id, user_id, text}

Denormalized (1 query):
  posts: {
    id, content,
    author_name: "Vikram",         ← duplicated from users
    recent_comments: [              ← embedded from comments
      {user_name: "Alice", text: "Great post!"}
    ]
  }
```

### Event Sourcing

Instead of storing current state, store **all events** that led to the current state.

```
Traditional:
  account: { balance: 500 }

Event Sourced:
  events: [
    { type: "deposit",  amount: 1000, at: "2024-01-01" }
    { type: "withdraw", amount: 300,  at: "2024-01-15" }
    { type: "withdraw", amount: 200,  at: "2024-02-01" }
  ]
  // Current balance = replay events = 1000 - 300 - 200 = 500
```

**Advantages:**
- Complete audit trail
- Can reconstruct state at any point in time
- Events can be replayed to build new views/projections
- Natural fit for distributed systems

**Used by:** Banking, e-commerce, event-driven microservices, Git (is essentially event sourcing for code).

### CQRS (Command Query Responsibility Segregation)

Separate the **write model** from the **read model**.

```
Commands (writes):
  [API] → [Command Handler] → [Write DB (normalized, event store)]
                                      ↓ (events)
Queries (reads):
  [API] → [Query Handler] → [Read DB (denormalized, optimized views)]
```

**Why:** Writes and reads have fundamentally different requirements:
- Writes need consistency, validation, business rules
- Reads need speed, flexible shapes, aggregations

**When to use:** High read:write ratio, complex domain models, event sourcing systems.

---

## 3.6 — Time-Series Data

For metrics, IoT, monitoring, financial data.

### Specialized Databases

| Database | Notes |
|----------|-------|
| **InfluxDB** | Purpose-built TSDB, SQL-like query language |
| **TimescaleDB** | PostgreSQL extension (get TSDB features + full SQL) |
| **Prometheus** | Pull-based monitoring, built-in alerting |
| **ClickHouse** | Column-oriented OLAP, extremely fast aggregations |
| **QuestDB** | High-performance, SQL-compatible |

### Key Patterns

1. **Downsampling:** Keep raw data for 7 days, 1-minute averages for 30 days, 1-hour averages for 1 year
2. **Partitioning by time:** Each day/week/month is a partition. Drop old partitions = instant cleanup
3. **Compression:** Time-series data compresses extremely well (delta encoding, gorilla compression)

---

## 3.7 — Blob/Object Storage

For images, videos, files, backups.

| Service | Notes |
|---------|-------|
| **S3** (AWS) | Industry standard, 11 nines durability, virtually unlimited |
| **GCS** (Google) | Similar to S3, strong consistency |
| **Azure Blob** | Microsoft's equivalent |
| **MinIO** | Self-hosted S3-compatible |

### Key Concepts

- **Buckets** — Top-level containers
- **Objects** — Files with metadata
- **Pre-signed URLs** — Temporary access without exposing credentials
- **Lifecycle policies** — Auto-move to cheaper storage after N days
- **Multipart upload** — Upload large files in parallel chunks

### Storage Tiers (S3 example)

| Tier | Cost | Access | Use Case |
|------|------|--------|----------|
| Standard | $$$  | Instant | Frequently accessed data |
| Infrequent Access | $$ | Instant (+ retrieval fee) | Accessed < 1x/month |
| Glacier | $ | Minutes to hours | Archives, compliance |
| Glacier Deep Archive | ¢ | 12-48 hours | Long-term legal/regulatory |

---

## 3.8 — Data Warehousing & Analytics

### OLTP vs OLAP

| Property | OLTP | OLAP |
|----------|------|------|
| Purpose | Day-to-day operations | Analytics & reporting |
| Queries | Simple, frequent | Complex, infrequent |
| Data shape | Rows (normalized) | Columns (denormalized, star schema) |
| Volume | GBs to TBs | TBs to PBs |
| Example | PostgreSQL, MySQL | BigQuery, Redshift, Snowflake |

### Star Schema

```
              [dim_time]
                  |
[dim_product] -- [fact_sales] -- [dim_customer]
                  |
              [dim_store]
```

- **Fact table:** The events (sales, clicks, page views). Huge.
- **Dimension tables:** The context (who, what, when, where). Smaller.

### Modern Data Stack

```
Sources → ETL/ELT → Data Warehouse → BI/Analytics

Sources: PostgreSQL, APIs, logs, S3
ETL: Fivetran, Airbyte, dbt
Warehouse: BigQuery, Snowflake, Redshift, ClickHouse
BI: Metabase, Looker, Tableau, Superset
```

---

## 3.9 — Exercises

1. **Database selection:** You're building Uber. What database(s) would you use for: (a) user profiles, (b) ride history, (c) real-time driver locations, (d) analytics on ride patterns? Justify each choice.

2. **Sharding design:** Your e-commerce platform has 100M users and 1B orders. Design a sharding strategy. What's your shard key? How do you handle the "get all orders for user X" query? How about "get all orders for product Y"?

3. **Replication trade-off:** Your system uses single-leader replication with 3 replicas. Replication lag is 200ms. A user posts a comment and immediately refreshes the page. How do you ensure they see their own comment?

4. **Event sourcing:** Design an event-sourced system for a collaborative document editor (like Google Docs). What events do you store? How do you handle concurrent edits?

---

**Next:** [Module 4 — Distributed Systems](04-distributed-systems.md) →
