# OS Study Path: Beginner → God Level

---

## Phase 1: Foundations (Weeks 1–4)

**Goal:** Understand what an OS does and why, at a conceptual level.

### Book: OSTEP (Operating Systems: Three Easy Pieces)
- Free online: https://pages.cs.wisc.edu/~remzi/OSTEP/
- Read in this order:

| Week | Chapters | Topic | Key Takeaway |
|------|----------|-------|--------------|
| 1 | 1–6 | Virtualization intro, processes | A process is just an illusion the OS creates |
| 1 | 7–11 | Scheduling (FCFS, SJF, RR, MLFQ) | Tradeoffs: latency vs throughput |
| 2 | 12–16 | Address spaces, memory API, segmentation | Every process thinks it has all memory |
| 2 | 17–24 | Paging, TLB, swapping | The MMU translates every address automatically |
| 3 | 25–28 | Concurrency, threads, locks | Shared mutable state is the root of all evil |
| 3 | 29–34 | Condition vars, semaphores, bugs | Deadlock, livelock, starvation |
| 4 | 35–42 | I/O, disks, filesystems, journaling | From spinning rust to reliable storage |
| 4 | 43–44 | Flash, data integrity | Modern storage guarantees |

### Daily Practice (During Phase 1):
```
30 min: Read OSTEP chapter
20 min: Do the homework questions at the end (they have simulators!)
10 min: Look at the corresponding file in YOUR P3-kernel and connect theory → code
```

### Checkpoints:
- [ ] Can explain virtual vs physical addresses without looking anything up
- [ ] Can draw the 2-level page table translation on paper
- [ ] Can explain why fork() + exec() is a two-step design
- [ ] Can explain what happens when malloc() runs out of heap (brk/mmap)
- [ ] Can describe 3 approaches to mutual exclusion and their tradeoffs

---

## Phase 2: Hands-On Kernel Hacking (Weeks 5–10)

**Goal:** Read and modify a real teaching kernel. Bridge theory → working code.

### Lab: MIT 6.S081 / xv6 (RISC-V version)
- Course page: https://pdos.csail.mit.edu/6.828/2023/
- xv6 book (read alongside): https://pdos.csail.mit.edu/6.828/2023/xv6/book-riscv-rev3.pdf
- xv6 source: ~8000 lines of C. Readable in a weekend.

| Week | Lab | What You'll Build |
|------|-----|-------------------|
| 5 | Lab: Utilities | Write user-space programs (sleep, find, xargs) using syscalls |
| 6 | Lab: System Calls | Add a new syscall to xv6 (trace, sysinfo) |
| 7 | Lab: Page Tables | Implement per-process kernel page tables, detect accessed pages |
| 8 | Lab: Traps | Handle page faults for lazy allocation and COW fork |
| 9 | Lab: Copy-on-Write | Real COW fork — the page fault handler allocates on write |
| 9 | Lab: Multithreading | User-level threads, synchronization |
| 10 | Lab: Locks | Redesign locks for better parallelism (per-CPU freelists) |
| 10 | Lab: File System | Add large files (double-indirect) and symbolic links |

### Why xv6:
- It's a UNIX. Same design as Linux but 100× smaller.
- Has: processes, virtual memory, syscalls, filesystem, pipes, signals
- You'll modify REAL kernel code — add features, fix bugs
- Prepares you to read Linux source later

### Daily Practice (During Phase 2):
```
Read one xv6 source file per day:
  Day 1: kernel/proc.c (process management — compare to your scheduler.cpp)
  Day 2: kernel/vm.c (page tables — compare to your paging.cpp)
  Day 3: kernel/trap.c (interrupt handling — compare to your idt.cpp)
  Day 4: kernel/fs.c (filesystem — compare to your ramdisk.cpp)
  Day 5: kernel/pipe.c (IPC — you built this in P1!)
```

### Checkpoints:
- [ ] Can add a new syscall to xv6 from scratch
- [ ] Can implement COW fork (handle page fault, allocate, copy, remap)
- [ ] Can explain every line in xv6's context switch (swtch.S)
- [ ] Can trace a file read from open() through VFS to disk block
- [ ] Can explain xv6's lock ordering and why it prevents deadlock

---

## Phase 3: Linux Internals (Weeks 11–20)

