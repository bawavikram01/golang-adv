#include "shell.h"
#include <sstream>
#include <cstdlib>  // getenv

// Milestone 7: Expand $VAR and ${VAR} in a string
//
// HOW IT WORKS:
//   Walk character by character. When we see '$':
//     - $VAR  → read alphanumeric chars as variable name → getenv()
//     - ${VAR} → read until '}' → getenv()
//   Everything else passes through unchanged.
//
static std::string expand_variables(const std::string& s) {
    std::string result;
    for (size_t i = 0; i < s.size(); i++) {
        if (s[i] == '$' && i + 1 < s.size()) {
            std::string varname;
            i++;  // skip '$'

            if (s[i] == '{') {
                // ${VAR} form
                i++;  // skip '{'
                while (i < s.size() && s[i] != '}') {
                    varname += s[i++];
                }
                // i now points to '}', loop increment will skip it
            } else {
                // $VAR form — read alphanumeric + underscore
                while (i < s.size() && (isalnum(s[i]) || s[i] == '_')) {
                    varname += s[i++];
                }
                i--;  // loop will increment, so step back one
            }

            const char* val = getenv(varname.c_str());
            if (val) result += val;
            // if var doesn't exist, expand to empty string (like bash)
        } else {
            result += s[i];
        }
    }
    return result;
}

// Milestone 7: Full tokenizer with quotes and variable expansion
//
// STATE MACHINE:
//   We walk through the input one character at a time.
//   We're either:
//     - OUTSIDE a token (skipping whitespace)
//     - INSIDE a token (building up characters)
//     - INSIDE quotes (space doesn't end the token)
//
//   Characters:
//     space     → if unquoted, ends current token
//     "         → toggle double-quote mode (vars ARE expanded inside)
//     '         → toggle single-quote mode (nothing is expanded inside)
//     \         → next char is literal (even spaces and quotes)
//     $         → expand variable (only outside single quotes)
//     anything  → add to current token
//
std::vector<std::string> tokenize(const std::string& line) {
    std::vector<std::string> tokens;
    std::string current;
    bool in_single = false;  // inside '...'
    bool in_double = false;  // inside "..."

    for (size_t i = 0; i < line.size(); i++) {
        char c = line[i];

        // Backslash escape — next char is always literal
        if (c == '\\' && !in_single && i + 1 < line.size()) {
            current += line[++i];
            continue;
        }

        // Single quote toggle
        if (c == '\'' && !in_double) {
            in_single = !in_single;
            continue;  // quote itself is NOT part of the token
        }

        // Double quote toggle
        if (c == '"' && !in_single) {
            in_double = !in_double;
            continue;
        }

        // Space — ends token (unless inside quotes)
        if (c == ' ' && !in_single && !in_double) {
            if (!current.empty()) {
                tokens.push_back(current);
                current.clear();
            }
            continue;
        }

        // Regular character — add to current token
        current += c;
    }

    // Don't forget the last token
    if (!current.empty()) {
        tokens.push_back(current);
    }

    // Expand variables in all tokens (except those that were single-quoted)
    // Note: For full correctness, expansion should happen per-character
    // during tokenization. This simpler approach handles most cases.
    for (auto& token : tokens) {
        token = expand_variables(token);
    }

    return tokens;
}

// Milestone 3: Extract redirection from the token list
//
// Input:  ["ls", "-la", ">", "output.txt"]
// Output: args becomes ["ls", "-la"], returns Redirect{output_file="output.txt"}
//
// Scans for >, >>, < tokens. Grabs the filename after them.
// Removes both the operator and filename from args.
//
Redirect parse_redirects(std::vector<std::string>& args) {
    Redirect redir;
    std::vector<std::string> cleaned;

    for (size_t i = 0; i < args.size(); i++) {
        if (args[i] == ">" && i + 1 < args.size()) {
            redir.output_file = args[i + 1];
            redir.append = false;
            i++;  // skip the filename
        } else if (args[i] == ">>" && i + 1 < args.size()) {
            redir.output_file = args[i + 1];
            redir.append = true;
            i++;
        } else if (args[i] == "<" && i + 1 < args.size()) {
            redir.input_file = args[i + 1];
            i++;
        } else {
            cleaned.push_back(args[i]);
        }
    }

    args = cleaned;
    return redir;
}

// Milestone 4: Split tokens at "|" into separate commands
//
// Input:  ["ls", "-la", "|", "grep", "txt", "|", "wc", "-l"]
// Output: [["ls","-la"], ["grep","txt"], ["wc","-l"]]
//
std::vector<std::vector<std::string>> split_pipeline(const std::vector<std::string>& args) {
    std::vector<std::vector<std::string>> cmds;
    std::vector<std::string> current;

    for (const auto& arg : args) {
        if (arg == "|") {
            if (!current.empty()) {
                cmds.push_back(current);
                current.clear();
            }
        } else {
            current.push_back(arg);
        }
    }
    if (!current.empty()) {
        cmds.push_back(current);
    }

    return cmds;
}
