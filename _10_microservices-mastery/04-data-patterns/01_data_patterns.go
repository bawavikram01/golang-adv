// =============================================================================
// LESSON 4: DATA PATTERNS — Saga, CQRS, Event Sourcing, Database-per-Service
// =============================================================================
//
// THE BIGGEST CHALLENGE IN MICROSERVICES: How do you manage data
// when each service owns its own database?
//
// In a monolith:  SELECT * FROM orders JOIN users JOIN payments → done.
// In microservices: that's IMPOSSIBLE. Each service = own database = own schema.
//
// This module covers every pattern for solving distributed data problems.
// =============================================================================

package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// PATTERN 1: Database-per-Service
// =============================================================================
//
// RULE: Each microservice OWNS its data. No other service can access
// another service's database directly. Communication only through APIs/events.
//
// WHY:
//   ✅ Independent deployment (schema changes don't break others)
//   ✅ Technology freedom (SQL for orders, MongoDB for catalog, Redis for sessions)
//   ✅ Encapsulation (service is the single source of truth)
//
// THE PROBLEM THIS CREATES:
//   ❌ No cross-service JOINs
//   ❌ No distributed transactions (2PC is slow and fragile)
//   ❌ Eventual consistency instead of strong consistency
//
// SOLUTIONS: Saga, CQRS, Event Sourcing, API Composition

// Simulated per-service databases
type UserDB struct {
	mu    sync.RWMutex
	users map[int64]User
}

type OrderDB struct {
	mu     sync.RWMutex
	orders map[int64]Order
}

type PaymentDB struct {
	mu       sync.RWMutex
	payments map[int64]Payment
}

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Order struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Products  []int64   `json:"products"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Payment struct {
	ID      int64   `json:"id"`
	OrderID int64   `json:"order_id"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
}

// =============================================================================
// PATTERN 2: SAGA Pattern — Distributed Transactions Without 2PC
// =============================================================================
//
// A Saga is a sequence of local transactions.
// Each step either succeeds → next step, or fails → compensating transactions.
//
// TWO IMPLEMENTATIONS:
//
// CHOREOGRAPHY-BASED SAGA:
//   Each service listens for events and acts.
//   Order Created → Payment processes → Inventory reserves → Shipping schedules
//   If payment fails → Order service refunds.
//   PROS: Loose coupling, simple for 2-3 steps.
//   CONS: Hard to debug, no central view of saga state.
//
// ORCHESTRATION-BASED SAGA:
//   A central orchestrator tells each service what to do.
//   SagaOrchestrator: "Payment service, charge $100" → "Inventory, reserve items"
//   If any step fails, orchestrator runs compensating transactions.
//   PROS: Easy to understand, centralized error handling.
//   CONS: Orchestrator is a single point, more coupling to orchestrator.
//
// RULE OF THUMB:
//   ≤3 services: choreography
//   >3 services or complex logic: orchestration

// --- Orchestration-Based Saga ---

type SagaStep struct {
	Name       string
	Execute    func() error
	Compensate func() error // rollback if later step fails
}

type SagaOrchestrator struct {
	steps     []SagaStep
	completed []int // indices of completed steps
}

func NewSagaOrchestrator() *SagaOrchestrator {
	return &SagaOrchestrator{}
}

func (s *SagaOrchestrator) AddStep(step SagaStep) {
	s.steps = append(s.steps, step)
}

func (s *SagaOrchestrator) Execute() error {
	fmt.Println("\n--- SAGA EXECUTION ---")

	for i, step := range s.steps {
		fmt.Printf("  [Step %d] Executing: %s\n", i+1, step.Name)

		if err := step.Execute(); err != nil {
			fmt.Printf("  ✗ Step %d failed: %v\n", i+1, err)
			fmt.Println("  → Running compensating transactions...")

			// Compensate in reverse order
			s.compensate()
			return fmt.Errorf("saga failed at step %s: %w", step.Name, err)
		}

		s.completed = append(s.completed, i)
		fmt.Printf("  ✓ Step %d completed: %s\n", i+1, step.Name)
	}

	fmt.Println("--- SAGA COMPLETED SUCCESSFULLY ---")
	return nil
}

