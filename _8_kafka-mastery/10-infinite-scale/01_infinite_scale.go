//go:build ignore
// =============================================================================
// LESSON 10.1: DESIGNING FOR INFINITE SCALE
// =============================================================================
//
// THIS IS THE LESSON THAT SEPARATES KAFKA USERS FROM KAFKA ARCHITECTS.
//
// WHAT YOU'LL LEARN:
// - Topic design patterns for millions of events/sec
// - How to scale consumers beyond partition count
// - Backpressure strategies: what to do when consumers can't keep up
// - Multi-tenancy: sharing Kafka clusters across teams
// - The "Kafka as the central nervous system" architecture
// - Real-world architectures from LinkedIn, Uber, Netflix
//
// THE KEY INSIGHT:
// Infinite scale doesn't mean "one huge Kafka cluster."
// It means designing LAYERS — each layer scales independently:
// - Topic layer: right partition count, right key strategy
// - Consumer layer: parallelism, backpressure, graceful degradation
// - Operational layer: multi-cluster, tiered storage, quotas
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== DESIGNING FOR INFINITE SCALE ===")
	fmt.Println()

	topicDesignPatterns()
	consumerScaling()
	backpressureStrategies()
	multiTenancy()
	realWorldArchitectures()
	scaleAntiPatterns()
}

// =============================================================================
// PART 1: TOPIC DESIGN PATTERNS
// =============================================================================
func topicDesignPatterns() {
	fmt.Println("--- TOPIC DESIGN PATTERNS ---")

	// PATTERN 1: SINGLE TOPIC PER EVENT TYPE
	// ───────────────────────────────────────
	// topic: user.created
	// topic: user.updated
	// topic: order.placed
	// topic: order.shipped
	//
	// Pros: Simple, clear ownership, independent retention
	// Cons: Many topics, harder to get "all user events in order"
	// Use: Microservices publishing domain events
	//
	// PATTERN 2: SINGLE TOPIC FOR ENTITY
	// ───────────────────────────────────
	// topic: users (key=user-id, value=any user event)
	// topic: orders (key=order-id, value=any order event)
	//
	// Pros: All events for an entity in order (same partition)
	// Cons: Different event shapes in same topic (needs schema evolution care)
	// Use: Event sourcing, when you need entity-level ordering
	//
	// PATTERN 3: TOPIC PER DATA SOURCE
	// ─────────────────────────────────
	// topic: clickstream-web
	// topic: clickstream-mobile
	// topic: clickstream-iot
	//
	// Merged by a stream processor into: clickstream-unified
	//
	// Pros: Independent ingestion, source-specific retention
	// Cons: Need a merge/union step
	// Use: Data lake ingestion, multi-source analytics
	//
	// PATTERN 4: TIERED TOPICS (fan-out)
	// ──────────────────────────────────
	//        raw.events (high volume, short retention)
	//            │
	//            ├──► processed.events (filtered, enriched)
	//            │        │
	//            │        ├──► analytics.events (aggregated)
	//            │        └──► alerts.events (threshold breached)
	//            └──► archive.events (long retention, compressed)
	//
	// Each tier has different partition counts, retention, and consumers.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  TOPIC NAMING CONVENTION:                                     │
	// │                                                              │
	// │  <domain>.<entity>.<event>                                    │
	// │  OR                                                          │
	// │  <team>.<domain>.<version>                                    │
	// │                                                              │
	// │  Examples:                                                    │
	// │  payments.orders.created                                      │
	// │  analytics.clickstream.v2                                     │
	// │  platform.users.profile-updated                               │
	// │                                                              │
	// │  Rules:                                                       │
	// │  - Lowercase, dot-separated (not underscores)                │
	// │  - Include domain/team ownership                              │
	// │  - Version suffix when schema breaks                         │
	// │  - Max ~249 chars (Kafka limit)                               │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  Single topic per event type: simplest, most common")
	fmt.Println("  Single topic per entity: when entity-level ordering needed")
	fmt.Println("  Tiered topics: raw → processed → analytics (each tier scales independently)")
	fmt.Println()
}

