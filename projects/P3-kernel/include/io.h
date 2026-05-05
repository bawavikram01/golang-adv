#pragma once
#include <stdint.h>

//
// io.h — x86 port I/O (inline assembly)
//
// These are how you talk to hardware on x86:
//   outb(port, data) — write a byte to a hardware port
//   inb(port)        — read a byte from a hardware port
//
// Examples:
//   outb(0x20, 0x20)  — send EOI to PIC
//   inb(0x60)         — read keyboard scancode
//

// Write a byte to an I/O port
static inline void outb(uint16_t port, uint8_t data) {
    asm volatile("outb %0, %1" : : "a"(data), "Nd"(port));
}

// Read a byte from an I/O port
static inline uint8_t inb(uint16_t port) {
    uint8_t result;
    asm volatile("inb %1, %0" : "=a"(result) : "Nd"(port));
    return result;
}

// Wait a tiny bit (for slow devices)
static inline void io_wait() {
    outb(0x80, 0);  // Port 0x80 is used for POST codes — safe to write junk
}
