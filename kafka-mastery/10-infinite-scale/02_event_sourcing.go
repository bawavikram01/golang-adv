//go:build ignore
// =============================================================================
// LESSON 10.2: EVENT SOURCING & CQRS — Kafka as the Source of Truth
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Event sourcing: storing events instead of current state
// - CQRS: separating reads and writes for independent scaling
// - Saga pattern: distributed transactions without 2PC
// - Outbox pattern: reliable event publishing from databases
// - Practical event sourcing pitfalls
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== EVENT SOURCING & CQRS ===")
	fmt.Println()

	eventSourcing()
	cqrs()
	sagaPattern()
	outboxPattern()
	practicalPitfalls()
}

func eventSourcing() {
	fmt.Println("--- EVENT SOURCING ---")

	// TRADITIONAL: Store CURRENT STATE
	//   User table: {id: 1, name: "Alice", balance: 500}
	//   After deposit: {id: 1, name: "Alice", balance: 700}
	//   History is LOST. You only know the current balance.
	//
	// EVENT SOURCING: Store EVENTS
	//   Event log:
	//   1. UserCreated(id=1, name="Alice")
	//   2. BalanceDeposited(id=1, amount=1000)
	//   3. BalanceWithdrawn(id=1, amount=500)
	//   4. BalanceDeposited(id=1, amount=200)
	//
	//   Current state = replay all events: balance = 0 + 1000 - 500 + 200 = 700
	//
	// KAFKA AS EVENT STORE:
	// ─────────────────────
	// Topic: user-events (compacted for snapshotting, or regular for full history)
	// Key: user-id
	// Value: event (BalanceDeposited, BalanceWithdrawn, etc.)
	//
	// WHY KAFKA IS A GOOD FIT:
	// - Immutable append-only log = perfect for events
	// - Ordering per key = events for same entity in order
	// - Replay capability = rebuild state from any point
	// - Multiple consumers = multiple read models from same events
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  EVENT SOURCING WITH KAFKA:                                    │
	// │                                                              │
	// │  Command ──► Aggregate ──► Event ──► Kafka Topic              │
	// │                                          │                   │
	// │                                          ├──► Read Model DB  │
	// │                                          ├──► Search Index   │
	// │                                          ├──► Analytics      │
	// │                                          └──► Notifications  │
	// │                                                              │
	// │  Each consumer builds its OWN view from the SAME events.     │
	// │  Add a new view? Just add a new consumer from offset 0.      │
	// └──────────────────────────────────────────────────────────────┘
	//
	// SNAPSHOTTING:
	// ─────────────
	// Replaying 1M events per user on startup is slow.
	// Every N events: produce a SNAPSHOT record (full current state).
	// On startup: read latest snapshot → replay events after snapshot.
	// Use a separate compacted topic for snapshots.

	fmt.Println("  Store events, not state. Replay events to derive current state.")
	fmt.Println("  Kafka: immutable, ordered, replayable = natural event store")
	fmt.Println("  Snapshots every N events for fast startup")
	fmt.Println()
}

func cqrs() {
	fmt.Println("--- CQRS ---")

	// CQRS = Command Query Responsibility Segregation
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │                                                              │
	// │  WRITE SIDE:                    READ SIDE:                   │
	// │  ──────────                     ─────────                    │
	// │  Commands ──► Aggregate         Queries ──► Read Model       │
	// │               │                             ▲                │
	// │               ▼                             │                │
	// │           Events ──► Kafka ──► Projector ──► DB              │
	// │                                                              │
	// │  Write model: optimized for consistency and business rules   │
	// │  Read model: optimized for queries (denormalized, indexed)   │
	// │  They're COMPLETELY SEPARATE databases!                       │
	// │                                                              │
	// │  BENEFITS:                                                    │
	// │  - Scale reads and writes independently                      │
	// │  - Multiple read models from same events                     │
	// │  - Read model: PostgreSQL, Elasticsearch, Redis, whatever   │
	// │  - Each optimized for its specific query patterns            │
	// │                                                              │
	// │  COSTS:                                                      │
	// │  - Eventual consistency (read model lags behind write)       │
	// │  - More infrastructure (multiple databases)                  │
	// │  - Complexity of maintaining projectors                      │
	// └──────────────────────────────────────────────────────────────┘

	fmt.Println("  Separate write (events) from read (materialized views)")
	fmt.Println("  Scale independently, multiple read models from same events")
	fmt.Println("  Cost: eventual consistency + operational complexity")
	fmt.Println()
}