// =============================================================================
// PART 2: CONSUMER SCALING
// =============================================================================
func consumerScaling() {
	fmt.Println("--- CONSUMER SCALING ---")

	// FUNDAMENTAL LIMIT:
	// ──────────────────
	// In a consumer group, max parallelism = number of partitions.
	// 30 partitions → max 30 consumers. The 31st consumer sits idle.
	//
	// STRATEGIES TO SCALE BEYOND THIS:
	//
	// STRATEGY 1: Multiple Consumer Groups
	// ─────────────────────────────────────
	// Group-A: real-time alerting (fast, drop if behind)
	// Group-B: data lake ingestion (batch, can lag)
	// Group-C: search indexing (moderate speed)
	//
	// Each group independently reads ALL data.
	// No interference between groups. Each scales to partition count.
	//
	// STRATEGY 2: Consumer + Thread Pool
	// ───────────────────────────────────
	// Each consumer distributes work to a thread pool.
	// Consumer polls → hands records to 10 worker threads → commit after all done.
	//
	// Effective parallelism: partitions × threads_per_consumer
	// 30 partitions × 10 threads = 300 effective workers!
	//
	// GOTCHA: Ordering within a partition is lost (threads process concurrently).
	// FIX: Route records with the same key to the same thread:
	//   thread = hash(record.key) % numThreads
	//   This preserves per-key ordering within per-partition ordering.
	//
	// STRATEGY 3: Consumer + External Queue
	// ─────────────────────────────────────
	// Consumer → push to in-memory queue → pool of workers
	// Decouple consumption from processing.
	// Useful when processing time varies wildly per record.
	//
	// STRATEGY 4: Increase Partitions (last resort)
	// ──────────────────────────────────────────────
	// If 30 partitions × 10 threads still isn't enough:
	// Increase partitions. But remember the costs from Lesson 05.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  SCALING DECISION TREE:                                       │
	// │                                                              │
	// │  Can you add more consumers (< partition count)?              │
	// │  ├── YES → add consumers                                     │
	// │  └── NO (at partition count) →                                │
	// │      Can you add thread pool inside consumers?               │
	// │      ├── YES → consumer + thread pool (preserve key ordering)│
	// │      └── NO (still not enough) →                              │
	// │          Can you optimize processing speed?                   │
	// │          ├── YES → optimize first (async I/O, batching)      │
	// │          └── NO → increase partition count + consumers        │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  Max parallelism per group = partition count")
	fmt.Println("  Scale beyond: thread pool per consumer (partition × threads)")
	fmt.Println("  Preserve ordering: hash(key) % threads → same key same thread")
	fmt.Println()
}

// =============================================================================
// PART 3: BACKPRESSURE STRATEGIES
// =============================================================================
func backpressureStrategies() {
	fmt.Println("--- BACKPRESSURE STRATEGIES ---")

	// When consumers can't keep up with producers:
	//
	// STRATEGY 1: LET LAG GROW (absorb spikes)
	// ─────────────────────────────────────────
	// Kafka IS the buffer. Increase retention to absorb the spike.
	// Consumer catches up when spike ends.
	// Pros: simple, no data loss
	// Cons: lag grows unbounded, potential page cache eviction
	//
	// STRATEGY 2: THROTTLE PRODUCERS
	// ──────────────────────────────
	// Kafka quotas: produce_throttle_time_ms
	// Set per clientId or per user quotas:
	//   producer_byte_rate = 10485760  (10 MB/s per producer)
	// When exceeded: broker delays the response (backpressure to producer)
	//
	// STRATEGY 3: DROP & SAMPLE
	// ─────────────────────────
	// Consumer skips records when behind:
	//   if consumerLag > threshold:
	//       process every 10th record (sampling)
	//   OR: seek to latest offset (skip the backlog)
	//
	// Use for: metrics, analytics where sampling is acceptable.
	//
	// STRATEGY 4: PRIORITY TOPICS
	// ───────────────────────────
	// Split events by priority:
	//   topic: events-high-priority (always processed)
	//   topic: events-low-priority (processed when capacity available)
	//
	// Consumer reads both, but pauses low-priority when behind on high-priority.
	//
	// STRATEGY 5: AUTO-SCALING CONSUMERS
	// ──────────────────────────────────
	// Monitor consumer lag → auto-scale consumer instances:
	//   if lag > 10K for > 5 min: scale up
	//   if lag < 1K for > 30 min: scale down
	//
	// Works perfectly with Kubernetes HPA + custom lag metric.

	fmt.Println("  Kafka IS the buffer (let lag grow for spikes)")
	fmt.Println("  Quotas: throttle producers at the broker level")
	fmt.Println("  Auto-scale consumers based on lag metric")
	fmt.Println()
}

