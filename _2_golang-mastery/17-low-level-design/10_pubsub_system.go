package lowleveldesign

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// LLD PROBLEM: PUB-SUB MESSAGING SYSTEM
// =============================================================================
// Design a publish-subscribe messaging system.
//
// Requirements:
// 1. Topics: Create, delete, list topics
// 2. Publishers: Publish messages to topics
// 3. Subscribers: Subscribe to topics, receive messages
// 4. Message ordering: Per-topic FIFO
// 5. At-least-once delivery
// 6. Subscriber groups (load balancing within a group)
// 7. Message acknowledgment
// 8. Dead letter queue for failed messages
// 9. Thread-safe
//
// Design Decisions:
// - Channel-based delivery (Go-idiomatic)
// - Interface-based (swappable backends: in-memory, Redis, Kafka)
// - Observer pattern for notifications
// - Strategy pattern for delivery (broadcast vs load-balanced)

// =============================================================================
// DOMAIN MODELS
// =============================================================================

type Message struct {
	ID        string
	Topic     string
	Payload   []byte
	Metadata  map[string]string
	CreatedAt time.Time
	Attempts  int
}

type Acknowledgment struct {
	MessageID    string
	SubscriberID string
	Success      bool
	Error        error
	AckedAt      time.Time
}

// =============================================================================
// SUBSCRIBER
// =============================================================================

type Subscriber struct {
	ID       string
	Group    string // empty = independent subscriber
	topics   map[string]bool
	messages chan *Message
	done     chan struct{}
	handler  MessageHandler
}

type MessageHandler func(msg *Message) error

func NewSubscriber(id, group string, bufferSize int, handler MessageHandler) *Subscriber {
	return &Subscriber{
		ID:       id,
		Group:    group,
		topics:   make(map[string]bool),
		messages: make(chan *Message, bufferSize),
		done:     make(chan struct{}),
		handler:  handler,
	}
}

// Start begins processing messages
func (s *Subscriber) Start(ctx context.Context) {
	go func() {
		defer close(s.done)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-s.messages:
				if !ok {
					return
				}
				if err := s.handler(msg); err != nil {
					fmt.Printf("[Subscriber %s] Error processing msg %s: %v\n",
						s.ID, msg.ID, err)
				}
			}
		}
	}()
}

// Deliver sends a message to this subscriber (non-blocking with buffer)
func (s *Subscriber) Deliver(msg *Message) bool {
	select {
	case s.messages <- msg:
		return true
	default:
		// Buffer full — message dropped (could go to DLQ)
		return false
	}
}

func (s *Subscriber) Stop() {
	close(s.messages)
	<-s.done
}

// =============================================================================
// TOPIC
// =============================================================================

type DeliveryMode int

const (
	Broadcast    DeliveryMode = iota // All subscribers get every message
	LoadBalanced                     // Round-robin within subscriber groups
)

type Topic struct {
	Name         string
	Mode         DeliveryMode
	subscribers  []*Subscriber
	groups       map[string][]*Subscriber // group -> subscribers
	mu           sync.RWMutex
	messageCount int64
	roundRobin   map[string]int // group -> next index
}

func NewTopic(name string, mode DeliveryMode) *Topic {
	return &Topic{
		Name:       name,
		Mode:       mode,
		groups:     make(map[string][]*Subscriber),
		roundRobin: make(map[string]int),
	}
}

func (t *Topic) AddSubscriber(sub *Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.subscribers = append(t.subscribers, sub)
	sub.topics[t.Name] = true

	if sub.Group != "" {
		t.groups[sub.Group] = append(t.groups[sub.Group], sub)
	}
	fmt.Printf("[Topic %s] Subscriber %s added (group: %q)\n", t.Name, sub.ID, sub.Group)
}

func (t *Topic) RemoveSubscriber(subID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, sub := range t.subscribers {
		if sub.ID == subID {
			t.subscribers = append(t.subscribers[:i], t.subscribers[i+1:]...)
			delete(sub.topics, t.Name)

			// Remove from group
			if sub.Group != "" {
				group := t.groups[sub.Group]
				for j, gs := range group {
					if gs.ID == subID {
						t.groups[sub.Group] = append(group[:j], group[j+1:]...)
						break
					}
				}
			}
			break
		}
	}
}

