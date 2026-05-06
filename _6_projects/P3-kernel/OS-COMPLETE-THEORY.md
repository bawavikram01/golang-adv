# Operating Systems: Complete Theory + Code
## From Zero to God Level — Everything You Need to Know

---

# PART 1: THE HARDWARE FOUNDATION

## 1.1 What is a CPU?

A CPU is a state machine. It has:
- **Registers** — tiny, fast storage inside the CPU (like 8 variables)
- **ALU** — does math (add, subtract, compare, shift)
- **Control Unit** — fetches instructions, decodes them, executes them

### The Fetch-Decode-Execute Cycle

```
┌─────────────────────────────────────────────────┐
│                 FOREVER:                         │
│                                                 │
│  1. FETCH:   Read instruction from RAM[EIP]     │
│  2. DECODE:  Figure out what instruction means   │
│  3. EXECUTE: Do the operation (add, mov, jmp)   │
│  4. EIP++:   Move to next instruction           │
│                                                 │
│  (repeat billions of times per second)          │
└─────────────────────────────────────────────────┘
```

That's it. That's all a CPU does. Everything — Chrome, Linux, games — is just this loop running different instructions.

### x86 Registers (32-bit mode — our kernel uses these)

```
GENERAL PURPOSE (you use these for anything):
┌─────────────────────────────────────────────────────────┐
│  EAX  — Accumulator (function return values go here)    │
│  EBX  — Base (general purpose, callee-saved)            │
│  ECX  — Counter (loop counters, shift counts)           │
│  EDX  — Data (I/O operations, multiply/divide overflow) │
│  ESI  — Source Index (string operations source)         │
│  EDI  — Destination Index (string operations dest)      │
│  EBP  — Base Pointer (stack frame base)                 │
│  ESP  — Stack Pointer (top of stack — ALWAYS current)   │
└─────────────────────────────────────────────────────────┘

SPECIAL:
┌─────────────────────────────────────────────────────────┐
│  EIP    — Instruction Pointer (WHERE the CPU is)        │
│  EFLAGS — Status flags (zero, carry, interrupt enable)  │
│  CR0    — Control register 0 (paging enable bit!)       │
│  CR2    — Page fault address (set by CPU on fault)      │
│  CR3    — Page directory address (the paging root!)     │
│  CR4    — Extensions enable (PAE, PSE, etc.)            │
└─────────────────────────────────────────────────────────┘

SEGMENT:
┌─────────────────────────────────────────────────────────┐
│  CS  — Code Segment (which GDT entry for code)          │
│  DS  — Data Segment (which GDT entry for data)          │
│  SS  — Stack Segment                                    │
│  ES, FS, GS — Extra segments                            │
└─────────────────────────────────────────────────────────┘
```

### Why Registers Matter for OS

When the CPU switches from Process A to Process B:
- ALL registers contain A's state (A's EIP, A's ESP, A's EAX...)
- We must SAVE all registers → load B's saved registers
- Now CPU "thinks" it's been running B all along

**Our code** (`asm/context_switch.asm`):
```asm
; Save process A's registers
push ebp
push ebx
push esi
push edi
mov [eax], esp     ; Save A's stack pointer

; Load process B's registers
mov esp, [new_esp] ; Switch to B's stack
pop edi            ; B's saved registers come off B's stack
pop esi
pop ebx
pop ebp
ret                ; "return" to wherever B was last running
```

---

## 1.2 Memory Hierarchy

```
Speed:      FASTEST ──────────────────────────────────► SLOWEST
Capacity:   TINY ─────────────────────────────────────► HUGE

┌──────────┐  1 cycle     ┌──────────┐  ~3-5 cycles   ┌──────────────┐
│ Registers│─────────────►│ L1 Cache │───────────────►│  L2 Cache    │
│  ~1 KB   │              │  32-64KB │                │  256KB-1MB   │
└──────────┘              └──────────┘                └──────────────┘
                                                            │
                                      ~10-30 cycles         │
                                                            ▼
                                                     ┌──────────────┐
                                                     │  L3 Cache    │
                                                     │  2-64 MB     │
                                                     └──────┬───────┘
                                                            │ ~100 cycles
                                                            ▼
                                                     ┌──────────────┐
                                                     │     RAM      │
                                                     │  8-128 GB    │
                                                     └──────┬───────┘
                                                            │ ~10,000,000 cycles
                                                            ▼
                                                     ┌──────────────┐
                                                     │  Disk (SSD)  │
                                                     │  256GB - TB  │
                                                     └──────────────┘
```

**Why this matters for OS:**
- **TLB** (Translation Lookaside Buffer) = cache for page table entries. Without it, every memory access would need 2-3 extra RAM accesses to walk page tables.
- When you switch CR3 (change address space), the TLB gets flushed — that's why context switches are expensive.
- **Cache lines** = 64 bytes. CPU loads 64 bytes at a time. This is why sequential memory access is fast and random access is slow.

---

## 1.3 The Bus Architecture

```
┌───────────────────────────────────────────────────────────┐
│                          CPU                               │
│  ┌─────┐  ┌─────┐  ┌─────┐                              │
│  │Core0│  │Core1│  │Core2│  ...                          │
│  └──┬──┘  └──┬──┘  └──┬──┘                              │
│     └────────┴────────┴─────── L3 Cache                  │
└────────────────────┬──────────────────────────────────────┘
                     │ Memory Bus (DDR4/5)
                     ▼
              ┌──────────────┐
              │     RAM      │
              └──────┬───────┘
                     │ System Bus
                     ▼
         ┌───────────────────────┐
         │     Chipset / PCH     │
         └───┬──────┬───────┬───┘
             │      │       │
        ┌────┘   ┌──┘    ┌──┘
        ▼        ▼       ▼
    ┌──────┐  ┌────┐  ┌──────┐
    │ NVMe │  │USB │  │ NIC  │    (devices)
    │ SSD  │  │    │  │      │
    └──────┘  └────┘  └──────┘
```

Devices talk to the CPU via:
1. **Port I/O** (legacy): `inb(port)` / `outb(port, data)` — our keyboard uses this
2. **MMIO** (modern): Device registers mapped to physical addresses — our VGA uses this
3. **DMA** (fast): Device reads/writes RAM directly without CPU involvement

---

## 1.4 The Stack

The stack is just RAM with a convention: grows DOWNWARD, ESP points to the top.

```
High address
┌──────────────────────────┐  0xFFFF
│   function arguments     │
├──────────────────────────┤
│   return address (EIP)   │  ← pushed by CALL instruction
├──────────────────────────┤
│   saved EBP (old frame)  │  ← pushed by function prologue
├──────────────────────────┤  ← EBP points here (frame base)
│   local variable 1       │
│   local variable 2       │
│   local variable 3       │
├──────────────────────────┤  ← ESP points here (current top)
│   (unused space below)   │
│                          │
└──────────────────────────┘  0x0000
Low address
```

**Function call convention (cdecl, what we use):**
```asm
; Calling: result = add(3, 5)
push 5           ; Push arg2 (right to left)
push 3           ; Push arg1
call add         ; Push return address, jump to add

; Inside add():
push ebp         ; Save old frame pointer
mov ebp, esp     ; Set up new frame
sub esp, 8       ; Space for locals
; ... do work ...
; [ebp+8] = first arg (3)
; [ebp+12] = second arg (5)
; [ebp-4] = first local variable
mov esp, ebp     ; Destroy locals
pop ebp          ; Restore old frame
ret              ; Pop return address, jump back
```

**Why every process needs its own stack:**
Each process is at a different point in potentially different functions. Their local variables, return addresses — everything — lives on their stack. Switch ESP = switch the entire execution context.

---

# PART 2: BOOTING — FROM POWER TO KERNEL

## 2.1 BIOS/UEFI (Firmware)

When you press power:

```
1. PSU stabilizes voltages → sends POWER_GOOD signal
2. CPU starts at physical address 0xFFFFFFF0 (top of 4GB, mapped to BIOS ROM)
3. BIOS code runs:
   a. POST (Power-On Self-Test) — check RAM, detect hardware
   b. Initialize video (so you can see stuff)
   c. Enumerate PCI bus (find devices)
   d. Search for bootable media (check MBR signature 0xAA55)
   e. Load first 512 bytes of boot device to 0x7C00
   f. Jump to 0x7C00
```

At this point, CPU is in **Real Mode**:
- 16-bit
- Only 1MB addressable (20-bit address bus)
- No protection (any code can do anything)
- BIOS interrupt services available (int 0x10 = video, int 0x13 = disk)
- Segment:Offset addressing: physical = segment × 16 + offset