// =============================================================================
// PART 4: MULTI-TENANCY
// =============================================================================
func multiTenancy() {
	fmt.Println("--- MULTI-TENANCY ---")

	// Sharing a Kafka cluster across teams/services:
	//
	// QUOTAS:
	// ───────
	// Per-client quotas (KIP-13):
	//   producer_byte_rate: max bytes/sec per producer client ID
	//   consumer_byte_rate: max bytes/sec per consumer client ID
	//   request_percentage: max fraction of broker request handler time
	//
	// Per-user quotas (KIP-55):
	//   Quotas tied to authenticated user (via SASL)
	//   More granular than client ID
	//
	// TOPIC-LEVEL ISOLATION:
	// ──────────────────────
	// kafka-configs.sh --alter --entity-type topics --entity-name my-topic \
	//   --add-config retention.ms=86400000,segment.bytes=1073741824
	//
	// Each team manages their own topic configs.
	// Central platform team manages broker configs.
	//
	// ACLs:
	// ─────
	// kafka-acls.sh --add --allow-principal User:team-a \
	//   --operation Write --topic 'team-a.*' --resource-pattern-type prefixed
	//
	// Teams can only write to topics with their prefix.
	// Central team can read all topics for monitoring.
	//
	// CLUSTER SIZING FOR MULTI-TENANCY:
	// ──────────────────────────────────
	// 1. Sum all teams' throughput requirements
	// 2. Add 40% headroom (some teams will burst)
	// 3. Add 10% for replication traffic
	// 4. Add capacity for cold consumers (historical reads)
	// 5. Plan for N-1 broker failure (can you survive losing one broker?)
	//
	// WHEN TO USE SEPARATE CLUSTERS:
	// ──────────────────────────────
	// - Hard latency SLAs (one team's behavior shouldn't affect another)
	// - Regulatory isolation (PCI, HIPAA, GDPR)
	// - Different availability requirements (prod vs dev)
	// - Team wants full control over configs

	fmt.Println("  Quotas: per-client or per-user byte rates")
	fmt.Println("  ACLs: topic-prefix-based access control")
	fmt.Println("  Separate clusters when: hard SLAs, regulatory, different availability")
	fmt.Println()
}

