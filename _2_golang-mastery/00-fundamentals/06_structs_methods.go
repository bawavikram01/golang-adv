//go:build ignore

// =============================================================================
// LESSON 0.6: STRUCTS & METHODS — Go's Object-Oriented Building Block
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Struct declaration, initialization, and field access
// - Methods: value vs pointer receivers
// - Embedding: Go's version of composition (not inheritance!)
// - Struct tags: JSON, DB, validation metadata
// - Anonymous structs and anonymous fields
// - Constructor patterns (NewXxx)
// - Functional options pattern
// - Struct comparison and copying
//
// THE KEY INSIGHT:
// Go has NO classes, NO inheritance, NO constructors, NO self/this keyword.
// Instead: structs hold data, methods are functions with a receiver parameter,
// and composition is achieved through embedding. This is simpler than OOP
// and scales better in large codebases because dependencies are explicit.
//
// RUN: go run 06_structs_methods.go
// =============================================================================

package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== STRUCTS & METHODS ===")
	fmt.Println()

	structBasics()
	methodsDeepDive()
	embedding()
	structTags()
	anonymousStructs()
	constructorPatterns()
	functionalOptions()
}

// =============================================================================
// PART 1: Struct Basics
// =============================================================================

type Point struct {
	X float64
	Y float64
}

type Person struct {
	Name   string
	Age    int
	Email  string
	Active bool
}

func structBasics() {
	fmt.Println("--- STRUCT BASICS ---")

	// ─── Declaration & initialization ───

	// 1. Named fields (preferred — order-independent, self-documenting)
	p1 := Person{
		Name:   "Vikram",
		Age:    25,
		Email:  "v@test.com",
		Active: true,
	}
	fmt.Printf("  Named: %+v\n", p1)

	// 2. Positional (all fields required, order matters — FRAGILE)
	p2 := Point{3.0, 4.0}
	fmt.Printf("  Positional: %+v\n", p2)

	// 3. Zero value (all fields have their zero values)
	var p3 Person // Name="", Age=0, Email="", Active=false
	fmt.Printf("  Zero: %+v\n", p3)

	// 4. Pointer with &
	p4 := &Person{Name: "Alice", Age: 30}
	fmt.Printf("  Pointer: %+v\n", *p4)

	// ─── Field access ───
	fmt.Printf("  p1.Name: %q\n", p1.Name)
	p1.Age = 26 // modify field
	fmt.Printf("  Modified age: %d\n", p1.Age)

	// Auto-dereference for pointers
	p4.Email = "alice@test.com" // same as (*p4).Email = ...
	fmt.Printf("  Pointer field: %q\n", p4.Email)

	// ─── Struct comparison ───
	a := Point{1, 2}
	b := Point{1, 2}
	c := Point{3, 4}
	fmt.Printf("  %v == %v: %v\n", a, b, a == b)
	fmt.Printf("  %v == %v: %v\n", a, c, a == c)
	// Structs are comparable only if ALL fields are comparable
	// If a field is a slice, map, or func → struct is NOT comparable

	// ─── Struct copying ───
	// Assigning a struct COPIES all fields (value semantics)
	original := Person{Name: "Bob", Age: 30}
	copied := original
	copied.Name = "Bobby"
	fmt.Printf("  Copy: original=%q, copy=%q (independent)\n", original.Name, copied.Name)

	fmt.Println()
}

// =============================================================================
// PART 2: Methods
// =============================================================================

type Rectangle struct {
	Width  float64
	Height float64
}

// Value receiver: immutable access
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Pointer receiver: can modify the struct
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

// Stringer interface (like toString in Java)
func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%.1f×%.1f)", r.Width, r.Height)
}

func methodsDeepDive() {
	fmt.Println("--- METHODS ---")

	r := Rectangle{Width: 5, Height: 3}
	fmt.Printf("  %s\n", r)
	fmt.Printf("  Area: %.1f\n", r.Area())
	fmt.Printf("  Perimeter: %.1f\n", r.Perimeter())

	r.Scale(2) // Go auto-takes address: (&r).Scale(2)
	fmt.Printf("  After Scale(2): %s\n", r)
	fmt.Printf("  Area: %.1f\n", r.Area())

	// ─── Method expressions ───
	// You can get a reference to a method as a function
	areaFn := Rectangle.Area // function(Rectangle) float64
	fmt.Printf("  Method expression: %.1f\n", areaFn(Rectangle{10, 5}))

	// ─── Method values (bound to a specific receiver) ───
	r2 := Rectangle{7, 3}
	boundArea := r2.Area // function() float64 — captures r2
	fmt.Printf("  Method value: %.1f\n", boundArea())

	// ─── Methods on non-struct types ───
	// You can define methods on ANY named type (not just structs)
	type Celsius float64
	// func (c Celsius) ToFahrenheit() float64 { return float64(c)*9/5 + 32 }
	// Can't add methods to unnamed types or types from other packages

	fmt.Println()
}

