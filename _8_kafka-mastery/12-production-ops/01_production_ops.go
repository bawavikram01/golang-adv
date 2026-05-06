//go:build ignore
// =============================================================================
// LESSON 12.1: PRODUCTION OPS — Running Kafka Like a Pro
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Critical metrics: the 3-tier alerting hierarchy
// - Alerting rules: what to page on, what to warn on
// - Capacity planning: formulas and real-world sizing
// - Rolling upgrades: zero-downtime broker upgrades
// - Incident response: 5 common scenarios and their runbooks
//
// THE KEY INSIGHT:
// Running Kafka in production is 80% monitoring and 20% responding.
// If you have the right metrics and alerts, most incidents are caught
// before they become outages. The #1 production mistake is not monitoring
// under-replicated partitions.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== PRODUCTION OPS ===")
	fmt.Println()

	criticalMetrics()
	alertingRules()
	capacityPlanning()
	rollingUpgrades()
	incidentResponse()
}

// =============================================================================
// PART 1: CRITICAL METRICS — The 3-tier hierarchy
// =============================================================================
func criticalMetrics() {
	fmt.Println("--- CRITICAL METRICS ---")

	// TIER 1: RED ALERT (page someone immediately)
	// ──────────────────────────────────────────────
	// These indicate data loss risk or active outage.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │ METRIC                          │ CONDITION   │ MEANING      │
	// ├──────────────────────────────────────────────────────────────┤
	// │ OfflinePartitionsCount           │ > 0         │ OUTAGE!      │
	// │   Partitions with NO leader.    │             │ Data unavail │
	// │                                 │             │              │
	// │ UnderMinIsrPartitionCount        │ > 0         │ DATA RISK!   │
	// │   ISR < min.insync.replicas.    │             │ Writes fail  │
	// │                                 │             │ with acks=all│
	// │                                 │             │              │
	// │ ActiveControllerCount            │ != 1        │ CLUSTER RISK!│
	// │   0 = no controller (no leader  │             │ No leader    │
	// │   elections possible).          │             │ elections    │
	// │   >1 = split brain (ZK mode)   │             │              │
	// └──────────────────────────────────────────────────────────────┘
	//
	// TIER 2: WARNING (investigate within 30 minutes)
	// ────────────────────────────────────────────────
	// ┌──────────────────────────────────────────────────────────────┐
	// │ UnderReplicatedPartitions        │ > 0         │ Replica lag  │
	// │   Some replicas behind leader.  │             │ May escalate │
	// │                                 │             │              │
	// │ IsrShrinksPerSec                 │ > 0 (5min) │ ISR changing │
	// │   Replicas falling out of ISR.  │             │ Broker sick? │
	// │                                 │             │              │
	// │ RequestHandlerAvgIdlePercent     │ < 0.3       │ Overloaded   │
	// │   Broker running out of threads.│             │ Add brokers  │
	// │                                 │             │              │
	// │ Consumer Group Lag               │ growing     │ Consumers    │
	// │   Consumer falling behind.      │             │ can't keep up│
	// └──────────────────────────────────────────────────────────────┘
	//
	// TIER 3: CAPACITY PLANNING (review weekly)
	// ──────────────────────────────────────────
	// ┌──────────────────────────────────────────────────────────────┐
	// │ Disk usage per broker            │ > 70%       │ Plan expand  │
	// │ Network utilization              │ > 60%       │ Add brokers  │
	// │ CPU utilization                  │ > 70% avg   │ Add brokers  │
	// │ Partition count per broker       │ > 2000      │ Rebalance    │
	// │ Leader partition imbalance       │ > 10% skew  │ Run election │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  Tier 1 (RED): OfflinePartitions, UnderMinIsr, ActiveController")
	fmt.Println("  Tier 2 (WARN): UnderReplicated, ISR shrinks, handler idle < 0.3")
	fmt.Println("  Tier 3 (PLAN): disk > 70%, network > 60%, partitions > 2000/broker")
	fmt.Println()
}

