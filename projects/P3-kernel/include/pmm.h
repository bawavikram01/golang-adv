#pragma once
#include <stdint.h>

//
// pmm.h — Physical Memory Manager (Bitmap Allocator)
//
// Manages physical RAM in 4KB pages.
// Each bit in the bitmap represents one page:
//   0 = free, 1 = used
//

#define PAGE_SIZE 4096

void pmm_init(uint32_t mboot_info_addr);

// Allocate a single 4KB physical page. Returns physical address.
uint32_t pmm_alloc();

// Free a previously allocated page.
void pmm_free(uint32_t page_addr);

// Mark a range of addresses as used (for kernel, BIOS regions)
void pmm_mark_region_used(uint32_t base, uint32_t length);

// Mark a range of addresses as free
void pmm_mark_region_free(uint32_t base, uint32_t length);

// Stats
uint32_t pmm_get_total_pages();
uint32_t pmm_get_used_pages();
uint32_t pmm_get_free_pages();
