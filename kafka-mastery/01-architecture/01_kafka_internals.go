//go:build ignore
// =============================================================================
// LESSON 1.1: KAFKA ARCHITECTURE & INTERNALS — The Brain of the System
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - What actually IS Kafka at the fundamental level
// - Broker internals: network threads, request handlers, purgatory
// - Controller: single-brain vs KRaft quorum
// - The request/response protocol at the wire level
// - Why Kafka is fundamentally different from RabbitMQ/SQS/Pulsar
//
// THE ONE INSIGHT THAT CHANGES EVERYTHING:
// Kafka is NOT a message queue. It's a distributed, partitioned, replicated
// commit log service. Messages aren't "consumed" — they're read from an offset
// in an immutable log. This single design choice makes everything else possible:
// replay, multi-consumer, compaction, infinite retention, and massive throughput.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== KAFKA ARCHITECTURE & INTERNALS ===")
	fmt.Println()

	// =========================================================================
	// PART 1: WHAT IS KAFKA — THE REAL ANSWER
	// =========================================================================
	//
	// Most people say "Kafka is a distributed messaging system." WRONG.
	// Kafka is a DISTRIBUTED COMMIT LOG.
	//
	// The difference matters:
	//
	// MESSAGE QUEUE (RabbitMQ, SQS):
	//   - Message is delivered to ONE consumer, then DELETED
	//   - Queue is a transient buffer
	//   - No replay capability
	//   - Ordering only within a single queue
	//   - Scales by adding more queues (no automatic partitioning)
	//
	// COMMIT LOG (Kafka):
	//   - Message is APPENDED to a log, stays FOREVER (until retention expires)
	//   - Any number of consumers can read from any position
	//   - Full replay capability — consumers track their own position
	//   - Ordering guaranteed within a partition
	//   - Scales by adding partitions (horizontal by design)
	//
	// This is why:
	//   - You can have 100 consumer groups reading the same data
	//   - You can "rewind" a consumer to reprocess a week of data
	//   - You can use Kafka as both a message bus AND a database (compacted topics)
	//   - LinkedIn processes 7+ TRILLION messages/day on Kafka
	//
	// ┌─────────────────────────────────────────────────────────────────┐
	// │                    THE KAFKA DATA MODEL                         │
	// │                                                                 │
	// │  Cluster ──► Broker(s)                                          │
	// │              └── Topic(s)                                        │
	// │                  └── Partition(s)                                │
	// │                      └── Segment(s)                             │
	// │                          └── Record(s)                          │
	// │                              ├── Key (optional, bytes)          │
	// │                              ├── Value (bytes)                  │
	// │                              ├── Headers (key-value pairs)      │
	// │                              ├── Timestamp                      │
	// │                              └── Offset (assigned by broker)    │
	// │                                                                 │
	// │  CRITICAL: Offset is immutable. Once record gets offset 42,    │
	// │  it's ALWAYS at offset 42. This enables deterministic replay.  │
	// └─────────────────────────────────────────────────────────────────┘

	brokerAnatomy()
	controllerRole()
	requestLifecycle()
	networkArchitecture()
	kafkaVsOthers()
}

