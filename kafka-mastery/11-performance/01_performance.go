//go:build ignore
// =============================================================================
// LESSON 11.1: PERFORMANCE — Benchmarking, Tuning, and Squeezing Every Drop
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Benchmarking Kafka: tools, methodology, what "good" looks like
// - Bottleneck analysis: disk, network, CPU, memory, request handlers
// - Broker tuning cheat sheet: the configs that matter most
// - OS tuning: page cache, file descriptors, network stack
// - JVM tuning: heap size, GC, and the configs that reduce pauses
// - Production performance checklist
//
// THE KEY INSIGHT:
// Kafka is already fast out of the box. Most performance problems come from
// misconfiguration, not Kafka's architecture. The key is to identify YOUR
// bottleneck (disk? network? consumer processing?) and tune specifically for it.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== PERFORMANCE MASTERY ===")
	fmt.Println()

	benchmarkingKafka()
	bottleneckAnalysis()
	brokerTuningCheatSheet()
	osTuning()
	jvmTuning()
	productionChecklist()
}

// =============================================================================
// PART 1: BENCHMARKING KAFKA
// =============================================================================
func benchmarkingKafka() {
	fmt.Println("--- BENCHMARKING KAFKA ---")

	// BUILT-IN TOOLS:
	// ────────────────
	// kafka-producer-perf-test.sh:
	//   Tests producer throughput and latency.
	//   bin/kafka-producer-perf-test.sh \
	//     --topic perf-test \
	//     --num-records 10000000 \
	//     --record-size 1024 \
	//     --throughput -1 \    # unlimited
	//     --producer-props \
	//       bootstrap.servers=localhost:9092 \
	//       acks=all \
	//       batch.size=65536 \
	//       linger.ms=10 \
	//       compression.type=zstd
	//
	// kafka-consumer-perf-test.sh:
	//   Tests consumer throughput.
	//   bin/kafka-consumer-perf-test.sh \
	//     --bootstrap-server localhost:9092 \
	//     --topic perf-test \
	//     --messages 10000000 \
	//     --threads 1
	//
	// BENCHMARKING RULES:
	// ────────────────────
	// 1. Run tests FROM OUTSIDE the broker machine (not localhost)
	// 2. Use realistic message sizes (your actual workload)
	// 3. Test with production configuration (RF=3, acks=all, etc.)
	// 4. Run multiple times and take the median
	// 5. Test at different concurrency levels (1, 5, 10 producers)
	// 6. Measure BOTH throughput and latency percentiles (p50, p95, p99)
	// 7. Test sustained load (not just burst) for at least 10 minutes
	//
	// WHAT "GOOD" LOOKS LIKE (per broker, modern hardware):
	// ─────────────────────────────────────────────────────
	// ┌────────────────────────────────────────────────────────────┐
	// │ Metric                │ Good          │ Excellent           │
	// ├────────────────────────────────────────────────────────────┤
	// │ Producer throughput   │ 100-500 MB/s  │ 500 MB/s - 1 GB/s  │
	// │ Consumer throughput   │ 200-600 MB/s  │ 600 MB/s - 2 GB/s  │
	// │ Produce latency p50  │ 2-5 ms        │ < 2 ms              │
	// │ Produce latency p99  │ 10-50 ms      │ < 10 ms             │
	// │ End-to-end latency   │ 5-20 ms       │ < 5 ms              │
	// │ Records/sec (1KB)    │ 100K-500K     │ 500K-1M+            │
	// └────────────────────────────────────────────────────────────┘
	//
	// Consumer is usually faster than producer because:
	// - Zero-copy (sendfile) avoids data copy to userspace
	// - Sequential reads hit page cache (no disk seek)
	// - No replication overhead (that's the producer's / broker's job)

	fmt.Println("  Use kafka-producer-perf-test and kafka-consumer-perf-test")
	fmt.Println("  Test with production config (RF=3, acks=all, real message sizes)")
	fmt.Println("  Good: 100-500 MB/s producer, <50ms p99 latency")
	fmt.Println()
}

