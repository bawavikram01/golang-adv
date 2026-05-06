# 6.4 — Books & Courses — The Canon

> These are the texts that database engineers swear by.  
> Not a reading list you skim. A library you build over years.

---

## 1. The Essential Five (Read These First)

### "Designing Data-Intensive Applications" — Martin Kleppmann (2017)

```
★★★★★ — THE most important book for modern data engineering.

If you read one book on databases, make it this one.
Not about any single database — about the PRINCIPLES behind all of them.

What you'll learn:
  Part I:  Foundations of data systems (encoding, replication, partitioning)
  Part II: Distributed data (consistency, consensus, transactions)
  Part III: Derived data (batch processing, stream processing)

Why it's essential:
  - Bridges theory and practice better than any other book
  - Covers the landscape: SQL, NoSQL, streaming, batch, consensus
  - Explains trade-offs honestly (no vendor bias)
  - Referenced in every database engineering interview
  - Makes you dangerous in system design discussions

Read when: You've finished Phases 1-3 of this roadmap.
Time: 2-3 weeks of focused reading.
```

### "Database Internals" — Alex Petrov (2019)

```
★★★★★ — The best book for understanding HOW databases work under the hood.

Two parts:
  Part I:  Storage engines (B-trees, LSM-trees, pages, buffer pool)
  Part II: Distributed systems (replication, consensus, anti-entropy)

Why it's essential:
  - Covers B-trees at a level of detail no other book does
  - Explains LSM-tree compaction strategies (leveled, tiered, FIFO)
  - Connects theory (papers) to practice (real systems)
  - Distributed section covers gossip, failure detection, Raft, Paxos

Read when: You're ready to build your own database.
Time: 2-3 weeks.
Pairs with: Phase 6.1 (Build Your Own DB) and 6.3 (Research Papers).
```

### "PostgreSQL 14 Internals" — Egor Rogov (2023)

```
★★★★★ — The deepest dive into PostgreSQL's architecture.

What you'll learn:
  - Buffer pool and shared memory layout
  - MVCC implementation (xmin, xmax, freezing, visibility map)
  - VACUUM internals (why it exists, how it works, when it fails)
  - WAL and crash recovery implementation
  - B-tree, GiST, GIN, BRIN index internals
  - Query planner and optimizer internals
  - Lock system architecture
  - TOAST (large object storage)

Why it's essential:
  - Written by a Postgres developer
  - Explains the actual source code flow
  - Diagrams of in-memory data structures
  - Free online: https://postgrespro.com/community/books/internals

Read when: You're reading PostgreSQL source code (Phase 6.2).
Time: 3-4 weeks.
```

### "The Art of PostgreSQL" — Dimitri Fontaine (2020)

```
★★★★☆ — Master-level SQL using PostgreSQL.

Not about internals — about USING PostgreSQL like an expert.

Covers:
  - Advanced SQL: window functions, CTEs, lateral joins
  - Full-text search, JSONB, array operations
  - Data modeling patterns
  - When to use stored procedures vs application code
  - Practical ETL with PostgreSQL
  - Real-world case studies

Why it's valuable:
  - Shows that PostgreSQL can replace many "specialized" tools
  - Practical, production-oriented examples
  - Bridges the gap between knowing SQL and mastering it

Read when: You've finished Phase 1 and want deep SQL mastery.
Time: 2 weeks.
```

### "SQL Performance Explained" — Markus Winand (2012)

```
★★★★★ — The definitive guide to SQL indexing and query tuning.

Everything you need to never write a slow query again.
Also available free online: https://use-the-index-luke.com

Covers:
  - B-tree internals (just enough to understand indexes)
  - The where clause (how predicates use indexes)
  - Index-only scans (covering indexes)
  - Sorting and grouping with indexes
  - Partial results (LIMIT + offset pitfalls)
  - Join operations (nested loop, hash, merge)
  - DML performance (insert/update/delete impact on indexes)

Why it's essential:
  - Practical, not academic
  - Examples work on PostgreSQL, MySQL, Oracle, SQL Server
  - You will use this knowledge every single day

Read when: Phase 1 (SQL) or Phase 5 (Performance tuning).
Time: 1 week (it's concise).
```