func (t *Topic) Publish(msg *Message) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.messageCount++
	msg.Topic = t.Name

	switch t.Mode {
	case Broadcast:
		// Every subscriber gets the message
		for _, sub := range t.subscribers {
			sub.Deliver(msg)
		}
	case LoadBalanced:
		// Independent subscribers get all messages
		// Group subscribers: round-robin within group
		delivered := make(map[string]bool)

		for _, sub := range t.subscribers {
			if sub.Group == "" {
				// Independent subscriber — broadcast
				sub.Deliver(msg)
			} else if !delivered[sub.Group] {
				// Pick one from the group (round-robin)
				group := t.groups[sub.Group]
				if len(group) > 0 {
					idx := t.roundRobin[sub.Group] % len(group)
					group[idx].Deliver(msg)
					t.roundRobin[sub.Group] = idx + 1
					delivered[sub.Group] = true
				}
			}
		}
	}
}

// =============================================================================
// BROKER (Main Orchestrator / Facade)
// =============================================================================

type Broker struct {
	topics map[string]*Topic
	mu     sync.RWMutex
	nextID int
}

func NewBroker() *Broker {
	return &Broker{
		topics: make(map[string]*Topic),
	}
}

func (b *Broker) CreateTopic(name string, mode DeliveryMode) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.topics[name]; exists {
		return fmt.Errorf("topic %q already exists", name)
	}
	b.topics[name] = NewTopic(name, mode)
	fmt.Printf("[Broker] Created topic: %s (mode: %v)\n", name, mode)
	return nil
}

func (b *Broker) DeleteTopic(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.topics[name]; !exists {
		return fmt.Errorf("topic %q not found", name)
	}
	delete(b.topics, name)
	return nil
}

func (b *Broker) Subscribe(topicName string, sub *Subscriber) error {
	b.mu.RLock()
	topic, exists := b.topics[topicName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("topic %q not found", topicName)
	}
	topic.AddSubscriber(sub)
	return nil
}

func (b *Broker) Unsubscribe(topicName, subID string) error {
	b.mu.RLock()
	topic, exists := b.topics[topicName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("topic %q not found", topicName)
	}
	topic.RemoveSubscriber(subID)
	return nil
}

func (b *Broker) Publish(topicName string, payload []byte, metadata map[string]string) error {
	b.mu.RLock()
	topic, exists := b.topics[topicName]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("topic %q not found", topicName)
	}

	b.mu.Lock()
	b.nextID++
	msgID := fmt.Sprintf("msg-%d", b.nextID)
	b.mu.Unlock()

	msg := &Message{
		ID:        msgID,
		Topic:     topicName,
		Payload:   payload,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	topic.Publish(msg)
	return nil
}

func (b *Broker) ListTopics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	names := make([]string, 0, len(b.topics))
	for name := range b.topics {
		names = append(names, name)
	}
	return names
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================

func ExamplePubSub() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker := NewBroker()

	// Create topics
	broker.CreateTopic("orders", Broadcast)
	broker.CreateTopic("notifications", LoadBalanced)

	// Create subscribers
	orderLogger := NewSubscriber("order-logger", "", 100, func(msg *Message) error {
		fmt.Printf("  📝 [Logger] Order event: %s\n", string(msg.Payload))
		return nil
	})

	orderAnalytics := NewSubscriber("order-analytics", "", 100, func(msg *Message) error {
		fmt.Printf("  📊 [Analytics] Processing: %s\n", string(msg.Payload))
		return nil
	})

	// Notification workers (load-balanced group)
	notifWorker1 := NewSubscriber("notif-1", "notif-workers", 100, func(msg *Message) error {
		fmt.Printf("  🔔 [Worker-1] Sending: %s\n", string(msg.Payload))
		return nil
	})

	notifWorker2 := NewSubscriber("notif-2", "notif-workers", 100, func(msg *Message) error {
		fmt.Printf("  🔔 [Worker-2] Sending: %s\n", string(msg.Payload))
		return nil
	})

	// Start subscribers
	orderLogger.Start(ctx)
	orderAnalytics.Start(ctx)
	notifWorker1.Start(ctx)
	notifWorker2.Start(ctx)

	// Subscribe to topics
	broker.Subscribe("orders", orderLogger)
	broker.Subscribe("orders", orderAnalytics)
	broker.Subscribe("notifications", notifWorker1)
	broker.Subscribe("notifications", notifWorker2)

	// Publish messages
	fmt.Println("\n--- Publishing order events (broadcast) ---")
	broker.Publish("orders", []byte("Order #1 created"), nil)
	broker.Publish("orders", []byte("Order #2 created"), nil)

	fmt.Println("\n--- Publishing notifications (load-balanced) ---")
	broker.Publish("notifications", []byte("Welcome email to user-1"), nil)
	broker.Publish("notifications", []byte("Welcome email to user-2"), nil)
	broker.Publish("notifications", []byte("Welcome email to user-3"), nil)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nTopics:", broker.ListTopics())
}
