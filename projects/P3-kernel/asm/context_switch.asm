; context_switch.asm — Save/Restore process context
;
; void context_switch(uint32_t* old_esp, uint32_t new_esp)
;
; 1. Push callee-saved registers onto the CURRENT stack
; 2. Save current ESP into *old_esp
; 3. Load new_esp as our stack pointer
; 4. Pop the NEW process's registers
; 5. 'ret' jumps to wherever the new process was last preempted
;
; This is THE fundamental operation of multitasking.
; After this function returns, you're running a DIFFERENT process.

section .text
global context_switch

context_switch:
    ; Save callee-saved registers of the CURRENT process
    push ebp
    push ebx
    push esi
    push edi

    ; Save current ESP: *old_esp = ESP
    ; old_esp is at [esp + 20] (4 pushes * 4 bytes + return address)
    mov eax, [esp + 20]     ; eax = old_esp pointer
    mov [eax], esp          ; *old_esp = current ESP

    ; Load new process's ESP
    ; new_esp is at [esp + 24]
    mov esp, [esp + 24]     ; ESP = new_esp  *** WE ARE NOW ON THE NEW STACK ***

    ; Restore new process's callee-saved registers
    pop edi
    pop esi
    pop ebx
    pop ebp

    ; 'ret' pops the return address from the new stack
    ; For a brand new process, this is the entry_point function
    ; For a resumed process, this is wherever it was in schedule()
    ret
