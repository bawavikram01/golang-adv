// ============================================================
// Lesson 4: Control Flow — Beyond the Basics
// ============================================================
// Compile: g++ -std=c++20 -Wall -Wextra control_flow.cpp -o control_flow && ./control_flow

#include <iostream>
#include <vector>
#include <string>
#include <optional>

int main() {
    // ================================================================
    // IF WITH INITIALIZER (C++17) — limits variable scope
    // ================================================================
    
    std::cout << "=== if with initializer (C++17) ===\n";
    
    std::vector<int> numbers = {10, 20, 30, 40, 50};
    
    // Old style:
    // auto it = std::find(numbers.begin(), numbers.end(), 30);
    // if (it != numbers.end()) { ... }
    // 'it' leaks into surrounding scope!
    
    // C++17 style: init statement in if
    if (auto it = std::find(numbers.begin(), numbers.end(), 30); it != numbers.end()) {
        std::cout << "Found: " << *it << " at index " << (it - numbers.begin()) << '\n';
    }
    // 'it' doesn't exist here — much cleaner!
    
    // ================================================================
    // SWITCH STATEMENT — deep understanding
    // ================================================================
    
    std::cout << "\n=== Switch ===\n";
    
    int choice = 2;
    switch (choice) {
        case 1:
            std::cout << "One\n";
            break;  // Without break, execution FALLS THROUGH to next case!
        case 2:
            std::cout << "Two\n";
            [[fallthrough]];  // C++17: explicit fallthrough (silences warnings)
        case 3:
            std::cout << "Two or Three path\n";
            break;
        default:
            std::cout << "Other\n";
            break;
    }
    
    // Switch with initializer (C++17):
    switch (auto len = numbers.size(); len) {
        case 0:  std::cout << "Empty\n"; break;
        case 1:  std::cout << "Single\n"; break;
        default: std::cout << "Multiple (" << len << " elements)\n"; break;
    }
    
    // ================================================================
    // LOOPS — All variants
    // ================================================================
    
    std::cout << "\n=== Loops ===\n";
    
    // Traditional for: init; condition; increment
    std::cout << "for: ";
    for (int i = 0; i < 5; ++i) {  // Prefer ++i over i++ (no wasted copy)
        std::cout << i << ' ';
    }
    std::cout << '\n';
    
    // Range-based for (C++11) — THE preferred way to iterate
    std::cout << "range-for: ";
    for (int n : numbers) {  // Copies each element
        std::cout << n << ' ';
    }
    std::cout << '\n';
    
    // Range-for with reference (no copy):
    std::cout << "range-for (ref): ";
    for (const int& n : numbers) {  // const ref: read-only, no copy
        std::cout << n << ' ';
    }
    std::cout << '\n';
    
    // Range-for with auto:
    for (const auto& n : numbers) {  // Most idiomatic for reading
        (void)n;
    }
    
    // Modify with range-for:
    for (auto& n : numbers) {  // Non-const ref: can modify
        n *= 2;
    }
    std::cout << "After doubling: ";
    for (const auto& n : numbers) {
        std::cout << n << ' ';
    }
    std::cout << '\n';
    
    // Range-for with structured bindings (C++17) — for maps/pairs:
    std::cout << "\nStructured bindings in range-for:\n";
    std::vector<std::pair<std::string, int>> scores = {
        {"Alice", 95}, {"Bob", 87}, {"Charlie", 92}
    };
    for (const auto& [name, score] : scores) {
        std::cout << "  " << name << ": " << score << '\n';
    }
    
    // While loop:
    int countdown = 3;
    std::cout << "\nwhile: ";
    while (countdown > 0) {
        std::cout << countdown << ' ';
        --countdown;
    }
    std::cout << "Go!\n";
    
    // Do-while: executes at least once
    int input = 0;
    do {
        ++input;
    } while (input < 0);  // Condition checked AFTER first iteration
    
    // ================================================================
    // STRUCTURED BINDINGS (C++17) — decompose objects
    // ================================================================
    
    std::cout << "\n=== Structured Bindings ===\n";
    
    // With arrays:
    int arr[] = {1, 2, 3};
    auto [first, second, third] = arr;
    std::cout << "Array: " << first << ", " << second << ", " << third << '\n';
    
    // With pairs:
    auto [name, age] = std::pair<std::string, int>{"Vikram", 25};
    std::cout << "Pair: " << name << ", " << age << '\n';
    
    // With structs:
    struct Point { double x, y; };
    auto [px, py] = Point{3.14, 2.71};
    std::cout << "Point: (" << px << ", " << py << ")\n";
    
    // ================================================================
    // TECHNIQUES: Early return & Guard clauses
    // ================================================================
    
    std::cout << "\n=== Guard Clause Pattern ===\n";
    
    // BAD: Deep nesting (arrow code)
    // if (condition1) {
    //     if (condition2) {
    //         if (condition3) {
    //             // actual work buried deep
    //         }
    //     }
    // }
    
    // GOOD: Guard clauses (check and return early)
    // See the function below this main for example
    
    auto validate = [](int value) -> std::string {
        if (value < 0) return "negative";
        if (value == 0) return "zero";
        if (value > 100) return "too large";
        return "valid: " + std::to_string(value);
    };
    
    std::cout << "validate(-1): " << validate(-1) << '\n';
    std::cout << "validate(0):  " << validate(0) << '\n';
    std::cout << "validate(50): " << validate(50) << '\n';
    std::cout << "validate(200):" << validate(200) << '\n';
    
    return 0;
}
