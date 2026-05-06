//go:build ignore
// =============================================================================
// LESSON 3.1: PRODUCER INTERNALS — Every Byte from Send() to Broker ACK
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - The complete internal pipeline: serializer → partitioner → accumulator → sender
// - Batching internals: how linger.ms and batch.size interact
// - Compression: when, where, and which codec (zstd wins almost always)
// - Idempotent producer: sequence numbers, producer IDs, exactly-once semantics
// - Sticky partitioner: why it replaced round-robin as default
// - Retry semantics: what happens on failure and how ordering is preserved
// - Memory management: buffer.memory, max.block.ms, and backpressure
//
// THE KEY INSIGHT:
// The producer is NOT a simple "send and forget" client. It's a sophisticated
// async pipeline with batching, compression, connection pooling, and retry logic.
// Understanding this pipeline is the difference between 10K msg/sec and 1M msg/sec.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== PRODUCER INTERNALS ===")
	fmt.Println()

	producerPipeline()
	batchingDeepDive()
	compressionStrategy()
	idempotentProducer()
	stickyPartitioner()
	retrySemantics()
	producerMemoryManagement()
	producerTuningMatrix()
}

// =============================================================================
// PART 1: THE PRODUCER PIPELINE
// =============================================================================
func producerPipeline() {
	fmt.Println("--- PRODUCER PIPELINE ---")

	// When you call producer.Send(record), here's the FULL internal flow:
	//
	// ┌──────────────────────────────────────────────────────────────────┐
	// │                   PRODUCER INTERNAL PIPELINE                      │
	// │                                                                  │
	// │  Application Thread                    Sender Thread (background)│
	// │  ═══════════════════                   ═════════════════════════ │
	// │                                                                  │
	// │  record.Send()                                                   │
	// │       │                                                          │
	// │       ▼                                                          │
	// │  ┌─────────────┐                                                 │
	// │  │ Interceptors│ (optional, rarely used in prod)                 │
	// │  └──────┬──────┘                                                 │
	// │         ▼                                                        │
	// │  ┌─────────────┐                                                 │
	// │  │ Serializer  │ Key serializer + Value serializer               │
	// │  │ (key+value) │ Converts objects → byte[]                       │
	// │  └──────┬──────┘                                                 │
	// │         ▼                                                        │
	// │  ┌─────────────┐                                                 │
	// │  │ Partitioner │ Determines target partition                     │
	// │  │             │ - Key != null: murmur2(key) % numPartitions     │
	// │  │             │ - Key == null: sticky partitioner               │
	// │  └──────┬──────┘                                                 │
	// │         ▼                                                        │
	// │  ┌─────────────────────────────────────┐                         │
	// │  │ RecordAccumulator                    │    ┌─────────────────┐ │
	// │  │ (per-partition batching buffer)      │───►│  Sender Thread  │ │
	// │  │                                     │    │                 │ │
	// │  │  Partition 0: [batch][batch]         │    │  Drains batches │ │
	// │  │  Partition 1: [batch]                │    │  Groups by broker│
	// │  │  Partition 2: [batch][batch][batch]  │    │  Creates request│ │
	// │  │                                     │    │  Sends to broker│ │
	// │  │  Total memory: buffer.memory (32MB) │    │  Handles ACK    │ │
	// │  └─────────────────────────────────────┘    │  Retries on fail│ │
	// │                                             └─────────────────┘ │
	// └──────────────────────────────────────────────────────────────────┘
	//
	// TWO THREADS:
	// ────────────
	// 1. Application thread: serialize → partition → append to batch
	//    This is YOUR thread. It's fast — just memory operations.
	//    UNLESS buffer.memory is full → blocks for max.block.ms
	//
	// 2. Sender thread: drain batches → compress → send to broker → handle response
	//    Background thread. One per producer instance.
	//    Makes actual network calls. Handles retries.
	//
	// The RecordAccumulator is the bridge between the two threads.
	// It's a ConcurrentMap<TopicPartition, Deque<RecordBatch>>.

	fmt.Println("  Pipeline: Serialize → Partition → Accumulate → Send (async)")
	fmt.Println("  App thread: fast memory ops. Sender thread: network + compression.")
	fmt.Println("  RecordAccumulator bridges the two with per-partition batch queues.")
	fmt.Println()
}

// =============================================================================
// PART 2: BATCHING — The secret to Kafka's throughput
// =============================================================================
func batchingDeepDive() {
	fmt.Println("--- BATCHING DEEP DIVE ---")

	// Batching is THE single most important throughput optimization in Kafka.
	// A batch amortizes: network overhead, disk I/O, compression, CRC computation.
	//
	// TWO KNOBS CONTROL BATCHING:
	//
	// batch.size (default: 16384 = 16 KB)
	//   Maximum size of a single batch in bytes.
	//   When a batch reaches this size → Sender drains it immediately.
	//   Larger batch = better compression, fewer requests, higher throughput.
	//   Typical production: 64 KB - 1 MB
	//
	// linger.ms (default: 0)
	//   How long to wait for more records before sending a batch.
	//   0 = send immediately when Sender thread is available (still batches
	//       if records accumulate faster than sending)
	//   5 = wait up to 5ms for more records → better batching, +5ms latency
	//   100 = aggressive batching → great throughput, +100ms latency
	//
	// HOW THEY INTERACT:
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  Batch is sent when EITHER condition is met:                  │
	// │                                                              │
	// │  1. batch.size reached (batch is "full")                     │
	// │  2. linger.ms expired since first record was added           │
	// │  3. Another batch for same broker is ready (piggybacking)    │
	// │  4. flush() or close() is called                             │
	// │  5. Memory pressure (buffer.memory running low)              │
	// │                                                              │
	// │  SCENARIOS:                                                   │
	// │                                                              │
	// │  High throughput (1M+ records/sec) + linger.ms=0:            │
	// │  → Batches fill up to batch.size before Sender drains them   │
	// │  → linger.ms is irrelevant (batches are always full)         │
	// │  → Increase batch.size to get bigger batches                 │
	// │                                                              │
	// │  Low throughput (100 records/sec) + linger.ms=0:             │
	// │  → Batches are tiny (1-2 records each)                       │
	// │  → Many small network requests → poor throughput              │
	// │  → Set linger.ms=50-100 to allow batches to accumulate       │
	// │                                                              │
	// │  Low throughput + linger.ms=100:                              │
	// │  → Batches accumulate for 100ms → reasonable batch sizes     │
	// │  → Tradeoff: +100ms latency for better throughput             │
	// └──────────────────────────────────────────────────────────────┘
	//
	// BATCH MEMORY LAYOUT:
	// ────────────────────
	// Each batch is allocated from buffer.memory as a ByteBuffer.
	// Records are appended into the batch in the RecordBatch wire format
	// (same format that's stored on the broker — no conversion needed).
	//
	// When compression is enabled, the batch is compressed just before sending.
	// So the in-memory batch is UNCOMPRESSED, uses full batch.size memory.
	// After compression, the actual network payload is smaller.
	//
	// MAX.REQUEST.SIZE (default: 1 MB):
	// Maximum total size of a ProduceRequest (can contain multiple batches).
	// If batch.size > max.request.size, you'll get errors.
	// Rule: max.request.size > batch.size (always!)
	// Also: broker's message.max.bytes must be ≥ max.request.size
	//
	// RECORDS-PER-BATCH SWEET SPOT:
	// ─────────────────────────────
	// Too few records per batch: high overhead per record
	// Too many records per batch: high latency, memory pressure
	//
	// For throughput optimization:
	//   batch.size=512KB-1MB, linger.ms=50-200, compression=zstd
	//   → Expect 500K-2M records/sec per producer instance
	//
	// For latency optimization:
	//   batch.size=16KB-64KB, linger.ms=0-5, compression=lz4 or snappy
	//   → Expect 1-5ms p99 produce latency

	fmt.Println("  batch.size: max batch bytes. linger.ms: max wait time.")
	fmt.Println("  High throughput → big batch.size + linger.ms=50-200")
	fmt.Println("  Low latency → small batch.size + linger.ms=0-5")
	fmt.Println("  Both → many partitions + linger.ms=0 + batch naturally fills up")
	fmt.Println()
}

