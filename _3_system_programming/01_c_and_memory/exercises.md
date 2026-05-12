# Module 1 Exercises — C and Memory

Complete these exercises to solidify your understanding.
**Rule: Write, compile, run, and debug. Don't just read.**

---

## Exercise 1: Memory Map Explorer

Write a program that prints the address of:
- A function (text segment)
- A global initialized variable
- A global uninitialized variable
- A local variable (stack)
- A malloc'd variable (heap)
- A string literal

Verify that the addresses follow the expected layout (text < data < bss < heap < ... < stack).

**Bonus:** Read `/proc/self/maps` from within your program to see the actual memory mappings.

---

## Exercise 2: Pointer Swap

Write a function `void swap(int *a, int *b)` that swaps two integers using pointers.
Then write `void swap_ptrs(int **a, int **b)` that swaps two pointers themselves.

Test both and explain the difference.

---

## Exercise 3: String Manipulation Without stdlib

Implement these functions WITHOUT using any string.h functions:
- `int my_strlen(const char *s)`
- `char *my_strcpy(char *dest, const char *src)`
- `int my_strcmp(const char *a, const char *b)`
- `char *my_strcat(char *dest, const char *src)`

This teaches you to think in terms of raw memory traversal.

---

## Exercise 4: Memory Leak Detector

Write a wrapper around malloc/free that tracks:
- Every allocation (address + size + file + line number)
- Every free
- At program exit, print any allocations that were never freed

Hint: Use macros to capture `__FILE__` and `__LINE__`:
```c
#define my_malloc(size) _my_malloc(size, __FILE__, __LINE__)
```

---

## Exercise 5: Buffer Overflow Attack (Educational)

1. Write a program with a buffer overflow vulnerability:
   ```c
   void vulnerable() {
       char buf[64];
       gets(buf);  // Never use gets() in real code!
   }
   ```
2. Compile WITHOUT protections: `gcc -fno-stack-protector -z execstack -no-pie`
3. Use GDB to examine the stack before and after overflow
4. Observe how the return address gets overwritten

**Purpose:** Understanding WHY systems code must validate buffer sizes.

---

## Exercise 6: Struct Packing Challenge

Given this struct, calculate the size manually (with padding), then verify:
```c
struct Mystery {
    char a;
    int b;
    char c;
    double d;
    short e;
};
```

Then reorder the members to minimize size. What's the minimum possible size?

---

## Exercise 7: Build a Dynamic Array

Implement a growable array (like Go's slices or Rust's Vec):
```c
typedef struct {
    int *data;
    size_t length;
    size_t capacity;
} DynArray;

DynArray *dynarray_new();
void dynarray_push(DynArray *arr, int value);
int dynarray_get(DynArray *arr, size_t index);
void dynarray_free(DynArray *arr);
```

Growth strategy: double capacity when full. This is how every ArrayList/Vector works internally.

---

## Exercise 8: Read /proc/self/maps

Write a C program that:
1. Opens `/proc/self/maps`
2. Reads and prints each memory region
3. Identifies which region contains: your main function, a stack variable, a heap allocation

This shows you the real virtual memory layout of your process.

---

## Validation

For each exercise:
1. Compile with: `gcc -Wall -Wextra -Werror -g -fsanitize=address`
2. The `-fsanitize=address` flag catches buffer overflows, use-after-free, leaks
3. Run under Valgrind: `valgrind --leak-check=full ./program`
4. Zero errors, zero leaks = you pass

---

## When You're Done

You should be able to answer:
- What's the difference between `int *p` and `int **p`?
- Why can't you return a pointer to a local variable?
- What happens when you `free()` the same pointer twice?
- How does `malloc` actually get memory from the OS?
- Why do structs have padding?

If yes → proceed to Module 2: How Programs Actually Run.
