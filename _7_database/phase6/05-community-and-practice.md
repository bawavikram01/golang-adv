# 6.5 — Community, Practice & Becoming God-Tier

> Knowledge without practice is philosophy.  
> Practice without knowledge is fumbling.  
> The god-tier engineer has both — and shares them with others.

---

## 1. Build These Projects

### Tier 1 — Prove You Understand (do at least 3)

```
1. Key-Value Store from Scratch
   Language: Rust or Go
   What to implement:
     - In-memory hash map with WAL for durability
     - GET, PUT, DELETE operations
     - Crash recovery from WAL
     - Benchmark: ops/sec, latency percentiles
   Stretch: add LSM-tree storage, bloom filters, compaction

2. B+ Tree on Disk
   Language: Rust, C, or Go
   What to implement:
     - Fixed-size pages (4KB)
     - Insert with page splits
     - Delete with merge/redistribute
     - Range scan via leaf pointers
     - Persistence (write pages to file)
   Stretch: add concurrency control (latches)

3. CDC Pipeline with Debezium
   Stack: PostgreSQL + Debezium + Kafka + Elasticsearch
   What to build:
     - Source: PostgreSQL with users/orders tables
     - CDC: Debezium captures all changes
     - Sink: Elasticsearch for search, S3 for archival
     - Dashboard: real-time event counts in Grafana
   Shows: you understand real-time data pipelines

4. Multi-Tenant SaaS Database Design
   What to design:
     - Shared database, shared schema (tenant_id column)
     - Row-Level Security policies
     - Per-tenant connection pooling
     - Tenant isolation testing
     - Migration strategy for 1000+ tenants
   Shows: production database design skills

5. Real-Time Analytics Dashboard
   Stack: ClickHouse + Kafka + Grafana
   What to build:
     - Generate event stream (page views, clicks)
     - Ingest into ClickHouse via Kafka
     - Materialized views for pre-aggregation
     - Grafana dashboards with sub-second refresh
   Shows: OLAP + streaming skills
```

### Tier 2 — Prove You're Advanced (do at least 2)

```
6. Distributed Key-Value Store
   What to implement:
     - Raft consensus (leader election + log replication)
     - Multiple nodes (3 or 5)
     - Linearizable reads and writes
     - Node failure recovery
   Based on: MIT 6.5840 Labs 2-3
   This is THE project that separates advanced from god-tier.

7. Simple SQL Database
   What to implement:
     - SQL parser (use sqlparser-rs or write your own)
     - Table scan, index scan
     - Filter, project, join (at least hash join)
     - CREATE TABLE, INSERT, SELECT, WHERE
     - Buffer pool with LRU replacement
   Based on: CMU 15-445 BusTub projects
   
8. Database Load Tester
   What to build:
     - Configurable workload generator (OLTP, OLAP, mixed)
     - Connection pooling
     - Latency histogram collection (p50, p95, p99)
     - Throughput measurement
     - Report generation
   Like a simplified pgbench/sysbench/YCSB
```

### Tier 3 — God-Tier Projects

```
9. Your Own Storage Engine for PostgreSQL
   PostgreSQL's Table Access Method API (since PG 12) lets you
   build custom storage engines.
   - Implement a columnar storage engine
   - Or an LSM-tree based engine
   - Register it as an extension
   This demonstrates TRUE mastery of PostgreSQL internals.

10. Contribute to an Open-Source Database
    Start with:
    - PostgreSQL: fix a bug, optimize a query plan, improve docs
    - DuckDB: active community, welcoming to contributors
    - CockroachDB: good "good first issue" labels
    - TiKV: if you know Rust
    Steps:
    1. Clone the repo, build from source
    2. Read contributing guidelines
    3. Find a "good first issue" or fix a bug you've encountered
    4. Submit a PR, respond to review feedback
    5. One merged PR to a major database = career-defining
```

---

## 2. System Design Practice

```
Database-heavy system design questions (practice these):

1. Design Twitter
   - Fan-out on write vs fan-out on read
   - Timeline cache (Redis sorted sets)
   - Tweet storage (sharded by user_id)
   - Social graph (follower/following — graph or adjacency list?)

2. Design a URL Shortener
   - Base62 encoding, ID generation (Snowflake IDs)
   - Read-heavy: cache with Redis, 301 redirects
   - Analytics: ClickHouse or TimescaleDB for click tracking
   - Sharding strategy for billions of URLs

3. Design Uber / Ride-Sharing
   - Geospatial indexing (PostGIS, H3, Geohash)
   - Real-time driver location (Redis with TTL)
   - Trip database (PostgreSQL, sharded by city)
   - Surge pricing (stream processing)

4. Design a Chat System (WhatsApp)
   - Message storage (Cassandra — write-heavy, time-series-like)
   - Message delivery (message queues per user)
   - Read receipts (last-read pointer per conversation)
   - Group chats (fan-out problem again)

5. Design a Notification System
   - Event-driven: CDC → Kafka → notification service
   - Template storage, preference storage
   - Rate limiting, deduplication
   - Multi-channel (push, email, SMS)

For each design:
  - Draw the data model (ER diagram)
  - Estimate data volume and growth
  - Choose the right database(s) for each component
  - Design the sharding/partitioning strategy
  - Identify bottlenecks and design for failure
```

---

## 3. Blogs & Resources to Follow

