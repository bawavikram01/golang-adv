// ============================================================
// Lesson 3: Operators — Deep Understanding
// ============================================================
// Compile: g++ -std=c++20 -Wall -Wextra operators.cpp -o operators && ./operators
//
// Operators are just syntactic sugar for function calls.
// You MUST understand precedence, associativity, and bitwise ops.

#include <iostream>
#include <cstdint>
#include <bitset>  // For printing binary representation

int main() {
    // ================================================================
    // ARITHMETIC OPERATORS
    // ================================================================
    
    std::cout << "=== Arithmetic ===\n";
    
    int a = 17, b = 5;
    std::cout << "17 / 5 = " << a / b << " (integer division truncates toward zero)\n";
    std::cout << "17 % 5 = " << a % b << " (modulo/remainder)\n";
    std::cout << "-17 / 5 = " << (-17) / 5 << " (truncates toward zero in C++11+)\n";
    std::cout << "-17 % 5 = " << (-17) % 5 << " (result has sign of dividend)\n";
    
    // DANGER: integer overflow is UNDEFINED BEHAVIOR for signed types!
    // int overflow = INT_MAX + 1;  // UB! Compiler can do ANYTHING
    
    // But unsigned overflow is DEFINED (wraps around):
    unsigned int u = 0;
    u = u - 1;  // Wraps to UINT_MAX (4294967295)
    std::cout << "unsigned 0 - 1 = " << u << " (defined wrap-around)\n";
    
    // ================================================================
    // BITWISE OPERATORS (Critical for systems programming!)
    // ================================================================
    
    std::cout << "\n=== Bitwise Operations ===\n";
    
    uint8_t x = 0b1100'1010;  // 202 decimal
    uint8_t y = 0b1010'0101;  // 165 decimal
    
    std::cout << "x        = " << std::bitset<8>(x) << " (" << +x << ")\n";
    std::cout << "y        = " << std::bitset<8>(y) << " (" << +y << ")\n";
    std::cout << "x & y    = " << std::bitset<8>(x & y) << " (AND: both bits 1)\n";
    std::cout << "x | y    = " << std::bitset<8>(x | y) << " (OR: either bit 1)\n";
    std::cout << "x ^ y    = " << std::bitset<8>(x ^ y) << " (XOR: bits differ)\n";
    std::cout << "~x       = " << std::bitset<8>(static_cast<uint8_t>(~x)) << " (NOT: flip all)\n";
    std::cout << "x << 2   = " << std::bitset<8>(static_cast<uint8_t>(x << 2)) << " (left shift)\n";
    std::cout << "x >> 2   = " << std::bitset<8>(static_cast<uint8_t>(x >> 2)) << " (right shift)\n";
    
    // COMMON BIT MANIPULATION PATTERNS (used in systems programming daily):
    std::cout << "\n=== Bit Manipulation Patterns ===\n";
    
    uint8_t flags = 0b0000'0000;
    
    // Set bit n:     flags |= (1 << n)
    flags |= (1 << 3);  // Set bit 3
    std::cout << "Set bit 3:    " << std::bitset<8>(flags) << '\n';
    
    // Clear bit n:   flags &= ~(1 << n)
    flags |= (1 << 5);  // Set bit 5 first
    flags &= ~(1 << 5); // Now clear bit 5
    std::cout << "Clear bit 5:  " << std::bitset<8>(flags) << '\n';
    
    // Toggle bit n:  flags ^= (1 << n)
    flags ^= (1 << 3);  // Toggle bit 3 (was 1, now 0)
    std::cout << "Toggle bit 3: " << std::bitset<8>(flags) << '\n';
    
    // Check bit n:   (flags >> n) & 1
    flags = 0b0010'1000;
    bool bit3 = (flags >> 3) & 1;  // true
    bool bit2 = (flags >> 2) & 1;  // false
    std::cout << "flags = " << std::bitset<8>(flags) << '\n';
    std::cout << "Bit 3 set? " << bit3 << ", Bit 2 set? " << bit2 << '\n';
    
    // Power of 2 check:  (n & (n-1)) == 0
    for (int n : {1, 2, 3, 4, 5, 8, 16, 15}) {
        bool is_pow2 = (n > 0) && ((n & (n - 1)) == 0);
        std::cout << n << " is power of 2? " << is_pow2 << '\n';
    }
    
    // ================================================================
    // LOGICAL OPERATORS & SHORT-CIRCUIT EVALUATION
    // ================================================================
    
    std::cout << "\n=== Short-Circuit Evaluation ===\n";
    
    int val = 0;
    // && short-circuits: if left is false, right is NEVER evaluated
    if (val != 0 && (10 / val > 2)) {
        // The division never happens because val != 0 is false!
        std::cout << "This is safe thanks to short-circuit!\n";
    }
    
    // || short-circuits: if left is true, right is NEVER evaluated
    // This is used CONSTANTLY in real code for guard checks
    
    // ================================================================
    // OPERATOR PRECEDENCE (simplified, most important ones)
    // ================================================================
    
    std::cout << "\n=== Precedence (High to Low) ===\n";
    std::cout << "1.  :: (scope resolution)\n";
    std::cout << "2.  a++ a-- (postfix), () [] . ->\n";
    std::cout << "3.  ++a --a +a -a ! ~ (type) * & sizeof (prefix/unary)\n";
    std::cout << "4.  * / % (multiplicative)\n";
    std::cout << "5.  + - (additive)\n";
    std::cout << "6.  << >> (shift)\n";
    std::cout << "7.  < <= > >= (relational)\n";
    std::cout << "8.  == != (equality)\n";
    std::cout << "9.  & (bitwise AND)\n";
    std::cout << "10. ^ (bitwise XOR)\n";
    std::cout << "11. | (bitwise OR)\n";
    std::cout << "12. && (logical AND)\n";
    std::cout << "13. || (logical OR)\n";
    std::cout << "14. ?: (ternary)\n";
    std::cout << "15. = += -= *= etc. (assignment)\n";
    std::cout << "16. , (comma)\n";
    
    // COMMON PRECEDENCE TRAP:
    int v = 5;
    // if (v & 0x0F == 0)  // WRONG! == has higher precedence than &
    // if ((v & 0x0F) == 0) // CORRECT! Always parenthesize bitwise ops
    std::cout << "\nALWAYS parenthesize bitwise operations in conditions!\n";
    
    // ================================================================
    // INCREMENT/DECREMENT: prefix vs postfix
    // ================================================================
    
    std::cout << "\n=== Prefix vs Postfix ===\n";
    int c = 5;
    std::cout << "c = " << c << '\n';
    std::cout << "++c = " << ++c << " (increment THEN use, c is now " << c << ")\n";
    std::cout << "c++ = " << c++ << " (use THEN increment, c is now " << c << ")\n";
    // For iterators and complex types, ALWAYS prefer ++i (no wasted copy)
    
    // ================================================================
    // SIZEOF: not a function, it's an operator!
    // ================================================================
    
    std::cout << "\n=== sizeof ===\n";
    int arr[10];
    std::cout << "sizeof(arr) = " << sizeof(arr) << " (whole array: 10 * 4 = 40)\n";
    std::cout << "sizeof(arr[0]) = " << sizeof(arr[0]) << " (one element: 4)\n";
    std::cout << "Array length = " << sizeof(arr) / sizeof(arr[0]) << '\n';
    // NOTE: sizeof on a pointer gives pointer size, NOT array size!
    // This is a common bug when passing arrays to functions (they decay to pointers)
    
    (void)v; (void)bit2;
    return 0;
}
