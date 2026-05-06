// math.cpp — SOURCE FILE (definitions / implementations)
// This is WHERE the actual code lives.
// The linker connects calls in main.o to these definitions.

#include "math.h"  // Good practice: include own header to verify consistency

int add(int a, int b) {
    return a + b;
}

int subtract(int a, int b) {
    return a - b;
}

int multiply(int a, int b) {
    return a * b;
}

double divide(int a, int b) {
    // Note: we return double for precision
    // We'll learn about error handling later (what if b == 0?)
    if (b == 0) {
        return 0.0;  // Simplified — real code would handle this better
    }
    return static_cast<double>(a) / b;  // cast to double before division
}
