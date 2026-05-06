//go:build ignore
// =============================================================================
// LESSON 6.1: REPLICATION — How Kafka Survives Broker Failures
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - ISR (In-Sync Replicas): the core of Kafka's durability model
// - High Watermark (HW) and Log End Offset (LEO): the two cursors
// - Leader epochs: preventing data divergence after failures
// - Unclean leader election: the availability vs durability tradeoff
// - min.insync.replicas: the golden trio config
// - Replication tuning: latency, throughput, and monitoring
//
// THE KEY INSIGHT:
// Kafka replication is NOT like database replication. It's a pull-based model
// where followers fetch from the leader. The ISR mechanism is Kafka's way of
// balancing durability with availability — and understanding it deeply is the
// difference between losing data and sleeping soundly at night.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== REPLICATION DEEP DIVE ===")
	fmt.Println()

	isrMechanism()
	hwAndLeo()
	leaderEpoch()
	uncleanLeaderElection()
	minInsyncReplicas()
	replicationTuning()
}

// =============================================================================
// PART 1: ISR — The core of Kafka's durability
// =============================================================================
func isrMechanism() {
	fmt.Println("--- IN-SYNC REPLICAS (ISR) ---")

	// Every partition has:
	//   - 1 LEADER: handles all reads and writes
	//   - N-1 FOLLOWERS: pull data from the leader (replication factor = N)
	//
	// ISR = the set of replicas that are "in sync" with the leader.
	// A replica is "in sync" if it has fetched all messages from the leader
	// within replica.lag.time.max.ms (default: 30 seconds).
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  REPLICATION FLOW (RF=3):                                     │
	// │                                                              │
	// │  Producer ──────► Leader (Broker 1)                          │
	// │                      │                                       │
	// │                      ├──► Follower 1 (Broker 2) fetches      │
	// │                      │    last fetch: 200ms ago → IN SYNC    │
	// │                      │                                       │
	// │                      └──► Follower 2 (Broker 3) fetches      │
	// │                           last fetch: 45 seconds ago → OUT!  │
	// │                                                              │
	// │  ISR = {Broker 1 (leader), Broker 2}                         │
	// │  OSR = {Broker 3} (Out-of-Sync Replicas)                     │
	// └──────────────────────────────────────────────────────────────┘
	//
	// HOW REPLICAS FALL OUT OF ISR:
	// ─────────────────────────────
	// 1. Follower crashes or becomes unreachable
	// 2. Follower is slow (disk I/O, network, GC pause)
	// 3. Follower is behind by more than replica.lag.time.max.ms
	//
	// NOTE: It's NOT based on lag in number of messages!
	// Old Kafka (< 0.9) used replica.lag.max.messages, which was removed
	// because bursty producers would cause constant ISR shrinking/expanding.
	// Time-based is more stable.
	//
	// WHEN REPLICAS REJOIN ISR:
	// ─────────────────────────
	// Once a follower catches up to the leader's LEO (Log End Offset),
	// it's added back to the ISR. This is checked on every fetch request.
	//
	// ISR CHANGE IMPACT:
	// ──────────────────
	// ISR changes are written to ZooKeeper/KRaft metadata log.
	// Frequent ISR changes (ISR shrinking/expanding) = something is WRONG.
	// Monitor: kafka.server:type=ReplicaManager,name=IsrShrinksPerSec
	//          kafka.server:type=ReplicaManager,name=IsrExpandsPerSec
	//
	// acks=all (producer) means: wait for ALL replicas in ISR to acknowledge.
	// If ISR = {leader only}, acks=all = acks=1 (no durability!).
	// → This is why min.insync.replicas is critical (see Part 5).

	fmt.Println("  ISR: set of replicas within replica.lag.time.max.ms of leader")
	fmt.Println("  Time-based (not message-count-based) for stability")
	fmt.Println("  acks=all waits for all ISR members to acknowledge")
	fmt.Println()
}

