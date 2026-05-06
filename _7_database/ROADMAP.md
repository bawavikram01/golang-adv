# Database Mastery Roadmap — From Zero to God Mode

> A structured, exhaustive roadmap covering every dimension of database engineering.
> Estimated depth: **5 phases**, **30+ topics**, **hundreds of sub-skills**.

---

## Phase 0 — Foundations (Pre-requisites)

### 0.1 Computer Science Fundamentals
- [ ] Data structures: arrays, linked lists, hash tables, trees (BST, B-tree, B+ tree, LSM tree), graphs, heaps, tries
- [ ] Algorithms: sorting, searching, hashing, graph traversal (BFS/DFS)
- [ ] Big-O complexity analysis (time & space)
- [ ] Bit manipulation and binary representations
- [ ] Memory hierarchy: registers → cache → RAM → SSD → HDD
- [ ] How OS manages memory (virtual memory, paging, mmap)

### 0.2 Operating Systems Basics
- [ ] Processes vs threads
- [ ] Context switching
- [ ] File systems (ext4, XFS, ZFS, NTFS)
- [ ] I/O models: blocking, non-blocking, multiplexing (epoll, kqueue)
- [ ] System calls: read, write, fsync, fdatasync
- [ ] Page cache and buffer cache

### 0.3 Networking Basics
- [ ] TCP/IP fundamentals
- [ ] Sockets and connection pooling
- [ ] Client-server architecture
- [ ] TLS/SSL basics
- [ ] DNS and service discovery

### 0.4 Discrete Math & Logic
- [ ] Set theory (unions, intersections, differences)
- [ ] Predicate logic and boolean algebra
- [ ] Relations and functions
- [ ] Proof techniques (used in correctness arguments)

---

## Phase 1 — Relational Databases & SQL Mastery

### 1.1 Relational Model Theory
- [ ] Edgar F. Codd's 12 rules
- [ ] Relations, tuples, attributes, domains
- [ ] Relational algebra: σ (select), π (project), ⋈ (join), ∪, ∩, −, ×
- [ ] Relational calculus: tuple relational calculus, domain relational calculus
- [ ] Functional dependencies
- [ ] Closure of attribute sets
- [ ] Candidate keys, super keys, primary keys, foreign keys
- [ ] Multivalued dependencies

### 1.2 SQL — Beginner
- [ ] DDL: CREATE, ALTER, DROP, TRUNCATE
- [ ] DML: INSERT, UPDATE, DELETE, MERGE/UPSERT
- [ ] DQL: SELECT, FROM, WHERE, GROUP BY, HAVING, ORDER BY, LIMIT/OFFSET
- [ ] Data types: numeric, string, date/time, boolean, binary, JSON, UUID
- [ ] Constraints: NOT NULL, UNIQUE, PRIMARY KEY, FOREIGN KEY, CHECK, DEFAULT
- [ ] Basic joins: INNER, LEFT, RIGHT, FULL OUTER, CROSS
- [ ] Aliases, DISTINCT, CASE WHEN, COALESCE, NULLIF
- [ ] Aggregate functions: COUNT, SUM, AVG, MIN, MAX
- [ ] Subqueries: scalar, row, table, correlated

### 1.3 SQL — Intermediate
- [ ] Window functions: ROW_NUMBER, RANK, DENSE_RANK, NTILE
- [ ] Window frames: ROWS BETWEEN, RANGE BETWEEN, GROUPS BETWEEN
- [ ] Analytic functions: LAG, LEAD, FIRST_VALUE, LAST_VALUE, NTH_VALUE
- [ ] Running totals, moving averages, percentiles
- [ ] Common Table Expressions (CTEs) — non-recursive and recursive
- [ ] Recursive queries: hierarchical data, graph traversal, series generation
- [ ] UNION, INTERSECT, EXCEPT (and ALL variants)
- [ ] GROUPING SETS, CUBE, ROLLUP
- [ ] LATERAL joins
- [ ] Pivoting and unpivoting data
- [ ] String functions, date functions, math functions
- [ ] Regular expressions in SQL
- [ ] JSON/JSONB querying and manipulation

