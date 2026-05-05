#include "idt.h"
#include "vga.h"
#include "io.h"

//
// idt.cpp — IDT setup + interrupt handlers
//

static IdtEntry idt[256];
static IdtPointer idt_ptr;
static IrqHandler irq_handlers[16] = {};

// ISR stubs from isr.asm
extern "C" {
    void isr0();  void isr1();  void isr2();  void isr3();
    void isr4();  void isr5();  void isr6();  void isr7();
    void isr8();  void isr9();  void isr10(); void isr11();
    void isr12(); void isr13(); void isr14(); void isr15();
    void isr16(); void isr17(); void isr18(); void isr19();
    void isr20(); void isr21(); void isr22(); void isr23();
    void isr24(); void isr25(); void isr26(); void isr27();
    void isr28(); void isr29(); void isr30(); void isr31();

    void irq0();  void irq1();  void irq2();  void irq3();
    void irq4();  void irq5();  void irq6();  void irq7();
    void irq8();  void irq9();  void irq10(); void irq11();
    void irq12(); void irq13(); void irq14(); void irq15();
}

void idt_set_gate(uint8_t num, uint32_t base, uint16_t selector, uint8_t flags) {
    idt[num].base_low  = base & 0xFFFF;
    idt[num].base_high = (base >> 16) & 0xFFFF;
    idt[num].selector  = selector;
    idt[num].zero      = 0;
    idt[num].flags     = flags;
}

// Remap the PIC (Programmable Interrupt Controller)
// By default, IRQs 0-7 map to interrupts 8-15, which conflicts with CPU exceptions.
// We remap them to 32-47.
static void pic_remap() {
    // Save masks
    uint8_t mask1 = inb(0x21);
    uint8_t mask2 = inb(0xA1);

    // Start initialization sequence (ICW1)
    outb(0x20, 0x11); io_wait();
    outb(0xA0, 0x11); io_wait();

    // ICW2: Set vector offsets
    outb(0x21, 0x20); io_wait();  // Master PIC: IRQs start at 32
    outb(0xA1, 0x28); io_wait();  // Slave PIC: IRQs start at 40

    // ICW3: Tell Master about Slave on IRQ2
    outb(0x21, 0x04); io_wait();
    outb(0xA1, 0x02); io_wait();

    // ICW4: 8086 mode
    outb(0x21, 0x01); io_wait();
    outb(0xA1, 0x01); io_wait();

    // Restore masks
    outb(0x21, mask1);
    outb(0xA1, mask2);
}

