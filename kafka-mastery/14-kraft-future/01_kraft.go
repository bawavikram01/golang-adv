//go:build ignore
// =============================================================================
// LESSON 14.1: KRaft & THE FUTURE — Post-ZooKeeper Kafka
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Why ZooKeeper had to go (the 6 problems)
// - KRaft architecture: dedicated vs combined controller mode
// - Metadata as a log: __cluster_metadata topic and snapshots
// - Controller quorum: Raft protocol, leader election, performance
// - ZooKeeper to KRaft migration path
// - Kafka's future: tiered storage, share groups, Kafka 4.0+
//
// THE KEY INSIGHT:
// KRaft is not just "Kafka without ZooKeeper." It's a fundamental redesign
// of how Kafka manages metadata. By treating metadata as a Kafka topic
// (replicated via Raft), the entire cluster becomes self-contained.
// This unlocks millions of partitions per cluster.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== KRAFT & THE FUTURE ===")
	fmt.Println()

	whyRemoveZooKeeper()
	kraftArchitecture()
	metadataAsLog()
	controllerQuorum()
	zkToKraftMigration()
	kafkaFuture()
}

// =============================================================================
// PART 1: WHY REMOVE ZOOKEEPER — The 6 problems
// =============================================================================
func whyRemoveZooKeeper() {
	fmt.Println("--- WHY REMOVE ZOOKEEPER ---")

	// ZooKeeper served Kafka well for years, but it created 6 fundamental problems:
	//
	// PROBLEM 1: OPERATIONAL COMPLEXITY
	// ──────────────────────────────────
	// Running Kafka = running TWO distributed systems (Kafka + ZooKeeper).
	// Different configs, different monitoring, different failure modes.
	// Two things to upgrade, two things to secure, two things to backup.
	// Doubles the operational burden.
	//
	// PROBLEM 2: SCALABILITY CEILING
	// ──────────────────────────────
	// ZooKeeper stores ALL metadata in memory (znodes).
	// With 200K+ partitions, ZK becomes the bottleneck:
	// - Slow leader elections (iterating all partitions)
	// - Slow metadata operations (creating/deleting topics)
	// - ZK watch storms (all brokers watching same znodes)
	//
	// Practical limit: ~200,000 partitions per cluster with ZK.
	// KRaft target: millions of partitions.
	//
	// PROBLEM 3: METADATA INCONSISTENCY
	// ──────────────────────────────────
	// Metadata was split between ZK (source of truth) and brokers (cache).
	// The controller read from ZK and pushed updates to brokers.
	// Race conditions: broker's local cache could be stale.
	// This caused: wrong leader, stale partition assignments, ghost brokers.
	//
	// PROBLEM 4: SLOW RECOVERY
	// ────────────────────────
	// When the controller failed, the new controller had to:
	// 1. Read ALL metadata from ZK (O(n) in partition count)
	// 2. Rebuild its in-memory state
	// 3. Send full metadata updates to all brokers
	// With 100K partitions: this took MINUTES.
	// During this time: no leader elections, no topic creation, no rebalancing.
	//
	// PROBLEM 5: SECURITY SURFACE
	// ───────────────────────────
	// ZooKeeper has its own authentication (SASL) and authorization (ACLs).
	// Completely separate from Kafka's security model.
	// Securing the ZK-Kafka connection is another gap to close.
	//
	// PROBLEM 6: DEVELOPMENT VELOCITY
	// ────────────────────────────────
	// Adding new features to Kafka often required changes to how metadata
	// was stored in ZK. ZK's data model (tree of znodes) didn't always
	// fit Kafka's needs, leading to workarounds and tech debt.

	fmt.Println("  6 problems: ops complexity, scale ceiling, metadata inconsistency,")
	fmt.Println("  slow recovery, security surface, development velocity")
	fmt.Println("  Biggest win: from ~200K partition limit to millions")
	fmt.Println()
}

