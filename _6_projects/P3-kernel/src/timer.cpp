#include "timer.h"
#include "idt.h"
#include "io.h"
#include "vga.h"

//
// timer.cpp — PIT (Programmable Interval Timer) at IRQ0
//
// The PIT runs at a base frequency of 1.193182 MHz.
// We set a divisor to get our desired tick rate.
//
// At 100 Hz → interrupt every 10ms (good for scheduling)
//

static uint32_t tick_count = 0;

static void timer_callback(Registers* /*regs*/) {
    tick_count++;
}

void timer_init(uint32_t frequency_hz) {
    register_irq_handler(0, timer_callback);

    // Calculate the divisor
    uint32_t divisor = 1193182 / frequency_hz;

    // Send command byte: channel 0, lobyte/hibyte, rate generator
    outb(0x43, 0x36);

    // Send divisor (low byte first, then high byte)
    outb(0x40, divisor & 0xFF);
    outb(0x40, (divisor >> 8) & 0xFF);
}

uint32_t timer_get_ticks() {
    return tick_count;
}