// =============================================================================
// PART 2: BOTTLENECK ANALYSIS
// =============================================================================
func bottleneckAnalysis() {
	fmt.Println("--- BOTTLENECK ANALYSIS ---")

	// ┌──────────────────────────────────────────────────────────────┐
	// │  BOTTLENECK           │ SYMPTOM                    │ FIX     │
	// ├──────────────────────────────────────────────────────────────┤
	// │  DISK I/O             │ High iowait, slow produce  │ SSDs,   │
	// │                       │ high request latency       │ more    │
	// │                       │ iostat shows 100% busy     │ brokers │
	// │                       │                            │         │
	// │  NETWORK              │ High NIC utilization,      │ 10GbE+, │
	// │                       │ replication lag, consumer  │ compress│
	// │                       │ throughput plateaus         │ more    │
	// │                       │                            │ brokers │
	// │                       │                            │         │
	// │  CPU                  │ High compression CPU,      │ zstd→   │
	// │                       │ SSL/TLS overhead,          │ lz4 or  │
	// │                       │ many small messages        │ snappy, │
	// │                       │                            │ batch   │
	// │                       │                            │         │
	// │  MEMORY (Page Cache)  │ Cold reads (fetch from     │ More    │
	// │                       │ disk instead of cache),    │ RAM,    │
	// │                       │ consumer lag causes old    │ reduce  │
	// │                       │ segment reads              │ retention│
	// │                       │                            │         │
	// │  REQUEST HANDLERS     │ All handler threads busy,  │ Increase│
	// │                       │ request queue growing,     │ num.io. │
	// │                       │ high RequestQueueTimeMs    │ threads │
	// └──────────────────────────────────────────────────────────────┘
	//
	// HOW TO IDENTIFY YOUR BOTTLENECK:
	// ─────────────────────────────────
	// 1. Check disk: iostat -x 1 → if %util > 80% → disk bound
	// 2. Check network: sar -n DEV 1 → if NIC > 70% capacity → network bound
	// 3. Check CPU: top/htop → if kafka process > 80% → CPU bound
	// 4. Check page cache: free -m → if available < active data set → memory bound
	// 5. Check request handlers: RequestHandlerAvgIdlePercent < 0.3 → handler bound
	//
	// MOST COMMON BOTTLENECK: DISK (spinning disks) or NETWORK (10Gbps NIC shared).
	// With SSDs: usually network-bound before disk-bound.

	fmt.Println("  Most common: disk (HDD) or network (shared NIC)")
	fmt.Println("  Monitor: iostat, sar -n DEV, RequestHandlerAvgIdlePercent")
	fmt.Println("  With SSDs: usually network-bound first")
	fmt.Println()
}

// =============================================================================
// PART 3: BROKER TUNING CHEAT SHEET
// =============================================================================
func brokerTuningCheatSheet() {
	fmt.Println("--- BROKER TUNING CHEAT SHEET ---")

	// ┌───────────────────────────────────────────────────────────────────────┐
	// │ CONFIG                          │ DEFAULT  │ RECOMMENDED             │
	// ├───────────────────────────────────────────────────────────────────────┤
	// │ num.network.threads             │ 3        │ 8 (CPUs dedicated to    │
	// │                                 │          │  network I/O)           │
	// │ num.io.threads                  │ 8        │ 16 (CPUs for disk I/O)  │
	// │ socket.send.buffer.bytes        │ 102400   │ 1048576 (1MB)           │
	// │ socket.receive.buffer.bytes     │ 102400   │ 1048576 (1MB)           │
	// │ socket.request.max.bytes        │ 100MB    │ 100MB (fine)            │
	// │ num.partitions                  │ 1        │ 6-12 (default for new   │
	// │                                 │          │  topics, override per)  │
	// │ log.retention.hours             │ 168 (7d) │ Based on use case       │
	// │ log.segment.bytes               │ 1GB      │ 1GB (fine for most)     │
	// │ log.retention.check.interval.ms│ 300000   │ 60000 (check more often)│
	// │ replica.fetch.max.bytes         │ 1MB      │ 10MB (if large messages)│
	// │ message.max.bytes               │ 1MB      │ Match producer's max    │
	// │ compression.type                │ producer │ producer (KEEP THIS!)   │
	// │ unclean.leader.election.enable  │ false    │ false (data safety)     │
	// │ auto.create.topics.enable       │ true     │ FALSE in production!    │
	// │ min.insync.replicas             │ 1        │ 2 (with RF=3)           │
	// │ default.replication.factor      │ 1        │ 3                       │
	// └───────────────────────────────────────────────────────────────────────┘
	//
	// THE Most-Missed Config: auto.create.topics.enable=false
	// ────────────────────────────────────────────────────────
	// When true (default!), producing to a non-existent topic creates it
	// with default settings (RF=1, 1 partition). This leads to:
	// - Topics with RF=1 in production (NO replication!)
	// - Typos creating garbage topics
	// - Uncontrolled topic proliferation
	// ALWAYS set to false. Create topics explicitly with proper configs.

	fmt.Println("  auto.create.topics.enable=FALSE (most-missed config!)")
	fmt.Println("  num.io.threads=16, num.network.threads=8")
	fmt.Println("  default.replication.factor=3, min.insync.replicas=2")
	fmt.Println()
}

