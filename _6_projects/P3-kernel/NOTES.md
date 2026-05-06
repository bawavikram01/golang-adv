# Operating Systems — Deep Theory Notes
### Learn the theory. Then see it in OUR code.

---

## Table of Contents

1. [What is an Operating System?](#1-what-is-an-operating-system)
2. [The Boot Process](#2-the-boot-process)
3. [CPU Modes & Privilege Levels](#3-cpu-modes--privilege-levels)
4. [Memory Segmentation & GDT](#4-memory-segmentation--gdt)
5. [Interrupts & Exceptions](#5-interrupts--exceptions)
6. [I/O — Talking to Hardware](#6-io--talking-to-hardware)
7. [Memory Management — Physical](#7-memory-management--physical)
8. [Memory Management — Virtual (Paging)](#8-memory-management--virtual-paging)
9. [Processes & Scheduling](#9-processes--scheduling)
10. [Context Switching](#10-context-switching)
11. [Concurrency & Synchronization](#11-concurrency--synchronization)
12. [Filesystems](#12-filesystems)
13. [System Calls](#13-system-calls)
14. [The Kernel vs. Userspace Boundary](#14-the-kernel-vs-userspace-boundary)

---

## 1. What is an Operating System?

### The Core Job

An OS is a **resource manager** and **abstraction layer**. It manages:

- **CPU time** — who runs, for how long (scheduler)
- **Memory** — who gets which bytes (memory manager)
- **Devices** — who talks to disk, keyboard, network (drivers)
- **Files** — organized access to persistent data (filesystem)

Without an OS, every program would need to:
- Manage its own memory (and could corrupt other programs)
- Directly program hardware (different code for every keyboard model)
- Somehow share the CPU with other programs (cooperative? chaos?)

### The Two Fundamental Abstractions

| Raw Hardware | OS Abstraction |
|--------------|---------------|
| Physical RAM addresses | Virtual address space per process |
| CPU cycles | Processes/threads with scheduled time slices |
| Disk sectors | Files and directories |
| Network packets | Sockets and connections |
| I/O ports | Device-independent read/write |

### Kernel Architectures

```
┌─────────────────────────────────────────────┐
│  MONOLITHIC (Linux, our kernel)             │
│                                             │
│  Everything in kernel space:                │
│  Scheduler, MM, FS, drivers, networking     │
│  + Fast (no IPC overhead)                   │
│  - One bug can crash everything             │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  MICROKERNEL (Minix, QNX, seL4)            │
│                                             │
│  Kernel only does: IPC, scheduling, MM      │
│  Everything else is a userspace server      │
│  + Robust (driver crash doesn't kill kernel)│
│  - Slower (lots of IPC messages)            │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  HYBRID (Windows NT, macOS XNU)            │
│                                             │
│  Mix: some services in kernel, some outside │
│  Pragmatic middle ground                    │
└─────────────────────────────────────────────┘
```

**Our kernel:** Monolithic — everything runs in ring 0. Same as Linux.

---

## 2. The Boot Process

### From Power Button to Your Code

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 1: FIRMWARE (you don't write this)                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. CPU powers on in "real mode" (16-bit, 1MB addressable)      │
│  2. Instruction pointer set to 0xFFFFFFF0 (in BIOS ROM)         │
│  3. BIOS runs POST (Power-On Self-Test):                        │
│     - Checks RAM                                                │
│     - Detects hardware (PCI bus scan)                           │
│     - Initializes video (so you see BIOS splash)                │
│  4. BIOS reads first 512 bytes of boot device (MBR)             │
│  5. Jumps to bootloader code at 0x7C00                          │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ PHASE 2: BOOTLOADER — GRUB (you configure this)                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Stage 1: MBR code (512 bytes, loads Stage 2)                   │
│  Stage 2: Full GRUB in memory                                   │
│     - Shows menu (grub.cfg)                                     │
│     - Reads filesystem to find kernel                           │
│     - Loads kernel ELF into RAM at 1MB+                         │
│     - Switches to 32-bit protected mode                         │
│     - Sets up multiboot info struct (memory map, modules)       │
│     - Jumps to kernel entry point (_start)                      │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ PHASE 3: KERNEL (THIS IS US — asm/boot.asm → src/kernel.cpp)    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Our _start:                                                    │
│     - Set up stack (ESP = stack_top)                            │
│     - Pass multiboot info to C++ (push EBX, EAX)               │
│     - Call kernel_main()                                        │
│                                                                 │
│  kernel_main:                                                   │
│     - Initialize VGA console                                    │
│     - Set up GDT (memory segments)                              │
│     - Set up IDT (interrupt handlers)                           │
│     - Initialize hardware (timer, keyboard)                     │
│     - Enable interrupts (STI)                                   │
│     - Enter idle loop (HLT)                                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### The Multiboot Standard

GRUB expects a "multiboot header" in the first 8KB of your kernel binary:

```
Offset  Field        Value           Meaning
0       magic        0x1BADB002      "I'm a multiboot kernel"
4       flags        0x00000003      Bit 0: align modules on 4K
                                     Bit 1: provide memory map
8       checksum     -(magic+flags)  Must sum to 0 (validation)
```

GRUB checks this. If valid, it loads the kernel and passes:
- **EAX** = 0x2BADB002 (confirmation "you were loaded by multiboot")
- **EBX** = pointer to multiboot info struct (memory map, module list)

**Our code:** `asm/boot.asm` lines 18-21

### Real Mode → Protected Mode → Long Mode

```
Real Mode (16-bit):
  - CPU starts here
  - Only 1MB addressable (20-bit addresses)
  - No memory protection (any code can access anything)
  - BIOS interrupts available (int 0x10 = print, int 0x13 = disk)
  - Segmentation: physical_addr = segment * 16 + offset

Protected Mode (32-bit):     ← OUR KERNEL RUNS HERE
  - 4GB addressable
  - Memory protection via segments (GDT) and pages
  - Ring levels (0-3) for privilege separation
  - BIOS interrupts NO LONGER WORK (must write your own drivers)

Long Mode (64-bit):
  - Requires 64-bit CPU
  - Flat address space, paging mandatory
  - 4-level page tables (PML4)
  - Ubuntu's kernel runs here
```

**Why we use 32-bit:** Same concepts as 64-bit, simpler to learn. 64-bit adds complexity (PML4 paging, canonical addresses, red zone) without teaching new fundamental concepts.

---

## 3. CPU Modes & Privilege Levels

### Ring Model

```
        ┌───────────────────────────┐
        │      Ring 3 (User)        │  ← Applications (ls, bash, Chrome)
        │  ┌───────────────────┐    │
        │  │   Ring 2          │    │  ← Rarely used (device drivers on some OSes)
        │  │  ┌─────────────┐  │    │
        │  │  │  Ring 1      │  │    │  ← Rarely used
        │  │  │  ┌────────┐  │  │    │
        │  │  │  │ Ring 0  │  │  │    │  ← Kernel (full hardware access)
        │  │  │  └────────┘  │  │    │
        │  │  └─────────────┘  │    │
        │  └───────────────────┘    │
        └───────────────────────────┘
```

**What each ring CAN'T do:**

Ring 3 (user mode) CANNOT:
- Execute `cli` / `sti` (disable/enable interrupts)
- Execute `in` / `out` (port I/O — talking to hardware)
- Write to control registers (CR0, CR3, CR4)
- Modify the GDT, IDT, or page tables
- Execute `hlt` (halt CPU)

If Ring 3 code tries any of these → CPU generates a **General Protection Fault** (interrupt 13) → kernel handles it (usually kills the process with SIGSEGV).

**How rings work mechanically:**

The current privilege level (CPL) is stored in the lower 2 bits of the CS register:
- CS = 0x08 → segment 1, ring 0 (kernel code)
- CS = 0x1B → segment 3, ring 3 (user code)

When an interrupt fires, CPU automatically switches to ring 0 (loads kernel's CS/SS from the TSS).

**Our code:** `src/gdt.cpp` — entries 1-2 are ring 0, entries 3-4 are ring 3.

### How Linux Uses Rings

Linux only uses ring 0 (kernel) and ring 3 (user). Rings 1-2 are ignored.
- Kernel code: ring 0 — can do anything
- User code: ring 3 — restricted
- Transition: via **system calls** (software interrupt or `syscall` instruction)

---

## 4. Memory Segmentation & GDT

### What Segmentation Is

In protected mode, EVERY memory access goes through a segment:

```
Logical Address:   segment_selector : offset
                        ↓
              CPU looks up GDT[selector]
                        ↓
Physical Address = base + offset  (with limit/permission checks)
```

### GDT Entry Format (8 bytes each)

```
Byte 7  Byte 6  Byte 5  Byte 4  Byte 3  Byte 2  Byte 1  Byte 0
┌───────┬───────┬───────┬───────┬───────┬───────┬───────┬───────┐
│Base   │Flags +│Access │Base   │Base           │Limit          │
│[31:24]│Limit  │Byte   │[23:16]│[15:0]         │[15:0]         │
│       │[19:16]│       │       │               │               │
└───────┴───────┴───────┴───────┴───────┴───────┴───────┴───────┘
```

**Access Byte breakdown:**
```
Bit 7: Present (1 = valid segment)
Bits 6-5: DPL (Descriptor Privilege Level — ring 0, 1, 2, or 3)
Bit 4: Descriptor type (1 = code/data, 0 = system)
Bit 3: Executable (1 = code, 0 = data)
Bit 2: Direction/Conforming
Bit 1: Readable/Writable
Bit 0: Accessed (CPU sets this)
```

**Kernel Code (0x9A):** `10011010`
- Present=1, DPL=00(ring 0), Type=1, Exec=1, Conform=0, Read=1, Accessed=0

**Kernel Data (0x92):** `10010010`
- Present=1, DPL=00(ring 0), Type=1, Exec=0, Conform=0, Write=1, Accessed=0

### Flat Model (What we use)

```
Segment 1 (code): base=0x00000000, limit=0xFFFFFFFF (all 4GB)
Segment 2 (data): base=0x00000000, limit=0xFFFFFFFF (all 4GB)

Result: logical address = physical address (no translation)
```

This effectively disables segmentation while keeping the CPU happy (it REQUIRES a GDT in protected mode).

### Loading the GDT

```asm
lgdt [gdt_pointer]    ; Tell CPU: "GDT is at this address, this size"

; Must reload ALL segment registers after loading GDT:
mov ax, 0x10          ; 0x10 = offset of kernel data segment (entry 2 * 8 bytes)
mov ds, ax            ; Data segment
mov es, ax            ; Extra segment
mov fs, ax            ; General purpose segment
mov gs, ax            ; General purpose segment
mov ss, ax            ; Stack segment

jmp 0x08:.done        ; 0x08 = kernel code segment — far jump reloads CS
.done:
```

**Why the far jump?** CS (code segment) can't be loaded with `mov`. The only way is a far jump (`jmp segment:offset`) or an interrupt return (`iret`).

**Our code:** `asm/boot.asm` → `gdt_flush` function

---

## 5. Interrupts & Exceptions

### Three Types of Interrupts

```
1. EXCEPTIONS (synchronous — caused by CPU executing your code)
   - Faults: can be fixed and instruction retried (e.g., page fault)
   - Traps: reported after instruction (e.g., breakpoint, int 3)
   - Aborts: unrecoverable (e.g., double fault)

2. HARDWARE INTERRUPTS / IRQs (asynchronous — hardware signals)
   - IRQ0: Timer (PIT/APIC)
   - IRQ1: Keyboard
   - IRQ14: Primary ATA disk
   - These fire at ANY time, regardless of what CPU is doing

3. SOFTWARE INTERRUPTS (int N instruction)
   - int 0x80: Linux system call (old method)
   - int 3: Debugger breakpoint
```

### CPU Exception Table (ISRs 0-31)

| # | Name | Type | Error Code? | What Triggers It |
|---|------|------|-------------|-----------------|
| 0 | Division Error | Fault | No | `div` or `idiv` by zero |
| 1 | Debug | Trap/Fault | No | Single-step, breakpoints |
| 2 | NMI | Interrupt | No | Non-maskable (hardware critical) |
| 3 | Breakpoint | Trap | No | `int 3` instruction |
| 6 | Invalid Opcode | Fault | No | CPU can't decode instruction |
| 8 | Double Fault | Abort | Yes (0) | Exception during exception handler |
| 13 | General Protection Fault | Fault | Yes | Privilege violation, bad segment |
| 14 | Page Fault | Fault | Yes | Access to unmapped/protected page |

### What the CPU Does When an Interrupt Fires

```
1. If privilege change (ring 3 → ring 0):
   - Load kernel stack pointer from TSS
   - Push old SS and ESP

2. Always:
   - Push EFLAGS (save interrupt flag, etc.)
   - Push CS
   - Push EIP (return address)
   - If exception with error code: push error code
   - Clear IF flag (disable further interrupts)
   - Load CS:EIP from IDT[interrupt_number]
   - Start executing handler
```

**Stack after interrupt (no privilege change):**
```
[top of stack]
  ESP+12: EFLAGS
  ESP+8:  CS
  ESP+4:  EIP         ← where to return
  ESP+0:  Error Code  ← only for some exceptions
```

### IDT Entry Format

```
Bits 0-15:  Handler address [15:0]    (low half)
Bits 16-31: Segment selector (0x08 = kernel code)
Bits 32-39: Reserved (0)
Bits 40-43: Gate type (0xE = 32-bit interrupt gate)
Bit 44:     0
Bits 45-46: DPL (who can trigger this with 'int' instruction)
Bit 47:     Present
Bits 48-63: Handler address [31:16]   (high half)
```

**Interrupt Gate vs Trap Gate:**
- Interrupt gate: CPU clears IF (disables further interrupts) — use for hardware IRQs
- Trap gate: CPU does NOT clear IF — use for software exceptions

### The PIC (8259A)

The PIC is the bridge between hardware devices and the CPU:

```
           ┌─────────────┐
IRQ0 ─────►│             │
IRQ1 ─────►│  Master PIC │──────► CPU INTR pin
IRQ2 ─────►│  (0x20-0x21)│
IRQ3 ─────►│             │
IRQ4 ─────►│             │
IRQ5 ─────►│             │
IRQ6 ─────►│             │
IRQ7 ─────►│             │
           └─────────────┘
                  ▲
                  │ IRQ2 (cascade)
           ┌─────┴───────┐
IRQ8 ─────►│             │
IRQ9 ─────►│  Slave PIC  │
IRQ10 ────►│  (0xA0-0xA1)│
IRQ11 ────►│             │
IRQ12 ────►│             │
IRQ13 ────►│             │
IRQ14 ────►│             │
IRQ15 ────►│             │
           └─────────────┘
```

**Remapping is essential** — default IRQ0-7 overlap with CPU exceptions 8-15:
```
IRQ0 (timer) → INT 8 (Double Fault!)  ← DISASTER
IRQ1 (keybd) → INT 9 (Coprocessor)    ← WRONG

After remapping:
IRQ0 (timer) → INT 32  ← Clean
IRQ1 (keybd) → INT 33  ← Clean
```

**Our code:** `src/idt.cpp` → `pic_remap()` function

### End of Interrupt (EOI)

After handling an IRQ, you MUST send EOI to the PIC. Otherwise it won't send more interrupts:
```cpp
outb(0x20, 0x20);  // EOI to master PIC
// If IRQ >= 8:
outb(0xA0, 0x20);  // EOI to slave PIC too
```

---

## 6. I/O — Talking to Hardware

### Two Methods

**1. Port-Mapped I/O (PMIO)** — x86 specific
```
CPU has a separate 16-bit I/O address space (0x0000 - 0xFFFF)
Access via IN/OUT instructions:

outb(0x60, data)   →  "Write 'data' to port 0x60"
inb(0x60)          →  "Read a byte from port 0x60"
```

Common ports:
| Port | Device |
|------|--------|
| 0x20-0x21 | Master PIC |
| 0x40-0x43 | PIT (timer) |
| 0x60 | Keyboard data |
| 0x64 | Keyboard command/status |
| 0xA0-0xA1 | Slave PIC |
| 0x3D4-0x3D5 | VGA cursor control |
| 0x1F0-0x1F7 | Primary ATA/IDE disk |
| 0x3F8 | COM1 serial port |

**2. Memory-Mapped I/O (MMIO)** — universal
```
Device registers mapped to physical memory addresses.
Read/write normal memory addresses → actually talks to hardware.

Example: VGA buffer at 0xB8000
```

MMIO is how modern hardware works (PCIe devices, APIC, framebuffers). Port I/O is legacy x86.

**Our code:** `include/io.h` — `inb()` and `outb()` inline assembly functions

### How inb/outb Work (Inline Assembly)

```cpp
static inline void outb(uint16_t port, uint8_t data) {
    asm volatile("outb %0, %1" : : "a"(data), "Nd"(port));
}
//  "outb %0, %1"  → the instruction
//  "a"(data)       → put 'data' in AL register (required by outb)
//  "Nd"(port)      → put 'port' in DX or use immediate (required by outb)
//  volatile        → don't optimize this away, it has side effects

static inline uint8_t inb(uint16_t port) {
    uint8_t result;
    asm volatile("inb %1, %0" : "=a"(result) : "Nd"(port));
    return result;
}
```

---

## 7. Memory Management — Physical

### The Problem

After boot, you have some amount of RAM. But how do you know:
- How much RAM is there?
- Which parts are usable? (Some is reserved for BIOS, MMIO, etc.)
- How to give out memory and get it back?

### Memory Map (from GRUB)

GRUB passes a memory map in the multiboot info struct:

```
Region 0: 0x00000000 - 0x0009FBFF (639 KB)   — USABLE
Region 1: 0x0009FC00 - 0x000FFFFF (1 KB)     — RESERVED (BIOS/video)
Region 2: 0x00100000 - 0x07FDFFFF (126 MB)   — USABLE (our kernel is here)
Region 3: 0x07FE0000 - 0x07FFFFFF (128 KB)   — RESERVED (ACPI)
Region 4: 0xFEC00000 - 0xFFFFFFFF (20 MB)    — RESERVED (MMIO, BIOS ROM)
```

Only "USABLE" regions can be allocated. We must NOT touch reserved regions.

### Physical Memory Layout (typical)

```
0x00000000 ┌──────────────────┐
           │   Real Mode IVT  │ (1 KB — interrupt vector table)
0x00000400 ├──────────────────┤
           │   BIOS Data Area │
0x00000500 ├──────────────────┤
           │   Free (usable)  │ (~638 KB)
0x0007FFFF ├──────────────────┤
           │   EBDA           │ (Extended BIOS Data Area)
0x0009FFFF ├──────────────────┤
           │   Video RAM      │ (VGA at 0xB8000, framebuffer)
0x000BFFFF ├──────────────────┤
           │   BIOS ROM       │ (mapped from ROM chip)
0x000FFFFF ├──────────────────┤  ← 1 MB mark
           │   OUR KERNEL     │ (loaded here by GRUB)
           │   .text          │
           │   .data          │
           │   .bss           │
0x00?????? ├──────────────────┤  ← __kernel_end (from linker.ld)
           │   FREE RAM       │ ← THIS is what we manage!
           │   ...            │
           │   (up to ~128MB+)│
0x07FFFFFF └──────────────────┘
```

### Allocation Strategy: Bitmap Allocator

We divide all usable RAM into 4KB pages and maintain a bitmap:

```
Page 0 (0x00000 - 0x00FFF): bit 0
Page 1 (0x01000 - 0x01FFF): bit 1
Page 2 (0x02000 - 0x02FFF): bit 2
...etc

Bitmap: [1][1][1]...[1][0][0][0][0]...
         ^               ^
         Kernel pages    Free pages
         (marked used)   (available to allocate)
```

- `pmm_alloc()` → find first 0 bit, set it to 1, return address
- `pmm_free(addr)` → set the bit back to 0

**Why 4KB pages?** x86 paging uses 4KB as the smallest unit. By allocating in 4KB chunks, every allocated page can be mapped into a virtual address space.

### How Linux Does It

Linux uses a **buddy allocator**:
- Free pages organized in lists by power-of-2 sizes: 1, 2, 4, 8, 16... pages
- To allocate 1 page: take from the "1-page" list
- If empty, split a "2-page" block
- To free: if the "buddy" (adjacent block) is also free, merge into larger block

This is like your P2 coalescing, but for whole pages!

On top of buddy allocator, Linux has **slab allocator** for small objects (like `struct task_struct` — you allocate thousands of them, all same size).

**Our code (M4):** `src/pmm.cpp` (bitmap allocator — simpler but same concept)

---

## 8. Memory Management — Virtual (Paging)

### The Fundamental Problem

Without paging:
- Every process sees the SAME physical memory
- Process A can corrupt process B's data
- No memory isolation
- Loading programs at fixed addresses is fragile

With paging:
- Every process has its OWN virtual address space
- Process A's address 0x400000 ≠ Process B's address 0x400000
- They map to DIFFERENT physical pages
- The kernel controls the mapping

### How Paging Works (x86 32-bit, 2-level)

```
Virtual Address (32 bits):
┌──────────┬──────────┬──────────────┐
│ Dir (10) │ Table(10)│ Offset (12)  │
└──────────┴──────────┴──────────────┘
     │           │            │
     │           │            └──► Byte within the 4KB page
     │           │
     │           └──► Index into Page Table (1024 entries)
     │
     └──► Index into Page Directory (1024 entries)
```

**Translation process:**
```
1. CPU reads CR3 register → Physical address of Page Directory
2. Page Directory[dir_index] → Physical address of Page Table
3. Page Table[table_index] → Physical address of the page
4. Final physical address = page_base + offset
```

**Visual:**
```
CR3 = 0x1000 (address of page directory)

Page Directory (at 0x1000):
┌─────────────────────────────────────┐
│ Entry 0:  Page Table at 0x5000      │
│ Entry 1:  Page Table at 0x6000      │
│ Entry 2:  NOT PRESENT               │ ← access = page fault!
│ ...                                 │
│ Entry 1023: ...                     │
└─────────────────────────────────────┘

Page Table at 0x5000:
┌─────────────────────────────────────┐
│ Entry 0:  Physical page 0x00000     │
│ Entry 1:  Physical page 0x01000     │
│ Entry 2:  Physical page 0xA7000     │ ← doesn't have to be sequential!
│ ...                                 │
└─────────────────────────────────────┘
```

### Page Table Entry Format

```
Bit 0:  Present (is this page mapped?)
Bit 1:  Read/Write (0=read-only, 1=writable)
Bit 2:  User/Supervisor (0=kernel only, 1=user accessible)
Bit 3:  Write-Through
Bit 4:  Cache Disable
Bit 5:  Accessed (CPU sets this when page is read)
Bit 6:  Dirty (CPU sets this when page is written)
Bit 7:  Page size (0=4KB, 1=4MB for directory entries)
Bits 12-31: Physical page frame number (address >> 12)
```

### Identity Mapping vs. Higher-Half Kernel

**Identity mapping:** virtual address = physical address
```
Virtual 0x100000 → Physical 0x100000
```
Simple, but means kernel and userspace fight over the same address range.

**Higher-half kernel:** kernel lives in upper virtual addresses
```
Virtual 0xC0000000 → Physical 0x00100000  (kernel)
Virtual 0x00400000 → Physical 0x07F00000  (userspace program)
```
Every process maps the kernel at 0xC0000000+. This is what Linux does.

**Our code (M5):** We identity-map the first few MB (simple). `src/paging.cpp`

### Enabling Paging

```cpp
// 1. Set up page directory and tables
// 2. Point CR3 to page directory
asm volatile("mov %0, %%cr3" : : "r"(page_directory_addr));

// 3. Enable paging (set bit 31 of CR0)
uint32_t cr0;
asm volatile("mov %%cr0, %0" : "=r"(cr0));
cr0 |= 0x80000000;
asm volatile("mov %0, %%cr0" : : "r"(cr0));
// NOW every memory access goes through page tables!
```

### Page Faults

When the CPU accesses an address that:
- Has `Present = 0` in page table → **Page Fault** (interrupt 14)
- Is marked read-only and you write to it → **Page Fault**
- Is marked kernel-only and user code accesses it → **Page Fault**

The page fault handler receives:
- **Error code:** bits telling you WHY (read/write? user/kernel? page not present?)
- **CR2 register:** the virtual address that caused the fault

This is how demand paging, copy-on-write, and memory-mapped files work!

### How Linux Uses This

```
Process A's virtual space:          Process B's virtual space:
0x00000000: [not mapped]            0x00000000: [not mapped]
0x00400000: [A's code]              0x00400000: [B's code]
0x00600000: [A's data]              0x00600000: [B's data]
0x08000000: [A's heap]              ...[shared library]
0xC0000000: [KERNEL]                0xC0000000: [KERNEL]  ← same mapping!
```

When CPU switches from A to B:
- Scheduler changes CR3 to B's page directory
- Boom — completely different memory view
- B can't see A's data (not in its page tables)
- But both can see the kernel (same kernel entries in both directories)

---

## 9. Processes & Scheduling

### What is a Process?

A process is a running program. It consists of:

```
┌─────────────────────────────────────┐
│ Process Control Block (PCB)         │
├─────────────────────────────────────┤
│ PID (unique identifier)             │
│ State (running/ready/blocked/dead)  │
│ Program Counter (EIP)               │
│ Registers (EAX, EBX, ... ESP, EBP) │
│ Page Directory (CR3)                │
│ Stack pointer                       │
│ Priority                            │
│ Open files                          │
│ Parent PID                          │
│ Signal handlers                     │
└─────────────────────────────────────┘
```

### Process States

```
        ┌────────────────────────────────────────────┐
        │                                            │
        ▼                                            │
    ┌────────┐    schedule    ┌─────────┐   I/O done │
    │ READY  │──────────────►│ RUNNING │            │
    │(in run │◄──────────────│         │            │
    │ queue) │   preempted   └────┬────┘            │
    └────────┘                    │                 │
                                  │ wait for I/O    │
                                  ▼                 │
                            ┌──────────┐            │
                            │ BLOCKED  │────────────┘
                            │(waiting) │
                            └──────────┘
```

- **Running:** CPU is executing this process right now
- **Ready:** Could run, waiting for CPU time
- **Blocked:** Waiting for something (disk read, network, sleep timer)

### Scheduling Algorithms

**Round-Robin (what we implement):**
```
Ready queue: [P1] → [P2] → [P3] → [P1] → ...

Each process gets a fixed time quantum (e.g., 10ms).
Timer fires → switch to next in queue.
Simple, fair, but not optimal for all workloads.
```

**Priority Scheduling:**
```
Higher priority processes run first.
Problem: starvation (low-priority processes never run)
Solution: aging (priority increases over time)
```

**Linux's CFS (Completely Fair Scheduler):**
```
Red-black tree sorted by "virtual runtime" (vruntime).
Process with LEAST vruntime runs next.
Every process gets fair share of CPU time.
Nice values adjust the rate vruntime increases.
```

### Preemptive vs. Cooperative Multitasking

**Cooperative (old Mac OS, Windows 3.1):**
- Process runs until it voluntarily yields: `yield()`
- If a program has an infinite loop → entire system hangs
- Bad.

**Preemptive (Linux, our kernel):**
- Timer interrupt fires every 10ms
- Kernel FORCIBLY saves process state and switches to another
- No process can hog the CPU
- This is what our M6 implements!

**Our code (M6):** Timer callback → `schedule()` → switch to next process

---

## 10. Context Switching

### The Most Important Operation in an OS

A context switch is: save the state of process A, load the state of process B.

```
Process A running:
  EAX=5, EBX=10, EIP=0x401000, ESP=0x7FFF1000, CR3=0x1000

Timer interrupt fires!
Kernel saves A's registers to A's PCB.

Kernel loads B's registers from B's PCB:
  EAX=42, EBX=99, EIP=0x402000, ESP=0x7FFF2000, CR3=0x2000

IRET → CPU now executing process B. Process A is frozen.
```

### What Gets Saved/Restored

```
┌─────────────────────────────────────────────────────┐
│ Saved by CPU automatically (on interrupt/exception): │
│   - EIP (instruction pointer)                        │
│   - CS  (code segment)                               │
│   - EFLAGS (flags register)                          │
│   - ESP, SS (if privilege change)                    │
├─────────────────────────────────────────────────────┤
│ Saved by us manually (in assembly stub):             │
│   - EAX, EBX, ECX, EDX, ESI, EDI, EBP, ESP         │
│   - DS, ES, FS, GS (data segments)                  │
│   - CR3 (page directory — if switching address spaces)│
│   - FPU/SSE state (if process uses floating point)   │
└─────────────────────────────────────────────────────┘
```

### Context Switch Cost

A context switch is EXPENSIVE:
1. Save ~20 registers to memory
2. Invalidate CPU caches (TLB flush when CR3 changes)
3. Cold cache for new process (cache misses)
4. Pipeline flush

**Cost:** ~1-5 microseconds on modern hardware. Sounds small, but at 1000 switches/sec, that's real overhead.

**Our code (M6):** `asm/context_switch.asm` — save ESP, swap to new process's ESP, restore

### The Assembly

```asm
; context_switch(old_esp_ptr, new_esp)
; Save current process's state, switch to new process
context_switch:
    ; Save callee-saved registers on current stack
    push ebp
    push ebx
    push esi
    push edi

    ; Save current ESP into old process's struct
    mov eax, [esp + 20]    ; old_esp_ptr
    mov [eax], esp         ; *old_esp_ptr = current ESP

    ; Load new process's ESP
    mov esp, [esp + 24]    ; new_esp

    ; Restore new process's registers
    pop edi
    pop esi
    pop ebx
    pop ebp

    ret   ; Returns to wherever the new process was last interrupted
```

---

## 11. Concurrency & Synchronization

### The Problem

Even with a single CPU, preemptive interrupts can cause races:

```cpp
// Process A:                     // Timer interrupt can fire HERE:
counter = counter + 1;
// Really:
//   mov eax, [counter]   ← interrupt fires after this
//   add eax, 1           ← but before this
//   mov [counter], eax   ← counter update lost!
```

### Disabling Interrupts (simplest, what we use)

```cpp
asm volatile("cli");  // Disable interrupts — no preemption
// ... critical section (modify shared data) ...
asm volatile("sti");  // Re-enable interrupts
```

**Pros:** Simple, works on single CPU.
**Cons:** Doesn't work on multi-core. Delays interrupt handling.

### Spinlocks (multi-core)

```cpp
// Busy-wait until lock is free
void spin_lock(int* lock) {
    while (__sync_lock_test_and_set(lock, 1)) {
        // spin (waste CPU cycles)
    }
}

void spin_unlock(int* lock) {
    __sync_lock_release(lock);
}
```

Uses atomic instructions (like `xchg`) that the CPU guarantees are indivisible.

### Mutexes, Semaphores (for userspace)

- **Mutex:** Binary lock. One thread holds it, others sleep.
- **Semaphore:** Counting lock. N threads can enter simultaneously.
- **Condition Variable:** Sleep until some condition is signaled.

These are implemented ON TOP of the kernel's spinlocks + scheduler:
```
pthread_mutex_lock():
   Try atomic lock
   If fails → syscall to kernel
   Kernel puts thread in BLOCKED state
   Scheduler runs another thread
   When lock is released → kernel wakes up blocked thread
```

---

## 12. Filesystems

### What a Filesystem Does

Maps human-friendly names to disk locations:
```
"/home/vikram/hello.c" → disk sectors 4571-4573
```

### Layers of Abstraction

```
Application:  open("/home/vikram/hello.c", O_RDONLY)
     ↓
VFS (Virtual Filesystem Switch):  abstracts different FS types
     ↓
Filesystem driver (ext4, FAT32, etc.): translates path → disk blocks
     ↓
Block layer: reads/writes disk sectors
     ↓
Disk driver: talks to hardware (SATA, NVMe)
     ↓
Hardware: physical disk
```

### Simple Filesystem Layout (our ramdisk)

```
┌──────────────────────────────────────────────┐
│ Header:                                      │
│   magic: 0xDEADBEEF                          │
│   num_files: 3                               │
├──────────────────────────────────────────────┤
│ File Entry 0:                                │
│   name: "hello.txt"                          │
│   offset: 512                                │
│   size: 13                                   │
├──────────────────────────────────────────────┤
│ File Entry 1:                                │
│   name: "kernel.log"                         │
│   offset: 525                                │
│   size: 100                                  │
├──────────────────────────────────────────────┤
│ ... (file contents at their offsets) ...     │
└──────────────────────────────────────────────┘
```

**Our code (M7):** `src/ramdisk.cpp` — reads this flat structure from memory.

### Real Filesystems (ext4)

ext4 uses:
- **Superblock:** filesystem metadata (total blocks, free count, etc.)
- **Inodes:** one per file, contains permissions, size, pointers to data blocks
- **Directory entries:** map filenames → inode numbers
- **Data blocks:** actual file content
- **Extent trees:** efficiently map large files to contiguous disk blocks
- **Journal:** write-ahead log for crash recovery

---

## 13. System Calls

### The Bridge Between Userspace and Kernel

User programs CAN'T access hardware directly (ring 3). They ask the kernel via syscalls:

```
User code (ring 3):
    mov eax, 4       ; syscall number 4 = sys_write
    mov ebx, 1       ; fd = stdout
    mov ecx, msg     ; buffer address
    mov edx, 13      ; length
    int 0x80         ; TRIGGER SYSCALL (old Linux method)
    ↓
CPU: ring 3 → ring 0 transition
    ↓
Kernel (ring 0):
    Looks up syscall table[4] → sys_write()
    Validates user pointer (is ecx a valid address for this process?)
    Writes to file descriptor 1
    Returns result in EAX
    ↓
IRET: ring 0 → ring 3
    ↓
User code continues, result in EAX
```

### Modern Syscall Method (sysenter/syscall)

`int 0x80` is slow (full interrupt mechanism). Modern CPUs have dedicated instructions:
- `syscall` / `sysret` (AMD, 64-bit)
- `sysenter` / `sysexit` (Intel, 32-bit)

These are faster — no IDT lookup, direct ring switch.

### Common Linux Syscalls

| # | Name | What it does |
|---|------|-------------|
| 1 | exit | Terminate process |
| 2 | fork | Create child process (copy of parent) |
| 3 | read | Read from file descriptor |
| 4 | write | Write to file descriptor |
| 5 | open | Open file, get fd |
| 6 | close | Close file descriptor |
| 11 | execve | Replace process image with new program |
| 45 | brk | Expand heap (your P2 sbrk!) |
| 90 | mmap | Map memory (your P2 mmap!) |

You used syscalls in P1 (shell): `fork()`, `execvp()`, `waitpid()`, `pipe()`, `dup2()` — all are thin wrappers around kernel syscalls.

---

## 14. The Kernel vs. Userspace Boundary

### The Full Picture

```
┌──────────────────────────────────────────────────────────────┐
│                        USERSPACE (Ring 3)                     │
│                                                              │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐       │
│  │  bash   │  │  nginx  │  │  Chrome  │  │   ls    │       │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘       │
│       │             │            │             │             │
│  ┌────┴─────────────┴────────────┴─────────────┴──────┐     │
│  │              C Library (glibc)                       │     │
│  │   printf → write()   malloc → brk()/mmap()         │     │
│  └────────────────────────────┬────────────────────────┘     │
│                               │                              │
├───────────────────────────────┼──────────────────────────────┤
│                               │ SYSCALL                      │
│                               ▼                              │
│                        KERNEL (Ring 0)                        │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐     │
│  │              System Call Table                        │     │
│  │   0: restart_syscall  1: exit  2: fork  3: read ... │     │
│  └────────────────────────────┬────────────────────────┘     │
│                               │                              │
│  ┌────────────┐  ┌───────────┴───┐  ┌──────────────┐       │
│  │  Scheduler │  │    VFS        │  │ Memory Mgmt  │       │
│  │            │  │  (open,read,  │  │ (page alloc, │       │
│  │ (CFS,     │  │   write)      │  │  page tables)│       │
│  │  timeslice)│  └───────┬───────┘  └──────────────┘       │
│  └────────────┘          │                                   │
│                          │                                   │
│  ┌───────────────────────┴───────────────────────────┐      │
│  │              Device Drivers                        │      │
│  │   disk, network, USB, GPU, keyboard, timer...     │      │
│  └───────────────────────────┬───────────────────────┘      │
│                              │                               │
├──────────────────────────────┼───────────────────────────────┤
│                              ▼                               │
│                        HARDWARE                              │
│         CPU, RAM, Disk, Network Card, Keyboard, GPU          │
└──────────────────────────────────────────────────────────────┘
```

### What We're Building vs. Ubuntu

| Layer | Our Kernel (P3) | Ubuntu (Linux) |
|-------|-----------------|----------------|
| Boot | Multiboot + GRUB | EFI stub + GRUB |
| Segments | Flat GDT | Same flat GDT + TSS |
| Interrupts | 48 handlers (M2) | 256 + APIC + MSI-X |
| Timer | PIT 100Hz (M3) | HPET/APIC tickless |
| Keyboard | Raw PS/2 (M3) | Input subsystem + evdev |
| Physical Mem | Bitmap (M4) | Buddy system + slab |
| Virtual Mem | 2-level paging (M5) | 4-level (PML4), demand paging, CoW |
| Processes | Round-robin (M6) | CFS, cgroups, namespaces |
| Filesystem | Flat ramdisk (M7) | VFS + ext4 + 50 FS types |
| Syscalls | None (future) | 400+ syscalls |
| Networking | None | Full TCP/IP + socket API |
| USB/PCI | None | Hundreds of drivers |

---

## Key Takeaways

1. **The OS is the first program that runs.** It sets up everything from scratch.
2. **Interrupts are the heartbeat.** Without timer interrupts, no multitasking.
3. **Paging gives isolation.** One process can't see another's memory.
4. **Context switches are mechanical.** Save registers, swap stack, restore registers.
5. **Syscalls are the API.** The ONLY way userspace talks to the kernel.
6. **Everything is built in layers.** Each layer trusts the one below and provides abstraction above.

---

## Recommended Reading Order

1. **Now:** Finish P3 milestones (hands-on > theory alone)
2. **Parallel reading:** OSDev Wiki articles as you implement each milestone
3. **After P3:** Read "Operating Systems: Three Easy Pieces" (free online — ostep.org)
4. **Deep dive:** "Computer Systems: A Programmer's Perspective" (Chapter 8-9: Exceptions, Virtual Memory)
5. **Linux specifics:** "The Linux Programming Interface" (Kerrisk) — the bible for Linux syscalls

---

*Last updated: May 2026*
*Paired with: P3-kernel code in this same directory*
