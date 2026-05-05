# P3-kernel Makefile
#
# Build a bootable i686 kernel using NASM + g++ cross-compiler
# Package into an ISO with GRUB for QEMU

ASM      = nasm
CXX      = g++
LD       = ld

ASMFLAGS = -f elf32
CXXFLAGS = -m32 -ffreestanding -O2 -Wall -Wextra -fno-exceptions -fno-rtti -Iinclude -std=c++17
LDFLAGS  = -m elf_i386 -T linker.ld -nostdlib

BUILD    = build
ISO_DIR  = iso
KERNEL   = $(ISO_DIR)/boot/kernel.bin
ISO      = os.iso

# Source files
ASM_SRCS = $(wildcard asm/*.asm)
CXX_SRCS = $(wildcard src/*.cpp)

# Object files
ASM_OBJS = $(patsubst asm/%.asm, $(BUILD)/%.asm.o, $(ASM_SRCS))
CXX_OBJS = $(patsubst src/%.cpp, $(BUILD)/%.o, $(CXX_SRCS))
OBJS     = $(ASM_OBJS) $(CXX_OBJS)

.PHONY: all clean run debug iso

all: $(KERNEL)

# Link everything into kernel binary
$(KERNEL): $(OBJS)
	@mkdir -p $(dir $@)
	$(LD) $(LDFLAGS) -o $@ $^

# Assemble .asm files
$(BUILD)/%.asm.o: asm/%.asm
	@mkdir -p $(BUILD)
	$(ASM) $(ASMFLAGS) $< -o $@

# Compile .cpp files
$(BUILD)/%.o: src/%.cpp
	@mkdir -p $(BUILD)
	$(CXX) $(CXXFLAGS) -c $< -o $@

# Create bootable ISO with GRUB
iso: $(KERNEL)
	grub-mkrescue -o $(ISO) $(ISO_DIR)

# QEMU command — VS Code snap contaminates the library path, so we run with a clean env
QEMU = env -i PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin HOME=$(HOME) DISPLAY=$(DISPLAY) TERM=xterm qemu-system-i386

# Run in QEMU (from ISO)
run: iso
	$(QEMU) -cdrom $(ISO) -serial stdio

# Run kernel directly (faster, no ISO needed)
run-kernel: $(KERNEL)
	$(QEMU) -kernel $(KERNEL) -serial stdio

# Debug with GDB
debug: iso
	$(QEMU) -cdrom $(ISO) -serial stdio -s -S &
	@echo "GDB: target remote localhost:1234"

clean:
	rm -rf $(BUILD) $(KERNEL) $(ISO)
