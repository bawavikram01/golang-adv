// ============================================================
// Phase 1 EXERCISES — Do these to solidify your understanding
// ============================================================
// Each exercise builds on the lessons. Write your solutions in
// separate files (ex01.cpp, ex02.cpp, etc.) in this directory.
//
// Compile any exercise:
//   g++ -std=c++20 -Wall -Wextra exNN.cpp -o exNN && ./exNN

// ================================================================
// EXERCISE 1: Compilation Pipeline
// ================================================================
// 1a. Run `g++ -E hello.cpp` and look at the output.
//     How many lines does <iostream> expand to? (hint: thousands!)
//
// 1b. Run `g++ -S hello.cpp` and open hello.s
//     Can you find the string "Hello, World!" in the assembly?
//
// 1c. Create a program with an intentional linker error:
//     - Declare a function in a header but don't define it
//     - Call it from main
//     - Observe the "undefined reference" error
//     - Now define it and fix the error

// ================================================================
// EXERCISE 2: Types
// ================================================================
// 2a. Write a program that prints sizeof() for every fundamental type.
//
// 2b. Demonstrate integer overflow:
//     - Assign INT_MAX to an int, add 1, print the result
//     - Assign 0 to an unsigned int, subtract 1, print the result
//     - Explain why they differ (signed = UB, unsigned = defined wrap)
//
// 2c. Write a program that demonstrates the signed/unsigned comparison bug:
//     - Compare (unsigned)1 > (int)-1
//     - Explain why the result is surprising
//
// 2d. Use brace initialization {} to trigger a compile error from narrowing:
//     - Try: int x{3.14};

// ================================================================
// EXERCISE 3: Bitwise Operations
// ================================================================
// 3a. Write a function that counts the number of set bits (1s) in an integer.
//     Signature: int count_bits(uint32_t n);
//     Test with: 0, 1, 255, 0xFFFFFFFF
//
// 3b. Write a function that checks if a number is a power of 2.
//     Don't use loops or division. (Hint: n & (n-1))
//
// 3c. Write a function that swaps two integers using XOR (no temp variable).
//     void xor_swap(int& a, int& b);
//
// 3d. Implement a simple "permissions" system using bit flags:
//     - READ = 1, WRITE = 2, EXECUTE = 4
//     - Write functions: grant(flags, perm), revoke(flags, perm), has(flags, perm)

// ================================================================
// EXERCISE 4: Functions
// ================================================================
// 4a. Write a function that takes a vector by const reference and returns
//     a new vector with each element doubled.
//
// 4b. Write overloaded "print" functions that handle:
//     - int, double, std::string, std::vector<int>
//
// 4c. Write a constexpr function that computes fibonacci(n).
//     Verify it works at compile time: constexpr auto f10 = fib(10);
//
// 4d. Write a higher-order function:
//     int apply_twice(int (*f)(int), int x);  // returns f(f(x))
//     Test with a function that doubles its argument.

// ================================================================
// EXERCISE 5: Pointers
// ================================================================
// 5a. Write a function void swap(int* a, int* b) that swaps using pointers.
//     Write another void swap(int& a, int& b) using references.
//     Which is cleaner at the call site?
//
// 5b. Write a function that finds the maximum element in a C-style array:
//     int* find_max(int* arr, size_t size);
//     Return a POINTER to the max element. Handle empty array (return nullptr).
//
// 5c. Implement a simple dynamic array manually:
//     - Allocate with new[]
//     - When full, allocate a bigger array, copy elements, delete old one
//     - This is essentially what std::vector does!
//     - Remember to free memory when done
//
// 5d. Demonstrate array decay:
//     - Write a function that takes int arr[] and sizeof(arr) inside it
//     - Compare with sizeof in main() to show size info is lost

// ================================================================
// EXERCISE 6: Mini Project — Command Line Calculator
// ================================================================
// Build a calculator that:
// - Takes 3 command-line arguments: num1 operator num2
//   Example: ./calc 10 + 5
// - Supports: + - * / %
// - Handles division by zero gracefully
// - Uses function pointers to dispatch operations
// - Split into: calc.cpp (main), operations.h, operations.cpp
//
// Concepts practiced:
// - argc/argv (command-line arguments)
// - String to number conversion (std::stod)
// - Function pointers
// - Multi-file compilation
// - Error handling

int main() {
    // This file is just for reading. Write solutions in separate files!
    return 0;
}
