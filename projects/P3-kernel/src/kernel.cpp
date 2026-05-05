#include "vga.h"
#include "gdt.h"
#include "idt.h"
#include "keyboard.h"
#include "timer.h"
#include "pmm.h"
#include "paging.h"
#include "scheduler.h"
#include "ramdisk.h"

//
// kernel.cpp — The kernel entry point
//
// GRUB → boot.asm → here. We initialize all subsystems, then
// create demo processes and let the scheduler take over.
//

// Demo processes for Milestone 6
static void process_a() {
    int counter = 0;
    for (;;) {
        if (counter % 500000 == 0) {
            vga_set_color(VGA_LIGHT_CYAN, VGA_BLACK);
            vga_print("[A]");
            vga_set_color(VGA_WHITE, VGA_BLACK);
        }
        counter++;
    }
}

static void process_b() {
    int counter = 0;
    for (;;) {
        if (counter % 500000 == 0) {
            vga_set_color(VGA_LIGHT_GREEN, VGA_BLACK);
            vga_print("[B]");
            vga_set_color(VGA_WHITE, VGA_BLACK);
        }
        counter++;
    }
}

static void process_c() {
    int counter = 0;
    for (;;) {
        if (counter % 500000 == 0) {
            vga_set_color(VGA_LIGHT_MAGENTA, VGA_BLACK);
            vga_print("[C]");
            vga_set_color(VGA_WHITE, VGA_BLACK);
        }
        counter++;
    }
}

// Timer tick counter — triggers scheduling every N ticks
static uint32_t schedule_counter = 0;

static void scheduler_timer_callback(Registers* /*regs*/) {
    schedule_counter++;
    // Switch processes every 2 ticks (20ms quantum)
    if (schedule_counter >= 2) {
        schedule_counter = 0;
        schedule();
    }
}

extern "C" void kernel_main(uint32_t magic, uint32_t* mboot_info) {
    // Milestone 1: VGA output
    vga_init();
    vga_set_color(VGA_LIGHT_GREEN, VGA_BLACK);
    vga_print("=================================\n");
    vga_print("  MiniOS Kernel v0.1\n");
    vga_print("=================================\n");
    vga_set_color(VGA_LIGHT_GREY, VGA_BLACK);

    // Verify multiboot magic
    if (magic != 0x2BADB002) {
        vga_set_color(VGA_LIGHT_RED, VGA_BLACK);
        vga_print("[FAIL] Not booted by multiboot-compliant bootloader!\n");
        return;
    }
    vga_print("[OK] Multiboot magic verified\n");

    // Milestone 2: GDT
    gdt_init();
    vga_print("[OK] GDT initialized\n");

    // Milestone 2: IDT + interrupts
    idt_init();
    vga_print("[OK] IDT initialized (PIC remapped)\n");

    // Milestone 3: Timer (100 Hz = every 10ms)
    timer_init(100);
    vga_print("[OK] PIT timer @ 100 Hz\n");

    // Milestone 3: Keyboard
    keyboard_init();
    vga_print("[OK] PS/2 keyboard driver loaded\n");

    // Milestone 4: Physical memory manager
    pmm_init((uint32_t)mboot_info);

    // Milestone 5: Paging
    paging_init();

    // Test paging: map a page at virtual 0x400000
    uint32_t test_page = pmm_alloc();
    if (test_page) {
        paging_map(0x00400000, test_page, PAGE_WRITABLE);
        // Write through the virtual address
        volatile uint32_t* ptr = (volatile uint32_t*)0x00400000;
        *ptr = 0xCAFEBABE;
        if (*ptr == 0xCAFEBABE) {
            vga_print("[PAGING] Test page mapped and verified at 0x400000\n");
        }
        paging_unmap(0x00400000);
        pmm_free(test_page);
    }

    // Milestone 7: Ramdisk (if a module was loaded)
    ramdisk_init((uint32_t)mboot_info);
    ramdisk_list();

    // Milestone 6: Scheduler
    scheduler_init();
    process_create(process_a, "Process A");
    process_create(process_b, "Process B");
    process_create(process_c, "Process C");

    // Enable interrupts
    asm volatile("sti");
    vga_print("[OK] Interrupts enabled\n\n");

    vga_set_color(VGA_YELLOW, VGA_BLACK);
    vga_print("Scheduling 3 processes (round-robin, 20ms quantum):\n");
    vga_set_color(VGA_WHITE, VGA_BLACK);

    // Override timer handler to include scheduling
    register_irq_handler(0, scheduler_timer_callback);

    // Start the first process (this never returns)
    schedule();

    // Should never reach here
    for (;;) {
        asm volatile("hlt");
    }
}
