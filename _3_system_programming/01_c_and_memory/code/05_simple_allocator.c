// simple_allocator.c — Build a tiny memory allocator to understand malloc
// This teaches you EXACTLY what malloc does under the hood.
#include <stdio.h>
#include <unistd.h>   // sbrk
#include <string.h>

// Each allocated block has a header
typedef struct block_header {
    size_t size;                  // Size of the data area
    int free;                     // Is this block free?
    struct block_header *next;    // Next block in the list
} block_header_t;

#define HEADER_SIZE sizeof(block_header_t)

static block_header_t *head = NULL;  // Start of our block list

// Find a free block that's big enough (first-fit strategy)
static block_header_t *find_free_block(size_t size) {
    block_header_t *current = head;
    while (current) {
        if (current->free && current->size >= size) {
            return current;
        }
        current = current->next;
    }
    return NULL;
}

// Request more memory from the OS
static block_header_t *request_space(size_t size) {
    block_header_t *block = sbrk(0);  // Current break (end of heap)
    void *request = sbrk(HEADER_SIZE + size);  // Extend heap

    if (request == (void *)-1) {
        return NULL;  // sbrk failed — out of memory
    }

    block->size = size;
    block->free = 0;
    block->next = NULL;

    // Add to linked list
    if (!head) {
        head = block;
    } else {
        block_header_t *current = head;
        while (current->next) {
            current = current->next;
        }
        current->next = block;
    }

    return block;
}

// Our malloc implementation
void *my_malloc(size_t size) {
    if (size == 0) return NULL;

    // Try to reuse a free block
    block_header_t *block = find_free_block(size);
    if (block) {
        block->free = 0;
        return (void *)(block + 1);  // Return pointer AFTER the header
    }

    // No free block — ask OS for more memory
    block = request_space(size);
    if (!block) return NULL;

    return (void *)(block + 1);
}

// Our free implementation
void my_free(void *ptr) {
    if (!ptr) return;

    // The header is right before the pointer
    block_header_t *block = (block_header_t *)ptr - 1;
    block->free = 1;
}

// Print allocator state
void dump_heap() {
    printf("\n--- Heap State ---\n");
    block_header_t *current = head;
    int i = 0;
    while (current) {
        printf("Block %d: addr=%p, size=%zu, free=%s\n",
               i, (void *)(current + 1), current->size,
               current->free ? "YES" : "NO");
        current = current->next;
        i++;
    }
    printf("------------------\n");
}

int main() {
    printf("=== Building a Simple Memory Allocator ===\n");
    printf("Header size: %zu bytes\n\n", HEADER_SIZE);

    // Allocate some memory
    int *a = my_malloc(sizeof(int));
    *a = 42;
    printf("Allocated int: %d at %p\n", *a, (void *)a);

    char *s = my_malloc(20);
    strcpy(s, "Hello, Systems!");
    printf("Allocated string: \"%s\" at %p\n", s, (void *)s);

    int *arr = my_malloc(5 * sizeof(int));
    for (int i = 0; i < 5; i++) arr[i] = i * 10;
    printf("Allocated array at %p\n", (void *)arr);

    dump_heap();

    // Free the string
    printf("\nFreeing string...\n");
    my_free(s);
    dump_heap();

    // Allocate again — should reuse the freed block
    char *s2 = my_malloc(10);
    strcpy(s2, "Reused!");
    printf("\nNew allocation reused free block: \"%s\" at %p\n", s2, (void *)s2);
    dump_heap();

    printf("\n--- What real malloc does better ---\n");
    printf("1. Splitting: split large free blocks into smaller ones\n");
    printf("2. Coalescing: merge adjacent free blocks\n");
    printf("3. Binning: separate free lists by size class\n");
    printf("4. mmap: use mmap() for large allocations instead of sbrk\n");
    printf("5. Thread safety: per-thread arenas (like glibc's ptmalloc)\n");

    return 0;
}
