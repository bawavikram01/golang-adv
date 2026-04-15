//go:build ignore
// =============================================================================
// LESSON 9.1: STREAM PROCESSING — Real-Time Computation on Kafka
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Stream-table duality: the foundational concept
// - Stateless processing: filter, map, flatMap
// - Stateful processing: aggregations, joins, windowing
// - Windowing: tumbling, hopping, sliding, session windows
// - Joins: stream-stream, stream-table, table-table (co-partitioning!)
// - State stores: RocksDB, changelog topics, interactive queries
// - Framework comparison: Kafka Streams vs Flink vs custom
//
// THE KEY INSIGHT:
// A Kafka topic is BOTH a stream AND a table:
// - Stream: append-only sequence of events (facts that happened)
// - Table: latest value per key (current state)
// Stream processing is about transforming streams into streams or tables.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== STREAM PROCESSING ===")
	fmt.Println()

	streamTableDuality()
	statelessProcessing()
	statefulProcessing()
	windowingDeepDive()
	joinsDeepDive()
	stateStores()
	frameworkComparison()
}

// =============================================================================
// PART 1: STREAM-TABLE DUALITY
// =============================================================================
func streamTableDuality() {
	fmt.Println("--- STREAM-TABLE DUALITY ---")

	// A STREAM is an unbounded sequence of events.
	// Each event is an immutable fact: "something happened."
	//
	// A TABLE is a snapshot of the latest value per key.
	// It represents "current state."
	//
	// THE DUALITY:
	// ─────────────
	// Stream → Table: Replay all events, keeping latest per key.
	//   Event stream: {key:A, val:1}, {key:B, val:2}, {key:A, val:3}
	//   Table state:  {A: 3, B: 2}
	//
	// Table → Stream: Capture every change to the table as an event.
	//   Table update: A=1 → emit {key:A, val:1}
	//   Table update: A=3 → emit {key:A, val:3}
	//   This is CDC (Change Data Capture)!
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  STREAM (topic, all events)         TABLE (compacted/state)  │
	// │                                                              │
	// │  offset 0: {A: 1}                   {A: 3}                   │
	// │  offset 1: {B: 2}                   {B: 2}                   │
	// │  offset 2: {A: 3}                   (latest per key)         │
	// │  offset 3: {C: 5}                   {C: 5}                   │
	// │  offset 4: {B: 7}                   {B: 7}                   │
	// │                                                              │
	// │  Stream = every event happened       Table = current truth   │
	// │  (immutable, append-only)           (mutable, queryable)     │
	// └──────────────────────────────────────────────────────────────┘
	//
	// In Kafka:
	// - Regular topic ≈ Stream (all events retained per retention policy)
	// - Compacted topic ≈ Table (only latest per key retained)
	// - KTable in Kafka Streams ≈ materialized table from a topic

	fmt.Println("  Stream: append-only sequence of immutable events")
	fmt.Println("  Table: latest value per key (current state)")
	fmt.Println("  Stream → Table: replay events. Table → Stream: capture changes.")
	fmt.Println()
}

// =============================================================================
// PART 2: STATELESS PROCESSING
// =============================================================================
func statelessProcessing() {
	fmt.Println("--- STATELESS PROCESSING ---")

	// Stateless = each record processed independently, no memory of previous records.
	//
	// FILTER: Keep records matching a predicate.
	//   Input:  [order:100, order:5, order:250, order:50]
	//   Filter: amount > 100
	//   Output: [order:250]
	//
	// MAP: Transform each record 1:1.
	//   Input:  [{"name":"alice","age":30}]
	//   Map:    extract name
	//   Output: ["alice"]
	//
	// FLATMAP: Transform each record into 0 or more records.
	//   Input:  ["hello world"]
	//   FlatMap: split by space
	//   Output: ["hello", "world"]
	//
	// BRANCH: Route records to different output topics based on predicates.
	//   Input:  [order:100, order:5, order:250]
	//   Branch: amount > 200 → high_value_orders
	//           amount > 50  → medium_orders
	//           else         → low_orders
	//
	// PERFORMANCE:
	// ────────────
	// Stateless operations are FAST:
	// - No state store lookups
	// - No changelog topic writes
	// - Scales linearly with partition count
	// - Typical: 100K-1M records/sec per thread
	//
	// If your pipeline is purely stateless, you might not need Kafka Streams.
	// A simple consumer loop with business logic works fine.

	fmt.Println("  Filter, Map, FlatMap, Branch — no state, no memory")
	fmt.Println("  Fast: 100K-1M records/sec per thread")
	fmt.Println("  Simple consumer loop often sufficient for stateless only")
	fmt.Println()
}

