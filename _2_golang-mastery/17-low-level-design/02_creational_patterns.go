package lowleveldesign

import (
	"fmt"
	"sync"
)

// =============================================================================
// CREATIONAL DESIGN PATTERNS IN GO
// =============================================================================
// These patterns control HOW objects are created.
// In Go, we don't have constructors — we have factory functions, builders,
// and some clever uses of closures and sync.Once.
//
// Patterns covered:
// 1. Factory Method
// 2. Abstract Factory
// 3. Builder
// 4. Singleton
// 5. Prototype
// 6. Object Pool

// =============================================================================
// 1. FACTORY METHOD PATTERN
// =============================================================================
// "Define an interface for creating an object, but let subclasses decide
//  which class to instantiate."
//
// In Go: A function that returns an interface based on input parameters.
// This is the MOST common pattern in Go codebases.

// --- Product interface ---
type Notification interface {
	Send(to, message string) error
	GetType() string
}

// --- Concrete products ---
type EmailNotification struct {
	from     string
	smtpHost string
}

func (e *EmailNotification) Send(to, message string) error {
	fmt.Printf("[EMAIL] From: %s, To: %s, Message: %s (via %s)\n",
		e.from, to, message, e.smtpHost)
	return nil
}

func (e *EmailNotification) GetType() string { return "email" }

type SMSNotification struct {
	phoneNumber string
	provider    string
}

func (s *SMSNotification) Send(to, message string) error {
	fmt.Printf("[SMS] From: %s, To: %s, Message: %s (via %s)\n",
		s.phoneNumber, to, message, s.provider)
	return nil
}

func (s *SMSNotification) GetType() string { return "sms" }

type PushNotification struct {
	appID string
}

func (p *PushNotification) Send(to, message string) error {
	fmt.Printf("[PUSH] App: %s, To: %s, Message: %s\n", p.appID, to, message)
	return nil
}

func (p *PushNotification) GetType() string { return "push" }

// --- Factory function ---
// This is the Go-idiomatic factory. No need for factory classes.
func NewNotification(notifType string, config map[string]string) (Notification, error) {
	switch notifType {
	case "email":
		return &EmailNotification{
			from:     config["from"],
			smtpHost: config["smtp_host"],
		}, nil
	case "sms":
		return &SMSNotification{
			phoneNumber: config["phone"],
			provider:    config["provider"],
		}, nil
	case "push":
		return &PushNotification{
			appID: config["app_id"],
		}, nil
	default:
		return nil, fmt.Errorf("unknown notification type: %s", notifType)
	}
}

// --- Registry-based Factory (more extensible) ---
// This approach follows OCP — add new types without modifying factory code.

type NotificationFactory func(config map[string]string) Notification

var notificationRegistry = make(map[string]NotificationFactory)

func RegisterNotification(name string, factory NotificationFactory) {
	notificationRegistry[name] = factory
}

func CreateNotification(name string, config map[string]string) (Notification, error) {
	factory, exists := notificationRegistry[name]
	if !exists {
		return nil, fmt.Errorf("notification type %q not registered", name)
	}
	return factory(config), nil
}

// Register at init time — each package self-registers
func init() {
	RegisterNotification("email", func(config map[string]string) Notification {
		return &EmailNotification{from: config["from"], smtpHost: config["smtp_host"]}
	})
	RegisterNotification("sms", func(config map[string]string) Notification {
		return &SMSNotification{phoneNumber: config["phone"], provider: config["provider"]}
	})
}

// =============================================================================
// 2. ABSTRACT FACTORY PATTERN
// =============================================================================
// "Provide an interface for creating FAMILIES of related objects."
//
// When you need to create multiple related objects that must be compatible.

// Scenario: UI toolkit that supports different themes

// --- Abstract products ---
type Button interface {
	Render() string
	OnClick(handler func())
}

type TextInput interface {
	Render() string
	SetValue(val string)
}

type Modal interface {
	Render() string
	Show()
	Hide()
}

// --- Abstract factory ---
type UIFactory interface {
	CreateButton(label string) Button
	CreateTextInput(placeholder string) TextInput
	CreateModal(title string) Modal
}

// --- Dark Theme Family ---
type DarkButton struct{ label string }

func (b *DarkButton) Render() string         { return fmt.Sprintf("[DARK BTN: %s]", b.label) }
func (b *DarkButton) OnClick(handler func()) { handler() }

type DarkTextInput struct{ placeholder string }

func (t *DarkTextInput) Render() string      { return fmt.Sprintf("[DARK INPUT: %s]", t.placeholder) }
func (t *DarkTextInput) SetValue(val string) { fmt.Printf("Dark input set: %s\n", val) }

type DarkModal struct{ title string }

func (m *DarkModal) Render() string { return fmt.Sprintf("[DARK MODAL: %s]", m.title) }
func (m *DarkModal) Show()          { fmt.Printf("Showing dark modal: %s\n", m.title) }
func (m *DarkModal) Hide()          { fmt.Println("Hiding dark modal") }

type DarkThemeFactory struct{}

