# 4.3 — Key-Value Stores

> The simplest data model: key → value.  
> This simplicity enables extreme performance and scalability.  
> Key-value stores are the backbone of caching, session management,  
> configuration, and distributed coordination.

---

## 1. Redis — The Swiss Army Knife

### Architecture

```
Redis is single-threaded (mostly):
  One main thread handles ALL commands sequentially.
  
  Why single-threaded?
    - No locking overhead (no mutexes, no deadlocks)
    - No context switching between threads
    - CPU is rarely the bottleneck (network and memory are)
    - Simple, correct code
  
  Redis 6.0+: I/O threads for network read/write (parsing/writing)
    Command execution is still single-threaded.
  Redis 7.0+: more I/O threading improvements.
  
  Event loop: epoll (Linux) / kqueue (macOS)
    → Non-blocking I/O, handles 100K+ ops/sec on single core

Memory layout:
  Uses custom allocators (jemalloc by default)
  Every key-value pair is a redisObject:
    type (string, list, set, hash, zset, stream)
    encoding (int, embstr, raw, ziplist, listpack, skiplist, hashtable...)
    ptr → actual data
    refcount, LRU bits
```

### Data Structures (This Is What Makes Redis Special)

```
Redis is NOT just key → string. It's key → data structure.

STRING:
  SET key "value"              O(1)
  GET key                      O(1)
  INCR counter                 O(1) — atomic increment
  SETNX key "value"            O(1) — set if not exists (distributed lock primitive)
  SET key "value" EX 300       O(1) — set with 5-min expiration
  
  Encoding: int (for integers ≤ 2^63), embstr (≤ 44 bytes), raw

LIST (doubly-linked list → quicklist):
  LPUSH/RPUSH mylist "a" "b"   O(N) for N elements pushed
  LPOP/RPOP mylist             O(1)
  LRANGE mylist 0 -1           O(S+N) — get all elements
  BLPOP mylist 30              blocking pop (queue!) — wait up to 30s
  
  Encoding: listpack (small) → quicklist (large)
  Use case: message queues, activity feeds, recent items

SET:
  SADD myset "a" "b" "c"      O(N)
  SMEMBERS myset               O(N)
  SISMEMBER myset "a"          O(1) — membership check
  SINTER set1 set2             O(N*M) — intersection
  SUNION set1 set2             set union
  
  Encoding: listpack (small) → hashtable (large)
  Use case: tags, unique visitors, common friends

HASH (field-value pairs inside a key):
  HSET user:1 name "Alice" age "30"
  HGET user:1 name
  HGETALL user:1               O(N)
  HINCRBY user:1 age 1         atomic field increment
  
  Encoding: listpack (small) → hashtable (large)
  Use case: objects, user profiles, session data

SORTED SET (ZSET — sorted by score):
  ZADD leaderboard 100 "alice" 200 "bob" 150 "carol"
  ZRANGE leaderboard 0 -1 WITHSCORES     — ascending by score
  ZREVRANGE leaderboard 0 9              — top 10
  ZRANGEBYSCORE leaderboard 100 200       — score range
  ZRANK leaderboard "alice"               — rank (0-based)
  
  Encoding: listpack (small) → skiplist + hashtable (large)
  
  Skiplist: probabilistic data structure for O(log N) operations
    Level 3: ────────────────────── 50 ────────────── → NIL
    Level 2: ────── 20 ─────────── 50 ──── 80 ──── → NIL
    Level 1: 10 ── 20 ── 30 ── 50 ── 60 ── 80 ── → NIL
  
  Use case: leaderboards, priority queues, rate limiters, time-series

STREAM (append-only log, like Kafka):
  XADD mystream * field1 value1 field2 value2   — auto-generated ID
  XREAD COUNT 10 STREAMS mystream 0              — read from beginning
  XRANGE mystream - +                            — range query
  
  Consumer groups (like Kafka consumer groups):
    XGROUP CREATE mystream mygroup 0
    XREADGROUP GROUP mygroup consumer1 COUNT 1 STREAMS mystream >
    XACK mystream mygroup <message-id>
  
  Use case: event streaming, audit logs, activity feeds
```

### Persistence

```
RDB (snapshotting):
  Periodic point-in-time snapshots to disk (dump.rdb)
  save 900 1      — snapshot if ≥1 key changed in 900 seconds
  save 300 10     — snapshot if ≥10 keys changed in 300 seconds
  ✓ Compact, fast recovery
  ✗ Data loss between snapshots

AOF (Append Only File):
  Log every write command to a file
  appendfsync always    — fsync after every command (safe, slow)
  appendfsync everysec  — fsync once per second (good compromise)
  appendfsync no        — let OS decide (fast, less safe)
  
  AOF rewrite: background process compacts the AOF file
  ✓ Better durability
  ✗ Larger files, slower recovery

RDB + AOF (recommended):
  Use both. On restart, Redis loads AOF (most complete).

Redis 7.0+ Multi-Part AOF:
  AOF split into base (RDB-like) + incremental files
  → More efficient, no full rewrite needed
```

