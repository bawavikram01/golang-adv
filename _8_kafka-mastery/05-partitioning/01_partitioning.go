//go:build ignore
// =============================================================================
// LESSON 5.1: PARTITIONING — The Art and Science of Data Distribution
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - How to choose the right number of partitions (formula + intuition)
// - Key design: high cardinality, entity-based, composite keys
// - Ordering guarantees: what Kafka promises (and what it doesn't)
// - Hot partitions: detection, root causes, and solutions
// - Partition expansion: why you can't shrink, and what breaks when you grow
// - Custom partitioners: when and how to write your own
//
// THE KEY INSIGHT:
// Partitions are Kafka's UNIT OF PARALLELISM. A topic with 10 partitions can
// have at most 10 concurrent consumers. Partitions are also the UNIT OF ORDERING.
// Kafka only guarantees ordering WITHIN a single partition.
// Get partitioning wrong → you get hot spots, ordering bugs, or scaling walls.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== PARTITIONING MASTERY ===")
	fmt.Println()

	partitionCountFormula()
	keyDesignStrategies()
	orderingGuarantees()
	hotPartitions()
	partitionExpansion()
	customPartitioners()
}

// =============================================================================
// PART 1: HOW MANY PARTITIONS? — The formula and the intuition
// =============================================================================
func partitionCountFormula() {
	fmt.Println("--- PARTITION COUNT FORMULA ---")

	// THE FORMULA:
	//
	//   partitions = max(T/Tp, T/Tc)
	//
	// Where:
	//   T  = target throughput (e.g., 1 GB/sec)
	//   Tp = throughput per partition on the PRODUCER side
	//   Tc = throughput per partition on the CONSUMER side
	//
	// TYPICAL THROUGHPUT PER PARTITION:
	// ─────────────────────────────────
	// Producer: 10-100 MB/sec per partition (depends on batch size, compression)
	// Consumer: 5-50 MB/sec per partition (depends on processing complexity)
	//
	// EXAMPLE:
	//   Target: 1 GB/sec
	//   Producer per partition: 50 MB/sec → 1000/50 = 20 partitions
	//   Consumer per partition: 20 MB/sec → 1000/20 = 50 partitions
	//   Answer: max(20, 50) = 50 partitions
	//
	// BUT THE FORMULA IS JUST THE START. CONSIDER:
	//
	// ┌────────────────────────────────────────────────────────────────────┐
	// │  FACTOR                │ IMPACT                                    │
	// ├────────────────────────────────────────────────────────────────────┤
	// │  Consumer parallelism  │ More partitions = more parallel consumers │
	// │  Future growth         │ Can't easily change later, overprovision  │
	// │  Broker memory         │ Each partition uses ~10KB broker memory   │
	// │  Leader elections      │ More partitions = slower leader election  │
	// │  End-to-end latency    │ More partitions = slightly higher latency│
	// │  File handles          │ Each partition = segment files (3+ FDs)   │
	// │  Replication traffic   │ More partitions × RF = more replication   │
	// │  Consumer rebalance    │ More partitions = slower rebalance        │
	// └────────────────────────────────────────────────────────────────────┘
	//
	// PRACTICAL GUIDELINES:
	// ─────────────────────
	// Low throughput (<10 MB/s): 6-12 partitions (still allows good parallelism)
	// Medium throughput (10-100 MB/s): 12-50 partitions
	// High throughput (100 MB/s - 1 GB/s): 50-200 partitions
	// Extreme (>1 GB/s): 200-500 partitions (consider multiple topics)
	//
	// NEVER exceed numPartitions per topic > numBrokers × 4000
	// (Confluent recommendation: ~4000 partitions per broker across all topics)
	//
	// THE GOLDILOCKS RULE:
	// ────────────────────
	// Too few: can't scale consumers, throughput ceiling
	// Too many: slow rebalance, high broker overhead, wasteful idle partitions
	// Just right: max(expected_consumers, throughput/partition_throughput) × 1.5

	fmt.Println("  Formula: partitions = max(T/Tp, T/Tc)")
	fmt.Println("  Typical: 12-50 for medium, 50-200 for high throughput")
	fmt.Println("  Rule: overprovision slightly (can't shrink later)")
	fmt.Println("  Limit: ~4000 partitions per broker across all topics")
	fmt.Println()
}