```
Must-follow blogs:

Technical depth:
  - Brandur Leach (brandur.org) — PostgreSQL, ACID, idempotency
  - use-the-index-luke.com (Markus Winand) — Indexing and SQL tuning
  - Percona Blog — MySQL/PostgreSQL/MongoDB production insights
  - pganalyze Blog — PostgreSQL performance analysis
  - CockroachDB Blog — Distributed systems engineering
  - Jepsen.io (Kyle Kingsbury) — Correctness testing of distributed DBs

Industry perspectives:
  - Andy Pavlo's blog + talks — Database landscape, opinions
  - The Morning Paper (archived) — Paper summaries by Adrian Colyer
  - Jack Vanlightly — RabbitMQ, Kafka, distributed systems
  - Murat Demirbas (muratbuffalo.blogspot.com) — Distributed systems papers

Newsletters:
  - DB Weekly (dbweekly.com)
  - Postgres Weekly (postgresweekly.com)
  - Console.dev — Developer tools including databases

Podcasts:
  - Software Engineering Daily (database episodes)
  - The Changelog (database episodes)

YouTube:
  - CMU Database Group (Andy Pavlo's lectures + talks)
  - Hussein Nasser — Database and backend topics
```

---

## 4. Conferences

```
Academic (papers presented here first):
  SIGMOD — ACM Conference on Management of Data (the top venue)
  VLDB — Very Large Data Bases (equally prestigious)
  ICDE — International Conference on Data Engineering

Industry:
  PGConf — PostgreSQL conferences (multiple worldwide)
  Percona Live — MySQL, PostgreSQL, MongoDB production
  re:Invent — AWS (database service announcements)
  Google Cloud Next — GCP (Spanner, AlloyDB, BigQuery)
  Kafka Summit — Streaming and event-driven architectures

Watch the talks on YouTube even if you can't attend.
VLDB and SIGMOD publish proceedings freely online.
```

---

## 5. SQL Practice Platforms

```
For interview prep and skill sharpening:

LeetCode SQL (leetcode.com/problemset/database/)
  - 50+ SQL problems, Easy to Hard
  - Great for interview prep
  - Focus on: window functions, CTEs, complex JOINs

HackerRank SQL (hackerrank.com/domains/sql)
  - Structured progression (basic → advanced)
  - Good for beginners building confidence

StrataScratch (stratascratch.com)
  - Real interview questions from FAANG companies
  - Python + SQL
  - Best for data engineering interviews

SQLZoo (sqlzoo.net)
  - Interactive tutorials
  - Good for quick practice

PgExercises (pgexercises.com)
  - PostgreSQL-specific
  - Practical scenarios (club membership system)
  - Progressive difficulty

DataLemur (datalemur.com)
  - SQL interview questions
  - Curated by a Meta data scientist
```

---

## 6. The God-Tier Checklist

```
You've reached god-tier when you can confidently:

□ Design
  □ Model a complex domain in 3NF and denormalize appropriately
  □ Design a star schema for analytics
  □ Choose between SQL, document, graph, time-series, search for any use case
  □ Design a sharding strategy with proper shard key selection
  □ Design a multi-region, high-availability database architecture

□ Build
  □ Implement a B+ tree that persists to disk
  □ Implement a WAL with crash recovery
  □ Implement MVCC visibility checks
  □ Build a query executor (at least Volcano model)
  □ Implement Raft consensus

□ Operate
  □ Read an EXPLAIN plan and know exactly what to fix
  □ Tune PostgreSQL for a specific workload (OLTP vs OLAP)
  □ Set up streaming replication with automatic failover
  □ Perform zero-downtime schema migrations
  □ Set up monitoring, alerting, and capacity planning
  □ Execute disaster recovery (backup → PITR → verify)

□ Debug
  □ Diagnose lock contention and deadlocks
  □ Identify and fix bloat (table + index)
  □ Trace a slow query from application to disk I/O
  □ Diagnose replication lag and fix it
  □ Handle XID wraparound before it becomes critical

□ Think
  □ Explain the CAP theorem AND its limitations
  □ Compare isolation levels (Read Committed vs Serializable vs SSI)
  □ Explain ARIES recovery in detail
  □ Discuss Dynamo vs Spanner trade-offs
  □ Explain why Aurora's "log is the database" architecture works

□ Teach
  □ Write blog posts that make complex topics accessible
  □ Give a talk on database internals at a meetup
  □ Mentor others on database design and optimization
  □ Contribute to open-source database projects
  □ Answer database questions on Stack Overflow with citations

When you can check every box: you are god-tier.
Not because you memorized facts — because you understand the WHY
behind every design decision in every database ever built.
```

---

## Key Takeaways

1. **Build projects, not just read.** The key-value store, B+ tree, and Raft implementation are the three projects that will teach you the most. Do all three.
2. **Contribute to open source.** One merged PR to PostgreSQL, DuckDB, or CockroachDB is worth more than 100 LeetCode problems for your career.
3. **System design = database design.** Every system design question is fundamentally about how data is stored, queried, replicated, and scaled. The god-tier database engineer is automatically a great system designer.
4. **Teach what you learn.** Writing a blog post forces you to truly understand a topic. Explaining ARIES or MVCC in your own words reveals every gap in your knowledge.
5. **Consistency over intensity.** One paper per week, one chapter per week, one project per quarter. In 12 months you'll have read 50+ papers, 5+ books, and built 3+ projects. That's god-tier.

---

**You've completed the entire roadmap.**

```
Phase 1 ✅ SQL & Relational Foundations
Phase 2 ✅ Database Internals
Phase 3 ✅ RDBMS Deep Dives
Phase 4 ✅ NoSQL & Distributed Systems
Phase 5 ✅ Production Engineering
Phase 6 ✅ God Tier

Total: 37 files, 6 phases, one path to mastery.
The knowledge is here. The execution is yours.
```
