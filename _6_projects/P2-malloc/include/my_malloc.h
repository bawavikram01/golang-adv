#ifndef MY_MALLOC_H
#define MY_MALLOC_H

#include <cstddef>  // size_t

//
// Block Header — metadata stored BEFORE each allocation
//
// Memory layout:
//   [Header][---- user data ----]
//            ^
//            This is the pointer returned to the user
//
// The header is INVISIBLE to the user. When they call free(ptr),
// we go BACK by sizeof(Header) to find the metadata.
//
struct alignas(16) Header {
    size_t size;       // Size of the user data (NOT including header)
    bool is_free;      // Is this block available?
    bool is_mmap;      // Was this allocated with mmap? (Milestone 5)
    Header* next;      // Next block in the list (sequential in memory)
};

// Alignment requirement (16 bytes for x86-64)
constexpr size_t ALIGNMENT = 16;

// Threshold for using mmap instead of sbrk (Milestone 5)
constexpr size_t MMAP_THRESHOLD = 128 * 1024;  // 128 KB

// Minimum block size worth splitting (Milestone 2)
// Don't split if the remainder would be smaller than this
constexpr size_t MIN_SPLIT_SIZE = sizeof(Header) + ALIGNMENT;

// Milestone 3: Allocation strategies
enum class AllocStrategy {
    FIRST_FIT,   // Take the first free block that fits (fast)
    BEST_FIT,    // Take the smallest block that fits (less waste)
    WORST_FIT    // Take the largest block (reduces tiny fragments)
};

// Set the allocation strategy (default: FIRST_FIT)
void set_alloc_strategy(AllocStrategy strategy);

// ------- Public API (matches libc interface) -------

void* my_malloc(size_t size);
void  my_free(void* ptr);
void* my_realloc(void* ptr, size_t new_size);
void* my_calloc(size_t count, size_t size);

// ------- Debug helpers -------

// Print the state of the heap (all blocks, free and used)
void heap_dump();

// Return total free bytes / total heap bytes
double fragmentation_ratio();

// Return total number of blocks in the list
int block_count();

#endif // MY_MALLOC_H
