#pragma once
#include <stdint.h>

//
// paging.h — Virtual Memory Manager
//
// x86 2-level paging:
//   Page Directory (1024 entries) → Page Tables (1024 entries each) → 4KB Pages
//
// Virtual address layout:
//   [Dir index (10 bits)] [Table index (10 bits)] [Offset (12 bits)]
//

#define PAGE_PRESENT   0x01
#define PAGE_WRITABLE  0x02
#define PAGE_USER      0x04

void paging_init();

// Map a virtual address to a physical address
void paging_map(uint32_t virtual_addr, uint32_t physical_addr, uint32_t flags);

// Unmap a virtual address
void paging_unmap(uint32_t virtual_addr);

// Get the physical address for a virtual address (returns 0 if not mapped)
uint32_t paging_get_physical(uint32_t virtual_addr);
