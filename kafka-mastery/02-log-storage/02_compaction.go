//go:build ignore
// =============================================================================
// LESSON 2.2: LOG COMPACTION — Kafka as a Database
// =============================================================================
//
// Log compaction is what transforms Kafka from a "message broker" into a
// "distributed change log" and eventually into a "streaming database."
//
// KEY INSIGHT: With compaction, a Kafka topic becomes a TABLE.
// Each key has exactly one latest value. New consumers can bootstrap
// state by reading the entire compacted topic from the beginning.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== LOG COMPACTION DEEP DIVE ===")
	fmt.Println()

	compactionAlgorithm()
	tombstones()
	compactionTuning()
	compactionUseCases()
}

// =============================================================================
// PART 1: THE COMPACTION ALGORITHM
// =============================================================================
func compactionAlgorithm() {
	fmt.Println("--- COMPACTION ALGORITHM ---")

	// The log is divided into two regions:
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  HEAD (clean)           │  TAIL (dirty)                      │
	// │  ──────────────         │  ────────────                      │
	// │  Already compacted.     │  Not yet compacted.                │
	// │  Each key appears       │  May have duplicate keys.          │
	// │  at most once.          │  Includes the active segment.      │
	// │                         │                                    │
	// │  Segments: 0, 65536     │  Segments: 131072, 196608 (active)│
	// │                         │                                    │
	// │  ◄── clean ──►│◄── dirty ──────────────────────────────►    │
	// │  cleanableRatio = dirty / (clean + dirty)                    │
	// │  Compaction triggered when cleanableRatio > min.cleanable... │
	// └──────────────────────────────────────────────────────────────┘
	//
	// THE ALGORITHM:
	// ──────────────
	// 1. Log cleaner thread selects the "dirtiest" partition
	//    (highest cleanable ratio)
	//
	// 2. Build an OFFSET MAP in memory:
	//    - Reads the dirty (tail) portion
	//    - For each record: map[key] = latestOffset
	//    - This map is kept in a fixed-size buffer (log.cleaner.dedupe.buffer.size)
	//    - Uses a hash map with 16-byte MD5 hash of key + 8-byte offset
	//    - Memory per entry: 24 bytes
	//    - With 128 MB dedupe buffer: ~5.6 million unique keys per pass
	//
	// 3. Re-copy the clean segments:
	//    - Read each record in the clean portion
	//    - If key exists in offset map AND this is NOT the latest offset → SKIP
	//    - Otherwise → COPY to new segment
	//
	// 4. Swap old segments with new compacted segments (atomic rename)
	//
	// 5. Delete old segment files
	//
	// IMPORTANT SAFETY PROPERTIES:
	// ─────────────────────────────
	// - Compaction NEVER reorders records
	// - Offsets are NEVER changed (gaps are fine)
	// - Active segment is NEVER compacted
	// - Records with min.compaction.lag.ms are left alone
	// - Consumers see a consistent view (atomic segment swap)
	//
	// PERFORMANCE IMPACT:
	// ───────────────────
	// - Compaction reads old segments sequentially (OK for throughput)
	// - Compaction writes new segments sequentially (OK for throughput)
	// - CPU: hashing keys for the offset map
	// - Memory: the dedupe buffer (configure appropriately)
	// - I/O: can compete with production reads/writes
	//   → Use log.cleaner.threads (default: 1) and log.cleaner.io.max.bytes.per.second
	//     to throttle compaction I/O

	fmt.Println("  Compaction: build offset map from dirty segments,")
	fmt.Println("  then rewrite clean segments keeping only latest per key")
	fmt.Println("  Active segment is NEVER compacted")
	fmt.Println("  Offsets are preserved (gaps are expected)")
	fmt.Println()
}

// =============================================================================
// PART 2: TOMBSTONES — Deleting keys from compacted topics
// =============================================================================
func tombstones() {
	fmt.Println("--- TOMBSTONES ---")

	// To DELETE a key from a compacted topic, produce a record with:
	//   Key = the key to delete
	//   Value = null (empty/nil)
	//
	// This is called a TOMBSTONE.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  TOMBSTONE LIFECYCLE:                                         │
	// │                                                              │
	// │  t=0    Producer sends: key="user-123", value=null           │
	// │         Record appended to log with offset 500               │
	// │                                                              │
	// │  t=1    First compaction pass:                                │
	// │         Tombstone is KEPT (it's the latest value for key)    │
	// │         Older records for "user-123" are removed             │
	// │                                                              │
	// │  t=T    After delete.retention.ms (default: 24h):            │
	// │         Next compaction pass removes the tombstone itself    │
	// │         Key "user-123" no longer exists in the topic         │
	// │                                                              │
	// │  WHY THE DELAY?                                              │
	// │  Consumers that are behind need time to see the tombstone    │
	// │  so they know to delete the key from their local state.     │
	// │  Without the delay, they'd never know the key was deleted!   │
	// └──────────────────────────────────────────────────────────────┘
	//
	// GOTCHA: delete.retention.ms is separate from retention.ms!
	// retention.ms → applies to DELETE cleanup policy
	// delete.retention.ms → how long tombstones survive in COMPACT policy
	// Confusing names, know the difference.

	fmt.Println("  Tombstone = record with key + null value → deletes the key")
	fmt.Println("  Kept for delete.retention.ms (24h default) so consumers see it")
	fmt.Println("  After that, the key is truly gone from the compacted topic")
	fmt.Println()
}