// =============================================================================
// PART 2: ALERTING RULES
// =============================================================================
func alertingRules() {
	fmt.Println("--- ALERTING RULES ---")

	// ┌──────────────────────────────────────────────────────────────────────┐
	// │  ALERT                  │ THRESHOLD               │ ACTION          │
	// ├──────────────────────────────────────────────────────────────────────┤
	// │  Offline partitions     │ > 0 for 1 min           │ PAGE immediately│
	// │  Under-min-ISR          │ > 0 for 2 min           │ PAGE immediately│
	// │  No active controller   │ == 0 for 1 min          │ PAGE immediately│
	// │  Under-replicated       │ > 0 for 5 min           │ WARN on-call    │
	// │  ISR shrinks (sustained)│ > 0 for 10 min          │ WARN on-call    │
	// │  Consumer lag growing   │ lag increasing 15 min    │ WARN team       │
	// │  Produce latency p99    │ > 500ms for 5 min       │ WARN team       │
	// │  Disk usage > 80%       │ Any broker              │ WARN + plan     │
	// │  Disk usage > 90%       │ Any broker              │ PAGE (imminent) │
	// │  Broker down            │ Broker unreachable 2min │ PAGE immediately│
	// └──────────────────────────────────────────────────────────────────────┘
	//
	// IMPLEMENTATION:
	// ───────────────
	// - JMX metrics → Prometheus (via jmx_exporter) → Grafana dashboards
	// - Consumer lag: use kafka-consumer-groups.sh or Burrow for lag monitoring
	// - PagerDuty/Opsgenie for Tier 1 alerts
	// - Slack/email for Tier 2 and 3

	fmt.Println("  Page on: offline partitions, under-min-ISR, no controller, disk >90%")
	fmt.Println("  Warn on: under-replicated, ISR shrinks, growing consumer lag")
	fmt.Println("  Stack: JMX → Prometheus → Grafana → PagerDuty")
	fmt.Println()
}

// =============================================================================
// PART 3: CAPACITY PLANNING
// =============================================================================
func capacityPlanning() {
	fmt.Println("--- CAPACITY PLANNING ---")

	// FORMULA:
	//
	// Storage per broker per day =
	//   (message_rate × avg_message_size × replication_factor × retention_days)
	//   ÷ number_of_brokers
	//
	// EXAMPLE:
	// - 100,000 messages/sec
	// - 1 KB average message size
	// - RF = 3
	// - 7 day retention
	// - 5 brokers
	//
	// Daily ingest: 100,000 × 1 KB × 86,400 sec = 8.64 TB/day (raw)
	// With RF=3: 8.64 TB × 3 = 25.92 TB/day total
	// Per broker: 25.92 TB / 5 = 5.18 TB/day
	// 7 days retention: 5.18 × 7 = 36.3 TB per broker
	//
	// With zstd compression (6x ratio): 36.3 / 6 = 6.05 TB per broker
	// Add 30% headroom: ~8 TB disk per broker
	//
	// NETWORK SIZING:
	// ────────────────
	// Inbound: message_rate × avg_size = 100 MB/sec
	// Replication: inbound × (RF - 1) = 200 MB/sec
	// Consumer: inbound × num_consumer_groups = 100-300 MB/sec
	//
	// Per broker: (100 + 200 + 300) / 5 brokers = 120 MB/sec
	// → 1 Gbps NIC is 125 MB/sec → BARELY enough!
	// → Use 10 Gbps minimum for production
	//
	// WHEN TO ADD BROKERS:
	// ────────────────────
	// - Disk > 70% → add before it hits 80%
	// - Network > 60% sustained → add brokers
	// - CPU > 70% sustained → add brokers
	// - Partitions per broker > 4000 → add brokers
	// - Rebalancing any of these? → topic reassignment with kafka-reassign-partitions

	fmt.Println("  Storage = rate × size × RF × retention / brokers")
	fmt.Println("  Network: 10 Gbps minimum, account for replication + consumers")
	fmt.Println("  Add brokers when: disk >70%, network >60%, CPU >70%")
	fmt.Println()
}