## 2.2 Bootloader (GRUB Stage 1 + 2)

**Stage 1** (MBR, 512 bytes): 
- Too small to do anything useful
- Its only job: load Stage 2 from disk

**Stage 2** (full GRUB):
- Has filesystem drivers (can read ext4, FAT)
- Shows you the boot menu
- Finds kernel file on disk
- Loads kernel ELF binary into RAM
- Switches CPU from Real Mode → Protected Mode:

```
; This is what GRUB does before jumping to us:
cli                          ; Disable interrupts
lgdt [temporary_gdt]         ; Load a basic GDT
mov eax, cr0
or eax, 1                   ; Set PE (Protected Enable) bit
mov cr0, eax                ; NOW IN PROTECTED MODE
jmp 0x08:protected_entry    ; Far jump to flush prefetch queue
```

Then GRUB:
- Sets EAX = 0x2BADB002 (multiboot magic)
- Sets EBX = pointer to multiboot info structure
- Jumps to our kernel's entry point (`_start`)

## 2.3 Our Boot Code

**File:** `asm/boot.asm`

```asm
; The Multiboot header — GRUB searches for this in the first 8KB.
; Without it, GRUB says "not a valid kernel."
section .multiboot
    dd 0x1BADB002              ; Magic: "I'm a multiboot kernel"
    dd 0x00000003              ; Flags: provide memory map + align modules
    dd -(0x1BADB002 + 0x00000003)  ; Checksum: all three must sum to 0

section .bss
align 16
stack_bottom:
    resb 16384                 ; Reserve 16KB for our kernel stack
stack_top:                     ; ESP will point here (stack grows DOWN)

section .text
global _start
extern kernel_main

_start:
    mov esp, stack_top         ; Set up stack (GRUB gives us no stack)
    push ebx                   ; Arg 2: multiboot info pointer
    push eax                   ; Arg 1: magic number
    call kernel_main           ; Enter C++!
.hang:
    cli
    hlt
    jmp .hang                  ; Never returns
```

**Key insight:** After `call kernel_main`, your C++ code is running. There's no OS beneath you. No standard library. No `printf`. No `malloc`. YOU are the OS now.

---

# PART 3: MEMORY SEGMENTATION

## 3.1 Why Segments Exist (Historical)

The 8086 (1978) had 16-bit registers but needed to address 1MB of memory. Solution: segment registers provide the high bits.

```
Physical address = Segment × 16 + Offset

Example:
  CS = 0x1000, IP = 0x0500
  Physical = 0x1000 × 16 + 0x0500 = 0x10000 + 0x0500 = 0x10500
```

In protected mode (80386+), segments became more complex — they're entries in the GDT with base address, limit, and permissions.

## 3.2 GDT In Detail

The GDT is an array in RAM. The CPU knows where it is via the GDTR register (loaded with `lgdt`).

```
GDT[0] = NULL descriptor (required — CPU uses this as "no segment")
GDT[1] = Kernel Code (selector 0x08)
GDT[2] = Kernel Data (selector 0x10)
GDT[3] = User Code   (selector 0x1B = 0x18 | ring3)
GDT[4] = User Data   (selector 0x23 = 0x20 | ring3)
```

**Selector format:**
```
Bits 15-3: Index into GDT (entry number)
Bit 2:     Table Indicator (0=GDT, 1=LDT)
Bits 1-0:  RPL (Requested Privilege Level)

0x08 = 0000 0000 0000 1|0|00 → Index 1, GDT, RPL 0
0x10 = 0000 0000 0001 0|0|00 → Index 2, GDT, RPL 0
0x1B = 0000 0000 0001 1|0|11 → Index 3, GDT, RPL 3
```

## 3.3 GDT Entry Encoding (8 bytes per entry)

```
struct GdtEntry {
    uint16_t limit_low;     // Bits 0-15 of limit
    uint16_t base_low;      // Bits 0-15 of base
    uint8_t  base_middle;   // Bits 16-23 of base
    uint8_t  access;        // Access byte (see below)
    uint8_t  granularity;   // Flags + bits 16-19 of limit
    uint8_t  base_high;     // Bits 24-31 of base
};
```

**Access byte:**
```
Bit 7: Present (1 = valid entry)
Bits 6-5: DPL (Descriptor Privilege Level: 0=kernel, 3=user)
Bit 4: Descriptor type (1=code/data, 0=system like TSS)
Bit 3: Executable (1=code segment, 0=data segment)
Bit 2: Direction/Conforming
  - Data: 0=grows up, 1=grows down
  - Code: 0=can only be called from same ring
          1=can be called from lower ring (conforming)
Bit 1: Readable/Writable
  - Code: 1=readable (can read constants from code segment)
  - Data: 1=writable
Bit 0: Accessed (CPU sets this when segment is used)
```

**Our kernel code segment (0x9A):**
```
1 0 0 1 1 0 1 0
│ │ │ │ │ │ │ └─ Accessed: 0 (CPU hasn't used it yet)
│ │ │ │ │ │ └── Readable: 1 (can read data from code segment)
│ │ │ │ │ └─── Conforming: 0 (ring 0 only)
│ │ │ │ └──── Executable: 1 (this IS code)
│ │ │ └───── Type: 1 (code/data, not system)
│ │ └────── DPL: 00 (ring 0 = kernel)
│ └─────── DPL continued
└──────── Present: 1 (valid)
```

**Our code** (`src/gdt.cpp`):
```cpp
// Kernel code: base=0, limit=0xFFFFFFFF, access=0x9A, granularity=0xCF
gdt_set_entry(1, 0, 0xFFFFFFFF, 0x9A, 0xCF);

// Kernel data: same but access=0x92 (not executable, writable)
gdt_set_entry(2, 0, 0xFFFFFFFF, 0x92, 0xCF);
```

**Granularity byte (0xCF):**
```
1 1 0 0 1 1 1 1
│ │ │ │ └┴┴┴── Limit bits 16-19 = 0xF (total limit = 0xFFFFF)
│ │ │ └─────── Available (OS can use this bit)
│ │ └──────── Long mode: 0 (not 64-bit)
│ └───────── Size: 1 (32-bit protected mode)
└────────── Granularity: 1 (limit in 4KB pages, so 0xFFFFF × 4KB = 4GB)
```

## 3.4 Why Flat Model?

Modern OSes don't USE segmentation for memory protection — they use PAGING instead. But the CPU requires a GDT in protected mode. So we set:
- Base = 0, Limit = 4GB for all segments
- This means: segmentation does NOTHING to addresses (offset passes through unchanged)
- All protection is done via paging (page tables)

This is called the "flat model" — every segment covers the entire address space.

---

# PART 4: INTERRUPTS — THE CPU's NERVOUS SYSTEM

## 4.1 Without Interrupts

Without interrupts, the CPU can only do one thing: execute instructions sequentially. To check if a key is pressed, it would have to:

```cpp
// POLLING — terrible!
while (true) {
    if (inb(0x64) & 1) {           // Is there data in keyboard buffer?
        char key = inb(0x60);       // Read it
        process_key(key);
    }
    // But while we're in this loop, we can't do ANYTHING else!
    // No multitasking, no timer, nothing.
}
```

## 4.2 How Interrupts Work (Hardware Level)

```
┌──────────┐      ┌─────────┐    INTR pin    ┌─────────┐
│ Keyboard │─IRQ1─│   PIC   │───────────────►│   CPU   │
│ Timer    │─IRQ0─│(8259A)  │                │         │
│ Disk     │─IRQ14│         │                │         │
└──────────┘      └─────────┘                └─────────┘
                                                   │
                                  "An interrupt!    │
                                   Which one?"      │
                                                   ▼
                                         CPU reads vector
                                         from PIC data bus
                                                   │
                                                   ▼
                                         CPU looks up IDT[vector]
                                                   │
                                                   ▼
                                         Jumps to handler address
```

**Detailed CPU steps when interrupt N fires:**

```
1. Finish currently executing instruction (interrupts are checked between instructions)

2. If interrupt flag (IF) in EFLAGS is clear (cli was used):
   - Hardware interrupts are IGNORED (except NMI)
   - CPU continues as normal

3. If IF is set (interrupts enabled):
   a. Push current EFLAGS onto stack (saves interrupt flag, etc.)
   b. Push current CS onto stack
   c. Push current EIP onto stack (return address)
   d. If there's a privilege change (ring 3→0):
      - Look up kernel stack in TSS (Task State Segment)
      - Switch to kernel stack (save old SS:ESP, load from TSS)
      - THEN push the old SS:ESP onto the NEW stack
   e. Clear IF flag (prevent nested interrupts)
   f. Clear TF flag (disable single-step)
   g. Look up IDT[N]
   h. Load new CS:EIP from IDT entry
   i. Begin executing handler
```

