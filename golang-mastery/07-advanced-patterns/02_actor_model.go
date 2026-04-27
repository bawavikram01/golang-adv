//go:build ignore

// =============================================================================
// ACTOR MODEL IN GO — Complete Deep Dive
// =============================================================================
//
// The Actor Model is a mathematical model for concurrent computation where
// "actors" are the universal primitives. Each actor can:
//   1. Receive messages
//   2. Create new actors
//   3. Send messages to other actors
//   4. Decide how to handle the next message (change internal state)
//
// KEY INSIGHT: Actors NEVER share memory. All communication happens via
// messages. This eliminates data races BY DESIGN.
//
// Go doesn't have a built-in actor framework (unlike Erlang/Akka), but
// goroutines + channels map naturally to the actor model. A goroutine IS
// essentially an actor, and a channel IS a mailbox.
//
// ┌─────────────────────────────────────────────────────────┐
// │                   ACTOR MODEL vs GO                     │
// │                                                         │
// │   Actor Concept        │  Go Equivalent                 │
// │   ─────────────────────┼──────────────────────────       │
// │   Actor                │  goroutine with a loop          │
// │   Mailbox              │  channel                        │
// │   Message              │  value sent on channel          │
// │   State                │  local variables in goroutine   │
// │   Supervision          │  parent goroutine + recovery    │
// │   Actor System         │  the Go runtime                 │
// │                                                         │
// │   "Don't communicate by sharing memory;                 │
// │    share memory by communicating." — Go Proverb         │
// └─────────────────────────────────────────────────────────┘
//
// WHY ACTOR MODEL?
// - No locks, no mutexes, no data races
// - Each actor reasons about its own state independently
// - Natural for distributed systems (actors can be on different machines)
// - Fault isolation (one actor crashing doesn't kill others)
// - Scales naturally (add more actors = more throughput)
//
// RUN: go run 02_actor_model.go
// =============================================================================

package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== ACTOR MODEL IN GO ===")
	fmt.Println()

	// PART 1: Basic Actor
	basicActorExample()

	// PART 2: Actor with Request-Reply
	requestReplyExample()

	// PART 3: Typed Actor (type-safe messages)
	typedActorExample()

	// PART 4: Actor Supervision (Erlang-style)
	supervisionExample()

	// PART 5: Actor Registry / Router
	routerExample()

	// PART 6: Real-World: Bank Account Actor
	bankAccountExample()

	// PART 7: Real-World: Chat Room
	chatRoomExample()

	// PART 8: Actor Hierarchy (Parent-Child)
	hierarchyExample()
}

// =============================================================================
// PART 1: Basic Actor — The Fundamental Pattern
// =============================================================================
//
// An actor is simply a goroutine that:
// 1. Owns private state
// 2. Reads messages from a channel (its mailbox)
// 3. Processes one message at a time (sequential!)
// 4. Can send messages to other actors
//
// ┌───────────┐  message   ┌──────────┐
// │  Sender   │──────────►│  MAILBOX  │ ← buffered channel
// └───────────┘            └────┬─────┘
//                               │ receive
//                          ┌────▼─────┐
//                          │  ACTOR   │ ← goroutine
//                          │  state   │ ← local vars
//                          │  logic   │ ← message handler
//                          └──────────┘
//
// CRITICAL: The actor processes messages SEQUENTIALLY.
// Even though many goroutines send concurrently, the actor sees
// one message at a time. This is what eliminates data races.

type Message struct {
	Type    string
	Payload interface{}
}

// Actor is a goroutine with a mailbox
type Actor struct {
	mailbox chan Message
	quit    chan struct{}
}

func NewActor(handler func(Message)) *Actor {
	a := &Actor{
		mailbox: make(chan Message, 100), // buffered mailbox
		quit:    make(chan struct{}),
	}
	go func() {
		for {
			select {
			case msg := <-a.mailbox:
				handler(msg) // process one at a time
			case <-a.quit:
				return
			}
		}
	}()
	return a
}

func (a *Actor) Send(msg Message) {
	a.mailbox <- msg
}

func (a *Actor) Stop() {
	close(a.quit)
}

func basicActorExample() {
	fmt.Println("--- PART 1: Basic Actor ---")

	// Create an actor that counts messages
	count := 0
	counter := NewActor(func(msg Message) {
		switch msg.Type {
		case "increment":
			count++
		case "print":
			fmt.Printf("  Counter: %d\n", count)
		}
	})

	// Send messages from multiple goroutines — no locks needed!
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Send(Message{Type: "increment"})
		}()
	}
	wg.Wait()

	// Give actor time to process all messages
	time.Sleep(10 * time.Millisecond)
	counter.Send(Message{Type: "print"})
	time.Sleep(10 * time.Millisecond)
	counter.Stop()

	fmt.Println()
}

