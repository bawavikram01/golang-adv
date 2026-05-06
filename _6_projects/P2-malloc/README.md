# P2: Build Your Own Memory Allocator (C++)

## What Is a Memory Allocator?

Every time you write `new`, `malloc()`, or `make()`, something has to:
1. Find a free chunk of memory big enough
2. Mark it as "in use"
3. Return a pointer to it
4. Later, when you `free()` it, mark it available again

That "something" is the **memory allocator**. You're building it.

---

## Concepts You Need Before Starting

### 1. The Memory Layout of a Process

```
┌────────────────────────┐  High addresses (0xFFFF...)
│        STACK           │  ← grows DOWN (local variables, function calls)
│          ↓             │
│                        │
│          ↑             │
│        HEAP            │  ← grows UP (malloc/new allocations)
├────────────────────────┤  ← "brk" pointer (end of heap)
│        BSS             │  ← uninitialized globals (zeroed)
│        DATA            │  ← initialized globals
│        TEXT            │  ← your compiled code (read-only)
└────────────────────────┘  Low addresses (0x0000...)
```

**The heap** is the region where dynamic allocations live. When you call `malloc(100)`, the allocator finds 100 bytes somewhere in the heap. If the heap is too small, the allocator asks the OS for more by moving the `brk` pointer up.

### 2. How the OS Gives You Memory

Two system calls:

```cpp
// Method 1: sbrk() — move the "brk" pointer
void* ptr = sbrk(4096);  // grow heap by 4096 bytes

// Method 2: mmap() — request a specific chunk from the OS
void* ptr = mmap(NULL, 4096, PROT_READ|PROT_WRITE,
                 MAP_PRIVATE|MAP_ANONYMOUS, -1, 0);
```

**`sbrk`** = simple, grows the heap contiguously. Good for small allocations.
**`mmap`** = flexible, gives you memory anywhere. Good for large allocations (>128KB).

### 3. The Free List

A **linked list** of blocks, embedded IN the memory itself:

```
┌──────────┬──────────┬──────────┬──────────┐
│  Header  │  DATA    │  Header  │  DATA    │ ...
│  {size,  │ (user)   │  {size,  │ (user)   │
│   free,  │          │   free,  │          │
│   next}  │          │   next}  │          │
└──────────┴──────────┴──────────┴──────────┘
```

### 4. Fragmentation — The Enemy

```
[USED 32B][FREE 16B][USED 64B][FREE 8B][USED 16B][FREE 24B]
```
48 bytes free total, but can't satisfy `malloc(32)`. Fix: **coalescing** adjacent free blocks.

### 5. Alignment

Return pointers aligned to 16 bytes:
```cpp
size_t align(size_t size) { return (size + 15) & ~15; }
```

---

## Build Plan — 6 Milestones

### Milestone 1: Simplest Possible Allocator
- `sbrk()` to grow heap, linked list of blocks, first-fit search
- **Done when:** Can malloc and free 1000 times without crashing.

### Milestone 2: Splitting + Coalescing
- Split large blocks, merge adjacent free blocks on free
- **Done when:** Fragmentation stays bounded after 10K cycles.

### Milestone 3: Multiple Allocation Strategies
- First-fit, best-fit, worst-fit — implement and benchmark all three
- **Done when:** Benchmark shows trade-offs.

### Milestone 4: Alignment + realloc (in-place)
- 16-byte alignment, realloc tries to grow in-place
- **Done when:** realloc never corrupts data.

### Milestone 5: mmap for Large Allocations
- Size > 128KB → mmap, munmap on free
- **Done when:** Large allocs use mmap.

### Milestone 6: Stress Test + Benchmark vs glibc
- 100K+ random ops, compare against system malloc
- **Done when:** No corruption, you can explain why glibc is faster.

---

## Compile & Run
```bash
make
./test              # Run test suite
make bench          # Run benchmarks
```

---

## Progress

- [ ] Milestone 1: Basic malloc/free with free list
- [ ] Milestone 2: Splitting + coalescing
- [ ] Milestone 3: First-fit, best-fit, worst-fit
- [ ] Milestone 4: Alignment + realloc
- [ ] Milestone 5: mmap for large allocations
- [ ] Milestone 6: Stress test + benchmark vs glibc
