/*
 * ============================================================
 *  CHAPTER 51: BYTECODE & METHODHANDLES
 * ============================================================
 *  At god level, you understand what javac ACTUALLY generates,
 *  how the JVM executes it, and how to manipulate invocations
 *  at the lowest level Java allows.
 *
 *  TOPICS:
 *    1. Understanding Bytecode (javap -c)
 *    2. Key Bytecode Instructions
 *    3. Constant Pool
 *    4. MethodHandles — the modern reflection
 *    5. MethodType
 *    6. Lookup — access control for MethodHandles
 *    7. VarHandle — memory-order-aware field access
 *    8. invokedynamic — how lambdas and string concat work
 *    9. Practical Uses
 * ============================================================
 *
 *  TO EXPLORE BYTECODE:
 *    javac Chapter51_Bytecode.java
 *    javap -c -p -v Chapter51_Bytecode
 *
 *  KEY BYTECODE INSTRUCTIONS:
 *  ──────────────────────────────────────────────────
 *  LOAD/STORE:
 *    iload_0      load int from local var 0
 *    aload_1      load reference from local var 1
 *    istore_2     store int to local var 2
 *    ldc "hello"  load constant from constant pool
 *
 *  ARITHMETIC:
 *    iadd, isub, imul, idiv     int math
 *    ladd, lsub                 long math
 *    fadd, dadd                 float/double math
 *    i2l, l2i, i2f              type conversions
 *
 *  OBJECT:
 *    new          allocate object (uninitialized)
 *    invokespecial  call <init> or super method
 *    invokevirtual  call instance method (virtual dispatch)
 *    invokestatic   call static method
 *    invokeinterface call interface method
 *    invokedynamic   call bootstrap → CallSite
 *    getfield/putfield   instance field access
 *    getstatic/putstatic  static field access
 *    checkcast    runtime type check (throws ClassCastException)
 *    instanceof   test type (returns 0 or 1)
 *
 *  ARRAY:
 *    newarray, anewarray   create array
 *    iaload, aaload        load from array
 *    iastore, aastore      store to array
 *    arraylength           get array length
 *
 *  CONTROL:
 *    ifeq, ifne, iflt, ifge, ifgt, ifle      branch on int condition
 *    if_icmpeq, if_acmpeq                    compare and branch
 *    goto          unconditional jump
 *    tableswitch   switch (dense cases)
 *    lookupswitch  switch (sparse cases)
 *
 *  STACK:
 *    dup    duplicate top of stack
 *    pop    discard top of stack
 *    swap   swap top two stack values
 *
 *  RETURN:
 *    ireturn, lreturn, freturn, dreturn, areturn, return
 *
 *  EXCEPTION:
 *    athrow       throw exception
 *    Exception table entries define try-catch ranges
 *
 *  ──────────────────────────────────────────────────
 *  STACK FRAME CONCEPT:
 *    Every method has a stack frame:
 *    ┌──────────────────────────┐
 *    │  Operand Stack           │  (values being computed)
 *    │  Local Variables Array   │  (params + locals)
 *    │  Constant Pool Reference │  (class constants)
 *    └──────────────────────────┘
 *
 *  EXAMPLE — what `int add(int a, int b) { return a + b; }`
 *  compiles to:
 *    0: iload_1          // push param 'a' onto stack
 *    1: iload_2          // push param 'b' onto stack
 *    2: iadd             // pop both, push sum
 *    3: ireturn           // return int on top of stack
 *
 *  LAMBDA COMPILATION:
 *    list.forEach(x -> System.out.println(x));
 *    compiles to:
 *      1. A private static method is generated:
 *         private static void lambda$main$0(Object x) {
 *             System.out.println(x);
 *         }
 *      2. invokedynamic calls LambdaMetafactory.metafactory()
 *         which generates a class implementing Consumer<T>
 *         that calls lambda$main$0
 *
 *    This is MORE efficient than anonymous inner classes
 *    because metafactory can optimize at runtime.
 *
 *  STRING CONCATENATION (Java 9+):
 *    String s = "Hello " + name + "!";
 *    compiles to invokedynamic calling:
 *      StringConcatFactory.makeConcatWithConstants(...)
 *    NOT StringBuilder anymore! The JVM can optimize
 *    the strategy at runtime.
 *
 * ============================================================
 */