// =============================================================================
// PART 2: KRAFT ARCHITECTURE
// =============================================================================
func kraftArchitecture() {
	fmt.Println("--- KRAFT ARCHITECTURE ---")

	// KRaft mode: Kafka uses an internal Raft-based consensus protocol
	// to manage metadata. No external dependency.
	//
	// TWO DEPLOYMENT MODES:
	//
	// MODE 1: DEDICATED CONTROLLERS (recommended for production)
	// ──────────────────────────────────────────────────────────
	// ┌──────────┐  ┌──────────┐  ┌──────────┐
	// │Controller│  │Controller│  │Controller│  (3 or 5 dedicated nodes)
	// │  Node 1  │  │  Node 2  │  │  Node 3  │  Only handle metadata
	// └────┬─────┘  └────┬─────┘  └────┬─────┘
	//      │             │             │
	// ┌────▼─────┐  ┌────▼─────┐  ┌────▼─────┐  ┌──────────┐
	// │ Broker 1 │  │ Broker 2 │  │ Broker 3 │  │ Broker N │
	// │ (data)   │  │ (data)   │  │ (data)   │  │ (data)   │
	// └──────────┘  └──────────┘  └──────────┘  └──────────┘
	//
	// Controllers are lightweight (no data serving).
	// Scale brokers independently of controllers.
	// Controller failure doesn't affect data serving.
	//
	// MODE 2: COMBINED (controllers + brokers on same nodes)
	// ──────────────────────────────────────────────────────
	// ┌───────────────┐  ┌───────────────┐  ┌───────────────┐
	// │  Node 1       │  │  Node 2       │  │  Node 3       │
	// │  Controller   │  │  Controller   │  │  Controller   │
	// │  + Broker     │  │  + Broker     │  │  + Broker     │
	// └───────────────┘  └───────────────┘  └───────────────┘
	//
	// Simpler deployment for small clusters (3-5 nodes).
	// Saves hardware. Good for dev/test or small production.
	// Not recommended for large clusters (controller work competes with data I/O).
	//
	// CONFIGURATION:
	// ──────────────
	// process.roles=controller  (dedicated controller)
	// process.roles=broker      (dedicated broker)
	// process.roles=controller,broker  (combined mode)
	//
	// controller.quorum.voters=1@controller1:9093,2@controller2:9093,3@controller3:9093
	// node.id=1  (unique per node)

	fmt.Println("  Dedicated controllers: separate from brokers (production)")
	fmt.Println("  Combined mode: controller+broker on same node (small clusters)")
	fmt.Println("  3 or 5 controllers recommended (Raft quorum needs odd count)")
	fmt.Println()
}

// =============================================================================
// PART 3: METADATA AS A LOG
// =============================================================================
func metadataAsLog() {
	fmt.Println("--- METADATA AS A LOG ---")

	// In KRaft, all cluster metadata is stored in a special internal topic:
	//   __cluster_metadata (single partition, replicated across controllers)
	//
	// This topic contains records like:
	// - TopicRecord: topic created with UUID
	// - PartitionRecord: partition assigned to brokers
	// - BrokerRegistrationRecord: broker joined the cluster
	// - ConfigRecord: configuration changed
	// - PartitionChangeRecord: leadership changed
	// - RemoveTopicRecord: topic deleted
	//
	// WHY THIS IS BRILLIANT:
	// ──────────────────────
	// 1. ATOMIC METADATA UPDATES
	//    Multiple metadata changes can be batched into one log append.
	//    "Create topic with 10 partitions" = 1 TopicRecord + 10 PartitionRecords
	//    All or nothing, no partial updates.
	//
	// 2. EVENT-SOURCED METADATA
	//    Brokers "replay" the metadata log to build their local state.
	//    Like a Kafka consumer, each broker tracks its metadata offset.
	//    On startup: replay from last snapshot + any new records.
	//
	// 3. NO METADATA INCONSISTENCY
	//    There's ONE source of truth: the metadata log.
	//    Every broker reads the same log → eventually consistent state.
	//    No more "ZK says one thing, broker says another."
	//
	// METADATA SNAPSHOTS:
	// ────────────────────
	// The metadata log grows forever if not compacted.
	// KRaft periodically creates SNAPSHOTS: compressed state at a point in time.
	// Old log records before the snapshot can be deleted.
	//
	// New broker / broker restarting after long downtime:
	// 1. Fetch latest snapshot from the active controller
	// 2. Apply snapshot (instant state rebuild)
	// 3. Fetch and apply any records after the snapshot
	//
	// This is MUCH faster than the old ZK approach (read all znodes one by one).

	fmt.Println("  __cluster_metadata: single topic containing ALL cluster metadata")
	fmt.Println("  Event-sourced: brokers replay the log to build local state")
	fmt.Println("  Snapshots: compressed state for fast broker startup")
	fmt.Println()
}

