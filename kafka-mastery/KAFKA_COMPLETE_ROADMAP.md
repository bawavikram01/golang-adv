# Kafka Complete Roadmap — From Zero to God Level

## Phase 1: Foundations (Week 1-2)
> "Understand what Kafka IS before you use it."

- [ ] Read Lesson 01: Architecture & Internals
  - Understand: Kafka = distributed commit log, NOT a message queue
  - Understand: broker anatomy (network threads, request handlers, purgatory)
  - Understand: controller role and why single-brain works
  - Can explain: what happens at the wire level during produce/consume

- [ ] Read Lesson 02: The Log (Storage Engine)
  - Understand: segments, indexes, mmap, sparse index lookup
  - Understand: page cache strategy and why Kafka doesn't manage its own cache
  - Understand: zero-copy (sendfile) and why it matters
  - Understand: retention (delete, compact, compact+delete)
  - Read Lesson 02b: Log compaction internals, tombstones, use cases

## Phase 2: Producer & Consumer Mastery (Week 2-3)
> "Every byte, every millisecond, every retry."

- [ ] Read Lesson 03: Producer Internals
  - Understand: the full pipeline (serialize → partition → batch → send)
  - Understand: batching (batch.size, linger.ms) and their interaction
  - Understand: compression (zstd > lz4 > snappy >> gzip)
  - Understand: idempotent producer (PID, sequence numbers, dedup)
  - Understand: sticky partitioner and why round-robin was replaced
  - Understand: retry semantics (delivery.timeout.ms is the real deadline)
  - Understand: memory management (buffer.memory, backpressure)

- [ ] Read Lesson 04: Consumer Internals
  - Understand: consumer group protocol (join, sync, heartbeat, leave)
  - Understand: rebalancing (eager vs cooperative, static membership)
  - Understand: offset management (auto-commit pitfalls, manual commit patterns)
  - Understand: fetch internals (min.bytes, max.wait.ms, max.poll.records)
  - Understand: the poll loop and max.poll.interval.ms
  - Understand: consumer lag (measurement, causes, death spiral)

## Phase 3: Distribution Design (Week 3-4)
> "Partitions and replication are the heart of Kafka's scalability."

