# Advanced Go Resources

> Curated list of the best resources to go from intermediate to god-level Go.
> Ordered by priority within each category.

---

## Official Resources (Free)

| Resource | URL | What You'll Learn |
|---|---|---|
| Go Specification | https://go.dev/ref/spec | The language bible — read cover to cover at least once |
| Go Memory Model | https://go.dev/ref/mem | Happens-before rules, synchronization guarantees |
| Go Blog | https://go.dev/blog | Deep dives from the core team |
| Effective Go | https://go.dev/doc/effective_go | Idiomatic patterns and conventions |
| Standard Library Docs | https://pkg.go.dev/std | Every package, every function, with examples |
| Go Diagnostics | https://go.dev/doc/diagnostics | Official profiling, tracing, debugging guide |
| Go Wiki | https://go.dev/wiki | Community-maintained guides and tips |
| Go FAQ | https://go.dev/doc/faq | Design decisions explained by the Go team |
| Go Tour | https://go.dev/tour | Interactive basics — best starting point |
| Module Reference | https://go.dev/ref/mod | Complete go.mod and module system docs |
| Release Notes | https://go.dev/doc/devel/release | What changed in each version |
| Go Runtime Source | https://github.com/golang/go/tree/master/src/runtime | Actual scheduler, GC, memory allocator code |

---

## Must-Read Official Blog Posts

### Concurrency
- **Go Concurrency Patterns** — https://go.dev/blog/pipelines
- **Advanced Go Concurrency Patterns** — https://go.dev/blog/io2013-talk-concurrency
- **Share Memory by Communicating** — https://go.dev/blog/codelab-share
- **Go Concurrency Patterns: Context** — https://go.dev/blog/context
- **Go Concurrency Patterns: Timing out, moving on** — https://go.dev/blog/concurrency-timeouts

### Internals & Performance
- **Getting to Go: The Journey of Go's Garbage Collector** — https://go.dev/blog/ismmkeynote
- **Go GC: Prioritizing Low Latency and Simplicity** — https://go.dev/blog/go15gc
- **A Guide to the Go Garbage Collector** — https://go.dev/doc/gc-guide
- **Profile-Guided Optimization** — https://go.dev/blog/pgo
- **An Introduction to Generics** — https://go.dev/blog/intro-generics
- **When to Use Generics** — https://go.dev/blog/when-generics
- **Structured Logging with slog** — https://go.dev/blog/slog

### Language Design
- **Go's Declaration Syntax** — https://go.dev/blog/declaration-syntax
- **Go Slices: Usage and Internals** — https://go.dev/blog/slices-intro
- **Strings, bytes, runes and characters in Go** — https://go.dev/blog/strings
- **Error Handling and Go** — https://go.dev/blog/error-handling-and-go
- **Working with Errors in Go 1.13** — https://go.dev/blog/go1.13-errors
- **The Laws of Reflection** — https://go.dev/blog/laws-of-reflection
- **Testable Examples in Go** — https://go.dev/blog/examples
- **Using Go Modules** — https://go.dev/blog/using-go-modules
- **Fuzzing is Beta Ready** — https://go.dev/blog/fuzz-beta
- **Range Over Function Types** — https://go.dev/blog/range-functions

---

## Books

### Tier 1 — Essential (Read These)

| Book | Author | Level | Focus |
|---|---|---|---|
| *The Go Programming Language* | Donovan & Kernighan | Intermediate | The "K&R of Go" — comprehensive foundation |
| *100 Go Mistakes and How to Avoid Them* | Teiva Harsanyi | Advanced | Real production pitfalls, covers every common mistake |
| *Concurrency in Go* | Katherine Cox-Buday | Advanced | Every concurrency pattern in depth |
| *Efficient Go* | Bartłomiej Płotka | Expert | Performance, profiling, optimization |

### Tier 2 — Highly Recommended

| Book | Author | Level | Focus |
|---|---|---|---|
| *Let's Go* | Alex Edwards | Intermediate | Production web apps |
| *Let's Go Further* | Alex Edwards | Advanced | Production APIs, auth, deployment |
| *Learning Go* (2nd ed.) | Jon Bodner | Intermediate-Advanced | Idiomatic Go with 1.21+ features |
| *Go in Action* | Kennedy, Ketelsen, Martin | Intermediate | Practical, well-structured |
| *Network Programming with Go* | Jan Newmarch | Advanced | TCP, HTTP, TLS, protocols |
| *Go with the Domain* | Three Dots Labs | Advanced | DDD and clean architecture in Go (free) |

### Tier 3 — Specialized

| Book | Author | Level | Focus |
|---|---|---|---|
| *Writing an Interpreter in Go* | Thorsten Ball | Advanced | Build a language in Go |
| *Writing a Compiler in Go* | Thorsten Ball | Advanced | Sequel — bytecode compiler + VM |
| *Distributed Services with Go* | Travis Jeffery | Advanced | Raft consensus, distributed systems |
| *Black Hat Go* | Steele, Patten, Kottmann | Advanced | Security tooling in Go |
| *Go Design Patterns* | Mario Castro | Intermediate | Classic patterns adapted for Go |

