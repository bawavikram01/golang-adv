#pragma once
#include <stdint.h>

//
// vga.h — VGA text mode driver (80x25, 16 colors)
//
// The VGA text buffer lives at physical address 0xB8000.
// Each character = 2 bytes: [ASCII char][color attribute]
// Color attribute: high nibble = background, low nibble = foreground
//

enum VgaColor : uint8_t {
    VGA_BLACK        = 0,
    VGA_BLUE         = 1,
    VGA_GREEN        = 2,
    VGA_CYAN         = 3,
    VGA_RED          = 4,
    VGA_MAGENTA      = 5,
    VGA_BROWN        = 6,
    VGA_LIGHT_GREY   = 7,
    VGA_DARK_GREY    = 8,
    VGA_LIGHT_BLUE   = 9,
    VGA_LIGHT_GREEN  = 10,
    VGA_LIGHT_CYAN   = 11,
    VGA_LIGHT_RED    = 12,
    VGA_LIGHT_MAGENTA= 13,
    VGA_YELLOW       = 14,
    VGA_WHITE        = 15,
};

void vga_init();
void vga_clear();
void vga_putchar(char c);
void vga_print(const char* str);
void vga_print_hex(uint32_t value);
void vga_print_dec(uint32_t value);
void vga_set_color(VgaColor fg, VgaColor bg);
