# Low Level Design Mastery — Go

A structured, hands-on repository to master Low Level Design using idiomatic Go.

## Roadmap (Follow in Order)

### Phase 1: Foundations
| # | Topic | Directory | Status |
|---|-------|-----------|--------|
| 1 | SOLID Principles | `01_solid/` | [ ] |
| 2 | Key OOP Concepts in Go (composition, embedding, interfaces) | `02_oop_in_go/` | [ ] |

### Phase 2: Design Patterns
| # | Topic | Directory | Status |
|---|-------|-----------|--------|
| 3 | Creational Patterns | `03_design_patterns/creational/` | [ ] |
| 4 | Structural Patterns | `03_design_patterns/structural/` | [ ] |
| 5 | Behavioral Patterns | `03_design_patterns/behavioral/` | [ ] |

### Phase 3: LLD Practice Problems (Interview-Level)
| # | Problem | Directory | Difficulty |
|---|---------|-----------|------------|
| 6 | Parking Lot System | `04_problems/parking_lot/` | Medium |
| 7 | Elevator System | `04_problems/elevator/` | Medium |
| 8 | Snake & Ladder Game | `04_problems/snake_ladder/` | Medium |
| 9 | Tic Tac Toe | `04_problems/tic_tac_toe/` | Easy |
| 10 | BookMyShow (Movie Booking) | `04_problems/bookmyshow/` | Hard |
| 11 | Splitwise (Expense Sharing) | `04_problems/splitwise/` | Hard |
| 12 | LRU Cache | `04_problems/lru_cache/` | Medium |
| 13 | Chess Game | `04_problems/chess/` | Hard |
| 14 | Vending Machine | `04_problems/vending_machine/` | Medium |
| 15 | ATM System | `04_problems/atm/` | Medium |
| 16 | Library Management | `04_problems/library/` | Medium |
| 17 | Hotel Booking | `04_problems/hotel_booking/` | Hard |
| 18 | Car Rental System | `04_problems/car_rental/` | Medium |
| 19 | Notification System | `04_problems/notification/` | Medium |
| 20 | Rate Limiter | `04_problems/rate_limiter/` | Medium |

### Phase 4: Advanced
| # | Topic | Directory | Status |
|---|-------|-----------|--------|
| 21 | CQRS Pattern | `05_advanced/cqrs/` | [ ] |
| 22 | Event Sourcing | `05_advanced/event_sourcing/` | [ ] |
| 23 | Concurrency Patterns in Go | `05_advanced/concurrency/` | [ ] |

## How to Study Each Topic

1. **Read** the `README.md` in each directory — understand the concept
2. **Study** the code — every file has detailed comments
3. **Run the tests** — `go test ./...` to verify understanding
4. **Modify & Break** — change code, see what breaks, understand why
5. **Solve problems** — attempt `04_problems/` without looking at solutions first

## Running

```bash
# Run all tests
go test ./...

# Run a specific package
go test ./01_solid/01_single_responsibility/...

# Run with verbose output
go test -v ./...
```