// =============================================================================
// PART 3: Embedding — Composition Over Inheritance
// =============================================================================

type Address struct {
	Street string
	City   string
	State  string
}

func (a Address) FullAddress() string {
	return fmt.Sprintf("%s, %s, %s", a.Street, a.City, a.State)
}

type Employee struct {
	Person  // embedded — fields and methods promoted
	Address // embedded — fields and methods promoted
	Company string
	Salary  float64
}

// Logger example
type Logger struct {
	Prefix string
}

func (l Logger) Log(msg string) {
	fmt.Printf("  [%s] %s\n", l.Prefix, msg)
}

type Server struct {
	Logger // embedded: Server "has a" Logger
	Port   int
}

func embedding() {
	fmt.Println("--- EMBEDDING ---")

	// ─── Field and method promotion ───
	e := Employee{
		Person:  Person{Name: "Alice", Age: 30, Active: true},
		Address: Address{Street: "123 Main St", City: "NYC", State: "NY"},
		Company: "Gopher Inc",
		Salary:  120000,
	}

	// Promoted fields: access directly (no e.Person.Name needed)
	fmt.Printf("  Name: %s (promoted from Person)\n", e.Name)
	fmt.Printf("  City: %s (promoted from Address)\n", e.City)
	fmt.Printf("  Company: %s\n", e.Company)

	// Promoted methods
	fmt.Printf("  Address: %s (promoted method)\n", e.FullAddress())

	// Can still access the embedded struct directly
	fmt.Printf("  Person: %+v\n", e.Person)

	// ─── Embedding is NOT inheritance ───
	// Employee IS NOT a Person. Employee HAS a Person.
	// You can't pass Employee where Person is expected (no polymorphism).
	// Use interfaces for polymorphism.

	// ─── Name conflict: outer wins ───
	type Base struct{ Name string }
	type Derived struct {
		Base
		Name string // shadows Base.Name
	}
	d := Derived{Base: Base{Name: "base"}, Name: "derived"}
	fmt.Printf("  Shadow: d.Name=%q, d.Base.Name=%q\n", d.Name, d.Base.Name)

	// ─── Embedding for delegation ───
	s := Server{
		Logger: Logger{Prefix: "SERVER"},
		Port:   8080,
	}
	s.Log("started") // calls s.Logger.Log("started")

	// ─── Embedding interfaces (advanced) ───
	// type ReadWriter struct {
	//     io.Reader  // embedded interface — gains Read method
	//     io.Writer  // embedded interface — gains Write method
	// }

	fmt.Println()
}

// =============================================================================
// PART 4: Struct Tags — Metadata for Field Behavior
// =============================================================================

