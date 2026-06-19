package lowleveldesign

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// =============================================================================
// BEHAVIORAL DESIGN PATTERNS IN GO
// =============================================================================
// These patterns define HOW objects communicate and distribute responsibility.
//
// Patterns covered:
// 1. Strategy
// 2. Observer
// 3. Command
// 4. State
// 5. Chain of Responsibility
// 6. Template Method
// 7. Iterator

// =============================================================================
// 1. STRATEGY PATTERN
// =============================================================================
// "Define a family of algorithms, encapsulate each one, and make them
//  interchangeable."
//
// In Go: Strategy is just an interface with different implementations.
// Bonus: Functions as strategies (no need for a struct if it's one method).

// --- Strategy interface ---
type SortStrategy interface {
	Sort(data []int) []int
	Name() string
}

// --- Concrete strategies ---
type BubbleSort struct{}

func (b *BubbleSort) Name() string { return "BubbleSort" }
func (b *BubbleSort) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	n := len(result)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

type QuickSort struct{}

func (q *QuickSort) Name() string { return "QuickSort" }
func (q *QuickSort) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	sort.Ints(result) // using stdlib for simplicity
	return result
}

// --- Context uses strategy ---
type Sorter struct {
	strategy SortStrategy
}

func NewSorter(strategy SortStrategy) *Sorter {
	return &Sorter{strategy: strategy}
}

func (s *Sorter) SetStrategy(strategy SortStrategy) {
	s.strategy = strategy
}

func (s *Sorter) Sort(data []int) []int {
	fmt.Printf("Sorting with %s\n", s.strategy.Name())
	return s.strategy.Sort(data)
}

// --- Go-idiomatic: Function as strategy ---
type SortFunc func([]int) []int

type FuncSorter struct {
	sortFn SortFunc
}

func NewFuncSorter(fn SortFunc) *FuncSorter {
	return &FuncSorter{sortFn: fn}
}

func (s *FuncSorter) Sort(data []int) []int {
	return s.sortFn(data)
}

// --- Pricing strategy (more practical example) ---
type PricingStrategy interface {
	Calculate(basePrice float64) float64
	Description() string
}

type RegularPricing struct{}

func (p *RegularPricing) Calculate(base float64) float64 { return base }
func (p *RegularPricing) Description() string            { return "Regular" }

type PremiumDiscount struct{ discountPct float64 }

func NewPremiumDiscount(pct float64) *PremiumDiscount     { return &PremiumDiscount{discountPct: pct} }
func (p *PremiumDiscount) Calculate(base float64) float64 { return base * (1 - p.discountPct/100) }
func (p *PremiumDiscount) Description() string {
	return fmt.Sprintf("Premium (%.0f%% off)", p.discountPct)
}

type BlackFridayPricing struct{}

func (p *BlackFridayPricing) Calculate(base float64) float64 { return base * 0.5 }
func (p *BlackFridayPricing) Description() string            { return "Black Friday (50% off)" }

// =============================================================================
// 2. OBSERVER PATTERN (Pub/Sub)
// =============================================================================
// "Define a one-to-many dependency so that when one object changes state,
//  all dependents are notified."
//
// Go twist: Use channels for async notification. Much cleaner than callbacks.

// --- Event types ---
type Event struct {
	Type    string
	Payload interface{}
	Time    time.Time
}

// --- Observer interface ---
type EventHandler interface {
	Handle(event Event)
	ID() string
}

// --- Subject (Event Bus) ---
type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler // eventType -> handlers
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Unsubscribe(eventType string, handlerID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	handlers := eb.handlers[eventType]
	for i, h := range handlers {
		if h.ID() == handlerID {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return
		}
	}
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	handlers := eb.handlers[event.Type]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		handler.Handle(event) // sync notification
	}
}

// Async publish with channels
func (eb *EventBus) PublishAsync(event Event) {
	eb.mu.RLock()
	handlers := eb.handlers[event.Type]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		go handler.Handle(event)
	}
}

// --- Concrete observers ---
type LoggingObserver struct{ id string }

func (o *LoggingObserver) ID() string { return o.id }
func (o *LoggingObserver) Handle(event Event) {
	fmt.Printf("[LOG] Event: %s at %v\n", event.Type, event.Time.Format(time.RFC3339))
}

type MetricsObserver struct{ id string }

func (o *MetricsObserver) ID() string { return o.id }
func (o *MetricsObserver) Handle(event Event) {
	fmt.Printf("[METRICS] Recording event: %s\n", event.Type)
}

