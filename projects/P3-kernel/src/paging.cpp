#include "paging.h"
#include "pmm.h"
#include "vga.h"
#include "idt.h"

//
// paging.cpp — x86 2-level paging
//
// We identity-map the first 4MB (kernel space).
// Then we can map additional pages as needed.
//
// Page Directory: 1024 entries, each points to a Page Table
// Page Table: 1024 entries, each points to a 4KB physical page
//
// Each entry format:
//   Bits 12-31: Physical address (page-aligned, so bottom 12 bits are flags)
//   Bit 0: Present
//   Bit 1: Read/Write
//   Bit 2: User/Supervisor
//

// Page directory — must be 4KB aligned
static uint32_t page_directory[1024] __attribute__((aligned(4096)));

// We pre-allocate one page table for the first 4MB (identity map)
static uint32_t first_page_table[1024] __attribute__((aligned(4096)));

// Page fault handler (used if we install custom ISR14 handling later)
static void page_fault_handler(Registers* regs) __attribute__((unused));
static void page_fault_handler(Registers* regs) {
    // CR2 contains the faulting address
    uint32_t fault_addr;
    asm volatile("mov %%cr2, %0" : "=r"(fault_addr));

    // Decode error code
    bool present = regs->err_code & 0x1;   // Page was present?
    bool write   = regs->err_code & 0x2;   // Was it a write?
    bool user    = regs->err_code & 0x4;   // From user mode?

    vga_set_color(VGA_LIGHT_RED, VGA_BLACK);
    vga_print("\n[PAGE FAULT] addr=");
    vga_print_hex(fault_addr);
    vga_print(" (");
    if (!present) vga_print("not-present ");
    if (write)    vga_print("write ");
    if (user)     vga_print("user ");
    vga_print(") at EIP=");
    vga_print_hex(regs->eip);
    vga_print("\n");
    vga_set_color(VGA_LIGHT_GREY, VGA_BLACK);

    // Halt
    asm volatile("cli; hlt");
}

void paging_init() {
    // Step 1: Clear page directory
    for (int i = 0; i < 1024; i++) {
        page_directory[i] = 0;  // Not present
    }

    // Step 2: Identity-map first 4MB using first_page_table
    // Virtual 0x00000000-0x003FFFFF → Physical 0x00000000-0x003FFFFF
    for (int i = 0; i < 1024; i++) {
        // Each entry maps a 4KB page: address | flags
        first_page_table[i] = (i * PAGE_SIZE) | PAGE_PRESENT | PAGE_WRITABLE;
    }

    // Step 3: Point page directory entry 0 to our page table
    page_directory[0] = ((uint32_t)first_page_table) | PAGE_PRESENT | PAGE_WRITABLE;

    // Step 4: Register page fault handler (ISR 14)
    // We override the default ISR handler for interrupt 14
    // (The generic isr_handler in idt.cpp already handles this,
    //  but we install a specific handler via a function pointer)
    // For now, the generic handler in idt.cpp prints the fault info.

    // Step 5: Load page directory into CR3
    asm volatile("mov %0, %%cr3" : : "r"(page_directory));

    // Step 6: Enable paging (set bit 31 of CR0)
    uint32_t cr0;
    asm volatile("mov %%cr0, %0" : "=r"(cr0));
    cr0 |= 0x80000000;
    asm volatile("mov %0, %%cr0" : : "r"(cr0));

    vga_print("[PAGING] Enabled — first 4MB identity-mapped\n");
}

void paging_map(uint32_t virtual_addr, uint32_t physical_addr, uint32_t flags) {
    uint32_t dir_index = virtual_addr >> 22;            // Top 10 bits
    uint32_t table_index = (virtual_addr >> 12) & 0x3FF; // Middle 10 bits

    // Check if page table exists for this directory entry
    if (!(page_directory[dir_index] & PAGE_PRESENT)) {
        // Allocate a new page table
        uint32_t new_table = pmm_alloc();
        if (!new_table) {
            vga_print("[PAGING] ERROR: out of memory for page table!\n");
            return;
        }

        // Zero it out
        uint32_t* table = (uint32_t*)new_table;
        for (int i = 0; i < 1024; i++) {
            table[i] = 0;
        }

        // Install in page directory
        page_directory[dir_index] = new_table | PAGE_PRESENT | PAGE_WRITABLE | flags;
    }

    // Get the page table
    uint32_t* table = (uint32_t*)(page_directory[dir_index] & 0xFFFFF000);

    // Map the page
    table[table_index] = (physical_addr & 0xFFFFF000) | (flags | PAGE_PRESENT);

    // Invalidate TLB entry for this address
    asm volatile("invlpg (%0)" : : "r"(virtual_addr) : "memory");
}

void paging_unmap(uint32_t virtual_addr) {
    uint32_t dir_index = virtual_addr >> 22;
    uint32_t table_index = (virtual_addr >> 12) & 0x3FF;

    if (!(page_directory[dir_index] & PAGE_PRESENT)) return;

    uint32_t* table = (uint32_t*)(page_directory[dir_index] & 0xFFFFF000);
    table[table_index] = 0;  // Not present

    asm volatile("invlpg (%0)" : : "r"(virtual_addr) : "memory");
}

uint32_t paging_get_physical(uint32_t virtual_addr) {
    uint32_t dir_index = virtual_addr >> 22;
    uint32_t table_index = (virtual_addr >> 12) & 0x3FF;

    if (!(page_directory[dir_index] & PAGE_PRESENT)) return 0;

    uint32_t* table = (uint32_t*)(page_directory[dir_index] & 0xFFFFF000);
    if (!(table[table_index] & PAGE_PRESENT)) return 0;

    return (table[table_index] & 0xFFFFF000) + (virtual_addr & 0xFFF);
}