void idt_init() {
    idt_ptr.limit = sizeof(idt) - 1;
    idt_ptr.base = (uint32_t)&idt;

    // Zero out all entries
    for (int i = 0; i < 256; i++) {
        idt_set_gate(i, 0, 0, 0);
    }

    // CPU exceptions (ISRs 0-31)
    // Flags: 0x8E = present(1) | ring0(00) | size32(1) | type_interrupt(110)
    idt_set_gate(0,  (uint32_t)isr0,  0x08, 0x8E);
    idt_set_gate(1,  (uint32_t)isr1,  0x08, 0x8E);
    idt_set_gate(2,  (uint32_t)isr2,  0x08, 0x8E);
    idt_set_gate(3,  (uint32_t)isr3,  0x08, 0x8E);
    idt_set_gate(4,  (uint32_t)isr4,  0x08, 0x8E);
    idt_set_gate(5,  (uint32_t)isr5,  0x08, 0x8E);
    idt_set_gate(6,  (uint32_t)isr6,  0x08, 0x8E);
    idt_set_gate(7,  (uint32_t)isr7,  0x08, 0x8E);
    idt_set_gate(8,  (uint32_t)isr8,  0x08, 0x8E);
    idt_set_gate(9,  (uint32_t)isr9,  0x08, 0x8E);
    idt_set_gate(10, (uint32_t)isr10, 0x08, 0x8E);
    idt_set_gate(11, (uint32_t)isr11, 0x08, 0x8E);
    idt_set_gate(12, (uint32_t)isr12, 0x08, 0x8E);
    idt_set_gate(13, (uint32_t)isr13, 0x08, 0x8E);
    idt_set_gate(14, (uint32_t)isr14, 0x08, 0x8E);
    idt_set_gate(15, (uint32_t)isr15, 0x08, 0x8E);
    idt_set_gate(16, (uint32_t)isr16, 0x08, 0x8E);
    idt_set_gate(17, (uint32_t)isr17, 0x08, 0x8E);
    idt_set_gate(18, (uint32_t)isr18, 0x08, 0x8E);
    idt_set_gate(19, (uint32_t)isr19, 0x08, 0x8E);
    idt_set_gate(20, (uint32_t)isr20, 0x08, 0x8E);
    idt_set_gate(21, (uint32_t)isr21, 0x08, 0x8E);
    idt_set_gate(22, (uint32_t)isr22, 0x08, 0x8E);
    idt_set_gate(23, (uint32_t)isr23, 0x08, 0x8E);
    idt_set_gate(24, (uint32_t)isr24, 0x08, 0x8E);
    idt_set_gate(25, (uint32_t)isr25, 0x08, 0x8E);
    idt_set_gate(26, (uint32_t)isr26, 0x08, 0x8E);
    idt_set_gate(27, (uint32_t)isr27, 0x08, 0x8E);
    idt_set_gate(28, (uint32_t)isr28, 0x08, 0x8E);
    idt_set_gate(29, (uint32_t)isr29, 0x08, 0x8E);
    idt_set_gate(30, (uint32_t)isr30, 0x08, 0x8E);
    idt_set_gate(31, (uint32_t)isr31, 0x08, 0x8E);

    // Remap PIC so IRQs don't conflict with CPU exceptions
    pic_remap();

    // Hardware IRQs (32-47)
    idt_set_gate(32, (uint32_t)irq0,  0x08, 0x8E);
    idt_set_gate(33, (uint32_t)irq1,  0x08, 0x8E);
    idt_set_gate(34, (uint32_t)irq2,  0x08, 0x8E);
    idt_set_gate(35, (uint32_t)irq3,  0x08, 0x8E);
    idt_set_gate(36, (uint32_t)irq4,  0x08, 0x8E);
    idt_set_gate(37, (uint32_t)irq5,  0x08, 0x8E);
    idt_set_gate(38, (uint32_t)irq6,  0x08, 0x8E);
    idt_set_gate(39, (uint32_t)irq7,  0x08, 0x8E);
    idt_set_gate(40, (uint32_t)irq8,  0x08, 0x8E);
    idt_set_gate(41, (uint32_t)irq9,  0x08, 0x8E);
    idt_set_gate(42, (uint32_t)irq10, 0x08, 0x8E);
    idt_set_gate(43, (uint32_t)irq11, 0x08, 0x8E);
    idt_set_gate(44, (uint32_t)irq12, 0x08, 0x8E);
    idt_set_gate(45, (uint32_t)irq13, 0x08, 0x8E);
    idt_set_gate(46, (uint32_t)irq14, 0x08, 0x8E);
    idt_set_gate(47, (uint32_t)irq15, 0x08, 0x8E);

    // Load IDT
    asm volatile("lidt %0" : : "m"(idt_ptr));
}

void register_irq_handler(uint8_t irq, IrqHandler handler) {
    irq_handlers[irq] = handler;
}

// Called from isr_common_stub (assembly)
extern "C" void isr_handler(Registers* regs) {
    vga_set_color(VGA_LIGHT_RED, VGA_BLACK);
    vga_print("\n!!! CPU Exception: ");
    vga_print_dec(regs->int_no);
    vga_print(" err=");
    vga_print_hex(regs->err_code);
    vga_print(" at EIP=");
    vga_print_hex(regs->eip);
    vga_print("\n");
    vga_set_color(VGA_LIGHT_GREY, VGA_BLACK);

    // Page fault — print the faulting address (CR2)
    if (regs->int_no == 14) {
        uint32_t cr2;
        asm volatile("mov %%cr2, %0" : "=r"(cr2));
        vga_print("  Page fault at address: ");
        vga_print_hex(cr2);
        vga_print("\n");
    }

    // Halt on unrecoverable exceptions
    if (regs->int_no < 32) {
        vga_print("  SYSTEM HALTED\n");
        asm volatile("cli; hlt");
    }
}

// Called from irq_common_stub (assembly)
extern "C" void irq_handler(Registers* regs) {
    // Dispatch to registered handler
    uint8_t irq = regs->int_no - 32;
    if (irq < 16 && irq_handlers[irq]) {
        irq_handlers[irq](regs);
    }

    // Send End-of-Interrupt to PIC
    if (regs->int_no >= 40) {
        outb(0xA0, 0x20);  // EOI to slave PIC
    }
    outb(0x20, 0x20);      // EOI to master PIC
}
