package lowleveldesign

import (
	"fmt"
	"math"
)

// =============================================================================
// SOLID PRINCIPLES IN GO — The Foundation of All Good LLD
// =============================================================================
//
// SOLID is not about OOP — it's about managing DEPENDENCIES and CHANGE.
// Go implements these differently than Java/C++, but the principles are identical.
//
// S — Single Responsibility Principle
// O — Open/Closed Principle
// L — Liskov Substitution Principle
// I — Interface Segregation Principle
// D — Dependency Inversion Principle

// =============================================================================
// 1. SINGLE RESPONSIBILITY PRINCIPLE (SRP)
// =============================================================================
// "A struct should have only ONE reason to change."
//
// BAD: One struct doing everything = impossible to test, modify, or extend.
// GOOD: Each struct owns ONE concern.

// ❌ BAD — This struct has 3 reasons to change:
// 1. Order business logic changes
// 2. Persistence logic changes
// 3. Notification logic changes
type BadOrderService struct{}

func (s *BadOrderService) CreateOrder(item string, qty int) {
	// business logic
	fmt.Printf("Order created: %s x %d\n", item, qty)
	// persistence — WHY IS THIS HERE?
	fmt.Println("Saving to database...")
	// notification — WHY IS THIS HERE?
	fmt.Println("Sending email notification...")
}

// ✅ GOOD — Each struct has ONE job

// Order represents the domain entity — only business rules
type Order struct {
	ID    string
	Item  string
	Qty   int
	Total float64
}

func NewOrder(id, item string, qty int, price float64) *Order {
	return &Order{
		ID:    id,
		Item:  item,
		Qty:   qty,
		Total: float64(qty) * price,
	}
}

// OrderRepository handles persistence — only storage
type OrderRepository interface {
	Save(order *Order) error
	FindByID(id string) (*Order, error)
}

// OrderNotifier handles notifications — only messaging
type OrderNotifier interface {
	NotifyCreated(order *Order) error
}

// OrderService orchestrates — delegates to specialists
type OrderService struct {
	repo     OrderRepository
	notifier OrderNotifier
}

func NewOrderService(repo OrderRepository, notifier OrderNotifier) *OrderService {
	return &OrderService{repo: repo, notifier: notifier}
}

func (s *OrderService) CreateOrder(id, item string, qty int, price float64) (*Order, error) {
	order := NewOrder(id, item, qty, price)
	if err := s.repo.Save(order); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}
	if err := s.notifier.NotifyCreated(order); err != nil {
		// Log but don't fail — notification is secondary
		fmt.Printf("warning: notification failed: %v\n", err)
	}
	return order, nil
}

// KEY INSIGHT: Now you can:
// - Change DB without touching business logic
// - Change notification channel without touching DB
// - Test business logic with mock repo and notifier
// - Each piece is independently deployable

// =============================================================================
// 2. OPEN/CLOSED PRINCIPLE (OCP)
// =============================================================================
// "Open for EXTENSION, closed for MODIFICATION."
//
// You should be able to add NEW behavior without changing EXISTING code.
// In Go, this is achieved through INTERFACES and COMPOSITION.

// ❌ BAD — Adding a new shape requires modifying this function
func BadCalculateArea(shapeType string, dimensions ...float64) float64 {
	switch shapeType {
	case "circle":
		return math.Pi * dimensions[0] * dimensions[0]
	case "rectangle":
		return dimensions[0] * dimensions[1]
	// Every new shape = modify this function = risk breaking existing shapes
	default:
		return 0
	}
}

// ✅ GOOD — Interface allows infinite extension without modification

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

// Adding a Triangle? ZERO modification to existing code:
type Triangle struct {
	A, B, C float64 // sides
}

func (t Triangle) Area() float64 {
	s := (t.A + t.B + t.C) / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }

// This function works with ANY shape — past, present, or future
func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// =============================================================================
// 3. LISKOV SUBSTITUTION PRINCIPLE (LSP)
// =============================================================================
// "If S is a subtype of T, objects of type T can be replaced with objects of
//  type S without breaking the program."
//
// In Go terms: If your code accepts an interface, ANY implementation of that
// interface must work correctly without the caller knowing the concrete type.

// ❌ BAD — Square violates LSP for Rectangle behavior
// This is the classic example. A Square IS-A Rectangle mathematically,
// but in code, substituting Square for Rectangle breaks expectations.

