# C++ Mastery Roadmap — From Zero to God Level

## Overview
This roadmap takes you from absolute beginner to system-level mastery.
Each phase builds on the previous one. Do NOT skip phases.

---

## PHASE 1: Core Foundations (Weeks 1–4)

### 1.1 Environment & Compilation
- [ ] How C++ code becomes an executable (preprocessing → compilation → assembly → linking)
- [ ] g++ compiler flags (-Wall, -Wextra, -std=c++20, -O2, -g)
- [ ] Header files vs source files
- [ ] Translation units
- [ ] One Definition Rule (ODR)
- [ ] #include guards and #pragma once

### 1.2 Type System & Variables
- [ ] Fundamental types: int, char, float, double, bool, void
- [ ] Fixed-width integers: int8_t, int16_t, int32_t, int64_t
- [ ] size_t, ptrdiff_t
- [ ] Type modifiers: signed, unsigned, short, long
- [ ] Type conversions: implicit, explicit (static_cast, reinterpret_cast, const_cast, dynamic_cast)
- [ ] auto type deduction
- [ ] const and constexpr

### 1.3 Operators & Expressions
- [ ] Arithmetic, relational, logical, bitwise operators
- [ ] Assignment operators
- [ ] Operator precedence and associativity
- [ ] Short-circuit evaluation
- [ ] Comma operator
- [ ] sizeof operator
- [ ] Ternary operator

### 1.4 Control Flow
- [ ] if/else, switch
- [ ] for, while, do-while
- [ ] Range-based for loops
- [ ] break, continue, goto (and why to avoid goto)
- [ ] Structured bindings in if/switch (C++17)

### 1.5 Functions
- [ ] Declaration vs definition
- [ ] Pass by value, reference, pointer
- [ ] Default arguments
- [ ] Function overloading
- [ ] Inline functions
- [ ] constexpr functions
- [ ] Recursion
- [ ] Function pointers

### 1.6 Arrays & Strings
- [ ] C-style arrays and their decay to pointers
- [ ] std::array
- [ ] std::string and std::string_view
- [ ] C-strings (char arrays, null termination)
- [ ] String operations and manipulation

### 1.7 Pointers & References
- [ ] Pointer arithmetic
- [ ] Pointer to pointer
- [ ] nullptr
- [ ] References vs pointers
- [ ] Dangling pointers and references
- [ ] void pointers
- [ ] const pointers vs pointer to const

---

## PHASE 2: Memory & Object Model (Weeks 5–9)

### 2.1 Memory Layout
- [ ] Stack vs Heap
- [ ] Text segment, Data segment, BSS
- [ ] Stack frames
- [ ] Memory alignment and padding
- [ ] new/delete vs malloc/free
- [ ] Placement new
- [ ] Memory leaks and how to detect them

### 2.2 Classes & Objects
- [ ] Class definition, access specifiers (public, protected, private)
- [ ] Constructors: default, parameterized, delegating
- [ ] Destructors
- [ ] Member initializer lists (and why they matter)
- [ ] this pointer
- [ ] static members
- [ ] const member functions
- [ ] Mutable keyword
- [ ] Friend functions and classes
- [ ] Nested classes

### 2.3 RAII (Resource Acquisition Is Initialization)
- [ ] The core C++ idiom
- [ ] Deterministic destruction
- [ ] Scope-based resource management
- [ ] File handles, mutexes, memory via RAII

### 2.4 Copy & Move Semantics
- [ ] Copy constructor
- [ ] Copy assignment operator
- [ ] Rule of Three
- [ ] Move constructor
- [ ] Move assignment operator
- [ ] Rule of Five (and Rule of Zero)
- [ ] std::move
- [ ] Rvalue references (&&)
- [ ] Value categories: lvalue, rvalue, xvalue, prvalue, glvalue

### 2.5 Operator Overloading
- [ ] Arithmetic operators
- [ ] Comparison operators (and <=> spaceship in C++20)
- [ ] Stream operators (<< and >>)
- [ ] Subscript operator []
- [ ] Function call operator ()
- [ ] Conversion operators
- [ ] User-defined literals