// =============================================================================
// PART 4: CONTROLLER QUORUM — Raft protocol
// =============================================================================
func controllerQuorum() {
	fmt.Println("--- CONTROLLER QUORUM ---")

	// The controller quorum uses a Raft-based protocol (KRaft = Kafka + Raft).
	// NOT the exact Raft paper implementation — adapted for Kafka's needs.
	//
	// KEY PROPERTIES:
	// ────────────────
	// 1. LEADER ELECTION
	//    - One controller is the ACTIVE controller (leader)
	//    - Others are followers (hot standby)
	//    - If the leader fails: followers elect a new leader
	//    - Election time: ~200ms (MUCH faster than ZK-based controller failover)
	//
	// 2. LOG REPLICATION
	//    - Active controller appends metadata records to the log
	//    - Followers replicate the log (like Raft followers)
	//    - Majority must acknowledge before a record is committed
	//    - 3 controllers: majority = 2 (tolerates 1 failure)
	//    - 5 controllers: majority = 3 (tolerates 2 failures)
	//
	// 3. BROKERS AS OBSERVERS
	//    - Brokers are NOT part of the Raft quorum
	//    - They FETCH metadata from the active controller (pull model)
	//    - Similar to how followers fetch data from leaders
	//    - Broker tracks: "I've seen metadata up to offset X"
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  KRAFT METADATA FLOW:                                         │
	// │                                                              │
	// │  Active Controller (Leader)                                   │
	// │  ┌──────────────────────────┐                                │
	// │  │ __cluster_metadata log   │                                │
	// │  │ [0][1][2][3][4][5][6][7] │                                │
	// │  └──────────┬───────────────┘                                │
	// │             │ Raft replication (push)                         │
	// │  ┌──────────▼──────────┐                                     │
	// │  │ Follower Controllers │                                    │
	// │  │ (hot standby)       │                                     │
	// │  └─────────────────────┘                                     │
	// │             │ Metadata fetch (pull)                           │
	// │  ┌──────────▼──────────┐                                     │
	// │  │ Broker 1: offset=5  │  "I've seen up to record 5"        │
	// │  │ Broker 2: offset=7  │  "I'm fully caught up"             │
	// │  │ Broker 3: offset=6  │  "One record behind"               │
	// │  └─────────────────────┘                                     │
	// └──────────────────────────────────────────────────────────────┘
	//
	// PERFORMANCE:
	// ────────────
	// Controller failover: ~200ms (Raft election)
	// vs ZK-based: 30 seconds - several minutes (reading all ZK state)
	//
	// Metadata propagation: milliseconds (brokers fetch continuously)
	// vs ZK-based: seconds (controller pushes to all brokers sequentially)
	//
	// Partition count: millions (log is compact, indexed)
	// vs ZK-based: ~200K (memory-limited znode tree)

	fmt.Println("  Raft-based leader election: ~200ms (vs minutes with ZK)")
	fmt.Println("  Brokers fetch metadata (pull model, like consumers)")
	fmt.Println("  3 controllers for most clusters, 5 for large/critical")
	fmt.Println()
}

