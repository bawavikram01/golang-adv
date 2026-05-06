//go:build ignore
// =============================================================================
// LESSON 7.1: EXACTLY-ONCE SEMANTICS — The Holy Grail of Stream Processing
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - The three delivery guarantees: at-most-once, at-least-once, exactly-once
// - Idempotent producer: exactly-once for single partitions
// - Kafka Transactions: exactly-once across partitions and topics
// - The consume-transform-produce pattern with EOS
// - Limitations and gotchas of exactly-once semantics
// - When exactly-once is impossible (spoiler: external systems)
//
// THE KEY INSIGHT:
// "Exactly-once" in Kafka doesn't mean "magic." It means:
// 1. Idempotent writes prevent duplicates within Kafka
// 2. Transactions make multi-partition writes atomic
// 3. Consumers in read_committed mode skip uncommitted data
// Combined: a consume → process → produce pipeline with no duplicates and no loss.
// BUT: this only works WITHIN KAFKA. Side effects to external systems
// (databases, APIs) still need application-level idempotency.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== EXACTLY-ONCE SEMANTICS ===")
	fmt.Println()

	deliveryGuarantees()
	kafkaTransactions()
	consumeTransformProduce()
	transactionInternals()
	eosLimitations()
}

// =============================================================================
// PART 1: THE THREE DELIVERY GUARANTEES
// =============================================================================
func deliveryGuarantees() {
	fmt.Println("--- DELIVERY GUARANTEES ---")

	// AT-MOST-ONCE:
	// ─────────────
	// "Fire and forget." Record might be lost, never duplicated.
	// Implementation: acks=0, or commit offset BEFORE processing.
	// Use case: Metrics, logging where losing some data is acceptable.
	//
	// AT-LEAST-ONCE:
	// ───────────────
	// "If in doubt, send again." Record might be duplicated, never lost.
	// Implementation: acks=all + retries + commit offset AFTER processing.
	// If processing succeeds but commit fails → re-processing on restart.
	// Use case: Most applications (if idempotent processing is implemented).
	//
	// EXACTLY-ONCE:
	// ─────────────
	// "Every record processed once and only once."
	// Implementation: Idempotent producer + Transactions + read_committed.
	// Use case: Financial systems, state management, CDC.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  DELIVERY GUARANTEES IN PRACTICE:                              │
	// │                                                              │
	// │  AT-MOST-ONCE:                                                │
	// │  produce(acks=0) → might fail silently                       │
	// │  OR: commitOffset → poll → process → crash → LOST            │
	// │                                                              │
	// │  AT-LEAST-ONCE:                                               │
	// │  produce(acks=all) + retry → guaranteed delivery              │
	// │  poll → process → commitOffset → crash → REPROCESS           │
	// │                                                              │
	// │  EXACTLY-ONCE (within Kafka):                                 │
	// │  beginTransaction()                                           │
	// │  produce(recordA) → topic-output                              │
	// │  produce(recordB) → topic-output                              │
	// │  commitOffsets(consumer-group-offsets) → __consumer_offsets   │
	// │  commitTransaction()                                          │
	// │  → ALL succeed atomically, OR all are rolled back             │
	// │  → Records are deduplicated via idempotent producer           │
	// │  → Consumer reads only committed data (isolation.level=       │
	// │    read_committed)                                            │
	// └──────────────────────────────────────────────────────────────┘
	//
	// THE TRUTH ABOUT EXACTLY-ONCE:
	// ─────────────────────────────
	// Exactly-once is only achievable within a CLOSED SYSTEM.
	// Kafka ↔ Kafka: transactions provide exactly-once.
	// Kafka → Database: NOT exactly-once (DB write might succeed, Kafka commit fail).
	//   Solution: make DB writes idempotent (e.g., upsert with offset as key).
	// Kafka → API call: NOT exactly-once (API call is a side effect).
	//   Solution: idempotency key in the API call.

	fmt.Println("  At-most-once: fire and forget (acks=0)")
	fmt.Println("  At-least-once: guaranteed delivery + possible duplicates (acks=all)")
	fmt.Println("  Exactly-once: transactions + idempotent producer (within Kafka only)")
	fmt.Println()
}