### 2.6 Inheritance & Polymorphism
- [ ] Single inheritance
- [ ] Multiple inheritance
- [ ] Virtual functions and vtable mechanism
- [ ] Pure virtual functions and abstract classes
- [ ] Override and final keywords
- [ ] Virtual destructors (critical!)
- [ ] Diamond problem and virtual inheritance
- [ ] Object slicing
- [ ] dynamic_cast and RTTI

---

## PHASE 3: Modern C++ (Weeks 10–15)

### 3.1 Smart Pointers
- [ ] std::unique_ptr (exclusive ownership)
- [ ] std::shared_ptr (shared ownership, reference counting)
- [ ] std::weak_ptr (breaking cycles)
- [ ] Custom deleters
- [ ] std::make_unique, std::make_shared
- [ ] When to use which

### 3.2 Templates
- [ ] Function templates
- [ ] Class templates
- [ ] Template specialization (full and partial)
- [ ] Non-type template parameters
- [ ] Template argument deduction
- [ ] Variadic templates and parameter packs
- [ ] Fold expressions (C++17)
- [ ] CTAD (Class Template Argument Deduction, C++17)
- [ ] Concepts (C++20)
- [ ] requires expressions

### 3.3 STL Containers
- [ ] Sequence: vector, deque, list, forward_list, array
- [ ] Associative: set, map, multiset, multimap
- [ ] Unordered: unordered_set, unordered_map
- [ ] Adaptors: stack, queue, priority_queue
- [ ] std::span (C++20)
- [ ] Container internals (how vector grows, hash table mechanics)

### 3.4 Iterators & Algorithms
- [ ] Iterator categories: input, output, forward, bidirectional, random access
- [ ] Iterator invalidation rules
- [ ] STL algorithms: sort, find, transform, accumulate, remove_if, etc.
- [ ] Ranges library (C++20)
- [ ] Views and lazy evaluation

### 3.5 Lambda Expressions
- [ ] Capture by value, reference, this
- [ ] Mutable lambdas
- [ ] Generic lambdas (auto parameters)
- [ ] Lambda as callback
- [ ] std::function
- [ ] Immediately Invoked Lambda Expressions (IILE)

### 3.6 Error Handling
- [ ] Exceptions: throw, try, catch
- [ ] Exception safety guarantees (basic, strong, nothrow)
- [ ] noexcept specifier
- [ ] std::expected (C++23)
- [ ] Error codes vs exceptions (when to use which)
- [ ] Custom exception hierarchies
- [ ] Stack unwinding

### 3.7 Move Semantics Deep Dive
- [ ] Perfect forwarding (std::forward)
- [ ] Universal references / forwarding references
- [ ] Reference collapsing rules
- [ ] Return value optimization (RVO/NRVO)
- [ ] Copy elision (guaranteed in C++17)

---

## PHASE 4: Advanced C++ (Weeks 16–22)

### 4.1 Template Metaprogramming
- [ ] Type traits (std::is_integral, std::remove_const, etc.)
- [ ] SFINAE (Substitution Failure Is Not An Error)
- [ ] std::enable_if
- [ ] if constexpr (C++17)
- [ ] Compile-time computation
- [ ] Tag dispatch
- [ ] Policy-based design
- [ ] Expression templates

