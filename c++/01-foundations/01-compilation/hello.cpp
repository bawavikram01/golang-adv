// ============================================================
// Lesson 1: Understanding the Compilation Pipeline
// ============================================================
// Compile this file step by step to see each stage:
//
//   g++ -E hello.cpp -o hello.i    # Preprocessing (see expanded output)
//   g++ -S hello.cpp -o hello.s    # Compilation (see assembly)
//   g++ -c hello.cpp -o hello.o    # Assembly (create object file)
//   g++ hello.o -o hello           # Linking (create executable)
//
// Or all at once:
//   g++ -std=c++20 -Wall -Wextra -g hello.cpp -o hello
//
// Then run:
//   ./hello

#include <iostream>  // This gets copy-pasted by preprocessor (thousands of lines!)

// This is a DEFINITION of main. Every C++ program needs exactly one.
int main() {
    // std::cout is defined in <iostream>
    // << is an overloaded operator (you'll learn this later)
    // std::endl flushes the buffer and adds newline
    // '\n' is preferred over std::endl (faster, no flush)
    
    std::cout << "Hello, World!\n";
    std::cout << "You are now on the path to mastering C++.\n";
    
    // return 0 means "success" to the OS
    // In main(), return 0 is implicit if omitted (only in main!)
    return 0;
}
