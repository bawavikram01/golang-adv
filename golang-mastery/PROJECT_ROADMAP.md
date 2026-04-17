# Go God-Level Project Roadmap

> Theory gets you 60%. These projects get you to 100%.
> Each project targets specific skills. Build them IN ORDER.

---

## Phase 1: Foundation Projects (Week 1-2)
*Goal: Prove you can write clean, idiomatic Go from scratch*

### Project 1: `gofind` — File Search CLI
**Build a grep/find hybrid CLI tool**

```
gofind --pattern "TODO" --ext ".go" --dir ./myproject
gofind --pattern "func.*Error" --regex --dir . --ignore vendor,node_modules
```

**Skills tested:**
- `os`, `filepath.Walk` / `fs.WalkDir`
- `flag` or `cobra` for CLI args
- `regexp` for pattern matching
- `bufio.Scanner` for line-by-line reading
- Goroutines for parallel directory scanning
- Proper error handling with `%w` wrapping
- `//go:build` tags for OS-specific code (Windows vs Linux paths)

**Stretch:** Add `--replace` flag, colorized output with ANSI codes, `.gitignore` support

**Definition of done:**
- [ ] Handles 100K+ files without OOM
- [ ] Faster than `grep -r` for large dirs (benchmark it)
- [ ] `go vet`, `staticcheck` pass with zero warnings
- [ ] Has unit tests with `t.Run` subtests
- [ ] Published to GitHub with proper `go.mod`

---

### Project 2: `jsonq` — JSON Query Tool
**Build a jq-lite that reads JSON from stdin and queries it**

```bash
echo '{"users":[{"name":"alice","age":30}]}' | jsonq '.users[0].name'
cat config.json | jsonq '.database.host'
cat data.json | jsonq '.items[] | select(.price > 100)'
```

**Skills tested:**
- `encoding/json` deep usage (RawMessage, Decoder, Token)
- `os.Stdin` / `io.Reader` composition
- String parsing (build a mini query parser)
- `any` / type assertions / type switches
- Recursive data structure traversal
- Pipeline pattern (stdin → parse → query → stdout)

**Stretch:** Support YAML input, output formatting (pretty/compact/table)

**Definition of done:**
- [ ] Handles 500MB JSON via streaming (not loading into memory)
- [ ] Meaningful error messages for bad queries
- [ ] Tests with `testdata/` fixtures
- [ ] Zero allocations in hot path (verify with benchmarks)

---

## Phase 2: Concurrency Projects (Week 3-4)
*Goal: Master goroutines, channels, and real concurrent systems*

### Project 3: `goscrape` — Concurrent Web Crawler
**Crawl websites concurrently with configurable depth and parallelism**

```
goscrape --url https://example.com --depth 3 --workers 10 --timeout 30s
```

**Skills tested:**
- `net/http` client with timeouts, redirects, TLS
- `context.Context` for cancellation and timeout
- Worker pool pattern with buffered channels
- `sync.WaitGroup` + `sync.Mutex` for coordination
- `sync.Map` or mutex-guarded map for visited URLs
- Rate limiting with `time.Ticker` or `golang.org/x/time/rate`
- Graceful shutdown on SIGINT/SIGTERM
- `net/url` for URL parsing and resolution

**Architecture:**
```
                    ┌──────────┐
  URLs to crawl ──→ │ channel  │ ──→ Worker 1 ──→ Results channel ──→ Output
                    │ (queue)  │ ──→ Worker 2         │
                    │          │ ──→ Worker N          │
                    └──────────┘                       ▼
                         ▲                        New URLs
                         └────────────────────────────┘
```

**Definition of done:**
- [ ] Doesn't crash on malformed HTML/URLs
- [ ] Respects `robots.txt`
- [ ] Memory stays flat (no goroutine leaks — test with `runtime.NumGoroutine`)
- [ ] Tested with `-race` flag
- [ ] Can handle 10K pages without issues

---

### Project 4: `taskq` — In-Memory Job Queue
**Build an in-memory task queue with workers, retries, and priorities**

```go
q := taskq.New(taskq.Config{Workers: 5, MaxRetries: 3})
q.Enqueue(taskq.Task{Name: "send-email", Payload: data, Priority: High})
q.Start(ctx)
```

**Skills tested:**
- Channel-based worker pool
- `container/heap` for priority queue
- Retry with exponential backoff
- `context.Context` for cancellation
- `sync` primitives (Mutex, RWMutex, Cond)
- Graceful shutdown (drain queue before exit)
- Interface design (`TaskHandler` interface)
- Custom error types for retry vs permanent failure
- Metrics (processed, failed, in-flight counts)

