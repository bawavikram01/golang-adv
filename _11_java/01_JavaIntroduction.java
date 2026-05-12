/*
 * ============================================================
 *  CHAPTER 01: JAVA INTRODUCTION & SETUP
 * ============================================================
 *
 *  WHAT IS JAVA?
 *  -------------
 *  Java is a high-level, object-oriented, platform-independent
 *  programming language created by James Gosling at Sun Microsystems
 *  in 1995 (now owned by Oracle).
 *
 *  KEY FEATURES OF JAVA:
 *  1. Platform Independent  - "Write Once, Run Anywhere" (WORA)
 *  2. Object-Oriented       - Everything revolves around objects
 *  3. Strongly Typed         - Every variable must have a type
 *  4. Automatic Memory Mgmt  - Garbage Collector handles memory
 *  5. Multithreaded          - Built-in support for threads
 *  6. Secure                 - No pointers, bytecode verification
 *  7. Robust                 - Exception handling, strong type checking
 *
 *  HOW JAVA WORKS — THE 3 PILLARS:
 *  ================================
 *
 *  1. JDK (Java Development Kit)
 *     - The FULL toolkit for developers
 *     - Contains: JRE + compiler (javac) + debugger + tools
 *     - YOU need this to WRITE and COMPILE Java code
 *
 *  2. JRE (Java Runtime Environment)
 *     - Contains: JVM + core libraries
 *     - YOU need this to RUN Java programs
 *     - End users only need JRE, not JDK
 *
 *  3. JVM (Java Virtual Machine)
 *     - The ENGINE that runs Java bytecode
 *     - Makes Java platform-independent
 *     - Different JVM for each OS, but same bytecode runs on all
 *
 *  COMPILATION FLOW:
 *  =================
 *
 *    YourCode.java  --[javac compiler]-->  YourCode.class (bytecode)
 *                                              |
 *                                         [JVM loads it]
 *                                              |
 *                                    Machine Code (runs on OS)
 *
 *    Step 1: You write .java file (source code)
 *    Step 2: javac compiles it to .class file (bytecode)
 *    Step 3: JVM interprets/compiles bytecode to machine code
 *    Step 4: Program runs on your OS
 *
 *  WHY BYTECODE?
 *  - .class files are NOT tied to any OS
 *  - Any machine with a JVM can run them
 *  - This is how "Write Once, Run Anywhere" works
 *
 *  ANATOMY OF A JAVA PROGRAM:
 *  ==========================
 *
 *  Every Java program needs:
 *  1. A class (the container)
 *  2. A main method (the entry point)
 *
 *  Rules:
 *  - File name MUST match the public class name (case-sensitive)
 *  - This file is "01_JavaIntroduction.java" but class name can't
 *    start with a number, so we name it properly
 *  - Every statement ends with a semicolon ;
 *  - Java is CASE-SENSITIVE: "Hello" != "hello"
 *  - Code blocks are enclosed in { }
 *
 *  HOW TO COMPILE AND RUN THIS FILE:
 *  ==================================
 *  Terminal:
 *    $ javac 01_JavaIntroduction.java
 *    $ java Chapter01_JavaIntroduction
 *
 * ============================================================
 */

// This is a single-line comment

/*
 * This is a multi-line comment.
 * Used for longer explanations.
 */

/**
 * This is a Javadoc comment.
 * Used to generate documentation.
 * It describes classes, methods, and fields.
 */

// "public" = accessible from anywhere
// "class"  = defines a class (blueprint)
// The class name MUST match the file name (for public classes)
public class Chapter01_JavaIntroduction {

    // "public"        = accessible from anywhere
    // "static"        = belongs to class, not an instance
    // "void"          = returns nothing
    // "main"          = the method name JVM looks for
    // "String[] args" = command-line arguments
    public static void main(String[] args) {

        // ---- PRINTING OUTPUT ----

        // println = print + new line
        System.out.println("Hello, World! Welcome to Java!");

        // print = no new line after
        System.out.print("This is on ");
        System.out.print("the same line.\n"); // \n = manual new line

        // printf = formatted output (like C)
        String name = "Vikram";
        int age = 25;
        System.out.printf("Name: %s, Age: %d%n", name, age);
        // %s = string, %d = integer, %f = float, %n = newline

        // ---- COMMON ESCAPE SEQUENCES ----
        System.out.println("\n--- Escape Sequences ---");
        System.out.println("New line: \\n  -> creates a new line");
        System.out.println("Tab: \tThis is tabbed");
        System.out.println("Backslash: \\\\  prints \\");
        System.out.println("Double quote: \\\"  prints \"");
        System.out.println("Single quote: \\'  prints \'");

        // ---- ANATOMY RECAP ----
        System.out.println("\n--- Key Points ---");
        System.out.println("1. Every Java program needs a class");
        System.out.println("2. Execution starts from main()");
        System.out.println("3. System.out.println() prints to console");
        System.out.println("4. Java is case-sensitive");
        System.out.println("5. Every statement ends with ;");
        System.out.println("6. File name must match public class name");

        // ---- COMMAND LINE ARGUMENTS ----
        System.out.println("\n--- Command Line Arguments ---");
        // args[] contains arguments passed when running:
        // java Chapter01_JavaIntroduction arg1 arg2
        if (args.length > 0) {
            System.out.println("You passed " + args.length + " arguments:");
            for (int i = 0; i < args.length; i++) {
                System.out.println("  args[" + i + "] = " + args[i]);
            }
        } else {
            System.out.println("No command-line arguments passed.");
            System.out.println("Try: java Chapter01_JavaIntroduction hello world");
        }
    }
}

/*
 * ============================================================
 *  EXERCISES — Try these yourself!
 * ============================================================
 *
 *  1. Change "Hello, World!" to your own greeting
 *  2. Print your name, age, and city using println
 *  3. Print a triangle pattern using println:
 *       *
 *       **
 *       ***
 *       ****
 *  4. Use printf to print: "I am [name], [age] years old, [height] meters tall"
 *     (use %s for name, %d for age, %.2f for height)
 *  5. Pass your name as a command-line argument and print it
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 02 — Variables & Data Types
 * ============================================================
 */
