# System Programming — From Backend Dev to God Level

## What is System Programming?

System programming is writing software that provides services to other software (not directly to users). It operates close to hardware, manages resources, and demands understanding of how computers *actually* work — memory layouts, CPU instructions, OS internals, concurrency primitives, and network stacks.

---

## Your Learning Roadmap

### Phase 1: Foundations (Weeks 1–4)
| # | Topic | Key Concepts |
|---|-------|-------------|
| 1 | C Mastery | Pointers, memory layout, stack vs heap, undefined behavior |
| 2 | How Programs Run | ELF format, linking, loading, compilation pipeline |
| 3 | Memory Management | Virtual memory, paging, mmap, allocators |
| 4 | Process Model | fork/exec, process lifecycle, signals |

### Phase 2: OS Internals (Weeks 5–8)
| # | Topic | Key Concepts |
|---|-------|-------------|
| 5 | System Calls | syscall interface, strace, kernel boundary |
| 6 | File Systems | VFS, inodes, file descriptors, I/O models |
| 7 | Concurrency | threads, mutexes, futexes, atomics, memory ordering |
| 8 | Scheduling | CFS, preemption, context switching |

### Phase 3: Networking & IPC (Weeks 9–12)
| # | Topic | Key Concepts |
|---|-------|-------------|
| 9 | Socket Programming | TCP/UDP, epoll/io_uring, non-blocking I/O |
| 10 | IPC Mechanisms | pipes, shared memory, message queues, Unix sockets |
| 11 | Building a Server | event loops, connection handling, zero-copy |
| 12 | Protocol Implementation | HTTP from scratch, wire formats |

### Phase 4: Advanced (Weeks 13–16)
| # | Topic | Key Concepts |
|---|-------|-------------|
| 13 | Linux Kernel Modules | writing a char device, kernel APIs |
| 14 | Performance | perf, flamegraphs, cache effects, SIMD |
| 15 | Security | ASLR, stack canaries, seccomp, capabilities |
| 16 | Rust for Systems | ownership, unsafe, FFI, async runtime internals |

---

## How Each Module Works

Each module contains:
- `concepts.md` — Theory explained clearly
- `code/` — Hands-on programs to write and run
- `exercises.md` — Challenges to solve yourself
- `Makefile` — Build instructions

---

## Tools You'll Use

- **GCC/Clang** — Compilation
- **GDB** — Debugging at instruction level
- **strace/ltrace** — Syscall/library tracing
- **perf** — Performance profiling
- **objdump/readelf** — Binary inspection
- **Valgrind** — Memory error detection

---

## Recommended References

1. *Computer Systems: A Programmer's Perspective* (CS:APP) — Bryant & O'Hallaron
2. *The Linux Programming Interface* — Michael Kerrisk
3. *Operating Systems: Three Easy Pieces* (free: ostep.org)
4. *Linux Kernel Development* — Robert Love
5. *Beej's Guide to Network Programming* (free)

---

## Start Here

Begin with `01_c_and_memory/` — every system programmer must think in C.

```bash
cd 01_c_and_memory
cat concepts.md
```
