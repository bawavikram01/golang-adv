//
// builtins.cpp — Commands that run inside the shell process
//
// WHY BUILTINS EXIST:
//
//   Most commands (ls, grep, cat) run as child processes.
//   But some commands MUST run in the shell itself:
//
//   cd /home   ← Must change the SHELL's directory, not a child's
//   exit       ← Must stop the SHELL, not a child
//   export X=1 ← Must set a variable in the SHELL's environment
//
//   If "cd" ran in a child process:
//     fork() → child calls chdir("/home") → child exits
//     Parent (shell) is still in the old directory. Nothing changed.
//
//   That's why we handle these BEFORE fork(), in main.cpp.
//

#include "shell.h"

#include <unistd.h>    // chdir, getenv
#include <iostream>
#include <cstdlib>     // setenv

bool is_builtin(const std::string& cmd) {
    return cmd == "cd" || cmd == "exit" || cmd == "export" || cmd == "pwd";
}

int run_builtin(const std::vector<std::string>& args) {

    // --- cd ---
    if (args[0] == "cd") {
        // cd with no args → go to HOME directory (just like bash)
        std::string target;
        if (args.size() < 2) {
            const char* home = getenv("HOME");
            if (!home) {
                std::cerr << "vsh: cd: HOME not set" << std::endl;
                return 1;
            }
            target = home;
        } else {
            target = args[1];
        }

        // chdir() is the syscall that actually changes the directory
        // It modifies THIS process's working directory
        if (chdir(target.c_str()) != 0) {
            perror("vsh: cd");
            return 1;
        }
        return 0;
    }

    // --- pwd (builtin version) ---
    if (args[0] == "pwd") {
        char cwd[1024];
        if (getcwd(cwd, sizeof(cwd))) {
            std::cout << cwd << std::endl;
        } else {
            perror("vsh: pwd");
            return 1;
        }
        return 0;
    }

    // --- export ---
    if (args[0] == "export") {
        if (args.size() < 2) {
            std::cerr << "vsh: export: usage: export NAME=VALUE" << std::endl;
            return 1;
        }
        // Parse "NAME=VALUE"
        std::string arg = args[1];
        size_t eq = arg.find('=');
        if (eq == std::string::npos) {
            std::cerr << "vsh: export: invalid format (use NAME=VALUE)" << std::endl;
            return 1;
        }
        std::string name = arg.substr(0, eq);
        std::string value = arg.substr(eq + 1);
        setenv(name.c_str(), value.c_str(), 1);
        return 0;
    }

    return 1;
}
