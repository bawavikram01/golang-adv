/*
 * ============================================================
 *  CHAPTER 61: JNI & NATIVE INTEROP
 * ============================================================
 *  The `native` keyword is part of the Java LANGUAGE. It lets
 *  you call C/C++ code from Java and vice versa. This is how
 *  Java talks to the OS, hardware, and legacy systems.
 *
 *  TOPICS:
 *    1. What Is JNI?
 *    2. The `native` Keyword
 *    3. JNI Workflow (javac → javah/javac -h → C → compile → run)
 *    4. JNI Type Mapping
 *    5. Passing Data Between Java and C
 *    6. Calling Java from C (Callbacks)
 *    7. Error Handling Across Boundaries
 *    8. JNA — Easier Alternative
 *    9. Panama / Foreign Function API (Future)
 *   10. When to Use Native Code
 * ============================================================
 *
 *  WHY JNI EXISTS:
 *    - Access OS APIs not exposed by Java
 *    - Use existing C/C++ libraries
 *    - Performance-critical code (SIMD, GPU)
 *    - Hardware access (serial ports, GPIO)
 *    - Legacy system integration
 *
 *  WHERE JNI IS USED IN THE JDK:
 *    - java.io (file I/O → OS syscalls)
 *    - java.net (sockets → OS networking)
 *    - java.awt (graphics → native windowing)
 *    - java.util.zip (zlib compression)
 *    - sun.misc.Unsafe (direct memory access)
 *    - Every System.* call
 *
 * ============================================================
 */

public class Chapter61_JNI {

    // ========================================================
    // 1. DECLARING NATIVE METHODS
    // ========================================================

    // 'native' = method body is in C/C++ (no Java implementation)
    // Just like 'abstract' — no body, but for a DIFFERENT reason
    public native int addNative(int a, int b);
    public native String greetNative(String name);
    public static native long currentTimeNanos();

    // Load the native library
    static {
        // Looks for libmylib.so (Linux), mylib.dll (Windows), libmylib.dylib (Mac)
        // System.loadLibrary("mylib");

        // Or load by absolute path:
        // System.load("/path/to/libmylib.so");
    }

    // ========================================================
    // 2. JNI WORKFLOW
    // ========================================================
    /*
     * STEP 1: Write Java class with native methods
     * ─────────────────────────────────────────────
     *   public class MyNative {
     *       public native int add(int a, int b);
     *       static { System.loadLibrary("mynative"); }
     *   }
     *
     * STEP 2: Generate header file
     * ────────────────────────────
     *   # Java 10+:
     *   javac -h . MyNative.java
     *   → generates MyNative.h
     *
     *   # Java 8 (deprecated):
     *   javah MyNative
     *
     * STEP 3: Generated header looks like:
     * ─────────────────────────────────────
     *   #include <jni.h>
     *
     *   JNIEXPORT jint JNICALL Java_MyNative_add
     *     (JNIEnv *, jobject, jint, jint);
     *
     *   // Naming: Java_{package}_{class}_{method}
     *   // package separators: . → _
     *   // JNIEnv* = pointer to JNI function table
     *   // jobject = 'this' (for instance methods)
     *   // jclass = class (for static methods)
     *
     * STEP 4: Implement in C
     * ──────────────────────
     *   #include "MyNative.h"
     *
     *   JNIEXPORT jint JNICALL Java_MyNative_add
     *     (JNIEnv *env, jobject obj, jint a, jint b) {
     *       return a + b;
     *   }
     *
     * STEP 5: Compile the native library
     * ───────────────────────────────────
     *   # Linux:
     *   gcc -shared -fPIC -o libmynative.so \
     *       -I${JAVA_HOME}/include \
     *       -I${JAVA_HOME}/include/linux \
     *       MyNative.c
     *
     *   # Mac:
     *   gcc -shared -o libmynative.dylib \
     *       -I${JAVA_HOME}/include \
     *       -I${JAVA_HOME}/include/darwin \
     *       MyNative.c
     *
     *   # Windows (MSVC):
     *   cl /LD MyNative.c /I %JAVA_HOME%\include /I %JAVA_HOME%\include\win32
     *
     * STEP 6: Run
     * ──────────
     *   java -Djava.library.path=. MyNative
     *   # or set LD_LIBRARY_PATH (Linux) / DYLD_LIBRARY_PATH (Mac)
     */