type EmailObserver struct {
	id    string
	email string
}

func (o *EmailObserver) ID() string { return o.id }
func (o *EmailObserver) Handle(event Event) {
	fmt.Printf("[EMAIL] Notifying %s about: %s\n", o.email, event.Type)
}

func ExampleObserver() {
	bus := NewEventBus()

	// Subscribe observers
	bus.Subscribe("user.created", &LoggingObserver{id: "logger"})
	bus.Subscribe("user.created", &EmailObserver{id: "welcome-email", email: "admin@example.com"})
	bus.Subscribe("order.placed", &MetricsObserver{id: "metrics"})
	bus.Subscribe("order.placed", &LoggingObserver{id: "logger"})

	// Publish events
	bus.Publish(Event{Type: "user.created", Payload: "user-123", Time: time.Now()})
	bus.Publish(Event{Type: "order.placed", Payload: "order-456", Time: time.Now()})
}

// =============================================================================
// 3. COMMAND PATTERN
// =============================================================================
// "Encapsulate a request as an object, allowing you to parameterize clients,
//  queue requests, and support undo operations."
//
// Perfect for: undo/redo, task queues, transaction logs, macro recording.

// --- Command interface ---
type Command interface {
	Execute() error
	Undo() error
	Description() string
}

// --- Receiver ---
type TextEditor struct {
	content string
	cursor  int
}

func NewTextEditor() *TextEditor {
	return &TextEditor{}
}

func (e *TextEditor) Content() string { return e.content }

// --- Concrete commands ---
type InsertCommand struct {
	editor   *TextEditor
	position int
	text     string
}

func NewInsertCommand(editor *TextEditor, position int, text string) *InsertCommand {
	return &InsertCommand{editor: editor, position: position, text: text}
}

func (c *InsertCommand) Execute() error {
	if c.position > len(c.editor.content) {
		c.position = len(c.editor.content)
	}
	c.editor.content = c.editor.content[:c.position] + c.text + c.editor.content[c.position:]
	return nil
}

func (c *InsertCommand) Undo() error {
	c.editor.content = c.editor.content[:c.position] + c.editor.content[c.position+len(c.text):]
	return nil
}

func (c *InsertCommand) Description() string {
	return fmt.Sprintf("Insert '%s' at position %d", c.text, c.position)
}

type DeleteCommand struct {
	editor      *TextEditor
	position    int
	length      int
	deletedText string // saved for undo
}

func NewDeleteCommand(editor *TextEditor, position, length int) *DeleteCommand {
	return &DeleteCommand{editor: editor, position: position, length: length}
}

func (c *DeleteCommand) Execute() error {
	end := c.position + c.length
	if end > len(c.editor.content) {
		end = len(c.editor.content)
	}
	c.deletedText = c.editor.content[c.position:end]
	c.editor.content = c.editor.content[:c.position] + c.editor.content[end:]
	return nil
}

func (c *DeleteCommand) Undo() error {
	c.editor.content = c.editor.content[:c.position] + c.deletedText + c.editor.content[c.position:]
	return nil
}

func (c *DeleteCommand) Description() string {
	return fmt.Sprintf("Delete %d chars at position %d", c.length, c.position)
}

// --- Invoker with history (undo/redo support) ---
type CommandHistory struct {
	undoStack []Command
	redoStack []Command
}

func NewCommandHistory() *CommandHistory {
	return &CommandHistory{}
}

func (h *CommandHistory) Execute(cmd Command) error {
	if err := cmd.Execute(); err != nil {
		return err
	}
	h.undoStack = append(h.undoStack, cmd)
	h.redoStack = nil // clear redo after new command
	return nil
}

func (h *CommandHistory) Undo() error {
	if len(h.undoStack) == 0 {
		return fmt.Errorf("nothing to undo")
	}
	cmd := h.undoStack[len(h.undoStack)-1]
	h.undoStack = h.undoStack[:len(h.undoStack)-1]
	if err := cmd.Undo(); err != nil {
		return err
	}
	h.redoStack = append(h.redoStack, cmd)
	return nil
}

func (h *CommandHistory) Redo() error {
	if len(h.redoStack) == 0 {
		return fmt.Errorf("nothing to redo")
	}
	cmd := h.redoStack[len(h.redoStack)-1]
	h.redoStack = h.redoStack[:len(h.redoStack)-1]
	if err := cmd.Execute(); err != nil {
		return err
	}
	h.undoStack = append(h.undoStack, cmd)
	return nil
}