// =============================================================================
// PART 2: KEY DESIGN STRATEGIES
// =============================================================================
func keyDesignStrategies() {
	fmt.Println("--- KEY DESIGN STRATEGIES ---")

	// The key determines WHICH PARTITION a record goes to.
	// Default: murmur2(keyBytes) % numPartitions
	//
	// KEY DESIGN PRINCIPLES:
	//
	// 1. HIGH CARDINALITY
	// ────────────────────
	// Bad key: country (10-50 values → hot partitions)
	// Bad key: boolean field (2 values → 2 useful partitions)
	// Good key: user_id (millions of values → even distribution)
	// Good key: order_id, session_id, device_id
	//
	// 2. ENTITY-BASED KEYS
	// ─────────────────────
	// Use the entity whose events need ordering as the key.
	//
	// E-commerce order events (created, paid, shipped, delivered):
	//   Key: order_id → all events for one order land in same partition
	//   Guarantees: consumer sees events in order per order!
	//
	// User activity (login, click, purchase, logout):
	//   Key: user_id → all activity for one user in same partition
	//   Enables: session reconstruction, user behavior analysis
	//
	// IoT device readings:
	//   Key: device_id → all readings from one device in order
	//
	// 3. COMPOSITE KEYS
	// ──────────────────
	// When you need ordering at a sub-entity level:
	//   Key: "tenant_123:user_456" → ordering per user per tenant
	//   Key: "region_us:sensor_789" → ordering per sensor per region
	//
	// 4. NULL KEYS
	// ────────────
	// When ordering doesn't matter (metrics, logs, clickstream):
	//   Key: null → sticky partitioner distributes evenly
	//   Pro: best distribution, no hot partitions
	//   Con: no ordering guarantee whatsoever
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  KEY DESIGN DECISION TREE:                                    │
	// │                                                              │
	// │  Need ordering for records? ──── NO → null key               │
	// │          │                                                    │
	// │         YES                                                   │
	// │          │                                                    │
	// │  Ordering by which entity? ──── entity_id as key             │
	// │          │                                                    │
	// │  High cardinality? ──── YES → good, use it                   │
	// │          │                                                    │
	// │         NO                                                    │
	// │          │                                                    │
	// │  Composite key (entity + sub_entity)                          │
	// │  or custom partitioner                                        │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  Key = entity whose events need ordering (user_id, order_id)")
	fmt.Println("  High cardinality = even distribution")
	fmt.Println("  null key = no ordering needed, best distribution")
	fmt.Println()
}

// =============================================================================
// PART 3: ORDERING GUARANTEES — What Kafka actually promises
// =============================================================================
func orderingGuarantees() {
	fmt.Println("--- ORDERING GUARANTEES ---")

	// KAFKA'S ORDERING GUARANTEE:
	// ───────────────────────────
	// Records within a SINGLE PARTITION are ordered by offset.
	// Consumers read records in offset order within each partition.
	// THAT'S IT. No cross-partition ordering.
	//
	// WHAT THIS MEANS:
	//
	// Partition 0: [A offset=0] [C offset=1] [E offset=2]
	// Partition 1: [B offset=0] [D offset=1] [F offset=2]
	//
	// Consumer may see: A, B, C, D, E, F  (fair interleaving)
	//              or:  A, C, B, E, D, F  (partition 0 faster)
	//              or:  B, D, A, F, C, E  (partition 1 faster)
	//
	// But WITHIN each partition:
	//   A always before C always before E ✓
	//   B always before D always before F ✓
	//
	// COMMON ORDERING TRAPS:
	// ──────────────────────
	// Trap 1: "I need global ordering"
	//   → Use 1 partition. But throughput limited to 1 consumer. Terrible.
	//   → Better: design your system to only need per-key ordering.
	//
	// Trap 2: "User A's events must happen before User B's events"
	//   → That's cross-partition ordering. Kafka can't do this.
	//   → Use timestamps + reordering buffer in the consumer.
	//
	// Trap 3: "Retries mess up ordering"
	//   → Without idempotent producer: YES! Use max.in.flight=1.
	//   → With idempotent producer: NO. max.in.flight=5 is safe.
	//
	// Trap 4: "Multiple consumers reading same partition"
	//   → In a consumer group: ONE consumer per partition. Ordering preserved.
	//   → Multiple groups reading same partition: each sees full order.
	//   → Manual assignment (no group): you control everything.

	fmt.Println("  Kafka guarantees ordering ONLY within a single partition")
	fmt.Println("  No cross-partition ordering (use timestamps if needed)")
	fmt.Println("  Key → partition mapping gives per-entity ordering")
	fmt.Println()
}

