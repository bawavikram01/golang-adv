package lowleveldesign

// =============================================================================
// LLD INTERVIEW CHEAT SHEET — THE FRAMEWORK
// =============================================================================
//
// When you get an LLD question in an interview, follow this EXACT framework:
//
// ┌─────────────────────────────────────────────────────────────────┐
// │  STEP 1: CLARIFY REQUIREMENTS (2-3 min)                        │
// │  - What are the core use cases? (must-have vs nice-to-have)    │
// │  - What are the constraints? (users, scale, latency)           │
// │  - What are the boundaries? (what NOT to design)               │
// └─────────────────────────────────────────────────────────────────┘
// ┌─────────────────────────────────────────────────────────────────┐
// │  STEP 2: IDENTIFY CORE OBJECTS (3-5 min)                       │
// │  - What are the main entities/nouns?                           │
// │  - What are the relationships?                                 │
// │  - What are the cardinalities? (1:1, 1:N, M:N)                │
// └─────────────────────────────────────────────────────────────────┘
// ┌─────────────────────────────────────────────────────────────────┐
// │  STEP 3: DEFINE INTERFACES (5-7 min)                           │
// │  - What operations does each entity support?                   │
// │  - What are the input/output contracts?                        │
// │  - Keep interfaces SMALL (ISP)                                 │
// └─────────────────────────────────────────────────────────────────┘
// ┌─────────────────────────────────────────────────────────────────┐
// │  STEP 4: IMPLEMENT CORE LOGIC (15-20 min)                      │
// │  - Start with the HAPPY PATH                                   │
// │  - Add error handling                                          │
// │  - Make it thread-safe if needed                               │
// └─────────────────────────────────────────────────────────────────┘
// ┌─────────────────────────────────────────────────────────────────┐
// │  STEP 5: DISCUSS PATTERNS & TRADE-OFFS (5 min)                 │
// │  - Which design patterns did you use? WHY?                     │
// │  - What are the trade-offs?                                    │
// │  - How would you extend this?                                  │
// └─────────────────────────────────────────────────────────────────┘
//
// =============================================================================
// COMMON LLD INTERVIEW QUESTIONS + PATTERNS TO USE
// =============================================================================
//
// ┌──────────────────────────┬───────────────────────────────────────────────┐
// │ Problem                  │ Key Patterns                                  │
// ├──────────────────────────┼───────────────────────────────────────────────┤
// │ Parking Lot              │ Strategy, Factory, Observer                   │
// │ Elevator System          │ State, Strategy, Command                     │
// │ LRU Cache                │ HashMap + Doubly Linked List                  │
// │ Rate Limiter             │ Strategy (Token Bucket/Sliding Window)        │
// │ Pub/Sub System           │ Observer, Mediator                           │
// │ Task Scheduler           │ Priority Queue, Worker Pool, Strategy         │
// │ Vending Machine          │ State Machine                                │
// │ ATM Machine              │ State, Chain of Responsibility               │
// │ File System              │ Composite, Iterator                          │
// │ Logger Framework         │ Singleton, Strategy, Decorator               │
// │ Chess/Card Game          │ State, Strategy, Command                     │
// │ URL Shortener            │ Factory, Strategy (encoding)                 │
// │ Hotel Booking            │ Strategy (pricing), Observer (notifications) │
// │ Movie Ticket Booking     │ Strategy (seat allocation), State            │
// │ Library Management       │ Observer, Strategy (fine calculation)        │
// │ Social Media Feed        │ Observer, Strategy (ranking)                 │
// │ Notification Service     │ Observer, Factory, Bridge                    │
// │ Payment System           │ Strategy, Adapter, Facade                   │
// │ Search Autocomplete      │ Trie, Strategy (ranking)                    │
// │ Connection Pool          │ Object Pool, Semaphore                      │
// └──────────────────────────┴───────────────────────────────────────────────┘
//
// =============================================================================
// GO-SPECIFIC LLD PATTERNS
// =============================================================================
//
// 1. FUNCTIONAL OPTIONS (Builder alternative)
//    type Option func(*Config)
//    func New(opts ...Option) *Thing
//
// 2. INTERFACE ON CONSUMER SIDE
//    Define interfaces where they are USED, not where they are IMPLEMENTED.
//    package handler:
//      type UserStore interface { GetUser(id string) (*User, error) }
//    package storage:
//      type PostgresStore struct { ... }  // implements handler.UserStore
//
// 3. EMBED FOR COMPOSITION
//    type ReadWriteCloser struct {
//        io.Reader
//        io.Writer
//        io.Closer
//    }
//
// 4. CHANNELS AS SYNCHRONIZATION
//    - chan struct{} for signaling (done, quit, ready)
//    - Buffered channels as semaphores
//    - Select for multiplexing
//
// 5. CONTEXT FOR CANCELLATION
//    Every long-running operation should accept context.Context.
//    Wire cancellation through the entire call chain.
//
// 6. ERRORS AS VALUES
//    - Wrap errors with context: fmt.Errorf("operation: %w", err)
//    - Custom error types for different failure modes
//    - errors.Is() and errors.As() for checking
//
// 7. TABLE-DRIVEN TESTS
//    tests := []struct {
//        name string
//        input Input
//        want  Output
//    }{ ... }
//
// =============================================================================
// PRINCIPLES RANKED BY IMPORTANCE FOR GO LLD
// =============================================================================
//
// 1. ⭐⭐⭐⭐⭐ INTERFACE SEGREGATION — Small interfaces (1-3 methods)
// 2. ⭐⭐⭐⭐⭐ DEPENDENCY INVERSION — Accept interfaces, return structs
// 3. ⭐⭐⭐⭐   SINGLE RESPONSIBILITY — One struct, one job
// 4. ⭐⭐⭐⭐   OPEN/CLOSED — Extend via new types, not modifying old ones
// 5. ⭐⭐⭐     COMPOSITION OVER INHERITANCE — Embed, don't extend
// 6. ⭐⭐⭐     LISKOV SUBSTITUTION — Implementations honor contracts
//
// =============================================================================
// RED FLAGS IN YOUR LLD (Things that scream "bad design")
// =============================================================================
//
// ❌ God struct with 10+ fields and 15+ methods
// ❌ Concrete types as function parameters (should be interfaces)
// ❌ Exported global variables (hidden dependencies)
// ❌ init() for important setup (untestable)
// ❌ Mutex inside a channel-based system (pick one paradigm)
// ❌ Interface with 5+ methods (too fat)
// ❌ Package circular imports (wrong dependency direction)
// ❌ Deep nesting (more than 3 levels of if/for)
// ❌ Comments explaining WHAT code does (code should be self-documenting)
// ❌ No error handling (panic in library code)
//
// =============================================================================
// GREEN FLAGS IN YOUR LLD (Things that scream "good design")
// =============================================================================
//
// ✅ Interfaces defined at point of use (consumer side)
// ✅ Constructor functions (NewX) that return interfaces
// ✅ Small, focused packages (one concern per package)
// ✅ Table-driven tests with clear names
// ✅ context.Context threaded through call chains
// ✅ Error wrapping with %w for debugging
// ✅ Graceful shutdown with signal handling
// ✅ Bounded concurrency (worker pools, semaphores)
// ✅ Dependency injection via constructor params
// ✅ Exported interfaces, unexported implementations
