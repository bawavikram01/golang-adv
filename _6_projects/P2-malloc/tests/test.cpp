//
// test.cpp — Complete test suite for all milestones
//

#include "my_malloc.h"
#include <cstdio>
#include <cstring>
#include <cassert>
#include <cstdlib>
#include <chrono>

static int tests_passed = 0;
static int tests_failed = 0;

#define TEST(name) printf("  TEST: %-50s", name);
#define PASS() do { printf("✓ PASS\n"); tests_passed++; } while(0)
#define FAIL(msg) do { printf("✗ FAIL: %s\n", msg); tests_failed++; } while(0)

// ============================================================
// Milestone 1: Basic malloc/free
// ============================================================

void test_basic_malloc() {
    TEST("malloc returns non-null");
    void* ptr = my_malloc(100);
    if (ptr != nullptr) PASS(); else FAIL("returned null");
    my_free(ptr);
}

void test_write_and_read() {
    TEST("write and read back data");
    char* ptr = (char*)my_malloc(50);
    strcpy(ptr, "hello world");
    if (strcmp(ptr, "hello world") == 0) PASS(); else FAIL("data corrupted");
    my_free(ptr);
}

void test_multiple_allocs() {
    TEST("multiple allocations don't overlap");
    char* a = (char*)my_malloc(100);
    char* b = (char*)my_malloc(100);
    char* c = (char*)my_malloc(100);

    memset(a, 'A', 100);
    memset(b, 'B', 100);
    memset(c, 'C', 100);

    bool ok = true;
    for (int i = 0; i < 100; i++) {
        if (a[i] != 'A' || b[i] != 'B' || c[i] != 'C') {
            ok = false;
            break;
        }
    }
    if (ok) PASS(); else FAIL("blocks overlap!");

    my_free(a);
    my_free(b);
    my_free(c);
}

void test_free_and_reuse() {
    TEST("freed memory gets reused");
    void* first = my_malloc(100);
    my_free(first);
    void* second = my_malloc(100);
    if (second == first) PASS(); else FAIL("not reusing freed memory");
    my_free(second);
}

void test_zero_malloc() {
    TEST("malloc(0) returns null");
    void* ptr = my_malloc(0);
    if (ptr == nullptr) PASS(); else FAIL("should return null");
}

void test_free_null() {
    TEST("free(null) doesn't crash");
    my_free(nullptr);
    PASS();
}

// ============================================================
// Milestone 2: Splitting + Coalescing
// ============================================================

void test_coalescing() {
    TEST("adjacent free blocks coalesce");
    char* a = (char*)my_malloc(100);
    char* b = (char*)my_malloc(100);
    char* c = (char*)my_malloc(100);

    my_free(a);
    my_free(b);  // Should merge with a

    // Now allocate 200 — should fit in the merged block
    char* big = (char*)my_malloc(200);
    if (big != nullptr && big == a) PASS(); else FAIL("coalescing didn't merge or reuse");
    my_free(big);
    my_free(c);
}

void test_splitting() {
    TEST("large block splits for small allocation");
    char* big = (char*)my_malloc(1000);
    my_free(big);

    int blocks_before = block_count();
    char* small = (char*)my_malloc(50);
    int blocks_after = block_count();

    // Splitting should create an additional block
    if (blocks_after > blocks_before) PASS(); else FAIL("no split occurred");
    my_free(small);
}

void test_coalesce_reduces_fragmentation() {
    TEST("repeated alloc/free doesn't fragment");
    // Allocate and free many times — coalescing should keep block count low
    for (int i = 0; i < 100; i++) {
        void* p = my_malloc(64);
        my_free(p);
    }
    int blocks = block_count();
    // After all frees + coalescing, should be very few blocks
    if (blocks <= 5) PASS(); else FAIL("too many blocks — coalescing broken");
}

// ============================================================
// Milestone 3: Allocation Strategies
// ============================================================

void test_best_fit() {
    TEST("best-fit picks smallest fitting block");
    set_alloc_strategy(AllocStrategy::BEST_FIT);

    // Create three free blocks of different sizes with barriers
    char* a = (char*)my_malloc(100);   // will become free 100
    char* bar1 = (char*)my_malloc(16); // barrier
    char* b = (char*)my_malloc(300);   // will become free 300
    char* bar2 = (char*)my_malloc(16); // barrier
    char* c = (char*)my_malloc(200);   // will become free 200
    char* guard = (char*)my_malloc(16);

    my_free(a);  // free: 100 bytes
    my_free(b);  // free: 300 bytes
    my_free(c);  // free: 200 bytes

    // Best-fit for 90 bytes should pick the 100-byte block (smallest that fits)
    char* fit = (char*)my_malloc(90);
    if (fit == a) PASS(); else FAIL("didn't pick smallest fitting block");

    my_free(fit);
    my_free(bar1);
    my_free(bar2);
    my_free(guard);
    set_alloc_strategy(AllocStrategy::FIRST_FIT);  // reset
}