**Goal:** Understand how a production kernel handles millions of processes, terabytes of RAM, and thousands of devices.

### Book: "Linux Kernel Development" by Robert Love (3rd ed)
Read chapters in this order:

| Week | Chapters | Focus |
|------|----------|-------|
| 11 | 1–3 | Kernel intro, process management, scheduling |
| 12 | 4–5 | CFS scheduler deep dive, system calls |
| 13 | 6–8 | Kernel data structures, interrupts, bottom halves |
| 14 | 9–10 | Kernel synchronization (spinlocks, RCU, per-CPU) |
| 15 | 11–12 | Timers, memory management (zones, buddy, slab) |
| 16 | 13–14 | VFS, block I/O layer |
| 17 | 15–16 | Process address space, page cache |
| 18 | 17–18 | Devices, modules |

### Supplement: "Understanding the Linux Kernel" (Bovet & Cesati)
- Heavier. Use as reference when Love doesn't go deep enough.
- Chapter 2 (Memory Addressing) and Chapter 8 (Memory Management) are gold.

### Hands-On Linux Kernel Activities:

```bash
# 1. Read actual kernel source (start with these files):
git clone --depth=1 https://github.com/torvalds/linux.git
less linux/kernel/sched/core.c      # Scheduler
less linux/mm/page_alloc.c          # Buddy allocator
less linux/mm/mmap.c                # mmap implementation
less linux/fs/read_write.c          # VFS read/write
less linux/arch/x86/kernel/traps.c  # Exception handlers

# 2. Write a kernel module:
# Simple character device that exposes data via /dev/mydevice
# Use: module_init, module_exit, file_operations

# 3. Use ftrace/perf to trace kernel execution:
sudo perf record -g -a -- sleep 5
sudo perf report                    # See kernel call stacks

# 4. Add a /proc entry:
# Write a module that creates /proc/myinfo showing custom kernel stats
```

### Checkpoints:
- [ ] Can explain CFS vruntime calculation and red-black tree usage
- [ ] Can trace a page fault through Linux source (mm/memory.c:handle_mm_fault)
- [ ] Can write and load a Linux kernel module
- [ ] Can explain RCU (Read-Copy-Update) and when to use it vs spinlocks
- [ ] Can explain the difference between softirq, tasklet, and workqueue
- [ ] Can use ftrace to trace scheduler decisions

---

## Phase 4: Specialization & Advanced Topics (Weeks 21–30)

**Goal:** Go deep in 2-3 areas. This is where you become dangerous.

### Choose Your Tracks (pick 2-3):

#### Track A: Memory Management Deep Dive
```
Read:
  - "What Every Programmer Should Know About Memory" (Ulrich Drepper, free PDF)
  - Linux mm/ source code: slub.c, vmscan.c, oom_kill.c
  - NUMA balancing: kernel/sched/fair.c (task placement)

Build:
  - Custom page replacement algorithm (modify mm/vmscan.c)
  - NUMA-aware memory allocator
  - Memory compaction tool
```

#### Track B: Scheduler & Real-Time
```
Read:
  - "Inside the Linux Scheduler" (CFS paper)
  - SCHED_DEADLINE: Earliest Deadline First in Linux
  - "A Complete Guide to Linux Process Scheduling" (PhD thesis, Nikita Desai)

Build:
  - Custom scheduler class in Linux (like SCHED_MY)
  - Real-time latency measurement tool
  - Compare CFS vs BFS (Brain Fuck Scheduler)
```

#### Track C: Filesystems & Storage
```
Read:
  - "Design and Implementation of ext4" (kernel docs)
  - ZFS architecture paper
  - "F2FS: A New Filesystem for Flash Storage"

Build:
  - Simple FUSE filesystem (in userspace — easier to debug)
  - Implement journaling in your P3 ramdisk
  - Write a log-structured filesystem
```

#### Track D: Networking Stack
```
Read:
  - "TCP/IP Illustrated" Vol 1 (Stevens)
  - Linux networking: net/ipv4/tcp.c, net/core/dev.c
  - DPDK / XDP / io_uring architecture

Build:
  - Raw socket packet sniffer
  - Simple TCP stack from scratch (SYN/ACK/FIN state machine)
  - XDP program for packet filtering
```

