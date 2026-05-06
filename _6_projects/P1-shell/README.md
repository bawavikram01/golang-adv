# P1: Build Your Own Shell (C++)

## What Is a Shell?

A shell is just a program that:
1. Prints a prompt (`$ `)
2. Reads a line of input (`ls -la | grep foo`)
3. Parses it into commands
4. Executes those commands using OS system calls
5. Waits for them to finish
6. Repeats

That's it. Bash, Zsh, Fish — they all do this at their core. You're building the same thing.

---

## Concepts You Need Before Starting

### 1. Process — What Even Is One?
When you run `./myprogram`, the OS creates a **process**: an isolated instance of your program with its own memory, registers, and file descriptors. Every command you type in a shell becomes a process.

**Key idea:** Your shell is a process. When you type `ls`, your shell *creates a child process* to run `ls`.

### 2. The Big Three System Calls

Your entire shell revolves around 3 system calls:

```
fork()  → Creates a copy of the current process (child process)
exec()  → Replaces the current process with a new program
wait()  → Parent waits for child to finishCreates a copy of the current process (child process)
```

Here's how they work together:

```
Your Shell (parent process)
    │
    ├── User types: ls -la
    │
    ├── fork()  ──→  Child process (exact copy of shell)
    │                    │
    │                    └── exec("ls", ["-la"])
    │                         │
    │                         └── ls runs, prints output, exits
    │
    └── wait()  ──→  Shell resumes when child exits
```

**That's 80% of a shell.** The rest is parsing and plumbing.

### 3. File Descriptors — How I/O Actually Works

Every process has a table of **file descriptors** (just integers):
- `0` = stdin (keyboard input)
- `1` = stdout (screen output)  
- `2` = stderr (error output)

When you do `ls > output.txt`, the shell:
1. Opens `output.txt` (gets fd 3)
2. Copies fd 3 onto fd 1 (`dup2(3, 1)`) — now stdout points to the file
3. Runs `ls` — it writes to stdout, which is now the file

**Key insight:** Programs don't know where their output goes. They always write to fd 1. The shell controls *where* fd 1 points.

### 4. Pipes — Connecting Processes

`ls | grep foo` means:
- `ls` writes to stdout → which goes into a **pipe** → `grep` reads from stdin

A pipe is just a kernel buffer with two ends:
- Write end (fd)
- Read end (fd)

```c++
int pipefd[2];
pipe(pipefd);  // pipefd[0] = read end, pipefd[1] = write end
```

The shell creates the pipe, then:
- Child 1 (`ls`): `dup2(pipefd[1], STDOUT)` — stdout goes into pipe
- Child 2 (`grep`): `dup2(pipefd[0], STDIN)` — stdin comes from pipe

### 5. Signals — Handling Ctrl+C

When you press Ctrl+C, the OS sends `SIGINT` to the foreground process group. Your shell needs to:
- **Not die** when Ctrl+C is pressed (ignore it in the shell itself)
- **Let the child die** (forward the signal to the child)

---

## Build Plan — 7 Milestones

Build these one at a time. Each milestone is a working shell with one more feature.

### Milestone 1: REPL Loop + Execute Simple Commands
```
$ ls
file1.txt  file2.txt
$ pwd
/home/vikram
$ echo hello world
hello world
```
**What you implement:** Read input → split by spaces → `fork()` → `exec()` → `wait()`
**System calls:** `fork`, `execvp`, `waitpid`

### Milestone 2: Built-in Commands
```
$ cd /home/vikram
$ cd ..
$ exit
```
**What you implement:** `cd` can't be a child process (it must change the shell's own directory). Handle it before fork.
**System calls:** `chdir`

### Milestone 3: Output Redirection (`>`, `>>`, `<`)
```
$ ls -la > output.txt
$ wc -l < output.txt
$ echo "appended" >> output.txt
```
**What you implement:** Parse `>`, `<`, `>>`. Open files, use `dup2()` to rewire stdin/stdout before `exec()`.
**System calls:** `open`, `dup2`, `close`

### Milestone 4: Pipes (`|`)
```
$ ls -la | grep txt
$ cat file.txt | sort | uniq | wc -l
```
**What you implement:** Create pipes between processes. Handle multi-stage pipelines (not just two).
**System calls:** `pipe`, `dup2`, `close`

