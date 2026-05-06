//go:build ignore
// =============================================================================
// LESSON 2.1: THE LOG — Kafka's Storage Engine
// =============================================================================
//
// THIS IS THE MOST IMPORTANT LESSON IN THE ENTIRE CURRICULUM.
// If you understand the log, you understand Kafka.
//
// WHAT YOU'LL LEARN:
// - The append-only log structure (segments, indexes, offsets)
// - How Kafka achieves insane disk throughput (sequential I/O)
// - How reads work: offset → segment → index → file position
// - Page cache: why Kafka lets the OS manage caching
// - Zero-copy: the sendfile syscall that makes consumers fast
// - Retention: time-based, size-based, and compaction
// - Segment lifecycle: active → closed → deleted/compacted
//
// THE KEY INSIGHT:
// Kafka's storage is a SEQUENCE OF SEGMENTS, each segment is a pair of files:
//   .log  — the actual record data (append-only)
//   .index — maps relative offset → physical file position
//   .timeindex — maps timestamp → offset (for time-based lookup)
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== THE LOG: KAFKA'S STORAGE ENGINE ===")
	fmt.Println()

	segmentStructure()
	indexMechanism()
	pageCacheStrategy()
	zeroCopyDeepDive()
	retentionPolicies()
	diskLayoutBestPractices()
}

