// struct_layout.c — How structs are laid out in memory (padding & alignment)
#include <stdio.h>
#include <stddef.h>  // offsetof

// BAD layout — wastes memory due to padding
struct BadLayout {
    char a;      // 1 byte + 7 bytes padding (to align next member)
    double b;    // 8 bytes
    char c;      // 1 byte + 3 bytes padding
    int d;       // 4 bytes
};               // Total: likely 24 bytes (not 14!)

// GOOD layout — ordered by size (largest first)
struct GoodLayout {
    double b;    // 8 bytes
    int d;       // 4 bytes
    char a;      // 1 byte
    char c;      // 1 byte + 2 bytes padding (struct must be multiple of largest alignment)
};               // Total: 16 bytes

// Packed struct — no padding (used in network protocols, file formats)
struct __attribute__((packed)) PackedLayout {
    char a;
    double b;
    char c;
    int d;
};               // Total: exactly 14 bytes (but may be slower to access!)

int main() {
    printf("=== Struct Memory Layout & Padding ===\n\n");

    printf("BadLayout:    size = %zu bytes\n", sizeof(struct BadLayout));
    printf("  offset of a: %zu\n", offsetof(struct BadLayout, a));
    printf("  offset of b: %zu\n", offsetof(struct BadLayout, b));
    printf("  offset of c: %zu\n", offsetof(struct BadLayout, c));
    printf("  offset of d: %zu\n", offsetof(struct BadLayout, d));

    printf("\nGoodLayout:   size = %zu bytes\n", sizeof(struct GoodLayout));
    printf("  offset of b: %zu\n", offsetof(struct GoodLayout, b));
    printf("  offset of d: %zu\n", offsetof(struct GoodLayout, d));
    printf("  offset of a: %zu\n", offsetof(struct GoodLayout, a));
    printf("  offset of c: %zu\n", offsetof(struct GoodLayout, c));

    printf("\nPackedLayout: size = %zu bytes\n", sizeof(struct PackedLayout));
    printf("  offset of a: %zu\n", offsetof(struct PackedLayout, a));
    printf("  offset of b: %zu\n", offsetof(struct PackedLayout, b));
    printf("  offset of c: %zu\n", offsetof(struct PackedLayout, c));
    printf("  offset of d: %zu\n", offsetof(struct PackedLayout, d));

    printf("\n--- WHY THIS MATTERS ---\n");
    printf("1. Cache performance: smaller structs = more fit in cache line (64 bytes)\n");
    printf("2. Network protocols: packed structs map directly to wire format\n");
    printf("3. Memory usage: millions of structs × padding = wasted MB\n");
    printf("\nRule: Order struct members from largest to smallest alignment.\n");

    return 0;
}