// =============================================================================
// PART 4: HOT PARTITIONS — Detection and solutions
// =============================================================================
func hotPartitions() {
	fmt.Println("--- HOT PARTITIONS ---")

	// A hot partition receives disproportionately more data than others.
	// This causes: one consumer falls behind, one broker disk fills faster,
	// one broker has higher CPU, tail latency increases.
	//
	// ROOT CAUSES:
	// ────────────
	// 1. Low cardinality key: country=US gets 40% of traffic
	// 2. Power law distribution: top 1% of users generate 50% of events
	// 3. Default partitioner + unlucky hash: murmur2 collision on popular keys
	// 4. Time-based keys: all data for "current hour" goes to one partition
	//
	// DETECTION:
	// ──────────
	// Monitor per-partition metrics:
	//   kafka.log:type=Log,name=Size,topic=X,partition=Y
	//   kafka.server:type=BrokerTopicMetrics,name=MessagesInPerSec (per partition)
	//   Consumer lag per partition (should be roughly even)
	//
	// SOLUTIONS:
	//
	// 1. SALTED KEYS
	// ───────────────
	// Add a random suffix to spread load:
	//   Original key: "user_hot_123"
	//   Salted key: "user_hot_123_" + random(0..9)
	//   → Same user splits across 10 partitions
	//   → Loses ordering for that user (tradeoff!)
	//   → Good when ordering isn't needed (metrics, logs)
	//
	// 2. CUSTOM PARTITIONER
	// ─────────────────────
	// Route specific hot keys to a dedicated set of partitions:
	//   if key in hotKeys → hash(key) % 10 + OFFSET
	//   else → defaultPartition(key)
	//
	// 3. SEPARATE TOPIC
	// ─────────────────
	// Route hot entities to a dedicated high-partition topic:
	//   hot_users → topic "events-premium" (100 partitions)
	//   everyone else → topic "events" (20 partitions)
	//
	// 4. COMPOSITE KEY REDESIGN
	// ─────────────────────────
	// If key is "country", add sub-entity:
	//   "US:session_abc123" → distributes US traffic across partitions
	//   Still maintains per-session ordering within US

	fmt.Println("  Hot partition: one partition gets disproportionate load")
	fmt.Println("  Solutions: salted keys, custom partitioner, separate topic")
	fmt.Println("  Monitor: per-partition size/rate metrics, consumer lag skew")
	fmt.Println()
}

