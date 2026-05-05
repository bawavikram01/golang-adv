#include "ramdisk.h"
#include "multiboot.h"
#include "vga.h"

//
// ramdisk.cpp — Read files from a GRUB-loaded module
//
// GRUB can load additional files (modules) alongside the kernel.
// We treat the first module as our ramdisk filesystem.
//

static uint32_t ramdisk_base = 0;
static RamdiskHeader* header = nullptr;
static FileEntry* file_table = nullptr;

// Simple string comparison (no libc in freestanding)
static bool str_equal(const char* a, const char* b) {
    while (*a && *b) {
        if (*a != *b) return false;
        a++; b++;
    }
    return *a == *b;
}

void ramdisk_init(uint32_t mboot_info_addr) {
    MultibootInfo* mboot = (MultibootInfo*)mboot_info_addr;

    // Check if GRUB loaded any modules
    if (mboot->mods_count == 0) {
        vga_set_color(VGA_YELLOW, VGA_BLACK);
        vga_print("[RAMDISK] No modules loaded (ramdisk not available)\n");
        vga_set_color(VGA_LIGHT_GREY, VGA_BLACK);
        return;
    }

    // Get the first module (our ramdisk)
    MultibootModule* mod = (MultibootModule*)mboot->mods_addr;
    ramdisk_base = mod->mod_start;
    uint32_t ramdisk_size = mod->mod_end - mod->mod_start;

    header = (RamdiskHeader*)ramdisk_base;

    // Validate magic
    if (header->magic != RAMDISK_MAGIC) {
        vga_set_color(VGA_LIGHT_RED, VGA_BLACK);
        vga_print("[RAMDISK] ERROR: invalid magic (");
        vga_print_hex(header->magic);
        vga_print(")\n");
        vga_set_color(VGA_LIGHT_GREY, VGA_BLACK);
        header = nullptr;
        return;
    }

    // File table starts right after the header
    file_table = (FileEntry*)(ramdisk_base + sizeof(RamdiskHeader));

    vga_print("[RAMDISK] Loaded: ");
    vga_print_dec(header->num_files);
    vga_print(" files, ");
    vga_print_dec(ramdisk_size);
    vga_print(" bytes\n");
}

void ramdisk_list() {
    if (!header) {
        vga_print("  (no ramdisk)\n");
        return;
    }

    vga_print("  Files in ramdisk:\n");
    for (uint32_t i = 0; i < header->num_files; i++) {
        vga_print("    ");
        vga_print(file_table[i].name);
        vga_print("  (");
        vga_print_dec(file_table[i].size);
        vga_print(" bytes)\n");
    }
}

const char* ramdisk_read(const char* name, uint32_t* size) {
    if (!header) return nullptr;

    for (uint32_t i = 0; i < header->num_files; i++) {
        if (str_equal(file_table[i].name, name)) {
            *size = file_table[i].size;
            return (const char*)(ramdisk_base + file_table[i].offset);
        }
    }
    return nullptr;  // Not found
}