// =============================================================================
// PART 4: OS TUNING
// =============================================================================
func osTuning() {
	fmt.Println("--- OS TUNING ---")

	// Kafka relies HEAVILY on the OS page cache. Tuning the OS is
	// as important as tuning Kafka itself.
	//
	// CRITICAL OS SETTINGS:
	// ─────────────────────
	//
	// 1. vm.swappiness = 1
	//    Tells Linux to avoid swapping to disk.
	//    Kafka should NEVER swap — it destroys latency.
	//    0 = no swap (risky — OOM killer may strike)
	//    1 = minimal swap (safest for Kafka)
	//    echo "vm.swappiness=1" >> /etc/sysctl.conf
	//
	// 2. vm.dirty_ratio = 60 (default: 20)
	//    vm.dirty_background_ratio = 5 (default: 10)
	//    Controls when dirty pages are flushed to disk.
	//    Kafka manages its own flushing — let it accumulate more dirty pages.
	//    Higher dirty_ratio = more data in page cache before forced flush.
	//
	// 3. File descriptors: ulimit -n 100000
	//    Kafka opens MANY files: log segments, indexes, network connections.
	//    Default (1024) is WAY too low.
	//    Formula: (partitions × segments_per_partition × 3) + network connections
	//    Safe: 100,000+ per broker process.
	//    Set in /etc/security/limits.conf:
	//    kafka  soft  nofile  100000
	//    kafka  hard  nofile  100000
	//
	// 4. Filesystem: XFS (recommended) or ext4
	//    XFS: better for large files, parallel allocation
	//    ext4: fine but slightly worse under heavy parallel writes
	//    NEVER: btrfs, ZFS (too much overhead for Kafka's workload)
	//
	// 5. Mount options: noatime
	//    mount -o noatime /dev/sda1 /kafka-data
	//    Disables access time tracking — reduces unnecessary writes.
	//
	// 6. Network: increase socket buffers
	//    net.core.wmem_default = 2097152
	//    net.core.rmem_default = 2097152
	//    net.core.wmem_max = 2097152
	//    net.core.rmem_max = 2097152
	//    net.ipv4.tcp_wmem = 4096 65536 2048000
	//    net.ipv4.tcp_rmem = 4096 65536 2048000

	fmt.Println("  vm.swappiness=1 (never swap)")
	fmt.Println("  ulimit -n 100000 (file descriptors)")
	fmt.Println("  XFS + noatime mount option")
	fmt.Println("  Increase network socket buffers")
	fmt.Println()
}

