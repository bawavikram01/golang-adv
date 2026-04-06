# Go (Golang) Complete Mastery Roadmap

> Every single topic from absolute zero to runtime-hacker level.
> Checkboxes track your progress. Topics marked with 🔥 are the ones that separate gods from everyone else.

---

## Phase 1 — Foundations (The Non-Negotiables)

### 1.1 Language Basics
- [ ] Installing Go, GOPATH vs Go Modules
- [ ] `go run`, `go build`, `go install`
- [ ] Package structure and `main` package
- [ ] Variables: `var`, `:=`, zero values
- [ ] Basic types: `int`, `float64`, `string`, `bool`, `byte`, `rune`
- [ ] Constants and `iota`
- [ ] Type conversions (no implicit conversions in Go)
- [ ] Comments and documentation (`godoc` format)

### 1.2 Control Flow
- [ ] `if`, `else`, `if` with init statement (`if err := ...; err != nil`)
- [ ] `for` loop (only loop in Go)
- [ ] `for range` over slices, maps, strings, channels
- [ ] `switch` (no fallthrough by default, type switch)
- [ ] `select` (channel multiplexing)
- [ ] `goto`, labels, `break`/`continue` with labels
- [ ] `defer` (LIFO order, evaluated immediately, runs on return)

### 1.3 Functions
- [ ] Multiple return values
- [ ] Named return values (and naked returns)
- [ ] Variadic functions (`...T`)
- [ ] First-class functions (functions as values)
- [ ] Anonymous functions (closures)
- [ ] `init()` function (runs before `main`, per package)
- [ ] Recursion and tail-call optimization (Go does NOT optimize tail calls)

### 1.4 Data Structures
- [ ] Arrays (fixed size, value type)
- [ ] Slices (dynamic, reference to underlying array)
- [ ] Slice internals: `len`, `cap`, growth strategy (2x then 1.25x)
- [ ] Slice tricks: append, copy, delete, insert, filter in-place
- [ ] `make` vs `new`
- [ ] Maps (hash table, unordered, not concurrent-safe)
- [ ] Map internals: buckets, load factor, evacuation
- [ ] Structs (value type, field alignment/padding)
- [ ] Struct embedding (composition over inheritance)
- [ ] Anonymous structs
- [ ] Struct tags (`json`, `db`, `validate`, `yaml`)

### 1.5 Pointers
- [ ] Pointer syntax (`*T`, `&v`)
- [ ] Pointer vs value semantics
- [ ] When to use pointers (mutation, large structs, interfaces)
- [ ] No pointer arithmetic (use `unsafe` package for that)
- [ ] `nil` pointers and nil checks

### 1.6 Strings
- [ ] Strings are immutable byte slices
- [ ] `rune` vs `byte` (Unicode code points vs raw bytes)
- [ ] `strings` package (Builder, Split, Join, Contains, Replace, Trim)
- [ ] `strconv` package (Atoi, Itoa, ParseFloat, FormatInt)
- [ ] `fmt` package (Sprintf, Fprintf, verbs: `%v`, `%+v`, `%#v`, `%T`)
- [ ] String concatenation performance (`+` vs `Builder` vs `Join`)
- [ ] `unicode/utf8` package

---

## Phase 2 — Core Go Idioms

### 2.1 Error Handling
- [ ] The `error` interface
- [ ] `errors.New()` and `fmt.Errorf()`
- [ ] Error wrapping with `%w` (Go 1.13+)
- [ ] `errors.Is()` — check error identity through wrapping chain
- [ ] `errors.As()` — extract typed error from chain
- [ ] `errors.Join()` — multi-error (Go 1.20+)
- [ ] 🔥 Sentinel errors vs custom error types vs wrapping
- [ ] 🔥 Error handling patterns: wrap at boundaries, log at top
- [ ] 🔥 Domain error design (Op, Kind, Entity pattern)
- [ ] `panic` and `recover` (when to use: truly unrecoverable only)

### 2.2 Interfaces
- [ ] Implicit satisfaction (no `implements` keyword)
- [ ] Interface as contract
- [ ] Empty interface `interface{}` / `any`
- [ ] Type assertion (`v.(Type)`, `v, ok := v.(Type)`)
- [ ] Type switch
- [ ] Interface embedding (composition)
- [ ] Standard interfaces: `io.Reader`, `io.Writer`, `fmt.Stringer`, `error`, `sort.Interface`
- [ ] 🔥 Accept interfaces, return structs
- [ ] 🔥 Small interfaces (1-2 methods ideal)
- [ ] 🔥 Define interfaces at the consumer, not producer
- [ ] 🔥 Nil interface vs interface holding nil pointer (the gotcha)

