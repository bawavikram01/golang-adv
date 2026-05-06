// =============================================================================
// LESSON 10: TESTING MICROSERVICES
// =============================================================================
//
// Testing microservices is HARDER than testing monoliths because:
//   - Services depend on other services (over the network)
//   - Tests need multiple services running
//   - Flaky tests from network issues
//   - Data consistency across services
//
// TESTING PYRAMID (bottom = more, top = fewer):
//
//        ╱  E2E Tests  ╲         ← Few, slow, expensive, high confidence
//       ╱ Contract Tests ╲       ← Medium, validate service interfaces
//      ╱ Integration Tests╲      ← Test with real DB, cache, etc.
//     ╱    Unit Tests      ╲     ← Many, fast, cheap, isolated
//
// This lesson covers every testing strategy for microservices.
// =============================================================================

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// PATTERN 1: Unit Testing with Test Doubles
// =============================================================================
//
// Unit tests in microservices: test ONE service in isolation.
// Replace all dependencies with test doubles (mocks, stubs, fakes).
//
// TEST DOUBLES:
//   Stub:  Returns canned responses. No behavior verification.
//   Mock:  Records calls. Verifies interactions (was method called? with what args?).
//   Fake:  Working implementation (in-memory DB instead of Postgres).
//   Spy:   Real implementation + records calls for verification.

// --- Production interface ---
type PaymentGateway interface {
	Charge(amount float64, currency, cardToken string) (string, error)
	Refund(paymentID string) error
}

type OrderRepository interface {
	Save(order OrderRecord) error
	FindByID(id string) (*OrderRecord, error)
	FindByUser(userID string) ([]OrderRecord, error)
}

type OrderRecord struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	PaymentID string  `json:"payment_id"`
}

// --- Fake: In-memory implementation for testing ---
type FakeOrderRepository struct {
	orders map[string]OrderRecord
	calls  []string // spy: records method calls
}

func NewFakeOrderRepository() *FakeOrderRepository {
	return &FakeOrderRepository{orders: make(map[string]OrderRecord)}
}

func (r *FakeOrderRepository) Save(order OrderRecord) error {
	r.calls = append(r.calls, fmt.Sprintf("Save(%s)", order.ID))
	r.orders[order.ID] = order
	return nil
}

func (r *FakeOrderRepository) FindByID(id string) (*OrderRecord, error) {
	r.calls = append(r.calls, fmt.Sprintf("FindByID(%s)", id))
	order, ok := r.orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	return &order, nil
}

func (r *FakeOrderRepository) FindByUser(userID string) ([]OrderRecord, error) {
	r.calls = append(r.calls, fmt.Sprintf("FindByUser(%s)", userID))
	var result []OrderRecord
	for _, o := range r.orders {
		if o.UserID == userID {
			result = append(result, o)
		}
	}
	return result, nil
}

// --- Mock: Payment gateway ---
type MockPaymentGateway struct {
	ChargeResponse string
	ChargeError    error
	RefundError    error
	ChargeCalls    []ChargeCall
	RefundCalls    []string
}

type ChargeCall struct {
	Amount    float64
	Currency  string
	CardToken string
}

func (m *MockPaymentGateway) Charge(amount float64, currency, cardToken string) (string, error) {
	m.ChargeCalls = append(m.ChargeCalls, ChargeCall{amount, currency, cardToken})
	return m.ChargeResponse, m.ChargeError
}

func (m *MockPaymentGateway) Refund(paymentID string) error {
	m.RefundCalls = append(m.RefundCalls, paymentID)
	return m.RefundError
}

// --- Service being tested ---
type OrderService struct {
	repo    OrderRepository
	payment PaymentGateway
}

func NewOrderService(repo OrderRepository, payment PaymentGateway) *OrderService {
	return &OrderService{repo: repo, payment: payment}
}