void test_worst_fit() {
    TEST("worst-fit picks largest block");
    set_alloc_strategy(AllocStrategy::WORST_FIT);

    char* a = (char*)my_malloc(100);
    char* bar1 = (char*)my_malloc(16);  // barrier prevents coalescing
    char* b = (char*)my_malloc(500);
    char* bar2 = (char*)my_malloc(16);  // barrier prevents coalescing
    char* c = (char*)my_malloc(200);
    char* guard = (char*)my_malloc(16);

    my_free(a);
    my_free(b);
    my_free(c);

    // Worst-fit should pick the 500-byte block
    char* fit = (char*)my_malloc(50);
    if (fit == b) PASS(); else FAIL("didn't pick largest block");

    my_free(fit);
    my_free(bar1);
    my_free(bar2);
    my_free(guard);
    set_alloc_strategy(AllocStrategy::FIRST_FIT);
}

// ============================================================
// Milestone 4: Alignment + Realloc
// ============================================================

void test_alignment() {
    TEST("all allocations are 16-byte aligned");
    bool ok = true;
    for (int i = 0; i < 50; i++) {
        void* ptr = my_malloc(i * 7 + 1);  // various odd sizes
        if (reinterpret_cast<uintptr_t>(ptr) % ALIGNMENT != 0) {
            ok = false;
        }
        my_free(ptr);
    }
    if (ok) PASS(); else FAIL("misaligned pointer returned");
}

void test_realloc_grow() {
    TEST("realloc grows and preserves data");
    char* ptr = (char*)my_malloc(50);
    strcpy(ptr, "hello");
    ptr = (char*)my_realloc(ptr, 200);
    if (strcmp(ptr, "hello") == 0) PASS(); else FAIL("data lost after realloc");
    my_free(ptr);
}

void test_realloc_shrink() {
    TEST("realloc shrink returns same pointer");
    char* ptr = (char*)my_malloc(200);
    char* shrunk = (char*)my_realloc(ptr, 50);
    if (shrunk == ptr) PASS(); else FAIL("unnecessary copy on shrink");
    my_free(shrunk);
}

void test_realloc_in_place() {
    TEST("realloc grows in-place when next block is free");
    char* a = (char*)my_malloc(100);
    char* b = (char*)my_malloc(100);
    char* c = (char*)my_malloc(100);
    strcpy(a, "keep this");

    my_free(b);  // b is now free, adjacent to a

    // Realloc a to 200 — should absorb b, no copy needed
    char* grown = (char*)my_realloc(a, 200);
    if (grown == a && strcmp(grown, "keep this") == 0) {
        PASS();
    } else {
        FAIL("didn't grow in-place");
    }
    my_free(grown);
    my_free(c);
}

void test_realloc_null() {
    TEST("realloc(NULL, n) acts like malloc");
    char* ptr = (char*)my_realloc(nullptr, 100);
    if (ptr != nullptr) PASS(); else FAIL("returned null");
    my_free(ptr);
}

void test_realloc_zero() {
    TEST("realloc(ptr, 0) acts like free");
    char* ptr = (char*)my_malloc(100);
    void* result = my_realloc(ptr, 0);
    if (result == nullptr) PASS(); else FAIL("should return null");
}

// ============================================================
// Milestone 5: mmap for large allocations
// ============================================================

void test_mmap_large_alloc() {
    TEST("large alloc (>128KB) succeeds");
    size_t large_size = 256 * 1024;  // 256 KB
    char* ptr = (char*)my_malloc(large_size);
    if (ptr != nullptr) PASS(); else FAIL("large alloc returned null");

    // Write to it to make sure it's real memory
    memset(ptr, 0xAB, large_size);
    my_free(ptr);  // should munmap
}

void test_mmap_doesnt_fragment_heap() {
    TEST("large allocs don't increase heap block count");
    int blocks_before = block_count();

    // Allocate and free several large blocks
    for (int i = 0; i < 5; i++) {
        char* p = (char*)my_malloc(256 * 1024);
        memset(p, 'X', 256 * 1024);
        my_free(p);
    }

    int blocks_after = block_count();
    if (blocks_after == blocks_before) PASS(); else FAIL("mmap blocks in heap list");
}

void test_mmap_data_integrity() {
    TEST("large alloc preserves written data");
    size_t size = 200 * 1024;
    char* ptr = (char*)my_malloc(size);
    // Write a pattern
    for (size_t i = 0; i < size; i++) {
        ptr[i] = (char)(i % 256);
    }
    // Verify
    bool ok = true;
    for (size_t i = 0; i < size; i++) {
        if (ptr[i] != (char)(i % 256)) { ok = false; break; }
    }
    if (ok) PASS(); else FAIL("data corrupted in mmap alloc");
    my_free(ptr);
}