### 2.3 Methods
- [ ] Value receivers vs pointer receivers
- [ ] Method sets (pointer type has all methods, value type only value-receiver methods)
- [ ] Methods on non-struct types (e.g., `type Celsius float64`)
- [ ] When to use value vs pointer receiver (mutation, consistency, size)

### 2.4 Packages & Modules
- [ ] `go mod init`, `go mod tidy`, `go.sum`
- [ ] Semantic versioning and module versioning
- [ ] `go.mod`: require, replace, exclude, retract
- [ ] Internal packages (`internal/`)
- [ ] Vendoring (`go mod vendor`)
- [ ] Package naming conventions (short, lowercase, no underscores)
- [ ] Circular dependency prevention
- [ ] `go work` — multi-module workspaces (Go 1.18+)

---

## Phase 3 — Concurrency (Go's Superpower)

### 3.1 Goroutines
- [ ] `go` keyword — launching goroutines
- [ ] Goroutine lifecycle
- [ ] Goroutine vs OS thread vs green thread
- [ ] 🔥 Goroutine cost (~2-8KB stack, grows dynamically)
- [ ] 🔥 Goroutine leaks — causes and detection
- [ ] `runtime.NumGoroutine()` for leak detection

### 3.2 Channels
- [ ] Unbuffered channels (synchronous rendezvous)
- [ ] Buffered channels (async up to capacity)
- [ ] Directional channels (`chan<-`, `<-chan`)
- [ ] `close()` and range over channel
- [ ] `select` with `default` (non-blocking)
- [ ] `select` with `time.After` and `time.Tick`
- [ ] 🔥 Channel ownership pattern (creator closes, writer closes)
- [ ] 🔥 Nil channel behavior in select (blocks forever = disables case)

### 3.3 Channel Patterns
- [ ] Generator (function returning `<-chan T`)
- [ ] Fan-out (multiple goroutines reading from one channel)
- [ ] Fan-in (merge multiple channels into one)
- [ ] Pipeline (chain of stages)
- [ ] 🔥 Or-done channel (read with cancellation)
- [ ] 🔥 Tee channel (split one channel into two)
- [ ] 🔥 Bridge channel (flatten channel-of-channels)
- [ ] Semaphore (buffered channel as concurrency limiter)
- [ ] Rate limiter (token bucket via ticker)

### 3.4 sync Package
- [ ] `sync.Mutex` and `sync.RWMutex`
- [ ] `sync.WaitGroup`
- [ ] `sync.Once` (exactly-once initialization)
- [ ] `sync.OnceValue`, `sync.OnceFunc` (Go 1.21+)
- [ ] `sync.Map` (when to use vs `map` + `RWMutex`)
- [ ] `sync.Pool` (object reuse to reduce GC pressure)
- [ ] 🔥 `sync.Cond` (conditional waiting, broadcast)
- [ ] 🔥 When to use channels vs mutexes

### 3.5 sync/atomic
- [ ] `atomic.Int64`, `atomic.Bool` (Go 1.19+ typed atomics)
- [ ] `atomic.Pointer[T]` (Go 1.19+)
- [ ] `CompareAndSwap` (CAS) for lock-free data structures
- [ ] 🔥 Memory ordering and happens-before guarantees
- [ ] 🔥 Lock-free stack/queue using CAS

### 3.6 Context
- [ ] `context.Background()`, `context.TODO()`
- [ ] `context.WithCancel` — manual cancellation
- [ ] `context.WithTimeout` — deadline-based
- [ ] `context.WithDeadline` — absolute time
- [ ] `context.WithValue` — request-scoped data
- [ ] `context.WithCancelCause` (Go 1.20+)
- [ ] `context.AfterFunc` (Go 1.21+)
- [ ] 🔥 Cascading cancellation (parent cancels all children)
- [ ] 🔥 Context value key types (unexported to prevent collision)
- [ ] Rules: first param, never store in struct, never nil

### 3.7 Advanced Concurrency
- [ ] `errgroup.Group` (golang.org/x/sync)
- [ ] `semaphore.Weighted` (golang.org/x/sync)
- [ ] `singleflight.Group` (deduplicate concurrent requests)
- [ ] 🔥 Structured concurrency (parent waits for all children)
- [ ] 🔥 Data race detection (`go test -race`, `go build -race`)
- [ ] 🔥 The Go Memory Model (happens-before, synchronization points)

