# Low Level Design Mastery — Complete Java Curriculum

> **35+ standalone Java files** covering OOP, SOLID, 17 Design Patterns, 10 Case Studies, Concurrency, UML, and Design Principles.
> Every file compiles independently: `javac File.java && java File`

## Learning Roadmap

```
                    ┌──────────────────────────┐
                    │   GOD-LEVEL LLD SKILL    │
                    └────────────┬─────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌────────▼────────┐   ┌─────────▼─────────┐   ┌────────▼────────┐
│  CASE STUDIES   │   │ KEY CONCEPTS       │   │ DESIGN          │
│  (10 systems)   │   │ (Concurrency, UML, │   │ PATTERNS        │
│                 │   │  Principles)        │   │ (17 patterns)   │
└────────┬────────┘   └─────────┬─────────┘   └────────┬────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                      ┌──────────▼───────────┐
                      │  SOLID PRINCIPLES    │
                      └──────────┬───────────┘
                                 │
                      ┌──────────▼───────────┐
                      │  OOP FUNDAMENTALS    │
                      │  (The Foundation)    │
                      └──────────────────────┘
```

---

## Module 1: OOP Fundamentals (Start Here!)

Study in order. Each file has theory + BAD example + GOOD example + real-world analogies.

| # | Topic | File | Key Concepts |
|---|-------|------|-------------|
| 1 | Classes & Objects | `_1_oops_fundamentals/_1_classes_and_objects/` | Fields, methods, constructors, `this`, `static`, `final`, equals/hashCode |
| 2 | Encapsulation | `_1_oops_fundamentals/_2_encapsulation/` | Access modifiers, getters/setters, validation, immutability |
| 3 | Inheritance | `_1_oops_fundamentals/_3_inheritance/` | `extends`, `super`, abstract classes, constructor chaining |
| 4 | Polymorphism | `_1_oops_fundamentals/_4_polymorphism/` | Overloading vs overriding, runtime dispatch, polymorphic collections |
| 5 | Abstraction | `_1_oops_fundamentals/_5_abstraction/` | Abstract classes vs interfaces, `default` methods, template method preview |
| 6 | Composition vs Inheritance | `_1_oops_fundamentals/_6_composition_vs_inheritance/` | HAS-A vs IS-A, strategy injection, favor composition |

---

## Module 2: SOLID Principles

The five pillars of clean, maintainable design.

| # | Principle | File | One-Liner |
|---|-----------|------|-----------|
| S | Single Responsibility | `_2_solid_principles/_1_single_responsibility/` | One class = one job = one reason to change |
| O | Open/Closed | `_2_solid_principles/_2_open_closed/` | Open for extension, closed for modification |
| L | Liskov Substitution | `_2_solid_principles/_3_liskov_substitution/` | Subtypes must be substitutable for base types |
| I | Interface Segregation | `_2_solid_principles/_4_interface_segregation/` | Many small interfaces > one fat interface |
| D | Dependency Inversion | `_2_solid_principles/_5_dependency_inversion/` | Depend on abstractions, not concretions |

---

## Module 3: Design Patterns (17 Patterns)

### Creational (How objects are created)

| # | Pattern | File | When to Use |
|---|---------|------|-------------|
| 1 | Singleton | `_3_design_patterns/_1_creational/_1_singleton/` | One instance globally (DB pool, config, logger) |
| 2 | Factory | `_3_design_patterns/_1_creational/_2_factory/` | Creating objects without specifying exact class |
| 3 | Builder | `_3_design_patterns/_1_creational/_3_builder/` | Complex objects with many optional params |
| 4 | Prototype | `_3_design_patterns/_1_creational/_4_prototype/` | Creating by cloning existing objects |

### Structural (How objects are composed)

| # | Pattern | File | When to Use |
|---|---------|------|-------------|
| 1 | Adapter | `_3_design_patterns/_2_structural/_1_adapter/` | Making incompatible interfaces work together |
| 2 | Decorator | `_3_design_patterns/_2_structural/_2_decorator/` | Adding behavior dynamically (layers) |
| 3 | Facade | `_3_design_patterns/_2_structural/_3_facade/` | Simplifying complex subsystem interfaces |
| 4 | Proxy | `_3_design_patterns/_2_structural/_4_proxy/` | Protection, caching, lazy loading, logging |
| 5 | Composite | `_3_design_patterns/_2_structural/_5_composite/` | Tree structures (file system, org chart, menus) |

### Behavioral (How objects communicate)

| # | Pattern | File | When to Use |
|---|---------|------|-------------|
| 1 | Strategy | `_3_design_patterns/_3_behavioral/_1_strategy/` | Swappable algorithms at runtime |
| 2 | Observer | `_3_design_patterns/_3_behavioral/_2_observer/` | Event handling, notifications, pub-sub |
| 3 | Command | `_3_design_patterns/_3_behavioral/_3_command/` | Undo/redo, queuing, logging operations |
| 4 | Template Method | `_3_design_patterns/_3_behavioral/_4_template_method/` | Algorithm skeleton with customizable steps |
| 5 | State | `_3_design_patterns/_3_behavioral/_5_state/` | Object behavior changes with internal state |
| 6 | Chain of Responsibility | `_3_design_patterns/_3_behavioral/_6_chain_of_responsibility/` | Pass request through handler chain (middleware, support) |
| 7 | Iterator | `_3_design_patterns/_3_behavioral/_7_iterator/` | Sequential access without exposing internals |
| 8 | Mediator | `_3_design_patterns/_3_behavioral/_8_mediator/` | Reduce many-to-many coupling (chat room, control tower) |

