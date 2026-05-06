#include "gdt.h"

//
// gdt.cpp — Set up a flat-model GDT
//
// We need at least 3 entries:
//   0: Null descriptor (CPU requires this)
//   1: Kernel code segment (0x08) — ring 0, the whole 4GB
//   2: Kernel data segment (0x10) — ring 0, the whole 4GB
//

static GdtEntry gdt[5];
static GdtPointer gdt_ptr;

// Defined in assembly — loads the GDT register and reloads segments
extern "C" void gdt_flush(uint32_t gdt_ptr_addr);

static void gdt_set_entry(int index, uint32_t base, uint32_t limit,
                          uint8_t access, uint8_t granularity) {
    gdt[index].base_low    = base & 0xFFFF;
    gdt[index].base_middle = (base >> 16) & 0xFF;
    gdt[index].base_high   = (base >> 24) & 0xFF;

    gdt[index].limit_low   = limit & 0xFFFF;
    gdt[index].granularity = ((limit >> 16) & 0x0F) | (granularity & 0xF0);

    gdt[index].access = access;
}

void gdt_init() {
    gdt_ptr.limit = sizeof(gdt) - 1;
    gdt_ptr.base = (uint32_t)&gdt;

    // Null descriptor
    gdt_set_entry(0, 0, 0, 0, 0);

    // Kernel code segment: base=0, limit=4GB, ring 0, executable
    // Access: present(1) | ring0(00) | type(1) | exec(1) | conforming(0) | readable(1) | accessed(0)
    // = 0b10011010 = 0x9A
    // Granularity: 4KB pages(1) | 32-bit(1) | 0 | 0 | limit[19:16]=0xF
    // = 0b11001111 = 0xCF
    gdt_set_entry(1, 0, 0xFFFFFFFF, 0x9A, 0xCF);

    // Kernel data segment: same but read/write, not executable
    // Access: 0b10010010 = 0x92
    gdt_set_entry(2, 0, 0xFFFFFFFF, 0x92, 0xCF);

    // User code segment (ring 3) — for later
    // Access: 0b11111010 = 0xFA
    gdt_set_entry(3, 0, 0xFFFFFFFF, 0xFA, 0xCF);

    // User data segment (ring 3) — for later
    // Access: 0b11110010 = 0xF2
    gdt_set_entry(4, 0, 0xFFFFFFFF, 0xF2, 0xCF);

    gdt_flush((uint32_t)&gdt_ptr);
}
