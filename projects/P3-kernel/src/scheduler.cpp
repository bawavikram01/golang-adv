#include "scheduler.h"
#include "vga.h"
#include "timer.h"
#include "idt.h"

//
// scheduler.cpp — Round-robin preemptive scheduler
//
// How it works:
//   1. Each process has its own 4KB stack
//   2. Timer fires every 10ms → calls schedule()
//   3. schedule() saves current ESP, loads next process's ESP
//   4. When we "return," we're now running a different process
//
// The context_switch is a function that:
//   - Pushes callee-saved registers (EBP, EBX, ESI, EDI) on current stack
//   - Saves current ESP to the current process struct
//   - Loads new process's ESP
//   - Pops that process's callee-saved registers
//   - Returns (to wherever that process was last interrupted)
//

static Process processes[MAX_PROCESSES];
static int current_process = -1;
static int num_processes = 0;

// Assembly function — does the actual ESP swap
extern "C" void context_switch(uint32_t* old_esp, uint32_t new_esp);

// How a process exits (shouldn't happen for our demo processes, but safety)
static void process_exit() {
    processes[current_process].state = ProcessState::DEAD;
    // Yield to scheduler — we'll never come back
    for (;;) {
        asm volatile("hlt");
    }
}

void scheduler_init() {
    for (int i = 0; i < MAX_PROCESSES; i++) {
        processes[i].state = ProcessState::UNUSED;
    }
    vga_print("[SCHED] Scheduler initialized\n");
}

int process_create(void (*entry_point)(), const char* name) {
    // Find a free slot
    int slot = -1;
    for (int i = 0; i < MAX_PROCESSES; i++) {
        if (processes[i].state == ProcessState::UNUSED) {
            slot = i;
            break;
        }
    }
    if (slot == -1) return -1;  // No free slots

    // Allocate a stack (we use a static array here for simplicity)
    // In a real OS, you'd use pmm_alloc() for the stack pages
    static uint8_t stacks[MAX_PROCESSES][PROCESS_STACK_SIZE] __attribute__((aligned(16)));

    Process& p = processes[slot];
    p.pid = slot;
    p.state = ProcessState::READY;
    p.name = name;
    p.stack_base = (uint32_t)&stacks[slot][0];

    // Set up the initial stack so that when context_switch pops registers
    // and does 'ret', it "returns" to entry_point.
    //
    // Stack layout (grows downward):
    //   [process_exit]   ← return address if entry_point() returns
    //   [entry_point]    ← where 'ret' in context_switch will jump
    //   [0 - EBP]       ← fake saved registers (context_switch expects these)
    //   [0 - EBX]
    //   [0 - ESI]
    //   [0 - EDI]       ← ESP points here
    //
    uint32_t* stack_top = (uint32_t*)(p.stack_base + PROCESS_STACK_SIZE);

    *(--stack_top) = (uint32_t)process_exit;  // Return address (safety net)
    *(--stack_top) = (uint32_t)entry_point;   // 'ret' will jump here
    *(--stack_top) = 0;  // EBP
    *(--stack_top) = 0;  // EBX
    *(--stack_top) = 0;  // ESI
    *(--stack_top) = 0;  // EDI

    p.esp = (uint32_t)stack_top;

    num_processes++;

    vga_print("[SCHED] Created process '");
    vga_print(name);
    vga_print("' (PID ");
    vga_print_dec(slot);
    vga_print(")\n");

    return slot;
}

void schedule() {
    if (num_processes == 0) return;

    int prev = current_process;

    // Find next READY process (round-robin)
    int next = (current_process + 1) % MAX_PROCESSES;
    int checked = 0;
    while (checked < MAX_PROCESSES) {
        if (processes[next].state == ProcessState::READY ||
            processes[next].state == ProcessState::RUNNING) {
            break;
        }
        next = (next + 1) % MAX_PROCESSES;
        checked++;
    }

    if (checked >= MAX_PROCESSES) return;  // No runnable process
    if (next == current_process) return;   // Already running this one

    // Switch
    if (prev >= 0 && processes[prev].state == ProcessState::RUNNING) {
        processes[prev].state = ProcessState::READY;
    }

    processes[next].state = ProcessState::RUNNING;
    current_process = next;

    if (prev < 0) {
        // First time — no old context to save, just load new one
        // We fake an "old esp" location on the kernel stack (it won't be used again)
        static uint32_t dummy_esp;
        context_switch(&dummy_esp, processes[next].esp);
    } else {
        context_switch(&processes[prev].esp, processes[next].esp);
    }
}

uint32_t current_pid() {
    if (current_process < 0) return 0;
    return processes[current_process].pid;
}