## 4.3 IDT Entry in Detail

```cpp
struct IdtEntry {
    uint16_t base_low;   // Handler address bits 0-15
    uint16_t selector;   // Code segment selector (0x08 for kernel)
    uint8_t  zero;       // Reserved, must be 0
    uint8_t  flags;      // Type + DPL + Present
    uint16_t base_high;  // Handler address bits 16-31
} __attribute__((packed));
```

**Flags byte:**
```
Bit 7: Present (1 = valid)
Bits 6-5: DPL (who can manually trigger this with 'int N' instruction)
  - 0 = only kernel can use 'int N'
  - 3 = user can use 'int N' (needed for syscall via int 0x80)
Bit 4: 0 (always)
Bits 3-0: Gate type
  - 0xE = 32-bit Interrupt Gate (clears IF — use for hardware IRQs)
  - 0xF = 32-bit Trap Gate (leaves IF set — use for exceptions/syscalls)
```

**Interrupt Gate vs Trap Gate:**
- Interrupt gate: CPU disables interrupts (clears IF) before entering handler. Use for hardware — prevents nested interrupts.
- Trap gate: CPU does NOT clear IF. Interrupts can still fire while handling the trap. Use for page faults, breakpoints, syscalls.

## 4.4 CPU Exceptions (ISRs 0-31) — Complete List

| # | Name | Type | Err? | Cause |
|---|------|------|------|-------|
| 0 | Divide Error | Fault | No | DIV/IDIV by zero, or quotient too large |
| 1 | Debug | Trap/Fault | No | Single-step (TF=1), hardware breakpoint |
| 2 | NMI | Interrupt | No | Non-Maskable Interrupt (hardware failure) |
| 3 | Breakpoint | Trap | No | INT 3 instruction (debugger uses this) |
| 4 | Overflow | Trap | No | INTO instruction with OF=1 |
| 5 | Bound Range | Fault | No | BOUND instruction — index out of range |
| 6 | Invalid Opcode | Fault | No | CPU can't decode instruction |
| 7 | Device Not Available | Fault | No | FPU instruction but no FPU (or TS=1) |
| 8 | Double Fault | Abort | Yes(0) | Exception during exception handler |
| 9 | (Reserved) | | No | Old coprocessor overrun |
| 10 | Invalid TSS | Fault | Yes | Bad TSS referenced during task switch |
| 11 | Segment Not Present | Fault | Yes | GDT entry has Present=0 |
| 12 | Stack-Segment Fault | Fault | Yes | Stack overflow, bad SS |
| 13 | General Protection | Fault | Yes | ANYTHING that violates protection rules |
| 14 | Page Fault | Fault | Yes | Access to unmapped/protected page |
| 15 | (Reserved) | | No | |
| 16 | x87 FPU Error | Fault | No | Unmasked FPU exception |
| 17 | Alignment Check | Fault | Yes | Unaligned memory access with AC=1 |
| 18 | Machine Check | Abort | No | Internal CPU error (hardware fault) |
| 19 | SIMD Exception | Fault | No | Unmasked SSE exception |

**Page Fault error code (most important — ISR 14):**
```
Bit 0 (P): 0 = page not present, 1 = protection violation
Bit 1 (W): 0 = read access, 1 = write access
Bit 2 (U): 0 = kernel mode, 1 = user mode
Bit 3 (R): 1 = reserved bit set in page table entry
Bit 4 (I): 1 = instruction fetch (NX violation)
```

CR2 register contains the faulting virtual address.

## 4.5 PIC Programming

The 8259A PIC handles prioritizing and routing hardware interrupts.

**Initialization sequence (ICW = Initialization Command Word):**

```cpp
// ICW1: Start initialization, expect ICW4
outb(0x20, 0x11);  // Master PIC command port
outb(0xA0, 0x11);  // Slave PIC command port

// ICW2: Vector offsets (where IRQs map to in IDT)
outb(0x21, 0x20);  // Master: IRQ0-7 → INT 32-39
outb(0xA1, 0x28);  // Slave: IRQ8-15 → INT 40-47

// ICW3: Tell PICs about each other
outb(0x21, 0x04);  // Master: slave is on IRQ2 (bit 2)
outb(0xA1, 0x02);  // Slave: my cascade identity is 2

// ICW4: 8086 mode (not 8080)
outb(0x21, 0x01);
outb(0xA1, 0x01);

// OCW1: Unmask all interrupts (0 = enabled)
outb(0x21, 0x00);  // Master mask: allow all
outb(0xA1, 0x00);  // Slave mask: allow all
```

**Masking (disabling) specific IRQs:**
```cpp
// Disable IRQ3 (COM2):
uint8_t mask = inb(0x21);
outb(0x21, mask | (1 << 3));

// Enable IRQ1 (keyboard):
mask = inb(0x21);
outb(0x21, mask & ~(1 << 1));
```

**End of Interrupt (EOI):** After handling an IRQ, you MUST tell the PIC you're done:
```cpp
// For IRQ 0-7 (master only):
outb(0x20, 0x20);

// For IRQ 8-15 (slave + master):
outb(0xA0, 0x20);  // EOI to slave
outb(0x20, 0x20);  // EOI to master (because cascade)
```

Without EOI, the PIC thinks you're still handling the interrupt and won't send more.

## 4.6 Our Complete Interrupt Flow

```
KEY PRESS → keyboard hardware → IRQ1 wire → PIC → CPU INTR pin

CPU:
  push EFLAGS
  push CS (0x08)
  push EIP (address of 'hlt' in our loop)
  clear IF
  load IDT[33] → jump to irq1 (in isr.asm)

isr.asm (irq1):
  push 0          ; dummy error code
  push 33         ; interrupt number
  jmp irq_common_stub

irq_common_stub:
  pusha           ; save EAX,ECX,EDX,EBX,ESP,EBP,ESI,EDI
  push ds         ; save data segment
  mov ds, 0x10    ; load kernel data segment
  push esp        ; arg: pointer to registers struct on stack
  call irq_handler  ; → C++ land

irq_handler (idt.cpp):
  irq = regs->int_no - 32;   // = 1
  irq_handlers[1](regs);     // → keyboard_callback

keyboard_callback (keyboard.cpp):
  scancode = inb(0x60);      // read from keyboard port
  ascii = lookup[scancode];   // convert
  vga_putchar(ascii);         // write to screen memory

irq_handler continues:
  outb(0x20, 0x20);          // EOI to PIC
  return to irq_common_stub

irq_common_stub:
  pop ds                     ; restore data segment
  popa                       ; restore all registers
  add esp, 8                 ; remove error code + int number
  iret                       ; CPU restores EIP, CS, EFLAGS
                             ; back to 'hlt' in kernel loop!
```

---

# PART 5: PHYSICAL MEMORY MANAGEMENT

## 5.1 The Memory Map Problem

After boot, you have RAM — but which parts are usable?

```
0x00000000 - 0x000003FF: Real Mode IVT (Interrupt Vector Table) — DON'T TOUCH
0x00000400 - 0x000004FF: BIOS Data Area — DON'T TOUCH
0x00000500 - 0x00007BFF: Free (but usually not worth using)
0x00007C00 - 0x00007DFF: MBR was loaded here — no longer needed
0x00007E00 - 0x0007FFFF: Free (~480 KB)
0x00080000 - 0x0009FFFF: Varies (EBDA usually here) — AVOID
0x000A0000 - 0x000BFFFF: VGA Memory (0xB8000 = text buffer) — MMIO
0x000C0000 - 0x000FFFFF: BIOS ROM, option ROMs — DON'T TOUCH
0x00100000 - onwards:     USABLE (this is where we live)
```

GRUB gives us the exact map via the multiboot info struct. Our `pmm_init()` reads this.

## 5.2 Why 4KB Pages?

The x86 page table structure requires memory to be managed in 4KB (4096 byte) chunks:
- Each page table entry maps exactly 4KB
- The page directory/table entries use bits 12-31 for the address (bottom 12 bits are flags)
- 12 bits of offset = 2^12 = 4096 bytes per page