// =============================================================================
// PART 2: HIGH WATERMARK (HW) AND LOG END OFFSET (LEO)
// =============================================================================
func hwAndLeo() {
	fmt.Println("--- HIGH WATERMARK & LOG END OFFSET ---")

	// Every replica maintains TWO offsets:
	//
	// LEO (Log End Offset): the offset of the NEXT message to be written.
	//   = last written offset + 1
	//   Each replica has its own LEO (may differ).
	//
	// HW (High Watermark): the highest offset COMMITTED (fully replicated).
	//   = min(LEO of all ISR members)
	//   Consumers can only read up to the HW.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  LOG STATE (Partition 0, RF=3, all in ISR):                   │
	// │                                                              │
	// │  Leader:     [0][1][2][3][4][5][6][7]     LEO=8              │
	// │  Follower 1: [0][1][2][3][4][5][6]        LEO=7              │
	// │  Follower 2: [0][1][2][3][4][5]           LEO=6              │
	// │                                                              │
	// │  HW = min(8, 7, 6) = 6                                       │
	// │  → Consumers can read offsets 0-5                             │
	// │  → Offsets 6-7 exist on leader but are NOT yet visible        │
	// │                                                              │
	// │  After Follower 2 catches up to LEO=8:                        │
	// │  HW = min(8, 7, 8) = 7 → consumers now see offset 6          │
	// │  After Follower 1 catches up to LEO=8:                        │
	// │  HW = min(8, 8, 8) = 8 → consumers now see offsets 6-7       │
	// └──────────────────────────────────────────────────────────────┘
	//
	// HW UPDATE PROTOCOL:
	// ────────────────────
	// 1. Producer writes to leader → leader LEO advances
	// 2. Followers FETCH from leader → each follower's LEO advances
	//    Fetch request includes the follower's current LEO
	// 3. Leader receives fetch with follower LEO → updates its view of
	//    that follower's LEO → recalculates HW = min(all ISR LEOs)
	// 4. New HW is piggy-backed in the NEXT fetch response to followers
	// 5. Followers update their local HW
	//
	// LATENCY IMPLICATION:
	// ────────────────────
	// It takes AT LEAST 2 fetch cycles for HW to advance:
	//   Cycle 1: Follower fetches data, updates its LEO
	//   Cycle 2: Follower fetches again, receives updated HW
	//
	// replica.fetch.wait.max.ms (default: 500ms) controls how often
	// followers fetch from the leader. Lower = faster HW advance = lower
	// end-to-end latency. But more network overhead.
	//
	// WHY CONSUMERS CAN'T READ PAST HW:
	// ──────────────────────────────────
	// If a consumer read offset 7 (not yet replicated) and the leader crashes,
	// the new leader might not have offset 7. The consumer would have data that
	// "never existed" from Kafka's perspective.
	// The HW ensures consumers only see data that survives leader failure.

	fmt.Println("  LEO: next offset to write (per replica)")
	fmt.Println("  HW: min(LEO of all ISR) = last committed offset")
	fmt.Println("  Consumers can only read up to HW (committed data)")
	fmt.Println("  HW advance takes 2 fetch cycles (~1 second by default)")
	fmt.Println()
}

// =============================================================================
// PART 3: LEADER EPOCHS — Preventing data divergence
// =============================================================================
func leaderEpoch() {
	fmt.Println("--- LEADER EPOCHS ---")

	// PROBLEM: After a leader failure, replicas can DIVERGE.
	//
	// Before leader epochs (Kafka < 0.11), this could happen:
	//
	// 1. Leader has [0,1,2,3,4], HW=3, Follower has [0,1,2], HW=3
	// 2. Leader crashes. Follower becomes new leader.
	// 3. Old leader comes back. It has [3,4] that new leader doesn't have.
	// 4. Old leader truncates to HW=3... but what if HW was wrong?
	//
	// SOLUTION: Leader Epoch (introduced in KIP-101)
	// ────────────────────────────────────────────────
	// An epoch is a monotonically increasing number for each leadership term.
	// Each record batch is tagged with the leader epoch at write time.
	//
	// When a follower starts up after a failure:
	// 1. It sends OffsetsForLeaderEpoch(epoch=X) to the current leader
	// 2. Leader responds: "for epoch X, the last offset was Y"
	// 3. Follower truncates to offset Y (not HW, which could be stale)
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  LEADER EPOCH EXAMPLE:                                        │
	// │                                                              │
	// │  Epoch 0 (Leader=Broker1):                                    │
	// │  Broker1: [0:e0][1:e0][2:e0][3:e0][4:e0]  (leader)           │
	// │  Broker2: [0:e0][1:e0][2:e0][3:e0]         (follower)        │
	// │                                                              │
	// │  Broker1 crashes! Broker2 becomes leader with epoch=1.       │
	// │                                                              │
	// │  Epoch 1 (Leader=Broker2):                                    │
	// │  Broker2: [0:e0][1:e0][2:e0][3:e0][4:e1][5:e1]  (leader)    │
	// │                                                              │
	// │  Broker1 comes back! Asks: "what's the end offset for e0?"  │
	// │  Broker2: "epoch 0 ended at offset 4"                        │
	// │  Broker1: truncates to offset 4 (removes [4:e0])             │
	// │  Broker1: [0:e0][1:e0][2:e0][3:e0] → fetches [4:e1][5:e1]  │
	// │                                                              │
	// │  Result: consistent! No divergence!                           │
	// └──────────────────────────────────────────────────────────────┘
	//
	// Without leader epochs, truncation was based on HW which could be
	// stale, causing both data loss AND data divergence between replicas.
	// Leader epochs solved one of Kafka's longest-standing correctness issues.

	fmt.Println("  Leader epoch: monotonic counter incremented on each leader change")
	fmt.Println("  Prevents data divergence after leader failure")
	fmt.Println("  Follower truncates based on epoch boundary, not stale HW")
	fmt.Println()
}