// =============================================================================
// PART 3: COMPACTION TUNING
// =============================================================================
func compactionTuning() {
	fmt.Println("--- COMPACTION TUNING ---")

	// BROKER-LEVEL SETTINGS:
	// ──────────────────────
	//
	// log.cleaner.threads (default: 1)
	//   Number of compaction threads. Increase for clusters with many
	//   compacted topics. Each thread handles one partition at a time.
	//
	// log.cleaner.dedupe.buffer.size (default: 128 MB per thread)
	//   Memory for the offset map. More memory = more keys per pass.
	//   If your compacted topic has 50M unique keys:
	//   Required: 50M × 24 bytes = 1.2 GB (increase this!)
	//
	// log.cleaner.io.buffer.size (default: 512 KB)
	//   I/O buffer for reading/writing segments during compaction.
	//   Larger = fewer syscalls = better throughput.
	//
	// log.cleaner.io.max.bytes.per.second (default: ~1.8 EB = unlimited)
	//   Throttle compaction I/O to avoid impacting production.
	//   Set to ~50 MB/s on busy clusters.
	//
	// log.cleaner.backoff.ms (default: 15s)
	//   Sleep between compaction checks when nothing needs compaction.
	//
	// TOPIC-LEVEL SETTINGS:
	// ─────────────────────
	//
	// min.cleanable.dirty.ratio (default: 0.5)
	//   Compact when dirty% > this. Lower = fresher compaction.
	//   0.1 = compact aggressively (more I/O, fresher view)
	//   0.9 = compact rarely (less I/O, stale data longer)
	//
	// min.compaction.lag.ms (default: 0)
	//   Don't compact records newer than this.
	//   Use when consumers need to see intermediate values.
	//   Example: CDC pipeline needs to see all changes for audit.
	//
	// max.compaction.lag.ms (default: Long.MAX)
	//   FORCE compaction after this, overriding dirty ratio.
	//   Critical for GDPR: "delete user data within 72 hours"
	//   Set to 72 hours to ensure tombstones are compacted within SLA.
	//
	// MONITORING COMPACTION:
	// ──────────────────────
	// kafka.log:type=LogCleaner,name=cleaner-recopy-percent
	//   → percentage of data recopied (lower = better compaction)
	// kafka.log:type=LogCleaner,name=max-clean-time-secs
	//   → longest compaction run (watch for increasing trend)
	// kafka.log:type=LogCleaner,name=max-buffer-utilization-percent
	//   → if 100%, you need more dedupe buffer

	fmt.Println("  Key tuning: dedupe buffer size, dirty ratio, compaction lag")
	fmt.Println("  For GDPR: use max.compaction.lag.ms to force timely deletion")
	fmt.Println("  Monitor: buffer utilization, recopy percent, clean time")
	fmt.Println()
}

// =============================================================================
// PART 4: COMPACTION USE CASES — Real-world patterns
// =============================================================================
func compactionUseCases() {
	fmt.Println("--- COMPACTION USE CASES ---")

	// USE CASE 1: CDC (Change Data Capture)
	// ──────────────────────────────────────
	// Database changes → Debezium → Kafka topic (compacted)
	// Key = table.primaryKey (e.g., "users.123")
	// Value = full row as JSON/Avro
	//
	// Result: Kafka topic IS the database table!
	// New microservice? Read compacted topic from beginning = full table snapshot.
	// Then continue consuming for real-time updates.
	// No need for database dumps, no ETL batch jobs.
	//
	// USE CASE 2: State Store Backing (Kafka Streams)
	// ────────────────────────────────────────────────
	// Kafka Streams uses compacted topics as "changelog topics"
	// for its local state stores (RocksDB).
	// If a stream processor crashes and restarts, it rebuilds state
	// by replaying the changelog topic.
	// Compaction ensures this replay is bounded in time.
	//
	// USE CASE 3: Configuration Distribution
	// ───────────────────────────────────────
	// Key = service name or config key
	// Value = configuration JSON
	// All services consume this topic from the beginning on startup.
	// They get the LATEST configuration for everything.
	// Changes propagate in real-time.
	//
	// USE CASE 4: Session Store
	// ─────────────────────────
	// Key = session ID
	// Value = session data
	// Tombstone when session expires.
	// Stateless web tier can bootstrap session state from Kafka.
	//
	// ANTI-PATTERN: Using compacted topics for time-series data
	// ─────────────────────────────────────────────────────────
	// Time-series has unique timestamps as keys → compaction is useless
	// (every key is unique, nothing to compact). Use DELETE retention instead.

	fmt.Println("  CDC: Kafka topic = database table (key=PK, value=row)")
	fmt.Println("  State stores: changelog topics for crash recovery")
	fmt.Println("  Configuration: latest config per key, real-time updates")
	fmt.Println("  Anti-pattern: time-series data (unique keys = no compaction benefit)")
}
