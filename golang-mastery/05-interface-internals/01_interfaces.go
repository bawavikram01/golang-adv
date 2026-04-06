// =============================================================================
// LESSON 5: INTERFACE INTERNALS & ADVANCED DESIGN
// =============================================================================
//
// Interfaces in Go are THE key abstraction mechanism. Understanding their
// internal representation, cost, and design principles separates advanced
// developers from beginners.
//
// INTERNAL REPRESENTATION:
//   Empty interface (any/interface{}):
//     type eface struct { _type *_type; data unsafe.Pointer }
//   Non-empty interface:
//     type iface struct { tab *itab; data unsafe.Pointer }
//   itab contains: interface type + concrete type + method table (vtable)
//
// COST: Interface method call ≈ 2 pointer dereferences + indirect function call
//       (~2-5ns overhead vs direct call). Inlining is usually defeated.
// =============================================================================

package main

import (
	"fmt"
	"io"
	"strings"
)

// =============================================================================
// PRINCIPLE 1: Accept interfaces, return structs
// =============================================================================
//
// Functions should ACCEPT interfaces (for flexibility) and RETURN concrete
// types (for usability). The caller decides what interface to store it as.

type UserRepository interface {
	FindByID(id int64) (*UserEntity, error)
	Save(user *UserEntity) error
}

type UserEntity struct {
	ID    int64
	Name  string
	Email string
}

// Concrete implementation returned (not the interface)
type PostgresUserRepo struct {
	dsn string
}

func NewPostgresUserRepo(dsn string) *PostgresUserRepo {
	return &PostgresUserRepo{dsn: dsn} // returns concrete type
}

func (r *PostgresUserRepo) FindByID(id int64) (*UserEntity, error) {
	return &UserEntity{ID: id, Name: "test", Email: "test@test.com"}, nil
}

func (r *PostgresUserRepo) Save(user *UserEntity) error { return nil }

// Consumer accepts the interface
func GetUser(repo UserRepository, id int64) (*UserEntity, error) {
	return repo.FindByID(id) // works with any implementation
}

// =============================================================================
// PRINCIPLE 2: Small interfaces — the power of single-method interfaces
// =============================================================================
//
// io.Reader, io.Writer, fmt.Stringer, sort.Interface, http.Handler
// Go's standard library is built on small interfaces.
// Compose large behaviors from small interfaces.

// Small, focused interface
type Validator interface {
	Validate() error
}

type Serializer interface {
	Serialize() ([]byte, error)
}

// Compose via embedding
type ValidatingSerializer interface {
	Validator
	Serializer
}

// Implementation satisfies both without knowing about the combined interface
type Order struct {
	ID    int
	Total float64
	Items []string
}

func (o Order) Validate() error {
	if o.Total < 0 {
		return fmt.Errorf("total cannot be negative")
	}
	if len(o.Items) == 0 {
		return fmt.Errorf("order must have items")
	}
	return nil
}

