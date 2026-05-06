# 4.10 — Message Queues & Streaming (as Databases)

> Kafka is not just a message queue — it's a distributed commit log.  
> It stores data durably, replays history, and enables event-driven architectures.  
> Modern streaming platforms blur the line between messaging and databases.

---

## 1. Apache Kafka — The Distributed Commit Log

### Architecture

```
┌────────────────────────────────────────────────────────┐
│                    Kafka Cluster                         │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ Broker 1  │  │ Broker 2  │  │ Broker 3  │              │
│  │           │  │           │  │           │              │
│  │ Topic A   │  │ Topic A   │  │ Topic A   │              │
│  │ Part 0 (L)│  │ Part 1 (L)│  │ Part 2 (L)│              │
│  │ Part 1 (F)│  │ Part 2 (F)│  │ Part 0 (F)│              │
│  │           │  │           │  │           │              │
│  └──────────┘  └──────────┘  └──────────┘              │
│                                                          │
│  L = Leader replica, F = Follower replica                │
│                                                          │
│  Metadata: KRaft (Kafka Raft) or ZooKeeper (legacy)      │
└────────────────────────────────────────────────────────┘

Topic: a named channel for events (like a table)
Partition: ordered, immutable sequence of records (like a log file)
Offset: position of a record within a partition (monotonically increasing)
Replication factor: copies per partition (typically 3)

Partition 0:
┌────┬────┬────┬────┬────┬────┬────┬────┐
│ 0  │ 1  │ 2  │ 3  │ 4  │ 5  │ 6  │ 7  │  ← offsets
└────┴────┴────┴────┴────┴────┴────┴────┘
  write →→→→→→→→→→→→→→→→→→→→→→→→→→→→→ (append-only)

Records within a partition are:
  - Ordered (offset = sequence number)
  - Immutable (once written, never modified)
  - Durable (replicated, flushed to disk)
  - Retained for a configurable period (7 days default, or forever)
```

### Producers and Consumers

```
Producer:
  - Sends records to a topic
  - Chooses partition: by key hash, round-robin, or custom
  - Batching: accumulates records and sends in batches (throughput vs latency)
  
  Key-based partitioning:
    All records with the same key → same partition → ordered processing
    hash(key) % num_partitions = target partition
    Example: key = user_id → all events for a user in order

  Acknowledgment levels:
    acks=0:   fire and forget (fastest, data loss possible)
    acks=1:   leader acknowledges (fast, loss if leader crashes before replication)
    acks=all: all in-sync replicas acknowledge (safest, slower)

Consumer Group:
  Multiple consumers sharing the work of reading a topic.
  Each partition is assigned to exactly ONE consumer in the group.
  
  Topic with 6 partitions, consumer group with 3 consumers:
    Consumer A: partitions 0, 1
    Consumer B: partitions 2, 3
    Consumer C: partitions 4, 5
  
  If Consumer B dies → rebalance:
    Consumer A: partitions 0, 1, 2
    Consumer C: partitions 3, 4, 5
  
  More consumers than partitions = some consumers idle.
  → Partitions = max parallelism of a consumer group.

  Offset management:
    Consumers track their position (offset) per partition.
    Committed offsets stored in __consumer_offsets topic.
    On restart: resume from last committed offset.
    
    At-most-once: commit before processing → may lose messages
    At-least-once: commit after processing → may duplicate messages
    Exactly-once: idempotent producer + transactional consumer (EOS)
```

### Kafka Internals

```
Storage on disk:
  Each partition = directory of segment files
  
  /data/topic-orders/partition-0/
    00000000000000000000.log      ← segment file (records)
    00000000000000000000.index    ← offset index (offset → file position)
    00000000000000000000.timeindex ← timestamp index
    00000000000054321000.log      ← next segment (starts at offset 54321000)
  
  Segments:
    - Active segment: currently being written to (append-only)
    - Closed segments: immutable, eligible for cleanup
    - Segment size: default 1 GB or 7 days

  Zero-copy:
    Kafka uses sendfile() system call to transfer data
    directly from disk → network socket, bypassing user space.
    → Extremely efficient for consumers reading sequential data.

Log compaction:
  Instead of deleting old segments by time:
  Keep LATEST value for each key, delete older values.
  
  Before compaction:
    offset 0: key=A, value=1
    offset 1: key=B, value=2
    offset 2: key=A, value=3  ← newer value for A
    offset 3: key=C, value=4
    offset 4: key=B, value=5  ← newer value for B
  
  After compaction:
    offset 2: key=A, value=3  ← latest A
    offset 3: key=C, value=4  ← latest C
    offset 4: key=B, value=5  ← latest B
  
  Use cases: change data capture, state snapshots, configuration
  → Kafka as a DATABASE: compacted topic = key-value store!

KRaft (Kafka Raft):
  Kafka 3.3+: replaces ZooKeeper with built-in Raft consensus.
  Metadata stored in internal topic (__cluster_metadata).
  Controller quorum manages cluster state.
  → Simpler operations (no separate ZooKeeper cluster).
```