// =============================================================================
// PART 4: UNCLEAN LEADER ELECTION
// =============================================================================
func uncleanLeaderElection() {
	fmt.Println("--- UNCLEAN LEADER ELECTION ---")

	// SCENARIO: All ISR replicas are dead. Only out-of-sync replicas remain.
	//
	// unclean.leader.election.enable (default: false since Kafka 0.11)
	//
	// If TRUE: An out-of-sync replica CAN become leader.
	//   Pro: AVAILABILITY — partition can accept writes again
	//   Con: DATA LOSS — all messages not replicated to this replica are GONE
	//        Also: consumers may see messages they already read disappear
	//
	// If FALSE: The partition WAITS until an ISR member comes back.
	//   Pro: NO DATA LOSS
	//   Con: UNAVAILABILITY — partition is offline until ISR member recovers
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  THE TRADEOFF:                                                │
	// │                                                              │
	// │  unclean.leader.election.enable=true                          │
	// │  ─────────────────────────────────                            │
	// │  → "I'd rather lose some data than have my entire pipeline    │
	// │     stall waiting for a dead broker to come back"             │
	// │  → Use for: metrics, logs, non-critical analytics             │
	// │  → NEVER for: financial data, orders, user data               │
	// │                                                              │
	// │  unclean.leader.election.enable=false (RECOMMENDED DEFAULT)   │
	// │  ────────────────────────────────────                          │
	// │  → "I'd rather be unavailable than serve incorrect data"      │
	// │  → Use for: anything where correctness > availability         │
	// │  → Most production systems should use this                    │
	// └──────────────────────────────────────────────────────────────┘
	//
	// In practice, with RF=3 and good broker distribution:
	// The probability that ALL 3 replicas are simultaneously dead is very low.
	// Only if you have: rack awareness disabled, or correlated failures
	// (e.g., all replicas on same rack, and the rack loses power).
	//
	// PREVENTION:
	// ───────────
	// 1. Use rack-awareness (broker.rack=rack-A, broker.rack=rack-B)
	//    Kafka distributes replicas across racks → single rack failure is safe
	// 2. RF=3 minimum for important topics (RF=2 means one failure = no ISR)
	// 3. Monitor ISR: if ISR shrinks to 1, that's already critical!

	fmt.Println("  Unclean=true: availability over durability (for non-critical data)")
	fmt.Println("  Unclean=false (default): durability over availability")
	fmt.Println("  Prevention: RF=3, rack awareness, monitor ISR size")
	fmt.Println()
}