// =============================================================================
// PART 2: BROKER ANATOMY — What's Inside a Kafka Broker
// =============================================================================
func brokerAnatomy() {
	fmt.Println("--- BROKER ANATOMY ---")

	// A Kafka broker is a JVM process with these key components:
	//
	// ┌────────────────────────────────────────────────────────────────────┐
	// │                         KAFKA BROKER                              │
	// │                                                                   │
	// │  ┌─────────────┐   ┌──────────────────┐   ┌──────────────────┐   │
	// │  │  Network     │   │  Request          │   │  Request         │   │
	// │  │  Threads     │──►│  Queue            │──►│  Handler         │   │
	// │  │  (Acceptor + │   │  (request.queue   │   │  Threads         │   │
	// │  │   Processor) │   │   .max.size)      │   │  (num.io.threads)│   │
	// │  └─────────────┘   └──────────────────┘   └──────────────────┘   │
	// │        │                                          │               │
	// │        │                                          ▼               │
	// │        │           ┌──────────────────┐   ┌──────────────────┐   │
	// │        │           │  Response Queue   │◄──│ Log Manager      │   │
	// │        │◄──────────│                   │   │ (append, read,   │   │
	// │        │           │                   │   │  flush, compact) │   │
	// │        │           └──────────────────┘   └──────────────────┘   │
	// │        │                                          │               │
	// │        ▼                                          ▼               │
	// │  ┌─────────────┐                          ┌──────────────────┐   │
	// │  │  Client      │                          │  Disk / Page     │   │
	// │  │  (Producers  │                          │  Cache           │   │
	// │  │   & Consumers│                          │  (OS-managed)    │   │
	// │  └─────────────┘                          └──────────────────┘   │
	// │                                                                   │
	// │  OTHER KEY COMPONENTS:                                            │
	// │  ├── ReplicaManager: handles fetch from followers, ISR tracking   │
	// │  ├── GroupCoordinator: manages consumer group membership + offsets│
	// │  ├── TransactionCoordinator: manages producer transactions        │
	// │  ├── Purgatory: holds requests waiting for conditions             │
	// │  │   (e.g., produce waiting for ISR acks, fetch waiting for data)│
	// │  └── KafkaApis: routes requests to appropriate handler            │
	// └────────────────────────────────────────────────────────────────────┘
	//
	// NETWORK THREADS:
	// ────────────────
	// num.network.threads (default: 3)
	// - Acceptor thread: accepts new TCP connections via Java NIO Selector
	// - Processor threads: read requests from socket → request queue
	//                      read responses from response queue → socket
	// - These are NON-BLOCKING I/O threads — they never touch disk
	// - They use Java NIO (epoll on Linux) for multiplexing
	//
	// KEY INSIGHT: Network threads should NEVER be the bottleneck.
	// If network threads are saturated, you have a serious problem because
	// they don't do any heavy lifting — they just shuttle bytes.
	// Monitor: kafka.network:type=SocketServer,name=NetworkProcessorAvgIdlePercent
	// If < 0.3, you have a network thread bottleneck.
	//
	// REQUEST HANDLER THREADS:
	// ────────────────────────
	// num.io.threads (default: 8)
	// - These actually PROCESS requests: append to log, read from log, etc.
	// - Despite the name "io.threads", most reads come from page cache (RAM)
	// - Writes are sequential appends (very fast even on spinning disks)
	//
	// Monitor: kafka.server:type=KafkaRequestHandlerPool,name=RequestHandlerAvgIdlePercent
	// If < 0.3, add more io threads or add brokers.
	//
	// PURGATORY:
	// ──────────
	// The most clever design in Kafka's broker.
	// Some requests can't be satisfied immediately:
	//   - ProduceRequest with acks=all: must wait for ALL ISR replicas to ack
	//   - FetchRequest with min.bytes: must wait for enough data to accumulate
	//
	// Instead of blocking a handler thread (which would be TERRIBLE for throughput),
	// Kafka puts the request in "purgatory" — a timer-based holding area.
	// When the condition is met (or timeout expires), the response is completed.
	//
	// This is essentially an implementation of the Reactor pattern with
	// delayed completion. It's why Kafka can handle 100K+ concurrent
	// produce requests with acks=all without running out of threads.

	fmt.Println("  A Kafka broker is NOT just a 'message store'.")
	fmt.Println("  It's a carefully designed reactor-pattern server with:")
	fmt.Println("  - Non-blocking network I/O (Java NIO / epoll)")
	fmt.Println("  - Request handler pool (actual processing)")
	fmt.Println("  - Purgatory (delayed completion for acks=all, fetch min.bytes)")
	fmt.Println("  - OS page cache for reads (zero-copy with sendfile)")
	fmt.Println()
}