// =============================================================================
// PART 2: Request-Reply Pattern (Ask Pattern)
// =============================================================================
//
// Sometimes you need a response from an actor. The caller sends a message
// that includes a reply channel. The actor sends the result back on it.
//
// ┌────────┐   {request, replyCh}   ┌─────────┐
// │ Caller │────────────────────────►│  Actor  │
// └───┬────┘                         └────┬────┘
//     │         result on replyCh         │
//     │◄──────────────────────────────────│
//
// This is the Go equivalent of Akka's "ask" pattern.

type Request struct {
	Key   string
	Reply chan string // caller provides this
}

func requestReplyExample() {
	fmt.Println("--- PART 2: Request-Reply ---")

	// Key-value store actor
	mailbox := make(chan interface{}, 100)

	type SetMsg struct {
		Key, Value string
	}
	type GetMsg struct {
		Key   string
		Reply chan string
	}

	// Actor goroutine — owns the map, no locks needed
	go func() {
		store := make(map[string]string) // private state
		for msg := range mailbox {
			switch m := msg.(type) {
			case SetMsg:
				store[m.Key] = m.Value
			case GetMsg:
				m.Reply <- store[m.Key]
			}
		}
	}()

	// Set values
	mailbox <- SetMsg{"name", "Go"}
	mailbox <- SetMsg{"version", "1.22"}

	// Get value with reply
	reply := make(chan string, 1)
	mailbox <- GetMsg{"name", reply}
	fmt.Printf("  Got: %s\n", <-reply)

	mailbox <- GetMsg{"version", reply}
	fmt.Printf("  Got: %s\n", <-reply)

	close(mailbox)
	fmt.Println()
}

// =============================================================================
// PART 3: Typed Actor — Type-Safe Messages with Generics
// =============================================================================
//
// The basic actor uses interface{} which isn't type-safe. With generics
// (Go 1.18+), we can create typed actors that only accept specific messages.

type TypedActor[M any] struct {
	mailbox chan M
	done    chan struct{}
}

func NewTypedActor[M any](bufSize int, handler func(M)) *TypedActor[M] {
	a := &TypedActor[M]{
		mailbox: make(chan M, bufSize),
		done:    make(chan struct{}),
	}
	go func() {
		defer close(a.done)
		for msg := range a.mailbox {
			handler(msg)
		}
	}()
	return a
}

func (a *TypedActor[M]) Send(msg M) {
	a.mailbox <- msg
}

func (a *TypedActor[M]) Stop() {
	close(a.mailbox)
	<-a.done // wait for all messages to be processed
}

// Typed message enum using sum types (sealed interface pattern)
type CounterMsg interface {
	isCounterMsg()
}

type Increment struct{ Amount int }
type GetCount struct{ Reply chan int }

func (Increment) isCounterMsg() {}
func (GetCount) isCounterMsg()  {}

func typedActorExample() {
	fmt.Println("--- PART 3: Typed Actor ---")

	count := 0
	counter := NewTypedActor[CounterMsg](100, func(msg CounterMsg) {
		switch m := msg.(type) {
		case Increment:
			count += m.Amount
		case GetCount:
			m.Reply <- count
		}
	})

	// Type-safe: can only send CounterMsg
	counter.Send(Increment{Amount: 5})
	counter.Send(Increment{Amount: 3})

	reply := make(chan int, 1)
	counter.Send(GetCount{Reply: reply})
	fmt.Printf("  Typed counter: %d\n", <-reply)

	counter.Stop() // graceful: processes remaining messages
	fmt.Println()
}

// =============================================================================
// PART 4: Supervision — Erlang's "Let It Crash" in Go
// =============================================================================
//
// In Erlang/Akka, supervisors monitor child actors and restart them on failure.
// Go doesn't have built-in supervision, but we can build it.
//
// ┌──────────────┐
// │  SUPERVISOR  │  monitors children, restarts on panic
// └──────┬───────┘
//        │ spawns & monitors
//   ┌────┴────┬─────────┐
//   │         │         │
// ┌─▼──┐  ┌──▼─┐  ┌───▼┐
// │ A1 │  │ A2 │  │ A3 │  child actors
// └────┘  └────┘  └────┘
//
// Supervision strategies:
// - ONE_FOR_ONE: restart only the failed actor
// - ONE_FOR_ALL: restart all children if one fails
// - REST_FOR_ONE: restart failed actor and all actors started after it

