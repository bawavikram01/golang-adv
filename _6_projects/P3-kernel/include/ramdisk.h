#pragma once
#include <stdint.h>

//
// ramdisk.h — Simple Flat Filesystem (loaded by GRUB as a module)
//
// Layout:
//   [RamdiskHeader]
//   [FileEntry 0]
//   [FileEntry 1]
//   ...
//   [file data bytes]
//

#define RAMDISK_MAGIC 0xDEADBEEF
#define MAX_FILENAME 32

struct RamdiskHeader {
    uint32_t magic;
    uint32_t num_files;
};

struct FileEntry {
    char name[MAX_FILENAME];
    uint32_t offset;  // Offset from start of ramdisk
    uint32_t size;
};

void ramdisk_init(uint32_t mboot_info_addr);

// List all files in the ramdisk
void ramdisk_list();

// Read a file by name. Returns pointer to data and fills 'size'.
// Returns nullptr if not found.
const char* ramdisk_read(const char* name, uint32_t* size);
