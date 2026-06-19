package lowleveldesign

import (
	"fmt"
	"io"
	"strings"
)

// =============================================================================
// STRUCTURAL DESIGN PATTERNS IN GO
// =============================================================================
// These patterns deal with object COMPOSITION — how to assemble objects
// into larger structures while keeping them flexible and efficient.
//
// Patterns covered:
// 1. Adapter
// 2. Decorator
// 3. Proxy
// 4. Composite
// 5. Facade
// 6. Bridge

// =============================================================================
// 1. ADAPTER PATTERN
// =============================================================================
// "Convert the interface of a class into another interface clients expect."
//
// Real world: You have a 3rd-party library with interface X,
// but your code needs interface Y. Adapter bridges the gap.

// Our system's interface — what WE expect
type PaymentGateway interface {
	Charge(amount float64, currency string, cardToken string) (string, error)
}

// Third-party Stripe SDK (we can't modify this)
type StripeSDK struct{}

func (s *StripeSDK) CreateCharge(amountCents int64, cur string, source string) (*StripeCharge, error) {
	return &StripeCharge{
		ID:     "ch_stripe_" + source[:8],
		Status: "succeeded",
	}, nil
}

type StripeCharge struct {
	ID     string
	Status string
}

// Third-party Razorpay SDK (we can't modify this either)
type RazorpaySDK struct{}

func (r *RazorpaySDK) CapturePayment(amountPaise int, currency string, token string) (string, error) {
	return "pay_razorpay_" + token[:6], nil
}

// --- Adapters: bridge between our interface and third-party SDKs ---

type StripeAdapter struct {
	sdk *StripeSDK
}

func NewStripeAdapter(sdk *StripeSDK) *StripeAdapter {
	return &StripeAdapter{sdk: sdk}
}

func (a *StripeAdapter) Charge(amount float64, currency string, cardToken string) (string, error) {
	// Convert dollars to cents (Stripe's format)
	amountCents := int64(amount * 100)
	charge, err := a.sdk.CreateCharge(amountCents, currency, cardToken)
	if err != nil {
		return "", fmt.Errorf("stripe charge failed: %w", err)
	}
	return charge.ID, nil
}

type RazorpayAdapter struct {
	sdk *RazorpaySDK
}

func NewRazorpayAdapter(sdk *RazorpaySDK) *RazorpayAdapter {
	return &RazorpayAdapter{sdk: sdk}
}

func (a *RazorpayAdapter) Charge(amount float64, currency string, cardToken string) (string, error) {
	// Convert to paise (Razorpay's format)
	amountPaise := int(amount * 100)
	paymentID, err := a.sdk.CapturePayment(amountPaise, currency, cardToken)
	if err != nil {
		return "", fmt.Errorf("razorpay payment failed: %w", err)
	}
	return paymentID, nil
}

// Now our code works with ANY payment gateway through one interface:
func ProcessCheckout(gateway PaymentGateway, amount float64) {
	txnID, err := gateway.Charge(amount, "INR", "tok_visa_1234567890")
	if err != nil {
		fmt.Printf("Payment failed: %v\n", err)
		return
	}
	fmt.Printf("Payment successful: %s\n", txnID)
}

// =============================================================================
// 2. DECORATOR PATTERN
// =============================================================================
// "Attach additional responsibilities to an object dynamically."
//
// Go's io.Reader/io.Writer is the PERFECT example of this pattern.
// Each wrapper adds behavior while maintaining the same interface.

// --- Base interface ---
type DataSource interface {
	Read() (string, error)
	Write(data string) error
}

// --- Concrete implementation ---
type FileDataSource struct {
	filename string
	content  string
}

func NewFileDataSource(filename string) *FileDataSource {
	return &FileDataSource{filename: filename}
}

func (f *FileDataSource) Read() (string, error) {
	return f.content, nil
}

func (f *FileDataSource) Write(data string) error {
	f.content = data
	fmt.Printf("[File] Writing to %s: %s\n", f.filename, data)
	return nil
}

// --- Decorator: Encryption ---
type EncryptionDecorator struct {
	wrapped DataSource
	key     string
}

func NewEncryptionDecorator(source DataSource, key string) *EncryptionDecorator {
	return &EncryptionDecorator{wrapped: source, key: key}
}

func (e *EncryptionDecorator) Write(data string) error {
	encrypted := fmt.Sprintf("ENC[%s]{%s}", e.key[:3], data) // simplified
	fmt.Printf("[Encrypt] Encrypting data with key %s\n", e.key[:3])
	return e.wrapped.Write(encrypted)
}