### Exactly-Once Semantics (EOS)

```
Kafka 0.11+ supports exactly-once processing:

Idempotent producer:
  enable.idempotence = true
  Producer assigns sequence numbers to records.
  Broker deduplicates based on (producer_id, sequence_number).
  → No duplicates from producer retries.

Transactional producer:
  Atomic writes across multiple partitions/topics.
  
  producer.initTransactions();
  producer.beginTransaction();
  producer.send(record1);  // to topic A
  producer.send(record2);  // to topic B
  producer.commitTransaction();  // atomic: both or neither
  
  Used for: consume-transform-produce pipelines
  (read from topic A, process, write to topic B — atomically)

Consumer: read_committed
  isolation.level = "read_committed"
  Consumer only sees committed transactional records.
  → End-to-end exactly-once: idempotent producer + transactions + read_committed consumer.
```

---

## 2. Apache Pulsar

```
Pulsar: next-generation messaging/streaming platform.

Key differences from Kafka:
  ┌─────────────────────────────────────────────────┐
  │             Pulsar Architecture                   │
  │                                                   │
  │  Brokers (stateless) ←→ BookKeeper (storage)     │
  │                                                   │
  │  Broker = serving layer (routing, protocol)       │
  │  BookKeeper = storage layer (distributed log)     │
  │                                                   │
  │  Separation of compute and storage!               │
  └─────────────────────────────────────────────────┘

  Kafka: brokers ARE storage (must rebalance data when adding brokers)
  Pulsar: brokers are stateless (add/remove without data movement)

Pulsar advantages over Kafka:
  ✓ Multi-tenancy (namespaces, resource isolation)
  ✓ Geo-replication built-in (cross-datacenter)
  ✓ Tiered storage (offload old data to S3 automatically)
  ✓ Topic-level subscriptions (not just partition-level)
  ✓ Shared subscriptions (multiple consumers on one partition)
  ✓ Delayed messages, dead letter queues built-in

Pulsar disadvantages:
  ✗ More complex architecture (3 components: broker + BookKeeper + ZK)
  ✗ Smaller ecosystem and community than Kafka
  ✗ Higher operational complexity
```

---

## 3. Redpanda — Kafka-Compatible, C++

```
Redpanda: Kafka API compatible, rewritten in C++.
No JVM, no ZooKeeper, no GC pauses.

Architecture:
  - Single binary (no separate processes)
  - Raft consensus per partition (like CockroachDB)
  - Thread-per-core (Seastar framework — same as ScyllaDB)
  - Direct I/O (bypass page cache, self-managed)

Advantages:
  ✓ Drop-in Kafka replacement (same protocol, same client libraries)
  ✓ 10x lower tail latency (no GC)
  ✓ Simpler operations (1 binary vs Kafka + ZooKeeper/KRaft)
  ✓ Lower resource usage per message

When to choose:
  - New deployment + want Kafka API → consider Redpanda
  - Existing Kafka + mature ecosystem → stay on Kafka
  - Need lowest possible latency → Redpanda
```

---

## 4. Event Sourcing and CQRS

```
Event Sourcing:
  Instead of storing CURRENT STATE, store the SEQUENCE OF EVENTS.
  
  Traditional (state):
    Account { id: 1, balance: 150 }
  
  Event sourced:
    AccountCreated { id: 1, balance: 0 }
    MoneyDeposited { id: 1, amount: 200 }
    MoneyWithdrawn { id: 1, amount: 50 }
    → Current state = replay events: 0 + 200 - 50 = 150
  
  Benefits:
    ✓ Complete audit trail (every change recorded)
    ✓ Time travel (rebuild state at any point)
    ✓ Event replay (reprocess events with new logic)
    ✓ Natural fit for Kafka (topic = event log)
  
  Challenges:
    ✗ Event schema evolution (old events must remain readable)
    ✗ Eventual consistency between command and query sides
    ✗ Complexity of rebuilding snapshots

CQRS (Command Query Responsibility Segregation):
  SEPARATE the write model from the read model.
  
  Command side (writes):
    → Events → Kafka → Event store
  
  Query side (reads):
    → Kafka consumer → Materialized view in PostgreSQL/Elasticsearch/Redis
    → Optimized for specific read patterns
  
  ┌──────────┐  events  ┌───────┐  consume  ┌──────────────┐
  │ Commands  │ ──────→ │ Kafka │ ────────→ │ Read DB       │
  │ (write)   │          │       │           │ (PostgreSQL,  │
  └──────────┘          │       │           │  Elasticsearch│
                         │       │           │  Redis)       │
                         └───────┘           └──────────────┘
  
  Why:
    Write model: append events (fast, simple)
    Read model: denormalized, query-optimized (different shape per use case)
    Multiple read models from same events (one for search, one for analytics)
```

