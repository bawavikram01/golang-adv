//go:build ignore
// =============================================================================
// LESSON 4.1: CONSUMER INTERNALS — The Group Protocol & Offset Management
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Consumer group protocol: join, sync, heartbeat, leave
// - Rebalancing: eager vs cooperative (incremental), static membership
// - Offset management: auto-commit pitfalls, manual commit strategies
// - Fetch internals: fetch.min.bytes, fetch.max.wait.ms, max.poll.records
// - Consumer lag: what it means, how to measure, how to fix
// - The poll loop: why max.poll.interval.ms matters more than you think
//
// THE KEY INSIGHT:
// Consumers don't "receive" messages. They PULL from a position (offset) in the log.
// The broker doesn't know or care what a consumer has "processed" — the consumer
// tracks its own position. This decoupling is what enables replay, multi-consumer,
// and independent scaling.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== CONSUMER INTERNALS ===")
	fmt.Println()

	consumerGroupProtocol()
	rebalanceStrategies()
	offsetManagement()
	fetchInternals()
	pollLoop()
	consumerLag()
}

// =============================================================================
// PART 1: CONSUMER GROUP PROTOCOL
// =============================================================================
func consumerGroupProtocol() {
	fmt.Println("--- CONSUMER GROUP PROTOCOL ---")

	// A consumer group is a set of consumers that cooperatively consume
	// from a set of topics. Each partition is assigned to EXACTLY ONE
	// consumer in the group.
	//
	// KEY PLAYERS:
	// ─────────────
	// Group Coordinator: A specific broker responsible for this consumer group.
	//   Determined by: hash(groupId) % __consumer_offsets partitions
	//   The coordinator manages membership, assignment, and offset storage.
	//
	// Group Leader: One consumer in the group, elected to perform partition assignment.
	//   The coordinator just facilitates; the LEADER does the actual assignment.
	//   This is a unique design: assignment logic resides in the CLIENT, not server.
	//   Why? So you can plug in custom assignors without broker changes.
	//
	// THE PROTOCOL (4 phases):
	// ────────────────────────
	//
	// ┌──────────────────────────────────────────────────────────────────┐
	// │  PHASE 1: FIND COORDINATOR                                       │
	// │  ─────────────────────────                                       │
	// │  Consumer → Any Broker: FindCoordinator(groupId)                 │
	// │  Broker → Consumer: Coordinator is Broker-X                      │
	// │                                                                  │
	// │  PHASE 2: JOIN GROUP                                              │
	// │  ─────────────────                                                │
	// │  Consumer → Coordinator: JoinGroup(groupId, protocols, metadata) │
	// │  Coordinator waits for ALL expected consumers (or timeout)        │
	// │  Coordinator selects a LEADER (first consumer to join)            │
	// │  Coordinator → Leader: JoinGroupResponse(members=all, isLeader)  │
	// │  Coordinator → Others: JoinGroupResponse(members=[], isFollower) │
	// │                                                                  │
	// │  PHASE 3: SYNC GROUP                                              │
	// │  ──────────────────                                               │
	// │  Leader computes assignment (e.g., RangeAssignor)                 │
	// │  Leader → Coordinator: SyncGroup(assignment for each member)      │
	// │  Others → Coordinator: SyncGroup(empty)                           │
	// │  Coordinator → ALL: SyncGroupResponse(your assigned partitions)  │
	// │                                                                  │
	// │  PHASE 4: HEARTBEAT + FETCH LOOP                                  │
	// │  ───────────────────────────────                                  │
	// │  Consumer sends heartbeats every heartbeat.interval.ms (3s)      │
	// │  Coordinator replies with either OK or REBALANCE_IN_PROGRESS     │
	// │  If REBALANCE: consumer must rejoin (back to Phase 2)            │
	// │  Meanwhile: consumer fetches data and processes records          │
	// └──────────────────────────────────────────────────────────────────┘
	//
	// HEARTBEAT MECHANISM:
	// ────────────────────
	// heartbeat.interval.ms (default: 3000ms)
	//   How often the consumer sends heartbeats to the coordinator.
	//   Must be < session.timeout.ms / 3 (to detect failures quickly).
	//
	// session.timeout.ms (default: 45000ms, previous default was 10000ms)
	//   If coordinator doesn't receive a heartbeat within this time,
	//   the consumer is considered DEAD and a rebalance is triggered.
	//   Lower = faster failure detection, but more false positives.
	//
	// max.poll.interval.ms (default: 300000ms = 5 minutes)
	//   If the consumer doesn't call poll() within this time,
	//   it's considered STUCK and removed from the group.
	//   This is SEPARATE from heartbeat! A consumer can send heartbeats
	//   (on a background thread) while being stuck in processing.
	//   The coordinator detects this via max.poll.interval.ms.

	fmt.Println("  Group protocol: FindCoordinator → JoinGroup → SyncGroup → Heartbeat+Fetch")
	fmt.Println("  Leader consumer computes partition assignment (client-side logic)")
	fmt.Println("  session.timeout.ms: heartbeat-based death detection")
	fmt.Println("  max.poll.interval.ms: poll-based stuck detection")
	fmt.Println()
}