// =============================================================================
// PART 3: THE CONTROLLER — Kafka's Single Brain
// =============================================================================
func controllerRole() {
	fmt.Println("--- CONTROLLER ROLE ---")

	// In classic Kafka (with ZooKeeper):
	// - ONE broker is elected as the Controller
	// - The Controller manages ALL cluster-level operations:
	//   1. Partition leader election (when a broker fails)
	//   2. Partition reassignment (when you rebalance)
	//   3. Topic creation/deletion
	//   4. ISR changes (when a replica falls behind or catches up)
	//
	// Controller election (ZooKeeper mode):
	// - All brokers try to create an ephemeral ZNode /controller
	// - First one wins, becomes the Controller
	// - Others watch the ZNode — if it disappears (Controller dies),
	//   they race to recreate it
	//
	// WHY SINGLE CONTROLLER?
	// ─────────────────────
	// It might seem like a bottleneck, but it's actually brilliant:
	// - Metadata operations are RARE compared to produce/consume
	// - Single writer avoids split-brain and consistency issues
	// - The Controller only handles METADATA, not data flow
	//
	// CONTROLLER PROBLEMS AT SCALE:
	// ─────────────────────────────
	// At LinkedIn-scale (100K+ partitions), the Controller was a bottleneck:
	// - When a broker with 10K partitions dies, the Controller must elect
	//   10K new leaders — this takes time
	// - During this "controlled shutdown", those partitions are UNAVAILABLE
	// - This is one reason KRaft was developed (see lesson 14)
	//
	// ┌───────────────────────────────────────────────────────────────┐
	// │              CONTROLLER FAILOVER TIMELINE                     │
	// │                                                               │
	// │  t=0    Broker-3 (controller) dies                            │
	// │  t=6s   ZooKeeper detects loss (session timeout)              │
	// │  t=6.1s Broker-1 wins controller election                     │
	// │  t=6.1s New controller reads FULL metadata from ZooKeeper     │
	// │  t=7s   Controller begins partition leader elections           │
	// │  t=7-9s Leaders elected for 10K partitions (sequential!)      │
	// │  t=9s   Cluster fully operational                             │
	// │                                                               │
	// │  TOTAL UNAVAILABILITY: ~3 seconds for affected partitions     │
	// │                                                               │
	// │  With KRaft: This is reduced to < 1 second because:           │
	// │  - No ZooKeeper session timeout wait                          │
	// │  - Metadata is replicated via Raft (already available)        │
	// │  - Controller failover is a Raft leader election (~100ms)     │
	// └───────────────────────────────────────────────────────────────┘

	fmt.Println("  Controller is the metadata brain of the cluster.")
	fmt.Println("  It handles: leader election, ISR changes, topic management.")
	fmt.Println("  NOT on the hot path — produce/consume bypass the controller.")
	fmt.Println()
}