type APIUser struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"-"` // never serialized
	CreatedAt time.Time `json:"created_at"`
	Score     int       `json:"score,string"`          // serialized as string
	Internal  string    `json:"-" db:"internal_field"` // different tag per package
}

func structTags() {
	fmt.Println("--- STRUCT TAGS ---")

	// Tags are string metadata attached to struct fields.
	// Convention: `key:"value" key2:"value2"`
	// Parsed by reflect package (used by encoding/json, database drivers, etc.)
	//
	// COMMON TAGS:
	// json:"name,omitempty"      — encoding/json
	// xml:"name,attr"            — encoding/xml
	// db:"column_name"           — sqlx, gorm
	// yaml:"name"                — yaml parser
	// validate:"required,min=1"  — validator
	// form:"field_name"          — HTTP form binding
	// env:"VAR_NAME"             — environment config

	u := APIUser{
		ID:        1,
		Username:  "vikram",
		Password:  "secret123",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Score:     95,
	}

	data, _ := json.MarshalIndent(u, "  ", "  ")
	fmt.Printf("  JSON:\n  %s\n", data)
	// Password is omitted (json:"-")
	// Email is omitted (omitempty, empty string)
	// Score is "95" not 95 (json:",string")

	fmt.Println()
}

// =============================================================================
// PART 5: Anonymous Structs
// =============================================================================
func anonymousStructs() {
	fmt.Println("--- ANONYMOUS STRUCTS ---")

	// Define and use a struct inline (no named type needed)
	// USE FOR: one-off data structures, test data, JSON parsing

	// ─── Inline struct variable ───
	point := struct {
		X, Y int
	}{10, 20}
	fmt.Printf("  Inline struct: %+v\n", point)

	// ─── Common: JSON parsing when you don't want a named type ───
	data := `{"name":"Go","year":2009}`
	var result struct {
		Name string `json:"name"`
		Year int    `json:"year"`
	}
	json.Unmarshal([]byte(data), &result)
	fmt.Printf("  JSON parse: %+v\n", result)

	// ─── Common: table-driven tests ───
	tests := []struct {
		input    int
		expected int
	}{
		{1, 2},
		{2, 4},
		{3, 6},
	}
	for _, tt := range tests {
		got := tt.input * 2
		fmt.Printf("  Test: %d*2 = %d (expected %d, pass=%v)\n",
			tt.input, got, tt.expected, got == tt.expected)
	}

	// ─── Anonymous field (unnamed field by type) ───
	type Data struct {
		int    // anonymous field, type name becomes field name
		string // anonymous field
	}
	d := Data{42, "hello"}
	fmt.Printf("  Anonymous fields: int=%d, string=%q\n", d.int, d.string)
	// Rarely used — embedding named structs is more common

	fmt.Println()
}

// =============================================================================
// PART 6: Constructor Patterns
// =============================================================================

type Server2 struct {
	host    string
	port    int
	maxConn int
	tls     bool
}

// ─── Simple constructor ───
func NewServer(host string, port int) *Server2 {
	return &Server2{
		host:    host,
		port:    port,
		maxConn: 100, // sensible defaults
		tls:     false,
	}
}

// ─── Constructor with validation ───
func NewValidatedServer(host string, port int) (*Server2, error) {
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", port)
	}
	return &Server2{host: host, port: port, maxConn: 100}, nil
}

func constructorPatterns() {
	fmt.Println("--- CONSTRUCTOR PATTERNS ---")

	// ─── Pattern 1: NewXxx constructor ───
	s := NewServer("localhost", 8080)
	fmt.Printf("  NewServer: host=%s port=%d\n", s.host, s.port)

	// ─── Pattern 2: Constructor with validation ───
	s2, err := NewValidatedServer("", 8080)
	fmt.Printf("  Validated (empty host): s=%v err=%v\n", s2, err)

	s3, err := NewValidatedServer("localhost", 8080)
	fmt.Printf("  Validated (good): host=%s err=%v\n", s3.host, err)

	// ─── Pattern 3: Config struct ───
	// When there are many optional parameters:
	type DBConfig struct {
		Host     string
		Port     int
		User     string
		Password string
		Database string
		MaxConns int
		Timeout  time.Duration
	}

	db := DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "admin",
		Database: "mydb",
		MaxConns: 25,
		Timeout:  5 * time.Second,
	}
	fmt.Printf("  Config struct: %s:%d/%s\n", db.Host, db.Port, db.Database)

	fmt.Println()
}

// =============================================================================
// PART 7: Functional Options Pattern
// =============================================================================

// The most flexible constructor pattern in Go.
// Used by: gRPC, Uber's zap logger, many production libraries.

type HTTPServer struct {
	host         string
	port         int
	maxBodySize  int64
	readTimeout  time.Duration
	writeTimeout time.Duration
	tls          bool
}

// Option is a function that configures HTTPServer
type Option func(*HTTPServer)

// Each option is a function that returns an Option
func WithPort(port int) Option {
	return func(s *HTTPServer) {
		s.port = port
	}
}

func WithTLS(enabled bool) Option {
	return func(s *HTTPServer) {
		s.tls = enabled
	}
}

func WithTimeouts(read, write time.Duration) Option {
	return func(s *HTTPServer) {
		s.readTimeout = read
		s.writeTimeout = write
	}
}

func WithMaxBodySize(size int64) Option {
	return func(s *HTTPServer) {
		s.maxBodySize = size
	}
}

func NewHTTPServer(host string, opts ...Option) *HTTPServer {
	// Defaults
	s := &HTTPServer{
		host:         host,
		port:         8080,
		maxBodySize:  1 << 20, // 1MB
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func functionalOptions() {
	fmt.Println("--- FUNCTIONAL OPTIONS ---")

	// Defaults only
	s1 := NewHTTPServer("localhost")
	fmt.Printf("  Defaults: port=%d tls=%v\n", s1.port, s1.tls)

	// Custom options
	s2 := NewHTTPServer("0.0.0.0",
		WithPort(443),
		WithTLS(true),
		WithTimeouts(5*time.Second, 10*time.Second),
		WithMaxBodySize(10<<20),
	)
	fmt.Printf("  Custom: port=%d tls=%v readTimeout=%v maxBody=%d\n",
		s2.port, s2.tls, s2.readTimeout, s2.maxBodySize)

	// WHY THIS PATTERN IS GREAT:
	// 1. Backward compatible: add new options without breaking existing callers
	// 2. Self-documenting: WithPort(443) reads better than positional args
	// 3. Flexible defaults: only override what you need
	// 4. Composable: options can be grouped into presets
	//    prodDefaults := []Option{WithTLS(true), WithPort(443)}
	//    s := NewHTTPServer("prod.com", prodDefaults...)

	fmt.Println()
}