// =============================================================================
// PART 2: REBALANCE STRATEGIES
// =============================================================================
func rebalanceStrategies() {
	fmt.Println("--- REBALANCE STRATEGIES ---")

	// Rebalancing is when partition assignments change within a consumer group.
	// This happens when: consumer joins, consumer leaves/dies, topics/partitions change.
	//
	// EAGER REBALANCING (pre-Kafka 2.4):
	// ───────────────────────────────────
	// ALL consumers REVOKE ALL partitions → JoinGroup → SyncGroup → Re-assigned
	//
	// Timeline:
	// t=0  Consumer-C3 dies
	// t=3s Coordinator detects (session timeout)
	// t=3s ALL consumers (C1, C2) revoke ALL partitions → STOP PROCESSING
	// t=3s JoinGroup + SyncGroup (new assignment)
	// t=4s C1 and C2 resume with new partitions
	//
	// PROBLEM: During rebalance, ALL consumers stop processing ALL partitions!
	// Called the "stop-the-world" rebalance. Can take 3-30 seconds.
	// At scale (100+ partitions), this is devastating for latency.
	//
	// COOPERATIVE REBALANCING (Kafka 2.4+, CooperativeStickyAssignor):
	// ─────────────────────────────────────────────────────────────────
	// Only AFFECTED partitions are revoked. Other consumers keep processing.
	//
	// Timeline:
	// t=0  Consumer-C3 dies
	// t=3s Coordinator detects (session timeout)
	// t=3s JoinGroup: Coordinator tells consumers about new membership
	// t=3s Consumers figure out which partitions need to MOVE
	// t=3s Only C1 revokes partition P5 (which needs to move to C2)
	//      C1 keeps P0, P1. C2 keeps P2, P3. THEY KEEP PROCESSING!
	// t=4s Second JoinGroup: C2 gets P5 assigned
	//
	// MUCH BETTER: Only the partitions that change ownership are affected.
	// The "stop-the-world" window is minimized.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  EAGER vs COOPERATIVE:                                        │
	// │                                                              │
	// │  EAGER:                                                       │
	// │  C1: [P0,P1,P2] ──── REVOKE ALL ──── [P0,P1] ✓              │
	// │  C2: [P3,P4,P5] ──── REVOKE ALL ──── [P3,P4,P5] ✓           │
	// │  C3: died                             (P2 moved to C2)       │
	// │       ◄──────── ALL stopped ────────►                        │
	// │                                                              │
	// │  COOPERATIVE:                                                  │
	// │  C1: [P0,P1,P2] ── revoke P2 only ── [P0,P1] ✓              │
	// │  C2: [P3,P4,P5] ── keep processing ── [P3,P4,P5,P2] ✓      │
	// │  C3: died                                                     │
	// │  C1 and C2 NEVER fully stopped!                              │
	// └──────────────────────────────────────────────────────────────┘
	//
	// PARTITION ASSIGNMENT STRATEGIES:
	// ────────────────────────────────
	// RangeAssignor (default before 3.0):
	//   Assigns ranges of partitions per topic. Can be uneven across consumers.
	//   Example: 3 partitions, 2 consumers → C1:[P0,P1], C2:[P2]
	//
	// RoundRobinAssignor:
	//   Round-robin across all partitions. More even distribution.
	//
	// StickyAssignor:
	//   Minimizes partition movement during rebalance. Eager protocol.
	//
	// CooperativeStickyAssignor (RECOMMENDED):
	//   Sticky + cooperative protocol. Minimal disruption.
	//   partition.assignment.strategy=
	//     org.apache.kafka.clients.consumer.CooperativeStickyAssignor
	//
	// STATIC MEMBERSHIP (group.instance.id):
	// ──────────────────────────────────────
	// Normal: consumer leaves → immediate rebalance
	// Static: consumer leaves → wait session.timeout.ms before rebalancing
	//
	// Set group.instance.id="consumer-1" on each consumer.
	// If a consumer restarts quickly (< session.timeout.ms):
	//   - It re-joins with the same instance ID
	//   - Gets back the SAME partitions
	//   - NO REBALANCE at all!
	//
	// Perfect for: rolling deployments, JVM restarts, transient failures.
	// Set session.timeout.ms=5-10 minutes for static membership.

	fmt.Println("  Eager: stop-the-world (all consumers revoke all partitions)")
	fmt.Println("  Cooperative: only affected partitions move (others keep processing)")
	fmt.Println("  CooperativeStickyAssignor = RECOMMENDED default")
	fmt.Println("  Static membership (group.instance.id) avoids rebalance on restart")
	fmt.Println()
}