func ExampleCommand() {
	editor := NewTextEditor()
	history := NewCommandHistory()

	history.Execute(NewInsertCommand(editor, 0, "Hello"))
	history.Execute(NewInsertCommand(editor, 5, " World"))
	fmt.Println(editor.Content()) // "Hello World"

	history.Undo()
	fmt.Println(editor.Content()) // "Hello"

	history.Redo()
	fmt.Println(editor.Content()) // "Hello World"

	history.Execute(NewDeleteCommand(editor, 5, 6))
	fmt.Println(editor.Content()) // "Hello"
}

// =============================================================================
// 4. STATE PATTERN
// =============================================================================
// "Allow an object to alter its behavior when its internal state changes."
//
// The object appears to change its class. Great for: vending machines,
// order processing, document workflows, TCP connections.

// --- State interface ---
type OrderState interface {
	Next(order *OnlineOrder) error
	Cancel(order *OnlineOrder) error
	String() string
}

// --- Context ---
type OnlineOrder struct {
	ID      string
	state   OrderState
	Items   []string
	history []string
}

func NewOnlineOrder(id string) *OnlineOrder {
	order := &OnlineOrder{
		ID: id,
	}
	order.SetState(&DraftState{})
	return order
}

func (o *OnlineOrder) SetState(state OrderState) {
	o.state = state
	o.history = append(o.history, state.String())
}

func (o *OnlineOrder) Next() error    { return o.state.Next(o) }
func (o *OnlineOrder) Cancel() error  { return o.state.Cancel(o) }
func (o *OnlineOrder) Status() string { return o.state.String() }

// --- Concrete states ---
type DraftState struct{}

func (s *DraftState) String() string { return "DRAFT" }
func (s *DraftState) Next(order *OnlineOrder) error {
	if len(order.Items) == 0 {
		return fmt.Errorf("cannot submit empty order")
	}
	fmt.Println("Order submitted for processing")
	order.SetState(&PendingState{})
	return nil
}
func (s *DraftState) Cancel(order *OnlineOrder) error {
	fmt.Println("Draft order discarded")
	order.SetState(&CancelledState{})
	return nil
}

type PendingState struct{}

func (s *PendingState) String() string { return "PENDING" }
func (s *PendingState) Next(order *OnlineOrder) error {
	fmt.Println("Payment confirmed, order processing")
	order.SetState(&ProcessingState{})
	return nil
}
func (s *PendingState) Cancel(order *OnlineOrder) error {
	fmt.Println("Pending order cancelled")
	order.SetState(&CancelledState{})
	return nil
}

type ProcessingState struct{}

func (s *ProcessingState) String() string { return "PROCESSING" }
func (s *ProcessingState) Next(order *OnlineOrder) error {
	fmt.Println("Order shipped!")
	order.SetState(&ShippedState{})
	return nil
}
func (s *ProcessingState) Cancel(order *OnlineOrder) error {
	return fmt.Errorf("cannot cancel order that is already being processed")
}

type ShippedState struct{}

func (s *ShippedState) String() string { return "SHIPPED" }
func (s *ShippedState) Next(order *OnlineOrder) error {
	fmt.Println("Order delivered!")
	order.SetState(&DeliveredState{})
	return nil
}
func (s *ShippedState) Cancel(order *OnlineOrder) error {
	return fmt.Errorf("cannot cancel shipped order")
}

type DeliveredState struct{}

func (s *DeliveredState) String() string { return "DELIVERED" }
func (s *DeliveredState) Next(order *OnlineOrder) error {
	return fmt.Errorf("order already delivered — no next state")
}
func (s *DeliveredState) Cancel(order *OnlineOrder) error {
	return fmt.Errorf("cannot cancel delivered order")
}

type CancelledState struct{}

func (s *CancelledState) String() string { return "CANCELLED" }
func (s *CancelledState) Next(order *OnlineOrder) error {
	return fmt.Errorf("cancelled order cannot proceed")
}
func (s *CancelledState) Cancel(order *OnlineOrder) error {
	return fmt.Errorf("order already cancelled")
}

func ExampleState() {
	order := NewOnlineOrder("ORD-001")
	order.Items = []string{"Laptop", "Mouse"}

	fmt.Println(order.Status()) // DRAFT
	order.Next()                // -> PENDING
	order.Next()                // -> PROCESSING
	order.Next()                // -> SHIPPED
	order.Next()                // -> DELIVERED
	err := order.Next()         // ERROR: already delivered
	fmt.Println(err)
}