### 1.4 SQL — Advanced
- [ ] Query optimization: reading and understanding EXPLAIN / EXPLAIN ANALYZE
- [ ] Cost-based vs rule-based optimization
- [ ] Join algorithms: nested loop, hash join, merge join
- [ ] Index scan vs sequential scan vs bitmap scan
- [ ] Materialized views and incremental refresh
- [ ] Dynamic SQL and prepared statements
- [ ] Stored procedures and functions (PL/pgSQL, PL/SQL, T-SQL)
- [ ] Triggers: BEFORE, AFTER, INSTEAD OF (row-level vs statement-level)
- [ ] Cursors (and why to avoid them)
- [ ] User-defined types and domains
- [ ] Table inheritance and partitioning via SQL
- [ ] SQL standards: SQL-92, SQL:1999, SQL:2003, SQL:2011, SQL:2016, SQL:2023
- [ ] Full-text search in SQL (tsvector, tsquery in PostgreSQL)

### 1.5 Normalization & Schema Design
- [ ] 1NF — eliminate repeating groups
- [ ] 2NF — eliminate partial dependencies
- [ ] 3NF — eliminate transitive dependencies
- [ ] BCNF (Boyce-Codd Normal Form)
- [ ] 4NF — eliminate multivalued dependencies
- [ ] 5NF (PJNF) — eliminate join dependencies
- [ ] 6NF — for temporal databases
- [ ] Denormalization: when and why
- [ ] Star schema and snowflake schema
- [ ] Fact tables and dimension tables
- [ ] Slowly Changing Dimensions (SCD Types 1-6)
- [ ] Entity-Relationship (ER) modeling
- [ ] EER (Enhanced ER) diagrams
- [ ] Data modeling tools: dbdiagram.io, ERDPlus, Lucidchart

---

## Phase 2 — Database Internals & Storage Engine Mastery

### 2.1 Storage Engines
- [ ] Page-oriented storage (slotted pages)
- [ ] Row-oriented (NSM) vs column-oriented (DSM) storage
- [ ] B-tree family: B-tree, B+ tree, B* tree
- [ ] LSM trees (Log-Structured Merge Trees)
- [ ] Write-Ahead Log (WAL)
- [ ] Buffer pool / buffer manager
- [ ] Page replacement policies: LRU, Clock, LRU-K, 2Q, ARC
- [ ] Heap files, sorted files, hash files
- [ ] Overflow pages and TOAST (in PostgreSQL)
- [ ] Compaction strategies: size-tiered, leveled, FIFO
- [ ] Bloom filters for LSM optimization
- [ ] Copy-on-write B-trees (LMDB, BoltDB)
- [ ] Fractal trees (TokuDB)

### 2.2 Indexing — Deep Dive
- [ ] Primary index vs secondary index
- [ ] Clustered vs non-clustered indexes
- [ ] Dense vs sparse indexes
- [ ] Composite (multi-column) indexes and column ordering
- [ ] Covering indexes
- [ ] Partial indexes (conditional indexes)
- [ ] Expression indexes / functional indexes
- [ ] Hash indexes
- [ ] GiST (Generalized Search Tree)
- [ ] GIN (Generalized Inverted Index)
- [ ] SP-GiST (Space-Partitioned GiST)
- [ ] BRIN (Block Range Index)
- [ ] R-tree indexes (spatial)
- [ ] Bitmap indexes
- [ ] Full-text indexes (inverted indexes)
- [ ] Trie-based indexes
- [ ] Skip list indexes
- [ ] Adaptive Radix Tree (ART)
- [ ] Index-only scans
- [ ] Index maintenance: bloat, reindex, online reindexing
- [ ] When NOT to index

### 2.3 Query Processing & Optimization
- [ ] Query parsing → AST → logical plan → physical plan
- [ ] Logical optimization: predicate pushdown, projection pushdown, constant folding
- [ ] Join ordering and join enumeration
- [ ] Cost models: I/O cost, CPU cost, network cost
- [ ] Statistics: histograms, most common values, n-distinct, correlation
- [ ] Cardinality estimation and its pitfalls
- [ ] Adaptive query execution
- [ ] Parallel query execution
- [ ] Vectorized execution (column-at-a-time)
- [ ] JIT compilation of queries (e.g., PostgreSQL JIT)
- [ ] Query plan caching and plan invalidation
- [ ] Prepared statement optimization
- [ ] Optimizer hints (MySQL, Oracle)
- [ ] Plan regression and plan stability