// =============================================================================
// PART 4: ROLLING UPGRADES
// =============================================================================
func rollingUpgrades() {
	fmt.Println("--- ROLLING UPGRADES ---")

	// ZERO-DOWNTIME UPGRADE PROCEDURE:
	// ─────────────────────────────────
	//
	// 1. PREPARE:
	//    - Read release notes for the target version
	//    - Check for breaking config changes
	//    - Test on staging cluster first (always!)
	//    - Ensure UnderReplicatedPartitions = 0 before starting
	//
	// 2. SET INTER-BROKER PROTOCOL VERSION:
	//    inter.broker.protocol.version=<CURRENT_VERSION>
	//    log.message.format.version=<CURRENT_VERSION>
	//    (In KRaft mode: only inter.broker.protocol.version)
	//    This ensures brokers can still talk during the mixed-version window.
	//
	// 3. ROLLING RESTART (one broker at a time):
	//    For each broker:
	//    a) Stop the broker gracefully (SIGTERM)
	//       - controlled.shutdown.enable=true (default)
	//       - Broker transfers leadership before stopping
	//       - Wait for controlled shutdown to complete
	//    b) Upgrade the binary / docker image
	//    c) Start the broker with new version
	//    d) Wait for it to rejoin the cluster
	//    e) Wait for UnderReplicatedPartitions to return to 0
	//    f) Wait at least 5 minutes (observe stability)
	//    g) Move to next broker
	//
	// 4. UPGRADE PROTOCOL VERSION:
	//    After ALL brokers are on the new version:
	//    - Update inter.broker.protocol.version to new version
	//    - Update log.message.format.version to new version
	//    - Rolling restart again (to apply new protocol)
	//
	// 5. VERIFY:
	//    - All brokers on new version
	//    - No under-replicated partitions
	//    - Produce and consume working normally
	//    - No errors in broker logs
	//
	// ROLLBACK PLAN:
	// ──────────────
	// - Keep the old binary/image available
	// - If issues: stop upgraded broker, revert to old version, start
	// - Works because inter.broker.protocol.version is still old

	fmt.Println("  Upgrade one broker at a time, wait for ISR recovery between each")
	fmt.Println("  Key: set inter.broker.protocol.version to OLD version during upgrade")
	fmt.Println("  Only bump protocol version AFTER all brokers are upgraded")
	fmt.Println()
}