// =============================================================================
// PART 5: JVM TUNING
// =============================================================================
func jvmTuning() {
	fmt.Println("--- JVM TUNING ---")

	// Kafka brokers run on the JVM. JVM tuning = fewer GC pauses =
	// more predictable latency.
	//
	// HEAP SIZE:
	// ──────────
	// KAFKA_HEAP_OPTS="-Xmx6g -Xms6g"
	//
	// 6 GB is the sweet spot for most deployments.
	// WHY NOT MORE?
	// - Kafka stores data in PAGE CACHE, not JVM heap
	// - Larger heap = longer GC pauses
	// - 6 GB covers: request handling, metadata, connections, batching
	// - Leave the rest of RAM for page cache!
	//
	// Example: 32 GB server → 6 GB heap + 26 GB page cache
	// The page cache is what makes Kafka fast (sequential reads from memory).
	//
	// GC SETTINGS:
	// ─────────────
	// Use G1GC (default since Java 9, recommended for Kafka):
	//
	// KAFKA_JVM_PERFORMANCE_OPTS="
	//   -XX:+UseG1GC
	//   -XX:MaxGCPauseMillis=20
	//   -XX:InitiatingHeapOccupancyPercent=35
	//   -XX:+ExplicitGCInvokesConcurrent
	//   -XX:G1HeapRegionSize=16M
	//   -XX:MetaspaceSize=96m
	//   -XX:MinMetaspaceFreeRatio=50
	//   -XX:MaxMetaspaceFreeRatio=80
	// "
	//
	// KEY SETTINGS:
	// MaxGCPauseMillis=20
	//   Target maximum GC pause time. G1 tries to stay under this.
	//   20ms means GC pauses rarely affect produce/consume latency.
	//
	// InitiatingHeapOccupancyPercent=35
	//   Start concurrent GC cycle when heap is 35% full.
	//   Lower than default (45) → GC starts earlier → less risk of
	//   full GC (which causes long pauses).
	//
	// MONITORING GC:
	// ──────────────
	// Enable GC logging:
	//   -Xlog:gc*:file=/var/log/kafka/gc.log:time,tags:filecount=10,filesize=100M
	//
	// Watch for:
	// - Full GC events (should NEVER happen in steady state)
	// - GC pause > 100ms (indicates heap too large or GC misconfigured)
	// - Allocation rate spikes (indicates a code path creating too many objects)

	fmt.Println("  Heap: 6 GB (-Xmx6g -Xms6g) — leave rest for page cache")
	fmt.Println("  G1GC with MaxGCPauseMillis=20")
	fmt.Println("  Full GC = something is WRONG (investigate immediately)")
	fmt.Println()
}

