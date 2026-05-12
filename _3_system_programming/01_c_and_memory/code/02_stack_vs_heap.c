// stack_vs_heap.c — Demonstrate allocation speed and behavior differences
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <string.h>

#define ITERATIONS 1000000

// Returns a pointer to a stack variable — THIS IS A BUG
// The compiler may warn about this
int *dangling_pointer_demo() {
    int local = 42;
    return &local;  // WARNING: local is destroyed after return
}

void stack_allocation_speed() {
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);

    for (int i = 0; i < ITERATIONS; i++) {
        int arr[100];  // Stack allocation — just moves stack pointer
        arr[0] = i;    // Prevent optimization
        (void)arr[99];
    }

    clock_gettime(CLOCK_MONOTONIC, &end);
    long ns = (end.tv_sec - start.tv_sec) * 1000000000L + (end.tv_nsec - start.tv_nsec);
    printf("Stack: %d allocations in %ld ns (%.1f ns/alloc)\n", ITERATIONS, ns, (double)ns / ITERATIONS);
}

void heap_allocation_speed() {
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);

    for (int i = 0; i < ITERATIONS; i++) {
        int *arr = malloc(100 * sizeof(int));  // Heap — allocator overhead
        arr[0] = i;
        free(arr);
    }

    clock_gettime(CLOCK_MONOTONIC, &end);
    long ns = (end.tv_sec - start.tv_sec) * 1000000000L + (end.tv_nsec - start.tv_nsec);
    printf("Heap:  %d allocations in %ld ns (%.1f ns/alloc)\n", ITERATIONS, ns, (double)ns / ITERATIONS);
}

int main() {
    printf("=== Stack vs Heap Allocation Speed ===\n\n");
    stack_allocation_speed();
    heap_allocation_speed();

    printf("\n=== Dangling Pointer Demo ===\n");
    int *bad = dangling_pointer_demo();
    printf("Dangling pointer value: %d (UNDEFINED — may be 42 or garbage)\n", *bad);

    printf("\n=== Stack Overflow Demo ===\n");
    printf("Default stack size on Linux: usually 8MB\n");
    printf("Try: ulimit -s  (shows stack size in KB)\n");
    printf("A recursive function without base case → SIGSEGV (stack overflow)\n");

    return 0;
}
