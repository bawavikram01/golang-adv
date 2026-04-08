// =============================================================================
// LESSON 3: ASYNCHRONOUS COMMUNICATION — Events, Message Queues, Pub/Sub
// =============================================================================
//
// Async = sender does NOT wait for a response.
// This is the backbone of scalable microservices architectures.
//
// THREE MESSAGING PATTERNS:
//   1. Point-to-Point (Queue)  — one sender, one receiver
//   2. Pub/Sub (Topics)        — one sender, many receivers
//   3. Event Streaming          — ordered, replayable log (Kafka-style)
//
// WHEN TO USE ASYNC:
//   ✅ Fire-and-forget (send email, push notification)
//   ✅ Fan-out to many services (order created → inventory + email + analytics)
//   ✅ Long-running operations (video processing, report generation)
//   ✅ Decoupling services (producer doesn't know consumers)
//   ✅ Load leveling (smooth out traffic spikes with a queue buffer)
// =============================================================================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// CORE CONCEPT: Events vs Commands vs Queries
// =============================================================================
//
// EVENT:   "Something happened" (past tense) — OrderCreated, PaymentCompleted
//          Producer doesn't care who listens. Adding listeners doesn't change producer.
//          Events are FACTS. They are immutable.
//
// COMMAND: "Do something" (imperative) — CreateOrder, ProcessPayment
//          Directed at a specific service. Expects it to be handled.
//
// QUERY:   "Tell me something" — GetUser, ListOrders
//          Always synchronous (REST/gRPC).
//
// RULE: Events for decoupling, Commands for orchestration, Queries for reads.

// =============================================================================
// PART 1: Event Types and Envelope
// =============================================================================

type Event struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Source    string            `json:"source"`
	Timestamp time.Time         `json:"timestamp"`
	Data      json.RawMessage   `json:"data"`
	Metadata  map[string]string `json:"metadata"`
}

type OrderCreatedEvent struct {
	OrderID    int64   `json:"order_id"`
	UserID     int64   `json:"user_id"`
	Products   []int64 `json:"products"`
	TotalPrice float64 `json:"total_price"`
}

type PaymentCompletedEvent struct {
	PaymentID int64   `json:"payment_id"`
	OrderID   int64   `json:"order_id"`
	Amount    float64 `json:"amount"`
	Method    string  `json:"method"`
}

type OrderShippedEvent struct {
	OrderID    int64  `json:"order_id"`
	TrackingNo string `json:"tracking_no"`
	Carrier    string `json:"carrier"`
}

func NewEvent(eventType, source string, data interface{}) Event {
	payload, _ := json.Marshal(data)
	return Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Data:      payload,
		Metadata:  make(map[string]string),
	}
}

// =============================================================================
// PART 2: In-Memory Event Bus (simulates Kafka/RabbitMQ/NATS)
// =============================================================================

type EventHandler func(ctx context.Context, event Event) error

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
	queue    chan Event
	dead     chan Event
}

func NewEventBus(bufferSize int) *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
		queue:    make(chan Event, bufferSize),
		dead:     make(chan Event, bufferSize),
	}
}

func (eb *EventBus) Subscribe(topic string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[topic] = append(eb.handlers[topic], handler)
}

func (eb *EventBus) Publish(event Event) {
	select {
	case eb.queue <- event:
	default:
		fmt.Printf("  ⚠ Queue full, dropping event: %s\n", event.Type)
	}
}

func (eb *EventBus) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eb.queue:
			eb.dispatch(ctx, event)
		}
	}
}

func (eb *EventBus) dispatch(ctx context.Context, event Event) {
	eb.mu.RLock()
	handlers := eb.handlers[event.Type]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			fmt.Printf("  ✗ Handler failed for %s: %v\n", event.Type, err)
			eb.dead <- event
		}
	}
}

// =============================================================================
// PART 3: Microservices that communicate via events
// =============================================================================

// --- Order Service (Event Producer) ---

type OrderService struct {
	bus    *EventBus
	orders map[int64]string
	mu     sync.Mutex
}

func NewOrderService(bus *EventBus) *OrderService {
	return &OrderService{bus: bus, orders: make(map[int64]string)}
}

func (s *OrderService) CreateOrder(userID int64, products []int64, total float64) int64 {
	s.mu.Lock()
	orderID := int64(len(s.orders) + 1)
	s.orders[orderID] = "created"
	s.mu.Unlock()

	fmt.Printf("\n[Order Service] Order %d created for user %d\n", orderID, userID)

	event := NewEvent("order.created", "order-service", OrderCreatedEvent{
		OrderID:    orderID,
		UserID:     userID,
		Products:   products,
		TotalPrice: total,
	})
	event.Metadata["correlation_id"] = fmt.Sprintf("corr_%d", orderID)
	s.bus.Publish(event)

	return orderID
}

// --- Payment Service (Event Consumer + Producer) ---

type PaymentService struct {
	bus *EventBus
}

func NewPaymentService(bus *EventBus) *PaymentService {
	svc := &PaymentService{bus: bus}
	bus.Subscribe("order.created", svc.handleOrderCreated)
	return svc
}