// =============================================================================
// PART 2: KAFKA TRANSACTIONS
// =============================================================================
func kafkaTransactions() {
	fmt.Println("--- KAFKA TRANSACTIONS ---")

	// A Kafka transaction atomically writes to MULTIPLE partitions/topics
	// AND commits consumer offsets — all-or-nothing!
	//
	// SETTING UP A TRANSACTIONAL PRODUCER:
	// ────────────────────────────────────
	// transactional.id = "my-app-instance-1"  ← MUST be unique per instance
	// enable.idempotence = true               ← forced by transactions
	// acks = all                              ← forced by transactions
	//
	// TRANSACTIONAL API (pseudo-code):
	//
	// producer.initTransactions()              // One-time init
	// while (true) {
	//     records = consumer.poll()
	//     producer.beginTransaction()
	//     for record in records {
	//         result = process(record)
	//         producer.send(outputTopic, result)
	//     }
	//     producer.sendOffsetsToTransaction(offsets, consumerGroupId)
	//     producer.commitTransaction()
	//     // If ANY step fails: producer.abortTransaction()
	// }
	//
	// WHAT HAPPENS DURING commitTransaction():
	// ─────────────────────────────────────────
	// 1. Producer sends EndTransactionRequest to TransactionCoordinator
	// 2. Coordinator writes PREPARE_COMMIT to __transaction_state topic
	// 3. Coordinator sends WriteTxnMarker to ALL partitions in the transaction
	//    - Each partition leader appends a COMMIT marker to the log
	//    - This marker tells consumers "everything before this in this
	//      transaction is committed and can be read"
	// 4. Coordinator writes COMPLETE_COMMIT to __transaction_state
	//
	// IF CRASH DURING COMMIT:
	// - If before PREPARE_COMMIT: transaction is aborted on recovery
	// - If after PREPARE_COMMIT: transaction is completed on recovery
	//   (two-phase commit: prepare phase is the point of no return)
	//
	// CONSUMER SIDE — read_committed:
	// ────────────────────────────────
	// isolation.level=read_committed (default: read_uncommitted)
	//
	// With read_committed, the consumer:
	// - Only reads records up to the LSO (Last Stable Offset)
	// - LSO = the offset of the first record in an OPEN transaction
	// - Records in aborted transactions are SKIPPED
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  TRANSACTION VIEW OF A PARTITION:                              │
	// │                                                              │
	// │  [A:tx1][B:tx1][C:tx2][D:tx1][COMMIT:tx1][E:tx2][ABORT:tx2]│
	// │                                                              │
	// │  read_uncommitted: reads ALL records (A,B,C,D,E)             │
	// │  read_committed:   reads only A,B,D (tx1 committed)          │
	// │                    skips C,E (tx2 aborted)                    │
	// │                                                              │
	// │  LSO moves forward as transactions complete.                 │
	// │  If a transaction stays open for too long:                    │
	// │  LSO is stuck → consumer can't read new records!             │
	// │  transaction.max.timeout.ms (default: 900000 = 15 min)       │
	// │  Broker aborts transactions older than this.                  │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  Transactions: atomic write to multiple partitions + offset commit")
	fmt.Println("  Consumer: isolation.level=read_committed to skip aborted records")
	fmt.Println("  transactional.id must be UNIQUE per producer instance")
	fmt.Println()
}

