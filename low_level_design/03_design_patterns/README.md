# Design Patterns in Go

All 23 GoF (Gang of Four) patterns adapted for idiomatic Go.

Go doesn't have classes or inheritance, so patterns look different:
- **Factory** → constructor functions (`NewXxx`)
- **Abstract Factory** → interface returning interfaces
- **Inheritance** → embedding + interfaces
- **Decorator** → wrapping (middleware pattern)

## Creational Patterns
| Pattern | Purpose | Go Idiom |
|---------|---------|----------|
| Factory Method | Create objects without specifying exact type | `NewXxx()` functions returning interfaces |
| Abstract Factory | Create families of related objects | Interface with multiple `Create` methods |
| Builder | Construct complex objects step-by-step | Functional options / method chaining |
| Singleton | Ensure single instance | `sync.Once` |
| Prototype | Clone existing objects | `Clone()` method |

## Structural Patterns
| Pattern | Purpose | Go Idiom |
|---------|---------|----------|
| Adapter | Make incompatible interfaces work together | Wrapper struct implementing target interface |
| Decorator | Add behavior dynamically | Middleware / wrapping |
| Facade | Simplify complex subsystem | Single struct composing multiple services |
| Proxy | Control access to an object | Same interface, extra logic |
| Composite | Tree structures | Interface with slice of same interface |
| Bridge | Separate abstraction from implementation | Two interface hierarchies |

## Behavioral Patterns
| Pattern | Purpose | Go Idiom |
|---------|---------|----------|
| Strategy | Swap algorithms at runtime | Interface field |
| Observer | Publish-subscribe | Channels or callback slices |
| Command | Encapsulate requests as objects | Interface with `Execute()` |
| State | Behavior changes with state | Interface field that gets swapped |
| Chain of Responsibility | Pass request through handlers | Middleware chain |
| Template Method | Define skeleton, defer steps | Embed struct + interface for hooks |
| Iterator | Sequential access | Channels or `Next()` method |
