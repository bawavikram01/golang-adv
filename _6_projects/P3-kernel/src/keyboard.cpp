#include "keyboard.h"
#include "idt.h"
#include "vga.h"
#include "io.h"

//
// keyboard.cpp — PS/2 keyboard driver
//
// IRQ1 fires on every keypress/release. We read the scancode from port 0x60.
// Scancode Set 1 (default): key press = code, key release = code | 0x80
//

// US QWERTY scancode-to-ASCII table (Set 1, lowercase only for now)
static const char scancode_to_ascii[128] = {
    0,  27, '1','2','3','4','5','6','7','8','9','0','-','=','\b',  // 0x00-0x0E
    '\t','q','w','e','r','t','y','u','i','o','p','[',']','\n',     // 0x0F-0x1C
    0,  'a','s','d','f','g','h','j','k','l',';','\'','`',          // 0x1D-0x29
    0,  '\\','z','x','c','v','b','n','m',',','.','/',0,            // 0x2A-0x35
    '*', 0, ' ',                                                    // 0x36-0x39
    // Rest is function keys, etc. — ignore for now
};

static void keyboard_callback(Registers* /*regs*/) {
    uint8_t scancode = inb(0x60);

    // Ignore key releases (bit 7 set)
    if (scancode & 0x80) return;

    // Convert to ASCII
    if (scancode < 128) {
        char c = scancode_to_ascii[scancode];
        if (c) {
            vga_putchar(c);
        }
    }
}

void keyboard_init() {
    register_irq_handler(1, keyboard_callback);
}
