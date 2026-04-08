// =============================================================================
// LESSON 1: MICROSERVICES FUNDAMENTALS — What, Why, When, and How to Decompose
// =============================================================================
//
// DEFINITION: Microservices are independently deployable services that
// communicate over a network, each owning its own data and business logic.
//
// THIS FILE COVERS:
//   1. Monolith vs Microservices vs Modular Monolith (comparison)
//   2. Service decomposition strategies
//   3. Domain-Driven Design (DDD) for service boundaries
//   4. The two-pizza team rule and Conway's Law
//   5. When NOT to use microservices
//
// This file is a runnable simulation of decomposing a monolith.
// =============================================================================

package main

import "fmt"

// =============================================================================
// CONCEPT 1: THE MONOLITH — Where everything starts
// =============================================================================
//
// A monolith is a single deployable unit containing ALL business logic.
// It's NOT inherently bad. Most companies SHOULD start with a monolith.
//
// Advantages:
//   - Simple to develop, test, deploy, debug
//   - Single database, no distributed transactions
//   - IDE support: find all references, refactor safely
//   - Low latency (in-process function calls)
//
// When it breaks:
//   - Team size > 10-15 (merge conflicts, slow CI, coordination overhead)
//   - Deploy frequency needed > daily (one bug blocks everything)
//   - Different scaling needs (orders need 100x compute, catalog needs 1x)
//   - Technology lock-in (can't use Python ML alongside Go API)

// Monolithic e-commerce — everything in one binary
type Monolith struct {
	// All domains mixed together
	users    map[int64]*User
	products map[int64]*Product
	orders   map[int64]*Order
	payments map[int64]*Payment
}

type User struct {
	ID    int64
	Name  string
	Email string
}

type Product struct {
	ID    int64
	Name  string
	Price float64
	Stock int
}

type Order struct {
	ID       int64
	UserID   int64
	Products []int64
	Total    float64
	Status   string
}

type Payment struct {
	ID      int64
	OrderID int64
	Amount  float64
	Status  string
}

// In a monolith, creating an order touches ALL domains directly:
func (m *Monolith) CreateOrder(userID int64, productIDs []int64) (*Order, error) {
	// Check user exists (User domain)
	user := m.users[userID]
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check products and calculate total (Product domain)
	var total float64
	for _, pid := range productIDs {
		p := m.products[pid]
		if p == nil {
			return nil, fmt.Errorf("product %d not found", pid)
		}
		if p.Stock <= 0 {
			return nil, fmt.Errorf("product %d out of stock", pid)
		}
		total += p.Price
		p.Stock-- // Direct mutation — tightly coupled!
	}

	// Create order (Order domain)
	order := &Order{
		ID:       int64(len(m.orders) + 1),
		UserID:   userID,
		Products: productIDs,
		Total:    total,
		Status:   "pending",
	}
	m.orders[order.ID] = order

	// Process payment (Payment domain)
	payment := &Payment{
		ID:      int64(len(m.payments) + 1),
		OrderID: order.ID,
		Amount:  total,
		Status:  "completed",
	}
	m.payments[payment.ID] = payment

	order.Status = "confirmed"
	return order, nil
}

// =============================================================================
// CONCEPT 2: DECOMPOSITION STRATEGIES
// =============================================================================
//
// How do you decide what becomes a service? There are 4 main strategies:
//
// STRATEGY 1: Decompose by Business Capability
//   "What does the business DO?"
//   Each capability = a service
//   - User Management Service
//   - Product Catalog Service
//   - Order Service
//   - Payment Service
//   - Shipping Service
//   - Notification Service
//
// STRATEGY 2: Decompose by Subdomain (DDD)
//   "What are the bounded contexts?"
//   Core Domain    → where competitive advantage lives (invest most here)
//   Supporting     → necessary but not differentiating
//   Generic        → commodity (use off-the-shelf: auth, email, payments)
//
// STRATEGY 3: Decompose by Team (Conway's Law)
//   "Organizations which design systems are constrained to produce designs
//    which are copies of the communication structures of these organizations."
//   Each team owns 1-3 services. Service boundaries = team boundaries.
//
// STRATEGY 4: Strangler Fig Pattern (for migrating from monolith)
//   Gradually replace monolith pieces with services.
//   Route traffic to new service. When stable, remove old code.