func (e *EncryptionDecorator) Read() (string, error) {
	data, err := e.wrapped.Read()
	if err != nil {
		return "", err
	}
	// simplified decryption
	decrypted := strings.TrimPrefix(data, fmt.Sprintf("ENC[%s]{", e.key[:3]))
	decrypted = strings.TrimSuffix(decrypted, "}")
	fmt.Printf("[Decrypt] Decrypting data\n")
	return decrypted, nil
}

// --- Decorator: Compression ---
type CompressionDecorator struct {
	wrapped DataSource
}

func NewCompressionDecorator(source DataSource) *CompressionDecorator {
	return &CompressionDecorator{wrapped: source}
}

func (c *CompressionDecorator) Write(data string) error {
	compressed := fmt.Sprintf("GZIP{%s}", data) // simplified
	fmt.Printf("[Compress] Compressing %d bytes\n", len(data))
	return c.wrapped.Write(compressed)
}

func (c *CompressionDecorator) Read() (string, error) {
	data, err := c.wrapped.Read()
	if err != nil {
		return "", err
	}
	decompressed := strings.TrimPrefix(data, "GZIP{")
	decompressed = strings.TrimSuffix(decompressed, "}")
	fmt.Printf("[Decompress] Decompressing data\n")
	return decompressed, nil
}

// --- Decorator: Logging ---
type LoggingDecorator struct {
	wrapped DataSource
	label   string
}

func NewLoggingDecorator(source DataSource, label string) *LoggingDecorator {
	return &LoggingDecorator{wrapped: source, label: label}
}

func (l *LoggingDecorator) Write(data string) error {
	fmt.Printf("[LOG:%s] Write called with %d bytes\n", l.label, len(data))
	return l.wrapped.Write(data)
}

func (l *LoggingDecorator) Read() (string, error) {
	fmt.Printf("[LOG:%s] Read called\n", l.label)
	return l.wrapped.Read()
}

// --- Usage: Stack decorators like middleware! ---
func ExampleDecorator() {
	// Plain file
	source := NewFileDataSource("secret.dat")

	// Wrap with encryption, then compression, then logging
	// Order matters: logging -> compression -> encryption -> file
	var decorated DataSource = source
	decorated = NewEncryptionDecorator(decorated, "my-secret-key")
	decorated = NewCompressionDecorator(decorated)
	decorated = NewLoggingDecorator(decorated, "audit")

	// Write goes: log -> compress -> encrypt -> file
	decorated.Write("sensitive data here")

	// Read goes: log -> decompress -> decrypt -> file
	data, _ := decorated.Read()
	fmt.Println("Read back:", data)
}

// --- Real-world Go Decorator: io.Reader wrapping ---
type CountingReader struct {
	reader    io.Reader
	BytesRead int64
}

func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{reader: r}
}

func (c *CountingReader) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	c.BytesRead += int64(n)
	return n, err
}

// =============================================================================
// 3. PROXY PATTERN
// =============================================================================
// "Provide a surrogate/placeholder for another object to control access."
//
// Types: Virtual Proxy (lazy init), Protection Proxy (access control),
//        Caching Proxy, Logging Proxy

// --- Subject interface ---
type DatabaseAccess interface {
	Query(sql string) ([]string, error)
	Execute(sql string) error
}

// --- Real subject ---
type RealDatabase struct {
	connectionString string
}

func (db *RealDatabase) Query(sql string) ([]string, error) {
	fmt.Printf("[DB] Executing query: %s\n", sql)
	return []string{"row1", "row2"}, nil
}

func (db *RealDatabase) Execute(sql string) error {
	fmt.Printf("[DB] Executing: %s\n", sql)
	return nil
}

// --- Protection Proxy: Access control ---
type SecureDatabaseProxy struct {
	db       DatabaseAccess
	userRole string
}

func NewSecureDatabaseProxy(db DatabaseAccess, role string) *SecureDatabaseProxy {
	return &SecureDatabaseProxy{db: db, userRole: role}
}

func (p *SecureDatabaseProxy) Query(sql string) ([]string, error) {
	// Everyone can read
	fmt.Printf("[Proxy] User role '%s' executing query\n", p.userRole)
	return p.db.Query(sql)
}

func (p *SecureDatabaseProxy) Execute(sql string) error {
	// Only admins can write
	if p.userRole != "admin" {
		return fmt.Errorf("access denied: role '%s' cannot execute write operations", p.userRole)
	}
	fmt.Printf("[Proxy] Admin executing write operation\n")
	return p.db.Execute(sql)
}

// --- Caching Proxy ---
type CachingDatabaseProxy struct {
	db    DatabaseAccess
	cache map[string][]string
}