// =============================================================================
// PART 5: PARTITION EXPANSION — Can't shrink, grow with care
// =============================================================================
func partitionExpansion() {
	fmt.Println("--- PARTITION EXPANSION ---")

	// YOU CAN ADD PARTITIONS. YOU CAN NEVER REMOVE THEM.
	//
	// WHAT HAPPENS WHEN YOU ADD PARTITIONS:
	// ─────────────────────────────────────
	// 1. Old data stays in old partitions (not redistributed)
	// 2. New records with keys are re-hashed to new+old partitions
	// 3. Same key may NOW go to a DIFFERENT partition
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  BEFORE: 4 partitions                                        │
	// │  murmur2("user_123") % 4 = 2 → Partition 2                  │
	// │                                                              │
	// │  AFTER: 8 partitions (added 4)                               │
	// │  murmur2("user_123") % 8 = 6 → Partition 6  ← DIFFERENT!    │
	// │                                                              │
	// │  Old events for user_123: Partition 2                         │
	// │  New events for user_123: Partition 6                         │
	// │  Ordering for user_123: BROKEN across the boundary            │
	// └──────────────────────────────────────────────────────────────┘
	//
	// SAFE TO EXPAND:
	// ───────────────
	// - Topics with null keys (no key-based routing to break)
	// - Topics where consumers handle reprocessing (e.g., compacted topics
	//   where consumers rebuild state from scratch)
	// - Topics where key→partition mapping doesn't matter
	//
	// DANGEROUS TO EXPAND:
	// ────────────────────
	// - Topics with keyed data where ordering matters
	// - Topics backing Kafka Streams state stores (breaks changelog mapping)
	// - Topics where consumers cache partition→entity assignments
	//
	// MIGRATION STRATEGY FOR KEYED TOPICS:
	// ────────────────────────────────────
	// 1. Create a NEW topic with desired partition count
	// 2. Dual-write: send to both old and new topic
	// 3. Migrate consumers from old to new topic
	// 4. Stop writing to old topic
	// 5. Decommission old topic after retention expires
	//
	// This is the ONLY safe way to repartition keyed data.

	fmt.Println("  Can only ADD partitions, never remove")
	fmt.Println("  Adding partitions changes key→partition mapping")
	fmt.Println("  Safe for null-key topics. Dangerous for keyed data.")
	fmt.Println("  Keyed repartitioning: create new topic + migrate consumers")
	fmt.Println()
}