type SupervisedActor struct {
	name   string
	work   func(ctx context.Context)
	cancel context.CancelFunc
}

type Supervisor struct {
	children []*SupervisedActor
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewSupervisor() *Supervisor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Supervisor{ctx: ctx, cancel: cancel}
}

func (s *Supervisor) Spawn(name string, work func(ctx context.Context)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithCancel(s.ctx)
	child := &SupervisedActor{name: name, work: work, cancel: cancel}
	s.children = append(s.children, child)

	go s.supervise(child, ctx)
}

func (s *Supervisor) supervise(child *SupervisedActor, ctx context.Context) {
	maxRestarts := 3
	for i := 0; i < maxRestarts; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("  [Supervisor] %s panicked: %v — restarting (%d/%d)\n",
						child.name, r, i+1, maxRestarts)
				}
			}()
			child.work(ctx)
		}()
	}
	fmt.Printf("  [Supervisor] %s exceeded max restarts, giving up\n", child.name)
}

func (s *Supervisor) Shutdown() {
	s.cancel()
}

func supervisionExample() {
	fmt.Println("--- PART 4: Supervision ---")

	sup := NewSupervisor()

	// Spawn an actor that randomly panics
	sup.Spawn("flaky-worker", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if rand.Intn(5) == 0 {
					panic("random failure!")
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
	})

	// Spawn a stable actor
	sup.Spawn("stable-worker", func(ctx context.Context) {
		<-ctx.Done()
		fmt.Println("  [stable-worker] shutting down gracefully")
	})

	time.Sleep(100 * time.Millisecond)
	sup.Shutdown()
	time.Sleep(50 * time.Millisecond) // let shutdown complete
	fmt.Println()
}

// =============================================================================
// PART 5: Actor Router — Distribute Work Across Actor Pool
// =============================================================================
//
// A router distributes messages across a pool of actors.
// Strategies: round-robin, random, consistent hashing, broadcast.
//
// ┌────────┐        ┌─────────────┐
// │ Sender │───────►│   ROUTER    │
// └────────┘        └──┬───┬───┬──┘
//                      │   │   │
//                   ┌──▼┐┌─▼─┐┌▼──┐
//                   │ A1││ A2││ A3│  worker actors
//                   └───┘└───┘└───┘

type Router struct {
	workers []*TypedActor[string]
	next    int
}

func NewRouter(n int, handler func(string)) *Router {
	r := &Router{}
	for i := 0; i < n; i++ {
		r.workers = append(r.workers, NewTypedActor[string](50, handler))
	}
	return r
}

// RoundRobin sends to workers in rotation
func (r *Router) Send(msg string) {
	r.workers[r.next%len(r.workers)].Send(msg)
	r.next++
}

// Broadcast sends to ALL workers
func (r *Router) Broadcast(msg string) {
	for _, w := range r.workers {
		w.Send(msg)
	}
}

func (r *Router) Stop() {
	for _, w := range r.workers {
		w.Stop()
	}
}

func routerExample() {
	fmt.Println("--- PART 5: Router ---")

	var mu sync.Mutex
	processed := 0

	router := NewRouter(3, func(msg string) {
		// Each worker processes independently
		mu.Lock()
		processed++
		mu.Unlock()
	})

	// Distribute 100 messages across 3 actors
	for i := 0; i < 100; i++ {
		router.Send(fmt.Sprintf("job-%d", i))
	}

	router.Stop() // wait for all messages
	fmt.Printf("  Processed %d messages across 3 actors\n", processed)
	fmt.Println()
}

// =============================================================================
// PART 6: Real-World Example — Bank Account Actor
// =============================================================================
//
// Classic actor model example: bank account that's safe without locks.
// All operations go through the actor's mailbox.

type AccountMsg interface{ isAccountMsg() }

type Deposit struct {
	Amount float64
	Reply  chan float64
}
type Withdraw struct {
	Amount float64
	Reply  chan error
}
type Balance struct {
	Reply chan float64
}

func (Deposit) isAccountMsg()  {}
func (Withdraw) isAccountMsg() {}
func (Balance) isAccountMsg()  {}

type BankAccount struct {
	*TypedActor[AccountMsg]
}

func NewBankAccount(initialBalance float64) *BankAccount {
	balance := initialBalance

	actor := NewTypedActor[AccountMsg](100, func(msg AccountMsg) {
		switch m := msg.(type) {
		case Deposit:
			balance += m.Amount
			m.Reply <- balance
		case Withdraw:
			if balance < m.Amount {
				m.Reply <- fmt.Errorf("insufficient funds: have %.2f, want %.2f",
					balance, m.Amount)
			} else {
				balance -= m.Amount
				m.Reply <- nil
			}
		case Balance:
			m.Reply <- balance
		}
	})

	return &BankAccount{actor}
}