type BadRectangle struct {
	width, height float64
}

func (r *BadRectangle) SetWidth(w float64)  { r.width = w }
func (r *BadRectangle) SetHeight(h float64) { r.height = h }
func (r *BadRectangle) Area() float64       { return r.width * r.height }

type BadSquare struct {
	BadRectangle
}

// This VIOLATES LSP — caller sets width, expects only width to change
func (s *BadSquare) SetWidth(w float64) {
	s.width = w
	s.height = w // SURPRISE! Height changes too!
}

func (s *BadSquare) SetHeight(h float64) {
	s.width = h
	s.height = h
}

// ✅ GOOD — Use interfaces that capture BEHAVIOR, not hierarchy

// Sizer captures what we actually need — computing dimensions
type Sizer interface {
	Area() float64
}

// Resizer captures the ability to resize
type Resizer interface {
	Resize(width, height float64)
}

type GoodRectangle struct {
	Width, Height float64
}

func (r *GoodRectangle) Area() float64 { return r.Width * r.Height }
func (r *GoodRectangle) Resize(w, h float64) {
	r.Width = w
	r.Height = h
}

type GoodSquare struct {
	Side float64
}

func (s *GoodSquare) Area() float64 { return s.Side * s.Side }

// Square doesn't implement Resizer — because you CAN'T independently resize sides
// This is honest. The TYPE SYSTEM now prevents misuse.

// KEY INSIGHT: In Go, LSP is about interface contracts.
// If you implement an interface, you MUST honor the contract.
// Don't implement an interface if you can't fulfill the behavioral promise.

// =============================================================================
// 4. INTERFACE SEGREGATION PRINCIPLE (ISP)
// =============================================================================
// "No client should be forced to depend on methods it doesn't use."
//
// Go's stdlib is the GOLD STANDARD of ISP:
// - io.Reader (1 method)
// - io.Writer (1 method)
// - io.Closer (1 method)
// - io.ReadWriter (composition of Reader + Writer)
//
// Compare with Java's bloated interfaces with 20+ methods.

// ❌ BAD — Fat interface forces implementors to fake methods
type BadAnimal interface {
	Walk()
	Swim()
	Fly()
	Speak()
	Eat()
	Sleep()
}

// A Dog would have to implement Fly() with a panic or no-op. TERRIBLE.

// ✅ GOOD — Small, focused interfaces

type Walker interface {
	Walk()
}

type Swimmer interface {
	Swim()
}

type Flyer interface {
	Fly()
}

type Speaker interface {
	Speak() string
}

// Compose when needed:
type LandAnimal interface {
	Walker
	Speaker
}

type Amphibian interface {
	Walker
	Swimmer
}

type Bird interface {
	Walker
	Flyer
}

// Dog only implements what it can actually do
type Dog struct{ Name string }

func (d Dog) Walk()         { fmt.Printf("%s is walking\n", d.Name) }
func (d Dog) Swim()         { fmt.Printf("%s is swimming\n", d.Name) }
func (d Dog) Speak() string { return "Woof!" }

// Dog satisfies Walker, Swimmer, Speaker, LandAnimal, Amphibian
// But NOT Flyer or Bird — which is correct!

// KEY INSIGHT: Define interfaces WHERE THEY ARE USED (consumer side),
// not where they are implemented (producer side).
// This is the Go way. Accept interfaces, return structs.

// =============================================================================
// 5. DEPENDENCY INVERSION PRINCIPLE (DIP)
// =============================================================================
// "High-level modules should not depend on low-level modules.
//  Both should depend on abstractions."
//
// This is THE most important principle for testable, maintainable code.

// ❌ BAD — High-level directly depends on low-level
type MySQLDatabase struct{}

func (db *MySQLDatabase) Query(sql string) string {
	return "mysql result"
}

type BadUserService struct {
	db *MySQLDatabase // CONCRETE dependency — can't swap, can't test
}

func (s *BadUserService) GetUser(id string) string {
	return s.db.Query("SELECT * FROM users WHERE id = " + id)
}

// ✅ GOOD — Depend on abstractions

// Database is the abstraction (defined by the consumer)
type Database interface {
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)
	Execute(query string, args ...interface{}) error
}