// =============================================================================
// PART 3: COMPRESSION — When, where, and which codec
// =============================================================================
func compressionStrategy() {
	fmt.Println("--- COMPRESSION STRATEGY ---")

	// Kafka supports 4 compression codecs:
	//
	// ┌───────────────┬────────────┬────────────┬──────────────────────┐
	// │ Codec         │ Ratio      │ Speed      │ CPU Usage            │
	// ├───────────────┼────────────┼────────────┼──────────────────────┤
	// │ none          │ 1.0x       │ ∞          │ 0                    │
	// │ gzip          │ Best (5-8x)│ Slowest    │ Highest              │
	// │ snappy        │ Good (3-4x)│ Fast       │ Low                  │
	// │ lz4           │ Good (3-4x)│ Fastest    │ Very Low             │
	// │ zstd          │ Best (5-8x)│ Fast       │ Medium (tunable)     │
	// └───────────────┴────────────┴────────────┴──────────────────────┘
	//
	// RECOMMENDATION:
	// ───────────────
	// zstd is the WINNER for almost all use cases.
	// - Compression ratio close to gzip
	// - Speed close to lz4
	// - Has adjustable compression levels (1-22)
	//   Level 3 (default): good balance
	//   Level 1: fastest, ratio similar to lz4
	//   Level 19+: slow, ratio approaching gzip's best
	//
	// USE lz4 if: CPU is your bottleneck and you can't afford any extra cycles.
	// USE snappy if: You're on older Kafka versions without zstd support.
	// NEVER use gzip in production: CPU cost is too high for the marginal gain.
	//
	// WHERE COMPRESSION HAPPENS:
	// ──────────────────────────
	// END-TO-END compression model:
	//   Producer → compresses batch → sends compressed bytes → Broker
	//   Broker → stores compressed bytes as-is → Disk
	//   Broker → sends compressed bytes as-is → Consumer
	//   Consumer → decompresses batch → application
	//
	// The broker NEVER decompresses data (with one exception).
	// This is WHY zero-copy works: bytes flow from disk to network unchanged.
	//
	// THE EXCEPTION: compression.type on the topic
	// If the topic has compression.type=zstd but the producer sends lz4,
	// the broker MUST decompress (lz4) and recompress (zstd).
	// This kills zero-copy and adds CPU load on the broker.
	// BEST PRACTICE: Set topic compression.type=producer (default)
	// and control compression on the producer side.
	//
	// COMPRESSION & BATCHING SYNERGY:
	// ───────────────────────────────
	// Compression works on the ENTIRE BATCH, not individual records.
	// A batch of 100 similar JSON records compresses MUCH better than
	// 100 individually compressed records.
	// This is another reason bigger batches = better throughput.
	//
	// REAL NUMBERS (1 KB JSON records, batch of 100):
	// ─────────────────────────────────────────────
	// No compression: 100 KB network, 100 KB disk
	// lz4:            22 KB network, 22 KB disk (4.5x ratio)
	// zstd:           15 KB network, 15 KB disk (6.7x ratio)
	// gzip:           14 KB network, 14 KB disk (7.1x ratio)
	// CPU per batch: lz4 < zstd << gzip

	fmt.Println("  zstd: best ratio + good speed (use this by default)")
	fmt.Println("  lz4: fastest (use when CPU-bound)")
	fmt.Println("  Compression is end-to-end: producer → broker → consumer")
	fmt.Println("  Set topic compression.type=producer (avoid broker recompression)")
	fmt.Println()
}