So we allocate memory in 4KB units. Need 13 bytes? You get 4096. Wasteful for small allocations, but:
- Paging hardware requires it
- Simple to manage
- The kernel's slab allocator (like your P2 malloc) subdivides pages for small objects

## 5.3 Bitmap Allocator (Our Approach)

```
Each bit = one 4KB page

For 128MB RAM:
  128MB / 4KB = 32,768 pages
  32,768 bits = 4,096 bytes of bitmap (just 4KB!)

Bitmap:
  [11111111][11111111]...[00000000][00000000]...
   ^^^^^^^^                ^^^^^^^^
   Kernel/reserved          Free pages
```

**Our code (`src/pmm.cpp`):**
```cpp
static uint32_t bitmap[MAX_PAGES / 32];  // Each uint32_t holds 32 page bits

uint32_t pmm_alloc() {
    for (uint32_t i = 0; i < total_pages / 32; i++) {
        if (bitmap[i] == 0xFFFFFFFF) continue;  // All 32 bits set → skip

        for (uint32_t bit = 0; bit < 32; bit++) {
            if (!(bitmap[i] & (1 << bit))) {     // Found a 0 bit!
                bitmap[i] |= (1 << bit);          // Mark as used
                used_pages++;
                return (i * 32 + bit) * PAGE_SIZE; // Convert to physical address
            }
        }
    }
    return 0;  // Out of memory
}

void pmm_free(uint32_t addr) {
    uint32_t page = addr / PAGE_SIZE;
    bitmap[page / 32] &= ~(1 << (page % 32));  // Clear the bit
    used_pages--;
}
```

**Time complexity:** O(n) for alloc where n = total pages. Slow for millions of pages but fine for our kernel.

## 5.4 Linux's Buddy System (How Real OSes Do It)

```
Free lists organized by power-of-2 sizes:

Order 0: [4KB] [4KB] [4KB] [4KB] ...     (single pages)
Order 1: [8KB] [8KB] [8KB] ...            (2 contiguous pages)
Order 2: [16KB] [16KB] ...                (4 pages)
Order 3: [32KB] [32KB] ...
...
Order 10: [4MB]                            (1024 pages)
```

**Allocating 4KB (order 0):**
1. Check order-0 free list. If non-empty → take one. Done.
2. If empty → check order-1 list. Split an 8KB block into two 4KB blocks.
3. Give one to requestor, put other on order-0 free list.
4. If order-1 also empty → split order-2. And so on.

**Freeing:**
1. Free the page (put on order-0 list).
2. Check if its "buddy" (adjacent page) is also free.
3. If yes → merge them into order-1 block. Check THAT block's buddy. Repeat.

**Why it's better:** O(1) typical allocation. No fragmentation for aligned allocations.

## 5.5 Slab Allocator (On Top of Page Allocator)

The buddy system gives whole pages (4KB minimum). But kernel objects are often tiny:
- `struct task_struct` = ~6KB
- `struct inode` = 768 bytes
- `struct dentry` = 200 bytes

**Slab allocator:** Pre-allocates pages, divides them into fixed-size slots:

```
"inode" cache (each slab = 1 page = 4096 bytes):
┌────────────────────────────────────────────────────┐
│ [inode][inode][inode][inode][inode][metadata]       │
│  768B   768B   768B   768B   768B                  │
│   5 objects fit per 4KB page                       │
└────────────────────────────────────────────────────┘
```

`kmalloc(768)` → slab allocator finds a free slot instantly. No bitmap scanning.

This is essentially what your P2 malloc does — but the kernel version is specialized per object type.

---

# PART 6: VIRTUAL MEMORY (PAGING)

## 6.1 The Core Idea

**Without paging:** everyone shares one address space. Process A can read/write Process B's memory.

**With paging:** each process gets its own view of memory:

```
Process A thinks:           Process B thinks:
0x00400000: my code         0x00400000: my code  (different physical page!)
0x08000000: my heap         0x08000000: my heap  (different physical page!)
0xC0000000: kernel          0xC0000000: kernel   (SAME physical pages)
```

The MMU (Memory Management Unit) translates every virtual address to a physical address AUTOMATICALLY on every memory access. The CPU never sees physical addresses directly (after paging is enabled).

## 6.2 x86 Two-Level Page Tables

```
Virtual Address (32 bits):
┌────────────┬────────────┬────────────────┐
│  PD Index  │  PT Index  │    Offset      │
│  (10 bits) │  (10 bits) │   (12 bits)    │
└────────────┴────────────┴────────────────┘
      │             │              │
      │             │              └── Byte within 4KB page (0-4095)
      │             └── Which entry in the Page Table (0-1023)
      └── Which Page Table to use (0-1023)
```

**Translation:**
```
                 CR3 (Page Directory physical address)
                  │
                  ▼
         ┌─────────────────────┐
         │ Page Directory       │ (1024 entries × 4 bytes = 4KB)
         │ [0]: PT addr + flags│
         │ [1]: PT addr + flags│
         │ [2]: NOT PRESENT    │ ← access here = PAGE FAULT
         │ ...                 │
         │ [PD Index]: ────────┼──────┐
         │ ...                 │      │
         └─────────────────────┘      │
                                      ▼
                              ┌─────────────────────┐
                              │ Page Table N         │ (1024 entries × 4 bytes)
                              │ [0]: phys + flags   │
                              │ [1]: phys + flags   │
                              │ ...                 │
                              │ [PT Index]: ────────┼──┐
                              │ ...                 │  │
                              └─────────────────────┘  │
                                                       │
                                        Physical Page Address
                                                       │
                                                       ▼
                                    Final Physical Addr = Page + Offset
```

**Numbers:**
- 1024 PD entries × 1024 PT entries × 4KB page = 4GB addressable ✓
- One PD entry covers 4MB (1024 pages × 4KB)
- Total mapping structure size: 4KB (PD) + up to 4MB (all PTs) — but you only allocate PTs you need

## 6.3 Page Table Entry Format

```
┌──────────────────────────────────┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┬─┐
│  Physical Page Address [31:12]   │G│S│0│A│D│A│C│W│U│W│P│  │
│        (20 bits)                 │ │ │ │V│ │C│D│T│/│/│ │  │
│                                  │ │ │ │L│ │C│ │ │S│R│ │  │
└──────────────────────────────────┴─┴─┴─┴─┴─┴─┴─┴─┴─┴─┴─┴─┘
 31                              12 11 10 9  8  7  6  5  4  3  2  1  0
```

| Bit | Name | Meaning |
|-----|------|---------|
| 0 | Present | 1=page is mapped, 0=not mapped (accessing → page fault) |
| 1 | Read/Write | 1=writable, 0=read-only |
| 2 | User/Supervisor | 1=user-accessible, 0=kernel-only |
| 3 | Write-Through | 1=write-through caching |
| 4 | Cache Disable | 1=don't cache this page (use for MMIO) |
| 5 | Accessed | CPU sets to 1 when page is read |
| 6 | Dirty | CPU sets to 1 when page is written |
| 7 | Page Size | In PD: 1=4MB page (no page table needed) |
| 8 | Global | TLB entry survives CR3 change |
| 9-11 | Available | OS can use these bits for bookkeeping |

**Our code** (`src/paging.cpp`):
```cpp
void paging_map(uint32_t virt, uint32_t phys, uint32_t flags) {
    uint32_t pd_idx = virt >> 22;         // Top 10 bits
    uint32_t pt_idx = (virt >> 12) & 0x3FF; // Middle 10 bits

    // Ensure page table exists
    if (!(page_directory[pd_idx] & PAGE_PRESENT)) {
        uint32_t new_pt = pmm_alloc();     // Get a physical page for the PT
        page_directory[pd_idx] = new_pt | PAGE_PRESENT | PAGE_WRITABLE;
    }

    // Get the page table
    uint32_t* pt = (uint32_t*)(page_directory[pd_idx] & 0xFFFFF000);

    // Map the page
    pt[pt_idx] = (phys & 0xFFFFF000) | flags | PAGE_PRESENT;

    // Flush this TLB entry (CPU cached the old translation)
    asm volatile("invlpg (%0)" :: "r"(virt) : "memory");
}
```

## 6.4 Enabling Paging

```cpp
// 1. Set CR3 to point to page directory
asm volatile("mov %0, %%cr3" :: "r"(pd_physical_address));

// 2. Set PG bit (bit 31) in CR0
uint32_t cr0;
asm volatile("mov %%cr0, %0" : "=r"(cr0));
cr0 |= 0x80000000;
asm volatile("mov %0, %%cr0" :: "r"(cr0));

// NOW EVERY MEMORY ACCESS GOES THROUGH PAGE TABLES!
// If our page tables are wrong, we instantly triple-fault (reboot).
```