// UserService depends on the INTERFACE, not the implementation
type UserService struct {
	db Database
}

func NewUserService(db Database) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUser(id string) (map[string]interface{}, error) {
	results, err := s.db.Query("SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return results[0], nil
}

// Now you can use:
// - MySQLDB in production
// - SQLiteDB in integration tests
// - MockDB in unit tests
// - PostgresDB when you migrate
// ALL without changing UserService!

// =============================================================================
// PUTTING IT ALL TOGETHER — A Real Example
// =============================================================================
// Let's design a Payment Processing system using all SOLID principles.

// --- Interfaces (ISP: small, focused) ---

type PaymentProcessor interface {
	ProcessPayment(amount float64, currency string) (*PaymentResult, error)
}

type RefundProcessor interface {
	ProcessRefund(transactionID string, amount float64) error
}

type PaymentValidator interface {
	Validate(amount float64, currency string) error
}

// --- Domain types (SRP: only data and business rules) ---

type PaymentResult struct {
	TransactionID string
	Status        string
	Amount        float64
	Currency      string
}

// --- Implementations (OCP: add new processors without modifying existing) ---

type StripeProcessor struct {
	apiKey string
}

func NewStripeProcessor(apiKey string) *StripeProcessor {
	return &StripeProcessor{apiKey: apiKey}
}

func (s *StripeProcessor) ProcessPayment(amount float64, currency string) (*PaymentResult, error) {
	// In real code: call Stripe API
	return &PaymentResult{
		TransactionID: "stripe_txn_123",
		Status:        "success",
		Amount:        amount,
		Currency:      currency,
	}, nil
}

func (s *StripeProcessor) ProcessRefund(txnID string, amount float64) error {
	fmt.Printf("Refunding %f on Stripe txn %s\n", amount, txnID)
	return nil
}

type PayPalProcessor struct {
	clientID string
}

func NewPayPalProcessor(clientID string) *PayPalProcessor {
	return &PayPalProcessor{clientID: clientID}
}

func (p *PayPalProcessor) ProcessPayment(amount float64, currency string) (*PaymentResult, error) {
	return &PaymentResult{
		TransactionID: "paypal_txn_456",
		Status:        "success",
		Amount:        amount,
		Currency:      currency,
	}, nil
}

// PayPal doesn't implement RefundProcessor — ISP: it doesn't have to!

// --- Validator (SRP: only validation) ---

type AmountValidator struct {
	minAmount float64
	maxAmount float64
}

func NewAmountValidator(min, max float64) *AmountValidator {
	return &AmountValidator{minAmount: min, maxAmount: max}
}

func (v *AmountValidator) Validate(amount float64, currency string) error {
	if amount < v.minAmount {
		return fmt.Errorf("amount %f below minimum %f", amount, v.minAmount)
	}
	if amount > v.maxAmount {
		return fmt.Errorf("amount %f exceeds maximum %f", amount, v.maxAmount)
	}
	return nil
}

// --- Service (DIP: depends on abstractions, not concretions) ---

type PaymentService struct {
	processor PaymentProcessor
	validator PaymentValidator
	// Could add: logger, metrics, audit trail — all via interfaces
}

func NewPaymentService(p PaymentProcessor, v PaymentValidator) *PaymentService {
	return &PaymentService{processor: p, validator: v}
}

func (s *PaymentService) Pay(amount float64, currency string) (*PaymentResult, error) {
	// Validate first
	if err := s.validator.Validate(amount, currency); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	// Process payment
	result, err := s.processor.ProcessPayment(amount, currency)
	if err != nil {
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}
	return result, nil
}

// Usage:
func ExampleSOLID() {
	// Wire up with Stripe
	stripe := NewStripeProcessor("sk_test_xxx")
	validator := NewAmountValidator(0.50, 10000.00)
	service := NewPaymentService(stripe, validator)

	result, err := service.Pay(49.99, "USD")
	if err != nil {
		fmt.Printf("Payment failed: %v\n", err)
		return
	}
	fmt.Printf("Payment succeeded: %+v\n", result)

	// Switch to PayPal? Change ONE line:
	paypal := NewPayPalProcessor("client_xxx")
	service2 := NewPaymentService(paypal, validator)
	_ = service2

	// LSP: Both processors work identically from PaymentService's perspective.
	// The service doesn't know or care which processor it's using.
}