// =============================================================================
// 5. CHAIN OF RESPONSIBILITY
// =============================================================================
// "Pass a request along a chain of handlers. Each handler decides whether
//  to process it or pass it to the next handler."
//
// Perfect for: middleware, logging pipelines, validation chains, approval flows.

// --- Handler interface ---
type Middleware interface {
	Handle(req *HTTPReq, next Middleware) *HTTPResp
}

type HTTPReq struct {
	Path    string
	Method  string
	Headers map[string]string
	Body    string
	UserID  string
}

type HTTPResp struct {
	Status int
	Body   string
}

// --- Concrete handlers ---
type AuthMiddleware struct {
	validTokens map[string]string // token -> userID
}

func NewAuthMiddleware(tokens map[string]string) *AuthMiddleware {
	return &AuthMiddleware{validTokens: tokens}
}

func (m *AuthMiddleware) Handle(req *HTTPReq, next Middleware) *HTTPResp {
	token := req.Headers["Authorization"]
	if token == "" {
		return &HTTPResp{Status: 401, Body: "missing authorization"}
	}
	userID, valid := m.validTokens[token]
	if !valid {
		return &HTTPResp{Status: 403, Body: "invalid token"}
	}
	req.UserID = userID
	fmt.Printf("[Auth] Authenticated user: %s\n", userID)
	return next.Handle(req, nil)
}

type RateLimitMiddleware struct {
	mu       sync.Mutex
	requests map[string]int
	limit    int
}

func NewRateLimitMiddleware(limit int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		requests: make(map[string]int),
		limit:    limit,
	}
}

func (m *RateLimitMiddleware) Handle(req *HTTPReq, next Middleware) *HTTPResp {
	m.mu.Lock()
	m.requests[req.UserID]++
	count := m.requests[req.UserID]
	m.mu.Unlock()

	if count > m.limit {
		return &HTTPResp{Status: 429, Body: "rate limit exceeded"}
	}
	fmt.Printf("[RateLimit] User %s: %d/%d requests\n", req.UserID, count, m.limit)
	return next.Handle(req, nil)
}

type LoggingMiddleware struct{}

func (m *LoggingMiddleware) Handle(req *HTTPReq, next Middleware) *HTTPResp {
	start := time.Now()
	fmt.Printf("[Log] %s %s started\n", req.Method, req.Path)
	resp := next.Handle(req, nil)
	fmt.Printf("[Log] %s %s completed in %v (status: %d)\n",
		req.Method, req.Path, time.Since(start), resp.Status)
	return resp
}

// --- Chain builder ---
type MiddlewareChain struct {
	middlewares []Middleware
	handler     func(*HTTPReq) *HTTPResp
}

func NewMiddlewareChain(handler func(*HTTPReq) *HTTPResp) *MiddlewareChain {
	return &MiddlewareChain{handler: handler}
}

func (c *MiddlewareChain) Use(m Middleware) {
	c.middlewares = append(c.middlewares, m)
}

func (c *MiddlewareChain) Handle(req *HTTPReq, _ Middleware) *HTTPResp {
	// Build chain from inside out
	var final Middleware = middlewareFunc(func(req *HTTPReq, _ Middleware) *HTTPResp {
		return c.handler(req)
	})

	// Wrap from last to first
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		mw := c.middlewares[i]
		nextMw := final
		final = middlewareFunc(func(req *HTTPReq, _ Middleware) *HTTPResp {
			return mw.Handle(req, nextMw)
		})
	}

	return final.Handle(req, nil)
}

// Helper to convert a function into Middleware interface
type middlewareFunc func(*HTTPReq, Middleware) *HTTPResp

func (f middlewareFunc) Handle(req *HTTPReq, next Middleware) *HTTPResp {
	return f(req, next)
}

// =============================================================================
// 6. TEMPLATE METHOD PATTERN
// =============================================================================
// "Define the skeleton of an algorithm, deferring some steps to subclasses."
//
// In Go: Use interface embedding or function fields (no abstract classes).

// --- Template with hooks ---
type DataPipeline interface {
	Extract() ([]byte, error)
	Transform(data []byte) ([]byte, error)
	Load(data []byte) error
}

