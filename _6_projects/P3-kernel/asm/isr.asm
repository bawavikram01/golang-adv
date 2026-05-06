; isr.asm — Interrupt Service Routine stubs
;
; When an interrupt fires, the CPU pushes some state automatically.
; We need assembly stubs that:
;   1. Push a dummy error code (if the CPU didn't push one)
;   2. Push the interrupt number
;   3. Save all registers
;   4. Call our C++ handler
;   5. Restore registers and iret
;
; ISRs 0-31 = CPU exceptions (some push error codes, some don't)
; IRQs 32-47 = Hardware interrupts (remapped PIC)

section .text

extern isr_handler      ; C++ function: void isr_handler(Registers* regs)
extern irq_handler      ; C++ function: void irq_handler(Registers* regs)

; Common ISR stub — saves all registers, calls C++ handler
isr_common_stub:
    ; At this point, stack has: [SS, ESP, EFLAGS, CS, EIP, error_code, int_no]
    ; (pushed by CPU + our stub)

    ; Save all general-purpose registers
    pusha               ; Pushes EAX, ECX, EDX, EBX, ESP, EBP, ESI, EDI

    ; Save data segment
    mov ax, ds
    push eax

    ; Load kernel data segment
    mov ax, 0x10        ; Kernel data segment selector
    mov ds, ax
    mov es, ax
    mov fs, ax
    mov gs, ax

    ; Pass pointer to the register struct (current ESP)
    push esp
    call isr_handler
    add esp, 4          ; Clean up pushed argument

    ; Restore data segment
    pop eax
    mov ds, ax
    mov es, ax
    mov fs, ax
    mov gs, ax

    ; Restore general-purpose registers
    popa

    ; Remove interrupt number and error code from stack
    add esp, 8

    ; Return from interrupt
    iret

; Common IRQ stub — same but calls irq_handler
irq_common_stub:
    pusha

    mov ax, ds
    push eax

    mov ax, 0x10
    mov ds, ax
    mov es, ax
    mov fs, ax
    mov gs, ax

    push esp
    call irq_handler
    add esp, 4

    pop eax
    mov ds, ax
    mov es, ax
    mov fs, ax
    mov gs, ax

    popa
    add esp, 8
    iret

; ============================================================
; ISR stubs (exceptions 0-31)
; Some push error codes automatically, some don't.
; We add a dummy 0 for those that don't, so the stack is uniform.
; ============================================================

%macro ISR_NOERRCODE 1
global isr%1
isr%1:
    push dword 0        ; Dummy error code
    push dword %1       ; Interrupt number
    jmp isr_common_stub
%endmacro

%macro ISR_ERRCODE 1
global isr%1
isr%1:
    ; CPU already pushed the error code
    push dword %1       ; Interrupt number
    jmp isr_common_stub
%endmacro

; CPU Exceptions
ISR_NOERRCODE 0   ; Division by zero
ISR_NOERRCODE 1   ; Debug
ISR_NOERRCODE 2   ; NMI
ISR_NOERRCODE 3   ; Breakpoint
ISR_NOERRCODE 4   ; Overflow
ISR_NOERRCODE 5   ; Bound range exceeded
ISR_NOERRCODE 6   ; Invalid opcode
ISR_NOERRCODE 7   ; Device not available
ISR_ERRCODE   8   ; Double fault
ISR_NOERRCODE 9   ; Coprocessor segment overrun
ISR_ERRCODE   10  ; Invalid TSS
ISR_ERRCODE   11  ; Segment not present
ISR_ERRCODE   12  ; Stack-segment fault
ISR_ERRCODE   13  ; General protection fault
ISR_ERRCODE   14  ; Page fault
ISR_NOERRCODE 15  ; Reserved
ISR_NOERRCODE 16  ; x87 FPU error
ISR_ERRCODE   17  ; Alignment check
ISR_NOERRCODE 18  ; Machine check
ISR_NOERRCODE 19  ; SIMD floating-point
ISR_NOERRCODE 20  ; Virtualization
ISR_NOERRCODE 21  ; Reserved
ISR_NOERRCODE 22  ; Reserved
ISR_NOERRCODE 23  ; Reserved
ISR_NOERRCODE 24  ; Reserved
ISR_NOERRCODE 25  ; Reserved
ISR_NOERRCODE 26  ; Reserved
ISR_NOERRCODE 27  ; Reserved
ISR_NOERRCODE 28  ; Reserved
ISR_NOERRCODE 29  ; Reserved
ISR_ERRCODE   30  ; Security exception
ISR_NOERRCODE 31  ; Reserved

; ============================================================
; IRQ stubs (hardware interrupts, remapped to 32-47)
; ============================================================

%macro IRQ 2
global irq%1
irq%1:
    push dword 0        ; Dummy error code
    push dword %2       ; Interrupt number (32+)
    jmp irq_common_stub
%endmacro

IRQ 0, 32   ; Timer (PIT)
IRQ 1, 33   ; Keyboard
IRQ 2, 34   ; Cascade (never raised)
IRQ 3, 35   ; COM2
IRQ 4, 36   ; COM1
IRQ 5, 37   ; LPT2
IRQ 6, 38   ; Floppy
IRQ 7, 39   ; LPT1 / spurious
IRQ 8, 40   ; CMOS RTC
IRQ 9, 41   ; Free
IRQ 10, 42  ; Free
IRQ 11, 43  ; Free
IRQ 12, 44  ; PS/2 Mouse
IRQ 13, 45  ; FPU
IRQ 14, 46  ; Primary ATA
IRQ 15, 47  ; Secondary ATA
