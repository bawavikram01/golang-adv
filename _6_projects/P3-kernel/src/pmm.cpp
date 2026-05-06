#include "pmm.h"
#include "multiboot.h"
#include "vga.h"

//
// pmm.cpp — Physical Memory Manager
//
// Uses a bitmap to track which 4KB pages are free/used.
// One bit per page: 0 = free, 1 = used.
//
// For 128MB RAM: 128MB / 4KB = 32768 pages = 4096 bytes of bitmap
//

// Bitmap — supports up to 4GB (1M pages = 128KB bitmap)
// We'll only use what we need based on actual RAM
#define MAX_PAGES (1024 * 1024)  // 4GB / 4KB = 1M pages max
static uint32_t bitmap[MAX_PAGES / 32];  // 128KB

static uint32_t total_pages = 0;
static uint32_t used_pages = 0;

// Kernel end address (defined in linker.ld)
extern "C" uint32_t __kernel_end;

// ---- Bitmap helpers ----

static inline void bitmap_set(uint32_t page) {
    bitmap[page / 32] |= (1 << (page % 32));
}

static inline void bitmap_clear(uint32_t page) {
    bitmap[page / 32] &= ~(1 << (page % 32));
}

static inline bool bitmap_test(uint32_t page) {
    return bitmap[page / 32] & (1 << (page % 32));
}

// ---- Public API ----

void pmm_init(uint32_t mboot_info_addr) {
    MultibootInfo* mboot = (MultibootInfo*)mboot_info_addr;

    // Step 1: Mark ALL pages as used (conservative default)
    for (uint32_t i = 0; i < MAX_PAGES / 32; i++) {
        bitmap[i] = 0xFFFFFFFF;
    }

    // Calculate total memory from mboot
    if (mboot->flags & MULTIBOOT_FLAG_MEM) {
        uint32_t total_kb = mboot->mem_lower + mboot->mem_upper + 1024;
        total_pages = total_kb / 4;  // 4KB per page
        if (total_pages > MAX_PAGES) total_pages = MAX_PAGES;
    } else {
        // Fallback: assume 128MB
        total_pages = 32768;
    }
    used_pages = total_pages;

    // Step 2: Parse memory map — mark usable regions as free
    if (mboot->flags & MULTIBOOT_FLAG_MMAP) {
        uint32_t offset = 0;
        while (offset < mboot->mmap_length) {
            MultibootMmapEntry* entry = (MultibootMmapEntry*)(mboot->mmap_addr + offset);

            if (entry->type == 1) {  // Usable memory
                pmm_mark_region_free((uint32_t)entry->addr, (uint32_t)entry->len);
            }

            offset += entry->size + sizeof(entry->size);
        }
    }

    // Step 3: Re-mark critical regions as USED (even if mmap said "usable"):

    // First 1MB — BIOS, IVT, VGA, etc.
    pmm_mark_region_used(0x00000000, 0x00100000);

    // Our kernel image (1MB to __kernel_end)
    uint32_t kernel_end = (uint32_t)&__kernel_end;
    pmm_mark_region_used(0x00100000, kernel_end - 0x00100000);

    // The bitmap itself lives in BSS (inside kernel), so already covered above.

    vga_print("[PMM] Initialized: ");
    vga_print_dec(pmm_get_free_pages());
    vga_print(" free pages (");
    vga_print_dec(pmm_get_free_pages() * 4);
    vga_print(" KB free)\n");
}

void pmm_mark_region_used(uint32_t base, uint32_t length) {
    uint32_t start_page = base / PAGE_SIZE;
    uint32_t num_pages = length / PAGE_SIZE;
    for (uint32_t i = 0; i < num_pages; i++) {
        if (!bitmap_test(start_page + i)) {
            bitmap_set(start_page + i);
            used_pages++;
        }
    }
}

void pmm_mark_region_free(uint32_t base, uint32_t length) {
    uint32_t start_page = base / PAGE_SIZE;
    uint32_t num_pages = length / PAGE_SIZE;
    for (uint32_t i = 0; i < num_pages; i++) {
        if (bitmap_test(start_page + i)) {
            bitmap_clear(start_page + i);
            used_pages--;
        }
    }
}

uint32_t pmm_alloc() {
    // First-fit: find the first free page
    for (uint32_t i = 0; i < total_pages / 32; i++) {
        if (bitmap[i] == 0xFFFFFFFF) continue;  // All 32 pages used, skip

        // Find which bit is free
        for (uint32_t bit = 0; bit < 32; bit++) {
            if (!(bitmap[i] & (1 << bit))) {
                uint32_t page = i * 32 + bit;
                bitmap_set(page);
                used_pages++;
                return page * PAGE_SIZE;  // Return physical address
            }
        }
    }
    // Out of memory!
    return 0;
}

void pmm_free(uint32_t page_addr) {
    uint32_t page = page_addr / PAGE_SIZE;
    if (bitmap_test(page)) {
        bitmap_clear(page);
        used_pages--;
    }
}

uint32_t pmm_get_total_pages() { return total_pages; }
uint32_t pmm_get_used_pages() { return used_pages; }
uint32_t pmm_get_free_pages() { return total_pages - used_pages; }