func NewCachingDatabaseProxy(db DatabaseAccess) *CachingDatabaseProxy {
	return &CachingDatabaseProxy{
		db:    db,
		cache: make(map[string][]string),
	}
}

func (p *CachingDatabaseProxy) Query(sql string) ([]string, error) {
	if result, found := p.cache[sql]; found {
		fmt.Printf("[Cache HIT] %s\n", sql)
		return result, nil
	}
	fmt.Printf("[Cache MISS] %s\n", sql)
	result, err := p.db.Query(sql)
	if err != nil {
		return nil, err
	}
	p.cache[sql] = result
	return result, nil
}

func (p *CachingDatabaseProxy) Execute(sql string) error {
	// Invalidate cache on writes
	p.cache = make(map[string][]string)
	return p.db.Execute(sql)
}

// =============================================================================
// 4. COMPOSITE PATTERN
// =============================================================================
// "Compose objects into tree structures and treat individual objects and
//  compositions uniformly."
//
// Classic examples: File systems, UI components, org charts, menus

// --- Component interface ---
type FileSystemNode interface {
	Name() string
	Size() int64
	Display(indent string)
	Search(name string) []FileSystemNode
}

// --- Leaf: File ---
type File struct {
	name string
	size int64
}

func NewFile(name string, size int64) *File {
	return &File{name: name, size: size}
}

func (f *File) Name() string { return f.name }
func (f *File) Size() int64  { return f.size }
func (f *File) Display(indent string) {
	fmt.Printf("%s📄 %s (%d bytes)\n", indent, f.name, f.size)
}
func (f *File) Search(name string) []FileSystemNode {
	if strings.Contains(f.name, name) {
		return []FileSystemNode{f}
	}
	return nil
}

// --- Composite: Directory ---
type Directory struct {
	name     string
	children []FileSystemNode
}

func NewDirectory(name string) *Directory {
	return &Directory{name: name}
}

func (d *Directory) Name() string { return d.name }

func (d *Directory) Size() int64 {
	var total int64
	for _, child := range d.children {
		total += child.Size() // Works recursively for nested dirs!
	}
	return total
}

func (d *Directory) Add(node FileSystemNode) {
	d.children = append(d.children, node)
}

func (d *Directory) Display(indent string) {
	fmt.Printf("%s📁 %s/ (%d bytes total)\n", indent, d.name, d.Size())
	for _, child := range d.children {
		child.Display(indent + "  ")
	}
}

func (d *Directory) Search(name string) []FileSystemNode {
	var results []FileSystemNode
	if strings.Contains(d.name, name) {
		results = append(results, d)
	}
	for _, child := range d.children {
		results = append(results, child.Search(name)...)
	}
	return results
}

func ExampleComposite() {
	root := NewDirectory("project")
	src := NewDirectory("src")
	src.Add(NewFile("main.go", 1200))
	src.Add(NewFile("handler.go", 3400))

	tests := NewDirectory("tests")
	tests.Add(NewFile("main_test.go", 800))

	root.Add(src)
	root.Add(tests)
	root.Add(NewFile("go.mod", 150))

	root.Display("")
	// 📁 project/ (5550 bytes total)
	//   📁 src/ (4600 bytes total)
	//     📄 main.go (1200 bytes)
	//     📄 handler.go (3400 bytes)
	//   📁 tests/ (800 bytes total)
	//     📄 main_test.go (800 bytes)
	//   📄 go.mod (150 bytes)
}

// =============================================================================
// 5. FACADE PATTERN
// =============================================================================
// "Provide a simplified interface to a complex subsystem."
//
// When you have multiple subsystems that need to coordinate,
// expose ONE simple API that handles the orchestration.

// --- Complex subsystems ---
type InventorySystem struct{}

func (i *InventorySystem) CheckStock(productID string) int {
	return 42 // simulated stock
}
func (i *InventorySystem) ReserveStock(productID string, qty int) error {
	fmt.Printf("[Inventory] Reserved %d of %s\n", qty, productID)
	return nil
}

type PaymentSystem struct{}

func (p *PaymentSystem) CreatePaymentIntent(amount float64) string {
	return "pi_abc123"
}
func (p *PaymentSystem) ConfirmPayment(intentID string) error {
	fmt.Printf("[Payment] Confirmed %s\n", intentID)
	return nil
}

type ShippingSystem struct{}

func (s *ShippingSystem) CalculateShipping(address string, weight float64) float64 {
	return 5.99
}
func (s *ShippingSystem) CreateShipment(orderID, address string) string {
	return "TRACK123"
}

type NotificationSystem struct{}

func (n *NotificationSystem) SendOrderConfirmation(email, orderID string) {
	fmt.Printf("[Notification] Order %s confirmation sent to %s\n", orderID, email)
}

