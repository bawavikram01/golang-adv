//
// vsh — Vikram's Shell
//
// Milestone 1: REPL loop + execute simple commands
//
// YOUR TASK:
// 1. Print a prompt
// 2. Read a line of input
// 3. Tokenize it into arguments
// 4. Fork a child process
// 5. In the child: exec the command
// 6. In the parent: wait for the child
// 7. Repeat
//
// Start here. Read the README for the concepts.
//

#include <iostream>
#include <string>
#include <csignal>     // sigaction
#include <sys/wait.h>  // waitpid, WNOHANG
#include "shell.h"

// Milestone 5: Reap finished background processes
// Without this, finished background processes become "zombies" —
// they're done but their entry stays in the process table because
// nobody called waitpid() on them.
// We check for finished children at every prompt, without blocking.
static void reap_background() {
    int status;
    pid_t pid;
    // WNOHANG = don't block. Return 0 if no child has finished yet.
    while ((pid = waitpid(-1, &status, WNOHANG)) > 0) {
        std::cout << "[done] PID " << pid << " exited with status "
                  << WEXITSTATUS(status) << std::endl;
    }
}

int main() {
    std::string line;

    // Milestone 6: Shell ignores SIGINT (Ctrl+C) and SIGTSTP (Ctrl+Z)
    //
    // WHY: When you press Ctrl+C, the OS sends SIGINT to every process
    // in the foreground process group. Without this, our shell would die
    // alongside the child. We want ONLY the child to die.
    //
    // sigaction is preferred over signal() because its behavior is
    // well-defined across all Unix systems.
    //
    struct sigaction sa;
    sa.sa_handler = SIG_IGN;    // SIG_IGN = ignore the signal
    sigemptyset(&sa.sa_mask);   // don't block other signals
    sa.sa_flags = 0;
    sigaction(SIGINT, &sa, nullptr);   // Ctrl+C won't kill the shell
    sigaction(SIGTSTP, &sa, nullptr);  // Ctrl+Z won't stop the shell
    sigaction(SIGTTOU, &sa, nullptr);  // tcsetpgrp won't stop the shell
    sigaction(SIGTTIN, &sa, nullptr);  // background read won't stop the shell

    while (true) {
        // Milestone 5: Reap any finished background processes before prompt
        reap_background();

        // Step 1: Print prompt
        std::cout << "vsh> ";

        // Step 2: Read input
        if (!std::getline(std::cin, line)) {
            // Handle Ctrl+D (EOF)
            std::cout << std::endl;
            break;
        }

        // Skip empty lines
        if (line.empty()) {
            continue;
        }

        // Step 3: Tokenize
        std::vector<std::string> args = tokenize(line);
        if (args.empty()) {
            continue;
        }

        // Handle "exit" built-in
        if (args[0] == "exit") {
            break;
        }

        // Milestone 5: Check if command ends with "&"
        // If so, remove it and run the command in the background (don't wait)
        bool background = false;
        if (!args.empty() && args.back() == "&") {
            background = true;
            args.pop_back();
            if (args.empty()) continue;
        }

        // Milestone 4: Check if this is a pipeline (contains |)
        std::vector<std::vector<std::string>> pipeline = split_pipeline(args);

        if (pipeline.size() > 1) {
            // It's a pipeline — handle pipes
            execute_pipeline(pipeline);
            continue;
        }

        // Single command — handle redirection + builtins as before
        // Milestone 3: Parse redirection before anything else
        Redirect redir = parse_redirects(args);
        if (args.empty()) continue;

        // Milestone 2: Handle built-in commands BEFORE fork()
        // cd, export, pwd must run in the shell process itself
        if (is_builtin(args[0])) {
            run_builtin(args);
            continue;
        }

        // Step 4-6: Execute external commands (fork + exec + wait)
        execute(args, redir, background);
    }

    return 0;
}