---

## 2. Deep Dive Books

### Systems & Theory

```
"Database System Concepts" — Silberschatz, Korth, Sudarshan
  The standard university textbook. Comprehensive but academic.
  Covers: relational algebra, SQL, storage, indexing, transactions, recovery.
  Read if: you want rigorous formal foundations.

"Fundamentals of Database Systems" — Elmasri & Navathe
  Alternative textbook. More detail on ER modeling and normalization.
  Read if: you prefer a different writing style than Silberschatz.

"Transaction Processing: Concepts and Techniques" — Gray & Reuter (1993)
  THE reference on transaction processing. Jim Gray (Turing Award winner).
  Deep coverage of ACID, locking, recovery, distributed transactions.
  Read if: you're implementing a transaction system.
  Warning: dense, 1000+ pages. Use as reference, not cover-to-cover.

"Readings in Database Systems" — Hellerstein & Stonebraker (Red Book, 5th ed.)
  Curated collection of the most important database papers with commentary.
  Free online: http://www.redbook.io
  Read if: you want guided entry into the research literature.
```

### Specific Technologies

```
"High Performance MySQL" — Schwartz, Zaitsev, Tkachenko (4th ed.)
  The MySQL bible. InnoDB internals, replication, query optimization.
  Read if: you work with MySQL in production.

"Redis in Action" — Josiah Carlson
  Practical Redis: data structures, caching patterns, messaging.
  Read if: you use Redis (and you probably do).

"MongoDB: The Definitive Guide" — Shannon Bradshaw (3rd ed.)
  WiredTiger, replication, sharding, aggregation pipeline.
  Read if: you work with MongoDB.

"Streaming Systems" — Akidau, Chernyak, Lax
  The theory behind streaming: windows, triggers, watermarks, exactly-once.
  Written by the creators of Apache Beam (Google Dataflow).
  Read if: you work with Flink, Kafka Streams, Spark Streaming.

"The Data Warehouse Toolkit" — Ralph Kimball (3rd ed.)
  The definitive guide to dimensional modeling.
  Star schemas, slowly changing dimensions, ETL best practices.
  Read if: you do data warehousing or analytics engineering.

"Understanding Distributed Systems" — Roberto Vitillo
  Gentler introduction to distributed systems than DDIA.
  Covers: networking, consensus, replication, caching.
  Read if: you want a stepping stone before DDIA.
```

---

## 3. Courses & Lectures

### CMU 15-445/645 — Introduction to Database Systems

```
★★★★★ — THE database course. Andy Pavlo. Free on YouTube.

THE single best free resource for learning database internals.

Topics:
  Lectures 1-5:   SQL, relational algebra, storage
  Lectures 6-10:  Buffer pool, hash tables, B+ trees
  Lectures 11-15: Sorting, joins, query processing
  Lectures 16-18: Query optimization
  Lectures 19-22: Concurrency control (2PL, MVCC, OCC)
  Lectures 23-25: Recovery (WAL, ARIES, checkpointing)
  Lectures 26:    Distributed databases

Projects (do these in BusTub):
  Project 1: Buffer Pool Manager
  Project 2: B+ Tree Index
  Project 3: Query Execution
  Project 4: Concurrency Control

YouTube: search "CMU 15-445 Fall 2023"
Course site: https://15445.courses.cs.cmu.edu
BusTub: https://github.com/cmu-db/bustub

Time: ~40 hours of lectures + ~80 hours of projects.
This alone will put you in the top 1% of database understanding.
```

### CMU 15-721 — Advanced Database Systems