func (s *PaymentService) handleOrderCreated(ctx context.Context, event Event) error {
	var data OrderCreatedEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("decode event: %w", err)
	}

	fmt.Printf("[Payment Service] Processing payment for order %d: $%.2f\n",
		data.OrderID, data.TotalPrice)
	time.Sleep(50 * time.Millisecond)

	paymentEvent := NewEvent("payment.completed", "payment-service", PaymentCompletedEvent{
		PaymentID: 1001,
		OrderID:   data.OrderID,
		Amount:    data.TotalPrice,
		Method:    "credit_card",
	})
	paymentEvent.Metadata["correlation_id"] = event.Metadata["correlation_id"]
	s.bus.Publish(paymentEvent)

	return nil
}

// --- Inventory Service (Event Consumer) ---

type InventoryService struct {
	stock map[int64]int
	mu    sync.Mutex
}

func NewInventoryService(bus *EventBus) *InventoryService {
	svc := &InventoryService{
		stock: map[int64]int{1: 100, 2: 50, 3: 25},
	}
	bus.Subscribe("order.created", svc.handleOrderCreated)
	return svc
}

func (s *InventoryService) handleOrderCreated(ctx context.Context, event Event) error {
	var data OrderCreatedEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, pid := range data.Products {
		if s.stock[pid] <= 0 {
			return fmt.Errorf("product %d out of stock", pid)
		}
		s.stock[pid]--
		fmt.Printf("[Inventory Service] Reserved product %d (remaining: %d)\n", pid, s.stock[pid])
	}
	return nil
}

// --- Notification Service (Event Consumer) ---

type NotificationService struct{}

func NewNotificationService(bus *EventBus) *NotificationService {
	svc := &NotificationService{}
	bus.Subscribe("order.created", svc.handleOrderCreated)
	bus.Subscribe("payment.completed", svc.handlePaymentCompleted)
	return svc
}

func (s *NotificationService) handleOrderCreated(ctx context.Context, event Event) error {
	var data OrderCreatedEvent
	json.Unmarshal(event.Data, &data)
	fmt.Printf("[Notification Service] Sending order confirmation email for order %d\n", data.OrderID)
	return nil
}

func (s *NotificationService) handlePaymentCompleted(ctx context.Context, event Event) error {
	var data PaymentCompletedEvent
	json.Unmarshal(event.Data, &data)
	fmt.Printf("[Notification Service] Sending payment receipt for order %d ($%.2f)\n",
		data.OrderID, data.Amount)
	return nil
}

// =============================================================================
// CONCEPTS: Message Delivery Guarantees
// =============================================================================
//
// AT-MOST-ONCE:  Fire and forget. Message may be lost.
// AT-LEAST-ONCE: Message delivered, but may be duplicated. HANDLERS MUST BE IDEMPOTENT.
// EXACTLY-ONCE:  Very hard. "at-least-once + idempotency = effectively exactly-once"
//
// RECOMMENDATION: Design for at-least-once + idempotent handlers.

// =============================================================================
// CONCEPTS: Message Brokers Comparison
// =============================================================================
//
// ┌──────────────┬──────────────┬──────────────┬──────────────┐
// │              │ Kafka         │ RabbitMQ      │ NATS          │
// ├──────────────┼──────────────┼──────────────┼──────────────┤
// │ Model        │ Log (stream)  │ Queue+Pub/Sub │ Pub/Sub      │
// │ Ordering     │ Per partition │ Per queue     │ Per subject   │
// │ Replay       │ ✅ Yes        │ ❌ No         │ ✅ JetStream  │
// │ Throughput   │ ~1M/s        │ ~50K/s        │ ~10M/s        │
// │ Best for     │ Event sourcing│ Task queues   │ Real-time     │
// └──────────────┴──────────────┴──────────────┴──────────────┘

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	bus := NewEventBus(100)

	orderSvc := NewOrderService(bus)
	_ = NewPaymentService(bus)
	_ = NewInventoryService(bus)
	_ = NewNotificationService(bus)

	go bus.Start(ctx)

	fmt.Println("=== Event-Driven Microservices ===")
	fmt.Println("Flow: Order Created → Payment + Inventory + Notification (parallel)")
	fmt.Println("      Payment Completed → Notification (receipt)")

	orderSvc.CreateOrder(1, []int64{1, 2}, 199.98)
	time.Sleep(500 * time.Millisecond)

	orderSvc.CreateOrder(2, []int64{3}, 399.99)
	time.Sleep(500 * time.Millisecond)
	cancel()

	fmt.Println("\n=== EVENT-DRIVEN KEY PATTERNS ===")
	fmt.Println("1. Events are facts (past tense): OrderCreated, not CreateOrder")
	fmt.Println("2. Producers don't know consumers — loose coupling")
	fmt.Println("3. Use correlation IDs to trace events across services")
	fmt.Println("4. Handlers MUST be idempotent (at-least-once delivery)")
	fmt.Println("5. Dead letter queue for failed events")
	fmt.Println("6. Event envelope: ID, type, source, timestamp, data, metadata")
}