---

## Video Courses

| Course | Instructor / Platform | Level | Notes |
|---|---|---|---|
| Ultimate Go | Bill Kennedy (Ardan Labs) | Advanced-Expert | The best paid Go course. Period. |
| Go Class | Matt Holiday (YouTube, free) | Intermediate-Advanced | 50+ lectures, thorough |
| Boot.dev Backend Track | boot.dev | Intermediate | Project-based learning |
| Go: The Complete Developer's Guide | Stephen Grider (Udemy) | Beginner-Intermediate | Good for absolute beginners |
| justforfunc | Francesc Campoy (YouTube, free) | Intermediate-Advanced | Practical deep dives |

---

## Must-Watch Conference Talks

### Foundational Talks
| Talk | Speaker | Event | Topic |
|---|---|---|---|
| Concurrency is Not Parallelism | Rob Pike | Heroku Waza 2012 | Core mental model for goroutines |
| Go Proverbs | Rob Pike | GopherFest 2015 | Go philosophy in 19 proverbs |
| Simplicity is Complicated | Rob Pike | dotGo 2015 | Why Go chooses simplicity |

### Deep Dives
| Talk | Speaker | Event | Topic |
|---|---|---|---|
| The Scheduler Saga | Kavya Joshi | GopherCon 2018 | GMP model deep dive |
| Understanding Channels | Kavya Joshi | GopherCon 2017 | Channel internals |
| So You Wanna Go Fast? | Tyler Treat | GopherCon 2017 | Performance pitfalls |
| Understanding the Go Compiler | Jesús Espino | GopherCon EU | SSA, optimizations, inlining |
| Advanced Testing with Go | Mitchell Hashimoto | GopherCon 2017 | Table tests, golden files, test helpers |
| How I Write HTTP Web Services | Mat Ryer | GopherCon 2019 | Clean HTTP patterns |
| Rethinking Classical Concurrency Patterns | Bryan Mills | GopherCon 2018 | When NOT to use channels |
| Understanding Allocations | Jacob Walker | GopherCon 2022 | Escape analysis, stack vs heap |
| Building a Container from Scratch | Liz Rice | Container Camp 2016 | Linux namespaces in Go |

### GopherCon Full Archives
- **GopherCon YouTube**: https://www.youtube.com/@GopherAcademy
- **dotGo YouTube**: https://www.youtube.com/@dotconferences
- **GopherCon EU**: https://www.youtube.com/@GopherConEurope

---

## Blogs & Writers

### Tier 1 — Essential Reading

| Author / Site | URL | Focus |
|---|---|---|
| Dave Cheney | https://dave.cheney.net | Performance, internals, philosophy |
| Eli Bendersky | https://eli.thegreenplace.net/tag/go | Deep technical Go posts |
| Russ Cox | https://research.swtch.com | Module design, generics decisions, versioning |
| Three Dots Labs | https://threedots.tech/post | DDD, clean architecture, CQRS in Go |
| Go Blog (official) | https://go.dev/blog | Core team posts |

### Tier 2 — Highly Valuable

| Author / Site | URL | Focus |
|---|---|---|
| Applied Go | https://appliedgo.net | Practical how-tos |
| Boldly Go | https://boldlygo.tech | Advanced patterns |
| Alex Edwards | https://www.alexedwards.net/blog | Web development, security |
| Ardan Labs Blog | https://www.ardanlabs.com/blog | Architecture, performance |
| Bitfield Consulting (John Arundel) | https://bitfieldconsulting.com/posts | Testing, clean code, idiomatic Go |
| Willem Schots | https://www.willem.dev | Testing, practical patterns |
| Carl Johnson | https://blog.carlmjohnson.net | Generics, stdlib deep dives |
| Paschalis Tsilias | https://tpaschalis.me | Internals, compiler, runtime |

---

## Newsletters

| Newsletter | URL | Frequency |
|---|---|---|
| Go Weekly | https://golangweekly.com | Weekly |
| Golang Bridge (community) | https://forum.golangbridge.org | Ongoing discussions |
| Reddit r/golang | https://reddit.com/r/golang | Daily |

---

## GitHub Repos to Study

### Go Runtime & Language

| Repo | Why Study It |
|---|---|
| `golang/go` | The runtime: `src/runtime/proc.go` (scheduler), `mgc.go` (GC), `malloc.go` (allocator) |
| `golang/go/src/sync` | How Mutex, WaitGroup, Pool are actually implemented |
| `golang/go/src/net/http` | How the HTTP server and client work internally |

### Production-Grade Go Projects