```
★★★★★ — The sequel. For after you've done 15-445.

Topics:
  - In-memory databases (no buffer pool needed)
  - MVCC design decisions (comparison of all approaches)
  - OLAP execution: vectorized vs compiled
  - Query optimization deep dive (Cascades, Volcano)
  - Networking and query execution models
  - Query compilation (JIT)
  - Vectorized execution
  - Modern storage (columnar, compression)

Each lecture covers 2-3 research papers.
Andy Pavlo makes papers accessible and entertaining.

YouTube: search "CMU 15-721 Spring 2024"
Course site: https://15721.courses.cs.cmu.edu

Time: ~30 hours of lectures. No projects (paper reading instead).
```

### MIT 6.5840 (formerly 6.824) — Distributed Systems

```
★★★★★ — The distributed systems course. Robert Morris / Frans Kaashoek.

Not database-specific, but essential for distributed databases.

Topics:
  - MapReduce
  - Raft consensus (you implement this!)
  - Fault tolerance
  - Linearizability, eventual consistency
  - Zookeeper
  - Distributed transactions (two-phase commit, Spanner)
  - CRDTs

Labs:
  Lab 1: MapReduce
  Lab 2: Raft (leader election, log replication, persistence)
  Lab 3: Fault-tolerant key-value store (on top of Raft)
  Lab 4: Sharded key-value store

Course site: https://pdos.csail.mit.edu/6.824/
Labs are in Go.

Time: ~30 hours of lectures + ~100 hours of labs.
The Raft lab alone teaches more about distributed systems than most books.
```

### Other Valuable Courses

```
UC Berkeley CS186 — Introduction to Database Systems
  Alternative to CMU 15-445  with different perspective.
  Available on YouTube. Java-based projects.

Stanford CS245 — Principles of Data-Intensive Systems
  Research-oriented. Covers recent papers and systems.
  Good if you want academic depth.

use-the-index-luke.com
  Free online "course" on SQL indexing and tuning.
  By Markus Winand (author of SQL Performance Explained).
  Interactive, with visual execution plan explanations.

PgExercises (pgexercises.com)
  Free SQL practice with a PostgreSQL focus.
  Progressively harder exercises.
```

---

## 4. Recommended Learning Path

```
The order matters. Each resource builds on the previous.

Month 1-2: Foundations
  □ SQL Performance Explained (or use-the-index-luke.com)
  □ CMU 15-445 Lectures 1-12 (storage, indexes, sorting)

Month 3-4: Core Systems
  □ CMU 15-445 Lectures 13-26 (optimization, concurrency, recovery)
  □ CMU 15-445 Projects (BusTub — buffer pool, B+ tree, executor)

Month 5-6: Distributed & Modern
  □ Designing Data-Intensive Applications (full book)
  □ MIT 6.5840 Lab 2 (implement Raft)

Month 7-8: Deep Internals
  □ Database Internals (Alex Petrov)
  □ PostgreSQL 14 Internals (Egor Rogov)
  □ Start reading PostgreSQL/SQLite source code

Month 9-10: Research & Build
  □ Tier 1 papers (one per week)
  □ CMU 15-721 (advanced topics)
  □ Start building your own database

Month 11-12: Mastery
  □ Finish your database project
  □ Read CockroachDB or TiKV source code
  □ Contribute to an open-source database
  □ Write blog posts about what you learned

After 12 months of this: you are god-tier.
```

---

## Key Takeaways

1. **DDIA (Kleppmann) is the #1 book.** It bridges theory and practice for data systems. Read it cover-to-cover. You'll reference it for your entire career.
2. **CMU 15-445 (Andy Pavlo) is the #1 course.** Free on YouTube, with hands-on projects. Doing the BusTub projects teaches you more about database internals than years of production work.
3. **SQL Performance Explained / use-the-index-luke.com** is the fastest ROI. One week of reading → immediately faster queries for the rest of your career.
4. **PostgreSQL 14 Internals** (free) pairs perfectly with reading PostgreSQL source code. Read a chapter, then read the corresponding source files.
5. **The 12-month path** (above) is realistic and will take you from competent to god-tier. Consistency matters more than speed — one paper/chapter per week compounds dramatically.

---

Next: [05-community-and-practice.md](05-community-and-practice.md) →