- [ ] Read Lesson 05: Partitioning & Ordering
  - Understand: partition count formula and diminishing returns
  - Understand: key design (high cardinality, entity-based)
  - Understand: ordering guarantees (per-partition only)
  - Understand: hot partitions (detection and fix)
  - Understand: partition expansion (and why you can't shrink)

- [ ] Read Lesson 06: Replication Deep Dive
  - Can draw: ISR protocol with HW and LEO on a whiteboard
  - Understand: leader epoch and why it prevents data divergence
  - Understand: unclean leader election tradeoff
  - Know by heart: RF=3, min.ISR=2, acks=all (the golden config)
  - Understand: replication tuning and monitoring

## Phase 4: Advanced Guarantees (Week 4-5)
> "Exactly-once is not magic — it's engineering."

- [ ] Read Lesson 07: Exactly-Once Semantics
  - Understand: at-most/at-least/exactly-once delivery semantics
  - Understand: Kafka transactions (beginTransaction → commitTransaction)
  - Understand: consume-transform-produce pattern
  - Understand: transaction internals (coordinator, state machine, fencing)
  - Know: EOS limitations (external systems, cross-cluster, performance)

- [ ] Read Lesson 08: Schema Evolution
  - Understand: why raw JSON fails at scale
  - Understand: Avro vs Protobuf vs JSON Schema
  - Understand: Schema Registry (wire format, caching, subjects)
  - Understand: compatibility modes (BACKWARD, FORWARD, FULL, TRANSITIVE)
  - Can apply: evolution rules for safe schema changes

## Phase 5: Processing & Architecture (Week 5-6)
> "Kafka as the central nervous system."

- [ ] Read Lesson 09: Stream Processing
  - Understand: stream-table duality
  - Understand: stateless vs stateful processing
  - Understand: windowing (tumbling, hopping, sliding, session)
  - Understand: joins (stream-stream, stream-table, co-partitioning requirement)
  - Understand: state stores (RocksDB + changelog)
  - Know: when to use Kafka Streams vs Flink vs custom consumer

- [ ] Read Lesson 10a: Designing for Infinite Scale
  - Understand: topic design patterns (per-event-type, per-entity, tiered)
  - Understand: consumer scaling beyond partition count
  - Understand: backpressure strategies
  - Understand: multi-tenancy (quotas, ACLs)
  - Know: real-world architectures (LinkedIn, Uber, Netflix)
  - Know: scale anti-patterns

- [ ] Read Lesson 10b: Event Sourcing & CQRS
  - Understand: event sourcing (store events, derive state)
  - Understand: CQRS (separate read/write models)
  - Understand: saga pattern (choreography vs orchestration)
  - Understand: outbox pattern (atomic DB + Kafka via CDC)

## Phase 6: Operations & Production (Week 6-7)
> "Running Kafka without losing sleep."

- [ ] Read Lesson 11: Performance & Tuning
  - Can run: proper benchmarks with kafka-producer-perf-test
  - Can identify: bottleneck type (disk, network, CPU, memory)
  - Know: broker tuning cheat sheet
  - Know: OS tuning (swappiness, dirty ratios, file descriptors)
  - Know: JVM tuning (6GB heap, G1GC, MaxGCPauseMillis=20)
  - Have: production readiness checklist completed

- [ ] Read Lesson 12: Production Operations
  - Know: critical metrics (3 tiers: red alert, warning, capacity)
  - Have: alerting rules configured
  - Know: capacity planning formula
  - Can do: zero-downtime rolling upgrade
  - Can handle: common incidents (lag spiral, disk full, leader imbalance)

## Phase 7: Global & Future (Week 7-8)
> "Multi-DC, disaster recovery, and what's next."

- [ ] Read Lesson 13: Multi-Datacenter & DR
  - Understand: active-passive vs active-active vs aggregate patterns
  - Understand: MirrorMaker 2 (config, topic naming, limitations)
  - Understand: offset translation and checkpoint connector
  - Know: RPO/RTO for each architecture pattern
  - Understand: stretched clusters (when/when not)
  - Have: failover runbook reviewed

- [ ] Read Lesson 14: KRaft & The Future
  - Understand: why ZooKeeper was removed (6 problems)
  - Understand: KRaft architecture (controllers, metadata log, Raft)
  - Understand: __cluster_metadata topic (metadata as event log)
  - Know: ZK to KRaft migration path
  - Know: upcoming features (tiered storage, share groups, Kafka 4.0)

## Mastery Validation

When you can confidently answer these questions, you've reached god level:

1. "Draw what happens byte-by-byte when a producer sends a record to a broker."
2. "Explain how ISR, HW, and LEO work together during a leader failover."
3. "Design a system processing 1M events/sec with exactly-once guarantees."
4. "A consumer group is 30 minutes behind. Diagnose and fix it."
5. "Design a multi-DC Kafka architecture with < 30 second RPO."
6. "Why is Kafka fast? Explain all 5 performance mechanisms."
7. "When should you NOT use Kafka? Give 5 scenarios."
8. "Design partition strategy for an e-commerce platform (orders, users, payments)."
9. "Explain the difference between KRaft and ZooKeeper at the protocol level."
10. "Implement exactly-once consume-transform-produce and explain fencing."

## Resources

### Essential Reading
- "Kafka: The Definitive Guide" (2nd edition, O'Reilly, Shapira et al.)
- "Designing Event-Driven Systems" (Ben Stopford, free Confluent ebook)
- "Making Sense of Stream Processing" (Martin Kleppmann)

### KIPs to Study
- KIP-500: Replace ZooKeeper with KRaft
- KIP-405: Tiered Storage
- KIP-932: Queues for Kafka (Share Groups)
- KIP-848: Next-gen Consumer Group Protocol
- KIP-392: Follower Fetching

### Blogs & Talks
- Jay Kreps: "The Log" (the foundational blog post)
- Confluent blog: detailed KIP deep dives
- LinkedIn engineering blog: Kafka at scale
- Uber engineering blog: multi-DC Kafka