**Stretch:** Add dead-letter queue, task scheduling (run at specific time), persistence to disk

**Definition of done:**
- [ ] Can handle 100K tasks/sec (benchmark it)
- [ ] No race conditions (`go test -race`)
- [ ] Clean shutdown drains all in-flight tasks
- [ ] 80%+ test coverage

---

## Phase 3: Network & API Projects (Week 5-7)
*Goal: Build production-quality HTTP services*

### Project 5: `goapi` — RESTful API with Everything
**Build a URL shortener service with full production features**

```
POST /api/shorten  {"url": "https://very-long-url.com/path"}  → {"short": "abc123"}
GET  /abc123       → 301 redirect to original URL
GET  /api/stats/abc123 → {"clicks": 42, "created": "...", "url": "..."}
```

**Skills tested:**
- `net/http` server (NOT a framework — raw stdlib first)
- Custom `http.Handler` and middleware chain
- `encoding/json` request/response handling
- `database/sql` with PostgreSQL or SQLite
- Database migrations (goose or golang-migrate)
- Input validation and sanitization
- Structured logging with `log/slog` (Go 1.21+)
- Configuration from env vars / config file
- Graceful shutdown (`http.Server.Shutdown`)
- Rate limiting per IP
- Request ID middleware, CORS, auth middleware
- Health check endpoint

**Project structure:**
```
goapi/
├── cmd/server/main.go       # entry point
├── internal/
│   ├── handler/              # HTTP handlers
│   ├── middleware/            # auth, logging, rate-limit
│   ├── model/                # domain types
│   ├── repository/           # database layer
│   └── service/              # business logic
├── migrations/               # SQL migrations
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

**Definition of done:**
- [ ] Handles 10K req/sec (benchmark with `wrk` or `hey`)
- [ ] Zero lint warnings from `golangci-lint`
- [ ] Integration tests with `httptest.Server`
- [ ] Docker image < 20MB (multi-stage build)
- [ ] Graceful shutdown tested
- [ ] SQL injection impossible (parameterized queries)

---

### Project 6: `gochat` — WebSocket Chat Server
**Real-time chat with rooms, typing indicators, message history**

**Skills tested:**
- WebSocket protocol (`gorilla/websocket` or `nhooyr.io/websocket`)
- Connection management (hub pattern)
- Concurrent map access for rooms
- Fan-out message broadcasting
- Heartbeat/ping-pong for dead connection detection
- JSON message serialization
- `context.Context` for connection lifecycle
- Memory management (bounded message history with ring buffer)

**Definition of done:**
- [ ] Supports 1000 concurrent connections
- [ ] Dead connections detected and cleaned up within 30s
- [ ] No goroutine leaks (monitor `runtime.NumGoroutine`)
- [ ] Messages delivered in order within a room

---

## Phase 4: Systems Programming (Week 8-10)
*Goal: Low-level Go mastery — this is where gods are made*

### Project 7: `gokv` — Key-Value Store with WAL
**Build a persistent key-value store from scratch**

```go
db, _ := gokv.Open("./data")
db.Set("user:1", []byte(`{"name":"alice"}`))
val, _ := db.Get("user:1")
db.Delete("user:1")
db.Close() // data survives restart
```

**Skills tested:**
- File I/O with `os.File`, `bufio`, `io`
- Write-Ahead Log (WAL) for crash safety
- Binary encoding (`encoding/binary`, `encoding/gob`)
- Memory-mapped files with `syscall.Mmap` (optional)
- `sync.RWMutex` for concurrent read/write
- Custom serialization format
- Compaction / garbage collection of old data
- Benchmarking with `testing.B`
- `unsafe.Pointer` for zero-copy tricks (optional)

**Architecture:**
```
Write path: Set("key", val) → WAL append → memtable update → ACK
Read path:  Get("key") → memtable lookup → (miss?) → scan WAL
Compaction:  Background goroutine merges WAL into sorted data file
```

**Definition of done:**
- [ ] Data survives process crash (kill -9)
- [ ] Read: < 1μs for memtable hit
- [ ] Write: < 10μs for WAL append
- [ ] Concurrent reads don't block each other
- [ ] Compaction runs without blocking reads
- [ ] Fuzz test key/value edge cases

---

### Project 8: `goproxy` — HTTP Reverse Proxy
**Build an nginx-lite reverse proxy with load balancing**

```yaml
# config.yml
listen: ":8080"
backends:
  - name: api
    prefix: /api/
    servers:
      - http://localhost:3001
      - http://localhost:3002
    strategy: round-robin
    health_check: /health
    timeout: 5s