### 4.2 Advanced Patterns
- [ ] CRTP (Curiously Recurring Template Pattern)
- [ ] Type erasure (std::any, std::function internals)
- [ ] Pimpl idiom
- [ ] Singleton (and why it's often bad)
- [ ] Factory pattern
- [ ] Observer pattern
- [ ] Visitor pattern (and std::variant + std::visit)
- [ ] Strategy pattern via templates

### 4.3 Concurrency & Parallelism
- [ ] std::thread
- [ ] std::mutex, std::lock_guard, std::unique_lock
- [ ] std::condition_variable
- [ ] std::async, std::future, std::promise
- [ ] std::atomic and memory orders
- [ ] Memory model (happens-before, sequenced-before)
- [ ] Lock-free data structures
- [ ] Thread pools
- [ ] std::jthread and stop tokens (C++20)
- [ ] Coroutines (C++20)

### 4.4 Memory Management Advanced
- [ ] Custom allocators
- [ ] Pool allocators, arena allocators
- [ ] std::pmr (polymorphic memory resources)
- [ ] Memory-mapped files
- [ ] Cache-friendly programming
- [ ] False sharing
- [ ] Placement new in depth

### 4.5 Compile-Time Programming
- [ ] constexpr everything
- [ ] consteval (C++20)
- [ ] constinit (C++20)
- [ ] Compile-time string processing
- [ ] Static reflection (upcoming C++26)

### 4.6 Undefined Behavior & Safety
- [ ] Common UB scenarios
- [ ] Strict aliasing rule
- [ ] Sequence points / sequencing
- [ ] Integer overflow
- [ ] Lifetime issues
- [ ] Sanitizers: ASan, UBSan, TSan, MSan
- [ ] Static analysis tools

---

## PHASE 5: Systems Programming (Weeks 23–32)

### 5.1 Linux System Calls
- [ ] Process creation: fork(), exec(), wait()
- [ ] File descriptors and I/O: open(), read(), write(), close()
- [ ] Memory: mmap(), munmap(), brk()
- [ ] Signals: signal(), sigaction()
- [ ] Process groups and sessions

### 5.2 File Systems & I/O
- [ ] Low-level I/O vs standard I/O
- [ ] Buffered vs unbuffered I/O
- [ ] epoll, poll, select for I/O multiplexing
- [ ] Asynchronous I/O (io_uring)
- [ ] std::filesystem (C++17)
- [ ] File locking (flock, fcntl)

### 5.3 Networking
- [ ] Socket programming (TCP/UDP)
- [ ] Berkeley sockets API
- [ ] Non-blocking I/O
- [ ] Event-driven architecture
- [ ] Building a TCP server from scratch
- [ ] Protocol design and parsing

### 5.4 IPC (Inter-Process Communication)
- [ ] Pipes and FIFOs
- [ ] Shared memory (POSIX and System V)
- [ ] Message queues
- [ ] Unix domain sockets
- [ ] Memory-mapped IPC

### 5.5 Dynamic Libraries & Linking
- [ ] Static vs shared libraries
- [ ] Symbol visibility
- [ ] dlopen/dlsym (runtime loading)
- [ ] ABI compatibility
- [ ] Name mangling
- [ ] Linker scripts

### 5.6 Performance & Profiling
- [ ] CPU caches and cache lines
- [ ] Branch prediction
- [ ] SIMD intrinsics
- [ ] perf, gprof, Cachegrind
- [ ] Flame graphs
- [ ] Benchmarking with Google Benchmark

---

## PHASE 6: Mastery Projects (Weeks 33–52)

### Project 1: Unix Shell
- Process management, pipes, I/O redirection, job control

### Project 2: Custom Memory Allocator
- malloc/free implementation, multiple strategies (first-fit, best-fit, buddy system)

### Project 3: Thread Pool with Work Stealing
- Lock-free queue, task scheduling, futures

### Project 4: HTTP/1.1 Server
- Async I/O with epoll/io_uring, connection pooling, static file serving

### Project 5: Key-Value Database
- B-tree on disk, WAL (Write-Ahead Logging), crash recovery, simple query language

### Project 6: Container Runtime (mini-Docker)
- Linux namespaces, cgroups, overlay filesystem, networking

### Project 7: Compiler/Interpreter
- Lexer, parser, AST, code generation (target: a simple language → x86-64 or bytecode)

---

## PHASE 7: God Level (Ongoing)

### Deepening
- [ ] Read the C++ Standard (ISO/IEC 14882)
- [ ] Contribute to open-source C++ projects (LLVM, GCC, Chromium, etc.)
- [ ] Write a C++ library and publish it
- [ ] Study compiler internals (Clang/LLVM)
- [ ] OS development (write a minimal kernel)
- [ ] Embedded systems programming
- [ ] Real-time systems
- [ ] Formal verification and correctness proofs

### Keep Up
- [ ] Follow C++ proposals (wg21.link)
- [ ] CppCon talks (YouTube)
- [ ] C++ Weekly (Jason Turner)
- [ ] ISO C++ blog (isocpp.org)