// =============================================================================
// PART 4: IDEMPOTENT PRODUCER — Exactly-once at the produce level
// =============================================================================
func idempotentProducer() {
	fmt.Println("--- IDEMPOTENT PRODUCER ---")

	// PROBLEM: Without idempotency, retries cause DUPLICATES.
	//
	// Scenario:
	// 1. Producer sends batch to broker
	// 2. Broker writes batch to log → ACK response
	// 3. ACK is LOST (network blip)
	// 4. Producer RETRIES the same batch
	// 5. Broker writes it AGAIN → DUPLICATE DATA!
	//
	// SOLUTION: Idempotent producer (enable.idempotence=true, default since Kafka 3.0)
	//
	// HOW IT WORKS:
	// ─────────────
	// 1. On first send, broker assigns the producer a PRODUCER ID (PID)
	//    - PID is a unique int64
	//    - Survives transient failures (stored in memory)
	//    - DOES NOT survive producer restart (new PID on restart)
	//      For cross-restart dedup, you need TRANSACTIONS (Lesson 07)
	//
	// 2. Each batch gets a SEQUENCE NUMBER (per partition)
	//    - Starts at 0 for each (PID, partition) pair
	//    - Incremented by the number of records in the batch
	//    - Broker tracks: expectedSequence[PID][partition]
	//
	// 3. On each produce:
	//    - Broker checks: is batch.sequence == expected?
	//    - YES → write to log, increment expected, ACK
	//    - NO (higher) → OUT_OF_ORDER_SEQUENCE error → producer retries
	//    - NO (same or lower) → DUPLICATE → ACK without writing (dedup!)
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  IDEMPOTENT PRODUCER DEDUP FLOW:                              │
	// │                                                              │
	// │  Producer                          Broker (Partition 0)      │
	// │  PID=5                             expected_seq[5][0] = 0    │
	// │                                                              │
	// │  Send(seq=0, records=[A,B]) ────►  0 == 0? YES               │
	// │                                    Write [A,B], expected = 2 │
	// │  ◄───── ACK (lost!) ───────────                              │
	// │                                                              │
	// │  RETRY(seq=0, records=[A,B]) ──►  0 < 2? DUPLICATE           │
	// │                                    ACK without writing!      │
	// │  ◄───── ACK ──────────────────                               │
	// │                                                              │
	// │  Send(seq=2, records=[C]) ──────► 2 == 2? YES                │
	// │                                   Write [C], expected = 3    │
	// │  ◄───── ACK ──────────────────                               │
	// └──────────────────────────────────────────────────────────────┘
	//
	// WHAT IDEMPOTENCE REQUIRES:
	// ──────────────────────────
	// max.in.flight.requests.per.connection ≤ 5 (default: 5)
	//   With idempotence, Kafka guarantees ordering even with 5 in-flight!
	//   How? The broker rejects out-of-order batches and the producer retries.
	//   Without idempotence, max.in.flight=1 is needed for ordering.
	//
	// retries > 0 (default: Integer.MAX_VALUE)
	//   Idempotence needs retries to work.
	//
	// acks=all (forced by idempotence)
	//   Idempotence implies acks=all. You can't have idempotent + acks=1.
	//
	// THE PRODUCER ID (PID) EPOCH:
	// ────────────────────────────
	// When a producer gets a new PID (restart or transaction boundary),
	// the broker's old sequence tracking for the old PID is irrelevant.
	//
	// Gotcha: PID + sequence is tracked per-partition ON THE LEADER.
	// If the leader changes, the new leader has this info from the replicated
	// log (sequence numbers are stored in the record batch headers).
	//
	// BROKER STATE FOR IDEMPOTENCE:
	// ─────────────────────────────
	// The broker keeps in memory:
	//   Map<PID, Map<Partition, ProducerStateEntry>>
	//   ProducerStateEntry = { lastSequence, lastTimestamp, epoch }
	//
	// This state is also written to a snapshot file on disk:
	//   /var/kafka-data/topic-partition/00000000000000123456.snapshot
	// To survive broker restart.
	//
	// Memory impact: ~200 bytes per active (PID, partition) pair.
	// 100K active producers × 100 partitions = 10M entries × 200B = ~2 GB
	// This is why very large numbers of transient producers can be problematic.

	fmt.Println("  Problem: retries cause duplicates")
	fmt.Println("  Solution: PID + sequence number per (producer, partition)")
	fmt.Println("  Broker deduplicates: same sequence → ACK without writing")
	fmt.Println("  Requires: acks=all, max.in.flight ≤ 5, retries=MAX")
	fmt.Println()
}

// =============================================================================
// PART 5: STICKY PARTITIONER — Why round-robin was replaced
// =============================================================================
func stickyPartitioner() {
	fmt.Println("--- STICKY PARTITIONER ---")

	// WHEN KEY IS NULL (no explicit key):
	//
	// OLD BEHAVIOR (before Kafka 2.4): Round-Robin
	// ─────────────────────────────────────────────
	// Each record goes to the next partition.
	// Record 0 → Partition 0
	// Record 1 → Partition 1
	// Record 2 → Partition 2
	// Record 3 → Partition 0 (wrap around)
	//
	// PROBLEM: Each batch has only 1 record!
	// If you have 10 partitions and send 10 records:
	//   - 10 batches of 1 record each
	//   - 10 network requests (one per partition per broker)
	//   - Terrible compression ratio (one record per batch)
	//   - Terrible throughput
	//
	// NEW BEHAVIOR (Kafka 2.4+): Sticky Partitioner
	// ───────────────────────────────────────────────
	// Sticks to ONE partition until the batch is full (or linger.ms expires).
	// Then switches to another partition.
	//
	// Record 0-99 → Partition 3  (one full batch)
	// Record 100-199 → Partition 7 (one full batch)
	// Record 200-299 → Partition 1 (one full batch)
	//
	// RESULT: Full batches, great compression, fewer network calls.
	// Throughput improvement: 50-100% for null-key workloads!
	//
	// WHEN KEY IS NOT NULL:
	// ─────────────────────
	// DefaultPartitioner: murmur2(keyBytes) % numPartitions
	// This guarantees: same key → same partition → ordering per key.
	//
	// GOTCHA: If you change the number of partitions, the same key
	// maps to a DIFFERENT partition! Existing ordering guarantees break.
	// This is why you should NEVER increase partitions on a keyed topic
	// without careful planning (see Lesson 05).

	fmt.Println("  Sticky partitioner: fills one batch per partition, then switches")
	fmt.Println("  50-100% throughput improvement over round-robin for null-key records")
	fmt.Println("  Keys: murmur2(key) % numPartitions (deterministic)")
	fmt.Println()
}

