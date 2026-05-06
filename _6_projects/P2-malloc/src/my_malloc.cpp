//
// my_malloc.cpp — A complete memory allocator
//
// ALL MILESTONES IMPLEMENTED:
//   1. Basic malloc/free with free list (sbrk)
//   2. Splitting + Coalescing
//   3. First-fit, Best-fit, Worst-fit strategies
//   4. Alignment + in-place realloc
//   5. mmap for large allocations
//   6. Ready for stress test + benchmark
//

#include "my_malloc.h"

#include <unistd.h>   // sbrk
#include <sys/mman.h> // mmap, munmap
#include <cstring>    // memcpy, memset
#include <cstdio>     // printf (for heap_dump)

// ============================================================
// GLOBALS
// ============================================================

static Header* head = nullptr;
static Header* tail = nullptr;
static AllocStrategy current_strategy = AllocStrategy::FIRST_FIT;

// ============================================================
// STRATEGY SETTER (Milestone 3)
// ============================================================

void set_alloc_strategy(AllocStrategy strategy) {
    current_strategy = strategy;
}

// ============================================================
// HELPERS
// ============================================================

static size_t align_up(size_t size) {
    return (size + (ALIGNMENT - 1)) & ~(ALIGNMENT - 1);
}

static void* header_to_data(Header* h) {
    return reinterpret_cast<void*>(reinterpret_cast<char*>(h) + sizeof(Header));
}

static Header* data_to_header(void* ptr) {
    return reinterpret_cast<Header*>(reinterpret_cast<char*>(ptr) - sizeof(Header));
}

// ============================================================
// MILESTONE 2: SPLITTING
// ============================================================
//
// When a free block is much bigger than needed, split it:
//   Before: [Header][ -------- 1000 bytes -------- ]
//   After:  [Header][100 bytes][Header][ 868 bytes ]
//                    ^used              ^new free block
//
// Only split if the leftover is big enough to be useful
// (at least MIN_SPLIT_SIZE = sizeof(Header) + ALIGNMENT)
//
static void split_block(Header* block, size_t needed) {
    size_t remaining = block->size - needed - sizeof(Header);

    // Only split if the remainder is worth keeping
    if (block->size >= needed + MIN_SPLIT_SIZE) {
        // Create a new free block after the allocated portion
        Header* new_block = reinterpret_cast<Header*>(
            reinterpret_cast<char*>(block) + sizeof(Header) + needed
        );
        new_block->size = remaining;
        new_block->is_free = true;
        new_block->is_mmap = false;
        new_block->next = block->next;

        // Update the original block
        block->size = needed;
        block->next = new_block;

        // Update tail if needed
        if (block == tail) {
            tail = new_block;
        }
    }
}

// ============================================================
// MILESTONE 2: COALESCING
// ============================================================
//
// After freeing a block, merge it with adjacent free blocks.
// This prevents fragmentation from accumulating.
//
//   Before: [FREE 100][FREE 200][USED 50]
//   After:  [FREE 300 + sizeof(Header)][USED 50]
//
static void coalesce(Header* block) {
    // Merge with NEXT block(s) if they're free
    while (block->next && block->next->is_free) {
        Header* absorbed = block->next;
        block->size += sizeof(Header) + absorbed->size;
        block->next = absorbed->next;

        // Update tail if we absorbed it
        if (absorbed == tail) {
            tail = block;
        }
    }
}

// ============================================================
// MILESTONE 3: ALLOCATION STRATEGIES
// ============================================================

// First-fit: return the first block that's big enough
static Header* find_first_fit(size_t size) {
    Header* current = head;
    while (current) {
        if (current->is_free && current->size >= size) {
            return current;
        }
        current = current->next;
    }
    return nullptr;
}

// Best-fit: return the SMALLEST block that's big enough
// Less wasted space per allocation, but slower (must scan all)
static Header* find_best_fit(size_t size) {
    Header* best = nullptr;
    Header* current = head;
    while (current) {
        if (current->is_free && current->size >= size) {
            if (!best || current->size < best->size) {
                best = current;
                // Perfect fit — no need to keep searching
                if (best->size == size) return best;
            }
        }
        current = current->next;
    }
    return best;
}

// Worst-fit: return the LARGEST free block
// Counterintuitive, but the leftover after splitting is bigger,
// which may reduce tiny unusable fragments
static Header* find_worst_fit(size_t size) {
    Header* worst = nullptr;
    Header* current = head;
    while (current) {
        if (current->is_free && current->size >= size) {
            if (!worst || current->size > worst->size) {
                worst = current;
            }
        }
        current = current->next;
    }
    return worst;
}

// Dispatch to the configured strategy
static Header* find_free_block(size_t size) {
    switch (current_strategy) {
        case AllocStrategy::FIRST_FIT:  return find_first_fit(size);
        case AllocStrategy::BEST_FIT:   return find_best_fit(size);
        case AllocStrategy::WORST_FIT:  return find_worst_fit(size);
    }
    return find_first_fit(size);  // fallback
}

// ============================================================
// GROW THE HEAP (sbrk)
// ============================================================

static Header* request_memory(size_t size) {
    size_t total = sizeof(Header) + size;

    void* block = sbrk(total);
    if (block == (void*)-1) {
        return nullptr;
    }

    Header* header = reinterpret_cast<Header*>(block);
    header->size = size;
    header->is_free = false;
    header->is_mmap = false;
    header->next = nullptr;

    if (tail) {
        tail->next = header;
    }
    tail = header;

    if (!head) {
        head = header;
    }

    return header;
}

