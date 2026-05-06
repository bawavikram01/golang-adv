# Lesson 1: How C++ Code Becomes an Executable

## The Compilation Pipeline

When you write C++ code and compile it, it goes through 4 stages:

```
Source Code (.cpp)
    │
    ▼ [1] Preprocessing
Expanded Source (macros expanded, headers inserted)
    │
    ▼ [2] Compilation
Assembly Code (.s)
    │
    ▼ [3] Assembly
Object Code (.o)
    │
    ▼ [4] Linking
Executable Binary
```

### Stage 1: Preprocessing (`g++ -E`)
- Processes all `#` directives: `#include`, `#define`, `#ifdef`, etc.
- `#include` literally copy-pastes the file contents
- Macros are expanded
- Conditional compilation is resolved
- Output: a single expanded text file (can be thousands of lines)

### Stage 2: Compilation (`g++ -S`)
- Parses the C++ code into an AST (Abstract Syntax Tree)
- Performs semantic analysis (type checking, overload resolution)
- Optimizations happen here (-O1, -O2, -O3)
- Generates assembly for the target architecture
- Output: `.s` file (assembly)

### Stage 3: Assembly (`g++ -c`)
- Converts assembly to machine code
- Creates an object file
- Contains machine instructions but NOT linked yet
- Symbols (function names) are recorded but addresses are unresolved
- Output: `.o` file (object code)

### Stage 4: Linking (`g++` or `ld`)
- Combines all `.o` files into one executable
- Resolves symbol references (connects function calls to definitions)
- Links standard library (libstdc++)
- Assigns final memory addresses
- Output: executable binary

## Key Concepts

### Translation Unit
A single `.cpp` file + all headers it includes = one translation unit.
Each translation unit is compiled independently.

### One Definition Rule (ODR)
- A variable/function can be DECLARED many times
- But DEFINED only ONCE across all translation units
- Exception: inline functions, templates, constexpr

### Header Guards
Prevent a header from being included multiple times in the same translation unit:

```cpp
// Old style
#ifndef MY_HEADER_H
#define MY_HEADER_H
// ... contents ...
#endif

// Modern (non-standard but universally supported)
#pragma once
```

## Important Compiler Flags

| Flag | Meaning |
|------|---------|
| -std=c++20 | Use C++20 standard |
| -Wall | Enable most warnings |
| -Wextra | Even more warnings |
| -Werror | Treat warnings as errors |
| -g | Include debug info (for gdb) |
| -O0 | No optimization (for debugging) |
| -O2 | Good optimization level |
| -O3 | Maximum optimization |
| -fsanitize=address | Enable AddressSanitizer |
| -c | Compile only (produce .o, don't link) |
| -S | Produce assembly only |
| -E | Preprocess only |