// =============================================================================
// PART 1: SEGMENT STRUCTURE — How data lives on disk
// =============================================================================
func segmentStructure() {
	fmt.Println("--- SEGMENT STRUCTURE ---")

	// Every partition is a DIRECTORY on disk:
	//   /var/kafka-data/my-topic-0/
	//
	// Inside that directory, data is split into SEGMENTS:
	//   00000000000000000000.log          ← first segment (base offset: 0)
	//   00000000000000000000.index        ← offset index for this segment
	//   00000000000000000000.timeindex    ← timestamp index for this segment
	//   00000000000000065536.log          ← second segment (base offset: 65536)
	//   00000000000000065536.index
	//   00000000000000065536.timeindex
	//   00000000000000131072.log          ← third segment (active)
	//   00000000000000131072.index
	//   00000000000000131072.timeindex
	//
	// FILE NAMING: The filename IS the base offset (20 digits, zero-padded).
	// This enables binary search across segments to find the right file.
	//
	// WHY SEGMENTS?
	// ─────────────
	// 1. RETENTION: Can delete entire segment files (fast! no rewriting)
	// 2. COMPACTION: Can compact old segments without touching active one
	// 3. MEMORY MAPPING: Each segment can be independently mmap'd
	// 4. FILE DESCRIPTORS: Close old segments, keep active one open
	//
	// SEGMENT LIFECYCLE:
	// ┌───────────┐    roll condition     ┌────────────┐    retention    ┌──────────┐
	// │  Active   │ ──────────────────►  │   Closed   │ ────────────► │ Deleted  │
	// │  Segment  │    met               │   Segment  │   policy met  │          │
	// └───────────┘                       └────────────┘               └──────────┘
	//                                          │
	//                                          │ compaction
	//                                          ▼
	//                                    ┌────────────┐
	//                                    │ Compacted  │
	//                                    │ Segment    │
	//                                    └────────────┘
	//
	// SEGMENT ROLL CONDITIONS (new segment created when ANY is true):
	// ──────────────────────────────────────────────────────────────
	// segment.bytes (default: 1 GB)
	//   Active segment reached this size → roll to new segment
	//
	// segment.ms (default: 7 days)
	//   Active segment has been open for this long → roll
	//   WHY THIS MATTERS: If a partition has low throughput, the active
	//   segment stays open forever. This means retention can't delete it!
	//   You might have a topic with retention=7days but data from 30 days
	//   ago because the segment never rolled.
	//   FIX: Set segment.ms = retention.ms (or something reasonable)
	//
	// segment.index.bytes (default: 10 MB)
	//   When the index file reaches this size → roll
	//
	// segment.jitter.ms (default: 0)
	//   Random jitter added to segment.ms to avoid all partitions rolling
	//   segments at the same time (thundering herd on disk I/O)
	//
	// ┌──────────────────────────────────────────────────────────────────┐
	// │              RECORD BATCH FORMAT (v2 / Magic=2)                   │
	// │                                                                  │
	// │  ┌─────────────────────────────────────────────────────────────┐ │
	// │  │ Batch Header                                                │ │
	// │  │ ├── baseOffset: int64        (first offset in batch)       │ │
	// │  │ ├── batchLength: int32       (total bytes)                 │ │
	// │  │ ├── partitionLeaderEpoch: int32                            │ │
	// │  │ ├── magic: int8 (=2)                                       │ │
	// │  │ ├── crc: uint32              (CRC of what follows)         │ │
	// │  │ ├── attributes: int16                                       │ │
	// │  │ │   ├── bits 0-2: compression (0=none,1=gzip,2=snappy,    │ │
	// │  │ │   │              3=lz4,4=zstd)                           │ │
	// │  │ │   ├── bit 3: timestampType (0=create, 1=logAppend)      │ │
	// │  │ │   ├── bit 4: isTransactional                             │ │
	// │  │ │   └── bit 5: isControlBatch                              │ │
	// │  │ ├── lastOffsetDelta: int32                                  │ │
	// │  │ ├── firstTimestamp: int64                                    │ │
	// │  │ ├── maxTimestamp: int64                                      │ │
	// │  │ ├── producerId: int64        (for idempotent/transactional)│ │
	// │  │ ├── producerEpoch: int16                                    │ │
	// │  │ ├── baseSequence: int32      (for idempotent dedup)        │ │
	// │  │ └── recordCount: int32                                      │ │
	// │  ├─────────────────────────────────────────────────────────────┤ │
	// │  │ Records (compressed if attributes say so)                   │ │
	// │  │ ├── Record 0                                                │ │
	// │  │ │   ├── length: varint                                      │ │
	// │  │ │   ├── attributes: int8                                    │ │
	// │  │ │   ├── timestampDelta: varint (from firstTimestamp)        │ │
	// │  │ │   ├── offsetDelta: varint (from baseOffset)              │ │
	// │  │ │   ├── keyLength: varint                                   │ │
	// │  │ │   ├── key: bytes                                          │ │
	// │  │ │   ├── valueLength: varint                                 │ │
	// │  │ │   ├── value: bytes                                        │ │
	// │  │ │   └── headers: [Header]                                   │ │
	// │  │ ├── Record 1                                                │ │
	// │  │ └── ...                                                      │ │
	// │  └─────────────────────────────────────────────────────────────┘ │
	// │                                                                  │
	// │  KEY INSIGHT: Records use VARINTS and DELTAS extensively.        │
	// │  This means small offsets/timestamps cost only 1 byte each.     │
	// │  A batch of 100 records with similar timestamps is TINY.         │
	// └──────────────────────────────────────────────────────────────────┘

	fmt.Println("  Partition = Directory → Segments (.log + .index + .timeindex)")
	fmt.Println("  Segment filename = base offset (20 digits, zero-padded)")
	fmt.Println("  Active segment receives writes; old segments are immutable")
	fmt.Println("  Record batch v2 uses varints + deltas for extreme compactness")
	fmt.Println()
}