---

## Module 4: LLD Case Studies (10 Interview Problems)

Full implementations of the most commonly asked LLD interview questions.

| # | System | File | Patterns Used | Difficulty |
|---|--------|------|--------------|------------|
| 1 | Parking Lot | `_4_case_studies/_1_parking_lot/` | Singleton, Strategy, Enum | ★★☆ |
| 2 | Elevator System | `_4_case_studies/_2_elevator_system/` | Strategy, State, Observer | ★★★ |
| 3 | Library Management | `_4_case_studies/_3_library_management/` | Singleton, Builder, Observer | ★★☆ |
| 4 | Tic-Tac-Toe | `_4_case_studies/_4_tic_tac_toe/` | Board-Player-Game separation | ★☆☆ |
| 5 | Splitwise | `_4_case_studies/_5_splitwise/` | Strategy (split types), Singleton | ★★★ |
| 6 | ATM Machine | `_4_case_studies/_6_atm_machine/` | State, Chain of Responsibility | ★★★ |
| 7 | LRU Cache | `_4_case_studies/_7_lru_cache/` | HashMap+DLL, Strategy (eviction) | ★★★ |
| 8 | Snake & Ladder | `_4_case_studies/_8_snake_and_ladder/` | Strategy (dice), Game loop | ★★☆ |
| 9 | BookMyShow | `_4_case_studies/_9_book_my_show/` | Singleton, synchronized booking | ★★★ |
| 10 | Rate Limiter | `_4_case_studies/_10_rate_limiter/` | Strategy, Decorator | ★★★ |

---

## Module 5: Key Concepts

Advanced topics that separate good from god-level.

| # | Topic | File | Covers |
|---|-------|------|--------|
| 1 | Design Principles | `_5_key_concepts/_1_design_principles/` | DRY, KISS, YAGNI, Law of Demeter, Coupling/Cohesion, Tell Don't Ask |
| 2 | UML Class Diagrams | `_5_key_concepts/_2_uml_class_diagrams/` | Class notation, all 6 relationships, multiplicity, interview approach |
| 3 | Concurrency Patterns | `_5_key_concepts/_3_concurrency_patterns/` | Thread-safe Singleton, Producer-Consumer, ReadWriteLock, Thread Pool, Atomics |

---

## How to Run Any File

```bash
cd _1_oops_fundamentals/_1_classes_and_objects/
javac ClassesAndObjects.java && java ClassesAndObjects
```

Each file is standalone — compile and run directly.

---

## Study Strategy for Mastery

```
Week 1-2:   OOP Fundamentals (Module 1)
            → Read, run, modify each example
            → Write your own classes from scratch

Week 3:     SOLID Principles (Module 2)
            → Identify violations in your own code
            → Refactor old code to follow SOLID

Week 4-5:   Design Patterns (Module 3)
            → Implement each pattern from memory
            → Find patterns in Java's standard library
            → Know when NOT to use a pattern

Week 6-7:   Key Concepts (Module 5)
            → Draw UML diagrams for your patterns
            → Make your singleton thread-safe
            → Practice Producer-Consumer from scratch

Week 8-10:  Case Studies (Module 4)
            → Design on whiteboard/paper FIRST
            → Identify which patterns fit
            → Implement, then compare with provided solution
            → Re-implement from memory 3 days later

Week 11+:   Practice Mode
            → Pick random case study, solve in 45 min
            → Peer review / mock interviews
            → Build your OWN case studies
```

---

## Pattern Selection Cheat Sheet

```
"I need to create objects..."
  └─ Complex construction?           → Builder
  └─ Don't know exact type?          → Factory
  └─ Only one instance?              → Singleton
  └─ Clone existing?                 → Prototype

"I need to structure objects..."
  └─ Incompatible interface?         → Adapter
  └─ Add features dynamically?       → Decorator
  └─ Simplify complex system?        → Facade
  └─ Control access / lazy load?     → Proxy
  └─ Tree / part-whole hierarchy?    → Composite

"I need objects to communicate..."
  └─ Swap algorithms at runtime?     → Strategy
  └─ Notify multiple objects?        → Observer
  └─ Undo/redo operations?           → Command
  └─ Same skeleton, different steps? → Template Method
  └─ Behavior changes with state?    → State
  └─ Pass through handler chain?     → Chain of Responsibility
  └─ Traverse without exposing?      → Iterator
  └─ Reduce coupling between many?   → Mediator
```

---

## Key Principles to Internalize

1. **Program to an interface, not an implementation**
2. **Favor composition over inheritance**
3. **Encapsulate what varies**
4. **Strive for loosely coupled designs**
5. **Classes should be open for extension, closed for modification**
6. **Depend on abstractions, not on concretions**
7. **DRY — Don't Repeat Yourself**
8. **KISS — Keep It Simple, Stupid**
9. **YAGNI — You Ain't Gonna Need It**
10. **Law of Demeter — Talk only to your immediate friends**

---

## Interview Approach (5-Step Framework)

```
1. CLARIFY    → Ask requirements, constraints, scale
2. IDENTIFY   → Core entities, relationships, actions
3. DIAGRAM    → UML class diagram on whiteboard
4. PATTERNS   → Which design patterns fit naturally?
5. CODE       → Implement core classes, show extensibility
```

> *"The goal is not to memorize patterns. The goal is to recognize WHEN a design problem calls for a pattern and apply it naturally."*
