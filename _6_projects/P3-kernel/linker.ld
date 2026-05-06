/* linker.ld — Kernel linker script
 *
 * Tells the linker:
 *   - Start the kernel at 1MB (0x100000) — standard for x86 kernels
 *   - Place sections in the right order
 *   - Expose symbols for section boundaries (useful for memory management)
 */

ENTRY(_start)

SECTIONS
{
    /* Kernel starts at 1MB — below this is BIOS/GRUB territory */
    . = 1M;

    /* Multiboot header must come first (GRUB scans first 8KB) */
    .multiboot ALIGN(4K) :
    {
        *(.multiboot)
    }

    /* Code section */
    .text ALIGN(4K) :
    {
        *(.text)
    }

    /* Read-only data */
    .rodata ALIGN(4K) :
    {
        *(.rodata)
    }

    /* Initialized data */
    .data ALIGN(4K) :
    {
        *(.data)
    }

    /* Uninitialized data (BSS) */
    .bss ALIGN(4K) :
    {
        __bss_start = .;
        *(COMMON)
        *(.bss)
        __bss_end = .;
    }

    /* End of kernel — everything after this is free memory */
    __kernel_end = .;
}