// =============================================================================
// CONCEPT 3: SIMULATING MICROSERVICE BOUNDARIES
// =============================================================================
// Now let's model the SAME system as microservices.
// Each service is an independent unit with its own data.

// --- User Service ---
// Owns: user data, authentication, profiles
// Exposes: GET /users/{id}, POST /users
type UserService struct {
	users map[int64]*User
}

func (s *UserService) GetUser(id int64) (*User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user %d not found", id)
	}
	return u, nil
}

// --- Product Service ---
// Owns: catalog, inventory, pricing
// Exposes: GET /products/{id}, POST /products/{id}/reserve
type ProductService struct {
	products map[int64]*Product
}

func (s *ProductService) GetProduct(id int64) (*Product, error) {
	p, ok := s.products[id]
	if !ok {
		return nil, fmt.Errorf("product %d not found", id)
	}
	return p, nil
}

func (s *ProductService) ReserveStock(id int64, qty int) error {
	p, ok := s.products[id]
	if !ok {
		return fmt.Errorf("product %d not found", id)
	}
	if p.Stock < qty {
		return fmt.Errorf("insufficient stock")
	}
	p.Stock -= qty
	return nil
}

// --- Order Service ---
// Owns: orders, order lifecycle
// CALLS: UserService, ProductService, PaymentService
// This is the ORCHESTRATOR in this design
type OrderService struct {
	orders   map[int64]*Order
	users    *UserService // Would be HTTP/gRPC client in reality
	products *ProductService
	payments *PaymentService
}

func (s *OrderService) CreateOrder(userID int64, productIDs []int64) (*Order, error) {
	// Step 1: Verify user exists (NETWORK CALL to User Service)
	_, err := s.users.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}

	// Step 2: Reserve products (NETWORK CALL to Product Service)
	var total float64
	for _, pid := range productIDs {
		p, err := s.products.GetProduct(pid)
		if err != nil {
			return nil, fmt.Errorf("product service: %w", err)
		}
		if err := s.products.ReserveStock(pid, 1); err != nil {
			// TODO: Need to release already-reserved products (compensation!)
			return nil, fmt.Errorf("product service reserve: %w", err)
		}
		total += p.Price
	}

	// Step 3: Create order locally
	order := &Order{
		ID:       int64(len(s.orders) + 1),
		UserID:   userID,
		Products: productIDs,
		Total:    total,
		Status:   "pending_payment",
	}
	s.orders[order.ID] = order

	// Step 4: Process payment (NETWORK CALL to Payment Service)
	err = s.payments.ProcessPayment(order.ID, total)
	if err != nil {
		// Payment failed — need COMPENSATION (release stock, cancel order)
		order.Status = "payment_failed"
		// s.products.ReleaseStock(...) — compensation action
		return order, fmt.Errorf("payment service: %w", err)
	}

	order.Status = "confirmed"
	return order, nil
}

// --- Payment Service ---
// Owns: payment processing, refunds, payment methods
type PaymentService struct {
	payments map[int64]*Payment
}

func (s *PaymentService) ProcessPayment(orderID int64, amount float64) error {
	payment := &Payment{
		ID:      int64(len(s.payments) + 1),
		OrderID: orderID,
		Amount:  amount,
		Status:  "completed",
	}
	s.payments[payment.ID] = payment
	return nil
}

