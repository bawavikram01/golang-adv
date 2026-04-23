# 5.10 — Cloud Database Services

> Running databases yourself teaches you everything.  
> Running databases on managed services keeps you sane.  
> Know both — master the internals, then let the cloud handle the toil.

---

## 1. AWS Database Services

### RDS (Relational Database Service)

```
Managed PostgreSQL, MySQL, MariaDB, Oracle, SQL Server.
You pick the engine; AWS handles patching, backups, failover.

What AWS manages:           What YOU manage:
  - OS patching               - Schema design
  - Engine patching           - Query optimization
  - Automated backups         - Indexing
  - Multi-AZ failover         - Connection pooling
  - Storage scaling           - Parameter tuning
  - Monitoring (CloudWatch)   - Security groups / IAM

Instance classes:
  db.t4g.micro   → dev/test (burstable, 2 vCPU, 1GB)
  db.r6g.xlarge  → production (4 vCPU, 32GB, memory-optimized)
  db.r6g.16xlarge → heavy production (64 vCPU, 512GB)

Storage types:
  gp3:  General purpose SSD (3000 IOPS baseline, scale to 16K)
  io2:  Provisioned IOPS (up to 256K IOPS) — latency-sensitive
  
Key features:
  - Multi-AZ: synchronous replication to standby → automatic failover (~60s)
  - Read replicas: async replication, up to 15 replicas (cross-region too)
  - Automated backups: daily snapshots + transaction logs → PITR (up to 35 days)
  - Performance Insights: built-in query analysis (like pg_stat_statements)
  - IAM authentication: no passwords, use IAM tokens
```

### Aurora

```
AWS-reengineered MySQL/PostgreSQL (same wire protocol, different storage).

Architecture:
  ┌─────────────────────────────────────────────────────┐
  │                  Compute Layer                       │
  │   Writer Instance ──── Reader Instance(s)           │
  │        │                     │                       │
  └────────┼─────────────────────┼───────────────────────┘
           │                     │
  ┌────────┼─────────────────────┼───────────────────────┐
  │        ▼     Shared Storage Layer     ▼              │
  │   ┌─────────────────────────────────────────────┐    │
  │   │  6 copies across 3 AZs (quorum writes: 4/6) │    │
  │   │  Auto-scales 10GB → 128TB                    │    │
  │   │  Continuous backup to S3 (no impact)          │    │
  │   └─────────────────────────────────────────────┘    │
  └──────────────────────────────────────────────────────┘

Why Aurora?
  - 5x MySQL / 3x PostgreSQL throughput (AWS claims)
  - Storage auto-scales (pay for what you use, no provisioning)
  - Failover in <30 seconds (reader promoted instantly)
  - Up to 15 read replicas with <20ms replica lag
  - Backtrack: rewind DB to any point in last 72 hours (no restore needed)
  - Global Database: <1 second cross-region replication
  - Cloning: instant zero-copy clone for dev/test

Aurora Serverless v2:
  - Scales compute automatically (0.5 → 256 ACUs)
  - Sub-second scaling (no cold starts for v2)
  - Good for: variable workloads, dev/test, infrequent access
  - Cost: pay per ACU-second (~$0.12/ACU-hour)

When to use Aurora vs RDS:
  Aurora: high availability critical, need auto-scaling storage, >1 read replica
  RDS: cost-sensitive, need Oracle/SQL Server, simpler workloads
```

### DynamoDB

```
Fully serverless key-value / document store (see Phase 4).

Capacity modes:
  On-demand:    Pay per request (~$1.25 per million writes)
                No capacity planning. Good for unpredictable traffic.
  Provisioned:  Reserve RCU/WCU (cheaper for steady workloads)
                Auto-scaling available.

Key features beyond basics:
  - Global Tables: multi-region, active-active replication
  - DAX (DynamoDB Accelerator): in-memory cache, microsecond reads
  - Streams: CDC for DynamoDB (triggers Lambda on changes)
  - PartiQL: SQL-compatible query language for DynamoDB
  - Export to S3: full table export in Parquet/JSON for analytics
  - Point-in-time recovery (PITR): restore to any second in last 35 days
```

### Other AWS Database Services

```
ElastiCache:    Managed Redis / Memcached
MemoryDB:       Redis-compatible with durability (WAL)
Neptune:        Graph database (Gremlin + SPARQL)
Timestream:     Time-series database (serverless)
QLDB:           Immutable ledger database (cryptographic verification)
DocumentDB:     MongoDB-compatible (Aurora storage engine underneath)
Keyspaces:      Cassandra-compatible (serverless)
Redshift:       Data warehouse (columnar, MPP, Parquet/Spectrum)
  Redshift Serverless: auto-scaling compute, pay per query
```

---

## 2. GCP Database Services

### Cloud SQL

```
Managed PostgreSQL, MySQL, SQL Server (equivalent to AWS RDS).
  - High availability with regional instances (synchronous replication)
  - Read replicas (cross-region)
  - Automated backups + PITR
  - Integrated with Cloud IAM + VPC
```