// =============================================================================
// PART 4: REQUEST LIFECYCLE — What happens when you call produce/consume
// =============================================================================
func requestLifecycle() {
	fmt.Println("--- REQUEST LIFECYCLE ---")

	// PRODUCE REQUEST LIFECYCLE (bird's eye):
	// ────────────────────────────────────────
	//
	// 1. CLIENT: Serializes key+value, applies partitioner → knows target partition
	// 2. CLIENT: Batches records by broker (RecordAccumulator)
	// 3. CLIENT: Sender thread picks a batch, creates ProduceRequest
	// 4. CLIENT: Compresses the batch (if configured)
	// 5. CLIENT: Sends to the LEADER broker for that partition (via metadata cache)
	//
	// 6. BROKER: Network thread reads request from socket
	// 7. BROKER: Puts on request queue
	// 8. BROKER: Request handler thread picks it up
	// 9. BROKER: Log.append() — writes to active segment file
	//    - Validates CRC, magic byte, compression codec
	//    - Assigns offsets (base offset + index within batch)
	//    - Appends to OS page cache (NOT fsync'd yet!)
	//    - Updates LEO (Log End Offset)
	//
	// 10. IF acks=0: Response sent immediately (may lose data!)
	// 11. IF acks=1: Response sent after leader writes (may lose if leader dies)
	// 12. IF acks=all: Request goes to Purgatory
	//     - Waits for ALL ISR replicas to fetch and ack
	//     - When last ISR replica catches up → complete the response
	//     - If timeout → return error to producer
	//
	// 13. BROKER: Response put on response queue
	// 14. BROKER: Network thread sends response to client
	//
	// ┌──────────────────────────────────────────────────────────────────┐
	// │           PRODUCE REQUEST — INTERNAL FLOW                        │
	// │                                                                  │
	// │  Producer                    Leader Broker                       │
	// │  ════════                    ═══════════════                     │
	// │  serialize()                                                      │
	// │  partition()                                                      │
	// │  batch + compress                                                 │
	// │  ──── ProduceRequest ────►   NetworkThread.read()                │
	// │                              RequestQueue.enqueue()              │
	// │                              IoThread.handle()                    │
	// │                              Log.append()                        │
	// │                              ╔═══════════════════════╗           │
	// │                              ║ if acks=all:          ║           │
	// │                              ║   Purgatory.add()     ║           │
	// │                              ║   wait for ISR fetch  ║           │
	// │                              ╚═══════════════════════╝           │
	// │  ◄── ProduceResponse ─────   ResponseQueue.send()               │
	// │  callback(metadata)                                              │
	// │                                                                  │
	// │  TOTAL LATENCY BREAKDOWN (typical, acks=all, 3 ISR):            │
	// │  ├── Network: 0.1-0.5ms (same DC)                               │
	// │  ├── Request queue wait: 0-2ms                                   │
	// │  ├── Log append: 0.01-0.1ms (page cache)                        │
	// │  ├── ISR replication: 0.5-5ms (depends on network + load)       │
	// │  └── Response: 0.1-0.5ms                                        │
	// │  TOTAL: 1-8ms (same DC, healthy cluster)                        │
	// └──────────────────────────────────────────────────────────────────┘
	//
	// FETCH REQUEST LIFECYCLE (consumer):
	// ───────────────────────────────────
	// 1. Consumer sends FetchRequest with: topic, partition, offset, max.bytes
	// 2. Leader broker reads from log starting at requested offset
	// 3. IF data in page cache: zero-copy transfer (sendfile syscall!)
	//    IF data on disk: page fault → disk read → served from page cache
	// 4. IF not enough data: request goes to Purgatory (fetch.min.bytes)
	// 5. Response sent with RecordBatch + highWatermark
	//
	// ZERO-COPY (sendfile):
	// ─────────────────────
	// This is one of Kafka's secret weapons for throughput.
	// Normal flow: Disk → Page Cache → User Space → Socket Buffer → NIC
	// Zero-copy:   Disk → Page Cache ───────────► Socket Buffer → NIC
	//              (skips the copy to JVM heap entirely!)
	//
	// Java: FileChannel.transferTo() → calls sendfile() on Linux
	// This only works when you DON'T need to transform the data,
	// which is why Kafka stores data in the exact wire format.

	fmt.Println("  Produce: serialize → partition → batch → compress → send to leader")
	fmt.Println("  Consume: fetch from offset → zero-copy from page cache → process")
	fmt.Println("  Zero-copy (sendfile) skips JVM heap entirely for consumer reads.")
	fmt.Println()
}

