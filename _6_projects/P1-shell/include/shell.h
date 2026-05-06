#ifndef SHELL_H
#define SHELL_H

#include <string>
#include <vector>

// Milestone 3: Redirection info
//
// When the user types:  ls -la > output.txt
// We parse out the ">" and "output.txt" into this struct,
// and the command becomes just ["ls", "-la"]
//
struct Redirect {
    std::string input_file;    // "< file"   → redirect stdin from file
    std::string output_file;   // "> file"   → redirect stdout to file
    bool append = false;       // ">> file"  → append instead of overwrite
};

// Tokenize the input line into arguments
std::vector<std::string> tokenize(const std::string& line);

// Milestone 3: Extract redirection operators from args
// Modifies args in-place (removes >, >>, < and their filenames)
// Returns a Redirect struct with the file info
Redirect parse_redirects(std::vector<std::string>& args);

// Execute a single command (fork + exec + wait)
int execute(const std::vector<std::string>& args, const Redirect& redir = {}, bool background = false);

// Milestone 4: Split tokens by "|" into separate commands
// ["ls", "-la", "|", "grep", "txt"] → [["ls","-la"], ["grep","txt"]]
std::vector<std::vector<std::string>> split_pipeline(const std::vector<std::string>& args);

// Milestone 4: Execute a pipeline of commands connected by pipes
int execute_pipeline(std::vector<std::vector<std::string>>& cmds);

// Built-in commands (Milestone 2)
bool is_builtin(const std::string& cmd);
int run_builtin(const std::vector<std::string>& args);

#endif // SHELL_H