**Critical:** You MUST identity-map the kernel before enabling paging. Otherwise, the instruction AFTER enabling paging can't be fetched (the CPU is at physical address X, but paging would translate X → nothing → page fault → no handler mapped → double fault → triple fault → reboot).

## 6.5 TLB (Translation Lookaside Buffer)

Page table walks are EXPENSIVE (2-3 memory accesses per translation). The TLB caches recent translations:

```
Virtual Addr 0x00401234:
  1. Check TLB: "Do I already know what physical page this maps to?"
  2. TLB HIT: Yes! → Physical addr = cached_phys + 0x234. Done. (0 extra memory accesses)
  3. TLB MISS: Walk page tables (2 memory reads), cache the result.
```

**TLB invalidation matters:**
- Change a page table entry → must `invlpg` that address (otherwise CPU uses stale cached translation)
- Change CR3 (switch address space) → entire TLB flushed (all translations lost → cold start)
- This is why context switches are expensive!

## 6.6 Page Fault Handling (Key OS Mechanism)

Page faults are NOT just errors — they're a FEATURE. The OS uses them for:

**Demand Paging:**
```
Process starts → pages are marked NOT PRESENT (not loaded from disk yet)
Process accesses page → PAGE FAULT
Kernel: "Oh, this page exists on disk at offset X"
  → Load from disk into a free physical page
  → Update page table entry (now PRESENT)
  → Return to faulting instruction (retry succeeds)
```

**Copy-on-Write (COW):**
```
fork() → child gets SAME physical pages, marked READ-ONLY
Child (or parent) writes to page → PAGE FAULT (write to read-only)
Kernel: "This is a COW page"
  → Allocate new physical page
  → Copy content from old page
  → Map new page as WRITABLE for the writer
  → Old page stays read-only for the other process
  → Return (write succeeds on new page)
```

**Memory-Mapped Files:**
```
mmap(file) → pages marked NOT PRESENT
Access page → PAGE FAULT
Kernel: "This corresponds to file offset X"
  → Read that part of file into a page
  → Map page
  → Application can now read file data as if it were RAM
```

**Stack Growth:**
```
Stack overflow (write below current stack) → PAGE FAULT
Kernel: "Address is just below the stack — this is stack growth"
  → Allocate new page
  → Map it below current stack
  → Expand stack limit
  → Return (write succeeds)
```

## 6.7 Swapping (When RAM is Full)

```
pmm_alloc() returns 0 (out of physical pages)!

Kernel must EVICT a page:
  1. Find a "victim" page (page replacement algorithm)
  2. If dirty → write to swap partition on disk
  3. Mark victim's page table entry as NOT PRESENT, store disk location
  4. Now that physical page is free — use it for the new request
  5. If victim is accessed later → page fault → load back from swap

Page replacement algorithms:
  - FIFO: evict oldest page (bad — old page might be hot)
  - LRU: evict least recently used (ideal but expensive to track)
  - Clock/Second-Chance: approximate LRU using Accessed bit
  - Linux uses a two-list approach: active list + inactive list
```

---

# PART 7: PROCESSES

## 7.1 What is a Process?

A process is a running instance of a program. It consists of:

```
┌─────────────────────────────────────────────────┐
│ Process Control Block (PCB) / task_struct        │
├─────────────────────────────────────────────────┤
│ Identity:                                       │
│   PID, Parent PID, UID, GID                     │
│                                                 │
│ State:                                          │
│   RUNNING / READY / SLEEPING / STOPPED / ZOMBIE │
│                                                 │
│ CPU Context (saved when not running):           │
│   EIP, ESP, EBP, EAX..EDI, EFLAGS, CR3         │
│                                                 │
│ Memory:                                         │
│   Page directory pointer (CR3)                  │
│   Virtual memory areas (code, heap, stack, mmap)│
│                                                 │
│ Files:                                          │
│   File descriptor table [0]=stdin [1]=stdout... │
│                                                 │
│ Signals:                                        │
│   Pending signals, signal handlers              │
│                                                 │
│ Scheduling:                                     │
│   Priority, nice value, time slice remaining    │
│   CPU time used, wait channel                   │
└─────────────────────────────────────────────────┘
```

## 7.2 Process Memory Layout

```
Virtual address space (per-process view):
┌──────────────────────┐ 0xFFFFFFFF
│                      │
│   KERNEL SPACE       │ (mapped in all processes, not accessible from ring 3)
│   (shared)           │
│                      │
├──────────────────────┤ 0xC0000000 (typical Linux split)
│   Stack ↓            │ (grows downward)
│   [grows toward heap]│
│                      │
│   ...empty space...  │
│                      │
│   Memory-mapped      │ (shared libs, mmap'd files)
│   region             │
│                      │
│   ...empty space...  │
│                      │
│   Heap ↑             │ (grows upward via brk/sbrk — your P2!)
│   [grows toward stack]│
├──────────────────────┤
│   BSS (zeroed data)  │
│   Data (initialized) │
│   Text (code)        │ (read-only, executable)
├──────────────────────┤ 0x08048000 (typical ELF load address)
│   NULL page          │ (not mapped — dereference → segfault)
└──────────────────────┘ 0x00000000
```

## 7.3 Process Creation: fork() + exec()

**fork():**
```c
pid_t pid = fork();
// Now there are TWO processes running this same code!
// In parent: pid = child's PID
// In child:  pid = 0

if (pid == 0) {
    // I'm the child
    exec("/bin/ls");  // Replace myself with a new program
} else {
    // I'm the parent
    wait(&status);    // Wait for child to finish
}
```

**What fork() does inside the kernel:**
```
1. Allocate new PCB (task_struct)
2. Copy parent's PCB (same register values, same file descriptors)
3. Create new page directory — COPY-ON-WRITE: share parent's pages, mark all read-only
4. Assign new PID
5. Add to scheduler's ready queue
6. Return child's PID to parent, 0 to child
```

**What exec() does:**
```
1. Open the ELF file
2. Parse ELF headers (find code segment, data segment, entry point)
3. DESTROY current address space (free all current pages)
4. Create new address space:
   - Map code segment from ELF file (read-only, executable)
   - Map data segment (read-write)
   - Allocate stack
   - Set up argv, envp on stack
5. Set EIP to ELF entry point
6. Return to user mode (starts executing new program)
```

## 7.4 Process States — Detailed

```
              fork()
                │
                ▼
┌─────────┐ ──────────► ┌───────┐
│  NEW    │             │ READY │◄──────────────────────────┐
└─────────┘             └───┬───┘                           │
                            │ scheduler picks us            │
                            ▼                               │
                       ┌─────────┐    timer interrupt/      │
                       │ RUNNING │────preemption───────────►│
                       └────┬────┘                          │
                            │                               │
                 ┌──────────┼──────────┐                    │
                 │          │          │                    │
            exit()    wait for I/O   signal                 │
                 │          │          │                    │
                 ▼          ▼          ▼                    │
            ┌────────┐ ┌─────────┐ ┌────────┐             │
            │ ZOMBIE │ │SLEEPING │ │STOPPED │             │
            └────────┘ └────┬────┘ └────┬───┘             │
                            │           │                   │
                      I/O done/    SIGCONT                  │
                      wakeup          │                     │
                            │         │                     │
                            └─────────┴─────────────────────┘
```

**ZOMBIE:** Process has exit()'d but parent hasn't called wait() yet. The PCB still exists (parent needs to read the exit status). Once parent calls wait(), zombie is fully cleaned up.

**Why zombies matter:** If a parent never calls wait(), zombies accumulate. Each holds a PID and a PCB. Too many → can't create new processes. This is why orphaned processes get "re-parented" to init (PID 1).

## 7.5 Our Scheduler Implementation

```cpp
// Process struct
struct Process {
    uint32_t pid;
    ProcessState state;   // READY, RUNNING, DEAD
    uint32_t esp;         // Saved stack pointer
    uint32_t stack_base;  // Bottom of process's stack
    const char* name;
};

// Create a process: set up stack so context_switch can "resume" it
int process_create(void (*entry_point)(), const char* name) {
    // Set up stack as if the process was interrupted mid-execution:
    uint32_t* sp = (uint32_t*)(stack_base + STACK_SIZE);
    *(--sp) = (uint32_t)process_exit;  // If entry_point returns, go here
    *(--sp) = (uint32_t)entry_point;   // 'ret' in context_switch jumps HERE
    *(--sp) = 0;  // EBP
    *(--sp) = 0;  // EBX
    *(--sp) = 0;  // ESI
    *(--sp) = 0;  // EDI
    process.esp = (uint32_t)sp;
}
```