func (s *SagaOrchestrator) compensate() {
	// Compensate in reverse order (last completed first)
	for i := len(s.completed) - 1; i >= 0; i-- {
		step := s.steps[s.completed[i]]
		fmt.Printf("  ↩ Compensating: %s\n", step.Name)
		if err := step.Compensate(); err != nil {
			// In production: log, alert, manual intervention
			fmt.Printf("  ⚠ Compensation failed for %s: %v (REQUIRES MANUAL FIX)\n", step.Name, err)
		}
	}
}

// =============================================================================
// PATTERN 3: CQRS — Command Query Responsibility Segregation
// =============================================================================
//
// CQRS = separate models for reading and writing.
//
// WRITE SIDE (Commands):
//   - Normalized, relational, consistent
//   - Validates business rules
//   - Publishes events after write
//
// READ SIDE (Queries):
//   - Denormalized, optimized for queries
//   - Projections built from events
//   - Eventually consistent with write side
//
// WHY CQRS:
//   ✅ Reads are 10-100x more frequent than writes → optimize separately
//   ✅ Read model can be a materialized view, search index, or cache
//   ✅ Write model can enforce complex business rules
//   ✅ Scale reads and writes independently
//
// WHEN TO USE:
//   ✅ Read and write patterns are very different
//   ✅ You need different storage for reads (Elasticsearch, Redis)
//   ✅ Complex domain with many business rules on writes
//
// WHEN NOT TO USE:
//   ❌ Simple CRUD app (overkill)
//   ❌ Strong consistency requirements everywhere

// --- Command Side ---
type OrderCommand interface {
	CommandName() string
}

type CreateOrderCommand struct {
	UserID   int64   `json:"user_id"`
	Products []int64 `json:"products"`
	Total    float64 `json:"total"`
}

func (c CreateOrderCommand) CommandName() string { return "CreateOrder" }

type CancelOrderCommand struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}

func (c CancelOrderCommand) CommandName() string { return "CancelOrder" }

// --- Command Handler (Write Side) ---
type OrderCommandHandler struct {
	db     *OrderDB
	events []DomainEvent // events generated by commands
	nextID int64
}

type DomainEvent struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

func NewOrderCommandHandler(db *OrderDB) *OrderCommandHandler {
	return &OrderCommandHandler{db: db, nextID: 1}
}

func (h *OrderCommandHandler) Handle(cmd OrderCommand) error {
	switch c := cmd.(type) {
	case CreateOrderCommand:
		return h.createOrder(c)
	case CancelOrderCommand:
		return h.cancelOrder(c)
	default:
		return fmt.Errorf("unknown command: %s", cmd.CommandName())
	}
}

func (h *OrderCommandHandler) createOrder(cmd CreateOrderCommand) error {
	h.db.mu.Lock()
	defer h.db.mu.Unlock()

	order := Order{
		ID:        h.nextID,
		UserID:    cmd.UserID,
		Products:  cmd.Products,
		Total:     cmd.Total,
		Status:    "created",
		CreatedAt: time.Now(),
	}
	h.db.orders[h.nextID] = order
	h.nextID++

	// Publish event for read side
	payload, _ := json.Marshal(order)
	h.events = append(h.events, DomainEvent{
		Type:      "OrderCreated",
		Payload:   payload,
		Timestamp: time.Now(),
	})

	fmt.Printf("[CQRS Write] Order %d created\n", order.ID)
	return nil
}

func (h *OrderCommandHandler) cancelOrder(cmd CancelOrderCommand) error {
	h.db.mu.Lock()
	defer h.db.mu.Unlock()

	order, ok := h.db.orders[cmd.OrderID]
	if !ok {
		return fmt.Errorf("order %d not found", cmd.OrderID)
	}
	order.Status = "cancelled"
	h.db.orders[cmd.OrderID] = order

	payload, _ := json.Marshal(map[string]interface{}{
		"order_id": cmd.OrderID,
		"reason":   cmd.Reason,
	})
	h.events = append(h.events, DomainEvent{
		Type:      "OrderCancelled",
		Payload:   payload,
		Timestamp: time.Now(),
	})

	fmt.Printf("[CQRS Write] Order %d cancelled: %s\n", cmd.OrderID, cmd.Reason)
	return nil
}