### 2.4 Concurrency Control
- [ ] ACID properties — deep understanding
- [ ] Transaction isolation levels: READ UNCOMMITTED, READ COMMITTED, REPEATABLE READ, SERIALIZABLE
- [ ] Anomalies: dirty reads, non-repeatable reads, phantom reads, write skew, lost updates
- [ ] Lock-based concurrency: shared locks, exclusive locks, intention locks
- [ ] Two-Phase Locking (2PL): strict, rigorous
- [ ] Deadlock detection (wait-for graphs) and prevention
- [ ] Lock granularity: row, page, table, database
- [ ] Optimistic Concurrency Control (OCC)
- [ ] Multi-Version Concurrency Control (MVCC)
  - [ ] PostgreSQL: tuple versioning with xmin/xmax
  - [ ] MySQL/InnoDB: undo logs and read views
  - [ ] Oracle: undo tablespace and SCN
- [ ] Snapshot Isolation (SI) and Serializable Snapshot Isolation (SSI)
- [ ] Timestamp ordering
- [ ] Vacuum and garbage collection of old versions

### 2.5 Recovery & Durability
- [ ] Write-Ahead Logging (WAL) protocol
- [ ] ARIES recovery algorithm (Analysis, Redo, Undo)
- [ ] Checkpointing: fuzzy checkpoints, sharp checkpoints
- [ ] Log sequence numbers (LSN)
- [ ] Physiological logging
- [ ] Shadow paging
- [ ] Double-write buffer (InnoDB)
- [ ] Crash recovery process
- [ ] Point-in-time recovery (PITR)
- [ ] Group commit optimization
- [ ] fsync and durability guarantees
- [ ] Battery-backed write cache considerations

### 2.6 Memory Management
- [ ] Buffer pool architecture
- [ ] Shared buffers vs OS page cache
- [ ] Memory allocators in databases
- [ ] Work memory for sorts and hash tables
- [ ] Connection memory overhead
- [ ] Huge pages / large pages
- [ ] NUMA-aware memory allocation
- [ ] Memory-mapped I/O (mmap) databases

---

## Phase 3 — Master Specific Database Systems

### 3.1 PostgreSQL — Deep Mastery
- [ ] Architecture: postmaster, backend processes, shared memory
- [ ] System catalogs: pg_class, pg_attribute, pg_index, pg_stat_*
- [ ] MVCC implementation details
- [ ] VACUUM: regular, full, auto-vacuum tuning
- [ ] HOT updates (Heap-Only Tuples)
- [ ] TOAST mechanism
- [ ] Tablespaces
- [ ] Table partitioning: range, list, hash (declarative)
- [ ] Connection pooling: PgBouncer, Pgpool-II
- [ ] Extensions ecosystem: PostGIS, pg_trgm, pg_stat_statements, TimescaleDB, Citus, pgvector
- [ ] Logical replication and logical decoding
- [ ] Streaming replication (sync and async)
- [ ] pg_basebackup, pg_dump, pg_restore
- [ ] Configuration tuning: shared_buffers, work_mem, effective_cache_size, wal_buffers, max_connections, etc.
- [ ] PL/pgSQL programming
- [ ] Foreign Data Wrappers (FDW)
- [ ] Advisory locks
- [ ] Listen/Notify for pub/sub
- [ ] Row-level security (RLS)
- [ ] pg_stat_statements for query analysis
- [ ] Custom background workers

### 3.2 MySQL / MariaDB
- [ ] Architecture: connection layer, server layer, storage engine layer
- [ ] InnoDB internals: clustered index, secondary indexes, buffer pool, change buffer, adaptive hash index
- [ ] InnoDB locking: record locks, gap locks, next-key locks
- [ ] MyISAM vs InnoDB vs RocksDB (MyRocks)
- [ ] MySQL replication: binlog, GTID, semi-sync, group replication
- [ ] MySQL Router and InnoDB Cluster
- [ ] Query cache (deprecated) and ProxySQL caching
- [ ] Partitioning in MySQL
- [ ] Performance Schema and sys schema
- [ ] MySQL slow query log
- [ ] pt-query-digest and Percona Toolkit
- [ ] MySQL 8.0+ features: window functions, CTEs, JSON improvements
- [ ] MariaDB-specific: Aria, ColumnStore, Spider, Galera Cluster