func (f *DarkThemeFactory) CreateButton(label string) Button {
	return &DarkButton{label: label}
}
func (f *DarkThemeFactory) CreateTextInput(placeholder string) TextInput {
	return &DarkTextInput{placeholder: placeholder}
}
func (f *DarkThemeFactory) CreateModal(title string) Modal {
	return &DarkModal{title: title}
}

// --- Light Theme Family ---
type LightButton struct{ label string }

func (b *LightButton) Render() string         { return fmt.Sprintf("[LIGHT BTN: %s]", b.label) }
func (b *LightButton) OnClick(handler func()) { handler() }

type LightTextInput struct{ placeholder string }

func (t *LightTextInput) Render() string      { return fmt.Sprintf("[LIGHT INPUT: %s]", t.placeholder) }
func (t *LightTextInput) SetValue(val string) { fmt.Printf("Light input set: %s\n", val) }

type LightModal struct{ title string }

func (m *LightModal) Render() string { return fmt.Sprintf("[LIGHT MODAL: %s]", m.title) }
func (m *LightModal) Show()          { fmt.Printf("Showing light modal: %s\n", m.title) }
func (m *LightModal) Hide()          { fmt.Println("Hiding light modal") }

type LightThemeFactory struct{}

func (f *LightThemeFactory) CreateButton(label string) Button {
	return &LightButton{label: label}
}
func (f *LightThemeFactory) CreateTextInput(placeholder string) TextInput {
	return &LightTextInput{placeholder: placeholder}
}
func (f *LightThemeFactory) CreateModal(title string) Modal {
	return &LightModal{title: title}
}

// --- Client code: works with ANY theme ---
func RenderLoginPage(factory UIFactory) {
	btn := factory.CreateButton("Login")
	input := factory.CreateTextInput("Enter email...")
	modal := factory.CreateModal("Welcome")

	fmt.Println(input.Render())
	fmt.Println(btn.Render())
	fmt.Println(modal.Render())
}

// =============================================================================
// 3. BUILDER PATTERN
// =============================================================================
// "Separate the construction of a complex object from its representation."
//
// WHEN TO USE: Object has many optional fields, or construction requires steps.
// In Go, this is extremely common for configs, HTTP requests, queries.

// --- The product ---
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
	Timeout int // seconds
	Retries int
	Auth    *AuthConfig
}

type AuthConfig struct {
	Type  string // "bearer", "basic"
	Token string
}

// --- The builder ---
type HTTPRequestBuilder struct {
	request *HTTPRequest
}

func NewHTTPRequestBuilder(method, url string) *HTTPRequestBuilder {
	return &HTTPRequestBuilder{
		request: &HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: make(map[string]string),
			Timeout: 30, // sensible default
			Retries: 3,  // sensible default
		},
	}
}

func (b *HTTPRequestBuilder) WithHeader(key, value string) *HTTPRequestBuilder {
	b.request.Headers[key] = value
	return b
}

func (b *HTTPRequestBuilder) WithBody(body []byte) *HTTPRequestBuilder {
	b.request.Body = body
	return b
}

func (b *HTTPRequestBuilder) WithTimeout(seconds int) *HTTPRequestBuilder {
	b.request.Timeout = seconds
	return b
}

func (b *HTTPRequestBuilder) WithRetries(n int) *HTTPRequestBuilder {
	b.request.Retries = n
	return b
}

func (b *HTTPRequestBuilder) WithBearerAuth(token string) *HTTPRequestBuilder {
	b.request.Auth = &AuthConfig{Type: "bearer", Token: token}
	return b
}

func (b *HTTPRequestBuilder) Build() *HTTPRequest {
	return b.request
}

// Usage:
func ExampleBuilder() {
	req := NewHTTPRequestBuilder("POST", "https://api.example.com/users").
		WithHeader("Content-Type", "application/json").
		WithHeader("X-Request-ID", "abc-123").
		WithBody([]byte(`{"name": "Vikram"}`)).
		WithTimeout(10).
		WithRetries(5).
		WithBearerAuth("my-token").
		Build()

	fmt.Printf("Request: %s %s (timeout: %ds, retries: %d)\n",
		req.Method, req.URL, req.Timeout, req.Retries)
}

// --- FUNCTIONAL OPTIONS PATTERN (Go-idiomatic alternative to Builder) ---
// This is what the Go community actually uses most. Study this pattern deeply.

type Server struct {
	host       string
	port       int
	maxConn    int
	timeout    int
	tlsEnabled bool
}

// Option is a function that configures the server
type ServerOption func(*Server)

func WithPort(port int) ServerOption {
	return func(s *Server) { s.port = port }
}

func WithMaxConnections(n int) ServerOption {
	return func(s *Server) { s.maxConn = n }
}

func WithTimeout(seconds int) ServerOption {
	return func(s *Server) { s.timeout = seconds }
}

func WithTLS() ServerOption {
	return func(s *Server) { s.tlsEnabled = true }
}