---

## 5. Kafka as a Database?

```
Kafka CAN function as a database in specific scenarios:

Kafka IS a database when:
  ✓ Compacted topics = key-value store (latest state per key)
  ✓ Durable, replicated, ordered log
  ✓ Kafka Streams KTables = materialized queryable state
  ✓ ksqlDB: SQL on streams + materialized views
  ✓ Infinite retention = permanent storage

Kafka is NOT a good database for:
  ✗ Point queries by key (no index, scan entire partition)
  ✗ Complex queries (JOIN, aggregate across topics = complex)
  ✗ Random access reads (optimized for sequential consumption)
  ✗ Updates (can only append; compaction is eventual)
  ✗ Transactions across topics (limited to consume-produce pattern)

ksqlDB (Confluent):
  -- Create a stream from a Kafka topic:
  CREATE STREAM orders (
    order_id VARCHAR KEY, customer_id VARCHAR, amount DOUBLE
  ) WITH (KAFKA_TOPIC='orders', VALUE_FORMAT='JSON');

  -- Materialized table (continuously updated):
  CREATE TABLE customer_totals AS
  SELECT customer_id, SUM(amount) AS total, COUNT(*) AS order_count
  FROM orders
  GROUP BY customer_id;

  -- Query the materialized state:
  SELECT * FROM customer_totals WHERE customer_id = 'user-123';
  -- Pull query: point lookup on materialized state
  -- Push query: continuous stream of updates

The real pattern:
  Kafka = durable event log (source of truth)
  Derived stores = READ-OPTIMIZED projections (PostgreSQL, Redis, Elasticsearch)
  Event sourcing + CQRS naturally leads here.
```

---

## 6. Comparison Table

```
System       Model           Ordering    Persistence  Throughput    Use Case
──────────── ─────────────── ─────────── ──────────── ────────────── ──────────────
Kafka        Distributed log Per-partition Durable     1M+ msg/sec   Event streaming, CDC
Pulsar       Distributed log Per-partition Durable     1M+ msg/sec   Multi-tenant streaming
Redpanda     Distributed log Per-partition Durable     1M+ msg/sec   Low-latency Kafka
RabbitMQ     Message broker  Per-queue    Durable      50K msg/sec   Task queues, RPC
NATS         Pub/sub         Per-subject  Optional     10M+ msg/sec  Lightweight messaging
Redis Streams Append log     Per-stream   Optional     500K+ msg/sec  Simple event log

When to use what:
  Kafka/Pulsar/Redpanda: event sourcing, CDC, stream processing, analytics pipeline
  RabbitMQ: task distribution, request-reply, routing patterns
  NATS: microservice communication, IoT, edge
  Redis Streams: lightweight event log when you already have Redis
```

---

## 7. RabbitMQ — The Message Broker

```
RabbitMQ: traditional message broker (AMQP protocol).

Architecture:
  Producer → Exchange → Binding → Queue → Consumer

Exchange types:
  Direct:  route by exact routing key match
  Topic:   route by pattern matching (*.error, order.#)
  Fanout:  broadcast to all bound queues
  Headers: route by message header attributes

Key differences from Kafka:
  RabbitMQ: messages are deleted AFTER consumer acknowledges
  Kafka: messages are retained regardless of consumption

  RabbitMQ: smart broker, dumb consumers (broker pushes to consumers)
  Kafka: dumb broker, smart consumers (consumers pull and track offset)

  RabbitMQ: message-level routing (fan-out, topic, headers)
  Kafka: partition-level distribution only

RabbitMQ Streams (3.9+):
  Append-only log (like Kafka topics).
  Multiple consumers can read from the same stream independently.
  → RabbitMQ moving toward Kafka's stream model for specific use cases.
```

---

## Key Takeaways

1. **Kafka is an append-only distributed commit log.** Records are immutable, ordered per partition, and retained for a configurable period. It's more "database" than "queue."

2. **Partition count = max consumer parallelism.** Design partitions for your throughput needs. Key-based partitioning guarantees per-key ordering.

3. **Exactly-once semantics** is achievable with idempotent producers + transactions + read_committed consumers. It's now production-ready.

4. **Log compaction turns Kafka into a key-value store.** Retained latest value per key = materialized state in a topic.

5. **CQRS + Event Sourcing + Kafka** is a powerful architecture: write events to Kafka, project to read-optimized stores (PostgreSQL, Elasticsearch, Redis).

6. **Kafka vs RabbitMQ**: Kafka for event streaming and replay. RabbitMQ for task distribution and routing. Different tools for different problems.

7. **Redpanda is a compelling Kafka alternative** — same API, lower latency, simpler operations. Consider it for new deployments.

---

← Back to [01-distributed-systems-theory.md](01-distributed-systems-theory.md)
