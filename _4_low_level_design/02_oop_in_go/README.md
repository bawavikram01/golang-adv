# OOP Concepts in Go

Go is NOT an OOP language in the traditional sense. It has NO:
- Classes
- Inheritance
- Constructors (as language feature)
- `this`/`self` keywords

But it achieves OOP goals through:

## 1. Encapsulation → Exported/Unexported

```go
type User struct {
    Name  string // Exported (public)
    email string // unexported (private to package)
}
```

## 2. Abstraction → Interfaces

```go
type Shape interface {
    Area() float64
}
// Any struct with Area() satisfies Shape — no "implements" keyword
```

## 3. Inheritance → Composition + Embedding

```go
type Animal struct { Name string }
func (a Animal) Speak() string { return "..." }

type Dog struct {
    Animal // embedded — Dog "inherits" Speak()
    Breed string
}
// dog.Speak() works directly
```

## 4. Polymorphism → Interfaces

```go
func PrintArea(s Shape) { fmt.Println(s.Area()) }
// Works with Circle, Rectangle, Triangle — any Shape
```

## Key Go Idioms for LLD

| OOP Concept | Go Equivalent |
|-------------|---------------|
| Class | struct |
| Method | func with receiver |
| Constructor | `NewXxx()` function |
| Interface | Implicit interface |
| Inheritance | Embedding |
| Abstract class | Interface + partial impl struct |
| Getter/Setter | Direct field access (exported) or methods |
| Private | Unexported (lowercase) |
| Public | Exported (Uppercase) |