// =============================================================================
// PART 3: CONSUME-TRANSFORM-PRODUCE — The EOS Pattern
// =============================================================================
func consumeTransformProduce() {
	fmt.Println("--- CONSUME-TRANSFORM-PRODUCE (CTP) ---")

	// The holy grail pattern: read input → process → write output
	// With EXACTLY-ONCE guarantees end-to-end within Kafka.
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  THE CTP PATTERN:                                             │
	// │                                                              │
	// │  Input Topic ──► Consumer ──► Business Logic ──► Producer    │
	// │                                                  │           │
	// │                                                  ▼           │
	// │                                            Output Topic      │
	// │                                                              │
	// │  IN A SINGLE TRANSACTION:                                    │
	// │  1. Read records from input topic                             │
	// │  2. Process them (transform, aggregate, filter)              │
	// │  3. Write results to output topic                             │
	// │  4. Commit input offsets                                      │
	// │  →  All succeed atomically or all fail                       │
	// │                                                              │
	// │  IF CRASH AFTER PROCESSING, BEFORE COMMIT:                   │
	// │  - Transaction is aborted                                     │
	// │  - Consumer re-reads from last committed offset               │
	// │  - Processing is retried                                      │
	// │  - No duplicates in output (same transaction was aborted)    │
	// │                                                              │
	// │  IF CRASH AFTER PREPARE_COMMIT:                               │
	// │  - Transaction is completed by new coordinator                │
	// │  - Still no duplicates                                        │
	// └──────────────────────────────────────────────────────────────┘
	//
	// TRANSACTIONAL.ID AND FENCING:
	// ─────────────────────────────
	// When a new producer instance with the same transactional.id starts:
	// 1. It calls initTransactions()
	// 2. The TransactionCoordinator sees an existing transactional.id
	// 3. It increments the EPOCH for that transactional.id
	// 4. Any old producer with the same transactional.id and lower epoch
	//    is FENCED: it gets ProducerFencedException on next request
	//    → It can no longer produce, preventing "zombie" producers
	//
	// This is critical for exactly-once: only ONE producer can be active
	// for a given transactional.id at any time.
	//
	// TRANSACTIONAL.ID NAMING:
	// ────────────────────────
	// Best practice: transactional.id = appId + "-" + inputPartition
	// One producer per input partition → each handles exactly one partition's data
	//
	// Example:
	//   transactional.id = "order-processor-partition-0"
	//   transactional.id = "order-processor-partition-1"
	//   transactional.id = "order-processor-partition-2"
	//
	// After rebalance, if partition-0 moves to a new instance:
	// New instance uses same transactional.id → fences old instance → clean handoff.

	fmt.Println("  CTP: consume + transform + produce + commit offsets in ONE transaction")
	fmt.Println("  Zombie fencing: same transactional.id → old producer is fenced out")
	fmt.Println("  Best: transactional.id = appId + inputPartition")
	fmt.Println()
}

// =============================================================================
// PART 4: TRANSACTION INTERNALS
// =============================================================================
func transactionInternals() {
	fmt.Println("--- TRANSACTION INTERNALS ---")

	// TRANSACTION COORDINATOR:
	// ────────────────────────
	// A specific broker responsible for a transactional.id.
	// Determined by: hash(transactional.id) % __transaction_state partitions
	// (Similar to how GroupCoordinator works for consumer groups)
	//
	// __transaction_state topic (internal, compacted):
	// Key: transactional.id
	// Value: transaction metadata (state, epoch, partitions, timeout)
	//
	// TRANSACTION STATE MACHINE:
	//
	// ┌──────────┐
	// │  Empty   │
	// └────┬─────┘
	//      │ initTransactions()
	//      ▼
	// ┌──────────┐
	// │  Init    │ ◄─── abortTransaction()
	// └────┬─────┘
	//      │ beginTransaction() + first send()
	//      ▼
	// ┌──────────┐
	// │ Ongoing  │ ◄─── more send() calls add partitions
	// └────┬─────┘
	//      │ commitTransaction() / abortTransaction()
	//      ├──────────────────────────┐
	//      ▼                          ▼
	// ┌────────────────┐    ┌──────────────────┐
	// │ PrepareCommit  │    │ PrepareAbort     │
	// └────┬───────────┘    └──────┬───────────┘
	//      │ markers written       │ markers written
	//      ▼                       ▼
	// ┌──────────┐          ┌──────────────┐
	// │ Complete │          │ CompleteAbort│
	// │ Commit   │          │              │
	// └──────────┘          └──────────────┘
	//
	// PERFORMANCE IMPACT OF TRANSACTIONS:
	// ───────────────────────────────────
	// 1. Extra writes to __transaction_state (~3 per transaction)
	// 2. Transaction markers appended to each partition in the transaction
	// 3. Consumers with read_committed have slightly higher latency
	//    (must wait for transaction to complete to read records)
	//
	// For high throughput: minimize the number of transactions:
	//   Don't commit after EVERY record.
	//   Commit after a BATCH of records (e.g., one poll() batch).
	//   transaction.timeout.ms (default: 60s) bounds how long a transaction can stay open.
	//
	// TYPICAL OVERHEAD:
	// - Extra latency: 5-15ms per transaction (for coordinator round-trips)
	// - Extra disk: ~100 bytes per transaction marker per partition
	// - Throughput impact: 3-20% reduction vs non-transactional

	fmt.Println("  TransactionCoordinator manages state via __transaction_state topic")
	fmt.Println("  State machine: Empty → Init → Ongoing → PrepareCommit → CompleteCommit")
	fmt.Println("  Performance: 5-15ms overhead per transaction, 3-20% throughput impact")
	fmt.Println()
}