// =============================================================================
// PART 3: STATEFUL PROCESSING
// =============================================================================
func statefulProcessing() {
	fmt.Println("--- STATEFUL PROCESSING ---")

	// Stateful = processing depends on accumulated state from previous records.
	//
	// AGGREGATE: Combine records by key into a summary value.
	//   Input:  user_A:click, user_A:click, user_A:purchase
	//   Aggregate: count events per user
	//   State:  {user_A: 3}
	//
	// REDUCE: Special case of aggregate where types don't change.
	//   Input:  sensor_1:25°, sensor_1:27°, sensor_1:23°
	//   Reduce: keep max temperature
	//   State:  {sensor_1: 27°}
	//
	// COUNT: Count records per key.
	//   Input:  topic_A:msg, topic_A:msg, topic_B:msg
	//   Count:  {topic_A: 2, topic_B: 1}
	//
	// WHERE STATE LIVES:
	// ──────────────────
	// State is stored in a LOCAL STATE STORE per stream task.
	// Default implementation: RocksDB (embedded key-value store).
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  STATEFUL PROCESSING ARCHITECTURE:                            │
	// │                                                              │
	// │  Input Topic (partitioned)                                    │
	// │  ┌─────┐ ┌─────┐ ┌─────┐                                    │
	// │  │ P0  │ │ P1  │ │ P2  │                                     │
	// │  └──┬──┘ └──┬──┘ └──┬──┘                                     │
	// │     │       │       │                                        │
	// │  ┌──▼──┐ ┌──▼──┐ ┌──▼──┐  Stream Tasks                      │
	// │  │Task0│ │Task1│ │Task2│  (one per partition)                 │
	// │  │     │ │     │ │     │                                      │
	// │  │State│ │State│ │State│  Local RocksDB                       │
	// │  │Store│ │Store│ │Store│  (fast, local disk)                  │
	// │  └──┬──┘ └──┬──┘ └──┬──┘                                     │
	// │     │       │       │                                        │
	// │  ┌──▼──────▼──────▼──┐                                      │
	// │  │ Changelog Topic     │  Backup of state changes            │
	// │  │ (compacted)         │  For recovery after failure         │
	// │  └─────────────────────┘                                     │
	// └──────────────────────────────────────────────────────────────┘
	//
	// CHANGELOG TOPICS:
	// ─────────────────
	// Every state change is written to a changelog topic (compacted).
	// If a task crashes, the new owner rebuilds state by replaying the changelog.
	// This is how Kafka Streams achieves fault-tolerant stateful processing.

	fmt.Println("  Aggregate, Reduce, Count — depend on accumulated state")
	fmt.Println("  State lives in local RocksDB per stream task")
	fmt.Println("  Changelog topics for fault-tolerant state recovery")
	fmt.Println()
}