### Clustering

```
Redis Cluster:
  - 16384 hash slots divided among master nodes
  - hash_slot = CRC16(key) % 16384
  - Each master handles a subset of slots
  - Each master has 0+ replicas
  - Gossip protocol for cluster state
  
  ┌──────────┐  ┌──────────┐  ┌──────────┐
  │Master 1  │  │Master 2  │  │Master 3  │
  │Slots 0-  │  │Slots     │  │Slots     │
  │5460      │  │5461-10922│  │10923-16383│
  │  │Replica│  │  │Replica│  │  │Replica│
  └──────────┘  └──────────┘  └──────────┘

  Multi-key operations: ONLY work if all keys are on the SAME slot.
  Hash tags force co-location: {user:1}:name and {user:1}:email → same slot.

Redis Sentinel (no clustering, just HA):
  - Monitors master + replicas
  - Automatic failover if master dies
  - Client connects to Sentinel, which redirects to current master
  - Simpler than Cluster, but no sharding
```

---

## 2. Memcached

```
Memcached: pure in-memory cache. Simpler than Redis.

Architecture:
  Multi-threaded (scales to many cores — unlike Redis)
  No persistence (RAM only)
  No data structures (string values only)
  No replication, no clustering (client-side consistent hashing)
  
  Slab allocator:
    Memory divided into slabs of fixed chunk sizes (64B, 128B, 256B, ...)
    Item stored in smallest slab class that fits.
    → No fragmentation (unlike malloc/free)
    → But: memory waste if item < slab class size
    → And: slab calcification (popular class runs out, others empty)

  LRU eviction per slab class.

When to use Memcached over Redis:
  ✓ Simple caching (string key → string value)
  ✓ Multi-threaded: better multi-core utilization out of box
  ✓ Lower memory overhead per key (~50 bytes metadata vs Redis ~70+)
  ✗ No persistence, no pub/sub, no data structures, no scripting
```

---

## 3. etcd — Distributed Coordination

```
etcd: strongly consistent key-value store, built on Raft.
Used by Kubernetes as its primary data store.

Properties:
  - Linearizable reads and writes
  - Watch API (get notified of changes)
  - Lease-based TTL (keys expire when lease dies)
  - Multi-version (MVCC — every key has revision history)
  - Transactions (multi-key compare-and-swap)

Use cases:
  - Service discovery (register/lookup services)
  - Leader election (using lease + compare-and-swap)
  - Distributed locking (lease-based locks)
  - Configuration management (watch for changes)
  - Kubernetes: stores ALL cluster state (pods, services, secrets)

etcd is NOT for:
  ✗ Large datasets (recommended < 8 GB)
  ✗ High throughput (designed for consistency, not throughput)
  ✗ General-purpose key-value storage

  etcdctl put /myapp/config '{"db":"postgres://..."}'
  etcdctl get /myapp/config
  etcdctl watch /myapp/ --prefix    -- watch all keys under /myapp/
```

---

## 4. Amazon DynamoDB

```
DynamoDB: fully managed NoSQL key-value + document database.
Inspired by Amazon's 2007 Dynamo paper + SimpleDB.

Data model:
  Table → Items (rows) → Attributes (columns)
  
  Primary key:
    Partition key (hash key):  hash determines partition placement
    OR
    Partition key + Sort key:  composite key for range queries within a partition

  -- Single item operations:
  PutItem, GetItem, UpdateItem, DeleteItem

  -- Query (within a single partition):
  Query(PK = 'user#123', SK begins_with 'order#')
  → Efficient: single partition scan
  
  -- Scan (full table scan):
  Scan() → reads EVERY item → expensive, avoid in production!

Secondary Indexes:
  GSI (Global Secondary Index): different partition key + sort key
    → Copies data to a new partition scheme (eventual consistency)
    → Can project all attributes or a subset
    → Has its own throughput capacity
  
  LSI (Local Secondary Index): same partition key, different sort key
    → Must be created at table creation time
    → Strongly consistent reads possible
    → Shares throughput with base table

Single-Table Design (DynamoDB pattern):
  Instead of multiple tables (relational thinking):
  → Put EVERYTHING in one table with overloaded PK/SK
  
  PK               SK                  Data
  ──────────────── ─────────────────── ──────────
  USER#123         PROFILE             {name, email, ...}
  USER#123         ORDER#2024-001      {total, status, ...}
  USER#123         ORDER#2024-002      {total, status, ...}
  ORDER#2024-001   ITEM#A              {product, qty, ...}
  ORDER#2024-001   ITEM#B              {product, qty, ...}
  
  Query(PK='USER#123') → get user profile + all orders
  Query(PK='USER#123', SK begins_with 'ORDER#') → just orders

Capacity modes:
  Provisioned: specify read/write capacity units (cheaper, predictable)
  On-demand: auto-scales (more expensive, zero planning)

DynamoDB Streams:
  Change data capture — ordered stream of item changes
  → Trigger Lambda functions on changes
  → Replicate to other systems (Elasticsearch, analytics)

Global Tables:
  Multi-region, multi-active replication
  All replicas accept writes, conflict resolution via LWW
```

