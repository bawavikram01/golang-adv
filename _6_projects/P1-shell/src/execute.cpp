//
// execute.cpp — The heart of the shell
//
// This is where you implement fork() + exec() + wait()
//
// READ THIS FIRST:
//
//   fork() returns TWICE:
//     - In the parent: returns the child's PID (> 0)
//     - In the child:  returns 0
//
//   So you write:
//     pid_t pid = fork();
//     if (pid == 0) {
//         // I am the child — run the command
//     } else {
//         // I am the parent — wait for child
//     }
//
//   execvp() replaces the current process with a new program.
//   If it returns, something went wrong.
//
//   waitpid() blocks until the child exits.
//

#include "shell.h"

#include <sys/wait.h>  // waitpid
#include <unistd.h>    // fork, execvp
#include <fcntl.h>     // open, O_WRONLY, O_CREAT, etc.
#include <csignal>     // signal, SIG_DFL
#include <cstring>     // for converting strings
#include <iostream>

// Milestone 3: Apply redirections in the child process
//
// HOW THIS WORKS:
//   open("file") returns a new fd (e.g., 3)
//   dup2(3, STDOUT_FILENO) makes fd 1 point to the same file as fd 3
//   close(3) — we don't need fd 3 anymore, fd 1 is enough
//   Now anything written to stdout goes to the file
//
// IMPORTANT: This runs AFTER fork(), in the child, BEFORE exec().
// The parent's file descriptors are not affected.
//
static bool apply_redirects(const Redirect& redir) {
    // Input redirection: cmd < file
    if (!redir.input_file.empty()) {
        int fd = open(redir.input_file.c_str(), O_RDONLY);
        if (fd < 0) {
            perror("vsh");
            return false;
        }
        dup2(fd, STDIN_FILENO);   // stdin now reads from file
        close(fd);                // cleanup the extra fd
    }

    // Output redirection: cmd > file  or  cmd >> file
    if (!redir.output_file.empty()) {
        int flags = O_WRONLY | O_CREAT;
        if (redir.append) {
            flags |= O_APPEND;    // >> : append to end
        } else {
            flags |= O_TRUNC;     // >  : overwrite (truncate)
        }
        int fd = open(redir.output_file.c_str(), flags, 0644);
        if (fd < 0) {
            perror("vsh");
            return false;
        }
        dup2(fd, STDOUT_FILENO);  // stdout now writes to file
        close(fd);
    }

    return true;
}

int execute(const std::vector<std::string>& args, const Redirect& redir, bool background) {
    // Convert vector<string> to char*[] for execvp
    // execvp needs a C-style array: {"ls", "-la", nullptr}
    std::vector<char*> argv;
    for (const auto& arg : args) {
        argv.push_back(const_cast<char*>(arg.c_str()));
    }
    argv.push_back(nullptr);

    // --- YOUR JOURNEY STARTS HERE ---

    // Step 1: Fork a child process
    pid_t pid = fork();

    if (pid < 0) {
        // fork failed
        perror("vsh: fork");
        return -1;
    }

    if (pid == 0) {
        // ---- CHILD PROCESS ----

        // Milestone 6: Restore default signal handlers in child
        // The shell ignores SIGINT/SIGTSTP, but children should NOT.
        // When the user presses Ctrl+C, the child should die normally.
        signal(SIGINT, SIG_DFL);
        signal(SIGTSTP, SIG_DFL);

        // Put child in its own process group
        // This separates it from the shell so we can control
        // which process group is "foreground" (receives signals)
        setpgid(0, 0);

        // Milestone 3: Rewire stdin/stdout BEFORE exec
        // This is the magic — the program we exec has no idea
        // its I/O has been redirected. It just reads/writes normally.
        if (!apply_redirects(redir)) {
            _exit(1);
        }

        // Replace this process with the command
        // execvp searches PATH for the command automatically
        execvp(argv[0], argv.data());

        // If we get here, exec failed (command not found)
        perror("vsh");
        _exit(1);  // Use _exit in child, not exit()
    }

    // ---- PARENT PROCESS ----
    if (background) {
        // Milestone 5: Don't wait — let it run in the background
        std::cout << "[bg] PID " << pid << std::endl;
        return 0;
    }

    // Milestone 6: Give the child's process group control of the terminal
    // This means Ctrl+C/Ctrl+Z go to the child, not the shell
    setpgid(pid, pid);  // ensure child is in its own group (race condition safety)
    tcsetpgrp(STDIN_FILENO, pid);  // child group gets the terminal

    // Foreground: Wait for the child to finish (or be stopped by Ctrl+Z)
    int status;
    waitpid(pid, &status, WUNTRACED);  // WUNTRACED: also return if child is stopped

    // Take the terminal back for the shell
    tcsetpgrp(STDIN_FILENO, getpgrp());

    // Check if child was stopped (Ctrl+Z)
    if (WIFSTOPPED(status)) {
        std::cout << std::endl << "[stopped] PID " << pid << std::endl;
        // TODO: Track stopped jobs for fg/bg commands
    }

    return WEXITSTATUS(status);
}