// =============================================================================
// PART 5: INCIDENT RESPONSE — 5 common scenarios
// =============================================================================
func incidentResponse() {
	fmt.Println("--- INCIDENT RESPONSE ---")

	// SCENARIO 1: BROKER WON'T START
	// ────────────────────────────────
	// Check: kafka-server-start.sh logs
	// Common causes:
	// - Port already in use (another process or zombie)
	// - Corrupted log segments → kafka-log-recovery-tools or delete segment
	// - Out of disk space → free space or expand volume
	// - ZK/KRaft connection refused → check ZK/controller cluster
	// Fix: Address root cause, then start. Other brokers cover the partitions.
	//
	// SCENARIO 2: CONSUMER LAG SPIRALING
	// ───────────────────────────────────
	// Check: kafka-consumer-groups.sh --describe --group <group>
	// Common causes:
	// - Consumer processing too slow (downstream bottleneck)
	// - Consumer exceptions causing reprocessing
	// - GC pauses causing session timeouts → rebalance loop
	// - Too few consumers for the throughput
	// Fix: Scale consumers, optimize processing, fix downstream.
	// Emergency: reset offsets to latest (loses data!) if lag is unrecoverable.
	//
	// SCENARIO 3: DISK FULL
	// ─────────────────────
	// URGENT: Broker will refuse writes and potentially crash.
	// Immediate: Reduce retention on large topics:
	//   kafka-configs.sh --alter --entity-type topics --entity-name big-topic \
	//     --add-config retention.ms=3600000  # 1 hour
	// Then: delete old segments, add disk, or add brokers and reassign.
	//
	// SCENARIO 4: LEADER IMBALANCE
	// ────────────────────────────
	// One broker has too many partition leaders → hot broker.
	// Check: kafka-preferred-replica-election.sh or auto.leader.rebalance.enable=true
	// Fix: Trigger preferred leader election:
	//   kafka-leader-election.sh --election-type preferred --all-topic-partitions
	//
	// SCENARIO 5: OUT-OF-SYNC REPLICAS
	// ─────────────────────────────────
	// UnderReplicatedPartitions > 0 for extended period.
	// Check: which broker is out of sync (describe topic)
	// Common causes:
	// - Broker disk slow (check iostat)
	// - Network issue between brokers
	// - Broker overloaded (too many partitions)
	// - GC pauses on the follower
	// Fix: Address the slow broker. If unrecoverable, replace the broker
	// and let replication catch up.

	fmt.Println("  Broker won't start: check logs, port, disk, ZK/KRaft connection")
	fmt.Println("  Consumer lag: scale consumers, fix processing, check GC")
	fmt.Println("  Disk full: reduce retention immediately, then add capacity")
	fmt.Println("  Leader imbalance: trigger preferred leader election")
	fmt.Println("  Out-of-sync: check disk I/O, network, GC on the slow broker")
	fmt.Println()
}




































































































































































































































}	fmt.Println("  Leader imbalance: kafka-leader-election --preferred")	fmt.Println("  Disk full: lower retention (DON'T delete files manually!)")	fmt.Println("  Lag spiraling: scale consumers, fix rebalances, or seek to latest")	fmt.Println("  Broker won't start: check logs, disk, corrupted segments")	// or replace the broker's disks.	// Fix: increase num.replica.fetchers, reduce partition count on broker,	// Check: GC pauses? (GC logs)	// Check: is the broker overloaded? (CPU, disk, network)	// Check: is the follower fetching? (check replica fetcher metrics)	// ────────────────────────────────	// SCENARIO 5: OUT-OF-SYNC REPLICAS	//	// Or: auto.leader.rebalance.enable = true (default, checks every 5 min)	// Fix: kafka-leader-election.sh --all-topic-partitions --election-type preferred	// After a broker failure and recovery, leaders may be unevenly distributed.	// ──────────────────────────────────────	// SCENARIO 4: PARTITION LEADER IMBALANCE	//	// 5. Add more disks to log.dirs	// 4. If emergency: kafka-delete-records.sh --topic X --offset-json-file offsets.json	// 3. Wait for log cleaner to delete old segments	// 2. Lower retention: kafka-configs.sh --alter --topic X --add-config retention.ms=3600000	// 1. DO NOT delete data files manually!	// ───────────────────────────────	// SCENARIO 3: DISK FULL ON BROKER	//	// 4. Nuclear: seek consumers to latest offset (lose backlog)	//    long processing time, session timeout	// 3. Growing: frequent rebalances? → check for consumer crashes, 	// 2. Growing: producer rate increased? → scale consumers	// 1. Check if lag is growing or stable (stable = consumer is just slow)	// ──────────────────────────────────	// SCENARIO 2: CONSUMER LAG SPIRALING	//	// from replicas.	// Common: corrupted segment file → move corrupted segment, broker recovers	// Check: Kafka server.log, disk space, ZooKeeper/KRaft connectivity	// ───────────────────────────────	// SCENARIO 1: BROKER WON'T START	fmt.Println("--- INCIDENT RESPONSE ---")func incidentResponse() {}	fmt.Println()	fmt.Println("  controlled.shutdown.enable=true for leader migration before stop")	fmt.Println("  Keep inter.broker.protocol.version = OLD during rolling upgrade")	fmt.Println("  Rolling upgrade: one broker at a time, check ISR after each")	// TOTAL: For 20 brokers × 15 min = 5 hours. Plan accordingly.	// TIME PER BROKER: 5-30 minutes depending on partition count and ISR catch-up.	//	//    Rolling restart again to pick up new protocol version.	//    Update inter.broker.protocol.version to new version.	// 7. AFTER ALL BROKERS UPGRADED:	//	//    Only then proceed to next broker.	//    Wait for ISR to be full for all partitions on this broker.	// 6. VERIFY: Wait for UnderReplicatedPartitions = 0	//	//    This ensures brokers can still communicate during mixed-version state.	//    Use inter.broker.protocol.version = OLD version during rolling upgrade	// 5. START: Start the broker with new version	//	// 4. UPGRADE: Install new Kafka version, update configs	//	//    controlled.shutdown.max.retries = 3	//    This migrates leaders BEFORE shutdown → minimal unavailability!	//    controlled.shutdown.enable = true (default)	// 3. STOP: Gracefully stop the broker	//	//    kafka-leader-election.sh --all-topic-partitions  (after restart)	//    Move leaders away from this broker before stopping it.	// 2. DRAIN (optional, reduces impact):	//	// 1. CHECK: Ensure UnderReplicatedPartitions = 0 cluster-wide	//	// For each broker (one at a time):	// ─────────────────────────────────	// ZERO-DOWNTIME UPGRADE PROCEDURE:	fmt.Println("--- ROLLING UPGRADES ---")func rollingUpgrades() {}	fmt.Println()	fmt.Println("  Rule: 50-100 MB/s ingress per broker, 4K partitions max")	fmt.Println("  brokers = max(network_need, disk_need, partition_need)")	// - Network: at least 10 Gbps for production clusters	// - Each broker: 1-2 TB per disk, 4-12 disks	// - Each broker: 4000 partitions max	// - Each broker: 50-100 MB/s ingress sustained	// RULE OF THUMB:	//	// Result: Need 90 brokers (disk-bound). Consider 10Gb NICs or shorter retention.	//	// - Brokers needed (disk): 900 / 10 = 90 brokers	// - Per broker disk: 10 TB	// - Data: 500 MB/s × 86400s × 7 days retention × 3 RF = ~900 TB	//	// - Brokers needed (network): 2000 / 100 = 20 brokers	// - Per broker network: 1 Gbps = 125 MB/s (with other traffic, ~100 MB/s usable)	// - Total internal network: 500 + 1000 + egress = ~2000 MB/s	// - Replication: 500 MB/s × 2 (RF=3, two followers) = 1000 MB/s between brokers	// - Ingress: 500 MB/s	// EXAMPLE:	//	// )	//   totalPartitions / maxPartitionsPerBroker	//   totalDiskGB / perBrokerDiskGB,	//   totalIngressMB / perBrokerNetworkCapacity * safetyFactor,	// brokers = max(	// ─────────────────────────	// FORMULA FOR BROKER COUNT:	fmt.Println("--- CAPACITY PLANNING ---")func capacityPlanning() {}	fmt.Println()	fmt.Println("  Trend alert on: growing lag, disk usage above 80%")	fmt.Println("  Always alert on: offline partitions, no controller, under min ISR")	// └──────────────────────────────────────────────────────────────────┘	// │  RequestHandler idle < 0.3│ Sustained 10 min   │ LOW             │	// │  Network util > 70%       │ Sustained 10 min   │ MEDIUM          │	// │  ISR shrinks > 10/min     │ Sustained 10 min   │ MEDIUM          │	// │  Disk usage > 80%         │ Any broker          │ MEDIUM          │	// │  Produce latency p99      │ > 500ms for 5 min  │ HIGH            │	// │  Consumer lag growing     │ > 10 min increasing│ HIGH            │	// │  UnderReplicated > 0      │ > 5 minutes        │ HIGH            │	// │  No active controller     │ > 30 seconds       │ CRITICAL (page) │	// │  UnderMinIsrPartitions > 0│ > 1 minute         │ CRITICAL (page) │	// │  OfflinePartitions > 0    │ Immediate          │ CRITICAL (page) │	// │  ─────────────────────────┼────────────────────┼─────────────────│	// │  ALERT                    │ THRESHOLD          │ SEVERITY        │	// ┌──────────────────────────────────────────────────────────────────┐	fmt.Println("--- ALERTING RULES ---")func alertingRules() {}	fmt.Println()	fmt.Println("  CAPACITY: disk, network, CPU trends (review weekly)")	fmt.Println("  WARNING: consumer lag, request latency p99 > 100ms, ISR shrinks")	fmt.Println("  RED ALERT: UnderReplicated, UnderMinISR, Offline, no Controller")	// messages in per second, bytes in/out per second.	// Disk usage per broker, network utilization, CPU utilization,	// ─────────────────────────────────────────────────	// TIER 3: CAPACITY PLANNING METRICS (weekly review)	//	//   If increasing: disk is getting saturated.	//   kafka.log:type=LogFlushStats,name=LogFlushRateAndTimeMs	// Log flush latency:	//	//   Frequent ISR shrinks → broker health issues.	//   kafka.server:type=ReplicaManager,name=IsrShrinksPerSec	// ISR Shrink/Expand Rate:	//	//   p99 > 100ms → something is slow.	//   kafka.network:type=RequestMetrics,name=TotalTimeMs,request=FetchConsumer	//   kafka.network:type=RequestMetrics,name=TotalTimeMs,request=Produce	// Request latency:	//	//   Or: external monitoring (Burrow, Prometheus kafka-exporter)	//   kafka.consumer:type=consumer-fetch-manager-metrics,name=records-lag-max	// Consumer Lag:	//	// ──────────────────────────────────────────────────────────	// TIER 2: WARNING METRICS (investigate during business hours)	//	//   If no broker has 1: cluster has no controller → cannot recover from failures.	//   Exactly ONE broker should have value=1. All others should be 0.	//   kafka.controller:type=KafkaController,name=ActiveControllerCount	// ActiveControllerCount (per broker):	//	//   If > 0: partitions have NO LEADER. Complete outage for those partitions.	//   kafka.controller:type=KafkaController,name=OfflinePartitionsCount	// OfflinePartitionsCount (controller only):	//	//   This is an ACTIVE OUTAGE for write operations.	//   If > 0: partitions can't accept writes (ISR < min.insync.replicas)!	//   kafka.server:type=ReplicaManager,name=UnderMinIsrPartitionCount	// UnderMinIsrPartitionCount (per broker):	//	//   Causes: broker overload, disk failure, network partitioning.	//   Should be 0. If > 0: replica is falling behind → data at risk.	//   kafka.server:type=ReplicaManager,name=UnderReplicatedPartitions	// UnderReplicatedPartitions (per broker):	//	// ──────────────────────────────────────────────────	// TIER 1: RED ALERT METRICS (page someone at 3 AM)	fmt.Println("--- CRITICAL METRICS ---")func criticalMetrics() {}	incidentResponse()	rollingUpgrades()	capacityPlanning()	alertingRules()	criticalMetrics()	fmt.Println()	fmt.Println("=== PRODUCTION OPERATIONS ===")func main() {import "fmt"package main// =============================================================================//// - Incident response: common failures and how to fix them// - Partition reassignment: rebalancing data across brokers// - Rolling upgrades: zero-downtime broker upgrades  // - Capacity planning: when to add brokers// - Alerting rules that actually matter// - Critical metrics to monitor (and what they mean)// WHAT YOU'LL LEARN://// =============================================================================// LESSON 12.1: PRODUCTION OPERATIONS — Running Kafka Without Losing Sleep// =============================================================================