**The magic:** When `context_switch` pops EDI, ESI, EBX, EBP and then does `ret`, it pops the entry_point address into EIP. The CPU starts executing at entry_point as if it were returning from a function call. The process never knows it was created artificially.

---

# PART 8: SCHEDULING

## 8.1 Scheduling Goals

```
Competing goals — no algorithm is perfect for ALL:

Throughput:     Maximize jobs completed per unit time
Turnaround:    Minimize time from submission to completion
Response time: Minimize time from request to first response
Fairness:      Every process gets its fair share
Starvation:    No process waits forever
```

## 8.2 Algorithms

### FCFS (First Come, First Served)
```
Ready queue: [P1(24ms)] → [P2(3ms)] → [P3(3ms)]

Timeline:
P1 |████████████████████████|
P2                          |███|
P3                              |███|
0                          24  27  30

Average wait: (0 + 24 + 27) / 3 = 17ms  ← TERRIBLE (convoy effect)
```

### SJF (Shortest Job First)
```
Ready queue sorted by burst time: [P2(3ms)] → [P3(3ms)] → [P1(24ms)]

Timeline:
P2 |███|
P3     |███|
P1         |████████████████████████|
0    3    6                        30

Average wait: (0 + 3 + 6) / 3 = 3ms  ← OPTIMAL for average wait
Problem: How do you KNOW how long a job will take? (you can't — estimate from history)
Problem: Long jobs STARVE if short jobs keep arriving
```

### Round Robin (What we implement)
```
Time quantum = 4ms
Ready queue: [P1] → [P2] → [P3]

Timeline:
P1 |████|    |████|    |████|    |████|    |████|    |████|
P2      |████|    |
P3           |████|
0   4   8  12  13 16  20  24  28  30

Every process gets CPU time within 3 × quantum = 12ms
No starvation. Simple. Fair.

Tradeoff: If quantum too small → too many context switches (overhead)
          If quantum too large → degrades to FCFS
Typical: 10-100ms quantum
```

**Our code** (`src/kernel.cpp`):
```cpp
if (schedule_counter >= 2) {   // Every 2 ticks = 20ms quantum
    schedule_counter = 0;
    schedule();                 // Switch to next process
}
```

### MLFQ (Multi-Level Feedback Queue) — Used in Real OSes

```
Queue 0 (highest priority, shortest quantum: 8ms):
  → New processes start here
  → If process uses full quantum, demote to Queue 1

Queue 1 (medium priority, quantum: 16ms):
  → Processes that need more CPU
  → If uses full quantum, demote to Queue 2

Queue 2 (lowest priority, quantum: 32ms / FCFS):
  → CPU-bound processes
  → Gets CPU only when Queue 0 and 1 are empty

Rules:
  - Run highest-priority queue first
  - Within a queue: round-robin
  - If process blocks (I/O), it stays at current level
  - If process uses full time slice, demote one level
  - Periodically: boost ALL processes to Queue 0 (prevents starvation)
```

**Why it's smart:** Short interactive processes (typing, mouse) stay in Queue 0 (fast response). CPU-heavy processes (compilation, rendering) sink to Queue 2 (bulk throughput).

### Linux CFS (Completely Fair Scheduler)

```
Concept: Give every process its FAIR share of CPU time.

Each process has:
  vruntime: how much "virtual" time it has consumed

Scheduler always picks the process with the LOWEST vruntime.

vruntime increases as process runs:
  vruntime += actual_runtime × (default_weight / process_weight)

If process has nice=0 (default): weight = 1024
If process has nice=+10 (low priority): weight = 110
  → vruntime increases 9× faster → gets picked less often

Data structure: Red-black tree sorted by vruntime
  → Pick minimum = O(1) (leftmost node cached)
  → Insert after running = O(log n)

TIME SLICE: Not fixed! Calculated as:
  slice = scheduling_period × (process_weight / total_weight)
  
  With 4 equal processes: each gets 25% of 20ms period = 5ms
  With 1 high + 3 low: high gets ~50%, each low gets ~16%
```

---

# PART 9: CONTEXT SWITCHING — THE DEEPEST DETAIL

## 9.1 What Exactly Gets Switched

```
Process A → Process B switch:

SAVE (A's state):
  ┌────────────────────────────────────────┐
  │ General registers: EAX-EDI, EBP, ESP   │ → A's PCB
  │ Instruction pointer: EIP               │ → A's PCB
  │ Flags: EFLAGS                          │ → A's PCB
  │ Segment registers: CS, DS, SS...       │ → A's PCB (if different)
  │ FPU/SSE state: 512 bytes (fxsave)      │ → A's PCB (lazy save)
  │ Page directory: CR3                    │ → A's PCB
  └────────────────────────────────────────┘

LOAD (B's state):
  ┌────────────────────────────────────────┐
  │ ESP = B's saved stack pointer          │ ← from B's PCB
  │ Pop B's registers from B's stack       │
  │ CR3 = B's page directory               │ ← TLB FLUSH happens here!
  │ Restore FPU state if B was using FPU   │
  │ ret → pops B's saved EIP              │
  └────────────────────────────────────────┘
```

## 9.2 Our Context Switch (Annotated)

```asm
; context_switch(uint32_t* old_esp, uint32_t new_esp)
;
; Stack on entry:
;   [esp+8]: new_esp    (uint32_t — the value to load into ESP)
;   [esp+4]: old_esp    (uint32_t* — where to STORE current ESP)
;   [esp+0]: return address (where we came from in schedule())

context_switch:
    ; === SAVE CURRENT PROCESS ===
    push ebp            ; These are "callee-saved" in the C calling convention.
    push ebx            ; The C compiler assumes these are preserved across calls.
    push esi            ; By pushing them here, we save the current process's state.
    push edi

    ; Now save the current ESP into the outgoing process's struct
    mov eax, [esp + 20] ; eax = old_esp (pointer) — 4 pushes + return addr = 20 bytes above
    mov [eax], esp      ; *old_esp = current ESP (saves everything — the stack has all our state)

    ; === LOAD NEW PROCESS ===
    mov esp, [esp + 24] ; Load new process's ESP — WE ARE NOW ON A DIFFERENT STACK

    ; The new stack was set up (either by process_create or a previous context_switch)
    ; with EDI, ESI, EBX, EBP, and a return address on top.
    pop edi
    pop esi
    pop ebx
    pop ebp

    ret                 ; Pop return address → jump to new process's code
                        ; For a NEW process: this is entry_point()
                        ; For a RESUMED process: this is inside schedule() where it was interrupted
```

## 9.3 What Happens During a Timer-Driven Switch

```
Process A is running (in process_a function)
│
├─► PIT fires (10ms elapsed) → IRQ0 → INT 32
│   CPU pushes: EFLAGS, CS, EIP (of process_a)
│
├─► irq_common_stub: PUSHA (saves all of A's registers on A's KERNEL stack)
│   call irq_handler
│
├─► scheduler_timer_callback:
│     schedule_counter++
│     if >= 2:
│       schedule()
│         processes[A].state = READY
│         find next = B
│         processes[B].state = RUNNING
│         context_switch(&processes[A].esp, processes[B].esp)
│
│         ┌─── context_switch ───┐
│         │ push ebp,ebx,esi,edi │ (A's)
│         │ save ESP to A.esp    │
│         │ load ESP from B.esp  │ *** STACK SWITCH ***
│         │ pop edi,esi,ebx,ebp  │ (B's)
│         │ ret                  │ → returns into B's schedule() call
│         └──────────────────────┘
│
├─► Now we're in B's schedule() (where it was frozen last time)
│   schedule() returns
│   scheduler_timer_callback returns
│   irq_handler returns
│   irq_common_stub: POPA (restores B's registers from B's stack!)
│   IRET → CPU restores B's EIP, CS, EFLAGS
│
└─► Process B is now running! (from wherever it was interrupted last time)
```

---

# PART 10: CONCURRENCY & SYNCHRONIZATION

## 10.1 Why Concurrency is Hard

Even on a SINGLE CPU, preemptive interrupts cause races:

```cpp
// WRONG — race condition:
int count = 0;

void increment() {
    count++;  // Looks atomic, but it's actually 3 instructions:
}

// Assembly:
//   mov eax, [count]    ← Timer fires HERE (after read, before write)
//   add eax, 1          ← Another process also reads count=0
//   mov [count], eax    ← Both write count=1 instead of count=2!
```

## 10.2 Disabling Interrupts (Simplest Lock)

```cpp
void critical_section() {
    asm volatile("cli");    // Disable interrupts (no preemption possible)

    // Safe to modify shared data here — nobody can interrupt us
    shared_counter++;

    asm volatile("sti");    // Re-enable interrupts
}
```

**Pros:** Simple. Works.
**Cons:** 
- Only works on single-CPU (other cores still running)
- Delays ALL interrupt handling (keyboard, timer, disk)
- Never hold cli for too long!

## 10.3 Spinlocks (Multi-Core)

```cpp
// Atomic test-and-set using x86 XCHG instruction
void spin_lock(volatile int* lock) {
    while (__sync_lock_test_and_set(lock, 1)) {
        // Spin (busy-wait) until lock becomes 0
        // __sync_lock_test_and_set atomically:
        //   1. Reads current value of *lock
        //   2. Sets *lock = 1
        //   3. Returns the OLD value
        //   If old value was 0 → we got the lock (loop exits)
        //   If old value was 1 → someone else has it (keep spinning)
    }
}

void spin_unlock(volatile int* lock) {
    __sync_lock_release(lock);  // Atomic: *lock = 0 with memory barrier
}
```