func bankAccountExample() {
	fmt.Println("--- PART 6: Bank Account Actor ---")

	account := NewBankAccount(1000.0)

	// Concurrent deposits — all safe, no locks
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reply := make(chan float64, 1)
			account.Send(Deposit{Amount: 100.0, Reply: reply})
			<-reply
		}()
	}
	wg.Wait()

	// Check balance
	balReply := make(chan float64, 1)
	account.Send(Balance{Reply: balReply})
	fmt.Printf("  Balance after 10 deposits: $%.2f\n", <-balReply)

	// Try overdraft
	errReply := make(chan error, 1)
	account.Send(Withdraw{Amount: 99999.0, Reply: errReply})
	if err := <-errReply; err != nil {
		fmt.Printf("  Withdraw error: %s\n", err)
	}

	account.Stop()
	fmt.Println()
}

// =============================================================================
// PART 7: Real-World Example — Chat Room (Multi-Actor System)
// =============================================================================
//
// A chat room where:
// - Each user is an actor
// - The room itself is an actor
// - Messages flow: User → Room → All Users (broadcast)
//
// ┌──────┐  join/say   ┌──────────┐  broadcast   ┌──────┐
// │ User ├────────────►│   Room   ├──────────────►│ User │
// └──────┘             │  Actor   │               └──────┘
//                      └────┬─────┘
//                           │ broadcast
//                      ┌────▼─────┐
//                      │  User 3  │
//                      └──────────┘

type RoomMsg interface{ isRoomMsg() }

type Join struct {
	Name  string
	Inbox chan string
}
type Say struct {
	From string
	Text string
}
type Leave struct {
	Name string
}

func (Join) isRoomMsg()  {}
func (Say) isRoomMsg()   {}
func (Leave) isRoomMsg() {}

type ChatRoom struct {
	*TypedActor[RoomMsg]
}

func NewChatRoom() *ChatRoom {
	members := make(map[string]chan string)

	actor := NewTypedActor[RoomMsg](100, func(msg RoomMsg) {
		switch m := msg.(type) {
		case Join:
			members[m.Name] = m.Inbox
			for name, inbox := range members {
				if name != m.Name {
					inbox <- fmt.Sprintf("[%s joined the chat]", m.Name)
				}
			}
		case Say:
			line := fmt.Sprintf("%s: %s", m.From, m.Text)
			for name, inbox := range members {
				if name != m.From {
					inbox <- line
				}
			}
		case Leave:
			delete(members, m.Name)
			for _, inbox := range members {
				inbox <- fmt.Sprintf("[%s left the chat]", m.Name)
			}
		}
	})

	return &ChatRoom{actor}
}

func chatRoomExample() {
	fmt.Println("--- PART 7: Chat Room ---")

	room := NewChatRoom()

	// Each user has their own inbox (they are also actors)
	aliceInbox := make(chan string, 10)
	bobInbox := make(chan string, 10)

	room.Send(Join{Name: "Alice", Inbox: aliceInbox})
	room.Send(Join{Name: "Bob", Inbox: bobInbox})
	room.Send(Say{From: "Alice", Text: "Hello everyone!"})
	room.Send(Say{From: "Bob", Text: "Hey Alice!"})
	room.Send(Leave{Name: "Alice"})

	// Give actor time to process
	time.Sleep(20 * time.Millisecond)

	// Drain inboxes
	fmt.Println("  Alice's inbox:")
	drainInbox(aliceInbox)
	fmt.Println("  Bob's inbox:")
	drainInbox(bobInbox)

	room.Stop()
	fmt.Println()
}

func drainInbox(ch chan string) {
	for {
		select {
		case msg := <-ch:
			fmt.Printf("    %s\n", msg)
		default:
			return
		}
	}
}

// =============================================================================
// PART 8: Actor Hierarchy — Parent-Child Relationships
// =============================================================================
//
// Actors can spawn child actors, creating a tree. The parent monitors
// children and handles their lifecycle.
//
// This is how Erlang/OTP works: supervision trees.
//
//                    ┌──────────┐
//                    │  System  │  root supervisor
//                    └────┬─────┘
//              ┌──────────┼──────────┐
//         ┌────▼───┐ ┌────▼───┐ ┌───▼────┐
//         │  HTTP  │ │  DB    │ │  Cache │  service actors
//         └────┬───┘ └────────┘ └───┬────┘
//         ┌────┴────┐          ┌────┴────┐
//      ┌──▼─┐  ┌───▼┐     ┌──▼─┐  ┌───▼┐
//      │ H1 │  │ H2 │     │ C1 │  │ C2 │  worker actors
//      └────┘  └────┘     └────┘  └────┘