### 3.3 Oracle Database
- [ ] Architecture: SGA, PGA, background processes (DBWR, LGWR, CKPT, SMON, PMON)
- [ ] Tablespaces, segments, extents, blocks
- [ ] Oracle RAC (Real Application Clusters)
- [ ] Data Guard (standby databases)
- [ ] ASM (Automatic Storage Management)
- [ ] Oracle Optimizer: hints, profiles, baselines
- [ ] PL/SQL programming
- [ ] Oracle partitioning options
- [ ] Flashback technology
- [ ] Exadata architecture

### 3.4 Microsoft SQL Server
- [ ] Architecture: TDS protocol, SQL OS, buffer pool, plan cache
- [ ] Storage: pages, extents, allocation units
- [ ] Columnstore indexes
- [ ] In-memory OLTP (Hekaton)
- [ ] Always On Availability Groups
- [ ] T-SQL programming
- [ ] Query Store
- [ ] Intelligent Query Processing
- [ ] PolyBase for external data
- [ ] Temporal tables

### 3.5 SQLite
- [ ] Architecture: single-file, serverless, zero-config
- [ ] B-tree implementation
- [ ] WAL mode vs rollback journal
- [ ] Concurrent readers, single writer
- [ ] Use cases: embedded, mobile, IoT, testing
- [ ] Virtual tables and R-tree module
- [ ] SQLite as an application file format
- [ ] Litestream for replication
- [ ] libSQL / Turso (distributed SQLite)

---

## Phase 4 — Distributed Systems & NoSQL

### 4.1 Distributed Systems Theory
- [ ] CAP theorem — deep understanding and critique
- [ ] PACELC theorem
- [ ] FLP impossibility result
- [ ] Consistency models: strong, sequential, causal, eventual, read-your-writes
- [ ] Linearizability vs serializability
- [ ] Consensus algorithms: Paxos, Multi-Paxos, Raft, Zab
- [ ] Vector clocks and Lamport timestamps
- [ ] Gossip protocols
- [ ] Consistent hashing
- [ ] Quorum-based replication (R + W > N)
- [ ] Two-Phase Commit (2PC) and Three-Phase Commit (3PC)
- [ ] Saga pattern for distributed transactions
- [ ] CRDTs (Conflict-free Replicated Data Types)
- [ ] Split-brain problem and fencing
- [ ] Failure detection: heartbeats, phi-accrual detector
- [ ] Shard rebalancing strategies

### 4.2 Distributed SQL / NewSQL
- [ ] CockroachDB: architecture, ranges, leaseholders, Raft consensus
- [ ] TiDB: TiKV + TiDB + PD architecture
- [ ] YugabyteDB: DocDB, tablet splitting, Raft
- [ ] Google Spanner: TrueTime, external consistency
- [ ] VoltDB: in-memory, partitioned, deterministic execution
- [ ] Calvin: deterministic database protocol
- [ ] Vitess: MySQL sharding middleware
- [ ] Citus: distributed PostgreSQL
- [ ] PlanetScale: Vitess-based platform
- [ ] Neon: serverless PostgreSQL with separation of storage and compute

### 4.3 Key-Value Stores
- [ ] Redis: data structures, persistence (RDB, AOF), clustering, Sentinel, Streams
- [ ] Redis internals: single-threaded event loop, SDS, ziplist, skiplist, dict
- [ ] Memcached: architecture, slab allocator, consistent hashing
- [ ] etcd: Raft consensus, watch, lease
- [ ] Amazon DynamoDB: partition keys, sort keys, GSI, LSI, DynamoDB Streams, single-table design
- [ ] Riak: masterless, vector clocks, CRDTs
- [ ] FoundationDB: ordered key-value, layer concept, simulation testing
- [ ] RocksDB: LSM internals, compaction, write stalls
- [ ] BadgerDB, Pebble, LevelDB

### 4.4 Document Databases
- [ ] MongoDB: BSON, replica sets, sharding, aggregation pipeline, change streams
- [ ] MongoDB internals: WiredTiger storage engine, oplog, chunk migration
- [ ] Schema design patterns for document DBs: embedding vs referencing, bucket pattern, outlier pattern
- [ ] CouchDB: multi-master replication, MapReduce views
- [ ] Couchbase: memory-first architecture, N1QL, XDCR
- [ ] Amazon DocumentDB
- [ ] FerretDB (MongoDB-compatible on PostgreSQL)