---

## 5. RocksDB — The Embedded Engine

```
RocksDB: embeddable LSM-tree key-value store (Facebook, forked from LevelDB).
NOT a standalone database — it's a storage engine used INSIDE other databases.

Used by:
  - TiKV (TiDB's storage)
  - CockroachDB (Pebble — RocksDB-inspired rewrite in Go)
  - YugabyteDB (DocDB, forked RocksDB)
  - MySQL (MyRocks storage engine)
  - Kafka (KRaft metadata storage)
  - Cassandra (optional storage engine)

Architecture:
  MemTable (in-memory, skip list) 
    → flush when full → SSTable (Sorted String Table) on disk
    → Background compaction merges SSTables
  
  Write path:
    1. Write to WAL (write-ahead log)
    2. Write to MemTable (skip list)
    3. Return success
    → Extremely fast writes (sequential I/O only)
  
  Read path:
    1. Check MemTable
    2. Check immutable MemTable (being flushed)
    3. Check SSTables L0 → L1 → L2 → ... (newest to oldest)
    4. Bloom filter on each SSTable to skip non-matching files
    → Reads can be slower (multiple levels to check)

  Compaction strategies:
    Level compaction: L0 → L1 → L2 (default, good read amplification)
    Universal compaction: merge all files (better write amplification)
    FIFO compaction: oldest files dropped (time-series/cache)

  Tuning knobs (hundreds):
    write_buffer_size:          MemTable size before flush
    max_write_buffer_number:    parallel MemTables
    level0_file_num_compaction_trigger: when to start compaction
    target_file_size_base:      SSTable size per level
    max_background_compactions: parallel compaction threads
    
  Write stalls:
    If compaction can't keep up with writes:
    → MemTable fills up → writes STALL (back-pressure)
    → Tuning compaction throughput is critical for write-heavy workloads
```

---

## 6. FoundationDB — The Simulation-Tested Foundation

```
FoundationDB: ordered key-value store with ACID transactions.
Acquired by Apple (2015). Powers Apple's iCloud.

Unique properties:
  - Ordered KV (keys are lexicographically sorted — scan-friendly)
  - Full ACID transactions across any keys (serializable)
  - Multi-model via "layers": document, graph, SQL built on top
  - Deterministic simulation testing (Flow framework)

Simulation testing:
  FoundationDB runs IN SIMULATION with:
    - Randomized network failures, disk failures, clock skew
    - Deterministic replay of any failure scenario
    - Millions of simulated years of testing
  → Found bugs that would take decades of real production to hit
  → Gold standard for distributed systems testing

Architecture:
  Stateless transaction processing → distributed storage
  Resolvers: check for conflicts (OCC — optimistic concurrency)
  Storage servers: hold range-partitioned data
  Coordinators: Paxos-based, store cluster configuration

Layer concept:
  FoundationDB provides: ordered KV + transactions
  Everything else is a "layer" built on top:
    Record Layer (Apple): SQL-like relational layer
    Document Layer: MongoDB-compatible API
    → Your application's data model is a layer

Limitations:
  - 5-second transaction limit (max active transaction duration)
  - 10 MB transaction size limit
  - Value size limit: 100 KB
  - Designed for OLTP, not analytics
```

---

## Key Takeaways

1. **Redis is a data structure server**, not just a cache. Sorted sets, streams, and Lua scripting make it a Swiss Army knife. But it's single-threaded — CPU-bound workloads need clustering.

2. **DynamoDB's single-table design** is unintuitive but powerful. Denormalize everything into one table with overloaded PK/SK. Design for access patterns, not entities.

3. **RocksDB is the engine inside everything.** TiDB, CockroachDB, YugabyteDB, MyRocks — they all use LSM-tree storage. Understanding RocksDB = understanding the storage layer of modern distributed databases.

4. **etcd is for coordination, not storage.** It provides linearizable reads/writes via Raft but keeps < 8 GB. Kubernetes depends on it entirely.

5. **Redis Cluster shards by hash slot**, which means multi-key operations only work within the same slot. Use hash tags `{...}` for co-location.

6. **FoundationDB's simulation testing** is why Apple trusts it with iCloud. Deterministic testing finds bugs that production testing cannot.

7. **LSM-tree trade-off**: fast writes (append-only) but slower reads (check multiple levels + Bloom filters) and write amplification (compaction). B-trees have the opposite trade-off.

---

Next: [04-document-databases.md](04-document-databases.md) →