func NewServer(host string, opts ...ServerOption) *Server {
	// Start with sensible defaults
	s := &Server{
		host:    host,
		port:    8080,
		maxConn: 100,
		timeout: 30,
	}
	// Apply options
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Usage:
func ExampleFunctionalOptions() {
	// Minimal — all defaults
	s1 := NewServer("localhost")

	// Customized
	s2 := NewServer("0.0.0.0",
		WithPort(443),
		WithMaxConnections(1000),
		WithTLS(),
	)

	fmt.Printf("Server 1: %s:%d\n", s1.host, s1.port)
	fmt.Printf("Server 2: %s:%d (TLS: %v)\n", s2.host, s2.port, s2.tlsEnabled)
}

// =============================================================================
// 4. SINGLETON PATTERN
// =============================================================================
// "Ensure a class has only ONE instance and provide global access."
//
// In Go: Use sync.Once. NEVER use init() for singletons — it's not lazy.
// WARNING: Singletons make testing hard. Use sparingly (configs, connection pools).

type DatabasePool struct {
	connections int
	host        string
}

var (
	dbPoolInstance *DatabasePool
	dbPoolOnce     sync.Once
)

func GetDatabasePool() *DatabasePool {
	dbPoolOnce.Do(func() {
		// This runs EXACTLY once, even with concurrent access
		dbPoolInstance = &DatabasePool{
			connections: 10,
			host:        "localhost:5432",
		}
		fmt.Println("Database pool initialized (this prints only once)")
	})
	return dbPoolInstance
}

// BETTER ALTERNATIVE: Dependency injection
// Instead of a singleton, create the pool once in main() and pass it around.
// This is more testable and explicit.

// =============================================================================
// 5. PROTOTYPE PATTERN
// =============================================================================
// "Create new objects by copying existing ones."
//
// When object creation is expensive, clone an existing one and modify.

type Cloneable interface {
	Clone() Cloneable
}

type DocumentTemplate struct {
	Title    string
	Content  string
	Metadata map[string]string
	Sections []string
}

func (d *DocumentTemplate) Clone() *DocumentTemplate {
	// Deep copy — don't share references!
	newMeta := make(map[string]string, len(d.Metadata))
	for k, v := range d.Metadata {
		newMeta[k] = v
	}
	newSections := make([]string, len(d.Sections))
	copy(newSections, d.Sections)

	return &DocumentTemplate{
		Title:    d.Title,
		Content:  d.Content,
		Metadata: newMeta,
		Sections: newSections,
	}
}

func ExamplePrototype() {
	// Create a template once
	template := &DocumentTemplate{
		Title:    "Monthly Report",
		Content:  "Report template content...",
		Metadata: map[string]string{"author": "system", "version": "1.0"},
		Sections: []string{"Summary", "Details", "Conclusion"},
	}

	// Clone and customize for different months
	janReport := template.Clone()
	janReport.Title = "January Report"
	janReport.Metadata["month"] = "January"

	febReport := template.Clone()
	febReport.Title = "February Report"
	febReport.Metadata["month"] = "February"

	// Original is untouched
	fmt.Println(template.Title)  // "Monthly Report"
	fmt.Println(janReport.Title) // "January Report"
	fmt.Println(febReport.Title) // "February Report"
}

// =============================================================================
// 6. OBJECT POOL PATTERN
// =============================================================================
// "Reuse expensive-to-create objects instead of creating/destroying them."
//
// Go has sync.Pool in the stdlib, but for typed pools with limits:

type Connection struct {
	ID     int
	Active bool
}

type ConnectionPool struct {
	mu          sync.Mutex
	connections []*Connection
	available   []*Connection
	maxSize     int
	nextID      int
}

func NewConnectionPool(maxSize int) *ConnectionPool {
	return &ConnectionPool{
		maxSize:     maxSize,
		connections: make([]*Connection, 0, maxSize),
		available:   make([]*Connection, 0, maxSize),
	}
}

func (p *ConnectionPool) Acquire() (*Connection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Try to reuse an available connection
	if len(p.available) > 0 {
		conn := p.available[len(p.available)-1]
		p.available = p.available[:len(p.available)-1]
		conn.Active = true
		return conn, nil
	}

	// Create new if under limit
	if len(p.connections) < p.maxSize {
		p.nextID++
		conn := &Connection{ID: p.nextID, Active: true}
		p.connections = append(p.connections, conn)
		return conn, nil
	}

	return nil, fmt.Errorf("pool exhausted (max: %d)", p.maxSize)
}

func (p *ConnectionPool) Release(conn *Connection) {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn.Active = false
	p.available = append(p.available, conn)
}

func (p *ConnectionPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.connections)
}

func (p *ConnectionPool) Available() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.available)
}

func ExampleObjectPool() {
	pool := NewConnectionPool(3)

	c1, _ := pool.Acquire()
	c2, _ := pool.Acquire()
	fmt.Printf("Pool size: %d, Available: %d\n", pool.Size(), pool.Available())

	pool.Release(c1)
	fmt.Printf("After release — Pool size: %d, Available: %d\n", pool.Size(), pool.Available())

	// Reuses c1's connection
	c3, _ := pool.Acquire()
	fmt.Printf("Reused connection: %d\n", c3.ID)
	_ = c2
}
