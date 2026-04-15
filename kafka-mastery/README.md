# Kafka Mastery — God-Level Apache Kafka

> From "I know what a topic is" to "I can design systems that process millions
> of events/sec with exactly-once semantics across data centers."
>
> Every file is heavily commented, teaches internals used at LinkedIn/Uber/Netflix
> scale, and connects theory to production reality.

## Philosophy

**You don't master Kafka by memorizing configs.**
You master it by understanding:
1. **WHY** the log-structured storage was designed this way
2. **HOW** the replication protocol guarantees durability under failure
3. **WHEN** exactly-once semantics actually work (and when they don't)
4. **WHERE** the bottlenecks live at 1M+ msg/sec
5. **WHAT** happens inside the broker when your producer calls `Send()`

## Prerequisites

- Solid Go fundamentals (see ../golang-mastery)
- Docker & Docker Compose (for running Kafka locally)
- Basic understanding of distributed systems concepts

## Quick Start

```bash
cd kafka-mastery

# Start a 3-broker Kafka cluster with KRaft (no ZooKeeper):
docker compose up -d

# Run any lesson:
go run ./01-architecture/01_kafka_internals.go

# Run benchmarks:
go test -bench=. -benchmem ./11-performance/

# Tear down:
docker compose down -v
```

---

## Curriculum

### 01 — Architecture & Internals
**Goal:** Understand Kafka's brain — how brokers, controllers, and the log work.

| File | Topic |
|------|-------|
| `01_kafka_internals.go` | Broker anatomy, controller election, metadata flow |

### 02 — The Log: Kafka's Storage Engine
**Goal:** Master the append-only log — segments, indexes, compaction, retention.

| File | Topic |
|------|-------|
| `01_log_storage.go` | Segments, indexes, mmap, zero-copy, page cache |
| `02_compaction.go` | Log compaction internals, tombstones, min.cleanable.dirty.ratio |

### 03 — Producer Internals
**Goal:** Understand every byte from `Send()` to broker ACK.

| File | Topic |
|------|-------|
| `01_producer_internals.go` | Batching, linger.ms, compression, sticky partitioner, idempotent producer |

### 04 — Consumer Internals
**Goal:** Master the consumer group protocol, rebalancing, and offset management.

| File | Topic |
|------|-------|
| `01_consumer_internals.go` | Group coordinator, rebalance protocols, offset commit strategies |

### 05 — Partitioning & Ordering Guarantees
**Goal:** Design partition strategies for any scale.

| File | Topic |
|------|-------|
| `01_partitioning.go` | Key-based, custom partitioners, ordering guarantees, hot partitions |

### 06 — Replication Deep Dive
**Goal:** Understand ISR, leader election, and durability guarantees at the protocol level.

| File | Topic |
|------|-------|
| `01_replication.go` | ISR, HW, LEO, leader epoch, unclean leader election, min.insync.replicas |

### 07 — Exactly-Once Semantics
**Goal:** Implement true exactly-once processing — idempotent producers, transactions, EOS.

| File | Topic |
|------|-------|
| `01_exactly_once.go` | Idempotent producer, transactional API, consume-transform-produce |

### 08 — Schema Evolution & Data Contracts
**Goal:** Never break consumers with schema changes.

| File | Topic |
|------|-------|
| `01_schema_evolution.go` | Avro/Protobuf/JSON Schema, compatibility modes, Schema Registry |

### 09 — Stream Processing
**Goal:** Build real-time processing pipelines with Kafka Streams concepts.

| File | Topic |
|------|-------|
| `01_stream_processing.go` | Stateless/stateful processing, windowing, joins, changelog topics |

### 10 — Designing for Infinite Scale
**Goal:** Architect systems that scale horizontally without limits.

| File | Topic |
|------|-------|
| `01_infinite_scale.go` | Topic design, partition scaling, consumer scaling, backpressure |
| `02_event_sourcing.go` | Event sourcing, CQRS, saga patterns with Kafka |

### 11 — Performance & Tuning
**Goal:** Squeeze every byte/sec out of your Kafka cluster.

| File | Topic |
|------|-------|
| `01_performance.go` | Benchmarking, bottleneck analysis, OS tuning, JVM tuning |

### 12 — Production Operations & War Stories
**Goal:** Operate Kafka in production without losing sleep.

| File | Topic |
|------|-------|
| `01_production_ops.go` | Monitoring, alerting, capacity planning, incident response |

### 13 — Multi-Datacenter & Disaster Recovery
**Goal:** Run Kafka across regions with RPO/RTO guarantees.

| File | Topic |
|------|-------|
| `01_multi_dc.go` | MirrorMaker 2, active-active, active-passive, offset translation |

### 14 — KRaft & The Future
**Goal:** Understand Kafka without ZooKeeper and what's coming next.

| File | Topic |
|------|-------|
| `01_kraft.go` | KRaft consensus, metadata topics, migration from ZooKeeper |

---

## The Mental Model

```
┌─────────────────────────────────────────────────────────────────────┐
│                    KAFKA = DISTRIBUTED COMMIT LOG                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Producer ──► Broker Cluster ──► Consumer Group                     │
│     │              │                   │                            │
│     │         ┌────┴────┐              │                            │
│     │         │ Topic   │              │                            │
│     ▼         │┌───────┐│              ▼                            │
│  Batching     ││Part-0 ││◄── Leader    Fetch + Poll Loop            │
│  Compression  ││Part-1 ││◄── Follower  Offset Management            │
│  Partitioning ││Part-2 ││◄── Follower  Rebalancing                  │
│               │└───────┘│                                           │
│               └─────────┘                                           │
│                                                                     │
│  KEY INSIGHT: Everything is an append-only log.                     │
│  Kafka doesn't "deliver" messages — consumers PULL from a position  │
│  in the log. This is what makes it infinitely scalable.             │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## God-Level Understanding Checklist

- [ ] Can explain what happens at the byte level when a producer sends a message
- [ ] Can draw the ISR protocol and explain leader epoch fencing
- [ ] Can design a system handling 1M+ events/sec with ordering guarantees
- [ ] Can implement exactly-once consume-transform-produce pipelines
- [ ] Can debug consumer lag and identify whether it's producer, broker, or consumer
- [ ] Can design multi-DC Kafka with < 1 minute RPO
- [ ] Can explain why Kafka is fast (zero-copy, page cache, sequential I/O, batching)
- [ ] Can design partition strategies that avoid hot spots at any scale
- [ ] Can implement schema evolution without breaking consumers
- [ ] Can tune a Kafka cluster for throughput vs latency tradeoffs
