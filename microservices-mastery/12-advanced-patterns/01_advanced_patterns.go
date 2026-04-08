// =============================================================================
// LESSON 12: ADVANCED PATTERNS — God-Level Microservices Architecture
// =============================================================================
//
// These are the patterns that separate senior engineers from the rest.
// Each pattern solves a SPECIFIC hard problem in distributed systems.
//
// THIS MODULE COVERS:
//   1. Strangler Fig Pattern    (migrating from monolith to microservices)
//   2. Sidecar Pattern          (attach functionality without changing service)
//   3. Ambassador Pattern       (proxy to external services)
//   4. Anti-Corruption Layer    (protect your domain from external systems)
//   5. Outbox Pattern           (reliable event publishing with DB transactions)
//   6. Change Data Capture      (stream DB changes as events)
//   7. Transactional Outbox     (exactly-once event delivery)
//   8. Backends for Frontends   (covered in 07, summarized here)
//   9. Bulkhead + Sidecar Combo
//   10. Distributed Locking
// =============================================================================

package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// PATTERN 1: Strangler Fig Pattern
// =============================================================================
//
// Named after the strangler fig tree that grows around a host tree and
// eventually replaces it.
//
// PROBLEM: You have a monolith. You want microservices. But you can't
//          rewrite everything at once (Big Bang rewrite = suicide).
//
// SOLUTION: Incrementally build microservices around the monolith.
//   1. New features → new microservice
//   2. Existing features → extract one by one
//   3. Route traffic: new path → microservice, old path → monolith
//   4. Eventually, monolith shrinks to nothing
//
// THE ROUTER:
//   API Gateway or reverse proxy decides: monolith or microservice?
//   Route by path, feature flag, or percentage.

type StranglerRouter struct {
	mu            sync.RWMutex
	migratedPaths map[string]string // path → microservice
	monolithURL   string
}

func NewStranglerRouter(monolithURL string) *StranglerRouter {
	return &StranglerRouter{
		migratedPaths: make(map[string]string),
		monolithURL:   monolithURL,
	}
}

func (r *StranglerRouter) MigratePath(path, microserviceURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.migratedPaths[path] = microserviceURL
	fmt.Printf("  [Strangler] Migrated %s → %s\n", path, microserviceURL)
}

func (r *StranglerRouter) Route(path string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if path is migrated to a microservice
	for prefix, msURL := range r.migratedPaths {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return msURL
		}
	}
	return r.monolithURL
}

func (r *StranglerRouter) MigrationProgress() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// In production: calculate based on traffic percentage
	return float64(len(r.migratedPaths)) * 10.0 // simplified
}

// =============================================================================
// PATTERN 2: Sidecar Pattern
// =============================================================================
//
// A SIDECAR is a helper container deployed alongside your main service.
// Same pod (Kubernetes), shares network namespace and storage.
//
// The sidecar handles cross-cutting concerns WITHOUT modifying the service:
//   ✅ Logging agent (Fluentd, Vector) — collects logs, ships to ELK
//   ✅ Proxy (Envoy) — handles mTLS, routing, retry
//   ✅ Config watcher — reloads config when ConfigMap changes
//   ✅ Secrets injector (Vault Agent) — injects secrets as files
//   ✅ Monitoring agent (Prometheus exporter)
//
// WHY SIDECAR (not library):
//   ✅ Language-agnostic (Go service + Python sidecar = fine)
//   ✅ Independent lifecycle (update sidecar without redeploying service)
//   ✅ Separation of concerns (service developer doesn't handle infra)

type SidecarConfig struct {
	Name     string
	Image    string
	Purpose  string
	Ports    []int
	Resource string
}

