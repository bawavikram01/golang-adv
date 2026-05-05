#pragma once
#include <stdint.h>

//
// scheduler.h — Round-Robin Preemptive Scheduler
//
// Each "process" is a kernel-mode task with its own stack.
// The timer interrupt triggers schedule() which switches context.
//

#define MAX_PROCESSES 16
#define PROCESS_STACK_SIZE 4096

enum class ProcessState {
    UNUSED,
    READY,
    RUNNING,
    DEAD
};

struct Process {
    uint32_t pid;
    ProcessState state;
    uint32_t esp;           // Saved stack pointer (for context switch)
    uint32_t stack_base;    // Bottom of this process's stack
    const char* name;       // For debug display
};

void scheduler_init();

// Create a new process from a function pointer
// Returns PID, or -1 on failure
int process_create(void (*entry_point)(), const char* name);

// Called by timer interrupt — picks next process and switches
void schedule();

// Get current process PID
uint32_t current_pid();