// =============================================================================
// PART 6: CUSTOM PARTITIONERS
// =============================================================================
func customPartitioners() {
	fmt.Println("--- CUSTOM PARTITIONERS ---")

	// WHEN TO WRITE A CUSTOM PARTITIONER:
	// ────────────────────────────────────
	// 1. Geographic routing: region-based partition assignment
	// 2. Priority: premium customers → dedicated partitions
	// 3. Hot key handling: spread known hot keys across more partitions
	// 4. Co-partitioning: ensure related topics use same partition logic
	//
	// EXAMPLE: Priority-based partitioner
	//
	// func partition(key, value, numPartitions) int {
	//     if isPremiumUser(key) {
	//         // Premium users: first 20% of partitions (dedicated resources)
	//         premiumPartitions := numPartitions / 5
	//         return hash(key) % premiumPartitions
	//     }
	//     // Regular users: remaining 80% of partitions
	//     regularStart := numPartitions / 5
	//     regularCount := numPartitions - regularStart
	//     return regularStart + (hash(key) % regularCount)
	// }
	//
	// EXAMPLE: Geo-aware partitioner
	//
	// func partition(key, value, numPartitions) int {
	//     region := extractRegion(key)  // "us-east", "eu-west", etc.
	//     switch region {
	//     case "us-east": return hash(key) % 10        // partitions 0-9
	//     case "us-west": return 10 + hash(key) % 10   // partitions 10-19
	//     case "eu-west": return 20 + hash(key) % 10   // partitions 20-29
	//     default:        return 30 + hash(key) % 20   // partitions 30-49
	//     }
	// }
	//
	// RULES FOR CUSTOM PARTITIONERS:
	// ──────────────────────────────
	// 1. DETERMINISTIC: same key → same partition (always!)
	// 2. EVEN: try to distribute evenly within each bucket
	// 3. FAST: called on EVERY record. No I/O, no allocations.
	// 4. STABLE: don't change logic unless you plan a migration

	fmt.Println("  Custom partitioners for: geo-routing, priority, hot-key spreading")
	fmt.Println("  Rules: deterministic, even distribution, fast, stable logic")
	fmt.Println("  Used when default murmur2 hash doesn't meet business needs")
	fmt.Println()
}
















































































































































































































































































































































































































}	fmt.Println("  Changing partitioner logic = same as adding partitions (key remap)")	fmt.Println("  Test distribution: CV < 10%, no partition > 2x mean")	fmt.Println("  Custom partitioners for: geo-affinity, priority, load balancing")	// 4. Check: no partition has > 2x mean traffic	//    (coefficient of variation < 10% = good distribution)	// 3. Check: stdev(partition_counts) / mean(partition_counts) < 0.1	// 2. Run keys through the partitioner	// 1. Generate realistic key distribution from production logs	// ─────────────────────	// TESTING PARTITIONERS:	//	// This has the same consequences as adding partitions.	// the same key may go to a DIFFERENT partition.	// IMPORTANT: If you change the partitioner or its logic,	//	// }	//     return normalStart + (hash(key) % normalRange)	//     normalRange := numPartitions - normalStart	//     normalStart := numPartitions / 4	//     // Normal keys: hash to remaining partitions	//	//     }	//         return hash(key) % hotRange     // partitions 0..hotRange-1	//         hotRange := numPartitions / 4  // reserve 25% for hot keys	//         // Spread hot keys across a dedicated range	//     if isHotKey(key) {	//     // Check if this is a known hot key	//	//     }	//         return stickyPartition.getAndMaybeSwitch(numPartitions)	//         // Sticky: choose a random partition, stick to it	//     if key == nil {	// func partition(key []byte, numPartitions int) int {	//	// EXAMPLE: Weight-aware partitioner (Go pseudo-code)	//	// 4. Compatibility: match another system's partitioning scheme	// 3. Load balancing: aware of partition load, routes away from hot ones	// 2. Priority routing: high-priority records to dedicated partitions	// 1. Geo-affinity: route records to partitions near specific data centers	// ──────────	// USE CASES:	//	// When the default murmur2 hash isn't enough, implement a custom partitioner.	fmt.Println("--- CUSTOM PARTITIONERS ---")func customPartitioners() {// =============================================================================// PART 6: CUSTOM PARTITIONERS// =============================================================================}	fmt.Println()	fmt.Println("  Plan partition count for 2-3 years of growth")	fmt.Println("  Adding partitions BREAKS key→partition mapping!")	fmt.Println("  Can add partitions but not remove them")	//    - After retention period: all data follows new routing	//    - Gradually old data expires	//    - Expand partitions	//    - Drain old data	// 4. If you MUST expand: accept the key-remapping and plan for it	// 3. For keyed topics: plan partition count for 2-3 years	// 2. Add 2x headroom for growth	// 1. Start with enough partitions (use the formula from Part 1)	// ───────────────────	// THE RIGHT APPROACH:	//	// WORKAROUND: Create a new topic with fewer partitions, migrate data.	//	// Removing a partition would lose those offsets and the data in it.	// Why? Consumers have committed offsets for each partition.	// IMPOSSIBLE in Kafka. You cannot reduce partition count.	// ────────────────────	// REMOVING PARTITIONS:	//	// - Consumer group needs rebalance to pick up new partitions	// - Compacted topics: same key now has data in TWO partitions	// - Ordering per key is BROKEN for the transition period	// CONSEQUENCES:	//	//    → Same key now routes to a DIFFERENT partition!	// 4. murmur2(key) % 20 ≠ murmur2(key) % 10 for many keys	// 3. EXISTING records stay in their original partitions (no data migration!)	// 2. NEW records are distributed across all partitions (including new ones)	// 1. New partitions are created on brokers (empty, offset starts at 0)	// What happens:	//	// kafka-topics.sh --alter --topic my-topic --partitions 20	// ──────────────────	// ADDING PARTITIONS:	fmt.Println("--- PARTITION EXPANSION ---")func partitionExpansion() {// =============================================================================// PART 5: PARTITION EXPANSION — Why you can't shrink// =============================================================================}	fmt.Println()	fmt.Println("  Fix: higher-cardinality key, salted key, custom partitioner")	fmt.Println("  Detect: monitor bytes-in-rate per partition")	fmt.Println("  Hot partition = one partition gets disproportionate traffic")	//    Over-provision the broker hosting that partition.	//    If 40% of events come from one entity, that's reality.	// 5. ACCEPT IT: Some domains have inherently skewed distributions	//	//    "events-other" topic with 10 partitions (by country)	//    "events-us" topic with 30 partitions (sub-partitioned by user_id)	// 4. SEPARATE TOPIC: Put the hot entity in its own topic	//	//    else: partition = default_hash(key)	//    if key in hotKeys: partition = special_range(key)	// 3. CUSTOM PARTITIONER: Route hot keys to a dedicated set of partitions	//	//    Tradeoff: records for "US" are now in 10 partitions → need aggregation	//    Spreads "US" across 10 different partitions	//    key = "US" + "-" + (hash(value) % 10)	// 2. SALTED KEY: Add a random suffix to spread hot keys	//	//    Tradeoff: lose country-level ordering (if you even need it)	//    Bad: key=country → Good: key=country+"-"+user_id	// 1. CHANGE THE KEY: Use higher-cardinality key	// ──────────	// SOLUTIONS:	//	// If max(partition rate) > 3x mean(partition rate) → you have a hot partition.	//   kafka.server:type=BrokerTopicMetrics,name=BytesInPerSec,topic=X,partition=Y	// Monitor bytes-in-rate per partition:	// ──────────	// DETECTION:	//	// 4. Partition 7 fills up faster → more disk I/O on that broker	// 3. Other consumers are idle → wasted resources	// 2. Consumer for partition 7 gets 40% of work → consumer overload	// 1. Partition 7's leader broker gets 40% of produce traffic → broker overload	// PROBLEMS:	//	//   Other 27 partitions: 37% combined	//   Partition 23 (maps to "IN"): 15% of traffic	//   Partition 12 (maps to "GB"): 8% of traffic	//   Partition 7 (maps to "US"): 40% of all traffic	// Topic with key=country_code, 30 partitions:	// EXAMPLE:	//	// A hot partition receives disproportionately more data than others.	fmt.Println("--- HOT PARTITIONS ---")func hotPartitions() {// =============================================================================// PART 4: HOT PARTITIONS — The silent performance killer// =============================================================================}	fmt.Println()	fmt.Println("  After partition expansion: same key may go to different partition!")	fmt.Println("  Across partitions: no order (fundamental)")	fmt.Println("  Within partition: total order (always)")	// └──────────────────────────────────────────────────────────────┘	// │                     │ ordered     │                          │	// │  With transactions  │ Atomic, not │ All succeed or all fail  │	// │                     │  partition) │                          │	// │  With idempotence   │ YES (per    │ Preserves order on retry │	// │  Across partitions  │ NO          │ Fundamental limitation   │	// │    partition expand │             │ count changes            │	// │  Same key, after    │ NO!         │ Key remapped if partition│	// │    partition         │             │                          │	// │  Same key, same     │ YES         │ Same key → same partition│	// │  Within partition   │ YES         │ Append order = read order│	// │  ───────────────────┼─────────────┼────────────────────────  │	// │  Scope              │ Guaranteed? │ Requirement              │	// │                                                              │	// │  ORDERING GUARANTEE MATRIX:                                   │	// ┌──────────────────────────────────────────────────────────────┐	//	// Transaction commits are ordered within each partition.	// They do NOT guarantee ORDER across partitions.	// Transactions guarantee ATOMICITY (all-or-nothing across partitions).	// TRAP 3: "I used transactions for ordering"	//	// Or: use event timestamps and sort in consumers (eventual ordering).	// You CANNOT have total ordering AND high throughput. Pick one.	// Then you need 1 partition. But max throughput = 1 partition throughput.	// TRAP 2: "I need total ordering across all events"	//	// User's records are now SPLIT across two partitions. No ordering.	// If you ADD partitions: murmur2("user-123") % newNumPartitions ≠ old!	// Records with key="user-123" go to murmur2("user-123") % numPartitions.	// TRAP 1: "I thought same user = same partition"	//	// ORDERING TRAPS:	//	//    Without idempotence: out-of-order possible with retries.	//    Per-partition order preserved even with retries.	// 3. With IDEMPOTENT producer + max.in.flight ≤ 5:	//	//    Consumer might see B before A, A before B, or simultaneously.	//    Record A in partition 0, record B in partition 1	// 2. Across PARTITIONS: NO ordering guarantee	//	//    Consumer will see A before B. Always.	//    If record A is written before record B → A.offset < B.offset	// 1. Within a SINGLE PARTITION: total order guaranteed	// ──────────────────────	// WHAT KAFKA GUARANTEES:	fmt.Println("--- ORDERING GUARANTEES ---")func orderingGuarantees() {// =============================================================================// PART 3: ORDERING GUARANTEES// =============================================================================}	fmt.Println()	fmt.Println("  No ordering needed → null key (sticky partitioner)")	fmt.Println("  High cardinality keys → even distribution")	fmt.Println("  Key = the entity you need ordering for")	// └──────────────────────────────────────────────────────────────┘	// │  └── NO → Consider composite key or accept hot partitions   │	// │  ├── YES (>10x partition count) → Good, even distribution   │	// │  Is the key high cardinality?                                │	// │                                                              │	// │      └── Per entity pair → key = entity1_id + entity2_id    │	// │      ├── Per device → key = device_id                        │	// │      ├── Per order → key = order_id                          │	// │      ├── Per user → key = user_id                            │	// │  └── YES → What entity needs ordering?                       │	// │  ├── NO → null key (best distribution, sticky partitioner)   │	// │  Do you need ordering?                                       │	// │                                                              │	// │  KEY DESIGN DECISION TREE:                                    │	// ┌──────────────────────────────────────────────────────────────┐	//	// Same as null key but without the distribution benefits.	// This is effectively UNIQUE per event → no ordering benefit.	//   key = tenantId + "-" + userId + "-" + timestamp	// DON'T over-specify keys:	//	// This ensures all events for a tenant+user combo go to the same partition.	//   key = tenantId + "-" + userId	// If you need ordering by (tenant + user):	// ───────────────	// COMPOSITE KEYS:	//	// "I don't need ordering" → null key (sticky partitioner, best distribution)	// "I need all events for ORDER-456 in order" → key = order_id	// "I need all events for USER-123 in order" → key = user_id	//	// Key = the ENTITY you need ordering for.	// ─────────────────────────	// THE KEY DESIGN PRINCIPLE:	//	// - boolean: true/false → only 2 partitions used	// - date: all today's records in one partition	// - status: "active"/"inactive" → 2 partitions get all data	// - country: USA gets 40% of traffic → hot partition	// ─────────────────────────────────────────────────	// BAD KEYS (low cardinality, skewed distribution):	//	// - device_id: unique per device → great distribution	// - session_id: unique per session → perfect distribution	// - order_id: unique per order → perfect distribution	// - user_id: millions of users → great distribution	// ─────────────────────────────────────────────────	// GOOD KEYS (high cardinality, even distribution):	//	//   partition = murmur2(keyBytes) % numPartitions	// The partition key determines WHICH partition a record goes to:	fmt.Println("--- PARTITION KEY DESIGN ---")func keyDesign() {// =============================================================================// PART 2: PARTITION KEY DESIGN// =============================================================================}	fmt.Println()	fmt.Println("  Costs: memory, file descriptors, failover time, rebalance time")	fmt.Println("  Sweet spot: 12-30 for most topics. Rarely > 100.")	fmt.Println("  Formula: max(throughput_need/partition_throughput, max_consumers)")	// └──────────────────────────────────────────────────────────────┘	// │  - KRaft pushes this to millions                              │	// │  - ≤ 200,000 partitions per cluster                           │	// │  - ≤ 4,000 partitions per broker                              │	// │  RULE OF THUMB (Confluent recommendation):                   │	// │                                                              │	// │  Consider: multiple topics, tiered storage, stream processing.│	// │  Almost always wrong. Re-think your architecture.            │	// │  1000+ partitions per topic:                                  │	// │                                                              │	// │  Monitor broker memory and controller overhead.              │	// │  Very high throughput. LinkedIn/Uber scale.                   │	// │  100+ partitions:                                             │	// │                                                              │	// │  High throughput topics. Plan for consumer scaling.           │	// │  50-100 partitions:                                           │	// │                                                              │	// │  Good balance of parallelism and overhead.                   │	// │  Medium throughput. Covers most production use cases.        │	// │  12-30 partitions:                                            │	// │                                                              │	// │  Low throughput topics. Good for configs, commands.           │	// │  ≤ 6 partitions:                                             │	// │                                                              │	// │  PARTITION COUNT GUIDELINES:                                   │	// ┌──────────────────────────────────────────────────────────────┐	//	//   - Terrible compression and throughput	//   - 100 records/sec across 100 partitions = 1 record per batch per partition	//   - If records spread across many partitions, each batch is small	// COST 5: Producer batching efficiency	//	//   - More partitions = longer rebalances	//   - More partitions = more fetch requests = more work per broker	//   - Replication fetch requests are per-partition	// COST 4: End-to-end latency	//	//   - KRaft improves this significantly (see Lesson 14)	//   - 100,000 partitions: potential MINUTES of unavailability	//   - 10,000 partitions on one broker: ~3-10 seconds failover	//   - Controller must elect leaders for all partitions on a dead broker	// COST 3: Controller failover time	//	//   - Check: ulimit -n (should be ≥ 100,000)	//   - 10,000 partitions = 30,000 file descriptors	//   - Each partition has ~3 open files (log, index, timeindex)	// COST 2: File descriptors	//	//   - This is BROKER memory — all partitions on that broker	//   - 10,000 partitions × 150 KB = 1.5 GB per broker	//   - Segment buffers, index mmap, replica state	// COST 1: Memory per partition (~150 KB on broker)	//	// ─────────────────────────────────	// WHY NOT THOUSANDS OF PARTITIONS?	//	// Growth headroom: 2x = 40 partitions	// Partitions needed: 500/25 = 20 partitions minimum	// Need: 500 MB/s throughput, consumer processes at 25 MB/s	// EXAMPLE:	//	// - consumer_count: max number of consumers you'll ever need	//   Consumer side: 10-50 MB/s per partition (depends on processing complexity)	//   Producer side: ~10-100 MB/s per partition (depends on record size, replication)	// - partition_throughput: what a single partition can handle	// - throughput_need: your target messages/sec or MB/sec	// WHERE:	//	// Partitions = max(throughput_need / partition_throughput, consumer_count)	// ────────	// FORMULA:	fmt.Println("--- HOW MANY PARTITIONS ---")func partitionCount() {// =============================================================================// PART 1: HOW MANY PARTITIONS — The most asked question in Kafka// =============================================================================}	customPartitioners()	partitionExpansion()	hotPartitions()	orderingGuarantees()	keyDesign()	partitionCount()	fmt.Println()	fmt.Println("=== PARTITIONING & ORDERING ===")func main() {import "fmt"package main// =============================================================================//// But MORE IS NOT ALWAYS BETTER. There are real costs to over-partitioning.// - Storage (partitions distribute across brokers)// - Read throughput (each partition maps to one consumer)// - Write throughput (each partition has an independent leader)// Partitions are Kafka's UNIT OF PARALLELISM. Everything scales with partitions:// THE KEY INSIGHT://// - How partitions map to consumer parallelism// - The partition expansion problem (and why you can't shrink)// - Custom partitioners: when the default isn't enough// - Ordering guarantees: what's guaranteed and what's NOT// - Partition key design: the art of avoiding hot partitions// - How to choose the right number of partitions// WHAT YOU'LL LEARN://// =============================================================================// LESSON 5.1: PARTITIONING & ORDERING — The Foundation of Kafka Scale// =============================================================================