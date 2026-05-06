// ============================================================
// Lesson 7: Pointers & References — The Heart of C++
// ============================================================
// Compile: g++ -std=c++20 -Wall -Wextra pointers.cpp -o pointers && ./pointers
//
// If you don't understand pointers deeply, you can NEVER master C++.
// Pointers are just variables that store memory addresses. That's it.
// Everything else builds on this simple fact.

#include <iostream>
#include <cstdint>

int main() {
    // ================================================================
    // WHAT IS A POINTER?
    // ================================================================
    // A pointer is a variable that holds a MEMORY ADDRESS.
    // Memory is just a huge array of bytes. Every byte has an address.
    // A pointer stores one of these addresses.
    
    std::cout << "=== Basic Pointers ===\n";
    
    int x = 42;
    int* ptr = &x;  // ptr holds the ADDRESS of x
    
    std::cout << "x value:        " << x << '\n';
    std::cout << "x address (&x): " << &x << '\n';       // Address of x
    std::cout << "ptr value:      " << ptr << '\n';       // Same as &x
    std::cout << "ptr deref (*ptr):" << *ptr << '\n';     // Dereference: get value AT address
    std::cout << "ptr address:    " << &ptr << '\n';      // ptr itself lives somewhere in memory
    std::cout << "ptr size:       " << sizeof(ptr) << " bytes\n";  // 8 on 64-bit system
    
    // Modifying through pointer:
    *ptr = 100;  // Changes x through the pointer!
    std::cout << "After *ptr=100, x = " << x << '\n';
    
    // ================================================================
    // POINTER ARITHMETIC
    // ================================================================
    
    std::cout << "\n=== Pointer Arithmetic ===\n";
    
    int arr[] = {10, 20, 30, 40, 50};
    int* p = arr;  // Array name decays to pointer to first element
    
    std::cout << "p points to: " << *p << '\n';       // 10
    std::cout << "p+1 points to: " << *(p+1) << '\n'; // 20
    std::cout << "p+2 points to: " << *(p+2) << '\n'; // 30
    
    // p+1 doesn't add 1 byte — it adds sizeof(int) bytes (4)!
    // Pointer arithmetic is SCALED by the pointed-to type size
    std::cout << "p address:   " << p << '\n';
    std::cout << "p+1 address: " << (p+1) << '\n';  // Differs by 4 bytes (sizeof int)
    
    // Array indexing is just pointer arithmetic:
    // arr[i] is EXACTLY equivalent to *(arr + i)
    std::cout << "arr[3] = " << arr[3] << ", *(arr+3) = " << *(arr+3) << '\n';
    
    // Iterating with pointers:
    std::cout << "Array via pointer: ";
    for (int* it = arr; it != arr + 5; ++it) {
        std::cout << *it << ' ';
    }
    std::cout << '\n';
    
    // ================================================================
    // POINTER TO POINTER
    // ================================================================
    
    std::cout << "\n=== Pointer to Pointer ===\n";
    
    int val = 42;
    int* p1 = &val;    // Pointer to int
    int** p2 = &p1;    // Pointer to pointer to int
    int*** p3 = &p2;   // Pointer to pointer to pointer to int
    
    std::cout << "val = " << val << '\n';
    std::cout << "*p1 = " << *p1 << '\n';
    std::cout << "**p2 = " << **p2 << '\n';
    std::cout << "***p3 = " << ***p3 << '\n';
    
    // Main use: when a function needs to modify a pointer itself
    // int** is used for "output parameter that is a pointer"
    
    // ================================================================
    // nullptr (C++11)
    // ================================================================
    
    std::cout << "\n=== nullptr ===\n";
    
    int* null_ptr = nullptr;  // Null pointer — points to nothing
    // int* old_null = NULL;  // C-style, avoid in C++
    // int* bad_null = 0;     // Works but misleading (looks like integer 0)
    
    // ALWAYS check before dereferencing a pointer that might be null:
    if (null_ptr != nullptr) {
        std::cout << *null_ptr << '\n';  // Safe
    } else {
        std::cout << "Pointer is null — can't dereference!\n";
    }
    // Dereferencing nullptr is UNDEFINED BEHAVIOR (usually crashes)
    
    // nullptr has type std::nullptr_t — resolves overloading ambiguity:
    // void f(int);      // f(0) calls this
    // void f(int*);     // f(nullptr) calls this
    
    // ================================================================
    // const WITH POINTERS (read right-to-left!)
    // ================================================================
    
    std::cout << "\n=== const and Pointers ===\n";
    
    int a = 10, b = 20;
    
    // Read declarations RIGHT TO LEFT:
    
    int* p_mut = &a;              // "pointer to int" — can change both
    const int* p_to_const = &a;   // "pointer to CONST int" — can't modify *p
    int* const const_p = &a;      // "CONST pointer to int" — can't modify p itself
    const int* const both = &a;   // "CONST pointer to CONST int" — can't modify either
    
    *p_mut = 99;       // OK: can modify value
    p_mut = &b;        // OK: can modify pointer
    
    // *p_to_const = 99; // ERROR: value is const
    p_to_const = &b;   // OK: pointer itself isn't const
    
    *const_p = 99;     // OK: value isn't const
    // const_p = &b;    // ERROR: pointer is const
    
    // *both = 99;      // ERROR
    // both = &b;       // ERROR
    
    std::cout << "Rule: read pointer declarations RIGHT TO LEFT\n";
    std::cout << "'const int*' = pointer to const int (value is const)\n";
    std::cout << "'int* const' = const pointer to int (pointer is const)\n";
    
    // ================================================================
    // REFERENCES — aliases for existing variables
    // ================================================================
    
    std::cout << "\n=== References ===\n";
    
    int original = 42;
    int& ref = original;  // ref IS original (just another name)
    
    std::cout << "original = " << original << '\n';
    std::cout << "ref = " << ref << '\n';
    std::cout << "&original = " << &original << '\n';
    std::cout << "&ref = " << &ref << '\n';  // Same address!
    
    ref = 100;  // Modifies original
    std::cout << "After ref=100, original = " << original << '\n';
    
    // References MUST be initialized and CANNOT be rebound:
    // int& bad_ref;        // ERROR: must initialize
    // int& ref2 = ref;
    // ref2 = b;            // This doesn't rebind! It copies b's value
    
    // ================================================================
    // POINTERS vs REFERENCES
    // ================================================================
    
    std::cout << "\n=== Pointers vs References ===\n";
    std::cout << "| Feature        | Pointer     | Reference    |\n";
    std::cout << "|----------------|-------------|------------- |\n";
    std::cout << "| Can be null    | YES         | NO           |\n";
    std::cout << "| Can rebind     | YES         | NO           |\n";
    std::cout << "| Syntax         | *, &, ->    | just use it  |\n";
    std::cout << "| Arithmetic     | YES         | NO           |\n";
    std::cout << "| Must init      | NO          | YES          |\n";
    std::cout << "| Can dangle     | YES         | YES          |\n";
    
    std::cout << "\nUse references when you CAN, pointers when you MUST.\n";
    std::cout << "Use pointers when: nullable, rebindable, or doing arithmetic.\n";
    
    // ================================================================
    // DANGLING POINTERS — the #1 memory bug
    // ================================================================
    
    std::cout << "\n=== Dangling Pointers ===\n";
    
    int* dangling = nullptr;
    {
        int temp = 42;
        dangling = &temp;
        std::cout << "Inside scope: *dangling = " << *dangling << '\n';
    }  // temp is destroyed here!
    // *dangling is now UB — temp's memory has been reclaimed
    std::cout << "After scope: dangling points to freed memory (UB if accessed)\n";
    
    // Other causes of dangling:
    // 1. Returning pointer/reference to local variable
    // 2. Using pointer after delete
    // 3. Iterator invalidation (vector reallocation)
    
    // ================================================================
    // void POINTER — generic pointer, no type info
    // ================================================================
    
    std::cout << "\n=== void Pointer ===\n";
    
    int iv = 42;
    double dv = 3.14;
    void* generic = &iv;  // OK: any pointer converts to void*
    
    // Can't dereference void* directly — must cast first:
    // *generic = 10;  // ERROR: void* has no type info
    std::cout << "void* to int: " << *static_cast<int*>(generic) << '\n';
    
    generic = &dv;
    std::cout << "void* to double: " << *static_cast<double*>(generic) << '\n';
    
    // void* is used in C-style APIs (malloc returns void*)
    // In C++, prefer templates for generic code instead of void*
    
    // ================================================================
    // DYNAMIC MEMORY (new/delete) — Preview
    // ================================================================
    
    std::cout << "\n=== Dynamic Memory (Brief) ===\n";
    
    // Stack allocation: automatic, scoped lifetime
    int stack_var = 10;  // Lives until end of scope
    
    // Heap allocation: manual lifetime management
    int* heap_var = new int(42);     // Allocate on heap
    std::cout << "Heap value: " << *heap_var << '\n';
    delete heap_var;                  // YOU must free it! Forgetting = memory leak
    heap_var = nullptr;               // Good practice: null after delete
    
    // Dynamic arrays:
    int* heap_arr = new int[5]{1, 2, 3, 4, 5};
    std::cout << "Heap array: " << heap_arr[2] << '\n';
    delete[] heap_arr;  // MUST use delete[] for arrays! (not delete)
    
    // In modern C++: NEVER use raw new/delete
    // Use std::unique_ptr, std::shared_ptr, std::vector instead
    // We'll cover this in Phase 3 (Smart Pointers)
    
    (void)null_ptr; (void)both; (void)p_to_const;
    return 0;
}