### 4.5 Wide-Column Stores
- [ ] Apache Cassandra: ring architecture, partitioners, replication strategy, compaction
- [ ] Cassandra data modeling: partition keys, clustering columns, denormalization-first
- [ ] CQL (Cassandra Query Language)
- [ ] ScyllaDB: shard-per-core architecture, seastar framework
- [ ] Apache HBase: RegionServers, WAL, MemStore, HFile, compaction
- [ ] Google Bigtable: tablet servers, SSTable, GFS

### 4.6 Graph Databases
- [ ] Property graph model vs RDF model
- [ ] Neo4j: Cypher query language, native graph storage, APOC library
- [ ] Graph algorithms: shortest path, PageRank, community detection, centrality
- [ ] Amazon Neptune
- [ ] ArangoDB (multi-model: document + graph + key-value)
- [ ] JanusGraph
- [ ] Apache TinkerPop / Gremlin query language
- [ ] SPARQL for RDF databases
- [ ] Graph database use cases: social networks, fraud detection, knowledge graphs, recommendation engines

### 4.7 Time-Series Databases
- [ ] TimescaleDB (PostgreSQL extension): hypertables, chunks, continuous aggregates
- [ ] InfluxDB: TSM storage engine, Flux query language
- [ ] Prometheus: pull-based metrics, PromQL
- [ ] QuestDB: column-oriented, zero-GC Java
- [ ] Apache IoTDB
- [ ] Amazon Timestream
- [ ] ClickHouse (also OLAP, but excellent for time-series)
- [ ] Data retention policies and downsampling

### 4.8 Search Engines (as Databases)
- [ ] Elasticsearch: inverted index, shards, replicas, analyzers, mappings
- [ ] Elasticsearch Query DSL, aggregations, full-text search, fuzzy matching
- [ ] ELK Stack (Elasticsearch + Logstash + Kibana)
- [ ] OpenSearch (fork of Elasticsearch)
- [ ] Apache Solr
- [ ] Meilisearch, Typesense (search-focused)
- [ ] Tantivy (Rust-based search library)
- [ ] Lucene internals: segments, merging, term dictionaries

### 4.9 Vector Databases
- [ ] What are embeddings and why vector search matters
- [ ] Similarity metrics: cosine, euclidean, dot product
- [ ] ANN algorithms: HNSW, IVF, PQ (product quantization), ScaNN
- [ ] pgvector (PostgreSQL extension)
- [ ] Pinecone
- [ ] Milvus
- [ ] Weaviate
- [ ] Qdrant
- [ ] Chroma
- [ ] Hybrid search: combining vector + keyword search

### 4.10 Message Queues & Streaming (as Databases)
- [ ] Apache Kafka: partitions, consumer groups, log compaction, exactly-once semantics
- [ ] Kafka internals: segments, indexing, zero-copy
- [ ] Apache Pulsar: multi-tenancy, geo-replication, tiered storage
- [ ] Amazon Kinesis
- [ ] NATS JetStream
- [ ] RabbitMQ: exchanges, queues, bindings
- [ ] Redpanda (Kafka-compatible, C++)
- [ ] Event sourcing and CQRS patterns with streaming

---

## Phase 5 — Production Engineering & Advanced Topics

### 5.1 Performance Tuning & Benchmarking
- [ ] Systematic query optimization methodology
- [ ] Identifying slow queries: slow query logs, pg_stat_statements, Performance Schema
- [ ] Index tuning wizard / advisor approaches
- [ ] Connection pool sizing (HikariCP, PgBouncer, ProxySQL)
- [ ] Benchmarking: sysbench, pgbench, YCSB, TPC-C, TPC-H, TPC-DS
- [ ] Load testing databases
- [ ] Profiling: perf, strace, eBPF for database analysis
- [ ] I/O profiling: iostat, blktrace
- [ ] Lock contention analysis
- [ ] Query plan regression detection
- [ ] Read replicas and query routing
- [ ] Caching strategies: application cache, query cache, materialized views, result cache
- [ ] Cache invalidation strategies
- [ ] Database proxy layers for performance

### 5.2 High Availability & Replication
- [ ] Single-leader replication
- [ ] Multi-leader replication
- [ ] Leaderless replication
- [ ] Synchronous vs asynchronous replication
- [ ] Semi-synchronous replication
- [ ] Replication lag and monitoring
- [ ] Failover: automatic vs manual
- [ ] Split-brain prevention
- [ ] Patroni for PostgreSQL HA
- [ ] Orchestrator for MySQL HA
- [ ] Keepalived / HAProxy for database HA
- [ ] Active-active vs active-passive setup
- [ ] Zero-downtime failover
- [ ] Read replicas: load balancing, lag handling