// =============================================================================
// PART 6: RETRY SEMANTICS — What happens on failure
// =============================================================================
func retrySemantics() {
	fmt.Println("--- RETRY SEMANTICS ---")

	// When a produce request fails, the producer retries automatically.
	//
	// RETRYABLE ERRORS (producer retries automatically):
	// ──────────────────────────────────────────────────
	// - LEADER_NOT_AVAILABLE: broker just became leader, not ready yet
	// - NOT_LEADER_FOR_PARTITION: metadata stale, refresh and retry on new leader
	// - REQUEST_TIMED_OUT: broker didn't respond in time
	// - NETWORK_EXCEPTION: connection lost
	// - KAFKA_STORAGE_EXCEPTION: broker disk issue (temporary)
	//
	// NON-RETRYABLE ERRORS (fail immediately):
	// ────────────────────────────────────────
	// - RECORD_TOO_LARGE: record exceeds max.request.size or message.max.bytes
	// - INVALID_REQUIRED_ACKS: invalid acks config
	// - TOPIC_AUTHORIZATION_FAILED: ACL denies write
	// - UNSUPPORTED_COMPRESSION_TYPE: broker doesn't support the codec
	//
	// RETRY CONFIGURATION:
	// ────────────────────
	// retries (default: 2147483647 = basically infinite)
	//   Number of times to retry a failed request.
	//   With delivery.timeout.ms, this is effectively infinite.
	//
	// delivery.timeout.ms (default: 120000 = 2 minutes)
	//   TOTAL time from producer.send() to ACK (including retries!).
	//   If the sum of all retry attempts exceeds this → fail permanently.
	//   This is the REAL timeout. retries count is secondary.
	//
	// retry.backoff.ms (default: 100ms)
	//   Delay between retry attempts.
	//
	// ORDERING AND RETRIES:
	// ─────────────────────
	// WITHOUT IDEMPOTENCE:
	//   max.in.flight.requests.per.connection > 1 + retries → REORDERING
	//   Batch 1 fails, Batch 2 succeeds, Batch 1 retry succeeds → out of order!
	//   FIX: set max.in.flight=1 (but reduces throughput)
	//
	// WITH IDEMPOTENCE (enable.idempotence=true):
	//   max.in.flight up to 5 is SAFE for ordering!
	//   The broker uses sequence numbers to detect and reject out-of-order batches.
	//   The producer then retries them in the correct order.
	//   This is why idempotent producer is strictly superior.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  RETRY TIMELINE:                                              │
	// │                                                              │
	// │  delivery.timeout.ms = 2000ms total                          │
	// │  retry.backoff.ms = 100ms                                    │
	// │                                                              │
	// │  t=0     Send attempt 1 ────► TIMEOUT (500ms)                │
	// │  t=600   Send attempt 2 ────► TIMEOUT (500ms)                │
	// │  t=1200  Send attempt 3 ────► SUCCESS (200ms)                │
	// │  t=1400  Callback with success                               │
	// │                                                              │
	// │  If attempt 3 also failed:                                    │
	// │  t=1800  Send attempt 4 ────► t=2000: delivery.timeout hit!  │
	// │  t=2000  Callback with TimeoutException                      │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  delivery.timeout.ms (120s) is the real deadline, not retries count")
	fmt.Println("  With idempotence: max.in.flight=5 is safe for ordering")
	fmt.Println("  Without idempotence: max.in.flight=1 needed for ordering")
	fmt.Println()
}

// =============================================================================
// PART 7: MEMORY MANAGEMENT — buffer.memory and backpressure
// =============================================================================
func producerMemoryManagement() {
	fmt.Println("--- PRODUCER MEMORY MANAGEMENT ---")

	// buffer.memory (default: 33554432 = 32 MB)
	// ──────────────────────────────────────────
	// Total memory the producer can use for batching records.
	// Shared across ALL partitions.
	//
	// When buffer.memory is full:
	// 1. producer.send() BLOCKS (on the application thread!)
	// 2. Blocks for up to max.block.ms (default: 60000 = 60 seconds)
	// 3. If still full after max.block.ms → TimeoutException
	//
	// This is BACKPRESSURE: the producer slows down the application
	// when brokers can't keep up or network is slow.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  BUFFER LIFECYCLE:                                            │
	// │                                                              │
	// │  buffer.memory = 32 MB                                       │
	// │  ┌────────────────────────────────────────────────────────┐  │
	// │  │ Free: 20 MB │ Batch P0: 4MB │ Batch P1: 8MB │         │  │
	// │  └────────────────────────────────────────────────────────┘  │
	// │                                                              │
	// │  Producer.Send() needs 16 MB for a new batch:               │
	// │  Free (20 MB) ≥ 16 MB → Allocate, continue.                  │
	// │                                                              │
	// │  Producer.Send() needs 24 MB but only 20 MB free:           │
	// │  BLOCK! Wait for Sender thread to drain batches and free up.│
	// │  After max.block.ms → throw TimeoutException.               │
	// └──────────────────────────────────────────────────────────────┘
	//
	// SIZING buffer.memory:
	// ─────────────────────
	// Need: numPartitions × batch.size (worst case: all partitions have full batches)
	// Example: 50 partitions × 512 KB batch = 25 MB minimum
	// Add headroom: 32-64 MB is good for most cases.
	// With 500+ partitions: may need 128 MB or more.
	//
	// IMPORTANT: buffer.memory is OFF-HEAP in Java producer (direct ByteBuffers).
	// It does NOT affect JVM heap or GC. But it does affect total process RSS.
	// In Go clients (franz-go, confluent-kafka-go), check the library's
	// memory management model.

	fmt.Println("  buffer.memory=32MB shared across all partition batches")
	fmt.Println("  Full buffer → send() BLOCKS for max.block.ms → TimeoutException")
	fmt.Println("  Size: numPartitions × batch.size + headroom")
	fmt.Println()
}