func sagaPattern() {
	fmt.Println("--- SAGA PATTERN ---")

	// Distributed transactions across services WITHOUT 2PC:
	//
	// CHOREOGRAPHY SAGA (event-driven):
	// ──────────────────────────────────
	// Each service reacts to events and publishes new events.
	//
	// 1. Order Service → publishes OrderCreated
	// 2. Payment Service → sees OrderCreated → charges card → publishes PaymentCompleted
	// 3. Inventory Service → sees PaymentCompleted → reserves stock → publishes StockReserved
	// 4. Shipping Service → sees StockReserved → ships order → publishes OrderShipped
	//
	// ON FAILURE (compensating actions):
	// 3b. Inventory Service → can't reserve → publishes StockFailed
	// 2b. Payment Service → sees StockFailed → refunds → publishes PaymentRefunded
	// 1b. Order Service → sees PaymentRefunded → marks order failed
	//
	// ORCHESTRATION SAGA (coordinator):
	// ──────────────────────────────────
	// A central "saga coordinator" tells services what to do:
	//
	// Saga Coordinator:
	// 1. Send "charge card" command → Payment Service
	// 2. On success: Send "reserve stock" command → Inventory Service
	// 3. On success: Send "ship order" command → Shipping Service
	// 4. On ANY failure: Send compensating commands in reverse
	//
	// KAFKA IMPLEMENTATION:
	// ─────────────────────
	// Commands and events flow through Kafka topics:
	//   topic: order-commands   (orchestrator → services)
	//   topic: order-events     (services → orchestrator)
	//   topic: order-saga-state (saga state changelog, compacted)
	//
	// The saga coordinator is a Kafka consumer/producer + state store.

	fmt.Println("  Choreography: services react to events (decoupled, harder to trace)")
	fmt.Println("  Orchestration: coordinator drives the flow (centralized, easier to trace)")
	fmt.Println("  Compensating actions instead of rollback")
	fmt.Println()
}

func outboxPattern() {
	fmt.Println("--- OUTBOX PATTERN ---")

	// PROBLEM: How to atomically update a database AND publish to Kafka?
	//
	// Naive approach:
	//   1. Write to database
	//   2. Publish to Kafka
	//   If step 2 fails: DB updated, event not published. INCONSISTENT!
	//
	// OUTBOX PATTERN:
	// ───────────────
	// 1. In the SAME database transaction:
	//    a. Write business data to business table
	//    b. Write event to OUTBOX table
	// 2. A separate process (CDC or poller) reads outbox table → publishes to Kafka
	// 3. Mark outbox entry as published (or CDC handles this automatically)
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  Service ──► BEGIN TX                                         │
	// │              INSERT INTO orders (...)                         │
	// │              INSERT INTO outbox (event_type, payload, ...)   │
	// │              COMMIT TX                                        │
	// │                                                              │
	// │  CDC (Debezium) ──► watches outbox table ──► Kafka topic     │
	// │  OR                                                          │
	// │  Poller ──► SELECT * FROM outbox WHERE published=false       │
	// │         ──► kafka.produce(events)                             │
	// │         ──► UPDATE outbox SET published=true                  │
	// └──────────────────────────────────────────────────────────────┘
	//
	// DEBEZIUM + OUTBOX: the gold standard for database → Kafka.
	// Debezium captures changes from DB's WAL → publishes to Kafka.
	// No polling, no missed events, near-realtime.

	fmt.Println("  Outbox: write event to DB table in same transaction as business data")
	fmt.Println("  CDC (Debezium) reads DB WAL → publishes to Kafka automatically")
	fmt.Println("  Guarantees: atomic DB write + event publish")
	fmt.Println()
}

func practicalPitfalls() {
	fmt.Println("--- PRACTICAL PITFALLS ---")

	// PITFALL 1: UNBOUNDED EVENT REPLAY
	// Entity has 10 years of events → replaying takes hours.
	// Fix: regular snapshots, or use compacted topics for latest state.
	//
	// PITFALL 2: SCHEMA EVOLUTION IN EVENT STORES
	// Old events have old schemas. New code must handle ALL versions.
	// Fix: upcasting old events to new schema during replay.
	//
	// PITFALL 3: EVENT ORDERING IN SAGAS
	// Events from different services arrive in different orders.
	// Fix: saga coordinator handles idempotency and out-of-order events.
	//
	// PITFALL 4: "EVENT STORM" ON NEW CONSUMER
	// New consumer reads from offset 0 → processes years of events.
	// Fix: bootstrap from snapshot/backup, then consume from recent offset.
	//
	// PITFALL 5: GDPR AND EVENT SOURCING
	// "Right to be forgotten" is hard when events are immutable.
	// Fix: crypto-shredding (encrypt per-user, delete key to "erase").

	fmt.Println("  Pitfalls: unbounded replay, schema evolution, GDPR")
	fmt.Println("  Solutions: snapshots, upcasting, crypto-shredding")
}