// =============================================================================
// PART 3: OFFSET MANAGEMENT
// =============================================================================
func offsetManagement() {
	fmt.Println("--- OFFSET MANAGEMENT ---")

	// Offset = the consumer's position in the log.
	// Committed offset = the position the consumer has CONFIRMED processing.
	//
	// Offsets are stored in __consumer_offsets (internal compacted topic).
	// Key = (groupId, topic, partition)
	// Value = (offset, metadata, timestamp)
	//
	// TWO STRATEGIES:
	//
	// AUTO-COMMIT (enable.auto.commit=true, default):
	// ────────────────────────────────────────────────
	// Every auto.commit.interval.ms (default: 5000ms), the consumer
	// automatically commits the offset of the last record returned by poll().
	//
	// DANGER: Records returned by poll() != records PROCESSED.
	// If your app crashes between poll() and processing:
	//   - Offsets were committed (auto-commit happened)
	//   - But records weren't fully processed
	//   - On restart: consumer starts AFTER the committed offset → DATA LOSS
	//
	// The opposite: if you process records but crash before auto-commit:
	//   - Offsets weren't committed
	//   - On restart: consumer re-reads from last committed offset → REPROCESSING
	//
	// AUTO-COMMIT IS "AT-MOST-ONCE" OR "AT-LEAST-ONCE" DEPENDING ON TIMING.
	// You have NO CONTROL. This is fine for metrics/logging. Not for money.
	//
	// MANUAL COMMIT (enable.auto.commit=false):
	// ──────────────────────────────────────────
	// You explicitly call commitSync() or commitAsync() after processing.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  MANUAL COMMIT PATTERNS:                                      │
	// │                                                              │
	// │  PATTERN 1: Commit after every batch                         │
	// │  ───────────────────────────────                              │
	// │  records = poll()                                             │
	// │  for record in records: process(record)                       │
	// │  commitSync()  ← commit AFTER all records processed          │
	// │                                                              │
	// │  Semantics: AT-LEAST-ONCE                                    │
	// │  If crash between process + commit: re-read records on restart│
	// │  Your processing MUST be idempotent!                          │
	// │                                                              │
	// │  PATTERN 2: Commit per record (fine-grained)                 │
	// │  ───────────────────────────────────────────                  │
	// │  records = poll()                                             │
	// │  for record in records:                                       │
	// │    process(record)                                            │
	// │    commitSync(partition, record.offset+1)                     │
	// │                                                              │
	// │  Lower reprocessing on failure but MUCH slower (commit per msg)│
	// │                                                              │
	// │  PATTERN 3: Async commit + sync on shutdown                  │
	// │  ──────────────────────────────────────────                   │
	// │  while running:                                               │
	// │    records = poll()                                           │
	// │    process(records)                                           │
	// │    commitAsync()  ← fire-and-forget, fast                     │
	// │  on shutdown:                                                 │
	// │    commitSync()   ← blocking, ensures last offset is saved    │
	// │                                                              │
	// │  Best balance: fast normal operation, safe shutdown.          │
	// │  If async commit fails: next commit will cover it.            │
	// └──────────────────────────────────────────────────────────────┘
	//
	// COMMIT SEMANTICS:
	// ─────────────────
	// commitSync(topicPartition, OffsetAndMetadata):
	//   Produces a record to __consumer_offsets topic
	//   Key: (groupId, topic, partition)
	//   Value: (committed_offset, metadata_string, commit_timestamp)
	//
	// The committed offset is: offset of NEXT record to read (not last processed!)
	//   If you processed offset 42, commit offset 43.
	//
	// GOTCHA: Committing to __consumer_offsets is a PRODUCE operation.
	// It can fail! Network issues, coordinator failover, etc.
	// commitSync retries. commitAsync does NOT retry (stale offset could overwrite newer).

	fmt.Println("  Auto-commit: easy but no control over at-most/at-least once")
	fmt.Println("  Manual commit: commitSync() after processing = at-least-once")
	fmt.Println("  Best practice: commitAsync() normally + commitSync() on shutdown")
	fmt.Println("  Committed offset = next offset to read (last_processed + 1)")
	fmt.Println()
}