### Milestone 5: Background Processes (`&`)
```
$ sleep 10 &
[1] 12345
$ ls
file1.txt
```
**What you implement:** If command ends with `&`, don't `wait()` — let it run in background. Track background PIDs.
**System calls:** `waitpid` with `WNOHANG`

### Milestone 6: Signal Handling (Ctrl+C, Ctrl+Z)
```
$ sleep 100
^C                  ← kills sleep, not the shell
$ sleep 100
^Z                  ← suspends sleep
[1]+ Stopped
$ fg                ← resumes it
```
**What you implement:** `sigaction` to handle SIGINT, SIGTSTP. Process groups for proper signal delivery.
**System calls:** `sigaction`, `setpgid`, `tcsetpgrp`, `kill`

### Milestone 7: Quote Handling + Environment Variables
```
$ echo "hello   world"
hello   world
$ export NAME=Vikram
$ echo $NAME
Vikram
$ echo "$NAME is learning"
Vikram is learning
```
**What you implement:** Proper tokenizer that handles quotes, variable expansion.
**System calls:** `setenv`, `getenv`

---

## What You Will Master After Building This

| Concept | Depth |
|---------|-------|
| **Processes** | How they're created (`fork`), replaced (`exec`), and managed (`wait`). You'll never wonder "what happens when I run a command" again. |
| **File Descriptors** | The universal I/O abstraction in Unix. Files, pipes, sockets — they're all fds. This concept carries into HTTP servers, databases, everything. |
| **Pipes & IPC** | How processes communicate. Pipes are the simplest form of inter-process communication, and you'll build them from scratch. |
| **Signals** | How the OS interrupts processes. Critical for understanding graceful shutdown in servers (SIGTERM), child process management, and debugging. |
| **Process Groups & Sessions** | How terminals manage foreground/background processes. Required knowledge for containers (P24) and process isolation. |
| **System Calls** | You'll call the raw Linux API directly — `fork`, `exec`, `pipe`, `dup2`, `open`, `wait`, `signal`. This is the interface between your code and the kernel. |
| **Parsing & Tokenization** | Splitting input into tokens, handling quotes, operator precedence (pipes before redirects). This skill reappears in every parser you'll ever write (HTTP, SQL, config files). |
| **C++ Systems Programming** | Manual resource management, POSIX API, error handling with `errno`, working without frameworks. |

### How This Connects to Your Backend Roadmap

```
P1 (Shell)
 ├── fork/exec/wait  ──→  P24 (Container Runtime) uses these same calls
 ├── File descriptors ──→  P6 (HTTP Server) — sockets are file descriptors
 ├── Pipes            ──→  P8 (Redis Clone) — epoll manages many fds at once
 ├── Signals          ──→  P17 (Production API) — graceful shutdown on SIGTERM
 ├── Parsing          ──→  P13 (SQL Database) — same tokenizer/parser pattern
 └── Process groups   ──→  P24 (Container Runtime) — PID namespaces
```

---

## How to Build It

1. Start with Milestone 1. Get a working REPL that runs `ls`, `pwd`, `echo`.
2. Add one milestone at a time. Test it, commit it, move on.
3. **Don't look at other shell implementations until you're stuck.** Struggling is where the learning happens.
4. When stuck, check `man 2 <syscall>` (e.g., `man 2 fork`) — the man pages are your best friend.
5. After each milestone, write a short note in this README about what you learned.

### Useful Commands While Building
```bash
man 2 fork          # Read about fork()
man 2 execvp        # Read about exec()
man 2 pipe          # Read about pipe()
man 2 dup2          # Read about dup2()
man 2 waitpid       # Read about wait()
man 7 signal        # Overview of all signals
strace bash -c "ls" # See what system calls bash makes (cheat sheet!)
```

---

## Compile & Run
```bash
make
./vsh               # Your shell!
```

---

## Progress

- [ ] Milestone 1: REPL + Execute commands
- [ ] Milestone 2: Built-in commands (cd, exit)
- [ ] Milestone 3: Redirection (>, >>, <)
- [ ] Milestone 4: Pipes (|)
- [ ] Milestone 5: Background processes (&)
- [ ] Milestone 6: Signal handling (Ctrl+C, Ctrl+Z)
- [ ] Milestone 7: Quotes + Environment variables