// ============================================================
// Milestone 4: Pipeline execution
// ============================================================
//
// HOW PIPES WORK:
//
//   pipe(pipefd) creates a one-way data channel:
//     pipefd[0] = read end   (data comes OUT here)
//     pipefd[1] = write end  (data goes IN here)
//
//   For "ls | grep txt":
//
//     1. Create a pipe
//     2. Fork child 1 (ls):   dup2(pipefd[1], STDOUT) → stdout goes into pipe
//     3. Fork child 2 (grep): dup2(pipefd[0], STDIN)  → stdin comes from pipe
//     4. Close pipe fds in parent (CRITICAL — or grep never gets EOF)
//     5. Wait for both children
//
//   For N commands, you need N-1 pipes.
//
//   THE #1 MISTAKE: Forgetting to close pipe fds.
//   If the write end of a pipe is still open anywhere, the reader
//   will block forever waiting for more data (never gets EOF).
//   Close ALL pipe fds you don't need, in EVERY process.
//
int execute_pipeline(std::vector<std::vector<std::string>>& cmds) {
    int n = cmds.size();
    // We need n-1 pipes. Each pipe is int[2].
    std::vector<int> pipefds(2 * (n - 1));

    // Create all pipes upfront
    for (int i = 0; i < n - 1; i++) {
        if (pipe(&pipefds[2 * i]) < 0) {
            perror("vsh: pipe");
            return -1;
        }
    }

    // Fork a child for each command
    std::vector<pid_t> pids;
    for (int i = 0; i < n; i++) {
        // Parse redirections for this command (handles > at end of pipeline)
        Redirect redir = parse_redirects(cmds[i]);

        // Build argv for execvp
        std::vector<char*> argv;
        for (const auto& arg : cmds[i]) {
            argv.push_back(const_cast<char*>(arg.c_str()));
        }
        argv.push_back(nullptr);

        pid_t pid = fork();
        if (pid < 0) {
            perror("vsh: fork");
            return -1;
        }

        if (pid == 0) {
            // ---- CHILD i ----

            // Milestone 6: Restore signals + own process group
            signal(SIGINT, SIG_DFL);
            signal(SIGTSTP, SIG_DFL);

            // If NOT the first command: read stdin from previous pipe
            if (i > 0) {
                dup2(pipefds[2 * (i - 1)], STDIN_FILENO);
            }

            // If NOT the last command: write stdout to next pipe
            if (i < n - 1) {
                dup2(pipefds[2 * i + 1], STDOUT_FILENO);
            }

            // Close ALL pipe fds in child — we've already dup2'd what we need
            // This is critical. If we don't close the write end of a pipe,
            // the reader will never get EOF and will hang forever.
            for (size_t j = 0; j < pipefds.size(); j++) {
                close(pipefds[j]);
            }

            // Apply any file redirections (> or <) on top of pipe redirections
            apply_redirects(redir);

            execvp(argv[0], argv.data());
            perror("vsh");
            _exit(1);
        }

        pids.push_back(pid);
    }

    // ---- PARENT ----
    // Close ALL pipe fds in parent — the children have their own copies
    for (size_t i = 0; i < pipefds.size(); i++) {
        close(pipefds[i]);
    }

    // Wait for ALL children
    int status = 0;
    for (pid_t pid : pids) {
        waitpid(pid, &status, 0);
    }

    // Return exit status of the last command (like bash does)
    return WEXITSTATUS(status);
}
