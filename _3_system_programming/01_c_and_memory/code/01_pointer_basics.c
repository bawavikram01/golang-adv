// pointer_basics.c — Understanding memory layout
#include <stdio.h>
#include <stdlib.h>

int global_var = 42;               // Data segment (initialized)
int uninitialized_global;          // BSS segment (zeroed by OS)

int main() {
    int stack_var = 10;
    int *heap_var = malloc(sizeof(int));
    if (!heap_var) {
        perror("malloc");
        return 1;
    }
    *heap_var = 20;

    printf("=== Memory Layout of This Process ===\n\n");
    printf("Text  (code):  %p\n", (void *)main);
    printf("Data  (init):  %p\n", (void *)&global_var);
    printf("BSS   (zero):  %p\n", (void *)&uninitialized_global);
    printf("Heap:          %p\n", (void *)heap_var);
    printf("Stack:         %p\n", (void *)&stack_var);

    printf("\n--- Observations ---\n");
    printf("Stack address > Heap address? %s\n",
           (void *)&stack_var > (void *)heap_var ? "YES (stack grows down)" : "NO");
    printf("BSS is zeroed? uninitialized_global = %d\n", uninitialized_global);

    free(heap_var);
    return 0;
}