var commonSidecars = []SidecarConfig{
	{
		Name: "envoy-proxy", Image: "envoyproxy/envoy:v1.28",
		Purpose: "mTLS, load balancing, circuit breaker",
		Ports:   []int{15001, 15006}, Resource: "CPU: 100m, Mem: 128Mi",
	},
	{
		Name: "fluentd", Image: "fluent/fluentd:v1.16",
		Purpose: "Log collection and shipping to ELK/Loki",
		Ports:   []int{24224}, Resource: "CPU: 50m, Mem: 64Mi",
	},
	{
		Name: "vault-agent", Image: "hashicorp/vault:1.15",
		Purpose: "Auto-inject secrets from Vault as files",
		Ports:   []int{}, Resource: "CPU: 25m, Mem: 32Mi",
	},
	{
		Name: "config-reloader", Image: "jimmidyson/configmap-reload:v0.9",
		Purpose: "Watch ConfigMap changes, trigger reload",
		Ports:   []int{}, Resource: "CPU: 10m, Mem: 16Mi",
	},
}

// =============================================================================
// PATTERN 3: Ambassador Pattern
// =============================================================================
//
// An AMBASSADOR is an outbound proxy that handles communication to EXTERNAL
// services (3rd party APIs, legacy systems, external databases).
//
// Unlike sidecar (general helper), ambassador specifically handles:
//   - Connection pooling to external service
//   - Retry with backoff
//   - Circuit breaking
//   - Protocol translation (e.g., your service speaks HTTP, external speaks SOAP)
//   - Auth token management (refresh tokens for 3rd party APIs)
//
// Your service calls the ambassador on localhost.
// The ambassador handles the messy external communication.

type Ambassador struct {
	name        string
	targetURL   string
	retries     int
	timeout     time.Duration
	circuitOpen bool
}

func NewAmbassador(name, targetURL string) *Ambassador {
	return &Ambassador{
		name:      name,
		targetURL: targetURL,
		retries:   3,
		timeout:   5 * time.Second,
	}
}

func (a *Ambassador) Call(method, path string) string {
	// In production: retry with backoff, circuit breaker, auth token refresh
	return fmt.Sprintf("[Ambassador:%s] %s %s%s (retries=%d, timeout=%v)",
		a.name, method, a.targetURL, path, a.retries, a.timeout)
}

// =============================================================================
// PATTERN 4: Anti-Corruption Layer (ACL)
// =============================================================================
//
// PROBLEM: You integrate with a legacy system or 3rd party API.
//          Their data model is DIFFERENT from yours. Their changes
//          break your code. Their weirdness leaks into your domain.
//
// SOLUTION: An Anti-Corruption Layer that TRANSLATES between their
//          domain model and yours. Your service never touches their types.
//
// YOUR SERVICE → ACL (translator) → External System
//
// The ACL:
//   - Converts their types to your types
//   - Handles their quirks (date formats, null handling, nested IDs)
//   - Isolates you from their changes (only ACL needs updating)

type LegacyOrder struct {
	// Legacy system's format (ugly, inconsistent)
	OrderNum   string `json:"ORDER_NUM"`
	CustID     string `json:"CUST_ID"`
	TotalAmt   string `json:"TOTAL_AMT"`       // string instead of float, wtf
	OrderDate  string `json:"ORD_DT"`          // "20240115" format
	StatusCode int    `json:"STAT_CD"`         // 1=new, 2=paid, 3=shipped
	LineItems  string `json:"LINE_ITEMS_JSON"` // JSON string inside JSON
}