#### Track E: Virtualization & Containers
```
Read:
  - "Hardware and Software Support for Virtualization" (Bugnion, Nieh, Tsafrir)
  - Linux KVM source: arch/x86/kvm/
  - Namespaces + cgroups (what Docker actually is)

Build:
  - Mini container runtime (use clone() with CLONE_NEWPID|CLONE_NEWNS)
  - Simple hypervisor using KVM API (/dev/kvm)
  - Understand: EPT (Extended Page Tables), VMCS, VM exits
```

---

## Phase 5: Research & Contribution (Ongoing)

**Goal:** Original thought. You don't just USE knowledge — you CREATE it.

### Read Papers:
```
Seminal papers every OS person should read:
  1. "The UNIX Time-Sharing System" (Ritchie & Thompson, 1974)
  2. "A Fast File System for UNIX" (McKusick, 1984)
  3. "Microkernel vs Monolithic" (Linus vs Tanenbaum debate, 1992)
  4. "Exokernel" (Engler, 1995) — OS as library
  5. "seL4: Formal Verification of an OS Kernel" (Klein et al., 2009)
  6. "LegoOS: Disaggregated OS for Hardware Resource Pool" (2018)
  7. "Demikernel: Library OS for Kernel-Bypass" (2021)
```

### Contribute:
```
1. Linux kernel mailing list (LKML) — read discussions
2. Pick a subsystem, subscribe to its mailing list
3. Start with: documentation fixes, compiler warning fixes
4. Graduate to: actual bug fixes (find bugs via syzkaller reports)
5. Eventually: propose features, write RFCs
```

### Build Something Novel:
```
Ideas for projects that push boundaries:
  - OS for persistent memory (Intel Optane) — no volatile/non-volatile distinction
  - Kernel for heterogeneous computing (CPU + GPU + FPGA unified scheduling)
  - Formally verified microkernel (like seL4 but for RISC-V)
  - OS that uses ML to predict page faults / scheduling decisions
```

---

## Daily Routine Template

```
┌─────────────────────────────────────────────────────────────┐
│ MORNING (1 hour):                                           │
│   - Read: 1 chapter / 20 pages / 1 paper                   │
│   - Handwrite notes (forces processing, not just reading)   │
│                                                             │
│ AFTERNOON (1-2 hours):                                      │
│   - Code: Lab exercise / kernel module / source reading     │
│   - Run experiments (boot modified kernel, trace with perf) │
│                                                             │
│ EVENING (30 min):                                           │
│   - Review: Can you explain today's topic to a rubber duck? │
│   - Connect: How does this relate to your P3 kernel?        │
│   - Note one thing you don't fully understand → tomorrow's  │
│     first task                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Resource Links

| Resource | URL | Use For |
|----------|-----|---------|
| OSTEP | https://pages.cs.wisc.edu/~remzi/OSTEP/ | Phase 1: Theory |
| MIT 6.S081 | https://pdos.csail.mit.edu/6.828/2023/ | Phase 2: xv6 labs |
| xv6 Source | https://github.com/mit-pdos/xv6-riscv | Phase 2: Reading |
| Linux Source | https://elixir.bootlin.com/linux/latest/source | Phase 3+: Browsable |
| OSDev Wiki | https://wiki.osdev.org/ | Reference for x86 details |
| LWN.net | https://lwn.net/ | Linux kernel news & deep articles |
| Brendan Gregg | https://www.brendangregg.com/ | Performance, tracing, BPF |

---

## Progress Tracker

```
Phase 1: [ ] OSTEP complete, [ ] All homework simulations done
Phase 2: [ ] xv6 labs 1-8 complete, [ ] Can read all xv6 source
Phase 3: [ ] Love book done, [ ] First kernel module written, [ ] Can trace in perf
Phase 4: [ ] Two specializations explored, [ ] Non-trivial project built
Phase 5: [ ] Read 5+ research papers, [ ] First kernel patch submitted
```

---

## The Mindset

```
"I don't know this yet" → GOOD. That's where growth lives.
"This is confusing" → Read it 3 times. Draw it. Implement it. Then it clicks.
"Why does Linux do it THIS way?" → There's always a historical reason. Find it.

You already built: a shell (P1), a memory allocator (P2), a kernel (P3).
You're not a beginner. You're someone who builds OS-level systems.
Now you're going from builder → architect → researcher.
```