// =============================================================================
// CONCEPT 4: KEY MICROSERVICE PRINCIPLES
// =============================================================================
//
// 1. SINGLE RESPONSIBILITY
//    Each service does one thing well.
//    User service doesn't know about orders.
//
// 2. OWN YOUR DATA
//    Each service has its OWN database.
//    No shared database! (If you share a DB, you have a distributed monolith)
//
//    User Service  → users_db (PostgreSQL)
//    Product Service → products_db (PostgreSQL)
//    Order Service → orders_db (MongoDB)
//    Payment Service → payments_db (PostgreSQL)
//
// 3. SMART ENDPOINTS, DUMB PIPES
//    Business logic lives in services, not in the message bus.
//    The network (HTTP, gRPC, message queue) is just transport.
//
// 4. DESIGN FOR FAILURE
//    Any network call can fail. Services MUST handle:
//    - Timeouts
//    - Retries (with idempotency)
//    - Circuit breakers
//    - Fallback responses
//
// 5. DECENTRALIZED GOVERNANCE
//    Each team picks their own tech stack, language, database.
//    Shared: API contracts, monitoring standards, deployment pipeline.
//
// 6. EVOLUTIONARY DESIGN
//    Services can be replaced, rewritten, or split further.
//    Start coarse-grained, split when needed.

// =============================================================================
// CONCEPT 5: WHEN NOT TO USE MICROSERVICES
// =============================================================================
//
// ❌ Small team (< 5 developers) — coordination overhead kills you
// ❌ Startup MVP — you don't know the domain boundaries yet
// ❌ Simple CRUD app — microservices add complexity for zero benefit
// ❌ No DevOps capability — you need CI/CD, containers, monitoring
// ❌ Tight performance requirements — network latency adds up
// ❌ Small domain — if everything is interconnected, don't split it
//
// THE GOLDEN RULE:
// "If you can't build a well-structured monolith,
//  what makes you think microservices will help?"
//                                    — Simon Brown
//
// START: Modular Monolith (clear boundaries, own packages, single deploy)
// SPLIT: When team/scale/deploy needs require it
// PATTERN: Strangler Fig — gradually extract services from the monolith

func main() {
	fmt.Println("=== MONOLITH APPROACH ===")
	mono := &Monolith{
		users: map[int64]*User{1: {ID: 1, Name: "Vikram", Email: "v@test.com"}},
		products: map[int64]*Product{
			1: {ID: 1, Name: "Go Book", Price: 49.99, Stock: 10},
			2: {ID: 2, Name: "Keyboard", Price: 149.99, Stock: 5},
		},
		orders:   make(map[int64]*Order),
		payments: make(map[int64]*Payment),
	}

	order, err := mono.CreateOrder(1, []int64{1, 2})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Monolith order: ID=%d Total=%.2f Status=%s\n", order.ID, order.Total, order.Status)
	}

	fmt.Println("\n=== MICROSERVICES APPROACH ===")
	userSvc := &UserService{users: map[int64]*User{1: {ID: 1, Name: "Vikram", Email: "v@test.com"}}}
	productSvc := &ProductService{products: map[int64]*Product{
		1: {ID: 1, Name: "Go Book", Price: 49.99, Stock: 10},
		2: {ID: 2, Name: "Keyboard", Price: 149.99, Stock: 5},
	}}
	paymentSvc := &PaymentService{payments: make(map[int64]*Payment)}
	orderSvc := &OrderService{
		orders:   make(map[int64]*Order),
		users:    userSvc,
		products: productSvc,
		payments: paymentSvc,
	}

	order, err = orderSvc.CreateOrder(1, []int64{1, 2})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Microservice order: ID=%d Total=%.2f Status=%s\n", order.ID, order.Total, order.Status)
	}

	fmt.Println("\n=== KEY DIFFERENCES ===")
	fmt.Println("Monolith:      Direct function calls, shared memory, one DB, one deploy")
	fmt.Println("Microservices: Network calls, separate DBs, independent deploys")
	fmt.Println()
	fmt.Println("Monolith trade:     Simple but coupled")
	fmt.Println("Microservice trade: Decoupled but complex (network, consistency, debugging)")
	fmt.Println()
	fmt.Println("=== DECOMPOSITION STRATEGIES ===")
	fmt.Println("1. By Business Capability — what the business does (User Mgmt, Orders, Payments)")
	fmt.Println("2. By Subdomain (DDD)    — Core / Supporting / Generic domains")
	fmt.Println("3. By Team (Conway's Law) — service boundary = team boundary")
	fmt.Println("4. Strangler Fig          — gradually extract from monolith")
}