// =============================================================================
// PART 2: INDEX MECHANISM — How Kafka finds a record by offset
// =============================================================================
func indexMechanism() {
	fmt.Println("--- INDEX MECHANISM ---")

	// PROBLEM: Consumer asks for "give me data starting at offset 123456"
	// SOLUTION: Two-level lookup
	//
	// STEP 1: Find the RIGHT SEGMENT
	// ─────────────────────────────
	// Segment filenames are base offsets: 0, 65536, 131072, ...
	// Binary search on filenames to find: largest base offset ≤ 123456
	// Result: segment 00000000000000065536 (covers offsets 65536-131071)
	//
	// Wait, 123456 > 131072? Then it's in the next segment.
	// Actually: binary search finds segment with base ≤ target offset.
	//
	// STEP 2: Find the POSITION within the segment
	// ──────────────────────────────────────────────
	// Open the .index file for that segment.
	// The index is a SPARSE index — not every offset is indexed!
	//
	// Index entries are created every log.index.interval.bytes (default: 4096).
	// After every ~4KB of records written, one index entry is added.
	//
	// Index entry format:
	//   relativeOffset: int32  (offset - baseOffset)
	//   position: int32        (physical byte position in .log file)
	//
	// Only 8 bytes per entry! The index is memory-mapped (mmap).
	//
	// STEP 3: Binary search in the index to find the largest entry ≤ target
	//
	// Example index content (base offset: 65536):
	//   relativeOffset=0     → position=0
	//   relativeOffset=120   → position=4096
	//   relativeOffset=245   → position=8192
	//   relativeOffset=380   → position=12288
	//
	// Looking for offset 65600 → relative offset = 64
	// Binary search finds: relativeOffset=0 (position=0)
	//
	// STEP 4: SEQUENTIAL SCAN from that position in the .log file
	// Walk through records until you find offset ≥ target
	//
	// ┌────────────────────────────────────────────────────────────────┐
	// │              OFFSET LOOKUP FLOW                                 │
	// │                                                                │
	// │  Consumer wants offset 123456                                  │
	// │  │                                                             │
	// │  ▼                                                             │
	// │  Binary search segment files  ──►  Segment 65536               │
	// │  │                                                             │
	// │  ▼                                                             │
	// │  mmap'd .index binary search  ──►  Position 12288              │
	// │  │                                                             │
	// │  ▼                                                             │
	// │  Sequential scan .log from 12288 ──► Found offset 123456!      │
	// │  │                                                             │
	// │  ▼                                                             │
	// │  Read records from here, send to consumer                      │
	// │                                                                │
	// │  TOTAL INDEX LOOKUP: O(log N) segments × O(log M) index entries│
	// │  Then O(K) sequential scan where K is very small (~4KB)        │
	// │  In practice: this takes microseconds.                         │
	// └────────────────────────────────────────────────────────────────┘
	//
	// TIME INDEX (.timeindex):
	// ────────────────────────
	// Same concept but maps: timestamp → offset
	// Used when consumer asks "give me data from timestamp X"
	// Entry format: timestamp:int64, offset:int32 (12 bytes per entry)
	//
	// IMPORTANT: Timestamps don't have to be monotonically increasing!
	// If producers have clock skew, timestamps can be out of order
	// within a partition. The time index stores the MAX timestamp
	// seen so far, which means lookups might return slightly older data.
	// This is why time-based consumer seeks are APPROXIMATE.

	fmt.Println("  Lookup: binary search segments → binary search sparse index → sequential scan")
	fmt.Println("  Index is sparse (every ~4KB), mmap'd, 8 bytes per entry")
	fmt.Println("  Time index enables timestamp-based seeks (approximate)")
	fmt.Println()
}

