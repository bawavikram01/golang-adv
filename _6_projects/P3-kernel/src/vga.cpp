#include "vga.h"
#include "io.h"

//
// vga.cpp — VGA text mode (80x25)
//
// The VGA buffer is at 0xB8000. Each cell is 2 bytes:
//   Byte 0: ASCII character
//   Byte 1: Color (bg << 4 | fg)
//

static const int VGA_WIDTH = 80;
static const int VGA_HEIGHT = 25;
static uint16_t* const VGA_BUFFER = (uint16_t*)0xB8000;

static int cursor_x = 0;
static int cursor_y = 0;
static uint8_t current_color = (VGA_BLACK << 4) | VGA_LIGHT_GREY;

static inline uint16_t vga_entry(char c, uint8_t color) {
    return (uint16_t)c | ((uint16_t)color << 8);
}

// Move the hardware cursor (blinking underscore)
static void update_cursor() {
    uint16_t pos = cursor_y * VGA_WIDTH + cursor_x;
    outb(0x3D4, 14);           // High byte
    outb(0x3D5, pos >> 8);
    outb(0x3D4, 15);           // Low byte
    outb(0x3D5, pos & 0xFF);
}

// Scroll the screen up by one line
static void scroll() {
    if (cursor_y >= VGA_HEIGHT) {
        // Move all lines up by one
        for (int i = 0; i < (VGA_HEIGHT - 1) * VGA_WIDTH; i++) {
            VGA_BUFFER[i] = VGA_BUFFER[i + VGA_WIDTH];
        }
        // Clear the last line
        for (int i = (VGA_HEIGHT - 1) * VGA_WIDTH; i < VGA_HEIGHT * VGA_WIDTH; i++) {
            VGA_BUFFER[i] = vga_entry(' ', current_color);
        }
        cursor_y = VGA_HEIGHT - 1;
    }
}

void vga_init() {
    vga_clear();
}

void vga_clear() {
    for (int i = 0; i < VGA_WIDTH * VGA_HEIGHT; i++) {
        VGA_BUFFER[i] = vga_entry(' ', current_color);
    }
    cursor_x = 0;
    cursor_y = 0;
    update_cursor();
}

void vga_set_color(VgaColor fg, VgaColor bg) {
    current_color = ((uint8_t)bg << 4) | (uint8_t)fg;
}

void vga_putchar(char c) {
    if (c == '\n') {
        cursor_x = 0;
        cursor_y++;
    } else if (c == '\t') {
        cursor_x = (cursor_x + 8) & ~7;  // Align to next 8-column boundary
    } else if (c == '\b') {
        if (cursor_x > 0) {
            cursor_x--;
            VGA_BUFFER[cursor_y * VGA_WIDTH + cursor_x] = vga_entry(' ', current_color);
        }
    } else {
        VGA_BUFFER[cursor_y * VGA_WIDTH + cursor_x] = vga_entry(c, current_color);
        cursor_x++;
    }

    // Wrap at edge of screen
    if (cursor_x >= VGA_WIDTH) {
        cursor_x = 0;
        cursor_y++;
    }

    scroll();
    update_cursor();
}

void vga_print(const char* str) {
    while (*str) {
        vga_putchar(*str++);
    }
}

void vga_print_hex(uint32_t value) {
    vga_print("0x");
    const char* hex = "0123456789ABCDEF";
    bool leading = true;
    for (int i = 28; i >= 0; i -= 4) {
        uint8_t nibble = (value >> i) & 0xF;
        if (nibble == 0 && leading && i > 0) continue;
        leading = false;
        vga_putchar(hex[nibble]);
    }
    if (leading) vga_putchar('0');
}

void vga_print_dec(uint32_t value) {
    if (value == 0) {
        vga_putchar('0');
        return;
    }
    char buf[12];
    int i = 0;
    while (value > 0) {
        buf[i++] = '0' + (value % 10);
        value /= 10;
    }
    // Print in reverse
    while (i > 0) {
        vga_putchar(buf[--i]);
    }
}