### Cloud Spanner

```
Globally distributed, strongly consistent, relational database.
THE unique GCP offering — nothing else like it.

Properties:
  - Horizontally scalable SQL database
  - External consistency (strongest consistency model possible)
  - 99.999% availability SLA (5 nines!)
  - Automatic sharding + rebalancing
  - Global distribution with TrueTime (atomic clocks + GPS)

How TrueTime enables global consistency:
  Traditional:  "What time is it?" → clock skew between nodes → uncertainty
  TrueTime:     "Time is between [earliest, latest]" → bounded uncertainty (~7ms)
  Spanner:      Waits for uncertainty window to pass → guarantees ordering
                No other database can do this without specialized hardware

-- Interleaved tables (parent-child co-location):
CREATE TABLE Singers (
  SingerId   INT64 NOT NULL,
  Name       STRING(1024),
) PRIMARY KEY (SingerId);

CREATE TABLE Albums (
  SingerId   INT64 NOT NULL,
  AlbumId    INT64 NOT NULL,
  Title      STRING(1024),
) PRIMARY KEY (SingerId, AlbumId),
  INTERLEAVE IN PARENT Singers ON DELETE CASCADE;
-- Albums physically stored with their Singer → fast joins

When to use: global applications needing strong consistency + SQL
Cost: expensive ($0.90/node-hour minimum, 3 nodes for production = ~$2K/month)
```

### AlloyDB

```
GCP's answer to Aurora. PostgreSQL-compatible with:
  - Separated compute and storage (like Aurora)
  - Columnar engine for analytics queries (auto-accelerates OLAP queries)
  - ML-driven adaptive caching
  - 4x faster than standard PostgreSQL (GCP claims)
  - 100% PostgreSQL compatible (unlike Spanner)

Good for: HTAP workloads (OLTP + OLAP on same database)
```

### BigQuery

```
Serverless data warehouse. Pay per query (bytes scanned).

Key features:
  - Columnar storage (Capacitor format)
  - Separated compute and storage
  - Auto-scales to petabytes
  - Built-in ML (BigQuery ML — train models with SQL)
  - BI Engine (in-memory analysis, sub-second queries)
  - Streaming inserts (real-time ingestion)

-- Cost optimization:
-- $6.25 per TB scanned (on-demand pricing)
-- Partition and cluster tables to reduce scanned data:
CREATE TABLE dataset.events
PARTITION BY DATE(event_timestamp)
CLUSTER BY user_id, event_type
AS SELECT * FROM dataset.raw_events;

-- Use column projection (SELECT only needed columns):
SELECT user_id, SUM(amount) FROM orders GROUP BY 1;  -- scans 2 columns
-- NOT: SELECT * FROM orders;                         -- scans ALL columns
```

---

## 3. Azure Database Services

### Cosmos DB

```
Multi-model, globally distributed database.
Microsoft's flagship NoSQL offering.

APIs (choose one per container):
  - NoSQL (document — native, recommended)
  - MongoDB (wire-protocol compatible)
  - Cassandra (wire-protocol compatible)
  - Gremlin (graph)
  - Table (key-value, Azure Table Storage compatible)
  - PostgreSQL (Citus-based distributed PostgreSQL)

Consistency levels (5 choices — unique to Cosmos DB):
  ┌─────────────────────────────────────────────────────┐
  │ Strong                                               │
  │   ↓  Bounded Staleness (max lag: K versions or T time)│
  │   ↓  Session (read-your-writes within session) ★default│
  │   ↓  Consistent Prefix (ordered, may be stale)       │
  │ Eventual                                             │
  └─────────────────────────────────────────────────────┘
  Each level trades latency for consistency.
  Session consistency is the sweet spot for most apps.

Partitioning:
  - Choose partition key carefully (high cardinality, even distribution)
  - All queries should include partition key (cross-partition = expensive)
  - Physical partitions auto-split at 50GB / 10K RU/s

Request Units (RU):
  1 RU = cost of reading 1 item (1KB) by ID
  Point read: 1 RU
  Write 1KB: ~5 RU
  Query: varies by complexity
  Provisioned: reserve RU/s (autoscale available)
  Serverless: pay per RU consumed
```

---

## 4. Serverless Databases