// ============================================================
// Milestone 6: Stress Tests + Benchmarks
// ============================================================

void test_stress_random() {
    TEST("stress: 100K random alloc/free cycles");
    const int N = 100000;
    const int SLOTS = 200;
    void* ptrs[SLOTS] = {};
    size_t sizes[SLOTS] = {};
    bool ok = true;

    for (int i = 0; i < N && ok; i++) {
        int slot = i % SLOTS;
        if (ptrs[slot]) {
            // Verify data integrity before freeing
            char expected = (char)(slot & 0xFF);
            char* data = (char*)ptrs[slot];
            // Check first byte
            if (data[0] != expected) {
                ok = false;
                break;
            }
            my_free(ptrs[slot]);
            ptrs[slot] = nullptr;
        }
        size_t size = (rand() % 4096) + 1;  // 1 to 4096 bytes
        ptrs[slot] = my_malloc(size);
        sizes[slot] = size;
        if (!ptrs[slot]) { ok = false; break; }
        // Write pattern
        memset(ptrs[slot], (char)(slot & 0xFF), size);
    }

    // Cleanup
    for (int i = 0; i < SLOTS; i++) {
        if (ptrs[i]) my_free(ptrs[i]);
    }

    if (ok) PASS(); else FAIL("corruption or OOM during stress");
}

void test_stress_mixed_sizes() {
    TEST("stress: mixed small + large allocations");
    const int N = 1000;
    void* ptrs[N] = {};
    bool ok = true;

    for (int i = 0; i < N; i++) {
        // Mix of small (16-512B) and large (128KB+) allocations
        size_t size;
        if (i % 20 == 0) {
            size = 128 * 1024 + (rand() % 65536);  // large
        } else {
            size = (rand() % 512) + 16;  // small
        }
        ptrs[i] = my_malloc(size);
        if (!ptrs[i]) { ok = false; break; }
        memset(ptrs[i], 'Z', size > 1024 ? 1024 : size);  // write first 1KB
    }

    // Free in reverse (tests coalescing patterns)
    for (int i = N - 1; i >= 0; i--) {
        if (ptrs[i]) my_free(ptrs[i]);
    }

    if (ok) PASS(); else FAIL("failed on mixed sizes");
}

void benchmark_throughput() {
    printf("\n  BENCH: Throughput (ops/sec)...\n");

    const int OPS = 500000;
    void* ptrs[100] = {};

    auto start = std::chrono::high_resolution_clock::now();

    for (int i = 0; i < OPS; i++) {
        int slot = i % 100;
        if (ptrs[slot]) my_free(ptrs[slot]);
        ptrs[slot] = my_malloc((i % 256) + 16);
    }
    for (int i = 0; i < 100; i++) {
        if (ptrs[i]) my_free(ptrs[i]);
    }

    auto end = std::chrono::high_resolution_clock::now();
    auto ms = std::chrono::duration_cast<std::chrono::milliseconds>(end - start).count();

    double ops_per_sec = (ms > 0) ? (double)OPS / ((double)ms / 1000.0) : 0;
    printf("         %d ops in %ld ms = %.0f ops/sec\n", OPS, ms, ops_per_sec);
    printf("         (system malloc benchmark skipped — sbrk conflicts with glibc)\n");
}

// ============================================================
// MAIN
// ============================================================

int main() {
    srand(42);  // deterministic randomness for reproducibility

    printf("\n====== P2 Memory Allocator — Full Test Suite ======\n\n");

    printf("--- Milestone 1: Basic malloc/free ---\n");
    test_basic_malloc();
    test_write_and_read();
    test_multiple_allocs();
    test_free_and_reuse();
    test_zero_malloc();
    test_free_null();

    printf("\n--- Milestone 2: Splitting + Coalescing ---\n");
    test_coalescing();
    test_splitting();
    test_coalesce_reduces_fragmentation();

    printf("\n--- Milestone 3: Allocation Strategies ---\n");
    test_best_fit();
    test_worst_fit();

    printf("\n--- Milestone 4: Alignment + Realloc ---\n");
    test_alignment();
    test_realloc_grow();
    test_realloc_shrink();
    test_realloc_in_place();
    test_realloc_null();
    test_realloc_zero();

    printf("\n--- Milestone 5: mmap for Large Allocations ---\n");
    test_mmap_large_alloc();
    test_mmap_doesnt_fragment_heap();
    test_mmap_data_integrity();

    printf("\n--- Milestone 6: Stress Tests ---\n");
    test_stress_random();
    test_stress_mixed_sizes();

    printf("\n======================================\n");
    printf("  Results: %d passed, %d failed\n", tests_passed, tests_failed);
    printf("======================================\n");

    // Benchmark
    benchmark_throughput();

    // Final heap state
    heap_dump();

    return tests_failed > 0 ? 1 : 0;
}
