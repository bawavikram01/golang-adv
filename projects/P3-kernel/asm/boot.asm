; boot.asm — Multiboot header + kernel entry stub
;
; This is where the CPU starts executing after GRUB loads us.
; We set up a small stack, then jump to the C++ kernel_main().
;
; GRUB requires a Multiboot header in the first 8KB of the binary.
; The header tells GRUB: "I'm a valid kernel, load me."
;

; Multiboot constants
MBOOT_MAGIC     equ 0x1BADB002     ; Magic number GRUB looks for
MBOOT_FLAGS     equ 0x00000003     ; Flags: align modules + provide memory map
MBOOT_CHECKSUM  equ -(MBOOT_MAGIC + MBOOT_FLAGS)  ; Must sum to 0

; ============================================================
; Multiboot Header (must be in first 8KB of binary)
; ============================================================
section .multiboot
align 4
    dd MBOOT_MAGIC
    dd MBOOT_FLAGS
    dd MBOOT_CHECKSUM

; ============================================================
; Stack (16KB — plenty for a simple kernel)
; ============================================================
section .bss
align 16
stack_bottom:
    resb 16384          ; 16 KB stack
stack_top:

; ============================================================
; Entry Point — GRUB jumps here
; ============================================================
section .text
global _start
global gdt_flush
extern kernel_main      ; Defined in kernel.cpp

_start:
    ; Set up our own stack (GRUB doesn't guarantee a valid one)
    mov esp, stack_top

    ; Push multiboot info pointer (EBX) and magic number (EAX)
    ; kernel_main(uint32_t magic, uint32_t* mboot_info)
    push ebx            ; Multiboot info struct pointer
    push eax            ; Multiboot magic number (should be 0x2BADB002)

    ; Call the C++ kernel entry point
    call kernel_main

    ; If kernel_main returns (it shouldn't), halt the CPU
.hang:
    cli                 ; Disable interrupts
    hlt                 ; Halt
    jmp .hang           ; In case of NMI, loop back

; ============================================================
; gdt_flush — Load the GDT and reload segment registers
; Called from C++: gdt_flush(uint32_t gdt_ptr_addr)
; ============================================================
gdt_flush:
    mov eax, [esp + 4]  ; Get pointer to GdtPointer struct
    lgdt [eax]           ; Load GDT register

    ; Reload segment registers with kernel data segment (0x10)
    mov ax, 0x10
    mov ds, ax
    mov es, ax
    mov fs, ax
    mov gs, ax
    mov ss, ax

    ; Far jump to reload CS with kernel code segment (0x08)
    jmp 0x08:.flush_done
.flush_done:
    ret