```
The trend: remove ALL operational overhead.

┌───────────────────── ─────────── ───────────────────────────────┐
│ Service              Provider    Notes                          │
├───────────────────── ─────────── ───────────────────────────────┤
│ Aurora Serverless v2  AWS        Scale 0.5-256 ACUs, sub-second│
│ Neon                  -          Serverless PostgreSQL, branching│
│ PlanetScale           -          Serverless MySQL (Vitess-based)│
│ CockroachDB Serverless Cockroach  Distributed SQL, free tier   │
│ Turso                 -          Edge SQLite (libSQL fork)      │
│ D1                    Cloudflare  SQLite at the edge            │
│ Supabase              -          PostgreSQL + auth + realtime   │
│ BigQuery              GCP        Pay per query                  │
│ DynamoDB on-demand    AWS        Pay per request                │
│ Cosmos DB serverless   Azure     Pay per RU                     │
└───────────────────── ─────────── ───────────────────────────────┘

Neon (notable):
  - Separates compute (Postgres) and storage (custom page server)
  - Scale to zero (pay nothing when idle)
  - Branching: create instant copy of your DB (like git branch)
    → Use for dev, testing, migrations, preview deployments
  - Copy-on-write: branches share pages until modified

PlanetScale (notable):
  - Built on Vitess (YouTube's MySQL sharding layer)
  - Non-blocking schema changes (like gh-ost, built in)
  - Deploy requests (schema change PRs with diff + review)
  - Read replicas, analytics replicas
```

---

## 5. Cost Optimization

```
Cloud databases can get VERY expensive. Key strategies:

1. Right-size instances:
   - Use Performance Insights / Cloud Monitoring to check actual CPU/memory usage
   - Most databases are over-provisioned by 2-4x
   - Start small, scale up when metrics justify it

2. Reserved instances (commitment discounts):
   - AWS: 1-year RI = ~30-40% savings, 3-year = ~50-60%
   - GCP: Committed Use Discounts (CUDs)
   - Azure: Reserved capacity

3. Storage optimization:
   - Aurora: auto-scales (no wasted storage)
   - RDS: watch for over-provisioned EBS volumes
   - Delete old snapshots (they cost $0.02-0.10/GB/month)

4. Read replica optimization:
   - Route read traffic to replicas (save writes for primary)
   - Consider caching (ElastiCache) before adding replicas

5. Data lifecycle:
   - Archive old data to S3/GCS (cheapest storage)
   - Use table partitioning + partition dropping
   - TTL in DynamoDB / Cosmos DB for auto-expiry

6. Query cost awareness:
   - BigQuery: partition + cluster tables (reduce scanned bytes)
   - DynamoDB: avoid scans, use queries with partition keys
   - Cosmos DB: include partition key in every query

7. Dev/test environments:
   - Use Aurora cloning (instant, zero-copy)
   - Neon branches (free, instant)
   - Scale down or stop dev databases on nights/weekends
   - Use Serverless tiers for dev (Aurora Serverless, Neon, PlanetScale)

Monthly cost reference (production, us-east-1):
  RDS db.r6g.xlarge (PG):    ~$550/month (on-demand)
  Aurora (same compute):      ~$650/month (storage separate)
  Cloud SQL (n2-standard-4):  ~$450/month
  Spanner (3 nodes):          ~$2,000/month
  DynamoDB (on-demand, 10M writes/mo): ~$12.50/month
```

---

## 6. Multi-Cloud and Portability

```
Avoiding vendor lock-in is a spectrum:

Maximum lock-in (hardest to leave):
  DynamoDB, Spanner, Cosmos DB, BigQuery
  → Proprietary APIs, unique consistency models, no equivalent elsewhere

Moderate lock-in:
  Aurora, AlloyDB, Cloud SQL
  → PostgreSQL/MySQL compatible but with proprietary extensions
  → Migration requires testing (Aurora-specific features won't work elsewhere)

Minimal lock-in:
  RDS PostgreSQL/MySQL (standard engines)
  Self-managed on EC2/GCE/VMs
  → Standard PostgreSQL, can move anywhere

Portability strategies:
  1. Use standard SQL as much as possible
  2. Abstract database access behind repository patterns
  3. Avoid proprietary extensions unless the value is high
  4. Use Terraform to define infrastructure (multi-cloud IaC)
  5. Test against standard PostgreSQL (even if running Aurora)
  6. Use open formats for data (Parquet, Iceberg) over proprietary

Reality check:
  - Most companies will NEVER migrate clouds
  - Lock-in risk is often overstated vs. operational benefits
  - Choose the best tool for the job; optimize for productivity
  - If portability matters: stick to standard PostgreSQL on managed service
```

---

## Key Takeaways

1. **Aurora** (AWS) and **AlloyDB** (GCP) separate compute from storage, giving auto-scaling, fast failover, and instant cloning. They're the managed PostgreSQL sweet spot.
2. **Cloud Spanner** is the only globally distributed, strongly consistent SQL database — enabled by hardware-based TrueTime. Nothing else like it.
3. **Serverless databases** (Neon, PlanetScale, Aurora Serverless v2) eliminate capacity planning. Scale to zero for dev, burst for production.
4. **Cost optimization** starts with right-sizing. Most production databases are over-provisioned by 2-4x. Monitor before scaling.
5. **Vendor lock-in is real but overrated.** Use standard PostgreSQL-compatible services, and you can migrate if needed. Optimize for productivity, not theoretical portability.

---

Next: [11-database-devops.md](11-database-devops.md) →
