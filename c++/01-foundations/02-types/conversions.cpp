// ============================================================
// Lesson 2b: Type Conversions — The Silent Killers
// ============================================================
// Compile: g++ -std=c++20 -Wall -Wextra -Wconversion conversions.cpp -o conversions && ./conversions
//
// Type conversions (casts) are one of the biggest sources of bugs in C++.
// Understanding them is CRITICAL for writing correct code.

#include <iostream>
#include <cstdint>

int main() {
    // ================================================================
    // IMPLICIT CONVERSIONS (happen automatically — often dangerous!)
    // ================================================================
    
    std::cout << "=== Implicit Conversions ===\n";
    
    // Integer promotion: smaller types promoted to int in expressions
    short a = 10;
    short b = 20;
    // a + b is computed as int, not short!
    auto result = a + b;  // result is int, not short
    std::cout << "short + short = int? " << sizeof(result) << " bytes\n";
    
    // Narrowing: losing data silently (DANGEROUS!)
    int big = 300;
    char small = big;  // char can only hold -128 to 127 (or 0-255 unsigned)
    std::cout << "int 300 → char: " << static_cast<int>(small) << " (DATA LOSS!)\n";
    
    // Signed/unsigned mismatch (EXTREMELY common bug source)
    unsigned int u = 1;
    int s = -1;
    // When comparing signed and unsigned, signed is converted to unsigned!
    if (u > s) {
        std::cout << "1 > -1? This won't print!\n";
    } else {
        std::cout << "SURPRISE: unsigned 1 is NOT > signed -1 (after conversion)\n";
        std::cout << "-1 as unsigned = " << static_cast<unsigned int>(s) << '\n';
    }
    
    // Float to int truncation
    double pi = 3.99;
    int truncated = pi;  // Truncates, doesn't round! → 3
    std::cout << "double 3.99 → int: " << truncated << " (truncated!)\n";
    
    // Bool conversions
    int zero = 0;
    int nonzero = 42;
    bool b1 = zero;     // 0 → false
    bool b2 = nonzero;  // any non-zero → true
    std::cout << "0 → bool: " << b1 << ", 42 → bool: " << b2 << '\n';
    
    // ================================================================
    // C++ NAMED CASTS (prefer these over C-style casts)
    // ================================================================
    
    std::cout << "\n=== Named Casts ===\n";
    
    // static_cast: compile-time checked, most common
    // Use for: numeric conversions, upcasting, void* to type*
    double d = 3.14;
    int i = static_cast<int>(d);  // Explicit: "yes, I know I'm losing precision"
    std::cout << "static_cast<int>(3.14) = " << i << '\n';
    
    // const_cast: add or remove const (DANGEROUS, rarely needed)
    // Use ONLY when interfacing with legacy code that doesn't use const
    const int ci = 42;
    // int* p = &ci;  // ERROR: can't get non-const pointer to const
    int* p = const_cast<int*>(&ci);  // Removes const (undefined behavior if you write!)
    std::cout << "const_cast result: " << *p << '\n';
    // *p = 100;  // UNDEFINED BEHAVIOR! ci was declared const
    
    // reinterpret_cast: bit reinterpretation (DANGEROUS, low-level)
    // Use for: viewing memory as different types, hardware programming
    int32_t value = 0x41424344;  // ASCII: 'A' 'B' 'C' 'D'
    char* bytes = reinterpret_cast<char*>(&value);
    std::cout << "reinterpret_cast bytes: ";
    for (int idx = 0; idx < 4; ++idx) {
        std::cout << bytes[idx];
    }
    std::cout << " (byte order depends on endianness!)\n";
    
    // dynamic_cast: runtime checked, for polymorphic types (needs virtual functions)
    // We'll cover this when we learn inheritance
    
    // ================================================================
    // C-STYLE CAST (AVOID! No compile-time checking)
    // ================================================================
    
    // C-style cast: (type)expression — does whatever it takes, no safety
    double bad = (double)i;      // Might be static_cast, reinterpret_cast, or const_cast
    std::cout << "C-style cast: " << bad << " (DON'T USE THIS)\n";
    
    // ================================================================
    // BRACE INITIALIZATION PREVENTS NARROWING (C++11)
    // ================================================================
    
    std::cout << "\n=== Brace Init (Narrowing Prevention) ===\n";
    
    // int narrow{3.14};  // COMPILE ERROR! Brace init prevents narrowing
    int safe{42};          // OK: no narrowing
    // int oops{big};      // ERROR if big might not fit (depends on value)
    
    std::cout << "Brace init safe: " << safe << '\n';
    std::cout << "Use {} initialization to catch narrowing bugs at compile time!\n";
    
    // ================================================================
    // KEY RULES TO REMEMBER
    // ================================================================
    
    std::cout << "\n=== Rules ===\n";
    std::cout << "1. NEVER mix signed and unsigned in comparisons\n";
    std::cout << "2. Use static_cast for intentional conversions\n";
    std::cout << "3. Use {} init to prevent accidental narrowing\n";
    std::cout << "4. Compile with -Wconversion to catch implicit narrows\n";
    std::cout << "5. Prefer fixed-width types (int32_t) when size matters\n";
    
    (void)p; (void)b1; (void)b2;
    
    return 0;
}