| Repo | Why Study It |
|---|---|
| `docker/docker` (moby) | Large-scale Go systems design |
| `kubernetes/kubernetes` | How Google structures massive Go codebases |
| `cockroachdb/cockroach` | Distributed SQL database — Raft, MVCC, distributed txns |
| `etcd-io/etcd` | Raft consensus + distributed key-value store |
| `hashicorp/consul` | Gossip protocol, service discovery |
| `hashicorp/vault` | Secrets management, plugin architecture |
| `hashicorp/terraform` | Plugin system, HCL parsing, state management |
| `traefik/traefik` | Reverse proxy, clean middleware architecture |
| `prometheus/prometheus` | Monitoring, TSDB, pull-based metrics |
| `grafana/grafana` | Large Go + frontend codebase |
| `nats-io/nats-server` | High-performance message broker, zero-dep |
| `dgraph-io/badger` | LSM-tree key-value store in pure Go |
| `vitessio/vitess` | MySQL sharding proxy (used by YouTube) |
| `minio/minio` | S3-compatible object storage |
| `caddyserver/caddy` | Automatic HTTPS web server, module architecture |

### Small But Excellent Code to Study

| Repo | Why Study It |
|---|---|
| `uber-go/zap` | Zero-allocation logging — study for performance |
| `uber-go/automaxprocs` | Container CPU awareness |
| `sourcegraph/conc` | Structured concurrency patterns |
| `samber/lo` | Generic utility library (lodash for Go) |
| `charmbracelet/bubbletea` | Terminal UI framework — beautiful Go code |
| `go-chi/chi` | Lightweight router — clean idiomatic design |
| `benbjohnson/litestream` | SQLite replication — small, elegant codebase |

---

## Tools to Master

### Profiling & Debugging

| Tool | Purpose | Command |
|---|---|---|
| `go tool pprof` | CPU, memory, goroutine profiling | `go tool pprof http://localhost:6060/debug/pprof/heap` |
| `go tool trace` | Execution trace visualization | `go test -trace=trace.out && go tool trace trace.out` |
| `dlv` (Delve) | Interactive debugger | `dlv debug ./main.go` |
| `benchstat` | Statistically compare benchmarks | `benchstat old.txt new.txt` |
| Flame graphs | Visual CPU profiling | Via pprof web UI |

### Code Quality

| Tool | Purpose | Install |
|---|---|---|
| `golangci-lint` | Meta-linter (runs 50+ linters) | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| `staticcheck` | Best static analyzer for Go | `go install honnef.co/go/tools/cmd/staticcheck@latest` |
| `govulncheck` | Vulnerability scanner | `go install golang.org/x/vuln/cmd/govulncheck@latest` |
| `gofumpt` | Stricter gofmt | `go install mvdan.cc/gofumpt@latest` |
| `errcheck` | Find unchecked errors | Included in golangci-lint |
| `deadcode` | Find unreachable functions | `go install golang.org/x/tools/cmd/deadcode@latest` |

### Code Generation

| Tool | Purpose | Install |
|---|---|---|
| `stringer` | String() for enums | `go install golang.org/x/tools/cmd/stringer@latest` |
| `mockgen` | Interface mock generation | `go install go.uber.org/mock/mockgen@latest` |
| `sqlc` | Type-safe SQL | https://sqlc.dev |
| `wire` | Compile-time DI | `go install github.com/google/wire/cmd/wire@latest` |
| `protoc-gen-go` | Protocol Buffers | `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` |
| `buf` | Protobuf toolchain | https://buf.build |

---

## Community

| Community | URL |
|---|---|
| Gophers Slack | https://gophers.slack.com (invite: https://invite.slack.golangbridge.org) |
| Reddit r/golang | https://reddit.com/r/golang |
| Go Forum | https://forum.golangbridge.org |
| Go Discord | https://discord.gg/golang |
| Hacker News (Go tag) | https://news.ycombinator.com (search "Go") |

---

## Learning Priority Order

```
Level 1 — Foundation
  ① go.dev/tour (interactive basics)
  ② "The Go Programming Language" book (Donovan & Kernighan)
  ③ go.dev/doc/effective_go

Level 2 — Intermediate
  ④ "100 Go Mistakes" book (every common pitfall)
  ⑤ Go Blog posts (concurrency, errors, slices)
  ⑥ Bill Kennedy's Ardan Labs course or Matt Holiday's YouTube series

Level 3 — Advanced
  ⑦ "Concurrency in Go" book (Cox-Buday)
  ⑧ go.dev/ref/spec + go.dev/ref/mem (full spec + memory model)
  ⑨ GopherCon talks (Kavya Joshi on scheduler, Tyler Treat on perf)
  ⑩ Dave Cheney's blog archive

Level 4 — Expert
  ⑪ "Efficient Go" book (Płotka)
  ⑫ Go runtime source code (proc.go, mgc.go, malloc.go)
  ⑬ Study CockroachDB / etcd / Docker source
  ⑭ Build a god-level project from the roadmap

Level 5 — God
  ⑮ Contribute to the Go standard library
  ⑯ Write a proposal for golang/go
  ⑰ Give a GopherCon talk
```

---

*Last updated: April 2026*
