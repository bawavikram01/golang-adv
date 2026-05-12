// pointer_arithmetic.c — How pointer math actually works
#include <stdio.h>
#include <stdint.h>
#include <string.h>

void basic_arithmetic() {
    printf("=== Basic Pointer Arithmetic ===\n");
    int arr[] = {10, 20, 30, 40, 50};
    int *p = arr;

    printf("arr address:  %p\n", (void *)arr);
    printf("p + 0 = %p → value: %d\n", (void *)(p + 0), *(p + 0));
    printf("p + 1 = %p → value: %d\n", (void *)(p + 1), *(p + 1));
    printf("p + 2 = %p → value: %d\n", (void *)(p + 2), *(p + 2));
    printf("\nNote: each step is %zu bytes (sizeof(int))\n", sizeof(int));
}

void pointer_to_different_types() {
    printf("\n=== Same Memory, Different Interpretations ===\n");

    // Store bytes, read as different types
    unsigned char bytes[8] = {0x01, 0x00, 0x00, 0x00,   // 1 as little-endian int32
                               0xFF, 0xFF, 0xFF, 0xFF};  // -1 as signed int32

    int *as_int = (int *)bytes;
    printf("First 4 bytes as int:  %d\n", as_int[0]);   // 1
    printf("Next 4 bytes as int:   %d\n", as_int[1]);   // -1

    short *as_short = (short *)bytes;
    printf("First 2 bytes as short: %d\n", as_short[0]); // 1
    printf("Next 2 bytes as short:  %d\n", as_short[1]); // 0

    printf("\nThis is how network protocols and file formats work!\n");
    printf("You cast raw bytes to structured types.\n");
}

void void_pointer_usage() {
    printf("\n=== void* — The Generic Pointer ===\n");

    int x = 42;
    float f = 3.14f;
    char c = 'A';

    // void* can point to anything — you just lose type info
    void *ptr;

    ptr = &x;
    printf("void* → int:   %d\n", *(int *)ptr);

    ptr = &f;
    printf("void* → float: %.2f\n", *(float *)ptr);

    ptr = &c;
    printf("void* → char:  %c\n", *(char *)ptr);

    printf("\nmalloc returns void* — you cast it to your type\n");
}

void array_decay() {
    printf("\n=== Array Decay — Arrays ARE Pointers (almost) ===\n");

    int arr[5] = {1, 2, 3, 4, 5};

    // arr "decays" to a pointer to its first element
    int *p = arr;  // same as: int *p = &arr[0]

    printf("arr == &arr[0]? %s\n", (void *)arr == (void *)&arr[0] ? "YES" : "NO");
    printf("sizeof(arr) = %zu (full array size)\n", sizeof(arr));
    printf("sizeof(p)   = %zu (just a pointer)\n", sizeof(p));

    printf("\nKey difference: sizeof(arr) knows the full size,\n");
    printf("but once passed to a function, it decays to pointer and size is lost.\n");
}

// This receives a POINTER, not an array — size info is lost
void print_array(int *arr, int n) {
    printf("\nInside function: sizeof(arr) = %zu (pointer size, NOT array size)\n", sizeof(arr));
    for (int i = 0; i < n; i++) {
        printf("%d ", arr[i]);
    }
    printf("\n");
}

int main() {
    basic_arithmetic();
    pointer_to_different_types();
    void_pointer_usage();
    array_decay();

    int nums[] = {10, 20, 30};
    print_array(nums, 3);

    return 0;
}