// =============================================================================
// PART 5: EOS LIMITATIONS — When exactly-once doesn't work
// =============================================================================
func eosLimitations() {
	fmt.Println("--- EOS LIMITATIONS ---")

	// LIMITATION 1: External side effects
	// ────────────────────────────────────
	// Kafka transactions only cover KAFKA operations.
	// If your processing sends an email, calls an API, or writes to a DB:
	//   consumer.poll() → process() → sendEmail() → producer.send()
	// The email is sent OUTSIDE the transaction. If the transaction aborts:
	//   Email already sent → can't un-send → NOT exactly-once.
	//
	// FIX: Make side effects idempotent.
	//   - Database: use upsert with (transactional.id, offset) as dedup key
	//   - API: include idempotency key in request
	//   - Email: check "already sent" flag before sending
	//
	// LIMITATION 2: Cross-cluster transactions
	// ─────────────────────────────────────────
	// Kafka transactions work WITHIN a single Kafka cluster.
	// If you have MirrorMaker 2 replicating across clusters:
	//   Transaction boundaries are NOT preserved.
	//   You get at-least-once across clusters, not exactly-once.
	//
	// LIMITATION 3: Consumer rebalance during transaction
	// ───────────────────────────────────────────────────
	// If a rebalance occurs while a transaction is open:
	//   - The old consumer's transaction must be aborted
	//   - The new consumer starts fresh from last committed offset
	//   - No data loss, but the aborted transaction's processing is wasted
	//   Static group membership helps reduce this.
	//
	// LIMITATION 4: Performance overhead
	// ──────────────────────────────────
	// Transactions add 5-15ms latency per commit.
	// For sub-millisecond latency requirements, EOS adds too much overhead.
	//
	// LIMITATION 5: Zombie detection depends on transactional.id
	// ──────────────────────────────────────────────────────────
	// If two producer instances use DIFFERENT transactional.ids,
	// they won't fence each other → possible duplicates.
	// The transactional.id scheme MUST guarantee uniqueness per "logical producer."
	//
	// PRACTICAL ADVICE:
	// ─────────────────
	// - Use EOS for Kafka-to-Kafka pipelines (stream processing)
	// - Use at-least-once + idempotent processing for Kafka-to-external
	// - Design ALL consumers to be idempotent regardless (defense in depth)
	// - Don't reach for EOS unless you truly need it — at-least-once
	//   with idempotent consumers is simpler and often sufficient

	fmt.Println("  EOS works WITHIN Kafka only (not for external side effects)")
	fmt.Println("  No cross-cluster EOS (MirrorMaker = at-least-once)")
	fmt.Println("  Practical: use at-least-once + idempotent consumers for most cases")
	fmt.Println("  Reserve EOS for Kafka-to-Kafka stream processing")
}