// =============================================================================
// PART 5: NETWORK ARCHITECTURE — The threading model
// =============================================================================
func networkArchitecture() {
	fmt.Println("--- NETWORK ARCHITECTURE ---")

	// Kafka uses a multi-layered Reactor pattern:
	//
	// LAYER 1: Acceptor Thread (1 per listener)
	// - Accepts new TCP connections
	// - Round-robins new connections to Processor threads
	//
	// LAYER 2: Processor Threads (num.network.threads per listener)
	// - Each manages many connections via Java NIO Selector
	// - Reads complete requests from socket → RequestChannel queue
	// - Reads complete responses from response queue → sends to socket
	// - NON-BLOCKING: never does I/O or any heavy processing
	//
	// LAYER 3: Request Handler Threads (num.io.threads, shared)
	// - Picks requests from RequestChannel queue
	// - Processes them (log append, log read, metadata lookup, etc.)
	// - Puts responses on the per-processor response queue
	//
	// ┌──────────────────────────────────────────────────────────┐
	// │              KAFKA REACTOR PATTERN                        │
	// │                                                          │
	// │  Clients ──► Acceptor ──┬──► Processor-0 ──┐             │
	// │                         ├──► Processor-1 ──┤             │
	// │                         └──► Processor-2 ──┤             │
	// │                                            │             │
	// │                                    ┌───────▼──────┐      │
	// │                                    │ RequestChannel│      │
	// │                                    │ (shared queue)│      │
	// │                                    └───────┬──────┘      │
	// │                                            │             │
	// │                              ┌─────────────┼────────┐    │
	// │                              ▼             ▼        ▼    │
	// │                          Handler-0    Handler-1  ... N   │
	// │                              │             │        │    │
	// │                              ▼             ▼        ▼    │
	// │                          Log/Metadata  operations        │
	// │                              │                           │
	// │                              ▼                           │
	// │                     Response Queues (per processor)      │
	// └──────────────────────────────────────────────────────────┘
	//
	// SIZING RULES OF THUMB:
	// ──────────────────────
	// num.network.threads:
	//   - Default: 3. Good for most clusters.
	//   - Increase if NetworkProcessorAvgIdlePercent < 0.3
	//   - Rarely needs to be > 8 (if so, you have other problems)
	//
	// num.io.threads:
	//   - Default: 8. Should be ≥ number of disks.
	//   - Increase if RequestHandlerAvgIdlePercent < 0.3
	//   - LinkedIn uses 8-16 depending on workload
	//
	// queued.max.requests:
	//   - Default: 500. Max requests in the request queue before blocking.
	//   - If this fills up, network threads BLOCK (backpressure!)
	//   - Monitor: kafka.network:type=RequestChannel,name=RequestQueueSize

	fmt.Println("  Kafka uses a 3-layer Reactor pattern:")
	fmt.Println("  Acceptor → Processor (NIO) → Request Handler (I/O)")
	fmt.Println("  This separates network I/O from disk I/O cleanly.")
	fmt.Println()
}