import java.lang.invoke.*;
import java.lang.invoke.MethodHandles.Lookup;
import java.util.*;

public class Chapter51_Bytecode {

    // Sample class to introspect
    private int value;
    private static int counter = 0;

    public Chapter51_Bytecode(int value) { this.value = value; }

    public int getValue() { return value; }
    private void setValue(int v) { this.value = v; }
    public static int getCounter() { return counter; }

    public int add(int a, int b) { return a + b; }

    public String greet(String name) {
        return "Hello, " + name + "! Value=" + value;
    }

    // ========================================================
    // VarHandle example field (Java 9+)
    // ========================================================
    private volatile int vhField = 0;

    // ========================================================
    // MAIN
    // ========================================================
    public static void main(String[] args) throws Throwable {

        System.out.println("=== CHAPTER 51: BYTECODE & METHODHANDLES ===\n");

        // ====================================================
        // 1. MethodHandles.Lookup — Access Control
        // ====================================================
        System.out.println("--- 1. MethodHandles.Lookup ---\n");

        // Lookup gives you access to find methods, fields, constructors
        // The Lookup object carries the ACCESS CONTEXT of where it was created
        Lookup lookup = MethodHandles.lookup();
        System.out.println("  Lookup class: " + lookup.lookupClass());
        System.out.println("  Lookup modes: " + lookup.lookupModes());

        // Public lookup — can only see public members of public classes
        Lookup publicLookup = MethodHandles.publicLookup();
        System.out.println("  Public lookup modes: " + publicLookup.lookupModes());

        // ====================================================
        // 2. MethodType — Describe Method Signatures
        // ====================================================
        System.out.println("\n--- 2. MethodType ---\n");

        // MethodType = return type + parameter types
        MethodType mt1 = MethodType.methodType(int.class, int.class, int.class);
        System.out.println("  (int, int) → int: " + mt1);

        MethodType mt2 = MethodType.methodType(String.class, String.class);
        System.out.println("  (String) → String: " + mt2);

        MethodType mt3 = MethodType.methodType(void.class);
        System.out.println("  () → void: " + mt3);

        // ====================================================
        // 3. Finding Methods — MethodHandle
        // ====================================================
        System.out.println("\n--- 3. Finding Methods ---\n");

        // Find instance method: add(int, int)
        MethodHandle addHandle = lookup.findVirtual(
            Chapter51_Bytecode.class,
            "add",
            MethodType.methodType(int.class, int.class, int.class)
        );
        System.out.println("  addHandle: " + addHandle);

        // Find static method
        MethodHandle counterHandle = lookup.findStatic(
            Chapter51_Bytecode.class,
            "getCounter",
            MethodType.methodType(int.class)
        );

        // Find constructor
        MethodHandle constructor = lookup.findConstructor(
            Chapter51_Bytecode.class,
            MethodType.methodType(void.class, int.class)
        );
        System.out.println("  constructor: " + constructor);

        // Find private method (works because lookup is from this class)
        MethodHandle setValueHandle = lookup.findVirtual(
            Chapter51_Bytecode.class,
            "setValue",
            MethodType.methodType(void.class, int.class)
        );
        System.out.println("  private setValue found: " + setValueHandle);

        // ====================================================
        // 4. Invoking MethodHandles
        // ====================================================
        System.out.println("\n--- 4. Invoking MethodHandles ---\n");

        Chapter51_Bytecode obj = new Chapter51_Bytecode(42);

        // invoke — polymorphic signature, arguments must match exactly
        int sum = (int) addHandle.invoke(obj, 3, 7);
        System.out.println("  add(3, 7) = " + sum);

        // invokeExact — NO type adaptation, must match precisely
        // You must cast the return to the EXACT type
        int sumExact = (int) addHandle.invokeExact(obj, 3, 7);
        System.out.println("  addExact(3, 7) = " + sumExact);

        // Construct via MethodHandle
        Chapter51_Bytecode constructed = (Chapter51_Bytecode) constructor.invoke(99);
        System.out.println("  Constructed value: " + constructed.getValue());

        // Call greet
        MethodHandle greetHandle = lookup.findVirtual(
            Chapter51_Bytecode.class,
            "greet",
            MethodType.methodType(String.class, String.class)
        );
        String greeting = (String) greetHandle.invoke(obj, "World");
        System.out.println("  greet: " + greeting);

        // ====================================================
        // 5. MethodHandle Transformations
        // ====================================================
        System.out.println("\n--- 5. MethodHandle Transformations ---\n");

        // bindTo — partial application (fix the receiver)
        MethodHandle boundAdd = addHandle.bindTo(obj);
        int r1 = (int) boundAdd.invoke(10, 20);
        System.out.println("  boundAdd(10, 20) = " + r1);

        // insertArguments — fix arguments at specific positions
        MethodHandle addTo5 = MethodHandles.insertArguments(addHandle, 1, 5);
        // Now takes (receiver, int) instead of (receiver, int, int)
        int r2 = (int) addTo5.invoke(obj, 8);
        System.out.println("  addTo5(8) = " + r2 + " [5 + 8]");

        // dropArguments — add ignored parameters
        MethodHandle addWithExtra = MethodHandles.dropArguments(
            boundAdd, 0, String.class);
        int r3 = (int) addWithExtra.invoke("ignored", 2, 3);
        System.out.println("  addWithExtra(\"ignored\", 2, 3) = " + r3);

        // filterReturnValue — transform the result
        MethodHandle stringAdd = MethodHandles.filterReturnValue(
            boundAdd,
            lookup.findStatic(String.class, "valueOf",
                MethodType.methodType(String.class, int.class))
        );
        String r4 = (String) stringAdd.invoke(10, 20);
        System.out.println("  stringAdd(10, 20) = \"" + r4 + "\"");

        // permuteArguments — reorder arguments
        // Swap the two int params of boundAdd
        MethodHandle swapped = MethodHandles.permuteArguments(
            boundAdd,
            MethodType.methodType(int.class, int.class, int.class),
            1, 0  // argument indices
        );
        // add does a+b, so swapping doesn't change sum, but matters for subtraction etc.

        // ====================================================
        // 6. Field Access via MethodHandle
        // ====================================================
        System.out.println("\n--- 6. Field Access ---\n");

        // Getter
        MethodHandle getter = lookup.findGetter(
            Chapter51_Bytecode.class, "value", int.class);
        int fieldVal = (int) getter.invoke(obj);
        System.out.println("  Get 'value': " + fieldVal);

        // Setter
        MethodHandle setter = lookup.findSetter(
            Chapter51_Bytecode.class, "value", int.class);
        setter.invoke(obj, 100);
        System.out.println("  Set 'value' to 100, now: " + obj.getValue());

        // Static field
        MethodHandle staticGetter = lookup.findStaticGetter(
            Chapter51_Bytecode.class, "counter", int.class);
        int cnt = (int) staticGetter.invoke();
        System.out.println("  Static 'counter': " + cnt);

        // ====================================================
        // 7. VarHandle (Java 9+) — Memory-Order-Aware Access
        // ====================================================
        System.out.println("\n--- 7. VarHandle ---\n");

        /*
         * VarHandle provides:
         *   - get/set           (plain access)
         *   - getVolatile/setVolatile
         *   - getOpaque/setOpaque
         *   - getAcquire/setRelease
         *   - compareAndSet
         *   - getAndAdd, getAndSet
         *
         * This is like AtomicFieldUpdater but more general and
         * with explicit memory ordering modes.
         */

        VarHandle vhHandle = lookup.findVarHandle(
            Chapter51_Bytecode.class, "vhField", int.class);

        // Plain set/get
        vhHandle.set(obj, 10);
        System.out.println("  VarHandle get: " + vhHandle.get(obj));

        // Compare and set (atomic)
        boolean casResult = vhHandle.compareAndSet(obj, 10, 20);
        System.out.println("  CAS 10→20: " + casResult + ", value=" + vhHandle.get(obj));

        boolean casFail = vhHandle.compareAndSet(obj, 10, 30);
        System.out.println("  CAS 10→30: " + casFail + " (failed, still " + vhHandle.get(obj) + ")");

        // Atomic add
        int prev = (int) vhHandle.getAndAdd(obj, 5);
        System.out.println("  getAndAdd(5): prev=" + prev + ", now=" + vhHandle.get(obj));

        // Array VarHandle
        VarHandle arrayVH = MethodHandles.arrayElementVarHandle(int[].class);
        int[] arr = {10, 20, 30};
        arrayVH.set(arr, 1, 99);
        System.out.println("  Array VarHandle: " + Arrays.toString(arr));

        // ====================================================
        // 8. MethodHandle vs Reflection
        // ====================================================
        System.out.println("\n--- 8. MethodHandle vs Reflection ---\n");

        System.out.println("  ┌─────────────────┬──────────────┬──────────────────┐");
        System.out.println("  │ Feature          │ Reflection   │ MethodHandle     │");
        System.out.println("  ├─────────────────┼──────────────┼──────────────────┤");
        System.out.println("  │ Speed            │ Slower       │ JIT-optimizable  │");
        System.out.println("  │ Access check     │ Every call   │ Once at lookup   │");
        System.out.println("  │ Type safety      │ Runtime only │ MethodType check │");
        System.out.println("  │ Composable       │ No           │ Yes (transform)  │");
        System.out.println("  │ invokedynamic    │ No           │ Yes              │");
        System.out.println("  │ Lambda support   │ No           │ Built on it      │");
        System.out.println("  └─────────────────┴──────────────┴──────────────────┘");

        // ====================================================
        // 9. SwitchPoint — Optimistic Assumptions
        // ====================================================
        System.out.println("\n--- 9. SwitchPoint ---\n");

        /*
         * SwitchPoint lets you create a "guard" that the JVM can
         * optimize around. While the guard is valid, one target
         * is called. After invalidation, the fallback is called.
         *
         * Used in dynamic language runtimes on the JVM (Nashorn, etc.)
         */

        SwitchPoint sp = new SwitchPoint();

        MethodHandle fastPath = MethodHandles.constant(String.class, "FAST");
        MethodHandle slowPath = MethodHandles.constant(String.class, "SLOW");

        MethodHandle guarded = sp.guardWithTest(fastPath, slowPath);

        System.out.println("  Before invalidation: " + (String) guarded.invokeExact());

        SwitchPoint.invalidateAll(new SwitchPoint[]{sp});
        System.out.println("  After invalidation:  " + (String) guarded.invokeExact());

        // ====================================================
        // 10. Practical: Method Dispatch Table
        // ====================================================
        System.out.println("\n--- 10. Dispatch Table Pattern ---\n");

        Map<String, MethodHandle> dispatch = new HashMap<>();

        MethodHandle printlnHandle = lookup.findVirtual(
            java.io.PrintStream.class, "println",
            MethodType.methodType(void.class, String.class)
        ).bindTo(System.out);

        MethodHandle toUpperHandle = lookup.findVirtual(
            String.class, "toUpperCase",
            MethodType.methodType(String.class)
        );

        MethodHandle toLowerHandle = lookup.findVirtual(
            String.class, "toLowerCase",
            MethodType.methodType(String.class)
        );

        dispatch.put("upper", toUpperHandle);
        dispatch.put("lower", toLowerHandle);

        String input = "Hello World";
        for (Map.Entry<String, MethodHandle> e : dispatch.entrySet()) {
            String result = (String) e.getValue().invoke(input);
            System.out.println("  " + e.getKey() + "(\"" + input + "\") = " + result);
        }

        System.out.println("\n✓ Bytecode & MethodHandles Complete!");
    }
}

/*
 * HOW TO EXPLORE BYTECODE:
 *   javac Chapter51_Bytecode.java
 *   javap -c -p Chapter51_Bytecode           # disassemble
 *   javap -c -p -v Chapter51_Bytecode        # verbose (constant pool)
 *
 * TRY THIS:
 *   1. Write a simple loop and check the bytecode
 *   2. Compare bytecode of method dispatch (virtual vs static)
 *   3. Look at what lambdas compile to
 *   4. Compare try-catch bytecode (exception table)
 *
 * EXERCISES:
 *   1. Create a MethodHandle that composes two string functions.
 *   2. Build a simple interpreter using a Map<String, MethodHandle> dispatch.
 *   3. Use VarHandle.compareAndSet to build a lock-free counter.
 *   4. Examine the bytecode of a switch statement with 5 cases vs 100 cases
 *      (tableswitch vs lookupswitch).
 *
 * NEXT: Chapter 52 — Dynamic Proxies & ClassLoaders
 */