    // ========================================================
    // 3. JNI TYPE MAPPING
    // ========================================================
    /*
     * ┌────────────────┬────────────────┬──────────────┐
     * │ Java Type       │ JNI Type       │ C Type       │
     * ├────────────────┼────────────────┼──────────────┤
     * │ boolean         │ jboolean       │ unsigned char│
     * │ byte            │ jbyte          │ signed char  │
     * │ char            │ jchar          │ unsigned short│
     * │ short           │ jshort         │ short        │
     * │ int             │ jint           │ int (32-bit) │
     * │ long            │ jlong          │ long long    │
     * │ float           │ jfloat         │ float        │
     * │ double          │ jdouble        │ double       │
     * │ void            │ void           │ void         │
     * ├────────────────┼────────────────┼──────────────┤
     * │ Object          │ jobject        │ pointer      │
     * │ String          │ jstring        │ pointer      │
     * │ Class           │ jclass         │ pointer      │
     * │ Throwable       │ jthrowable     │ pointer      │
     * │ int[]           │ jintArray      │ pointer      │
     * │ Object[]        │ jobjectArray   │ pointer      │
     * └────────────────┴────────────────┴──────────────┘
     */

    // ========================================================
    // 4. WORKING WITH STRINGS IN JNI
    // ========================================================
    /*
     * Java Strings are UTF-16. C strings are char* (usually UTF-8).
     * JNI provides conversion functions:
     *
     *   // Java String → C string
     *   const char *cStr = (*env)->GetStringUTFChars(env, jStr, NULL);
     *   // use cStr...
     *   (*env)->ReleaseStringUTFChars(env, jStr, cStr);  // MUST release!
     *
     *   // C string → Java String
     *   jstring result = (*env)->NewStringUTF(env, "hello");
     *   return result;
     *
     * MEMORY RULE: Every Get* must have a matching Release*
     * Forgetting Release → native memory leak!
     */

    // ========================================================
    // 5. WORKING WITH ARRAYS
    // ========================================================
    /*
     *   // Java int[] → C int*
     *   jint *elements = (*env)->GetIntArrayElements(env, jArray, NULL);
     *   jsize length = (*env)->GetArrayLength(env, jArray);
     *
     *   for (int i = 0; i < length; i++) {
     *       elements[i] *= 2;  // modify
     *   }
     *
     *   // Release (0 = copy back, JNI_ABORT = don't copy back)
     *   (*env)->ReleaseIntArrayElements(env, jArray, elements, 0);
     *
     *   // Create new Java array
     *   jintArray result = (*env)->NewIntArray(env, 10);
     *   jint buf[10] = {1,2,3,4,5,6,7,8,9,10};
     *   (*env)->SetIntArrayRegion(env, result, 0, 10, buf);
     *   return result;
     *
     * HIGH-PERFORMANCE:
     *   GetPrimitiveArrayCritical / ReleasePrimitiveArrayCritical
     *   → May return direct pointer (no copy) but disables GC!
     *   → MUST be short-lived, no JNI calls between get/release
     */

    // ========================================================
    // 6. CALLING JAVA FROM C (Callbacks)
    // ========================================================
    /*
     *   // Get class, method ID, and call
     *   jclass cls = (*env)->GetObjectClass(env, obj);
     *   jmethodID mid = (*env)->GetMethodID(env, cls, "myMethod", "(I)V");
     *   (*env)->CallVoidMethod(env, obj, mid, 42);
     *
     * METHOD SIGNATURES (JNI format):
     *   (I)V           → void method(int)
     *   (II)I          → int method(int, int)
     *   (Ljava/lang/String;)V  → void method(String)
     *   ()Ljava/lang/String;   → String method()
     *   ([I)V          → void method(int[])
     *   (ILjava/lang/String;)Z → boolean method(int, String)
     *
     * TYPE CODES:
     *   B=byte C=char D=double F=float I=int J=long
     *   S=short Z=boolean V=void L=object [=array
     *
     * USE javap -s MyClass TO SEE SIGNATURES:
     *   javap -s -p MyClass
     *   → shows descriptor for each method
     */

    // ========================================================
    // 7. ERROR HANDLING
    // ========================================================
    /*
     * C HAS NO EXCEPTIONS. JNI has exception-checking functions:
     *
     *   // After a JNI call, check for pending exception:
     *   if ((*env)->ExceptionCheck(env)) {
     *       (*env)->ExceptionDescribe(env);  // print to stderr
     *       (*env)->ExceptionClear(env);     // clear the exception
     *       return;  // or handle the error
     *   }
     *
     *   // Throw a Java exception from C:
     *   jclass excClass = (*env)->FindClass(env, "java/lang/RuntimeException");
     *   (*env)->ThrowNew(env, excClass, "Something went wrong in native code");
     *   return;  // MUST return after throwing!
     *
     * CRITICAL RULE:
     *   After ThrowNew(), you MUST return from the native method.
     *   The exception is "pending" — it will be thrown when control
     *   returns to Java.
     */