// =============================================================================
// PART 4: WINDOWING — Time-based grouping of events
// =============================================================================
func windowingDeepDive() {
	fmt.Println("--- WINDOWING ---")

	// Windows group events by time for bounded aggregations.
	// "Count clicks in the last 5 minutes" needs a window.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  WINDOW TYPES:                                                │
	// │                                                              │
	// │  TUMBLING (fixed, non-overlapping):                           │
	// │  ├──5min──┤├──5min──┤├──5min──┤                              │
	// │  00:00   05:00   10:00   15:00                               │
	// │  Each event belongs to exactly ONE window.                    │
	// │  Use for: regular business reporting, per-minute metrics.     │
	// │                                                              │
	// │  HOPPING (fixed, overlapping):                                │
	// │  ├──5min──┤                                                   │
	// │    ├──5min──┤                                                 │
	// │      ├──5min──┤                                               │
	// │  Advance: 1 min, Size: 5 min                                 │
	// │  Each event belongs to MULTIPLE windows (5 in this case).    │
	// │  Use for: moving averages, smoothed metrics.                  │
	// │                                                              │
	// │  SLIDING (based on event timestamps):                         │
	// │  Window exists only when events are within 'difference' of   │
	// │  each other. No fixed grid.                                   │
	// │  Use for: "events within 10 sec of each other"               │
	// │                                                              │
	// │  SESSION (gap-based):                                         │
	// │  ├─events─┤  gap  ├─events──┤  gap  ├─events─┤              │
	// │  Sessions end after inactivity gap (e.g., 30 min no events). │
	// │  Session size varies!                                         │
	// │  Use for: user sessions, activity periods, burst detection.   │
	// └──────────────────────────────────────────────────────────────┘
	//
	// EVENT TIME vs PROCESSING TIME:
	// ──────────────────────────────
	// Event time: timestamp embedded in the record (when it happened)
	// Processing time: wall clock when the stream processor receives it
	//
	// WHY EVENT TIME MATTERS:
	// A mobile app sends a click event at 10:00:00 but due to network delay,
	// it arrives at 10:05:00. Should it go in the 10:00 window or 10:05 window?
	// Event time: 10:00 (correct). Processing time: 10:05 (misleading).
	//
	// LATE ARRIVALS:
	// ──────────────
	// What if an event for 09:55 arrives at 10:10? The window is "closed."
	// grace period: how long to keep a window open for late events.
	//   Window(5min).grace(1min) → window 09:55-10:00 stays open until 10:01
	// After grace: late events are DROPPED (written to dead-letter topic if configured)
	//
	// WATERMARKS:
	// ───────────
	// Kafka Streams uses the lowest timestamp across all partitions as the
	// "stream time." Windows close when stream time advances past the grace period.
	// One slow partition can hold back ALL window closure!

	fmt.Println("  Tumbling: fixed, non-overlapping windows")
	fmt.Println("  Hopping: fixed, overlapping (moving average)")
	fmt.Println("  Session: gap-based, variable size")
	fmt.Println("  Always use event time, configure grace period for late arrivals")
	fmt.Println()
}