// =============================================================================
// PART 3: PAGE CACHE — Why Kafka doesn't manage its own cache
// =============================================================================
func pageCacheStrategy() {
	fmt.Println("--- PAGE CACHE STRATEGY ---")

	// This is one of Kafka's most counter-intuitive design decisions.
	//
	// TRADITIONAL DATABASE: Manages its own buffer pool (e.g., PostgreSQL shared_buffers)
	//   - Must carefully manage memory
	//   - Double-buffering: data in process cache AND OS page cache
	//   - GC pressure (for JVM-based systems)
	//   - Cache lost on process restart
	//
	// KAFKA: Uses the OS page cache as its ONLY cache
	//   - Kafka writes to a file → OS puts it in page cache
	//   - Consumer reads from file → OS serves from page cache (if warm)
	//   - Kafka broker JVM heap is SMALL (typically 4-6 GB for a terabyte+ of data)
	//   - Cache SURVIVES broker restart!
	//
	// WHY THIS WORKS:
	// ───────────────
	// 1. Kafka's access pattern is almost entirely SEQUENTIAL
	//    - Producers append to the end of the log
	//    - Consumers read sequentially from a recent offset
	//    - The OS read-ahead algorithm is PERFECT for this
	//
	// 2. Most consumer reads are from "the tail" (recent data)
	//    - If consumers are keeping up, they read data that was JUST written
	//    - That data is already in page cache from the producer write
	//    - Cache hit rate for caught-up consumers: ~100%
	//
	// 3. Kafka writes are append-only
	//    - No random writes → no cache invalidation issues
	//    - OS can optimize write-back with large sequential flushes
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │              PAGE CACHE SCENARIOS                              │
	// │                                                              │
	// │  SCENARIO 1: Consumer is caught up (normal operation)        │
	// │  ─────────────────────────────────────────────────────       │
	// │  Producer writes → page cache (warm)                         │
	// │  Consumer reads → page cache HIT → zero-copy → socket       │
	// │  Latency: < 1ms for reads. No disk I/O at all!              │
	// │                                                              │
	// │  SCENARIO 2: Consumer fell behind by 10 minutes              │
	// │  ────────────────────────────────────────────────            │
	// │  Consumer reads → might be in page cache (depends on RAM)   │
	// │  If 64 GB RAM and data rate is 100 MB/s:                    │
	// │  Page cache holds ~640 seconds = ~10 minutes                 │
	// │  So even 10 minutes behind = cache hit!                     │
	// │                                                              │
	// │  SCENARIO 3: Consumer does full replay from the beginning   │
	// │  ──────────────────────────────────────────────              │
	// │  Consumer reads OLD data → page cache MISS → disk read      │
	// │  OS read-ahead kicks in → pre-fetches next blocks           │
	// │  Sequential read throughput: 200+ MB/s (even HDD)           │
	// │                                                              │
	// │  DANGER: Replay consumers can EVICT hot cache entries for    │
	// │  caught-up consumers, causing a cascade of cache misses!    │
	// │  FIX: Isolate replay consumers to different brokers, or     │
	// │  use follower fetching (KIP-392) to read from followers.    │
	// └──────────────────────────────────────────────────────────────┘
	//
	// MEMORY PLANNING:
	// ────────────────
	// Rule: Give Kafka brokers at LEAST as much free RAM as the Kafka data
	// that active consumers need to read.
	//
	// Example:
	//   - Data rate: 500 MB/s across all partitions
	//   - Consumers are typically < 30 seconds behind
	//   - Needed page cache: 500 MB/s × 30s = 15 GB
	//   - Kafka JVM heap: 6 GB
	//   - OS + other processes: 4 GB
	//   - Broker RAM: 15 + 6 + 4 = 25 GB minimum
	//   - Recommended: 32-64 GB (more page cache headroom)
	//
	// CRITICAL: Don't give Kafka JVM too much heap!
	//   - More JVM heap = less page cache = worse performance
	//   - Kafka barely uses heap (most data flows through page cache)
	//   - 6 GB heap is enough for most production clusters
	//
	// vm.swappiness:
	//   - Set to 1 (not 0!) to prevent the OOM killer
	//   - Swapping Kafka data is WORSE than disk reads
	//   - If the OS starts swapping page cache, you have too little RAM

	fmt.Println("  Kafka uses OS page cache as its only cache (no JVM double-buffer)")
	fmt.Println("  Caught-up consumers → 100% cache hit → zero-copy → insane throughput")
	fmt.Println("  Give the broker RAM = JVM heap (6GB) + page cache (match data volume)")
	fmt.Println()
}

