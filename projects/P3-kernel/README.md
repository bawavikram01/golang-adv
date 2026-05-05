# P3 — Mini OS Kernel (C++ + Assembly)

A minimal operating system kernel that boots in QEMU, handles interrupts, manages memory with paging, runs multiple processes with preemptive scheduling, and reads files from a ramdisk.

---

## What You'll Learn

| Concept | Why It Matters |
|---------|---------------|
| Boot process (GRUB → kernel) | How every OS starts — BIOS/UEFI → bootloader → your code |
| Protected mode + GDT | CPU privilege rings, memory segmentation |
| IDT + interrupts | How hardware talks to software (keyboard, timer) |
| Paging + virtual memory | Every process has its own address space — this is how |
| Physical page allocator | Managing raw RAM pages (bitmap allocator) |
| Preemptive scheduling | Timer interrupt forces context switches — no process hogs CPU |
| Ramdisk filesystem | Reading files from memory — simplest "disk" |

---

## Architecture

```
┌─────────────────────────────────────────┐
│              User Processes              │
├─────────────────────────────────────────┤
│          Scheduler (Round-Robin)         │
├─────────────────────────────────────────┤
│  Virtual Memory  │  Physical Allocator  │
├─────────────────────────────────────────┤
│  IDT / ISRs / IRQs (PIC remapped)      │
├─────────────────────────────────────────┤
│  GDT  │  VGA Console  │  Keyboard/Timer │
├─────────────────────────────────────────┤
│  GRUB Multiboot → boot.asm → kernel    │
└─────────────────────────────────────────┘
```

---

## Milestones

### M1: Boot + VGA Output
- GRUB multiboot header in assembly
- Jump to C++ `kernel_main()`
- Write text to VGA buffer (0xB8000)
- **Done when:** "Hello from kernel!" appears on screen in QEMU

### M2: GDT + IDT + Interrupts
- Set up Global Descriptor Table (flat model: code + data segments)
- Set up Interrupt Descriptor Table (256 entries)
- Write ISR stubs in assembly (pushes error code, calls C++ handler)
- Remap PIC (8259) to IRQ 32–47
- Handle division by zero, page fault, general protection fault
- **Done when:** Keyboard input triggers IRQ1 handler, prints scan codes

### M3: Keyboard + Timer Drivers
- PS/2 keyboard driver: scancode → ASCII translation
- PIT (Programmable Interval Timer): tick counter, configurable frequency
- Shell-like input: type characters, see them on screen
- **Done when:** Type on keyboard → characters appear. Timer ticks visible.

### M4: Physical Memory Manager
- Parse GRUB memory map (multiboot info struct)
- Bitmap allocator: each bit = one 4KB page
- `pmm_alloc()` → returns a free physical page
- `pmm_free()` → marks a page as available
- **Done when:** Can allocate/free pages, reports total/used/free memory

### M5: Paging (Virtual Memory)
- Identity-map the first 4MB (kernel lives here)
- Create page directory + page tables
- Enable paging (set CR3, flip CR0 bit)
- Map/unmap arbitrary virtual → physical pages
- Handle page faults (ISR 14)
- **Done when:** Paging enabled, page fault handler prints fault address

### M6: Process Scheduler
- Process struct: PID, state, register context (saved ESP, EIP, etc.)
- Context switch in assembly (save/restore registers)
- Round-robin scheduler triggered by timer IRQ
- Create 2+ kernel-mode "processes" (functions that print and yield)
- **Done when:** Two processes alternate printing to screen via preemptive switching

### M7: Ramdisk Filesystem
- Simple flat filesystem: header with file entries (name + offset + size)
- GRUB loads ramdisk module into memory
- Kernel reads file list, can `open()` and `read()` files by name
- **Done when:** Kernel lists files from ramdisk, prints contents of a test file

---

## How to Run

```bash
# Build the kernel
make

# Run in QEMU
make run

# Run with debug (GDB attached)
make debug
```

---

## File Structure

```
P3-kernel/
├── asm/
│   ├── boot.asm          # Multiboot header, GDT, jump to kernel
│   ├── isr.asm           # Interrupt service routine stubs
│   └── context_switch.asm # Save/restore process context
├── src/
│   ├── kernel.cpp        # kernel_main() — entry point
│   ├── vga.cpp           # VGA text mode driver
│   ├── gdt.cpp           # GDT setup
│   ├── idt.cpp           # IDT setup + interrupt handlers
│   ├── pic.cpp           # PIC (8259) remapping
│   ├── keyboard.cpp      # PS/2 keyboard driver
│   ├── timer.cpp         # PIT timer driver
│   ├── pmm.cpp           # Physical memory manager (bitmap)
│   ├── paging.cpp        # Virtual memory / page tables
│   ├── scheduler.cpp     # Round-robin process scheduler
│   └── ramdisk.cpp       # Ramdisk filesystem reader
├── include/
│   ├── vga.h
│   ├── gdt.h
│   ├── idt.h
│   ├── pic.h
│   ├── keyboard.h
│   ├── timer.h
│   ├── pmm.h
│   ├── paging.h
│   ├── scheduler.h
│   ├── ramdisk.h
│   └── io.h             # inb/outb port I/O
├── linker.ld             # Linker script (kernel at 1MB)
├── grub.cfg              # GRUB menu entry
├── Makefile
└── README.md
```

---

## Tools Needed

```bash
# Install cross-compiler + QEMU (Ubuntu/Debian)
sudo apt install nasm g++ make qemu-system-i386 grub-pc-bin xorriso mtools
```

**Target:** i686 (32-bit protected mode) — simpler than 64-bit, same concepts.

---

## Resources

- [OSDev Wiki — Bare Bones](https://wiki.osdev.org/Bare_Bones) — The starting tutorial
- [OSDev Wiki — GDT](https://wiki.osdev.org/GDT_Tutorial)
- [OSDev Wiki — IDT](https://wiki.osdev.org/IDT)
- [James Molloy's Kernel Tutorial](http://www.jamesmolloy.co.uk/tutorial_html/)
- [os-tutorial (GitHub)](https://github.com/cfenollosa/os-tutorial)
- [Intel Manual Vol 3A](https://www.intel.com/content/www/us/en/developer/articles/technical/intel-sdm.html) — Chapter 6 (Interrupts)