// =============================================================================
// PART 5: JOINS — Combining streams and tables
// =============================================================================
func joinsDeepDive() {
	fmt.Println("--- JOINS ---")

	// Joins combine data from two sources based on key.
	//
	// ┌──────────────────────────────────────────────────────────────────┐
	// │  JOIN TYPE         │ LEFT           │ RIGHT          │ RESULT   │
	// ├──────────────────────────────────────────────────────────────────┤
	// │  Stream-Stream     │ Stream (events)│ Stream (events)│ Stream   │
	// │  INNER JOIN        │ Must match key │ Within window  │ matched  │
	// │  LEFT JOIN         │ All left       │ If right match │ all left │
	// │                    │                │                │          │
	// │  Stream-Table      │ Stream (events)│ Table (lookup) │ Stream   │
	// │  (enrichment)      │ Each event     │ Current state  │ enriched │
	// │                    │                │                │          │
	// │  Table-Table       │ Table          │ Table          │ Table    │
	// │  (materialized)    │ State          │ State          │ combined │
	// └──────────────────────────────────────────────────────────────────┘
	//
	// STREAM-STREAM JOIN:
	// ────────────────────
	// Needs a WINDOW: "join click events with impression events within 1 hour"
	// Both sides buffered in state stores until the window expires.
	// Memory intensive for large windows!
	//
	// STREAM-TABLE JOIN:
	// ──────────────────
	// No window needed: lookup the table's current value for each stream record.
	// Classic use: enrich events with reference data.
	//   Stream: order_events (key: order_id)
	//   Table:  customer_info (key: customer_id)
	//   BUT: keys must match! Need to re-key order events by customer_id first.
	//
	// CO-PARTITIONING REQUIREMENT:
	// ────────────────────────────
	// For a join to work, BOTH inputs must be:
	//   1. Same number of partitions
	//   2. Same partitioning strategy (same key → same partition number)
	//
	// Why? Stream tasks are assigned partitions. Task 0 gets partition 0
	// from BOTH topics. If "customer_123" is in partition 3 of topic A
	// but partition 7 of topic B, Task 3 can't join them!
	//
	// FIX: Use through() / repartition() to rekey before joining.
	//   stream.selectKey((k, v) -> v.customerId)
	//         .through("repartitioned-topic")    // rekey by customer_id
	//         .join(customerTable, ...)           // NOW co-partitioned
	//
	// GLOBAL KTABLE:
	// ──────────────
	// Alternative to co-partitioning: use a GlobalKTable.
	// It replicates the ENTIRE table to every stream task instance.
	// Pros: No co-partitioning needed, join on any field.
	// Cons: Full copy on every instance → only for small tables (<1 GB).

	fmt.Println("  Stream-Stream: windowed join (both sides buffered)")
	fmt.Println("  Stream-Table: enrich events with lookup data")
	fmt.Println("  CRITICAL: co-partitioning required (same partitions, same keys)")
	fmt.Println("  GlobalKTable: replicate small tables everywhere (no co-partition needed)")
	fmt.Println()
}

// =============================================================================
// PART 6: STATE STORES — RocksDB, changelog, and interactive queries
// =============================================================================
func stateStores() {
	fmt.Println("--- STATE STORES ---")

	// Kafka Streams uses RocksDB as the default state store.
	// It's an embedded key-value store (LSM-tree based).
	//
	// WHY ROCKSDB:
	// ────────────
	// 1. Fast: optimized for write-heavy workloads (LSM tree)
	// 2. Disk-backed: can handle state larger than memory
	// 3. Embedded: no external dependency, runs in-process
	// 4. Compression: automatic data compression on disk
	//
	// STATE STORE + CHANGELOG = FAULT TOLERANCE:
	// ────────────────────────────────────────────
	// Every write to the state store is ALSO written to a changelog topic.
	// The changelog topic is compacted (only latest per key).
	//
	// If a task fails and moves to another instance:
	// 1. New instance creates empty RocksDB
	// 2. Replays the changelog topic from the beginning
	// 3. Rebuilds the EXACT same state
	// 4. Resumes processing from the last committed offset
	//
	// STANDBY REPLICAS:
	// ─────────────────
	// num.standby.replicas (default: 0)
	// If set to 1+, Kafka Streams maintains warm backups of state stores
	// on other instances. On failure, failover is instant (no changelog replay).
	// Cost: 2x state storage, extra network for keeping standby in sync.
	//
	// INTERACTIVE QUERIES:
	// ────────────────────
	// State stores can be queried directly via an API!
	// Instead of writing aggregation results to an output topic and reading
	// from another consumer, you query the KTable's state store in-process.
	//
	// Use case: "What's the current count for user_123?"
	//   store.get("user_123") → 42  (instant, no Kafka round-trip)
	//
	// For distributed queries (state might be on another instance):
	//   Kafka Streams exposes metadata: "which instance has user_123's partition?"
	//   Your application routes the query to the right instance via HTTP/gRPC.

	fmt.Println("  RocksDB: embedded LSM-tree, disk-backed, fast writes")
	fmt.Println("  Changelog topics: fault-tolerant state recovery")
	fmt.Println("  Standby replicas: instant failover (warm backup)")
	fmt.Println("  Interactive queries: query state stores directly (no output topic)")
	fmt.Println()
}