// =============================================================================
// PART 4: ZERO-COPY — The sendfile syscall
// =============================================================================
func zeroCopyDeepDive() {
	fmt.Println("--- ZERO-COPY DEEP DIVE ---")

	// NORMAL DATA TRANSFER (without zero-copy):
	// ──────────────────────────────────────────
	//
	// Step 1: read() syscall
	//   Disk → Kernel Buffer (page cache) → User Buffer (JVM heap)
	//   Context switch: user → kernel → user (2 switches)
	//   CPU copy: kernel buffer → user buffer
	//
	// Step 2: write() syscall (to socket)
	//   User Buffer (JVM heap) → Kernel Buffer (socket) → NIC
	//   Context switch: user → kernel → user (2 switches)
	//   CPU copy: user buffer → socket buffer
	//
	// TOTAL: 4 context switches, 2 CPU copies, 2 DMA copies
	// The data touches JVM heap → GC pressure!
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  NORMAL TRANSFER:                                            │
	// │  Disk ─DMA→ Page Cache ─CPU→ User Space ─CPU→ Socket ─DMA→ NIC│
	// │              (kernel)         (JVM heap)       (kernel)       │
	// │  4 copies, 4 context switches                                │
	// └──────────────────────────────────────────────────────────────┘
	//
	// ZERO-COPY TRANSFER (sendfile syscall):
	// ──────────────────────────────────────
	//
	// Java: FileChannel.transferTo() → Linux: sendfile()
	//
	// Step 1: sendfile() syscall
	//   Disk → Kernel Buffer (page cache) → Socket Buffer → NIC
	//   Context switch: user → kernel → user (2 switches)
	//   NO CPU copy! Data goes directly from page cache to socket.
	//   (On modern kernels with scatter-gather DMA, even the socket
	//   buffer copy is avoided — DMA reads directly from page cache)
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  ZERO-COPY:                                                  │
	// │  Disk ─DMA→ Page Cache ────────────DMA────────────────→ NIC │
	// │              (kernel)     (no user space involved!)          │
	// │  2 DMA copies, 2 context switches, ZERO CPU copies          │
	// └──────────────────────────────────────────────────────────────┘
	//
	// WHY THIS IS HUGE:
	// ─────────────────
	// 1. CPU is free for other work (producer handling, replication)
	// 2. No JVM heap allocation → no GC pressure at all
	// 3. Fewer context switches → less kernel overhead
	// 4. Network throughput approaches line rate
	//
	// A single Kafka broker can saturate a 10 Gbps NIC for consumer reads.
	// Without zero-copy, you'd need significantly more CPU.
	//
	// WHEN ZERO-COPY DOESN'T WORK:
	// ─────────────────────────────
	// - SSL/TLS: Data must pass through the JVM for encryption
	//   Kafka 2.x: SSL disables zero-copy entirely
	//   Kafka 3.x: Can use OS-level TLS offload on some platforms
	//
	// - Compressed batches: Zero-copy still works! Because Kafka stores
	//   batches in their compressed form. The broker doesn't decompress.
	//   Compression is end-to-end: producer → broker → consumer.
	//
	// This is also why BROKER-SIDE MESSAGE TRANSFORMATION is bad:
	// If the broker needs to transform messages, it must read into JVM heap,
	// modify, write back — killing zero-copy and adding GC pressure.
	// Keep your messages opaque bytes from the broker's perspective!

	fmt.Println("  Normal: Disk → Page Cache → JVM Heap → Socket → NIC (slow, GC pressure)")
	fmt.Println("  Zero-copy: Disk → Page Cache → NIC (skips JVM entirely!)")
	fmt.Println("  Result: Single broker can saturate 10 Gbps NIC for consumer reads")
	fmt.Println("  SSL/TLS disables zero-copy (data must pass through JVM for encryption)")
	fmt.Println()
}