// =============================================================================
// PART 5: ZK TO KRAFT MIGRATION
// =============================================================================
func zkToKraftMigration() {
	fmt.Println("--- ZK TO KRAFT MIGRATION ---")

	// Migration from ZooKeeper mode to KRaft mode is supported since Kafka 3.4+
	// (fully GA in Kafka 3.6+).
	//
	// MIGRATION PATH:
	// ────────────────
	//
	// 1. PREREQUISITES
	//    - Upgrade to Kafka 3.6+ (or the latest 3.x)
	//    - All brokers running and healthy
	//    - Ensure all inter.broker.protocol.version and log.message.format.version
	//      are set to the current version
	//
	// 2. DEPLOY KRaft CONTROLLERS
	//    - Set up 3 (or 5) KRaft controller nodes
	//    - Configure them with:
	//      process.roles=controller
	//      controller.quorum.voters=...
	//      node.id=... (unique IDs, different from broker IDs)
	//    - Start controllers (they form an empty quorum)
	//
	// 3. RUN MIGRATION COMMAND
	//    bin/kafka-metadata.sh migrate \
	//      --zookeeper zk1:2181 \
	//      --controller controller1:9093
	//
	//    This copies ALL metadata from ZK into the KRaft metadata log:
	//    - Topic configurations
	//    - Partition assignments
	//    - ACLs
	//    - Client quotas
	//    - Broker configs
	//
	// 4. DUAL-WRITE MODE
	//    After migration, the cluster enters dual-write mode:
	//    - KRaft controller is the source of truth
	//    - Changes are ALSO written to ZK (for rollback safety)
	//    - Brokers gradually switch to fetching from KRaft controller
	//
	// 5. FINALIZE (no rollback after this!)
	//    bin/kafka-metadata.sh finalize-migration
	//    - Stops dual-writing to ZK
	//    - ZK connection severed
	//    - Cluster is fully KRaft
	//
	// 6. DECOMMISSION ZOOKEEPER
	//    - Stop ZK nodes
	//    - Clean up ZK data
	//    - Remove ZK configs from broker properties
	//
	// ROLLBACK:
	// ─────────
	// Before finalize: restart brokers with ZK config → back to ZK mode.
	// After finalize: NO ROLLBACK. You're committed to KRaft.
	// → Test thoroughly on staging before finalizing!

	fmt.Println("  Migration: deploy controllers → migrate metadata → dual-write → finalize")
	fmt.Println("  Dual-write mode allows rollback before finalization")
	fmt.Println("  After finalize: no rollback to ZK (test on staging first!)")
	fmt.Println()
}