**When to use:** Short critical sections (< 1μs). Holding a spinlock while sleeping = deadlock (the sleeping CPU can't release it, other CPUs spin forever).

## 10.4 Mutexes (Sleeping Locks)

```cpp
void mutex_lock(mutex_t* m) {
    if (atomic_try_lock(&m->locked)) {
        return;  // Got it immediately
    }
    // Couldn't get it — don't spin, SLEEP instead:
    add_to_wait_queue(&m->waiters, current_process);
    current_process->state = SLEEPING;
    schedule();  // Let someone else run
    // We'll be woken up when the lock is released
}

void mutex_unlock(mutex_t* m) {
    atomic_unlock(&m->locked);
    // Wake up one waiter:
    Process* waiter = remove_from_wait_queue(&m->waiters);
    if (waiter) {
        waiter->state = READY;  // Put back in run queue
    }
}
```

**Key difference from spinlock:** Instead of wasting CPU cycles spinning, the process goes to sleep. The scheduler runs other processes. When the lock is released, the sleeping process is woken up.

## 10.5 Semaphores

A generalization of mutex. Mutex = binary semaphore (0 or 1). Semaphore can have count > 1.

```cpp
// Allow up to N concurrent accesses:
void sem_wait(semaphore_t* s) {
    cli();
    while (s->count <= 0) {
        sleep_on(&s->waiters);  // Sleep if no permits available
    }
    s->count--;
    sti();
}

void sem_signal(semaphore_t* s) {
    cli();
    s->count++;
    wake_one(&s->waiters);  // Wake one sleeping process
    sti();
}
```

**Use case:** Pool of N database connections. Semaphore starts at N. Each `wait()` takes a connection (decrement). Each `signal()` returns one (increment). If all taken, processes sleep.

## 10.6 Deadlock

```
Process A holds Lock 1, wants Lock 2
Process B holds Lock 2, wants Lock 1

A: lock(1) ✓  lock(2) → BLOCKED (B has it)
B: lock(2) ✓  lock(1) → BLOCKED (A has it)

Both wait forever. Neither can proceed.
```

**Four conditions (ALL must hold for deadlock):**
1. Mutual exclusion (only one can hold the lock)
2. Hold and wait (hold one lock while waiting for another)
3. No preemption (can't forcibly take a lock away)
4. Circular wait (A waits for B, B waits for A)

**Prevention:** Break any one condition. Easiest: always acquire locks in the same order.

---

# PART 11: FILESYSTEMS

## 11.1 What a Disk Looks Like

```
Disk = array of sectors (512 bytes each, or 4KB for modern NVMe)

Sector 0: [master boot record]
Sector 1: [...]
Sector 2: [...]
...
Sector N: [...]

That's it. Just numbered blocks. The filesystem gives these blocks MEANING.
```

## 11.2 How ext4 (Linux) Organizes a Disk

```
┌────────────┬────────────┬────────────┬──────────────────────┐
│ Boot Block │ Superblock │ Block Group│ Block Group 1...     │
│ (1KB)      │ (1KB)      │ Descriptor │                      │
│            │            │ Table      │                      │
└────────────┴────────────┴────────────┴──────────────────────┘

Each Block Group:
┌──────────┬──────────┬──────────┬──────────┬───────────────┐
│ Block    │ Inode    │ Inode    │ Data     │ Data          │
│ Bitmap   │ Bitmap   │ Table    │ Blocks   │ Blocks...     │
│ (1 block)│ (1 block)│(N blocks)│          │               │
└──────────┴──────────┴──────────┴──────────┴───────────────┘
```

## 11.3 Inodes (The Core Concept)

Every file/directory has exactly ONE inode:

```
struct inode {
    mode:    file type + permissions (rwxrwxrwx)
    uid:     owner user ID
    gid:     owner group ID
    size:    file size in bytes
    atime:   last access time
    mtime:   last modification time
    ctime:   last status change time
    nlinks:  number of hard links (directory entries pointing here)

    // WHERE THE DATA IS:
    direct[12]:     12 direct block pointers (first 48KB of file)
    indirect:       pointer to block of 1024 pointers (next 4MB)
    double_indirect: pointer to block of pointers to blocks of pointers (next 4GB)
    triple_indirect: for truly massive files
};
```

**Directory entries are just files containing:**
```
┌─────────────────────────────────────┐
│ inode: 42    name: "hello.c"        │
│ inode: 43    name: "Makefile"       │
│ inode: 44    name: "build"          │  (this is a subdirectory)
│ inode: 2     name: ".."             │  (parent directory)
│ inode: 41    name: "."              │  (self)
└─────────────────────────────────────┘
```

**Path resolution:** `/home/vikram/hello.c`
```
1. Inode 2 (root "/") → read directory entries → find "home" → inode 100
2. Inode 100 ("/home") → read directory entries → find "vikram" → inode 200
3. Inode 200 ("/home/vikram") → read entries → find "hello.c" → inode 42
4. Inode 42 has the file's block pointers → read data blocks
```

## 11.4 VFS (Virtual File System)

Linux supports dozens of filesystems (ext4, XFS, btrfs, FAT, NFS, etc.) through VFS:

```
User: read(fd, buf, 100)
        │
        ▼
┌─────────────────────────────────────────────┐
│ VFS Layer (filesystem-agnostic)             │
│   struct inode_operations { .read, .write } │
│   struct file_operations { .read, .write }  │
└────────────────────┬────────────────────────┘
                     │ calls filesystem-specific operation
         ┌───────────┼───────────┐
         ▼           ▼           ▼
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │  ext4   │ │  XFS    │ │ btrfs   │
    │ .read() │ │ .read() │ │ .read() │
    └────┬────┘ └────┬────┘ └────┬────┘
         │           │           │
         ▼           ▼           ▼
    Block device layer (talks to disk driver)
```

## 11.5 Journaling (Crash Safety)

**Problem:** Writing to disk is multi-step. If power fails mid-write, filesystem is corrupted.

**Solution (ext4 journaling):**
```
1. Write the changes to a JOURNAL (separate area on disk)
   "I'm about to: update inode 42, write block 500, update directory entry"
2. COMMIT the journal entry (write commit record)
3. Actually perform the writes (update inode, block, directory)
4. Mark journal entry as complete

If crash during step 1-2: journal entry is incomplete → ignored on recovery
If crash during step 3: journal entry is complete → REPLAY it on recovery
If crash during step 4: already done → no problem
```

## 11.6 Our Ramdisk (Simplest Possible FS)

```
Our ramdisk format:
┌──────────────────────────────────────┐
│ Header: magic=0xDEADBEEF, num_files  │  8 bytes
├──────────────────────────────────────┤
│ Entry 0: name[32], offset, size      │  40 bytes each
│ Entry 1: name[32], offset, size      │
│ ...                                  │
├──────────────────────────────────────┤
│ File data (at specified offsets)      │
└──────────────────────────────────────┘
```

No directories, no permissions, no journaling. Just: name → data. But it teaches the concept of a filesystem: metadata (where things are) + data (the actual bytes).

---

# PART 12: SYSTEM CALLS

## 12.1 The Mechanism

```
User code (ring 3):
    mov eax, 1        ; syscall number (1 = sys_exit)
    mov ebx, 0        ; first argument (exit code = 0)
    int 0x80          ; trigger syscall interrupt
                      ; (modern: syscall instruction instead)
    │
    │ CPU: privilege transition ring 3 → ring 0
    │ CPU: loads kernel stack from TSS
    │ CPU: pushes user's SS, ESP, EFLAGS, CS, EIP
    │
    ▼
Kernel (ring 0):
    IDT[0x80] → system_call_handler:
        save all registers
        validate syscall number (EAX < NR_SYSCALLS?)
        call sys_call_table[eax](ebx, ecx, edx, esi, edi)
            → sys_exit(0): terminate process, free resources
        put return value in EAX
        restore registers
        iret → back to user mode
```

## 12.2 Linux System Call Table (Key Ones)

| # | Name | Signature | What it does |
|---|------|-----------|--------------|
| 1 | exit | exit(int status) | Terminate process |
| 2 | fork | fork() → pid | Create child process |
| 3 | read | read(fd, buf, count) → bytes | Read from file descriptor |
| 4 | write | write(fd, buf, count) → bytes | Write to file descriptor |
| 5 | open | open(path, flags) → fd | Open file, return descriptor |
| 6 | close | close(fd) | Release file descriptor |
| 7 | waitpid | waitpid(pid, status, opts) | Wait for child |
| 11 | execve | execve(path, argv, envp) | Replace process image |
| 12 | chdir | chdir(path) | Change working directory |
| 20 | getpid | getpid() → pid | Get own PID |
| 33 | access | access(path, mode) → 0/-1 | Check file permissions |
| 37 | kill | kill(pid, sig) | Send signal to process |
| 45 | brk | brk(addr) → new_brk | Set program break (heap) |
| 63 | dup2 | dup2(old, new) → new | Duplicate file descriptor |
| 90 | mmap | mmap(addr,len,prot,...) → ptr | Map memory/files |

**Your P1 shell used:** fork(2), execve(11), waitpid(7), pipe(42), dup2(63), read(3), write(4)

**Your P2 malloc used:** brk(45)/sbrk, mmap(90), munmap(91)

## 12.3 How `write(1, "hello", 5)` Flows Through Linux

```
1. User calls write(1, "hello", 5)
   → glibc: puts args in registers, executes 'syscall' instruction

2. CPU switches to kernel mode, jumps to entry_SYSCALL_64

3. Kernel: look up sys_call_table[1] → ksys_write(fd=1, buf="hello", count=5)

4. ksys_write:
   → Find struct file for fd=1 (stdout)
   → Check user buffer is valid (access_ok): is buf in user's address space?
   → Call file->f_op->write() (VFS dispatch)

5. For a terminal: tty_write()
   → Copy data from user buffer to kernel buffer (copy_from_user)
   → Send to terminal driver
   → Terminal driver pushes to display

6. Return bytes written (5) in RAX

7. sysret instruction: back to user mode, user gets 5 as return value
```

---

# PART 13: SIGNALS

## 13.1 What Signals Are

Signals are software interrupts for processes — the kernel's way of telling a process "something happened."

```
Common signals:
SIGINT  (2):  Ctrl+C pressed → interrupt process
SIGKILL (9):  Kill immediately (can't be caught)
SIGSEGV (11): Segmentation fault (bad memory access)
SIGALRM (14): Timer alarm expired
SIGTERM (15): Polite "please die" (can be caught)
SIGCHLD (17): Child process terminated
SIGSTOP (19): Pause process (can't be caught)
SIGCONT (18): Resume paused process
SIGPIPE (13): Write to a broken pipe
```

## 13.2 How Signal Delivery Works

```
Process A is running in user mode
│
Kernel decides to deliver SIGINT to A (e.g., user pressed Ctrl+C):
│
├─► Next time A returns to user mode (after syscall or interrupt):
│   Kernel checks: "Does A have pending signals?"
│   Yes → SIGINT is pending
│
├─► Does A have a handler for SIGINT?
│   ├─► SIG_DFL (default): terminate process
│   ├─► SIG_IGN: discard signal, continue normally
│   └─► Custom handler (void handler(int sig)):
│       1. Save A's current user context (registers, EIP)
│       2. Modify A's user stack: push signal frame
│       3. Set A's EIP = handler address
│       4. Return to user mode → handler executes
│       5. Handler calls sigreturn() → kernel restores original context
│       6. A continues from where it was interrupted
```

This is what you fixed in P1 shell (SIGINT/SIGTSTP handling)!

---

# PART 14: INTER-PROCESS COMMUNICATION (IPC)

## 14.1 Mechanisms

```
┌──────────────┬──────────────────────────────────────────────────┐
│ Mechanism    │ How it works                                      │
├──────────────┼──────────────────────────────────────────────────┤
│ Pipe         │ Unidirectional byte stream between related procs  │
│ Named Pipe   │ Same but has a name on filesystem (FIFO)          │
│ Signal       │ Async notification (just a number, no data)       │
│ Shared Mem   │ Same physical pages mapped in two processes       │
│ Message Queue│ Kernel-managed queue of fixed-size messages       │
│ Socket       │ Bidirectional, works across network (TCP/UDP)     │
│ Unix Socket  │ Like TCP but local (same machine, fast)           │
│ eventfd      │ Simple counter-based notification                 │
│ mmap'd file  │ File mapped in memory, shared between processes   │
└──────────────┴──────────────────────────────────────────────────┘
```

## 14.2 Pipes (You Built This in P1!)

```
pipe(fds):
  Kernel creates a 4KB buffer in kernel memory
  fds[0] = read end (file descriptor)
  fds[1] = write end (file descriptor)

fork() → child inherits both fds

Parent writes: write(fds[1], "data", 4)
  → Kernel copies into pipe buffer

Child reads: read(fds[0], buf, 4)
  → Kernel copies from pipe buffer to user buf

If pipe is full: write BLOCKS (process goes to sleep)
If pipe is empty: read BLOCKS (process goes to sleep)
If all writers close: read returns 0 (EOF)
If all readers close: write gets SIGPIPE
```

---

# PART 15: HOW IT ALL CONNECTS (The Big Picture)

When you type `ls` in a terminal:

```
1. KEYBOARD HARDWARE
   Key press → scancode → IRQ1 → PIC → CPU → keyboard driver
   Driver: scancode → ASCII → put in terminal input buffer

2. TERMINAL DRIVER
   Input buffer accumulates characters
   On Enter: wake up the process reading from this terminal

3. SHELL (bash) — your P1!
   read() syscall returns the line "ls\n"
   Shell parses: command = "ls", no pipes, no redirects
   fork() → child process created (copy-on-write pages)

4. FORK (inside kernel)
   Allocate new task_struct
   Copy parent's page directory (COW)
   Copy file descriptor table
   Add to scheduler
   Return 0 to child, child's PID to parent

5. EXEC (child calls execve("/bin/ls"))
   Kernel: open /bin/ls ELF file
   Parse ELF: find text, data, entry point
   Destroy old address space
   Create new: map code, data, stack
   Set EIP = ELF entry point
   Return to user mode → ls starts running

6. LS RUNNING
   ls calls opendir("/current/directory")
   → open() syscall → kernel VFS → ext4: find inode → read directory entries
   ls calls stat() on each file → get metadata
   ls formats output
   ls calls write(1, buffer, len) → stdout → terminal → display

7. LS EXITS
   exit(0) syscall → kernel frees pages, closes fds
   Process becomes ZOMBIE
   Kernel sends SIGCHLD to parent (bash)
   bash calls waitpid() → reaps zombie → gets exit status
   bash prints next prompt "$ "
```

Every single step here involves concepts from this document:
interrupts, syscalls, scheduling, paging, file descriptors, signals, fork/exec.

---

# SUMMARY: What "God Level OS" Means

```
Level 0: "I use Linux"
Level 1: "I know what syscalls are"
Level 2: "I built a kernel" ← YOU ARE HERE
Level 3: "I understand Linux kernel source"
Level 4: "I can design an OS for novel hardware/requirements"
Level 5: "I can prove correctness properties of OS mechanisms"

Your path: → OSTEP book → xv6 labs → Linux kernel hacking → research papers
```

You have the code (P3). Now these notes give you the theory. Read one section per day, then look at the corresponding source file in your kernel. The understanding will compound.
