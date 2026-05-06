// ============================================================
// Lesson 6: Arrays & Strings — The Full Picture
// ============================================================
// Compile: g++ -std=c++20 -Wall -Wextra arrays_strings.cpp -o arrays_strings && ./arrays_strings

#include <iostream>
#include <array>        // std::array
#include <vector>       // std::vector (dynamic array)
#include <string>       // std::string
#include <string_view>  // std::string_view (C++17)
#include <cstring>      // strlen, strcmp, etc. (C functions)
#include <algorithm>    // sort, etc.
#include <span>         // std::span (C++20)

int main() {
    // ================================================================
    // C-STYLE ARRAYS (know them, but prefer std::array/vector)
    // ================================================================
    
    std::cout << "=== C-Style Arrays ===\n";
    
    int arr[5] = {10, 20, 30, 40, 50};  // Fixed size, on the stack
    int arr2[] = {1, 2, 3};             // Size deduced from initializer (3)
    int arr3[5] = {1, 2};              // Rest initialized to 0: {1, 2, 0, 0, 0}
    int arr4[5] = {};                   // All zeros
    
    // Accessing elements:
    std::cout << "arr[0] = " << arr[0] << '\n';  // No bounds checking!
    // arr[10] = 99;  // UB! Out of bounds — no error at compile time!
    
    // Arrays DECAY to pointers when passed to functions:
    // void func(int arr[]) ← this is actually void func(int* arr)!
    // You LOSE the size information. This is why C-style arrays are dangerous.
    
    // Array size from sizeof (only works in same scope, NOT after passing):
    size_t size = sizeof(arr) / sizeof(arr[0]);  // 5
    std::cout << "Array size: " << size << '\n';
    
    // 2D arrays:
    int matrix[3][4] = {
        {1, 2, 3, 4},
        {5, 6, 7, 8},
        {9, 10, 11, 12}
    };
    std::cout << "matrix[1][2] = " << matrix[1][2] << '\n';  // 7
    
    // ================================================================
    // std::array (C++11) — Fixed size, safe, zero overhead
    // ================================================================
    
    std::cout << "\n=== std::array ===\n";
    
    std::array<int, 5> safe_arr = {10, 20, 30, 40, 50};
    
    std::cout << "Size: " << safe_arr.size() << '\n';       // Knows its size!
    std::cout << "Front: " << safe_arr.front() << '\n';
    std::cout << "Back: " << safe_arr.back() << '\n';
    // std::cout << safe_arr.at(10);  // throws std::out_of_range (bounds checked!)
    
    // Can be sorted, compared, etc.:
    std::array<int, 5> arr_b = {50, 10, 40, 20, 30};
    std::sort(arr_b.begin(), arr_b.end());
    std::cout << "Sorted: ";
    for (const auto& v : arr_b) std::cout << v << ' ';
    std::cout << '\n';
    
    // Works with range-based for:
    for (auto& v : safe_arr) {
        v *= 2;
    }
    
    // ================================================================
    // std::vector — Dynamic array (THE workhorse container)
    // ================================================================
    
    std::cout << "\n=== std::vector ===\n";
    
    std::vector<int> vec;                       // Empty
    std::vector<int> vec2(10);                  // 10 elements, all 0
    std::vector<int> vec3(5, 42);              // 5 elements, all 42
    std::vector<int> vec4 = {1, 2, 3, 4, 5};  // Initializer list
    
    // Adding elements:
    vec.push_back(10);   // Add to end
    vec.push_back(20);
    vec.push_back(30);
    vec.emplace_back(40);  // Construct in place (more efficient for objects)
    
    // Size vs Capacity:
    std::cout << "Size: " << vec.size() << '\n';        // Number of elements
    std::cout << "Capacity: " << vec.capacity() << '\n'; // Allocated space
    // When size > capacity, vector REALLOCATES (expensive!)
    // Strategy: doubles capacity each time (amortized O(1) push_back)
    
    vec.reserve(100);  // Pre-allocate if you know the size
    std::cout << "After reserve(100) — Capacity: " << vec.capacity() << '\n';
    
    // Removing elements:
    vec.pop_back();           // Remove last
    vec.erase(vec.begin());   // Remove first (O(n) — shifts everything!)
    
    // Accessing:
    std::cout << "vec[0] = " << vec[0] << '\n';     // No bounds check
    std::cout << "vec.at(0) = " << vec.at(0) << '\n';  // Bounds checked
    
    // IMPORTANT: vector invalidation rules
    // push_back/emplace_back/insert/resize MAY invalidate ALL iterators
    // if reallocation occurs. This is a common source of bugs!
    
    // ================================================================
    // C-STRINGS (null-terminated char arrays)
    // ================================================================
    
    std::cout << "\n=== C-Strings ===\n";
    
    const char* cstr = "Hello";  // String literal → stored in read-only memory
    char mutable_str[] = "Hello";  // Array copy → can be modified
    
    // String literals are null-terminated: "Hello" = {'H','e','l','l','o','\0'}
    std::cout << "Length: " << strlen(cstr) << '\n';  // 5 (doesn't count '\0')
    std::cout << "sizeof: " << sizeof(mutable_str) << '\n';  // 6 (includes '\0')
    
    // C-string functions (from <cstring>):
    // strlen(s)           — length
    // strcmp(s1, s2)      — compare (0 if equal)
    // strcpy(dst, src)    — copy (DANGEROUS! no bounds check)
    // strncpy(dst, src, n) — bounded copy
    // strcat(dst, src)    — concatenate
    
    // ================================================================
    // std::string — THE string type in C++
    // ================================================================
    
    std::cout << "\n=== std::string ===\n";
    
    std::string s1 = "Hello";            // From literal
    std::string s2("World");             // Constructor
    std::string s3(5, 'x');              // "xxxxx"
    std::string s4 = s1 + ", " + s2;    // Concatenation with +
    
    std::cout << "s4 = " << s4 << '\n';
    std::cout << "Length: " << s4.length() << '\n';  // or .size()
    std::cout << "Char at 0: " << s4[0] << '\n';
    std::cout << "Substr: " << s4.substr(0, 5) << '\n';
    
    // Searching:
    size_t pos = s4.find("World");
    if (pos != std::string::npos) {
        std::cout << "'World' found at position " << pos << '\n';
    }
    
    // Modifying:
    s4.append("!!!");
    s4.insert(5, " Beautiful");
    std::cout << "Modified: " << s4 << '\n';
    
    // Comparing:
    if (s1 == "Hello") std::cout << "Strings compare with ==\n";
    if (s1 < s2) std::cout << "Lexicographic comparison works\n";
    
    // Converting:
    std::string num_str = std::to_string(42);     // int → string
    int num = std::stoi("123");                   // string → int
    double dbl = std::stod("3.14");               // string → double
    std::cout << "to_string(42) = " << num_str << '\n';
    std::cout << "stoi(\"123\") = " << num << '\n';
    std::cout << "stod(\"3.14\") = " << dbl << '\n';
    
    // ================================================================
    // std::string_view (C++17) — Non-owning view into a string
    // ================================================================
    
    std::cout << "\n=== std::string_view ===\n";
    
    // string_view does NOT own the string — just a pointer + length
    // ZERO COST: no heap allocation, no copy
    // Use for function parameters when you just need to READ a string
    
    std::string_view sv = "Hello, string_view!";  // No allocation!
    std::cout << "View: " << sv << '\n';
    std::cout << "Substr view: " << sv.substr(0, 5) << '\n';  // Also no allocation!
    
    // DANGER: string_view can dangle if the underlying string is destroyed!
    // std::string_view sv2;
    // {
    //     std::string temp = "temporary";
    //     sv2 = temp;  // sv2 points to temp's buffer
    // }  // temp destroyed here!
    // std::cout << sv2;  // DANGLING! UB!
    
    // When to use what:
    // std::string        → when you need to OWN and modify the string
    // std::string_view   → when you just need to READ (function parameters)
    // const char*        → when interfacing with C APIs
    
    // ================================================================
    // std::span (C++20) — Non-owning view into a contiguous sequence
    // ================================================================
    
    std::cout << "\n=== std::span (C++20) ===\n";
    
    // Like string_view but for any array/vector
    int raw_arr[] = {1, 2, 3, 4, 5};
    std::span<int> sp(raw_arr);  // View into the array
    
    std::cout << "Span size: " << sp.size() << '\n';
    std::cout << "First 3: ";
    for (int v : sp.first(3)) {
        std::cout << v << ' ';
    }
    std::cout << '\n';
    
    // Works with vectors too:
    std::vector<int> v = {10, 20, 30, 40, 50};
    std::span<int> sp2(v);
    std::cout << "Vector span: ";
    for (int val : sp2) {
        std::cout << val << ' ';
    }
    std::cout << '\n';
    
    (void)arr2; (void)arr3; (void)arr4; (void)matrix;
    (void)vec2; (void)vec3; (void)vec4;
    (void)cstr;
    
    return 0;
}