// =============================================================================
// PART 4: FETCH INTERNALS
// =============================================================================
func fetchInternals() {
	fmt.Println("--- FETCH INTERNALS ---")

	// When the consumer calls poll(), it returns records from a local buffer.
	// The actual network fetch happens in a background thread.
	//
	// FETCH TUNING:
	//
	// fetch.min.bytes (default: 1)
	//   Minimum data the broker should return per fetch request.
	//   If less data is available, broker waits (request goes to purgatory).
	//   Set higher (e.g., 50KB-1MB) to improve throughput at cost of latency.
	//
	// fetch.max.wait.ms (default: 500ms)
	//   How long the broker waits if fetch.min.bytes is not met.
	//   After this time, returns whatever it has (even if < min.bytes).
	//   Controls maximum additional latency from fetch.min.bytes.
	//
	// fetch.max.bytes (default: 52428800 = 50 MB)
	//   Maximum data the broker returns across ALL partitions in one fetch.
	//   Not a hard limit: at least one record batch is always returned.
	//
	// max.partition.fetch.bytes (default: 1048576 = 1 MB)
	//   Maximum data per partition in one fetch.
	//   If records are large (> 1 MB), increase this.
	//
	// max.poll.records (default: 500)
	//   Maximum records returned by poll() to your application.
	//   This is from the LOCAL buffer, not the network fetch.
	//   Lower = process fewer records per loop = lower latency per batch
	//   Higher = more records per loop = higher throughput
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  FETCH PIPELINE:                                              │
	// │                                                              │
	// │  Background                poll() method                     │
	// │  Fetcher Thread            (your thread)                     │
	// │  ════════════              ══════════════                     │
	// │                                                              │
	// │  FetchRequest ──► Broker                                     │
	// │                                                              │
	// │  ◄── FetchResponse (batches)                                 │
	// │  │                                                           │
	// │  ▼                                                           │
	// │  Decompress batches                                          │
	// │  │                                                           │
	// │  ▼                                                           │
	// │  ┌─────────────────────┐         poll() called               │
	// │  │ Completed Fetches   │────────►return up to                │
	// │  │ Buffer              │         max.poll.records            │
	// │  └─────────────────────┘         records                     │
	// │                                                              │
	// │  Meanwhile, pre-fetches                                      │
	// │  next batch from broker                                      │
	// │  (pipelining!)                                               │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  Background thread fetches from broker → local buffer → poll() returns")
	fmt.Println("  fetch.min.bytes: throughput knob (wait for more data before returning)")
	fmt.Println("  max.poll.records: controls records per poll() iteration")
	fmt.Println()
}