// The template function — algorithm skeleton
func RunETLPipeline(ctx context.Context, pipeline DataPipeline) error {
	fmt.Println("=== Starting ETL Pipeline ===")

	// Step 1: Extract
	fmt.Println("Step 1: Extracting data...")
	data, err := pipeline.Extract()
	if err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	// Check context cancellation between steps
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Step 2: Transform
	fmt.Println("Step 2: Transforming data...")
	transformed, err := pipeline.Transform(data)
	if err != nil {
		return fmt.Errorf("transform failed: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Step 3: Load
	fmt.Println("Step 3: Loading data...")
	if err := pipeline.Load(transformed); err != nil {
		return fmt.Errorf("load failed: %w", err)
	}

	fmt.Println("=== Pipeline Complete ===")
	return nil
}

// Concrete pipeline: CSV to Database
type CSVToDBPipeline struct {
	csvPath string
	dbConn  string
}

func (p *CSVToDBPipeline) Extract() ([]byte, error) {
	fmt.Printf("  Reading CSV from: %s\n", p.csvPath)
	return []byte("name,age\nAlice,30\nBob,25"), nil
}

func (p *CSVToDBPipeline) Transform(data []byte) ([]byte, error) {
	fmt.Printf("  Transforming %d bytes of CSV data\n", len(data))
	return []byte("INSERT INTO users VALUES ('Alice',30),('Bob',25)"), nil
}

func (p *CSVToDBPipeline) Load(data []byte) error {
	fmt.Printf("  Executing on %s: %s\n", p.dbConn, string(data))
	return nil
}

// Concrete pipeline: API to S3
type APIToS3Pipeline struct {
	apiURL string
	bucket string
}

func (p *APIToS3Pipeline) Extract() ([]byte, error) {
	fmt.Printf("  Fetching from API: %s\n", p.apiURL)
	return []byte(`[{"id":1},{"id":2}]`), nil
}

func (p *APIToS3Pipeline) Transform(data []byte) ([]byte, error) {
	fmt.Printf("  Converting %d bytes to parquet format\n", len(data))
	return data, nil
}

func (p *APIToS3Pipeline) Load(data []byte) error {
	fmt.Printf("  Uploading %d bytes to s3://%s/data.parquet\n", len(data), p.bucket)
	return nil
}

// =============================================================================
// 7. ITERATOR PATTERN
// =============================================================================
// "Provide a way to access elements of a collection sequentially without
//  exposing the underlying representation."
//
// Go 1.22+ has range-over-func. Before that, use channels or callback iterators.

// --- Classic iterator ---
type Iterator[T any] interface {
	HasNext() bool
	Next() T
	Reset()
}

// --- Collection that produces iterators ---
type TreeNode struct {
	Value int
	Left  *TreeNode
	Right *TreeNode
}

// In-order iterator for binary tree
type InOrderIterator struct {
	stack   []*TreeNode
	current *TreeNode
}

func NewInOrderIterator(root *TreeNode) *InOrderIterator {
	it := &InOrderIterator{current: root}
	return it
}

func (it *InOrderIterator) HasNext() bool {
	return it.current != nil || len(it.stack) > 0
}

func (it *InOrderIterator) Next() int {
	for it.current != nil {
		it.stack = append(it.stack, it.current)
		it.current = it.current.Left
	}
	node := it.stack[len(it.stack)-1]
	it.stack = it.stack[:len(it.stack)-1]
	it.current = node.Right
	return node.Value
}

// --- Go-idiomatic: Channel-based iterator ---
func InOrderTraversal(root *TreeNode) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		var traverse func(*TreeNode)
		traverse = func(node *TreeNode) {
			if node == nil {
				return
			}
			traverse(node.Left)
			ch <- node.Value
			traverse(node.Right)
		}
		traverse(root)
	}()
	return ch
}

// --- Go-idiomatic: Callback iterator (zero allocation) ---
func InOrderWalk(root *TreeNode, visit func(int) bool) {
	if root == nil {
		return
	}
	InOrderWalk(root.Left, visit)
	if !visit(root.Value) {
		return
	}
	InOrderWalk(root.Right, visit)
}

func ExampleIterator() {
	tree := &TreeNode{
		Value: 5,
		Left: &TreeNode{
			Value: 3,
			Left:  &TreeNode{Value: 1},
			Right: &TreeNode{Value: 4},
		},
		Right: &TreeNode{
			Value: 7,
			Left:  &TreeNode{Value: 6},
			Right: &TreeNode{Value: 9},
		},
	}

	// Channel-based
	fmt.Print("In-order: ")
	for val := range InOrderTraversal(tree) {
		fmt.Printf("%d ", val)
	}
	fmt.Println()

	// Callback-based (find first > 5)
	fmt.Print("First > 5: ")
	InOrderWalk(tree, func(v int) bool {
		if v > 5 {
			fmt.Println(v)
			return false // stop iteration
		}
		return true
	})
}
