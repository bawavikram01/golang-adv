# 5.3 — Sharding & Partitioning

> Partitioning splits a table into pieces on ONE server.  
> Sharding distributes those pieces across MULTIPLE servers.  
> Both are tools for scale — but sharding changes everything about your application.

---

## 1. Table Partitioning (Single Node)

```sql
-- PostgreSQL declarative partitioning (PG 10+):

-- RANGE partitioning (by time — the most common):
CREATE TABLE events (
    id BIGSERIAL,
    event_type TEXT,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL
) PARTITION BY RANGE (created_at);

CREATE TABLE events_2024_q1 PARTITION OF events
    FOR VALUES FROM ('2024-01-01') TO ('2024-04-01');
CREATE TABLE events_2024_q2 PARTITION OF events
    FOR VALUES FROM ('2024-04-01') TO ('2024-07-01');
CREATE TABLE events_2024_q3 PARTITION OF events
    FOR VALUES FROM ('2024-07-01') TO ('2024-10-01');
CREATE TABLE events_2024_q4 PARTITION OF events
    FOR VALUES FROM ('2024-10-01') TO ('2025-01-01');

-- Partition pruning (automatic):
SELECT * FROM events WHERE created_at >= '2024-07-15';
-- Planner scans ONLY events_2024_q3 and events_2024_q4
-- Other partitions are eliminated at planning time

-- LIST partitioning (by discrete values):
CREATE TABLE orders (
    id BIGSERIAL, region TEXT, total DECIMAL
) PARTITION BY LIST (region);

CREATE TABLE orders_us PARTITION OF orders FOR VALUES IN ('us');
CREATE TABLE orders_eu PARTITION OF orders FOR VALUES IN ('eu');
CREATE TABLE orders_ap PARTITION OF orders FOR VALUES IN ('ap');
CREATE TABLE orders_default PARTITION OF orders DEFAULT;

-- HASH partitioning (even distribution):
CREATE TABLE sessions (
    id UUID, user_id INT, data JSONB
) PARTITION BY HASH (user_id);

CREATE TABLE sessions_0 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE sessions_1 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE sessions_2 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE sessions_3 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 3);

-- Benefits of partitioning:
-- 1. Partition pruning: queries touch fewer data
-- 2. Faster maintenance: VACUUM, REINDEX per partition
-- 3. Instant data retention: DROP partition (instead of DELETE + VACUUM)
-- 4. Parallel scans: different partitions on different workers
-- 5. Tablespace placement: hot partitions on SSD, cold on HDD

-- Gotchas:
-- Primary key MUST include partition key
-- Foreign keys to partitioned tables: limited support
-- Too many partitions = planner overhead (keep < 1000)
-- CrossPartition unique constraints not supported (must include partition key)
```

---

## 2. Sharding Strategies

```
Sharding = distributing data across multiple database servers.

Application-Level Sharding:
  Application decides which shard to query.
  shard_id = hash(user_id) % num_shards
  connection = shard_connections[shard_id]
  
  ✓ Full control, no middleware overhead
  ✗ Every query must know the shard key
  ✗ Cross-shard queries = application-level scatter-gather
  ✗ Resharding = application-level migration

Proxy-Based Sharding:
  Application → Proxy → Routes to correct shard
  
  Vitess (MySQL): transparent sharding proxy
  Citus (PostgreSQL): extension-based, coordinator routes
  ProxySQL: query routing based on rules
  
  ✓ Application doesn't need to know about shards
  ✗ Proxy adds latency
  ✗ Some queries can't be routed (require scatter-gather)

Shard Key Selection (THE most critical decision):

  Good shard keys:
    ✓ user_id (even distribution, most queries are per-user)
    ✓ tenant_id (multi-tenant SaaS, natural isolation)
    ✓ order_id (if independent order processing)
  
  Bad shard keys:
    ✗ timestamp (all writes go to latest shard = hotspot)
    ✗ country (uneven distribution: US >> Liechtenstein)
    ✗ status (few distinct values, massive imbalance)
    ✗ auto-increment ID (sequential = one hot shard)
  
  Rules:
    1. High cardinality (many unique values)
    2. Even distribution across shards
    3. Matches your most common access pattern
    4. Stable (doesn't change per row)
```

---

## 3. Cross-Shard Operations

```
The moment you shard, these become HARD:

Cross-shard JOIN:
  SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id
  If users and orders are on different shards:
    → Scatter-gather: query all shards, assemble results in application
    → Or: co-locate users and orders on the same shard (same shard key)

Cross-shard aggregation:
  SELECT COUNT(*), SUM(total) FROM orders WHERE status = 'pending'
  → Query every shard, sum results in application/proxy
  → Expensive: N network round trips for N shards

Cross-shard transaction:
  Transfer $100 from user A (shard 1) to user B (shard 2)
  → Need distributed transaction (2PC) or Saga pattern
  → Vitess: supports cross-shard transactions (with overhead)
  → Application-level: Saga with compensating actions

Cross-shard unique constraint:
  "Email must be unique across all shards"
  → Cannot enforce at database level across shards
  → Options: separate lookup table, external service, eventual uniqueness check

Design principle: MINIMIZE cross-shard operations.
  Co-locate related data on the same shard.
  Accept denormalization to avoid cross-shard JOINs.
```

---

## 4. Resharding

```
Adding/removing shards requires moving data.

Approaches:

1. Double-write migration:
   Phase 1: Write to old AND new shard layout
   Phase 2: Backfill new shards with existing data
   Phase 3: Verify consistency
   Phase 4: Switch reads to new layout
   Phase 5: Stop writing to old layout
   → Zero downtime but complex

2. Logical replication + switchover:
   Set up logical replication from old shards to new shard layout
   Wait for sync → switchover
   → Works for PostgreSQL logical replication

3. Vitess resharding:
   Built-in VReplication streams data between shards
   Atomic switchover (cuts over read/write traffic)
   
4. Pre-shard: start with more shards than needed
   Initially: 256 logical shards on 4 physical servers (64 each)
   Scale: move logical shards to new servers (no data splitting)
   → Much simpler than splitting existing shards
```

---

## Key Takeaways

1. **Partition before you shard.** Table partitioning on a single server is 10x simpler than sharding. Exhaust vertical scaling first.
2. **Range partitioning by time** is the most common and useful pattern. Partition pruning + instant data retention (DROP partition).
3. **Shard key = destiny.** Get it wrong and you'll have hot shards and cross-shard query nightmares. Design for your primary access pattern.
4. **Co-locate related data** on the same shard. Shard users and their orders by `user_id` → local JOINs, local transactions.
5. **Pre-shard with more logical shards** than physical nodes. Moving whole shards between servers is far easier than splitting shards.

---

Next: [04-backup-and-dr.md](04-backup-and-dr.md) →