### 5.3 Sharding & Partitioning
- [ ] Horizontal partitioning (sharding) vs vertical partitioning
- [ ] Shard key selection strategies
- [ ] Hash-based sharding
- [ ] Range-based sharding
- [ ] Directory-based sharding
- [ ] Geo-based sharding
- [ ] Hot shard detection and remediation
- [ ] Cross-shard queries and transactions
- [ ] Resharding / rebalancing
- [ ] Table partitioning: range, list, hash, composite
- [ ] Partition pruning
- [ ] Vitess, Citus, ProxySQL for sharding

### 5.4 Backup & Disaster Recovery
- [ ] Backup types: full, incremental, differential
- [ ] Logical backups: pg_dump, mysqldump, mongodump
- [ ] Physical backups: pg_basebackup, Percona XtraBackup, RMAN
- [ ] Continuous archiving and WAL shipping
- [ ] Point-in-time recovery (PITR)
- [ ] Backup verification and restore testing
- [ ] RTO (Recovery Time Objective) and RPO (Recovery Point Objective)
- [ ] Disaster recovery planning
- [ ] Cross-region backup replication
- [ ] Backup encryption and security
- [ ] pgBackRest, Barman for PostgreSQL
- [ ] Percona XtraBackup for MySQL

### 5.5 Security
- [ ] Authentication: password, certificate, LDAP, Kerberos, SCRAM-SHA-256
- [ ] Authorization: GRANT, REVOKE, role-based access control
- [ ] Row-Level Security (RLS)
- [ ] Column-level encryption
- [ ] Transparent Data Encryption (TDE)
- [ ] Encryption at rest and in transit (TLS)
- [ ] SQL injection: understanding, prevention, parameterized queries
- [ ] Audit logging
- [ ] Database firewall
- [ ] Secrets management (Vault, AWS Secrets Manager)
- [ ] Principle of least privilege
- [ ] Connection security and network isolation
- [ ] Data masking and anonymization
- [ ] GDPR, HIPAA, SOC2 compliance considerations

### 5.6 Monitoring & Observability
- [ ] Key metrics: QPS, latency (p50/p95/p99), connections, cache hit ratio, replication lag, disk I/O, lock waits
- [ ] pg_stat_activity, pg_stat_user_tables, pg_stat_bgwriter
- [ ] Performance Schema (MySQL)
- [ ] Monitoring stacks: Prometheus + Grafana, Datadog, New Relic
- [ ] postgres_exporter, mysqld_exporter for Prometheus
- [ ] pgwatch2, PMM (Percona Monitoring and Management)
- [ ] Alerting strategies and thresholds
- [ ] Query log analysis
- [ ] Dead tuple ratio monitoring (PostgreSQL)
- [ ] Bloat detection and remediation
- [ ] Slow query detection and automated analysis
- [ ] Distributed tracing integration

### 5.7 Schema Migration & Evolution
- [ ] Migration tools: Flyway, Liquibase, Alembic, golang-migrate, Atlas, sqitch
- [ ] Zero-downtime migrations
- [ ] Expand-and-contract pattern
- [ ] Online DDL: pt-online-schema-change, gh-ost, pg_repack
- [ ] Adding columns, dropping columns safely
- [ ] Index creation: CONCURRENTLY (PostgreSQL), online DDL (MySQL)
- [ ] Foreign key addition without locking
- [ ] Data migration strategies
- [ ] Backward-compatible schema changes
- [ ] Version control for database schemas
- [ ] Blue-green deployments for databases

### 5.8 Data Warehousing & OLAP
- [ ] OLTP vs OLAP
- [ ] Data warehouse architecture: Kimball vs Inmon
- [ ] Star schema, snowflake schema, data vault
- [ ] ETL vs ELT
- [ ] Columnar storage benefits for analytics
- [ ] ClickHouse: MergeTree engine, materialized views, distributed tables
- [ ] Apache Druid: real-time ingestion, segments, rollup
- [ ] Apache Pinot: real-time analytics
- [ ] Amazon Redshift, Google BigQuery, Snowflake
- [ ] Apache Hive, Presto/Trino
- [ ] DuckDB (embedded OLAP)
- [ ] Apache Parquet, ORC file formats
- [ ] Apache Iceberg, Delta Lake, Apache Hudi (table formats)
- [ ] Materialized views for pre-aggregation
- [ ] Approximate query processing: HyperLogLog, Count-Min Sketch, t-digest