// =============================================================================
// PART 5: RETENTION — When does data get deleted?
// =============================================================================
func retentionPolicies() {
	fmt.Println("--- RETENTION POLICIES ---")

	// Kafka supports three retention modes:
	//
	// MODE 1: DELETE (default)
	// ────────────────────────
	// Oldest segments are deleted when retention criteria are met.
	//
	// retention.ms (default: 7 days / 604800000 ms)
	//   Delete segments where MAX timestamp in segment < now - retention.ms
	//   NOTE: It's the MAX timestamp, not the first/last record!
	//   If clock skew causes a future timestamp in the segment,
	//   the ENTIRE segment stays past expected retention period.
	//
	// retention.bytes (default: -1 = unlimited)
	//   When total partition size > retention.bytes, delete oldest segments
	//   This is PER-PARTITION, not per-topic!
	//   If topic has 10 partitions and retention.bytes=10GB:
	//   Each partition retains up to 10GB → total = 100GB!
	//
	// DELETION MECHANICS:
	// - The log cleaner thread checks segments periodically
	// - Candidate segments (closed + past retention) are marked for deletion
	// - .deleted suffix is added to the filename
	// - After file.delete.delay.ms (default: 60s), files are actually removed
	// - Active segment is NEVER deleted (even if past retention!)
	//
	// MODE 2: COMPACT
	// ────────────────
	// Keeps only the LATEST value for each key.
	// This turns Kafka into a TABLE (key-value store!).
	//
	// cleanup.policy=compact
	// - Log cleaner thread reads old segments
	// - Builds an offset map: key → latest offset
	// - Rewrites segments keeping only the latest record per key
	// - Records with null value (tombstones) are kept for
	//   delete.retention.ms (default: 24h), then removed
	//
	// USE CASES:
	// - CDC (Change Data Capture): topic = database table, key = primary key
	// - State stores: Kafka Streams uses compacted topics for state
	// - Configuration: store latest config per service, replay on restart
	//
	// ┌──────────────────────────────────────────────────────────────────┐
	// │  LOG COMPACTION EXAMPLE:                                          │
	// │                                                                  │
	// │  BEFORE COMPACTION:                                               │
	// │  Offset: 0   1   2   3   4   5   6   7   8   9   10             │
	// │  Key:    A   B   A   C   A   B   C   A   B   A   C              │
	// │  Value:  v1  v1  v2  v1  v3  v2  v2  v4  v3  v5  v3             │
	// │                                                                  │
	// │  AFTER COMPACTION:                                                │
	// │  Offset: 9   8   10                                              │
	// │  Key:    A   B   C                                               │
	// │  Value:  v5  v3  v3                                              │
	// │                                                                  │
	// │  IMPORTANT: Offsets are PRESERVED. Offset 9 is still offset 9.  │
	// │  Compaction never changes offsets — it only removes old records. │
	// │  Consumers seeking to offset 3 will jump to offset 8 (next      │
	// │  available offset).                                              │
	// └──────────────────────────────────────────────────────────────────┘
	//
	// COMPACTION INTERNALS:
	// ─────────────────────
	// min.cleanable.dirty.ratio (default: 0.5)
	//   The log cleaner compacts a partition when:
	//   (dirty bytes) / (total bytes) > min.cleanable.dirty.ratio
	//   "Dirty" = segments after the last compacted segment
	//   Lower ratio = more frequent compaction but more I/O
	//
	// min.compaction.lag.ms (default: 0)
	//   Records must be at least this old before compaction removes them
	//   Useful to ensure consumers have time to see all versions
	//
	// max.compaction.lag.ms (default: MAX_LONG)
	//   Force compaction after this time, even if dirty ratio not met
	//
	// MODE 3: COMPACT + DELETE
	// ────────────────────────
	// cleanup.policy=compact,delete
	// First compacts (keeps latest per key), THEN applies time/size retention.
	// Useful for: "keep latest state, but don't keep it forever"
	//
	// __consumer_offsets topic uses compact,delete:
	// - Keeps latest committed offset per consumer group+partition
	// - Deletes offsets for consumer groups that no longer exist

	fmt.Println("  DELETE: Remove segments older than retention.ms or larger than retention.bytes")
	fmt.Println("  COMPACT: Keep only latest value per key (Kafka as a table!)")
	fmt.Println("  COMPACT+DELETE: Latest per key, but also time-limited")
	fmt.Println("  Active segment is NEVER deleted (gotcha for low-throughput topics)")
	fmt.Println()
}

