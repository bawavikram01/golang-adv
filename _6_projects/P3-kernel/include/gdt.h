#pragma once
#include <stdint.h>

//
// gdt.h — Global Descriptor Table
//
// The GDT defines memory segments. In modern x86 (flat model),
// we set up minimal segments that cover all 4GB:
//   - Null descriptor (required)
//   - Kernel code segment (ring 0, execute/read)
//   - Kernel data segment (ring 0, read/write)
//   - User code segment (ring 3, execute/read) — for later
//   - User data segment (ring 3, read/write) — for later
//

struct GdtEntry {
    uint16_t limit_low;
    uint16_t base_low;
    uint8_t  base_middle;
    uint8_t  access;
    uint8_t  granularity;  // Also contains limit bits 16:19
    uint8_t  base_high;
} __attribute__((packed));

struct GdtPointer {
    uint16_t limit;    // Size of GDT - 1
    uint32_t base;     // Address of first GDT entry
} __attribute__((packed));

void gdt_init();