// =============================================================================
// PART 6: KAFKA'S FUTURE
// =============================================================================
func kafkaFuture() {
	fmt.Println("--- KAFKA'S FUTURE ---")

	// TIERED STORAGE (KIP-405) — Available since Kafka 3.6
	// ─────────────────────────────────────────────────────
	// Problem: Kafka stores ALL data on broker disks. Long retention = huge disks.
	//
	// Solution: Two tiers of storage:
	//   Hot tier: local broker disks (recent data, fast access)
	//   Cold tier: object storage like S3, GCS, Azure Blob (old data, cheap)
	//
	// How it works:
	//   - Recent segments stay on local disk
	//   - Old segments automatically uploaded to object storage
	//   - Local copies deleted (freeing disk space)
	//   - Consumer reads old data? Fetched from object storage transparently
	//
	// Impact: Keep 30 days on disk + 1 year in S3, at 1/10th the cost.
	// Run fewer brokers (less disk per broker).
	//
	// SHARE GROUPS (KIP-932) — Coming soon
	// ─────────────────────────────────────
	// Problem: Consumer groups have 1 consumer per partition max.
	// 10 partitions, 20 consumers → 10 consumers are idle.
	//
	// Solution: Share groups allow multiple consumers to share a partition.
	// Records are "locked" to a consumer until acknowledged.
	// Similar to traditional message queues (RabbitMQ, SQS).
	//
	// Use cases:
	// - Job/task queues where partition-level ordering isn't needed
	// - Scaling consumers beyond partition count
	// - Replacing message queues with Kafka (unified platform)
	//
	// KAFKA 4.0+ (upcoming)
	// ─────────────────────
	// - KRaft-only (ZooKeeper support removed entirely)
	// - Tiered storage GA
	// - Share groups
	// - Continued focus on: ease of operations, cloud-native deployment
	// - Queue and stream processing in one platform
	//
	// THE VISION:
	// ───────────
	// Kafka evolves from "distributed log" to "unified event streaming platform":
	// - Streams (traditional Kafka topics)
	// - Queues (share groups)
	// - Storage (tiered storage, infinite retention)
	// - Processing (Kafka Streams, Flink integration)
	// - One platform for all event-driven architectures.

	fmt.Println("  Tiered storage: hot (disk) + cold (S3) = cheap infinite retention")
	fmt.Println("  Share groups: multiple consumers per partition (queue semantics)")
	fmt.Println("  Kafka 4.0: ZK removed, unified streaming + queuing platform")
	fmt.Println()
}


































































































































































































































































































































}	fmt.Println("  Future: log + queue + storage + compute = unified data platform")	fmt.Println("  Share groups: Kafka as a message queue (SQS semantics on Kafka)")	fmt.Println("  Tiered storage: cold data on S3, 10x retention at 1/10 cost")	// It's becoming the central nervous system for ALL data flow in an org.	//	//   - Connect (Kafka Connect, CDC)	//   - Compute (Kafka Streams, ksqlDB)	//   - Storage (tiered storage, infinite retention)	//   - Queue (share groups)	//   - Log (traditional Kafka)	// "unified streaming platform":	// Kafka is evolving from "distributed commit log" to	// ────────────────	// THE BIG PICTURE:	//	// - Improved compaction performance	// - Leader election optimization	// - Client quotas improvements	// - Improved consumer group protocol (KIP-848)	// - JBOD improvements: better disk failure handling	// ────────────────────────	// OTHER UPCOMING FEATURES:	//	//   Some use cases don't need ordering — they need pure parallelism.	//   Like SQS/RabbitMQ semantics, but on Kafka's infrastructure.	//	//   being processed and re-delivered if processing fails.	//   the SAME partition concurrently. Records are "locked" while	//   Multiple consumers in a share group can process records from	// "Share groups" — Kafka as a message queue (not just a log!):	// ───────────────────────	// KAFKA QUEUES (KIP-932):	//	//   After: 90 days retention (S3 is cheap)	//   Before: 7 days retention (disk is expensive)	// IMPACT: 10x retention at 1/10th cost.	//	// Consumer reads: hot data from broker, cold data from object storage.	// Broker keeps recent segments locally, moves old ones to remote.	//	//   Cold data: S3/GCS (cheap, infinite capacity)	//   Hot data: local broker disks (page cache, zero-copy)	// Store cold data on cheap object storage (S3, GCS, HDFS):	// ─────────────────────────	// TIERED STORAGE (KIP-405):	fmt.Println("--- KAFKA'S FUTURE ---")func kafkaFuture() {}	fmt.Println()	fmt.Println("  Kafka 4.0: ZooKeeper completely removed, KRaft only")	fmt.Println("  Migration: add controllers → dual-write → switch → decommission ZK")	// If running ZK mode: migrate to KRaft before upgrading to Kafka 4.x.	// TIMELINE RECOMMENDATION:	//	// - New clusters created in KRaft mode by default	// - KRaft is the ONLY mode	// - ZooKeeper support completely REMOVED	// Kafka 4.0 (2024+):	//	//    No more ZK!	// 5. DECOMMISSION: Shut down ZooKeeper ensemble	//	//    Brokers now read metadata from KRaft controllers only	// 4. SWITCH: Restart brokers without ZK config	//	//    Brokers still talk to ZK	//    Controller writes to BOTH ZK and __cluster_metadata	// 3. DUAL-WRITE MODE:	//	//    Migrates ZK metadata to __cluster_metadata	// 2. MIGRATE: Run kafka-metadata.sh --snapshot	//	//    Configure controller.quorum.voters	// 1. PREPARE: Add KRaft controllers alongside existing brokers	// ─────────────────────────────	// MIGRATION PATH (Kafka 3.4+):	fmt.Println("--- ZOOKEEPER TO KRAFT MIGRATION ---")func migration() {}	fmt.Println()	fmt.Println("  Brokers PULL metadata incrementally (no thundering herd)")	fmt.Println("  vs ZooKeeper: 6-30s session timeout + metadata reload")	fmt.Println("  Raft consensus: 3 or 5 controllers, leader election in ~200ms")	//   fetches only X+1 onwards	//   Incremental: broker already has metadata up to offset X,	//   Each broker fetches at its own pace, no thundering herd	// New (KRaft): Brokers pull metadata from __cluster_metadata log	//	//   Slow for large clusters (thousands of partitions per update)	// Old (ZK): Controller pushes metadata to brokers via UpdateMetadata RPC	// ─────────────────────	// METADATA PROPAGATION:	//	// KRaft: Raft election (~200ms) + no metadata read needed (already local!)	// ZK: session timeout (~6s) + controller reads full metadata (~1-10s)	// ──────────────────────────────────	// WHY THIS IS FASTER THAN ZOOKEEPER:	//	// 6. Committed when majority of controllers have it	// 5. Leader appends to metadata log, replicates to followers	// 4. Majority vote (⌊N/2⌋ + 1) → new leader	// 3. Candidate votes for itself, asks others to vote	// 2. If a follower doesn't hear for election.timeout: starts election	// 1. Leader sends heartbeats to followers	// ────────────────────────────	// RAFT PROTOCOL (simplified):	//	// - If leader dies: Raft election in ~200ms (vs 6-30 seconds with ZK!)	// - Others are followers (replicate the metadata log)	// - ONE is the active controller (Raft leader)	// - Typically 3 controllers (tolerates 1 failure) or 5 (tolerates 2)	//	// KRaft uses a Raft-based consensus protocol for controllers:	fmt.Println("--- CONTROLLER QUORUM ---")func controllerQuorum() {}	fmt.Println()	fmt.Println("  Brokers consume this log to stay current + snapshots for fast catch-up")	fmt.Println("  Metadata = ordered event log (beautifully recursive!)")	fmt.Println("  __cluster_metadata: internal topic with all cluster state")	// └──────────────────────────────────────────────────────────────┘	// │  New broker: load snapshot @50, replay 51-100.                │	// │  [SNAPSHOT at offset 50]                                      │	// │  ...                                                         │	// │              ISR=[1,3])  ← broker 2 fell out of ISR          │	// │  Offset 100: PartitionChangeRecord(topic=uuid123, part=0,   │	// │  ...                                                         │	// │  Offset 4: PartitionRecord(topic=uuid123, partition=1, ...) │	// │            replicas=[1,2,3], leader=1, ISR=[1,2,3])         │	// │  Offset 3: PartitionRecord(topic=uuid123, partition=0,      │	// │  Offset 2: TopicRecord(name="orders", id=uuid123)           │	// │  Offset 1: BrokerRegistrationRecord(brokerId=2, ...)        │	// │  Offset 0: BrokerRegistrationRecord(brokerId=1, ...)        │	// │                                                              │	// │  METADATA LOG EXAMPLE:                                        │	// ┌──────────────────────────────────────────────────────────────┐	//	// Snapshots are stored locally on each controller node.	// instead of replaying the entire log from offset 0.	// New brokers or catching-up brokers can start from the snapshot	// Periodically, the controller takes a snapshot of current metadata state.	// ──────────	// SNAPSHOTS:	//	// - No watches, no sessions, no ephemeral nodes	// - Snapshots reduce the log size (periodic metadata snapshot)	// - Brokers can catch up after downtime (replay the log)	// - Metadata changes are ORDERED (log ordering)	// BENEFITS:	//	// It's literally a Kafka consumer reading a Kafka topic!	// Each broker fetches from __cluster_metadata to stay up-to-date.	// ─────────────────────────	// BROKERS CONSUME THIS LOG:	//	//   - Timestamp	//   - Epoch (controller epoch, like leader epoch)	//   - Offset (position in the metadata log)	// Each record has:	//	//   - ... many more record types	//   - FeatureLevelRecord: feature flags	//   - BrokerRegistrationRecord: broker joined/left	//   - ConfigRecord: configuration change	//   - PartitionChangeRecord: ISR change, leader change	//   - PartitionRecord: partition details (replicas, ISR, leader)	//   - TopicRecord: new topic created	// This topic is a LOG of metadata EVENTS:	//	//   __cluster_metadata (single partition, replicated across controllers)	// In KRaft, cluster metadata is stored in a special internal topic:	fmt.Println("--- METADATA AS LOG ---")func metadataAsLog() {}	fmt.Println()	fmt.Println("  Brokers pull metadata from controllers (no ZooKeeper)")	fmt.Println("  Combined mode (broker+controller) = dev/small clusters")	fmt.Println("  Dedicated controllers (3-5) + broker nodes = production setup")	// listeners = PLAINTEXT://:9092,CONTROLLER://:9093	// controller.listener.names = CONTROLLER	// controller.quorum.voters = 1@host1:9093,2@host2:9093,3@host3:9093	// node.id = 1                          # Unique per node	// process.roles = broker | controller | broker,controller	// ────────────	// KEY CONFIGS:	//	// └──────────────────────────────────────────────────────────────┘	// │  No ZooKeeper in the picture!                                 │	// │  Brokers fetch metadata from controllers (pull model)        │	// │                       │                                       │	// │  └──────┘  └──────┘  │  └──────┘  └──────┘                 │	// │  │  1   │  │  2   │  │  │  3   │  │  4   │                 │	// │  │Broker│  │Broker│  │  │Broker│  │Broker│                 │	// │  ┌──────┐  ┌──────┐  │  ┌──────┐  ┌──────┐                 │	// │                       │                                       │	// │                       │ (metadata log)                        │	// │                       │ Raft consensus                        │	// │        └──────────────┼──────────────┘                        │	// │        │              │              │                        │	// │  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘               │	// │  │  (Leader)  │  │ (Follower) │  │ (Follower) │               │	// │  │Controller-1│  │Controller-2│  │Controller-3│               │	// │  ┌───────────┐  ┌───────────┐  ┌───────────┐               │	// │                                                              │	// │    Simpler but controller workload competes with broker.      │	// │    Each node is BOTH broker and controller.                    │	// │    process.roles = broker,controller                          │	// │  All nodes: 3+                                                │	// │  ──────────────────────────────────────────────               │	// │  OPTION 2: Combined Mode (small clusters, dev/test)          │	// │                                                              │	// │    Handle produce, consume, replication. No metadata votes.   │	// │    process.roles = broker                                     │	// │  Broker nodes: N (as many as needed)                          │	// │                                                              │	// │    Only handle metadata. No produce/consume.                  │	// │    process.roles = controller                                 │	// │  Controller nodes: 3 (or 5 for large clusters)               │	// │  ──────────────────────────────                               │	// │  OPTION 1: Dedicated Controllers (recommended for production)│	// │                                                              │	// │  KRAFT CLUSTER TOPOLOGY:                                      │	// ┌──────────────────────────────────────────────────────────────┐	fmt.Println("--- KRAFT ARCHITECTURE ---")func kraftArchitecture() {}	fmt.Println()	fmt.Println("  KRaft: single system, metadata in log, millisecond failover")	fmt.Println("  ZooKeeper: separate system, metadata bottleneck, slow failover")	// - Event-driven metadata (no watches needed)	// - Single security model	// - No split-brain (Raft guarantees single leader)	// - Controller failover in milliseconds (Raft election)	// - Metadata in Kafka's own log (scales to millions of partitions)	// - Single system to operate	// KRaft ELIMINATES ALL OF THESE:	//	//    Kafka uses watches extensively → many round-trips to ZK.	//    ZK watches are one-shot (fire once, then must re-register).	// 6. WATCHER LIMITATIONS	//	//    Securing both is complex and error-prone.	//    ZK has its own auth/ACL model, different from Kafka's.	// 5. SECURITY MODEL MISMATCH	//	//    ZK network partitions → confusing state.	//    If ZK ensemble has issues, Kafka brokers might disagree on who's controller.	// 4. SPLIT-BRAIN RISK	//	//    During failover: no leader elections possible → partitions unavailable.	//    New controller must read FULL metadata from ZK → slow for large clusters.	//    When controller dies: ZK session timeout (6-30 seconds) before detection.	// 3. CONTROLLER FAILOVER	//	//    At LinkedIn scale: metadata changes were slow.	//    ZK has a limit: ~1 million znodes comfortably, performance degrades after.	//    All metadata goes through ZooKeeper (topics, partitions, ISR, configs).	// 2. METADATA BOTTLENECK	//	//    Different failure modes, different tuning, different expertise.	//    Two separate clusters to monitor, upgrade, secure, scale.	//    Kafka needs brokers + ZooKeeper ensemble (3-5 nodes).	// 1. SEPARATE SYSTEM TO OPERATE	//	// ───────────────────	// ZOOKEEPER PROBLEMS:	fmt.Println("--- WHY REMOVE ZOOKEEPER ---")func whyRemoveZookeeper() {}	kafkaFuture()	migration()	controllerQuorum()	metadataAsLog()	kraftArchitecture()	whyRemoveZookeeper()	fmt.Println()	fmt.Println("=== KRAFT & THE FUTURE ===")func main() {import "fmt"package main// =============================================================================//// Kafka, the distributed log system, uses a distributed log for its own metadata.// Kafka stores it in its OWN LOG. This is beautifully recursive:// metadata architecture. Instead of storing metadata in an external system,// KRaft doesn't just replace ZooKeeper — it fundamentally changes Kafka's// THE KEY INSIGHT://// - What's coming: tiered storage, queues, Kafka 4.0+// - Migration from ZooKeeper to KRaft// - Controller quorum: how KRaft controllers elect leaders// - Metadata as an event log: the __cluster_metadata topic// - KRaft consensus protocol: how it works// - Why ZooKeeper was removed (and why it took 10 years)// WHAT YOU'LL LEARN://// =============================================================================// LESSON 14.1: KRAFT & THE FUTURE — Kafka Without ZooKeeper// =============================================================================