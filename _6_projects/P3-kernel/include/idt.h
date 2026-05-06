#pragma once
#include <stdint.h>

//
// idt.h — Interrupt Descriptor Table
//
// The IDT maps interrupt numbers (0-255) to handler functions.
// When interrupt N fires, CPU looks up IDT[N] and jumps there.
//

// Register state pushed by our ISR stubs (matches the push order in isr.asm)
struct Registers {
    uint32_t ds;                                     // Data segment
    uint32_t edi, esi, ebp, esp, ebx, edx, ecx, eax; // pusha
    uint32_t int_no, err_code;                       // Pushed by stub
    uint32_t eip, cs, eflags, useresp, ss;           // Pushed by CPU
} __attribute__((packed));

struct IdtEntry {
    uint16_t base_low;    // Low 16 bits of handler address
    uint16_t selector;    // Kernel code segment selector (0x08)
    uint8_t  zero;        // Always 0
    uint8_t  flags;       // Type + DPL + Present
    uint16_t base_high;   // High 16 bits of handler address
} __attribute__((packed));

struct IdtPointer {
    uint16_t limit;       // Size of IDT - 1
    uint32_t base;        // Address of first entry
} __attribute__((packed));

void idt_init();
void idt_set_gate(uint8_t num, uint32_t base, uint16_t selector, uint8_t flags);

// Register an IRQ handler callback
using IrqHandler = void (*)(Registers*);
void register_irq_handler(uint8_t irq, IrqHandler handler);