// =============================================================================
// PART 5: min.insync.replicas — The Golden Config
// =============================================================================
func minInsyncReplicas() {
	fmt.Println("--- MIN.INSYNC.REPLICAS ---")

	// min.insync.replicas (default: 1)
	// Only used when producer sets acks=all.
	// Defines: the MINIMUM number of ISR members that must acknowledge
	// a write for it to succeed.
	//
	// THE GOLDEN TRIO:
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │                                                              │
	// │  replication.factor = 3                                       │
	// │  min.insync.replicas = 2                                      │
	// │  acks = all (producer setting)                                │
	// │                                                              │
	// │  WHAT THIS MEANS:                                             │
	// │  ─────────────────                                            │
	// │  • 3 copies of each message (including leader)                │
	// │  • Write succeeds if ≥ 2 replicas ACK                         │
	// │  • Write FAILS if < 2 replicas in ISR                         │
	// │    (NOT_ENOUGH_REPLICAS exception → producer retries)         │
	// │  • Tolerates 1 broker failure without data loss OR downtime   │
	// │  • Tolerates 1 broker failure for writes (ISR shrinks to 2)   │
	// │  • If 2 brokers fail: reads work (1 replica left as leader)   │
	// │    but writes fail (ISR < min.insync)                         │
	// │                                                              │
	// │  Balance of:                                                   │
	// │  ✓ Durability: survive 1 broker failure without data loss     │
	// │  ✓ Availability: survive 1 broker failure for both R+W        │
	// │  ✓ Performance: only wait for 2 acks, not 3                   │
	// │  ✓ Cost: 3x storage (industry standard, not excessive)        │
	// │                                                              │
	// └──────────────────────────────────────────────────────────────┘
	//
	// OTHER CONFIGURATIONS:
	//
	// RF=3, min.insync=1, acks=all
	//   → acks=all with ISR=1 is the same as acks=1 (no durability gain!)
	//   → If 2 replicas die, writes still succeed but data may be lost
	//   → DON'T USE THIS. It gives false sense of durability.
	//
	// RF=3, min.insync=3, acks=all
	//   → EVERY write must reach ALL 3 replicas
	//   → If ANY broker dies, writes FAIL immediately
	//   → Too aggressive. You lose write availability on any single failure.
	//   → Only for extremely critical financial data (and you'll hate your SLA)
	//
	// RF=5, min.insync=3, acks=all
	//   → Tolerates 2 simultaneous broker failures
	//   → Used in multi-AZ setups where AZ failures happen
	//   → 5x storage cost, higher replication latency

	fmt.Println("  GOLDEN CONFIG: RF=3, min.insync=2, acks=all")
	fmt.Println("  Tolerates 1 broker failure for both reads AND writes")
	fmt.Println("  min.insync=1 is useless with acks=all")
	fmt.Println("  min.insync=RF is too aggressive (any failure blocks writes)")
	fmt.Println()
}