// --- Query Side (Read Model / Projection) ---
type OrderReadModel struct {
	mu     sync.RWMutex
	orders map[int64]OrderView // denormalized view optimized for queries
}

type OrderView struct {
	OrderID   int64   `json:"order_id"`
	UserName  string  `json:"user_name"` // denormalized from user service
	Products  []int64 `json:"products"`
	Total     float64 `json:"total"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"` // formatted for display
}

func NewOrderReadModel() *OrderReadModel {
	return &OrderReadModel{orders: make(map[int64]OrderView)}
}

func (rm *OrderReadModel) ProcessEvent(event DomainEvent) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	switch event.Type {
	case "OrderCreated":
		var order Order
		json.Unmarshal(event.Payload, &order)
		rm.orders[order.ID] = OrderView{
			OrderID:   order.ID,
			UserName:  fmt.Sprintf("User-%d", order.UserID), // in production: lookup from cache
			Products:  order.Products,
			Total:     order.Total,
			Status:    order.Status,
			CreatedAt: order.CreatedAt.Format("2006-01-02"),
		}
		fmt.Printf("[CQRS Read] Projection updated: order %d added\n", order.ID)

	case "OrderCancelled":
		var data map[string]interface{}
		json.Unmarshal(event.Payload, &data)
		orderID := int64(data["order_id"].(float64))
		if view, ok := rm.orders[orderID]; ok {
			view.Status = "cancelled"
			rm.orders[orderID] = view
			fmt.Printf("[CQRS Read] Projection updated: order %d cancelled\n", orderID)
		}
	}
}

func (rm *OrderReadModel) GetOrder(id int64) (OrderView, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	v, ok := rm.orders[id]
	return v, ok
}

func (rm *OrderReadModel) GetAllOrders() []OrderView {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	result := make([]OrderView, 0, len(rm.orders))
	for _, v := range rm.orders {
		result = append(result, v)
	}
	return result
}

// =============================================================================
// PATTERN 4: Event Sourcing
// =============================================================================
//
// Instead of storing CURRENT STATE, store ALL EVENTS that led to the state.
//
// Traditional DB:  Order { status: "shipped" }  — you lost history
// Event Sourcing:  [Created, Paid, PackageReady, Shipped] — full history
//
// To get current state: replay all events from the beginning.
//
// EVENT STORE:
//   Append-only log. Events are immutable. Never update, never delete.
//   Each entity (aggregate) has its own event stream.
//
// SNAPSHOTS:
//   If 10,000 events for one order → slow to replay.
//   Take periodic snapshots: store state at event #9000, replay from there.
//
// WHY EVENT SOURCING:
//   ✅ Complete audit trail (every change recorded)
//   ✅ Time travel (reconstruct state at any point in time)
//   ✅ Debug production issues (replay events to reproduce)
//   ✅ Build new read models from historical events
//
// WHEN NOT TO USE:
//   ❌ Simple CRUD (way overkill)
//   ❌ Need to delete data (GDPR — need crypto-shredding)
//   ❌ Small team that can't manage complexity

type EventStore struct {
	mu     sync.RWMutex
	events map[string][]StoredEvent // aggregate ID → events
}

type StoredEvent struct {
	AggregateID string          `json:"aggregate_id"`
	Version     int             `json:"version"` // ordering within aggregate
	Type        string          `json:"type"`
	Data        json.RawMessage `json:"data"`
	Timestamp   time.Time       `json:"timestamp"`
}

func NewEventStore() *EventStore {
	return &EventStore{events: make(map[string][]StoredEvent)}
}