// =============================================================================
// PART 5: THE POLL LOOP — Why it's trickier than it looks
// =============================================================================
func pollLoop() {
	fmt.Println("--- THE POLL LOOP ---")

	// The consumer MUST call poll() regularly. If it doesn't call poll()
	// within max.poll.interval.ms (default: 5 minutes), the coordinator
	// considers it dead and triggers a rebalance.
	//
	// TYPICAL POLL LOOP:
	//
	// while (running) {
	//     records = consumer.poll(Duration.ofMillis(100))
	//     for record in records {
	//         process(record)  // YOUR BUSINESS LOGIC
	//     }
	//     consumer.commitAsync()
	// }
	//
	// THE TRAP:
	// If process(record) takes too long:
	//   - poll() isn't called for > max.poll.interval.ms
	//   - Coordinator triggers rebalance
	//   - Partitions reassigned to other consumers
	//   - Your consumer is now a "zombie" processing stale data
	//   - When it finally calls poll(), it discovers it's no longer in the group
	//   - Committed offsets might overwrite newer commits from the new owner!
	//
	// SOLUTIONS:
	// ──────────
	// 1. REDUCE max.poll.records: fewer records per batch = less processing time
	//
	// 2. INCREASE max.poll.interval.ms: more time for processing
	//    But: slower failure detection. Bad tradeoff.
	//
	// 3. SEPARATE THREADS: poll() in one thread, process in a thread pool
	//    But: complex offset management. Must track which offsets are "done."
	//    This is the Kafka consumer's biggest architectural challenge.
	//
	// 4. PAUSE/RESUME: Pause partition fetching while processing
	//    consumer.pause(partitions) → poll() returns nothing but sends heartbeats
	//    When done: consumer.resume(partitions) → poll() returns data again
	//    This is the cleanest approach for slow processing.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  THE PAUSE/RESUME PATTERN:                                    │
	// │                                                              │
	// │  records = poll()                                             │
	// │  consumer.pause(allPartitions)   ← stop fetching             │
	// │                                                              │
	// │  executor.submit(() -> {                                      │
	// │    processAll(records)          ← slow processing in threads │
	// │    consumer.commitSync()                                      │
	// │    consumer.resume(allPartitions) ← resume fetching          │
	// │  })                                                          │
	// │                                                              │
	// │  while(!done) {                                               │
	// │    poll(100ms)  ← returns 0 records (paused) but heartbeats  │
	// │  }                                                           │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  poll() MUST be called within max.poll.interval.ms (5 min default)")
	fmt.Println("  If processing is slow: reduce max.poll.records or use pause/resume")
	fmt.Println("  pause/resume: cleanest pattern for slow processing operations")
	fmt.Println()
}

// =============================================================================
// PART 6: CONSUMER LAG — The critical metric
// =============================================================================
func consumerLag() {
	fmt.Println("--- CONSUMER LAG ---")

	// Consumer lag = (latest offset in partition) - (consumer's committed offset)
	//
	// This tells you HOW FAR BEHIND a consumer is.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  PARTITION P0:                                                │
	// │                                                              │
	// │  ──[0][1][2][3][4][5][6][7][8][9][10][11]──►                │
	// │                    ▲                    ▲                     │
	// │                    │                    │                     │
	// │            committed offset=4    LEO=12 (Log End Offset)     │
	// │                                                              │
	// │  LAG = 12 - 4 = 8 records behind                             │
	// │                                                              │
	// │  Also useful: LAG IN TIME                                    │
	// │  If each record has a timestamp:                              │
	// │  lag_time = now - timestamp of record at committed offset    │
	// │  This is more meaningful: "30 seconds behind" vs "8 records" │
	// └──────────────────────────────────────────────────────────────┘
	//
	// MONITORING LAG:
	// ───────────────
	// Tool: kafka-consumer-groups.sh --describe --group my-group
	// Metric: kafka.consumer:type=consumer-fetch-manager-metrics,
	//         name=records-lag-max
	// External: Burrow (LinkedIn's lag monitoring), Prometheus + Grafana
	//
	// LAG CAUSES:
	// ───────────
	// 1. Producer rate > consumer rate
	//    Fix: add consumers (up to partition count), increase max.poll.records,
	//    optimize processing logic
	//
	// 2. Frequent rebalances
	//    Fix: use CooperativeStickyAssignor, static membership
	//    Symptom: lag spikes at regular intervals
	//
	// 3. Slow processing
	//    Fix: optimize business logic, use async processing,
	//    reduce max.poll.records
	//
	// 4. GC pauses (JVM consumers)
	//    Fix: tune GC, reduce heap allocation in processing loop
	//
	// 5. Broker throttling (quota exceeded)
	//    Fix: increase quota or add consumers to spread load
	//
	// LAG ALERTING:
	// ─────────────
	// WARNING: lag > 10,000 records for > 5 minutes
	// CRITICAL: lag > 100,000 records or growing continuously
	// EMERGENCY: lag > 1M records or lag time > SLA requirement
	//
	// THE LAG DEATH SPIRAL:
	// ─────────────────────
	// Lag increases → consumer reads from disk (page cache miss) →
	// reads are slower → lag increases further →
	// consumer falls further behind → even colder data on disk →
	// reads are even slower → LAG EXPLODES
	//
	// FIX: Add consumers, or accept data loss and skip ahead:
	//   consumer.seek(partition, latestOffset)

	fmt.Println("  Lag = latest offset - committed offset (how far behind)")
	fmt.Println("  Lag causes: slow processing, rebalances, producer > consumer rate")
	fmt.Println("  Watch for lag death spiral: lag → cold reads → slower → more lag")
	fmt.Println("  Nuclear option: seek to latest offset (skip the backlog)")
}
