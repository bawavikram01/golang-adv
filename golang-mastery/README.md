# Go Mastery — Advanced Golang Learning Path

> From intermediate to god-level Go. Every file is runnable, heavily commented,
> and teaches concepts used in production systems at scale.

## How to Use

```bash
cd golang-mastery

# Run any lesson directly:
go run ./01-memory-model/01_stack_vs_heap.go

# Run with escape analysis output:
go run -gcflags="-m -m" ./01-memory-model/01_stack_vs_heap.go

# Run benchmarks:
go test -bench=. -benchmem ./06-performance-profiling/

# Run with GC tracing:
GODEBUG=gctrace=1 go run ./01-memory-model/03_gc_internals.go
```

---

## Curriculum

### 01 — Memory Model & Internals
**Goal:** Understand WHERE Go allocates memory and WHY it matters.

| File | Topic |
|------|-------|
| `01_stack_vs_heap.go` | Stack vs heap allocation, escape analysis rules |
| `02_escape_analysis.go` | Preventing escapes, sync.Pool, inlining budget |
| `03_gc_internals.go` | Tri-color GC, GOGC/GOMEMLIMIT tuning, finalizers |

**Key commands:**
```bash
go run -gcflags="-m" ./01-memory-model/01_stack_vs_heap.go
GODEBUG=gctrace=1 go run ./01-memory-model/03_gc_internals.go
```

---

### 02 — Advanced Concurrency
**Goal:** Master every concurrency pattern used in production Go.

| File | Topic |
|------|-------|
| `01_channel_patterns.go` | Generator, fan-out/fan-in, pipeline, semaphore, rate limiter, tee, or-done |
| `02_sync_primitives.go` | RWMutex, sync.Once, sync.Map, sync.Cond, atomics, lock-free stack |
| `03_context_deep_dive.go` | Cascading cancellation, deadlines, values, AfterFunc, CancelCause, graceful shutdown |

---

### 03 — Reflection & Unsafe
**Goal:** Break Go's type system (carefully) for metaprogramming.

| File | Topic |
|------|-------|
| `01_reflection.go` | Type introspection, value modification, struct validator, dynamic calls, dynamic types |
| `02_unsafe.go` | Struct padding, pointer arithmetic, zero-copy string↔[]byte, interface internals |

---

### 04 — Advanced Generics
**Goal:** Write type-safe, reusable code with Go 1.18+ generics.

| File | Topic |
|------|-------|
| `01_generics.go` | Custom constraints, Result/Optional types, generic BST, functional utilities (Map/Filter/Reduce), middleware chains, type-safe pool |

---

### 05 — Interface Internals
**Goal:** Design elegant APIs with Go's most powerful abstraction.

| File | Topic |
|------|-------|
| `01_interfaces.go` | iface/eface internals, nil gotcha, small interfaces, functional options, decorator pattern, capability detection |

---

### 06 — Performance Profiling
**Goal:** Measure, profile, and optimize like a systems programmer.

| File | Topic |
|------|-------|
| `01_benchmarks_test.go` | Benchmark patterns, string concat comparison, sub-benchmarks, pprof, BCE, compiler optimizations |

**Key commands:**
```bash
go test -bench=. -benchmem ./06-performance-profiling/
go test -bench=BenchmarkConcat -cpuprofile=cpu.prof ./06-performance-profiling/
go tool pprof -http=:8080 cpu.prof
```

---

### 07 — Advanced Patterns
**Goal:** Build resilient systems with battle-tested patterns.

| File | Topic |
|------|-------|
| `01_patterns.go` | Circuit breaker, worker pool, pub/sub event bus, pipeline with backpressure, retry with exponential backoff, singleflight, structured concurrency |

---

### 08 — Runtime & Scheduler
**Goal:** Understand the GMP model and runtime internals.

| File | Topic |
|------|-------|
| `01_runtime.go` | GMP model, GOMAXPROCS tuning, stack growth, Gosched/Goexit/LockOSThread, goroutine leak detection, memory ballast |

---

### 09 — Code Generation & AST
**Goal:** Write programs that write programs.

| File | Topic |
|------|-------|
| `01_codegen.go` | go/ast parsing, AST transformation, template-based code gen, go:generate, code analysis |

---

### 10 — Production Systems
**Goal:** Ship Go that runs at scale without waking you up at 3AM.

| File | Topic |
|------|-------|
| `01_production.go` | Structured errors, slog logging, graceful shutdown, health checks, middleware stack, DI without frameworks, configuration |

---

## Learning Order (Recommended)

```
01 Memory Model ──→ 08 Runtime & Scheduler ──→ 02 Advanced Concurrency
       ↓                                              ↓
05 Interface Internals ──→ 04 Advanced Generics ──→ 07 Advanced Patterns
       ↓                                              ↓
03 Reflection & Unsafe ──→ 09 Code Generation   ──→ 06 Performance Profiling
                                                       ↓
                                               10 Production Systems
```

## Quick Reference

| Want to... | Look at |
|------------|---------|
| Reduce GC pressure | 01 (sync.Pool, escape analysis) |
| Fix a goroutine leak | 08 (leak detection) |
| Build a worker pool | 07 (worker pool pattern) |
| Profile CPU/memory | 06 (pprof, benchmarks) |
| Design clean APIs | 05 (interfaces, functional options) |
| Handle errors properly | 10 (AppError, errors.Is/As) |
| Write generic code | 04 (constraints, data structures) |
| Inspect types at runtime | 03 (reflection) |
| Generate boilerplate | 09 (AST, templates) |
| Ship to production | 10 (shutdown, logging, middleware) |