    // ========================================================
    // 8. JNA — Easier Alternative
    // ========================================================
    /*
     * JNA (Java Native Access) lets you call C functions WITHOUT
     * writing any C code or header files. It uses libffi at runtime.
     *
     *   import com.sun.jna.*;
     *   import com.sun.jna.platform.win32.*;
     *
     *   // Declare the native library interface
     *   public interface CLib extends Library {
     *       CLib INSTANCE = Native.load("c", CLib.class);
     *       int printf(String format, Object... args);
     *       int getpid();
     *   }
     *
     *   // Use it — no C code needed!
     *   CLib.INSTANCE.printf("Hello from C! PID=%d\n", CLib.INSTANCE.getpid());
     *
     * JNA vs JNI:
     *   ┌──────────┬──────────────────┬──────────────────┐
     *   │ Aspect    │ JNI              │ JNA              │
     *   ├──────────┼──────────────────┼──────────────────┤
     *   │ C code    │ Required         │ Not required     │
     *   │ Speed     │ Direct call      │ Overhead (libffi)│
     *   │ Setup     │ Complex          │ Simple           │
     *   │ Safety    │ Can crash JVM    │ Safer            │
     *   │ Callbacks │ Complex          │ Easy             │
     *   └──────────┴──────────────────┴──────────────────┘
     *
     * JNR-FFI: Similar to JNA but faster (used by JRuby).
     */

    // ========================================================
    // 9. PANAMA — Foreign Function & Memory API (Java 22+)
    // ========================================================
    /*
     * Project Panama provides SUPPORTED alternatives to JNI:
     *
     * FOREIGN FUNCTION API: Call C functions without JNI
     *
     *   // Look up C function
     *   Linker linker = Linker.nativeLinker();
     *   SymbolLookup stdlib = linker.defaultLookup();
     *   MethodHandle strlen = linker.downcallHandle(
     *       stdlib.find("strlen").orElseThrow(),
     *       FunctionDescriptor.of(JAVA_LONG, ADDRESS)
     *   );
     *
     *   // Call it!
     *   try (Arena arena = Arena.ofConfined()) {
     *       MemorySegment cStr = arena.allocateFrom("Hello");
     *       long len = (long) strlen.invoke(cStr);
     *       System.out.println("strlen = " + len);  // 5
     *   }
     *
     * FOREIGN MEMORY API: Safe off-heap memory
     *
     *   try (Arena arena = Arena.ofConfined()) {
     *       // Allocate off-heap memory
     *       MemorySegment segment = arena.allocate(100);
     *       segment.set(JAVA_INT, 0, 42);
     *       int val = segment.get(JAVA_INT, 0);
     *
     *       // Memory is automatically freed when arena closes
     *   }
     *
     * JEXTRACT: Auto-generate Java bindings from C headers
     *   jextract --source -t com.example -l mylib myheader.h
     *   → Generates Java code to call everything in myheader.h
     *
     * Panama is the FUTURE — it will eventually replace JNI.
     * Benefits: no C code needed, safe memory, supported API.
     */

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== CHAPTER 61: JNI & NATIVE INTEROP ===\n");

        // --- 1. The native keyword ---
        System.out.println("--- 1. The 'native' Keyword ---\n");
        System.out.println("  'native' is a Java keyword like 'abstract'.");
        System.out.println("  It means: \"this method is implemented in C/C++\".");
        System.out.println("  You can't call it without loading the native library first.");
        System.out.println("  UnsatisfiedLinkError if library not loaded.\n");

        // --- 2. Where JDK uses native ---
        System.out.println("--- 2. Native Methods in the JDK ---\n");

        // Count native methods in some JDK classes
        String[] classes = {"java.lang.Object", "java.lang.System",
                           "java.lang.Thread", "java.lang.ClassLoader"};
        for (String className : classes) {
            try {
                Class<?> clazz = Class.forName(className);
                long nativeCount = java.util.Arrays.stream(clazz.getDeclaredMethods())
                    .filter(m -> java.lang.reflect.Modifier.isNative(m.getModifiers()))
                    .count();
                System.out.println("  " + clazz.getSimpleName() + ": "
                    + nativeCount + " native methods");
            } catch (ClassNotFoundException e) { /* skip */ }
        }