// ============================================================
// MILESTONE 5: mmap FOR LARGE ALLOCATIONS
// ============================================================
//
// For allocations > MMAP_THRESHOLD (128KB), bypass the heap
// entirely and ask the OS for a dedicated memory region.
//
// Benefits:
//   - Memory is returned to the OS immediately on free (munmap)
//   - No heap fragmentation for large blocks
//   - Page-aligned automatically
//
// This is exactly what glibc malloc does for large allocations.
//
static void* mmap_alloc(size_t size) {
    size_t total = sizeof(Header) + size;

    void* region = mmap(nullptr, total,
                        PROT_READ | PROT_WRITE,
                        MAP_PRIVATE | MAP_ANONYMOUS,
                        -1, 0);
    if (region == MAP_FAILED) {
        return nullptr;
    }

    Header* header = reinterpret_cast<Header*>(region);
    header->size = size;
    header->is_free = false;
    header->is_mmap = true;
    header->next = nullptr;
    // mmap blocks are NOT added to the linked list — they're standalone

    return header_to_data(header);
}

static void mmap_free(Header* header) {
    size_t total = sizeof(Header) + header->size;
    munmap(header, total);  // return memory to OS immediately
}

// ============================================================
// MALLOC (All milestones combined)
// ============================================================

void* my_malloc(size_t size) {
    if (size == 0) return nullptr;

    // Milestone 4: Align the requested size to ALIGNMENT boundary
    size = align_up(size);

    // Milestone 5: Large allocations use mmap
    if (size >= MMAP_THRESHOLD) {
        return mmap_alloc(size);
    }

    // Milestone 3: Find a free block using the configured strategy
    Header* block = find_free_block(size);

    if (block) {
        block->is_free = false;

        // Milestone 2: Split if the block is much bigger than needed
        split_block(block, size);

        return header_to_data(block);
    }

    // No free block found — grow the heap
    block = request_memory(size);
    if (!block) return nullptr;

    return header_to_data(block);
}

// ============================================================
// FREE (All milestones combined)
// ============================================================

void my_free(void* ptr) {
    if (!ptr) return;

    Header* header = data_to_header(ptr);

    // Milestone 5: mmap'd blocks get munmap'd immediately
    if (header->is_mmap) {
        mmap_free(header);
        return;
    }

    // Mark as free
    header->is_free = true;

    // Milestone 2: Coalesce with adjacent free blocks
    coalesce(header);

    // Also try to coalesce BACKWARDS by scanning from head
    // (handles the case where the PREVIOUS block is also free)
    Header* current = head;
    while (current) {
        if (current->is_free) {
            coalesce(current);
        }
        current = current->next;
    }
}

// ============================================================
// REALLOC (Milestone 4: in-place growth)
// ============================================================

void* my_realloc(void* ptr, size_t new_size) {
    if (!ptr) return my_malloc(new_size);
    if (new_size == 0) {
        my_free(ptr);
        return nullptr;
    }

    Header* header = data_to_header(ptr);
    new_size = align_up(new_size);

    // Case 1: Current block is already big enough — maybe split excess
    if (header->size >= new_size) {
        // Milestone 2: Split off excess if it's worth it
        split_block(header, new_size);
        return ptr;
    }

    // Milestone 4: Case 2: Absorb the next block if it's free
    // This avoids an expensive copy when memory grows in-place
    if (header->next && header->next->is_free) {
        size_t combined = header->size + sizeof(Header) + header->next->size;
        if (combined >= new_size) {
            // Absorb next block
            Header* absorbed = header->next;
            header->size += sizeof(Header) + absorbed->size;
            header->next = absorbed->next;
            if (absorbed == tail) {
                tail = header;
            }
            // Split off excess
            split_block(header, new_size);
            return ptr;  // Same pointer, no copy needed!
        }
    }

    // Case 3: Must move — allocate new, copy old data, free old block
    void* new_ptr = my_malloc(new_size);
    if (!new_ptr) return nullptr;
    memcpy(new_ptr, ptr, header->size);
    my_free(ptr);
    return new_ptr;
}

// ============================================================
// CALLOC
// ============================================================

void* my_calloc(size_t count, size_t size) {
    // Check for overflow
    if (count != 0 && size > (size_t)-1 / count) {
        return nullptr;
    }
    size_t total = count * size;
    void* ptr = my_malloc(total);
    if (ptr) {
        memset(ptr, 0, total);
    }
    return ptr;
}

// ============================================================
// DEBUG: HEAP DUMP
// ============================================================

void heap_dump() {
    Header* current = head;
    int i = 0;
    size_t total_size = 0;
    size_t free_size = 0;

    printf("\n=== HEAP DUMP ===\n");
    while (current) {
        printf("  Block %d: addr=%p  size=%-6zu  %s\n",
               i, header_to_data(current), current->size,
               current->is_free ? "FREE" : "USED");
        total_size += current->size + sizeof(Header);
        if (current->is_free) free_size += current->size;
        current = current->next;
        i++;
    }
    printf("  Total: %zu bytes | Free: %zu bytes | Blocks: %d\n", total_size, free_size, i);
    printf("  Fragmentation: %.1f%%\n", fragmentation_ratio() * 100.0);
    printf("=================\n\n");
}

double fragmentation_ratio() {
    size_t total = 0;
    size_t free_bytes = 0;
    Header* current = head;
    while (current) {
        total += current->size;
        if (current->is_free) free_bytes += current->size;
        current = current->next;
    }
    if (total == 0) return 0.0;
    return static_cast<double>(free_bytes) / total;
}

int block_count() {
    int count = 0;
    Header* current = head;
    while (current) {
        count++;
        current = current->next;
    }
    return count;
}
