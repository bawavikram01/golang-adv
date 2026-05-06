// ============================================================
// Lesson 1b: Multi-file Compilation (How Real Programs Work)
// ============================================================
// Compile: g++ -std=c++20 -Wall main.cpp math.cpp -o calculator
// 
// What happens:
// 1. main.cpp → main.o  (compiled independently)
// 2. math.cpp → math.o  (compiled independently)
// 3. main.o + math.o → calculator (linked together)
//
// main.cpp sees math.h (declarations) so it knows what functions EXIST
// But the actual code lives in math.cpp

#include <iostream>
#include "math.h"  // "" means look in current directory first, then system paths
                   // <> means look in system paths only

int main() {
    int a = 10, b = 3;
    
    std::cout << a << " + " << b << " = " << add(a, b) << '\n';
    std::cout << a << " - " << b << " = " << subtract(a, b) << '\n';
    std::cout << a << " * " << b << " = " << multiply(a, b) << '\n';
    std::cout << a << " / " << b << " = " << divide(a, b) << '\n';
    
    return 0;
}