// =============================================================================
// PART 8: PRODUCER TUNING MATRIX — Quick reference
// =============================================================================
func producerTuningMatrix() {
	fmt.Println("--- PRODUCER TUNING MATRIX ---")

	// ┌────────────────────────────────────────────────────────────────────────┐
	// │  OPTIMIZE FOR     │ KEY SETTINGS                                       │
	// ├────────────────────────────────────────────────────────────────────────┤
	// │  MAX THROUGHPUT   │ batch.size=512KB-1MB                               │
	// │                   │ linger.ms=50-200                                    │
	// │                   │ compression.type=zstd                              │
	// │                   │ buffer.memory=128MB                                │
	// │                   │ acks=1 (if you can tolerate some loss)             │
	// │                   │ max.in.flight=5                                     │
	// ├────────────────────────────────────────────────────────────────────────┤
	// │  MIN LATENCY      │ batch.size=16KB                                    │
	// │                   │ linger.ms=0                                         │
	// │                   │ compression.type=lz4 (or none)                     │
	// │                   │ acks=1                                              │
	// │                   │ max.in.flight=1-5                                   │
	// ├────────────────────────────────────────────────────────────────────────┤
	// │  DURABILITY       │ acks=all                                            │
	// │                   │ enable.idempotence=true                             │
	// │                   │ min.insync.replicas=2 (broker/topic config)        │
	// │                   │ retries=MAX (default)                               │
	// │                   │ delivery.timeout.ms=120000 (default)               │
	// ├────────────────────────────────────────────────────────────────────────┤
	// │  ORDERING         │ enable.idempotence=true (allows max.in.flight=5)  │
	// │                   │ OR max.in.flight=1 (without idempotence)           │
	// ├────────────────────────────────────────────────────────────────────────┤
	// │  GOD TIER         │ enable.idempotence=true                             │
	// │  (throughput +    │ acks=all                                            │
	// │   durability +    │ compression.type=zstd                              │
	// │   ordering)       │ batch.size=256KB                                    │
	// │                   │ linger.ms=20                                         │
	// │                   │ buffer.memory=64MB                                  │
	// │                   │ max.in.flight=5                                     │
	// │                   │ min.insync.replicas=2                               │
	// └────────────────────────────────────────────────────────────────────────┘

	fmt.Println("  God tier config: idempotent + acks=all + zstd + batch.size=256KB")
	fmt.Println("  This gives you: ordering + durability + good throughput")
	fmt.Println("  Only sacrifice: ~20ms extra latency from linger.ms")
	fmt.Println()
}






































































































































































































































































































































































































































































































































































































}	fmt.Println()	fmt.Println("  Only sacrifice: ~20ms extra latency from linger.ms")	fmt.Println("  This gives you: ordering + durability + good throughput")	fmt.Println("  God tier config: idempotent + acks=all + zstd + batch.size=256KB")	// └────────────────────────────────────────────────────────────────────────┘	// │                   │ min.insync.replicas=2                               │	// │                   │ max.in.flight=5                                     │	// │                   │ buffer.memory=64MB                                  │	// │                   │ linger.ms=20                                         │	// │   ordering)       │ batch.size=256KB                                    │	// │   durability +    │ compression.type=zstd                              │	// │  (throughput +    │ acks=all                                            │	// │  GOD TIER         │ enable.idempotence=true                             │	// ├────────────────────────────────────────────────────────────────────────┤	// │                   │ OR max.in.flight=1 (without idempotence)           │	// │  ORDERING         │ enable.idempotence=true (allows max.in.flight=5)  │	// ├────────────────────────────────────────────────────────────────────────┤	// │                   │ delivery.timeout.ms=120000 (default)               │	// │                   │ retries=MAX (default)                               │	// │                   │ min.insync.replicas=2 (broker/topic config)        │	// │                   │ enable.idempotence=true                             │	// │  DURABILITY       │ acks=all                                            │	// ├────────────────────────────────────────────────────────────────────────┤	// │                   │ max.in.flight=1-5                                   │	// │                   │ acks=1                                              │	// │                   │ compression.type=lz4 (or none)                     │	// │                   │ linger.ms=0                                         │	// │  MIN LATENCY      │ batch.size=16KB                                    │	// ├────────────────────────────────────────────────────────────────────────┤	// │                   │ max.in.flight=5                                     │	// │                   │ acks=1 (if you can tolerate some loss)             │	// │                   │ buffer.memory=128MB                                │	// │                   │ compression.type=zstd                              │	// │                   │ linger.ms=50-200                                    │	// │  MAX THROUGHPUT   │ batch.size=512KB-1MB                               │	// ├────────────────────────────────────────────────────────────────────────┤	// │  OPTIMIZE FOR     │ KEY SETTINGS                                       │	// ┌────────────────────────────────────────────────────────────────────────┐	fmt.Println("--- PRODUCER TUNING MATRIX ---")func producerTuningMatrix() {// =============================================================================// PART 8: PRODUCER TUNING MATRIX — Quick reference// =============================================================================}	fmt.Println()	fmt.Println("  Size: numPartitions × batch.size + headroom")	fmt.Println("  Full buffer → send() BLOCKS for max.block.ms → TimeoutException")	fmt.Println("  buffer.memory=32MB shared across all partition batches")	// memory management model.	// In Go clients (franz-go, confluent-kafka-go), check the library's	// It does NOT affect JVM heap or GC. But it does affect total process RSS.	// IMPORTANT: buffer.memory is OFF-HEAP in Java producer (direct ByteBuffers).	//	// With 500+ partitions: may need 128 MB or more.	// Add headroom: 32-64 MB is good for most cases.	// Example: 50 partitions × 512 KB batch = 25 MB minimum	// Need: numPartitions × batch.size (worst case: all partitions have full batches)	// ─────────────────────	// SIZING buffer.memory:	//	// └──────────────────────────────────────────────────────────────┘	// │  After max.block.ms → throw TimeoutException.               │	// │  BLOCK! Wait for Sender thread to drain batches and free up.│	// │  Producer.Send() needs 24 MB but only 20 MB free:           │	// │                                                              │	// │  Free (20 MB) ≥ 16 MB → Allocate, continue.                  │	// │  Producer.Send() needs 16 MB for a new batch:               │	// │                                                              │	// │  └────────────────────────────────────────────────────────┘  │	// │  │ Free: 20 MB │ Batch P0: 4MB │ Batch P1: 8MB │         │  │	// │  ┌────────────────────────────────────────────────────────┐  │	// │  buffer.memory = 32 MB                                       │	// │                                                              │	// │  BUFFER LIFECYCLE:                                            │	// ┌──────────────────────────────────────────────────────────────┐	//	// when brokers can't keep up or network is slow.	// This is BACKPRESSURE: the producer slows down the application	//	// 3. If still full after max.block.ms → TimeoutException	// 2. Blocks for up to max.block.ms (default: 60000 = 60 seconds)	// 1. producer.send() BLOCKS (on the application thread!)	// When buffer.memory is full:	//	// Shared across ALL partitions.	// Total memory the producer can use for batching records.	// ──────────────────────────────────────────	// buffer.memory (default: 33554432 = 32 MB)	fmt.Println("--- PRODUCER MEMORY MANAGEMENT ---")func producerMemoryManagement() {// =============================================================================// PART 7: MEMORY MANAGEMENT — buffer.memory and backpressure// =============================================================================}	fmt.Println()	fmt.Println("  Without idempotence: max.in.flight=1 needed for ordering")	fmt.Println("  With idempotence: max.in.flight=5 is safe for ordering")	fmt.Println("  delivery.timeout.ms (120s) is the real deadline, not retries count")	// └──────────────────────────────────────────────────────────────┘	// │  t=2000  Callback with TimeoutException                      │	// │  t=1800  Send attempt 4 ────► t=2000: delivery.timeout hit!  │	// │  If attempt 3 also failed:                                    │	// │                                                              │	// │  t=1400  Callback with success                               │	// │  t=1200  Send attempt 3 ────► SUCCESS (200ms)                │	// │  t=600   Send attempt 2 ────► TIMEOUT (500ms)                │	// │  t=0     Send attempt 1 ────► TIMEOUT (500ms)                │	// │                                                              │	// │  retry.backoff.ms = 100ms                                    │	// │  delivery.timeout.ms = 2000ms total                          │	// │                                                              │	// │  RETRY TIMELINE:                                              │	// ┌──────────────────────────────────────────────────────────────┐	//	//   This is why idempotent producer is strictly superior.	//   The producer then retries them in the correct order.	//   The broker uses sequence numbers to detect and reject out-of-order batches.	//   max.in.flight up to 5 is SAFE for ordering!	// WITH IDEMPOTENCE (enable.idempotence=true):	//	//   FIX: set max.in.flight=1 (but reduces throughput)	//   Batch 1 fails, Batch 2 succeeds, Batch 1 retry succeeds → out of order!	//   max.in.flight.requests.per.connection > 1 + retries → REORDERING	// WITHOUT IDEMPOTENCE:	// ─────────────────────	// ORDERING AND RETRIES:	//	//   Delay between retry attempts.	// retry.backoff.ms (default: 100ms)	//	//   This is the REAL timeout. retries count is secondary.	//   If the sum of all retry attempts exceeds this → fail permanently.	//   TOTAL time from producer.send() to ACK (including retries!).	// delivery.timeout.ms (default: 120000 = 2 minutes)	//	//   With delivery.timeout.ms, this is effectively infinite.	//   Number of times to retry a failed request.	// retries (default: 2147483647 = basically infinite)	// ────────────────────	// RETRY CONFIGURATION:	//	// - UNSUPPORTED_COMPRESSION_TYPE: broker doesn't support the codec	// - TOPIC_AUTHORIZATION_FAILED: ACL denies write	// - INVALID_REQUIRED_ACKS: invalid acks config	// - RECORD_TOO_LARGE: record exceeds max.request.size or message.max.bytes	// ────────────────────────────────────────	// NON-RETRYABLE ERRORS (fail immediately):	//	// - KAFKA_STORAGE_EXCEPTION: broker disk issue (temporary)	// - NETWORK_EXCEPTION: connection lost	// - REQUEST_TIMED_OUT: broker didn't respond in time	// - NOT_LEADER_FOR_PARTITION: metadata stale, refresh and retry on new leader	// - LEADER_NOT_AVAILABLE: broker just became leader, not ready yet	// ──────────────────────────────────────────────────	// RETRYABLE ERRORS (producer retries automatically):	//	// When a produce request fails, the producer retries automatically.	fmt.Println("--- RETRY SEMANTICS ---")func retrySemantics() {// =============================================================================// PART 6: RETRY SEMANTICS — What happens on failure// =============================================================================}	fmt.Println()	fmt.Println("  Keys: murmur2(key) % numPartitions (deterministic)")	fmt.Println("  50-100% throughput improvement over round-robin for null-key records")	fmt.Println("  Sticky partitioner: fills one batch per partition, then switches")	// without careful planning (see Lesson 05).	// This is why you should NEVER increase partitions on a keyed topic	// maps to a DIFFERENT partition! Existing ordering guarantees break.	// GOTCHA: If you change the number of partitions, the same key	//	// This guarantees: same key → same partition → ordering per key.	// DefaultPartitioner: murmur2(keyBytes) % numPartitions	// ─────────────────────	// WHEN KEY IS NOT NULL:	//	// Throughput improvement: 50-100% for null-key workloads!	// RESULT: Full batches, great compression, fewer network calls.	//	// Record 200-299 → Partition 1 (one full batch)	// Record 100-199 → Partition 7 (one full batch)	// Record 0-99 → Partition 3  (one full batch)	//	// Then switches to another partition.	// Sticks to ONE partition until the batch is full (or linger.ms expires).	// ───────────────────────────────────────────────	// NEW BEHAVIOR (Kafka 2.4+): Sticky Partitioner	//	//   - Terrible throughput	//   - Terrible compression ratio (one record per batch)	//   - 10 network requests (one per partition per broker)	//   - 10 batches of 1 record each	// If you have 10 partitions and send 10 records:	// PROBLEM: Each batch has only 1 record!	//	// Record 3 → Partition 0 (wrap around)	// Record 2 → Partition 2	// Record 1 → Partition 1	// Record 0 → Partition 0	// Each recordgoes to the next partition.	// ─────────────────────────────────────────────	// OLD BEHAVIOR (before Kafka 2.4): Round-Robin	//	// WHEN KEY IS NULL (no explicit key):	fmt.Println("--- STICKY PARTITIONER ---")func stickyPartitioner() {// =============================================================================// PART 5: STICKY PARTITIONER — Why round-robin was replaced// =============================================================================}	fmt.Println()	fmt.Println("  Requires: acks=all, max.in.flight ≤ 5, retries=MAX")	fmt.Println("  Broker deduplicates: same sequence → ACK without writing")	fmt.Println("  Solution: PID + sequence number per (producer, partition)")	fmt.Println("  Problem: retries cause duplicates")	// This is why very large numbers of transient producers can be problematic.	// 100K active producers × 100 partitions = 10M entries × 200B = ~2 GB	// Memory impact: ~200 bytes per active (PID, partition) pair.	//	// To survive broker restart.	//   /var/kafka-data/topic-partition/00000000000000123456.snapshot	// This state is also written to a snapshot file on disk:	//	//   ProducerStateEntry = { lastSequence, lastTimestamp, epoch }	//   Map<PID, Map<Partition, ProducerStateEntry>>	// The broker keeps in memory:	// ─────────────────────────────	// BROKER STATE FOR IDEMPOTENCE:	//	// log (sequence numbers are stored in the record batch headers).	// If the leader changes, the new leader has this info from the replicated	// Gotcha: PID + sequence is tracked per-partition ON THE LEADER.	//	// the broker's old sequence tracking for the old PID is irrelevant.	// When a producer gets a new PID (restart or transaction boundary),	// ────────────────────────────	// THE PRODUCER ID (PID) EPOCH:	//	//   Idempotence implies acks=all. You can't have idempotent + acks=1.	// acks=all (forced by idempotence)	//	//   Idempotence needs retries to work.	// retries > 0 (default: Integer.MAX_VALUE)	//	//   Without idempotence, max.in.flight=1 is needed for ordering.	//   How? The broker rejects out-of-order batches and the producer retries.	//   With idempotence, Kafka guarantees ordering even with 5 in-flight!	// max.in.flight.requests.per.connection ≤ 5 (default: 5)	// ──────────────────────────	// WHAT IDEMPOTENCE REQUIRES:	//	// └──────────────────────────────────────────────────────────────┘	// │  ◄───── ACK ──────────────────                               │	// │                                   Write [C], expected = 3    │	// │  Send(seq=2, records=[C]) ──────► 2 == 2? YES                │	// │                                                              │	// │  ◄───── ACK ──────────────────                               │	// │                                    ACK without writing!      │	// │  RETRY(seq=0, records=[A,B]) ──►  0 < 2? DUPLICATE           │	// │                                                              │	// │  ◄───── ACK (lost!) ───────────                              │	// │                                    Write [A,B], expected = 2 │	// │  Send(seq=0, records=[A,B]) ────►  0 == 0? YES               │	// │                                                              │	// │  PID=5                             expected_seq[5][0] = 0    │	// │  Producer                          Broker (Partition 0)      │	// │                                                              │	// │  IDEMPOTENT PRODUCER DEDUP FLOW:                              │	// ┌──────────────────────────────────────────────────────────────┐	//	//    - NO (same or lower) → DUPLICATE → ACK without writing (dedup!)	//    - NO (higher) → OUT_OF_ORDER_SEQUENCE error → producer retries	//    - YES → write to log, increment expected, ACK	//    - Broker checks: is batch.sequence == expected?	// 3. On each produce:	//	//    - Broker tracks: expectedSequence[PID][partition]	//    - Incremented by the number of records in the batch	//    - Starts at 0 for each (PID, partition) pair	// 2. Each batch gets a SEQUENCE NUMBER (per partition)	//	//      For cross-restart dedup, you need TRANSACTIONS (Lesson 07)	//    - DOES NOT survive producer restart (new PID on restart)	//    - Survives transient failures (stored in memory)	//    - PID is a unique int64	// 1. On first send, broker assigns the producer a PRODUCER ID (PID)	// ─────────────	// HOW IT WORKS:	//	// SOLUTION: Idempotent producer (enable.idempotence=true, default since Kafka 3.0)	//	// 5. Broker writes it AGAIN → DUPLICATE DATA!	// 4. Producer RETRIES the same batch	// 3. ACK is LOST (network blip)	// 2. Broker writes batch to log → ACK response	// 1. Producer sends batch to broker	// Scenario:	//	// PROBLEM: Without idempotency, retries cause DUPLICATES.	fmt.Println("--- IDEMPOTENT PRODUCER ---")func idempotentProducer() {// =============================================================================// PART 4: IDEMPOTENT PRODUCER — Exactly-once at the produce level// =============================================================================}	fmt.Println()	fmt.Println("  Set topic compression.type=producer (avoid broker recompression)")	fmt.Println("  Compression is end-to-end: producer → broker → consumer")	fmt.Println("  lz4: fastest (use when CPU-bound)")	fmt.Println("  zstd: best ratio + good speed (use this by default)")	// CPU per batch: lz4 < zstd << gzip	// gzip:           14 KB network, 14 KB disk (7.1x ratio)	// zstd:           15 KB network, 15 KB disk (6.7x ratio)	// lz4:            22 KB network, 22 KB disk (4.5x ratio)	// No compression: 100 KB network, 100 KB disk	// ─────────────────────────────────────────────	// REAL NUMBERS (1 KB JSON records, batch of 100):	//	// This is another reason bigger batches = better throughput.	// 100 individually compressed records.	// A batch of 100 similar JSON records compresses MUCH better than	// Compression works on the ENTIRE BATCH, not individual records.	// ───────────────────────────────	// COMPRESSION & BATCHING SYNERGY:	//	// and control compression on the producer side.	// BEST PRACTICE: Set topic compression.type=producer (default)	// This kills zero-copy and adds CPU load on the broker.	// the broker MUST decompress (lz4) and recompress (zstd).	// If the topic has compression.type=zstd but the producer sends lz4,	// THE EXCEPTION: compression.type on the topic	//	// This is WHY zero-copy works: bytes flow from disk to network unchanged.	// The broker NEVER decompresses data (with one exception).	//	//   Consumer → decompresses batch → application	//   Broker → sends compressed bytes as-is → Consumer	//   Broker → stores compressed bytes as-is → Disk	//   Producer → compresses batch → sends compressed bytes → Broker	// END-TO-END compression model:	// ──────────────────────────	// WHERE COMPRESSION HAPPENS:	//	// NEVER use gzip in production: CPU cost is too high for the marginal gain.	// USE snappy if: You're on older Kafka versions without zstd support.	// USE lz4 if: CPU is your bottleneck and you can't afford any extra cycles.	//	//   Level 19+: slow, ratio approaching gzip's best	//   Level 1: fastest, ratio similar to lz4	//   Level 3 (default): good balance	// - Has adjustable compression levels (1-22)	// - Speed close to lz4	// - Compression ratio close to gzip	// zstd is the WINNER for almost all use cases.	// ───────────────	// RECOMMENDATION:	//	// └───────────────┴────────────┴────────────┴──────────────────────┘	// │ zstd          │ Best (5-8x)│ Fast       │ Medium (tunable)     │	// │ lz4           │ Good (3-4x)│ Fastest    │ Very Low             │	// │ snappy        │ Good (3-4x)│ Fast       │ Low                  │	// │ gzip          │ Best (5-8x)│ Slowest    │ Highest              │	// │ none          │ 1.0x       │ ∞          │ 0                    │	// ├───────────────┼────────────┼────────────┼──────────────────────┤	// │ Codec         │ Ratio      │ Speed      │ CPU Usage            │	// ┌───────────────┬────────────┬────────────┬──────────────────────┐	//	// Kafka supports 4 compression codecs:	fmt.Println("--- COMPRESSION STRATEGY ---")func compressionStrategy() {// =============================================================================// PART 3: COMPRESSION — When, where, and which codec// =============================================================================}	fmt.Println()	fmt.Println("  Both → many partitions + linger.ms=0 + batch naturally fills up")	fmt.Println("  Low latency → small batch.size + linger.ms=0-5")	fmt.Println("  High throughput → big batch.size + linger.ms=50-200")	fmt.Println("  batch.size: max batch bytes. linger.ms: max wait time.")	//   → Expect 1-5ms p99 produce latency	//   batch.size=16KB-64KB, linger.ms=0-5, compression=lz4 or snappy	// For latency optimization:	//	//   → Expect 500K-2M records/sec per producer instance	//   batch.size=512KB-1MB, linger.ms=50-200, compression=zstd	// For throughput optimization:	//	// Too many records per batch: high latency, memory pressure	// Too few records per batch: high overhead per record	// ─────────────────────────────	// RECORDS-PER-BATCH SWEET SPOT:	//	// Also: broker's message.max.bytes must be ≥ max.request.size	// Rule: max.request.size > batch.size (always!)	// If batch.size > max.request.size, you'll get errors.	// Maximum total size of a ProduceRequest (can contain multiple batches).	// MAX.REQUEST.SIZE (default: 1 MB):	//	// After compression, the actual network payload is smaller.	// So the in-memory batch is UNCOMPRESSED, uses full batch.size memory.	// When compression is enabled, the batch is compressed just before sending.	//	// (same format that's stored on the broker — no conversion needed).	// Records are appended into the batch in the RecordBatch wire format	// Each batch is allocated from buffer.memory as a ByteBuffer.	// ────────────────────	// BATCH MEMORY LAYOUT:	//	// └──────────────────────────────────────────────────────────────┘	// │  → Tradeoff: +100ms latency for better throughput             │	// │  → Batches accumulate for 100ms → reasonable batch sizes     │	// │  Low throughput + linger.ms=100:                              │	// │                                                              │	// │  → Set linger.ms=50-100 to allow batches to accumulate       │	// │  → Many small network requests → poor throughput              │	// │  → Batches are tiny (1-2 records each)                       │	// │  Low throughput (100 records/sec) + linger.ms=0:             │	// │                                                              │	// │  → Increase batch.size to get bigger batches                 │	// │  → linger.ms is irrelevant (batches are always full)         │	// │  → Batches fill up to batch.size before Sender drains them   │	// │  High throughput (1M+ records/sec) + linger.ms=0:            │	// │                                                              │	// │  SCENARIOS:                                                   │	// │                                                              │	// │  5. Memory pressure (buffer.memory running low)              │	// │  4. flush() or close() is called                             │	// │  3. Another batch for same broker is ready (piggybacking)    │	// │  2. linger.ms expired since first record was added           │	// │  1. batch.size reached (batch is "full")                     │	// │                                                              │	// │  Batch is sent when EITHER condition is met:                  │	// ┌──────────────────────────────────────────────────────────────┐	//	// HOW THEY INTERACT:	//	//   100 = aggressive batching → great throughput, +100ms latency	//   5 = wait up to 5ms for more records → better batching, +5ms latency	//       if records accumulate faster than sending)	//   0 = send immediately when Sender thread is available (still batches	//   How long to wait for more records before sending a batch.	// linger.ms (default: 0)	//	//   Typical production: 64 KB - 1 MB	//   Larger batch = better compression, fewer requests, higher throughput.	//   When a batch reaches this size → Sender drains it immediately.	//   Maximum size of a single batch in bytes.	// batch.size (default: 16384 = 16 KB)	//	// TWO KNOBS CONTROL BATCHING:	//	// A batch amortizes: network overhead, disk I/O, compression, CRC computation.	// Batching is THE single most important throughput optimization in Kafka.	fmt.Println("--- BATCHING DEEP DIVE ---")func batchingDeepDive() {// =============================================================================// PART 2: BATCHING — The secret to Kafka's throughput// =============================================================================}	fmt.Println()	fmt.Println("  RecordAccumulator bridges the two with per-partition batch queues.")	fmt.Println("  App thread: fast memory ops. Sender thread: network + compression.")	fmt.Println("  Pipeline: Serialize → Partition → Accumulate → Send (async)")	// It's a ConcurrentMap<TopicPartition, Deque<RecordBatch>>.	// The RecordAccumulator is the bridge between the two threads.	//	//    Makes actual network calls. Handles retries.	//    Background thread. One per producer instance.	// 2. Sender thread: drain batches → compress → send to broker → handle response	//	//    UNLESS buffer.memory is full → blocks for max.block.ms	//    This is YOUR thread. It's fast — just memory operations.	// 1. Application thread: serialize → partition → append to batch	// ────────────	// TWO THREADS:	//	// └──────────────────────────────────────────────────────────────────┘	// │                                             └─────────────────┘ │	// │  └─────────────────────────────────────┘    │  Retries on fail│ │	// │  │  Total memory: buffer.memory (32MB) │    │  Handles ACK    │ │	// │  │                                     │    │  Sends to broker│ │	// │  │  Partition 2: [batch][batch][batch]  │    │  Creates request│ │	// │  │  Partition 1: [batch]                │    │  Groups by broker│	// │  │  Partition 0: [batch][batch]         │    │  Drains batches │ │	// │  │                                     │    │                 │ │	// │  │ (per-partition batching buffer)      │───►│  Sender Thread  │ │	// │  │ RecordAccumulator                    │    ┌─────────────────┐ │	// │  ┌─────────────────────────────────────┐                         │	// │         ▼                                                        │	// │  └──────┬──────┘                                                 │	// │  │             │ - Key == null: sticky partitioner               │	// │  │             │ - Key != null: murmur2(key) % numPartitions     │	// │  │ Partitioner │ Determines target partition                     │	// │  ┌─────────────┐                                                 │	// │         ▼                                                        │	// │  └──────┬──────┘                                                 │	// │  │ (key+value) │ Converts objects → byte[]                       │	// │  │ Serializer  │ Key serializer + Value serializer               │	// │  ┌─────────────┐                                                 │	// │         ▼                                                        │	// │  └──────┬──────┘                                                 │	// │  │ Interceptors│ (optional, rarely used in prod)                 │	// │  ┌─────────────┐                                                 │	// │       ▼                                                          │	// │       │                                                          │	// │  record.Send()                                                   │	// │                                                                  │	// │  ═══════════════════                   ═════════════════════════ │	// │  Application Thread                    Sender Thread (background)│	// │                                                                  │	// │                   PRODUCER INTERNAL PIPELINE                      │	// ┌──────────────────────────────────────────────────────────────────┐	//	// When you call producer.Send(record), here's the FULL internal flow:	fmt.Println("--- PRODUCER PIPELINE ---")func producerPipeline() {// =============================================================================// PART 1: THE PRODUCER PIPELINE// =============================================================================}	producerTuningMatrix()	producerMemoryManagement()	retrySemantics()	stickyPartitioner()	idempotentProducer()	compressionStrategy()	batchingDeepDive()	producerPipeline()	fmt.Println()	fmt.Println("=== PRODUCER INTERNALS ===")func main() {import "fmt"package main// =============================================================================//// Understanding this pipeline is the difference between 10K msg/sec and 1M msg/sec.// async pipeline with batching, compression, connection pooling, and retry logic.// The producer is NOT a simple "send and forget" client. It's a sophisticated// THE KEY INSIGHT://// - Memory management: buffer.memory, max.block.ms, and backpressure// - Retry semantics: what happens on failure and how ordering is preserved// - Sticky partitioner: why it replaced round-robin as default// - Idempotent producer: sequence numbers, producer IDs, exactly-once semantics// - Compression: when, where, and which codec (zstd wins almost always)// - Batching internals: how linger.ms and batch.size interact// - The complete internal pipeline: serializer → partitioner → accumulator → sender// WHAT YOU'LL LEARN://// =============================================================================// LESSON 3.1: PRODUCER INTERNALS — Every Byte from Send() to Broker ACK// =============================================================================