# Low-Level Design Mastery in Go

## What is Low-Level Design?

LLD is about translating high-level architecture into detailed class/struct design with:
- Clear responsibilities (who does what)
- Clean interfaces (how components talk)
- Extensibility (how to add features without breaking things)
- Testability (how to verify correctness)

## Roadmap

| # | Topic | File |
|---|-------|------|
| 1 | SOLID Principles | `01_solid_principles.go` |
| 2 | Creational Patterns | `02_creational_patterns.go` |
| 3 | Structural Patterns | `03_structural_patterns.go` |
| 4 | Behavioral Patterns | `04_behavioral_patterns.go` |
| 5 | Concurrency Patterns for LLD | `05_concurrency_lld.go` |
| 6 | LLD: Parking Lot System | `06_parking_lot.go` |
| 7 | LLD: Elevator System | `07_elevator_system.go` |
| 8 | LLD: Cache (LRU) | `08_lru_cache.go` |
| 9 | LLD: Rate Limiter | `09_rate_limiter.go` |
| 10 | LLD: Pub-Sub System | `10_pubsub_system.go` |
| 11 | LLD: Task Scheduler | `11_task_scheduler.go` |
| 12 | LLD: Snake Game | `12_snake_game.go` |

## Go's LLD Philosophy

Go is NOT Java. Go's approach to LLD:
- **Composition over Inheritance** — embed structs, don't extend classes
- **Implicit Interfaces** — satisfy interfaces without declaring it
- **Small Interfaces** — `io.Reader`, `io.Writer`, `fmt.Stringer` (1-2 methods)
- **Package-level encapsulation** — exported vs unexported (Capital vs lowercase)
- **Concurrency as first-class** — goroutines and channels ARE design tools