### 5.9 Data Pipeline & Integration
- [ ] CDC (Change Data Capture): Debezium, Maxwell, logical decoding
- [ ] Apache Kafka Connect
- [ ] Apache Flink for stream processing
- [ ] Apache Spark for batch processing
- [ ] dbt (data build tool) for transformation
- [ ] Airbyte, Fivetran for data ingestion
- [ ] Apache Airflow for orchestration
- [ ] Data lake architecture
- [ ] Lakehouse architecture
- [ ] Data mesh concepts
- [ ] Schema registry (Confluent, Apicurio)
- [ ] Data quality and validation frameworks

### 5.10 Cloud Database Services
- [ ] AWS: RDS, Aurora, DynamoDB, Redshift, ElastiCache, Neptune, Timestream, DocumentDB, Keyspaces, MemoryDB
- [ ] GCP: Cloud SQL, Cloud Spanner, Bigtable, BigQuery, Firestore, Memorystore, AlloyDB
- [ ] Azure: SQL Database, Cosmos DB, Database for PostgreSQL/MySQL, Synapse Analytics, Cache for Redis
- [ ] Serverless databases: Aurora Serverless, PlanetScale, Neon, Turso, D1 (Cloudflare)
- [ ] Database-as-a-Service trade-offs
- [ ] Multi-cloud database strategies
- [ ] Cost optimization for cloud databases

### 5.11 Database DevOps & Automation
- [ ] Infrastructure as Code: Terraform, Pulumi, CloudFormation for databases
- [ ] Ansible/Chef/Puppet for database configuration management
- [ ] Database CI/CD pipelines
- [ ] Automated testing with test databases
- [ ] Database containerization: Docker, Kubernetes operators
- [ ] Kubernetes operators: CloudNativePG, Zalando Postgres Operator, Percona operators
- [ ] GitOps for database configuration
- [ ] Chaos engineering for databases
- [ ] Automated failover testing
- [ ] Synthetic monitoring

---

## Phase 6 — Becoming God-Tier

### 6.1 Build Your Own Database
- [ ] Implement a simple key-value store from scratch
- [ ] Build a B+ tree on disk
- [ ] Implement a buffer pool manager
- [ ] Build a WAL-based recovery system
- [ ] Implement a simple SQL parser
- [ ] Build a query executor (volcano model)
- [ ] Implement MVCC
- [ ] Build a simple query optimizer
- [ ] Resources:
  - [ ] CMU 15-445/645 (Andy Pavlo) — Database Systems
  - [ ] CMU 15-721 — Advanced Database Systems
  - [ ] "Database Design and Implementation" by Edward Sciore
  - [ ] Let's Build a Simple Database (cstack.github.io)
  - [ ] Toydb, mini-lsm, bustub projects

### 6.2 Read the Source Code
- [ ] PostgreSQL source code (C)
- [ ] SQLite source code (C) — beautifully readable
- [ ] RocksDB source code (C++)
- [ ] DuckDB source code (C++)
- [ ] TiKV source code (Rust)
- [ ] etcd source code (Go)
- [ ] CockroachDB source code (Go)
- [ ] FoundationDB source code (C++)

### 6.3 Research Papers — Must Read
- [ ] "A Relational Model of Data for Large Shared Data Banks" — Codd (1970)
- [ ] "Access Path Selection in a Relational DBMS" — Selinger et al. (1979)
- [ ] "ARIES: A Transaction Recovery Method" — Mohan et al. (1992)
- [ ] "The Design of POSTGRES" — Stonebraker & Rowe (1986)
- [ ] "Architecture of a Database System" — Hellerstein, Stonebraker, Hamilton (2007)
- [ ] "Spanner: Google's Globally-Distributed Database" — Corbett et al. (2012)
- [ ] "Dynamo: Amazon's Highly Available Key-value Store" — DeCandia et al. (2007)
- [ ] "Bigtable: A Distributed Storage System for Structured Data" — Chang et al. (2006)
- [ ] "The Log-Structured Merge-Tree (LSM-Tree)" — O'Neil et al. (1996)
- [ ] "Calvin: Fast Distributed Transactions for Partitioned Database Systems" (2012)
- [ ] "Socrates: The New SQL Server in the Cloud" — Antonopoulos et al. (2019)
- [ ] "Amazon Aurora: Design Considerations for High Throughput Cloud-Native Relational Databases" (2017)
- [ ] "CockroachDB: The Resilient Geo-Distributed SQL Database" (2020)
- [ ] "Napa: Powering Scalable Data Warehousing with Robust Query Performance at Google" (2021)
- [ ] "The Snowflake Elastic Data Warehouse" (2016)
- [ ] "Looking Back at Postgres" — Stonebraker (2019)
- [ ] "What's Really New with NewSQL?" — Pavlo & Aslett (2016)
- [ ] "An Empirical Evaluation of In-Memory Multi-Version Concurrency Control" — Wu et al. (2017)