        // Show some native methods
        System.out.println("\n  Key native methods in JDK:");
        System.out.println("    Object.hashCode()         → native (identity hash)");
        System.out.println("    Object.clone()            → native (memory copy)");
        System.out.println("    Object.wait()/notify()    → native (OS monitor)");
        System.out.println("    System.currentTimeMillis()→ native (OS clock)");
        System.out.println("    System.arraycopy()        → native (optimized copy)");
        System.out.println("    Thread.start0()           → native (OS thread)");
        System.out.println("    Thread.sleep()            → native (OS sleep)");

        // --- 3. JNI build process ---
        System.out.println("\n--- 3. Build Process ---\n");
        System.out.println("  1. javac -h . MyNative.java   → compile + generate .h");
        System.out.println("  2. Write MyNative.c            → implement native methods");
        System.out.println("  3. gcc -shared -fPIC \\");
        System.out.println("       -I$JAVA_HOME/include \\");
        System.out.println("       -I$JAVA_HOME/include/linux \\");
        System.out.println("       -o libmynative.so MyNative.c");
        System.out.println("  4. java -Djava.library.path=. MyNative");

        // --- 4. Common pitfalls ---
        System.out.println("\n--- 4. JNI Pitfalls ---\n");
        System.out.println("  ❌ Forgetting ReleaseStringUTFChars → memory leak");
        System.out.println("  ❌ Forgetting ReleaseIntArrayElements → memory leak");
        System.out.println("  ❌ Caching JNIEnv* across threads → crash");
        System.out.println("     (JNIEnv is per-thread! Use JavaVM* instead)");
        System.out.println("  ❌ Not checking for exceptions after JNI calls");
        System.out.println("  ❌ Continuing after ThrowNew → undefined behavior");
        System.out.println("  ❌ Using local references after method returns");
        System.out.println("     (use NewGlobalRef for persistent references)");
        System.out.println("  ❌ Buffer overflows in C → JVM crash (segfault)");
        System.out.println("  ❌ Wrong method signature → NoSuchMethodError");

        // --- 5. Alternatives comparison ---
        System.out.println("\n--- 5. Native Interop Options ---\n");
        System.out.println("  ┌───────────────┬────────────┬──────────┬──────────────┐");
        System.out.println("  │ Technology     │ C Code?    │ Speed    │ Complexity   │");
        System.out.println("  ├───────────────┼────────────┼──────────┼──────────────┤");
        System.out.println("  │ JNI            │ Required   │ Fastest  │ High         │");
        System.out.println("  │ JNA            │ No         │ Good     │ Low          │");
        System.out.println("  │ JNR-FFI        │ No         │ Better   │ Low          │");
        System.out.println("  │ Panama (22+)   │ No         │ Fast     │ Medium       │");
        System.out.println("  │ ProcessBuilder │ No         │ Slow     │ Low          │");
        System.out.println("  └───────────────┴────────────┴──────────┴──────────────┘");

        // --- 6. System.loadLibrary details ---
        System.out.println("\n--- 6. Library Loading ---\n");
        System.out.println("  System.loadLibrary(\"mylib\") searches:");
        System.out.println("    1. java.library.path system property");
        System.out.println("    2. OS-specific paths:");
        System.out.println("       Linux:   LD_LIBRARY_PATH, /usr/lib, /lib");
        System.out.println("       Mac:     DYLD_LIBRARY_PATH, /usr/local/lib");
        System.out.println("       Windows: PATH, system32");
        System.out.println("    3. File naming:");
        System.out.println("       loadLibrary(\"foo\") →");
        System.out.println("         Linux:   libfoo.so");
        System.out.println("         Mac:     libfoo.dylib");
        System.out.println("         Windows: foo.dll");

        System.out.println("\n  java.library.path = " +
            System.getProperty("java.library.path", "").split(":")[0] + "...");

        System.out.println("\n✓ JNI & Native Interop Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Write a complete JNI example:
 *    - Java class with native int add(int, int)
 *    - Generate header with javac -h
 *    - Implement in C
 *    - Compile and run
 * 2. Call the C function getpid() from Java using JNI.
 * 3. Try JNA: add the dependency and call a C library function
 *    without any C code.
 * 4. If you have Java 22+: use the Foreign Function API to call
 *    strlen() from the C standard library.
 *
 * NEXT: Chapter 62 — JVM Internals Deep Dive
 */
