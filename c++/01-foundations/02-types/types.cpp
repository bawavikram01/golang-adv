// ============================================================
// Lesson 2: The C++ Type System — Your Foundation
// ============================================================
// Compile: g++ -std=c++20 -Wall -Wextra types.cpp -o types && ./types
//
// C++ is STATICALLY TYPED: every variable's type is known at compile time.
// C++ is STRONGLY TYPED: implicit conversions are limited and often warned about.
// Understanding types deeply is what separates beginners from experts.

#include <iostream>
#include <cstdint>   // Fixed-width integers
#include <limits>    // std::numeric_limits
#include <climits>   // INT_MAX, etc.

int main() {
    // ================================================================
    // FUNDAMENTAL TYPES
    // ================================================================
    
    // Boolean — true (1) or false (0)
    bool is_learning = true;
    
    // Character types — store characters OR small integers
    char letter = 'A';           // At least 8 bits, could be signed or unsigned!
    signed char sc = -128;       // Explicitly signed: -128 to 127
    unsigned char uc = 255;      // Explicitly unsigned: 0 to 255
    char8_t utf8_char = u8'A';   // C++20: UTF-8 character
    char16_t utf16_char = u'A';  // UTF-16 character
    char32_t utf32_char = U'A';  // UTF-32 character (full Unicode)
    wchar_t wide_char = L'A';    // Wide character (platform-dependent size)
    
    // Integer types (sizes are MINIMUM guarantees, actual size is platform-dependent)
    short s = 32767;             // At least 16 bits
    int i = 2147483647;          // At least 16 bits (usually 32 on modern systems)
    long l = 2147483647L;        // At least 32 bits
    long long ll = 9223372036854775807LL;  // At least 64 bits
    
    // Unsigned variants (cannot be negative, double the positive range)
    unsigned short us = 65535;
    unsigned int ui = 4294967295U;
    unsigned long ul = 4294967295UL;
    unsigned long long ull = 18446744073709551615ULL;
    
    // Floating point
    float f = 3.14f;            // ~7 decimal digits precision, 32 bits
    double d = 3.141592653589;  // ~15 decimal digits precision, 64 bits
    long double ld = 3.14159265358979323846L;  // Extended precision (80-128 bits)
    
    // void — represents "no type" (used for functions that return nothing)
    // You can't create a void variable!
    
    // ================================================================
    // FIXED-WIDTH INTEGERS (USE THESE when size matters!)
    // ================================================================
    // Defined in <cstdint> — guaranteed exact sizes
    
    int8_t   i8  = -128;          // Exactly 8 bits, signed
    int16_t  i16 = -32768;        // Exactly 16 bits, signed
    int32_t  i32 = -2147483648;   // Exactly 32 bits, signed
    int64_t  i64 = -9223372036854775807LL - 1;  // Exactly 64 bits, signed
    
    uint8_t  u8  = 255;           // Exactly 8 bits, unsigned
    uint16_t u16 = 65535;         // Exactly 16 bits, unsigned
    uint32_t u32 = 4294967295U;   // Exactly 32 bits, unsigned
    uint64_t u64 = 18446744073709551615ULL;  // Exactly 64 bits, unsigned
    
    // size_t — unsigned type for sizes and indices (usually 64-bit on modern systems)
    size_t size = sizeof(int);  // sizeof returns size_t
    
    // ptrdiff_t — signed type for pointer differences
    int arr[5] = {1, 2, 3, 4, 5};
    ptrdiff_t diff = &arr[4] - &arr[0];  // = 4
    
    // ================================================================
    // PRINTING SIZES AND LIMITS
    // ================================================================
    
    std::cout << "=== Type Sizes (in bytes) ===\n";
    std::cout << "bool:        " << sizeof(bool) << '\n';
    std::cout << "char:        " << sizeof(char) << '\n';      // Always 1
    std::cout << "short:       " << sizeof(short) << '\n';
    std::cout << "int:         " << sizeof(int) << '\n';
    std::cout << "long:        " << sizeof(long) << '\n';
    std::cout << "long long:   " << sizeof(long long) << '\n';
    std::cout << "float:       " << sizeof(float) << '\n';
    std::cout << "double:      " << sizeof(double) << '\n';
    std::cout << "long double: " << sizeof(long double) << '\n';
    std::cout << "size_t:      " << sizeof(size_t) << '\n';
    std::cout << "pointer:     " << sizeof(void*) << '\n';
    
    std::cout << "\n=== Integer Limits ===\n";
    std::cout << "int min:       " << std::numeric_limits<int>::min() << '\n';
    std::cout << "int max:       " << std::numeric_limits<int>::max() << '\n';
    std::cout << "uint max:      " << std::numeric_limits<unsigned int>::max() << '\n';
    std::cout << "int64_t max:   " << std::numeric_limits<int64_t>::max() << '\n';
    
    std::cout << "\n=== Float Properties ===\n";
    std::cout << "float digits10:  " << std::numeric_limits<float>::digits10 << '\n';
    std::cout << "double digits10: " << std::numeric_limits<double>::digits10 << '\n';
    std::cout << "float epsilon:   " << std::numeric_limits<float>::epsilon() << '\n';
    
    // ================================================================
    // TYPE DEDUCTION WITH auto
    // ================================================================
    
    auto x = 42;          // int (integer literals default to int)
    auto y = 42L;         // long
    auto z = 42ULL;       // unsigned long long
    auto pi = 3.14;       // double (floating literals default to double)
    auto pi_f = 3.14f;    // float
    auto ch = 'A';        // char
    auto str = "hello";   // const char* (NOT std::string!)
    
    // auto is great for:
    // 1. Long type names (iterators, etc.)
    // 2. When the type is obvious from the right side
    // 3. Generic code
    // 
    // auto is BAD when:
    // 1. The intended type isn't clear
    // 2. You want a specific type (e.g., int32_t vs int)
    
    // ================================================================
    // const and constexpr
    // ================================================================
    
    const int MAX_SIZE = 100;       // Runtime constant — cannot be modified after init
    constexpr int ARRAY_SIZE = 50;  // Compile-time constant — MUST be known at compile time
    
    // constexpr is stronger than const:
    // - constexpr: value computed at compile time (guaranteed)
    // - const: value won't change, but may be determined at runtime
    
    int runtime_value = 42;         // Determined at runtime
    const int c = runtime_value;    // OK: const can hold runtime values
    // constexpr int ce = runtime_value;  // ERROR: constexpr needs compile-time value
    
    // ================================================================
    // LITERAL SUFFIXES
    // ================================================================
    
    // Integer suffixes
    auto a1 = 42;       // int
    auto a2 = 42U;      // unsigned int
    auto a3 = 42L;      // long
    auto a4 = 42UL;     // unsigned long
    auto a5 = 42LL;     // long long
    auto a6 = 42ULL;    // unsigned long long
    
    // Float suffixes
    auto b1 = 3.14;     // double
    auto b2 = 3.14f;    // float
    auto b3 = 3.14L;    // long double
    
    // Integer bases
    auto hex = 0xFF;          // Hexadecimal (255)
    auto oct = 0777;          // Octal (511)
    auto bin = 0b1010'1010;   // Binary with digit separator (170)
    auto big = 1'000'000;     // Digit separator for readability (1000000)
    
    std::cout << "\n=== Literals ===\n";
    std::cout << "hex 0xFF = " << hex << '\n';
    std::cout << "oct 0777 = " << oct << '\n';
    std::cout << "bin 0b10101010 = " << bin << '\n';
    std::cout << "big 1'000'000 = " << big << '\n';
    
    // Suppress unused variable warnings
    (void)is_learning; (void)letter; (void)sc; (void)uc;
    (void)utf8_char; (void)utf16_char; (void)utf32_char; (void)wide_char;
    (void)s; (void)i; (void)l; (void)ll;
    (void)us; (void)ui; (void)ul; (void)ull;
    (void)f; (void)d; (void)ld;
    (void)i8; (void)i16; (void)i32; (void)i64;
    (void)u8; (void)u16; (void)u32; (void)u64;
    (void)size; (void)diff;
    (void)x; (void)y; (void)z; (void)pi; (void)pi_f; (void)ch; (void)str;
    (void)MAX_SIZE; (void)ARRAY_SIZE; (void)c;
    (void)a1; (void)a2; (void)a3; (void)a4; (void)a5; (void)a6;
    (void)b1; (void)b2; (void)b3;
    
    return 0;
}