type CleanOrder struct {
	// Your clean domain model
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id"`
	Total      float64     `json:"total"`
	OrderedAt  time.Time   `json:"ordered_at"`
	Status     string      `json:"status"`
	Items      []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Anti-Corruption Layer translates legacy → clean domain
type OrderACL struct{}

func (acl *OrderACL) TranslateOrder(legacy LegacyOrder) (*CleanOrder, error) {
	// Parse amount
	var total float64
	fmt.Sscanf(legacy.TotalAmt, "%f", &total)

	// Parse date
	orderedAt, _ := time.Parse("20060102", legacy.OrderDate)

	// Map status codes
	statusMap := map[int]string{1: "pending", 2: "paid", 3: "shipped", 4: "delivered"}
	status := statusMap[legacy.StatusCode]
	if status == "" {
		status = "unknown"
	}

	// Parse nested JSON items
	var items []OrderItem
	if legacy.LineItems != "" {
		json.Unmarshal([]byte(legacy.LineItems), &items)
	}

	return &CleanOrder{
		ID:         legacy.OrderNum,
		CustomerID: legacy.CustID,
		Total:      total,
		OrderedAt:  orderedAt,
		Status:     status,
		Items:      items,
	}, nil
}

// =============================================================================
// PATTERN 5: Outbox Pattern (Transactional Outbox)
// =============================================================================
//
// THE PROBLEM: You need to save data to DB AND publish an event.
//   1. Save order to DB     ✓
//   2. Publish OrderCreated  ← What if this fails? DB has order, but event is lost.
//
// Or:
//   1. Publish OrderCreated  ✓
//   2. Save order to DB      ← What if this fails? Event sent, but no order in DB.
//
// You CANNOT do a distributed transaction across DB + message broker (2PC is bad).
//
// SOLUTION: OUTBOX PATTERN
//   1. In the SAME DB TRANSACTION:
//      a. INSERT order INTO orders
//      b. INSERT event INTO outbox_table
//   2. A separate process (CDC or poller) reads outbox_table → publishes to Kafka
//   3. Mark outbox event as published
//
// This ensures atomicity: either both order + event are saved, or neither.
//
// OUTBOX TABLE SCHEMA:
//   id         UUID PRIMARY KEY
//   event_type VARCHAR     ("OrderCreated")
//   payload    JSONB       (event data)
//   created_at TIMESTAMP
//   published  BOOLEAN     (false → unprocessed)

type OutboxEntry struct {
	ID        string          `json:"id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
	Published bool            `json:"published"`
}

type OutboxStore struct {
	mu      sync.Mutex
	entries []OutboxEntry
}

func NewOutboxStore() *OutboxStore {
	return &OutboxStore{}
}

// SaveOrderWithEvent — simulates a single DB transaction
func (s *OutboxStore) SaveOrderWithEvent(order CleanOrder, eventType string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// In production: BEGIN TRANSACTION
	// INSERT INTO orders ...
	// INSERT INTO outbox ...
	// COMMIT

	payload, _ := json.Marshal(order)
	entry := OutboxEntry{
		ID:        fmt.Sprintf("outbox_%d", time.Now().UnixNano()),
		EventType: eventType,
		Payload:   payload,
		CreatedAt: time.Now(),
		Published: false,
	}
	s.entries = append(s.entries, entry)
	fmt.Printf("  [Outbox] Saved order %s + event %s in single transaction\n",
		order.ID, eventType)
}

// PollAndPublish — outbox relay (runs periodically)
func (s *OutboxStore) PollAndPublish() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, entry := range s.entries {
		if !entry.Published {
			// In production: publish to Kafka/RabbitMQ
			fmt.Printf("  [Outbox Relay] Publishing: %s (type: %s)\n", entry.ID, entry.EventType)
			s.entries[i].Published = true
		}
	}
}

// =============================================================================
// PATTERN 6: Change Data Capture (CDC)
// =============================================================================
//
// CDC = capture changes from the database transaction log (WAL)
// and stream them as events.
//
// Instead of your app publishing events manually:
//   App → saves to DB → Debezium reads WAL → publishes to Kafka
//
// WHY CDC:
//   ✅ Your app doesn't need to know about events (zero code changes)
//   ✅ Captures ALL changes (even from legacy apps that don't publish events)
//   ✅ Guaranteed delivery (based on DB transaction log)
//   ✅ Works with Outbox pattern (Debezium reads outbox table from WAL)
//
// TOOLS: Debezium (most popular), Maxwell, DynamoDB Streams, Mongo Change Streams
//
// FLOW:
//   App → Postgres (WAL) → Debezium → Kafka → Consumers

// =============================================================================
// PATTERN 7: Distributed Locking
// =============================================================================
//
// PROBLEM: Two instances of order-service both try to process the same order.
//          Double charge! Double shipping!
//
// SOLUTION: Distributed lock. Only one instance can hold the lock.
//
// IMPLEMENTATIONS:
//   Redis:      SETNX + TTL (Redlock algorithm for HA)
//   Zookeeper:  Ephemeral sequential nodes
//   etcd:       Lease-based locks
//   Database:   SELECT ... FOR UPDATE (simple but slow)
//
// RULES:
//   ✅ Always set a TTL (lock expires if holder crashes)
//   ✅ Use fencing tokens (monotonic ID to prevent stale lock holders)
//   ✅ Prefer idempotency over locking when possible

type DistributedLock struct {
	mu    sync.Mutex
	locks map[string]lockEntry
}

type lockEntry struct {
	holder    string
	expiresAt time.Time
	fenceID   int64 // monotonically increasing
}

func NewDistributedLock() *DistributedLock {
	return &DistributedLock{locks: make(map[string]lockEntry)}
}

func (dl *DistributedLock) Acquire(resource, holder string, ttl time.Duration) (int64, bool) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// Check if lock exists and is still valid
	if existing, ok := dl.locks[resource]; ok {
		if time.Now().Before(existing.expiresAt) {
			return 0, false // locked by someone else
		}
	}

	fenceID := time.Now().UnixNano()
	dl.locks[resource] = lockEntry{
		holder:    holder,
		expiresAt: time.Now().Add(ttl),
		fenceID:   fenceID,
	}

	return fenceID, true
}

func (dl *DistributedLock) Release(resource, holder string) bool {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	existing, ok := dl.locks[resource]
	if !ok || existing.holder != holder {
		return false // not your lock
	}

	delete(dl.locks, resource)
	return true
}

func main() {
	// =========================================================================
	// DEMO 1: Strangler Fig Pattern
	// =========================================================================
	fmt.Println("=== STRANGLER FIG PATTERN ===")

	router := NewStranglerRouter("http://monolith:8080")

	// Phase 1: Migrate user service
	router.MigratePath("/api/users", "http://user-service:8080")

	// Phase 2: Migrate order service
	router.MigratePath("/api/orders", "http://order-service:8080")

	// Test routing
	paths := []string{"/api/users/123", "/api/orders/456", "/api/legacy/reports"}
	for _, path := range paths {
		target := router.Route(path)
		fmt.Printf("  %s → %s\n", path, target)
	}
	fmt.Printf("  Migration progress: %.0f%%\n", router.MigrationProgress())

	// =========================================================================
	// DEMO 2: Sidecar Pattern
	// =========================================================================
	fmt.Println("\n=== SIDECAR PATTERN ===")

	for _, sc := range commonSidecars {
		fmt.Printf("  %s (%s)\n    Purpose: %s\n    Resource: %s\n",
			sc.Name, sc.Image, sc.Purpose, sc.Resource)
	}

	// =========================================================================
	// DEMO 3: Ambassador Pattern
	// =========================================================================
	fmt.Println("\n=== AMBASSADOR PATTERN ===")

	stripe := NewAmbassador("stripe", "https://api.stripe.com")
	twilio := NewAmbassador("twilio", "https://api.twilio.com")

	fmt.Printf("  %s\n", stripe.Call("POST", "/v1/charges"))
	fmt.Printf("  %s\n", twilio.Call("POST", "/2010/Messages.json"))

	// =========================================================================
	// DEMO 4: Anti-Corruption Layer
	// =========================================================================
	fmt.Println("\n=== ANTI-CORRUPTION LAYER ===")

	legacy := LegacyOrder{
		OrderNum:   "ORD-2024-001",
		CustID:     "C-12345",
		TotalAmt:   "199.99",
		OrderDate:  "20240115",
		StatusCode: 2,
		LineItems:  `[{"product_id":"P1","quantity":2,"price":99.99}]`,
	}

	acl := &OrderACL{}
	clean, _ := acl.TranslateOrder(legacy)

	fmt.Printf("  Legacy: ORDER_NUM=%s, TOTAL_AMT=%s, STAT_CD=%d\n",
		legacy.OrderNum, legacy.TotalAmt, legacy.StatusCode)
	fmt.Printf("  Clean:  ID=%s, Total=%.2f, Status=%s, Items=%d\n",
		clean.ID, clean.Total, clean.Status, len(clean.Items))

	// =========================================================================
	// DEMO 5: Outbox Pattern
	// =========================================================================
	fmt.Println("\n=== OUTBOX PATTERN ===")

	outbox := NewOutboxStore()

	// Save order + event atomically
	outbox.SaveOrderWithEvent(*clean, "OrderCreated")
	outbox.SaveOrderWithEvent(CleanOrder{
		ID: "ORD-002", CustomerID: "C-456", Total: 49.99, Status: "pending",
	}, "OrderCreated")

	// Outbox relay polls and publishes
	fmt.Println("  [Relay running...]")
	outbox.PollAndPublish()

	// =========================================================================
	// DEMO 6: Distributed Locking
	// =========================================================================
	fmt.Println("\n=== DISTRIBUTED LOCKING ===")

	dl := NewDistributedLock()

	// Instance 1 acquires lock
	fenceID, ok := dl.Acquire("order-123", "instance-1", 10*time.Second)
	fmt.Printf("  instance-1 acquire order-123: %v (fence: %d)\n", ok, fenceID)

	// Instance 2 tries to acquire same lock
	_, ok = dl.Acquire("order-123", "instance-2", 10*time.Second)
	fmt.Printf("  instance-2 acquire order-123: %v (blocked!)\n", ok)

	// Instance 1 releases
	dl.Release("order-123", "instance-1")
	fmt.Println("  instance-1 released order-123")

	// Now instance 2 can acquire
	fenceID2, ok := dl.Acquire("order-123", "instance-2", 10*time.Second)
	fmt.Printf("  instance-2 acquire order-123: %v (fence: %d)\n", ok, fenceID2)

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println("\n=== ADVANCED PATTERNS CHEAT SHEET ===")
	fmt.Println("┌──────────────────────────────┬──────────────────────────────────────┐")
	fmt.Println("│ Pattern                      │ Solves                               │")
	fmt.Println("├──────────────────────────────┼──────────────────────────────────────┤")
	fmt.Println("│ Strangler Fig                │ Monolith → microservices migration   │")
	fmt.Println("│ Sidecar                      │ Cross-cutting concerns (logs, proxy) │")
	fmt.Println("│ Ambassador                   │ External service communication       │")
	fmt.Println("│ Anti-Corruption Layer         │ Legacy system integration            │")
	fmt.Println("│ Outbox                       │ Reliable event publishing            │")
	fmt.Println("│ CDC (Change Data Capture)    │ DB changes → events (zero code)      │")
	fmt.Println("│ Distributed Lock             │ Prevent concurrent processing        │")
	fmt.Println("│ Saga                         │ Distributed transactions (see 04)    │")
	fmt.Println("│ CQRS                         │ Read/write optimization (see 04)     │")
	fmt.Println("│ Event Sourcing               │ Full audit trail (see 04)            │")
	fmt.Println("└──────────────────────────────┴──────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("GOD-LEVEL ARCHITECTURE:")
	fmt.Println("  Event-Driven + CQRS + Event Sourcing + Saga + Outbox + CDC")
	fmt.Println("  + Service Mesh (Istio) + GitOps (ArgoCD) + Observability (OTel)")
	fmt.Println("  = The ultimate microservices stack 🏆")
}