```

**Skills tested:**
- `net/http` transport layer, `httputil.ReverseProxy`
- TCP connection pooling
- Load balancing algorithms (round-robin, least-connections, weighted)
- Health checking with background goroutines
- Circuit breaker pattern
- Request/response streaming (no buffering large bodies)
- TLS termination
- Access logging with `log/slog`
- Hot config reload with `SIGHUP`
- Metrics endpoint (request count, latency histogram, error rate)
- `pprof` integration for profiling in production

**Definition of done:**
- [ ] < 1ms added latency (p99)
- [ ] Handles backend failure gracefully (auto-failover)
- [ ] Zero downtime config reload
- [ ] Profile with `pprof`, optimize hot paths
- [ ] Load test: 50K req/sec passthrough

---

## Phase 5: Advanced & Open Source (Week 11-14)
*Goal: Operate at the level of Go core contributors*

### Project 9: `gotest` — Custom Test Framework
**Build a test runner that extends `go test`**

```go
func TestUserCreation(t *testing.T) {
    suite.Run(t, &UserSuite{})
}

type UserSuite struct {
    suite.Suite
    db *sql.DB
}

func (s *UserSuite) SetupTest()    { /* per-test setup */ }
func (s *UserSuite) TeardownTest() { /* per-test cleanup */ }
```

**Skills tested:**
- `testing` package internals
- `reflect` for discovering test methods
- `go/ast` and `go/parser` for code analysis
- `os/exec` for running subprocesses
- Interface design (test lifecycle hooks)
- `text/template` for report generation
- Plugin architecture (formatters, reporters)

---

### Project 10: `golsp` — Mini Language Server
**Build a language server for a simple config language**

**Skills tested:**
- LSP protocol over JSON-RPC
- `go/ast`, `go/parser`, `go/token` for code analysis
- Concurrent document handling
- `context.Context` for request cancellation
- TCP/stdio communication
- Code generation with `go generate`

---

## Phase 6: Read The Source (Ongoing)
*This is what separates good from god-level*

### Source Code Reading List (in order):
```
1. errors         — tiny, teaches interface design
2. sync.Once      — 30 lines, atomic + mutex pattern
3. sync.WaitGroup — channels + atomic + race-free design
4. context        — 500 lines, cancellation tree
5. net/http       — Server, Handler, ServeMux, Transport
6. encoding/json  — reflection-heavy, real-world complexity
7. database/sql   — connection pooling, driver interface
8. sync.Map       — lock-free concurrent map
9. runtime/proc.go — the Go scheduler itself
```

**How to read:**
```bash
# Find the source
go env GOROOT
# Read it
less $(go env GOROOT)/src/sync/once.go
```

For each package:
- [ ] Read every line, understand every line
- [ ] Write notes explaining the design decisions
- [ ] Try to reimplement key parts from memory
- [ ] Read the tests too — they show edge cases

---

## Progress Tracker

| # | Project | Status | Key Learning |
|---|---------|--------|-------------|
| 1 | gofind (File Search CLI) | ⬜ | io, fs, goroutines, CLI |
| 2 | jsonq (JSON Query) | ⬜ | json, streaming, parsing |
| 3 | goscrape (Web Crawler) | ⬜ | concurrency, http client, context |
| 4 | taskq (Job Queue) | ⬜ | channels, sync, design patterns |
| 5 | goapi (REST API) | ⬜ | http server, sql, middleware, production |
| 6 | gochat (WebSocket Chat) | ⬜ | websocket, fan-out, connection mgmt |
| 7 | gokv (Key-Value Store) | ⬜ | file I/O, binary, WAL, systems |
| 8 | goproxy (Reverse Proxy) | ⬜ | networking, load balancing, profiling |
| 9 | gotest (Test Framework) | ⬜ | reflect, AST, testing internals |
| 10 | golsp (Language Server) | ⬜ | AST, protocol, code analysis |

---

## Rules For Every Project

1. **No tutorials.** Design the architecture yourself. Get stuck. Unstick yourself.
2. **No frameworks** until you've built it with stdlib first.
3. **Benchmark everything.** If you can't measure it, you don't understand it.
4. **Run `go vet`, `staticcheck`, `golangci-lint`** on every commit.
5. **Write tests first** for critical paths. Fuzz test edge cases.
6. **Always run with `-race`** during development.
7. **Profile with `pprof`** at least once per project.
8. **Read error messages carefully.** Go's compiler errors are precise.
9. **`git commit` working code often.** Don't lose progress.
10. **When stuck > 2 hours:** Read the stdlib source for how Go itself solves it.
