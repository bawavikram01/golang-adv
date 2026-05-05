#pragma once
#include <stdint.h>

//
// multiboot.h — Multiboot information structures
//
// GRUB passes a pointer to this struct in EBX.
// We use it to get the memory map (which regions are usable).
//

#define MULTIBOOT_FLAG_MEM     0x001  // mem_lower, mem_upper valid
#define MULTIBOOT_FLAG_MMAP    0x040  // mmap_* fields valid

struct MultibootInfo {
    uint32_t flags;
    uint32_t mem_lower;      // KB of memory below 1MB
    uint32_t mem_upper;      // KB of memory above 1MB
    uint32_t boot_device;
    uint32_t cmdline;
    uint32_t mods_count;     // Number of modules loaded
    uint32_t mods_addr;      // Address of module list
    uint32_t syms[4];
    uint32_t mmap_length;    // Size of memory map buffer
    uint32_t mmap_addr;      // Address of memory map
} __attribute__((packed));

// Memory map entry (variable-length, but we use the common 24-byte version)
struct MultibootMmapEntry {
    uint32_t size;           // Size of this entry (not counting this field)
    uint64_t addr;           // Start of memory region
    uint64_t len;            // Length of region
    uint32_t type;           // 1 = usable, 2+ = reserved
} __attribute__((packed));

// Module entry (for ramdisk)
struct MultibootModule {
    uint32_t mod_start;      // Physical address of module start
    uint32_t mod_end;        // Physical address of module end
    uint32_t cmdline;        // Module command line
    uint32_t reserved;
} __attribute__((packed));