---

## Phase 4 — Standard Library Deep Dive

### 4.1 I/O
- [ ] `io.Reader`, `io.Writer` — the universal interfaces
- [ ] `io.Copy`, `io.TeeReader`, `io.LimitReader`, `io.MultiReader`
- [ ] `io.Pipe` (synchronous in-memory pipe)
- [ ] `bufio.Reader`, `bufio.Writer`, `bufio.Scanner`
- [ ] `os.File` — Open, Create, Read, Write, Seek
- [ ] `io/fs` — filesystem abstraction (Go 1.16+)
- [ ] `os.DirFS`, `embed.FS`, `testing/fstest.MapFS`

### 4.2 Encoding
- [ ] `encoding/json` — Marshal, Unmarshal, Encoder, Decoder
- [ ] JSON struct tags, `omitempty`, `-`, custom MarshalJSON/UnmarshalJSON
- [ ] `json.RawMessage` — delayed decoding
- [ ] `json.Number` — arbitrary precision numbers
- [ ] `encoding/xml`
- [ ] `encoding/csv`
- [ ] `encoding/gob` — Go-native binary encoding
- [ ] `encoding/binary` — raw byte encoding (BigEndian, LittleEndian)
- [ ] Protocol Buffers (`google.golang.org/protobuf`)

### 4.3 net/http
- [ ] `http.ListenAndServe`
- [ ] `http.Handler` and `http.HandlerFunc`
- [ ] `http.ServeMux` (Go 1.22+ pattern matching: `GET /users/{id}`)
- [ ] `http.Request` — Body, Headers, URL, Context, Form, MultipartForm
- [ ] `http.ResponseWriter` — WriteHeader, Write, Header
- [ ] Middleware pattern (func(http.Handler) http.Handler)
- [ ] `http.Client` — Transport, Timeout, Redirects
- [ ] 🔥 HTTP timeout architecture (5 layers)
- [ ] 🔥 Connection pooling and `Transport` tuning
- [ ] 🔥 `ResponseWriter.(http.Flusher)` for SSE/streaming
- [ ] 🔥 `ResponseWriter.(http.Hijacker)` for WebSocket/protocol upgrade
- [ ] `httptest.NewServer`, `httptest.NewRecorder` for testing

### 4.4 Networking
- [ ] `net.Listener`, `net.Conn` — raw TCP
- [ ] `net.Dial`, `net.DialTimeout`
- [ ] `net.TCPConn` — SetKeepAlive, SetNoDelay
- [ ] UDP: `net.ListenPacket`, `net.PacketConn`
- [ ] Unix domain sockets
- [ ] 🔥 Custom protocol framing (length-prefixed, delimiter-based)
- [ ] DNS: `net.Resolver`, custom DNS resolvers
- [ ] TLS: `crypto/tls` configuration