// =============================================================================
// PART 6: REPLICATION TUNING AND MONITORING
// =============================================================================
func replicationTuning() {
	fmt.Println("--- REPLICATION TUNING ---")

	// KEY REPLICATION CONFIGS:
	// ────────────────────────
	//
	// replica.lag.time.max.ms (default: 30000 = 30 sec)
	//   Maximum time a follower can be behind before being removed from ISR.
	//   Lower: faster detection of slow replicas, but may cause ISR flapping
	//          during brief GC pauses or network blips.
	//   Higher: more tolerant but slow to detect truly dead replicas.
	//   Recommendation: 10000-30000ms (10-30 seconds)
	//
	// replica.fetch.max.bytes (default: 1048576 = 1 MB)
	//   Maximum data a follower fetches per request.
	//   If your message sizes are large (>100 KB), increase this.
	//   Must be larger than your max message size.
	//
	// replica.fetch.wait.max.ms (default: 500ms)
	//   Max time a follower waits for data on a fetch (long-poll).
	//   Lower: faster replication, lower commit latency.
	//   Higher: less network overhead.
	//   For low-latency: 100-200ms
	//
	// num.replica.fetchers (default: 1)
	//   Number of threads each follower uses to fetch from leaders.
	//   If a broker has many partitions to replicate, increase this (2-4).
	//
	// MONITORING REPLICATION HEALTH:
	// ──────────────────────────────
	//
	// ┌─────────────────────────────────────────────────────────────────┐
	// │ METRIC                            │ ALERT THRESHOLD              │
	// ├─────────────────────────────────────────────────────────────────┤
	// │ UnderReplicatedPartitions          │ > 0 (RED ALERT)            │
	// │ IsrShrinksPerSec                   │ > 0 sustained (WARNING)    │
	// │ IsrExpandsPerSec                   │ Correlate with shrinks     │
	// │ UnderMinIsrPartitionCount          │ > 0 (CRITICAL — data risk)│
	// │ ReplicaMaxLag                      │ > replica.lag.time.max.ms │
	// │ OfflinePartitionsCount             │ > 0 (CRITICAL — outage)    │
	// │ ActiveControllerCount              │ != 1 (CRITICAL)            │
	// └─────────────────────────────────────────────────────────────────┘
	//
	// UnderReplicatedPartitions > 0 means:
	// - A follower is behind or dead
	// - Data is at risk of loss if leader also dies
	// - This should ALWAYS trigger an alert
	//
	// UnderMinIsrPartitionCount > 0 means:
	// - ISR < min.insync.replicas
	// - Writes with acks=all will START FAILING
	// - This is a CRITICAL alert — page someone

	fmt.Println("  replica.lag.time.max.ms=10-30s: ISR detection speed")
	fmt.Println("  Monitor: UnderReplicatedPartitions MUST be 0")
	fmt.Println("  Any UnderMinIsrPartitionCount > 0 = page someone immediately")
	fmt.Println()
}









































































































































































































































































































































































































}	fmt.Println("  Alert: UnderMinIsrPartitionCount > 0 → writes are failing!")	fmt.Println("  Monitor: UnderReplicatedPartitions (must be 0)")	fmt.Println("  Key: replica.lag.time.max.ms (10-30s), num.replica.fetchers (1-4)")	//   Alert IMMEDIATELY.	//   Partitions where ISR < min.insync.replicas → WRITES WILL FAIL!	// kafka.server:type=ReplicaManager,name=UnderMinIsrPartitionCount	//	//   Rate of ISR expand events. Should follow shrinks (recovery).	// kafka.server:type=ReplicaManager,name=IsrExpandsPerSec	//	//   Rate of ISR shrink events. Spikes indicate follower problems.	// kafka.server:type=ReplicaManager,name=IsrShrinksPerSec	//	//   Should be 0 in steady state. Alert if > 0 for > 5 minutes.	//   Number of partitions where ISR < replication.factor	// kafka.server:type=ReplicaManager,name=UnderReplicatedPartitions	// ──────────────────────────────	// MONITORING REPLICATION HEALTH:	//	//   Similar to consumer's fetch.max.wait.ms.	//   How long the leader waits for data when follower's fetch finds nothing.	// replica.fetch.wait.max.ms (default: 500)	//	//   Increase if partitions have large records or high throughput.	//   Maximum bytes per partition per fetch request.	// replica.fetch.max.bytes (default: 1048576 = 1 MB)	//	//   Each thread handles a subset of partitions.	//   increase to 2-4.	//   If brokers have many partitions and followers can't keep up:	//   Fetch threads per follower broker.	// num.replica.fetchers (default: 1)	//	//   Sweet spot: 10-30s depending on cluster health.	//   Too high (> 60s): slow followers stay in ISR → acks=all is slow	//   Too low (< 10s): followers drop from ISR during GC pauses → false alarms	//   Time before a slow follower is removed from ISR.	// replica.lag.time.max.ms (default: 30000)	//	// KEY SETTINGS THAT AFFECT ISR STABILITY:	fmt.Println("--- REPLICATION TUNING ---")func replicationTuning() {// =============================================================================// PART 6: REPLICATION TUNING// =============================================================================}	fmt.Println()	fmt.Println("  NEVER set min.insync.replicas = replication.factor")	fmt.Println("  = tolerates 1 failure for writes, 2 for reads, no data loss")	fmt.Println("  GOLDEN CONFIG: RF=3, min.insync.replicas=2, acks=all")	// └──────────────────────────────────────────────────────────────┘	// │  - 5 copies = expensive but bulletproof                      │	// │  - Uber/LinkedIn use this for critical topics                │	// │  - Tolerates 2 broker failures for writes                    │	// │  RF=5, min.isr=3, acks=all:                                  │	// │                                                              │	// │  - DON'T USE THIS. Any maintenance → topic unavailable.      │	// │  - Maximum durability but terrible availability              │	// │  - Tolerates 0 broker failures (writes fail if ANY is down) │	// │  RF=3, min.isr=3, acks=all:                                  │	// │                                                              │	// │  - Data is ALWAYS on at least 2 brokers before ACK          │	// │  - Tolerates 2 broker failures (for reads, no new writes)   │	// │  - Tolerates 1 broker failure (for writes)                   │	// │  RF=3, min.isr=2, acks=all: ← THE GOLD STANDARD             │	// │                                                              │	// │  - If ISR = {leader}, acks=all = acks=1 = POSSIBLE LOSS     │	// │  - Tolerates 0 broker failures (for guaranteed durability!)  │	// │  - Tolerates 2 broker failures (for reads)                   │	// │  RF=3, min.isr=1, acks=all:                                  │	// │                                                              │	// │  CONFIGURATION MATRIX:                                        │	// ┌──────────────────────────────────────────────────────────────┐	//	//   but NO DATA IS LOST (as long as at least 1 replica survives)	// - 2 brokers fail: partition becomes unavailable (can't meet min ISR)	// - 1 broker can fail without data loss OR unavailability	// - Produce succeeds if at least 2 replicas ack (leader + 1 follower)	// - 3 copies of every record (on 3 different brokers)	// This means:	//	// replication.factor=3, min.insync.replicas=2, acks=all	// ─────────────────────────	// THE GOLDEN CONFIGURATION:	//	// This prevents the "ISR-of-one" problem.	//	//   Error: NOT_ENOUGH_REPLICAS	//   The broker will REJECT produce requests if ISR size < 2.	// min.insync.replicas=2 means:	//	// But what if ISR = {leader only}? Then acks=all = acks=1!	// With acks=all, the producer waits for ALL ISR replicas to ack.	//	// min.insync.replicas (broker or topic level, default: 1)	fmt.Println("--- min.insync.replicas ---")func minInsyncReplicas() {// =============================================================================// PART 5: min.insync.replicas — The Durability Dial// =============================================================================}	fmt.Println()	fmt.Println("  Default is false. Enable only for non-critical data.")	fmt.Println("  unclean.leader.election=true: prefer availability (may lose data)")	fmt.Println("  unclean.leader.election=false: prefer durability over availability")	// └──────────────────────────────────────────────────────────────┘	// │  - Anything where HA is handled at the application level      │	// │  - State stores / CDC: gaps cause inconsistency              │	// │  - Financial transactions: cannot lose data                   │	// │  WHEN TO KEEP DISABLED:                                       │	// │                                                              │	// │  - When availability > durability                             │	// │  - Analytics: approximate data is fine                        │	// │  - Metrics/logging topics: losing some data is acceptable     │	// │  WHEN TO ENABLE UNCLEAN ELECTION:                             │	// │                                                              │	// │  unclean=true:  ░░░░░░░░░░░░ DURABILITY  ████ AVAILABILITY  │	// │  unclean=false: ████████████ DURABILITY  ░░░░ AVAILABILITY  │	// │                                                              │	// │  AVAILABILITY vs DURABILITY TRADEOFF:                         │	// ┌──────────────────────────────────────────────────────────────┐	//	//   - The log has a GAP (some offsets are missing)	//   - Records that were on B1 but not on B2/B3 are LOST	//   Partition is available again, BUT:	//   B2 or B3 can become leader even though they're not in ISR.	// WITH unclean.leader.election=true:	//	//   If B1's disk is dead → partition is permanently unavailable!	//   No data loss, but no availability either.	//   Partition is UNAVAILABLE until B1 comes back.	// WITH unclean.leader.election=false (default):	//	// - No more ISR replicas alive.	// - B1 crashes!	// - ISR = {B1}	// - B2 and B3 fall out of ISR (too slow)	// - Partition has 3 replicas: B1 (leader, ISR), B2 (ISR), B3 (ISR)	// SCENARIO:	//	// unclean.leader.election.enable (default: false since Kafka 0.11)	fmt.Println("--- UNCLEAN LEADER ELECTION ---")func uncleanElection() {// =============================================================================// PART 4: UNCLEAN LEADER ELECTION// =============================================================================}	fmt.Println()	fmt.Println("  No more HW-based truncation ambiguity")	fmt.Println("  On restart: follower asks 'where did my epoch end?' → truncates")	fmt.Println("  Leader epoch prevents data divergence after leader failover")	// └──────────────────────────────────────────────────────────────┘	// │  ✓ Consistent!                                               │	// │  Broker-1 fetches D,E from Broker-2 → [A:e0][B:e0][D:e1][E:e1]│	// │  Broker-1 truncates: removes C (offset 2) → [A:e0][B:e0]    │	// │  Broker-2 responds: epoch 0 ended at offset 2                 │	// │  Sends: OffsetsForLeaderEpoch(epoch=0) to Broker-2            │	// │  Broker-1 comes back with [A:e0][B:e0][C:e0]                 │	// │                                                              │	// │  New records: [A:e0][B:e0][D:e1][E:e1]                       │	// │  Broker-2 had [A:e0][B:e0] (never got C)                     │	// │  Epoch 1: Broker-2 is leader                                  │	// │  Broker-1 crashes. Broker-2 becomes leader.                  │	// │                                                              │	// │  [A:e0][B:e0][C:e0]  (records A,B,C written under epoch 0)  │	// │  Epoch 0: Broker-1 is leader                                  │	// │                                                              │	// │  LEADER EPOCH EXAMPLE:                                        │	// ┌──────────────────────────────────────────────────────────────┐	//	// 3. Follower truncates to that offset → ensures consistent state	// 2. Leader responds with: the last offset of that epoch	// 1. Follower sends OffsetsForLeaderEpoch(myLastEpoch) to the leader	// On follower restart:	//	//   Batch header: partitionLeaderEpoch field	// The epoch is stored with each record batch in the log:	//	// When a new leader is elected, the epoch increments.	// Each leader has a monotonically increasing EPOCH number.	// ──────────────────────────────────────────────────	// SOLUTION: Leader Epoch (introduced in Kafka 0.11)	//	// at the same offset.	// This could cause data DIVERGENCE: two replicas with different data	//	//    It might keep C which conflicts with D. BAD.	//    But what if timing was different and HW was already 3?	//    It sees HW was 2, so it keeps [A, B] and truncates C. Good.	// 5. Old leader comes back. It has [A, B, C].	// 4. New leader receives new record D → [A, B, D]. HW=3.	// 3. Follower becomes new leader. It has [A, B]. HW=2.	// 2. Leader crashes.	// 1. Leader has [A, B, C]. HW=2 (A, B committed). C not yet replicated.	// Scenario (simplified):	//	// Before leader epochs, replicas used HW for truncation on restart.	// ───────────────────────────────────────	// PROBLEM (the old HW-based truncation):	fmt.Println("--- LEADER EPOCH ---")func leaderEpoch() {// =============================================================================// PART 3: LEADER EPOCH — Preventing data divergence// =============================================================================}	fmt.Println()	fmt.Println("  HW update is piggybacked on fetch responses (one round-trip delay)")	fmt.Println("  HW = min(LEO of ISR) → consumers can only read up to HW")	fmt.Println("  LEO = next offset to write (per replica)")	// at HW or below → visible to consumers → durable in ISR.	// This means: a producer ACK with acks=all guarantees the record is	//	// Only then does it: update HW, acknowledge the producer.	// ALL ISR replicas (i.e., all ISR LEOs ≥ record's offset + 1).	// When acks=all, the leader waits until the record is replicated to	// ─────────────────	// acks=all AND HW:	//	// This delay caused data inconsistency issues → fixed by leader epochs.	// followers knowing about it (one fetch round-trip).	// This means there's a DELAY between leader updating HW and	// IMPORTANT: HW update is PIGGYBACKED on fetch responses.	//	// 5. Consumers can now see records up to offset 6 (HW-1).	//    Followers update their local HW.	// 4. Leader includes HW in next fetch response.	//    Leader now knows all LEOs: {8, 8, 7}. HW = min(ISR) = 7.	// 3. Follower B3 fetches → gets record 6. B3.LEO = 7.	//    Leader now knows B2's LEO = 8 from the fetch response	// 2. Follower B2 fetches from leader → gets record 7. B2.LEO = 8	// 1. Producer sends record to leader → leader appends, LEO = 8	// ───────────────────	// HW UPDATE PROTOCOL:	//	// └──────────────────────────────────────────────────────────────┘	// │            consumers read up to here                         │	// │                      │                                       │	// │                      HW      LEO                             │	// │                      ▲       ▲                               │	// │  ──[0][1][2][3][4][5]│[6][7]──►                             │	// │                                                              │	// │  They become visible when ALL ISR replicas catch up.        │	// │  Offsets 6 and 7 exist on leader but are NOT visible yet!   │	// │  Consumers can read: offsets 0-5 (up to HW-1)               │	// │                                                              │	// │  HW = min(LEO of all ISR) = 6                                │	// │                                                              │	// │  B3 (Follower): [0][1][2][3][4][5]         LEO=6            │	// │  B2 (Follower): [0][1][2][3][4][5][6]      LEO=7            │	// │  B1 (Leader):   [0][1][2][3][4][5][6][7]   LEO=8            │	// │                                                              │	// │  PARTITION P0: replication.factor=3, ISR={B1,B2,B3}          │	// ┌──────────────────────────────────────────────────────────────┐	//	//   This ensures consumers never see data that could be lost.	//   Consumers can only read up to HW (not LEO!).	//   The offset up to which ALL ISR replicas have replicated.	// HW (High Watermark):	//	//   Each replica has its own LEO.	//   If last record has offset 9, LEO = 10.	//   The offset of the NEXT record to be written.	// LEO (Log End Offset):	//	// Every replica tracks TWO positions:	fmt.Println("--- HW (High Watermark) AND LEO (Log End Offset) ---")func hwAndLeo() {// =============================================================================// PART 2: HW and LEO — The Two Cursors// =============================================================================}	fmt.Println()	fmt.Println("  acks=all waits only for ISR replicas → if ISR shrinks, less safe!")	fmt.Println("  NOT about message count — about recency of last fetch")	fmt.Println("  ISR = replicas that fetched within replica.lag.time.max.ms (30s)")	//   Each thread handles a subset of partitions.	//   Increase if followers can't keep ISR membership (increase to 2-4).	//   Number of fetch threads per follower broker for pulling from leaders.	// num.replica.fetchers (default: 1)	//	// And they have their own fetch thread (ReplicaFetcherThread).	// But they use replica.fetch.max.bytes (default: 1 MB) per partition.	// They send FetchRequests just like consumers do.	// Followers are essentially consumers of the leader's log.	// ────────────────────	// HOW FOLLOWERS FETCH:	//	// └──────────────────────────────────────────────────────────────┘	// │  This is where min.insync.replicas saves you.                │	// │  If ISR = {Leader}, acks=all is effectively acks=1!          │	// │  acks=all only waits for ISR replicas.                       │	// │  WHY THIS MATTERS:                                            │	// │                                                              │	// │  Broker-3 is OUT-OF-SYNC (still a replica, just not in ISR) │	// │  ISR = {Broker-1, Broker-2}                                  │	// │  If Broker-3 stops fetching for 30 seconds:                  │	// │                                                              │	// │  (All fetched within last 30 seconds)                        │	// │  ISR = {Broker-1, Broker-2, Broker-3}                        │	// │                                                              │	// │  Broker-3 (Follower):[0][1][2][3][4][5][6]            LEO=7 │	// │  Broker-2 (Follower):[0][1][2][3][4][5][6][7][8]     LEO=9 │	// │  Broker-1 (Leader):  [0][1][2][3][4][5][6][7][8][9]  LEO=10│	// │                                                              │	// │  PARTITION P0: replication.factor=3                           │	// ┌──────────────────────────────────────────────────────────────┐	//	//   Notifies controller → controller updates metadata	//   Leader adds follower back to ISR	// Follower catches up:	//	//   Produce requests with acks=all now only wait for remaining ISR	//   Notifies controller → controller updates metadata	//   Leader shrinks ISR: removes slow follower	// Follower falls behind (no fetch for > 30s):	// ────────────	// ISR CHANGES:	//	// fetched within the last 30 seconds (just slow at fetching).	// A follower can be 100,000 messages behind but still in ISR if it	// NOTE: It's NOT about how many messages behind. It's about TIME.	//	// Default replica.lag.time.max.ms: 30000 (30 seconds)	// In other words: the follower has fetched from the leader recently.	//	//   (leader.LEO - follower.LEO) was fetched within replica.lag.time.max.ms	// A follower is in the ISR if:	// ─────────────────────	// WHAT "IN SYNC" MEANS:	//	//   - ISR: subset of replicas that are "caught up" to the leader	//   - N-1 Followers: replicate data from the leader	//   - 1 Leader: handles ALL produce and consume requests	// Every partition has:	fmt.Println("--- ISR MECHANISM ---")func isrMechanism() {// =============================================================================// PART 1: ISR — In-Sync Replicas// =============================================================================}	replicationTuning()	minInsyncReplicas()	uncleanElection()	leaderEpoch()	hwAndLeo()	isrMechanism()	fmt.Println()	fmt.Println("=== REPLICATION DEEP DIVE ===")func main() {import "fmt"package main// =============================================================================//// when it's safe to acknowledge a produce request.// They fetch data exactly like a consumer does. The ISR mechanism then determines// no logical decoding. Followers are just CONSUMERS of the leader's log.// Kafka replication is NOT like database replication. There's no WAL shipping,// THE KEY INSIGHT://// - Follower fetching: how replicas stay in sync// - min.insync.replicas: the durability dial// - Unclean leader election: when Kafka CHOOSES to lose data// - Leader epoch: how Kafka prevents data divergence after leader change// - HW (High Watermark) and LEO (Log End Offset): the two cursors// - ISR (In-Sync Replicas): what "in sync" actually means// WHAT YOU'LL LEARN://// =============================================================================// LESSON 6.1: REPLICATION DEEP DIVE — How Kafka Never Loses Your Data// =============================================================================