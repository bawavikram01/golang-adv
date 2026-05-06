// ============================================================
// Lesson 5: Functions — Everything You Need to Know
// ============================================================
// Compile: g++ -std=c++20 -Wall -Wextra functions.cpp -o functions && ./functions
//
// Functions are the building blocks of C++ programs.
// Understanding pass-by-value, pass-by-reference, and overloading is essential.

#include <iostream>
#include <string>
#include <vector>
#include <array>

// ================================================================
// DECLARATION vs DEFINITION
// ================================================================

// DECLARATION (prototype) — tells compiler it exists
int add(int a, int b);          // Just the signature, no body
void print_array(const int* arr, size_t size);

// DEFINITION — the actual implementation (can also serve as declaration)
int add(int a, int b) {
    return a + b;
}

// ================================================================
// PASS BY VALUE — function gets a COPY
// ================================================================

void increment_value(int n) {
    n += 1;  // Modifies local copy only!
    std::cout << "  Inside function: n = " << n << '\n';
}

// ================================================================
// PASS BY REFERENCE — function gets the ORIGINAL variable
// ================================================================

void increment_ref(int& n) {    // & makes it a reference
    n += 1;  // Modifies the original!
    std::cout << "  Inside function: n = " << n << '\n';
}

// ================================================================
// PASS BY CONST REFERENCE — read-only access, no copy
// ================================================================
// Use for large objects you don't want to modify or copy

void print_vector(const std::vector<int>& vec) {
    // vec.push_back(1);  // ERROR: vec is const
    for (const auto& v : vec) {
        std::cout << v << ' ';
    }
    std::cout << '\n';
}

// ================================================================
// PASS BY POINTER — function gets memory address
// ================================================================

void increment_ptr(int* p) {
    if (p != nullptr) {  // Always null-check pointers!
        *p += 1;  // Dereference and modify
    }
}

// Pointer to const: can't modify through this pointer
void read_only_ptr(const int* p) {
    // *p = 10;  // ERROR: can't write through pointer-to-const
    std::cout << "  Value: " << *p << '\n';
}

// ================================================================
// RETURNING VALUES
// ================================================================

// Return by value (most common — compiler optimizes via RVO)
std::vector<int> make_vector(int n) {
    std::vector<int> result;
    for (int i = 0; i < n; ++i) {
        result.push_back(i * i);
    }
    return result;  // NOT expensive! RVO eliminates the copy
}

// NEVER return a reference to a local variable!
// int& bad_function() {
//     int local = 42;
//     return local;  // DANGLING REFERENCE! UB!
// }

// ================================================================
// DEFAULT ARGUMENTS
// ================================================================

void greet(const std::string& name, const std::string& greeting = "Hello") {
    std::cout << greeting << ", " << name << "!\n";
}
// Rules: defaults go right-to-left (no gaps allowed)
// void bad(int a = 1, int b, int c = 3);  // ERROR: gap at b

// ================================================================
// FUNCTION OVERLOADING — same name, different parameters
// ================================================================

int square(int x) { return x * x; }
double square(double x) { return x * x; }
// int square(int x) { ... }  // ERROR: can't differ only by return type

// Overload resolution rules (simplified):
// 1. Exact match
// 2. Promotion (char → int, float → double)
// 3. Standard conversion (int → double)
// 4. User-defined conversion

// ================================================================
// CONSTEXPR FUNCTIONS — computed at compile time
// ================================================================

constexpr int factorial(int n) {
    if (n <= 1) return 1;
    return n * factorial(n - 1);
}
// If called with compile-time args, computed at compile time:
// constexpr int f5 = factorial(5);  // Computed at compile time! No runtime cost

// If called with runtime args, runs at runtime (still valid):
// int x; std::cin >> x; int fx = factorial(x);  // Runs at runtime

// ================================================================
// INLINE FUNCTIONS
// ================================================================

// Suggestion to compiler: replace call with function body (avoids call overhead)
// Modern compilers mostly decide this themselves regardless of keyword
inline int max_of(int a, int b) {
    return (a > b) ? a : b;
}

// ================================================================
// FUNCTION POINTERS — functions ARE just addresses in memory
// ================================================================

int multiply(int a, int b) { return a * b; }
int subtract(int a, int b) { return a - b; }

// Type: int(*)(int, int) — pointer to function taking 2 ints, returning int
using BinaryOp = int(*)(int, int);  // Type alias for readability

int apply(BinaryOp op, int a, int b) {
    return op(a, b);
}

// ================================================================
// MAIN
// ================================================================

int main() {
    std::cout << "=== Pass by Value ===\n";
    int x = 10;
    std::cout << "Before: x = " << x << '\n';
    increment_value(x);
    std::cout << "After:  x = " << x << " (unchanged! function got a copy)\n";
    
    std::cout << "\n=== Pass by Reference ===\n";
    std::cout << "Before: x = " << x << '\n';
    increment_ref(x);
    std::cout << "After:  x = " << x << " (changed! function modified original)\n";
    
    std::cout << "\n=== Pass by Pointer ===\n";
    std::cout << "Before: x = " << x << '\n';
    increment_ptr(&x);  // Pass address with &
    std::cout << "After:  x = " << x << " (changed via pointer)\n";
    
    std::cout << "\n=== Returning Vectors (RVO) ===\n";
    auto vec = make_vector(5);  // No copy! RVO kicks in
    std::cout << "Vector: ";
    print_vector(vec);
    
    std::cout << "\n=== Default Arguments ===\n";
    greet("Vikram");                    // Uses default "Hello"
    greet("Vikram", "Good morning");    // Overrides default
    
    std::cout << "\n=== Overloading ===\n";
    std::cout << "square(5) = " << square(5) << " (calls int version)\n";
    std::cout << "square(2.5) = " << square(2.5) << " (calls double version)\n";
    
    std::cout << "\n=== constexpr ===\n";
    constexpr int f10 = factorial(10);  // Computed at COMPILE TIME
    std::cout << "10! = " << f10 << " (computed at compile time)\n";
    
    std::cout << "\n=== Function Pointers ===\n";
    std::cout << "apply(multiply, 6, 7) = " << apply(multiply, 6, 7) << '\n';
    std::cout << "apply(subtract, 10, 3) = " << apply(subtract, 10, 3) << '\n';
    
    // Store in a variable:
    BinaryOp op = multiply;
    std::cout << "op(4, 5) = " << op(4, 5) << '\n';
    op = add;
    std::cout << "op(4, 5) = " << op(4, 5) << '\n';
    
    // ================================================================
    // WHEN TO USE WHAT:
    // ================================================================
    std::cout << "\n=== Guidelines ===\n";
    std::cout << "Pass by value:      small types (int, double, char, bool, pointers)\n";
    std::cout << "Pass by const ref:  large types you only READ (string, vector, etc.)\n";
    std::cout << "Pass by ref:        when you need to MODIFY the argument\n";
    std::cout << "Pass by pointer:    when nullptr is a valid argument (optional param)\n";
    std::cout << "Return by value:    almost always (RVO makes it free)\n";
    
    return 0;
}

void print_array(const int* arr, size_t size) {
    for (size_t i = 0; i < size; ++i) {
        std::cout << arr[i] << ' ';
    }
    std::cout << '\n';
}