// =============================================================================
// PART 5: REAL-WORLD ARCHITECTURES
// =============================================================================
func realWorldArchitectures() {
	fmt.Println("--- REAL-WORLD ARCHITECTURES ---")

	// LINKEDIN (where Kafka was born):
	// ─────────────────────────────────
	// - 7+ trillion messages/day (2023)
	// - 100+ Kafka clusters
	// - 100K+ partitions per cluster
	// - Uses Kafka for: activity tracking, metrics, CDC, stream processing
	//
	// Architecture:
	//   Source → "Tracking" Kafka → MirrorMaker → "Standard" Kafka → Consumers
	//   Near-realtime pipeline with separate ingestion and consumption clusters.
	//
	// UBER:
	// ─────
	// - 40+ PB/day (2023)
	// - 1300+ brokers, multiple clusters
	// - Key pattern: "uReplicator" (custom MirrorMaker for cross-DC)
	// - Consumer proxy service: centralized consumer management
	//
	//   ┌──────────┐     ┌──────────┐     ┌──────────┐
	//   │ DC-West  │ ──► │ Aggregate│ ──► │ DC-East  │
	//   │ ingestion│     │ cluster  │     │ consumers│
	//   └──────────┘     └──────────┘     └──────────┘
	//       └──────────────────────────────────┘
	//                   uReplicator
	//
	// NETFLIX:
	// ────────
	// - Apache Kafka + custom routing layer
	// - Key pattern: "Keystone" data pipeline
	//   Frontend events → Kafka → real-time processing → S3/ElasticSearch
	// - Multi-region: Kafka clusters in every AWS region
	// - Uses Kafka for: A/B test events, personalization, analytics
	//
	// COMMON PATTERNS ACROSS ALL:
	// ───────────────────────────
	// 1. SEPARATE INGESTION AND CONSUMPTION CLUSTERS
	//    Don't let slow consumers affect producer SLAs.
	//
	// 2. TIERED ARCHITECTURE
	//    Raw → processed → aggregated → served
	//    Each tier scales independently.
	//
	// 3. CONSUMER PROXY SERVICES
	//    Centralize consumer group management, monitoring, auto-scaling.
	//
	// 4. SCHEMA GOVERNANCE
	//    Central Schema Registry, compatibility enforcement, schema review process.
	//
	// 5. SELF-SERVICE + GUARDRAILS
	//    Teams create topics via API, but with sane defaults enforced.

	fmt.Println("  LinkedIn: 7T+ msg/day, 100+ clusters, separate ingestion/consumption")
	fmt.Println("  Uber: 40+ PB/day, custom replicator, consumer proxy service")
	fmt.Println("  Common: tiered architecture, separate clusters, self-service + guardrails")
	fmt.Println()
}

// =============================================================================
// PART 6: SCALE ANTI-PATTERNS
// =============================================================================
func scaleAntiPatterns() {
	fmt.Println("--- SCALE ANTI-PATTERNS ---")

	// ANTI-PATTERN 1: ONE MASSIVE TOPIC
	// ──────────────────────────────────
	// "Let's put ALL events in one topic with 10,000 partitions!"
	// Problem: different event types have different retention, throughput,
	// and consumer needs. One topic can't serve all.
	//
	// ANTI-PATTERN 2: TOO MANY TINY TOPICS
	// ─────────────────────────────────────
	// "One topic per customer!" with 100K customers = 100K topics.
	// Each topic = metadata overhead, controller load, monitoring complexity.
	// Fix: key by customer within a shared topic.
	//
	// ANTI-PATTERN 3: SYNCHRONOUS REQUEST/REPLY VIA KAFKA
	// ────────────────────────────────────────────────────
	// Producer sends request → waits for response on reply topic.
	// Kafka is NOT designed for request/reply (high latency).
	// Use gRPC/HTTP for sync, Kafka for async.
	//
	// ANTI-PATTERN 4: KAFKA AS A DATABASE
	// ───────────────────────────────────
	// "Let's query Kafka directly for user data!"
	// Kafka is optimized for sequential access, NOT random lookups.
	// Use compacted topics for state transfer, NOT as a primary database.
	// Materialize into a proper DB (PostgreSQL, Redis, Elasticsearch).
	//
	// ANTI-PATTERN 5: IGNORING CONSUMER GROUP MANAGEMENT
	// ──────────────────────────────────────────────────
	// Hundreds of abandoned consumer groups → __consumer_offsets grows,
	// coordinator overhead increases, monitoring noise.
	// Clean up unused consumer groups regularly.

	fmt.Println("  Don't: one massive topic, topic-per-customer, sync request/reply")
	fmt.Println("  Don't: use Kafka as a database for random lookups")
	fmt.Println("  Do: right-sized topics, async patterns, materialize to proper DBs")
}
