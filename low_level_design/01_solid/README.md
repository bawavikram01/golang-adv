# SOLID Principles in Go

SOLID is the foundation of good low-level design. Every pattern and every
well-designed system follows these principles.

| # | Principle | One-Liner | Go Idiom |
|---|-----------|-----------|----------|
| S | Single Responsibility | A struct/package does ONE thing | Small packages, small interfaces |
| O | Open/Closed | Open for extension, closed for modification | Interfaces + composition |
| L | Liskov Substitution | Subtypes must be substitutable | Interface contracts + tests |
| I | Interface Segregation | Don't force unused methods | Small interfaces (io.Reader, io.Writer) |
| D | Dependency Inversion | Depend on abstractions, not concretions | Accept interfaces, return structs |

## Go's Advantage

Go naturally encourages SOLID through:
- **Implicit interfaces** — no `implements` keyword, just satisfy the contract
- **Composition over inheritance** — struct embedding, not class hierarchies
- **Small interfaces** — stdlib uses 1-2 method interfaces extensively
- **Package-level encapsulation** — unexported types enforce boundaries