### 4.5 Time
- [ ] `time.Time`, `time.Duration`, `time.Now()`
- [ ] `time.After`, `time.Tick`, `time.NewTimer`, `time.NewTicker`
- [ ] Time formatting (Go's reference time: `Mon Jan 2 15:04:05 MST 2006`)
- [ ] `time.Parse`, `time.Format`
- [ ] Timezone handling (`time.LoadLocation`)
- [ ] 🔥 Monotonic clock readings (Go 1.9+)
- [ ] 🔥 Injecting a clock interface for testability

### 4.6 Logging
- [ ] `log` package (basic, avoid in production)
- [ ] `log/slog` (Go 1.21+ structured logging)
- [ ] `slog.Handler` — JSONHandler, TextHandler
- [ ] `slog.With()` — add persistent context
- [ ] Custom slog handlers
- [ ] Log levels: Debug, Info, Warn, Error
- [ ] 🔥 Request-scoped logging via context

### 4.7 Cryptography
- [ ] `crypto/sha256`, `crypto/md5` — hashing
- [ ] `crypto/hmac` — message authentication
- [ ] `crypto/aes`, `crypto/cipher` — encryption
- [ ] `crypto/rand` — secure random numbers
- [ ] `crypto/rsa`, `crypto/ecdsa` — asymmetric crypto
- [ ] `crypto/tls` — TLS configuration
- [ ] `crypto/x509` — certificate parsing
- [ ] `golang.org/x/crypto/bcrypt` — password hashing

### 4.8 Other Important Packages
- [ ] `regexp` — RE2 regular expressions (guaranteed linear time)
- [ ] `sort` — sort.Slice, sort.SliceStable, sort.Search
- [ ] `slices` package (Go 1.21+) — Contains, Sort, BinarySearch, Compact
- [ ] `maps` package (Go 1.21+) — Keys, Values, Clone
- [ ] `math`, `math/big` (arbitrary precision)
- [ ] `path/filepath` — OS-aware path manipulation
- [ ] `os/exec` — run external commands
- [ ] `os/signal` — signal handling (SIGTERM, SIGINT)
- [ ] `flag` — command-line argument parsing
- [ ] `text/template`, `html/template` — safe templating
- [ ] `database/sql` — database interface

---

## Phase 5 — Generics (Go 1.18+)

- [ ] Type parameters `[T any]`
- [ ] Constraints: `any`, `comparable`
- [ ] `cmp.Ordered` constraint
- [ ] Custom constraints (interface with type sets)
- [ ] The `~` operator (underlying type)
- [ ] Union type sets (`int | float64 | string`)
- [ ] Generic functions
- [ ] Generic types (structs, interfaces)
- [ ] 🔥 Generic data structures (BST, linked list, pool)
- [ ] 🔥 Generic functional utilities (Map, Filter, Reduce, GroupBy)
- [ ] 🔥 Generic middleware/decorator pattern
- [ ] 🔥 Monadic types (Result[T], Optional[T])
- [ ] Limitations: no method-level type parameters, no specialization
- [ ] Implementation: monomorphization vs dictionaries (Go uses both)

---

## Phase 6 — Testing Mastery

### 6.1 Basics
- [ ] `testing.T` — `Error`, `Fatal`, `Log`, `Skip`
- [ ] `testing.B` — benchmarks with `b.N`
- [ ] Test file naming: `*_test.go`
- [ ] Test function naming: `Test<Name>(t *testing.T)`
- [ ] `go test -v -run=<regex>`
- [ ] `go test -count=1` (disable test caching)

### 6.2 Patterns
- [ ] Table-driven tests (struct slice + loop)
- [ ] 🔥 Subtests with `t.Run()` (individually runnable)
- [ ] 🔥 `t.Parallel()` — concurrent test execution
- [ ] `t.Helper()` — correct error line reporting
- [ ] `t.Cleanup()` — guaranteed cleanup (even on panic)
- [ ] `t.TempDir()` — auto-cleaned temp directory
- [ ] `TestMain(m *testing.M)` — global setup/teardown
- [ ] Golden file testing (snapshot testing)
- [ ] Example tests (`func Example<Name>()` with `// Output:`)

### 6.3 Mocking
- [ ] Interface-based mocking (manual mock structs)
- [ ] Function field pattern (`SaveFn func(...)`)
- [ ] Tracking calls for assertion
- [ ] `httptest.NewServer` — mock HTTP servers
- [ ] `testing/fstest.MapFS` — mock filesystem
- [ ] Clock interface for time-dependent code
- [ ] `io.NopCloser`, `strings.NewReader` for io mocking

### 6.4 Fuzzing (Go 1.18+)
- [ ] `testing.F` and `f.Fuzz()`
- [ ] Seed corpus with `f.Add()`
- [ ] `go test -fuzz=<name> -fuzztime=<duration>`
- [ ] Invariant checking in fuzz functions
- [ ] 🔥 Crash corpus in `testdata/fuzz/`
- [ ] Round-trip fuzzing (parse → serialize → parse)

### 6.5 Coverage & Profiling
- [ ] `go test -cover`
- [ ] `go test -coverprofile=coverage.out`
- [ ] `go tool cover -html=coverage.out`
- [ ] `go test -cpuprofile=cpu.prof -memprofile=mem.prof`
- [ ] Race detector: `go test -race`
- [ ] 🔥 `benchstat` — statistically compare benchmarks
- [ ] 🔥 `go test -bench=. -benchmem` — allocations per op

### 6.6 Integration Testing
- [ ] Build tags for integration tests (`//go:build integration`)
- [ ] Custom test flags (`-integration`)
- [ ] `t.Skip()` for conditional tests
- [ ] Testcontainers (database/redis in Docker for tests)
- [ ] Whitebox vs blackbox tests (`package foo` vs `package foo_test`)
- [ ] `export_test.go` pattern (exposing unexported for external tests)

---

## Phase 7 — Memory & Runtime Internals

### 7.1 Memory Model
- [ ] Stack vs heap allocation
- [ ] 🔥 Escape analysis (`go build -gcflags="-m"`)
- [ ] 🔥 What causes escape: returning pointers, interface boxing, closures, large objects
- [ ] 🔥 Preventing escape: pass pointers down, not up
- [ ] Value types vs reference types
- [ ] `unsafe.Sizeof`, `unsafe.Alignof`, `unsafe.Offsetof`
- [ ] 🔥 Struct field ordering for minimal padding

### 7.2 Garbage Collector
- [ ] 🔥 Tri-color mark-and-sweep (concurrent collector)
- [ ] GC phases: mark setup (STW), marking, mark termination (STW), sweeping
- [ ] Write barrier (ensures correctness during concurrent marking)
- [ ] `GOGC` — GC trigger threshold
- [ ] `GOMEMLIMIT` (Go 1.19+) — soft memory limit
- [ ] 🔥 GC tuning for latency vs throughput
- [ ] `runtime.GC()` — force GC
- [ ] `runtime.ReadMemStats()` — heap stats
- [ ] `GODEBUG=gctrace=1` — GC trace output
- [ ] `runtime.SetFinalizer` — last-resort cleanup
- [ ] `runtime.KeepAlive` — prevent premature collection
- [ ] 🔥 Pointer vs non-pointer types and GC scan cost

### 7.3 Scheduler (GMP Model)
- [ ] 🔥 G (Goroutine), M (OS thread), P (Logical processor)
- [ ] Local run queue (LRQ) vs global run queue (GRQ)
- [ ] 🔥 Work stealing algorithm
- [ ] `GOMAXPROCS` tuning
- [ ] Cooperative vs preemptive scheduling (Go 1.14+: async preemption via SIGURG)
- [ ] System calls and M handoff
- [ ] `runtime.Gosched()` — yield
- [ ] `runtime.LockOSThread()` — pin to OS thread
- [ ] `runtime.Goexit()` — exit goroutine running defers
- [ ] Network poller (epoll/kqueue integration)
- [ ] 🔥 `GODEBUG=schedtrace=1000` — scheduler trace

### 7.4 Stack Management
- [ ] 🔥 Goroutine stack growth (contiguous stacks, copy-on-grow)
- [ ] Initial stack size (~2-8KB)
- [ ] Stack shrinking (during GC)
- [ ] `runtime.Stack()` — dump goroutine stacks
- [ ] `GOTRACEBACK` environment variable
- [ ] `SIGQUIT` — dump all goroutines

---

## Phase 8 — Reflection & Unsafe

### 8.1 Reflection
- [ ] `reflect.Type` — type introspection
- [ ] `reflect.Value` — value inspection and modification
- [ ] `reflect.Kind` — underlying type kind
- [ ] `reflect.TypeOf()`, `reflect.ValueOf()`
- [ ] Struct field iteration and tag parsing
- [ ] `CanSet()`, `CanAddr()` — modifiability
- [ ] Dynamic function calls (`reflect.Value.Call()`)
- [ ] 🔥 Creating types at runtime (`reflect.StructOf`, `reflect.SliceOf`)
- [ ] 🔥 `reflect.Select` — dynamic select statement
- [ ] 🔥 Building a struct validator with reflection
- [ ] Performance: reflection is 10-100x slower than direct code

### 8.2 unsafe Package
- [ ] `unsafe.Pointer` — generic pointer type
- [ ] 🔥 Legal pointer conversion rules (4 rules from the spec)
- [ ] 🔥 Never store `uintptr` in a variable (GC doesn't track it)
- [ ] `unsafe.Sizeof`, `unsafe.Alignof`, `unsafe.Offsetof`
- [ ] `unsafe.Add` — pointer arithmetic (Go 1.17+)
- [ ] `unsafe.Slice` — create slice from pointer (Go 1.17+)
- [ ] `unsafe.String`, `unsafe.StringData` (Go 1.20+)
- [ ] 🔥 Zero-copy string ↔ []byte conversion
- [ ] 🔥 Inspecting interface internals (iface/eface layout)
- [ ] Type punning (reinterpret bits as different type)

---

## Phase 9 — Build System & Toolchain

### 9.1 Build
- [ ] `go build`, `go install`, `go run`
- [ ] Build tags / constraints (`//go:build linux && amd64`)
- [ ] File naming conventions (`_linux.go`, `_test.go`)
- [ ] Cross-compilation (`GOOS`, `GOARCH`)
- [ ] `CGO_ENABLED=0` for static binaries
- [ ] `-ldflags` — inject version info (`-X main.version=1.0.0`)
- [ ] `-ldflags="-s -w"` — strip debug info (smaller binary)
- [ ] `-trimpath` — reproducible builds
- [ ] Build modes: `default`, `pie`, `c-shared`, `c-archive`, `plugin`
- [ ] `//go:embed` — embed files in binary (Go 1.16+)

### 9.2 Go Generate
- [ ] `//go:generate` directive
- [ ] `stringer` — String() for enums
- [ ] `mockgen` — interface mocks
- [ ] `protoc-gen-go` — Protocol Buffer code
- [ ] `sqlc` — type-safe SQL
- [ ] `wire` — compile-time DI
- [ ] `enumer` — enum with JSON/text marshaling
- [ ] `ent` — entity framework

### 9.3 Code Analysis (AST)
- [ ] `go/token` — token types and positions
- [ ] `go/scanner` — lexical scanning
- [ ] `go/parser` — AST parsing
- [ ] `go/ast` — AST traversal and inspection
- [ ] `go/printer` — AST to source code
- [ ] `go/format` — `gofmt` formatting
- [ ] `go/types` — type checking
- [ ] 🔥 AST transformation (rewrite code programmatically)
- [ ] 🔥 Building custom linters with `go/analysis`

### 9.4 Compiler Directives
- [ ] `//go:noinline` — prevent inlining
- [ ] `//go:nosplit` — no stack split check
- [ ] `//go:norace` — skip race detector
- [ ] `//go:noescape` — hint that args don't escape
- [ ] 🔥 `//go:linkname` — access unexported symbols across packages
- [ ] `//go:build` — conditional compilation

### 9.5 Tooling
- [ ] `gofmt` / `goimports` — formatting
- [ ] `go vet` — static analysis
- [ ] `staticcheck` — advanced analysis (honnef.co/go/tools)
- [ ] `golangci-lint` — meta-linter (runs many linters)
- [ ] `dlv` (Delve) — debugger
- [ ] `gopls` — language server
- [ ] `go doc`, `godoc` — documentation
- [ ] `go list -m all` — dependency tree

---

## Phase 10 — Performance & Profiling

### 10.1 Benchmarking
- [ ] `testing.B` and `b.N`
- [ ] `b.ResetTimer()`, `b.StopTimer()`, `b.StartTimer()`
- [ ] `b.ReportAllocs()`
- [ ] Sub-benchmarks (`b.Run`)
- [ ] Table-driven benchmarks
- [ ] 🔥 Global sink pattern (prevent dead-code elimination)
- [ ] `benchstat` for statistical comparison

### 10.2 Profiling
- [ ] `go test -cpuprofile`, `-memprofile`, `-blockprofile`, `-mutexprofile`
- [ ] `go tool pprof` — interactive profiler
- [ ] `pprof` web UI (`-http=:8080`)
- [ ] 🔥 `net/http/pprof` — production profiling endpoint
- [ ] `go tool trace` — execution tracer
- [ ] 🔥 Flame graphs for CPU profiling
- [ ] Goroutine profiling (detect leaks and contention)

### 10.3 Optimization Techniques
- [ ] Pre-allocate slices and maps (`make([]T, 0, n)`)
- [ ] `strings.Builder` over `+` concatenation
- [ ] `sync.Pool` for hot-path object reuse
- [ ] 🔥 Avoid interface boxing in hot paths
- [ ] 🔥 Stack allocation vs heap (escape analysis control)
- [ ] Array vs slice for fixed sizes
- [ ] 🔥 Struct-of-arrays vs array-of-structs for cache efficiency
- [ ] `io.LimitReader` to prevent OOM
- [ ] 🔥 Bounds check elimination (BCE)
- [ ] 🔥 Inlining budget (~80 AST nodes, check with `-gcflags="-m"`)

---

## Phase 11 — Design Patterns & Architecture

### 11.1 Creational
- [ ] Functional options pattern (`WithXxx()` functions)
- [ ] Factory functions (`NewXxx()`)
- [ ] Builder pattern (method chaining)
- [ ] 🔥 `sync.Once` for singleton (not `init()`)
- [ ] Object pool (`sync.Pool`)

### 11.2 Structural
- [ ] Composition via struct embedding
- [ ] Interface decoration / wrapping (middleware)
- [ ] Adapter (wrap external API to internal interface)
- [ ] Facade (simplify complex subsystem)

### 11.3 Behavioral
- [ ] Strategy (interface with swappable implementations)
- [ ] Observer / Pub-Sub (event bus)
- [ ] Iterator (`for range`, custom iterators with channels or callbacks)
- [ ] 🔥 Range-over-function iterators (Go 1.23+)
- [ ] Command (encapsulate operations)

### 11.4 Concurrency Patterns
- [ ] Worker pool
- [ ] Pipeline
- [ ] Fan-out / Fan-in
- [ ] 🔥 Circuit breaker
- [ ] 🔥 Retry with exponential backoff + jitter
- [ ] 🔥 Singleflight (deduplicate concurrent requests)
- [ ] 🔥 Bulkhead (isolate failure domains)
- [ ] Rate limiter (token bucket, sliding window)
- [ ] Graceful shutdown with signal handling
- [ ] Health check pattern (liveness + readiness)

### 11.5 Clean Architecture
- [ ] Dependency injection via constructors (no frameworks)
- [ ] Repository pattern (data access abstraction)
- [ ] Service layer (business logic)
- [ ] Handler/controller layer (HTTP/gRPC)
- [ ] 🔥 Hexagonal architecture / Ports & Adapters
- [ ] Domain-driven design in Go

---

## Phase 12 — Production Go

### 12.1 HTTP APIs
- [ ] RESTful API design
- [ ] Middleware chains (logging, auth, recovery, CORS, rate-limit)
- [ ] Request validation
- [ ] Response formatting (JSON, error responses)
- [ ] API versioning
- [ ] OpenAPI / Swagger generation
- [ ] gRPC with Protocol Buffers
- [ ] gRPC-Gateway (REST → gRPC)
- [ ] GraphQL in Go

### 12.2 Databases
- [ ] `database/sql` — connection pool, Prepared statements
- [ ] `sql.DB.SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`
- [ ] PostgreSQL with `pgx`
- [ ] MySQL with `go-sql-driver/mysql`
- [ ] SQLite with `modernc.org/sqlite` (pure Go, no CGo)
- [ ] ORM: `GORM`, `ent`, `sqlboiler`
- [ ] Query builder: `squirrel`, `goqu`
- [ ] Type-safe SQL: `sqlc`
- [ ] Migrations: `goose`, `migrate`, `atlas`
- [ ] Redis: `go-redis/redis`
- [ ] MongoDB: `mongo-go-driver`
- [ ] 🔥 Connection pooling tuning
- [ ] 🔥 Transaction management patterns

### 12.3 Configuration
- [ ] Environment variables (`os.Getenv`)
- [ ] `viper` (multi-source config)
- [ ] `envconfig` (struct-based env config)
- [ ] 🔥 Configuration with defaults + env override + file + flags
- [ ] 12-factor app configuration principles

### 12.4 Observability
- [ ] Structured logging (`log/slog`)
- [ ] Distributed tracing (OpenTelemetry)
- [ ] Metrics (Prometheus client)
- [ ] Health endpoints (`/health`, `/ready`)
- [ ] 🔥 Request-scoped logging with trace IDs
- [ ] Span creation and propagation

### 12.5 Deployment
- [ ] Multi-stage Docker builds
- [ ] Minimal base images (`scratch`, `distroless`, `alpine`)
- [ ] Kubernetes readiness/liveness probes
- [ ] 🔥 Graceful shutdown (SIGTERM → drain → close)
- [ ] Graceful restart (zero-downtime deploys)
- [ ] `uber-go/automaxprocs` (container CPU awareness)

### 12.6 Security
- [ ] Input validation and sanitization
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention (`html/template` auto-escaping)
- [ ] CSRF protection
- [ ] Rate limiting
- [ ] JWT / OAuth2 / OIDC
- [ ] TLS configuration (min version, cipher suites)
- [ ] Secrets management (never in code or env vars)
- [ ] CORS configuration
- [ ] `crypto/rand` for tokens (never `math/rand`)

---

## Phase 13 — CGo & System Programming

- [ ] CGo basics: `import "C"`, `C.int`, `C.CString`
- [ ] Memory management: `C.malloc`, `C.free`
- [ ] 🔥 CGo call overhead (~100x slower than Go calls)
- [ ] Passing Go pointers to C (restrictions)
- [ ] `runtime.KeepAlive` with CGo
- [ ] Callbacks from C to Go
- [ ] Pure Go alternatives: `purego`, `ebitengine/purego`
- [ ] `syscall` and `golang.org/x/sys` packages
- [ ] Linux namespaces and cgroups from Go
- [ ] `mmap` from Go
- [ ] eBPF from Go (`cilium/ebpf`)

---

## Phase 14 — Advanced Topics

### 14.1 Code Generation
- [ ] `go/ast` for parsing and transforming Go code
- [ ] Template-based code generation (`text/template`)
- [ ] `go generate` workflow
- [ ] 🔥 Writing custom code generators
- [ ] 🔥 Writing custom linters with `go/analysis`

### 14.2 WebAssembly
- [ ] `GOOS=js GOARCH=wasm` compilation
- [ ] `syscall/js` — JavaScript interop
- [ ] `GOOS=wasip1 GOARCH=wasm` — WASI target (Go 1.21+)
- [ ] TinyGo for smaller Wasm binaries

### 14.3 Plugin System
- [ ] `plugin` package (Linux/macOS only)
- [ ] Limitations (same Go version, same dependencies)
- [ ] Alternative: HashiCorp `go-plugin` (process-based)
- [ ] Alternative: Wasm plugins (e.g., with `wazero`)

### 14.4 Iterators (Go 1.23+)
- [ ] `iter.Seq[V]` and `iter.Seq2[K, V]`
- [ ] Range-over-function
- [ ] `slices.Values`, `slices.All`, `maps.Keys`, `maps.Values`
- [ ] Writing custom iterators
- [ ] Pull-style iterators with `iter.Pull`

### 14.5 Weak Pointers (Go 1.24+)
- [ ] `weak.Pointer[T]`
- [ ] Use cases: caches, canonicalization maps
- [ ] `unique.Handle[T]` for interning

---

## Phase 15 — Ecosystem & Frameworks

### 15.1 Web Frameworks & Routers
- [ ] Standard library `net/http` (Go 1.22+ with patterns)
- [ ] `chi` — lightweight, idiomatic router
- [ ] `gin` — performance-focused
- [ ] `echo` — minimalist
- [ ] `fiber` — Express-inspired (uses fasthttp)
- [ ] When to use stdlib vs framework (stdlib is usually enough)

### 15.2 CLI Tools
- [ ] `cobra` — CLI framework
- [ ] `urfave/cli` — alternative CLI framework
- [ ] `bubbletea` — terminal UI framework
- [ ] `charm` ecosystem (lipgloss, bubbles, wish)

### 15.3 Essential Libraries
- [ ] `zap` / `zerolog` — high-performance structured logging
- [ ] `testify` — test assertions (use sparingly, stdlib is fine)
- [ ] `go-chi/chi` — HTTP router
- [ ] `gorilla/websocket` — WebSocket
- [ ] `nats` / `kafka-go` / `rabbitmq` — message queues
- [ ] `ristretto` — concurrent cache
- [ ] `go-playground/validator` — struct validation
- [ ] `golang.org/x/exp` — experimental packages
- [ ] `golang.org/x/sync` — errgroup, semaphore, singleflight
- [ ] `golang.org/x/time/rate` — rate limiter

---

## God-Level Projects to Build

After studying, build these to prove mastery:

1. **Custom HTTP framework** with middleware, routing, context propagation
2. **Distributed key-value store** with consistent hashing, replication, gossip protocol
3. **TCP proxy / Load balancer** with health checks, connection pooling, circuit breakers
4. **Custom database** (LSM tree or B+ tree based) with WAL and crash recovery
5. **Container runtime** (minimal Docker clone using Linux namespaces, cgroups, overlay fs)
6. **Compiler or interpreter** for a simple language, written in Go
7. **Distributed task queue** (like Celery/Sidekiq) with workers, retries, dead-letter
8. **Real-time event streaming system** with pub/sub, partitioning, consumer groups
9. **Service mesh sidecar proxy** (like a minimal Envoy, handling mTLS, retries, observability)
10. **Build your own Go linter / static analyzer** using `go/analysis`

---

## Mastery Verification Checklist

You're a Go god when you can:

- [ ] Read the Go runtime source code and understand the scheduler
- [ ] Profile a production service and reduce p99 latency by 10x
- [ ] Write zero-allocation hot paths verified by benchmarks
- [ ] Design APIs that other developers love using
- [ ] Debug a goroutine leak in production using pprof
- [ ] Explain why an interface with a nil pointer is not nil
- [ ] Build a lock-free data structure with atomic CAS
- [ ] Write a code generator that parses Go AST and emits code
- [ ] Deploy a Go service to Kubernetes with proper health checks and graceful shutdown
- [ ] Contribute a patch to the Go standard library or runtime

---

*Total topics: ~500+ | Estimated time to mastery: 12-18 months of dedicated practice*
*Last updated: April 2026*