func (o Order) Serialize() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"id":%d,"total":%.2f}`, o.ID, o.Total)), nil
}

// Function requires both capabilities — Order satisfies this implicitly
func ProcessOrder(vs ValidatingSerializer) error {
	if err := vs.Validate(); err != nil {
		return err
	}
	data, err := vs.Serialize()
	if err != nil {
		return err
	}
	fmt.Printf("  Processed: %s\n", data)
	return nil
}

// =============================================================================
// PRINCIPLE 3: Interface satisfaction is implicit — the nil interface gotcha
// =============================================================================

type MyError struct {
	Code    int
	Message string
}

func (e *MyError) Error() string {
	return fmt.Sprintf("error %d: %s", e.Code, e.Message)
}

func mightFail(fail bool) error {
	// WRONG: This returns a non-nil interface even when err is nil!
	var err *MyError // nil pointer
	if fail {
		err = &MyError{Code: 500, Message: "internal error"}
	}
	return err // BUG: interface{tab: *MyError type, data: nil} != nil!
}

func mightFailCorrect(fail bool) error {
	// CORRECT: Return nil explicitly for the interface
	if fail {
		return &MyError{Code: 500, Message: "internal error"}
	}
	return nil // interface itself is nil
}

// =============================================================================
// PRINCIPLE 4: Type switches and type assertions — advanced patterns
// =============================================================================

// Multi-interface type switch for behavior detection
type Reader interface{ Read() }
type Writer interface{ Write() }
type Closer interface{ Close() }

type File struct{}
func (File) Read()  { fmt.Println("  reading") }
func (File) Write() { fmt.Println("  writing") }
func (File) Close() { fmt.Println("  closing") }

type Pipe struct{}
func (Pipe) Read()  { fmt.Println("  pipe reading") }
func (Pipe) Write() { fmt.Println("  pipe writing") }

func handleIO(v interface{}) {
	// Check capabilities at runtime
	if r, ok := v.(Reader); ok {
		fmt.Print("  Has Read: ")
		r.Read()
	}
	if w, ok := v.(Writer); ok {
		fmt.Print("  Has Write: ")
		w.Write()
	}
	if c, ok := v.(Closer); ok {
		fmt.Print("  Has Close: ")
		c.Close()
	} else {
		fmt.Println("  No Close capability")
	}
}

// =============================================================================
// PRINCIPLE 5: Functional options pattern — elegant API design with interfaces
// =============================================================================

type ServerConfig struct {
	host    string
	port    int
	maxConn int
	tls     bool
}

type Option func(*ServerConfig)

func WithHost(host string) Option {
	return func(c *ServerConfig) { c.host = host }
}

func WithPort(port int) Option {
	return func(c *ServerConfig) { c.port = port }
}

func WithMaxConn(n int) Option {
	return func(c *ServerConfig) { c.maxConn = n }
}

func WithTLS() Option {
	return func(c *ServerConfig) { c.tls = true }
}

func NewServer(opts ...Option) *ServerConfig {
	cfg := &ServerConfig{
		host:    "localhost",
		port:    8080,
		maxConn: 100,
		tls:     false,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// =============================================================================
// PRINCIPLE 6: Interface wrapping/decoration (middleware pattern)
// =============================================================================

// Wrapping io.Writer to add capabilities
type CountingWriter struct {
	writer    io.Writer
	BytesWritten int
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{writer: w}
}

func (cw *CountingWriter) Write(p []byte) (int, error) {
	n, err := cw.writer.Write(p)
	cw.BytesWritten += n
	return n, err
}

// Chaining decorators
type UppercaseWriter struct {
	writer io.Writer
}

func NewUppercaseWriter(w io.Writer) *UppercaseWriter {
	return &UppercaseWriter{writer: w}
}

func (uw *UppercaseWriter) Write(p []byte) (int, error) {
	upper := []byte(strings.ToUpper(string(p)))
	return uw.writer.Write(upper)
}

func main() {
	// Principle 1: Accept interfaces, return structs
	fmt.Println("=== Accept Interfaces, Return Structs ===")
	repo := NewPostgresUserRepo("postgres://localhost/db")
	user, _ := GetUser(repo, 1) // repo automatically satisfies UserRepository
	fmt.Printf("User: %+v\n", user)

	// Principle 2: Small interfaces
	fmt.Println("\n=== Composable Small Interfaces ===")
	order := Order{ID: 1, Total: 99.99, Items: []string{"widget"}}
	ProcessOrder(order)

	// Principle 3: Nil interface gotcha
	fmt.Println("\n=== Nil Interface Gotcha ===")
	err := mightFail(false)
	fmt.Printf("mightFail(false) == nil: %v (BUG! Should be true)\n", err == nil)
	err = mightFailCorrect(false)
	fmt.Printf("mightFailCorrect(false) == nil: %v (CORRECT)\n", err == nil)

	// Principle 4: Type switches
	fmt.Println("\n=== Capability Detection ===")
	fmt.Println("File capabilities:")
	handleIO(File{})
	fmt.Println("Pipe capabilities:")
	handleIO(Pipe{})

	// Principle 5: Functional options
	fmt.Println("\n=== Functional Options ===")
	srv := NewServer(WithHost("0.0.0.0"), WithPort(443), WithTLS())
	fmt.Printf("Server: %+v\n", srv)

	// Principle 6: Writer decoration chain
	fmt.Println("\n=== Interface Decoration Chain ===")
	var buf strings.Builder
	counter := NewCountingWriter(&buf)
	upper := NewUppercaseWriter(counter)

	fmt.Fprint(upper, "hello world")
	fmt.Printf("Output: %q\n", buf.String())
	fmt.Printf("Bytes written: %d\n", counter.BytesWritten)
}