### 6.4 Books — The Canon
- [ ] **"Database Internals"** — Alex Petrov ★★★★★
- [ ] **"Designing Data-Intensive Applications" (DDIA)** — Martin Kleppmann ★★★★★
- [ ] **"Fundamentals of Database Systems"** — Elmasri & Navathe
- [ ] **"Database System Concepts"** — Silberschatz, Korth, Sudarshan
- [ ] **"Transaction Processing: Concepts and Techniques"** — Gray & Reuter
- [ ] **"The Art of PostgreSQL"** — Dimitri Fontaine
- [ ] **"PostgreSQL 14 Internals"** — Egor Rogov
- [ ] **"High Performance MySQL"** — Schwartz, Zaitsev, Tkachenko
- [ ] **"Redis in Action"** — Josiah Carlson
- [ ] **"MongoDB: The Definitive Guide"** — Shannon Bradshaw
- [ ] **"Streaming Systems"** — Akidau, Chernyak, Lax
- [ ] **"The Data Warehouse Toolkit"** — Ralph Kimball
- [ ] **"SQL Performance Explained"** — Markus Winand (use-the-index-luke.com)
- [ ] **"Understanding Distributed Systems"** — Roberto Vitillo

### 6.5 Courses & Lectures
- [ ] CMU 15-445/645 — Intro to Database Systems (Andy Pavlo, YouTube)
- [ ] CMU 15-721 — Advanced Database Systems (Andy Pavlo, YouTube)
- [ ] MIT 6.824 — Distributed Systems (Robert Morris)
- [ ] Stanford CS245 — Principles of Data-Intensive Systems
- [ ] UC Berkeley CS186 — Introduction to Database Systems
- [ ] use-the-index-luke.com — SQL indexing and tuning (free)

### 6.6 Community & Practice
- [ ] Contribute to open-source database projects
- [ ] Read database blogs: Brandur, Use The Index Luke, Percona Blog, PgAnalyze, CockroachDB Blog, Jepsen
- [ ] Follow Jepsen.io for correctness testing of distributed databases
- [ ] Attend database conferences: SIGMOD, VLDB, ICDE, PGConf, Percona Live
- [ ] Practice on: LeetCode SQL, HackerRank SQL, StrataScratch, SQLZoo
- [ ] Build projects:
  - [ ] Design Twitter's database schema
  - [ ] Build a URL shortener with analytics
  - [ ] Design an e-commerce data model
  - [ ] Build a real-time analytics dashboard
  - [ ] Implement a distributed key-value store
  - [ ] Build a CDC pipeline with Debezium + Kafka
  - [ ] Design a multi-tenant SaaS database
- [ ] Write blog posts explaining database concepts
- [ ] Answer database questions on Stack Overflow
- [ ] Give talks at meetups

---

## Learning Order — Suggested Path

```
Phase 0 (2-4 weeks)     → Foundations
Phase 1 (2-3 months)    → SQL + Relational mastery
Phase 2 (3-4 months)    → Internals deep dive
Phase 3 (3-4 months)    → Master 2-3 specific systems deeply
Phase 4 (3-6 months)    → Distributed systems + NoSQL breadth
Phase 5 (ongoing)       → Production engineering skills
Phase 6 (lifetime)      → Build, read source, read papers
```

> **Key Principle:** Depth over breadth. Master PostgreSQL deeply before touching 10 databases superficially. Understand B-trees before memorizing CREATE INDEX syntax. Read DDIA before designing distributed systems.

---

*"The database is the center of gravity of every serious application."*
*Become the person who truly understands what happens between the query and the disk.*