// =============================================================================
// PART 6: DISK LAYOUT — Production disk configuration
// =============================================================================
func diskLayoutBestPractices() {
	fmt.Println("--- DISK LAYOUT BEST PRACTICES ---")

	// DISK SELECTION:
	// ───────────────
	// Kafka's I/O is 99% sequential → HDDs work fine for throughput!
	// But SSDs help with:
	//   1. Tail latency (p99) when page cache misses cause real disk reads
	//   2. Log compaction (lots of random reads during rewrite)
	//   3. Higher partition count per broker (more random I/O)
	//
	// LinkedIn runs Kafka on HDDs for most clusters.
	// Uber and Confluent Cloud use SSDs for latency-sensitive workloads.
	//
	// FILESYSTEM:
	// ───────────
	// XFS or ext4. Both work well.
	// XFS: Better for large files, better parallel I/O
	// ext4: Good default, well-tested
	//
	// Mount options: noatime,nodiratime (avoid updating access timestamps)
	//
	// MULTIPLE DISKS:
	// ───────────────
	// log.dirs=/disk1/kafka,/disk2/kafka,/disk3/kafka
	// Kafka distributes partitions across configured log directories.
	// Each partition is fully contained in ONE directory.
	//
	// JBOD (Just a Bunch of Disks) vs RAID:
	//   JBOD: Used by LinkedIn, Uber. Kafka's replication handles disk failure.
	//         If one disk dies, only partitions on that disk are affected.
	//         Other disks continue serving. Followers on other brokers take over.
	//   RAID-10: Redundancy at the hardware level. More operational overhead.
	//            Not needed if replication.factor ≥ 3.
	//
	// FLUSH SETTINGS (fsync):
	// ───────────────────────
	// log.flush.interval.messages (default: MAX_LONG = never by message count)
	// log.flush.interval.ms (default: MAX_LONG = never by time)
	//
	// Kafka does NOT fsync by default!
	// It relies on replication for durability, not disk fsync.
	// The OS flushes dirty pages periodically (vm.dirty_ratio, vm.dirty_background_ratio).
	//
	// WHY NO FSYNC?
	// fsync is EXPENSIVE on HDDs (forces a seek + write) and even on SSDs.
	// With acks=all and replication.factor=3, data is in page cache of 3 brokers.
	// The probability of ALL 3 losing power simultaneously is negligible.
	// LinkedIn ran fsync=never for years without data loss.
	//
	// WHEN TO FSYNC:
	// - Single-broker deployment (no replication safety net)
	// - Regulatory requirements that mandate durability to disk
	// - Topics where replication.factor=1 (avoid this in production!)
	//
	// OS TUNING:
	// ──────────
	// vm.dirty_ratio = 80            (% of RAM before blocking writes)
	// vm.dirty_background_ratio = 5  (% of RAM before background flush)
	// vm.swappiness = 1              (avoid swapping, but don't disable)
	// net.core.wmem_max = 2097152    (socket write buffer)
	// net.core.rmem_max = 2097152    (socket read buffer)
	// fs.file-max = 100000           (max open files system-wide)

	fmt.Println("  HDDs work fine (sequential I/O), SSDs reduce tail latency")
	fmt.Println("  JBOD > RAID-10 when replication.factor ≥ 3")
	fmt.Println("  No fsync by default — replication handles durability")
	fmt.Println("  Mount: noatime, XFS/ext4, give maximum RAM to page cache")
	fmt.Println()
}