func (es *EventStore) Append(aggregateID string, eventType string, data interface{}) StoredEvent {
	es.mu.Lock()
	defer es.mu.Unlock()

	payload, _ := json.Marshal(data)
	version := len(es.events[aggregateID]) + 1

	event := StoredEvent{
		AggregateID: aggregateID,
		Version:     version,
		Type:        eventType,
		Data:        payload,
		Timestamp:   time.Now(),
	}

	es.events[aggregateID] = append(es.events[aggregateID], event)
	return event
}

func (es *EventStore) GetEvents(aggregateID string) []StoredEvent {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.events[aggregateID]
}

// Reconstruct order state from events
type OrderAggregate struct {
	ID     string
	Status string
	Total  float64
	Items  []string
}

func ReplayOrderEvents(events []StoredEvent) OrderAggregate {
	var order OrderAggregate

	for _, event := range events {
		switch event.Type {
		case "OrderCreated":
			var data struct {
				ID    string   `json:"id"`
				Items []string `json:"items"`
				Total float64  `json:"total"`
			}
			json.Unmarshal(event.Data, &data)
			order.ID = data.ID
			order.Items = data.Items
			order.Total = data.Total
			order.Status = "created"

		case "OrderPaid":
			order.Status = "paid"

		case "OrderShipped":
			order.Status = "shipped"

		case "OrderCancelled":
			order.Status = "cancelled"
		}
	}
	return order
}

// =============================================================================
// PATTERN 5: API Composition — Cross-Service Queries
// =============================================================================
//
// Problem: Customer wants to see order + user + payment info on one page.
//          Data lives in 3 different services.
//
// Solution: API Composer (or Gateway) calls multiple services and merges.
//
// RISKS:
//   - Performance (multiple HTTP calls)
//   - Availability (if one service is down, whole query fails)
//   - Data consistency (service A and B might be out of sync)

type OrderDTO struct {
	OrderID      int64   `json:"order_id"`
	UserName     string  `json:"user_name"`
	UserEmail    string  `json:"user_email"`
	Products     []int64 `json:"products"`
	Total        float64 `json:"total"`
	PaymentState string  `json:"payment_status"`
}

type APIComposer struct {
	userDB    *UserDB
	orderDB   *OrderDB
	paymentDB *PaymentDB
}

func (c *APIComposer) GetOrderDetails(orderID int64) (*OrderDTO, error) {
	c.orderDB.mu.RLock()
	order, ok := c.orderDB.orders[orderID]
	c.orderDB.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("order not found")
	}

	c.userDB.mu.RLock()
	user := c.userDB.users[order.UserID]
	c.userDB.mu.RUnlock()

	c.paymentDB.mu.RLock()
	payment := c.paymentDB.payments[orderID]
	c.paymentDB.mu.RUnlock()

	return &OrderDTO{
		OrderID:      order.ID,
		UserName:     user.Name,
		UserEmail:    user.Email,
		Products:     order.Products,
		Total:        order.Total,
		PaymentState: payment.Status,
	}, nil
}