// =============================================================================
// PART 7: FRAMEWORK COMPARISON
// =============================================================================
func frameworkComparison() {
	fmt.Println("--- FRAMEWORK COMPARISON ---")

	// ┌────────────────┬────────────────┬────────────────┬──────────────┐
	// │ Feature        │ Kafka Streams  │ Apache Flink   │ Custom       │
	// ├────────────────┼────────────────┼────────────────┼──────────────┤
	// │ Deployment     │ Library (JAR)  │ Cluster        │ Any          │
	// │ Operations     │ Simple         │ Complex        │ You build it │
	// │ Exactly-once   │ Yes (Kafka)    │ Yes (any sink) │ Manual       │
	// │ State mgmt     │ Built-in       │ Built-in       │ You build it │
	// │ Windowing      │ Good           │ Excellent      │ You build it │
	// │ Throughput     │ High           │ Very High      │ Varies       │
	// │ Latency        │ Low (ms)       │ Low (ms)       │ Lowest       │
	// │ Multi-source   │ Kafka only     │ Any source     │ Any          │
	// │ SQL support    │ KSQL (separate)│ Flink SQL      │ No           │
	// │ Language       │ Java/Scala     │ Java/Scala/Py  │ Any          │
	// │ Learning curve │ Low            │ High           │ Depends      │
	// └────────────────┴────────────────┴────────────────┴──────────────┘
	//
	// WHEN TO USE EACH:
	//
	// KAFKA STREAMS: When your source AND sink are Kafka.
	//   "I need to do aggregation/enrichment/joins on Kafka topics."
	//   Simplest to deploy (just a JVM app), no cluster to manage.
	//   Scales by adding more app instances.
	//
	// FLINK: When you need complex event processing or non-Kafka sources.
	//   "I need to join Kafka with a database, process with complex windows,
	//    and write to S3 + Elasticsearch."
	//   CEP (Complex Event Processing), advanced windowing, SQL interface.
	//   Higher operational complexity (cluster management).
	//
	// CUSTOM (consumer loop): When processing is simple and stateless.
	//   "I need to filter messages and write to another topic."
	//   No framework overhead, any language (Go, Python, Rust...).
	//   You handle scaling, offset management, error handling.
	//
	// For Go specifically:
	// ─────────────────────
	// - Kafka Streams is Java-only (no Go port)
	// - In Go: use consumer loop + your own state management
	// - Libraries: franz-go for high-performance Kafka client
	// - If you need state: embed badger or bbolt for local state stores
	// - If you need windowing: implement time-bucketed aggregation manually
	// - If complex: consider Flink with Go producer/consumer at edges

	fmt.Println("  Kafka Streams: library, Kafka-only, simple deployment")
	fmt.Println("  Flink: cluster, any source/sink, complex processing")
	fmt.Println("  Custom (Go): consumer loop + local state. No framework overhead.")
	fmt.Println()
}











































































































































































































































































































}	fmt.Println("  Custom consumer: simplest cases, full control")	fmt.Println("  Flink: cluster-based, multi-source, most powerful")	fmt.Println("  Kafka Streams: library-based, Kafka-centric, lowest operational overhead")	// Go consumer (franz-go): When you're a Go shop and need simple stream processing	// Custom consumer: Simple transformations, no joins, no windows	// Apache Flink: Complex event processing, multi-source, need SQL interface	// Kafka Streams: Kafka-in, Kafka-out, moderate complexity, want library not cluster	// ─────────────────	// WHEN TO USE WHAT:	//	// └───────────────┴──────────────────┴───────────────────┴──────────────┘	// │ Best for      │ Kafka-centric    │ Complex pipelines │ Simple cases │	// │ Complexity    │ Low-Medium       │ High              │ High at scale│	// │ Throughput    │ Very high        │ Very high         │ Depends      │	// │ Latency       │ ~ms              │ ~ms               │ ~ms          │	// │ Language      │ Java/Scala       │ Java/Scala/Python │ Any          │	// │ Joins         │ Stream+Table     │ All types         │ DIY          │	// │ Windowing     │ Event time       │ Event time + more │ DIY          │	// │ EOS           │ Yes (Kafka only) │ Yes (broader)     │ DIY          │	// │ State         │ RocksDB+changelog│ RocksDB+checkpoint│ You build it │	// │ Scaling       │ Add instances    │ JobManager manages│ Manual       │	// │ Deployment    │ Library (JAR)    │ Cluster (JobMgr)  │ Any process  │	// ├───────────────┼──────────────────┼───────────────────┼──────────────┤	// │ Feature       │ Kafka Streams    │ Apache Flink      │ Custom Consumer│	// ┌───────────────┬──────────────────┬───────────────────┬──────────────┐	fmt.Println("--- FRAMEWORK COMPARISON ---")func frameworkComparison() {}	fmt.Println()	fmt.Println("  Interactive queries: query state stores in real-time")	fmt.Println("  Fast recovery: local RocksDB first, changelog replay as fallback")	fmt.Println("  State stores: RocksDB local + Kafka changelog backup")	// which instance has it. Use the Metadata API to find the right instance.	// But: state is PARTITIONED. To query ANY key, you need to know	//	// This turns Kafka Streams into a queryable database.	// → Look up in local state store → instant answer.	// "What is the current count for user-123?"	// State stores can be queried in real-time!	// ────────────────────	// INTERACTIVE QUERIES:	//	// - Lightweight: embedded database, no separate process	// - Range scans: efficient for windowed operations	// - Point lookups: O(1) amortized	// - Can hold more data than memory (spills to disk)	// - LSM-tree storage: fast writes (append-only)	// ─────────────	// WHY ROCKSDB?	//	// └──────────────────────────────────────────────────────────────┘	// │     (slow, but guaranteed to recover)                        │	// │  2. If not → replay changelog topic → rebuild state          │	// │  1. Check local RocksDB → if valid, use it (fast!)          │	// │  ON RESTART:                 ▼                               │	// │                              │                               │	// │                    (compacted)                                │	// │                    changelog topic                            │	// │                              │                               │	// │  └───────────────────────────┬───────────────────────────┘   │	// │  │  - Backed up to Kafka changelog topic                  │   │	// │  │  - Persisted to local disk (fast!)                     │   │	// │  │  - Contains state for assigned partitions ONLY         │   │	// │  │  - Local to THIS instance                              │   │	// │  │  State Store (RocksDB)                                 │   │	// │  ┌───────────────────────────────────────────────────────┐   │	// │  Stream Processor Instance                                    │	// ┌──────────────────────────────────────────────────────────────┐	//	// STATE STORE ARCHITECTURE:	fmt.Println("--- STATE STORES ---")func stateStores() {}	fmt.Println()	fmt.Println("  CRITICAL: inputs must be co-partitioned (same count + same key)")	fmt.Println("  Stream-Stream: correlation within time window")	fmt.Println("  Stream-Table: enrichment (O(1) local lookup)")	// └──────────────────────────────────────────────────────────────┘	// │  If not co-partitioned: you must repartition before joining.  │	// │  partitioning strategy (default murmur2).                     │	// │  Both inputs MUST have the same partition count AND same       │	// │  CO-PARTITIONING REQUIREMENT:                                  │	// │                                                              │	// │  Result: users + addresses (continuously updated)              │	// │  Table2: addresses (user-id → address)                        │	// │  Table1: users (user-id → profile)                            │	// │  ──────────────────────────────────                            │	// │  TABLE-TABLE JOIN (materialized view):                         │	// │                                                              │	// │  Requirement: co-partitioned by join key.                      │	// │  stream2's buffer within the time window. And vice versa.     │	// │  For each record in stream1, look for matching records in     │	// │  Implementation: both streams buffered in windowed state stores.│	// │                                                              │	// │  Result: matched impressions+clicks within 5 minutes         │	// │  Stream2: ad-clicks (user-id, ad-id, timestamp)              │	// │  Stream1: ad-impressions (user-id, ad-id, timestamp)         │	// │  ────────────────────────────────                              │	// │  STREAM-STREAM JOIN (correlation):                             │	// │                                                              │	// │  Requirement: both keyed by user-id (co-partitioned).        │	// │  For each order, look up user in local store. O(1) lookup!   │	// │  Implementation: user table is loaded into local state store. │	// │                                                              │	// │  Result: enriched orders (order-id, user-id, name, amount)   │	// │  Table: users (user-id → name, email)                         │	// │  Stream: orders (order-id, user-id, amount)                  │	// │  ─────────────────────────────                                │	// │  STREAM-TABLE JOIN (enrichment):                               │	// │                                                              │	// │  JOIN TYPES:                                                   │	// ┌──────────────────────────────────────────────────────────────┐	fmt.Println("--- JOINS ---")func joins() {}	fmt.Println()	fmt.Println("  ALWAYS use event time, configure grace period for late events")	fmt.Println("  Session: gap-based variable length")	fmt.Println("  Hopping: fixed overlapping (sliding aggregate)")	fmt.Println("  Tumbling: fixed non-overlapping windows")	// After grace period, late events are DROPPED.	// How long to wait for late events before closing a window.	// GRACE PERIOD:	//	// Late events (network delays, retries) should go into the correct window.	// ALWAYS USE EVENT TIME for windowing!	//	// Processing time: when the stream processor sees the record	// Event time: timestamp in the record (when it actually happened)	// ──────────────────────────────	// EVENT TIME vs PROCESSING TIME:	//	// └──────────────────────────────────────────────────────────────┘	// │  Use: "user session duration" where session = burst of events │	// │  Inactivity gap (e.g., 30 min). Session closes after gap.    │	// │                                                              │	// │    └──── session 1 ────┘       └─session 2─┘       └─s3──┘  │	// │  ──[event..event..event]──gap──[event..event]──gap──[event]─ │	// │                                                              │	// │  SESSION WINDOW (gap-based, variable length):                 │	// │                                                              │	// │  Use: "join events within 10 seconds of each other"          │	// │  Centered on each event ± time difference.                    │	// │                                                              │	// │  SLIDING WINDOW (event-triggered, used for joins):            │	// │                                                              │	// │  Each event can be in MULTIPLE windows!                       │	// │  Use: "5-minute moving average, updated every minute"        │	// │                                                              │	// │  Advance every 1 minute.                                      │	// │            [──────5min──────]                                  │	// │       [──────5min──────]                                      │	// │  ──[──────5min──────]                                        │	// │                                                              │	// │  HOPPING WINDOW (fixed, overlapping):                         │	// │                                                              │	// │  Use: "count events per 5 minutes" (each event in ONE window)│	// │                                                              │	// │    └─ count events ─┘└─ count events ─┘└─ count events ─┘   │	// │  ──[──────5min──────][──────5min──────][──────5min──────]──  │	// │                                                              │	// │  TUMBLING WINDOW (fixed, non-overlapping):                    │	// ┌──────────────────────────────────────────────────────────────┐	//	// Windowing groups records by TIME for aggregation.	fmt.Println("--- WINDOWING ---")func windowing() {}	fmt.Println()	fmt.Println("  State stores: local RocksDB + Kafka changelog for recovery")	fmt.Println("  Aggregate, reduce, join: require state across records")	//   State is partitioned by key (same key → same partition → same store)	//   Local key-value store (RocksDB) + Kafka changelog topic backup	// SOLUTION: State Stores (see Part 6)	//	// 3. State must be replayable (if state is lost, rebuild from Kafka)	// 2. State must be partitioned (you can't keep ALL state on one machine)	// 1. State must survive process restarts → needs persistence	// STATEFUL = HARD because:	//	//   enrich orders with user data → need user table in memory	// JOIN: Combine records from two streams/tables	//	//   latest order per customer → need to remember last order	// REDUCE: Combine all records for a key into one value	//	//   count page views per user → need to remember previous count	// AGGREGATE: Combine records (count, sum, average)	//	// Operations that require maintaining state across records:	fmt.Println("--- STATEFUL PROCESSING ---")func statefulProcessing() {}	fmt.Println()	fmt.Println("  Each record processed independently → easy to parallelize")	fmt.Println("  Filter, map, flatMap, branch: no state, infinitely scalable")	// Each partition can be processed independently.	// These are simple: no state, no coordination, easy to scale.	//	//   input: all events → branch(isError? → errors topic, else → normal topic)	// BRANCH: Route records to different topics based on conditions	//	//   input: sentence → flatMap(split words) → individual words	// FLATMAP: Transform each record into 0 or more records	//	//   input: raw event → map(event → enrichedEvent) → enriched events	// MAP: Transform each record	//	//   input: all orders → filter(order.amount > 1000) → high-value orders	// FILTER: Keep only records matching a condition	//	// Operations that process each record independently:	fmt.Println("--- STATELESS PROCESSING ---")func statelessProcessing() {}	fmt.Println()	fmt.Println("  Duality: stream ↔ table (two views of the same data)")	fmt.Println("  Table = current state built from stream (compacted topic)")	fmt.Println("  Stream = unbounded sequence of events (Kafka topic)")	// └──────────────────────────────────────────────────────────────┘	// │  A materialized view turns a Kafka stream INTO a table.       │	// │  CDC turns a database table INTO a Kafka stream.              │	// │  A Kafka topic without compaction IS a stream.                │	// │  A Kafka topic with log compaction IS a table.                │	// │                                                              │	// │  →  STREAM of changes (this IS a Kafka topic!)               │	// │  Table changed: user-1 updated → emit(user-1, Alice Smith)   │	// │  Table changed: user-2 added → emit(user-2, Bob)             │	// │  Table changed: user-1 added → emit(user-1, Alice)           │	// │  TABLE → STREAM: Log every change as an event                │	// │                                                              │	// │  →  TABLE: {user-1: "Alice Smith", user-2: "Bob"}            │	// │  UPDATE user-1, name=Alice Smith                              │	// │  INSERT user-2, name=Bob                                      │	// │  INSERT user-1, name=Alice                                    │	// │  STREAM → TABLE: Apply each event to build current state     │	// │                                                              │	// │  THE DUALITY:                                                 │	// ┌──────────────────────────────────────────────────────────────┐	fmt.Println("--- STREAM-TABLE DUALITY ---")func streamTableDuality() {}	frameworkComparison()	stateStores()	joins()	windowing()	statefulProcessing()	statelessProcessing()	streamTableDuality()	fmt.Println()	fmt.Println("=== STREAM PROCESSING ===")func main() {import "fmt"package main// =============================================================================//// and a table's changes produce a stream. This is the stream-table duality.// Tables and streams are DUAL: a stream of changes creates a table,// Instead of just moving data, you COMPUTE on data as it flows.// Stream processing turns Kafka from a "pipe" into a "computer."// THE KEY INSIGHT://// - When to use Kafka Streams vs Flink vs custom consumers// - Exactly-once stream processing// - State stores: local state backed by changelog topics// - Joins: stream-stream, stream-table, table-table// - Windowing: tumbling, hopping, sliding, session// - Stateless vs stateful stream processing// WHAT YOU'LL LEARN://// =============================================================================// LESSON 9.1: STREAM PROCESSING — Real-Time Computation on Kafka// =============================================================================