// =============================================================================
// PART 6: KAFKA vs EVERYTHING ELSE — Why Kafka wins at scale
// =============================================================================
func kafkaVsOthers() {
	fmt.Println("--- KAFKA vs OTHER SYSTEMS ---")

	// ┌─────────────┬──────────────┬──────────────┬──────────────┬──────────────┐
	// │ Feature     │ Kafka        │ RabbitMQ     │ AWS SQS      │ Pulsar       │
	// ├─────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
	// │ Model       │ Commit log   │ Message queue│ Message queue│ Commit log   │
	// │ Delivery    │ Pull (poll)  │ Push         │ Pull         │ Push/Pull    │
	// │ Ordering    │ Per-partition│ Per-queue    │ Best-effort  │ Per-partition│
	// │ Replay      │ Yes (offset) │ No           │ No           │ Yes (cursor) │
	// │ Throughput  │ 1M+ msg/sec  │ 50K msg/sec  │ Managed      │ 1M+ msg/sec  │
	// │ Retention   │ Configurable │ Until consumed│ 14 days max │ Configurable │
	// │ Storage     │ Broker disk  │ Broker RAM/  │ AWS managed  │ Separate     │
	// │             │              │ disk         │              │ (BookKeeper) │
	// │ Consumers   │ Many groups  │ Queue = 1    │ Queue = 1    │ Many subs    │
	// │ Scaling     │ Add partitions│ Not easy    │ Automatic    │ Add partitions│
	// │ Exactly-once│ Yes (EOS)    │ No (at-most) │ No           │ Yes          │
	// └─────────────┴──────────────┴──────────────┴──────────────┴──────────────┘
	//
	// WHY KAFKA WINS FOR EVENT STREAMING:
	// ───────────────────────────────────
	// 1. SEQUENTIAL I/O: Kafka writes/reads sequentially. HDDs do 100+ MB/s
	//    sequential but only 100-200 IOPS random. SSDs still benefit from
	//    sequential patterns due to write amplification avoidance.
	//
	// 2. PAGE CACHE: Kafka doesn't manage its own cache — it lets the OS do it.
	//    This means: No GC pressure, no double-buffering, and the cache survives
	//    broker restarts! This is why Kafka brokers barely use heap.
	//
	// 3. ZERO-COPY: sendfile() syscall transfers data from page cache directly
	//    to the network socket, skipping JVM heap entirely.
	//
	// 4. BATCHING: Producer batches many records into one network request.
	//    Broker writes the batch as a single sequential write.
	//    Consumer fetches many records in one network request.
	//    Amortizes network overhead, disk overhead, and compression overhead.
	//
	// 5. COMPRESSION: Batches are compressed (gzip, snappy, lz4, zstd).
	//    Compression ratio is MUCH better for batches than individual records.
	//    end-to-end: producer compresses, broker stores compressed, consumer decompresses.
	//
	// 6. PULL MODEL: Consumers pull at their own pace.
	//    No need for broker to track delivery status per-consumer.
	//    Consumer manages its own offset — broker just serves bytes.
	//
	// WHEN KAFKA IS THE WRONG CHOICE:
	// ────────────────────────────────
	// - Request/reply patterns: Use RabbitMQ or gRPC
	// - Low-latency sub-millisecond: Use Aeron or shared memory
	// - Small messages (< 100 bytes) with low throughput: Overhead too high
	// - Complex routing: RabbitMQ's exchange model is better
	// - Serverless / no-ops requirement: Use SQS/SNS or managed Kafka

	fmt.Println("  Kafka wins at scale due to:")
	fmt.Println("  1. Sequential I/O (100MB/s+ even on HDD)")
	fmt.Println("  2. OS page cache (no JVM GC pressure)")
	fmt.Println("  3. Zero-copy reads (sendfile syscall)")
	fmt.Println("  4. Batch everything (network, disk, compression)")
	fmt.Println("  5. Pull model (consumers control their own pace)")
	fmt.Println()

	// =========================================================================
	// THE FUNDAMENTAL EQUATION OF KAFKA SCALE:
	// =========================================================================
	//
	// Throughput = (Partitions × Partition_Throughput)
	//            = (Partitions × min(Network_BW, Disk_BW, Consumer_Speed))
	//
	// Latency = Network_RTT + Queue_Wait + Log_Append + Replication_Wait
	//
	// SCALE STRATEGY:
	// ───────────────
	// Want more throughput? → Add partitions + consumers
	// Want less latency?   → Better hardware, tune linger.ms/batch.size
	// Want both?           → More partitions + better hardware
	//
	// THE CEILING:
	// ────────────
	// - Single partition: ~50-100 MB/s (limited by leader broker's disk/network)
	// - Single broker: ~200-500 MB/s aggregate (network-bound typically)
	// - LinkedIn cluster: 100+ brokers, 100K+ partitions, 7T+ messages/day
	// - Uber's cluster: 1300+ brokers, processing 40+ PB/day
	//
	// IMPORTANT: More partitions is NOT always better.
	// See Lesson 05 for the partition ceiling and diminishing returns.

	fmt.Println("  Throughput = Partitions × min(Network, Disk, Consumer Speed)")
	fmt.Println("  LinkedIn: 7T+ messages/day. Uber: 40+ PB/day.")
	fmt.Println("  Scale = more partitions + more brokers + more consumers")
}