func (s *OrderService) PlaceOrder(userID string, amount float64, cardToken string) (*OrderRecord, error) {
	// Charge payment
	paymentID, err := s.payment.Charge(amount, "USD", cardToken)
	if err != nil {
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	// Save order
	order := OrderRecord{
		ID:        fmt.Sprintf("ord_%d", time.Now().UnixNano()),
		UserID:    userID,
		Amount:    amount,
		Status:    "confirmed",
		PaymentID: paymentID,
	}

	if err := s.repo.Save(order); err != nil {
		// Compensate: refund payment
		s.payment.Refund(paymentID)
		return nil, fmt.Errorf("order save failed: %w", err)
	}

	return &order, nil
}

// =============================================================================
// PATTERN 2: Consumer-Driven Contract Testing
// =============================================================================
//
// Problem: Service A depends on Service B's API.
//          Service B changes Response format → Service A breaks.
//          How do you prevent this?
//
// CONTRACT TEST:
//   Consumer (Service A) writes a CONTRACT: "I expect this response format"
//   Provider (Service B) runs the contract in their CI pipeline.
//   If Service B breaks the contract → their build fails BEFORE deployment.
//
// TOOLS: Pact (most popular), Spring Cloud Contract
//
// FLOW:
//   1. Consumer writes test: "When I call GET /users/1, I expect {id, name, email}"
//   2. Pact generates a CONTRACT file (JSON)
//   3. Contract is shared (Pact Broker or Git)
//   4. Provider runs contract tests → verifies their API matches
//   5. If provider changes break contracts → CI fails

type Contract struct {
	Consumer     string                `json:"consumer"`
	Provider     string                `json:"provider"`
	Interactions []ContractInteraction `json:"interactions"`
}

type ContractInteraction struct {
	Description string           `json:"description"`
	Request     ContractRequest  `json:"request"`
	Response    ContractResponse `json:"response"`
}

type ContractRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type ContractResponse struct {
	Status int               `json:"status"`
	Body   map[string]string `json:"body"`
}

func VerifyContract(contract Contract, provider func(method, path string) (int, map[string]string)) []string {
	var failures []string

	for _, interaction := range contract.Interactions {
		status, body := provider(interaction.Request.Method, interaction.Request.Path)

		// Check status code
		if status != interaction.Response.Status {
			failures = append(failures, fmt.Sprintf(
				"%s: expected status %d, got %d",
				interaction.Description, interaction.Response.Status, status))
		}

		// Check response body fields exist
		for key, expectedValue := range interaction.Response.Body {
			actualValue, exists := body[key]
			if !exists {
				failures = append(failures, fmt.Sprintf(
					"%s: missing field '%s' in response",
					interaction.Description, key))
			} else if expectedValue != "*" && actualValue != expectedValue {
				failures = append(failures, fmt.Sprintf(
					"%s: field '%s' expected '%s', got '%s'",
					interaction.Description, key, expectedValue, actualValue))
			}
		}
	}

	return failures
}

// =============================================================================
// PATTERN 3: Integration Testing
// =============================================================================
//
// Test service WITH real dependencies (database, cache, message queue).
// Use testcontainers-go or docker-compose for ephemeral dependencies.
//
// STRATEGY:
//   1. Spin up Postgres container
//   2. Run migrations
//   3. Test service against real DB
//   4. Tear down container
//
// TOOLS: testcontainers-go, docker-compose, localstack (for AWS services)
//
// EXAMPLE (conceptual):
//   func TestOrderRepository_Integration(t *testing.T) {
//       ctx := context.Background()
//       pg, _ := postgres.RunContainer(ctx,
//           testcontainers.WithImage("postgres:15"),
//           postgres.WithDatabase("orders"),
//       )
//       defer pg.Terminate(ctx)
//
//       db := connectDB(pg.GetDSN())
//       repo := NewPostgresOrderRepository(db)
//
//       order := Order{ID: "1", UserID: "u1", Amount: 99.99}
//       err := repo.Save(order)
//       assert.NoError(t, err)
//
//       found, err := repo.FindByID("1")
//       assert.Equal(t, order, found)
//   }

// =============================================================================
// PATTERN 4: Component Testing
// =============================================================================
//
// Test ONE complete microservice with ALL its dependencies stubbed at the
// network boundary. The service runs as a real HTTP server, but calls to
// other services are stubbed (WireMock, mountebank).
//
// FLOW:
//   1. Start the actual microservice
//   2. Stub external dependencies (WireMock for HTTP, localstack for AWS)
//   3. Send real HTTP requests to the service
//   4. Assert responses

// =============================================================================
// PATTERN 5: Chaos Engineering
// =============================================================================
//
// "Everything fails all the time." — Werner Vogels, AWS CTO
//
// Chaos Engineering = intentionally inject failures to find weaknesses.
//
// PRINCIPLES:
//   1. Define "steady state" (normal behavior metrics)
//   2. Hypothesize: "The system will handle X failure gracefully"
//   3. Inject failure in production (start small!)
//   4. Observe: Did the system handle it? Or did it break?
//
// FAILURE INJECTION:
//   - Kill random service instances (Chaos Monkey)
//   - Add network latency (tc netem, Toxiproxy)
//   - Inject HTTP errors (500, 503 responses)
//   - Fill disk space
//   - Consume CPU/memory
//   - Partition network (split brain)
//
// TOOLS: Chaos Monkey (Netflix), LitmusChaos, Gremlin, Toxiproxy, tc

type ChaosExperiment struct {
	Name       string
	Hypothesis string
	Injection  string
	Duration   time.Duration
	Rollback   string
}

var chaosExperiments = []ChaosExperiment{
	{
		Name:       "Service instance failure",
		Hypothesis: "System continues serving with N-1 instances",
		Injection:  "Kill 1 pod of order-service (kubectl delete pod ...)",
		Duration:   5 * time.Minute,
		Rollback:   "Kubernetes auto-restarts the pod",
	},
	{
		Name:       "Network latency",
		Hypothesis: "Circuit breaker opens, fallback response returned in <500ms",
		Injection:  "Add 2s latency to payment-service (Toxiproxy)",
		Duration:   3 * time.Minute,
		Rollback:   "Remove Toxiproxy toxic",
	},
	{
		Name:       "Database failure",
		Hypothesis: "Read requests served from cache, writes queued",
		Injection:  "Stop Postgres primary (docker stop postgres)",
		Duration:   2 * time.Minute,
		Rollback:   "Restart Postgres, verify data consistency",
	},
	{
		Name:       "DNS failure",
		Hypothesis: "Service discovery falls back to cached endpoints",
		Injection:  "Block DNS resolution (iptables -A OUTPUT -p udp --dport 53 -j DROP)",
		Duration:   1 * time.Minute,
		Rollback:   "Remove iptables rule",
	},
}

// =============================================================================
// PATTERN 6: End-to-End Testing
// =============================================================================
//
// Test the ENTIRE system from a user's perspective.
//
// WARNING: E2E tests are:
//   ❌ Slow (minutes to run)
//   ❌ Flaky (network, timing, data dependencies)
//   ❌ Expensive to maintain
//   ❌ Hard to debug when they fail
//
// RULES:
//   - Keep E2E tests to a MINIMUM (5-10 critical user journeys)
//   - Run in a staging environment, NOT production
//   - Focus on happy paths + critical error paths
//   - Use contract tests to catch most issues before E2E

func main() {
	// =========================================================================
	// DEMO 1: Unit Testing with Mocks/Fakes
	// =========================================================================
	fmt.Println("=== UNIT TESTING WITH TEST DOUBLES ===")

	fakeRepo := NewFakeOrderRepository()
	mockPayment := &MockPaymentGateway{
		ChargeResponse: "pay_123",
		ChargeError:    nil,
	}

	orderSvc := NewOrderService(fakeRepo, mockPayment)

	// Test: successful order placement
	order, err := orderSvc.PlaceOrder("user-1", 99.99, "tok_visa")
	if err != nil {
		fmt.Printf("  ✗ FAIL: %v\n", err)
	} else {
		fmt.Printf("  ✓ Order created: %s, Status: %s\n", order.ID, order.Status)
	}

	// Verify mock was called correctly
	if len(mockPayment.ChargeCalls) == 1 {
		call := mockPayment.ChargeCalls[0]
		fmt.Printf("  ✓ Payment charged: $%.2f %s\n", call.Amount, call.Currency)
	}

	// Verify fake repo stored the order
	saved, _ := fakeRepo.FindByID(order.ID)
	if saved != nil {
		fmt.Printf("  ✓ Order saved in repository: %s\n", saved.PaymentID)
	}

	// Verify spy recorded calls
	fmt.Printf("  ✓ Repository calls: %v\n", fakeRepo.calls)

	// Test: payment failure
	fmt.Println("\n  Testing payment failure:")
	mockPayment.ChargeError = fmt.Errorf("card declined")
	_, err = orderSvc.PlaceOrder("user-2", 50.00, "tok_declined")
	if err != nil {
		fmt.Printf("  ✓ Expected error: %v\n", err)
	}

	// =========================================================================
	// DEMO 2: Contract Testing
	// =========================================================================
	fmt.Println("\n=== CONTRACT TESTING ===")

	// Consumer defines the contract
	contract := Contract{
		Consumer: "order-service",
		Provider: "user-service",
		Interactions: []ContractInteraction{
			{
				Description: "get user by ID",
				Request:     ContractRequest{Method: "GET", Path: "/users/1"},
				Response: ContractResponse{
					Status: 200,
					Body:   map[string]string{"id": "1", "name": "*", "email": "*"},
				},
			},
			{
				Description: "get non-existent user",
				Request:     ContractRequest{Method: "GET", Path: "/users/999"},
				Response: ContractResponse{
					Status: 404,
					Body:   map[string]string{"error": "*"},
				},
			},
		},
	}

	contractJSON, _ := json.MarshalIndent(contract, "  ", "  ")
	fmt.Printf("  Contract:\n  %s\n", contractJSON)

	// Provider verifies the contract
	fmt.Println("\n  Verifying contract against provider:")
	failures := VerifyContract(contract, func(method, path string) (int, map[string]string) {
		// Simulate the user-service provider
		if path == "/users/1" {
			return 200, map[string]string{"id": "1", "name": "Vikram", "email": "v@dev.com"}
		}
		if path == "/users/999" {
			return 404, map[string]string{"error": "user not found"}
		}
		return 500, nil
	})

	if len(failures) == 0 {
		fmt.Println("  ✓ All contract interactions verified!")
	} else {
		for _, f := range failures {
			fmt.Printf("  ✗ %s\n", f)
		}
	}

	// Test: contract broken by provider change
	fmt.Println("\n  Testing BROKEN contract (provider removed 'email' field):")
	failures = VerifyContract(contract, func(method, path string) (int, map[string]string) {
		if path == "/users/1" {
			return 200, map[string]string{"id": "1", "name": "Vikram"} // email removed!
		}
		return 404, map[string]string{"error": "not found"}
	})
	for _, f := range failures {
		fmt.Printf("  ✗ %s\n", f)
	}

	// =========================================================================
	// DEMO 3: Chaos Engineering Experiments
	// =========================================================================
	fmt.Println("\n=== CHAOS ENGINEERING ===")
	for _, exp := range chaosExperiments {
		fmt.Printf("  Experiment: %s\n", exp.Name)
		fmt.Printf("    Hypothesis: %s\n", exp.Hypothesis)
		fmt.Printf("    Injection:  %s\n", exp.Injection)
		fmt.Printf("    Duration:   %v\n", exp.Duration)
		fmt.Printf("    Rollback:   %s\n\n", exp.Rollback)
	}

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println("=== TESTING STRATEGY ===")
	fmt.Println("┌──────────────────────────┬─────────┬───────────┬──────────────────┐")
	fmt.Println("│ Type                     │ Speed   │ Count     │ Purpose          │")
	fmt.Println("├──────────────────────────┼─────────┼───────────┼──────────────────┤")
	fmt.Println("│ Unit Tests               │ <1ms    │ Hundreds  │ Business logic   │")
	fmt.Println("│ Integration Tests        │ 1-10s   │ Dozens    │ DB, cache, queue │")
	fmt.Println("│ Contract Tests           │ <1s     │ Per API   │ Service compat.  │")
	fmt.Println("│ Component Tests          │ 5-30s   │ Per svc   │ Full svc w/ stubs│")
	fmt.Println("│ E2E Tests                │ 1-5min  │ 5-10      │ Critical journeys│")
	fmt.Println("│ Chaos Tests              │ 1-5min  │ Periodic  │ Resilience       │")
	fmt.Println("└──────────────────────────┴─────────┴───────────┴──────────────────┘")
	fmt.Println()
	fmt.Println("KEY TOOLS:")
	fmt.Printf("  Unit:        %s\n", "Go testing, testify, mockgen")
	fmt.Printf("  Integration: %s\n", "testcontainers-go, docker-compose")
	fmt.Printf("  Contract:    %s\n", "Pact, Spring Cloud Contract")
	fmt.Printf("  E2E:         %s\n", "Playwright, Cypress, k6")
	fmt.Printf("  Chaos:       %s\n", "LitmusChaos, Gremlin, Toxiproxy")
	_ = strings.Join(nil, "") // imported
}