type ActorRef struct {
	Name     string
	mailbox  chan interface{}
	ctx      context.Context
	cancel   context.CancelFunc
	children []*ActorRef
	mu       sync.Mutex
}

func SpawnActor(parent context.Context, name string, handler func(context.Context, interface{})) *ActorRef {
	ctx, cancel := context.WithCancel(parent)
	ref := &ActorRef{
		Name:    name,
		mailbox: make(chan interface{}, 50),
		ctx:     ctx,
		cancel:  cancel,
	}

	go func() {
		defer cancel()
		for {
			select {
			case msg := <-ref.mailbox:
				handler(ctx, msg)
			case <-ctx.Done():
				// Parent cancelled → all children auto-cancelled via context
				fmt.Printf("    [%s] shutting down\n", name)
				return
			}
		}
	}()

	return ref
}

func (a *ActorRef) SpawnChild(name string, handler func(context.Context, interface{})) *ActorRef {
	child := SpawnActor(a.ctx, name, handler)
	a.mu.Lock()
	a.children = append(a.children, child)
	a.mu.Unlock()
	return child
}

func (a *ActorRef) Tell(msg interface{}) {
	select {
	case a.mailbox <- msg:
	case <-a.ctx.Done():
	}
}

func (a *ActorRef) Shutdown() {
	a.cancel() // cascades to all children via context
}

func hierarchyExample() {
	fmt.Println("--- PART 8: Actor Hierarchy ---")

	rootCtx := context.Background()

	// Root actor
	system := SpawnActor(rootCtx, "system", func(ctx context.Context, msg interface{}) {
		fmt.Printf("    [system] received: %v\n", msg)
	})

	// Spawn child actors
	httpActor := system.SpawnChild("http-server", func(ctx context.Context, msg interface{}) {
		fmt.Printf("    [http-server] handling: %v\n", msg)
	})

	// Spawn grandchild (child of http)
	httpActor.SpawnChild("http-handler-1", func(ctx context.Context, msg interface{}) {
		fmt.Printf("    [http-handler-1] processing: %v\n", msg)
	})

	// Send messages
	system.Tell("health-check")
	httpActor.Tell("GET /api/users")

	time.Sleep(20 * time.Millisecond)

	// Shutdown cascades: system → http-server → http-handler-1
	fmt.Println("  Shutting down system (cascades to all children):")
	system.Shutdown()
	time.Sleep(50 * time.Millisecond)

	fmt.Println()
	printSummary()
}

func printSummary() {
	fmt.Println("=== ACTOR MODEL CHEAT SHEET ===")
	fmt.Println()
	fmt.Println("  CORE PATTERN:")
	fmt.Println("    goroutine (actor) + channel (mailbox) + select loop")
	fmt.Println()
	fmt.Println("  RULES:")
	fmt.Println("    1. Never share state between actors")
	fmt.Println("    2. Communicate only via messages (channels)")
	fmt.Println("    3. Process one message at a time")
	fmt.Println("    4. Actors can create child actors")
	fmt.Println()
	fmt.Println("  PATTERNS:")
	fmt.Println("    Fire-and-forget:  actor.Send(msg)")
	fmt.Println("    Request-reply:    send msg with reply chan, wait on it")
	fmt.Println("    Router:           distribute across actor pool")
	fmt.Println("    Supervision:      restart children on panic")
	fmt.Println("    Hierarchy:        context cancellation cascades")
	fmt.Println()
	fmt.Println("  WHEN TO USE:")
	fmt.Println("    ✓ Managing stateful entities (users, sessions, devices)")
	fmt.Println("    ✓ Coordinating concurrent workflows")
	fmt.Println("    ✓ Building event-driven systems")
	fmt.Println("    ✓ When you'd reach for a mutex-protected struct")
	fmt.Println()
	fmt.Println("  WHEN NOT TO USE:")
	fmt.Println("    ✗ Simple request-response (just use functions)")
	fmt.Println("    ✗ Pure computation (use worker pools instead)")
	fmt.Println("    ✗ When shared memory with sync.Mutex is simpler")
	fmt.Println()
	fmt.Println("  GO LIBRARIES:")
	fmt.Println("    github.com/asynkron/protoactor-go  (closest to Akka)")
	fmt.Println("    github.com/ergo-services/ergo      (Erlang/OTP for Go)")
	fmt.Println("    Or just: goroutine + channel (this file's approach)")
}