func main() {
	// =========================================================================
	// DEMO 1: Saga Pattern (Orchestration)
	// =========================================================================
	fmt.Println("=== SAGA PATTERN (Orchestration) ===")

	paymentDone := false
	inventoryReserved := false

	saga := NewSagaOrchestrator()

	saga.AddStep(SagaStep{
		Name: "Validate Order",
		Execute: func() error {
			fmt.Println("    Validating order data...")
			return nil // success
		},
		Compensate: func() error {
			fmt.Println("    Nothing to compensate for validation")
			return nil
		},
	})

	saga.AddStep(SagaStep{
		Name: "Process Payment",
		Execute: func() error {
			fmt.Println("    Charging $199.99...")
			paymentDone = true
			return nil
		},
		Compensate: func() error {
			fmt.Println("    Refunding $199.99...")
			paymentDone = false
			return nil
		},
	})

	saga.AddStep(SagaStep{
		Name: "Reserve Inventory",
		Execute: func() error {
			inventoryReserved = true
			fmt.Println("    Reserving 2 items from warehouse...")
			return nil
		},
		Compensate: func() error {
			fmt.Println("    Releasing reserved inventory...")
			inventoryReserved = false
			return nil
		},
	})

	saga.AddStep(SagaStep{
		Name: "Schedule Shipping",
		Execute: func() error {
			// Simulate failure!
			return fmt.Errorf("no shipping carriers available")
		},
		Compensate: func() error {
			fmt.Println("    Cancelling shipment...")
			return nil
		},
	})

	err := saga.Execute()
	fmt.Printf("Saga result: %v\n", err)
	fmt.Printf("Payment reversed: %v, Inventory released: %v\n", !paymentDone, !inventoryReserved)

	// =========================================================================
	// DEMO 2: CQRS
	// =========================================================================
	fmt.Println("\n=== CQRS PATTERN ===")

	orderDB := &OrderDB{orders: make(map[int64]Order)}
	cmdHandler := NewOrderCommandHandler(orderDB)
	readModel := NewOrderReadModel()

	// Write side: execute commands
	cmdHandler.Handle(CreateOrderCommand{UserID: 1, Products: []int64{10, 20}, Total: 59.99})
	cmdHandler.Handle(CreateOrderCommand{UserID: 2, Products: []int64{30}, Total: 129.99})
	cmdHandler.Handle(CancelOrderCommand{OrderID: 1, Reason: "changed mind"})

	// Sync events to read model (in production: via event bus)
	for _, event := range cmdHandler.events {
		readModel.ProcessEvent(event)
	}

	// Query side: read from projection
	fmt.Println("\n[CQRS Query] All orders:")
	for _, view := range readModel.GetAllOrders() {
		fmt.Printf("  Order %d: %s — $%.2f — Status: %s\n",
			view.OrderID, view.UserName, view.Total, view.Status)
	}

	// =========================================================================
	// DEMO 3: Event Sourcing
	// =========================================================================
	fmt.Println("\n=== EVENT SOURCING ===")

	store := NewEventStore()

	// Append events (not storing state — storing history)
	store.Append("order-42", "OrderCreated", map[string]interface{}{
		"id": "order-42", "items": []string{"laptop", "mouse"}, "total": 1299.99,
	})
	store.Append("order-42", "OrderPaid", map[string]interface{}{
		"method": "credit_card",
	})
	store.Append("order-42", "OrderShipped", map[string]interface{}{
		"carrier": "FedEx", "tracking": "FX123456",
	})

	// Reconstruct state from events
	events := store.GetEvents("order-42")
	fmt.Printf("Event history for order-42 (%d events):\n", len(events))
	for _, e := range events {
		fmt.Printf("  v%d: %s at %s\n", e.Version, e.Type, e.Timestamp.Format("15:04:05"))
	}

	order := ReplayOrderEvents(events)
	fmt.Printf("Current state: ID=%s, Status=%s, Total=$%.2f, Items=%v\n",
		order.ID, order.Status, order.Total, order.Items)

	// Time travel: replay only first 2 events
	partialOrder := ReplayOrderEvents(events[:2])
	fmt.Printf("State at v2:   ID=%s, Status=%s (before shipping)\n",
		partialOrder.ID, partialOrder.Status)

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println("\n=== DATA PATTERN DECISION GUIDE ===")
	fmt.Println("┌────────────────────────────────────┬──────────────────────────────┐")
	fmt.Println("│ Problem                            │ Pattern                      │")
	fmt.Println("├────────────────────────────────────┼──────────────────────────────┤")
	fmt.Println("│ Cross-service transactions         │ Saga (orchestration/choreo)  │")
	fmt.Println("│ Read/write performance asymmetry   │ CQRS                         │")
	fmt.Println("│ Need full audit trail              │ Event Sourcing               │")
	fmt.Println("│ Cross-service queries              │ API Composition              │")
	fmt.Println("│ Data isolation                     │ Database-per-Service         │")
	fmt.Println("│ Eventual consistency + events      │ Event Sourcing + CQRS combo  │")
	fmt.Println("└────────────────────────────────────┴──────────────────────────────┘")
}
