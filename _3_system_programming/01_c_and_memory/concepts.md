# Module 1: C and Memory — The Foundation of System Programming

## Why C?

Every OS kernel, database engine, and systems library is written in C (or now Rust).
As a backend dev, you've been shielded from what happens below your runtime. C removes that shield.

**Your mental model shift:** In backend dev, you think in objects/requests/responses.
In systems programming, you think in **bytes, addresses, and CPU cycles**.

---

## 1. Memory Layout of a Process

When your program runs, the OS gives it a virtual address space:

```
High Address
┌─────────────────────┐
│      Stack          │  ← Local variables, return addresses (grows DOWN)
│         ↓           │
├─────────────────────┤
│                     │
│   (unmapped gap)    │
│                     │
├─────────────────────┤
│         ↑           │
│       Heap          │  ← malloc/free (grows UP)
├─────────────────────┤
│       BSS           │  ← Uninitialized globals (zeroed)
├─────────────────────┤
│       Data          │  ← Initialized globals
├─────────────────────┤
│       Text          │  ← Your compiled machine code (read-only)
└─────────────────────┘
Low Address
```

**Key insight:** The stack and heap grow toward each other. Stack overflows happen when they collide.

---

## 2. Pointers — Addresses Are Just Numbers

A pointer is a variable that holds a **memory address**. That's it. No magic.

```c
int x = 42;        // x lives at some address, say 0x7ffd1234
int *p = &x;       // p holds the value 0x7ffd1234
*p = 100;          // go to address in p, write 100 there → x is now 100
```

### Pointer Arithmetic

```c
int arr[5] = {10, 20, 30, 40, 50};
int *p = arr;       // p points to arr[0]

p + 1;              // points to arr[1] (advances by sizeof(int) = 4 bytes)
*(p + 2);           // value at arr[2] = 30
```

**Critical:** `p + 1` doesn't add 1 byte — it adds `sizeof(*p)` bytes.

---

## 3. Stack vs Heap

| | Stack | Heap |
|---|-------|------|
| **Allocation** | Automatic (entering function) | Manual (`malloc`/`free`) |
| **Speed** | ~1 CPU cycle (move stack pointer) | ~100s of cycles (allocator logic) |
| **Size** | Small (8MB default on Linux) | Limited by RAM + swap |
| **Lifetime** | Until function returns | Until you `free()` it |
| **Fragmentation** | Never | Can fragment badly |

```c
void stack_example() {
    int local = 42;          // Stack — gone when function returns
    int arr[1000];           // Stack — 4000 bytes, instant allocation
}

void heap_example() {
    int *p = malloc(sizeof(int) * 1000);  // Heap — persists until free
    if (!p) { /* handle OOM */ }
    // ... use p ...
    free(p);                              // YOU must free it
    p = NULL;                             // Avoid dangling pointer
}
```

---

## 4. Common Bugs (and why they matter in systems code)

### Buffer Overflow
```c
char buf[10];
strcpy(buf, "This string is way too long!");  // Writes past buf → UNDEFINED BEHAVIOR
// In systems code: this is a SECURITY VULNERABILITY (stack smashing)
```

### Use After Free
```c
int *p = malloc(sizeof(int));
*p = 42;
free(p);
printf("%d\n", *p);  // UNDEFINED — memory might be reused
```

### Memory Leak
```c
void leak() {
    int *p = malloc(1000);
    // forgot free(p) — in a long-running system daemon, this kills you
}
```

---

## 5. The Compilation Pipeline

```
Source (.c) → Preprocessor → Compiler → Assembler → Linker → Executable
   foo.c    →    foo.i     →  foo.s   →   foo.o   →         → a.out
```

See each step:
```bash
gcc -E foo.c -o foo.i    # Preprocessor output
gcc -S foo.c -o foo.s    # Assembly
gcc -c foo.c -o foo.o    # Object file
gcc foo.o -o foo         # Link into executable
```

---

## 6. Your First Systems Program

```c
// pointer_basics.c — Understanding memory addresses
#include <stdio.h>
#include <stdlib.h>

int global_var = 42;                    // Data segment
int uninitialized_global;              // BSS segment

int main() {
    int stack_var = 10;                // Stack
    int *heap_var = malloc(sizeof(int)); // Heap
    *heap_var = 20;

    printf("Text  (code):  %p\n", (void*)main);
    printf("Data  (init):  %p\n", (void*)&global_var);
    printf("BSS   (zero):  %p\n", (void*)&uninitialized_global);
    printf("Heap:          %p\n", (void*)heap_var);
    printf("Stack:         %p\n", (void*)&stack_var);

    printf("\nNotice: Stack addr > Heap addr (stack grows down)\n");
    printf("global_var = %d (should be 42)\n", global_var);
    printf("uninitialized_global = %d (should be 0)\n", uninitialized_global);

    free(heap_var);
    return 0;
}
```

Compile and run:
```bash
gcc -Wall -Wextra -g pointer_basics.c -o pointer_basics
./pointer_basics
```

---

## Next Steps

After reading this, go to:
1. `code/` — Write and run the programs
2. `exercises.md` — Test your understanding
3. Then move to Module 2: How Programs Actually Run