// =============================================================================
// PART 6: PRODUCTION PERFORMANCE CHECKLIST
// =============================================================================
func productionChecklist() {
	fmt.Println("--- PRODUCTION PERFORMANCE CHECKLIST ---")

	// ┌──────────────────────────────────────────────────────────────┐
	// │  BEFORE GO-LIVE CHECKLIST:                                    │
	// │                                                              │
	// │  Hardware:                                                     │
	// │  □ SSDs for Kafka data directories (NVMe preferred)          │
	// │  □ 10 Gbps+ network (25 Gbps for heavy workloads)           │
	// │  □ 32+ GB RAM per broker (6 GB heap + page cache)            │
	// │  □ 8+ CPU cores per broker                                    │
	// │  □ Dedicated disks for Kafka (not shared with OS)            │
	// │                                                              │
	// │  Kafka Config:                                                │
	// │  □ auto.create.topics.enable=false                            │
	// │  □ default.replication.factor=3                               │
	// │  □ min.insync.replicas=2                                      │
	// │  □ unclean.leader.election.enable=false                       │
	// │  □ compression.type=producer (topic level)                    │
	// │  □ num.io.threads=16, num.network.threads=8                  │
	// │                                                              │
	// │  OS Config:                                                    │
	// │  □ vm.swappiness=1                                            │
	// │  □ ulimit -n 100000                                           │
	// │  □ XFS filesystem with noatime                                │
	// │  □ Network socket buffers increased                           │
	// │                                                              │
	// │  JVM Config:                                                   │
	// │  □ -Xmx6g -Xms6g                                             │
	// │  □ G1GC with MaxGCPauseMillis=20                              │
	// │  □ GC logging enabled                                         │
	// │                                                              │
	// │  Producer Config:                                              │
	// │  □ enable.idempotence=true                                    │
	// │  □ acks=all                                                    │
	// │  □ compression.type=zstd                                      │
	// │  □ batch.size tuned for workload (64KB-1MB)                  │
	// │  □ linger.ms tuned (0 for latency, 20-200 for throughput)    │
	// │                                                              │
	// │  Consumer Config:                                              │
	// │  □ fetch.min.bytes tuned (1 for latency, 1MB for throughput) │
	// │  □ max.poll.interval.ms set correctly                         │
	// │  □ session.timeout.ms = 45 seconds (Kafka 3.0+)              │
	// │                                                              │
	// │  Monitoring:                                                   │
	// │  □ Under-replicated partitions alert (> 0 = red)             │
	// │  □ Consumer lag monitoring                                    │
	// │  □ Request latency percentiles (p99 < 100ms)                 │
	// │  □ Disk usage and growth rate                                 │
	// │  □ Network utilization per broker                             │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  Hardware: SSDs, 10GbE+, 32GB+ RAM, 8+ cores")
	fmt.Println("  Config: RF=3, min.ISR=2, auto.create=false, G1GC 6GB heap")
	fmt.Println("  Monitor: under-replicated partitions, consumer lag, p99 latency")
	fmt.Println()
}










































































































































































































































































































}	fmt.Println("  Monitor: under-replicated partitions, consumer lag, disk/network")	fmt.Println("  JVM=6GB, G1GC, swappiness=1, noatime, 128K file descriptors")	fmt.Println("  RF=3, min.ISR=2, unclean=false, auto.create=false")	// └──────────────────────────────────────────────────────────────┘	// │  [x] Schema Registry for serialization governance             │	// │  [x] ACLs for topic-level access control                      │	// │  [x] Quotas configured for multi-tenant clusters              │	// │  [x] auto.leader.rebalance.enable = true                      │	// │  [x] log.dirs = multiple disks (JBOD)                         │	// │  [x] Alerting: consumer lag, disk usage, network saturation   │	// │  [x] Monitoring: UnderReplicatedPartitions, ISR metrics        │	// │  [x] log.retention set appropriately                          │	// │  [x] Compression = zstd (producer side)                       │	// │  [x] File descriptors ≥ 128K                                  │	// │  [x] noatime mount option                                     │	// │  [x] vm.swappiness = 1                                        │	// │  [x] G1GC with MaxGCPauseMillis=20                            │	// │  [x] JVM heap = 6 GB (not more!)                              │	// │  [x] default.replication.factor = 3                           │	// │  [x] auto.create.topics.enable = false                        │	// │  [x] unclean.leader.election.enable = false                   │	// │  [x] min.insync.replicas = 2                                  │	// │  [x] replication.factor = 3                                   │	// │                                                              │	// │  PRODUCTION TUNING CHECKLIST:                                  │	// ┌──────────────────────────────────────────────────────────────┐	fmt.Println("--- PRODUCTION CHECKLIST ---")func productionChecklist() {}	fmt.Println()	fmt.Println("  GC pauses must be << replica.lag.time.max.ms")	fmt.Println("  G1GC with MaxGCPauseMillis=20")	fmt.Println("  Heap: 4-6 GB (not more — save RAM for page cache)")	//   "	//     -Djava.awt.headless=true	//     -XX:+ExplicitGCInvokesConcurrent	//     -XX:MaxMetaspaceFreeRatio=80	//     -XX:MinMetaspaceFreeRatio=50	//     -XX:MetaspaceSize=96m	//     -XX:G1HeapRegionSize=16M	//     -XX:InitiatingHeapOccupancyPercent=35	//     -XX:MaxGCPauseMillis=20	//     -XX:+UseG1GC	//   KAFKA_JVM_PERFORMANCE_OPTS="	//   KAFKA_HEAP_OPTS="-Xms6g -Xmx6g"	// FULL JVM FLAGS (production example):	//	// If GC pauses exceed replica.lag.time.max.ms → follower drops from ISR.	// Key: Keep GC pauses < 20ms to avoid ISR shrinks!	//	//   -XX:InitiatingHeapOccupancyPercent=35  # Start concurrent GC earlier	//   -XX:G1HeapRegionSize=16M     # Region size (for 6 GB heap)	//   -XX:MaxGCPauseMillis=20      # Target max GC pause	//   -XX:+UseG1GC	// Kafka 3.x+ recommends G1GC (default):	// ──────────	// GC CHOICE:	//	// DO NOT give Kafka 32 GB heap — you're stealing page cache!	// Recommended: 4-6 GB heap for production brokers.	// Kafka barely uses JVM heap (data flows through page cache, not heap).	// ──────────	// HEAP SIZE:	fmt.Println("--- JVM TUNING ---")func jvmTuning() {}	fmt.Println()	fmt.Println("  XFS with noatime, JBOD or RAID-10")	fmt.Println("  Raise file descriptors to 128K+")	fmt.Println("  vm.swappiness=1, vm.dirty_background_ratio=5")	// Kafka benefits from THP for large sequential allocations.	// DO NOT disable THP for Kafka (unlike databases).	// TRANSPARENT HUGE PAGES (THP):	//	// - Avoid RAID-5/6 (bad write amplification). Use RAID-10 or JBOD.	// - XFS: better for Kafka's sequential write patterns	// - XFS or ext4 with: noatime,nodiratime	// ───────────	// FILESYSTEM:	//	// # kafka hard nofile 128000	// # kafka soft nofile 128000	// # Also set in /etc/security/limits.conf:	// fs.file-max = 1000000              # System-wide max open files	// # File descriptors	//	// net.ipv4.tcp_max_syn_backlog = 8096	// net.core.netdev_max_backlog = 5000 # Packet queue	// net.ipv4.tcp_rmem = 4096 65536 2097152	// net.ipv4.tcp_wmem = 4096 65536 2097152	// net.core.rmem_max = 2097152        # Max socket read buffer (2 MB)	// net.core.wmem_max = 2097152        # Max socket write buffer (2 MB)	// # Networking	//	// vm.dirty_writeback_centisecs = 500  # Flush interval (5 seconds)	// vm.dirty_ratio = 80                # Block writes at 80% dirty pages	// vm.dirty_background_ratio = 5      # Start flushing at 5% dirty pages	// vm.swappiness = 1                  # Minimze swap (don't use 0 → OOM killer)	// # Virtual memory	// ────────────────────────	// LINUX KERNEL PARAMETERS:	fmt.Println("--- OS TUNING ---")func osTuning() {}	fmt.Println()	fmt.Println("  Let producers choose compression (compression.type=producer)")	fmt.Println("  Key: num.io.threads ≥ disks, no fsync, RF=3, min.ISR=2")	// └──────────────────────────────────────────────────────────────────┘	// │  group.initial.rebalance.delay.ms = 3000 (3s, for rolling deploy)│	// │  offsets.retention.minutes = 10080 (7 days)                       │	// │  CONSUMER FACING:                                                  │	// │                                                                  │	// │  compression.type = producer (let producer choose)                │	// │  message.max.bytes = 10485760 (10 MB, match max.request.size)    │	// │  PRODUCER FACING:                                                  │	// │                                                                  │	// │  replica.lag.time.max.ms = 30000 (default)                       │	// │  unclean.leader.election.enable = false                           │	// │  min.insync.replicas = 2                                          │	// │  default.replication.factor = 3                                    │	// │  REPLICATION:                                                      │	// │                                                                  │	// │  log.flush.interval.ms = Don't set                                │	// │  log.flush.interval.messages = Don't set (use replication)        │	// │  log.retention.bytes = -1 (unlimited, default)                    │	// │  log.retention.hours = 168 (7 days, default)                     │	// │  log.segment.bytes = 1073741824 (1 GB, default)                  │	// │  LOG:                                                              │	// │                                                                  │	// │  socket.request.max.bytes = 104857600 (100 MB)                   │	// │  socket.receive.buffer.bytes = 1048576 (1 MB)                    │	// │  socket.send.buffer.bytes = 1048576 (1 MB, increase for WAN)     │	// │  NETWORKING:                                                       │	// │                                                                  │	// │  log.cleaner.threads = 1-2 (increase if many compacted topics)   │	// │  num.replica.fetchers = 2-4 (increase if ISR problems)           │	// │  num.io.threads = 8-16 (default 8, ≥ disks)                      │	// │  num.network.threads = 3-8 (default 3, rarely needs changing)    │	// │  THREADING:                                                        │	// │                                                                  │	// │  BROKER TUNING CHEAT SHEET:                                       │	// ┌──────────────────────────────────────────────────────────────────┐	fmt.Println("--- BROKER TUNING ---")func brokerTuning() {}	fmt.Println()	fmt.Println("  Memory: page cache too small → more RAM, less JVM heap")	fmt.Println("  CPU: > 80% → SSL/compression bottleneck, offload to producer")	fmt.Println("  Network: TX/RX near limit → better compression, more brokers")	fmt.Println("  Disk: iostat %util > 90% → SSD or more disks")	// └──────────────────────────────────────────────────────────────┘	// │  - Fix: increase num.network.threads (unusual bottleneck)    │	// │  - Symptom: connection timeouts, request queueing            │	// │  - Check: NetworkProcessorAvgIdlePercent < 0.3               │	// │  NETWORK THREAD BOTTLENECK:                                    │	// │                                                              │	// │  - Fix: increase num.io.threads, add brokers                 │	// │  - Symptom: produce/consume latency spikes                   │	// │  - Check: RequestHandlerAvgIdlePercent < 0.3                  │	// │  REQUEST HANDLER BOTTLENECK:                                   │	// │                                                              │	// │         fewer partitions per broker                           │	// │  - Fix: more RAM, reduce JVM heap (give more to page cache), │	// │  - Symptom: consumer reads from disk, not page cache         │	// │  - Check: free -h → page cache < expected working set        │	// │  MEMORY BOTTLENECK:                                           │	// │                                                              │	// │         more brokers to distribute load                       │	// │  - Fix: offload compression to producers, hardware crypto,   │	// │  - Symptom: usually from SSL/TLS or compression on broker    │	// │  - Check: top → Kafka process using > 80% CPU                │	// │  CPU BOTTLENECK:                                              │	// │                                                              │	// │         factor (not recommended), add brokers                 │	// │  - Fix: better compression, 10G/25G NICs, reduce replication │	// │  - Symptom: produce timeout, consumer lag across all topics  │	// │  - Check: sar -n DEV 1 → TX/RX near bandwidth limit          │	// │  NETWORK BOTTLENECK:                                          │	// │                                                              │	// │         partitions per broker                                 │	// │  - Fix: faster disks (SSD), more disks (JBOD), reduce        │	// │  - Symptom: high produce latency, follower falling behind    │	// │  - Check: iostat -x 1 → %util > 90% for data disks          │	// │  DISK I/O BOTTLENECK:                                         │	// │                                                              │	// │  BOTTLENECK IDENTIFICATION:                                    │	// ┌──────────────────────────────────────────────────────────────┐	//	// STEP 1: Identify the bottleneck	fmt.Println("--- BOTTLENECK ANALYSIS ---")func bottleneckAnalysis() {}	fmt.Println()	fmt.Println("  Good: 50-100 MB/s per partition, 1-3 GB/s aggregate on 10 brokers")	fmt.Println("  Test for 10+ minutes with production-like configs")	fmt.Println("  Use kafka-producer-perf-test and kafka-consumer-perf-test")	//   Throughput: 1-3 GB/s	// Cluster aggregate (10 brokers, 100 partitions):	//	//   Throughput: 100-300 MB/s (if caught up, from page cache)	// Single consumer, single partition:	//	//   Throughput: 200-500 MB/s (depends on broker network/disk)	// Single producer, many partitions (30+), acks=all:	//	//   p99 latency: 5-15ms	//   Throughput: 50-100 MB/s (50K-100K records/sec)	// Single producer, acks=all, 1 KB records, zstd, single partition:	// ────────────────────────────	// WHAT GOOD NUMBERS LOOK LIKE:	//	// 6. Monitor BROKER metrics during the test (identify bottleneck)	// 5. Test with replication.factor=3 (production config)	// 4. Test with acks=all (not acks=0, that's unrealistic)	// 3. Test throughput AND latency separately (they're different optimizations)	// 2. Run for at least 10 minutes (let caches warm, GC stabilize)	// 1. Test with PRODUCTION-LIKE data (size, keys, compression ratio)	// ───────────────────	// BENCHMARKING RULES:	//	//     --consumer-props bootstrap.servers=localhost:9092	//     --producer-props bootstrap.servers=localhost:9092 \	//     --num-messages 10000 \	//     --topic latency-test \	//   kafka-e2e-latency.sh \	// End-to-end latency:	//	//     --bootstrap-server localhost:9092	//     --messages 10000000 \	//     --topic perf-test \	//   kafka-consumer-perf-test.sh \	// Consumer benchmark:	//	//       linger.ms=10	//       batch.size=65536 \	//       compression.type=zstd \	//       acks=all \	//       bootstrap.servers=localhost:9092 \	//     --producer-props \	//     --throughput -1 \     ← unlimited (max throughput)	//     --record-size 1024 \	//     --num-records 10000000 \	//     --topic perf-test \	//   kafka-producer-perf-test.sh \	// Producer benchmark:	// ────────────────	// BUILT-IN TOOLS:	fmt.Println("--- BENCHMARKING ---")func benchmarking() {}	productionChecklist()	jvmTuning()	osTuning()	brokerTuning()	bottleneckAnalysis()	benchmarking()	fmt.Println()	fmt.Println("=== PERFORMANCE & TUNING ===")func main() {import "fmt"package main// =============================================================================//// - The complete production tuning checklist// - JVM tuning: GC, heap, off-heap// - Broker tuning: threads, memory, OS settings// - Consumer tuning for throughput vs latency// - Producer tuning for throughput vs latency// - Bottleneck analysis: disk, network, CPU, or memory?// - Benchmarking Kafka: the right way to measure throughput and latency// WHAT YOU'LL LEARN://// =============================================================================// LESSON 11.1: PERFORMANCE & TUNING — Squeeze Every Byte/Sec// =============================================================================