// --- Facade: One simple interface for the entire checkout flow ---
type CheckoutFacade struct {
	inventory    *InventorySystem
	payment      *PaymentSystem
	shipping     *ShippingSystem
	notification *NotificationSystem
}

func NewCheckoutFacade() *CheckoutFacade {
	return &CheckoutFacade{
		inventory:    &InventorySystem{},
		payment:      &PaymentSystem{},
		shipping:     &ShippingSystem{},
		notification: &NotificationSystem{},
	}
}

type CheckoutRequest struct {
	ProductID string
	Quantity  int
	Amount    float64
	Address   string
	Email     string
}

type CheckoutResult struct {
	OrderID    string
	TrackingID string
	Total      float64
}

// ONE method handles the entire complex flow
func (f *CheckoutFacade) PlaceOrder(req CheckoutRequest) (*CheckoutResult, error) {
	// 1. Check stock
	stock := f.inventory.CheckStock(req.ProductID)
	if stock < req.Quantity {
		return nil, fmt.Errorf("insufficient stock: have %d, need %d", stock, req.Quantity)
	}

	// 2. Reserve inventory
	if err := f.inventory.ReserveStock(req.ProductID, req.Quantity); err != nil {
		return nil, fmt.Errorf("reserve stock: %w", err)
	}

	// 3. Calculate total with shipping
	shippingCost := f.shipping.CalculateShipping(req.Address, 1.0)
	total := req.Amount + shippingCost

	// 4. Process payment
	intentID := f.payment.CreatePaymentIntent(total)
	if err := f.payment.ConfirmPayment(intentID); err != nil {
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	// 5. Create shipment
	orderID := "ORD-001"
	trackingID := f.shipping.CreateShipment(orderID, req.Address)

	// 6. Send notification
	f.notification.SendOrderConfirmation(req.Email, orderID)

	return &CheckoutResult{
		OrderID:    orderID,
		TrackingID: trackingID,
		Total:      total,
	}, nil
}

// =============================================================================
// 6. BRIDGE PATTERN
// =============================================================================
// "Decouple an abstraction from its implementation so they can vary
//  independently."
//
// When you have TWO dimensions of variation (e.g., shape + rendering,
// notification + channel, message + format).

// Dimension 1: Message format
type MessageFormatter interface {
	Format(title, body string) string
}

type PlainTextFormatter struct{}

func (f *PlainTextFormatter) Format(title, body string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s", title, strings.Repeat("-", len(title)), body, "")
}

type HTMLFormatter struct{}

func (f *HTMLFormatter) Format(title, body string) string {
	return fmt.Sprintf("<h1>%s</h1><p>%s</p>", title, body)
}

type JSONFormatter struct{}

func (f *JSONFormatter) Format(title, body string) string {
	return fmt.Sprintf(`{"title":"%s","body":"%s"}`, title, body)
}

// Dimension 2: Delivery channel
type MessageSender interface {
	Send(recipient string, content string) error
}

type EmailSender struct{ smtpHost string }

func (e *EmailSender) Send(recipient string, content string) error {
	fmt.Printf("[Email -> %s] %s\n", recipient, content)
	return nil
}

type SlackSender struct{ webhook string }

func (s *SlackSender) Send(recipient string, content string) error {
	fmt.Printf("[Slack -> %s] %s\n", recipient, content)
	return nil
}

// Bridge: Combines format + channel without explosion of classes
type NotificationBridge struct {
	formatter MessageFormatter
	sender    MessageSender
}

func NewNotificationBridge(formatter MessageFormatter, sender MessageSender) *NotificationBridge {
	return &NotificationBridge{formatter: formatter, sender: sender}
}

func (n *NotificationBridge) Notify(recipient, title, body string) error {
	content := n.formatter.Format(title, body)
	return n.sender.Send(recipient, content)
}

// Without Bridge: You'd need EmailPlainText, EmailHTML, EmailJSON,
// SlackPlainText, SlackHTML, SlackJSON = 6 classes (and growing exponentially!)
// With Bridge: 3 formatters + 2 senders = 5 structs that compose into 6 combinations.

func ExampleBridge() {
	// HTML email
	n1 := NewNotificationBridge(&HTMLFormatter{}, &EmailSender{smtpHost: "smtp.example.com"})
	n1.Notify("user@example.com", "Welcome!", "Thanks for signing up")

	// JSON slack message
	n2 := NewNotificationBridge(&JSONFormatter{}, &SlackSender{webhook: "https://hooks.slack.com/xxx"})
	n2.Notify("#alerts", "Deploy Complete", "v2.1.0 deployed successfully")
}
